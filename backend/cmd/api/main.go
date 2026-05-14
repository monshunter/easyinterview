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
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	apipractice "github.com/monshunter/easyinterview/backend/internal/api/practice"
	"github.com/monshunter/easyinterview/backend/internal/auth"
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
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	storepractice "github.com/monshunter/easyinterview/backend/internal/store/practice"
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
	srv := &http.Server{
		Addr:              loader.GetString("app.listenAddr"),
		Handler:           buildAPIHandlerWithUploadAndHandlers(loader, flagsClient, authService, targetJobRuntime.Handler, practiceRoutes, uploadRoutes, resumeRuntime.Routes()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	targetJobRuntime.Start(ctx)
	resumeRuntime.Start(ctx)

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
		mux.Handle("GET /api/v1/resumes", auth.SessionMiddleware(authService, "listResumes", http.HandlerFunc(resume.Handler.ListResumes)))
		mux.Handle("POST /api/v1/resumes", auth.SessionMiddleware(authService, "registerResume", registerResume))
		mux.Handle("GET /api/v1/resumes/{resumeAssetId}", auth.SessionMiddleware(authService, "getResume", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resume.Handler.GetResume(w, r, r.PathValue("resumeAssetId"))
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
		mux.Handle("POST /api/v1/practice/sessions", auth.SessionMiddleware(authService, "startPracticeSession", http.HandlerFunc(practice.Handler.StartPracticeSession)))
		mux.Handle("GET /api/v1/practice/sessions/{sessionId}", auth.SessionMiddleware(authService, "getPracticeSession", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			practice.Handler.GetPracticeSession(w, r, r.PathValue("sessionId"))
		})))
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
	handler := apipractice.NewHandler(apipractice.HandlerOptions{
		Service: domainpractice.NewService(domainpractice.ServiceOptions{
			Store:    store,
			Registry: registryClient,
			AI:       ai,
			NewID:    idx.NewID,
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
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: store,
		Handlers: map[string]targetjob.JobHandler{
			string(jobs.JobTypeTargetImport):  executor,
			string(jobs.JobTypeSourceRefresh): &targetjob.SourceRefreshHandler{Store: store},
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
