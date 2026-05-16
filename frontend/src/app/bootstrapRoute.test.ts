import { describe, expect, it } from "vitest";

import { parseInitialRouteHash } from "./bootstrapRoute";

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
});
