import { useCallback, useEffect, useRef, useState } from "react";

import type { TargetJob } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export interface UseWorkspaceTargetJobsResult {
  loading: boolean;
  jobs: TargetJob[];
  error: Error | null;
}

/**
 * Calls listTargetJobs to fetch all candidate plans for Plan Switcher Modal.
 */
export function useWorkspaceTargetJobs(): UseWorkspaceTargetJobsResult {
  const runtime = useAppRuntimeOptional();
  const [loading, setLoading] = useState(true);
  const [jobs, setJobs] = useState<TargetJob[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const inFlightRef = useRef(false);

  const fetch = useCallback(() => {
    if (!runtime) {
      setLoading(false);
      return;
    }
    if (inFlightRef.current) return;

    let cancelled = false;
    inFlightRef.current = true;
    setLoading(true);
    setError(null);

    runtime.client
      .listTargetJobs({ query: { pageSize: "12" } })
      .then((page) => {
        if (cancelled) return;
        setJobs(page.items);
        setError(null);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        setError(err instanceof Error ? err : new Error(String(err)));
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
          inFlightRef.current = false;
        }
      });
  }, [runtime]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { loading, jobs, error };
}
