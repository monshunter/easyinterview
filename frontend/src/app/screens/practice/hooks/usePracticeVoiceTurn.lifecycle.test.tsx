// @vitest-environment jsdom

import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type { PracticeVoiceTurnResult } from "../../../../api/generated/types";
import {
  AppRuntimeContext,
  type AppRuntimeValue,
} from "../../../runtime/AppRuntimeProvider";
import { usePracticeVoiceTurn } from "./usePracticeVoiceTurn";

class FakeMediaRecorder {
  static instances: FakeMediaRecorder[] = [];

  static isTypeSupported(mimeType: string): boolean {
    return mimeType === "audio/webm";
  }

  readonly mimeType: string;
  readonly stream: MediaStream;
  state: RecordingState = "inactive";
  ondataavailable: ((event: BlobEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onstop: ((event: Event) => void) | null = null;

  constructor(stream: MediaStream, options?: MediaRecorderOptions) {
    this.stream = stream;
    this.mimeType = options?.mimeType ?? "audio/webm";
    FakeMediaRecorder.instances.push(this);
  }

  start(): void {
    this.state = "recording";
  }

  stop(): void {
    if (this.state === "inactive") return;
    this.state = "inactive";
    const data = new Blob(["OggS"], { type: this.mimeType });
    this.ondataavailable?.({ data } as BlobEvent);
    this.onstop?.(new Event("stop"));
  }
}

const originalMediaDevicesDescriptor = Object.getOwnPropertyDescriptor(
  navigator,
  "mediaDevices",
);

describe("usePracticeVoiceTurn microphone lifecycle", () => {
  let trackStop: ReturnType<typeof vi.fn>;
  let stream: MediaStream;
  let getUserMedia: ReturnType<typeof vi.fn>;
  let createPracticeVoiceTurn: ReturnType<typeof vi.fn>;
  let client: EasyInterviewClient;

  beforeEach(() => {
    FakeMediaRecorder.instances = [];
    trackStop = vi.fn();
    stream = {
      getTracks: () => [{ stop: trackStop } as unknown as MediaStreamTrack],
    } as unknown as MediaStream;
    getUserMedia = vi.fn().mockResolvedValue(stream);
    Object.defineProperty(navigator, "mediaDevices", {
      configurable: true,
      value: { getUserMedia },
    });
    vi.stubGlobal(
      "MediaRecorder",
      FakeMediaRecorder as unknown as typeof MediaRecorder,
    );

    createPracticeVoiceTurn = vi
      .fn()
      .mockResolvedValue({ voiceTurnId: "voice-turn-1" } as PracticeVoiceTurnResult);
    client = { createPracticeVoiceTurn } as unknown as EasyInterviewClient;
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    if (originalMediaDevicesDescriptor) {
      Object.defineProperty(
        navigator,
        "mediaDevices",
        originalMediaDevicesDescriptor,
      );
    } else {
      Reflect.deleteProperty(navigator, "mediaDevices");
    }
  });

  it("connects once and reuses the same stream for separate utterances", async () => {
    const { result } = renderVoiceTurnHook(client);

    await act(async () => {
      await Promise.all([
        result.current.connectMicrophone(),
        result.current.connectMicrophone(),
      ]);
    });

    expect(getUserMedia).toHaveBeenCalledTimes(1);
    expect(result.current.getMediaStream()).toBe(stream);

    await act(async () => {
      await result.current.startRecording();
    });
    const firstRecorder = FakeMediaRecorder.instances[0];
    expect(firstRecorder?.stream).toBe(stream);

    await act(async () => {
      await result.current.stopAndSubmit();
    });
    expect(trackStop).not.toHaveBeenCalled();

    await act(async () => {
      await result.current.startRecording();
    });
    const secondRecorder = FakeMediaRecorder.instances[1];
    expect(secondRecorder).not.toBe(firstRecorder);
    expect(secondRecorder?.stream).toBe(stream);
    expect(getUserMedia).toHaveBeenCalledTimes(1);

    await act(async () => {
      await result.current.stopAndSubmit();
    });
    expect(createPracticeVoiceTurn).toHaveBeenCalledTimes(2);
    expect(trackStop).not.toHaveBeenCalled();

    act(() => {
      result.current.releaseMicrophone();
    });
    expect(trackStop).toHaveBeenCalledTimes(1);
    expect(result.current.getMediaStream()).toBeNull();
  });

  it("discards only the current utterance and keeps the microphone connected", async () => {
    const { result } = renderVoiceTurnHook(client);

    await act(async () => {
      await result.current.connectMicrophone();
      await result.current.startRecording();
    });
    const discardedRecorder = FakeMediaRecorder.instances[0];

    act(() => {
      result.current.discardUtterance();
    });

    expect(discardedRecorder?.state).toBe("inactive");
    expect(result.current.state.kind).toBe("idle");
    expect(createPracticeVoiceTurn).not.toHaveBeenCalled();
    expect(trackStop).not.toHaveBeenCalled();

    await act(async () => {
      await result.current.startRecording();
    });
    expect(FakeMediaRecorder.instances[1]?.stream).toBe(stream);
    expect(getUserMedia).toHaveBeenCalledTimes(1);
  });

  it("requires an explicit connection before starting an utterance", async () => {
    const { result } = renderVoiceTurnHook(client);

    await act(async () => {
      await result.current.startRecording();
    });

    expect(getUserMedia).not.toHaveBeenCalled();
    expect(FakeMediaRecorder.instances).toHaveLength(0);
    expect(result.current.state).toEqual({
      kind: "error",
      message: "microphone is not connected",
    });
  });

  it("releases the recorder and tracks on unmount", async () => {
    const { result, unmount } = renderVoiceTurnHook(client);

    await act(async () => {
      await result.current.connectMicrophone();
      await result.current.startRecording();
    });
    const recorder = FakeMediaRecorder.instances[0];

    unmount();

    expect(recorder?.state).toBe("inactive");
    expect(trackStop).toHaveBeenCalledTimes(1);
    expect(createPracticeVoiceTurn).not.toHaveBeenCalled();
  });

  it("keeps reset as a release-compatible alias", async () => {
    const { result } = renderVoiceTurnHook(client);

    await act(async () => {
      await result.current.connectMicrophone();
    });
    act(() => {
      result.current.reset();
    });

    expect(trackStop).toHaveBeenCalledTimes(1);
    expect(result.current.state.kind).toBe("idle");
  });

  it("snapshots a hang-up utterance before releasing tracks and still returns the late result", async () => {
    const response = deferred<PracticeVoiceTurnResult>();
    createPracticeVoiceTurn.mockReturnValue(response.promise);
    const { result } = renderVoiceTurnHook(client);

    await act(async () => {
      await result.current.connectMicrophone();
      await result.current.startRecording();
    });

    let settlement!: Promise<PracticeVoiceTurnResult>;
    act(() => {
      settlement = result.current.stopAndSubmit({
        releaseMicrophoneAfterSnapshot: true,
      });
    });

    await waitFor(() => expect(createPracticeVoiceTurn).toHaveBeenCalledTimes(1));
    expect(trackStop).toHaveBeenCalledTimes(1);
    expect(result.current.getMediaStream()).toBeNull();
    expect(result.current.state.kind).toBe("idle");

    const lateResult = {
      voiceTurnId: "voice-turn-late",
    } as PracticeVoiceTurnResult;
    await act(async () => {
      response.resolve(lateResult);
      await expect(settlement).resolves.toBe(lateResult);
    });
    expect(result.current.state.kind).toBe("idle");
  });
});

function renderVoiceTurnHook(client: EasyInterviewClient) {
  const runtimeValue: AppRuntimeValue = {
    client,
    runtime: { status: "ready", config: {} as never },
    auth: { status: "unauthenticated" },
    refreshAuth: () => {},
  };
  const wrapper = ({ children }: { children: ReactNode }) => (
    <AppRuntimeContext.Provider value={runtimeValue}>
      {children}
    </AppRuntimeContext.Provider>
  );

  return renderHook(
    () =>
      usePracticeVoiceTurn({
        sessionId: "session-1",
        turnId: "turn-1",
        lang: "zh",
        practiceMode: "assisted",
      }),
    { wrapper },
  );
}

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
