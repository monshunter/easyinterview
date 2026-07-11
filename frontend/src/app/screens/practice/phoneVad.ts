export interface PhoneVadState {
  hasSpeech: boolean;
  silenceMs: number;
  submitted: boolean;
}

export interface PhoneVadFrame {
  speechDetected: boolean;
  durationMs: number;
  paused: boolean;
  ttsPlaying: boolean;
}

export interface PhoneVadDecision {
  state: PhoneVadState;
  shouldSubmit: boolean;
}

export function createPhoneVadState(): PhoneVadState {
  return {
    hasSpeech: false,
    silenceMs: 0,
    submitted: false,
  };
}

export function rearmPhoneVad(): PhoneVadState {
  return createPhoneVadState();
}

export function stepPhoneVad(
  state: PhoneVadState,
  frame: PhoneVadFrame,
  silenceThresholdMs: number,
): PhoneVadDecision {
  if (frame.paused || frame.ttsPlaying) {
    return { state: rearmPhoneVad(), shouldSubmit: false };
  }

  if (state.submitted) {
    return { state, shouldSubmit: false };
  }

  if (frame.speechDetected) {
    return {
      state: { hasSpeech: true, silenceMs: 0, submitted: false },
      shouldSubmit: false,
    };
  }

  if (!state.hasSpeech) {
    return { state, shouldSubmit: false };
  }

  const silenceMs = state.silenceMs + Math.max(0, frame.durationMs);
  const shouldSubmit = silenceMs >= silenceThresholdMs;

  return {
    state: {
      hasSpeech: true,
      silenceMs,
      submitted: shouldSubmit,
    },
    shouldSubmit,
  };
}
