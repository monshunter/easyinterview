package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var migrationFilePattern = regexp.MustCompile(`^([0-9]{6})_[a-z0-9]+(?:_[a-z0-9]+)*\.(up|down)\.sql$`)

// CreateMigrationFiles creates the next paired up/down migration files.
func CreateMigrationFiles(dir, name string) ([]string, error) {
	if dir == "" {
		return nil, fmt.Errorf("migrations dir is required")
	}
	cleanName, err := normalizeMigrationName(name)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	next, err := nextMigrationVersion(dir)
	if err != nil {
		return nil, err
	}
	base := fmt.Sprintf("%06d_%s", next, cleanName)
	files := []string{
		filepath.Join(dir, base+".up.sql"),
		filepath.Join(dir, base+".down.sql"),
	}
	for _, file := range files {
		f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return nil, err
		}
		if err := f.Close(); err != nil {
			return nil, err
		}
	}
	return files, nil
}

func normalizeMigrationName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("migration NAME is required")
	}
	name = strings.ReplaceAll(name, "-", "_")
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	if ok, _ := regexp.MatchString(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`, name); !ok {
		return "", fmt.Errorf("migration NAME must be lower_snake_case")
	}
	return name, nil
}

func nextMigrationVersion(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	versions := make([]int, 0)
	seen := map[int]bool{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := migrationFilePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(match[1], "%d", &n); err != nil {
			return 0, err
		}
		if !seen[n] {
			versions = append(versions, n)
			seen[n] = true
		}
	}
	sort.Ints(versions)
	if len(versions) == 0 {
		return 1, nil
	}
	return versions[len(versions)-1] + 1, nil
}
