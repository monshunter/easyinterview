import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

import {
  formatRouteUrl,
  isSafeRouteParam,
  parseUrlToRoute,
  ROUTE_TO_PATH,
  serializeRouteToUrl,
} from "./routeUrl";

describe("ROUTE_TO_PATH catalog", () => {
  it("does not expose canonical-route helpers without repository consumers", () => {
    const source = readFileSync(resolve(process.cwd(), "src/app/routeUrl.ts"), "utf8");
    expect(source).not.toContain("routeUrlsEqual");
  });

  it("covers every retained route with a canonical path", () => {
    expect(ROUTE_TO_PATH).toEqual({
      home: "/",
      workspace: "/workspace",
      resume_versions: "/resume-versions",
      parse: "/parse",
      practice: "/practice",
      reports: "/reports",
      generating: "/generating",
      report: "/report",
      settings: "/settings",
      auth_login: "/auth/login",
      auth_verify: "/auth/verify",
      auth_profile_setup: "/auth/profile",
      auth_logout: "/auth/logout",
    });
  });
});

describe("serializeRouteToUrl", () => {
  it("returns root path for home with no params", () => {
    expect(serializeRouteToUrl({ name: "home", params: {} })).toEqual({
      path: "/",
      search: "",
    });
  });

  it("retains targetJobId as the sole workspace detail locator", () => {
    expect(
      formatRouteUrl({
        name: "workspace",
        params: {
          targetJobId: "tj-1",
          resumeId: "rv-1",
          planId: "plan-1",
          autoStartPractice: "1",
        },
      }),
    ).toBe("/workspace?targetJobId=tj-1");
  });

  it("drops empty and unknown params", () => {
    expect(
      formatRouteUrl({
        name: "workspace",
        params: {
          targetJobId: "tj-1",
          rawText: "raw JD body",
          unknownKey: "x",
          planId: "",
        },
      }),
    ).toBe("/workspace?targetJobId=tj-1");
  });

  it("serializes the out-of-scope jd_match route name to the home canonical path (D-17)", () => {
    expect(
      formatRouteUrl({
        name: "jd_match",
        params: { tab: "search", query: "principal engineer" },
      }),
    ).toBe("/");
  });

  it("serializes standalone insight route names to home", () => {
    expect(
      formatRouteUrl({
        name: "standalone_insight",
        params: {
          targetJobId: "tj-1",
          jdId: "jd-1",
          companyId: "company-private",
        },
      }),
    ).toBe("/");
  });

  it("serializes Reports with targetJobId as its only safe context", () => {
    expect(
      formatRouteUrl({
        name: "reports",
        params: {
          targetJobId: "tj-1",
          section: "reports",
          reportId: "rpt-1",
          status: "ready",
          roundId: "round-1",
          rawText: "private JD body",
        },
      }),
    ).toBe("/reports?targetJobId=tj-1");
  });

  it("retains only reportId for generating/report deep links", () => {
    expect(
      formatRouteUrl({
        name: "generating",
        params: { sessionId: "s-1", reportId: "rpt-1" },
      }),
    ).toBe("/generating?reportId=rpt-1");

    expect(
      formatRouteUrl({
        name: "report",
        params: {
          sessionId: "s-1",
          reportId: "rpt-1",
          reportStatus: "failed",
          errorCode: "AI_PROVIDER_TIMEOUT",
        },
      }),
    ).toBe("/report?reportId=rpt-1");

    expect(
      formatRouteUrl({
        name: "resume_versions",
        params: { resumeId: "v-1", tab: "rewrites", tailorRunId: "tr-1" },
      }),
    ).toBe("/resume-versions?resumeId=v-1");
  });

  it("drops legacy phone mode params under canonical path", () => {
    expect(
      formatRouteUrl({
        name: "practice",
        params: { mode: "phone", modality: "phone", sessionId: "s-1" },
      }),
    ).toBe("/practice?sessionId=s-1");
  });

  it("drops out-of-scope voice mode values from canonical params", () => {
    expect(
      formatRouteUrl({
        name: "practice",
        params: { mode: "voice", modality: "voice", sessionId: "s-1" },
      }),
    ).toBe("/practice?sessionId=s-1");
  });

  it("emits only the opaque Home import handoff identifier", () => {
    expect(
      formatRouteUrl({
        name: "home",
        params: {
          opaquePendingImportId: "imp-1",
          pendingImportId: "legacy-imp",
          source: "paste",
          resumeId: "resume-secret",
        },
      }),
    ).toBe("/?opaquePendingImportId=imp-1");
  });

  it("retains only targetJobId on the Parse command/progress URL", () => {
    expect(
      formatRouteUrl({
        name: "parse",
        params: {
          targetJobId: "tj-1",
          resumeId: "rv-1",
          jdId: "jd-1",
          importId: "imp-1",
          source: "paste",
        },
      }),
    ).toBe("/parse?targetJobId=tj-1");
  });

  it("drops the retired Parse reports section and all report business authority", () => {
    expect(
      formatRouteUrl({
        name: "parse",
        params: {
          targetJobId: "tj-1",
          section: "reports",
          reportId: "report-route-must-not-own",
          reportStatus: "ready",
          status: "failed",
          roundId: "round-route-must-not-own",
          roundSequence: "99",
        },
      }),
    ).toBe("/parse?targetJobId=tj-1");

    expect(
      parseUrlToRoute(
        "/parse?targetJobId=tj-1&section=reports&reportId=route-report&status=ready&roundId=route-round",
      ),
    ).toEqual({
      name: "parse",
      params: { targetJobId: "tj-1" },
    });
  });

  it("drops every Parse section value", () => {
    expect(
      formatRouteUrl({
        name: "parse",
        params: { targetJobId: "tj-1", section: "timeline" },
      }),
    ).toBe("/parse?targetJobId=tj-1");
    expect(
      parseUrlToRoute("/parse?targetJobId=tj-1&section=timeline"),
    ).toEqual({ name: "parse", params: { targetJobId: "tj-1" } });
  });

  it("auth pendingAction cannot restore the retired Parse reports section", () => {
    expect(
      parseUrlToRoute(
        "/auth/login?pendingRoute=parse&pendingType=open_protected_route&pendingLabel=parse&targetJobId=tj-1&section=reports&reportId=rpt-1&status=ready&roundId=round-1",
      ),
    ).toEqual({
      name: "auth_login",
      params: {
        pendingRoute: "parse",
        pendingType: "open_protected_route",
        pendingLabel: "parse",
        targetJobId: "tj-1",
      },
    });
  });


  it("normalizes out-of-scope route names back to retained routes", () => {
    expect(serializeRouteToUrl({ name: "welcome", params: {} }).path).toBe("/");
    expect(serializeRouteToUrl({ name: "growth", params: {} }).path).toBe("/");
    expect(serializeRouteToUrl({ name: "plan", params: {} }).path).toBe(
      "/workspace",
    );
    expect(serializeRouteToUrl({ name: "mistakes", params: {} }).path).toBe(
      "/report",
    );
    expect(serializeRouteToUrl({ name: "drill", params: {} }).path).toBe(
      "/practice",
    );
    expect(serializeRouteToUrl({ name: "followup", params: {} }).path).toBe(
      "/practice",
    );
    expect(serializeRouteToUrl({ name: "experiences", params: {} }).path).toBe(
      "/resume-versions",
    );
    expect(serializeRouteToUrl({ name: "star", params: {} }).path).toBe(
      "/resume-versions",
    );
    expect(serializeRouteToUrl({ name: "onboarding", params: {} }).path).toBe(
      "/resume-versions",
    );
    expect(serializeRouteToUrl({ name: "auth_register", params: {} }).path).toBe(
      "/auth/login",
    );
    expect(serializeRouteToUrl({ name: "voice", params: {} }).path).toBe("/");
    expect(serializeRouteToUrl({ name: "debrief", params: {} }).path).toBe("/");
    expect(serializeRouteToUrl({ name: "debrief_full", params: {} }).path).toBe(
      "/",
    );
    expect(serializeRouteToUrl({ name: "profile", params: {} }).path).toBe("/");
  });

  it("auth_login carries pendingAction safe params union with target route", () => {
    expect(
      formatRouteUrl({
        name: "auth_login",
        params: {
          next: "/workspace",
          email: "alice@example.com",
          pendingRoute: "workspace",
          pendingType: "start_practice",
          pendingLabel: "立即面试",
          planId: "plan-1",
          targetJobId: "tj-1",
          jdId: "jd-1",
          resumeId: "rv-1",
          roundId: "round-1",
          rawText: "raw JD body",
        },
      }),
    ).toBe(
      "/auth/login?email=alice%40example.com&next=%2Fworkspace&pendingLabel=%E7%AB%8B%E5%8D%B3%E9%9D%A2%E8%AF%95&pendingRoute=workspace&pendingType=start_practice&targetJobId=tj-1",
    );
  });

  it("auth_login without pendingRoute keeps only base safe params", () => {
    expect(
      formatRouteUrl({
        name: "auth_login",
        params: {
          next: "/workspace",
          email: "alice@example.com",
          planId: "plan-1",
        },
      }),
    ).toBe("/auth/login?email=alice%40example.com&next=%2Fworkspace");
  });

  it("auth_verify drops raw auth token query because email code is form-only", () => {
    expect(
      formatRouteUrl({
        name: "auth_verify",
        params: {
          email: "alice@example.com",
          token: "123456",
          pendingRoute: "workspace",
          pendingType: "start_practice",
          pendingLabel: "立即面试",
          targetJobId: "tj-1",
        },
      }),
    ).toBe(
      "/auth/verify?email=alice%40example.com&pendingLabel=%E7%AB%8B%E5%8D%B3%E9%9D%A2%E8%AF%95&pendingRoute=workspace&pendingType=start_practice&targetJobId=tj-1",
    );
    expect(
      formatRouteUrl({
        name: "auth_login",
        params: { email: "alice@example.com", token: "123456" },
      }),
    ).toBe("/auth/login?email=alice%40example.com");
  });

  it("auth_profile_setup carries pendingAction params but drops raw form data", () => {
    expect(
      formatRouteUrl({
        name: "auth_profile_setup",
        params: {
          email: "alice@example.com",
          pendingRoute: "practice",
          pendingType: "start_practice",
          pendingLabel: "立即面试",
          planId: "plan-1",
          displayName: "Alice",
          token: "123456",
        },
      }),
    ).toBe(
      "/auth/profile?email=alice%40example.com&pendingLabel=%E7%AB%8B%E5%8D%B3%E9%9D%A2%E8%AF%95&pendingRoute=practice&pendingType=start_practice&planId=plan-1",
    );
  });

  it("drops report replay params from workspace because replay starts from report owner", () => {
    expect(
      formatRouteUrl({
        name: "workspace",
        params: {
          targetJobId: "tj-1",
          resumeId: "rv-1",
          planId: "plan-1",
          autoStartPractice: "1",
          sourceSessionId: "s-prior",
          replayItems: "turn-1,turn-3",
          evidenceGaps: "technical_depth|narrative",
          nextRoundId: "round-2",
        },
      }),
    ).toBe("/workspace?targetJobId=tj-1");
  });

  it("drops raw payload, AI prompt, auth secret keys from URL even when present", () => {
    const PRIVATE_KEYS = [
      "rawText",
      "rawDescription",
      "sourceUrl",
      "query",
      "label",
      "guidedAnswers",
      "parsedSummary",
      "structuredProfile",
      "suggestion",
      "originalBullet",
      "suggestedBullet",
      "questionText",
      "answerText",
      "notes",
      "prompt",
      "response",
      "file",
      "token",
      "password",
    ];
    for (const key of PRIVATE_KEYS) {
      const url = formatRouteUrl({
        name: "workspace",
        params: { targetJobId: "tj-1", [key]: "leaked-value" },
      });
      expect(url, `workspace must drop ${key}`).toBe(
        "/workspace?targetJobId=tj-1",
      );
    }
    for (const key of PRIVATE_KEYS) {
      const url = formatRouteUrl({
        name: "auth_login",
        params: {
          pendingRoute: "workspace",
          pendingType: "start_practice",
          pendingLabel: "立即面试",
          targetJobId: "tj-1",
          [key]: "leaked-value",
        },
      });
      expect(url, `auth_login must drop ${key}`).not.toContain("leaked-value");
    }
  });
});

describe("parseUrlToRoute", () => {
  it("parses root path to home with empty params", () => {
    expect(parseUrlToRoute("/")).toEqual({ name: "home", params: {} });
  });

  it("normalizes the out-of-scope /auth/reset path back to the login entry", () => {
    // product-scope D-16 — auth_reset is outside the current route catalog;
    // the path must land on auth_login instead of materializing a reset screen.
    expect(parseUrlToRoute("/auth/reset")).toEqual({
      name: "auth_login",
      params: {},
    });
  });

  it("parses targetJobId as the sole canonical workspace detail locator", () => {
    expect(
      parseUrlToRoute(
        "/workspace?targetJobId=tj-1&resumeId=rv-1&planId=plan-1&autoStartPractice=1",
      ),
    ).toEqual({
      name: "workspace",
      params: { targetJobId: "tj-1" },
    });
  });

  it("parses Reports with targetJobId only", () => {
    expect(
      parseUrlToRoute(
        "/reports?targetJobId=tj-1&section=reports&reportId=rpt-1&status=ready&roundId=round-1",
      ),
    ).toEqual({
      name: "reports",
      params: { targetJobId: "tj-1" },
    });
  });

  it("parses canonical report deep links with reportId as the sole locator", () => {
    expect(
      parseUrlToRoute(
        "/report?sessionId=s-1&reportId=rpt-1&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT",
      ),
    ).toEqual({
      name: "report",
      params: {
        reportId: "rpt-1",
      },
    });
  });

  it("drops out-of-scope voice mode values during canonical parse", () => {
    expect(
      parseUrlToRoute("/practice?sessionId=s-1&mode=voice&modality=voice"),
    ).toEqual({
      name: "practice",
      params: {
        sessionId: "s-1",
      },
    });
    expect(
      parseUrlToRoute("/practice?sessionId=s-1&mode=phone&modality=phone"),
    ).toEqual({
      name: "practice",
      params: {
        sessionId: "s-1",
      },
    });
  });

  it("drops unknown / unsafe params during parse", () => {
    expect(
      parseUrlToRoute("/workspace?targetJobId=tj-1&rawText=raw+jd&query=secret"),
    ).toEqual({
      name: "workspace",
      params: { targetJobId: "tj-1" },
    });
  });

  it("falls back to home on unknown path", () => {
    expect(parseUrlToRoute("/unknown")).toEqual({ name: "home", params: {} });
    expect(parseUrlToRoute("/standalone-insight")).toEqual({
      name: "home",
      params: {},
    });
    expect(parseUrlToRoute("/voice")).toEqual({ name: "home", params: {} });
    expect(parseUrlToRoute("/welcome")).toEqual({ name: "home", params: {} });
    expect(parseUrlToRoute("/auth/register")).toEqual({
      name: "home",
      params: {},
    });
  });

  it("parses auth pendingAction restore URL", () => {
    expect(
      parseUrlToRoute(
        "/auth/login?pendingRoute=workspace&pendingType=start_practice&pendingLabel=%E7%AB%8B%E5%8D%B3%E9%9D%A2%E8%AF%95&planId=plan-1&targetJobId=tj-1",
      ),
    ).toEqual({
      name: "auth_login",
      params: {
        pendingRoute: "workspace",
        pendingType: "start_practice",
        pendingLabel: "立即面试",
        targetJobId: "tj-1",
      },
    });
  });

  it("drops auth_verify token query because email code is form-only", () => {
    expect(
      parseUrlToRoute("/auth/verify?email=alice%40example.com&token=123456"),
    ).toEqual({
      name: "auth_verify",
      params: {
        email: "alice@example.com",
      },
    });
    expect(parseUrlToRoute("/auth/login?token=123456")).toEqual({
      name: "auth_login",
      params: {},
    });
  });

  it("parses auth_profile_setup pendingAction restore URL", () => {
    expect(
      parseUrlToRoute(
        "/auth/profile?pendingRoute=workspace&pendingType=complete_profile_resume&pendingLabel=workspace&planId=plan-1&displayName=Alice",
      ),
    ).toEqual({
      name: "auth_profile_setup",
      params: {
        pendingRoute: "workspace",
        pendingType: "complete_profile_resume",
        pendingLabel: "workspace",
      },
    });
  });

  it("supports plain URL input without leading slash", () => {
    expect(parseUrlToRoute("workspace?targetJobId=tj-1")).toEqual({
      name: "workspace",
      params: { targetJobId: "tj-1" },
    });
  });

  it("strips fragment during canonical parse (hash adapter remains separate)", () => {
    expect(
      parseUrlToRoute("/workspace?targetJobId=tj-1#anything"),
    ).toEqual({ name: "workspace", params: { targetJobId: "tj-1" } });
  });

  it("normalizes the opaque home query-only deep link", () => {
    expect(
      parseUrlToRoute(
        "/?opaquePendingImportId=imp-1&pendingImportId=legacy&source=paste&resumeId=rv-1",
      ),
    ).toEqual({
      name: "home",
      params: { opaquePendingImportId: "imp-1" },
    });
  });

  it("returns home when input is empty or malformed", () => {
    expect(parseUrlToRoute("")).toEqual({ name: "home", params: {} });
    expect(parseUrlToRoute("?targetJobId=tj-1")).toEqual({
      name: "home",
      params: {},
    });
  });
});

describe("isSafeRouteParam", () => {
  it("approves only the minimal cross-owner safe params", () => {
    expect(isSafeRouteParam("home", "opaquePendingImportId", {})).toBe(true);
    expect(isSafeRouteParam("home", "pendingImportId", {})).toBe(false);
    expect(isSafeRouteParam("home", "source", {})).toBe(false);
    expect(isSafeRouteParam("home", "resumeId", {})).toBe(false);
    expect(isSafeRouteParam("workspace", "targetJobId", {})).toBe(true);
    expect(isSafeRouteParam("workspace", "autoStartPractice", {})).toBe(false);
    expect(isSafeRouteParam("workspace", "sourceSessionId", {})).toBe(false);
    expect(isSafeRouteParam("workspace", "sourceReportId", {})).toBe(false);
    expect(isSafeRouteParam("workspace", "replayItems", {})).toBe(false);
    expect(isSafeRouteParam("workspace", "evidenceGaps", {})).toBe(false);
    expect(isSafeRouteParam("workspace", "nextRoundId", {})).toBe(false);
    expect(isSafeRouteParam("report", "reportId", {})).toBe(true);
    expect(isSafeRouteParam("report", "reportStatus", {})).toBe(false);
    expect(isSafeRouteParam("report", "errorCode", {})).toBe(false);
    expect(isSafeRouteParam("generating", "sessionId", {})).toBe(false);
    expect(isSafeRouteParam("resume_versions", "tailorRunId", {})).toBe(false);
    expect(isSafeRouteParam("parse", "targetJobId", {})).toBe(true);
    expect(isSafeRouteParam("parse", "resumeId", {})).toBe(false);
    expect(isSafeRouteParam("parse", "jdId", {})).toBe(false);
    expect(isSafeRouteParam("parse", "importId", {})).toBe(false);
    expect(isSafeRouteParam("parse", "source", {})).toBe(false);
  });

  it("denies raw payload / AI prompt / auth secret keys on every route", () => {
    const forbidden = [
      "rawText",
      "rawDescription",
      "sourceUrl",
      "query",
      "label",
      "guidedAnswers",
      "parsedSummary",
      "structuredProfile",
      "suggestion",
      "originalBullet",
      "suggestedBullet",
      "questionText",
      "answerText",
      "notes",
      "prompt",
      "response",
      "file",
      "password",
    ];
    for (const key of forbidden) {
      expect(isSafeRouteParam("home", key, {})).toBe(false);
      expect(isSafeRouteParam("workspace", key, {})).toBe(false);
      expect(isSafeRouteParam("practice", key, {})).toBe(false);
      expect(isSafeRouteParam("auth_login", key, {
        pendingRoute: "workspace",
      })).toBe(false);
    }
    expect(isSafeRouteParam("auth_verify", "token", {})).toBe(false);
    expect(isSafeRouteParam("auth_login", "token", {})).toBe(false);
    expect(isSafeRouteParam("workspace", "token", {})).toBe(false);
    // product-scope D-17: the jd_match -> parse reverse handoff param is
    // outside the current parse allowlist.
    expect(isSafeRouteParam("parse", "sourceJobMatchId", {})).toBe(false);
  });

  it("expands auth_login allowlist with target route safe params when pendingRoute is present", () => {
    expect(
      isSafeRouteParam("auth_login", "planId", { pendingRoute: "workspace" }),
    ).toBe(false);
    expect(
      isSafeRouteParam("auth_login", "planId", { pendingRoute: "practice" }),
    ).toBe(true);
    expect(
      isSafeRouteParam("auth_login", "tailorRunId", {
        pendingRoute: "resume_versions",
      }),
    ).toBe(false);
    expect(
      isSafeRouteParam("auth_login", "debriefId", {
        pendingRoute: "debrief",
      }),
    ).toBe(false);
  });

  it("auth_login without pendingRoute keeps only base allowlist", () => {
    expect(isSafeRouteParam("auth_login", "planId", {})).toBe(false);
    expect(isSafeRouteParam("auth_login", "next", {})).toBe(true);
    expect(isSafeRouteParam("auth_login", "email", {})).toBe(true);
  });
});
