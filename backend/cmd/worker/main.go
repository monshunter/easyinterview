// Command worker is the easyinterview async worker entry point. P0
// implementation is a placeholder that loads configuration through the
// platform/config loader (spec D-1) and exits — C8 backend-async-runtime
// owns the asynq scheduling logic.
package main

import (
	"fmt"
	"os"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/secrets"
)

func main() {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}

	loader, err := config.Load(config.Options{
		AppEnv:    appEnv,
		ConfigDir: "config",
		EnvBindings: map[string]string{
			"WORKER_LISTEN_ADDR": "worker.listenAddr",
		},
		SecretSource: secrets.EnvSecretSource{},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "worker: load config: %v\n", err)
		os.Exit(1)
	}
	if err := loader.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "worker: config validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("worker: configuration loaded; listenAddr=%s; queue weights critical/default/low=%d/%d/%d\n",
		loader.GetString("worker.listenAddr"),
		loader.GetInt("async.queueWeights.critical"),
		loader.GetInt("async.queueWeights.default"),
		loader.GetInt("async.queueWeights.low"),
	)
	fmt.Println("worker: TODO C8 backend-async-runtime will own the asynq dispatcher loop.")
}
