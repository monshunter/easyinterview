import { describe, expect, it } from 'vitest';

import {
  EVENT_NAME_REPORT_GENERATED,
  EVENT_NAME_TARGET_IMPORT_REQUESTED,
  type ReportGeneratedPayload,
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
      mistakeCount: 2,
      modelId: 'gpt-test',
      preparednessLevel: 'basically_ready',
      promptVersion: 'p1',
      reportId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
      rubricVersion: 'r1',
      sessionId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
      targetJobId: '0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e',
    };

    expect(targetImport.sourceType).toBe('url');
    expect(reportGenerated.preparednessLevel).toBe('basically_ready');
  });
});
