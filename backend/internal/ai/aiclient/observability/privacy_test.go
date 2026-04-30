package observability_test

import (
	"context"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
)

// sensitiveTokens are unique substrings the test plants into the prompt
// and response. The decorator MUST NOT let any of them surface in metric
// labels, log fields, ai_task_runs rows, or audit_events metadata.
var sensitiveTokens = []string{
	"SSN-123-45-6789",
	"OFFER-LETTER-ACME",
	"INTERVIEWER-NAME-Alice",
	"COMPANY-PROPRIETARY-XYZ",
}

func sensitivePayload() aiclient.CompletePayload {
	return aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "system", Content: "Background: candidate " + sensitiveTokens[0] + " applied to " + sensitiveTokens[1] + "."},
			{Role: "user", Content: "Interviewer notes: " + sensitiveTokens[2] + " regarding " + sensitiveTokens[3]},
		},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.followup",
			PromptVersion: "p1",
			RubricVersion: "r1",
			Language:      "en",
			TaskRun: aiclient.AITaskRunContext{
				TaskType:     aiclient.AITaskRunTaskFollowupGenerate,
				ResourceType: aiclient.AITaskRunResourceTargetJob,
				ResourceID:   "018f0d59-0f7a-7b58-9f2f-65cc4d8e8b1d",
			},
		},
	}
}

func TestPrivacy_NoPlaintextLeaksAnywhere(t *testing.T) {
	wrap, registry, logger, runs, audit := newTestStack(t)

	if _, _, err := wrap.Complete(context.Background(), "practice.followup.default", sensitivePayload()); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	for _, token := range sensitiveTokens {
		// Metric label values across every registered counter family.
		for _, family := range []string{
			observability.MetricRunsTotal,
			observability.MetricInputTokensTotal,
			observability.MetricOutputTokensTotal,
			observability.MetricCostUSDTotal,
			observability.MetricOutputValidationFailures,
			observability.MetricFallbackTotal,
		} {
			for _, labels := range registry.CounterLabelValues(family) {
				for _, lv := range labels {
					if strings.Contains(lv, token) {
						t.Errorf("plaintext token %q leaked into metric %q label %q", token, family, lv)
					}
				}
			}
		}

		// Log field values.
		for _, entry := range logger.Entries() {
			if anyContains(entry.Fields, token) {
				t.Errorf("plaintext token %q leaked into log fields: %+v", token, entry.Fields)
			}
		}

		// ai_task_runs rows.
		for _, row := range runs.Rows() {
			if anyTaskRunContains(row, token) {
				t.Errorf("plaintext token %q leaked into ai_task_runs row: %+v", token, row)
			}
		}

		// audit_events metadata.
		for _, row := range audit.Rows() {
			meta := row.Metadata
			if strings.Contains(meta.PromptHash, token) ||
				strings.Contains(meta.ResponseHash, token) ||
				strings.Contains(meta.ProfileName, token) {
				t.Errorf("plaintext token %q leaked into audit metadata: %+v", token, meta)
			}
		}
	}

	// Sanity check: hashes must be present and non-empty (the decorator
	// is not silently dropping the audit row).
	rows := audit.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 audit row, got %d", len(rows))
	}
	if rows[0].Metadata.PromptHash == "" || rows[0].Metadata.ResponseHash == "" {
		t.Fatalf("audit hashes missing: %+v", rows[0].Metadata)
	}
	if rows[0].Metadata.PromptCharLength == 0 || rows[0].Metadata.ResponseCharLength == 0 {
		t.Fatalf("audit char lengths zero: %+v", rows[0].Metadata)
	}
}

func anyContains(fields observability.LogFields, token string) bool {
	candidates := []string{
		fields.Provider, fields.ModelID, fields.ModelProfileName, fields.ModelProfileVersion,
		fields.PromptVersion, fields.RubricVersion, fields.TaskType, fields.Language,
		fields.Route, fields.ValidationStatus, fields.ErrorCode,
	}
	for _, c := range candidates {
		if strings.Contains(c, token) {
			return true
		}
	}
	for _, fb := range fields.FallbackChain {
		if strings.Contains(fb, token) {
			return true
		}
	}
	return false
}

func anyTaskRunContains(row aiclient.AITaskRunRow, token string) bool {
	candidates := []string{
		row.Provider, row.ModelFamily, row.ModelID, string(row.TaskType),
		row.PromptVersion, row.RubricVersion, row.ModelProfileName, row.ModelProfileVersion,
		row.Language, row.Route, string(row.ValidationStatus), row.ErrorCode,
		row.Metadata.PromptHash, row.Metadata.ResponseHash, row.Metadata.ProfileName,
	}
	for _, c := range candidates {
		if strings.Contains(c, token) {
			return true
		}
	}
	for _, fb := range row.FallbackChain {
		if strings.Contains(fb, token) {
			return true
		}
	}
	return false
}
