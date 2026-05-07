package ai

import "testing"

func TestAIVocabularyFieldSet(t *testing.T) {
	want := []FieldName{
		FieldModelProfileName,
		FieldModelProfileVersion,
		FieldProvider,
		FieldCapability,
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
		FieldFromProvider,
		FieldFromModelFamily,
		FieldToProvider,
		FieldToModelFamily,
		FieldToolInvocations,
		FieldPartialMetaReason,
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
		FieldProvider:            "provider",
		FieldCapability:          "capability",
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
		FieldFromProvider:        "from_provider",
		FieldFromModelFamily:     "from_model_family",
		FieldToProvider:          "to_provider",
		FieldToModelFamily:       "to_model_family",
		FieldToolInvocations:     "tool_invocations",
		FieldPartialMetaReason:   "partial_meta_reason",
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
		"provider":              FieldProvider,
		"capability":            FieldCapability,
		"model_family":          FieldModelFamily,
		"fallback_chain":        FieldFallbackChain,
		"route":                 FieldRoute,
		"validation_status":     FieldValidationStatus,
		"output_schema_version": FieldOutputSchemaVersion,
		"tool_invocations":      FieldToolInvocations,
		"partial_meta_reason":   FieldPartialMetaReason,
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

func TestAICapabilityVocabulary(t *testing.T) {
	want := []Capability{
		CapabilityChat,
		CapabilityStt,
		CapabilityRealtime,
		CapabilityJudge,
	}
	if len(AllCapabilities) != len(want) {
		t.Fatalf("AllCapabilities length = %d, want %d", len(AllCapabilities), len(want))
	}
	for i := range want {
		if AllCapabilities[i] != want[i] {
			t.Errorf("AllCapabilities[%d] = %q, want %q", i, AllCapabilities[i], want[i])
		}
	}
	if !IsCapability("judge") {
		t.Fatal("IsCapability(judge) = false, want true")
	}
	if IsCapability("image") {
		t.Fatal("IsCapability(image) = true, want false")
	}
}

func TestAIProviderRegistryAndProfileFieldVocabulary(t *testing.T) {
	providerRegistryFields := map[string]ProviderRegistryFieldName{
		"name":         ProviderRegistryFieldNameName,
		"protocol":     ProviderRegistryFieldNameProtocol,
		"base_url_env": ProviderRegistryFieldNameBaseUrlEnv,
		"api_key_env":  ProviderRegistryFieldNameApiKeyEnv,
		"capabilities": ProviderRegistryFieldNameCapabilities,
		"version":      ProviderRegistryFieldNameVersion,
	}
	for wire, field := range providerRegistryFields {
		if string(field) != wire {
			t.Errorf("provider registry field %s maps to %q", wire, field)
		}
		if !IsProviderRegistryFieldName(wire) {
			t.Errorf("provider registry field %s missing from IsProviderRegistryFieldName", wire)
		}
	}

	modelProfileFields := map[string]ModelProfileFieldName{
		"name":               ModelProfileFieldNameName,
		"capability":         ModelProfileFieldNameCapability,
		"status":             ModelProfileFieldNameStatus,
		"unsupported_reason": ModelProfileFieldNameUnsupportedReason,
		"provider_ref":       ModelProfileFieldNameProviderRef,
		"fallback":           ModelProfileFieldNameFallback,
		"timeout_ms":         ModelProfileFieldNameTimeoutMs,
		"privacy_policy":     ModelProfileFieldNamePrivacyPolicy,
	}
	for wire, field := range modelProfileFields {
		if string(field) != wire {
			t.Errorf("model profile field %s maps to %q", wire, field)
		}
		if !IsModelProfileFieldName(wire) {
			t.Errorf("model profile field %s missing from IsModelProfileFieldName", wire)
		}
	}
}
