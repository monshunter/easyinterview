import { useCallback, useRef, useState } from "react";

import type {
  RequestResumeTailorRequest,
  ResumeTailorRunWithJob,
} from "../../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../../runtime/AppRuntimeProvider";
import { generateIdempotencyKey } from "../../../../../lib/conventions/idempotency";

export type RequestResumeTailorErrorKind =
  | "validation"
  | "cross_user"
  | "idempotency_conflict"
  | "generic";

export interface RequestResumeTailorErrorEnvelope {
  kind: RequestResumeTailorErrorKind;
  status?: number;
  message: string;
  raw?: unknown;
}

export class RequestResumeTailorError extends Error {
  readonly kind: RequestResumeTailorErrorKind;
  readonly status?: number;
  readonly raw?: unknown;
  constructor(envelope: RequestResumeTailorErrorEnvelope) {
    super(envelope.message);
    this.name = "RequestResumeTailorError";
    this.kind = envelope.kind;
    this.status = envelope.status;
    this.raw = envelope.raw;
  }
}

export interface UseRequestResumeTailorResult {
  request: (
    body: RequestResumeTailorRequest,
  ) => Promise<ResumeTailorRunWithJob>;
  pending: boolean;
  lastError: RequestResumeTailorError | null;
  resetError: () => void;
}

const extractEnvelope = (error: unknown) => {
  if (!(error instanceof Error)) return { message: String(error) };
  const msg = error.message ?? String(error);
  const statusMatch = /HTTP\s+(\d{3})/.exec(msg);
  return {
    status: statusMatch ? Number(statusMatch[1]) : undefined,
    message: msg,
  };
};

function parseError(error: unknown): RequestResumeTailorError {
  const env = extractEnvelope(error);
  if (env.status === 422) {
    return new RequestResumeTailorError({
      kind: "validation",
      status: 422,
      message: env.message,
    });
  }
  if (env.status === 404) {
    return new RequestResumeTailorError({
      kind: "cross_user",
      status: 404,
      message: env.message,
    });
  }
  if (env.status === 409) {
    return new RequestResumeTailorError({
      kind: "idempotency_conflict",
      status: 409,
      message: env.message,
    });
  }
  return new RequestResumeTailorError({
    kind: "generic",
    status: env.status,
    message: env.message,
  });
}

const fingerprintOf = (body: RequestResumeTailorRequest): string =>
  JSON.stringify(body);

export function useRequestResumeTailor(): UseRequestResumeTailorResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;
  const keyRef = useRef<Map<string, string>>(new Map());
  const [pending, setPending] = useState(false);
  const [lastError, setLastError] = useState<RequestResumeTailorError | null>(
    null,
  );

  const resetError = useCallback(() => setLastError(null), []);

  const request = useCallback(
    async (body: RequestResumeTailorRequest) => {
      if (!client) {
        const err = new RequestResumeTailorError({
          kind: "generic",
          message: "runtime client not mounted",
        });
        setLastError(err);
        throw err;
      }
      const fp = fingerprintOf(body);
      const cached = keyRef.current.get(fp);
      const key = cached ?? generateIdempotencyKey();
      if (!cached) keyRef.current.set(fp, key);
      setPending(true);
      setLastError(null);
      try {
        return await client.requestResumeTailor(body, { idempotencyKey: key });
      } catch (rawErr) {
        const parsed = parseError(rawErr);
        setLastError(parsed);
        if (parsed.kind === "validation") keyRef.current.delete(fp);
        throw parsed;
      } finally {
        setPending(false);
      }
    },
    [client],
  );

  return { request, pending, lastError, resetError };
}
