// Hand-written id helpers for frontend code. Generated constants live in
// generated.ts and are produced by `make codegen-conventions` from
// shared/conventions.yaml — do not duplicate the prefix or regex here.

import { v7 as uuidv7 } from 'uuid';

import {
  TMP_ID_PREFIX,
  SAMPLE_UUID_V7,
  UUID_V7_REGEX,
} from './generated';

export { TMP_ID_PREFIX, SAMPLE_UUID_V7, UUID_V7_REGEX };

/** Mints a fresh UUIDv7 string suitable for new business primary keys. */
export function newId(): string {
  return uuidv7();
}

/**
 * Throws when the input is not a syntactically valid UUIDv7 or carries the
 * browser-only `tmp_` prefix. Use this at every server-bound boundary that
 * accepts client-supplied identifiers.
 */
export function requireServerId(id: string): string {
  if (!id) {
    throw new Error(`requireServerId: empty id`);
  }
  if (id.startsWith(TMP_ID_PREFIX)) {
    throw new Error(`requireServerId: tmp_ prefixed id rejected (${id})`);
  }
  if (!UUID_V7_REGEX.test(id)) {
    throw new Error(`requireServerId: not a valid UUIDv7 (${id})`);
  }
  return id;
}

/** True iff `id` is a syntactically valid UUIDv7 without the tmp_ prefix. */
export function isServerId(id: string): boolean {
  if (!id || id.startsWith(TMP_ID_PREFIX)) {
    return false;
  }
  return UUID_V7_REGEX.test(id);
}
