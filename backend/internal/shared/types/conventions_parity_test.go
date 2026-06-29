package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sharedai "github.com/monshunter/easyinterview/backend/internal/shared/ai"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type parityFixture struct {
	Enums                    map[string][]string `json:"enums"`
	ErrorCodes               []string            `json:"errorCodes"`
	AICapabilities           []string            `json:"aiCapabilities"`
	AIProviderRegistryFields []string            `json:"aiProviderRegistryFields"`
	AIModelProfileFields     []string            `json:"aiModelProfileFields"`
	AIVocabularyFields       []string            `json:"aiVocabularyFields"`
	Serialization            struct {
		PageInfo map[string]any `json:"pageInfo"`
		APIError map[string]any `json:"apiError"`
	} `json:"serialization"`
}

func loadParityFixture(t *testing.T) parityFixture {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	path := filepath.Join(wd, "..", "..", "..", "..", "shared", "fixtures", "conventions-parity.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read conventions parity fixture: %v", err)
	}
	var fixture parityFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("parse conventions parity fixture: %v", err)
	}
	return fixture
}

func TestConventionsParityFixture_EnumSets(t *testing.T) {
	fixture := loadParityFixture(t)
	got := map[string][]string{
		"TargetJobStatus":      stringsOf(AllTargetJobStatuses),
		"TargetJobParseStatus": stringsOf(AllTargetJobParseStatuses),
		"PracticeMode":         stringsOf(AllPracticeModes),
		"PracticeGoal":         stringsOf(AllPracticeGoals),
		"InterviewerRole":      stringsOf(AllInterviewerRoles),
		"SessionStatus":        stringsOf(AllSessionStatuses),
		"ReportStatus":         stringsOf(AllReportStatuses),
		"ReadinessTier":        stringsOf(AllReadinessTiers),
		"DimensionStatus":      stringsOf(AllDimensionStatuses),
		"Confidence":           stringsOf(AllConfidences),
		"QuestionReviewStatus": stringsOf(AllQuestionReviewStatuses),
		"PrivacyRequestType":   stringsOf(AllPrivacyRequestTypes),
		"PrivacyRequestStatus": stringsOf(AllPrivacyRequestStatuses),
	}
	if len(got) != 13 {
		t.Fatalf("generated enum type count = %d, want 13", len(got))
	}
	if !reflect.DeepEqual(got, fixture.Enums) {
		t.Fatalf("Go enum sets differ from fixture\ngot:  %#v\nwant: %#v", got, fixture.Enums)
	}
}

func TestPracticeModeIsBinary(t *testing.T) {
	want := []string{"assisted", "strict"}
	if got := stringsOf(AllPracticeModes); !reflect.DeepEqual(got, want) {
		t.Fatalf("PracticeMode values = %#v, want %#v", got, want)
	}
}

func TestConventionsParityFixture_ErrorCodesAndAIVocabulary(t *testing.T) {
	fixture := loadParityFixture(t)
	if !reflect.DeepEqual(sharederrors.AllCodes, fixture.ErrorCodes) {
		t.Fatalf("Go error codes differ from fixture\ngot:  %#v\nwant: %#v", sharederrors.AllCodes, fixture.ErrorCodes)
	}
	gotCapabilities := stringsOf(sharedai.AllCapabilities)
	if !reflect.DeepEqual(gotCapabilities, fixture.AICapabilities) {
		t.Fatalf("Go AI capabilities differ from fixture\ngot:  %#v\nwant: %#v", gotCapabilities, fixture.AICapabilities)
	}
	gotRegistryFields := stringsOf(sharedai.AllProviderRegistryFieldNames)
	if !reflect.DeepEqual(gotRegistryFields, fixture.AIProviderRegistryFields) {
		t.Fatalf("Go AI provider registry fields differ from fixture\ngot:  %#v\nwant: %#v", gotRegistryFields, fixture.AIProviderRegistryFields)
	}
	gotProfileFields := stringsOf(sharedai.AllModelProfileFieldNames)
	if !reflect.DeepEqual(gotProfileFields, fixture.AIModelProfileFields) {
		t.Fatalf("Go AI model profile fields differ from fixture\ngot:  %#v\nwant: %#v", gotProfileFields, fixture.AIModelProfileFields)
	}
	gotAIFields := stringsOf(sharedai.AllFieldNames)
	if !reflect.DeepEqual(gotAIFields, fixture.AIVocabularyFields) {
		t.Fatalf("Go AI vocabulary fields differ from fixture\ngot:  %#v\nwant: %#v", gotAIFields, fixture.AIVocabularyFields)
	}
}

func TestPracticeNotFoundErrorCodesRegistered(t *testing.T) {
	want := []string{"PRACTICE_PLAN_NOT_FOUND", "PRACTICE_SESSION_NOT_FOUND"}
	for _, code := range want {
		if !contains(sharederrors.AllCodes, code) {
			t.Fatalf("shared error code %s is not registered in AllCodes: %#v", code, sharederrors.AllCodes)
		}
	}
}

func TestResumeVocabularyRegistered(t *testing.T) {
	if !contains(sharederrors.AllCodes, sharederrors.CodeResumeExportNotAvailable) {
		t.Fatalf("shared error code %s is not registered in AllCodes: %#v", sharederrors.CodeResumeExportNotAvailable, sharederrors.AllCodes)
	}
}

func TestConventionsParityFixture_SerializationShapes(t *testing.T) {
	fixture := loadParityFixture(t)
	if fixture.Serialization.PageInfo == nil {
		t.Fatal("fixture missing serialization.pageInfo")
	}
	if fixture.Serialization.APIError == nil {
		t.Fatal("fixture missing serialization.apiError")
	}

	nextCursor := "cursor_01"
	page := PageInfo{
		NextCursor: &nextCursor,
		PageSize:   20,
		HasMore:    true,
	}
	assertJSONShape(t, "PageInfo", page, fixture.Serialization.PageInfo)

	apiError := sharederrors.APIError{
		Code:      sharederrors.CodeValidationFailed,
		Message:   "request validation failed",
		RequestID: "req_01HV",
		Retryable: false,
		Details:   map[string]any{"field": "email"},
	}
	assertJSONShape(t, "APIError", apiError, fixture.Serialization.APIError)
}

func stringsOf[T ~string](values []T) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func contains[T comparable](values []T, want T) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func assertJSONShape(t *testing.T, label string, value any, want map[string]any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("%s marshal: %v", label, err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("%s unmarshal: %v", label, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s JSON shape differs\ngot:  %#v\nwant: %#v\nraw: %s", label, got, want, raw)
	}
}
