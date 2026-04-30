package migrations

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Env abstracts process environment lookups for the migration CLI.
type Env interface {
	Getenv(key string) string
}

// StaticEnv is a test helper that satisfies Env.
type StaticEnv map[string]string

// Getenv returns the configured value for key.
func (e StaticEnv) Getenv(key string) string {
	return e[key]
}

type processEnv struct{}

func (processEnv) Getenv(key string) string {
	return os.Getenv(key)
}

// ProcessEnv returns an Env backed by os.Getenv.
func ProcessEnv() Env {
	return processEnv{}
}

// Run executes the migration CLI and returns a process-style exit code.
func Run(ctx context.Context, args []string, env Env, stdout, stderr io.Writer) int {
	if env == nil {
		env = ProcessEnv()
	}
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	opts, command, commandArgs, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: %v\n", err)
		printUsage(stderr)
		return 2
	}

	switch command {
	case "help":
		printUsage(stdout)
		return 0
	case "create":
		name := ""
		if len(commandArgs) > 0 {
			name = commandArgs[0]
		}
		files, err := CreateMigrationFiles(opts.MigrationsDir, name)
		if err != nil {
			fmt.Fprintf(stderr, "ERROR: %v\n", err)
			return 1
		}
		for _, file := range files {
			fmt.Fprintf(stdout, "created %s\n", file)
		}
		return 0
	case "privacy-matrix":
		if !hasFlag(commandArgs, "--dry-run") {
			fmt.Fprintln(stderr, "ERROR: privacy-matrix currently supports --dry-run only")
			return 2
		}
		WritePrivacyMatrix(stdout)
		return 0
	}

	if env.Getenv("DATABASE_URL") == "" {
		fmt.Fprintln(stderr, "ERROR: DATABASE_URL is required for migration commands")
		return 1
	}
	if command == "down" && env.Getenv("APP_ENV") == "prod" && env.Getenv("MIGRATE_DOWN_FORCE") != "1" {
		fmt.Fprintln(stderr, "ERROR: refusing migrate-down in APP_ENV=prod; set MIGRATE_DOWN_FORCE=1 during an approved operation window")
		return 1
	}

	if err := RunCommand(ctx, Command{
		Name:             command,
		DatabaseURL:      env.Getenv("DATABASE_URL"),
		MigrationsDir:    opts.MigrationsDir,
		BackfillManifest: opts.BackfillManifest,
		AppEnv:           env.Getenv("APP_ENV"),
		DropExtensions:   env.Getenv("MIGRATE_DROP_EXTENSIONS") == "1",
		ForceBackfill:    env.Getenv("MIGRATE_BACKFILL_FORCE") == "1",
		Stdout:           stdout,
	}); err != nil {
		fmt.Fprintf(stderr, "ERROR: %v\n", err)
		return 1
	}
	return 0
}

type options struct {
	MigrationsDir    string
	BackfillManifest string
}

func defaultOptions() options {
	return options{
		MigrationsDir:    filepath.Clean("../migrations"),
		BackfillManifest: filepath.Clean("../migrations/backfill/manifest.yaml"),
	}
}

func parseArgs(args []string) (options, string, []string, error) {
	opts := defaultOptions()
	command := ""
	commandArgs := make([]string, 0)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help", "help":
			return opts, "help", nil, nil
		case "--migrations-dir":
			if i+1 >= len(args) {
				return opts, "", nil, fmt.Errorf("--migrations-dir requires a value")
			}
			i++
			opts.MigrationsDir = args[i]
		case "--backfill-manifest":
			if i+1 >= len(args) {
				return opts, "", nil, fmt.Errorf("--backfill-manifest requires a value")
			}
			i++
			opts.BackfillManifest = args[i]
		default:
			if strings.HasPrefix(arg, "--") && command == "" {
				return opts, "", nil, fmt.Errorf("unknown flag %s", arg)
			}
			if command == "" {
				command = arg
				continue
			}
			commandArgs = append(commandArgs, arg)
		}
	}

	if command == "" {
		return opts, "help", nil, nil
	}
	switch command {
	case "up", "down", "status", "check", "create", "privacy-matrix":
		return opts, command, commandArgs, nil
	default:
		return opts, "", nil, fmt.Errorf("unknown command %s", command)
	}
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: migrate [--migrations-dir DIR] [--backfill-manifest FILE] <up|down|status|check|create|privacy-matrix>")
}
