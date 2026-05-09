import { useCallback, useEffect, useRef, useState } from "react";

import type { PracticePlan } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useInterviewContext } from "../../../interview-context/InterviewContext";

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
  const { ctx, dispatch } = useInterviewContext();
  const planId = ctx.planId;

  const [loading, setLoading] = useState(!!planId);
  const [data, setData] = useState<PracticePlan | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [ready, setReady] = useState(false);
  const inFlightRef = useRef(false);

  const fetch = useCallback(() => {
    if (!runtime || !planId) {
      setLoading(false);
      return;
    }

    if (inFlightRef.current) return;

    let cancelled = false;
    inFlightRef.current = true;
    setLoading(true);
    setError(null);

    runtime.client
      .getPracticePlan(planId)
      .then((plan) => {
        if (cancelled) return;
        setData(plan);
        const isReady = plan.status === "ready";
        setReady(isReady);
        if (isReady) {
          dispatch({ type: "MERGE_PRACTICE_PLAN", plan: plan as unknown as { id: string; [key: string]: unknown } });
        }
        // archived or other non-ready → don't merge, caller resets planId
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        setError(err instanceof Error ? err : new Error(String(err)));
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
          inFlightRef.current = false;
        }
      });
  }, [runtime, planId, dispatch]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { loading, data, error, ready };
}
