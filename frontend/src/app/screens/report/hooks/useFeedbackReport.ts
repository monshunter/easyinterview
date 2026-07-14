import { useCallback, useEffect, useRef, useState } from "react";

import type { EasyInterviewClient } from "../../../../api/generated/client";
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

interface ReportOwner {
  client: EasyInterviewClient | null;
  reportId: string;
}

interface OwnedFeedbackReport {
  client: EasyInterviewClient;
  reportId: string;
  value: FeedbackReport;
}

const HTTP_NOT_FOUND_MARKER = "HTTP 404";

/**
 * Single-shot `getFeedbackReport(reportId)` loader for ReportScreen.
 *
 * - 4-state machine (idle / loading / data / error / notFound).
 * - Cross-user 404 (REPORT_NOT_FOUND) is surfaced as `notFound`, never as a
 *   generic error. Higher layers map this to the dashboard not-found UI.
 * - Read-only: requests never carry an Idempotency-Key header.
 * - `refresh()` retries with the same reportId; cleanup ignores stale inflight
 *   results without opting the safe GET out of generated-client single-flight.
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
  const [stateOwner, setStateOwner] = useState<ReportOwner>(() => ({
    client,
    reportId,
  }));
  const [ownedData, setOwnedData] = useState<OwnedFeedbackReport | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [errorCode, setErrorCode] = useState<ApiErrorCode | string | null>(null);
  const [refreshSeq, setRefreshSeq] = useState(0);
  const runSeqRef = useRef(0);

  const refresh = useCallback(() => {
    setRefreshSeq((value) => value + 1);
  }, []);

  useEffect(() => {
    setStateOwner({ client, reportId });
    setOwnedData(null);
    setError(null);
    setErrorCode(null);
    if (!reportId) {
      setState("error");
      setError(new Error("missing reportId"));
      return;
    }
    if (!client) {
      setState("idle");
      return;
    }
    setState("loading");

    const seq = runSeqRef.current + 1;
    runSeqRef.current = seq;
    let cancelled = false;

    client
      .getFeedbackReport(reportId)
      .then((next) => {
        if (cancelled || runSeqRef.current !== seq) return;
        setOwnedData({ client, reportId, value: next });
        setState("data");
        if (next.status === "failed") setErrorCode(next.errorCode ?? null);
      })
      .catch((err: unknown) => {
        if (cancelled || runSeqRef.current !== seq) return;
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
      cancelled = true;
    };
  }, [client, reportId, refreshSeq]);

  const stateOwnerMatches =
    stateOwner.client === client && stateOwner.reportId === reportId;
  const data =
    ownedData?.client === client && ownedData.reportId === reportId
      ? ownedData.value
      : null;

  return {
    state: stateOwnerMatches ? state : initial,
    data,
    error: stateOwnerMatches ? error : null,
    errorCode: stateOwnerMatches ? errorCode : null,
    refresh,
  };
}
