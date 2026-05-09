import { useEffect, useRef, useState } from "react";

import type { ResumeAsset } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import { normalizeServerBoundId } from "../../../interview-context/apiIds";

export interface UseWorkspaceResumeResult {
  loading: boolean;
  data: ResumeAsset | null;
  error: Error | null;
  empty: boolean;
}

/**
 * Phase 3 hook: loads ResumeAsset via generated client when
 * InterviewContext.resumeVersionId exists, writing results through
 * MERGE_RESUME. Returns empty state immediately when resumeVersionId
 * is missing (no request sent). On 404, sets resumeVersionId=null
 * to trigger WorkspaceMissingResumeState.
 */
export function useWorkspaceResume(): UseWorkspaceResumeResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const { ctx, dispatch } = useInterviewContext();
  const resumeVersionId = normalizeServerBoundId(ctx.resumeVersionId);

  const [loading, setLoading] = useState(!!resumeVersionId);
  const [data, setData] = useState<ResumeAsset | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const requestSeqRef = useRef(0);

  useEffect(() => {
    if (!client || !resumeVersionId) {
      setLoading(false);
      setData(null);
      setError(null);
      if (ctx.resumeVersionId && !resumeVersionId) {
        dispatch({ type: "CLEAR_RESUME" });
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
      .getResume(resumeVersionId)
      .then((resume) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setData(resume);
        setError(null);
        dispatch({
          type: "MERGE_RESUME",
          resume: resume as unknown as { id: string; [key: string]: unknown },
        });
      })
      .catch((err: unknown) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        const error = err instanceof Error ? err : new Error(String(err));
        setData(null);
        setError(error);
        if (isNotFound(error)) {
          dispatch({ type: "CLEAR_RESUME" });
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
  }, [client, resumeVersionId, ctx.resumeVersionId, dispatch]);

  return { loading, data, error, empty: !resumeVersionId };
}

function isNotFound(error: Error): boolean {
  return /^HTTP 404\b/.test(error.message);
}
