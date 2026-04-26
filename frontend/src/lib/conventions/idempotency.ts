// Hand-written Idempotency-Key helper, dual to backend/internal/shared/idx/
// idempotency.go. Wire format: `v1.{unixSeconds}.{uuidv7}`. Spec: 00-shared-
// conventions §3.4 (24h TTL); shared-conventions-codified spec §6 C-4.

import { newId, requireServerId, UUID_V7_REGEX } from '../ids';
import { IDEMPOTENCY_KEY_TTL_SECONDS } from './pagination';

/** Wire-format version. Bumping signals an incompatible format change. */
export const IDEMPOTENCY_KEY_VERSION = 'v1';

const DECIMAL_UNIX_SECONDS = /^\d+$/;

export interface IdempotencyKey {
  version: string;
  issuedAt: Date;
  uuid: string;
}

/** Returns a fresh Idempotency-Key valid for {@link IDEMPOTENCY_KEY_TTL_SECONDS} seconds. */
export function generateIdempotencyKey(now: Date = new Date()): string {
  return formatIdempotencyKey(now, newId());
}

/** Formats an arbitrary timestamp + UUIDv7 into the wire format. */
export function formatIdempotencyKey(issuedAt: Date, uuid: string): string {
  const unixSeconds = Math.floor(issuedAt.getTime() / 1000);
  return `${IDEMPOTENCY_KEY_VERSION}.${unixSeconds}.${uuid}`;
}

/**
 * Validates and decodes a key without checking expiry. Throws when the input
 * is empty, the version is wrong, the timestamp is non-numeric, or the UUID
 * portion is not a valid UUIDv7 (including any value with the `tmp_` prefix).
 */
export function parseIdempotencyKey(raw: string): IdempotencyKey {
  if (!raw) {
    throw new Error('parseIdempotencyKey: empty');
  }
  const parts = raw.split('.');
  if (parts.length !== 3) {
    throw new Error(`parseIdempotencyKey: expected 3 parts, got ${parts.length}`);
  }
  // Length check above narrows the array contract; cast lets noUncheckedIndexedAccess pass.
  const [version, unixStr, uuid] = parts as [string, string, string];
  if (version !== IDEMPOTENCY_KEY_VERSION) {
    throw new Error(`parseIdempotencyKey: version ${version} (only ${IDEMPOTENCY_KEY_VERSION} accepted)`);
  }
  if (!DECIMAL_UNIX_SECONDS.test(unixStr)) {
    throw new Error(`parseIdempotencyKey: timestamp not decimal Unix seconds: ${unixStr}`);
  }
  const unixSec = Number(unixStr);
  if (!Number.isSafeInteger(unixSec)) {
    throw new Error(`parseIdempotencyKey: timestamp outside safe integer range: ${unixStr}`);
  }
  // requireServerId throws on tmp_ prefix and on non-UUIDv7 strings, dual to Go side.
  requireServerId(uuid);
  if (!UUID_V7_REGEX.test(uuid)) {
    // Defensive: requireServerId already enforces this, but keep the explicit check.
    throw new Error(`parseIdempotencyKey: bad uuid ${uuid}`);
  }
  return {
    version,
    issuedAt: new Date(unixSec * 1000),
    uuid,
  };
}

/** True when the key was issued more than {@link IDEMPOTENCY_KEY_TTL_SECONDS} ago. */
export function isIdempotencyKeyExpired(raw: string, now: Date = new Date()): boolean {
  const parsed = parseIdempotencyKey(raw);
  const ageSeconds = (now.getTime() - parsed.issuedAt.getTime()) / 1000;
  return ageSeconds > IDEMPOTENCY_KEY_TTL_SECONDS;
}
