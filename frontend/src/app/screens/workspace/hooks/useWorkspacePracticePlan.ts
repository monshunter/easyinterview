import { useEffect, useRef, useState } from "react";

import type { PracticePlan } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import { normalizeServerBoundId } from "../../../interview-context/apiIds";

export interface UseWorkspacePracticePlanResult {
  loading: boolean;
  data: PracticePlan | null;
  error: Error | null;
  ready: boolean;
}

/**
 * Phase 4 hook: refreshes PracticePlan data on mount when InterviewContext.planId
 * exists. status='ready' → MERGE_PRACTICE_PLAN. status='archived' or 404 →
 * reset planId=null. Does NOT assume un-declared plan statuses.
 */
export function useWorkspacePracticePlan(): UseWorkspacePracticePlanResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const { ctx, dispatch } = useInterviewContext();
  const planId = normalizeServerBoundId(ctx.planId);

  const [loading, setLoading] = useState(!!planId);
  const [data, setData] = useState<PracticePlan | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [ready, setReady] = useState(false);
  const requestSeqRef = useRef(0);

  useEffect(() => {
    if (!client || !planId) {
      setLoading(false);
      setData(null);
      setError(null);
      setReady(false);
      if (ctx.planId && !planId) {
        dispatch({ type: "CLEAR_PRACTICE_PLAN" });
      }
      return;
    }

    let active = true;
    const requestSeq = requestSeqRef.current + 1;
    requestSeqRef.current = requestSeq;
    setLoading(true);
    setData(null);
    setError(null);

    client
      .getPracticePlan(planId)
      .then((plan) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setData(plan);
        const isReady = plan.status === "ready";
        setReady(isReady);
        if (isReady) {
          dispatch({
            type: "MERGE_PRACTICE_PLAN",
            plan: plan as unknown as { id: string; [key: string]: unknown },
          });
        } else {
          dispatch({ type: "CLEAR_PRACTICE_PLAN" });
        }
      })
      .catch((err: unknown) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        const error = err instanceof Error ? err : new Error(String(err));
        setData(null);
        setError(error);
        setReady(false);
        if (isNotFound(error)) {
          dispatch({ type: "CLEAR_PRACTICE_PLAN" });
        }
      })
      .finally(() => {
        if (active && requestSeqRef.current === requestSeq) {
          setLoading(false);
        }
      });
    return () => {
      active = false;
    };
  }, [client, planId, ctx.planId, dispatch]);

  return { loading, data, error, ready };
}

function isNotFound(error: Error): boolean {
  return /^HTTP 404\b/.test(error.message);
}
