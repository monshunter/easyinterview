package targetjob

import (
	"encoding/json"
	"fmt"
	"strings"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/events"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

// ForbiddenOutboxFields enumerates substrings that must never appear in any
// outbox event payload, async_jobs payload, or audit metadata produced by
// this domain. The list is enforced at construction time by the Build*Payload
// helpers below; spec §4.3 / D-8 plus plan 5.1 codify the privacy redline.
//
// The match is case-insensitive substring against the marshalled JSON, so
// any caller-supplied map / struct tagged with one of these names is
// rejected before it can leak to outbox / log / metric label / audit.
var ForbiddenOutboxFields = []string{
	"raw_jd_text",
	"rawjdtext",
	"source_url",
	"sourceurl",
	"file_object_url",
	"fileobjecturl",
	"prompt_body",
	"promptbody",
	"response_body",
	"responsebody",
	"provider_secret",
	"providersecret",
	"authorization",
}

// TargetImportRequestedInput captures the three fields B3 allows for the
// events.EventNameTargetImportRequested outbox event. The struct is intentionally
// closed: any additional metadata the caller wants to record must be
// added via B3 spec revision first, never piggy-backed here.
type TargetImportRequestedInput struct {
	TargetJobID    string
	TargetLanguage string
	UserID         string
}

// BuildTargetImportRequestedPayload validates input shape and produces a
// B3-typed outbox event payload.
func BuildTargetImportRequestedPayload(in TargetImportRequestedInput) (events.TargetImportRequestedPayload, error) {
	if in.TargetJobID == "" {
		return events.TargetImportRequestedPayload{}, fmt.Errorf("%s: targetJobId is required", events.EventNameTargetImportRequested)
	}
	if in.UserID == "" {
		return events.TargetImportRequestedPayload{}, fmt.Errorf("%s: userId is required", events.EventNameTargetImportRequested)
	}
	if in.TargetLanguage == "" {
		return events.TargetImportRequestedPayload{}, fmt.Errorf("%s: targetLanguage is required", events.EventNameTargetImportRequested)
	}
	out := events.TargetImportRequestedPayload{
		TargetJobID:    in.TargetJobID,
		TargetLanguage: in.TargetLanguage,
		UserID:         in.UserID,
	}
	if err := assertNoForbiddenOutboxFields(out); err != nil {
		return events.TargetImportRequestedPayload{}, err
	}
	return out, nil
}

// TargetParsedInput captures the structured-only fields B3 allows for the
// events.EventNameTargetParsed event. Raw AI output, prompts, and response bodies are
// not part of this struct (and would be rejected by assertNoForbidden).
type TargetParsedInput struct {
	TargetJobID      string
	UserID           string
	AnalysisStatus   sharedtypes.TargetJobParseStatus
	RequirementCount int
	CoreThemes       []string
}

// BuildTargetParsedPayload validates input and produces a B3-typed payload
// for the events.EventNameTargetParsed outbox event.
func BuildTargetParsedPayload(in TargetParsedInput) (events.TargetParsedPayload, error) {
	if in.TargetJobID == "" {
		return events.TargetParsedPayload{}, fmt.Errorf("%s: targetJobId is required", events.EventNameTargetParsed)
	}
	if in.UserID == "" {
		return events.TargetParsedPayload{}, fmt.Errorf("%s: userId is required", events.EventNameTargetParsed)
	}
	if in.AnalysisStatus == "" {
		return events.TargetParsedPayload{}, fmt.Errorf("%s: analysisStatus is required", events.EventNameTargetParsed)
	}
	if in.RequirementCount < 0 {
		return events.TargetParsedPayload{}, fmt.Errorf("%s: requirementCount cannot be negative", events.EventNameTargetParsed)
	}
	out := events.TargetParsedPayload{
		AnalysisStatus:   in.AnalysisStatus,
		CoreThemes:       append([]string{}, in.CoreThemes...),
		RequirementCount: in.RequirementCount,
		TargetJobID:      in.TargetJobID,
		UserID:           in.UserID,
	}
	if err := assertNoForbiddenOutboxFields(out); err != nil {
		return events.TargetParsedPayload{}, err
	}
	return out, nil
}

// TargetAnalysisFailedInput captures the three structured fields B3 allows
// for events.EventNameTargetAnalysisFailed. Error envelopes / messages must reach this
// helper as a B1 error code only — never as raw provider error strings.
type TargetAnalysisFailedInput struct {
	TargetJobID string
	ErrorCode   string
	Retryable   bool
}

// BuildTargetAnalysisFailedPayload validates input and produces a B3-typed
// payload for the events.EventNameTargetAnalysisFailed outbox event.
func BuildTargetAnalysisFailedPayload(in TargetAnalysisFailedInput) (events.TargetAnalysisFailedPayload, error) {
	if in.TargetJobID == "" {
		return events.TargetAnalysisFailedPayload{}, fmt.Errorf("%s: targetJobId is required", events.EventNameTargetAnalysisFailed)
	}
	if in.ErrorCode == "" {
		return events.TargetAnalysisFailedPayload{}, fmt.Errorf("%s: errorCode is required", events.EventNameTargetAnalysisFailed)
	}
	if _, ok := sharederrors.CodeRegistry[in.ErrorCode]; !ok {
		return events.TargetAnalysisFailedPayload{}, fmt.Errorf("%s: errorCode must be a documented B1 code", events.EventNameTargetAnalysisFailed)
	}
	out := events.TargetAnalysisFailedPayload{
		ErrorCode:   in.ErrorCode,
		Retryable:   in.Retryable,
		TargetJobID: in.TargetJobID,
	}
	if err := assertNoForbiddenAnalysisFailedPayload(out); err != nil {
		return events.TargetAnalysisFailedPayload{}, err
	}
	return out, nil
}

func assertNoForbiddenAnalysisFailedPayload(out events.TargetAnalysisFailedPayload) error {
	redline := events.TargetAnalysisFailedPayload{
		ErrorCode:   "B1_DOCUMENTED_ERROR_CODE",
		Retryable:   out.Retryable,
		TargetJobID: out.TargetJobID,
	}
	return assertNoForbiddenOutboxFields(redline)
}

// TargetImportJobPayload is the JSON contract for the jobs.JobTypeTargetImport async
// job row. It mirrors the same redline as the outbox event: only structured
// references, never raw JD text or full URLs.
type TargetImportJobPayload struct {
	TargetJobID    string `json:"targetJobId"`
	UserID         string `json:"userId"`
	TargetLanguage string `json:"targetLanguage"`
}

// BuildTargetImportJobPayload validates a jobs.JobTypeTargetImport async_jobs payload
// and returns it as a JSON-encoded byte slice ready to write to
// `async_jobs.payload`. The same forbidden-token negative scan runs over
// the marshalled bytes.
func BuildTargetImportJobPayload(in TargetImportJobPayload) ([]byte, error) {
	if in.TargetJobID == "" || in.UserID == "" || in.TargetLanguage == "" {
		return nil, fmt.Errorf("%s payload requires targetJobId, userId, targetLanguage", jobs.JobTypeTargetImport)
	}
	if err := assertNoForbiddenOutboxFields(in); err != nil {
		return nil, err
	}
	return json.Marshal(in)
}

// assertNoForbiddenOutboxFields marshals the value to JSON and rejects it if
// any forbidden token (case-insensitive substring) appears in either the
// field names or values. This catches both struct-shape leaks and
// caller-supplied free-text injection of forbidden values.
func assertNoForbiddenOutboxFields(v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal payload for redline check: %w", err)
	}
	lower := strings.ToLower(string(raw))
	for _, f := range ForbiddenOutboxFields {
		if strings.Contains(lower, f) {
			return fmt.Errorf("payload contains forbidden token %q", f)
		}
	}
	return nil
}
