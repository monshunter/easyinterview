import { describe, expect, it } from "vitest";

import {
  createPhoneVadState,
  rearmPhoneVad,
  stepPhoneVad,
} from "./phoneVad";

const SILENCE_THRESHOLD_MS = 600;

function frame(
  overrides: Partial<Parameters<typeof stepPhoneVad>[1]> = {},
): Parameters<typeof stepPhoneVad>[1] {
  return {
    speechDetected: false,
    durationMs: 100,
    paused: false,
    ttsPlaying: false,
    ...overrides,
  };
}

describe("phone VAD decision kernel", () => {
  it("does not submit silence before real speech is detected", () => {
    let state = createPhoneVadState();

    for (let index = 0; index < 3; index += 1) {
      const decision = stepPhoneVad(
        state,
        frame({ durationMs: SILENCE_THRESHOLD_MS }),
        SILENCE_THRESHOLD_MS,
      );
      state = decision.state;
      expect(decision.shouldSubmit).toBe(false);
    }

    expect(state).toEqual({
      hasSpeech: false,
      silenceMs: 0,
      submitted: false,
    });
  });

  it("submits once after real speech is followed by continuous silence", () => {
    let state = createPhoneVadState();

    state = stepPhoneVad(
      state,
      frame({ speechDetected: true, durationMs: 20 }),
      SILENCE_THRESHOLD_MS,
    ).state;

    const earlySilence = stepPhoneVad(
      state,
      frame({ durationMs: SILENCE_THRESHOLD_MS - 1 }),
      SILENCE_THRESHOLD_MS,
    );
    expect(earlySilence.shouldSubmit).toBe(false);

    const thresholdReached = stepPhoneVad(
      earlySilence.state,
      frame({ durationMs: 1 }),
      SILENCE_THRESHOLD_MS,
    );
    expect(thresholdReached.shouldSubmit).toBe(true);

    const laterSilence = stepPhoneVad(
      thresholdReached.state,
      frame({ durationMs: SILENCE_THRESHOLD_MS }),
      SILENCE_THRESHOLD_MS,
    );
    expect(laterSilence.shouldSubmit).toBe(false);
    expect(laterSilence.state.submitted).toBe(true);
  });

  it.each([
    ["paused", { paused: true }],
    ["TTS playback", { ttsPlaying: true }],
  ])("does not submit during %s", (_label, gate) => {
    let state = createPhoneVadState();
    state = stepPhoneVad(
      state,
      frame({ speechDetected: true }),
      SILENCE_THRESHOLD_MS,
    ).state;
    state = stepPhoneVad(
      state,
      frame({ durationMs: SILENCE_THRESHOLD_MS - 1 }),
      SILENCE_THRESHOLD_MS,
    ).state;

    const gated = stepPhoneVad(
      state,
      frame({ durationMs: 1, ...gate }),
      SILENCE_THRESHOLD_MS,
    );
    expect(gated.shouldSubmit).toBe(false);
    expect(gated.state).toEqual(createPhoneVadState());

    const staleSilence = stepPhoneVad(
      gated.state,
      frame({ durationMs: SILENCE_THRESHOLD_MS }),
      SILENCE_THRESHOLD_MS,
    );
    expect(staleSilence.shouldSubmit).toBe(false);
  });

  it("can submit a new utterance after rearm", () => {
    let state = createPhoneVadState();
    state = stepPhoneVad(
      state,
      frame({ speechDetected: true }),
      SILENCE_THRESHOLD_MS,
    ).state;
    const firstSubmission = stepPhoneVad(
      state,
      frame({ durationMs: SILENCE_THRESHOLD_MS }),
      SILENCE_THRESHOLD_MS,
    );
    expect(firstSubmission.shouldSubmit).toBe(true);

    state = rearmPhoneVad();
    const silenceAfterRearm = stepPhoneVad(
      state,
      frame({ durationMs: SILENCE_THRESHOLD_MS }),
      SILENCE_THRESHOLD_MS,
    );
    expect(silenceAfterRearm.shouldSubmit).toBe(false);

    state = stepPhoneVad(
      silenceAfterRearm.state,
      frame({ speechDetected: true }),
      SILENCE_THRESHOLD_MS,
    ).state;
    const secondSubmission = stepPhoneVad(
      state,
      frame({ durationMs: SILENCE_THRESHOLD_MS }),
      SILENCE_THRESHOLD_MS,
    );
    expect(secondSubmission.shouldSubmit).toBe(true);
  });
});
