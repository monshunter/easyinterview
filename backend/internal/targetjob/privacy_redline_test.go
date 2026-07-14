package targetjob_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

// TestPackageSourcesContainNoForbiddenLogStrings is the package-level
// privacy gate: no production source file in the targetjob domain may
// contain forbidden sentinel strings inside string literals that would
// flow into log lines, error messages, metric labels, or audit metadata.
//
// We intentionally allow these tokens to appear inside the `Forbidden*`
// allowlist constants (those define the redline) and the redactor
// implementation — those are the gate, not the leak.
func TestPackageSourcesContainNoForbiddenLogStrings(t *testing.T) {
	forbiddenLiterals := []string{
		`"raw_jd_text"`,
		`"sourceUrl":"http`,
		`"promptBody"`,
		`"responseBody"`,
		`"providerSecret"`,
		`"Authorization":"Bearer`,
	}
	// Skip the gate definition files: they declare the redline, they do
	// not leak through it.
	gateDefs := map[string]bool{
		"payload.go":        true,
		"parse_executor.go": true,
	}
	files := goSourceFiles(t)
	for _, f := range files {
		if gateDefs[filepath.Base(f)] {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, kw := range forbiddenLiterals {
			if bytes.Contains(raw, []byte(kw)) {
				t.Errorf("file %s contains forbidden literal %q in production source", f, kw)
			}
		}
	}
}

func goSourceFiles(t *testing.T) []string {
	t.Helper()
	out := []string{}
	matches, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	for _, m := range matches {
		if strings.HasSuffix(m, "_test.go") {
			continue
		}
		out = append(out, m)
	}
	return out
}

// TestRedactErrorMessage_StripsForbiddenTokens covers the runtime redactor
// used by ParseExecutor.fail. Error messages that would otherwise leak
// raw_jd_text, Authorization headers, or Bearer tokens collapse to a
// generic redacted marker before the outcome reaches the async_jobs row.
func TestParseExecutor_RedactsForbiddenTokensInErrorMessage(t *testing.T) {
	store := &pipelineFakeStore{}
	registry := &fakeRegistry{err: errors.New("provider says raw_jd_text=hello world")}
	exec := targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store:    store,
		Registry: registry,
		AI:       &fakeAIClient{},
		NewID:    idSeq("redact"),
		Now:      func() time.Time { return time.Now().UTC() },
	})
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", RawJDText: "x"}
	outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.Succeeded {
		t.Fatal("expected failure, got success")
	}
	if strings.Contains(outcome.ErrorMessage, "raw_jd_text") {
		t.Fatalf("error message must redact raw_jd_text, got %q", outcome.ErrorMessage)
	}
	if outcome.ErrorMessage != "AI_PROVIDER_CONFIG_INVALID" {
		t.Fatalf("registry failure must persist code-based safe summary, got %q", outcome.ErrorMessage)
	}
}

func TestParseExecutor_RedactsPromptResponseAndProviderSecretInErrorMessage(t *testing.T) {
	for _, leaked := range []string{
		"provider secret leaked from upstream",
		"prompt body: private JD text",
		"response body: model returned private JD text",
		"Private JD body that must not leak",
	} {
		t.Run(leaked, func(t *testing.T) {
			exec, store, _, ai := newParseExecutorWithFakes(t)
			ai.err = errors.New(leaked)
			store.target = targetjob.TargetJobRecord{
				ID:             "tgt-privacy",
				TargetLanguage: "en",
				RawJDText:      "x",
			}
			outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-privacy"})
			if outcome.Succeeded {
				t.Fatal("expected failure, got success")
			}
			if strings.Contains(outcome.ErrorMessage, leaked) ||
				strings.Contains(outcome.ErrorMessage, "provider secret") ||
				strings.Contains(outcome.ErrorMessage, "prompt body") ||
				strings.Contains(outcome.ErrorMessage, "response body") ||
				strings.Contains(outcome.ErrorMessage, "Private JD body") {
				t.Fatalf("error message leaked forbidden token: %q", outcome.ErrorMessage)
			}
			if outcome.ErrorMessage != "AI_FALLBACK_EXHAUSTED" {
				t.Fatalf("AI failure must persist code-based safe summary, got %q", outcome.ErrorMessage)
			}
		})
	}
}

// TestParseExecutor_OutboxPayloadsContainOnlyAllowedTokens scans the
// outbox payload bytes the executor produces and asserts none of the
// forbidden tokens leak into the wire shape, even when the pasted JD text
// contains them.
func TestParseExecutor_OutboxPayloadsContainOnlyAllowedTokens(t *testing.T) {
	exec, store, _, ai := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	store.target = targetjob.TargetJobRecord{
		ID: "tgt-1", UserID: "user-1",
		TargetLanguage: "en",
		RawJDText:      "JD with secret Authorization: Bearer ABC123",
	}
	outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-1"})
	if !outcome.Succeeded {
		t.Fatalf("happy path failed: %+v", outcome)
	}
	for _, kw := range []string{"raw_jd_text", "Authorization:", "Bearer ", "JD with secret"} {
		if bytes.Contains(store.parsedOutboxPayload, []byte(kw)) {
			t.Errorf("target.parsed payload leaked forbidden token %q: %s", kw, string(store.parsedOutboxPayload))
		}
	}
}

// TestImportTargetJob_DedupeKeyIsUserScopedAcrossServices exercises spec
// 5.3 again at the redline level: even when two services share the same
// pepper, hashes for different users must diverge so handler / store
// behaviour cannot accidentally treat unrelated users as colliding.
func TestImportTargetJob_DedupeKeyIsUserScopedAcrossServices(t *testing.T) {
	store1 := &pipelineFakeStore{}
	idx1 := 0
	store2 := &pipelineFakeStore{}
	idx2 := 0
	gen1 := func() string {
		idx1++
		return fmt.Sprintf("svc1-id-%d", idx1)
	}
	gen2 := func() string {
		idx2++
		return fmt.Sprintf("svc2-id-%d", idx2)
	}
	now := time.Date(2026, 5, 9, 23, 0, 0, 0, time.UTC)
	svc1 := targetjob.NewService(targetjob.ServiceOptions{Store: store1, NewID: gen1, Now: func() time.Time { return now }, DedupePepper: "shared-pepper"})
	svc2 := targetjob.NewService(targetjob.ServiceOptions{Store: store2, NewID: gen2, Now: func() time.Time { return now }, DedupePepper: "shared-pepper"})

	if _, err := svc1.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID: "user-A", IdempotencyKey: "key-overlap", TargetLanguage: "en", ResumeID: "resume-A",
		RawText: "JD A",
	}); err != nil {
		t.Fatalf("svc1 import: %v", err)
	}
	if _, err := svc2.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID: "user-B", IdempotencyKey: "key-overlap", TargetLanguage: "en", ResumeID: "resume-B",
		RawText: "JD B",
	}); err != nil {
		t.Fatalf("svc2 import: %v", err)
	}
	if store1.importIn == nil || store2.importIn == nil {
		t.Fatal("import inputs must be captured")
	}
	if store1.importIn.DedupeKey == "" || store2.importIn.DedupeKey == "" {
		t.Fatal("dedupe keys must be populated")
	}
	if store1.importIn.DedupeKey == store2.importIn.DedupeKey {
		t.Fatalf("dedupe key must be user-scoped, both got %s", store1.importIn.DedupeKey)
	}
}
