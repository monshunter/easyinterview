import { useCallback, useEffect, useRef, useState } from "react";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type {
  ApiErrorCode,
  FeedbackReport,
} from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { isValidFeedbackReport } from "../../report/reportContract";

export type ReportGenerationPollState =
  | "idle"
  | "polling"
  | "ready"
  | "failed"
  | "invalid"
  | "timeout"
  | "error"
  | "paused";

export interface UseReportGenerationPollOptions {
  reportId: string;
  /** Fired when the poller observes status='ready'. */
  onReady?: (report: FeedbackReport) => void;
  /**
   * Fired when the poller observes a contract-valid status='failed' response
   * or HTTP 404 (errorCode = REPORT_NOT_FOUND).
   */
  onFailed?: (errorCode: ApiErrorCode | string) => void;
  /** Exponential backoff start; multiplied by `backoffFactor` per attempt. */
  initialDelayMs?: number;
  backoffFactor?: number;
  maxDelayMs?: number;
  maxAttempts?: number;
  /** Inject a custom waiter for fake-timer tests; default uses window.setTimeout. */
  scheduler?: PollScheduler;
}

export interface UseReportGenerationPollResult {
  state: ReportGenerationPollState;
  attemptCount: number;
  report: FeedbackReport | null;
  errorCode: ApiErrorCode | string | null;
  retry: () => void;
}

export interface PollScheduler {
  schedule: (ms: number, cb: () => void) => () => void;
}

interface PollOwner {
  client: EasyInterviewClient | null;
  reportId: string;
}

interface OwnedFeedbackReport {
  client: EasyInterviewClient;
  reportId: string;
  value: FeedbackReport;
}

const DEFAULT_INITIAL_DELAY_MS = 1500;
const DEFAULT_BACKOFF_FACTOR = 1.5;
const DEFAULT_MAX_DELAY_MS = 8000;
/**
 * Keeps status polling active for about 6 minutes. The backend may spend up to
 * four 60-second provider calls with report-specific 10s / 20s / 40s retry
 * backoffs, so the UI must not declare a timeout while that durable retry
 * budget is still running.
 */
export const REPORT_GENERATION_POLL_MAX_ATTEMPTS = 49;

const HTTP_NOT_FOUND_MARKER = "HTTP 404";

const defaultScheduler: PollScheduler = {
  schedule(ms, cb) {
    const timer = setTimeout(cb, ms);
    return () => clearTimeout(timer);
  },
};

/**
 * Polls `getFeedbackReport(reportId)` until the AI generation either succeeds,
 * fails, or hits max attempts. Surface contract:
 *
 * - State machine: idle → polling ↔ paused / error → ready | failed | invalid | timeout.
 * - Exponential backoff: initial 1.5s × 1.5 capped at 8s; max attempts 49.
 * - Visibility / focus: poller suspends while the tab is hidden and resumes
 *   on visible / focus. Suspension does not consume an attempt.
 * - Read-only contract: requests are sent through the generated client and
 *   never carry an `Idempotency-Key` header — `getFeedbackReport` is a pure
 *   read per openapi.yaml.
 * - HTTP 404 maps to `failed` + errorCode `REPORT_NOT_FOUND` (cross-user
 *   isolation per backend-review D-15 / B1).
 * - Unmount cancels the inflight request and prevents further state updates.
 */
export function useReportGenerationPoll(
  options: UseReportGenerationPollOptions,
): UseReportGenerationPollResult {
  const {
    reportId,
    onReady,
    onFailed,
    initialDelayMs = DEFAULT_INITIAL_DELAY_MS,
    backoffFactor = DEFAULT_BACKOFF_FACTOR,
    maxDelayMs = DEFAULT_MAX_DELAY_MS,
    maxAttempts = REPORT_GENERATION_POLL_MAX_ATTEMPTS,
    scheduler = defaultScheduler,
  } = options;

  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;

  const initialState: ReportGenerationPollState = !reportId
    ? "error"
    : !client
      ? "idle"
      : "polling";

  const [state, setState] = useState<ReportGenerationPollState>(initialState);
  const [attemptCount, setAttemptCount] = useState(0);
  const [stateOwner, setStateOwner] = useState<PollOwner>(() => ({
    client,
    reportId,
  }));
  const [ownedReport, setOwnedReport] = useState<OwnedFeedbackReport | null>(null);
  const [errorCode, setErrorCode] = useState<ApiErrorCode | string | null>(null);

  const onReadyRef = useRef(onReady);
  const onFailedRef = useRef(onFailed);
  onReadyRef.current = onReady;
  onFailedRef.current = onFailed;

  const stateRef = useRef<ReportGenerationPollState>(initialState);
  stateRef.current = state;

  // Used to invalidate scheduled timers / inflight responses when the caller
  // unmounts or retries.
  const runSeqRef = useRef(0);
  const resumePlanRef = useRef<{ attempt: number; delay: number } | null>(null);

  const retry = useCallback(() => {
    if (!reportId || !client) return;
    resumePlanRef.current = null;
    setAttemptCount(0);
    setErrorCode(null);
    setState("polling");
    runSeqRef.current += 1;
  }, [client, reportId]);

  useEffect(() => {
    setStateOwner({ client, reportId });
    setAttemptCount(0);
    setOwnedReport(null);
    setErrorCode(null);
    if (!reportId) {
      setState("error");
      return;
    }
    if (!client) {
      setState("idle");
      return;
    }
    // (Re)mount: start fresh.
    resumePlanRef.current = null;
    runSeqRef.current += 1;
    setState("polling");
  }, [client, reportId]);

  useEffect(() => {
    if (!reportId || !client) return;
    if (state !== "polling") return;

    const seq = runSeqRef.current;
    const controller = new AbortController();
    let cancelTimer: (() => void) | null = null;
    let cancelled = false;

    const finalize = (next: ReportGenerationPollState, code?: string | null) => {
      if (cancelled || runSeqRef.current !== seq) return;
      resumePlanRef.current = null;
      if (code !== undefined) setErrorCode(code);
      setState(next);
    };

    const scheduleAttempt = (attempt: number, delay: number) => {
      resumePlanRef.current = { attempt, delay };
      cancelTimer = scheduler.schedule(delay, () => {
        cancelTimer = null;
        if (stateRef.current === "paused" || stateRef.current === "ready") {
          return;
        }
        resumePlanRef.current = null;
        runAttempt(attempt);
      });
    };

    const runAttempt = (attempt: number) => {
      if (cancelled || runSeqRef.current !== seq) return;
      const nextDelay = Math.min(
        maxDelayMs,
        initialDelayMs * Math.pow(backoffFactor, attempt - 1),
      );
      resumePlanRef.current = {
        attempt: attempt + 1,
        delay: attempt < maxAttempts ? nextDelay : 0,
      };
      setAttemptCount(attempt);
      client
        .getFeedbackReport(reportId, { signal: controller.signal })
        .then((next) => {
          if (cancelled || runSeqRef.current !== seq) return;
          if (!isValidFeedbackReport(next, reportId)) {
            finalize("invalid", "AI_OUTPUT_INVALID");
            return;
          }
          setOwnedReport({ client, reportId, value: next });
          if (next.status === "ready") {
            finalize("ready", null);
            onReadyRef.current?.(next);
            return;
          }
          if (next.status === "failed") {
            const code = next.errorCode ?? "UNKNOWN";
            finalize("failed", code);
            onFailedRef.current?.(code);
            return;
          }
          // Still generating — schedule the next backoff if we have budget.
          if (attempt >= maxAttempts) {
            finalize("timeout");
            return;
          }
          scheduleAttempt(attempt + 1, nextDelay);
        })
        .catch((err: unknown) => {
          if (cancelled || runSeqRef.current !== seq) return;
          if (isAbortError(err)) return;
          const message = err instanceof Error ? err.message : String(err);
          if (message.startsWith(HTTP_NOT_FOUND_MARKER)) {
            const code = "REPORT_NOT_FOUND";
            finalize("failed", code);
            onFailedRef.current?.(code);
            return;
          }
          if (attempt >= maxAttempts) {
            finalize("error");
            return;
          }
          // 5xx / network — retry with backoff but charge this attempt.
          scheduleAttempt(attempt + 1, nextDelay);
        });
    };

    const resumePlan = resumePlanRef.current;
    if (resumePlan) {
      if (resumePlan.attempt > maxAttempts) {
        finalize("timeout");
      } else {
        scheduleAttempt(resumePlan.attempt, resumePlan.delay);
      }
    } else {
      runAttempt(1);
    }

    return () => {
      cancelled = true;
      controller.abort();
      if (cancelTimer) cancelTimer();
    };
  }, [
    backoffFactor,
    client,
    initialDelayMs,
    maxAttempts,
    maxDelayMs,
    reportId,
    scheduler,
    state,
  ]);

  // Visibility / focus pause-resume. A scheduled wait keeps its planned
  // attempt, while an aborted in-flight read remains a started attempt and
  // resumes at n+1. Both paths preserve the monotonic max-attempt cap.
  useEffect(() => {
    if (!reportId || !client) return;
    if (typeof document === "undefined" || typeof window === "undefined") return;

    const pause = () => {
      if (stateRef.current === "polling") {
        stateRef.current = "paused";
        setState("paused");
      }
    };
    const resume = () => {
      if (stateRef.current === "paused") {
        stateRef.current = "polling";
        setState("polling");
      }
    };
    const onVisibility = () => {
      if (document.visibilityState === "hidden") pause();
      else resume();
    };
    const onBlur = () => pause();
    const onFocus = () => resume();

    document.addEventListener("visibilitychange", onVisibility);
    window.addEventListener("blur", onBlur);
    window.addEventListener("focus", onFocus);
    return () => {
      document.removeEventListener("visibilitychange", onVisibility);
      window.removeEventListener("blur", onBlur);
      window.removeEventListener("focus", onFocus);
    };
  }, [client, reportId]);

  const stateOwnerMatches =
    stateOwner.client === client && stateOwner.reportId === reportId;
  const report =
    ownedReport?.client === client && ownedReport.reportId === reportId
      ? ownedReport.value
      : null;

  return {
    state: stateOwnerMatches ? state : initialState,
    attemptCount: stateOwnerMatches ? attemptCount : 0,
    report,
    errorCode: stateOwnerMatches ? errorCode : null,
    retry,
  };
}

function isAbortError(err: unknown): boolean {
  if (!err) return false;
  if (typeof err === "object" && "name" in err) {
    return (err as { name?: string }).name === "AbortError";
  }
  return false;
}
