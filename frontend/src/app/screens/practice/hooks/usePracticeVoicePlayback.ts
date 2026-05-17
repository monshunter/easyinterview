import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import type {
  PracticeSessionEventRequest,
  PracticeVoiceTTSChunk,
  PracticeVoiceTurnResult,
} from "../../../../api/generated/types";
import { newId } from "../../../../lib/ids";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export type PracticeVoicePlaybackState =
  | { kind: "idle" }
  | { kind: "playing"; chunkId: string }
  | { kind: "completed"; chunkId: string }
  | { kind: "interrupted"; chunkId: string }
  | { kind: "error"; message: string };

export interface UsePracticeVoicePlaybackInput {
  sessionId: string;
  result: PracticeVoiceTurnResult | null;
}

export interface UsePracticeVoicePlaybackResult {
  state: PracticeVoicePlaybackState;
  bargeIn: () => Promise<boolean>;
}

interface ActivePlayback {
  audio: HTMLAudioElement;
  assistantText: string;
  chunk: PracticeVoiceTTSChunk;
  startedAtMs: number;
  token: number;
  voiceTurnId: string;
}

export function usePracticeVoicePlayback({
  sessionId,
  result,
}: UsePracticeVoicePlaybackInput): UsePracticeVoicePlaybackResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const [state, setState] = useState<PracticeVoicePlaybackState>({
    kind: "idle",
  });
  const activeRef = useRef<ActivePlayback | null>(null);
  const tokenRef = useRef(0);
  const startedVoiceTurnRef = useRef<string | null>(null);

  const sendEvent = useCallback(
    async (
      kind: PracticeSessionEventRequest["kind"],
      payload: Record<string, unknown>,
    ) => {
      if (!client) throw new Error("usePracticeVoicePlayback: client missing");
      if (!sessionId) {
        throw new Error("usePracticeVoicePlayback: sessionId missing");
      }
      const body: PracticeSessionEventRequest = {
        clientEventId: newId(),
        kind,
        occurredAt: new Date().toISOString(),
        payload,
      };
      await client.appendSessionEvent(sessionId, body);
    },
    [client, sessionId],
  );

  const stopActivePlayback = useCallback(() => {
    tokenRef.current += 1;
    activeRef.current?.audio.pause();
    activeRef.current = null;
  }, []);

  const completePlayback = useCallback(
    async (token: number) => {
      const active = activeRef.current;
      if (!active || tokenRef.current !== token || active.token !== token) return;
      activeRef.current = null;
      const payloadBase = {
        voiceTurnId: active.voiceTurnId,
        chunkId: active.chunk.chunkId,
        playbackOffsetMs: active.chunk.durationMs,
      };
      try {
        await sendEvent("tts_chunk_played", {
          ...payloadBase,
          playedTextHash: active.chunk.textHash,
          playedTextLength: active.assistantText.length,
        });
        await sendEvent("assistant_context_committed", {
          ...payloadBase,
          committedTextHash: active.chunk.textHash,
          committedTextLength: active.assistantText.length,
        });
        if (tokenRef.current !== token) return;
        setState({ kind: "completed", chunkId: active.chunk.chunkId });
      } catch (err) {
        if (tokenRef.current !== token) return;
        setState({ kind: "error", message: errorMessage(err) });
      }
    },
    [sendEvent],
  );

  const playChunk = useCallback(
    async (
      voiceTurnId: string,
      chunk: PracticeVoiceTTSChunk,
      assistantText: string,
    ) => {
      const token = tokenRef.current + 1;
      tokenRef.current = token;
      const audio = new Audio(chunk.audioRef);
      const active: ActivePlayback = {
        audio,
        assistantText,
        chunk,
        startedAtMs: Date.now(),
        token,
        voiceTurnId,
      };
      activeRef.current = active;
      audio.onended = () => {
        void completePlayback(token);
      };
      audio.onerror = () => {
        if (tokenRef.current !== token) return;
        activeRef.current = null;
        setState({ kind: "error", message: "tts playback failed" });
      };

      try {
        await sendEvent("tts_chunk_started", {
          voiceTurnId,
          chunkId: chunk.chunkId,
          playbackOffsetMs: 0,
        });
        if (tokenRef.current !== token) return;
        await audio.play();
        if (tokenRef.current !== token) return;
        setState({ kind: "playing", chunkId: chunk.chunkId });
      } catch (err) {
        if (tokenRef.current !== token) return;
        activeRef.current = null;
        setState({ kind: "error", message: errorMessage(err) });
      }
    },
    [completePlayback, sendEvent],
  );

  const bargeIn = useCallback(async (): Promise<boolean> => {
    const active = activeRef.current;
    if (!active) return false;
    tokenRef.current += 1;
    active.audio.pause();
    activeRef.current = null;
    const playbackOffsetMs = Math.max(
      0,
      Math.min(
        active.chunk.durationMs,
        Math.round(active.audio.currentTime * 1000) ||
          Date.now() - active.startedAtMs,
      ),
    );
    setState({ kind: "interrupted", chunkId: active.chunk.chunkId });
    try {
      await sendEvent("barge_in_detected", {
        voiceTurnId: active.voiceTurnId,
        chunkId: active.chunk.chunkId,
        playbackOffsetMs,
        userSpeechStartedAt: new Date().toISOString(),
      });
    } catch (err) {
      setState({ kind: "error", message: errorMessage(err) });
    }
    return true;
  }, [sendEvent]);

  useEffect(() => {
    const firstChunk = result?.ttsChunks[0] ?? null;
    if (!client || !sessionId || !result || !firstChunk) {
      if (result && !firstChunk) {
        setState({ kind: "idle" });
      }
      if (!result) {
        startedVoiceTurnRef.current = null;
        setState({ kind: "idle" });
      }
      return;
    }
    if (startedVoiceTurnRef.current === result.voiceTurnId) return;
    startedVoiceTurnRef.current = result.voiceTurnId;
    void playChunk(result.voiceTurnId, firstChunk, result.assistantTextDraft);
    return () => {
      stopActivePlayback();
    };
  }, [client, playChunk, result, sessionId, stopActivePlayback]);

  return useMemo<UsePracticeVoicePlaybackResult>(
    () => ({
      state,
      bargeIn,
    }),
    [bargeIn, state],
  );
}

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}
