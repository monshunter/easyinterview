import { describe, expect, it } from 'vitest';

import {
  EVENT_NAME_REPORT_GENERATED,
  EVENT_NAME_RESUME_TAILOR_COMPLETED,
  EVENT_NAME_TARGET_IMPORT_REQUESTED,
  type ReportGeneratedPayload,
  type ResumeTailorCompletedPayload,
  type TargetImportRequestedPayload,
} from './events';

describe('generated event contract', () => {
  it('exports internal event name constants', () => {
    expect(EVENT_NAME_TARGET_IMPORT_REQUESTED).toBe('target.import.requested');
    expect(EVENT_NAME_REPORT_GENERATED).toBe('report.generated');
  });

  it('exposes typed payload surfaces', () => {
    const targetImport: TargetImportRequestedPayload = {
      sourceType: 'url',
      targetJobId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
      targetLanguage: 'en-US',
      userId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
    };
    const reportGenerated: ReportGeneratedPayload = {
      modelId: 'model-profile:contract.default',
      preparednessLevel: 'basically_ready',
      promptVersion: 'p1',
      questionIssueCount: 2,
      reportId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
      rubricVersion: 'r1',
      sessionId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
      targetJobId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
    };

    expect(targetImport.sourceType).toBe('url');
    expect(reportGenerated.preparednessLevel).toBe('basically_ready');
  });

  it('locks resume tailor mode to B2/B4-aligned values', () => {
    const gapReview: ResumeTailorCompletedPayload = {
      mode: 'gap_review',
      resumeAssetId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
      status: 'ready',
      tailorRunId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
      targetJobId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
    };
    const bulletSuggestions: ResumeTailorCompletedPayload = {
      ...gapReview,
      mode: 'bullet_suggestions',
    };
    const retiredModePayload: ResumeTailorCompletedPayload = {
      ...gapReview,
      // @ts-expect-error inline was retired by B3 D-14.
      mode: 'inline',
    };

    expect(EVENT_NAME_RESUME_TAILOR_COMPLETED).toBe('resume.tailor.completed');
    expect(gapReview.mode).toBe('gap_review');
    expect(bulletSuggestions.mode).toBe('bullet_suggestions');
    expect(retiredModePayload.mode).toBe('inline');
  });
});
