import { useCallback, useEffect, useRef, useState } from "react";

import type { Resume } from "../../../../api/generated/types";
import { useDisplayPreferencesOptional } from "../../../display/DisplayPreferencesProvider";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export interface UseResumeAssetResult {
  loading: boolean;
  data: Resume | null;
  error: Error | null;
  notFound: boolean;
  retry: () => void;
}

/**
 * Loads a single flat resume via `getResume(resumeId)`.
 * Used as the detail-view primary loader.
 */
export function useResumeAsset(resumeId: string | null): UseResumeAssetResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const isAuthenticated = runtime?.auth.status === "authenticated";
  const lang = useDisplayPreferencesOptional()?.lang ?? "zh";

  const [loading, setLoading] = useState<boolean>(
    !!client && isAuthenticated && !!resumeId,
  );
  const [data, setData] = useState<Resume | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [reloadSeq, setReloadSeq] = useState(0);
  const requestSeqRef = useRef(0);

  const retry = useCallback(() => {
    setReloadSeq((value) => value + 1);
  }, []);

  useEffect(() => {
    if (!client || !isAuthenticated || !resumeId) {
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
      .getResume(resumeId, { headers: { "Accept-Language": lang } })
      .then((resume) => {
        if (!active || requestSeqRef.current !== requestSeq) return;
        setData(resume);
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
  }, [client, isAuthenticated, resumeId, reloadSeq, lang]);

  const notFound = error
    ? error.message.startsWith("HTTP 404 ") || error.message.includes("404")
    : false;

  return { loading, data, error, notFound, retry };
}
