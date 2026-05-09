import { useCallback, useEffect, useRef, useState } from "react";

import type { TargetJob } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import { normalizeServerBoundId } from "../../../interview-context/apiIds";

export interface UseWorkspaceTargetJobResult {
  loading: boolean;
  data: TargetJob | null;
  error: Error | null;
  empty: boolean;
  notFound: boolean;
  retry: () => void;
}

/**
 * Phase 2 hook: loads TargetJob data via generated client when
 * InterviewContext.targetJobId exists, writing results through
 * MERGE_TARGET_JOB. Returns empty state immediately when targetJobId
 * is missing (no request sent).
 */
export function useWorkspaceTargetJob(): UseWorkspaceTargetJobResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const { ctx, dispatch } = useInterviewContext();
  const targetJobId = normalizeServerBoundId(ctx.targetJobId);

  const [loading, setLoading] = useState(!!targetJobId);
  const [data, setData] = useState<TargetJob | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [reloadSeq, setReloadSeq] = useState(0);
  const requestSeqRef = useRef(0);

  const retry = useCallback(() => {
    setReloadSeq((value) => value + 1);
  }, []);

  useEffect(() => {
    if (!client || !targetJobId) {
      setLoading(false);
      setData(null);
      setError(null);
      return;
    }

    let active = true;
    const requestSeq = requestSeqRef.current + 1;
    requestSeqRef.current = requestSeq;
    setLoading(true);
    setData(null);
    setError(null);

    client
      .getTargetJob(targetJobId)
      .then((job) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setData(job);
        setError(null);
        dispatch({
          type: "MERGE_TARGET_JOB",
          targetJob: job as unknown as { id: string; [key: string]: unknown },
        });
      })
      .catch((err: unknown) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setData(null);
        setError(err instanceof Error ? err : new Error(String(err)));
      })
      .finally(() => {
        if (active && requestSeqRef.current === requestSeq) {
          setLoading(false);
        }
      });
    return () => {
      active = false;
    };
  }, [client, targetJobId, dispatch, reloadSeq]);

  const notFound = error ? isHttpStatus(error, 404) : false;
  return { loading, data, error, empty: !targetJobId, notFound, retry };
}

function isHttpStatus(error: Error, status: number): boolean {
  return error.message.startsWith(`HTTP ${status} `);
}
