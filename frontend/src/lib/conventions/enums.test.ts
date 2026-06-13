import { describe, it, expect } from 'vitest';
import * as enums from './enums';
import {
  ALL_PRACTICE_MODES,
  ALL_SESSION_STATUSES,
  ALL_JOB_STATUSES,
  ALL_PRIVACY_REQUEST_TYPES,
  ALL_PRIVACY_REQUEST_STATUSES,
  type PracticeMode,
  type SessionStatus,
  type JobStatus,
} from './enums';

describe('enum union literals', () => {
  it('PracticeMode allows strict', () => {
    const mode: PracticeMode = 'strict';
    expect(ALL_PRACTICE_MODES).toContain(mode);
  });
  it('PracticeMode is the binary assisted/strict set', () => {
    expect(ALL_PRACTICE_MODES).toEqual(['assisted', 'strict']);
  });
  it('SessionStatus allows waiting_user_input', () => {
    const status: SessionStatus = 'waiting_user_input';
    expect(ALL_SESSION_STATUSES).toContain(status);
  });
  it('JobStatus allows dead', () => {
    const status: JobStatus = 'dead';
    expect(ALL_JOB_STATUSES).toContain(status);
  });
  it('PrivacyRequest splits into two enums (§5.13)', () => {
    expect(ALL_PRIVACY_REQUEST_TYPES).toEqual(['export', 'delete']);
    expect(ALL_PRIVACY_REQUEST_STATUSES).toContain('queued');
    expect(ALL_PRIVACY_REQUEST_STATUSES).toContain('cancelled');
  });
  it('D-20 retired the resume version-tree enums (flatten)', () => {
    // ResumeVersionType / ResumeSeedStrategy / ResumeTailorSuggestionStatus
    // were dropped when resumes flattened to a single asset (no version tree,
    // no structured master, accept-only rewrites). Guard zero residue.
    expect(enums).not.toHaveProperty('ALL_RESUME_VERSION_TYPES');
    expect(enums).not.toHaveProperty('ALL_RESUME_SEED_STRATEGIES');
    expect(enums).not.toHaveProperty('ALL_RESUME_TAILOR_SUGGESTION_STATUSES');
  });
});
