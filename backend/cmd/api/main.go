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
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
	"github.com/monshunter/easyinterview/backend/internal/platform/secrets"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
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

	srv := &http.Server{
		Addr:              loader.GetString("app.listenAddr"),
		Handler:           buildAPIHandler(loader, flagsClient, authService, db),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
	mux := http.NewServeMux()
	authHandler := auth.NewHandler(auth.HandlerOptions{
		Passwordless: authService,
		CookiePolicy: pointer(auth.CookiePolicyForAppEnv(loader.AppEnv())),
	})
	targetJobHandler := targetjob.NewHandler(targetjob.HandlerOptions{
		Service: targetjob.NewService(targetjob.ServiceOptions{
			Store:        targetjob.NewSQLStore(db),
			NewID:        idx.NewID,
			DedupePepper: loader.GetSecret("auth.challengeTokenPepper").Reveal(),
		}),
		Session: func(ctx context.Context) (string, bool) {
			current, ok := auth.CurrentSessionFromContext(ctx)
			if !ok || strings.TrimSpace(current.UserID) == "" {
				return "", false
			}
			return current.UserID, true
		},
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
	mux.Handle("GET /api/v1/targets", auth.SessionMiddleware(authService, "listTargetJobs", http.HandlerFunc(targetJobHandler.ListTargetJobs)))
	mux.Handle("POST /api/v1/targets/import", auth.SessionMiddleware(authService, "importTargetJob", http.HandlerFunc(targetJobHandler.ImportTargetJob)))
	mux.Handle("GET /api/v1/targets/{targetJobId}", auth.SessionMiddleware(authService, "getTargetJob", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetJobHandler.GetTargetJob(w, r, r.PathValue("targetJobId"))
	})))
	mux.Handle("PATCH /api/v1/targets/{targetJobId}", auth.SessionMiddleware(authService, "updateTargetJob", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetJobHandler.UpdateTargetJob(w, r, r.PathValue("targetJobId"))
	})))
	return mux
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
