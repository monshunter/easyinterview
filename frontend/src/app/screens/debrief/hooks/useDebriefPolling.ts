import { useCallback, useEffect, useRef, useState } from "react";

import type { Debrief, Job } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import type { DebriefPollingState } from "../types";

export interface UseDebriefPollingArgs {
  debriefJobId: string | null;
  debriefId: string | null;
  /** Optional: skip polling (e.g. before submission, or when component unmounts). */
  enabled?: boolean;
}

export interface UseDebriefPollingState {
  state: DebriefPollingState;
  job: Job | null;
  debrief: Debrief | null;
  errorCode: string | null;
  attempts: number;
  /** Manual restart for the Phase 5.3 timeout / failure CTAs. */
  restart: () => void;
}

const INITIAL_INTERVAL_MS = 1500;
const MAX_INTERVAL_MS = 8000;
const BACKOFF = 1.5;
const MAX_ATTEMPTS = 30;

/**
 * Phase 5.2 — two-track polling. Phase A polls `getJob(debriefJobId)` with
 * exponential backoff (1.5s × 1.5 capped at 8s, max 30 attempts). When the
 * job reports `succeeded` the hook moves to Phase B and reads the
 * enriched debrief via `getDebrief(debriefId)`. `failed` and max-attempts
 * surface dedicated `failed` / `timeout` states. Polling pauses when the
 * document loses visibility and resumes on visibilitychange / focus.
 */
export function useDebriefPolling({
  debriefJobId,
  debriefId,
  enabled = true,
}: UseDebriefPollingArgs): UseDebriefPollingState {
  const runtime = useAppRuntimeOptional();
  const [polling, setPolling] = useState<DebriefPollingState>("idle");
  const [job, setJob] = useState<Job | null>(null);
  const [debrief, setDebrief] = useState<Debrief | null>(null);
  const [errorCode, setErrorCode] = useState<string | null>(null);
  const [attempts, setAttempts] = useState(0);
  const [restartTick, setRestartTick] = useState(0);
  const cancelledRef = useRef(false);
  const pausedRef = useRef(false);

  const restart = useCallback(() => {
    setPolling("idle");
    setJob(null);
    setDebrief(null);
    setErrorCode(null);
    setAttempts(0);
    setRestartTick((tick) => tick + 1);
  }, []);

  useEffect(() => {
    if (!runtime || !enabled || !debriefJobId || !debriefId) return;
    cancelledRef.current = false;
    pausedRef.current = false;
    setPolling("running");
    setAttempts(0);

    const onVisibility = () => {
      pausedRef.current = document.visibilityState !== "visible";
    };
    document.addEventListener("visibilitychange", onVisibility);
    window.addEventListener("focus", onVisibility);
    window.addEventListener("blur", onVisibility);
    onVisibility();

    let interval = INITIAL_INTERVAL_MS;
    let attempt = 0;

    const tick = async (): Promise<void> => {
      if (cancelledRef.current) return;
      if (pausedRef.current) {
        window.setTimeout(tick, INITIAL_INTERVAL_MS);
        return;
      }
      try {
        const result = await runtime.client.getJob(debriefJobId);
        if (cancelledRef.current) return;
        setJob(result);
        if (result.status === "succeeded") {
          try {
            const enriched = await runtime.client.getDebrief(debriefId);
            if (cancelledRef.current) return;
            setDebrief(enriched);
            setPolling("succeeded");
          } catch (err) {
            if (cancelledRef.current) return;
            setErrorCode(
              err instanceof Error && /HTTP/.test(err.message)
                ? err.message
                : "DEBRIEF_FETCH_FAILED",
            );
            setPolling("failed");
          }
          return;
        }
        if (result.status === "failed") {
          setErrorCode(result.errorCode ?? "UNKNOWN");
          setPolling("failed");
          return;
        }
        attempt += 1;
        setAttempts(attempt);
        if (attempt >= MAX_ATTEMPTS) {
          setPolling("timeout");
          return;
        }
        interval = Math.min(interval * BACKOFF, MAX_INTERVAL_MS);
        window.setTimeout(tick, interval);
      } catch (err) {
        if (cancelledRef.current) return;
        setErrorCode(
          err instanceof Error ? err.message : "POLLING_FAILED",
        );
        setPolling("failed");
      }
    };

    void tick();

    return () => {
      cancelledRef.current = true;
      document.removeEventListener("visibilitychange", onVisibility);
      window.removeEventListener("focus", onVisibility);
      window.removeEventListener("blur", onVisibility);
    };
  }, [runtime, enabled, debriefJobId, debriefId, restartTick]);

  return {
    state: polling,
    job,
    debrief,
    errorCode,
    attempts,
    restart,
  };
}
