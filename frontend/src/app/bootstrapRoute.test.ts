import { describe, expect, it } from "vitest";

import { parseInitialRouteHash } from "./bootstrapRoute";
import { normalizeRoute } from "./normalizeRoute";
import { formatRouteUrl } from "./routeUrl";

describe("parseInitialRouteHash", () => {
  it("extracts report hash routes for static pixel parity entry", () => {
    expect(
      parseInitialRouteHash(
        "#route=report&reportId=report-1&sessionId=session-1&targetJobId=target-1",
      ),
    ).toEqual({
      name: "report",
      params: {
        reportId: "report-1",
        sessionId: "session-1",
        targetJobId: "target-1",
      },
    });
  });

  it("extracts generating hash routes and leaves normalization to App", () => {
    expect(
      parseInitialRouteHash("#route=generating&reportId=report-1&sessionId=session-1"),
    ).toEqual({
      name: "generating",
      params: {
        reportId: "report-1",
        sessionId: "session-1",
      },
    });
  });

  it("returns undefined when the hash does not carry a route", () => {
    expect(parseInitialRouteHash("")).toBeUndefined();
    expect(parseInitialRouteHash("#reportId=report-1")).toBeUndefined();
  });

  it("hash adapter and canonical codec drop out-of-scope workspace context params", () => {
    const loose = parseInitialRouteHash(
      "#route=workspace&targetJobId=tj-1&resumeId=rv-1&planId=plan-1",
    );
    expect(loose).toBeDefined();
    const normalized = normalizeRoute(loose!);
    expect(normalized).toEqual({
      name: "workspace",
      params: {
        targetJobId: "tj-1",
        resumeId: "rv-1",
        planId: "plan-1",
      },
    });
    expect(formatRouteUrl(normalized)).toBe("/workspace");
  });

  it("out-of-scope aliases via hash normalize to retained routes without materializing standalone screens", () => {
    const cases: Array<[string, { name: string; path: string }]> = [
      ["#route=welcome", { name: "home", path: "/" }],
      ["#route=growth", { name: "home", path: "/" }],
      ["#route=plan", { name: "workspace", path: "/workspace" }],
      ["#route=mistakes", { name: "report", path: "/report" }],
      ["#route=drill", { name: "practice", path: "/practice" }],
      ["#route=experiences", { name: "resume_versions", path: "/resume-versions" }],
      ["#route=voice", { name: "home", path: "/" }],
    ];
    for (const [hash, expected] of cases) {
      const loose = parseInitialRouteHash(hash);
      expect(loose, `hash ${hash} must extract loose route`).toBeDefined();
      const normalized = normalizeRoute(loose!);
      expect(normalized.name).toBe(expected.name);
      expect(formatRouteUrl(normalized).startsWith(expected.path)).toBe(true);
    }
  });

  it("hash phone mode practice entry maps to canonical /practice with mode/modality", () => {
    const loose = parseInitialRouteHash(
      "#route=practice&mode=phone&modality=phone&sessionId=s-1",
    );
    const normalized = normalizeRoute(loose!);
    expect(formatRouteUrl(normalized)).toBe(
      "/practice?modality=phone&mode=phone&sessionId=s-1",
    );
  });

  it("hash route drops out-of-scope voice mode values before canonical rewrite", () => {
    const loose = parseInitialRouteHash(
      "#route=practice&mode=voice&modality=voice&sessionId=s-1",
    );
    const normalized = normalizeRoute(loose!);
    expect(formatRouteUrl(normalized)).toBe(
      "/practice?sessionId=s-1",
    );
  });
});
