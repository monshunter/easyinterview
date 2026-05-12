import { useCallback, useEffect, useRef, useState } from "react";

import type { PaginatedResumeVersion } from "../../../../api/generated/types";
import { useDisplayPreferencesOptional } from "../../../display/DisplayPreferencesProvider";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export interface UseResumeVersionsResult {
  loading: boolean;
  data: PaginatedResumeVersion | null;
  error: Error | null;
  retry: () => void;
}

/**
 * Phase 1 fixture-backed transport ignores the path-param `resumeAssetId`
 * and returns the active scenario's version collection wholesale. The screen
 * groups the response client-side by `version.resumeAssetId`. We still pass
 * a non-empty resumeAssetId so the generated client builds a valid URL.
 */
export function useResumeVersions(
  primerAssetId: string | null,
): UseResumeVersionsResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const isAuthenticated = runtime?.auth.status === "authenticated";
  const lang = useDisplayPreferencesOptional()?.lang ?? "zh";

  const [loading, setLoading] = useState<boolean>(
    !!client && isAuthenticated && !!primerAssetId,
  );
  const [data, setData] = useState<PaginatedResumeVersion | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [reloadSeq, setReloadSeq] = useState(0);
  const requestSeqRef = useRef(0);

  const retry = useCallback(() => {
    setReloadSeq((value) => value + 1);
  }, []);

  useEffect(() => {
    if (!client || !isAuthenticated || !primerAssetId) {
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
      .listResumeVersions(primerAssetId, {
        headers: { "Accept-Language": lang },
      })
      .then((paginated) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setData(paginated);
      })
      .catch((err: unknown) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
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
  }, [client, isAuthenticated, primerAssetId, reloadSeq, lang]);

  return { loading, data, error, retry };
}
