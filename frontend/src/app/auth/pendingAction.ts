/**
 * pendingAction model — see docs/ui-design/auth-and-entry.md §6 / §8 and
 * docs/spec/frontend-shell/spec.md §4.
 *
 * Pending actions encode "the user wanted to do X but had to log in first"
 * so that after `verifyAuthEmailChallenge` succeeds the App can restore the
 * original route + business params (planId / targetJobId / jdId /
 * resumeId / roundId etc).
 *
 * The pending action is carried as opaque route params on the auth_* routes
 * so we can survive a hash-based redirect through the email verify link
 * without storing anything in localStorage.
 */

import { normalizeRoute, normalizeRouteName, type LooseRoute } from "../normalizeRoute";
import type { RouteName } from "../routes";
import { isSafeRouteParam } from "../routeUrl";

/** Interview-context keys that pending actions MUST round-trip intact. */
export const PENDING_ACTION_INTERVIEW_KEYS = [
  "planId",
  "targetJobId",
  "jobId",
  "jdId",
  "resumeId",
  "roundId",
  "roundName",
  "practiceGoal",
  "sessionId",
] as const;

const RESERVED_KEYS = [
  "pendingRoute",
  "pendingType",
  "pendingLabel",
  "returnTo",
  "email",
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

function filterSafePendingParams(
  target: RouteName,
  params: Record<string, string>,
): Record<string, string> {
  // Plan 004 §3.1: pendingAction must round-trip canonical route identity +
  // safe params only. Raw payload / AI prompt / auth secret keys never enter
  // the encoded action even if a screen accidentally passes them.
  const safe: Record<string, string> = {};
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === "") continue;
    if (!isSafeRouteParam(target, key, params)) continue;
    safe[key] = value;
  }
  return safe;
}

/**
 * Flattens a {@link PendingAction} into a route params object, ready to be
 * carried on the auth_* routes. Filters action params through the canonical
 * URL safe-param allowlist so raw payloads never leak into pendingAction
 * even when a caller passes them by mistake.
 */
export function encodePendingAction(
  action: PendingAction,
): Record<string, string> {
  const safeParams = filterSafePendingParams(action.route, action.params);
  return {
    pendingRoute: action.route,
    pendingType: action.type,
    pendingLabel: action.label,
    ...safeParams,
  };
}

/**
 * Inverse of {@link encodePendingAction}. Returns `null` when no pending
 * action is encoded in the params; otherwise rebuilds the loose route.
 * Applies the same canonical safe-param allowlist on the restore path so a
 * forged auth URL cannot bypass the encode-side filter.
 */
export function decodePendingActionRoute(
  params: Record<string, string>,
): LooseRoute | null {
  const route = params.pendingRoute;
  if (!route) return null;
  const target = normalizeRouteName(route);
  const restored: Record<string, string> = {};
  for (const [key, value] of Object.entries(params)) {
    if (RESERVED_KEY_SET.has(key)) continue;
    if (!isSafeRouteParam(target, key, params)) continue;
    restored[key] = value;
  }
  return normalizeRoute({ name: target, params: restored });
}
