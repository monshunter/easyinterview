import { useCallback, useEffect, useRef, useState } from "react";

import type { ResumeAsset } from "../../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../../runtime/AppRuntimeProvider";
import { useDisplayPreferencesOptional } from "../../../../display/DisplayPreferencesProvider";

export interface UseResumeParsingPollingOptions {
  initialDelayMs?: number;
  backoffFactor?: number;
  maxAttempts?: number;
  maxTotalMs?: number;
}

export interface ParsingPollingSnapshot {
  status: "idle" | "polling" | "ready" | "failed";
  attempts: number;
  asset: ResumeAsset | null;
  errorCode: string | null;
}

export interface UseResumeParsingPollingResult {
  snapshot: ParsingPollingSnapshot;
  retry: () => void;
  cancel: () => void;
}

const DEFAULT_INITIAL_DELAY_MS = 1_500;
const DEFAULT_BACKOFF = 1.4;
const DEFAULT_MAX_ATTEMPTS = 8;
const DEFAULT_MAX_TOTAL_MS = 30_000;

export function useResumeParsingPolling(
  resumeAssetId: string | null,
  options: UseResumeParsingPollingOptions = {},
): UseResumeParsingPollingResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;
  const lang = useDisplayPreferencesOptional()?.lang ?? "zh";
  const initialDelayMs = options.initialDelayMs ?? DEFAULT_INITIAL_DELAY_MS;
  const backoffFactor = options.backoffFactor ?? DEFAULT_BACKOFF;
  const maxAttempts = options.maxAttempts ?? DEFAULT_MAX_ATTEMPTS;
  const maxTotalMs = options.maxTotalMs ?? DEFAULT_MAX_TOTAL_MS;

  const [snapshot, setSnapshot] = useState<ParsingPollingSnapshot>({
    status: resumeAssetId ? "polling" : "idle",
    attempts: 0,
    asset: null,
    errorCode: null,
  });
  const [retryEpoch, setRetryEpoch] = useState(0);
  const cancelRef = useRef(false);
  const sessionRef = useRef(0);

  useEffect(() => {
    if (!resumeAssetId || !client) {
      setSnapshot({ status: "idle", attempts: 0, asset: null, errorCode: null });
      return;
    }
    cancelRef.current = false;
    const sessionId = sessionRef.current + 1;
    sessionRef.current = sessionId;
    setSnapshot({
      status: "polling",
      attempts: 0,
      asset: null,
      errorCode: null,
    });

    const started = Date.now();
    let timer: number | undefined;
    let attempt = 0;
    let delay = initialDelayMs;

    const poll = async () => {
      if (cancelRef.current || sessionRef.current !== sessionId) return;
      attempt += 1;
      try {
        const asset = await client.getResume(resumeAssetId, {
          headers: { "Accept-Language": lang },
        });
        if (cancelRef.current || sessionRef.current !== sessionId) return;
        if (asset.parseStatus === "ready") {
          setSnapshot({
            status: "ready",
            attempts: attempt,
            asset,
            errorCode: null,
          });
          return;
        }
        if (asset.parseStatus === "failed") {
          setSnapshot({
            status: "failed",
            attempts: attempt,
            asset,
            errorCode: "AI_TIMEOUT_RETRYABLE",
          });
          return;
        }
        if (attempt >= maxAttempts || Date.now() - started >= maxTotalMs) {
          setSnapshot({
            status: "failed",
            attempts: attempt,
            asset,
            errorCode: "PARSE_TIMEOUT",
          });
          return;
        }
        setSnapshot((prev) => ({ ...prev, attempts: attempt, asset }));
        timer = window.setTimeout(poll, delay);
        delay = delay * backoffFactor;
      } catch (error) {
        if (cancelRef.current || sessionRef.current !== sessionId) return;
        setSnapshot({
          status: "failed",
          attempts: attempt,
          asset: null,
          errorCode: error instanceof Error ? error.message : "UNKNOWN",
        });
      }
    };

    void poll();

    return () => {
      cancelRef.current = true;
      if (typeof timer === "number") window.clearTimeout(timer);
    };
  }, [
    backoffFactor,
    client,
    initialDelayMs,
    lang,
    maxAttempts,
    maxTotalMs,
    resumeAssetId,
    retryEpoch,
  ]);

  const retry = useCallback(() => {
    if (!resumeAssetId) return;
    cancelRef.current = true;
    setSnapshot({
      status: "polling",
      attempts: 0,
      asset: null,
      errorCode: null,
    });
    setRetryEpoch((value) => value + 1);
  }, [resumeAssetId]);

  const cancel = useCallback(() => {
    cancelRef.current = true;
    setSnapshot({ status: "idle", attempts: 0, asset: null, errorCode: null });
  }, []);

  return { snapshot, retry, cancel };
}
