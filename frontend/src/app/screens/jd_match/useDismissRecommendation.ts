import { useCallback, useState } from "react";

import type {
  JobMatchRecommendation,
  MarkNotRelevantReason,
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
    return `ik-jdmatch-dismiss-${crypto.randomUUID()}`;
  }
  return `ik-jdmatch-dismiss-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

const DEFAULT_DISMISS_REASON: MarkNotRelevantReason = "not_relevant";

export interface UseDismissRecommendationOptions {
  /**
   * Hides the recommendation locally and returns a `revert` callback that
   * restores it to its original position + selected state. The hook calls the
   * revert function when the server returns 4xx / 5xx so the parent's view
   * stays consistent with the persisted truth source.
   */
  applyOptimisticHide: (rec: JobMatchRecommendation) => () => void;
}

export interface UseDismissRecommendationResult {
  dismiss: (rec: JobMatchRecommendation) => Promise<void>;
  pendingIds: ReadonlySet<string>;
}

/**
 * Phase 3.4 hook: Mark not relevant loop with optimistic hide, default reason
 * enum (no UI prompt), Idempotency-Key per call, and window.eiToast feedback.
 */
export function useDismissRecommendation(
  opts: UseDismissRecommendationOptions,
): UseDismissRecommendationResult {
  const runtime = useAppRuntimeOptional();
  const { t } = useI18n();
  const [pendingIds, setPendingIds] = useState<Set<string>>(new Set());

  const setPending = useCallback((id: string, on: boolean) => {
    setPendingIds((prev) => {
      const next = new Set(prev);
      if (on) next.add(id);
      else next.delete(id);
      return next;
    });
  }, []);

  const dismiss = useCallback(
    async (rec: JobMatchRecommendation): Promise<void> => {
      const client = runtime?.client;
      if (!client) return;

      const revert = opts.applyOptimisticHide(rec);
      setPending(rec.id, true);
      const ik = generateIdempotencyKey();
      try {
        await client.markJobNotRelevant(
          rec.id,
          { reason: DEFAULT_DISMISS_REASON },
          { idempotencyKey: ik },
        );
        callToast(t("jdMatch.recommended.toastDismissed"), "neutral");
      } catch {
        revert();
        callToast(t("jdMatch.recommended.toastDismissError"), "danger");
      } finally {
        setPending(rec.id, false);
      }
    },
    [runtime, opts, t, setPending],
  );

  return { dismiss, pendingIds };
}
