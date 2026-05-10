import { useCallback, useEffect, useRef, useState } from "react";

import type {
  JobMatchRecommendation,
  SearchJobsFilters,
} from "../../../api/generated/types";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";

function generateIdempotencyKey(): string {
  if (
    typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
  ) {
    return `ik-jdmatch-search-${crypto.randomUUID()}`;
  }
  return `ik-jdmatch-search-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export interface UseSearchJobsResult {
  searching: boolean;
  results: JobMatchRecommendation[];
  error: Error | null;
  hasRunOnce: boolean;
  run: (query: string, filters?: SearchJobsFilters | null) => Promise<void>;
  abort: () => void;
  reset: () => void;
}

/**
 * Phase 4.2 hook: dispatch jd_match search via the generated client.
 *
 * Contract:
 * - Sends `{ query, filters?, profileSnapshotId? }` body with a unique
 *   Idempotency-Key per call.
 * - Surfaces `searching`, `results`, `error`, `hasRunOnce`.
 * - Aborts in-flight calls when `abort()` is invoked or the consumer
 *   unmounts. The component layer drives the 5-step AGENT panel via
 *   `searching=true` only — the hook does NOT register any setInterval /
 *   setTimeout for step advancement.
 */
export function useSearchJobs(): UseSearchJobsResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const [searching, setSearching] = useState(false);
  const [results, setResults] = useState<JobMatchRecommendation[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const [hasRunOnce, setHasRunOnce] = useState(false);
  const abortRef = useRef<AbortController | null>(null);
  const seqRef = useRef(0);

  useEffect(
    () => () => {
      seqRef.current += 1;
      abortRef.current?.abort();
      abortRef.current = null;
    },
    [],
  );

  const run = useCallback(
    async (
      query: string,
      filters?: SearchJobsFilters | null,
    ): Promise<void> => {
      if (!client) return;
      // Cancel previous in-flight call. The latest call's response wins.
      abortRef.current?.abort();
      const controller =
        typeof AbortController !== "undefined" ? new AbortController() : null;
      abortRef.current = controller;
      const seq = seqRef.current + 1;
      seqRef.current = seq;
      setSearching(true);
      setError(null);
      const ik = generateIdempotencyKey();
      try {
        const response = await client.searchJobs(
          { query, filters: filters ?? undefined },
          {
            idempotencyKey: ik,
            signal: controller?.signal,
          },
        );
        if (seqRef.current !== seq) return;
        setResults(response.items);
        setError(null);
        setHasRunOnce(true);
      } catch (err: unknown) {
        if (seqRef.current !== seq) return;
        if (
          err instanceof DOMException &&
          err.name === "AbortError"
        ) {
          // Aborted: leave existing results untouched.
        } else {
          setResults([]);
          setError(err instanceof Error ? err : new Error(String(err)));
          setHasRunOnce(true);
        }
      } finally {
        if (seqRef.current === seq) {
          setSearching(false);
        }
      }
    },
    [client],
  );

  const abort = useCallback(() => {
    seqRef.current += 1;
    abortRef.current?.abort();
    abortRef.current = null;
    setSearching(false);
  }, []);

  const reset = useCallback(() => {
    seqRef.current += 1;
    abortRef.current?.abort();
    abortRef.current = null;
    setSearching(false);
    setResults([]);
    setError(null);
    setHasRunOnce(false);
  }, []);

  return { searching, results, error, hasRunOnce, run, abort, reset };
}
