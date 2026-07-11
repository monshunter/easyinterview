/**
 * @vitest-environment jsdom
 */

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { startPhoneVadMonitor } from "./phoneVadMonitor";

let frames: FrameRequestCallback[];
let nextSamples: number[];

class FakeAnalyser {
  fftSize = 8;
  smoothingTimeConstant = 0;

  getFloatTimeDomainData(target: Float32Array) {
    const value = nextSamples.shift() ?? 0;
    target.fill(value);
  }

  disconnect() {}
}

class FakeAudioContext {
  createAnalyser() {
    return new FakeAnalyser();
  }

  createMediaStreamSource() {
    return { connect: vi.fn(), disconnect: vi.fn() };
  }

  close() {
    return Promise.resolve();
  }
}

describe("phone VAD browser monitor", () => {
  beforeEach(() => {
    frames = [];
    nextSamples = [];
    vi.stubGlobal("AudioContext", FakeAudioContext);
    vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
      frames.push(callback);
      return frames.length;
    });
    vi.stubGlobal("cancelAnimationFrame", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("confirms speech, submits after silence, and rearms without an echo submit", () => {
    const onSpeechStart = vi.fn();
    const onSilenceSubmit = vi.fn();
    const monitor = startPhoneVadMonitor({} as MediaStream, {
      silenceThresholdMs: 100,
      speechRmsThreshold: 0.1,
      speechStartFrames: 2,
      getPaused: () => false,
      getTtsPlaying: () => false,
      onSpeechStart,
      onSilenceSubmit,
    });
    expect(monitor).not.toBeNull();

    nextSamples.push(0.2, 0.2, 0, 0);
    runFrame(10);
    runFrame(20);
    expect(onSpeechStart).toHaveBeenCalledTimes(1);
    expect(monitor!.hasSpeech()).toBe(true);
    runFrame(70);
    runFrame(120);
    expect(onSilenceSubmit).toHaveBeenCalledTimes(1);

    monitor!.rearm();
    nextSamples.push(0, 0);
    runFrame(220);
    runFrame(320);
    expect(onSilenceSubmit).toHaveBeenCalledTimes(1);
    expect(monitor!.hasSpeech()).toBe(false);
    monitor!.stop();
  });
});

function runFrame(now: number): void {
  const callback = frames.shift();
  if (!callback) throw new Error("expected a scheduled VAD frame");
  callback(now);
}
