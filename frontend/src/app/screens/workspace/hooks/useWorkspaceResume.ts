import { useCallback, useEffect, useRef, useState } from "react";

import type { ResumeAsset } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useInterviewContext } from "../../../interview-context/InterviewContext";

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
  const { ctx, dispatch } = useInterviewContext();
  const resumeVersionId = ctx.resumeVersionId;

  const [loading, setLoading] = useState(!!resumeVersionId);
  const [data, setData] = useState<ResumeAsset | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const inFlightRef = useRef(false);

  const fetch = useCallback(() => {
    if (!runtime || !resumeVersionId) {
      setLoading(false);
      return;
    }

    if (inFlightRef.current) return;

    let cancelled = false;
    inFlightRef.current = true;
    setLoading(true);
    setError(null);

    runtime.client
      .getResume(resumeVersionId)
      .then((resume) => {
        if (cancelled) return;
        setData(resume);
        setError(null);
        dispatch({ type: "MERGE_RESUME", resume: resume as unknown as { id: string; [key: string]: unknown } });
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        const error = err instanceof Error ? err : new Error(String(err));
        setError(error);
        if (isNotFound(error)) {
          dispatch({ type: "CLEAR_RESUME" });
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
          inFlightRef.current = false;
        }
      });
  }, [runtime, resumeVersionId, dispatch]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { loading, data, error, empty: !resumeVersionId };
}

function isNotFound(error: Error): boolean {
  return /^HTTP 404\b/.test(error.message);
}
