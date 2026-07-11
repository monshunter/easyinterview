/**
 * @vitest-environment jsdom
 */

import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type {
  PracticeSessionEventRequest,
  PracticeVoiceTurnResult,
} from "../../../../api/generated/types";
import {
  AppRuntimeContext,
  type AppRuntimeValue,
} from "../../../runtime/AppRuntimeProvider";
import {
  type PracticeVoiceInterruptionReason,
  usePracticeVoicePlayback,
} from "./usePracticeVoicePlayback";

const SESSION_ID = "session-voice";
const VOICE_TURN_ID = "voice-turn-1";
const ASSISTANT_TEXT = "abcdefghij";

class FakeAudioElement {
  static instances: FakeAudioElement[] = [];

  currentTime = 0;
  onended: (() => void) | null = null;
  onerror: (() => void) | null = null;
  paused = true;
  pause = vi.fn(() => {
    this.paused = true;
  });
  play = vi.fn(async () => {
    this.paused = false;
  });

  constructor(readonly src: string) {
    FakeAudioElement.instances.push(this);
  }

  finish(): void {
    this.paused = true;
    this.onended?.();
  }
}

function voiceResult(voiceTurnId = VOICE_TURN_ID): PracticeVoiceTurnResult {
  return {
    voiceTurnId,
    assistantTextDraft: ASSISTANT_TEXT,
    ttsChunks: [
      {
        audioRef: "data:audio/mpeg;base64,VEVTVA==",
        byteLength: 4,
        chunkId: `chunk-${voiceTurnId}`,
        contentType: "audio/mpeg",
        durationMs: 4_000,
        sequence: 0,
        textHash: `sha256:${voiceTurnId}`,
      },
    ],
  } as PracticeVoiceTurnResult;
}

function buildClient() {
  const events: PracticeSessionEventRequest[] = [];
  const appendSessionEvent = vi.fn(
    async (_sessionId: string, body: PracticeSessionEventRequest) => {
      events.push(body);
      return { acknowledged: true } as never;
    },
  );
  const client = { appendSessionEvent } as unknown as EasyInterviewClient;
  return { client, events };
}

function runtimeWrapper(client: EasyInterviewClient) {
  const value: AppRuntimeValue = {
    client,
    runtime: { status: "loading" },
    auth: { status: "unauthenticated" },
    refreshAuth: vi.fn(),
  };

  return function RuntimeWrapper({ children }: { children: ReactNode }) {
    return (
      <AppRuntimeContext.Provider value={value}>
        {children}
      </AppRuntimeContext.Provider>
    );
  };
}

function eventKinds(events: PracticeSessionEventRequest[]) {
  return events.map((event) => event.kind);
}

describe("usePracticeVoicePlayback interrupt lifecycle", () => {
  beforeEach(() => {
    FakeAudioElement.instances = [];
    vi.stubGlobal("Audio", FakeAudioElement);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("pauses synchronously and reports played prefix before user-speech barge-in", async () => {
    const { client, events } = buildClient();
    const initialResult = voiceResult();
    const { result, rerender } = renderHook(
      ({ playbackResult }: { playbackResult: PracticeVoiceTurnResult | null }) =>
        usePracticeVoicePlayback({
          sessionId: SESSION_ID,
          result: playbackResult,
          enabled: true,
        }),
      {
        initialProps: { playbackResult: initialResult },
        wrapper: runtimeWrapper(client),
      },
    );

    await waitFor(() => expect(result.current.state.kind).toBe("playing"));
    const audio = FakeAudioElement.instances[0]!;
    audio.currentTime = 1;

    let interruption!: Promise<boolean>;
    act(() => {
      interruption = result.current.interrupt("user_speech");
    });
    expect(audio.pause).toHaveBeenCalledTimes(1);
    expect(audio.paused).toBe(true);

    await act(async () => {
      await expect(interruption).resolves.toBe(true);
    });
    expect(eventKinds(events)).toEqual([
      "tts_chunk_started",
      "tts_chunk_played",
      "barge_in_detected",
    ]);
    expect(events[1]?.payload).toEqual(
      expect.objectContaining({
        playbackOffsetMs: 1_000,
        playedTextLength: 3,
      }),
    );
    expect(result.current.state.kind).toBe("interrupted");

    audio.finish();
    rerender({ playbackResult: { ...initialResult } });
    await act(async () => Promise.resolve());
    expect(FakeAudioElement.instances).toHaveLength(1);
    expect(eventKinds(events)).toEqual([
      "tts_chunk_started",
      "tts_chunk_played",
      "barge_in_detected",
    ]);
  });

  it("commits the played prefix on mode switch without emitting barge-in", async () => {
    const { client, events } = buildClient();
    const playbackResult = voiceResult();
    const { result } = renderHook(
      () =>
        usePracticeVoicePlayback({
          sessionId: SESSION_ID,
          result: playbackResult,
          enabled: true,
        }),
      { wrapper: runtimeWrapper(client) },
    );

    await waitFor(() => expect(result.current.state.kind).toBe("playing"));
    const audio = FakeAudioElement.instances[0]!;
    audio.currentTime = 2;

    let interruption!: Promise<boolean>;
    act(() => {
      interruption = result.current.interrupt("mode_switch");
    });
    expect(audio.pause).toHaveBeenCalledTimes(1);

    await act(async () => {
      await expect(interruption).resolves.toBe(true);
    });
    expect(eventKinds(events)).toEqual([
      "tts_chunk_started",
      "tts_chunk_played",
      "assistant_context_committed",
    ]);
    expect(eventKinds(events)).not.toContain("barge_in_detected");
    expect(events[2]?.payload).toEqual(
      expect.objectContaining({
        playbackOffsetMs: 2_000,
        committedTextLength: 5,
      }),
    );

    audio.finish();
    await act(async () => Promise.resolve());
    expect(eventKinds(events)).toEqual([
      "tts_chunk_started",
      "tts_chunk_played",
      "assistant_context_committed",
    ]);
  });

  it("suppresses a result received while disabled even if playback is enabled later", async () => {
    const { client, events } = buildClient();
    const lateResult = voiceResult();
    const { rerender } = renderHook(
      ({ enabled, playbackResult }) =>
        usePracticeVoicePlayback({
          sessionId: SESSION_ID,
          result: playbackResult,
          enabled,
        }),
      {
        initialProps: {
          enabled: false,
          playbackResult: lateResult,
        },
        wrapper: runtimeWrapper(client),
      },
    );

    expect(FakeAudioElement.instances).toHaveLength(0);
    expect(events).toHaveLength(0);

    rerender({ enabled: true, playbackResult: lateResult });
    await act(async () => Promise.resolve());
    expect(FakeAudioElement.instances).toHaveLength(0);
    expect(events).toHaveLength(0);

    rerender({ enabled: true, playbackResult: voiceResult("voice-turn-2") });
    await waitFor(() => expect(FakeAudioElement.instances).toHaveLength(1));
    await waitFor(() =>
      expect(eventKinds(events)).toContain("tts_chunk_started"),
    );

    rerender({ enabled: true, playbackResult: lateResult });
    await act(async () => Promise.resolve());
    expect(FakeAudioElement.instances).toHaveLength(1);
  });

  it("stops active playback when disabled and does not resume the same result", async () => {
    const { client, events } = buildClient();
    const playbackResult = voiceResult();
    const { rerender } = renderHook(
      ({ enabled }: { enabled: boolean }) =>
        usePracticeVoicePlayback({
          sessionId: SESSION_ID,
          result: playbackResult,
          enabled,
        }),
      {
        initialProps: { enabled: true },
        wrapper: runtimeWrapper(client),
      },
    );

    await waitFor(() => expect(FakeAudioElement.instances).toHaveLength(1));
    const audio = FakeAudioElement.instances[0]!;
    await waitFor(() => expect(audio.paused).toBe(false));

    rerender({ enabled: false });
    expect(audio.pause).toHaveBeenCalledTimes(1);
    expect(audio.paused).toBe(true);
    audio.finish();
    await act(async () => Promise.resolve());
    expect(eventKinds(events)).toEqual(["tts_chunk_started"]);

    rerender({ enabled: true });
    await act(async () => Promise.resolve());
    expect(FakeAudioElement.instances).toHaveLength(1);
    expect(eventKinds(events)).toEqual(["tts_chunk_started"]);
  });

  it.each<{
    reason: PracticeVoiceInterruptionReason;
    expectedKinds: PracticeSessionEventRequest["kind"][];
  }>([
    {
      reason: "user_speech",
      expectedKinds: ["tts_chunk_started", "barge_in_detected"],
    },
    {
      reason: "mode_switch",
      expectedKinds: ["tts_chunk_started"],
    },
  ])(
    "does not send an invalid zero-length played prefix for $reason",
    async ({ reason, expectedKinds }) => {
      const { client, events } = buildClient();
      const playbackResult = voiceResult();
      const { result } = renderHook(
        () =>
          usePracticeVoicePlayback({
            sessionId: SESSION_ID,
            result: playbackResult,
            enabled: true,
          }),
        { wrapper: runtimeWrapper(client) },
      );

      await waitFor(() => expect(result.current.state.kind).toBe("playing"));
      const audio = FakeAudioElement.instances[0]!;
      expect(audio.currentTime).toBe(0);

      let interruption!: Promise<boolean>;
      act(() => {
        interruption = result.current.interrupt(reason);
      });
      expect(audio.pause).toHaveBeenCalledTimes(1);
      await act(async () => {
        await expect(interruption).resolves.toBe(true);
      });

      expect(eventKinds(events)).toEqual(expectedKinds);
      expect(eventKinds(events)).not.toContain("tts_chunk_played");
      expect(eventKinds(events)).not.toContain("assistant_context_committed");
      if (reason === "user_speech") {
        expect(events.at(-1)?.payload).toEqual(
          expect.objectContaining({ playbackOffsetMs: 0 }),
        );
      } else {
        expect(eventKinds(events)).not.toContain("barge_in_detected");
      }
    },
  );

  it("rejects interruption when a played-prefix event cannot be persisted", async () => {
    const events: PracticeSessionEventRequest[] = [];
    const appendSessionEvent = vi.fn(
      async (_sessionId: string, body: PracticeSessionEventRequest) => {
        events.push(body);
        if (body.kind === "tts_chunk_played") {
          throw new Error("delayed append failed");
        }
        return { acknowledged: true } as never;
      },
    );
    const client = { appendSessionEvent } as unknown as EasyInterviewClient;
    const playbackResult = voiceResult();
    const { result } = renderHook(
      () =>
        usePracticeVoicePlayback({
          sessionId: SESSION_ID,
          result: playbackResult,
          enabled: true,
        }),
      { wrapper: runtimeWrapper(client) },
    );

    await waitFor(() => expect(result.current.state.kind).toBe("playing"));
    FakeAudioElement.instances[0]!.currentTime = 1;

    await act(async () => {
      await expect(result.current.interrupt("user_speech")).rejects.toThrow(
        "delayed append failed",
      );
    });
    expect(eventKinds(events)).toEqual([
      "tts_chunk_started",
      "tts_chunk_played",
    ]);
    expect(result.current.state).toEqual({
      kind: "error",
      message: "delayed append failed",
    });
  });
});
