/**
 * Route catalog and chrome behavior for the formal frontend App shell.
 *
 * Truth source: docs/spec/frontend-shell/spec.md §2.1, docs/ui-design/auth-and-entry.md,
 * docs/ui-design/ui-architecture.md, ui-design/src/app.jsx.
 *
 * Five primary nav entries: home / jd_match / workspace / resume_versions / debrief.
 * Context routes: parse / practice / generating / report / company_intel.
 * User-menu routes: profile / settings / auth_logout.
 * Auth pages: auth_login / auth_register / auth_verify / auth_reset / auth_logout.
 *
 * D1 frontend-shell keeps the formal route catalog to current product routes.
 * Retired aliases are handled outside this catalog and must never appear as
 * materialized app routes.
 */

export const PRIMARY_NAV_ROUTES = [
  "home",
  "jd_match",
  "workspace",
  "resume_versions",
  "debrief",
] as const;

export const CONTEXT_ROUTES = [
  "parse",
  "practice",
  "generating",
  "report",
  "company_intel",
] as const;

export const USER_MENU_ROUTES = ["profile", "settings", "auth_logout"] as const;

export const AUTH_ROUTES = [
  "auth_login",
  "auth_register",
  "auth_verify",
  "auth_reset",
  "auth_logout",
] as const;

const ALL_ROUTE_NAMES = [
  ...PRIMARY_NAV_ROUTES,
  ...CONTEXT_ROUTES,
  ...USER_MENU_ROUTES,
  ...AUTH_ROUTES,
] as const;

export type RouteName = (typeof ALL_ROUTE_NAMES)[number];

export interface Route {
  name: RouteName;
  params: Record<string, string>;
}

const KNOWN_ROUTE_NAMES = new Set<string>(ALL_ROUTE_NAMES);

export function isKnownRouteName(value: string): value is RouteName {
  return KNOWN_ROUTE_NAMES.has(value);
}

const NO_CHROME_ROUTES = new Set<RouteName>(["practice", "generating"]);

/** Returns true when the route should hide the App chrome (TopBar etc). */
export function isChromeHidden(name: RouteName): boolean {
  return NO_CHROME_ROUTES.has(name);
}

/** Routes that carry InterviewContext across navigation per ui-design/src/app.jsx line 76. */
export const INTERVIEW_CONTEXT_ROUTES: ReadonlySet<string> = new Set([
  "workspace",
  "practice",
  "generating",
  "report",
  "debrief",
  "company_intel",
]);

export function shouldCarryInterviewContext(name: string): boolean {
  return INTERVIEW_CONTEXT_ROUTES.has(name);
}

export const DEFAULT_ROUTE: Route = { name: "home", params: {} };
