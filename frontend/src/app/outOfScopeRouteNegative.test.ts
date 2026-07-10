/**
 * Plan 004 Phase 4.2 — out-of-scope route negative regression.
 *
 * Asserts that out-of-scope routes (`welcome`, `growth`, `plan`, `mistakes`,
 * `drill`, `followup`, `experiences`, `star`, `onboarding`, standalone
 * `voice`, `debrief`, `profile`) never materialize as:
 *   - canonical paths emitted by `formatRouteUrl`
 *   - SPA fallback paths in `spaFallback.FRONTEND_CANONICAL_PATHS`
 *   - TopBar primary nav entries
 *   - normalized standalone route names (must collapse to retained routes)
 */
import { describe, expect, it } from "vitest";

import { normalizeRouteName } from "./normalizeRoute";
import { formatRouteUrl, ROUTE_TO_PATH, serializeRouteToUrl } from "./routeUrl";
import {
  AUTH_ROUTES,
  CONTEXT_ROUTES,
  PRIMARY_NAV_ROUTES,
  USER_MENU_ROUTES,
} from "./routes";

import { FRONTEND_CANONICAL_PATHS } from "../../scripts/spaFallback.mjs";

const OUT_OF_SCOPE_ALIASES = [
  "welcome",
  "growth",
  "plan",
  "mistakes",
  "drill",
  "followup",
  "experiences",
  "star",
  "onboarding",
  "voice",
  "debrief",
  "debrief_full",
  "profile",
] as const;

const OUT_OF_SCOPE_PATHS = [
  "/welcome",
  "/growth",
  "/plan",
  "/mistakes",
  "/drill",
  "/followup",
  "/experiences",
  "/star",
  "/onboarding",
  "/voice",
  "/debrief",
  "/debrief-full",
  "/profile",
] as const;

describe("Plan 004 Phase 4.2 — out-of-scope route negative regression", () => {
  it("ROUTE_TO_PATH does not include out-of-scope paths", () => {
    const allPaths = Object.values(ROUTE_TO_PATH);
    for (const path of OUT_OF_SCOPE_PATHS) {
      expect(
        allPaths.includes(path),
        `routeUrl.ROUTE_TO_PATH must not include out-of-scope path ${path}`,
      ).toBe(false);
    }
  });

  it("SPA fallback FRONTEND_CANONICAL_PATHS does not include out-of-scope paths", () => {
    for (const path of OUT_OF_SCOPE_PATHS) {
      expect(
        FRONTEND_CANONICAL_PATHS.includes(path),
        `spaFallback.FRONTEND_CANONICAL_PATHS must not include out-of-scope path ${path}`,
      ).toBe(false);
    }
  });

  it("formatRouteUrl normalizes out-of-scope aliases to retained canonical paths (no standalone)", () => {
    const expectations: Record<string, string> = {
      welcome: "/",
      growth: "/",
      plan: "/workspace",
      mistakes: "/report",
      drill: "/practice",
      followup: "/practice",
      experiences: "/resume-versions",
      star: "/resume-versions",
      onboarding: "/resume-versions",
      voice: "/",
      debrief: "/",
      debrief_full: "/",
      profile: "/",
    };
    for (const alias of OUT_OF_SCOPE_ALIASES) {
      const url = formatRouteUrl({ name: alias, params: {} });
      expect(
        url,
        `formatRouteUrl(${alias}) must collapse to a retained path`,
      ).toBe(expectations[alias]);
    }
  });

  it("serializeRouteToUrl emits no out-of-scope alias under any caller-provided params", () => {
    const outOfScopePathSet = new Set<string>(OUT_OF_SCOPE_PATHS);
    for (const alias of OUT_OF_SCOPE_ALIASES) {
      const parts = serializeRouteToUrl({
        name: alias,
        params: { token: "secret", rawText: "leak" },
      });
      expect(outOfScopePathSet.has(parts.path)).toBe(false);
    }
  });

  it("normalizeRouteName never returns an out-of-scope alias", () => {
    const outOfScopeAliasSet = new Set<string>(OUT_OF_SCOPE_ALIASES);
    for (const alias of OUT_OF_SCOPE_ALIASES) {
      const normalized = normalizeRouteName(alias);
      expect(
        outOfScopeAliasSet.has(normalized),
        `normalizeRouteName(${alias}) must NOT return another out-of-scope alias; got ${normalized}`,
      ).toBe(false);
    }
  });

  it("PRIMARY_NAV_ROUTES contains exactly the 3 retained primary nav entries (D-22)", () => {
    expect(PRIMARY_NAV_ROUTES).toEqual([
      "home",
      "workspace",
      "resume_versions",
    ]);
    for (const alias of OUT_OF_SCOPE_ALIASES) {
      expect(
        (PRIMARY_NAV_ROUTES as readonly string[]).includes(alias),
        `PRIMARY_NAV_ROUTES must not contain out-of-scope alias ${alias}`,
      ).toBe(false);
    }
  });

  it("CONTEXT_ROUTES / USER_MENU_ROUTES / AUTH_ROUTES contain no out-of-scope aliases", () => {
    const all = [
      ...CONTEXT_ROUTES,
      ...USER_MENU_ROUTES,
      ...AUTH_ROUTES,
    ] as readonly string[];
    for (const alias of OUT_OF_SCOPE_ALIASES) {
      expect(
        all.includes(alias),
        `Retained route catalog must not contain out-of-scope alias ${alias}`,
      ).toBe(false);
    }
  });
});
