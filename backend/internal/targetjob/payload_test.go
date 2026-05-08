package targetjob_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestBuildTargetImportRequestedPayload_HappyPathMapsAPISource(t *testing.T) {
	p, err := targetjob.BuildTargetImportRequestedPayload(targetjob.TargetImportRequestedInput{
		APISourceType:  targetjob.SourceTypeManualText,
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		TargetLanguage: "zh-CN",
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
	})
	if err != nil {
		t.Fatalf("BuildTargetImportRequestedPayload: %v", err)
	}
	if p.SourceType != events.TargetImportSourceTypeText {
		t.Fatalf("manual_text must map to event-local text, got %q", p.SourceType)
	}
	if p.UserID != "018f2a40-0000-7000-9000-0000000000b1" {
		t.Fatalf("UserID lost: %+v", p)
	}
}

func TestBuildTargetImportRequestedPayload_RejectsManualForm(t *testing.T) {
	_, err := targetjob.BuildTargetImportRequestedPayload(targetjob.TargetImportRequestedInput{
		APISourceType:  targetjob.SourceTypeManualForm,
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		TargetLanguage: "en",
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
	})
	if !errors.Is(err, events.ErrManualFormNotEventSource) {
		t.Fatalf("expected ErrManualFormNotEventSource, got %v", err)
	}
}

func TestBuildTargetImportRequestedPayload_RejectsForbiddenIDs(t *testing.T) {
	// Inject a forbidden token through a UUID-shaped string so the
	// substring scan still catches it. Real callers wouldn't pass these,
	// but the gate must reject them anyway.
	_, err := targetjob.BuildTargetImportRequestedPayload(targetjob.TargetImportRequestedInput{
		APISourceType:  targetjob.SourceTypeURL,
		TargetJobID:    "raw_jd_text-leak", // smuggle the forbidden string through a string field
		TargetLanguage: "en",
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
	})
	if err == nil || !strings.Contains(err.Error(), "raw_jd_text") {
		t.Fatalf("expected forbidden-token rejection, got %v", err)
	}
}

func TestBuildTargetImportRequestedPayload_RequiresMandatoryFields(t *testing.T) {
	cases := []struct {
		name string
		in   targetjob.TargetImportRequestedInput
	}{
		{"missing target", targetjob.TargetImportRequestedInput{APISourceType: targetjob.SourceTypeURL, TargetLanguage: "en", UserID: "u"}},
		{"missing user", targetjob.TargetImportRequestedInput{APISourceType: targetjob.SourceTypeURL, TargetJobID: "t", TargetLanguage: "en"}},
		{"missing language", targetjob.TargetImportRequestedInput{APISourceType: targetjob.SourceTypeURL, TargetJobID: "t", UserID: "u"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := targetjob.BuildTargetImportRequestedPayload(tc.in); err == nil {
				t.Fatalf("%s: expected error", tc.name)
			}
		})
	}
}

func TestBuildTargetParsedPayload_HappyPath(t *testing.T) {
	p, err := targetjob.BuildTargetParsedPayload(targetjob.TargetParsedInput{
		TargetJobID:      "018f2a40-0000-7000-9000-0000000000a1",
		UserID:           "018f2a40-0000-7000-9000-0000000000b1",
		AnalysisStatus:   sharedtypes.TargetJobParseStatusReady,
		RequirementCount: 3,
		CoreThemes:       []string{"api", "scaling"},
	})
	if err != nil {
		t.Fatalf("BuildTargetParsedPayload: %v", err)
	}
	if p.RequirementCount != 3 || len(p.CoreThemes) != 2 {
		t.Fatalf("unexpected payload shape: %+v", p)
	}
}

func TestBuildTargetParsedPayload_RejectsNegativeCount(t *testing.T) {
	_, err := targetjob.BuildTargetParsedPayload(targetjob.TargetParsedInput{
		TargetJobID:      "t",
		UserID:           "u",
		AnalysisStatus:   sharedtypes.TargetJobParseStatusReady,
		RequirementCount: -1,
	})
	if err == nil {
		t.Fatal("expected error on negative requirementCount")
	}
}

func TestBuildTargetAnalysisFailedPayload_HappyPath(t *testing.T) {
	p, err := targetjob.BuildTargetAnalysisFailedPayload(targetjob.TargetAnalysisFailedInput{
		TargetJobID: "018f2a40-0000-7000-9000-0000000000a1",
		ErrorCode:   "TARGET_IMPORT_SOURCE_INVALID",
		Retryable:   false,
	})
	if err != nil {
		t.Fatalf("BuildTargetAnalysisFailedPayload: %v", err)
	}
	if p.ErrorCode != "TARGET_IMPORT_SOURCE_INVALID" || p.Retryable {
		t.Fatalf("unexpected payload: %+v", p)
	}
}

func TestBuildTargetAnalysisFailedPayload_AllowsDocumentedProviderSecretCode(t *testing.T) {
	p, err := targetjob.BuildTargetAnalysisFailedPayload(targetjob.TargetAnalysisFailedInput{
		TargetJobID: "018f2a40-0000-7000-9000-0000000000a1",
		ErrorCode:   "AI_PROVIDER_SECRET_MISSING",
		Retryable:   false,
	})
	if err != nil {
		t.Fatalf("documented B1 provider-secret code must be allowed: %v", err)
	}
	if p.ErrorCode != "AI_PROVIDER_SECRET_MISSING" {
		t.Fatalf("unexpected code: %+v", p)
	}
}

func TestBuildTargetAnalysisFailedPayload_RejectsUndocumentedError(t *testing.T) {
	_, err := targetjob.BuildTargetAnalysisFailedPayload(targetjob.TargetAnalysisFailedInput{
		TargetJobID: "t",
		ErrorCode:   "Authorization: Bearer leaked",
		Retryable:   false,
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "documented b1 code") {
		t.Fatalf("expected documented-code rejection, got %v", err)
	}
}

func TestBuildTargetImportJobPayload_HappyPath(t *testing.T) {
	raw, err := targetjob.BuildTargetImportJobPayload(targetjob.TargetImportJobPayload{
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
		SourceType:     "manual_text",
		TargetLanguage: "en",
	})
	if err != nil {
		t.Fatalf("BuildTargetImportJobPayload: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal job payload: %v", err)
	}
	for _, want := range []string{"targetJobId", "userId", "sourceType", "targetLanguage"} {
		if _, ok := got[want]; !ok {
			t.Errorf("missing field %q in job payload", want)
		}
	}
	for _, forbidden := range []string{"rawJdText", "raw_jd_text", "promptBody", "prompt_body", "Authorization"} {
		if _, ok := got[forbidden]; ok {
			t.Errorf("forbidden field %q leaked into job payload", forbidden)
		}
	}
}

func TestB3OutboxPayloadStructsContainNoForbiddenFieldNames(t *testing.T) {
	cases := []struct {
		name string
		t    reflect.Type
	}{
		{"TargetImportRequestedPayload", reflect.TypeFor[events.TargetImportRequestedPayload]()},
		{"TargetParsedPayload", reflect.TypeFor[events.TargetParsedPayload]()},
		{"TargetAnalysisFailedPayload", reflect.TypeFor[events.TargetAnalysisFailedPayload]()},
	}
	forbidden := []string{
		"raw", "rawJd", "rawJdText", "rawJDText",
		"sourceUrl", "fileUrl", "fileObjectUrl",
		"promptBody", "responseBody", "providerSecret",
		"authorization",
	}
	for _, tc := range cases {
		for i := 0; i < tc.t.NumField(); i++ {
			fname := strings.ToLower(tc.t.Field(i).Name)
			tag := strings.ToLower(tc.t.Field(i).Tag.Get("json"))
			for _, kw := range forbidden {
				kwl := strings.ToLower(kw)
				if strings.Contains(fname, kwl) || strings.Contains(tag, kwl) {
					t.Errorf("%s.%s contains forbidden token %q (json tag %q)", tc.name, tc.t.Field(i).Name, kw, tag)
				}
			}
		}
	}
}
