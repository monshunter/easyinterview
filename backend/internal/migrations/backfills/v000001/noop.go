package v000001

import (
	"context"
	"database/sql"

	"github.com/monshunter/easyinterview/backend/internal/migrations"
)

func init() {
	migrations.RegisterBackfill("baseline_noop", run)
}

func run(_ context.Context, _ *sql.DB, _ migrations.BackfillMode) error {
	return nil
}
