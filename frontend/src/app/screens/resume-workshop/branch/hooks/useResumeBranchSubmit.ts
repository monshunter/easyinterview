import { useCallback, useRef, useState } from "react";

import type {
  BranchResumeVersionAccepted,
  ResumeVersion,
} from "../../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../../runtime/AppRuntimeProvider";
import { generateIdempotencyKey } from "../../../../../lib/conventions/idempotency";
import { mapBranchFormToBranchResumeVersionRequest } from "../adapters/mapBranchFormToRequest";
import type {
  BranchSubmitContext,
} from "../adapters/mapBranchFormToRequest";
import type { ResumeBranchFormDraft } from "../ResumeBranchFlow";

export type BranchSubmitOutcome =
  | {
      kind: "version";
      version: ResumeVersion;
      idempotencyKey: string;
    }
  | {
      kind: "accepted";
      accepted: BranchResumeVersionAccepted;
      idempotencyKey: string;
    };

export type BranchSubmitErrorKind =
  | "validation"
  | "parent_missing"
  | "target_missing"
  | "idempotency_conflict"
  | "cross_user"
  | "generic";

export interface BranchSubmitErrorEnvelope {
  kind: BranchSubmitErrorKind;
  message: string;
  status?: number;
  field?: string;
  raw?: unknown;
}

export class BranchSubmitError extends Error {
  readonly kind: BranchSubmitErrorKind;
  readonly status?: number;
  readonly field?: string;
  readonly raw?: unknown;
  constructor(envelope: BranchSubmitErrorEnvelope) {
    super(envelope.message);
    this.name = "BranchSubmitError";
    this.kind = envelope.kind;
    this.status = envelope.status;
    this.field = envelope.field;
    this.raw = envelope.raw;
  }
}

export interface UseResumeBranchSubmitResult {
  submit: (
    draft: ResumeBranchFormDraft,
    context: BranchSubmitContext,
  ) => Promise<BranchSubmitOutcome>;
  submitting: boolean;
  lastError: BranchSubmitError | null;
  resetError: () => void;
  /**
   * Test seam: returns the Idempotency-Key currently cached for the active
   * payload fingerprint. Production code should not introspect this; tests use
   * it to verify replay reuse vs. fresh-key generation.
   */
  peekIdempotencyKey: () => string | null;
}

interface IdempotencyEntry {
  fingerprint: string;
  key: string;
}

function buildFingerprint(
  draft: ResumeBranchFormDraft,
  context: BranchSubmitContext,
): string {
  return JSON.stringify({
    name: draft.name.trim(),
    target: draft.target.trim(),
    focus: draft.focus,
    seed: draft.seed,
    parentVersionId: context.parentVersionId,
    targetJobId: context.targetJobId,
  });
}

const isBranchResumeVersionAccepted = (
  response: unknown,
): response is BranchResumeVersionAccepted => {
  if (!response || typeof response !== "object") return false;
  const obj = response as Record<string, unknown>;
  return (
    typeof obj.resumeVersionId === "string" &&
    obj.job !== undefined &&
    obj.version !== undefined
  );
};

function extractErrorEnvelope(error: unknown): {
  status?: number;
  code?: string;
  message: string;
  details?: Record<string, unknown>;
} {
  if (!(error instanceof Error)) {
    return { message: String(error) };
  }
  const msg = error.message ?? String(error);
  const statusMatch = /HTTP\s+(\d{3})/.exec(msg);
  const status = statusMatch ? Number(statusMatch[1]) : undefined;
  let code: string | undefined;
  let details: Record<string, unknown> | undefined;
  const bodyMatch = /:\s+(\{[\s\S]*\})\s*$/.exec(msg);
  if (bodyMatch?.[1]) {
    try {
      const parsed = JSON.parse(bodyMatch[1]) as {
        error?: { code?: string; details?: Record<string, unknown> };
      };
      code = parsed.error?.code;
      details = parsed.error?.details;
    } catch {
      // fall through; fall back to regex code detection
    }
  }
  if (!code && /VALIDATION_FAILED/i.test(msg)) code = "VALIDATION_FAILED";
  return { status, code, message: msg, details };
}

function parseBranchSubmitError(error: unknown): BranchSubmitError {
  const env = extractErrorEnvelope(error);
  const code = env.code ?? "";
  if (env.status === 422 || code === "VALIDATION_FAILED") {
    const field =
      typeof env.details?.field === "string"
        ? env.details.field
        : undefined;
    return new BranchSubmitError({
      kind: "validation",
      status: env.status ?? 422,
      message: env.message,
      field,
      raw: env.details,
    });
  }
  if (env.status === 404) {
    const reason =
      typeof env.details?.reason === "string"
        ? env.details.reason
        : undefined;
    if (reason === "PARENT_NOT_FOUND") {
      return new BranchSubmitError({
        kind: "parent_missing",
        status: 404,
        message: env.message,
        raw: env.details,
      });
    }
    if (reason === "TARGET_JOB_NOT_FOUND") {
      return new BranchSubmitError({
        kind: "target_missing",
        status: 404,
        message: env.message,
        raw: env.details,
      });
    }
    // Cross-user / unknown 404 — both are surfaced as a generic 404 to avoid
    // leaking ownership info; the UI maps both to the same toast copy.
    return new BranchSubmitError({
      kind: "cross_user",
      status: 404,
      message: env.message,
      raw: env.details,
    });
  }
  if (env.status === 409) {
    return new BranchSubmitError({
      kind: "idempotency_conflict",
      status: 409,
      message: env.message,
      raw: env.details,
    });
  }
  return new BranchSubmitError({
    kind: "generic",
    status: env.status,
    message: env.message,
    raw: env.details,
  });
}

export function useResumeBranchSubmit(): UseResumeBranchSubmitResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;
  const idemRef = useRef<IdempotencyEntry | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [lastError, setLastError] = useState<BranchSubmitError | null>(null);

  const resetError = useCallback(() => setLastError(null), []);
  const peekIdempotencyKey = useCallback(
    () => idemRef.current?.key ?? null,
    [],
  );

  const submit = useCallback(
    async (
      draft: ResumeBranchFormDraft,
      context: BranchSubmitContext,
    ): Promise<BranchSubmitOutcome> => {
      if (!client) {
        const err = new BranchSubmitError({
          kind: "generic",
          message: "runtime client not mounted",
        });
        setLastError(err);
        throw err;
      }
      const fingerprint = buildFingerprint(draft, context);
      let key: string;
      if (idemRef.current?.fingerprint === fingerprint) {
        key = idemRef.current.key;
      } else {
        key = generateIdempotencyKey();
        idemRef.current = { fingerprint, key };
      }
      const body = mapBranchFormToBranchResumeVersionRequest(draft, context);
      setSubmitting(true);
      setLastError(null);
      try {
        const response = await client.branchResumeVersion(body, {
          idempotencyKey: key,
        });
        if (isBranchResumeVersionAccepted(response)) {
          return { kind: "accepted", accepted: response, idempotencyKey: key };
        }
        return {
          kind: "version",
          version: response as ResumeVersion,
          idempotencyKey: key,
        };
      } catch (rawErr) {
        const parsed = parseBranchSubmitError(rawErr);
        setLastError(parsed);
        // VALIDATION_FAILED means the body was wrong; reset the IK cache so a
        // user-corrected retry uses a fresh key (per plan §2.1 IK replay
        // contract).
        if (parsed.kind === "validation") {
          idemRef.current = null;
        }
        throw parsed;
      } finally {
        setSubmitting(false);
      }
    },
    [client],
  );

  return { submit, submitting, lastError, resetError, peekIdempotencyKey };
}
