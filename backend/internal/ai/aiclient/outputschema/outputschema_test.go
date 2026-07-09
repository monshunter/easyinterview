package outputschema_test

import (
	"encoding/json"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/outputschema"
)

func TestValidate(t *testing.T) {
	objectSchema := json.RawMessage(`{"type":"object","required":["a"],"properties":{"a":{"type":"string"},"b":{"enum":["x","y"]}}}`)

	cases := []struct {
		name    string
		schema  json.RawMessage
		content string
		wantErr bool
	}{
		{"valid object", objectSchema, `{"a":"hello"}`, false},
		{"valid with enum", objectSchema, `{"a":"hello","b":"x"}`, false},
		{"missing required", objectSchema, `{"b":"x"}`, true},
		{"enum mismatch", objectSchema, `{"a":"hello","b":"z"}`, true},
		{"wrong type", objectSchema, `{"a":123}`, true},
		{"empty content", objectSchema, ``, true},
		{"trailing tokens", objectSchema, `{"a":"hello"} {"a":"again"}`, true},
		{"markdown fenced object", objectSchema, "```json\n{\"a\":\"hello\"}\n```", false},
		{"markdown fenced object with prose rejected", objectSchema, "Here is the JSON:\n```json\n{\"a\":\"hello\"}\n```", true},
		{"markdown fenced object with trailing prose rejected", objectSchema, "```json\n{\"a\":\"hello\"}\n```\nDone.", true},
		{"array of objects", json.RawMessage(`{"type":"array","items":{"type":"object","required":["k"],"properties":{"k":{"type":"string"}}}}`), `[{"k":"v"}]`, false},
		{"array item invalid", json.RawMessage(`{"type":"array","items":{"type":"object","required":["k"]}}`), `[{"other":"v"}]`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := outputschema.Validate(tc.schema, tc.content)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
