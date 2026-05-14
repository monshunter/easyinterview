import { useCallback, useEffect, useRef, useState } from "react";

import type { PracticeSession } from "../../../../api/generated/types";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export type PracticeSessionLoaderState =
  | "idle"
  | "loading"
  | "data"
  | "sessionLost"
  | "error";

export interface UsePracticeSessionLoaderResult {
  state: PracticeSessionLoaderState;
  data: PracticeSession | null;
  error: Error | null;
  refresh: () => void;
}

/**
 * Item 1.3 — loads `getPracticeSession(sessionId)` from generated client and
 * tracks idle / loading / data / sessionLost (404 + missing) / error (5xx)
 * five states. Auto-refreshes on document visibility change to visible,
 * window focus, and window online events. Successful payloads dispatch
 * MERGE_SESSION into InterviewContext.
 *
 * Spec D-12 / D-13 boundary: this hook owns the read path only; mutations
 * flow through usePracticeEvents (Phase 2) / useCompletePracticeSession
 * (Phase 4).
 *
 * Optional `explicitSessionId` overrides the InterviewContext read so the
 * caller (PracticeScreen) can supply the route-param sessionId directly
 * without depending on a separate HYDRATE_FROM_ROUTE dispatch.
 */
export function usePracticeSessionLoader(
  explicitSessionId?: string,
): UsePracticeSessionLoaderResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const { ctx, dispatch } = useInterviewContext();
  const sessionId = explicitSessionId ?? ctx.sessionId ?? "";

  const initialState: PracticeSessionLoaderState = !sessionId
    ? "sessionLost"
    : client
      ? "loading"
      : "idle";

  const [state, setState] = useState<PracticeSessionLoaderState>(initialState);
  const [data, setData] = useState<PracticeSession | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [reloadSeq, setReloadSeq] = useState(0);
  const requestSeqRef = useRef(0);

  const refresh = useCallback(() => {
    setReloadSeq((value) => value + 1);
  }, []);

  useEffect(() => {
    if (!sessionId) {
      setState("sessionLost");
      setData(null);
      setError(null);
      return;
    }
    if (!client) {
      setState("idle");
      return;
    }

    let active = true;
    const seq = requestSeqRef.current + 1;
    requestSeqRef.current = seq;
    setState("loading");
    setError(null);

    client
      .getPracticeSession(sessionId)
      .then((session) => {
        if (!active || requestSeqRef.current !== seq) return;
        setData(session);
        setError(null);
        setState("data");
        dispatch({
          type: "MERGE_SESSION",
          session: session as unknown as { id: string; [key: string]: unknown },
        });
      })
      .catch((err: unknown) => {
        if (!active || requestSeqRef.current !== seq) return;
        const wrapped = err instanceof Error ? err : new Error(String(err));
        setError(wrapped);
        setData(null);
        setState(isHttpStatus(wrapped, 404) ? "sessionLost" : "error");
      });

    return () => {
      active = false;
    };
  }, [client, sessionId, dispatch, reloadSeq]);

  useEffect(() => {
    if (!sessionId || !client) return;

    const onFocus = () => refresh();
    const onOnline = () => refresh();
    const onVisibility = () => {
      if (typeof document !== "undefined" && document.visibilityState === "visible") {
        refresh();
      }
    };

    window.addEventListener("focus", onFocus);
    window.addEventListener("online", onOnline);
    document.addEventListener("visibilitychange", onVisibility);
    return () => {
      window.removeEventListener("focus", onFocus);
      window.removeEventListener("online", onOnline);
      document.removeEventListener("visibilitychange", onVisibility);
    };
  }, [sessionId, client, refresh]);

  return { state, data, error, refresh };
}

function isHttpStatus(error: Error, status: number): boolean {
  return error.message.startsWith(`HTTP ${status} `);
}
