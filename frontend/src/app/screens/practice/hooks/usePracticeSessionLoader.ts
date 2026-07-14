import { useCallback, useEffect, useRef, useState } from "react";

import {
  ApiClientError,
  type EasyInterviewClient,
} from "../../../../api/generated/client";
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
  adopt: (session: PracticeSession, options?: { readToken?: number }) => boolean;
  read: (options?: { signal?: AbortSignal }) => PracticeSessionRead;
}

export interface PracticeSessionRead {
  token: number;
  result: Promise<PracticeSession>;
}

interface PracticeSessionLoaderSnapshot {
  sessionId: string;
  state: PracticeSessionLoaderState;
  data: PracticeSession | null;
  error: Error | null;
}

const PRACTICE_SESSION_READ_TIMEOUT_MS = 10_000;

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
  const latestReadTokenRef = useRef(0);

  const refresh = useCallback(() => {
    setReloadSeq((value) => value + 1);
  }, []);

  const adopt = useCallback(
    (session: PracticeSession, options?: { readToken?: number }): boolean => {
      if (!sessionId || session.id !== sessionId) return false;
      if (options?.readToken !== undefined) {
        if (options.readToken !== latestReadTokenRef.current) return false;
      } else {
        latestReadTokenRef.current += 1;
      }
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

  const read = useCallback((options?: { signal?: AbortSignal }) => {
    latestReadTokenRef.current += 1;
    const token = latestReadTokenRef.current;
    const result = !client
      ? Promise.reject<PracticeSession>(new Error("usePracticeSessionLoader: client not mounted"))
      : !sessionId
        ? Promise.reject<PracticeSession>(new Error("usePracticeSessionLoader: sessionId missing"))
        : readPracticeSessionBounded(client, sessionId, options?.signal);
    return { token, result };
  }, [client, sessionId]);

  const isReadCurrent = useCallback(
    (token: number) => latestReadTokenRef.current === token,
    [],
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
    const controller = new AbortController();
    setSnapshot((previous) => ({
      sessionId,
      state: "loading",
      data: previous.sessionId === sessionId ? previous.data : null,
      error: null,
    }));

    const pendingRead = read({ signal: controller.signal });
    pendingRead.result
      .then((session) => {
        if (!active) return;
        if (!isReadCurrent(pendingRead.token)) return;
        if (adopt(session, { readToken: pendingRead.token })) return;
        setSnapshot({
          sessionId,
          state: "error",
          data: null,
          error: new Error("practice session response id mismatch"),
        });
      })
      .catch((err: unknown) => {
        if (!active) return;
        if (!isReadCurrent(pendingRead.token)) return;
        const wrapped = err instanceof Error ? err : new Error(String(err));
        const sessionLost = isHttpStatus(wrapped, 404);
        setSnapshot((previous) => ({
          sessionId,
          state: sessionLost ? "sessionLost" : "error",
          data: sessionLost
            ? null
            : previous.sessionId === sessionId
              ? previous.data
              : null,
          error: wrapped,
        }));
      });

    return () => {
      active = false;
      controller.abort();
    };
  }, [adopt, client, isReadCurrent, read, sessionId, reloadSeq]);

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
    read,
  };
}

function readPracticeSessionBounded(
  client: EasyInterviewClient,
  sessionId: string,
  externalSignal?: AbortSignal,
): Promise<PracticeSession> {
  const controller = new AbortController();
  const abortFromCaller = () => controller.abort();
  if (externalSignal?.aborted) {
    controller.abort();
  } else {
    externalSignal?.addEventListener("abort", abortFromCaller, { once: true });
  }
  const timeout = window.setTimeout(
    () => controller.abort(),
    PRACTICE_SESSION_READ_TIMEOUT_MS,
  );
  const aborted = new Promise<never>((_resolve, reject) => {
    const rejectAbort = () => reject(new ApiClientError("abort", null, null));
    if (controller.signal.aborted) {
      rejectAbort();
      return;
    }
    controller.signal.addEventListener("abort", rejectAbort, { once: true });
  });

  return Promise.race([
    client.getPracticeSession(sessionId, { signal: controller.signal }),
    aborted,
  ]).finally(() => {
    window.clearTimeout(timeout);
    externalSignal?.removeEventListener("abort", abortFromCaller);
  });
}

function isHttpStatus(error: Error, status: number): boolean {
  return error instanceof ApiClientError
    && error.kind === "http"
    && error.status === status;
}
