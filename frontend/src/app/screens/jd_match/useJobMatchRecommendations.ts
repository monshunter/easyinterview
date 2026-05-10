import { useCallback, useEffect, useRef, useState } from "react";

import type {
  JobMatchRecommendation,
  PageInfo,
} from "../../../api/generated/types";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";

export interface UseJobMatchRecommendationsResult {
  loading: boolean;
  items: JobMatchRecommendation[];
  error: Error | null;
  pageInfo: PageInfo | null;
  retry: () => void;
}

/**
 * Phase 3.2 hook: loads JobMatchRecommendation list via generated client on
 * mount and exposes a `retry` callback that refetches without remounting.
 * Returns inert state when no AppRuntimeProvider is mounted.
 */
export function useJobMatchRecommendations(): UseJobMatchRecommendationsResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const [loading, setLoading] = useState<boolean>(!!client);
  const [items, setItems] = useState<JobMatchRecommendation[]>([]);
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [retryNonce, setRetryNonce] = useState(0);
  const requestSeqRef = useRef(0);

  useEffect(() => {
    if (!client) {
      setLoading(false);
      setItems([]);
      setPageInfo(null);
      setError(null);
      return;
    }

    let active = true;
    const requestSeq = requestSeqRef.current + 1;
    requestSeqRef.current = requestSeq;
    setLoading(true);
    setError(null);

    client
      .listJobRecommendations()
      .then((page) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setItems(page.items);
        setPageInfo(page.pageInfo);
        setError(null);
      })
      .catch((err: unknown) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setItems([]);
        setPageInfo(null);
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
  }, [client, retryNonce]);

  const retry = useCallback(() => {
    setRetryNonce((n) => n + 1);
  }, []);

  return { loading, items, error, pageInfo, retry };
}
