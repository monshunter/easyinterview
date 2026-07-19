// @vitest-environment jsdom
/**
 * Code-level canonical routing and out-of-scope route regression.
 *
 * Truth source: docs/spec/frontend-shell/plans/004-url-addressable-routing/
 * This Vitest/jsdom suite is part of the frontend unit regression, not E2E.
 *
 * Validates that:
 *   - URL fragments never act as a parallel routing input.
 *   - Out-of-scope aliases (`welcome`, `growth`, `plan`, `mistakes`, `drill`,
 *     `followup`, `experiences`, `star`, `onboarding`, standalone `voice`,
 *     `debrief`, `debrief_full`, `profile`)
 *     never materialize standalone screens, canonical paths, scenario
 *     names or TopBar entries.
 *   - Server SPA fallback returns `index.html` for canonical frontend
 *     paths and never swallows `/api/*`, `/openapi/*`, `/health`,
 *     `/assets/*` or scenario script paths.
 */
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import { App } from "../App";
import { formatRouteUrl, ROUTE_TO_PATH } from "../routeUrl";
import { normalizeRouteName } from "../normalizeRoute";
import {
  FRONTEND_CANONICAL_PATHS,
  isCanonicalFrontendPath,
} from "../../../scripts/spaFallback.mjs";

function resetWindow(): void {
  delete (window as { __EASYINTERVIEW_INITIAL_ROUTE__?: unknown })
    .__EASYINTERVIEW_INITIAL_ROUTE__;
  window.history.replaceState(null, "", "/");
}

beforeEach(resetWindow);
afterEach(resetWindow);

describe("canonical and out-of-scope route negative regression", () => {
  it("ignores fragment route data and strips the fragment", () => {
    window.history.replaceState(null, "", "/#route=workspace&targetJobId=tj-1");
    render(<App />);
    expect(screen.getByTestId("home-hero-title")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/");
    expect(window.location.search).toBe("");
    expect(window.location.hash).toBe("");
  });

  it("unknown / malformed canonical path falls back to home without crashing", () => {
    window.history.replaceState(null, "", "/totally-unknown?foo=bar");
    render(<App />);
    expect(screen.getByTestId("home-hero-title")).toBeInTheDocument();
  });

  it("direct-open of `/voice` path normalizes to home (no standalone voice route)", () => {
    window.history.replaceState(null, "", "/voice?mode=voice");
    render(<App />);
    expect(screen.getByTestId("home-hero-title")).toBeInTheDocument();
  });

  it("ROUTE_TO_PATH must not include out-of-scope paths and must NOT have standalone /voice", () => {
    const all = Object.values(ROUTE_TO_PATH);
    expect(all).not.toContain("/voice");
    expect(all).not.toContain("/welcome");
    expect(all).not.toContain("/growth");
    expect(all).not.toContain("/plan");
    expect(all).not.toContain("/mistakes");
    expect(all).not.toContain("/drill");
    expect(all).not.toContain("/followup");
    expect(all).not.toContain("/experiences");
    expect(all).not.toContain("/star");
    expect(all).not.toContain("/onboarding");
    expect(all).not.toContain("/debrief");
    expect(all).not.toContain("/profile");
  });

  it("formatRouteUrl maps every out-of-scope alias to a retained canonical path", () => {
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
    for (const [alias, expectedPath] of Object.entries(expectations)) {
      expect(formatRouteUrl({ name: alias, params: {} })).toBe(expectedPath);
      expect(normalizeRouteName(alias)).not.toBe(alias);
    }
  });

  it("SPA fallback explicitly serves the known /reports path", () => {
    expect(
      isCanonicalFrontendPath(
        "/reports?targetJobId=01918fa0-0000-7000-8000-000000002000",
      ),
    ).toBe(true);
  });

  it("SPA host fallback covers every canonical frontend path and denies /api/*, /openapi/*, /health, /assets/*, file requests", () => {
    for (const canonical of Object.values(ROUTE_TO_PATH)) {
      expect(FRONTEND_CANONICAL_PATHS).toContain(canonical);
      expect(isCanonicalFrontendPath(canonical)).toBe(true);
    }
    for (const denied of [
      "/api/health",
      "/api/me",
      "/openapi/openapi.yaml",
      "/health",
      "/assets/main.js",
      "/index.html",
      "/workspace.json",
    ]) {
      expect(
        isCanonicalFrontendPath(denied),
        `${denied} must not be served by the SPA fallback`,
      ).toBe(false);
    }
  });
});
