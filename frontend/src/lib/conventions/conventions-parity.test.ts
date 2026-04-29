import { describe, it, expect } from 'vitest';
import fixture from '../../../../shared/fixtures/conventions-parity.json';
import {
  ALL_TARGET_JOB_STATUSES,
  ALL_TARGET_JOB_PARSE_STATUSES,
  ALL_PRACTICE_MODES,
  ALL_PRACTICE_GOALS,
  ALL_INTERVIEWER_ROLES,
  ALL_SESSION_STATUSES,
  ALL_REPORT_STATUSES,
  ALL_READINESS_TIERS,
  ALL_DIMENSION_STATUSES,
  ALL_CONFIDENCES,
  ALL_MISTAKE_STATUSES,
  ALL_DEBRIEF_STATUSES,
  ALL_PRIVACY_REQUEST_TYPES,
  ALL_PRIVACY_REQUEST_STATUSES,
  ALL_ERROR_CODES,
  ERROR_CODES,
  ALL_AI_VOCABULARY_FIELDS,
  type ApiError,
  type PageInfo,
} from '.';

interface ParityFixture {
  enums: Record<string, readonly string[]>;
  errorCodes: readonly string[];
  aiVocabularyFields: readonly string[];
  serialization: {
    pageInfo: Record<string, unknown>;
    apiError: Record<string, unknown>;
  };
}

const parity = fixture as ParityFixture;

describe('cross-language conventions parity fixture', () => {
  it('matches all 14 generated enum literal sets', () => {
    const actual: Record<string, readonly string[]> = {
      TargetJobStatus: ALL_TARGET_JOB_STATUSES,
      TargetJobParseStatus: ALL_TARGET_JOB_PARSE_STATUSES,
      PracticeMode: ALL_PRACTICE_MODES,
      PracticeGoal: ALL_PRACTICE_GOALS,
      InterviewerRole: ALL_INTERVIEWER_ROLES,
      SessionStatus: ALL_SESSION_STATUSES,
      ReportStatus: ALL_REPORT_STATUSES,
      ReadinessTier: ALL_READINESS_TIERS,
      DimensionStatus: ALL_DIMENSION_STATUSES,
      Confidence: ALL_CONFIDENCES,
      MistakeStatus: ALL_MISTAKE_STATUSES,
      DebriefStatus: ALL_DEBRIEF_STATUSES,
      PrivacyRequestType: ALL_PRIVACY_REQUEST_TYPES,
      PrivacyRequestStatus: ALL_PRIVACY_REQUEST_STATUSES,
    };

    expect(Object.keys(actual)).toHaveLength(14);
    expect(actual).toEqual(parity.enums);
  });

  it('matches error codes including AI baseline codes', () => {
    expect(ALL_ERROR_CODES).toEqual(parity.errorCodes);
    expect(ERROR_CODES.AI_PROVIDER_TIMEOUT).toBe('AI_PROVIDER_TIMEOUT');
    expect(ERROR_CODES.AI_OUTPUT_INVALID).toBe('AI_OUTPUT_INVALID');
    expect(ERROR_CODES.AI_FALLBACK_EXHAUSTED).toBe('AI_FALLBACK_EXHAUSTED');
  });

  it('matches AI vocabulary fields', () => {
    expect(ALL_AI_VOCABULARY_FIELDS).toEqual(parity.aiVocabularyFields);
  });

  it('keeps PageInfo and ApiError canonical JSON shapes', () => {
    const pageInfo: PageInfo = {
      nextCursor: 'cursor_01',
      pageSize: 20,
      hasMore: true,
    };
    const apiError: ApiError = {
      code: ERROR_CODES.VALIDATION_FAILED,
      message: 'request validation failed',
      requestId: 'req_01HV',
      retryable: false,
      details: { field: 'email' },
    };

    expect(JSON.parse(JSON.stringify(pageInfo))).toEqual(parity.serialization.pageInfo);
    expect(JSON.parse(JSON.stringify(apiError))).toEqual(parity.serialization.apiError);
  });
});
