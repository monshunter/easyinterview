import { useCallback, useEffect, useRef, useState } from "react";

import type {
  SuggestDebriefQuestionsRequest,
  SuggestedDebriefQuestion,
} from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export interface UseSuggestDebriefQuestionsArgs {
  targetJobId?: string | null;
  sessionId?: string | null;
  resumeVersionId?: string | null;
  language: string;
  count?: number;
  enabled: boolean;
}

export interface SuggestDebriefError {
  code: string;
  message: string;
}

export interface UseSuggestDebriefQuestionsState {
  suggestions: SuggestedDebriefQuestion[] | null;
  loading: boolean;
  error: SuggestDebriefError | null;
  refetch: () => void;
}

const DEBOUNCE_MS = 500;

function parseError(err: unknown): SuggestDebriefError {
  if (typeof err === "object" && err !== null && "code" in err) {
    const e = err as { code?: string; message?: string };
    return {
      code: e.code ?? "UNKNOWN",
      message: e.message ?? "unknown error",
    };
  }
  if (err instanceof Error) {
    const match = /HTTP \d{3}\s+([A-Z_]+)/.exec(err.message);
    return {
      code: match?.[1] ?? "UNKNOWN",
      message: err.message,
    };
  }
  return { code: "UNKNOWN", message: String(err) };
}

/**
 * Phase 4.1 — async loader for AI-suggested debrief questions. Debounces on
 * input change by 500 ms so rapid context-strip edits coalesce into a single
 * call. Exposes `refetch` so the "重新生成推荐" CTA can re-trigger without
 * changing inputs.
 */
export function useSuggestDebriefQuestions({
  targetJobId,
  sessionId,
  resumeVersionId,
  language,
  count = 6,
  enabled,
}: UseSuggestDebriefQuestionsArgs): UseSuggestDebriefQuestionsState {
  const runtime = useAppRuntimeOptional();
  const [state, setState] = useState<UseSuggestDebriefQuestionsState>({
    suggestions: null,
    loading: false,
    error: null,
    refetch: () => undefined,
  });
  const fetchTokenRef = useRef(0);

  const run = useCallback(async () => {
    if (!runtime || !enabled || !targetJobId) return;
    const token = ++fetchTokenRef.current;
    setState((prev) => ({ ...prev, loading: true, error: null }));
    const body: SuggestDebriefQuestionsRequest = {
      targetJobId,
      sessionId: sessionId ?? undefined,
      resumeVersionId: resumeVersionId ?? undefined,
      language,
      count,
    } as SuggestDebriefQuestionsRequest;
    try {
      const res = await runtime.client.suggestDebriefQuestions(body);
      if (token !== fetchTokenRef.current) return;
      setState((prev) => ({
        ...prev,
        suggestions: res.suggestions ?? [],
        loading: false,
        error: null,
      }));
    } catch (err) {
      if (token !== fetchTokenRef.current) return;
      setState((prev) => ({
        ...prev,
        loading: false,
        error: parseError(err),
      }));
    }
  }, [runtime, enabled, targetJobId, sessionId, resumeVersionId, language, count]);

  useEffect(() => {
    if (!enabled || !targetJobId) {
      setState((prev) => ({ ...prev, suggestions: null, loading: false, error: null }));
      return;
    }
    const handle = window.setTimeout(() => {
      void run();
    }, DEBOUNCE_MS);
    return () => window.clearTimeout(handle);
  }, [enabled, run, targetJobId, sessionId, resumeVersionId, language, count]);

  const refetch = useCallback(() => {
    void run();
  }, [run]);

  return { ...state, refetch };
}
