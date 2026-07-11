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
  /** False outside phone mode. Disabled playback also suppresses late results. */
  enabled: boolean;
}

export type PracticeVoiceInterruptionReason = "user_speech" | "mode_switch";

export interface UsePracticeVoicePlaybackResult {
  state: PracticeVoicePlaybackState;
  interrupt: (reason: PracticeVoiceInterruptionReason) => Promise<boolean>;
  /** Compatibility alias while existing callers migrate to interrupt(reason). */
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
  enabled = true,
}: UsePracticeVoicePlaybackInput): UsePracticeVoicePlaybackResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const [state, setState] = useState<PracticeVoicePlaybackState>({
    kind: "idle",
  });
  const activeRef = useRef<ActivePlayback | null>(null);
  const enabledRef = useRef(enabled);
  const handledVoiceTurnKeysRef = useRef<Set<string>>(new Set());
  const tokenRef = useRef(0);
  enabledRef.current = enabled;

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
      if (
        !enabledRef.current ||
        !active ||
        tokenRef.current !== token ||
        active.token !== token
      ) {
        return;
      }
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
        if (!enabledRef.current || tokenRef.current !== token) return;
        await sendEvent("assistant_context_committed", {
          ...payloadBase,
          committedTextHash: active.chunk.textHash,
          committedTextLength: active.assistantText.length,
        });
        if (!enabledRef.current || tokenRef.current !== token) return;
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
        if (!enabledRef.current || tokenRef.current !== token) return;
        await audio.play();
        if (!enabledRef.current || tokenRef.current !== token) {
          audio.pause();
          if (activeRef.current?.token === token) activeRef.current = null;
          return;
        }
        setState({ kind: "playing", chunkId: chunk.chunkId });
      } catch (err) {
        if (tokenRef.current !== token) return;
        activeRef.current = null;
        setState({ kind: "error", message: errorMessage(err) });
      }
    },
    [completePlayback, sendEvent],
  );

  const interrupt = useCallback(
    async (reason: PracticeVoiceInterruptionReason): Promise<boolean> => {
      const active = activeRef.current;
      if (!active) return false;
      tokenRef.current += 1;
      active.audio.pause();
      activeRef.current = null;
      const playbackOffsetMs = currentPlaybackOffsetMs(active);
      setState({ kind: "interrupted", chunkId: active.chunk.chunkId });
      try {
        const playedTextLength = estimatePlayedTextLength(
          active.assistantText,
          active.chunk.durationMs,
          playbackOffsetMs,
        );
        if (playedTextLength > 0) {
          await sendEvent("tts_chunk_played", {
            voiceTurnId: active.voiceTurnId,
            chunkId: active.chunk.chunkId,
            playbackOffsetMs,
            playedTextHash: active.chunk.textHash,
            playedTextLength,
          });
        }
        if (reason === "user_speech") {
          await sendEvent("barge_in_detected", {
            voiceTurnId: active.voiceTurnId,
            chunkId: active.chunk.chunkId,
            playbackOffsetMs,
            userSpeechStartedAt: new Date().toISOString(),
          });
        } else if (playedTextLength > 0) {
          await sendEvent("assistant_context_committed", {
            voiceTurnId: active.voiceTurnId,
            chunkId: active.chunk.chunkId,
            playbackOffsetMs,
            committedTextHash: active.chunk.textHash,
            committedTextLength: playedTextLength,
          });
        }
      } catch (err) {
        setState({ kind: "error", message: errorMessage(err) });
        throw err instanceof Error ? err : new Error(errorMessage(err));
      }
      return true;
    },
    [sendEvent],
  );

  const bargeIn = useCallback(
    () => interrupt("user_speech"),
    [interrupt],
  );

  useEffect(() => {
    const firstChunk = result?.ttsChunks[0] ?? null;
    if (!enabled) {
      if (result) {
        handledVoiceTurnKeysRef.current.add(
          voiceTurnKey(sessionId, result.voiceTurnId),
        );
      }
      stopActivePlayback();
      setState({ kind: "idle" });
      return;
    }
    if (!client || !sessionId || !result || !firstChunk) {
      if (result && !firstChunk) {
        setState({ kind: "idle" });
      }
      if (!result) {
        setState({ kind: "idle" });
      }
      return;
    }
    const key = voiceTurnKey(sessionId, result.voiceTurnId);
    if (handledVoiceTurnKeysRef.current.has(key)) return;
    handledVoiceTurnKeysRef.current.add(key);
    void playChunk(result.voiceTurnId, firstChunk, result.assistantTextDraft);
    return () => {
      stopActivePlayback();
    };
  }, [client, enabled, playChunk, result, sessionId, stopActivePlayback]);

  return useMemo<UsePracticeVoicePlaybackResult>(
    () => ({
      state,
      interrupt,
      bargeIn,
    }),
    [bargeIn, interrupt, state],
  );
}

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

function voiceTurnKey(sessionId: string, voiceTurnId: string): string {
  return `${sessionId}\u0000${voiceTurnId}`;
}

function currentPlaybackOffsetMs(active: ActivePlayback): number {
  const mediaOffsetMs = Number.isFinite(active.audio.currentTime)
    ? Math.round(active.audio.currentTime * 1000)
    : Date.now() - active.startedAtMs;
  return Math.max(0, Math.min(active.chunk.durationMs, mediaOffsetMs));
}

function estimatePlayedTextLength(
  assistantText: string,
  durationMs: number,
  playbackOffsetMs: number,
): number {
  const totalLength = Array.from(assistantText).length;
  if (totalLength === 0 || durationMs <= 0 || playbackOffsetMs <= 0) return 0;
  const ratio = Math.min(1, playbackOffsetMs / durationMs);
  return Math.max(1, Math.min(totalLength, Math.round(totalLength * ratio)));
}
