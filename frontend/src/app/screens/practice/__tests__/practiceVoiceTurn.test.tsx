/**
 * @vitest-environment jsdom
 *
 * practice-voice-mvp current phone contract:
 * - no manual record/submit/restart controls
 * - VAD submits speech after silence and TTS completion rearms listening
 * - only real speech-start barge-ins; hang-up exits to text and releases media
 */

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { UUID_V7_REGEX } from "../../../../lib/ids";
import {
  SESSION_A,
  TURN_A,
  buildPracticeClient,
  eventCalls,
  mountPracticeScreen,
  readBody,
  voiceTurnCalls,
} from "./practiceTestUtils";
import type { PracticeVoiceTurnResult } from "../../../../api/generated/types";

const IDEMPOTENCY_KEY_REGEX = /^v1\.\d+\.[0-9a-f-]{36}$/;
const DATA_AUDIO_REF_REGEX = /^data:audio\/[a-z0-9.+-]+;base64,/i;
const TURN_B = "01918fa0-0000-7000-8000-000000006001";
let fakeAudioStop: ReturnType<typeof vi.fn> | null = null;
let fakeVadLevel = 0;
let fakeVadNow = 0;
let nextFrameID = 1;
let fakeVadFrames: Map<number, FrameRequestCallback>;

describe("practice phone mode controller (item 4.2)", () => {
  beforeEach(() => {
    localStorage.setItem("ei-lang", "zh");
    installFakeAudioCapture();
    installFakeAudioPlayback();
    installFakeVad();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
    localStorage.clear();
  });

  it("submits the real voice turn from current phone mode", async () => {
    const { client, calls } = buildPracticeClient();
    mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone", practiceMode: "assisted" },
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
    const waveformBars = screen.getByTestId("practice-phone-waveform-bars");
    expect(waveformBars.children).toHaveLength(66);
    const firstBar = waveformBars.children[0] as HTMLElement;
    const firstHeight = firstBar.style.height;
    await waitFor(() => expect(firstBar.style.height).not.toBe(firstHeight));
    expect(screen.getByTestId("practice-phone-waveform-status")).toHaveAttribute(
      "data-icon",
      "mic",
    );
    expect(
      screen.getByTestId("practice-phone-captions-toggle").querySelector("svg"),
    ).not.toBeNull();
    expect(screen.getByTestId("practice-phone-hangup")).toBeDefined();
    expect(screen.queryByTestId("practice-phone-restart")).toBeNull();
    expect(screen.queryByTestId("practice-phone-question")).toBeNull();
    expect(screen.queryByText(/STAR/)).toBeNull();
    expect(screen.queryByTestId("practice-voice-record-toggle")).toBeNull();
    expect(screen.queryByTestId("practice-voice-submit")).toBeNull();
    expect(screen.queryByTestId("practice-voice-manual-fallback")).toBeNull();
    expect(screen.queryByTestId("practice-voice-annotated-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-expression-panel")).toBeNull();

    await submitDefaultPhoneTurn(calls);
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

    expect(screen.queryByTestId("practice-transcript")).toBeNull();
    await user.click(screen.getByTestId("practice-phone-captions-toggle"));
    expect(screen.getByTestId("practice-phone-captions-session-tag")).toHaveTextContent(
      "同一会话记录",
    );
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

    await submitDefaultPhoneTurn(calls);

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
        playedTextHash:
          "sha256:5607054a237c1bfd67f743bc6ff53c756d9d90b4e15b1826dc999b60e601e76e",
        playbackOffsetMs: 2840,
      }),
    );
    expect(committed.payload).toEqual(
      expect.objectContaining({
        voiceTurnId: "01918fa0-0000-7000-8000-00000000f201",
        chunkId: "voice-chunk-001",
        committedTextHash:
          "sha256:5607054a237c1bfd67f743bc6ff53c756d9d90b4e15b1826dc999b60e601e76e",
        playbackOffsetMs: 2840,
      }),
    );
    const playedCall = eventCalls(calls).find((call) =>
      call.bodyText?.includes('"tts_chunk_played"'),
    )!;
    expect(playedCall.headers.get("Idempotency-Key")).toBeNull();
    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );
    expect(navigator.mediaDevices.getUserMedia).toHaveBeenCalledTimes(1);
  });

  it("renders assistant text fallback when the tts-failed fixture returns TTS_PROVIDER_FAILED", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { createPracticeVoiceTurn: "tts-failed" },
    });
    mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone", practiceMode: "assisted" },
    });

    const user = userEvent.setup();
    await submitDefaultPhoneTurn(calls);

    expect(await screen.findByTestId("practice-phone-error")).toHaveTextContent(
      "语音播放暂时不可用",
    );
    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );

    await user.click(screen.getByTestId("practice-phone-captions-toggle"));
    expect(
      await screen.findByText("我先做风险分组，再让高风险团队先跑试点。"),
    ).toBeDefined();
    expect(
      screen.getByText("高风险团队试点时，你如何证明迁移没有影响线上体验？"),
    ).toBeDefined();
    expect(eventBodies(calls).some((body) => body.kind === "tts_chunk_started"))
      .toBe(false);
    expect(FakeAudioElement.instances).toHaveLength(0);
  });

  it("submits the next phone answer against the turn adopted from the prior voice response", async () => {
    const { client } = buildPracticeClient();
    const firstResult = voiceResultForTurn(TURN_B, 2, "第二题：请说明量化结果。");
    const secondResult = voiceResultForTurn(
      TURN_B,
      2,
      "请补充结果与目标的差异。",
    );
    const createVoiceTurn = vi
      .spyOn(client, "createPracticeVoiceTurn")
      .mockResolvedValueOnce(firstResult)
      .mockResolvedValueOnce(secondResult);
    mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone" },
    });

    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );
    act(() => emitSpeechThenSilence());
    await waitFor(() => expect(createVoiceTurn).toHaveBeenCalledTimes(1));
    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );

    act(() => emitSpeechThenSilence());
    await waitFor(() => expect(createVoiceTurn).toHaveBeenCalledTimes(2));

    expect(createVoiceTurn.mock.calls[0]?.[1].turnId).toBe(TURN_A);
    expect(createVoiceTurn.mock.calls[1]?.[1].turnId).toBe(TURN_B);
  });

  it("keeps the final user answer without adding an empty AI message when the voice turn completes the session", async () => {
    const { client } = buildPracticeClient();
    const completedResult = {
      voiceTurnId: "01918fa0-0000-7000-8000-00000000f299",
      userTranscriptFinal: "这是我的最后一个回答。",
      assistantTextDraft: "",
      ttsChunks: [],
      providerMetaSummary: {
        sttProfile: "practice.voice.stt.default",
        chatProfile: "practice.followup.default",
        ttsProfile: "practice.voice.tts.default",
        sttProvider: "fixture-stt",
        chatProvider: "fixture-chat",
        ttsProvider: "fixture-tts",
      },
      session: {
        id: SESSION_A,
        planId: "01918fa0-0000-7000-8000-000000004000",
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
        status: "completed",
        language: "zh-CN",
        hintsEnabled: true,
        turnCount: 1,
        currentTurn: null,
        createdAt: "2026-04-28T12:00:00Z",
        updatedAt: "2026-07-11T18:00:00Z",
      },
      ttsError: null,
    } satisfies PracticeVoiceTurnResult;
    const createVoiceTurn = vi
      .spyOn(client, "createPracticeVoiceTurn")
      .mockResolvedValue(completedResult);
    mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone" },
    });
    const user = userEvent.setup();

    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );
    act(() => emitSpeechThenSilence());
    await waitFor(() => expect(createVoiceTurn).toHaveBeenCalledTimes(1));
    await user.click(screen.getByTestId("practice-phone-captions-toggle"));

    expect(await screen.findByText("这是我的最后一个回答。")).toBeDefined();
    const messages = screen.getAllByTestId(/practice-transcript-message-/);
    expect(messages).toHaveLength(2);
    expect(messages[1]).toHaveAttribute("data-role", "user");
    expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
      "data-state",
      "paused",
    );
    expect(FakeAudioElement.instances).toHaveLength(0);
  });

  it("keeps the same session and localizes a double-invalid chat failure", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { createPracticeVoiceTurn: "chat-output-invalid" },
    });
    const { nav } = mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone" },
    });
    const user = userEvent.setup();

    await submitDefaultPhoneTurn(calls);

    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-error")).toHaveTextContent(
        "暂时无法生成符合本场语言的问题",
      ),
    );
    expect(screen.getByTestId("practice-screen")).toHaveAttribute(
      "data-session-id",
      SESSION_A,
    );
    expect(screen.getByTestId("practice-phone-surface")).toBeDefined();
    expect(nav).not.toHaveBeenCalled();
    expect(screen.queryByText(/did not match the session language/i)).toBeNull();
    expect(eventBodies(calls).some((body) => body.kind === "tts_chunk_started"))
      .toBe(false);
    expect(FakeAudioElement.instances).toHaveLength(0);

    await user.click(screen.getByTestId("practice-phone-hangup"));

    expect(nav).toHaveBeenCalledTimes(1);
    expect(nav).toHaveBeenCalledWith({
      name: "practice",
      params: expect.objectContaining({
        sessionId: SESSION_A,
        mode: "text",
        modality: "text",
      }),
    });
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
    expect(screen.getByTestId("practice-phone-waveform-status")).toHaveAttribute(
      "data-icon",
      "pause",
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

  it("reports partial playback before barge_in_detected on real speech-start", async () => {
    const { client, calls } = buildPracticeClient();
    mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone", practiceMode: "assisted" },
    });

    await submitDefaultPhoneTurn(calls);
    await waitFor(() =>
      expect(eventBodies(calls).some((body) => body.kind === "tts_chunk_started"))
        .toBe(true),
    );
    FakeAudioElement.instances[0]!.currentTime = 1.42;

    act(() => emitSpeechStart());

    await waitFor(() => {
      const bodies = eventBodies(calls);
      expect(bodies.some((body) => body.kind === "barge_in_detected")).toBe(true);
      expect(bodies.some((body) => body.kind === "tts_chunk_played")).toBe(true);
    });
    const bodies = eventBodies(calls);
    const playedIndex = bodies.findIndex((body) => body.kind === "tts_chunk_played");
    const bargeInIndex = bodies.findIndex(
      (body) => body.kind === "barge_in_detected",
    );
    expect(playedIndex).toBeGreaterThanOrEqual(0);
    expect(bargeInIndex).toBeGreaterThan(playedIndex);
    const played = bodies[playedIndex]!;
    expect(played.payload).toEqual(
      expect.objectContaining({
        voiceTurnId: "01918fa0-0000-7000-8000-00000000f201",
        chunkId: "voice-chunk-001",
        playedTextHash:
          "sha256:5607054a237c1bfd67f743bc6ff53c756d9d90b4e15b1826dc999b60e601e76e",
      }),
    );
    const bargeIn = bodies[bargeInIndex]!;
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

  it("hang-up exits to text, releases the microphone, and never reports barge-in", async () => {
    const { client, calls } = buildPracticeClient();
    const { nav } = mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone" },
    });
    const user = userEvent.setup();
    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );

    await user.click(screen.getByTestId("practice-phone-hangup"));

    expect(fakeAudioStop).toHaveBeenCalledTimes(1);
    expect(voiceTurnCalls(calls)).toHaveLength(0);
    expect(eventBodies(calls).some((body) => body.kind === "barge_in_detected"))
      .toBe(false);
    expect(nav).toHaveBeenCalledWith({
      name: "practice",
      params: expect.objectContaining({
        mode: "text",
        modality: "text",
        sessionId: expect.any(String),
      }),
    });
  });

  it("hang-up settles a non-empty utterance, stops capture, and never plays the late TTS", async () => {
    const { client, calls } = buildPracticeClient();
    const { nav } = mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone" },
    });
    const user = userEvent.setup();
    await waitFor(() =>
      expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
        "data-capture-state",
        "recording",
      ),
    );

    act(() => emitSpeechStart());
    await user.click(screen.getByTestId("practice-phone-hangup"));

    expect(nav).toHaveBeenCalledWith({
      name: "practice",
      params: expect.objectContaining({ mode: "text", modality: "text" }),
    });
    await waitFor(() => expect(voiceTurnCalls(calls)).toHaveLength(1));
    expect(fakeAudioStop).toHaveBeenCalledTimes(1);
    expect(eventBodies(calls).some((body) => body.kind === "barge_in_detected"))
      .toBe(false);
    await act(async () => Promise.resolve());
    expect(FakeAudioElement.instances).toHaveLength(0);
  });

  it("hang-up during TTS commits only the heard prefix and suppresses late completion", async () => {
    const { client, calls } = buildPracticeClient();
    const { nav } = mountPracticeScreen({
      client,
      routeParams: { mode: "phone", modality: "phone" },
    });
    const user = userEvent.setup();
    await submitDefaultPhoneTurn(calls);
    await waitFor(() =>
      expect(eventBodies(calls).some((body) => body.kind === "tts_chunk_started"))
        .toBe(true),
    );
    const audio = FakeAudioElement.instances[0]!;
    audio.currentTime = 1.42;

    await user.click(screen.getByTestId("practice-phone-hangup"));

    await waitFor(() => {
      const kinds = eventBodies(calls).map((body) => body.kind);
      expect(kinds).toContain("tts_chunk_played");
      expect(kinds).toContain("assistant_context_committed");
      expect(kinds).not.toContain("barge_in_detected");
    });
    expect(audio.paused).toBe(true);
    expect(fakeAudioStop).toHaveBeenCalledTimes(1);
    const eventCount = eventCalls(calls).length;
    act(() => audio.finish());
    await act(async () => Promise.resolve());
    expect(eventCalls(calls)).toHaveLength(eventCount);
    expect(nav).toHaveBeenCalledWith({
      name: "practice",
      params: expect.objectContaining({ mode: "text", modality: "text" }),
    });
  });
});

async function submitDefaultPhoneTurn(
  calls: Parameters<typeof voiceTurnCalls>[0],
): Promise<void> {
  await waitFor(() =>
    expect(screen.getByTestId("practice-phone-call-state")).toHaveAttribute(
      "data-capture-state",
      "recording",
    ),
  );
  act(() => emitSpeechThenSilence());
  await waitFor(() => expect(voiceTurnCalls(calls)).toHaveLength(1));
}

function eventBodies(
  calls: Parameters<typeof eventCalls>[0],
): Record<string, unknown>[] {
  return eventCalls(calls).map(readBody);
}

function voiceResultForTurn(
  turnId: string,
  turnIndex: number,
  questionText: string,
): PracticeVoiceTurnResult {
  return {
    voiceTurnId: `01918fa0-0000-7000-8000-00000000f3${turnIndex
      .toString()
      .padStart(2, "0")}`,
    userTranscriptFinal: `第 ${turnIndex} 次电话回答`,
    assistantTextDraft: questionText,
    ttsChunks: [],
    providerMetaSummary: {
      sttProfile: "practice.voice.stt.default",
      chatProfile: "practice.followup.default",
      ttsProfile: "practice.voice.tts.default",
      sttProvider: "fixture-stt",
      chatProvider: "fixture-chat",
      ttsProvider: "fixture-tts",
    },
    session: {
      id: SESSION_A,
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
      status: "running",
      language: "zh-CN",
      hintsEnabled: true,
      turnCount: turnIndex,
      currentTurn: {
        id: turnId,
        turnIndex,
        questionText,
        status: "asked",
      },
      createdAt: "2026-07-11T00:00:00Z",
      updatedAt: "2026-07-11T00:00:01Z",
    },
    ttsError: {
      code: "TTS_PROVIDER_FAILED",
      message: "TTS unavailable in continuity test",
      retryable: true,
    },
  };
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

function installFakeVad(): void {
  fakeVadLevel = 0;
  fakeVadNow = performance.now();
  nextFrameID = 1;
  fakeVadFrames = new Map();
  vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
    const id = nextFrameID++;
    fakeVadFrames.set(id, callback);
    return id;
  });
  vi.stubGlobal("cancelAnimationFrame", (id: number) => {
    fakeVadFrames.delete(id);
  });

  class FakeAnalyser {
    fftSize = 8;
    smoothingTimeConstant = 0;

    getFloatTimeDomainData(target: Float32Array) {
      target.fill(fakeVadLevel);
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

  vi.stubGlobal("AudioContext", FakeAudioContext);
}

function emitSpeechStart(): void {
  for (let index = 0; index < 3; index += 1) {
    runVadFrame(0.2, 30);
  }
}

function emitSpeechThenSilence(): void {
  emitSpeechStart();
  for (let index = 0; index < 4; index += 1) {
    runVadFrame(0, 250);
  }
}

function runVadFrame(level: number, durationMs: number): void {
  const entry = fakeVadFrames.entries().next().value as
    | [number, FrameRequestCallback]
    | undefined;
  if (!entry) throw new Error("expected a scheduled phone VAD frame");
  fakeVadFrames.delete(entry[0]);
  fakeVadLevel = level;
  fakeVadNow += durationMs;
  entry[1](fakeVadNow);
}
