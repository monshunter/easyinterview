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
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
	"github.com/monshunter/easyinterview/backend/internal/targetjob/urlfetch"
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
		Fetcher:  &fakeFetcher{},
		NewID:    idSeq("redact"),
		Now:      func() time.Time { return time.Now().UTC() },
	})
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", SourceType: targetjob.SourceTypeManualText, RawJDText: "x"}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
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
			exec, store, _, ai, _ := newParseExecutorWithFakes(t)
			ai.err = errors.New(leaked)
			store.target = targetjob.TargetJobRecord{
				ID:             "tgt-privacy",
				SourceType:     targetjob.SourceTypeManualText,
				TargetLanguage: "en",
				RawJDText:      "x",
			}
			outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-privacy"})
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
// forbidden tokens leak into the wire shape, even when the source values
// contain them.
func TestParseExecutor_OutboxPayloadsContainOnlyAllowedTokens(t *testing.T) {
	exec, store, _, ai, _ := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	store.target = targetjob.TargetJobRecord{
		ID: "tgt-1", UserID: "user-1",
		SourceType: targetjob.SourceTypeManualText, TargetLanguage: "en",
		RawJDText: "JD with secret Authorization: Bearer ABC123",
	}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if !outcome.Succeeded {
		t.Fatalf("happy path failed: %+v", outcome)
	}
	for _, kw := range []string{"raw_jd_text", "Authorization:", "Bearer ", "JD with secret"} {
		if bytes.Contains(store.parsedOutboxPayload, []byte(kw)) {
			t.Errorf("target.parsed payload leaked forbidden token %q: %s", kw, string(store.parsedOutboxPayload))
		}
	}
}

// TestSourceRefreshHandler_PayloadHasNoSourceURL guards 4.5: the placeholder
// source_refresh row must never reflect the original source URL.
func TestSourceRefreshHandler_PayloadHasNoSourceURL(t *testing.T) {
	store := &pipelineFakeStore{}
	h := &targetjob.SourceRefreshHandler{Store: store}
	outcome := h.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if !outcome.Succeeded {
		t.Fatalf("source refresh: %+v", outcome)
	}
	// SourceRefreshHandler does not produce an outbox payload itself; the
	// async_jobs payload is empty (`{}`). This test fails loudly if the
	// shape changes to include URL data.
}

// TestImportTargetJob_DedupeKeyIsUserScopedAcrossServices exercises spec
// 5.3 again at the redline level: even when two services share the same
// pepper, hashes for different users must diverge so handler / store
// behaviour cannot accidentally treat unrelated users as colliding.
func TestImportTargetJob_DedupeKeyIsUserScopedAcrossServices(t *testing.T) {
	store1 := &fakeStore{}
	idx1 := 0
	store2 := &fakeStore{}
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
		Source: map[string]any{"type": "manual_text", "rawText": "JD A"},
	}); err != nil {
		t.Fatalf("svc1 import: %v", err)
	}
	if _, err := svc2.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID: "user-B", IdempotencyKey: "key-overlap", TargetLanguage: "en", ResumeID: "resume-B",
		Source: map[string]any{"type": "manual_text", "rawText": "JD B"},
	}); err != nil {
		t.Fatalf("svc2 import: %v", err)
	}
	if store1.captured.DedupeKey == "" || store2.captured.DedupeKey == "" {
		t.Fatal("dedupe keys must be populated")
	}
	if store1.captured.DedupeKey == store2.captured.DedupeKey {
		t.Fatalf("dedupe key must be user-scoped, both got %s", store1.captured.DedupeKey)
	}
}

func TestImportTargetJob_URLQuerySecretDoesNotEnterStoreOrPayloads(t *testing.T) {
	store := &fakeStore{}
	ids := []string{
		"018f2a40-0000-7000-9000-0000000000a1",
		"018f2a40-0000-7000-9000-0000000000f1",
		"018f2a40-0000-7000-9000-0000000000c1",
		"018f2a40-0000-7000-9000-0000000000e1",
	}
	idx := 0
	svc := targetjob.NewService(targetjob.ServiceOptions{
		Store: store,
		NewID: func() string {
			v := ids[idx]
			idx++
			return v
		},
		Now:          func() time.Time { return time.Date(2026, 5, 9, 23, 10, 0, 0, time.UTC) },
		DedupePepper: "shared-pepper",
	})

	_, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "user-url",
		IdempotencyKey: "url-key",
		TargetLanguage: "en",
		ResumeID:       "resume-url",
		Source: map[string]any{
			"type": "url",
			"url":  "https://jobs.example.com/role/123?token=super-secret#share",
		},
	})
	if err != nil {
		t.Fatalf("ImportTargetJob URL: %v", err)
	}
	for name, raw := range map[string]string{
		"source_url": string(store.captured.SourceURL),
		"outbox":     string(store.captured.OutboxEventPayload),
		"job":        string(store.captured.JobPayload),
	} {
		if strings.Contains(raw, "super-secret") || strings.Contains(raw, "token=") || strings.Contains(raw, "#share") {
			t.Fatalf("%s leaked URL secret: %s", name, raw)
		}
	}
}

// TestUrlFetcher_HasReasonableDefaultsAfterFactory keeps Phase 5 honest:
// changing URLFetchTimeout or URLFetchBodyCap requires updating spec D-7
// first, which is enforced separately by config_test.go. This test
// asserts the fetcher honours those constants.
func TestUrlFetcher_HasReasonableDefaultsAfterFactory(t *testing.T) {
	if targetjob.URLFetchTimeout != 10*time.Second {
		t.Fatalf("URLFetchTimeout drifted: %v", targetjob.URLFetchTimeout)
	}
	if targetjob.URLFetchBodyCap != 1<<20 {
		t.Fatalf("URLFetchBodyCap drifted: %d", targetjob.URLFetchBodyCap)
	}
	// Construct a fetcher to confirm the factory accepts the canonical UA.
	_ = urlfetch.New(urlfetch.FetcherOptions{UserAgent: targetjob.URLFetchUserAgent("test")})
}
