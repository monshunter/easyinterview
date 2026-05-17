// Command api is the easyinterview HTTP entry point. It owns the
// minimal runtime-config endpoint required by secrets-and-config spec
// §6 C-1 / C-2 / C-6 and is the first allowlisted callsite of os.Getenv
// outside internal/platform/config (spec §4.1).
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/bootstrap"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	apidebriefs "github.com/monshunter/easyinterview/backend/internal/api/debriefs"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	apijobs "github.com/monshunter/easyinterview/backend/internal/api/jobs"
	apipractice "github.com/monshunter/easyinterview/backend/internal/api/practice"
	apireports "github.com/monshunter/easyinterview/backend/internal/api/reports"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	domaindebrief "github.com/monshunter/easyinterview/backend/internal/debrief"
	domainjobs "github.com/monshunter/easyinterview/backend/internal/jobs"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
	"github.com/monshunter/easyinterview/backend/internal/platform/secrets"
	domainpractice "github.com/monshunter/easyinterview/backend/internal/practice"
	privacyrunner "github.com/monshunter/easyinterview/backend/internal/privacy/runner"
	domainresume "github.com/monshunter/easyinterview/backend/internal/resume"
	resumehandler "github.com/monshunter/easyinterview/backend/internal/resume/handler"
	resumejobs "github.com/monshunter/easyinterview/backend/internal/resume/jobs"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	domainreview "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	storeai "github.com/monshunter/easyinterview/backend/internal/store/ai"
	storedebrief "github.com/monshunter/easyinterview/backend/internal/store/debrief"
	storejobs "github.com/monshunter/easyinterview/backend/internal/store/jobs"
	storepractice "github.com/monshunter/easyinterview/backend/internal/store/practice"
	storereview "github.com/monshunter/easyinterview/backend/internal/store/review"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
	"github.com/monshunter/easyinterview/backend/internal/targetjob/urlfetch"
	uploadhandler "github.com/monshunter/easyinterview/backend/internal/upload/handler"
	"github.com/monshunter/easyinterview/backend/internal/upload/objectstore"
	uploadservice "github.com/monshunter/easyinterview/backend/internal/upload/service"
	uploadstore "github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func main() {
	var (
		dumpConfig bool
		configDir  string
	)
	flag.BoolVar(&dumpConfig, "dump-config", false, "print the merged configuration as JSON and exit")
	flag.StringVar(&configDir, "config-dir", "config", "directory containing config.yaml + {APP_ENV}.yaml")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}

	loader, err := config.LoadCanonical(config.CanonicalOptions{
		AppEnv:       appEnv,
		ConfigDir:    configDir,
		SecretSource: secrets.EnvSecretSource{},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: load config: %v\n", err)
		os.Exit(1)
	}

	if err := loader.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "api: config validation failed: %v\n", err)
		os.Exit(1)
	}

	if dumpConfig {
		dump := map[string]any{
			"app.env":                     loader.GetString("app.env"),
			"app.listenAddr":              loader.GetString("app.listenAddr"),
			"log.level":                   loader.GetString("log.level"),
			"runtime.appVersion":          loader.GetString("runtime.appVersion"),
			"runtime.defaultUiLanguage":   loader.GetString("runtime.defaultUiLanguage"),
			"featureFlag.source":          loader.GetString("featureFlag.source"),
			"async.queueWeights.critical": loader.GetInt("async.queueWeights.critical"),
			"async.queueWeights.default":  loader.GetInt("async.queueWeights.default"),
			"async.queueWeights.low":      loader.GetInt("async.queueWeights.low"),
		}
		_ = json.NewEncoder(os.Stdout).Encode(dump)
		return
	}

	flagsClient, err := buildFlagsClient(loader, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: feature flag init: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", loader.GetString("database.url"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: database init: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	authService, mailDispatcher, err := buildAuthService(loader, db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: auth init: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mailDispatcher.Shutdown(shutdownCtx)
	}()

	uploadRoutes, err := buildUploadRoutes(loader, db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: upload runtime init: %v\n", err)
		os.Exit(1)
	}
	targetJobRuntime, err := buildTargetJobRuntime(loader, db, logger, uploadRoutes.Service)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: target job runtime init: %v\n", err)
		os.Exit(1)
	}
	resumeRuntime, err := buildResumeRuntime(loader, db, logger, uploadRoutes, targetJobRuntime.AI.Client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: resume runtime init: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = targetJobRuntime.Shutdown(shutdownCtx)
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = resumeRuntime.Shutdown(shutdownCtx)
	}()
	practiceRoutes, err := buildPracticeRoutes(loader, db, targetJobRuntime.AI.Client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: practice runtime init: %v\n", err)
		os.Exit(1)
	}
	reportRuntime, err := buildReportRuntime(loader, db, logger, targetJobRuntime.AI.Client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: report runtime init: %v\n", err)
		os.Exit(1)
	}
	jobsRoutes := buildJobsRoutes(db)
	debriefRoutes, err := buildDebriefRoutes(loader, db, targetJobRuntime.AI.Client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: debrief runtime init: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = reportRuntime.Shutdown(shutdownCtx)
	}()
	srv := &http.Server{
		Addr:              loader.GetString("app.listenAddr"),
		Handler:           buildAPIHandlerWithUploadReportDebriefJobsAndHandlers(loader, flagsClient, authService, targetJobRuntime.Handler, practiceRoutes, uploadRoutes, resumeRuntime.Routes(), reportRuntime.Routes(), debriefRoutes, jobsRoutes),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	targetJobRuntime.Start(ctx)
	resumeRuntime.Start(ctx)
	reportRuntime.Start(ctx)

	go func() {
		logger.Info("api: listening", "addr", srv.Addr, "env", loader.AppEnv())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api: serve failed", "error", err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func buildAuthService(loader *config.Loader, db *sql.DB) (*auth.PasswordlessService, *auth.BackgroundMailDispatcher, error) {
	challengePepper := strings.TrimSpace(loader.GetSecret("auth.challengeTokenPepper").Reveal())
	sessionCookieSecret := strings.TrimSpace(loader.GetSecret("auth.sessionCookieSecret").Reveal())
	var missing []string
	if challengePepper == "" {
		missing = append(missing, "AUTH_CHALLENGE_TOKEN_PEPPER")
	}
	if sessionCookieSecret == "" {
		missing = append(missing, "SESSION_COOKIE_SECRET")
	}
	if len(missing) > 0 {
		return nil, nil, fmt.Errorf("missing required auth secret(s): %s", strings.Join(missing, ", "))
	}
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "/api/v1/auth/email/verify"})
	dispatcher := auth.NewBackgroundMailDispatcher(auth.BackgroundMailDispatcherOptions{Writer: sink})
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               auth.NewSQLStore(db),
		Dispatcher:          dispatcher,
		DeliverySecrets:     sink,
		ChallengePepper:     challengePepper,
		SessionCookieSecret: sessionCookieSecret,
	})
	return service, dispatcher, nil
}

func buildAPIHandler(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.PasswordlessService, db *sql.DB) http.Handler {
	upload, _ := buildUploadRoutes(loader, db)
	return buildAPIHandlerWithUploadAndHandlers(loader, flagsClient, authService, buildTargetJobHandler(loader, targetjob.NewSQLStore(db)), practiceRoutes{}, upload, resumeRoutes{})
}

func buildAPIHandlerWithTargetJobHandler(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.PasswordlessService, targetJobHandler *targetjob.Handler) http.Handler {
	return buildAPIHandlerWithHandlers(loader, flagsClient, authService, targetJobHandler, practiceRoutes{})
}

type practiceRoutes struct {
	Handler     *apipractice.Handler
	Idempotency *idempotency.Middleware
}

type reportRoutes struct {
	Handler *apireports.Handler
}

type debriefRoutes struct {
	Handler     *apidebriefs.Handler
	Idempotency *idempotency.Middleware
}

type jobsRoutes struct {
	Handler *apijobs.Handler
}

type discardAIAuditWriter struct{}

func (discardAIAuditWriter) WriteAuditEvent(context.Context, aiclient.AuditEventRow) error {
	return nil
}

type uploadRoutes struct {
	Handler     *uploadhandler.Handler
	Idempotency *idempotency.Middleware
	Service     *uploadservice.Service
	Objects     objectstore.ObjectStore
}

type resumeRoutes struct {
	Handler     *resumehandler.Handler
	Idempotency *idempotency.Middleware
}

func buildAPIHandlerWithHandlers(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.PasswordlessService, targetJobHandler *targetjob.Handler, practice practiceRoutes) http.Handler {
	return buildAPIHandlerWithUploadAndHandlers(loader, flagsClient, authService, targetJobHandler, practice, uploadRoutes{}, resumeRoutes{})
}

func buildAPIHandlerWithUploadAndHandlers(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.PasswordlessService, targetJobHandler *targetjob.Handler, practice practiceRoutes, upload uploadRoutes, resume resumeRoutes) http.Handler {
	return buildAPIHandlerWithUploadReportAndHandlers(loader, flagsClient, authService, targetJobHandler, practice, upload, resume, reportRoutes{})
}

func buildAPIHandlerWithUploadReportAndHandlers(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.PasswordlessService, targetJobHandler *targetjob.Handler, practice practiceRoutes, upload uploadRoutes, resume resumeRoutes, reports reportRoutes) http.Handler {
	return buildAPIHandlerWithUploadReportDebriefAndHandlers(loader, flagsClient, authService, targetJobHandler, practice, upload, resume, reports, debriefRoutes{})
}

func buildAPIHandlerWithUploadReportDebriefAndHandlers(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.PasswordlessService, targetJobHandler *targetjob.Handler, practice practiceRoutes, upload uploadRoutes, resume resumeRoutes, reports reportRoutes, debrief debriefRoutes) http.Handler {
	return buildAPIHandlerWithUploadReportDebriefJobsAndHandlers(loader, flagsClient, authService, targetJobHandler, practice, upload, resume, reports, debrief, jobsRoutes{})
}

func buildAPIHandlerWithUploadReportDebriefJobsAndHandlers(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.PasswordlessService, targetJobHandler *targetjob.Handler, practice practiceRoutes, upload uploadRoutes, resume resumeRoutes, reports reportRoutes, debrief debriefRoutes, jobs jobsRoutes) http.Handler {
	mux := http.NewServeMux()
	authHandler := auth.NewHandler(auth.HandlerOptions{
		Passwordless: authService,
		CookiePolicy: pointer(auth.CookiePolicyForAppEnv(loader.AppEnv())),
	})
	mux.Handle("POST /api/v1/auth/email/start", auth.SessionMiddleware(authService, "startAuthEmailChallenge", http.HandlerFunc(authHandler.StartAuthEmailChallenge)))
	mux.Handle("GET /api/v1/auth/email/verify", auth.SessionMiddleware(authService, "verifyAuthEmailChallenge", http.HandlerFunc(authHandler.VerifyAuthEmailChallenge)))
	mux.Handle("POST /api/v1/auth/logout", auth.SessionMiddleware(authService, "logout", http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /api/v1/me", auth.SessionMiddleware(authService, "getMe", http.HandlerFunc(authHandler.GetMe)))
	mux.Handle("DELETE /api/v1/me", auth.SessionMiddleware(authService, "deleteMe", http.HandlerFunc(authHandler.DeleteMe)))
	mux.Handle("GET /api/v1/runtime-config", auth.SessionMiddleware(authService, "getRuntimeConfig", config.NewRuntimeConfigHandler(config.RuntimeConfigHandlerOptions{
		Loader: loader,
		Flags:  flagsClient,
		FlagContextFunc: func(r *http.Request) featureflag.FlagContext {
			return featureflag.FlagContext{AppEnv: loader.AppEnv()}
		},
		SessionResolver: authService.RuntimeConfigSessionResolver(),
	})))
	if upload.Handler != nil {
		createUploadPresign := http.HandlerFunc(upload.Handler.CreateUploadPresign)
		if upload.Idempotency != nil {
			createUploadPresign = upload.Idempotency.Handler("upload", "createUploadPresign", requestUserFromContext, createUploadPresign).ServeHTTP
		}
		mux.Handle("POST /api/v1/uploads/presign", auth.SessionMiddleware(authService, "createUploadPresign", createUploadPresign))
	}
	if resume.Handler != nil {
		registerResume := http.HandlerFunc(resume.Handler.RegisterResume)
		if resume.Idempotency != nil {
			registerResume = resume.Idempotency.Handler("resume", "registerResume", requestUserFromContext, registerResume).ServeHTTP
		}
		confirmStructuredMaster := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.ConfirmResumeStructuredMaster(w, r, r.PathValue("resumeAssetId"))
		})
		if resume.Idempotency != nil {
			confirmStructuredMaster = requireIdempotencyKey(http.StatusUnprocessableEntity, resume.Idempotency.Handler("resume", "confirmResumeStructuredMaster", requestUserFromContext, confirmStructuredMaster)).ServeHTTP
		}
		updateResumeVersion := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.UpdateResumeVersion(w, r, r.PathValue("resumeVersionId"))
		})
		if resume.Idempotency != nil {
			updateResumeVersion = requireIdempotencyKey(http.StatusUnprocessableEntity, resume.Idempotency.Handler("resume", "updateResumeVersion", requestUserFromContext, updateResumeVersion)).ServeHTTP
		}
		branchResumeVersion := http.HandlerFunc(resume.Handler.BranchResumeVersion)
		if resume.Idempotency != nil {
			branchResumeVersion = requireIdempotencyKey(http.StatusUnprocessableEntity, resume.Idempotency.Handler("resume", "branchResumeVersion", requestUserFromContext, branchResumeVersion)).ServeHTTP
		}
		mux.Handle("GET /api/v1/resumes", auth.SessionMiddleware(authService, "listResumes", http.HandlerFunc(resume.Handler.ListResumes)))
		mux.Handle("POST /api/v1/resumes", auth.SessionMiddleware(authService, "registerResume", registerResume))
		mux.Handle("POST /api/v1/resume-versions", auth.SessionMiddleware(authService, "branchResumeVersion", branchResumeVersion))
		mux.Handle("GET /api/v1/resumes/{resumeAssetId}", auth.SessionMiddleware(authService, "getResume", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.GetResume(w, r, r.PathValue("resumeAssetId"))
		})))
		mux.Handle("GET /api/v1/resume-versions/{resumeVersionId}", auth.SessionMiddleware(authService, "getResumeVersion", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.GetResumeVersion(w, r, r.PathValue("resumeVersionId"))
		})))
		mux.Handle("PATCH /api/v1/resume-versions/{resumeVersionId}", auth.SessionMiddleware(authService, "updateResumeVersion", updateResumeVersion))
		mux.Handle("GET /api/v1/resumes/{resumeAssetId}/versions", auth.SessionMiddleware(authService, "listResumeVersions", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.ListResumeVersions(w, r, r.PathValue("resumeAssetId"))
		})))
		mux.Handle("POST /api/v1/resumes/{resumeAssetId}/structured-master", auth.SessionMiddleware(authService, "confirmResumeStructuredMaster", confirmStructuredMaster))
	}
	if reports.Handler != nil {
		mux.Handle("GET /api/v1/reports/{reportId}", auth.SessionMiddleware(authService, "getFeedbackReport", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reports.Handler.GetFeedbackReport(w, r, r.PathValue("reportId"))
		})))
		mux.Handle("GET /api/v1/targets/{targetJobId}/reports", auth.SessionMiddleware(authService, "listTargetJobReports", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reports.Handler.ListTargetJobReports(w, r, r.PathValue("targetJobId"))
		})))
	}
	if debrief.Handler != nil {
		createDebrief := http.HandlerFunc(debrief.Handler.CreateDebrief)
		if debrief.Idempotency != nil {
			createDebrief = debrief.Idempotency.Handler("debrief", "createDebrief", requestUserFromContext, createDebrief).ServeHTTP
		}
		mux.Handle("POST /api/v1/debriefs", auth.SessionMiddleware(authService, "createDebrief", createDebrief))
		mux.Handle("POST /api/v1/debriefs/question-suggestions", auth.SessionMiddleware(authService, "suggestDebriefQuestions", http.HandlerFunc(debrief.Handler.SuggestDebriefQuestions)))
		mux.Handle("GET /api/v1/debriefs/{debriefId}", auth.SessionMiddleware(authService, "getDebrief", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			debrief.Handler.GetDebrief(w, r, r.PathValue("debriefId"))
		})))
	}
	if jobs.Handler != nil {
		mux.Handle("GET /api/v1/jobs/{jobId}", auth.SessionMiddleware(authService, "getJob", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jobs.Handler.GetJob(w, r, r.PathValue("jobId"))
		})))
	}
	mux.Handle("GET /api/v1/targets", auth.SessionMiddleware(authService, "listTargetJobs", http.HandlerFunc(targetJobHandler.ListTargetJobs)))
	mux.Handle("POST /api/v1/targets/import", auth.SessionMiddleware(authService, "importTargetJob", http.HandlerFunc(targetJobHandler.ImportTargetJob)))
	mux.Handle("GET /api/v1/targets/{targetJobId}", auth.SessionMiddleware(authService, "getTargetJob", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetJobHandler.GetTargetJob(w, r, r.PathValue("targetJobId"))
	})))
	mux.Handle("PATCH /api/v1/targets/{targetJobId}", auth.SessionMiddleware(authService, "updateTargetJob", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetJobHandler.UpdateTargetJob(w, r, r.PathValue("targetJobId"))
	})))
	if practice.Handler != nil {
		createPracticePlan := http.HandlerFunc(practice.Handler.CreatePracticePlan)
		if practice.Idempotency != nil {
			createPracticePlan = practice.Idempotency.Handler("practice", "createPracticePlan", requestUserFromContext, createPracticePlan).ServeHTTP
		}
		mux.Handle("POST /api/v1/practice/plans", auth.SessionMiddleware(authService, "createPracticePlan", createPracticePlan))
		mux.Handle("GET /api/v1/practice/plans/{planId}", auth.SessionMiddleware(authService, "getPracticePlan", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			practice.Handler.GetPracticePlan(w, r, r.PathValue("planId"))
		})))
		mux.Handle("GET /api/v1/practice/sessions", auth.SessionMiddleware(authService, "listPracticeSessions", http.HandlerFunc(practice.Handler.ListPracticeSessions)))
		mux.Handle("POST /api/v1/practice/sessions", auth.SessionMiddleware(authService, "startPracticeSession", http.HandlerFunc(practice.Handler.StartPracticeSession)))
		mux.Handle("GET /api/v1/practice/sessions/{sessionId}", auth.SessionMiddleware(authService, "getPracticeSession", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			practice.Handler.GetPracticeSession(w, r, r.PathValue("sessionId"))
		})))
		createPracticeVoiceTurn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			practice.Handler.CreatePracticeVoiceTurn(w, r, r.PathValue("sessionId"))
		})
		if practice.Idempotency != nil {
			createPracticeVoiceTurn = practice.Idempotency.Handler("practice", "createPracticeVoiceTurn", requestUserFromContext, createPracticeVoiceTurn).ServeHTTP
		}
		mux.Handle("POST /api/v1/practice/sessions/{sessionId}/voice-turns", auth.SessionMiddleware(authService, "createPracticeVoiceTurn", createPracticeVoiceTurn))
		completePracticeSession := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			practice.Handler.CompletePracticeSession(w, r, r.PathValue("sessionId"))
		})
		if practice.Idempotency != nil {
			completePracticeSession = practice.Idempotency.Handler("practice", "completePracticeSession", requestUserFromContext, completePracticeSession).ServeHTTP
		}
		mux.Handle("POST /api/v1/practice/sessions/{sessionId}/complete", auth.SessionMiddleware(authService, "completePracticeSession", completePracticeSession))
		mux.Handle("POST /api/v1/practice/sessions/{sessionId}/events", auth.SessionMiddleware(authService, "appendSessionEvent", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			practice.Handler.AppendSessionEvent(w, r, r.PathValue("sessionId"))
		})))
	}
	return mux
}

type reportRuntime struct {
	Handler *apireports.Handler
	Runner  *domainreview.Runner
	Reaper  *domainreview.Reaper
	Service *domainreview.Service
}

func (r *reportRuntime) Routes() reportRoutes {
	if r == nil {
		return reportRoutes{}
	}
	return reportRoutes{Handler: r.Handler}
}

func (r *reportRuntime) Start(ctx context.Context) {
	if r == nil {
		return
	}
	if r.Runner != nil {
		r.Runner.Start(ctx)
	}
	if r.Reaper != nil {
		r.Reaper.Start(ctx)
	}
}

func (r *reportRuntime) Shutdown(ctx context.Context) error {
	if r == nil {
		return nil
	}
	var errs []error
	if r.Runner != nil {
		errs = append(errs, r.Runner.Stop(ctx))
	}
	if r.Reaper != nil {
		errs = append(errs, r.Reaper.Stop(ctx))
	}
	return errors.Join(errs...)
}

func buildReportRuntime(loader *config.Loader, db *sql.DB, logger *slog.Logger, ai aiclient.AIClient) (*reportRuntime, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if ai == nil {
		return nil, fmt.Errorf("report AI client is required")
	}
	repo := storereview.NewRepository(db)
	registryClient, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: registryDirOrDefault(loader, "ai.promptsDir", "config/prompts"),
		RubricsDir: registryDirOrDefault(loader, "ai.rubricsDir", "config/rubrics"),
	})
	if err != nil {
		return nil, fmt.Errorf("build report prompt registry: %w", err)
	}
	taskRuns := storeai.NewTaskRunWriter(db)
	observedAI := ai
	if resolverProvider, ok := ai.(interface {
		Resolver() aiclient.ProfileResolver
	}); ok {
		wrapped, err := observability.New(ai,
			observability.WithRegisterer(observability.NewInMemoryRegistry()),
			observability.WithLogger(observability.NewMemoryLogger()),
			observability.WithAITaskRunWriter(taskRuns),
			observability.WithAuditEventWriter(discardAIAuditWriter{}),
			observability.WithProfileResolver(resolverProvider.Resolver()),
		)
		if err != nil {
			return nil, fmt.Errorf("build report AI observability: %w", err)
		}
		observedAI = wrapped
	}
	service := domainreview.NewService(domainreview.ServiceOptions{
		Registry:   registryClient,
		AI:         observedAI,
		AITaskRuns: taskRuns,
		Repository: repo,
		NewID:      idx.NewID,
	})
	return &reportRuntime{
		Handler: apireports.NewHandler(apireports.HandlerOptions{
			Service: service,
			Session: currentUserFromContext,
		}),
		Runner: domainreview.NewRunner(domainreview.RunnerOptions{
			Store:        repo,
			Service:      service,
			PollInterval: 5 * time.Second,
			Logger:       logger,
		}),
		Reaper: domainreview.NewReaper(domainreview.ReaperOptions{
			Store:        repo,
			LeaseTimeout: 5 * time.Minute,
			Interval:     150 * time.Second,
			Logger:       logger,
		}),
		Service: service,
	}, nil
}

func buildDebriefRoutes(loader *config.Loader, db *sql.DB, ai aiclient.AIClient) (debriefRoutes, error) {
	if ai == nil {
		return debriefRoutes{}, fmt.Errorf("debrief AI client is required")
	}
	registryClient, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: registryDirOrDefault(loader, "ai.promptsDir", "config/prompts"),
		RubricsDir: registryDirOrDefault(loader, "ai.rubricsDir", "config/rubrics"),
	})
	if err != nil {
		return debriefRoutes{}, fmt.Errorf("build debrief prompt registry: %w", err)
	}
	repo := storedebrief.NewRepository(db)
	service := domaindebrief.NewService(domaindebrief.ServiceOptions{
		Store:             repo,
		SuggestionContext: repo,
		Registry:          registryClient,
		AI:                ai,
		AITaskRuns:        storeai.NewTaskRunWriter(db),
		Audit:             repo,
		NewID:             idx.NewID,
	})
	return debriefRoutes{
		Handler: apidebriefs.NewHandler(apidebriefs.HandlerOptions{
			Service: service,
			Session: currentUserFromContext,
		}),
		Idempotency: idempotency.New(idempotency.MiddlewareOptions{
			Store:     idempotency.NewSQLStore(db),
			KeyPepper: loader.GetSecret("auth.challengeTokenPepper").Reveal(),
			TTL:       time.Duration(sharedtypes.IdempotencyKeyTTLSeconds) * time.Second,
		}),
	}, nil
}

func buildJobsRoutes(db *sql.DB) jobsRoutes {
	return jobsRoutes{
		Handler: apijobs.NewHandler(apijobs.HandlerOptions{
			Service: domainjobs.NewService(domainjobs.ServiceOptions{
				Store: storejobs.NewRepository(db),
			}),
			Session: currentUserFromContext,
		}),
	}
}

func buildUploadRoutes(loader *config.Loader, db *sql.DB) (uploadRoutes, error) {
	presignTTL := time.Duration(loader.GetInt("upload.presignTTLSeconds")) * time.Second
	objects, err := objectstore.NewFromConfig(objectstore.FactoryConfig{
		Provider:       loader.GetString("objectStorage.provider"),
		FilesystemRoot: filepath.Join(os.TempDir(), "easyinterview-upload-objects"),
		MinIO: objectstore.MinIOConfig{
			Endpoint:  loader.GetString("objectStorage.endpoint"),
			Bucket:    loader.GetString("objectStorage.bucket"),
			AccessKey: loader.GetSecret("objectStorage.accessKey").Reveal(),
			SecretKey: loader.GetSecret("objectStorage.secretKey").Reveal(),
		},
	})
	if err != nil {
		return uploadRoutes{}, err
	}
	service := uploadservice.New(uploadservice.Options{
		Repository: uploadstore.NewRepository(db),
		Objects:    objects,
		NewID:      idx.NewID,
	})
	return uploadRoutes{
		Handler: uploadhandler.New(uploadhandler.Options{
			Service:    service,
			Session:    currentUserFromContext,
			PresignTTL: presignTTL,
			MaxBytesByPurpose: map[string]int64{
				string(uploadstore.PurposeResume):              int64(loader.GetInt("upload.maxBytes.resume")),
				string(uploadstore.PurposeTargetJobAttachment): int64(loader.GetInt("upload.maxBytes.targetJobAttachment")),
				string(uploadstore.PurposePrivacyExport):       int64(loader.GetInt("upload.maxBytes.privacyExport")),
			},
		}),
		Idempotency: idempotency.New(idempotency.MiddlewareOptions{
			Store:     idempotency.NewSQLStore(db),
			KeyPepper: loader.GetSecret("auth.challengeTokenPepper").Reveal(),
			TTL:       presignTTL,
		}),
		Service: service,
		Objects: objects,
	}, nil
}

func buildPracticeRoutes(loader *config.Loader, db *sql.DB, ai aiclient.AIClient) (practiceRoutes, error) {
	registryClient, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: registryDirOrDefault(loader, "ai.promptsDir", "config/prompts"),
		RubricsDir: registryDirOrDefault(loader, "ai.rubricsDir", "config/rubrics"),
	})
	if err != nil {
		return practiceRoutes{}, fmt.Errorf("build practice prompt registry: %w", err)
	}
	store := storepractice.NewSQLRepository(db)
	taskRuns := storeai.NewTaskRunWriter(db)
	observedAI := ai
	if resolverProvider, ok := ai.(interface {
		Resolver() aiclient.ProfileResolver
	}); ok {
		wrapped, err := observability.New(ai,
			observability.WithRegisterer(observability.NewInMemoryRegistry()),
			observability.WithLogger(observability.NewMemoryLogger()),
			observability.WithAITaskRunWriter(taskRuns),
			observability.WithAuditEventWriter(discardAIAuditWriter{}),
			observability.WithProfileResolver(resolverProvider.Resolver()),
		)
		if err != nil {
			return practiceRoutes{}, fmt.Errorf("build practice AI observability: %w", err)
		}
		observedAI = wrapped
	}
	handler := apipractice.NewHandler(apipractice.HandlerOptions{
		Service: domainpractice.NewService(domainpractice.ServiceOptions{
			Store:      store,
			Registry:   registryClient,
			AI:         observedAI,
			AITaskRuns: taskRuns,
			NewID:      idx.NewID,
		}),
		Session:              currentUserFromContext,
		IdempotencyKeyPepper: loader.GetSecret("auth.challengeTokenPepper").Reveal(),
	})
	return practiceRoutes{
		Handler: handler,
		Idempotency: idempotency.New(idempotency.MiddlewareOptions{
			Store:     idempotency.NewSQLStore(db),
			KeyPepper: loader.GetSecret("auth.challengeTokenPepper").Reveal(),
		}),
	}, nil
}

type targetJobRuntime struct {
	Handler *targetjob.Handler
	Drainer *targetjob.Drainer
	AI      *bootstrap.Runtime
	ParseAI aiclient.AIClient
}

func (r *targetJobRuntime) Start(ctx context.Context) {
	if r == nil || r.Drainer == nil {
		return
	}
	r.Drainer.Start(ctx)
}

func (r *targetJobRuntime) Shutdown(ctx context.Context) error {
	if r == nil {
		return nil
	}
	var err error
	if r.Drainer != nil {
		err = r.Drainer.Shutdown(ctx)
	}
	if r.AI != nil {
		r.AI.Close()
	}
	return err
}

type resumeRuntime struct {
	Handler     *resumehandler.Handler
	Idempotency *idempotency.Middleware
	Drainer     *targetjob.Drainer
	ParseAI     aiclient.AIClient
}

func (r *resumeRuntime) Routes() resumeRoutes {
	if r == nil {
		return resumeRoutes{}
	}
	return resumeRoutes{Handler: r.Handler, Idempotency: r.Idempotency}
}

func (r *resumeRuntime) Start(ctx context.Context) {
	if r == nil || r.Drainer == nil {
		return
	}
	r.Drainer.Start(ctx)
}

func (r *resumeRuntime) Shutdown(ctx context.Context) error {
	if r == nil || r.Drainer == nil {
		return nil
	}
	return r.Drainer.Shutdown(ctx)
}

func buildResumeRuntime(loader *config.Loader, db *sql.DB, logger *slog.Logger, upload uploadRoutes, ai aiclient.AIClient) (*resumeRuntime, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if ai == nil {
		return nil, fmt.Errorf("resume AI client is required")
	}
	store := resumestore.NewRepository(db)
	registryClient, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: registryDirOrDefault(loader, "ai.promptsDir", "config/prompts"),
		RubricsDir: registryDirOrDefault(loader, "ai.rubricsDir", "config/rubrics"),
	})
	if err != nil {
		return nil, fmt.Errorf("build resume prompt registry: %w", err)
	}
	parseAI := ai
	if targetjob.IsTestAppEnv(loader.AppEnv()) {
		parseAI = resumejobs.NewDeterministicParseAIClient(parseAI)
	}
	parseHandler := resumejobs.NewParseHandler(resumejobs.ParseHandlerOptions{
		Store:    store,
		Registry: resumejobs.NewRegistryAdapter(registryClient),
		AI:       parseAI,
		Objects:  upload.Objects,
		NewID:    idx.NewID,
	})
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: store,
		Handlers: map[string]targetjob.JobHandler{
			string(jobs.JobTypeResumeParse): parseHandler,
		},
		Logger: logger,
	})
	service := domainresume.NewService(domainresume.ServiceOptions{
		Store:          store,
		UploadRegister: upload.Service,
		NewID:          idx.NewID,
		DedupePepper:   loader.GetSecret("auth.challengeTokenPepper").Reveal(),
	})
	return &resumeRuntime{
		Handler: resumehandler.New(resumehandler.Options{
			Service: service,
			Session: currentUserFromContext,
		}),
		Idempotency: idempotency.New(idempotency.MiddlewareOptions{
			Store:     idempotency.NewSQLStore(db),
			KeyPepper: loader.GetSecret("auth.challengeTokenPepper").Reveal(),
			TTL:       time.Duration(sharedtypes.IdempotencyKeyTTLSeconds) * time.Second,
		}),
		Drainer: drainer,
		ParseAI: parseAI,
	}, nil
}

func buildTargetJobRuntime(loader *config.Loader, db *sql.DB, logger *slog.Logger, uploadFiles privacyrunner.UploadFileDeleter) (*targetJobRuntime, error) {
	if logger == nil {
		logger = slog.Default()
	}
	store := targetjob.NewSQLStore(db)
	aiRuntime, err := bootstrap.NewClient(bootstrap.Options{
		Config: aiclient.Config{
			AppEnv:               loader.AppEnv(),
			ProviderRegistryPath: loader.GetString("ai.providerRegistryPath"),
			ModelProfilePath:     loader.GetString("ai.modelProfilePath"),
		},
		SecretSource:      secrets.EnvSecretSource{},
		AllowStubProvider: targetjob.IsTestAppEnv(loader.AppEnv()),
		OnWarn: func(err error) {
			logger.Warn("targetjob.ai reload warning", "error", err.Error())
		},
	})
	if err != nil {
		return nil, fmt.Errorf("build targetjob AI runtime: %w", err)
	}

	fetcher := urlfetch.New(urlfetch.FetcherOptions{
		UserAgent: targetjob.URLFetchUserAgent(loader.GetString("runtime.appVersion")),
		Timeout:   targetjob.URLFetchTimeout,
		BodyCap:   targetjob.URLFetchBodyCap,
	})
	var parseAI aiclient.AIClient = aiRuntime.Client
	if targetjob.IsTestAppEnv(loader.AppEnv()) {
		parseAI = targetjob.NewDeterministicParseAIClient(parseAI)
	}

	registryClient, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: registryDirOrDefault(loader, "ai.promptsDir", "config/prompts"),
		RubricsDir: registryDirOrDefault(loader, "ai.rubricsDir", "config/rubrics"),
	})
	if err != nil {
		return nil, fmt.Errorf("build prompt registry: %w", err)
	}

	executor := targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store:    store,
		Registry: targetjob.NewRegistryAdapter(registryClient),
		AI:       parseAI,
		Fetcher:  fetcher,
		NewID:    idx.NewID,
	})
	debriefStore := storedebrief.NewRepository(db)
	taskRuns := storeai.NewTaskRunWriter(db)
	debriefGenerateHandler := domaindebrief.NewGenerateHandler(domaindebrief.GenerateHandlerOptions{
		Store:      debriefStore,
		Registry:   registryClient,
		AI:         aiRuntime.Client,
		AITaskRuns: taskRuns,
		Audit:      debriefStore,
		NewID:      idx.NewID,
	})
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: store,
		Handlers: map[string]targetjob.JobHandler{
			string(jobs.JobTypeTargetImport):    executor,
			string(jobs.JobTypeSourceRefresh):   &targetjob.SourceRefreshHandler{Store: store},
			string(jobs.JobTypeDebriefGenerate): debriefGenerateHandler,
			string(jobs.JobTypePrivacyDelete): privacyrunner.NewPrivacyDeleteHandler(privacyrunner.PrivacyDeleteHandlerOptions{
				Requests:    privacyrunner.NewSQLStore(db),
				UploadFiles: uploadFiles,
			}),
		},
		Logger: logger,
	})
	return &targetJobRuntime{
		Handler: buildTargetJobHandler(loader, store),
		Drainer: drainer,
		AI:      aiRuntime,
		ParseAI: parseAI,
	}, nil
}

func buildTargetJobHandler(loader *config.Loader, store targetjob.Store) *targetjob.Handler {
	return targetjob.NewHandler(targetjob.HandlerOptions{
		Service: targetjob.NewService(targetjob.ServiceOptions{
			Store:        store,
			NewID:        idx.NewID,
			DedupePepper: loader.GetSecret("auth.challengeTokenPepper").Reveal(),
		}),
		Session: currentUserFromContext,
	})
}

func currentUserFromContext(ctx context.Context) (string, bool) {
	current, ok := auth.CurrentSessionFromContext(ctx)
	if !ok || strings.TrimSpace(current.UserID) == "" {
		return "", false
	}
	return current.UserID, true
}

func requestUserFromContext(r *http.Request) (string, bool) {
	return currentUserFromContext(r.Context())
}

func requireIdempotencyKey(status int, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(r.Header.Get(idempotency.HeaderName)) == "" {
			writeRouteAPIError(w, status, sharederrors.CodeValidationFailed, "Idempotency-Key header is required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeRouteAPIError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	meta := sharederrors.CodeRegistry[code]
	raw, _ := json.Marshal(api.ApiErrorResponse{Error: api.ApiError{
		Code:      code,
		Message:   message,
		RequestID: "",
		Retryable: meta.Retryable,
	}})
	_, _ = w.Write(raw)
}

// registryDirOrDefault returns the configured F3 truth-source path or
// `defaultPath` (relative to repo root) when the config key is empty.
func registryDirOrDefault(loader *config.Loader, key, defaultPath string) string {
	if v := strings.TrimSpace(loader.GetString(key)); v != "" {
		return v
	}
	return defaultPath
}

func pointer[T any](value T) *T {
	return &value
}

func buildFlagsClient(loader *config.Loader, logger *slog.Logger) (featureflag.FeatureFlagClient, error) {
	source := loader.GetString("featureFlag.source")
	switch source {
	case "", "file":
		path := loader.GetString("featureFlag.filePath")
		if path == "" {
			path = "config/feature-flags.yaml"
		}
		return featureflag.NewFileProvider(featureflag.FileProviderOptions{
			Path:           path,
			ReloadInterval: 30 * time.Second,
			Logger:         logger,
		})
	case "posthog":
		path := loader.GetString("featureFlag.filePath")
		if path == "" {
			path = "config/feature-flags.yaml"
		}
		publicFlags, err := featureflag.LoadPublicFlagMap(path)
		if err != nil {
			return nil, fmt.Errorf("load feature flag public allowlist: %w", err)
		}
		return featureflag.NewPostHogProvider(featureflag.PostHogProviderOptions{
			Host:       loader.GetString("featureFlag.posthogHost"),
			APIKey:     loader.GetSecret("featureFlag.posthogProjectApiKey").Reveal(),
			SelfHosted: loader.GetBool("featureFlag.posthogSelfHosted"),
			AppEnv:     loader.AppEnv(),
			Public:     publicFlags,
			Logger:     logger,
		})
	default:
		return nil, fmt.Errorf("unknown FEATURE_FLAG_SOURCE %q", source)
	}
}
