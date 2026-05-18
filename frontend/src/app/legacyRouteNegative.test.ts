/**
 * Plan 004 Phase 4.2 — legacy route negative regression.
 *
 * Asserts that retired routes (`welcome`, `growth`, `plan`, `mistakes`,
 * `drill`, `followup`, `experiences`, `star`, `onboarding`, standalone
 * `voice`) never materialize as:
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

const RETIRED_ALIASES = [
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
  "debrief_full",
] as const;

const RETIRED_PATHS = [
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
  "/debrief-full",
] as const;

describe("Plan 004 Phase 4.2 — legacy route negative regression", () => {
  it("ROUTE_TO_PATH does not include retired paths", () => {
    const allPaths = Object.values(ROUTE_TO_PATH);
    for (const path of RETIRED_PATHS) {
      expect(
        allPaths.includes(path),
        `routeUrl.ROUTE_TO_PATH must not include retired path ${path}`,
      ).toBe(false);
    }
  });

  it("SPA fallback FRONTEND_CANONICAL_PATHS does not include retired paths", () => {
    for (const path of RETIRED_PATHS) {
      expect(
        FRONTEND_CANONICAL_PATHS.includes(path),
        `spaFallback.FRONTEND_CANONICAL_PATHS must not include retired path ${path}`,
      ).toBe(false);
    }
  });

  it("formatRouteUrl normalizes retired aliases to retained canonical paths (no standalone)", () => {
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
      debrief_full: "/debrief",
    };
    for (const alias of RETIRED_ALIASES) {
      const url = formatRouteUrl({ name: alias, params: {} });
      expect(
        url,
        `formatRouteUrl(${alias}) must collapse to a retained path`,
      ).toBe(expectations[alias]);
    }
  });

  it("serializeRouteToUrl emits no retired alias under any caller-provided params", () => {
    const retiredPathSet = new Set<string>(RETIRED_PATHS);
    for (const alias of RETIRED_ALIASES) {
      const parts = serializeRouteToUrl({
        name: alias,
        params: { token: "secret", rawText: "leak" },
      });
      expect(retiredPathSet.has(parts.path)).toBe(false);
    }
  });

  it("normalizeRouteName never returns a retired alias", () => {
    const retiredAliasSet = new Set<string>(RETIRED_ALIASES);
    for (const alias of RETIRED_ALIASES) {
      const normalized = normalizeRouteName(alias);
      expect(
        retiredAliasSet.has(normalized),
        `normalizeRouteName(${alias}) must NOT return another retired alias; got ${normalized}`,
      ).toBe(false);
    }
  });

  it("PRIMARY_NAV_ROUTES contains exactly the 5 retained primary nav entries", () => {
    expect(PRIMARY_NAV_ROUTES).toEqual([
      "home",
      "jd_match",
      "workspace",
      "resume_versions",
      "debrief",
    ]);
    for (const alias of RETIRED_ALIASES) {
      expect(
        (PRIMARY_NAV_ROUTES as readonly string[]).includes(alias),
        `PRIMARY_NAV_ROUTES must not contain retired alias ${alias}`,
      ).toBe(false);
    }
  });

  it("CONTEXT_ROUTES / USER_MENU_ROUTES / AUTH_ROUTES contain no retired aliases", () => {
    const all = [
      ...CONTEXT_ROUTES,
      ...USER_MENU_ROUTES,
      ...AUTH_ROUTES,
    ] as readonly string[];
    for (const alias of RETIRED_ALIASES) {
      expect(
        all.includes(alias),
        `Retained route catalog must not contain retired alias ${alias}`,
      ).toBe(false);
    }
  });
});
