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
