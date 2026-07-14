package targetjob_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestIsTestAppEnvAcceptsOnlyTest(t *testing.T) {
	cases := map[string]bool{
		"test":       true,
		"TEST":       true,
		" test ":     true,
		"":           false,
		"dev":        false,
		"production": false,
		"kind":       false,
		"staging":    false,
		"testing":    false, // sub-string but not equal — must fail
	}
	for v, want := range cases {
		if got := targetjob.IsTestAppEnv(v); got != want {
			t.Errorf("IsTestAppEnv(%q) = %v, want %v", v, got, want)
		}
	}
}

// TestTargetJobFilesContainNoAppLevelProxyKeys prevents a removed URL intake
// concern from returning through domain configuration.
func TestTargetJobFilesContainNoAppLevelProxyKeys(t *testing.T) {
	forbidden := []string{
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"NO_PROXY",
		"OUTBOUND_PROXY",
		"FETCH_PROXY",
		"TARGETJOB_PROXY",
		"TARGET_JOB_PROXY",
		"EI_OUTBOUND_PROXY",
	}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, kw := range forbidden {
			if bytes.Contains(raw, []byte(kw)) {
				t.Errorf("file %s contains forbidden proxy key %q; route via deployment instead", f, kw)
			}
		}
	}
}
