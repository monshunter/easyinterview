package main

import (
	"context"
	"os"

	"github.com/monshunter/easyinterview/backend/internal/migrations"
	_ "github.com/monshunter/easyinterview/backend/internal/migrations/backfills/v000001"
)

func main() {
	os.Exit(migrations.Run(context.Background(), os.Args[1:], migrations.ProcessEnv(), os.Stdout, os.Stderr))
}
