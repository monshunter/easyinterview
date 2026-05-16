import { useCallback, useEffect, useRef, useState } from "react";

import type {
  ApiErrorCode,
  FeedbackReport,
} from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export type UseFeedbackReportState =
  | "idle"
  | "loading"
  | "data"
  | "error"
  | "notFound";

export interface UseFeedbackReportResult {
  state: UseFeedbackReportState;
  data: FeedbackReport | null;
  errorCode: ApiErrorCode | string | null;
  error: Error | null;
  refresh: () => void;
}

const HTTP_NOT_FOUND_MARKER = "HTTP 404";

/**
 * Single-shot `getFeedbackReport(reportId)` loader for ReportScreen.
 *
 * - 4-state machine (idle / loading / data / error / notFound).
 * - Cross-user 404 (REPORT_NOT_FOUND) is surfaced as `notFound`, never as a
 *   generic error. Higher layers map this to the dashboard not-found UI.
 * - Read-only: requests never carry an Idempotency-Key header.
 * - `refresh()` retries with the same reportId; unmount cancels inflight.
 */
export function useFeedbackReport(reportId: string): UseFeedbackReportResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;

  const initial: UseFeedbackReportState = !reportId
    ? "error"
    : client
      ? "loading"
      : "idle";

  const [state, setState] = useState<UseFeedbackReportState>(initial);
  const [data, setData] = useState<FeedbackReport | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [errorCode, setErrorCode] = useState<ApiErrorCode | string | null>(null);
  const [refreshSeq, setRefreshSeq] = useState(0);
  const runSeqRef = useRef(0);

  const refresh = useCallback(() => {
    setRefreshSeq((value) => value + 1);
  }, []);

  useEffect(() => {
    if (!reportId) {
      setState("error");
      setData(null);
      setError(new Error("missing reportId"));
      setErrorCode(null);
      return;
    }
    if (!client) {
      setState("idle");
      return;
    }
    setState("loading");
    setData(null);
    setError(null);
    setErrorCode(null);

    const seq = runSeqRef.current + 1;
    runSeqRef.current = seq;
    const controller = new AbortController();

    client
      .getFeedbackReport(reportId, { signal: controller.signal })
      .then((next) => {
        if (runSeqRef.current !== seq) return;
        setData(next);
        setState("data");
        if (next.status === "failed") setErrorCode(next.errorCode ?? null);
      })
      .catch((err: unknown) => {
        if (runSeqRef.current !== seq) return;
        if (isAbortError(err)) return;
        const wrapped = err instanceof Error ? err : new Error(String(err));
        setError(wrapped);
        if (wrapped.message.startsWith(HTTP_NOT_FOUND_MARKER)) {
          setState("notFound");
          setErrorCode("REPORT_NOT_FOUND");
          return;
        }
        setState("error");
      });

    return () => {
      controller.abort();
    };
  }, [client, reportId, refreshSeq]);

  return { state, data, error, errorCode, refresh };
}

function isAbortError(err: unknown): boolean {
  if (!err) return false;
  if (typeof err === "object" && "name" in err) {
    return (err as { name?: string }).name === "AbortError";
  }
  return false;
}
