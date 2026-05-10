import { describe, expect, it } from "vitest";

import type { JobMatchRecommendation } from "../../../api/generated/types";

import {
  applySearchFilter,
  recommendationMatchesFilter,
} from "./searchFilters";

function rec(
  overrides: Partial<JobMatchRecommendation> = {},
): JobMatchRecommendation {
  return {
    id: "jm-x",
    title: "x",
    company: "x",
    companyTag: null,
    level: null,
    location: "Shanghai",
    comp: null,
    posted: "now",
    score: 80,
    fit: { must: 0, total: 0, plus: 0, totalPlus: 0 },
    reasons: [],
    risks: [],
    highlights: [],
    seen: true,
    saved: false,
    sourceUrl: null,
    sourceLabel: null,
    networkNote: null,
    similarInterviewers: null,
    interviewHypotheses: [],
    provenance: {
      promptVersion: "p.v1",
      rubricVersion: "r.v1",
      modelId: "m",
      language: "zh-CN",
      featureFlag: "none",
      dataSourceVersion: "v1",
    },
    ...overrides,
  };
}

describe("recommendationMatchesFilter (item 4.4)", () => {
  it("all keeps every recommendation", () => {
    const r = rec();
    expect(recommendationMatchesFilter(r, "all")).toBe(true);
  });

  it("strong keeps only score >= 85", () => {
    expect(recommendationMatchesFilter(rec({ score: 92 }), "strong")).toBe(true);
    expect(recommendationMatchesFilter(rec({ score: 85 }), "strong")).toBe(true);
    expect(recommendationMatchesFilter(rec({ score: 84 }), "strong")).toBe(false);
    expect(recommendationMatchesFilter(rec({ score: 60 }), "strong")).toBe(false);
  });

  it("remote matches 'Remote' (en, case-insensitive) and '远程' (zh)", () => {
    expect(
      recommendationMatchesFilter(
        rec({ location: "Remote · APAC" }),
        "remote",
      ),
    ).toBe(true);
    expect(
      recommendationMatchesFilter(rec({ location: "远程优先" }), "remote"),
    ).toBe(true);
    expect(
      recommendationMatchesFilter(rec({ location: "REMOTE" }), "remote"),
    ).toBe(true);
    expect(
      recommendationMatchesFilter(
        rec({ location: "Shanghai · Hybrid" }),
        "remote",
      ),
    ).toBe(false);
  });

  it("unseen keeps only items where seen=false", () => {
    expect(recommendationMatchesFilter(rec({ seen: false }), "unseen")).toBe(
      true,
    );
    expect(recommendationMatchesFilter(rec({ seen: true }), "unseen")).toBe(
      false,
    );
  });
});

describe("applySearchFilter (item 4.4)", () => {
  it("all returns the same array reference (no filtering work)", () => {
    const list = [rec({ id: "a" }), rec({ id: "b" })];
    expect(applySearchFilter(list, "all")).toBe(list);
  });

  it("strong filters down to score >= 85", () => {
    const list = [
      rec({ id: "a", score: 92 }),
      rec({ id: "b", score: 78 }),
      rec({ id: "c", score: 86 }),
    ];
    const got = applySearchFilter(list, "strong");
    expect(got.map((r) => r.id)).toEqual(["a", "c"]);
  });

  it("remote filters to Remote / 远程 only", () => {
    const list = [
      rec({ id: "a", location: "Remote" }),
      rec({ id: "b", location: "Beijing · On-site" }),
      rec({ id: "c", location: "远程优先" }),
    ];
    expect(applySearchFilter(list, "remote").map((r) => r.id)).toEqual([
      "a",
      "c",
    ]);
  });

  it("unseen filters to !seen", () => {
    const list = [
      rec({ id: "a", seen: false }),
      rec({ id: "b", seen: true }),
      rec({ id: "c", seen: false }),
    ];
    expect(applySearchFilter(list, "unseen").map((r) => r.id)).toEqual([
      "a",
      "c",
    ]);
  });
});
