package targetjob_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestActiveScopeNegativeSearch enforces plan §6.3: the active TargetJob
// implementation must never reintroduce concepts the current product
// scope has dropped. Each forbidden token below maps to a removed module
// or stale alias; finding any of them in a non-test source file means
// the implementation has drifted away from the active spec.
func TestActiveScopeNegativeSearch(t *testing.T) {
	cases := []struct {
		token  string
		reason string
	}{
		{`mistake.`, "Mistakes module is not in the active P0 scope"},
		{`growth.`, "Growth tracking module is not in the active P0 scope"},
		{`"voice"`, "voice is no longer a top-level route; speech features live under practice"},
		{`practice.voice.tts`, "TTS is owned by ai-provider-and-model-routing, not targetjob"},
		{`"jd.parse"`, "stale feature_key alias replaced by target.import.parse"},
		{`"target.parse"`, "stale feature_key alias replaced by target.import.parse"},
		{`embedding`, "embedding capability was removed from the active AI provider scope"},
		{`rerank`, "rerank capability was removed from the active AI provider scope"},
		{`interview_round`, "interview_round was rejected as a separate module; rounds live inside practice"},
		// F3 prompt-rubric-registry/001-baseline phase 3 removal scope:
		// the StaticPromptRegistry bridge plus its four constants must
		// stay deleted now that RegistryAdapter is the only route to F3.
		{`StaticPromptRegistry`, "out-of-scope after plan 001-baseline phase 3; use RegistryAdapter instead"},
		{`defaultTargetImportPromptVersion`, "out-of-scope after plan 001-baseline phase 3; F3 owns prompt versions now"},
		{`defaultTargetImportRubricVersion`, "out-of-scope after plan 001-baseline phase 3; F3 owns rubric versions now"},
		{`defaultTargetImportModelProfileName`, "out-of-scope after plan 001-baseline phase 3; A3 owns profile names now"},
		{`defaultTargetImportDataSourceVersion`, "out-of-scope after plan 001-baseline phase 3; F3 owns data source versions now"},
	}

	matches, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	for _, f := range matches {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, c := range cases {
			if bytes.Contains(raw, []byte(c.token)) {
				t.Errorf("file %s reintroduces forbidden token %q (%s)", f, c.token, c.reason)
			}
		}
	}
}

// TestActiveScopeNegativeSearchInUrlfetch covers the urlfetch sub-package
// alongside the main targetjob package. The same forbidden tokens apply.
func TestActiveScopeNegativeSearchInUrlfetch(t *testing.T) {
	tokens := []string{`mistake.`, `growth.`, `"jd.parse"`, `"target.parse"`, `embedding`, `rerank`, `interview_round`}
	files, err := filepath.Glob("urlfetch/*.go")
	if err != nil {
		t.Fatalf("glob urlfetch: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, kw := range tokens {
			if bytes.Contains(raw, []byte(kw)) {
				t.Errorf("urlfetch file %s contains forbidden token %q", f, kw)
			}
		}
	}
}

func TestPackageDocReflectsCompletedScenarioGateState(t *testing.T) {
	raw, err := os.ReadFile("doc.go")
	if err != nil {
		t.Fatalf("read doc.go: %v", err)
	}
	for _, stale := range []string{"remain unchecked", "scenarios stay parked", "until F3 lands"} {
		if strings.Contains(string(raw), stale) {
			t.Fatalf("doc.go still carries stale BDD handoff state %q", stale)
		}
	}
}
