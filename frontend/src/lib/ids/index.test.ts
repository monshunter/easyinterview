import { describe, it, expect } from 'vitest';
import {
  newId,
  requireServerId,
  isServerId,
  TMP_ID_PREFIX,
  UUID_V7_REGEX,
} from './index';

describe('newId', () => {
  it('returns a UUIDv7 string', () => {
    const id = newId();
    expect(UUID_V7_REGEX.test(id)).toBe(true);
  });
  it('produces unique values', () => {
    expect(newId()).not.toBe(newId());
  });
});

describe('requireServerId', () => {
  it('accepts a freshly minted UUIDv7', () => {
    const id = newId();
    expect(requireServerId(id)).toBe(id);
  });
  it('rejects empty input', () => {
    expect(() => requireServerId('')).toThrow();
  });
  it('rejects tmp_ prefixed input', () => {
    expect(() => requireServerId(`${TMP_ID_PREFIX}abc`)).toThrow();
    expect(() =>
      requireServerId(`${TMP_ID_PREFIX}0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e`),
    ).toThrow();
  });
  it('rejects non-UUIDv7 strings', () => {
    expect(() => requireServerId('not-a-uuid')).toThrow();
    expect(() =>
      requireServerId('00000000-0000-0000-0000-000000000000'),
    ).toThrow();
  });
});

describe('isServerId', () => {
  it('returns true for valid UUIDv7', () => {
    expect(isServerId(newId())).toBe(true);
  });
  it('returns false for tmp_ prefix', () => {
    expect(isServerId(`${TMP_ID_PREFIX}xxx`)).toBe(false);
  });
  it('returns false for empty', () => {
    expect(isServerId('')).toBe(false);
  });
});
