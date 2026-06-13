import { useEffect, useRef, useState } from "react";

import type { Resume } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import { normalizeServerBoundId } from "../../../interview-context/apiIds";

export interface UseWorkspaceResumeResult {
  loading: boolean;
  data: Resume | null;
  error: Error | null;
  empty: boolean;
}

/**
 * Phase 3 hook: loads the flat Resume via generated client when
 * InterviewContext.resumeId exists, writing results through
 * MERGE_RESUME. Returns empty state immediately when resumeId
 * is missing (no request sent). On 404, sets resumeId=null
 * to trigger WorkspaceMissingResumeState.
 */
export function useWorkspaceResume(): UseWorkspaceResumeResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const { ctx, dispatch } = useInterviewContext();
  const resumeId = normalizeServerBoundId(ctx.resumeId);

  const [loading, setLoading] = useState(!!resumeId);
  const [data, setData] = useState<Resume | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const requestSeqRef = useRef(0);

  useEffect(() => {
    if (!client || !resumeId) {
      setLoading(false);
      setData(null);
      setError(null);
      if (ctx.resumeId && !resumeId) {
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
      .getResume(resumeId)
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
  }, [client, resumeId, ctx.resumeId, dispatch]);

  return { loading, data, error, empty: !resumeId };
}

function isNotFound(error: Error): boolean {
  return /^HTTP 404\b/.test(error.message);
}
