// @vitest-environment jsdom
/**
 * E2E.P0.090 — Hash routing + out-of-scope route negative regression.
 *
 * Truth source: docs/spec/frontend-shell/plans/004-url-addressable-routing/
 * bdd-plan.md §2 (E2E.P0.090) + bdd-checklist.md.
 *
 * Validates that:
 *   - `#route=...` static-preview / pixel-parity entries still bootstrap
 *     through `normalizeRoute` and produce equivalent canonical paths.
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
import { render, screen, waitFor } from "@testing-library/react";

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

describe("E2E.P0.090 hash routing + out-of-scope route negative regression", () => {
  it("`#route=home` bootstrap renders home and rewrites URL to `/`", () => {
    window.history.replaceState(null, "", "/#route=home");
    render(<App />);
    expect(screen.getByTestId("home-hero-label")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/");
    expect(window.location.hash).toBe("");
  });

  it("`#route=workspace&targetJobId=...` bootstrap rewrites to target-scoped detail", () => {
    window.history.replaceState(
      null,
      "",
      "/#route=workspace&targetJobId=tj-1",
    );
    render(<App />);
    expect(window.location.pathname).toBe("/workspace");
    expect(window.location.search).toBe("?targetJobId=tj-1");
    expect(window.location.hash).toBe("");
    expect(screen.getByTestId("workspace-detail-loading")).toBeInTheDocument();
  });

  it("Reports hash bootstrap keeps targetJobId only and never adds a TopBar entry", async () => {
    window.history.replaceState(
      null,
      "",
      "/#route=reports&targetJobId=01918fa0-0000-7000-8000-000000002000&section=reports&reportId=rpt-hostile&status=ready&roundId=round-hostile",
    );
    render(<App />);
    await waitFor(() => screen.getByTestId("reports-screen"));
    expect(window.location.pathname + window.location.search).toBe(
      "/reports?targetJobId=01918fa0-0000-7000-8000-000000002000",
    );
    expect(window.location.hash).toBe("");
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(screen.queryByTestId("topbar-nav-reports")).not.toBeInTheDocument();
  });

  it("legacy Parse report params are stripped instead of restoring an embedded report section", async () => {
    window.history.replaceState(
      null,
      "",
      "/#route=parse&targetJobId=tj-1&section=reports&reportId=rpt-hostile&status=ready&roundId=round-hostile",
    );
    render(<App />);
    await waitFor(() => expect(window.location.pathname).toBe("/parse"));
    expect(window.location.search).toBe("?targetJobId=tj-1");
    expect(window.location.hash).toBe("");
    for (const forbidden of ["section", "reportId", "status", "roundId"]) {
      expect(window.location.search).not.toContain(`${forbidden}=`);
    }
  });

  it("legacy phone hash values are dropped and voice stays disabled", () => {
    window.history.replaceState(
      null,
      "",
      "/#route=practice&mode=phone&modality=phone&sessionId=01918fa0-0000-7000-8000-000000005000",
    );
    render(<App />);
    expect(window.location.pathname).toBe("/practice");
    expect(window.location.search).not.toContain("mode=phone");
    expect(window.location.search).not.toContain("modality=phone");
    expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-conversation")).toBeInTheDocument();
    expect(screen.getByTestId("practice-topbar-phone-toggle")).toBeDisabled();
  });

  it("out-of-scope voice mode hash values are dropped without mounting phone surface", () => {
    window.history.replaceState(
      null,
      "",
      "/#route=practice&mode=voice&modality=voice&sessionId=01918fa0-0000-7000-8000-000000005000",
    );
    render(<App />);
    expect(window.location.pathname).toBe("/practice");
    expect(window.location.search).toContain(
      "sessionId=01918fa0-0000-7000-8000-000000005000",
    );
    expect(window.location.search).not.toContain("mode=voice");
    expect(window.location.search).not.toContain("modality=voice");
    expect(screen.getByTestId("practice-input")).toBeInTheDocument();
    expect(
      screen.queryByTestId("practice-phone-waveform"),
    ).not.toBeInTheDocument();
  });

  it("`#route=voice` normalizes to home (standalone voice route never materializes)", () => {
    window.history.replaceState(null, "", "/#route=voice");
    render(<App />);
    expect(window.location.pathname).toBe("/");
    expect(window.location.hash).toBe("");
    expect(screen.getByTestId("home-hero-label")).toBeInTheDocument();
  });

  it("out-of-scope aliases via hash all normalize to retained routes (welcome / growth / plan / mistakes / drill / followup / experiences / star / onboarding / debrief / debrief_full / profile)", () => {
    const cases: Array<[string, string]> = [
      ["#route=welcome", "/"],
      ["#route=growth", "/"],
      ["#route=plan", "/workspace"],
      ["#route=mistakes", "/report"],
      ["#route=drill", "/practice"],
      ["#route=followup", "/practice"],
      ["#route=experiences", "/resume-versions"],
      ["#route=star", "/resume-versions"],
      ["#route=onboarding", "/resume-versions"],
      ["#route=debrief", "/"],
      ["#route=debrief_full", "/"],
      ["#route=profile", "/"],
    ];
    for (const [hash, expectedPath] of cases) {
      resetWindow();
      window.history.replaceState(null, "", `/${hash}`);
      const { unmount } = render(<App />);
      expect(
        window.location.pathname,
        `${hash} must rewrite to ${expectedPath}`,
      ).toBe(expectedPath);
      expect(window.location.hash).toBe("");
      unmount();
    }
  });

  it("unknown / malformed canonical path falls back to home without crashing", () => {
    window.history.replaceState(null, "", "/totally-unknown?foo=bar");
    render(<App />);
    expect(screen.getByTestId("home-hero-label")).toBeInTheDocument();
  });

  it("direct-open of `/voice` path normalizes to home (no standalone voice route)", () => {
    window.history.replaceState(null, "", "/voice?mode=voice");
    render(<App />);
    expect(screen.getByTestId("home-hero-label")).toBeInTheDocument();
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
