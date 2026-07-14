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
		{`target_job_sources`, "paste-only TargetJob persistence has no source table"},
		{`source_refresh`, "paste-only TargetJob parsing has no source refresh job"},
		{`SourceRefreshHandler`, "paste-only TargetJob parsing has no source refresh handler"},
		{`URLFetcher`, "paste-only TargetJob parsing has no URL fetch boundary"},
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

func TestURLFetchPackageIsDeleted(t *testing.T) {
	files, err := filepath.Glob("urlfetch/*.go")
	if err != nil {
		t.Fatalf("glob urlfetch package: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("urlfetch package must be absent, found %v", files)
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
