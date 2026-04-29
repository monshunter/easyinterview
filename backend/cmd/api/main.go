// Command api is the easyinterview HTTP entry point. It owns the
// minimal runtime-config endpoint required by secrets-and-config spec
// §6 C-1 / C-2 / C-6 and is the first allowlisted callsite of os.Getenv
// outside internal/platform/config (spec §4.1).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/featureflag"
	"github.com/monshunter/easyinterview/backend/internal/platform/secrets"
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

	loader, err := config.Load(config.Options{
		AppEnv:    appEnv,
		ConfigDir: configDir,
		EnvBindings: map[string]string{
			"APP_ENV":                 "app.env",
			"APP_LISTEN_ADDR":         "app.listenAddr",
			"WORKER_LISTEN_ADDR":      "worker.listenAddr",
			"DATABASE_URL":            "database.url",
			"REDIS_URL":               "redis.url",
			"OBJECT_STORAGE_ENDPOINT": "objectStorage.endpoint",
			"OBJECT_STORAGE_BUCKET":   "objectStorage.bucket",
			"OTEL_EXPORTER_OTLP_ENDPOINT": "observability.otlpEndpoint",
			"LOG_LEVEL":                "log.level",
			"AI_GATEWAY_BASE_URL":      "ai.gatewayBaseURL",
			"AI_MODEL_PROFILE_PATH":    "ai.modelProfilePath",
			"FEATURE_FLAG_SOURCE":      "featureFlag.source",
			"FEATURE_FLAG_FILE_PATH":   "featureFlag.filePath",
			"POSTHOG_HOST":             "featureFlag.posthogHost",
			"POSTHOG_SELF_HOSTED":      "featureFlag.posthogSelfHosted",
			"POSTHOG_PUBLIC_KEY":       "featureFlag.posthogPublicKey",
			"EMAIL_PROVIDER":           "email.provider",
		},
		SecretBindings: map[string]string{
			"objectStorage.accessKey":          "OBJECT_STORAGE_ACCESS_KEY",
			"objectStorage.secretKey":          "OBJECT_STORAGE_SECRET_KEY",
			"auth.sessionCookieSecret":         "SESSION_COOKIE_SECRET",
			"auth.challengeTokenPepper":        "AUTH_CHALLENGE_TOKEN_PEPPER",
			"ai.gatewayApiKey":                 "AI_GATEWAY_API_KEY",
			"email.providerApiKey":             "EMAIL_PROVIDER_API_KEY",
			"featureFlag.posthogProjectApiKey": "POSTHOG_PROJECT_API_KEY",
		},
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
			"app.env":               loader.GetString("app.env"),
			"app.listenAddr":        loader.GetString("app.listenAddr"),
			"log.level":             loader.GetString("log.level"),
			"runtime.appVersion":    loader.GetString("runtime.appVersion"),
			"runtime.defaultUiLanguage": loader.GetString("runtime.defaultUiLanguage"),
			"featureFlag.source":    loader.GetString("featureFlag.source"),
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

	mux := http.NewServeMux()
	mux.Handle("/api/v1/runtime-config", config.NewRuntimeConfigHandler(config.RuntimeConfigHandlerOptions{
		Loader: loader,
		Flags:  flagsClient,
		FlagContextFunc: func(r *http.Request) featureflag.FlagContext {
			return featureflag.FlagContext{AppEnv: loader.AppEnv()}
		},
		// SessionResolver is owned by C1 backend-auth; unauthenticated
		// requests default to opt-out per spec D-2.
		SessionResolver: nil,
	}))

	srv := &http.Server{
		Addr:              loader.GetString("app.listenAddr"),
		Handler:           mux,
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
		return featureflag.NewPostHogProvider(featureflag.PostHogProviderOptions{
			Host:       loader.GetString("featureFlag.posthogHost"),
			APIKey:     loader.GetSecret("featureFlag.posthogProjectApiKey").Reveal(),
			SelfHosted: loader.GetBool("featureFlag.posthogSelfHosted"),
			AppEnv:     loader.AppEnv(),
			Logger:     logger,
		})
	default:
		return nil, fmt.Errorf("unknown FEATURE_FLAG_SOURCE %q", source)
	}
}
