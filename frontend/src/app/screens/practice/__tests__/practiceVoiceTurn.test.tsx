/**
 * @vitest-environment jsdom
 *
 * practice-voice-mvp item 4.2 current contract:
 * - user-visible "voice" routes render as phone mode
 * - phone mode has call-layer controls only: captions, hang up, restart
 * - the phone controller still records a voice turn, submits
 *   createPracticeVoiceTurn, plays returned TTS, and reports playback events
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
const DATA_AUDIO_REF_REGEX = /^data:audio\/[a-z0-9.+-]+;base64,/i;
let fakeAudioStop: ReturnType<typeof vi.fn> | null = null;

describe("practice phone mode controller (item 4.2)", () => {
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

  it("normalizes legacy voice routes to phone UI while submitting the real voice turn", async () => {
    const { client, calls } = buildPracticeClient();
    mountPracticeScreen({
      client,
      routeParams: { mode: "voice", modality: "voice", practiceMode: "assisted" },
    });

    const user = userEvent.setup();
    expect(await screen.findByTestId("practice-phone-surface")).toBeDefined();
    expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
      "data-state",
      "connected",
    );
    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );

    expect(screen.getByText("电话模式进行中")).toBeDefined();
    expect(screen.getByTestId("practice-phone-waveform")).toBeDefined();
    expect(screen.getByTestId("practice-phone-captions-toggle")).toBeDefined();
    expect(screen.getByTestId("practice-phone-hangup")).toBeDefined();
    expect(screen.getByTestId("practice-phone-restart")).toBeDefined();
    expect(screen.queryByTestId("practice-phone-question")).toBeNull();
    expect(screen.queryByText(/STAR/)).toBeNull();
    expect(screen.queryByTestId("practice-voice-record-toggle")).toBeNull();
    expect(screen.queryByTestId("practice-voice-submit")).toBeNull();
    expect(screen.queryByTestId("practice-voice-manual-fallback")).toBeNull();
    expect(screen.queryByTestId("practice-voice-annotated-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-expression-panel")).toBeNull();

    await user.click(screen.getByTestId("practice-phone-hangup"));

    await waitFor(() => expect(voiceTurnCalls(calls)).toHaveLength(1));
    expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
      "data-state",
      "ended",
    );
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

    expect(screen.queryByTestId("practice-transcript")).toBeNull();
    await user.click(screen.getByTestId("practice-phone-captions-toggle"));
    expect(
      await screen.findByText("我主导了设计系统迁移，先把 12 个团队按风险分组。"),
    ).toBeDefined();
    expect(
      screen.getByText("你提到按风险分组。请具体说明你如何处理最高风险团队的迁移窗口。"),
    ).toBeDefined();
  });

  it("autoplays the returned TTS chunk and reports started, played, and committed context", async () => {
    const { client, calls } = buildPracticeClient();
    mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone", practiceMode: "assisted" },
    });

    const user = userEvent.setup();
    await submitDefaultPhoneTurn(user, calls);

    await waitFor(() => {
      expect(eventBodies(calls).some((body) => body.kind === "tts_chunk_started"))
        .toBe(true);
    });
    expect(FakeAudioElement.instances[0]?.src).toMatch(DATA_AUDIO_REF_REGEX);
    expect(FakeAudioElement.instances[0]?.src).not.toMatch(
      /^fixture-audio:\/\//,
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

  it("stops and discards active microphone capture when phone mode is paused", async () => {
    const { client, calls } = buildPracticeClient();
    mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone", practiceMode: "assisted" },
    });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );

    await user.click(screen.getByTestId("practice-topbar-pause"));

    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-state",
        "paused",
      ),
    );
    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "idle",
      ),
    );
    expect(fakeAudioStop).toHaveBeenCalledTimes(1);
    expect(voiceTurnCalls(calls)).toHaveLength(0);

    await user.click(screen.getByTestId("practice-topbar-pause"));

    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-state",
        "connected",
      ),
    );
    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );
    expect(navigator.mediaDevices.getUserMedia).toHaveBeenCalledTimes(2);
  });

  it("reports barge_in_detected when restarting during active assistant playback", async () => {
    const { client, calls } = buildPracticeClient();
    mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone", practiceMode: "assisted" },
    });

    const user = userEvent.setup();
    await submitDefaultPhoneTurn(user, calls);
    await waitFor(() =>
      expect(eventBodies(calls).some((body) => body.kind === "tts_chunk_started"))
        .toBe(true),
    );
    FakeAudioElement.instances[0]!.currentTime = 1.42;

    await user.click(screen.getByTestId("practice-phone-restart"));

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
    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );
    expect(voiceTurnCalls(calls)).toHaveLength(1);
  });
});

async function submitDefaultPhoneTurn(
  user: ReturnType<typeof userEvent.setup>,
  calls: Parameters<typeof voiceTurnCalls>[0],
): Promise<void> {
  await waitFor(() =>
    expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
      "data-capture-state",
      "recording",
    ),
  );
  await user.click(screen.getByTestId("practice-phone-hangup"));
  await waitFor(() => expect(voiceTurnCalls(calls)).toHaveLength(1));
}

function eventBodies(
  calls: Parameters<typeof eventCalls>[0],
): Record<string, unknown>[] {
  return eventCalls(calls).map(readBody);
}

function installFakeAudioCapture(): void {
  fakeAudioStop = vi.fn();
  const tracks = [{ stop: fakeAudioStop }];
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
