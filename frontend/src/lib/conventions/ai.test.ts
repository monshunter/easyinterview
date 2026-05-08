import { describe, it, expect } from 'vitest';
import {
  AI_CAPABILITIES,
  ALL_AI_CAPABILITIES,
  isAICapability,
  AI_PROVIDER_REGISTRY_FIELDS,
  ALL_AI_PROVIDER_REGISTRY_FIELDS,
  isAIProviderRegistryField,
  AI_MODEL_PROFILE_FIELDS,
  ALL_AI_MODEL_PROFILE_FIELDS,
  isAIModelProfileField,
  AI_VOCABULARY_FIELDS,
  ALL_AI_VOCABULARY_FIELDS,
  isAIVocabularyField,
  type AIVocabularyField,
} from '.';

describe('AI vocabulary generated constants', () => {
  it('keeps the full wire field set in declaration order', () => {
    const expected: readonly AIVocabularyField[] = [
      AI_VOCABULARY_FIELDS.MODEL_PROFILE_NAME,
      AI_VOCABULARY_FIELDS.MODEL_PROFILE_VERSION,
      AI_VOCABULARY_FIELDS.PROVIDER,
      AI_VOCABULARY_FIELDS.CAPABILITY,
      AI_VOCABULARY_FIELDS.MODEL_FAMILY,
      AI_VOCABULARY_FIELDS.MODEL_ID,
      AI_VOCABULARY_FIELDS.FALLBACK_CHAIN,
      AI_VOCABULARY_FIELDS.ROUTE,
      AI_VOCABULARY_FIELDS.VALIDATION_STATUS,
      AI_VOCABULARY_FIELDS.OUTPUT_SCHEMA_VERSION,
      AI_VOCABULARY_FIELDS.PROMPT_VERSION,
      AI_VOCABULARY_FIELDS.RUBRIC_VERSION,
      AI_VOCABULARY_FIELDS.LANGUAGE,
      AI_VOCABULARY_FIELDS.FEATURE_FLAG,
      AI_VOCABULARY_FIELDS.DATA_SOURCE_VERSION,
      AI_VOCABULARY_FIELDS.FROM_PROVIDER,
      AI_VOCABULARY_FIELDS.FROM_MODEL_FAMILY,
      AI_VOCABULARY_FIELDS.TO_PROVIDER,
      AI_VOCABULARY_FIELDS.TO_MODEL_FAMILY,
      AI_VOCABULARY_FIELDS.TOOL_INVOCATIONS,
      AI_VOCABULARY_FIELDS.PARTIAL_META_REASON,
    ];

    expect(ALL_AI_VOCABULARY_FIELDS).toEqual(expected);
  });

  it('maps TS constant names to canonical snake_case wire names', () => {
    expect(AI_VOCABULARY_FIELDS.MODEL_PROFILE_NAME).toBe('model_profile_name');
    expect(AI_VOCABULARY_FIELDS.MODEL_PROFILE_VERSION).toBe('model_profile_version');
    expect(AI_VOCABULARY_FIELDS.PROVIDER).toBe('provider');
    expect(AI_VOCABULARY_FIELDS.CAPABILITY).toBe('capability');
    expect(AI_VOCABULARY_FIELDS.MODEL_FAMILY).toBe('model_family');
    expect(AI_VOCABULARY_FIELDS.MODEL_ID).toBe('model_id');
    expect(AI_VOCABULARY_FIELDS.FALLBACK_CHAIN).toBe('fallback_chain');
    expect(AI_VOCABULARY_FIELDS.ROUTE).toBe('route');
    expect(AI_VOCABULARY_FIELDS.VALIDATION_STATUS).toBe('validation_status');
    expect(AI_VOCABULARY_FIELDS.OUTPUT_SCHEMA_VERSION).toBe('output_schema_version');
    expect(AI_VOCABULARY_FIELDS.PROMPT_VERSION).toBe('prompt_version');
    expect(AI_VOCABULARY_FIELDS.RUBRIC_VERSION).toBe('rubric_version');
    expect(AI_VOCABULARY_FIELDS.LANGUAGE).toBe('language');
    expect(AI_VOCABULARY_FIELDS.FEATURE_FLAG).toBe('feature_flag');
    expect(AI_VOCABULARY_FIELDS.DATA_SOURCE_VERSION).toBe('data_source_version');
    expect(AI_VOCABULARY_FIELDS.FROM_PROVIDER).toBe('from_provider');
    expect(AI_VOCABULARY_FIELDS.FROM_MODEL_FAMILY).toBe('from_model_family');
    expect(AI_VOCABULARY_FIELDS.TO_PROVIDER).toBe('to_provider');
    expect(AI_VOCABULARY_FIELDS.TO_MODEL_FAMILY).toBe('to_model_family');
    expect(AI_VOCABULARY_FIELDS.TOOL_INVOCATIONS).toBe('tool_invocations');
    expect(AI_VOCABULARY_FIELDS.PARTIAL_META_REASON).toBe('partial_meta_reason');
  });

  it('validates documented fields only', () => {
    expect(isAIVocabularyField('model_profile_name')).toBe(true);
    expect(isAIVocabularyField('modelProfileName')).toBe(false);
  });

  it('covers fields currently consumed by A3 AICallMeta', () => {
    const a3Fields = [
      AI_VOCABULARY_FIELDS.MODEL_PROFILE_NAME,
      AI_VOCABULARY_FIELDS.MODEL_PROFILE_VERSION,
      AI_VOCABULARY_FIELDS.PROVIDER,
      AI_VOCABULARY_FIELDS.CAPABILITY,
      AI_VOCABULARY_FIELDS.MODEL_FAMILY,
      AI_VOCABULARY_FIELDS.FALLBACK_CHAIN,
      AI_VOCABULARY_FIELDS.ROUTE,
      AI_VOCABULARY_FIELDS.VALIDATION_STATUS,
      AI_VOCABULARY_FIELDS.OUTPUT_SCHEMA_VERSION,
      AI_VOCABULARY_FIELDS.TOOL_INVOCATIONS,
      AI_VOCABULARY_FIELDS.PARTIAL_META_REASON,
    ] as const;

    expect(a3Fields).toEqual([
      'model_profile_name',
      'model_profile_version',
      'provider',
      'capability',
      'model_family',
      'fallback_chain',
      'route',
      'validation_status',
      'output_schema_version',
      'tool_invocations',
      'partial_meta_reason',
    ]);
    for (const field of a3Fields) {
      expect(isAIVocabularyField(field)).toBe(true);
    }
  });

  it('documents capability literals and provider/profile schema fields', () => {
    expect(ALL_AI_CAPABILITIES).toEqual(['chat', 'stt', 'tts', 'realtime', 'judge']);
    expect(AI_CAPABILITIES.JUDGE).toBe('judge');
    expect(isAICapability('judge')).toBe(true);
    expect(isAICapability('image')).toBe(false);

    expect(ALL_AI_PROVIDER_REGISTRY_FIELDS).toEqual([
      'name',
      'protocol',
      'base_url_env',
      'api_key_env',
      'capabilities',
      'version',
    ]);
    expect(AI_PROVIDER_REGISTRY_FIELDS.API_KEY_ENV).toBe('api_key_env');
    expect(isAIProviderRegistryField('base_url_env')).toBe(true);

    expect(ALL_AI_MODEL_PROFILE_FIELDS).toContain('provider_ref');
    expect(AI_MODEL_PROFILE_FIELDS.UNSUPPORTED_REASON).toBe('unsupported_reason');
    expect(isAIModelProfileField('privacy_policy')).toBe(true);
  });
});
