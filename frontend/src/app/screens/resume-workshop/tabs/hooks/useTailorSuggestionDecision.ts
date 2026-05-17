import { useCallback, useRef, useState } from "react";

import type { ResumeVersion } from "../../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../../runtime/AppRuntimeProvider";
import { generateIdempotencyKey } from "../../../../../lib/conventions/idempotency";

export type SuggestionDecisionErrorKind =
  | "already_decided"
  | "cross_user"
  | "validation"
  | "generic";

export interface SuggestionDecisionErrorEnvelope {
  kind: SuggestionDecisionErrorKind;
  status?: number;
  message: string;
  raw?: unknown;
}

export class SuggestionDecisionError extends Error {
  readonly kind: SuggestionDecisionErrorKind;
  readonly status?: number;
  readonly raw?: unknown;
  constructor(envelope: SuggestionDecisionErrorEnvelope) {
    super(envelope.message);
    this.name = "SuggestionDecisionError";
    this.kind = envelope.kind;
    this.status = envelope.status;
    this.raw = envelope.raw;
  }
}

export interface SuggestionDecisionOutcome {
  version: ResumeVersion;
  idempotencyKey: string;
}

export interface UseTailorSuggestionDecisionResult {
  decide: (
    versionId: string,
    suggestionId: string,
  ) => Promise<SuggestionDecisionOutcome>;
  pendingFor: Record<string, boolean>;
  lastError: SuggestionDecisionError | null;
  resetError: () => void;
  peekIdempotencyKey: (
    versionId: string,
    suggestionId: string,
  ) => string | null;
}

interface IdempotencyKeyEntry {
  versionId: string;
  suggestionId: string;
  key: string;
}

const compositeId = (versionId: string, suggestionId: string): string =>
  `${versionId}::${suggestionId}`;

const extractEnvelope = (
  error: unknown,
): {
  status?: number;
  code?: string;
  message: string;
  reason?: string;
} => {
  if (!(error instanceof Error)) return { message: String(error) };
  const msg = error.message ?? String(error);
  const statusMatch = /HTTP\s+(\d{3})/.exec(msg);
  const status = statusMatch ? Number(statusMatch[1]) : undefined;
  let code: string | undefined;
  let reason: string | undefined;
  const bodyMatch = /:\s+(\{[\s\S]*\})\s*$/.exec(msg);
  if (bodyMatch?.[1]) {
    try {
      const parsed = JSON.parse(bodyMatch[1]) as {
        error?: { code?: string; details?: Record<string, unknown> };
      };
      code = parsed.error?.code;
      const det = parsed.error?.details;
      if (det && typeof det.reason === "string") reason = det.reason;
    } catch {
      // ignore
    }
  }
  if (!code && /VALIDATION_FAILED/i.test(msg)) code = "VALIDATION_FAILED";
  return { status, code, message: msg, reason };
};

export function parseSuggestionDecisionError(
  error: unknown,
): SuggestionDecisionError {
  const env = extractEnvelope(error);
  const reason = env.reason;
  const code = env.code ?? "";
  if (
    env.status === 409 ||
    reason === "SUGGESTION_ALREADY_DECIDED"
  ) {
    return new SuggestionDecisionError({
      kind: "already_decided",
      status: env.status ?? 409,
      message: env.message,
      raw: env,
    });
  }
  if (env.status === 422 || code === "VALIDATION_FAILED") {
    return new SuggestionDecisionError({
      kind: "validation",
      status: env.status ?? 422,
      message: env.message,
      raw: env,
    });
  }
  if (env.status === 404) {
    return new SuggestionDecisionError({
      kind: "cross_user",
      status: 404,
      message: env.message,
      raw: env,
    });
  }
  return new SuggestionDecisionError({
    kind: "generic",
    status: env.status,
    message: env.message,
    raw: env,
  });
}

export type DecisionCallback = (
  versionId: string,
  suggestionId: string,
  opts: { idempotencyKey: string },
) => Promise<ResumeVersion>;

/**
 * Shared base hook for accept / reject suggestion decisions. The actual
 * generated client method is injected so the hook stays agnostic of
 * accept-vs-reject semantics (both share the IK / error mapping contract
 * per plan §4 + backend D-12).
 */
export function useTailorSuggestionDecision(
  pickClientMethod: (
    client: NonNullable<ReturnType<typeof useAppRuntimeOptional>>["client"],
  ) => DecisionCallback,
): UseTailorSuggestionDecisionResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;
  const keyRefs = useRef<Map<string, IdempotencyKeyEntry>>(new Map());
  const [pendingFor, setPendingFor] = useState<Record<string, boolean>>({});
  const [lastError, setLastError] = useState<SuggestionDecisionError | null>(
    null,
  );

  const resetError = useCallback(() => setLastError(null), []);
  const peekIdempotencyKey = useCallback(
    (versionId: string, suggestionId: string) =>
      keyRefs.current.get(compositeId(versionId, suggestionId))?.key ?? null,
    [],
  );

  const decide = useCallback(
    async (
      versionId: string,
      suggestionId: string,
    ): Promise<SuggestionDecisionOutcome> => {
      if (!client) {
        const err = new SuggestionDecisionError({
          kind: "generic",
          message: "runtime client not mounted",
        });
        setLastError(err);
        throw err;
      }
      const id = compositeId(versionId, suggestionId);
      const cached = keyRefs.current.get(id);
      const key = cached?.key ?? generateIdempotencyKey();
      if (!cached) {
        keyRefs.current.set(id, { versionId, suggestionId, key });
      }
      setPendingFor((prev) => ({ ...prev, [id]: true }));
      setLastError(null);
      try {
        const version = await pickClientMethod(client)(
          versionId,
          suggestionId,
          { idempotencyKey: key },
        );
        return { version, idempotencyKey: key };
      } catch (rawErr) {
        const parsed = parseSuggestionDecisionError(rawErr);
        setLastError(parsed);
        if (parsed.kind === "validation") {
          // Validation usually means malformed request body. accept/reject are
          // bodyless, so this is mostly defensive; clear the cached key so a
          // user-corrected retry rotates the IK.
          keyRefs.current.delete(id);
        }
        throw parsed;
      } finally {
        setPendingFor((prev) => {
          const next = { ...prev };
          delete next[id];
          return next;
        });
      }
    },
    [client, pickClientMethod],
  );

  return { decide, pendingFor, lastError, resetError, peekIdempotencyKey };
}
