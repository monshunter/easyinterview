import { describe, expect, it } from "vitest";

import {
  formatRouteUrl,
  isSafeRouteParam,
  parseUrlToRoute,
  ROUTE_TO_PATH,
  serializeRouteToUrl,
} from "./routeUrl";

describe("ROUTE_TO_PATH catalog", () => {
  it("covers every retained route with a canonical path", () => {
    expect(ROUTE_TO_PATH).toEqual({
      home: "/",
      workspace: "/workspace",
      resume_versions: "/resume-versions",
      debrief: "/debrief",
      parse: "/parse",
      practice: "/practice",
      generating: "/generating",
      report: "/report",
      profile: "/profile",
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

  it("emits sorted query string for workspace handoff params", () => {
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
    ).toBe(
      "/workspace?autoStartPractice=1&planId=plan-1&resumeId=rv-1&targetJobId=tj-1",
    );
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

  it("serializes the retired jd_match route name to the home canonical path (D-17)", () => {
    expect(
      formatRouteUrl({
        name: "jd_match",
        params: { tab: "search", query: "principal engineer" },
      }),
    ).toBe("/");
  });

  it("serializes the retired company_intel route name to workspace (D-18)", () => {
    expect(
      formatRouteUrl({
        name: "company_intel",
        params: {
          targetJobId: "tj-1",
          jdId: "jd-1",
          companyId: "company-private",
        },
      }),
    ).toBe("/workspace?jdId=jd-1&targetJobId=tj-1");
  });

  it("retains generating/report/resume_versions/debrief deep-link params", () => {
    expect(
      formatRouteUrl({
        name: "generating",
        params: { sessionId: "s-1", reportId: "rpt-1" },
      }),
    ).toBe("/generating?reportId=rpt-1&sessionId=s-1");

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
    ).toBe(
      "/report?errorCode=AI_PROVIDER_TIMEOUT&reportId=rpt-1&reportStatus=failed&sessionId=s-1",
    );

    expect(
      formatRouteUrl({
        name: "resume_versions",
        params: { resumeId: "v-1", tab: "rewrites", tailorRunId: "tr-1" },
      }),
    ).toBe("/resume-versions?resumeId=v-1&tab=rewrites&tailorRunId=tr-1");

    expect(
      formatRouteUrl({
        name: "debrief",
        params: {
          targetJobId: "tj-1",
          debriefId: "d-1",
          debriefJobId: "j-1",
        },
      }),
    ).toBe("/debrief?debriefId=d-1&debriefJobId=j-1&targetJobId=tj-1");
  });

  it("emits practice voice mode params under canonical path", () => {
    expect(
      formatRouteUrl({
        name: "practice",
        params: { mode: "voice", modality: "voice", sessionId: "s-1" },
      }),
    ).toBe("/practice?modality=voice&mode=voice&sessionId=s-1");
  });

  it("emits home import handoff params", () => {
    expect(
      formatRouteUrl({
        name: "home",
        params: { pendingImportId: "imp-1", source: "paste" },
      }),
    ).toBe("/?pendingImportId=imp-1&source=paste");
  });


  it("normalizes legacy route names back to retained routes", () => {
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
      "/auth/login?email=alice%40example.com&jdId=jd-1&next=%2Fworkspace&pendingLabel=%E7%AB%8B%E5%8D%B3%E9%9D%A2%E8%AF%95&pendingRoute=workspace&pendingType=start_practice&planId=plan-1&resumeId=rv-1&roundId=round-1&targetJobId=tj-1",
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

  it("preserves report replay params for autoStartPractice / sourceSessionId / replayItems / evidenceGaps / nextRoundId on workspace", () => {
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
    ).toBe(
      "/workspace?autoStartPractice=1&evidenceGaps=technical_depth%7Cnarrative&nextRoundId=round-2&planId=plan-1&replayItems=turn-1%2Cturn-3&resumeId=rv-1&sourceSessionId=s-prior&targetJobId=tj-1",
    );
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

  it("normalizes the retired /auth/reset path back to the login entry", () => {
    // product-scope D-16 — auth_reset is no longer a live route; the legacy
    // path must land on auth_login instead of materializing a reset screen.
    expect(parseUrlToRoute("/auth/reset")).toEqual({
      name: "auth_login",
      params: {},
    });
  });

  it("parses canonical workspace deep link", () => {
    expect(
      parseUrlToRoute(
        "/workspace?targetJobId=tj-1&resumeId=rv-1&planId=plan-1&autoStartPractice=1",
      ),
    ).toEqual({
      name: "workspace",
      params: {
        targetJobId: "tj-1",
        resumeId: "rv-1",
        planId: "plan-1",
        autoStartPractice: "1",
      },
    });
  });

  it("parses canonical report deep link with reportStatus + errorCode", () => {
    expect(
      parseUrlToRoute(
        "/report?sessionId=s-1&reportId=rpt-1&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT",
      ),
    ).toEqual({
      name: "report",
      params: {
        sessionId: "s-1",
        reportId: "rpt-1",
        reportStatus: "failed",
        errorCode: "AI_PROVIDER_TIMEOUT",
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
        planId: "plan-1",
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
        planId: "plan-1",
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

  it("normalizes home query-only deep link", () => {
    expect(parseUrlToRoute("/?pendingImportId=imp-1&source=paste")).toEqual({
      name: "home",
      params: { pendingImportId: "imp-1", source: "paste" },
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
  it("approves cross-owner safe params (handoff keys must survive)", () => {
    expect(isSafeRouteParam("home", "pendingImportId", {})).toBe(true);
    expect(isSafeRouteParam("workspace", "autoStartPractice", {})).toBe(true);
    expect(isSafeRouteParam("workspace", "sourceSessionId", {})).toBe(true);
    expect(isSafeRouteParam("workspace", "sourceReportId", {})).toBe(true);
    expect(isSafeRouteParam("workspace", "replayItems", {})).toBe(true);
    expect(isSafeRouteParam("workspace", "evidenceGaps", {})).toBe(true);
    expect(isSafeRouteParam("workspace", "nextRoundId", {})).toBe(true);
    expect(isSafeRouteParam("report", "reportStatus", {})).toBe(true);
    expect(isSafeRouteParam("report", "errorCode", {})).toBe(true);
    expect(isSafeRouteParam("resume_versions", "tailorRunId", {})).toBe(true);
    expect(isSafeRouteParam("debrief", "debriefJobId", {})).toBe(true);
    expect(isSafeRouteParam("parse", "resumeId", {})).toBe(true);
    expect(isSafeRouteParam("home", "resumeId", {})).toBe(true);
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
    // product-scope D-17: the retired jd_match -> parse reverse handoff
    // param is no longer on the parse allowlist.
    expect(isSafeRouteParam("parse", "sourceJobMatchId", {})).toBe(false);
  });

  it("expands auth_login allowlist with target route safe params when pendingRoute is present", () => {
    expect(
      isSafeRouteParam("auth_login", "planId", { pendingRoute: "workspace" }),
    ).toBe(true);
    expect(
      isSafeRouteParam("auth_login", "tailorRunId", {
        pendingRoute: "resume_versions",
      }),
    ).toBe(true);
    expect(
      isSafeRouteParam("auth_login", "debriefId", {
        pendingRoute: "debrief",
      }),
    ).toBe(true);
  });

  it("auth_login without pendingRoute keeps only base allowlist", () => {
    expect(isSafeRouteParam("auth_login", "planId", {})).toBe(false);
    expect(isSafeRouteParam("auth_login", "next", {})).toBe(true);
    expect(isSafeRouteParam("auth_login", "email", {})).toBe(true);
  });
});
