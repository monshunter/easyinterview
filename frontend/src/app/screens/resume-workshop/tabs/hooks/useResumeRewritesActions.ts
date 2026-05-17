import { useCallback, useState } from "react";

import type { ResumeVersion } from "../../../../../api/generated/types";
import { useAcceptResumeTailorSuggestion } from "./useAcceptResumeTailorSuggestion";
import { useRejectResumeTailorSuggestion } from "./useRejectResumeTailorSuggestion";
import {
  SuggestionDecisionError,
} from "./useTailorSuggestionDecision";
import { useUpdateResumeVersion } from "./useUpdateResumeVersion";

export interface ManualEditEntry {
  suggestionId: string;
  text: string;
  savedAt: string;
}

export interface ResumeRewritesActions {
  onAccept: (suggestionId: string) => Promise<void>;
  onReject: (suggestionId: string) => Promise<void>;
  onSaveManualEdit: (
    suggestionId: string,
    text: string,
  ) => Promise<void>;
  /** Suggestion id of the bullet whose manual edit was saved but accept failed. */
  manualPendingFor: string | null;
  /** Optional consumer-side hook to surface toasts / errors. */
  lastDecisionError: SuggestionDecisionError | null;
}

export interface UseResumeRewritesActionsOptions {
  version: ResumeVersion;
  onVersionRefreshed?: () => void;
  /**
   * Test seam: lets the host inject a deterministic clock so manual-edit
   * payloads are reproducible.
   */
  nowProvider?: () => Date;
}

const mergeManualEdits = (
  existing: unknown,
  entry: ManualEditEntry,
): ManualEditEntry[] => {
  if (!Array.isArray(existing)) return [entry];
  const cleaned = existing.filter((row): row is ManualEditEntry => {
    if (!row || typeof row !== "object") return false;
    const obj = row as Record<string, unknown>;
    return (
      typeof obj.suggestionId === "string" &&
      typeof obj.text === "string" &&
      typeof obj.savedAt === "string"
    );
  });
  const next = cleaned.filter((row) => row.suggestionId !== entry.suggestionId);
  next.push(entry);
  return next;
};

export function useResumeRewritesActions(
  options: UseResumeRewritesActionsOptions,
): ResumeRewritesActions {
  const accept = useAcceptResumeTailorSuggestion();
  const reject = useRejectResumeTailorSuggestion();
  const update = useUpdateResumeVersion();
  const [manualPendingFor, setManualPendingFor] = useState<string | null>(null);

  const onAccept = useCallback(
    async (suggestionId: string) => {
      await accept.decide(options.version.id, suggestionId);
      options.onVersionRefreshed?.();
    },
    [accept, options],
  );

  const onReject = useCallback(
    async (suggestionId: string) => {
      await reject.decide(options.version.id, suggestionId);
      options.onVersionRefreshed?.();
    },
    [reject, options],
  );

  const onSaveManualEdit = useCallback(
    async (suggestionId: string, text: string) => {
      const now = (options.nowProvider ?? (() => new Date()))();
      const entry: ManualEditEntry = {
        suggestionId,
        text,
        savedAt: now.toISOString(),
      };
      const existing = (
        options.version.structuredProfile as Record<string, unknown>
      ).manualEdits;
      const nextProfile = {
        ...(options.version.structuredProfile as Record<string, unknown>),
        manualEdits: mergeManualEdits(existing, entry),
      };
      // Step 1: persist the manual edit on the version. If this fails, do not
      // continue to accept (otherwise the bullet would flip terminal without
      // its edit text being saved).
      await update.update({
        versionId: options.version.id,
        payload: { structuredProfile: nextProfile },
      });
      // Step 2: mark the suggestion terminal via the bodyless accept call. If
      // this step fails, surface the "saved manual edit pending retry" state
      // so the UI can re-attempt just the accept without rewriting the edit.
      try {
        await accept.decide(options.version.id, suggestionId);
        setManualPendingFor((prev) =>
          prev === suggestionId ? null : prev,
        );
        options.onVersionRefreshed?.();
      } catch (acceptErr) {
        setManualPendingFor(suggestionId);
        throw acceptErr;
      }
    },
    [accept, options, update],
  );

  return {
    onAccept,
    onReject,
    onSaveManualEdit,
    manualPendingFor,
    lastDecisionError:
      accept.lastError ?? reject.lastError ?? null,
  };
}
