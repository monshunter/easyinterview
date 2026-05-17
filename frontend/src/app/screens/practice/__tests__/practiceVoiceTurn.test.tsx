/**
 * @vitest-environment jsdom
 *
 * practice-voice-mvp item 4.2:
 * - record audio with MediaRecorder
 * - submit createPracticeVoiceTurn through the generated client + fixture
 * - render final transcript / assistant draft and keep manual transcript
 *   fallback usable when TTS fails after chat succeeds
 */

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { UUID_V7_REGEX } from "../../../../lib/ids";
import {
  TURN_A,
  buildPracticeClient,
  eventCalls,
  mountPracticeScreen,
  readBody,
  voiceTurnCalls,
} from "./practiceTestUtils";

const IDEMPOTENCY_KEY_REGEX = /^v1\.\d+\.[0-9a-f-]{36}$/;

describe("practice voice turn controller (item 4.2)", () => {
  beforeEach(() => {
    localStorage.setItem("ei-lang", "zh");
    installFakeAudioCapture();
    installFakeAudioPlayback();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
    localStorage.clear();
  });

  it("captures audio, posts createPracticeVoiceTurn, and renders transcript output", async () => {
    const { client, calls } = buildPracticeClient();
    mountPracticeScreen({
      client,
      routeParams: { mode: "voice", modality: "voice", practiceMode: "assisted" },
    });

    const user = userEvent.setup();
    await user.click(await screen.findByTestId("practice-voice-record-toggle"));
    await waitFor(() =>
      expect(screen.getByTestId("practice-voice-capture-status")).toHaveAttribute(
        "data-state",
        "recording",
      ),
    );
    expect(
      (screen.getByTestId("practice-voice-submit") as HTMLButtonElement).disabled,
    ).toBe(false);

    await user.click(screen.getByTestId("practice-voice-submit"));

    await waitFor(() => expect(voiceTurnCalls(calls)).toHaveLength(1));
    const call = voiceTurnCalls(calls)[0]!;
    expect(call.headers.get("Idempotency-Key")).toMatch(IDEMPOTENCY_KEY_REGEX);

    const body = readBody(call);
    expect(body.turnId).toBe(TURN_A);
    expect(body.practiceMode).toBe("assisted");
    expect(body.language).toBe("zh-CN");
    expect(UUID_V7_REGEX.test(body.clientVoiceTurnId as string)).toBe(true);
    expect(body.audio).toEqual(
      expect.objectContaining({
        contentBase64: "T2dnUw==",
        contentType: "audio/webm",
        byteLength: 4,
      }),
    );
    expect(
      (body.audio as { durationMs?: number }).durationMs,
    ).toBeGreaterThan(0);

    expect(
      await screen.findByText("我主导了设计系统迁移，先把 12 个团队按风险分组。"),
    ).toBeDefined();
    expect(
      screen.getByText("你提到按风险分组。请具体说明你如何处理最高风险团队的迁移窗口。"),
    ).toBeDefined();
    expect(screen.getByTestId("practice-voice-tts-status").textContent).toContain(
      "1",
    );
  });

  it("submits manual transcript fallback and keeps assistant text visible on TTS failure", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { createPracticeVoiceTurn: "tts-failed" },
    });
    mountPracticeScreen({
      client,
      routeParams: { mode: "voice", modality: "voice", practiceMode: "assisted" },
    });

    const user = userEvent.setup();
    await user.type(
      await screen.findByTestId("practice-voice-manual-fallback"),
      "我先做风险分组，再让高风险团队先跑试点。",
    );
    await user.click(screen.getByTestId("practice-voice-record-toggle"));
    await user.click(screen.getByTestId("practice-voice-submit"));

    await waitFor(() => expect(voiceTurnCalls(calls)).toHaveLength(1));
    const body = readBody(voiceTurnCalls(calls)[0]!);
    expect(body.manualTranscriptFallback).toBe(
      "我先做风险分组，再让高风险团队先跑试点。",
    );

    expect(await screen.findByText("我先做风险分组，再让高风险团队先跑试点。"))
      .toBeDefined();
    expect(screen.getByText("高风险团队试点时，你如何证明迁移没有影响线上体验？"))
      .toBeDefined();
    expect(screen.getByTestId("practice-voice-tts-error").textContent).toContain(
      "TTS_PROVIDER_FAILED",
    );
  });

  it("autoplays the returned TTS chunk and reports started, played, and committed context", async () => {
    const { client, calls } = buildPracticeClient();
    mountPracticeScreen({
      client,
      routeParams: { mode: "voice", modality: "voice", practiceMode: "assisted" },
    });

    const user = userEvent.setup();
    await submitDefaultVoiceTurn(user);

    await waitFor(() => {
      expect(eventBodies(calls).some((body) => body.kind === "tts_chunk_started"))
        .toBe(true);
    });
    expect(FakeAudioElement.instances[0]?.src).toBe(
      "fixture-audio://practice-voice/default/chunk-001",
    );
    expect(screen.getByTestId("practice-voice-playback-status")).toHaveAttribute(
      "data-state",
      "playing",
    );

    FakeAudioElement.instances[0]!.finish();

    await waitFor(() => {
      const bodies = eventBodies(calls);
      expect(bodies.some((body) => body.kind === "tts_chunk_played")).toBe(true);
      expect(bodies.some((body) => body.kind === "assistant_context_committed"))
        .toBe(true);
    });
    const played = eventBodies(calls).find(
      (body) => body.kind === "tts_chunk_played",
    )!;
    const committed = eventBodies(calls).find(
      (body) => body.kind === "assistant_context_committed",
    )!;
    expect(played.payload).toEqual(
      expect.objectContaining({
        voiceTurnId: "01918fa0-0000-7000-8000-00000000f201",
        chunkId: "voice-chunk-001",
        playedTextHash: "sha256:voice-default-chunk-001",
        playbackOffsetMs: 2840,
      }),
    );
    expect(committed.payload).toEqual(
      expect.objectContaining({
        voiceTurnId: "01918fa0-0000-7000-8000-00000000f201",
        chunkId: "voice-chunk-001",
        committedTextHash: "sha256:voice-default-chunk-001",
        playbackOffsetMs: 2840,
      }),
    );
    const playedCall = eventCalls(calls).find((call) =>
      call.bodyText?.includes('"tts_chunk_played"'),
    )!;
    expect(playedCall.headers.get("Idempotency-Key")).toBeNull();
  });

  it("stops active TTS playback and reports barge_in_detected before recording the next user speech", async () => {
    const { client, calls } = buildPracticeClient();
    mountPracticeScreen({
      client,
      routeParams: { mode: "voice", modality: "voice", practiceMode: "assisted" },
    });

    const user = userEvent.setup();
    await submitDefaultVoiceTurn(user);
    await waitFor(() =>
      expect(eventBodies(calls).some((body) => body.kind === "tts_chunk_started"))
        .toBe(true),
    );
    FakeAudioElement.instances[0]!.currentTime = 1.42;

    await user.click(screen.getByTestId("practice-voice-record-toggle"));

    await waitFor(() => {
      expect(eventBodies(calls).some((body) => body.kind === "barge_in_detected"))
        .toBe(true);
    });
    const played = eventBodies(calls).find(
      (body) => body.kind === "tts_chunk_played",
    )!;
    expect(played.payload).toEqual(
      expect.objectContaining({
        voiceTurnId: "01918fa0-0000-7000-8000-00000000f201",
        chunkId: "voice-chunk-001",
        playedTextHash: "sha256:voice-default-chunk-001",
      }),
    );
    expect((played.payload as { playedTextLength?: number }).playedTextLength)
      .toBeGreaterThan(0);
    const bargeIn = eventBodies(calls).find(
      (body) => body.kind === "barge_in_detected",
    )!;
    expect(bargeIn.payload).toEqual(
      expect.objectContaining({
        voiceTurnId: "01918fa0-0000-7000-8000-00000000f201",
        chunkId: "voice-chunk-001",
      }),
    );
    expect((bargeIn.payload as { playbackOffsetMs?: number }).playbackOffsetMs)
      .toBeGreaterThanOrEqual(0);
    const bargeInCall = eventCalls(calls).find((call) =>
      call.bodyText?.includes('"barge_in_detected"'),
    )!;
    expect(bargeInCall.headers.get("Idempotency-Key")).toBeNull();
    expect(FakeAudioElement.instances[0]!.paused).toBe(true);
    expect(screen.getByTestId("practice-voice-capture-status")).toHaveAttribute(
      "data-state",
      "recording",
    );
  });
});

async function submitDefaultVoiceTurn(
  user: ReturnType<typeof userEvent.setup>,
): Promise<void> {
  await user.click(await screen.findByTestId("practice-voice-record-toggle"));
  await waitFor(() =>
    expect(screen.getByTestId("practice-voice-capture-status")).toHaveAttribute(
      "data-state",
      "recording",
    ),
  );
  await user.click(screen.getByTestId("practice-voice-submit"));
  await screen.findByText("我主导了设计系统迁移，先把 12 个团队按风险分组。");
}

function eventBodies(calls: ReturnType<typeof eventCalls>): Record<string, unknown>[] {
  return eventCalls(calls).map(readBody);
}

function installFakeAudioCapture(): void {
  const tracks = [{ stop: vi.fn() }];
  Object.defineProperty(navigator, "mediaDevices", {
    configurable: true,
    value: {
      getUserMedia: vi.fn().mockResolvedValue({
        getTracks: () => tracks,
      }),
    },
  });

  class FakeMediaRecorder {
    static isTypeSupported() {
      return true;
    }

    readonly mimeType = "audio/webm";
    state: RecordingState = "inactive";
    ondataavailable: ((event: BlobEvent) => void) | null = null;
    onstop: ((event: Event) => void) | null = null;

    start() {
      this.state = "recording";
    }

    stop() {
      this.state = "inactive";
      const data = new Blob(["OggS"], { type: this.mimeType });
      this.ondataavailable?.({ data } as BlobEvent);
      this.onstop?.(new Event("stop"));
    }
  }

  vi.stubGlobal("MediaRecorder", FakeMediaRecorder);
}

class FakeAudioElement {
  static instances: FakeAudioElement[] = [];

  src = "";
  paused = true;
  currentTime = 0;
  onended: (() => void) | null = null;

  constructor(src?: string) {
    this.src = src ?? "";
    FakeAudioElement.instances.push(this);
  }

  play() {
    this.paused = false;
    return Promise.resolve();
  }

  pause() {
    this.paused = true;
  }

  finish() {
    this.paused = true;
    this.currentTime = 2.84;
    this.onended?.();
  }
}

function installFakeAudioPlayback(): void {
  FakeAudioElement.instances = [];
  vi.stubGlobal("Audio", FakeAudioElement);
}
