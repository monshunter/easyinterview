import { useCallback, useMemo, useRef, useState } from "react";

import type {
  CompletePracticeSessionRequest,
  ReportWithJob,
} from "../../../../api/generated/types";
import { newIdempotencyBatch } from "../../../../lib/conventions/idempotency";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export type CompleteState =
  | { kind: "idle" }
  | { kind: "loading" }
  | { kind: "success"; report: ReportWithJob }
  | {
      kind: "error";
      message: string;
      code: number | null;
      attempts: number;
      retryable: boolean;
      fallbackBackToWorkspace: boolean;
    };

export interface UseCompletePracticeSessionResult {
  state: CompleteState;
  ready: boolean;
  complete: () => Promise<ReportWithJob>;
  reset: () => void;
}

const FAILURE_FALLBACK_THRESHOLD = 3;

/**
 * Item 4.1 — completePracticeSession orchestrator.
 *
 * - Body is exactly { clientCompletedAt: ISO8601 }.
 * - Idempotency-Key header is minted via lib/conventions/idempotency.ts
 *   `newIdempotencyBatch().complete` and reused across retries until the
 *   server acknowledges (success / non-retryable error). StrictMode-style
 *   double invocations within the same render cycle dedupe to one POST
 *   via an in-flight Promise cache.
 * - 3 consecutive failures flip `state.fallbackBackToWorkspace=true` so
 *   PracticeScreen can surface the back-to-workspace exit.
 */
export function useCompletePracticeSession(
  explicitSessionId?: string,
): UseCompletePracticeSessionResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const { ctx } = useInterviewContext();
  const sessionId = explicitSessionId ?? ctx.sessionId ?? "";

  const idempotencyKeyRef = useRef<string | null>(null);
  const inFlightRef = useRef<Promise<ReportWithJob> | null>(null);
  const successReportRef = useRef<ReportWithJob | null>(null);
  const attemptsRef = useRef(0);

  const [state, setState] = useState<CompleteState>({ kind: "idle" });

  const reset = useCallback(() => {
    idempotencyKeyRef.current = null;
    inFlightRef.current = null;
    successReportRef.current = null;
    attemptsRef.current = 0;
    setState({ kind: "idle" });
  }, []);

  const complete = useCallback(async (): Promise<ReportWithJob> => {
    if (!client) throw new Error("useCompletePracticeSession: client missing");
    if (!sessionId) {
      throw new Error("useCompletePracticeSession: sessionId missing");
    }
    if (successReportRef.current) {
      return successReportRef.current;
    }
    if (inFlightRef.current) {
      return inFlightRef.current;
    }
    if (!idempotencyKeyRef.current) {
      idempotencyKeyRef.current = newIdempotencyBatch().complete;
    }
    setState({ kind: "loading" });
    const body: CompletePracticeSessionRequest = {
      clientCompletedAt: new Date().toISOString(),
    };
    attemptsRef.current += 1;
    const headers: Record<string, string> = {
      "Idempotency-Key": idempotencyKeyRef.current,
    };
    const promise = client
      .completePracticeSession(sessionId, body, { headers })
      .then((report) => {
        // Success — cache the report so subsequent clicks short-circuit
        // without minting a new key (handoff is one-shot per session).
        inFlightRef.current = null;
        idempotencyKeyRef.current = null;
        successReportRef.current = report;
        attemptsRef.current = 0;
        setState({ kind: "success", report });
        return report;
      })
      .catch((err: unknown) => {
        inFlightRef.current = null;
        const wrapped = err instanceof Error ? err : new Error(String(err));
        const code = parseHttpStatus(wrapped.message);
        const retryable = code === null || (code >= 500 && code < 600);
        const fallbackBackToWorkspace =
          attemptsRef.current >= FAILURE_FALLBACK_THRESHOLD;
        // Non-retryable errors (4xx) clear the inflight key so a fresh
        // attempt mints a new id; retryable errors keep the same key.
        if (!retryable) {
          idempotencyKeyRef.current = null;
        }
        setState({
          kind: "error",
          message: wrapped.message,
          code,
          attempts: attemptsRef.current,
          retryable,
          fallbackBackToWorkspace,
        });
        throw wrapped;
      });
    inFlightRef.current = promise;
    return promise;
  }, [client, sessionId]);

  return useMemo<UseCompletePracticeSessionResult>(
    () => ({
      state,
      ready: !!client && !!sessionId,
      complete,
      reset,
    }),
    [state, client, sessionId, complete, reset],
  );
}

function parseHttpStatus(message: string): number | null {
  const m = /^HTTP (\d{3}) /.exec(message);
  return m ? Number(m[1]) : null;
}
