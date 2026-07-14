package errors

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWrap_BasicShape(t *testing.T) {
	err := Wrap(CodeAuthUnauthorized, "needs auth", false)
	if err.Code != "AUTH_UNAUTHORIZED" {
		t.Errorf("Code = %q, want AUTH_UNAUTHORIZED", err.Code)
	}
	if err.Message != "needs auth" {
		t.Errorf("Message = %q, want %q", err.Message, "needs auth")
	}
	if err.Retryable {
		t.Errorf("Retryable = true, want false")
	}
}

func TestAPIError_ImplementsErrorInterface(t *testing.T) {
	var iface error = Wrap(CodeRateLimited, "slow down", true)
	if !strings.Contains(iface.Error(), "RATE_LIMITED") {
		t.Errorf("Error() = %q, expected to contain RATE_LIMITED", iface.Error())
	}
}

func TestAPIError_JSONShape(t *testing.T) {
	e := Wrap(CodeValidationFailed, "bad input", false)
	e.RequestID = "req_01HV"
	e.Details = map[string]any{"field": "email"}

	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	wantKeys := []string{"code", "message", "requestId", "retryable", "details"}
	for _, k := range wantKeys {
		if _, ok := got[k]; !ok {
			t.Errorf("APIError JSON missing key %q (raw: %s)", k, raw)
		}
	}
	if got["code"] != "VALIDATION_FAILED" {
		t.Errorf("code = %v, want VALIDATION_FAILED", got["code"])
	}
	if got["retryable"] != false {
		t.Errorf("retryable = %v, want false", got["retryable"])
	}
}

func TestAPIError_OmitDetailsWhenNil(t *testing.T) {
	e := Wrap(CodeReportNotReady, "wait", true)
	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(raw), "details") {
		t.Errorf("expected `details` to be omitted when nil; got %s", raw)
	}
}

func TestWrap_UsesGeneratedConstants(t *testing.T) {
	// Ensures hand-written Wrap helper interoperates with the generated CodeRegistry.
	for _, code := range AllCodes {
		meta, ok := CodeRegistry[code]
		if !ok {
			t.Errorf("CodeRegistry missing entry for %s", code)
			continue
		}
		e := Wrap(code, meta.Message, meta.Retryable)
		if e.Code != code {
			t.Errorf("Wrap(%s).Code = %q, want %q", code, e.Code, code)
		}
		if e.Retryable != meta.Retryable {
			t.Errorf("Wrap(%s).Retryable = %v, want %v", code, e.Retryable, meta.Retryable)
		}
	}
}

// TargetJob paste-only intake keeps generic validation/import failures and
// removes source-specific errors that only applied to URL/file ingestion.
func TestTargetJobPasteOnlyErrorCodes_Documented(t *testing.T) {
	cases := []struct {
		code      string
		message   string
		retryable bool
	}{
		{"VALIDATION_FAILED", "request validation failed", false},
		{"TARGET_IMPORT_FAILED", "failed to import target job", true},
		{"TARGET_JOB_NOT_FOUND", "target job not found", false},
		{"TARGET_INVALID_STATE_TRANSITION", "target job state transition is not allowed", false},
	}
	for _, tc := range cases {
		meta, ok := CodeRegistry[tc.code]
		if !ok {
			t.Errorf("CodeRegistry missing %s", tc.code)
			continue
		}
		if meta.Retryable != tc.retryable {
			t.Errorf("%s retryable = %v, want %v", tc.code, meta.Retryable, tc.retryable)
		}
		if meta.Message != tc.message {
			t.Errorf("%s message = %q, want %q", tc.code, meta.Message, tc.message)
		}
		if !contains(AllCodes, tc.code) {
			t.Errorf("AllCodes missing %s", tc.code)
		}
	}

	for _, removed := range []string{
		"TARGET_IMPORT_SOURCE_INVALID",
		"TARGET_IMPORT_SOURCE_UNAVAILABLE",
	} {
		if _, ok := CodeRegistry[removed]; ok {
			t.Errorf("CodeRegistry still contains removed source-specific code %s", removed)
		}
		if contains(AllCodes, removed) {
			t.Errorf("AllCodes still contains removed source-specific code %s", removed)
		}
	}
}

func TestResumeExportErrorCode_Documented(t *testing.T) {
	meta, ok := CodeRegistry[CodeResumeExportNotAvailable]
	if !ok {
		t.Fatalf("CodeRegistry missing %s", CodeResumeExportNotAvailable)
	}
	if meta.Retryable {
		t.Fatalf("%s retryable = true, want false", CodeResumeExportNotAvailable)
	}
	if meta.Message == "" {
		t.Fatalf("%s message must not be empty", CodeResumeExportNotAvailable)
	}
	if !contains(AllCodes, CodeResumeExportNotAvailable) {
		t.Fatalf("AllCodes missing %s", CodeResumeExportNotAvailable)
	}
}

func TestReportNotFoundErrorCode_Documented(t *testing.T) {
	meta, ok := CodeRegistry[CodeReportNotFound]
	if !ok {
		t.Fatalf("CodeRegistry missing %s", CodeReportNotFound)
	}
	if meta.Retryable {
		t.Fatalf("%s retryable = true, want false", CodeReportNotFound)
	}
	if meta.Message != "feedback report not found or not accessible" {
		t.Fatalf("%s message = %q", CodeReportNotFound, meta.Message)
	}
	if !contains(AllCodes, CodeReportNotFound) {
		t.Fatalf("AllCodes missing %s", CodeReportNotFound)
	}
}

func TestReportContextTooLargeErrorCode_Documented(t *testing.T) {
	meta, ok := CodeRegistry[CodeReportContextTooLarge]
	if !ok {
		t.Fatalf("CodeRegistry missing %s", CodeReportContextTooLarge)
	}
	if meta.Message != "report context exceeds supported generation size" {
		t.Fatalf("%s message = %q", CodeReportContextTooLarge, meta.Message)
	}
	if meta.Retryable {
		t.Fatalf("%s retryable = true, want false", CodeReportContextTooLarge)
	}
	if !contains(AllCodes, CodeReportContextTooLarge) {
		t.Fatalf("AllCodes missing %s", CodeReportContextTooLarge)
	}
}

func contains[T comparable](values []T, want T) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
