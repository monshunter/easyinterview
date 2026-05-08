package targetjob_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestURLFetchTimeoutMatchesSpecD7(t *testing.T) {
	if targetjob.URLFetchTimeout != 10*time.Second {
		t.Fatalf("URLFetchTimeout drifted from spec D-7 (10s): %v", targetjob.URLFetchTimeout)
	}
}

func TestURLFetchBodyCapMatchesSpecD7(t *testing.T) {
	if targetjob.URLFetchBodyCap != 1<<20 {
		t.Fatalf("URLFetchBodyCap drifted from spec D-7 (1 MiB): %d", targetjob.URLFetchBodyCap)
	}
}

func TestURLFetchUserAgentMatchesSpecD7(t *testing.T) {
	ua := targetjob.URLFetchUserAgent("v1.0.0")
	if !strings.HasPrefix(ua, "EasyInterview JD-Crawler/v1.0.0") {
		t.Fatalf("UA must use exact crawler prefix per spec D-7: %q", ua)
	}
	if !strings.Contains(ua, "+https://") {
		t.Fatalf("UA must include contact URL per spec D-7: %q", ua)
	}
	if !strings.Contains(targetjob.URLFetchUserAgent(""), "JD-Crawler/dev") {
		t.Fatal("empty version must default to dev so boot paths still emit a valid UA")
	}
	if !strings.Contains(targetjob.URLFetchUserAgent("  v9  "), "JD-Crawler/v9") {
		t.Fatal("version trimming dropped surrounding whitespace")
	}
}

func TestMustNotIntroduceAppLevelConfigKeyPanicsWithA4Hint(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("panic value not string: %v", r)
		}
		if !strings.Contains(msg, "secrets-and-config") {
			t.Fatalf("panic must mention A4 owner spec: %q", msg)
		}
		if !strings.Contains(msg, "TARGETJOB_PROXY") {
			t.Fatalf("panic must echo the offending key: %q", msg)
		}
	}()
	targetjob.MustNotIntroduceAppLevelConfigKey("TARGETJOB_PROXY")
}

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

// TestTargetJobSourceFilesContainNoAppLevelProxyKeys enforces spec D-7 / plan
// 1.2: outbound proxy is not an app-level concern of this domain. If users
// need a corporate egress proxy, the deployment platform handles it.
func TestTargetJobSourceFilesContainNoAppLevelProxyKeys(t *testing.T) {
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

// TestTargetJobSourceFilesDoNotReadOsGetenv enforces A4 ownership of
// app-level config: this domain receives configuration through constructor
// parameters, not via direct os.Getenv calls.
func TestTargetJobSourceFilesDoNotReadOsGetenv(t *testing.T) {
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
		if bytes.Contains(raw, []byte("os.Getenv")) {
			t.Errorf("file %s calls os.Getenv directly; route through A4 secrets-and-config", f)
		}
		if bytes.Contains(raw, []byte("os.LookupEnv")) {
			t.Errorf("file %s calls os.LookupEnv directly; route through A4 secrets-and-config", f)
		}
	}
}
