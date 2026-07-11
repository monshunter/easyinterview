import {
  createPhoneVadState,
  rearmPhoneVad,
  stepPhoneVad,
  type PhoneVadState,
} from "./phoneVad";

export interface PhoneVadMonitorOptions {
  silenceThresholdMs?: number;
  speechRmsThreshold?: number;
  speechStartFrames?: number;
  getPaused: () => boolean;
  getTtsPlaying: () => boolean;
  onSpeechStart: () => void;
  onSilenceSubmit: () => void;
}

export interface PhoneVadMonitor {
  stop: () => void;
  rearm: () => void;
  hasSpeech: () => boolean;
}

const DEFAULT_SILENCE_THRESHOLD_MS = 900;
const DEFAULT_SPEECH_RMS_THRESHOLD = 0.025;
const DEFAULT_SPEECH_START_FRAMES = 3;

/**
 * Web Audio adapter for the pure `phoneVad` kernel. It owns no call state;
 * callers provide current pause/playback gates and lifecycle callbacks.
 */
export function startPhoneVadMonitor(
  stream: MediaStream,
  options: PhoneVadMonitorOptions,
): PhoneVadMonitor | null {
  const AudioContextCtor =
    window.AudioContext ??
    (window as typeof window & { webkitAudioContext?: typeof AudioContext })
      .webkitAudioContext;
  if (!AudioContextCtor) return null;

  const audioContext = new AudioContextCtor();
  const analyser = audioContext.createAnalyser();
  analyser.fftSize = 512;
  analyser.smoothingTimeConstant = 0.2;
  const source = audioContext.createMediaStreamSource(stream);
  source.connect(analyser);

  const samples = new Float32Array(analyser.fftSize);
  const silenceThresholdMs =
    options.silenceThresholdMs ?? DEFAULT_SILENCE_THRESHOLD_MS;
  const speechRmsThreshold =
    options.speechRmsThreshold ?? DEFAULT_SPEECH_RMS_THRESHOLD;
  const speechStartFrames =
    options.speechStartFrames ?? DEFAULT_SPEECH_START_FRAMES;
  let vadState: PhoneVadState = createPhoneVadState();
  let speechFrames = 0;
  let speechLatched = false;
  let lastFrameAt = performance.now();
  let frameId = 0;
  let stopped = false;

  const frame = (now: number) => {
    if (stopped) return;
    analyser.getFloatTimeDomainData(samples);
    const speechDetected = rms(samples) >= speechRmsThreshold;
    speechFrames = speechDetected ? speechFrames + 1 : 0;
    const confirmedSpeech = speechFrames >= speechStartFrames;

    if (confirmedSpeech && !speechLatched) {
      speechLatched = true;
      options.onSpeechStart();
    } else if (!speechDetected) {
      speechLatched = false;
    }

    const durationMs = Math.max(0, Math.min(250, now - lastFrameAt));
    lastFrameAt = now;
    const decision = stepPhoneVad(
      vadState,
      {
        speechDetected: confirmedSpeech,
        durationMs,
        paused: options.getPaused(),
        ttsPlaying: options.getTtsPlaying(),
      },
      silenceThresholdMs,
    );
    vadState = decision.state;
    if (decision.shouldSubmit) options.onSilenceSubmit();
    frameId = requestAnimationFrame(frame);
  };

  frameId = requestAnimationFrame(frame);

  return {
    stop: () => {
      if (stopped) return;
      stopped = true;
      cancelAnimationFrame(frameId);
      source.disconnect();
      analyser.disconnect();
      void audioContext.close();
    },
    rearm: () => {
      vadState = rearmPhoneVad();
      speechFrames = 0;
      speechLatched = false;
      lastFrameAt = performance.now();
    },
    hasSpeech: () => vadState.hasSpeech,
  };
}

function rms(samples: Float32Array): number {
  if (samples.length === 0) return 0;
  let sum = 0;
  for (const sample of samples) sum += sample * sample;
  return Math.sqrt(sum / samples.length);
}
