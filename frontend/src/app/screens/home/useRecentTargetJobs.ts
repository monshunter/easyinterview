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
  const client = runtime?.client;
  const isAuthenticated = runtime?.auth.status === "authenticated";
  const [jobs, setJobs] = useState<TargetJob[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetch = useCallback(() => {
    if (!client || !isAuthenticated) {
      setJobs([]);
      setLoading(false);
      setError(null);
      return;
    }

    let cancelled = false;
    setLoading(true);

    client
      .listTargetJobs({ query: { analysisStatus: "ready", pageSize: "12" } })
      .then((page) => {
        if (!cancelled) {
          setJobs(
            Array.isArray(page.items)
              ? page.items.filter(isVisibleRecentTargetJob)
              : [],
          );
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
  }, [client, isAuthenticated]);

  useEffect(() => {
    const cancel = fetch();
    return cancel;
  }, [fetch]);

  return { jobs, loading, error, refetch: fetch };
}

function isVisibleRecentTargetJob(job: TargetJob): boolean {
  return job.analysisStatus === "ready" && job.title.trim().length > 0;
}
