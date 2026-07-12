package main

import (
	"context"
	"os"

	"github.com/monshunter/easyinterview/backend/internal/migrations"
	_ "github.com/monshunter/easyinterview/backend/internal/migrations/backfills/v000017"
)

// osEnv adapts os.Getenv to migrations.Env. The adapter lives in cmd/migrate
// (an A4 lint-getenv-boundary allow-list prefix per spec §4.1) so the
// migrations package stays free of direct os.Getenv calls.
type osEnv struct{}

func (osEnv) Getenv(key string) string { return os.Getenv(key) }

func main() {
	os.Exit(migrations.Run(context.Background(), os.Args[1:], osEnv{}, os.Stdout, os.Stderr))
}
