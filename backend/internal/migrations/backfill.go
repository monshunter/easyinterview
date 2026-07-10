package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// BackfillMode is the execution mode recorded in schema_backfills.
type BackfillMode string

const (
	BackfillModeDryRun BackfillMode = "dry_run"
	BackfillModeApply  BackfillMode = "apply"
)

// BackfillEntry is one manifest-owned backfill definition.
type BackfillEntry struct {
	Version    int    `yaml:"version"`
	Name       string `yaml:"name"`
	Checksum   string `yaml:"checksum"`
	Reversible bool   `yaml:"reversible"`
	DryRun     bool   `yaml:"dryRun"`
}

// BackfillFunc executes a registered backfill in dry-run or apply mode.
type BackfillFunc func(context.Context, *sql.DB, BackfillMode) error

// BackfillRegistry maps manifest names to compiled Go backfill functions.
type BackfillRegistry map[string]BackfillFunc

var (
	globalBackfillMu       sync.RWMutex
	globalBackfillRegistry = BackfillRegistry{}
)

// RegisterBackfill registers a compiled backfill implementation.
func RegisterBackfill(name string, fn BackfillFunc) {
	globalBackfillMu.Lock()
	defer globalBackfillMu.Unlock()
	globalBackfillRegistry[name] = fn
}

// RegisteredBackfills returns a snapshot of compiled backfill registrations.
func RegisteredBackfills() BackfillRegistry {
	globalBackfillMu.RLock()
	defer globalBackfillMu.RUnlock()
	out := make(BackfillRegistry, len(globalBackfillRegistry))
	for name, fn := range globalBackfillRegistry {
		out[name] = fn
	}
	return out
}

// LoadBackfillManifest reads the optional backfill manifest path.
func LoadBackfillManifest(path string) ([]BackfillEntry, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var manifest struct {
		Backfills []BackfillEntry `yaml:"backfills"`
	}
	if err := yaml.Unmarshal(b, &manifest); err != nil {
		return nil, err
	}
	for _, entry := range manifest.Backfills {
		if entry.Version <= 0 || entry.Name == "" || entry.Checksum == "" {
			return nil, fmt.Errorf("invalid backfill manifest entry: version, name, and checksum are required")
		}
	}
	return manifest.Backfills, nil
}

// RunBackfills runs manifest-defined backfills after schema migration up.
func RunBackfills(cmd Command) error {
	entries, err := LoadBackfillManifest(cmd.BackfillManifest)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	db, err := sql.Open("postgres", cmd.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	return RunBackfillEntries(context.Background(), db, cmd, entries, RegisteredBackfills())
}

// RunBackfillEntries executes entries and writes schema_backfills ledger rows.
func RunBackfillEntries(ctx context.Context, db *sql.DB, cmd Command, entries []BackfillEntry, registry BackfillRegistry) error {
	if cmd.ForceBackfill && cmd.AppEnv == "prod" {
		return fmt.Errorf("refusing forced backfill in APP_ENV=prod")
	}
	for _, entry := range entries {
		fn := registry[entry.Name]
		if fn == nil {
			return fmt.Errorf("backfill %q is declared but not registered", entry.Name)
		}
		if entry.DryRun {
			if err := runBackfillMode(ctx, db, cmd, entry, BackfillModeDryRun, fn); err != nil {
				return err
			}
		}
		if err := runBackfillMode(ctx, db, cmd, entry, BackfillModeApply, fn); err != nil {
			return err
		}
	}
	return nil
}

func runBackfillMode(ctx context.Context, db *sql.DB, cmd Command, entry BackfillEntry, mode BackfillMode, fn BackfillFunc) error {
	exists, err := backfillSuccessExists(ctx, db, entry, mode)
	if err != nil {
		return err
	}
	if exists && !cmd.ForceBackfill {
		return nil
	}
	if exists && cmd.ForceBackfill {
		if err := fn(ctx, db, mode); err != nil {
			return err
		}
		_, err := db.ExecContext(ctx, `
			update schema_backfills
			set started_at = now(), completed_at = now(), status = 'succeeded', error_message = null
			where version = $1 and mode = $2 and checksum = $3
		`, entry.Version, string(mode), entry.Checksum)
		return err
	}

	if _, err := db.ExecContext(ctx, `
		insert into schema_backfills (version, name, mode, status, checksum, started_at)
		values ($1, $2, $3, 'running', $4, now())
	`, entry.Version, entry.Name, string(mode), entry.Checksum); err != nil {
		return err
	}
	if err := fn(ctx, db, mode); err != nil {
		_, _ = db.ExecContext(ctx, `
			update schema_backfills
			set status = 'failed', completed_at = now(), error_message = $4
			where version = $1 and mode = $2 and checksum = $3 and status = 'running'
		`, entry.Version, string(mode), entry.Checksum, err.Error())
		return err
	}
	_, err = db.ExecContext(ctx, `
		update schema_backfills
		set status = 'succeeded', completed_at = now(), error_message = null
		where version = $1 and mode = $2 and checksum = $3 and status = 'running'
	`, entry.Version, string(mode), entry.Checksum)
	return err
}

func backfillSuccessExists(ctx context.Context, db *sql.DB, entry BackfillEntry, mode BackfillMode) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
		select exists (
			select 1 from schema_backfills
			where version = $1 and mode = $2 and checksum = $3 and status = 'succeeded'
		)
	`, entry.Version, string(mode), entry.Checksum).Scan(&exists)
	return exists, err
}
