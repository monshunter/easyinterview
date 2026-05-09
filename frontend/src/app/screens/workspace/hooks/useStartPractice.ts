import { useCallback, useRef, useState } from "react";

import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import { buildCreatePlanRequest } from "../../../interview-context/buildCreatePlanRequest";
import { newIdempotencyBatch } from "../../../../lib/conventions/idempotency";
import { useI18n } from "../../../i18n/messages";

export type StartState =
  | { kind: "idle" }
  | { kind: "loading" }
  | { kind: "error"; message: string; retryable: boolean }
  | { kind: "success"; sessionId: string };

/**
 * Phase 4: Implements the dual-step "Start Interview" contract:
 * 1. If no planId or plan not ready → createPracticePlan (with Idempotency-Key)
 * 2. startPracticeSession (with Idempotency-Key)
 * 3. On success → navigate to practice
 * Retries reuse the same idempotency batch for dedupe.
 */
export function useStartPractice() {
  const runtime = useAppRuntimeOptional();
  const { ctx, dispatch } = useInterviewContext();
  const { lang } = useI18n();

  const [state, setState] = useState<StartState>({ kind: "idle" });
  const batchRef = useRef<{ create: string; start: string } | null>(null);
  const attemptRef = useRef(0);
  const inFlightRef = useRef(false);

  const start = useCallback(async (): Promise<StartState> => {
    if (!runtime || inFlightRef.current) return state;

    inFlightRef.current = true;
    setState({ kind: "loading" });

    // Generate stable idempotency batch on first attempt; retries reuse
    if (!batchRef.current) {
      batchRef.current = newIdempotencyBatch();
    }
    const batch = batchRef.current;

    try {
      let planId = ctx.planId;
      if (!planId) {
        const plan = await runtime.client.createPracticePlan(
          buildCreatePlanRequest(ctx, lang),
          { idempotencyKey: batch.create },
        );
        planId = plan.id;
        dispatch({ type: "MERGE_PRACTICE_PLAN", plan: plan as unknown as { id: string; [key: string]: unknown } });
      }

      const session = await runtime.client.startPracticeSession(
        { planId: planId!, hintsEnabled: ctx.practiceMode === "assisted" },
        { idempotencyKey: batch.start },
      );

      dispatch({ type: "MERGE_SESSION", session: session as unknown as { id: string; [key: string]: unknown } });

      setState({ kind: "success", sessionId: session.id });
      inFlightRef.current = false;
      return { kind: "success" as const, sessionId: session.id };
    } catch (err: unknown) {
      attemptRef.current += 1;
      const message = err instanceof Error ? err.message : String(err);
      const retryable = attemptRef.current < 3;
      const result: StartState = { kind: "error", message, retryable };
      setState(result);
      inFlightRef.current = false;
      return result;
    }
  }, [runtime, ctx, lang, dispatch, state]);

  const reset = useCallback(() => {
    batchRef.current = null;
    attemptRef.current = 0;
    setState({ kind: "idle" });
    inFlightRef.current = false;
  }, []);

  return { state, start, reset };
}
