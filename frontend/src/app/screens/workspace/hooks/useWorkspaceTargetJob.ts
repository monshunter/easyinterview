import { useCallback, useEffect, useRef, useState } from "react";

import type { TargetJob } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useInterviewContext } from "../../../interview-context/InterviewContext";

export interface UseWorkspaceTargetJobResult {
  loading: boolean;
  data: TargetJob | null;
  error: Error | null;
  empty: boolean;
}

/**
 * Phase 2 hook: loads TargetJob data via generated client when
 * InterviewContext.targetJobId exists, writing results through
 * MERGE_TARGET_JOB. Returns empty state immediately when targetJobId
 * is missing (no request sent).
 */
export function useWorkspaceTargetJob(): UseWorkspaceTargetJobResult {
  const runtime = useAppRuntimeOptional();
  const { ctx, dispatch } = useInterviewContext();
  const targetJobId = ctx.targetJobId;

  const [loading, setLoading] = useState(!!targetJobId);
  const [data, setData] = useState<TargetJob | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const inFlightRef = useRef(false);

  const fetch = useCallback(() => {
    if (!runtime || !targetJobId) {
      setLoading(false);
      return;
    }

    if (inFlightRef.current) return;

    let cancelled = false;
    inFlightRef.current = true;
    setLoading(true);
    setError(null);

    runtime.client
      .getTargetJob(targetJobId)
      .then((job) => {
        if (cancelled) return;
        setData(job);
        setError(null);
        dispatch({ type: "MERGE_TARGET_JOB", targetJob: job as unknown as { id: string; [key: string]: unknown } });
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
  }, [runtime, targetJobId, dispatch]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { loading, data, error, empty: !targetJobId };
}
