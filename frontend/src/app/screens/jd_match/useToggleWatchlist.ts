import { useCallback, useRef, useState } from "react";

import type { JobMatchRecommendation } from "../../../api/generated/types";
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
  if (typeof fn === "function") {
    fn(message, { tone });
  }
}

function generateIdempotencyKey(): string {
  if (
    typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
  ) {
    return `ik-jdmatch-watchlist-${crypto.randomUUID()}`;
  }
  return `ik-jdmatch-watchlist-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export interface UseToggleWatchlistOptions {
  applyOptimistic: (jobMatchId: string, savedNext: boolean) => void;
}

export interface UseToggleWatchlistResult {
  toggleSave: (rec: JobMatchRecommendation) => Promise<void>;
  pendingIds: ReadonlySet<string>;
}

/**
 * Phase 3.3 hook: Save / Unsave loop with optimistic apply, per-id sequence
 * race handling, Idempotency-Key per call, and window.eiToast feedback. Does
 * NOT own the recommendations list — the consumer hands in `applyOptimistic`
 * to flip `saved` on its local state.
 */
export function useToggleWatchlist(
  opts: UseToggleWatchlistOptions,
): UseToggleWatchlistResult {
  const runtime = useAppRuntimeOptional();
  const { t } = useI18n();
  const seqRef = useRef<Map<string, number>>(new Map());
  const [pendingIds, setPendingIds] = useState<Set<string>>(new Set());

  const setPending = useCallback((id: string, on: boolean) => {
    setPendingIds((prev) => {
      const next = new Set(prev);
      if (on) next.add(id);
      else next.delete(id);
      return next;
    });
  }, []);

  const toggleSave = useCallback(
    async (rec: JobMatchRecommendation): Promise<void> => {
      const client = runtime?.client;
      if (!client) return;

      const next = !rec.saved;
      opts.applyOptimistic(rec.id, next);

      const seq = (seqRef.current.get(rec.id) ?? 0) + 1;
      seqRef.current.set(rec.id, seq);
      setPending(rec.id, true);

      const ik = generateIdempotencyKey();
      try {
        if (next) {
          await client.addToWatchlist(
            { jobMatchId: rec.id },
            { idempotencyKey: ik },
          );
        } else {
          await client.removeFromWatchlist(rec.id, { idempotencyKey: ik });
        }
        if (seqRef.current.get(rec.id) !== seq) return;
        callToast(
          next
            ? t("jdMatch.recommended.toastSaved")
            : t("jdMatch.recommended.toastUnsaved"),
          "ok",
        );
      } catch {
        if (seqRef.current.get(rec.id) !== seq) return;
        opts.applyOptimistic(rec.id, rec.saved);
        callToast(t("jdMatch.recommended.toastSaveError"), "danger");
      } finally {
        setPending(rec.id, false);
      }
    },
    [runtime, opts, t, setPending],
  );

  return { toggleSave, pendingIds };
}
