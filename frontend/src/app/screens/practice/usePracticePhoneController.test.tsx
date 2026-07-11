/**
 * @vitest-environment jsdom
 */

import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { PracticeVoiceTurnResult } from "../../../api/generated/types";
import type {
  PhoneVadMonitor,
  PhoneVadMonitorOptions,
} from "./phoneVadMonitor";
import { usePracticePhoneController } from "./usePracticePhoneController";

const STREAM = {} as MediaStream;

function voiceResult(): PracticeVoiceTurnResult {
  return {
    voiceTurnId: "voice-turn-1",
    userTranscriptFinal: "我的回答",
    assistantTextDraft: "请继续说明影响。",
    ttsChunks: [{ chunkId: "chunk-1" }],
    session: {
      id: "session-1",
      status: "running",
    },
  } as PracticeVoiceTurnResult;
}

describe("usePracticePhoneController", () => {
  it("connects once, VAD-submits, then rearms recording after TTS completes", async () => {
    let vadOptions: PhoneVadMonitorOptions | null = null;
    const monitor: PhoneVadMonitor = {
      stop: vi.fn(),
      rearm: vi.fn(),
      hasSpeech: vi.fn(() => true),
    };
    const createVadMonitor = vi.fn(
      (_stream: MediaStream, options: PhoneVadMonitorOptions) => {
        vadOptions = options;
        return monitor;
      },
    );
    const resultValue = voiceResult();
    const connectMicrophone = vi.fn(async () => undefined);
    const startRecording = vi.fn(async () => undefined);
    const stopAndSubmit = vi.fn(async () => resultValue);
    const onTurnResult = vi.fn();
    const discardUtterance = vi.fn();
    const releaseMicrophone = vi.fn();
    const interrupt = vi.fn(async () => false);
    const getMediaStream = () => STREAM;

    const { rerender } = renderHook(
      ({ captureKind, playbackKind }) =>
        usePracticePhoneController({
          enabled: true,
          active: true,
          voiceTurn: {
            ready: true,
            state: { kind: captureKind } as never,
            connectMicrophone,
            getMediaStream,
            startRecording,
            stopAndSubmit,
            discardUtterance,
            releaseMicrophone,
          },
          voicePlayback: {
            state: { kind: playbackKind } as never,
            interrupt,
          },
          onTurnResult,
          onError: vi.fn(),
          createVadMonitor,
        }),
      {
        initialProps: {
          captureKind: "idle",
          playbackKind: "idle",
        },
      },
    );

    await waitFor(() => expect(connectMicrophone).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(startRecording).toHaveBeenCalledTimes(1));
    expect(createVadMonitor).toHaveBeenCalledWith(STREAM, expect.any(Object));

    rerender({ captureKind: "recording", playbackKind: "idle" });
    await act(async () => {
      vadOptions!.onSilenceSubmit();
      await Promise.resolve();
    });
    await waitFor(() => expect(stopAndSubmit).toHaveBeenCalledTimes(1));
    expect(onTurnResult).toHaveBeenCalledWith(resultValue);

    rerender({ captureKind: "success", playbackKind: "completed" });
    await waitFor(() => expect(startRecording).toHaveBeenCalledTimes(2));
    expect(connectMicrophone).toHaveBeenCalledTimes(1);
    expect(monitor.rearm).toHaveBeenCalled();
  });

  it("does not rearm recording when the submitted voice turn completes the session", async () => {
    let vadOptions: PhoneVadMonitorOptions | null = null;
    const runningResult = voiceResult();
    const completedResult: PracticeVoiceTurnResult = {
      ...runningResult,
      assistantTextDraft: "",
      ttsChunks: [],
      session: {
        ...runningResult.session,
        status: "completed",
        currentTurn: null,
      },
    };
    const startRecording = vi.fn(async () => undefined);
    const stopAndSubmit = vi.fn(async () => completedResult);
    const onTurnResult = vi.fn();

    renderHook(() =>
      usePracticePhoneController({
        enabled: true,
        active: true,
        voiceTurn: {
          ready: true,
          state: { kind: "recording" },
          connectMicrophone: vi.fn(async () => undefined),
          getMediaStream: () => STREAM,
          startRecording,
          stopAndSubmit,
          discardUtterance: vi.fn(),
          releaseMicrophone: vi.fn(),
        },
        voicePlayback: {
          state: { kind: "idle" },
          interrupt: vi.fn(async () => false),
        },
        onTurnResult,
        onError: vi.fn(),
        createVadMonitor: (_stream, options) => {
          vadOptions = options;
          return {
            stop: vi.fn(),
            rearm: vi.fn(),
            hasSpeech: vi.fn(() => true),
          };
        },
      }),
    );

    await waitFor(() => expect(vadOptions).not.toBeNull());
    act(() => vadOptions!.onSilenceSubmit());

    await waitFor(() =>
      expect(onTurnResult).toHaveBeenCalledWith(completedResult),
    );
    expect(startRecording).not.toHaveBeenCalled();
  });

  it("only speech-start barge-ins while exit uses mode_switch and releases resources", async () => {
    let vadOptions: PhoneVadMonitorOptions | null = null;
    const monitor: PhoneVadMonitor = {
      stop: vi.fn(),
      rearm: vi.fn(),
      hasSpeech: vi.fn(() => false),
    };
    const interrupt = vi.fn(async () => true);
    const startRecording = vi.fn(async () => undefined);
    const releaseMicrophone = vi.fn();

    const { result } = renderHook(() =>
      usePracticePhoneController({
        enabled: true,
        active: true,
        voiceTurn: {
          ready: true,
          state: { kind: "success", result: voiceResult() },
          connectMicrophone: vi.fn(async () => undefined),
          getMediaStream: () => STREAM,
          startRecording,
          stopAndSubmit: vi.fn(async () => voiceResult()),
          discardUtterance: vi.fn(),
          releaseMicrophone,
        },
        voicePlayback: {
          state: { kind: "playing", chunkId: "chunk-1" },
          interrupt,
        },
        onTurnResult: vi.fn(),
        onError: vi.fn(),
        createVadMonitor: (_stream, options) => {
          vadOptions = options;
          return monitor;
        },
      }),
    );

    await waitFor(() => expect(vadOptions).not.toBeNull());
    act(() => vadOptions!.onSpeechStart());
    expect(interrupt).toHaveBeenCalledWith("user_speech");
    await waitFor(() => expect(startRecording).toHaveBeenCalled());

    act(() => result.current.exitPhoneMode());
    expect(monitor.stop).toHaveBeenCalledTimes(1);
    expect(releaseMicrophone).toHaveBeenCalledTimes(1);
    expect(interrupt).toHaveBeenCalledWith("mode_switch");
  });

  it("waits for delayed played-prefix and barge-in events before submitting the interrupted answer", async () => {
    let vadOptions: PhoneVadMonitorOptions | null = null;
    const interruption = deferred<boolean>();
    const stopAndSubmit = vi.fn(async () => voiceResult());
    const startRecording = vi.fn(async () => undefined);
    const onError = vi.fn();
    const interrupt = vi.fn(() => interruption.promise);

    const { rerender } = renderHook(
      ({ captureKind, playbackKind }) =>
        usePracticePhoneController({
          enabled: true,
          active: true,
          voiceTurn: {
            ready: true,
            state: { kind: captureKind } as never,
            connectMicrophone: vi.fn(async () => undefined),
            getMediaStream: () => STREAM,
            startRecording,
            stopAndSubmit,
            discardUtterance: vi.fn(),
            releaseMicrophone: vi.fn(),
          },
          voicePlayback: {
            state: { kind: playbackKind } as never,
            interrupt,
          },
          onTurnResult: vi.fn(),
          onError,
          createVadMonitor: (_stream, options) => {
            vadOptions = options;
            return {
              stop: vi.fn(),
              rearm: vi.fn(),
              hasSpeech: vi.fn(() => true),
            };
          },
        }),
      {
        initialProps: {
          captureKind: "submitting",
          playbackKind: "idle",
        },
      },
    );

    await waitFor(() => expect(vadOptions).not.toBeNull());
    rerender({ captureKind: "success", playbackKind: "playing" });
    act(() => vadOptions!.onSpeechStart());
    await waitFor(() => expect(startRecording).toHaveBeenCalledTimes(1));

    rerender({ captureKind: "recording", playbackKind: "interrupted" });
    act(() => vadOptions!.onSilenceSubmit());
    await act(async () => Promise.resolve());
    expect(stopAndSubmit).not.toHaveBeenCalled();

    await act(async () => {
      interruption.resolve(true);
      await interruption.promise;
    });
    await waitFor(() => expect(stopAndSubmit).toHaveBeenCalledTimes(1));
    expect(onError).not.toHaveBeenCalled();
  });

  it("does not submit and reports an error when the barge-in event barrier fails", async () => {
    let vadOptions: PhoneVadMonitorOptions | null = null;
    const interruption = deferred<boolean>();
    const stopAndSubmit = vi.fn(async () => voiceResult());
    const discardUtterance = vi.fn();
    const onError = vi.fn();

    const { rerender } = renderHook(
      ({ captureKind, playbackKind }) =>
        usePracticePhoneController({
          enabled: true,
          active: true,
          voiceTurn: {
            ready: true,
            state: { kind: captureKind } as never,
            connectMicrophone: vi.fn(async () => undefined),
            getMediaStream: () => STREAM,
            startRecording: vi.fn(async () => undefined),
            stopAndSubmit,
            discardUtterance,
            releaseMicrophone: vi.fn(),
          },
          voicePlayback: {
            state: { kind: playbackKind } as never,
            interrupt: vi.fn(() => interruption.promise),
          },
          onTurnResult: vi.fn(),
          onError,
          createVadMonitor: (_stream, options) => {
            vadOptions = options;
            return {
              stop: vi.fn(),
              rearm: vi.fn(),
              hasSpeech: vi.fn(() => true),
            };
          },
        }),
      {
        initialProps: {
          captureKind: "submitting",
          playbackKind: "idle",
        },
      },
    );

    await waitFor(() => expect(vadOptions).not.toBeNull());
    rerender({ captureKind: "success", playbackKind: "playing" });
    act(() => vadOptions!.onSpeechStart());
    rerender({ captureKind: "recording", playbackKind: "interrupted" });
    act(() => vadOptions!.onSilenceSubmit());

    await act(async () => {
      interruption.resolve(false);
      await interruption.promise;
    });
    await waitFor(() => expect(onError).toHaveBeenCalledTimes(1));
    expect(stopAndSubmit).not.toHaveBeenCalled();
    expect(discardUtterance).toHaveBeenCalledTimes(1);
  });

  it("settles a spoken in-flight utterance on hang-up but discards an empty one", async () => {
    const spokenSubmit = vi.fn(async () => voiceResult());
    const spokenResult = vi.fn();
    const spokenRelease = vi.fn();
    const { result: spokenController } = renderHook(() =>
      usePracticePhoneController({
        enabled: true,
        active: true,
        voiceTurn: {
          ready: true,
          state: { kind: "recording" },
          connectMicrophone: vi.fn(async () => undefined),
          getMediaStream: () => STREAM,
          startRecording: vi.fn(async () => undefined),
          stopAndSubmit: spokenSubmit,
          discardUtterance: vi.fn(),
          releaseMicrophone: spokenRelease,
        },
        voicePlayback: {
          state: { kind: "idle" },
          interrupt: vi.fn(async () => false),
        },
        onTurnResult: spokenResult,
        onError: vi.fn(),
        createVadMonitor: () => ({
          stop: vi.fn(),
          rearm: vi.fn(),
          hasSpeech: vi.fn(() => true),
        }),
      }),
    );

    await waitFor(() => expect(spokenSubmit).not.toHaveBeenCalled());
    act(() => spokenController.current.exitPhoneMode());
    await waitFor(() => expect(spokenSubmit).toHaveBeenCalledTimes(1));
    expect(spokenSubmit).toHaveBeenCalledWith({
      releaseMicrophoneAfterSnapshot: true,
    });
    await waitFor(() => expect(spokenResult).toHaveBeenCalledWith(voiceResult()));
    expect(spokenRelease).not.toHaveBeenCalled();

    const emptySubmit = vi.fn(async () => voiceResult());
    const emptyDiscard = vi.fn();
    const emptyRelease = vi.fn();
    const { result: emptyController } = renderHook(() =>
      usePracticePhoneController({
        enabled: true,
        active: true,
        voiceTurn: {
          ready: true,
          state: { kind: "recording" },
          connectMicrophone: vi.fn(async () => undefined),
          getMediaStream: () => STREAM,
          startRecording: vi.fn(async () => undefined),
          stopAndSubmit: emptySubmit,
          discardUtterance: emptyDiscard,
          releaseMicrophone: emptyRelease,
        },
        voicePlayback: {
          state: { kind: "idle" },
          interrupt: vi.fn(async () => false),
        },
        onTurnResult: vi.fn(),
        onError: vi.fn(),
        createVadMonitor: () => ({
          stop: vi.fn(),
          rearm: vi.fn(),
          hasSpeech: vi.fn(() => false),
        }),
      }),
    );

    act(() => emptyController.current.exitPhoneMode());
    expect(emptySubmit).not.toHaveBeenCalled();
    expect(emptyDiscard).toHaveBeenCalledTimes(1);
    expect(emptyRelease).toHaveBeenCalledTimes(1);
  });
});

interface Deferred<T> {
  promise: Promise<T>;
  resolve: (value: T) => void;
}

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((onResolve) => {
    resolve = onResolve;
  });
  return { promise, resolve };
}
