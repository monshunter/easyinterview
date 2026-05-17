/**
 * pendingAction model — see docs/ui-design/auth-and-entry.md §6 / §8 and
 * docs/spec/frontend-shell/spec.md §4.
 *
 * Pending actions encode "the user wanted to do X but had to log in first"
 * so that after `verifyAuthEmailChallenge` succeeds the App can restore the
 * original route + business params (planId / targetJobId / jdId /
 * resumeVersionId / roundId etc).
 *
 * The pending action is carried as opaque route params on the auth_* routes
 * so we can survive a hash-based redirect through the email verify link
 * without storing anything in localStorage.
 */

import { normalizeRoute, type LooseRoute } from "../normalizeRoute";
import type { RouteName } from "../routes";

/** Interview-context keys that pending actions MUST round-trip intact. */
export const PENDING_ACTION_INTERVIEW_KEYS = [
  "planId",
  "targetJobId",
  "jobId",
  "jdId",
  "resumeVersionId",
  "roundId",
  "roundName",
  "mode",
  "modality",
  // frontend-debrief plan 001 Phase 5.4: the debrief workflow needs to
  // restore its own context after sign-in (e.g. `start_debrief_interview`).
  "practiceGoal",
  "debriefId",
  "debriefJobId",
  "sessionId",
] as const;

const RESERVED_KEYS = [
  "pendingRoute",
  "pendingType",
  "pendingLabel",
  "returnTo",
  "email",
  "displayName",
] as const;

const RESERVED_KEY_SET = new Set<string>(RESERVED_KEYS);

export interface PendingAction {
  /** Programmatic action key (e.g., `start_practice`, `start_drill`). */
  type: string;
  /** User-facing label, surfaced on the auth screen as "登录后继续 X". */
  label: string;
  /** Target route to restore after successful verify. */
  route: RouteName;
  /** Route params to restore after successful verify. */
  params: Record<string, string>;
}

/**
 * Flattens a {@link PendingAction} into a route params object, ready to be
 * carried on the auth_* routes.
 */
export function encodePendingAction(
  action: PendingAction,
): Record<string, string> {
  return {
    pendingRoute: action.route,
    pendingType: action.type,
    pendingLabel: action.label,
    ...action.params,
  };
}

/**
 * Inverse of {@link encodePendingAction}. Returns `null` when no pending
 * action is encoded in the params; otherwise rebuilds the loose route.
 */
export function decodePendingActionRoute(
  params: Record<string, string>,
): LooseRoute | null {
  const route = params.pendingRoute;
  if (!route) return null;
  const restored: Record<string, string> = {};
  for (const [key, value] of Object.entries(params)) {
    if (RESERVED_KEY_SET.has(key)) continue;
    restored[key] = value;
  }
  return normalizeRoute({ name: route, params: restored });
}
