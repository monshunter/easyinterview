package config_test

import (
	"path/filepath"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

func TestGettersDotPathAndTypes(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
app:
  listenAddr: ":8080"
log:
  level: info
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
featureFlag:
  posthogSelfHosted: true
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := loader.GetString("app.listenAddr"); got != ":8080" {
		t.Errorf("GetString: %q", got)
	}
	if got := loader.GetInt("async.queueWeights.critical"); got != 6 {
		t.Errorf("GetInt: %d", got)
	}
	if got := loader.GetBool("featureFlag.posthogSelfHosted"); got != true {
		t.Errorf("GetBool: %v", got)
	}
}

func TestGettersDoNotPanicOnMissing(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), "")
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := loader.GetString("does.not.exist"); got != "" {
		t.Errorf("GetString missing: %q", got)
	}
	if got := loader.GetInt("does.not.exist"); got != 0 {
		t.Errorf("GetInt missing: %d", got)
	}
	if got := loader.GetBool("does.not.exist"); got {
		t.Errorf("GetBool missing: true")
	}
}
