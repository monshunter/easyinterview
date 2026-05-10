import type { JobMatchRecommendation } from "../../../api/generated/types";

import type { SearchResultFilter } from "./SearchTab";

/**
 * Phase 4.4 client-side filter predicate.
 *
 * - all     → every recommendation passes
 * - strong  → score >= 85
 * - remote  → location contains "remote" / "远程" (case-insensitive)
 * - unseen  → !seen
 */
export function recommendationMatchesFilter(
  rec: JobMatchRecommendation,
  filter: SearchResultFilter,
): boolean {
  switch (filter) {
    case "all":
      return true;
    case "strong":
      return rec.score >= 85;
    case "remote":
      return /remote|远程/i.test(rec.location);
    case "unseen":
      return !rec.seen;
    default:
      return true;
  }
}

export function applySearchFilter(
  recs: JobMatchRecommendation[],
  filter: SearchResultFilter,
): JobMatchRecommendation[] {
  if (filter === "all") return recs;
  return recs.filter((r) => recommendationMatchesFilter(r, filter));
}
