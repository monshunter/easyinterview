// Package outputschema is the single A3-owned implementation of the
// provider-neutral output-schema validation subset (spec §4.1: type / required
// / properties / items / enum, with description as a non-validating
// annotation). It is extracted so the observability decorator and the F3
// LLM judge fail-close path validate evaluated model output against the exact
// same subset semantics, instead of maintaining a second non-equivalent
// validator (plan prompt-rubric-registry/004 §2.5).
package outputschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Schema is the parsed output-schema subset. Only the keys A3 validates are
// modeled; unknown keys are ignored on decode.
type Schema struct {
	Type       string            `json:"type"`
	Required   []string          `json:"required"`
	Properties map[string]Schema `json:"properties"`
	Items      *Schema           `json:"items"`
	Enum       []any             `json:"enum"`
}

// Validate decodes schemaRaw and content and verifies content satisfies the
// schema subset. Content must be exactly one JSON value (trailing tokens are
// rejected). An empty content string is invalid.
func Validate(schemaRaw json.RawMessage, content string) error {
	if content == "" {
		return errors.New("empty content")
	}
	var schema Schema
	if err := json.Unmarshal(schemaRaw, &schema); err != nil {
		return fmt.Errorf("parse output_schema: %w", err)
	}
	var v any
	dec := json.NewDecoder(strings.NewReader(NormalizeJSONContent(content)))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err != nil {
			return err
		}
		return errors.New("multiple JSON values in output")
	}
	return validateAgainstSchema(schema, v, "$")
}

// NormalizeJSONContent accepts the one provider deviation this app can safely
// recover from: the whole JSON value wrapped in a markdown code fence. It does
// not strip prose or multiple JSON values; those remain validation failures.
func NormalizeJSONContent(content string) string {
	content = strings.TrimSpace(strings.TrimPrefix(content, "\ufeff"))
	if !strings.HasPrefix(content, "```") {
		return content
	}
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || !isJSONFenceOpening(lines[0]) || strings.TrimSpace(lines[len(lines)-1]) != "```" {
		return content
	}
	return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
}

func isJSONFenceOpening(line string) bool {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "```") {
		return false
	}
	info := strings.TrimSpace(strings.TrimPrefix(line, "```"))
	return info == "" || strings.EqualFold(info, "json")
}

func validateAgainstSchema(schema Schema, value any, path string) error {
	if schema.Type != "" && !matchesSchemaType(schema.Type, value) {
		return fmt.Errorf("%s expected %s", path, schema.Type)
	}
	if len(schema.Enum) > 0 && !valueInEnum(value, schema.Enum) {
		return fmt.Errorf("%s value is not in enum", path)
	}

	if len(schema.Required) > 0 || len(schema.Properties) > 0 {
		obj, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%s expected object", path)
		}
		for _, key := range schema.Required {
			if _, ok := obj[key]; !ok {
				return fmt.Errorf("%s missing required field %q", path, key)
			}
		}
		for key, childSchema := range schema.Properties {
			child, ok := obj[key]
			if !ok {
				continue
			}
			if err := validateAgainstSchema(childSchema, child, path+"."+key); err != nil {
				return err
			}
		}
	}

	if schema.Items != nil {
		items, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s expected array", path)
		}
		for i, item := range items {
			if err := validateAgainstSchema(*schema.Items, item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func valueInEnum(value any, enum []any) bool {
	for _, candidate := range enum {
		if jsonValuesEqual(value, candidate) {
			return true
		}
	}
	return false
}

func jsonValuesEqual(a, b any) bool {
	if an, ok := a.(json.Number); ok {
		return jsonNumberEqual(an, b)
	}
	if bn, ok := b.(json.Number); ok {
		return jsonNumberEqual(bn, a)
	}
	return reflect.DeepEqual(a, b)
}

func jsonNumberEqual(n json.Number, other any) bool {
	switch v := other.(type) {
	case json.Number:
		return n.String() == v.String()
	case float64:
		nf, err := n.Float64()
		return err == nil && nf == v
	default:
		return false
	}
}

func matchesSchemaType(schemaType string, value any) bool {
	switch schemaType {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(json.Number)
		return ok
	case "integer":
		n, ok := value.(json.Number)
		if !ok {
			return false
		}
		_, err := n.Int64()
		return err == nil
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}
