import { useCallback, useEffect, useRef } from "react";

import type { PracticeVoiceTurnResult } from "../../../api/generated/types";
import type {
  PracticeVoicePlaybackState,
  PracticeVoiceInterruptionReason,
} from "./hooks/usePracticeVoicePlayback";
import type { PracticeVoiceTurnState } from "./hooks/usePracticeVoiceTurn";
import type { StopPracticeVoiceTurnOptions } from "./hooks/usePracticeVoiceTurn";
import {
  startPhoneVadMonitor,
  type PhoneVadMonitor,
  type PhoneVadMonitorOptions,
} from "./phoneVadMonitor";

export interface PracticePhoneVoiceTurnPort {
  ready: boolean;
  state: PracticeVoiceTurnState;
  connectMicrophone: () => Promise<void>;
  getMediaStream: () => MediaStream | null;
  startRecording: () => Promise<void>;
  stopAndSubmit: (
    options?: StopPracticeVoiceTurnOptions,
  ) => Promise<PracticeVoiceTurnResult>;
  discardUtterance: () => void;
  releaseMicrophone: () => void;
}

export interface PracticePhonePlaybackPort {
  state: PracticeVoicePlaybackState;
  interrupt: (reason: PracticeVoiceInterruptionReason) => Promise<boolean>;
}

interface UsePracticePhoneControllerInput {
  enabled: boolean;
  active: boolean;
  voiceTurn: PracticePhoneVoiceTurnPort;
  voicePlayback: PracticePhonePlaybackPort;
  onTurnResult: (result: PracticeVoiceTurnResult) => void;
  onError: (error: unknown) => void;
  createVadMonitor?: (
    stream: MediaStream,
    options: PhoneVadMonitorOptions,
  ) => PhoneVadMonitor | null;
}

export interface UsePracticePhoneControllerResult {
  exitPhoneMode: () => void;
}

/** Coordinates call-scoped capture, VAD turn timing, playback and exit. */
export function usePracticePhoneController({
  enabled,
  active,
  voiceTurn,
  voicePlayback,
  onTurnResult,
  onError,
  createVadMonitor = startPhoneVadMonitor,
}: UsePracticePhoneControllerInput): UsePracticePhoneControllerResult {
  const monitorRef = useRef<PhoneVadMonitor | null>(null);
  const submittingRef = useRef(false);
  const bargeInBarrierRef = useRef<Promise<void> | null>(null);
  const exitInProgressRef = useRef(false);
  const callActiveRef = useRef(enabled && active);
  const captureKindRef = useRef(voiceTurn.state.kind);
  const playbackKindRef = useRef(voicePlayback.state.kind);
  const onTurnResultRef = useRef(onTurnResult);
  const onErrorRef = useRef(onError);
  const readyRef = useRef(voiceTurn.ready);
  const connectMicrophoneRef = useRef(voiceTurn.connectMicrophone);
  const getMediaStreamRef = useRef(voiceTurn.getMediaStream);
  const startRecordingRef = useRef(voiceTurn.startRecording);
  const stopAndSubmitRef = useRef(voiceTurn.stopAndSubmit);
  const discardUtteranceRef = useRef(voiceTurn.discardUtterance);
  const releaseMicrophoneRef = useRef(voiceTurn.releaseMicrophone);
  const interruptPlaybackRef = useRef(voicePlayback.interrupt);
  const createVadMonitorRef = useRef(createVadMonitor);

  callActiveRef.current = enabled && active;
  captureKindRef.current = voiceTurn.state.kind;
  playbackKindRef.current = voicePlayback.state.kind;
  onTurnResultRef.current = onTurnResult;
  onErrorRef.current = onError;
  readyRef.current = voiceTurn.ready;
  connectMicrophoneRef.current = voiceTurn.connectMicrophone;
  getMediaStreamRef.current = voiceTurn.getMediaStream;
  startRecordingRef.current = voiceTurn.startRecording;
  stopAndSubmitRef.current = voiceTurn.stopAndSubmit;
  discardUtteranceRef.current = voiceTurn.discardUtterance;
  releaseMicrophoneRef.current = voiceTurn.releaseMicrophone;
  interruptPlaybackRef.current = voicePlayback.interrupt;
  createVadMonitorRef.current = createVadMonitor;

  const stopMonitor = useCallback(() => {
    monitorRef.current?.stop();
    monitorRef.current = null;
  }, []);

  const resumeListening = useCallback(async () => {
    if (
      !callActiveRef.current ||
      !readyRef.current ||
      captureKindRef.current === "recording" ||
      captureKindRef.current === "submitting"
    ) {
      return;
    }
    monitorRef.current?.rearm();
    try {
      await startRecordingRef.current();
    } catch (error) {
      onErrorRef.current(error);
    }
  }, []);

  const submitCurrentUtterance = useCallback(() => {
    if (
      !callActiveRef.current ||
      submittingRef.current ||
      captureKindRef.current !== "recording"
    ) {
      return;
    }
    submittingRef.current = true;
    void (async () => {
      const barrier = bargeInBarrierRef.current;
      try {
        await barrier;
      } catch (error) {
        bargeInBarrierRef.current = null;
        discardUtteranceRef.current();
        onErrorRef.current(error);
        return;
      }
      bargeInBarrierRef.current = null;
      try {
        const result = await stopAndSubmitRef.current();
        captureKindRef.current = "success";
        onTurnResultRef.current(result);
        if (
          result.session.status === "running" &&
          (result.ttsChunks.length === 0 || Boolean(result.ttsError)) &&
          callActiveRef.current
        ) {
          await resumeListening();
        }
      } catch (error) {
        onErrorRef.current(error);
      } finally {
        submittingRef.current = false;
      }
    })();
  }, [resumeListening]);

  const handleSpeechStart = useCallback(() => {
    if (!callActiveRef.current || playbackKindRef.current !== "playing") return;
    monitorRef.current?.rearm();
    if (!bargeInBarrierRef.current) {
      const barrier = interruptPlaybackRef.current("user_speech")
        .then((interrupted) => {
          if (!interrupted) {
            throw new Error("phone barge-in events were not acknowledged");
          }
        });
      bargeInBarrierRef.current = barrier;
      void barrier.catch(() => undefined);
    }
    if (
      captureKindRef.current !== "recording" &&
      captureKindRef.current !== "submitting"
    ) {
      void startRecordingRef.current().catch((error: unknown) =>
        onErrorRef.current(error),
      );
    }
  }, []);

  const connectAndListen = useCallback(async () => {
    if (!callActiveRef.current || !readyRef.current) return;
    try {
      await connectMicrophoneRef.current();
      if (!callActiveRef.current) return;
      const stream = getMediaStreamRef.current();
      if (!stream) return;
      if (!monitorRef.current) {
        monitorRef.current = createVadMonitorRef.current(stream, {
          getPaused: () => !callActiveRef.current,
          getTtsPlaying: () => playbackKindRef.current === "playing",
          onSpeechStart: handleSpeechStart,
          onSilenceSubmit: submitCurrentUtterance,
        });
      }
      await resumeListening();
    } catch (error) {
      onErrorRef.current(error);
    }
  }, [handleSpeechStart, resumeListening, submitCurrentUtterance]);

  const exitPhoneMode = useCallback(() => {
    if (exitInProgressRef.current) return;
    exitInProgressRef.current = true;
    const monitor = monitorRef.current;
    const shouldSettleUtterance =
      captureKindRef.current === "recording" &&
      Boolean(monitor?.hasSpeech()) &&
      !submittingRef.current;
    callActiveRef.current = false;
    stopMonitor();
    if (shouldSettleUtterance) {
      submittingRef.current = true;
      const options: StopPracticeVoiceTurnOptions = {
        releaseMicrophoneAfterSnapshot: true,
      };
      if (bargeInBarrierRef.current) {
        options.beforeSubmit = bargeInBarrierRef.current;
      }
      void stopAndSubmitRef
        .current(options)
        .then((result) => onTurnResultRef.current(result))
        .catch((error: unknown) => onErrorRef.current(error))
        .finally(() => {
          bargeInBarrierRef.current = null;
          submittingRef.current = false;
        });
    } else {
      if (captureKindRef.current === "recording") {
        discardUtteranceRef.current();
      }
      releaseMicrophoneRef.current();
    }
    void interruptPlaybackRef
      .current("mode_switch")
      .catch((error: unknown) => onErrorRef.current(error));
  }, [stopMonitor]);

  useEffect(() => {
    if (enabled && active) {
      exitInProgressRef.current = false;
      void connectAndListen();
      return;
    }
    exitPhoneMode();
  }, [active, connectAndListen, enabled, exitPhoneMode, voiceTurn.ready]);

  useEffect(() => {
    if (
      enabled &&
      active &&
      (voicePlayback.state.kind === "completed" ||
        voicePlayback.state.kind === "error")
    ) {
      void resumeListening();
    }
  }, [active, enabled, resumeListening, voicePlayback.state.kind]);

  useEffect(
    () => () => {
      callActiveRef.current = false;
      stopMonitor();
      if (!submittingRef.current) {
        releaseMicrophoneRef.current();
      }
      void interruptPlaybackRef.current("mode_switch").catch(() => undefined);
    }, [stopMonitor],
  );

  return { exitPhoneMode };
}
