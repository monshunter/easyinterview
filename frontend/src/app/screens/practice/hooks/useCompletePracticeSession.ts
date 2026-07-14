import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { ApiClientError } from "../../../../api/generated/client";
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

interface CompletionScope {
  sessionId: string;
  active: boolean;
  idempotencyKey: string | null;
  inFlight: Promise<ReportWithJob> | null;
  successReport: ReportWithJob | null;
  attempts: number;
}

interface CompletionStateSnapshot {
  sessionId: string;
  state: CompleteState;
}

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

  const scopeRef = useRef<CompletionScope>(createCompletionScope(sessionId));
  if (scopeRef.current.sessionId !== sessionId) {
    scopeRef.current = createCompletionScope(sessionId);
  }
  const renderScope = scopeRef.current;
  const [stateSnapshot, setStateSnapshot] = useState<CompletionStateSnapshot>(() => ({
    sessionId,
    state: { kind: "idle" },
  }));
  const state = stateSnapshot.sessionId === sessionId
    ? stateSnapshot.state
    : { kind: "idle" } satisfies CompleteState;

  useEffect(() => {
    renderScope.active = true;
    return () => {
      renderScope.active = false;
    };
  }, [renderScope]);

  const reset = useCallback(() => {
    const scope = scopeRef.current;
    if (scope.sessionId !== sessionId) return;
    scope.idempotencyKey = null;
    scope.inFlight = null;
    scope.successReport = null;
    scope.attempts = 0;
    setStateSnapshot({ sessionId, state: { kind: "idle" } });
  }, [sessionId]);

  const complete = useCallback(async (): Promise<ReportWithJob> => {
    if (!client) throw new Error("useCompletePracticeSession: client missing");
    if (!sessionId) {
      throw new Error("useCompletePracticeSession: sessionId missing");
    }
    const scope = scopeRef.current.sessionId === sessionId
      ? scopeRef.current
      : createCompletionScope(sessionId);
    if (scopeRef.current !== scope) {
      scopeRef.current = scope;
    }
    const isCurrentScope = () => (
      scope.active
      && scopeRef.current === scope
      && scope.sessionId === sessionId
    );
    if (scope.successReport) {
      return scope.successReport;
    }
    if (scope.inFlight) {
      return scope.inFlight;
    }
    if (!scope.idempotencyKey) {
      scope.idempotencyKey = newIdempotencyBatch().complete;
    }
    setStateSnapshot({ sessionId, state: { kind: "loading" } });
    const body: CompletePracticeSessionRequest = {
      clientCompletedAt: new Date().toISOString(),
    };
    scope.attempts += 1;
    const headers: Record<string, string> = {
      "Idempotency-Key": scope.idempotencyKey,
    };
    const promise = client
      .completePracticeSession(sessionId, body, { headers })
      .then((report) => {
        if (!isCurrentScope()) return report;
        // Success — cache the report so subsequent clicks short-circuit
        // without minting a new key (handoff is one-shot per session).
        scope.inFlight = null;
        scope.idempotencyKey = null;
        scope.successReport = report;
        scope.attempts = 0;
        setStateSnapshot({ sessionId, state: { kind: "success", report } });
        return report;
      })
      .catch((err: unknown) => {
        const wrapped = err instanceof Error ? err : new Error(String(err));
        if (!isCurrentScope()) throw wrapped;
        scope.inFlight = null;
        const code = wrapped instanceof ApiClientError ? wrapped.status : null;
        const retryable = wrapped instanceof ApiClientError
          && (wrapped.kind === "transport"
            || (wrapped.kind === "http"
              && (wrapped.apiError?.error.retryable === true
                || (wrapped.apiError === null && code !== null && code >= 500 && code < 600))));
        const fallbackBackToWorkspace =
          scope.attempts >= FAILURE_FALLBACK_THRESHOLD;
        // Non-retryable errors (4xx) clear the inflight key so a fresh
        // attempt mints a new id; retryable errors keep the same key.
        if (!retryable) {
          scope.idempotencyKey = null;
        }
        setStateSnapshot({
          sessionId,
          state: {
            kind: "error",
            message: wrapped.message,
            code,
            attempts: scope.attempts,
            retryable,
            fallbackBackToWorkspace,
          },
        });
        throw wrapped;
      });
    scope.inFlight = promise;
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

function createCompletionScope(sessionId: string): CompletionScope {
  return {
    sessionId,
    active: true,
    idempotencyKey: null,
    inFlight: null,
    successReport: null,
    attempts: 0,
  };
}
