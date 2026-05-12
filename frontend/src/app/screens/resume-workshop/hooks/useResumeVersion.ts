import { useEffect, useRef, useState } from "react";

import type { ResumeVersion } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export interface UseResumeVersionResult {
  loading: boolean;
  data: ResumeVersion | null;
  error: Error | null;
  notFound: boolean;
}

export function useResumeVersion(versionId: string | null): UseResumeVersionResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const isAuthenticated = runtime?.auth.status === "authenticated";

  const [loading, setLoading] = useState<boolean>(
    !!client && isAuthenticated && !!versionId,
  );
  const [data, setData] = useState<ResumeVersion | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const requestSeqRef = useRef(0);

  useEffect(() => {
    if (!client || !isAuthenticated || !versionId) {
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
      .getResumeVersion(versionId)
      .then((version) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setData(version);
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
  }, [client, isAuthenticated, versionId]);

  const notFound = error
    ? error.message.startsWith("HTTP 404 ") ||
      error.message.includes("404")
    : false;

  return { loading, data, error, notFound };
}
