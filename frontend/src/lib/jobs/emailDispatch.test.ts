import { describe, expect, it } from 'vitest';

import { buildEmailDispatchPayload } from './jobs';

describe('email_dispatch payload helper', () => {
  it('rejects redacted fields', () => {
    for (const field of ['rawEmailCode', 'emailVerificationUrl', 'recipientEmail', 'emailBody']) {
      expect(() => buildEmailDispatchPayload({
        authChallengeId: 'challenge-1',
        [field]: 'secret',
      })).toThrow(field);
    }
  });

  it('accepts allowed fields', () => {
    expect(buildEmailDispatchPayload({
      authChallengeId: 'challenge-1',
      userId: 'user-1',
      templateKey: 'login',
      locale: 'en-US',
      deliverySecretRef: 'secret-ref-1',
      dedupeKey: 'dedupe-1',
    })).toEqual({
      authChallengeId: 'challenge-1',
      userId: 'user-1',
      templateKey: 'login',
      locale: 'en-US',
      deliverySecretRef: 'secret-ref-1',
      dedupeKey: 'dedupe-1',
    });
  });
});
