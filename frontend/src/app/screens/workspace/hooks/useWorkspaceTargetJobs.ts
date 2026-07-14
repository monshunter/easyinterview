import { useCallback, useEffect, useState } from "react";

import type { TargetJob } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export interface UseWorkspaceTargetJobsResult {
  loading: boolean;
  jobs: TargetJob[];
  error: Error | null;
}

/**
 * Calls listTargetJobs to fetch ready candidate plans for workspace surfaces.
 */
export function useWorkspaceTargetJobs(): UseWorkspaceTargetJobsResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const isAuthenticated = runtime?.auth.status === "authenticated";
  const [loading, setLoading] = useState(true);
  const [jobs, setJobs] = useState<TargetJob[]>([]);
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
    setError(null);

    client
      .listTargetJobs({ query: { analysisStatus: "ready", pageSize: "12" } })
      .then((page) => {
        if (cancelled) return;
        setJobs(page.items.filter(isVisibleWorkspaceTargetJob));
        setError(null);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        setError(err instanceof Error ? err : new Error(String(err)));
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [client, isAuthenticated]);

  useEffect(() => {
    const cancel = fetch();
    return cancel;
  }, [fetch]);

  return { loading, jobs, error };
}

export function isVisibleWorkspaceTargetJob(job: TargetJob): boolean {
  return job.analysisStatus === "ready" && job.title.trim().length > 0;
}
