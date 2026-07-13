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
	"regexp"
	"strings"
	"unicode/utf8"
)

// Schema is the parsed provider-neutral output-schema subset shared by A3 and
// F3. Pointer fields distinguish an omitted constraint from an explicit zero
// or false value.
type Schema struct {
	Type                 string            `json:"type"`
	Description          string            `json:"description"`
	Required             []string          `json:"required"`
	Properties           map[string]Schema `json:"properties"`
	AdditionalProperties *bool             `json:"additionalProperties"`
	Items                *Schema           `json:"items"`
	Enum                 []any             `json:"enum"`
	Minimum              *float64          `json:"minimum"`
	Maximum              *float64          `json:"maximum"`
	MinLength            *int              `json:"minLength"`
	MaxLength            *int              `json:"maxLength"`
	Pattern              string            `json:"pattern"`
	MinItems             *int              `json:"minItems"`
	MaxItems             *int              `json:"maxItems"`
	UniqueItems          *bool             `json:"uniqueItems"`
}

// Validate decodes schemaRaw and content and verifies content satisfies the
// schema subset. Content must be exactly one JSON value (trailing tokens are
// rejected). An empty content string is invalid.
func Validate(schemaRaw json.RawMessage, content string) error {
	if content == "" {
		return errors.New("empty content")
	}
	schema, err := parseSchema(schemaRaw)
	if err != nil {
		return err
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

// ValidateSchema verifies that schemaRaw uses only the supported, internally
// consistent provider-neutral subset.
func ValidateSchema(schemaRaw json.RawMessage) error {
	_, err := parseSchema(schemaRaw)
	return err
}

// RequireClosedObjects additionally requires every declared object to set
// additionalProperties=false. Grounded report v0.2 uses this before a loader
// snapshot can publish.
func RequireClosedObjects(schemaRaw json.RawMessage) error {
	schema, err := parseSchema(schemaRaw)
	if err != nil {
		return err
	}
	return requireClosedObjects(schema, "$")
}

func parseSchema(schemaRaw json.RawMessage) (Schema, error) {
	var schema Schema
	dec := json.NewDecoder(strings.NewReader(string(schemaRaw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&schema); err != nil {
		return Schema{}, fmt.Errorf("parse output_schema: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Schema{}, errors.New("parse output_schema: trailing content")
	}
	if err := validateSchemaNode(schema, "$"); err != nil {
		return Schema{}, err
	}
	return schema, nil
}

func validateSchemaNode(schema Schema, path string) error {
	if schema.Type != "" {
		switch schema.Type {
		case "object", "array", "string", "number", "integer", "boolean", "null":
		default:
			return fmt.Errorf("output_schema %s has unsupported type %q", path, schema.Type)
		}
	}
	if schema.AdditionalProperties != nil && schema.Type != "object" {
		return fmt.Errorf("output_schema %s additionalProperties requires object", path)
	}
	for _, bound := range []*int{schema.MinLength, schema.MaxLength, schema.MinItems, schema.MaxItems} {
		if bound != nil && *bound < 0 {
			return fmt.Errorf("output_schema %s bounds must be non-negative", path)
		}
	}
	if schema.MinLength != nil || schema.MaxLength != nil || schema.Pattern != "" {
		if schema.Type != "string" {
			return fmt.Errorf("output_schema %s string bounds require string type", path)
		}
		if schema.Pattern != "" {
			if _, err := regexp.Compile(schema.Pattern); err != nil {
				return fmt.Errorf("output_schema %s invalid pattern: %w", path, err)
			}
		}
	}
	if schema.MinItems != nil || schema.MaxItems != nil || schema.UniqueItems != nil {
		if schema.Type != "array" {
			return fmt.Errorf("output_schema %s array bounds require array type", path)
		}
	}
	if schema.Minimum != nil || schema.Maximum != nil {
		if schema.Type != "number" && schema.Type != "integer" {
			return fmt.Errorf("output_schema %s numeric bounds require number/integer type", path)
		}
	}
	if schema.MinLength != nil && schema.MaxLength != nil && *schema.MinLength > *schema.MaxLength {
		return fmt.Errorf("output_schema %s minLength exceeds maxLength", path)
	}
	if schema.MinItems != nil && schema.MaxItems != nil && *schema.MinItems > *schema.MaxItems {
		return fmt.Errorf("output_schema %s minItems exceeds maxItems", path)
	}
	if schema.Minimum != nil && schema.Maximum != nil && *schema.Minimum > *schema.Maximum {
		return fmt.Errorf("output_schema %s minimum exceeds maximum", path)
	}
	for key, child := range schema.Properties {
		if err := validateSchemaNode(child, path+"."+key); err != nil {
			return err
		}
	}
	if schema.Items != nil {
		if err := validateSchemaNode(*schema.Items, path+"[]"); err != nil {
			return err
		}
	}
	return nil
}

func requireClosedObjects(schema Schema, path string) error {
	if schema.Type == "object" && (schema.AdditionalProperties == nil || *schema.AdditionalProperties) {
		return fmt.Errorf("output_schema %s must set additionalProperties=false", path)
	}
	for key, child := range schema.Properties {
		if err := requireClosedObjects(child, path+"."+key); err != nil {
			return err
		}
	}
	if schema.Items != nil {
		return requireClosedObjects(*schema.Items, path+"[]")
	}
	return nil
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

	if schema.Type == "object" || len(schema.Required) > 0 || len(schema.Properties) > 0 || schema.AdditionalProperties != nil {
		obj, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%s expected object", path)
		}
		for _, key := range schema.Required {
			if _, ok := obj[key]; !ok {
				return fmt.Errorf("%s missing required field %q", path, key)
			}
		}
		if schema.AdditionalProperties != nil && !*schema.AdditionalProperties {
			for key := range obj {
				if _, ok := schema.Properties[key]; !ok {
					return fmt.Errorf("%s unknown field %q", path, key)
				}
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

	if schema.Type == "array" || schema.Items != nil || schema.MinItems != nil || schema.MaxItems != nil || schema.UniqueItems != nil {
		items, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s expected array", path)
		}
		if schema.MinItems != nil && len(items) < *schema.MinItems {
			return fmt.Errorf("%s has %d items, minimum is %d", path, len(items), *schema.MinItems)
		}
		if schema.MaxItems != nil && len(items) > *schema.MaxItems {
			return fmt.Errorf("%s has %d items, maximum is %d", path, len(items), *schema.MaxItems)
		}
		if schema.UniqueItems != nil && *schema.UniqueItems {
			for i := range items {
				for j := 0; j < i; j++ {
					if jsonValuesEqual(items[i], items[j]) {
						return fmt.Errorf("%s items %d and %d are duplicates", path, j, i)
					}
				}
			}
		}
		if schema.Items != nil {
			for i, item := range items {
				if err := validateAgainstSchema(*schema.Items, item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	}

	if text, ok := value.(string); ok {
		length := utf8.RuneCountInString(text)
		if schema.MinLength != nil && length < *schema.MinLength {
			return fmt.Errorf("%s length %d is below %d", path, length, *schema.MinLength)
		}
		if schema.MaxLength != nil && length > *schema.MaxLength {
			return fmt.Errorf("%s length %d exceeds %d", path, length, *schema.MaxLength)
		}
		if schema.Pattern != "" {
			matched, err := regexp.MatchString(schema.Pattern, text)
			if err != nil || !matched {
				return fmt.Errorf("%s does not match pattern %q", path, schema.Pattern)
			}
		}
	}
	if number, ok := value.(json.Number); ok {
		numeric, err := number.Float64()
		if err != nil {
			return fmt.Errorf("%s invalid number", path)
		}
		if schema.Minimum != nil && numeric < *schema.Minimum {
			return fmt.Errorf("%s value %v is below %v", path, numeric, *schema.Minimum)
		}
		if schema.Maximum != nil && numeric > *schema.Maximum {
			return fmt.Errorf("%s value %v exceeds %v", path, numeric, *schema.Maximum)
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
