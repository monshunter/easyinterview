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

export interface StopPracticeVoiceTurnOptions {
  /** Stop the call-scoped stream after the recorder emitted its final blob. */
  releaseMicrophoneAfterSnapshot?: boolean;
  /** Event-ordering barrier that must settle before createVoiceTurn is sent. */
  beforeSubmit?: Promise<void>;
}

export interface UsePracticeVoiceTurnResult {
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
  const connectPromiseRef = useRef<Promise<void> | null>(null);
  const connectionGenerationRef = useRef(0);
  const chunksRef = useRef<Blob[]>([]);
  const startedAtRef = useRef(0);
  const snapshotPendingRef = useRef(false);
  const releaseRequestedRef = useRef(false);
  const releaseStateRequestedRef = useRef(false);

  const [state, setState] = useState<PracticeVoiceTurnState>({ kind: "idle" });

  const discardRecorder = useCallback(() => {
    const recorder = recorderRef.current;
    recorderRef.current = null;
    chunksRef.current = [];
    startedAtRef.current = 0;
    if (!recorder || recorder.state === "inactive") return;
    recorder.ondataavailable = null;
    recorder.onstop = null;
    recorder.onerror = null;
    recorder.stop();
  }, []);

  const releaseResources = useCallback(
    (updateState: boolean) => {
      if (snapshotPendingRef.current) {
        releaseRequestedRef.current = true;
        releaseStateRequestedRef.current =
          releaseStateRequestedRef.current || updateState;
        return;
      }
      connectionGenerationRef.current += 1;
      connectPromiseRef.current = null;
      discardRecorder();
      streamRef.current?.getTracks().forEach((track) => track.stop());
      streamRef.current = null;
      if (updateState) setState({ kind: "idle" });
    },
    [discardRecorder],
  );

  useEffect(
    () => () => {
      releaseResources(false);
    },
    [releaseResources],
  );

  const connectMicrophone = useCallback(async () => {
    if (streamRef.current) return;
    if (connectPromiseRef.current) {
      await connectPromiseRef.current;
      return;
    }
    if (!navigator.mediaDevices?.getUserMedia) {
      setState({ kind: "error", message: "microphone capture is unavailable" });
      return;
    }

    const connectionGeneration = connectionGenerationRef.current;
    const request = navigator.mediaDevices
      .getUserMedia({ audio: true })
      .then((stream) => {
        if (connectionGenerationRef.current !== connectionGeneration) {
          stream.getTracks().forEach((track) => track.stop());
          return;
        }
        streamRef.current = stream;
        setState({ kind: "idle" });
      })
      .catch((err: unknown) => {
        if (connectionGenerationRef.current === connectionGeneration) {
          setState({ kind: "error", message: errorMessage(err) });
        }
      })
      .finally(() => {
        if (connectPromiseRef.current === request) {
          connectPromiseRef.current = null;
        }
      });
    connectPromiseRef.current = request;
    await request;
  }, []);

  const getMediaStream = useCallback(() => streamRef.current, []);

  const discardUtterance = useCallback(() => {
    discardRecorder();
    setState({ kind: "idle" });
  }, [discardRecorder]);

  const releaseMicrophone = useCallback(() => {
    releaseResources(true);
  }, [releaseResources]);

  const reset = releaseMicrophone;

  const startRecording = useCallback(async () => {
    if (!client) {
      setState({ kind: "error", message: "practice phone client is missing" });
      return;
    }
    if (!sessionId || !turnId) {
      setState({ kind: "error", message: "practice phone session is not ready" });
      return;
    }
    if (typeof MediaRecorder === "undefined") {
      setState({ kind: "error", message: "browser audio recorder is unavailable" });
      return;
    }
    const stream = streamRef.current;
    if (!stream) {
      setState({ kind: "error", message: "microphone is not connected" });
      return;
    }
    if (recorderRef.current?.state === "recording") {
      setState({ kind: "error", message: "recording is already active" });
      return;
    }

    try {
      chunksRef.current = [];
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
      discardRecorder();
      setState({ kind: "error", message: errorMessage(err) });
    }
  }, [client, discardRecorder, sessionId, turnId]);

  const stopAndSubmit = useCallback(async (
    options: StopPracticeVoiceTurnOptions = {},
  ): Promise<PracticeVoiceTurnResult> => {
    if (!client) throw new Error("usePracticeVoiceTurn: client missing");
    if (!sessionId) throw new Error("usePracticeVoiceTurn: sessionId missing");
    if (!turnId) throw new Error("usePracticeVoiceTurn: turnId missing");
    const recorder = recorderRef.current;
    if (!recorder || recorder.state !== "recording") {
      const err = new Error("recording is not active");
      setState({ kind: "error", message: err.message });
      throw err;
    }

    const connectionGeneration = connectionGenerationRef.current;
    let releasedAfterSnapshot = false;
    try {
      snapshotPendingRef.current = true;
      setState({ kind: "submitting" });
      await stopRecorder(recorder);
      if (recorderRef.current === recorder) recorderRef.current = null;
      recorder.ondataavailable = null;
      recorder.onstop = null;
      recorder.onerror = null;
      const durationMs = Math.max(1, Date.now() - startedAtRef.current);
      const chunks = chunksRef.current;
      chunksRef.current = [];
      startedAtRef.current = 0;
      const audioBlob = new Blob(chunks, {
        type: normalizeAudioContentType(
          chunks[0]?.type || recorder.mimeType,
        ),
      });
      snapshotPendingRef.current = false;
      const releaseAfterSnapshot =
        options.releaseMicrophoneAfterSnapshot || releaseRequestedRef.current;
      const updateStateAfterRelease =
        options.releaseMicrophoneAfterSnapshot ||
        releaseStateRequestedRef.current;
      releaseRequestedRef.current = false;
      releaseStateRequestedRef.current = false;
      if (releaseAfterSnapshot) {
        releaseResources(updateStateAfterRelease);
        releasedAfterSnapshot = true;
      }
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
      };
      await options.beforeSubmit;
      const result = await client.createPracticeVoiceTurn(sessionId, body, {
        idempotencyKey: generateIdempotencyKey(),
      });
      if (connectionGenerationRef.current === connectionGeneration) {
        setState({ kind: "success", result });
      }
      return result;
    } catch (err) {
      const releaseAfterFailure =
        options.releaseMicrophoneAfterSnapshot || releaseRequestedRef.current;
      const updateStateAfterFailure =
        options.releaseMicrophoneAfterSnapshot ||
        releaseStateRequestedRef.current;
      snapshotPendingRef.current = false;
      releaseRequestedRef.current = false;
      releaseStateRequestedRef.current = false;
      if (recorderRef.current === recorder) discardRecorder();
      if (releaseAfterFailure && !releasedAfterSnapshot) {
        releaseResources(updateStateAfterFailure);
      }
      const message = errorMessage(err);
      if (connectionGenerationRef.current === connectionGeneration) {
        setState({ kind: "error", message });
      }
      throw err instanceof Error ? err : new Error(message);
    }
  }, [
    client,
    discardRecorder,
    lang,
    practiceMode,
    releaseResources,
    sessionId,
    turnId,
  ]);

  return useMemo<UsePracticeVoiceTurnResult>(
    () => ({
      ready: !!client && !!sessionId && !!turnId,
      state,
      connectMicrophone,
      getMediaStream,
      startRecording,
      stopAndSubmit,
      discardUtterance,
      releaseMicrophone,
      reset,
    }),
    [
      client,
      connectMicrophone,
      discardUtterance,
      getMediaStream,
      releaseMicrophone,
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

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}
