import { useCallback, useEffect, useRef, useState } from "react";

import type {
  CreateSavedSearchRequest,
  SavedSearch,
} from "../../../api/generated/types";
import { useI18n } from "../../i18n/messages";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";

type ToastTone = "ok" | "warn" | "danger" | "neutral";
type EiToast = (
  message: string,
  opts?: { tone?: ToastTone; duration?: number },
) => void;

function callToast(message: string, tone: ToastTone): void {
  if (typeof window === "undefined") return;
  const fn = (window as unknown as { eiToast?: EiToast }).eiToast;
  if (typeof fn === "function") fn(message, { tone });
}

function generateIdempotencyKey(): string {
  if (
    typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
  ) {
    return `ik-jdmatch-saved-search-${crypto.randomUUID()}`;
  }
  return `ik-jdmatch-saved-search-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export interface UseSavedSearchesResult {
  loading: boolean;
  items: SavedSearch[];
  error: Error | null;
  retry: () => void;
}

/**
 * Phase 4.3 hook: load `listSavedSearches` once when `active` becomes true.
 * The Search tab provides `active` so the call is deferred until the tab is
 * actually opened (matches plan 4.3 contract: "list 调一次").
 */
export function useSavedSearches(active: boolean): UseSavedSearchesResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const [loading, setLoading] = useState(false);
  const [items, setItems] = useState<SavedSearch[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const [retryNonce, setRetryNonce] = useState(0);
  const seqRef = useRef(0);
  const calledRef = useRef(false);

  useEffect(() => {
    if (!client || !active) return;
    if (calledRef.current && retryNonce === 0) return;
    calledRef.current = true;
    let cancelled = false;
    const seq = seqRef.current + 1;
    seqRef.current = seq;
    setLoading(true);
    setError(null);
    client
      .listSavedSearches()
      .then((response) => {
        if (cancelled || seqRef.current !== seq) return;
        setItems(response.items);
        setError(null);
      })
      .catch((err: unknown) => {
        if (cancelled || seqRef.current !== seq) return;
        setItems([]);
        setError(err instanceof Error ? err : new Error(String(err)));
      })
      .finally(() => {
        if (!cancelled && seqRef.current === seq) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [client, active, retryNonce]);

  const retry = useCallback(() => setRetryNonce((n) => n + 1), []);

  return { loading, items, error, retry };
}

export interface UseCreateSavedSearchOptions {
  onCreated?: (saved: SavedSearch) => void;
}

export interface UseCreateSavedSearchResult {
  creating: boolean;
  error: Error | null;
  create: (req: CreateSavedSearchRequest) => Promise<void>;
}

/**
 * Phase 4.3 hook: create a saved search. Calls the generated client with a
 * unique Idempotency-Key, dispatches an eiToast on success/failure, and
 * surfaces the created entry through the `onCreated` callback so the
 * consumer can prepend it to the applied list without a re-fetch round-trip.
 */
export function useCreateSavedSearch(
  opts: UseCreateSavedSearchOptions = {},
): UseCreateSavedSearchResult {
  const runtime = useAppRuntimeOptional();
  const { t } = useI18n();
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const create = useCallback(
    async (req: CreateSavedSearchRequest): Promise<void> => {
      const client = runtime?.client;
      if (!client) return;
      setCreating(true);
      setError(null);
      const ik = generateIdempotencyKey();
      try {
        const saved = await client.createSavedSearch(req, {
          idempotencyKey: ik,
        });
        opts.onCreated?.(saved);
        callToast(t("jdMatch.search.savedSearchCreated"), "ok");
      } catch (err: unknown) {
        const wrapped = err instanceof Error ? err : new Error(String(err));
        setError(wrapped);
        callToast(t("jdMatch.search.savedSearchCreateError"), "danger");
      } finally {
        setCreating(false);
      }
    },
    [runtime, opts, t],
  );

  return { creating, error, create };
}
