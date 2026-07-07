import { describe, expect, it } from 'vitest';

import {
  API_FACING_JOB_TYPES,
  ASYNQ_TASK_EMAIL_DISPATCH,
  ASYNQ_TASK_PRIVACY_DELETE,
  ASYNQ_TASK_TARGET_IMPORT,
  isSourceEventOnly,
  JOB_TRIGGER_EVENT_SEMANTIC_SOURCE_EVENT_ONLY,
  JOB_TRIGGER_EVENT_SEMANTIC_TRIGGER_CREATES_JOB,
  JOB_TRIGGER_EVENT_SEMANTICS,
  JOB_TYPE_EMAIL_DISPATCH,
  JOB_TYPE_REPORT_GENERATE,
  JOB_TYPE_SOURCE_REFRESH,
  JOB_TYPE_TARGET_IMPORT,
  EMAIL_DISPATCH_REDACTED_FIELDS,
} from './jobs';

describe('generated job contract', () => {
  it('exports canonical and asynq job names', () => {
    expect(JOB_TYPE_TARGET_IMPORT).toBe('target_import');
    expect(JOB_TYPE_EMAIL_DISPATCH).toBe('email_dispatch');
    expect(ASYNQ_TASK_TARGET_IMPORT).toBe('target.import');
    expect(ASYNQ_TASK_PRIVACY_DELETE).toBe('privacy.delete');
    expect(ASYNQ_TASK_EMAIL_DISPATCH).toBe('email.dispatch');
  });

  it('exports trigger event semantics and source-event predicate', () => {
    expect(JOB_TRIGGER_EVENT_SEMANTIC_SOURCE_EVENT_ONLY).toBe('source_event_only');
    expect(JOB_TRIGGER_EVENT_SEMANTIC_TRIGGER_CREATES_JOB).toBe('trigger_creates_job');
    expect(JOB_TRIGGER_EVENT_SEMANTICS[JOB_TYPE_REPORT_GENERATE]).toBe(JOB_TRIGGER_EVENT_SEMANTIC_SOURCE_EVENT_ONLY);
    expect(isSourceEventOnly(JOB_TYPE_REPORT_GENERATE)).toBe(true);
    expect(isSourceEventOnly(JOB_TYPE_TARGET_IMPORT)).toBe(false);
  });

  it('keeps internal-only jobs out of the API-facing subset', () => {
    expect(API_FACING_JOB_TYPES).toContain('target_import');
    expect(API_FACING_JOB_TYPES).not.toContain(JOB_TYPE_SOURCE_REFRESH);
    expect(API_FACING_JOB_TYPES).not.toContain(JOB_TYPE_EMAIL_DISPATCH);
  });

  it('exports email_dispatch redaction policy', () => {
    expect(EMAIL_DISPATCH_REDACTED_FIELDS).toContain('rawEmailCode');
    expect(EMAIL_DISPATCH_REDACTED_FIELDS).toContain('emailBody');
  });
});
