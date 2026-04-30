package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ValidateMigrationFiles enforces the B4 flat NNNNNN_name.{up,down}.sql contract.
func ValidateMigrationFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	type pair struct {
		name string
		up   bool
		down bool
	}
	pairs := map[int]*pair{}
	var versions []int
	var problems []string

	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "backfill" {
				continue
			}
			problems = append(problems, fmt.Sprintf("%s is a directory; migrations must be flat", entry.Name()))
			continue
		}
		if filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		match := migrationFilePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			problems = append(problems, fmt.Sprintf("invalid migration file name: %s", entry.Name()))
			continue
		}
		var version int
		if _, err := fmt.Sscanf(match[1], "%d", &version); err != nil {
			problems = append(problems, fmt.Sprintf("invalid version in %s: %v", entry.Name(), err))
			continue
		}
		if _, ok := pairs[version]; !ok {
			pairs[version] = &pair{name: strings.TrimSuffix(strings.TrimSuffix(entry.Name(), ".up.sql"), ".down.sql")}
			versions = append(versions, version)
		}
		if match[2] == "up" {
			pairs[version].up = true
		} else {
			pairs[version].down = true
		}
	}

	sort.Ints(versions)
	for i, version := range versions {
		expected := i + 1
		if version != expected {
			problems = append(problems, fmt.Sprintf("expected version %06d, found %06d", expected, version))
		}
		p := pairs[version]
		if !p.up {
			problems = append(problems, fmt.Sprintf("missing up migration for %s", p.name))
		}
		if !p.down {
			problems = append(problems, fmt.Sprintf("missing down migration for %s", p.name))
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("migration file contract failed: %s", strings.Join(problems, "; "))
	}
	return nil
}
