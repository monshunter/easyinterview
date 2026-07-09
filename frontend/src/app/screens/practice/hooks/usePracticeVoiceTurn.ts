import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import type {
  CreatePracticeVoiceTurnRequest,
  PracticeMode,
  PracticeVoiceAudioInput,
  PracticeVoiceTurnResult,
} from "../../../../api/generated/types";
import { generateIdempotencyKey } from "../../../../lib/conventions/idempotency";
import { newId } from "../../../../lib/ids";
import type { Lang } from "../../../i18n/messages";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export type PracticeVoiceTurnState =
  | { kind: "idle" }
  | { kind: "recording" }
  | { kind: "submitting" }
  | { kind: "success"; result: PracticeVoiceTurnResult }
  | { kind: "error"; message: string };

export interface UsePracticeVoiceTurnInput {
  sessionId: string;
  turnId: string;
  lang: Lang;
  practiceMode: PracticeMode;
}

export interface UsePracticeVoiceTurnResult {
  ready: boolean;
  state: PracticeVoiceTurnState;
  manualTranscriptFallback: string;
  setManualTranscriptFallback: (next: string) => void;
  startRecording: () => Promise<void>;
  stopAndSubmit: () => Promise<PracticeVoiceTurnResult>;
  reset: () => void;
}

const SUPPORTED_CONTENT_TYPES = [
  "audio/webm",
  "audio/wav",
  "audio/mpeg",
] as const;

export function usePracticeVoiceTurn({
  sessionId,
  turnId,
  lang,
  practiceMode,
}: UsePracticeVoiceTurnInput): UsePracticeVoiceTurnResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const recorderRef = useRef<MediaRecorder | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const startedAtRef = useRef(0);

  const [state, setState] = useState<PracticeVoiceTurnState>({ kind: "idle" });
  const [manualTranscriptFallback, setManualTranscriptFallback] = useState("");

  const discardRecorder = useCallback(() => {
    const recorder = recorderRef.current;
    if (!recorder || recorder.state === "inactive") return;
    recorder.ondataavailable = null;
    recorder.onstop = null;
    recorder.onerror = null;
    recorder.stop();
  }, []);

  const cleanupStream = useCallback(() => {
    streamRef.current?.getTracks().forEach((track) => track.stop());
    streamRef.current = null;
    recorderRef.current = null;
  }, []);

  useEffect(
    () => () => {
      discardRecorder();
      cleanupStream();
    },
    [cleanupStream, discardRecorder],
  );

  const reset = useCallback(() => {
    discardRecorder();
    chunksRef.current = [];
    startedAtRef.current = 0;
    cleanupStream();
    setState({ kind: "idle" });
  }, [cleanupStream, discardRecorder]);

  const startRecording = useCallback(async () => {
    if (!client) {
      setState({ kind: "error", message: "practice voice client is missing" });
      return;
    }
    if (!sessionId || !turnId) {
      setState({ kind: "error", message: "practice voice session is not ready" });
      return;
    }
    if (!navigator.mediaDevices?.getUserMedia) {
      setState({ kind: "error", message: "microphone capture is unavailable" });
      return;
    }
    if (typeof MediaRecorder === "undefined") {
      setState({ kind: "error", message: "browser audio recorder is unavailable" });
      return;
    }

    try {
      cleanupStream();
      chunksRef.current = [];
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const contentType = chooseAudioContentType();
      const recorder = new MediaRecorder(stream, { mimeType: contentType });
      recorder.ondataavailable = (event) => {
        if (event.data.size > 0) chunksRef.current.push(event.data);
      };
      streamRef.current = stream;
      recorderRef.current = recorder;
      startedAtRef.current = Date.now();
      recorder.start();
      setState({ kind: "recording" });
    } catch (err) {
      cleanupStream();
      setState({ kind: "error", message: errorMessage(err) });
    }
  }, [cleanupStream, client, sessionId, turnId]);

  const stopAndSubmit = useCallback(async (): Promise<PracticeVoiceTurnResult> => {
    if (!client) throw new Error("usePracticeVoiceTurn: client missing");
    if (!sessionId) throw new Error("usePracticeVoiceTurn: sessionId missing");
    if (!turnId) throw new Error("usePracticeVoiceTurn: turnId missing");
    const recorder = recorderRef.current;
    if (!recorder || recorder.state !== "recording") {
      const err = new Error("recording is not active");
      setState({ kind: "error", message: err.message });
      throw err;
    }

    try {
      setState({ kind: "submitting" });
      await stopRecorder(recorder);
      const durationMs = Math.max(1, Date.now() - startedAtRef.current);
      const audioBlob = new Blob(chunksRef.current, {
        type: normalizeAudioContentType(
          chunksRef.current[0]?.type || recorder.mimeType,
        ),
      });
      cleanupStream();
      if (audioBlob.size <= 0) {
        throw new Error("recorded audio is empty");
      }
      const body: CreatePracticeVoiceTurnRequest = {
        clientVoiceTurnId: newId(),
        turnId,
        audio: {
          contentBase64: await blobToBase64(audioBlob),
          contentType: normalizeAudioContentType(audioBlob.type),
          durationMs,
          byteLength: audioBlob.size,
        },
        language: lang === "zh" ? "zh-CN" : "en",
        practiceMode,
        ...manualFallbackBody(manualTranscriptFallback),
      };
      const result = await client.createPracticeVoiceTurn(sessionId, body, {
        idempotencyKey: generateIdempotencyKey(),
      });
      setManualTranscriptFallback("");
      setState({ kind: "success", result });
      return result;
    } catch (err) {
      cleanupStream();
      const message = errorMessage(err);
      setState({ kind: "error", message });
      throw err instanceof Error ? err : new Error(message);
    }
  }, [
    cleanupStream,
    client,
    lang,
    manualTranscriptFallback,
    practiceMode,
    sessionId,
    turnId,
  ]);

  return useMemo<UsePracticeVoiceTurnResult>(
    () => ({
      ready: !!client && !!sessionId && !!turnId,
      state,
      manualTranscriptFallback,
      setManualTranscriptFallback,
      startRecording,
      stopAndSubmit,
      reset,
    }),
    [
      client,
      manualTranscriptFallback,
      reset,
      sessionId,
      startRecording,
      state,
      stopAndSubmit,
      turnId,
    ],
  );
}

function chooseAudioContentType(): PracticeVoiceAudioInput["contentType"] {
  const recorder = MediaRecorder as typeof MediaRecorder & {
    isTypeSupported?: (mimeType: string) => boolean;
  };
  return (
    SUPPORTED_CONTENT_TYPES.find(
      (contentType) =>
        !recorder.isTypeSupported || recorder.isTypeSupported(contentType),
    ) ?? "audio/webm"
  );
}

function normalizeAudioContentType(raw: string): PracticeVoiceAudioInput["contentType"] {
  return SUPPORTED_CONTENT_TYPES.includes(
    raw as PracticeVoiceAudioInput["contentType"],
  )
    ? (raw as PracticeVoiceAudioInput["contentType"])
    : "audio/webm";
}

function stopRecorder(recorder: MediaRecorder): Promise<void> {
  return new Promise((resolve, reject) => {
    recorder.onstop = () => resolve();
    recorder.onerror = () => reject(new Error("audio recorder failed"));
    recorder.stop();
  });
}

async function blobToBase64(blob: Blob): Promise<string> {
  if (typeof blob.arrayBuffer !== "function") {
    return blobToBase64WithFileReader(blob);
  }
  const bytes = new Uint8Array(await blob.arrayBuffer());
  let binary = "";
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte);
  });
  return btoa(binary);
}

function blobToBase64WithFileReader(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error("recorded audio cannot be read"));
    reader.onload = () => {
      const result = typeof reader.result === "string" ? reader.result : "";
      const [, base64 = ""] = result.split(",", 2);
      resolve(base64);
    };
    reader.readAsDataURL(blob);
  });
}

function manualFallbackBody(
  value: string,
): Pick<CreatePracticeVoiceTurnRequest, "manualTranscriptFallback"> {
  const trimmed = value.trim();
  return trimmed ? { manualTranscriptFallback: trimmed } : {};
}

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}
