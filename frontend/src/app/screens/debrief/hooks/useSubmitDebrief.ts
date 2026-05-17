import { useCallback, useState } from "react";

import type {
  CreateDebriefRequest,
  DebriefQuestionInput,
  DebriefRoundType,
  DebriefWithJob,
  InterviewerRole,
} from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import type { DebriefEntry } from "../types";

export interface SubmitDebriefArgs {
  targetJobId: string;
  roundType: DebriefRoundType;
  interviewerRole?: InterviewerRole;
  language: string;
  entries: DebriefEntry[];
  notes?: string;
}

export type SubmitDebriefStatus =
  | "idle"
  | "submitting"
  | "succeeded"
  | "auth_required"
  | "validation_failed"
  | "failed";

export interface SubmitDebriefError {
  code: string;
  message: string;
  details?: unknown;
}

export interface UseSubmitDebriefState {
  status: SubmitDebriefStatus;
  error: SubmitDebriefError | null;
  result: DebriefWithJob | null;
  /** Submit the debrief; resolves to the submission outcome, never throws. */
  submit: (args: SubmitDebriefArgs) => Promise<{
    status: SubmitDebriefStatus;
    error?: SubmitDebriefError;
    result?: DebriefWithJob;
  }>;
  reset: () => void;
}

function generateIdempotencyKey(): string {
  if (typeof globalThis.crypto?.randomUUID === "function") {
    return globalThis.crypto.randomUUID();
  }
  // Fallback: timestamp + random suffix — only used when crypto.randomUUID
  // is unavailable (older jsdom). Idempotency keys are server-validated.
  return `ik-${Date.now().toString(16)}-${Math.random().toString(16).slice(2, 10)}`;
}

function entriesToQuestionInputs(entries: DebriefEntry[]): DebriefQuestionInput[] {
  return entries.map((entry) => ({
    questionText: entry.questionText.trim(),
    myAnswerSummary: (entry.myAnswerSummary ?? "").trim(),
    interviewerReaction: (
      entry.interviewerReaction ??
      entry.reflection ??
      ""
    ).trim(),
  }));
}

function parseError(err: unknown): SubmitDebriefError {
  if (typeof err === "object" && err !== null) {
    const e = err as { code?: string; message?: string; status?: number; details?: unknown };
    if (e.code) {
      return { code: e.code, message: e.message ?? "submit failed", details: e.details };
    }
  }
  if (err instanceof Error) {
    const match = /HTTP (\d{3})\s*([A-Z_]+)?/.exec(err.message);
    if (match) {
      const status = match[1];
      const code = match[2];
      if (status === "401") return { code: "UNAUTHENTICATED", message: err.message };
      if (status === "409") return { code: "IDEMPOTENCY_KEY_MISMATCH", message: err.message };
      if (status === "422") return { code: code ?? "VALIDATION_FAILED", message: err.message };
      return { code: code ?? "UNKNOWN", message: err.message };
    }
    return { code: "UNKNOWN", message: err.message };
  }
  return { code: "UNKNOWN", message: String(err) };
}

/**
 * Phase 5.1 — `createDebrief` submission. Generates a fresh
 * `Idempotency-Key` per attempt, maps debrief entries into the wire shape,
 * and writes `debriefId` / `debriefJobId` into `InterviewContext` via the
 * Phase 5.4 `SET_DEBRIEF_CONTEXT` reducer action without touching `jobId`.
 */
export function useSubmitDebrief(): UseSubmitDebriefState {
  const runtime = useAppRuntimeOptional();
  const { dispatch } = useInterviewContext();
  const [state, setState] = useState<{
    status: SubmitDebriefStatus;
    error: SubmitDebriefError | null;
    result: DebriefWithJob | null;
  }>({ status: "idle", error: null, result: null });

  const submit = useCallback<UseSubmitDebriefState["submit"]>(
    async (args) => {
      if (!runtime) {
        const error: SubmitDebriefError = {
          code: "RUNTIME_UNAVAILABLE",
          message: "runtime is not mounted",
        };
        setState({ status: "failed", error, result: null });
        return { status: "failed", error };
      }
      const ik = generateIdempotencyKey();
      const body: CreateDebriefRequest = {
        targetJobId: args.targetJobId,
        roundType: args.roundType,
        interviewerRole: args.interviewerRole,
        language: args.language,
        questions: entriesToQuestionInputs(args.entries),
        notes: args.notes,
      };
      setState({ status: "submitting", error: null, result: null });
      const attempt = async (
        key: string,
      ): Promise<{ ok: true; data: DebriefWithJob } | { ok: false; error: SubmitDebriefError }> => {
        try {
          const data = await runtime.client.createDebrief(body, {
            headers: { "Idempotency-Key": key },
          });
          return { ok: true, data };
        } catch (err) {
          return { ok: false, error: parseError(err) };
        }
      };
      let first = await attempt(ik);
      if (!first.ok && first.error.code === "IDEMPOTENCY_KEY_MISMATCH") {
        // Phase 5.1 — auto-retry once with a fresh IK on 409 mismatch.
        first = await attempt(generateIdempotencyKey());
      }
      if (first.ok) {
        dispatch({
          type: "SET_DEBRIEF_CONTEXT",
          payload: {
            debriefId: first.data.debriefId,
            debriefJobId: first.data.job?.id,
            practiceGoal: "debrief",
          },
        });
        setState({ status: "succeeded", error: null, result: first.data });
        return { status: "succeeded", result: first.data };
      }
      const code = first.error.code;
      let status: SubmitDebriefStatus = "failed";
      if (code === "UNAUTHENTICATED") status = "auth_required";
      else if (code === "VALIDATION_FAILED") status = "validation_failed";
      setState({ status, error: first.error, result: null });
      return { status, error: first.error };
    },
    [dispatch, runtime],
  );

  const reset = useCallback(() => {
    setState({ status: "idle", error: null, result: null });
  }, []);

  return { ...state, submit, reset };
}
