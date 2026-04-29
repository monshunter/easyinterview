package ai

import "testing"

func TestAIVocabularyFieldSet(t *testing.T) {
	want := []FieldName{
		FieldModelProfileName,
		FieldModelProfileVersion,
		FieldModelFamily,
		FieldModelId,
		FieldFallbackChain,
		FieldRoute,
		FieldValidationStatus,
		FieldOutputSchemaVersion,
		FieldPromptVersion,
		FieldRubricVersion,
		FieldLanguage,
		FieldFeatureFlag,
		FieldDataSourceVersion,
	}

	if len(AllFieldNames) != len(want) {
		t.Fatalf("AllFieldNames length = %d, want %d", len(AllFieldNames), len(want))
	}
	for i := range want {
		if AllFieldNames[i] != want[i] {
			t.Errorf("AllFieldNames[%d] = %q, want %q", i, AllFieldNames[i], want[i])
		}
	}
}

func TestAIVocabularyWireNames(t *testing.T) {
	cases := map[FieldName]string{
		FieldModelProfileName:    "model_profile_name",
		FieldModelProfileVersion: "model_profile_version",
		FieldModelFamily:         "model_family",
		FieldModelId:             "model_id",
		FieldFallbackChain:       "fallback_chain",
		FieldRoute:               "route",
		FieldValidationStatus:    "validation_status",
		FieldOutputSchemaVersion: "output_schema_version",
		FieldPromptVersion:       "prompt_version",
		FieldRubricVersion:       "rubric_version",
		FieldLanguage:            "language",
		FieldFeatureFlag:         "feature_flag",
		FieldDataSourceVersion:   "data_source_version",
	}
	for field, want := range cases {
		if string(field) != want {
			t.Errorf("%s = %q, want %q", field, string(field), want)
		}
	}
}

func TestAIVocabularyFieldValidation(t *testing.T) {
	if !IsFieldName("model_profile_name") {
		t.Fatal("IsFieldName(model_profile_name) = false, want true")
	}
	if IsFieldName("modelProfileName") {
		t.Fatal("IsFieldName(modelProfileName) = true, want false")
	}
}

func TestA3ConsumedAIVocabularyFields(t *testing.T) {
	cases := map[string]FieldName{
		"model_profile_name":    FieldModelProfileName,
		"model_profile_version": FieldModelProfileVersion,
		"model_family":          FieldModelFamily,
		"fallback_chain":        FieldFallbackChain,
		"route":                 FieldRoute,
		"validation_status":     FieldValidationStatus,
		"output_schema_version": FieldOutputSchemaVersion,
	}
	for wire, field := range cases {
		if string(field) != wire {
			t.Errorf("A3 field %s maps to %q", wire, field)
		}
		if !IsFieldName(wire) {
			t.Errorf("A3 field %s missing from IsFieldName", wire)
		}
	}
}
