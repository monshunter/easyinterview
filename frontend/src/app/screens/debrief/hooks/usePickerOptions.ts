import { useEffect, useState } from "react";

import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import type { PickerOption } from "../components/DebriefContextPickerModal";

export interface UsePickerOptionsState<T> {
  options: PickerOption<T>[];
  loading: boolean;
  error: string | null;
  /** When true the picker showed a server-filter fallback (client-side filter). */
  fallback?: "client-side-status-filter";
}

interface UsePickerOptionsArgs<T> {
  enabled: boolean;
  /**
   * Async loader that returns the picker options. Implementations call the
   * generated client and map the response items to PickerOption<T>.
   */
  load: () => Promise<{
    options: PickerOption<T>[];
    fallback?: UsePickerOptionsState<T>["fallback"];
  }>;
}

/**
 * Tiny async loader for picker options. Resets when `enabled` flips, so the
 * picker can re-fetch each time it's opened without leaking stale state.
 */
export function usePickerOptions<T>({
  enabled,
  load,
}: UsePickerOptionsArgs<T>): UsePickerOptionsState<T> {
  const runtime = useAppRuntimeOptional();
  const [state, setState] = useState<UsePickerOptionsState<T>>({
    options: [],
    loading: false,
    error: null,
  });

  useEffect(() => {
    if (!enabled) {
      setState({ options: [], loading: false, error: null });
      return;
    }
    let cancelled = false;
    setState({ options: [], loading: true, error: null });
    load()
      .then((result) => {
        if (cancelled) return;
        setState({
          options: result.options,
          loading: false,
          error: null,
          fallback: result.fallback,
        });
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        const message =
          err instanceof Error ? err.message : "failed to load options";
        setState({ options: [], loading: false, error: message });
      });
    return () => {
      cancelled = true;
    };
    // runtime is read inside `load` callbacks; we intentionally do not key on
    // it here so callers that rebuild `load` each render trigger re-fetches.
  }, [enabled, load]);

  // Touch runtime so the hook short-circuits to a no-op when the AppRuntime
  // provider is absent (route-state probes that don't mount auth).
  void runtime;
  return state;
}
