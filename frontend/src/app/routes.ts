/**
 * Route catalog and chrome behavior for the formal frontend App shell.
 *
 * Truth source: docs/spec/frontend-shell/spec.md §2.1, docs/ui-design/auth-and-entry.md,
 * docs/ui-design/ui-architecture.md, and this formal route catalog.
 *
 * Three primary nav entries: home / workspace / resume_versions
 * (out-of-scope product entries are outside the current route catalog).
 * Context routes: parse / practice / reports / generating / report / report_conversation.
 * User-menu routes: settings / auth_logout.
 * Auth pages: auth_login / auth_verify / auth_profile_setup / auth_logout.
 * auth_reset is outside the current route catalog per product-scope D-16; it normalizes back to auth_login.
 *
 * D1 frontend-shell keeps the formal route catalog to current product routes.
 * Out-of-scope aliases are handled outside this catalog and must never appear as
 * materialized app routes.
 */

export const PRIMARY_NAV_ROUTES = [
  "home",
  "workspace",
  "resume_versions",
] as const;

export const CONTEXT_ROUTES = [
  "parse",
  "practice",
  "reports",
  "generating",
  "report",
  "report_conversation",
] as const;

export const USER_MENU_ROUTES = ["settings", "auth_logout"] as const;

export const AUTH_ROUTES = [
  "auth_login",
  "auth_verify",
  "auth_profile_setup",
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

/** Only Practice hydrates mutable interview context; report routes resolve their own locators. */
export const INTERVIEW_CONTEXT_ROUTES: ReadonlySet<string> = new Set([
  "practice",
]);

export function shouldCarryInterviewContext(name: string): boolean {
  return INTERVIEW_CONTEXT_ROUTES.has(name);
}

export const DEFAULT_ROUTE: Route = { name: "home", params: {} };
