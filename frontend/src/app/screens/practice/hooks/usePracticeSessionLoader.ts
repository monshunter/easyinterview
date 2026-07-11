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
  adopt: (session: PracticeSession) => boolean;
}

interface PracticeSessionLoaderSnapshot {
  sessionId: string;
  state: PracticeSessionLoaderState;
  data: PracticeSession | null;
  error: Error | null;
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

  const [snapshot, setSnapshot] = useState<PracticeSessionLoaderSnapshot>(() => ({
    sessionId,
    state: initialState,
    data: null,
    error: null,
  }));
  const [reloadSeq, setReloadSeq] = useState(0);
  const requestSeqRef = useRef(0);

  const refresh = useCallback(() => {
    setReloadSeq((value) => value + 1);
  }, []);

  const adopt = useCallback(
    (session: PracticeSession): boolean => {
      if (!sessionId || session.id !== sessionId) return false;
      setSnapshot({
        sessionId,
        state: "data",
        data: session,
        error: null,
      });
      dispatch({
        type: "MERGE_SESSION",
        session: session as unknown as { id: string; [key: string]: unknown },
      });
      return true;
    },
    [dispatch, sessionId],
  );

  useEffect(() => {
    if (!sessionId) {
      setSnapshot({
        sessionId: "",
        state: "sessionLost",
        data: null,
        error: null,
      });
      return;
    }
    if (!client) {
      setSnapshot({ sessionId, state: "idle", data: null, error: null });
      return;
    }

    let active = true;
    const seq = requestSeqRef.current + 1;
    requestSeqRef.current = seq;
    setSnapshot((previous) => ({
      sessionId,
      state: "loading",
      data: previous.sessionId === sessionId ? previous.data : null,
      error: null,
    }));

    client
      .getPracticeSession(sessionId)
      .then((session) => {
        if (!active || requestSeqRef.current !== seq) return;
        if (adopt(session)) return;
        setSnapshot({
          sessionId,
          state: "error",
          data: null,
          error: new Error("practice session response id mismatch"),
        });
      })
      .catch((err: unknown) => {
        if (!active || requestSeqRef.current !== seq) return;
        const wrapped = err instanceof Error ? err : new Error(String(err));
        setSnapshot({
          sessionId,
          state: isHttpStatus(wrapped, 404) ? "sessionLost" : "error",
          data: null,
          error: wrapped,
        });
      });

    return () => {
      active = false;
    };
  }, [adopt, client, sessionId, reloadSeq]);

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

  const currentSnapshot: PracticeSessionLoaderSnapshot =
    snapshot.sessionId === sessionId
      ? snapshot
      : {
          sessionId,
          state: initialState,
          data: null,
          error: null,
        };

  return {
    state: currentSnapshot.state,
    data: currentSnapshot.data,
    error: currentSnapshot.error,
    refresh,
    adopt,
  };
}

function isHttpStatus(error: Error, status: number): boolean {
  return error.message.startsWith(`HTTP ${status} `);
}
