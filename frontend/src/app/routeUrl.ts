/**
 * URL <-> Route codec for the formal frontend App shell.
 *
 * Truth source: docs/spec/frontend-shell/spec.md §4 / C-11..C-13,
 * docs/spec/frontend-shell/plans/004-url-addressable-routing/plan.md §4 + §5,
 * docs/ui-design/auth-and-entry.md, ui-design/src/app.jsx ROUTE_ALIASES.
 *
 * Browser History canonical URL contract — turns a `LooseRoute` into a
 * canonical path + sorted query string, and a canonical URL back into a
 * normalized `Route`. The codec owns the cross-owner safe-param allowlist
 * and never lets raw payload / AI prompt / auth secret keys reach the URL,
 * pendingAction, history.state or storage.
 *
 * `#route=...` static-preview parsing remains in `bootstrapRoute.ts` so
 * pixel parity harnesses keep working through the migration.
 */

import { normalizeRouteName, type LooseRoute } from "./normalizeRoute";
import {
  DEFAULT_ROUTE,
  isKnownRouteName,
  type Route,
  type RouteName,
} from "./routes";

/** Canonical URL pathname for each retained route. */
export const ROUTE_TO_PATH: Readonly<Record<RouteName, string>> = {
  home: "/",
  workspace: "/workspace",
  resume_versions: "/resume-versions",
  parse: "/parse",
  practice: "/practice",
  generating: "/generating",
  report: "/report",
  settings: "/settings",
  auth_login: "/auth/login",
  auth_verify: "/auth/verify",
  auth_profile_setup: "/auth/profile",
  auth_logout: "/auth/logout",
};

/**
 * Explicit non-current paths that still have a current retained destination per
 * product-scope D-16 / D-17 / D-22 and frontend-shell spec §2.2.
 */
export const NON_CURRENT_PATH_TO_ROUTE: ReadonlyMap<string, RouteName> = new Map([
  ["/auth/reset", "auth_login"],
  // product-scope D-17: jd_match is outside the current route catalog; saved
  // deep links land on home where JD intake lives.
  ["/jd-match", "home"],
  // product-scope D-22: non-current product entries are outside the current route catalog.
  ["/debrief", "home"],
  ["/profile", "home"],
]);

const PATH_TO_ROUTE: ReadonlyMap<string, RouteName> = (() => {
  const map = new Map<string, RouteName>();
  for (const name of Object.keys(ROUTE_TO_PATH) as RouteName[]) {
    map.set(ROUTE_TO_PATH[name], name);
  }
  return map;
})();

const WORKSPACE_SAFE = new Set([
  "targetJobId",
  "jobId",
  "resumeId",
  "sourceReportId",
  "planId",
  "roundId",
  "roundName",
  "jdId",
  "sessionId",
  "sourceSessionId",
  "replayItems",
  "evidenceGaps",
  "nextRoundId",
  "mode",
  "modality",
  "practiceMode",
  "practiceGoal",
  "hintUsed",
  "hintCount",
  "autoStartPractice",
  "language",
]);

const PRACTICE_SAFE = new Set([
  "sessionId",
  "planId",
  "targetJobId",
  "jobId",
  "jdId",
  "resumeId",
  "sourceReportId",
  "roundId",
  "roundName",
  "mode",
  "modality",
  "practiceMode",
  "practiceGoal",
  "language",
]);

const GENERATING_SAFE = new Set([
  "sessionId",
  "reportId",
  "planId",
  "targetJobId",
  "jobId",
  "jdId",
  "resumeId",
  "roundId",
  "roundName",
  "mode",
  "modality",
  "practiceMode",
  "practiceGoal",
  "hintUsed",
  "hintCount",
]);

const REPORT_SAFE = new Set([
  "sessionId",
  "reportId",
  "targetJobId",
  "jobId",
  "jdId",
  "resumeId",
  "roundId",
  "roundName",
  "mode",
  "modality",
  "practiceMode",
  "practiceGoal",
  "hintUsed",
  "hintCount",
  "reportStatus",
  "errorCode",
]);

const RESUME_VERSIONS_SAFE = new Set([
  "resumeId",
  "flow",
  "createMode",
  "targetJobId",
]);

const PARSE_SAFE = new Set([
  "jdId",
  "targetJobId",
  "resumeId",
  "importId",
  "source",
  // product-scope D-17 keeps the jd_match -> parse reverse handoff outside
  // the current parse allowlist.
]);

const HOME_SAFE = new Set(["pendingImportId", "source", "resumeId"]);

const SETTINGS_SAFE = new Set(["tab"]);

/**
 * pendingAction reserved keys carried on auth routes. Mirrors the constant
 * in `auth/pendingAction.ts`; keeping a parallel set here avoids a circular
 * import between routeUrl <-> pendingAction.
 */
const PENDING_ACTION_RESERVED = new Set([
  "pendingRoute",
  "pendingType",
  "pendingLabel",
]);

const AUTH_LOGIN_BASE = new Set([
  "next",
  "email",
  ...PENDING_ACTION_RESERVED,
]);
const AUTH_VERIFY_BASE = new Set(["email", ...PENDING_ACTION_RESERVED]);
const AUTH_PROFILE_SETUP_BASE = new Set(["email", ...PENDING_ACTION_RESERVED]);
const AUTH_LOGOUT_BASE = new Set(["next"]);

const ROUTE_SAFE_PARAMS: Readonly<Record<RouteName, ReadonlySet<string>>> = {
  home: HOME_SAFE,
  workspace: WORKSPACE_SAFE,
  resume_versions: RESUME_VERSIONS_SAFE,
  parse: PARSE_SAFE,
  practice: PRACTICE_SAFE,
  generating: GENERATING_SAFE,
  report: REPORT_SAFE,
  settings: SETTINGS_SAFE,
  auth_login: AUTH_LOGIN_BASE,
  auth_verify: AUTH_VERIFY_BASE,
  auth_profile_setup: AUTH_PROFILE_SETUP_BASE,
  auth_logout: AUTH_LOGOUT_BASE,
};

const AUTH_ROUTES_WITH_PENDING_ACTION = new Set<RouteName>([
  "auth_login",
  "auth_verify",
  "auth_profile_setup",
]);

function resolveAllowedParamKeys(
  routeName: RouteName,
  params: Record<string, string>,
): ReadonlySet<string> {
  const base = ROUTE_SAFE_PARAMS[routeName];
  if (!AUTH_ROUTES_WITH_PENDING_ACTION.has(routeName)) return base;
  const pendingRoute = params.pendingRoute;
  if (!pendingRoute || !isKnownRouteName(pendingRoute)) return base;
  const union = new Set(base);
  for (const k of ROUTE_SAFE_PARAMS[pendingRoute]) union.add(k);
  return union;
}

/** Returns true when `key` is permitted on the canonical URL for `routeName`. */
export function isSafeRouteParam(
  routeName: RouteName,
  key: string,
  contextParams: Record<string, string>,
): boolean {
  return resolveAllowedParamKeys(routeName, contextParams).has(key);
}

export interface RoutePathParts {
  path: string;
  search: string;
}

/**
 * Serializes a loose route to canonical `{path, search}` parts. Unknown
 * route names normalize through `normalizeRouteName`; unknown / empty /
 * unsafe params are dropped; remaining params are sorted alphabetically so
 * canonical URLs are stable across renderers.
 */
export function serializeRouteToUrl(input: LooseRoute): RoutePathParts {
  const name = normalizeRouteName(input.name);
  const path = ROUTE_TO_PATH[name];
  const params = input.params ?? {};
  const allowed = resolveAllowedParamKeys(name, params);
  const usp = new URLSearchParams();
  const keys = Object.keys(params)
    .filter((key) => {
      if (!allowed.has(key)) return false;
      const value = params[key];
      return value !== undefined && value !== null && value !== "";
    })
    .sort();
  for (const key of keys) {
    const value = params[key];
    if (value !== undefined) usp.set(key, value);
  }
  const search = usp.toString();
  return { path, search: search ? `?${search}` : "" };
}

/** Convenience: returns `path + search` as a single string. */
export function formatRouteUrl(input: LooseRoute): string {
  const parts = serializeRouteToUrl(input);
  return parts.path + parts.search;
}

/**
 * Parses a canonical URL (path + optional query, with or without leading
 * slash) into a normalized `Route`. Unknown paths fall back to `home`.
 * Unsafe params are silently dropped — they never enter App state. Fragment
 * (`#...`) is ignored here so the canonical parser stays orthogonal to the
 * non-current `#route=...` hash adapter.
 */
export function parseUrlToRoute(rawUrl: string): Route {
  const trimmed = (rawUrl ?? "").trim();
  if (!trimmed) return DEFAULT_ROUTE;
  let url: URL;
  try {
    const withScheme = /^[a-z][a-z0-9+\-.]*:\/\//i.test(trimmed);
    const normalizedInput = withScheme
      ? trimmed
      : trimmed.startsWith("/") || trimmed.startsWith("?")
        ? trimmed
        : `/${trimmed}`;
    url = new URL(normalizedInput, "http://easyinterview.local");
  } catch {
    return DEFAULT_ROUTE;
  }
  const pathname = url.pathname || "/";
  const name =
    PATH_TO_ROUTE.get(pathname) ??
    NON_CURRENT_PATH_TO_ROUTE.get(pathname) ??
    DEFAULT_ROUTE.name;
  const rawParams: Record<string, string> = {};
  for (const [key, value] of url.searchParams.entries()) rawParams[key] = value;
  const allowed = resolveAllowedParamKeys(name, rawParams);
  const params: Record<string, string> = {};
  for (const key of Object.keys(rawParams)) {
    const value = rawParams[key];
    if (allowed.has(key) && value !== undefined && value !== "")
      params[key] = value;
  }
  return { name, params };
}

/**
 * Returns true when two loose routes serialize to the same canonical URL.
 * Used by the route store to suppress redundant `pushState` calls when
 * navigation lands on the same canonical address.
 */
export function routeUrlsEqual(a: LooseRoute, b: LooseRoute): boolean {
  return formatRouteUrl(a) === formatRouteUrl(b);
}
