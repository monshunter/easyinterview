import { useCallback, useMemo, useRef } from "react";

import type {
  PracticeSessionEventRequest,
  SessionEventResult,
} from "../../../../api/generated/types";
import { newId } from "../../../../lib/ids";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export interface SubmitAnswerInput {
  turnId: string;
  answerText: string;
}

export interface TurnTargetedInput {
  turnId: string;
}

export interface UsePracticeEventsResult {
  ready: boolean;
  submitAnswer: (input: SubmitAnswerInput) => Promise<SessionEventResult>;
  requestHint: (input: TurnTargetedInput) => Promise<SessionEventResult>;
  pauseSession: () => Promise<SessionEventResult>;
  resumeSession: () => Promise<SessionEventResult>;
}

type EventKind = PracticeSessionEventRequest["kind"];

interface InflightRecord {
  fingerprint: string;
  clientEventId: string;
  promise?: Promise<SessionEventResult>;
}

/**
 * Item 2.1 — usePracticeEvents.
 *
 * Four user-facing mutations map 1:1 to the current spec D-12 event kinds.
 * Each mutation
 * builds a `PracticeSessionEventRequest` with a UUIDv7 `clientEventId` and
 * an `occurredAt` ISO timestamp; the request is sent through the generated
 * client (no `Idempotency-Key` header — append events are non-side-effect
 * per spec D-12 / D-13). Retries of the SAME logical user action reuse the
 * same `clientEventId` via an in-flight fingerprint cache; a fresh action
 * (different payload contents) mints a new id.
 */
export function usePracticeEvents(
  explicitSessionId?: string,
): UsePracticeEventsResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const { ctx } = useInterviewContext();
  const sessionId = explicitSessionId ?? ctx.sessionId ?? "";
  const inflightRef = useRef<Map<EventKind, InflightRecord>>(new Map());

  const send = useCallback(
    async (
      kind: EventKind,
      payload: Record<string, unknown>,
    ): Promise<SessionEventResult> => {
      if (!client) throw new Error("usePracticeEvents: client not mounted");
      if (!sessionId) {
        throw new Error("usePracticeEvents: sessionId missing — cannot send");
      }
      const fingerprint = stableFingerprint(kind, payload);
      const existing = inflightRef.current.get(kind);
      const reuseId = existing && existing.fingerprint === fingerprint
        ? existing.clientEventId
        : null;
      const clientEventId = reuseId ?? newId();
      // Always overwrite the per-kind record so a fresh fingerprint replaces
      // the previous one. Retries with identical fingerprint reuse the id.
      inflightRef.current.set(kind, { fingerprint, clientEventId });

      const body: PracticeSessionEventRequest = {
        clientEventId,
        kind,
        occurredAt: new Date().toISOString(),
        payload,
      };

      try {
        const result = await client.appendSessionEvent(sessionId, body);
        // Successful submission clears the inflight slot only when the
        // server acknowledged. Retry chains keep the id until success.
        if (result.acknowledged) {
          inflightRef.current.delete(kind);
        }
        return result;
      } catch (err) {
        // Keep the inflight record so the next retry reuses clientEventId.
        throw err;
      }
    },
    [client, sessionId],
  );

  const submitAnswer = useCallback<UsePracticeEventsResult["submitAnswer"]>(
    (input) => send("answer_submitted", { turnId: input.turnId, answerText: input.answerText }),
    [send],
  );
  const requestHint = useCallback<UsePracticeEventsResult["requestHint"]>(
    (input) => send("hint_requested", { turnId: input.turnId }),
    [send],
  );
  const pauseSession = useCallback<UsePracticeEventsResult["pauseSession"]>(
    () => send("session_paused", {}),
    [send],
  );
  const resumeSession = useCallback<UsePracticeEventsResult["resumeSession"]>(
    () => send("session_resumed", {}),
    [send],
  );

  return useMemo<UsePracticeEventsResult>(
    () => ({
      ready: !!client && !!sessionId,
      submitAnswer,
      requestHint,
      pauseSession,
      resumeSession,
    }),
    [client, sessionId, submitAnswer, requestHint, pauseSession, resumeSession],
  );
}

function stableFingerprint(
  kind: EventKind,
  payload: Record<string, unknown>,
): string {
  const keys = Object.keys(payload).sort();
  const orderedPairs = keys.map((k) => `${k}=${JSON.stringify(payload[k])}`);
  return `${kind}::${orderedPairs.join("&")}`;
}
