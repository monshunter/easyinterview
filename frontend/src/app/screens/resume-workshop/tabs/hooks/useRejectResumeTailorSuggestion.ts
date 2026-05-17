import {
  useTailorSuggestionDecision,
  type UseTailorSuggestionDecisionResult,
} from "./useTailorSuggestionDecision";

/**
 * Reject a tailor suggestion. Mirrors {@link useAcceptResumeTailorSuggestion}:
 * bodyless POST with an Idempotency-Key, no structured_profile mutation,
 * terminal status update only.
 */
export function useRejectResumeTailorSuggestion(): UseTailorSuggestionDecisionResult {
  return useTailorSuggestionDecision((client) =>
    (versionId, suggestionId, opts) =>
      client.rejectResumeTailorSuggestion(versionId, suggestionId, opts),
  );
}
