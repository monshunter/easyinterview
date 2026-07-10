/**
 * Route normalization for the formal frontend App shell.
 *
 * Truth source: ui-design/src/app.jsx ROUTE_ALIASES, docs/ui-design/auth-and-entry.md §9,
 * docs/ui-design/module-map.md §5.2,
 * docs/spec/frontend-shell/spec.md §2.2 / §4.
 *
 * Design constraint (frontend-shell §4): out-of-scope route names must never
 * materialize standalone screens. Only explicitly mapped aliases normalize to a
 * current route; all other inputs fall back to `home`.
 */

import { isKnownRouteName, type Route, type RouteName } from "./routes";

const ROUTE_ALIASES: Record<string, RouteName> = {
  welcome: "home",
  growth: "home",
  plan: "workspace",
  mistakes: "report",
  drill: "practice",
  followup: "practice",
  experiences: "resume_versions",
  star: "resume_versions",
  resume: "resume_versions",
  onboarding: "resume_versions",
  auth_register: "auth_login",
  // product-scope D-17: job recommendations are outside current scope; the
  // out-of-scope jd_match entry folds back into home (JD intake lives there).
  jd_match: "home",
  // product-scope D-16: email-code is the only sign-in flow.
  auth_reset: "auth_login",
  // product-scope D-22: debrief and user profile are outside current scope.
  // Saved state and deep links fold back to home instead of materializing screens.
  debrief: "home",
  debrief_full: "home",
  profile: "home",
};

export function normalizeRouteName(raw: string): RouteName {
  const alias = ROUTE_ALIASES[raw];
  if (alias) return alias;
  if (isKnownRouteName(raw)) return raw;
  return "home";
}

export interface LooseRoute {
  name: string;
  params?: Record<string, string>;
}

export function normalizeRoute(input: LooseRoute): Route {
  return {
    name: normalizeRouteName(input.name),
    params: input.params ?? {},
  };
}
