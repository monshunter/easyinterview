import { useCallback, useEffect, useRef, useState } from "react";

import type { JobMatchRecommendation } from "../../../api/generated/types";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";

export interface UseJobRecommendationResult {
  loading: boolean;
  data: JobMatchRecommendation | null;
  error: Error | null;
  retry: () => void;
}

/**
 * Loads the selected recommendation detail through the required detail
 * operation. The list response is only a summary source for cards.
 */
export function useJobRecommendation(
  jobMatchId: string | null,
): UseJobRecommendationResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const [loading, setLoading] = useState<boolean>(!!client && !!jobMatchId);
  const [data, setData] = useState<JobMatchRecommendation | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [retryNonce, setRetryNonce] = useState(0);
  const requestSeqRef = useRef(0);

  useEffect(() => {
    if (!client || !jobMatchId) {
      setLoading(false);
      setData(null);
      setError(null);
      return;
    }

    let active = true;
    const requestSeq = requestSeqRef.current + 1;
    requestSeqRef.current = requestSeq;
    setLoading(true);
    setError(null);

    client
      .getJobRecommendation(jobMatchId)
      .then((detail) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setData(detail);
        setError(null);
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
  }, [client, jobMatchId, retryNonce]);

  const retry = useCallback(() => {
    setRetryNonce((n) => n + 1);
  }, []);

  return { loading, data, error, retry };
}
