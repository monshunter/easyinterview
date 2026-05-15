package ai

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
)

func TestTaskRunWriterInsertsTypedColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 5, 14, 9, 30, 0, 0, time.UTC)
	row := aiclient.AITaskRunRow{
		ID:                   "01918fa0-9000-7000-8000-000000000001",
		UserID:               "01918fa0-9000-7000-8000-000000000002",
		Capability:           aiclient.AITaskRunTaskHintGenerate,
		ResourceType:         aiclient.AITaskRunResourceTargetJob,
		ResourceID:           "01918fa0-9000-7000-8000-000000000003",
		Provider:             "stub",
		ModelFamily:          "stub",
		ModelID:              "stub-chat-1",
		PromptVersion:        "hint.prompt.v1",
		RubricVersion:        "not_applicable",
		ModelProfileName:     "practice.turn_observe.default",
		ModelProfileVersion:  "1.0.0",
		FeatureKey:           "practice.turn.lightweight_observe",
		FeatureFlag:          "none",
		DataSourceVersion:    "registry.v1",
		Language:             "en",
		InputTokens:          7,
		OutputTokens:         3,
		LatencyMs:            25,
		CostUSDMicros:        12,
		Status:               aiclient.AITaskRunStatusSuccess,
		Route:                "practice.turn.lightweight_observe",
		ValidationStatus:     aiclient.ValidationStatusOK,
		OutputSchemaVersion:  "hint.v1",
		FallbackChain:        []string{"stub/stub-chat-1"},
		RawResponseObjectKey: "object-key",
		Metadata:             aiclient.AuditMetadata{PromptHash: "prompt-hash", ResponseHash: "response-hash", PromptCharLength: 11, ResponseCharLength: 5, ProfileName: "practice.turn_observe.default"},
		StartedAt:            now,
		CompletedAt:          now.Add(25 * time.Millisecond),
	}

	mock.ExpectExec(`insert into ai_task_runs`).
		WithArgs(
			row.ID, row.UserID, string(row.Capability), string(row.ResourceType), row.ResourceID,
			row.Provider, row.ModelFamily, row.ModelID, row.PromptVersion, row.RubricVersion,
			row.ModelProfileName, row.ModelProfileVersion, row.FeatureKey, row.FeatureFlag,
			row.DataSourceVersion, row.Language, row.InputTokens, row.OutputTokens, row.LatencyMs,
			row.CostUSDMicros, string(row.Status), row.Route, string(row.ValidationStatus), row.OutputSchemaVersion,
			nil, `["stub/stub-chat-1"]`, row.RawResponseObjectKey, sqlmock.AnyArg(), row.StartedAt,
			row.CompletedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := NewTaskRunWriter(db).WriteAITaskRun(context.Background(), row); err != nil {
		t.Fatalf("WriteAITaskRun returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
