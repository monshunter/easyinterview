import { useEffect, useRef, useState } from "react";

import type {
  MarketSignal,
  WatchlistItem,
} from "../../../api/generated/types";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";

export interface UseWatchlistResult {
  loading: boolean;
  items: WatchlistItem[];
  error: Error | null;
}

/**
 * Phase 5.2 hook: load `listWatchlist` once when `active` becomes true.
 * The Watchlist tab provides `active` so the call is deferred until the tab
 * is actually opened ("切 Watchlist tab 调一次").
 */
export function useWatchlist(active: boolean): UseWatchlistResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const [loading, setLoading] = useState(false);
  const [items, setItems] = useState<WatchlistItem[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const calledRef = useRef(false);
  const seqRef = useRef(0);

  useEffect(() => {
    if (!client || !active) return;
    if (calledRef.current) return;
    calledRef.current = true;
    let cancelled = false;
    const seq = seqRef.current + 1;
    seqRef.current = seq;
    setLoading(true);
    setError(null);
    client
      .listWatchlist()
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
  }, [client, active]);

  return { loading, items, error };
}

export interface UseMarketSignalsResult {
  loading: boolean;
  signals: MarketSignal[];
  error: Error | null;
}

/**
 * Phase 5.2 hook: load `getMarketSignals` once when `active` becomes true.
 */
export function useMarketSignals(active: boolean): UseMarketSignalsResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const [loading, setLoading] = useState(false);
  const [signals, setSignals] = useState<MarketSignal[]>([]);
  const [error, setError] = useState<Error | null>(null);
  const calledRef = useRef(false);
  const seqRef = useRef(0);

  useEffect(() => {
    if (!client || !active) return;
    if (calledRef.current) return;
    calledRef.current = true;
    let cancelled = false;
    const seq = seqRef.current + 1;
    seqRef.current = seq;
    setLoading(true);
    setError(null);
    client
      .getMarketSignals()
      .then((response) => {
        if (cancelled || seqRef.current !== seq) return;
        setSignals(response.signals);
        setError(null);
      })
      .catch((err: unknown) => {
        if (cancelled || seqRef.current !== seq) return;
        setSignals([]);
        setError(err instanceof Error ? err : new Error(String(err)));
      })
      .finally(() => {
        if (!cancelled && seqRef.current === seq) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [client, active]);

  return { loading, signals, error };
}
