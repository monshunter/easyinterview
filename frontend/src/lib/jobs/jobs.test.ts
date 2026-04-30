import { describe, expect, it } from 'vitest';

import {
  API_FACING_JOB_TYPES,
  ASYNQ_TASK_EMAIL_DISPATCH,
  JOB_TYPE_EMAIL_DISPATCH,
  JOB_TYPE_TARGET_IMPORT,
  EMAIL_DISPATCH_REDACTED_FIELDS,
} from './jobs';

describe('generated job contract', () => {
  it('exports canonical and asynq job names', () => {
    expect(JOB_TYPE_TARGET_IMPORT).toBe('target_import');
    expect(JOB_TYPE_EMAIL_DISPATCH).toBe('email_dispatch');
    expect(ASYNQ_TASK_EMAIL_DISPATCH).toBe('email.dispatch');
  });

  it('keeps internal-only jobs out of the API-facing subset', () => {
    expect(API_FACING_JOB_TYPES).toContain('target_import');
    expect(API_FACING_JOB_TYPES).not.toContain('email_dispatch');
  });

  it('exports email_dispatch redaction policy', () => {
    expect(EMAIL_DISPATCH_REDACTED_FIELDS).toContain('rawMagicLinkToken');
    expect(EMAIL_DISPATCH_REDACTED_FIELDS).toContain('emailBody');
  });
});
