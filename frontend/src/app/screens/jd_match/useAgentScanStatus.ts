import { useEffect, useRef, useState } from "react";

import type { AgentScanStatus } from "../../../api/generated/types";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";

export interface UseAgentScanStatusResult {
  loading: boolean;
  data: AgentScanStatus | null;
  error: Error | null;
}

const RECOMMENDED_TAB = "recommended";

/**
 * Phase 2.2 hook: loads AgentScanStatus via generated client whenever
 * `activeTab` is "recommended" — once on mount and once each time the tab
 * transitions back to "recommended" from another tab. No setInterval / SSE /
 * WebSocket is registered (D-10 forbids polling/streaming for this signal).
 */
export function useAgentScanStatus(activeTab: string): UseAgentScanStatusResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const [loading, setLoading] = useState<boolean>(false);
  const [data, setData] = useState<AgentScanStatus | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const requestSeqRef = useRef(0);

  useEffect(() => {
    if (!client || activeTab !== RECOMMENDED_TAB) {
      return;
    }

    let active = true;
    const requestSeq = requestSeqRef.current + 1;
    requestSeqRef.current = requestSeq;
    setLoading(true);
    setError(null);

    client
      .getAgentScanStatus()
      .then((status) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setData(status);
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
  }, [client, activeTab]);

  return { loading, data, error };
}
