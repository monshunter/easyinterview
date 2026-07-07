import { describe, expect, it } from "vitest";

import { normalizeRoute, normalizeRouteName } from "./normalizeRoute";

describe("normalizeRouteName", () => {
  it("maps every retained non-current alias documented in ui-design to a current route", () => {
    // Sourced from ui-design/src/app.jsx ROUTE_ALIASES + auth-and-entry.md §9.1.
    // `voice` is intentionally excluded: current product-scope keeps voice
    // only as practice route params, not as a route alias.
    expect(normalizeRouteName("welcome")).toBe("home");
    expect(normalizeRouteName("growth")).toBe("home");
    expect(normalizeRouteName("plan")).toBe("workspace");
    expect(normalizeRouteName("mistakes")).toBe("report");
    expect(normalizeRouteName("drill")).toBe("practice");
    expect(normalizeRouteName("followup")).toBe("practice");
    expect(normalizeRouteName("experiences")).toBe("resume_versions");
    expect(normalizeRouteName("star")).toBe("resume_versions");
    expect(normalizeRouteName("resume")).toBe("resume_versions");
    expect(normalizeRouteName("onboarding")).toBe("resume_versions");
  });

  it("does not preserve the non-current standalone voice route alias", () => {
    expect(normalizeRouteName("voice")).toBe("home");
  });

  it("normalizes the non-current auth_reset route to the single login entry", () => {
    // product-scope D-16 / frontend-shell spec v1.22 C-17 — email code is the
    // only sign-in flow; auth_reset must not materialize a standalone screen.
    expect(normalizeRouteName("auth_reset")).toBe("auth_login");
  });

  it("normalizes non-current debrief/profile aliases to home", () => {
    // product-scope D-22 keeps debrief and user-profile outside the current
    // route catalog. Saved local route names fold back to the core intake home.
    expect(normalizeRouteName("debrief")).toBe("home");
    expect(normalizeRouteName("debrief_full")).toBe("home");
    expect(normalizeRouteName("profile")).toBe("home");
  });

  it("does not retain standalone insight route aliases", () => {
    expect(normalizeRouteName("standalone_insight")).toBe("home");
  });

  it("preserves valid current route names", () => {
    expect(normalizeRouteName("home")).toBe("home");
    expect(normalizeRouteName("workspace")).toBe("workspace");
    expect(normalizeRouteName("auth_login")).toBe("auth_login");
    expect(normalizeRouteName("settings")).toBe("settings");
  });

  it("falls back to home for unknown names", () => {
    expect(normalizeRouteName("totally-bogus")).toBe("home");
    expect(normalizeRouteName("")).toBe("home");
  });
});

describe("normalizeRoute", () => {
  it("normalizes name and preserves params", () => {
    const result = normalizeRoute({
      name: "drill",
      params: { planId: "plan-1", roundId: "round-2" },
    });
    expect(result.name).toBe("practice");
    expect(result.params).toEqual({ planId: "plan-1", roundId: "round-2" });
  });

  it("defaults missing params to empty record", () => {
    const result = normalizeRoute({ name: "welcome" });
    expect(result.name).toBe("home");
    expect(result.params).toEqual({});
  });
});
