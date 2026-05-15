package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
)

// TaskRunWriter persists A3 ai_task_runs rows into the B4 baseline table.
type TaskRunWriter struct {
	db *sql.DB
}

func NewTaskRunWriter(db *sql.DB) *TaskRunWriter {
	return &TaskRunWriter{db: db}
}

func (w *TaskRunWriter) WriteAITaskRun(ctx context.Context, row aiclient.AITaskRunRow) error {
	if w == nil || w.db == nil {
		return fmt.Errorf("ai task run writer: db is required")
	}
	fallbackChain, err := json.Marshal(row.FallbackChain)
	if err != nil {
		return fmt.Errorf("marshal ai_task_runs fallback_chain: %w", err)
	}
	metadata, err := json.Marshal(row.Metadata)
	if err != nil {
		return fmt.Errorf("marshal ai_task_runs metadata: %w", err)
	}
	_, err = w.db.ExecContext(ctx, `
insert into ai_task_runs (
  id, user_id, task_type, resource_type, resource_id,
  provider, model_family, model_id, prompt_version, rubric_version,
  model_profile_name, model_profile_version, feature_key, feature_flag,
  data_source_version, language, input_tokens, output_tokens, latency_ms,
  cost_usd_micros, status, route, validation_status, output_schema_version,
  error_code, fallback_chain, raw_response_object_key, metadata, started_at,
  completed_at
) values (
  $1, $2, $3, $4, $5,
  $6, $7, $8, $9, $10,
  $11, $12, $13, $14,
  $15, $16, $17, $18, $19,
  $20, $21, $22, $23, $24,
  $25, $26::jsonb, $27, $28::jsonb, $29,
  $30
)`,
		row.ID, nullableString(row.UserID), string(row.Capability), string(row.ResourceType), row.ResourceID,
		row.Provider, nullableString(row.ModelFamily), row.ModelID, nullableString(row.PromptVersion), nullableString(row.RubricVersion),
		nullableString(row.ModelProfileName), nullableString(row.ModelProfileVersion), row.FeatureKey, row.FeatureFlag,
		row.DataSourceVersion, row.Language, row.InputTokens, row.OutputTokens, row.LatencyMs,
		row.CostUSDMicros, string(row.Status), nullableString(row.Route), nullableString(string(row.ValidationStatus)), nullableString(row.OutputSchemaVersion),
		nullableString(row.ErrorCode), string(fallbackChain), nullableString(row.RawResponseObjectKey), string(metadata), row.StartedAt,
		nullableTime(row.CompletedAt),
	)
	if err != nil {
		return fmt.Errorf("insert ai_task_runs: %w", err)
	}
	return nil
}

func nullableString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func nullableTime(value interface{ IsZero() bool }) any {
	if value.IsZero() {
		return nil
	}
	return value
}
