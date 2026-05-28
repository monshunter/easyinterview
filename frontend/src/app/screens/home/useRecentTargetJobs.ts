import { useCallback, useEffect, useState } from "react";

import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { TargetJob } from "../../../api/generated/types";

export interface UseRecentTargetJobsResult {
  jobs: TargetJob[];
  loading: boolean;
  error: Error | null;
  refetch: () => void;
}

export function useRecentTargetJobs(): UseRecentTargetJobsResult {
  const runtime = useAppRuntimeOptional();
  const [jobs, setJobs] = useState<TargetJob[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetch = useCallback(() => {
    if (!runtime || runtime.auth.status !== "authenticated") {
      setJobs([]);
      setLoading(false);
      setError(null);
      return;
    }

    let cancelled = false;
    setLoading(true);

    runtime.client
      .listTargetJobs({ query: { pageSize: "12" } })
      .then((page) => {
        if (!cancelled) {
          setJobs(Array.isArray(page.items) ? page.items : []);
          setError(null);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)));
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [runtime]);

  useEffect(() => {
    const cancel = fetch();
    return cancel;
  }, [fetch]);

  return { jobs, loading, error, refetch: fetch };
}
