// Command api is the easyinterview HTTP entry point. It owns the public
// runtime-config endpoint required by secrets-and-config spec §6 C-1 / C-2 /
// C-6 and is the first allowlisted callsite of os.Getenv outside
// internal/platform/config (spec §4.1).
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
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
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	apijobs "github.com/monshunter/easyinterview/backend/internal/api/jobs"
	apipractice "github.com/monshunter/easyinterview/backend/internal/api/practice"
	apireports "github.com/monshunter/easyinterview/backend/internal/api/reports"
	"github.com/monshunter/easyinterview/backend/internal/auth"
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
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	storeai "github.com/monshunter/easyinterview/backend/internal/store/ai"
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

	authService, mailWriter, err := buildAuthService(loader, db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: auth init: %v\n", err)
		os.Exit(1)
	}

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
	defer targetJobRuntime.Close()
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
	// Single in-process job kernel (spec D-1 / D-8): every domain handler is
	// registered on one runner.Runtime that owns lease / retry / reaper /
	// graceful shutdown.
	asyncCfg, err := loader.AsyncConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: async config: %v\n", err)
		os.Exit(1)
	}
	kernel, err := buildRunnerKernel(runnerKernelOptions{
		DB:     db,
		Async:  asyncCfg,
		Logger: logger,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "api: runner kernel: %v\n", err)
		os.Exit(1)
	}
	registerRunnerHandlers(kernel, targetJobRuntime.Handlers)
	registerRunnerHandlers(kernel, resumeRuntime.Handlers)
	registerRunnerHandlers(kernel, reportRuntime.Handlers)
	kernel.Register(string(jobs.JobTypeEmailDispatch), auth.NewEmailDispatchHandler(mailWriter))

	srv := &http.Server{
		Addr:              loader.GetString("app.listenAddr"),
		Handler:           withLocalDevCORS(loader.AppEnv(), localDevCORSOrigins(loader), buildAPIHandlerWithUploadReportJobsAndHandlers(loader, flagsClient, authService, targetJobRuntime.Handler, practiceRoutes, uploadRoutes, resumeRuntime.Routes(), reportRuntime.Routes(), jobsRoutes)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	kernel.Start(ctx)

	go func() {
		logger.Info("api: listening", "addr", srv.Addr, "env", loader.AppEnv())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api: serve failed", "error", err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(asyncCfg.ShutdownGraceSeconds)*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	if err := kernel.Shutdown(shutdownCtx); err != nil {
		logger.Warn("api: runner kernel shutdown", "error", err.Error())
	}
}

// registerRunnerHandlers registers every handler in the map onto the kernel.
func registerRunnerHandlers(rt *runner.Runtime, handlers map[string]runner.Handler) {
	for jobType, handler := range handlers {
		rt.Register(jobType, handler)
	}
}

type runnerKernelOptions struct {
	DB          *sql.DB
	Async       config.AsyncConfig
	Logger      *slog.Logger
	Metrics     runner.Metrics
	OutboxStore runner.OutboxStore
	Now         func() time.Time
}

func buildRunnerKernel(opts runnerKernelOptions) (*runner.Runtime, error) {
	cfg := runner.ConfigFromSeconds(
		opts.Async.ScanIntervalSeconds,
		opts.Async.LeaseTimeoutSeconds,
		opts.Async.ReaperIntervalSeconds,
		opts.Async.ShutdownGraceSeconds,
		runner.QueueWeights{
			Critical: opts.Async.QueueWeights.Critical,
			Default:  opts.Async.QueueWeights.Default,
			Low:      opts.Async.QueueWeights.Low,
		},
	)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	metrics := opts.Metrics
	if metrics == nil {
		metrics = runner.NewKernelMetrics(runner.NewInMemoryMetricsRegistry())
	}
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	kernel := runner.New(runner.Options{
		Store:   runner.NewSQLStore(opts.DB),
		Config:  cfg,
		Logger:  logger,
		Metrics: metrics,
		Now:     now,
	})
	outboxStore := opts.OutboxStore
	if outboxStore == nil {
		outboxStore = runner.NewSQLOutboxStore(opts.DB)
	}
	kernel.SetOutboxDispatcher(runner.NewOutboxDispatcher(runner.OutboxDispatcherOptions{
		Store:        outboxStore,
		ScanInterval: cfg.ScanInterval,
		Logger:       logger,
		Metrics:      metrics,
		Now:          now,
	}))
	return kernel, nil
}

func buildAuthService(loader *config.Loader, db *sql.DB) (*auth.EmailCodeService, auth.DeliveryWriter, error) {
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
	verifyBaseURL := strings.TrimSpace(loader.GetString("email.verifyBaseURL"))
	if verifyBaseURL == "" {
		verifyBaseURL = "/api/v1/auth/email/verify"
	}
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: verifyBaseURL})
	writer := auth.DeliveryWriter(sink)
	if strings.EqualFold(strings.TrimSpace(loader.GetString("email.provider")), "mailpit") {
		host := strings.TrimSpace(loader.GetString("email.smtpHost"))
		if host == "" {
			host = "127.0.0.1"
		}
		port := loader.GetInt("email.smtpPort")
		if port <= 0 {
			port = 1025
		}
		from := strings.TrimSpace(loader.GetString("email.fromAddress"))
		if from == "" {
			from = "noreply@easyinterview.local"
		}
		writer = auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
			SMTPAddr:             net.JoinHostPort(host, fmt.Sprintf("%d", port)),
			FromAddress:          from,
			VerifyBaseURL:        verifyBaseURL,
			DeliverySecrets:      sink,
			LookupChallengeEmail: auth.SQLChallengeEmailLookup(db),
		})
	}
	// Producer enqueues email_dispatch async_jobs rows (spec D-10); the kernel
	// EmailDispatchHandler delivers them through the configured writer.
	enqueuer := auth.NewEmailDispatchEnqueuer(db, idx.NewID, func() time.Time { return time.Now().UTC() })
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:               auth.NewSQLStore(db),
		Dispatcher:          enqueuer,
		DeliverySecrets:     sink,
		ChallengePepper:     challengePepper,
		SessionCookieSecret: sessionCookieSecret,
	})
	return service, writer, nil
}

func buildAPIHandler(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.EmailCodeService, db *sql.DB) http.Handler {
	upload, _ := buildUploadRoutes(loader, db)
	return buildAPIHandlerWithUploadAndHandlers(loader, flagsClient, authService, buildTargetJobHandler(loader, targetjob.NewSQLStore(db)), practiceRoutes{}, upload, resumeRoutes{})
}

func buildAPIHandlerWithTargetJobHandler(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.EmailCodeService, targetJobHandler *targetjob.Handler) http.Handler {
	return buildAPIHandlerWithHandlers(loader, flagsClient, authService, targetJobHandler, practiceRoutes{})
}

type practiceRoutes struct {
	Handler     *apipractice.Handler
	Idempotency *idempotency.Middleware
}

type reportRoutes struct {
	Handler *apireports.Handler
}

type jobsRoutes struct {
	Handler *apijobs.Handler
}

type discardAIAuditWriter struct{}

func (discardAIAuditWriter) WriteAuditEvent(context.Context, aiclient.AuditEventRow) error {
	return nil
}

func aiObservabilityOptions(loader *config.Loader, taskRuns aiclient.AITaskRunWriter, resolver aiclient.ProfileResolver) []observability.Option {
	opts := []observability.Option{
		observability.WithRegisterer(observability.NewInMemoryRegistry()),
		observability.WithLogger(observability.NewMemoryLogger()),
		observability.WithAITaskRunWriter(taskRuns),
		observability.WithAuditEventWriter(discardAIAuditWriter{}),
		observability.WithProfileResolver(resolver),
	}
	if loader != nil && loader.GetBool("ai.debugPrintRawOutput") {
		opts = append(opts, observability.WithRawOutputDebugWriter(os.Stderr))
	}
	return opts
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

func buildAPIHandlerWithHandlers(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.EmailCodeService, targetJobHandler *targetjob.Handler, practice practiceRoutes) http.Handler {
	return buildAPIHandlerWithUploadAndHandlers(loader, flagsClient, authService, targetJobHandler, practice, uploadRoutes{}, resumeRoutes{})
}

func buildAPIHandlerWithUploadAndHandlers(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.EmailCodeService, targetJobHandler *targetjob.Handler, practice practiceRoutes, upload uploadRoutes, resume resumeRoutes) http.Handler {
	return buildAPIHandlerWithUploadReportAndHandlers(loader, flagsClient, authService, targetJobHandler, practice, upload, resume, reportRoutes{})
}

func buildAPIHandlerWithUploadReportAndHandlers(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.EmailCodeService, targetJobHandler *targetjob.Handler, practice practiceRoutes, upload uploadRoutes, resume resumeRoutes, reports reportRoutes) http.Handler {
	return buildAPIHandlerWithUploadReportJobsAndHandlers(loader, flagsClient, authService, targetJobHandler, practice, upload, resume, reports, jobsRoutes{})
}

func buildAPIHandlerWithUploadReportJobsAndHandlers(loader *config.Loader, flagsClient featureflag.FeatureFlagClient, authService *auth.EmailCodeService, targetJobHandler *targetjob.Handler, practice practiceRoutes, upload uploadRoutes, resume resumeRoutes, reports reportRoutes, jobs jobsRoutes) http.Handler {
	mux := http.NewServeMux()
	authHandler := auth.NewHandler(auth.HandlerOptions{
		EmailCode:    authService,
		CookiePolicy: pointer(auth.CookiePolicyForAppEnv(loader.AppEnv())),
	})
	mux.Handle("POST /api/v1/auth/email/start", auth.SessionMiddleware(authService, "startAuthEmailChallenge", http.HandlerFunc(authHandler.StartAuthEmailChallenge)))
	mux.Handle("GET /api/v1/auth/email/verify", auth.SessionMiddleware(authService, "verifyAuthEmailChallenge", http.HandlerFunc(authHandler.VerifyAuthEmailChallenge)))
	mux.Handle("POST /api/v1/auth/logout", auth.SessionMiddleware(authService, "logout", http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /api/v1/me", auth.SessionMiddleware(authService, "getMe", http.HandlerFunc(authHandler.GetMe)))
	mux.Handle("PATCH /api/v1/me", auth.SessionMiddleware(authService, "completeMyProfile", http.HandlerFunc(authHandler.CompleteMyProfile)))
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
		updateResume := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.UpdateResume(w, r, r.PathValue("resumeId"))
		})
		if resume.Idempotency != nil {
			updateResume = requireIdempotencyKey(http.StatusUnprocessableEntity, resume.Idempotency.Handler("resume", "updateResume", requestUserFromContext, updateResume)).ServeHTTP
		}
		duplicateResume := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.DuplicateResume(w, r, r.PathValue("resumeId"))
		})
		if resume.Idempotency != nil {
			duplicateResume = requireIdempotencyKey(http.StatusUnprocessableEntity, resume.Idempotency.Handler("resume", "duplicateResume", requestUserFromContext, duplicateResume)).ServeHTTP
		}
		archiveResume := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.ArchiveResume(w, r, r.PathValue("resumeId"))
		})
		if resume.Idempotency != nil {
			archiveResume = requireIdempotencyKey(http.StatusUnprocessableEntity, resume.Idempotency.Handler("resume", "archiveResume", requestUserFromContext, archiveResume)).ServeHTTP
		}
		exportResume := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.ExportResume(w, r, r.PathValue("resumeId"))
		})
		if resume.Idempotency != nil {
			exportResume = requireIdempotencyKey(http.StatusUnprocessableEntity, resume.Idempotency.Handler("resume", "exportResume", requestUserFromContext, exportResume)).ServeHTTP
		}
		requestResumeTailor := http.HandlerFunc(resume.Handler.RequestResumeTailor)
		if resume.Idempotency != nil {
			requestResumeTailor = requireIdempotencyKey(http.StatusUnprocessableEntity, resume.Idempotency.Handler("resume", "requestResumeTailor", requestUserFromContext, requestResumeTailor)).ServeHTTP
		}
		mux.Handle("GET /api/v1/resumes", auth.SessionMiddleware(authService, "listResumes", http.HandlerFunc(resume.Handler.ListResumes)))
		mux.Handle("POST /api/v1/resumes", auth.SessionMiddleware(authService, "registerResume", registerResume))
		mux.Handle("POST /api/v1/resume/tailor", auth.SessionMiddleware(authService, "requestResumeTailor", requestResumeTailor))
		mux.Handle("GET /api/v1/resume/tailor-runs/{tailorRunId}", auth.SessionMiddleware(authService, "getResumeTailorRun", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.GetResumeTailorRun(w, r, r.PathValue("tailorRunId"))
		})))
		mux.Handle("GET /api/v1/resumes/{resumeId}", auth.SessionMiddleware(authService, "getResume", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.GetResume(w, r, r.PathValue("resumeId"))
		})))
		mux.Handle("GET /api/v1/resumes/{resumeId}/source", auth.SessionMiddleware(authService, "getResumeSource", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.GetResumeSource(w, r, r.PathValue("resumeId"))
		})))
		mux.Handle("PATCH /api/v1/resumes/{resumeId}", auth.SessionMiddleware(authService, "updateResume", updateResume))
		mux.Handle("POST /api/v1/resumes/{resumeId}/duplicate", auth.SessionMiddleware(authService, "duplicateResume", duplicateResume))
		mux.Handle("POST /api/v1/resumes/{resumeId}/archive", auth.SessionMiddleware(authService, "archiveResume", archiveResume))
		mux.Handle("POST /api/v1/resumes/{resumeId}/exports", auth.SessionMiddleware(authService, "exportResume", exportResume))
	}
	if reports.Handler != nil {
		mux.Handle("GET /api/v1/reports/{reportId}", auth.SessionMiddleware(authService, "getFeedbackReport", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reports.Handler.GetFeedbackReport(w, r, r.PathValue("reportId"))
		})))
		mux.Handle("GET /api/v1/targets/{targetJobId}/reports", auth.SessionMiddleware(authService, "listTargetJobReports", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reports.Handler.ListTargetJobReports(w, r, r.PathValue("targetJobId"))
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
	mux.Handle("POST /api/v1/targets/{targetJobId}/archive", auth.SessionMiddleware(authService, "archiveTargetJob", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetJobHandler.ArchiveTargetJob(w, r, r.PathValue("targetJobId"))
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
	Handler  *apireports.Handler
	Handlers map[string]runner.Handler
	Service  *domainreview.Service
}

func (r *reportRuntime) Routes() reportRoutes {
	if r == nil {
		return reportRoutes{}
	}
	return reportRoutes{Handler: r.Handler}
}

// Handles reports whether this runtime contributes a handler for jobType.
func (r *reportRuntime) Handles(jobType string) bool {
	if r == nil {
		return false
	}
	_, ok := r.Handlers[jobType]
	return ok
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
			aiObservabilityOptions(loader, taskRuns, resolverProvider.Resolver())...,
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
		Handlers: map[string]runner.Handler{
			string(jobs.JobTypeReportGenerate): domainreview.NewGenerateHandler(domainreview.GenerateHandlerOptions{
				Store:   repo,
				Service: service,
			}),
		},
		Service: service,
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
			aiObservabilityOptions(loader, taskRuns, resolverProvider.Resolver())...,
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
	Handler  *targetjob.Handler
	Handlers map[string]runner.Handler
	AI       *bootstrap.Runtime
	ParseAI  aiclient.AIClient
}

// Handles reports whether this runtime contributes a handler for jobType.
func (r *targetJobRuntime) Handles(jobType string) bool {
	if r == nil {
		return false
	}
	_, ok := r.Handlers[jobType]
	return ok
}

// Close releases the AI runtime resources owned by this runtime. Job lifecycle
// (lease / reaper / shutdown) is owned by the shared runner.Runtime kernel.
func (r *targetJobRuntime) Close() {
	if r == nil {
		return
	}
	if r.AI != nil {
		r.AI.Close()
	}
}

type resumeRuntime struct {
	Handler     *resumehandler.Handler
	Idempotency *idempotency.Middleware
	Handlers    map[string]runner.Handler
	ParseAI     aiclient.AIClient
}

func (r *resumeRuntime) Routes() resumeRoutes {
	if r == nil {
		return resumeRoutes{}
	}
	return resumeRoutes{Handler: r.Handler, Idempotency: r.Idempotency}
}

// Handles reports whether this runtime contributes a handler for jobType.
func (r *resumeRuntime) Handles(jobType string) bool {
	if r == nil {
		return false
	}
	_, ok := r.Handlers[jobType]
	return ok
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
	tailorHandler := resumejobs.NewTailorHandler(resumejobs.TailorHandlerOptions{
		Store:      store,
		Registry:   resumejobs.NewRegistryAdapter(registryClient),
		AI:         ai,
		AITaskRuns: storeai.NewTaskRunWriter(db),
		NewID:      idx.NewID,
	})
	handlers := map[string]runner.Handler{
		string(jobs.JobTypeResumeParse):  runner.FromTargetjobHandler(parseHandler),
		string(jobs.JobTypeResumeTailor): runner.FromTargetjobHandler(tailorHandler),
	}
	service := domainresume.NewService(domainresume.ServiceOptions{
		Store:            store,
		UploadRegister:   upload.Service,
		SourceObjects:    upload.Objects,
		NewID:            idx.NewID,
		DedupePepper:     loader.GetSecret("auth.challengeTokenPepper").Reveal(),
		MaxActiveResumes: loader.GetInt("resume.maxActive"),
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
		Handlers: handlers,
		ParseAI:  parseAI,
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
	taskRuns := storeai.NewTaskRunWriter(db)
	if db != nil {
		wrapped, err := observability.New(parseAI,
			aiObservabilityOptions(loader, taskRuns, aiRuntime.Client.Resolver())...,
		)
		if err != nil {
			return nil, fmt.Errorf("build targetjob AI observability: %w", err)
		}
		parseAI = wrapped
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
	privacyDeleteHandler := privacyrunner.NewPrivacyDeleteHandler(privacyrunner.PrivacyDeleteHandlerOptions{
		Requests:    privacyrunner.NewSQLStore(db),
		UploadFiles: uploadFiles,
	})
	handlers := map[string]runner.Handler{
		string(jobs.JobTypeTargetImport):  runner.FromTargetjobHandler(executor),
		string(jobs.JobTypeSourceRefresh): runner.FromTargetjobHandler(&targetjob.SourceRefreshHandler{Store: store}),
		string(jobs.JobTypePrivacyDelete): runner.FromTargetjobHandler(privacyDeleteHandler),
	}
	return &targetJobRuntime{
		Handler:  buildTargetJobHandler(loader, store),
		Handlers: handlers,
		AI:       aiRuntime,
		ParseAI:  parseAI,
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

func withLocalDevCORS(appEnv string, allowed map[string]struct{}, next http.Handler) http.Handler {
	if appEnv != "dev" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept,Accept-Language,Content-Type,Idempotency-Key,Traceparent,X-Client-Version,X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			w.Header().Add("Vary", "Origin")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		} else if origin != "" && r.Method == http.MethodOptions {
			http.Error(w, "CORS origin is not allowed", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func localDevCORSOrigins(loader *config.Loader) map[string]struct{} {
	allowed := map[string]struct{}{}
	addURLOrigin(allowed, loader.GetString("email.verifyBaseURL"))
	return allowed
}

func addURLOrigin(out map[string]struct{}, raw string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.HasPrefix(trimmed, "/") {
		return
	}
	u, err := url.Parse(trimmed)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return
	}
	out[u.Scheme+"://"+u.Host] = struct{}{}
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
