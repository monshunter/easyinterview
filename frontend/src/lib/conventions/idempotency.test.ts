import { describe, it, expect } from 'vitest';
import {
  IDEMPOTENCY_KEY_VERSION,
  generateIdempotencyKey,
  parseIdempotencyKey,
  isIdempotencyKeyExpired,
  formatIdempotencyKey,
} from './idempotency';
import { newId } from '../ids';
import { IDEMPOTENCY_KEY_TTL_SECONDS } from './pagination';

describe('generateIdempotencyKey', () => {
  it('emits the v1 prefix and three dot-separated parts', () => {
    const key = generateIdempotencyKey();
    const parts = key.split('.');
    expect(parts).toHaveLength(3);
    expect(parts[0]).toBe(IDEMPOTENCY_KEY_VERSION);
  });
  it('round-trips through parseIdempotencyKey', () => {
    const key = generateIdempotencyKey();
    const parsed = parseIdempotencyKey(key);
    expect(parsed.version).toBe(IDEMPOTENCY_KEY_VERSION);
    expect(parsed.uuid.length).toBeGreaterThan(0);
    expect(parsed.issuedAt.getTime()).toBeGreaterThan(0);
  });
});

describe('parseIdempotencyKey', () => {
  it.each([
    ['', 'empty'],
    ['v1.123', 'two-parts'],
    ['v0.123.0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e', 'bad-version'],
    ['v1.notnum.0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e', 'bad-timestamp'],
    ['v1.123.not-a-uuid', 'bad-uuid'],
    ['v1.123.tmp_0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e', 'tmp-prefix'],
  ])('rejects %s case', (input) => {
    expect(() => parseIdempotencyKey(input)).toThrow();
  });
});

describe('isIdempotencyKeyExpired', () => {
  it('fresh key is not expired', () => {
    const key = generateIdempotencyKey();
    expect(isIdempotencyKeyExpired(key)).toBe(false);
  });
  it('25-hour-old key is reported expired', () => {
    const past = new Date(Date.now() - 25 * 3600 * 1000);
    const key = formatIdempotencyKey(past, newId());
    expect(isIdempotencyKeyExpired(key)).toBe(true);
  });
  it('TTL matches the generated constant from §3.4', () => {
    expect(IDEMPOTENCY_KEY_TTL_SECONDS).toBe(86400);
  });
});
