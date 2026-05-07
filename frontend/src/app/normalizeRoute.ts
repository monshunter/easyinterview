/**
 * Route normalization for the formal frontend App shell.
 *
 * Truth source: ui-design/src/app.jsx ROUTE_ALIASES, docs/ui-design/auth-and-entry.md §9
 * (welcome removed), docs/ui-design/removed-modules-and-scope.md (growth / mistakes /
 * drill / followup / experiences / star removed; standalone voice alias deleted),
 * docs/spec/frontend-shell/spec.md §2.2 / §4.
 *
 * Design constraint (frontend-shell §4): old route names must NEVER materialize
 * standalone screens. They normalize to the current retained route, or to `home`
 * when no obvious target exists.
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
