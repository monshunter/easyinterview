import { useCallback, useEffect, useRef, useState } from "react";

import type { PaginatedResume } from "../../../../api/generated/types";
import { useDisplayPreferencesOptional } from "../../../display/DisplayPreferencesProvider";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export interface UseResumesResult {
  loading: boolean;
  data: PaginatedResume | null;
  error: Error | null;
  retry: () => void;
}

export function useResumes(): UseResumesResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const isAuthenticated = runtime?.auth.status === "authenticated";
  const lang = useDisplayPreferencesOptional()?.lang ?? "zh";

  const [loading, setLoading] = useState<boolean>(!!client && isAuthenticated);
  const [data, setData] = useState<PaginatedResume | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [reloadSeq, setReloadSeq] = useState(0);
  const requestSeqRef = useRef(0);

  const retry = useCallback(() => {
    setReloadSeq((value) => value + 1);
  }, []);

  useEffect(() => {
    if (!client || !isAuthenticated) {
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
      .listResumes({ headers: { "Accept-Language": lang } })
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
  }, [client, isAuthenticated, reloadSeq, lang]);

  return { loading, data, error, retry };
}
