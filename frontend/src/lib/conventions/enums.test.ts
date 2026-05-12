import { describe, it, expect } from 'vitest';
import {
  ALL_PRACTICE_MODES,
  ALL_SESSION_STATUSES,
  ALL_JOB_STATUSES,
  ALL_PRIVACY_REQUEST_TYPES,
  ALL_PRIVACY_REQUEST_STATUSES,
  ALL_RESUME_VERSION_TYPES,
  ALL_RESUME_SEED_STRATEGIES,
  ALL_RESUME_TAILOR_SUGGESTION_STATUSES,
  type PracticeMode,
  type SessionStatus,
  type JobStatus,
  type ResumeVersionType,
  type ResumeSeedStrategy,
  type ResumeTailorSuggestionStatus,
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
  it('Resume Workshop additive enums are generated from B1', () => {
    const versionType: ResumeVersionType = 'structured_master';
    const seedStrategy: ResumeSeedStrategy = 'ai_select';
    const suggestionStatus: ResumeTailorSuggestionStatus = 'rejected';
    expect(ALL_RESUME_VERSION_TYPES).toEqual(['structured_master', 'targeted']);
    expect(ALL_RESUME_SEED_STRATEGIES).toEqual(['copy_master', 'blank', 'ai_select']);
    expect(ALL_RESUME_TAILOR_SUGGESTION_STATUSES).toEqual(['pending', 'accepted', 'rejected']);
    expect(ALL_RESUME_VERSION_TYPES).toContain(versionType);
    expect(ALL_RESUME_SEED_STRATEGIES).toContain(seedStrategy);
    expect(ALL_RESUME_TAILOR_SUGGESTION_STATUSES).toContain(suggestionStatus);
  });
});
