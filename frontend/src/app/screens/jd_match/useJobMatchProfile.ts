import { useEffect, useRef, useState } from "react";

import type { JobMatchProfile } from "../../../api/generated/types";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";

export interface UseJobMatchProfileResult {
  loading: boolean;
  data: JobMatchProfile | null;
  error: Error | null;
}

/**
 * Phase 2.2 hook: loads JobMatchProfile via generated client once on mount.
 * Returns inert state when no AppRuntimeProvider is mounted.
 */
export function useJobMatchProfile(): UseJobMatchProfileResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const [loading, setLoading] = useState<boolean>(!!client);
  const [data, setData] = useState<JobMatchProfile | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const requestSeqRef = useRef(0);

  useEffect(() => {
    if (!client) {
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
      .getJobMatchProfile()
      .then((profile) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setData(profile);
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
  }, [client]);

  return { loading, data, error };
}
