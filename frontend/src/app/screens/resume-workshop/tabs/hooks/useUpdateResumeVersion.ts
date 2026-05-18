import { useCallback, useRef, useState } from "react";

import type {
  ResumeVersion,
  UpdateResumeVersionRequest,
} from "../../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../../runtime/AppRuntimeProvider";
import { generateIdempotencyKey } from "../../../../../lib/conventions/idempotency";

export type UpdateResumeVersionErrorKind =
  | "validation"
  | "cross_user"
  | "idempotency_conflict"
  | "generic";

export interface UpdateResumeVersionErrorEnvelope {
  kind: UpdateResumeVersionErrorKind;
  status?: number;
  message: string;
  field?: string;
  raw?: unknown;
}

export class UpdateResumeVersionError extends Error {
  readonly kind: UpdateResumeVersionErrorKind;
  readonly status?: number;
  readonly field?: string;
  readonly raw?: unknown;
  constructor(envelope: UpdateResumeVersionErrorEnvelope) {
    super(envelope.message);
    this.name = "UpdateResumeVersionError";
    this.kind = envelope.kind;
    this.status = envelope.status;
    this.field = envelope.field;
    this.raw = envelope.raw;
  }
}

const DISALLOWED_KEYS: ReadonlySet<string> = new Set([
  "versionType",
  "resumeAssetId",
  "parentVersionId",
  "targetJobId",
  "seedStrategy",
]);

/**
 * Filter the partial update body so that the wire never carries forbidden
 * editable-fields. Throws when a caller (test or component) attempts to send
 * one — this surfaces the bug eagerly instead of letting the server reject it.
 */
export function filterUpdateResumeVersionPayload(
  partial: Record<string, unknown>,
): UpdateResumeVersionRequest {
  for (const key of Object.keys(partial)) {
    if (DISALLOWED_KEYS.has(key)) {
      throw new Error(
        `useUpdateResumeVersion: refusing to send disallowed field "${key}" (D-12 + plan 003 §6.2 mapper contract)`,
      );
    }
  }
  const next: UpdateResumeVersionRequest = {};
  if (typeof partial.displayName === "string") next.displayName = partial.displayName;
  if (
    typeof partial.focusAngle === "string" ||
    partial.focusAngle === null
  ) {
    next.focusAngle = partial.focusAngle as string | null;
  }
  if (typeof partial.matchScore === "number" || partial.matchScore === null) {
    next.matchScore = partial.matchScore as number | null;
  }
  if (
    typeof partial.structuredProfile === "object" &&
    partial.structuredProfile !== null
  ) {
    const profile = {
      ...(partial.structuredProfile as Record<string, unknown>),
    };
    delete profile.provenance;
    next.structuredProfile = profile;
  }
  return next;
}

export interface UpdateResumeVersionInput {
  versionId: string;
  payload: Record<string, unknown>;
}

export interface UpdateResumeVersionResult {
  update: (input: UpdateResumeVersionInput) => Promise<{
    version: ResumeVersion;
    idempotencyKey: string;
  }>;
  pendingFor: Record<string, boolean>;
  lastError: UpdateResumeVersionError | null;
  resetError: () => void;
  peekIdempotencyKey: (
    versionId: string,
    payload: Record<string, unknown>,
  ) => string | null;
}

const fingerprintOf = (versionId: string, body: UpdateResumeVersionRequest): string =>
  JSON.stringify({ versionId, ...body });

const extractEnvelope = (error: unknown) => {
  if (!(error instanceof Error)) return { message: String(error) };
  const msg = error.message ?? String(error);
  const statusMatch = /HTTP\s+(\d{3})/.exec(msg);
  const status = statusMatch ? Number(statusMatch[1]) : undefined;
  let code: string | undefined;
  let field: string | undefined;
  const bodyMatch = /:\s+(\{[\s\S]*\})\s*$/.exec(msg);
  if (bodyMatch?.[1]) {
    try {
      const parsed = JSON.parse(bodyMatch[1]) as {
        error?: { code?: string; details?: Record<string, unknown> };
      };
      code = parsed.error?.code;
      const det = parsed.error?.details;
      if (det && typeof det.field === "string") field = det.field;
    } catch {
      // ignore
    }
  }
  if (!code && /VALIDATION_FAILED/i.test(msg)) code = "VALIDATION_FAILED";
  return { status, code, message: msg, field };
};

function parseUpdateError(error: unknown): UpdateResumeVersionError {
  const env = extractEnvelope(error);
  const code = env.code ?? "";
  if (env.status === 422 || code === "VALIDATION_FAILED") {
    return new UpdateResumeVersionError({
      kind: "validation",
      status: env.status ?? 422,
      message: env.message,
      field: env.field,
      raw: env,
    });
  }
  if (env.status === 404) {
    return new UpdateResumeVersionError({
      kind: "cross_user",
      status: 404,
      message: env.message,
      raw: env,
    });
  }
  if (env.status === 409) {
    return new UpdateResumeVersionError({
      kind: "idempotency_conflict",
      status: 409,
      message: env.message,
      raw: env,
    });
  }
  return new UpdateResumeVersionError({
    kind: "generic",
    status: env.status,
    message: env.message,
    raw: env,
  });
}

export function useUpdateResumeVersion(): UpdateResumeVersionResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;
  const keyRefs = useRef<Map<string, string>>(new Map());
  const [pendingFor, setPendingFor] = useState<Record<string, boolean>>({});
  const [lastError, setLastError] = useState<UpdateResumeVersionError | null>(
    null,
  );

  const resetError = useCallback(() => setLastError(null), []);

  const peekIdempotencyKey = useCallback(
    (versionId: string, payload: Record<string, unknown>) => {
      try {
        const filtered = filterUpdateResumeVersionPayload(payload);
        return keyRefs.current.get(fingerprintOf(versionId, filtered)) ?? null;
      } catch {
        return null;
      }
    },
    [],
  );

  const update = useCallback<UpdateResumeVersionResult["update"]>(
    async ({ versionId, payload }) => {
      if (!client) {
        const err = new UpdateResumeVersionError({
          kind: "generic",
          message: "runtime client not mounted",
        });
        setLastError(err);
        throw err;
      }
      const filtered = filterUpdateResumeVersionPayload(payload);
      const fp = fingerprintOf(versionId, filtered);
      const cached = keyRefs.current.get(fp);
      const key = cached ?? generateIdempotencyKey();
      if (!cached) keyRefs.current.set(fp, key);
      setPendingFor((prev) => ({ ...prev, [versionId]: true }));
      setLastError(null);
      try {
        const version = await client.updateResumeVersion(versionId, filtered, {
          idempotencyKey: key,
        });
        return { version, idempotencyKey: key };
      } catch (rawErr) {
        const parsed = parseUpdateError(rawErr);
        setLastError(parsed);
        if (parsed.kind === "validation") {
          keyRefs.current.delete(fp);
        }
        throw parsed;
      } finally {
        setPendingFor((prev) => {
          const next = { ...prev };
          delete next[versionId];
          return next;
        });
      }
    },
    [client],
  );

  return { update, pendingFor, lastError, resetError, peekIdempotencyKey };
}
