import {
  useTailorSuggestionDecision,
  type UseTailorSuggestionDecisionResult,
} from "./useTailorSuggestionDecision";

/**
 * Accept a tailor suggestion. The generated client posts a bodyless request
 * (D-12 contract) and the hook supplies the Idempotency-Key so retries are
 * deduplicated server-side. The accept call never patches structured_profile;
 * suggestion-status terminal updates are the only side effect.
 */
export function useAcceptResumeTailorSuggestion(): UseTailorSuggestionDecisionResult {
  return useTailorSuggestionDecision((client) =>
    (versionId, suggestionId, opts) =>
      client.acceptResumeTailorSuggestion(versionId, suggestionId, opts),
  );
}
