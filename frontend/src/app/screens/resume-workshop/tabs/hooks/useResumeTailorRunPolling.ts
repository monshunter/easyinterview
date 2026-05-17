import { useEffect, useRef, useState } from "react";

import type { ResumeTailorRun } from "../../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../../runtime/AppRuntimeProvider";

export type TailorPollingPhase =
  | "idle"
  | "polling"
  | "ready"
  | "failed"
  | "timeout"
  | "error";

export interface TailorPollingState {
  phase: TailorPollingPhase;
  attempt: number;
  run: ResumeTailorRun | null;
  lastError: Error | null;
}

export interface UseResumeTailorRunPollingOptions {
  /** Initial delay (ms) before the first poll attempt. Default 1500ms. */
  initialDelayMs?: number;
  /** Backoff multiplier per attempt. Default 1.4x. */
  backoffFactor?: number;
  /** Maximum poll attempts before timeout. Default 12 (~60s with default 1500/1.4). */
  maxAttempts?: number;
  /** Notified once when phase=ready, with the resolved run for cache invalidation. */
  onReady?: (run: ResumeTailorRun) => void;
  /** Notified once when phase enters failed/timeout/error. */
  onFailure?: (state: TailorPollingState) => void;
}

export interface UseResumeTailorRunPollingResult extends TailorPollingState {
  /** Restart polling (resets attempt counter). */
  retry: () => void;
}

const TERMINAL_RUN_STATUSES = new Set<ResumeTailorRun["status"]>([
  "ready",
  "failed",
]);

export function useResumeTailorRunPolling(
  tailorRunId: string | null,
  options: UseResumeTailorRunPollingOptions = {},
): UseResumeTailorRunPollingResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;
  const initialDelayMs = options.initialDelayMs ?? 1500;
  const backoffFactor = options.backoffFactor ?? 1.4;
  const maxAttempts = options.maxAttempts ?? 12;
  const onReadyRef = useRef(options.onReady);
  const onFailureRef = useRef(options.onFailure);
  onReadyRef.current = options.onReady;
  onFailureRef.current = options.onFailure;

  const [state, setState] = useState<TailorPollingState>({
    phase: tailorRunId ? "polling" : "idle",
    attempt: 0,
    run: null,
    lastError: null,
  });
  const [retryNonce, setRetryNonce] = useState(0);

  useEffect(() => {
    if (!tailorRunId || !client) {
      setState({ phase: "idle", attempt: 0, run: null, lastError: null });
      return;
    }

    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | null = null;

    setState({ phase: "polling", attempt: 0, run: null, lastError: null });

    const tick = async (attempt: number) => {
      if (cancelled) return;
      if (attempt >= maxAttempts) {
        setState((prev) => {
          const next: TailorPollingState = {
            phase: "timeout",
            attempt,
            run: prev.run,
            lastError: prev.lastError,
          };
          onFailureRef.current?.(next);
          return next;
        });
        return;
      }
      const nextAttempt = attempt + 1;
      setState((prev) => ({ ...prev, phase: "polling", attempt: nextAttempt }));
      try {
        const run = await client.getResumeTailorRun(tailorRunId);
        if (cancelled) return;
        if (TERMINAL_RUN_STATUSES.has(run.status)) {
          const phase: TailorPollingPhase =
            run.status === "ready" ? "ready" : "failed";
          const next: TailorPollingState = {
            phase,
            attempt: nextAttempt,
            run,
            lastError: null,
          };
          setState(next);
          if (phase === "ready") onReadyRef.current?.(run);
          else onFailureRef.current?.(next);
          return;
        }
        // not terminal: continue with exponential backoff
        setState((prev) => ({ ...prev, run, attempt: nextAttempt }));
        const wait = initialDelayMs * Math.pow(backoffFactor, attempt);
        timer = setTimeout(() => void tick(nextAttempt), wait);
      } catch (rawErr) {
        if (cancelled) return;
        const err = rawErr instanceof Error ? rawErr : new Error(String(rawErr));
        const next: TailorPollingState = {
          phase: "error",
          attempt: nextAttempt,
          run: null,
          lastError: err,
        };
        setState(next);
        onFailureRef.current?.(next);
      }
    };

    timer = setTimeout(() => void tick(0), initialDelayMs);

    return () => {
      cancelled = true;
      if (timer) clearTimeout(timer);
    };
  }, [
    backoffFactor,
    client,
    initialDelayMs,
    maxAttempts,
    retryNonce,
    tailorRunId,
  ]);

  return {
    ...state,
    retry: () => setRetryNonce((n) => n + 1),
  };
}
