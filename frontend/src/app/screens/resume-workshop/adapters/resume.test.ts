import { describe, expect, it } from "vitest";

import type {
  ResumeAsset,
  ResumeVersion,
} from "../../../../api/generated/types";

import {
  mapBulletSuggestionToUi,
  mapResumeAssetToUiSource,
  mapResumeVersionToUi,
  type ResumeSuggestionInput,
} from "./resume";

const baseAsset: ResumeAsset = {
  id: "01918fa0-0000-7000-8000-000000001000",
  title: "Alice Example - Senior Frontend Engineer",
  language: "zh-CN",
  parseStatus: "ready",
  status: "active",
  sourceType: "upload",
  fileObjectId: "01918fa0-0000-7000-8000-000000001100",
  originalText: null,
  guidedAnswers: {},
  parsedTextSnapshot: "Senior frontend engineer.",
  parsedSummary: { headline: "Senior frontend engineer focused on platform delivery" },
  createdAt: "2026-04-22T09:30:00Z",
  updatedAt: "2026-05-12T08:00:00Z",
  deletedAt: null,
};

const baseMaster: ResumeVersion = {
  id: "0195f2d0-0001-7000-8000-000000000201",
  resumeAssetId: "01918fa0-0000-7000-8000-000000001000",
  parentVersionId: null,
  versionType: "structured_master",
  targetJobId: null,
  displayName: "Structured master",
  seedStrategy: null,
  focusAngle: null,
  structuredProfile: {
    headline: "Senior frontend engineer",
    summary: "Owns complex product surfaces.",
    sections: [
      { title: "Experience", bullets: ["Bullet a.", "Bullet b."] },
    ],
  },
  matchScore: null,
  promptVersion: "resume_profile.v1",
  rubricVersion: "not_applicable",
  modelId: "fixture-model:resume-version-profile",
  provider: "fixture-provider",
  provenance: {
    promptVersion: "resume_profile.v1",
    rubricVersion: "not_applicable",
    modelId: "fixture-model:resume-version-profile",
    language: "zh-CN",
    featureFlag: "resume-workshop-additive",
    dataSourceVersion: "resume_asset.v1",
  },
  suggestions: [],
  createdAt: "2026-05-12T08:20:00Z",
  updatedAt: "2026-05-12T08:20:00Z",
  deletedAt: null,
};

const baseTargeted: ResumeVersion = {
  ...baseMaster,
  id: "0195f2d0-0001-7000-8000-000000000202",
  parentVersionId: "0195f2d0-0001-7000-8000-000000000201",
  versionType: "targeted",
  targetJobId: "01918fa0-0000-7000-8000-000000002000",
  displayName: "Northstar Systems frontend target",
  seedStrategy: "copy_master",
  focusAngle: "Evidence for platform ownership and pragmatic delivery",
  matchScore: 0.84,
  suggestions: [
    {
      id: "sug-1",
      originalBullet: "Improved checkout reliability.",
      suggestedBullet:
        "Reduced checkout incident follow-ups by tightening release guardrails.",
      reason: "Adds measurable ownership evidence.",
      status: "accepted",
    },
  ],
  createdAt: "2026-05-12T08:24:00Z",
  updatedAt: "2026-05-12T08:24:00Z",
};

describe("mapResumeAssetToUiSource", () => {
  it("maps the basic asset fields to the UI ResumeSource shape", () => {
    const ui = mapResumeAssetToUiSource(baseAsset);
    expect(ui.id).toBe(baseAsset.id);
    expect(ui.name).toBe(baseAsset.title);
    expect(ui.createdAt).toBe("2026-04-22");
    expect(ui.status).toBe("active");
    expect(ui.summary).toBe("Senior frontend engineer focused on platform delivery");
    expect(ui.text).toContain("Senior frontend engineer.");
  });

  it("derives langTag from BCP-47 language tag (zh-CN → 中, en-US → EN)", () => {
    expect(mapResumeAssetToUiSource({ ...baseAsset, language: "zh-CN" }).langTag).toBe(
      "中",
    );
    expect(mapResumeAssetToUiSource({ ...baseAsset, language: "en-US" }).langTag).toBe(
      "EN",
    );
    expect(mapResumeAssetToUiSource({ ...baseAsset, language: "fr" }).langTag).toBe(
      "FR",
    );
  });

  it("propagates archived status", () => {
    const archived = { ...baseAsset, status: "archived" as const };
    expect(mapResumeAssetToUiSource(archived).status).toBe("archived");
  });

  it("handles null parsedSummary / parsedTextSnapshot without throwing", () => {
    const sparse: ResumeAsset = {
      ...baseAsset,
      parsedSummary: null,
      parsedTextSnapshot: null,
      originalText: null,
    };
    const ui = mapResumeAssetToUiSource(sparse);
    expect(ui.summary).toBe("");
    expect(ui.text).toEqual([]);
  });
});

describe("mapResumeVersionToUi", () => {
  it("maps a structured_master version to MASTER tag with no parent chain", () => {
    const ui = mapResumeVersionToUi(baseMaster);
    expect(ui.id).toBe(baseMaster.id);
    expect(ui.originalId).toBe(baseMaster.resumeAssetId);
    expect(ui.tag).toBe("MASTER");
    expect(ui.parentVersionId).toBeNull();
    expect(ui.match).toBeNull();
    expect(ui.archived).toBe(false);
  });

  it("maps a targeted version to TARGETED tag and preserves the parent chain", () => {
    const ui = mapResumeVersionToUi(baseTargeted);
    expect(ui.tag).toBe("TARGETED");
    expect(ui.parentVersionId).toBe(baseMaster.id);
    expect(ui.target).toBe(
      "Evidence for platform ownership and pragmatic delivery",
    );
  });

  it("formats matchScore from 0..1 to a 0..100 integer percentage", () => {
    expect(mapResumeVersionToUi(baseTargeted).match).toBe(84);
    expect(
      mapResumeVersionToUi({ ...baseTargeted, matchScore: 0.715 }).match,
    ).toBe(72);
  });

  it("counts bullets and accepted suggestions from the suggestions array", () => {
    const withSuggestions: ResumeVersion = {
      ...baseTargeted,
      suggestions: [
        {
          id: "s-1",
          originalBullet: "a",
          suggestedBullet: "b",
          reason: "r",
          status: "accepted",
        },
        {
          id: "s-2",
          originalBullet: "c",
          suggestedBullet: "d",
          reason: "r",
          status: "pending",
        },
      ],
    };
    const ui = mapResumeVersionToUi(withSuggestions);
    expect(ui.accepted).toBe(1);
    expect(ui.bullets).toBeGreaterThanOrEqual(2);
  });

  it("marks deletedAt versions as archived", () => {
    const deleted: ResumeVersion = {
      ...baseMaster,
      deletedAt: "2026-05-13T00:00:00Z",
    };
    expect(mapResumeVersionToUi(deleted).archived).toBe(true);
  });
});

describe("mapBulletSuggestionToUi", () => {
  it("maps the suggestion to the UI bullet shape with the why list split", () => {
    const input: ResumeSuggestionInput = {
      id: "sug-1",
      originalBullet: "Improved checkout reliability.",
      suggestedBullet:
        "Reduced checkout incident follow-ups by tightening release guardrails.",
      reason:
        "Adds measurable ownership evidence. | Names the affected surface.",
      status: "pending",
      section: "Experience · Star-Ring",
    };
    const bullet = mapBulletSuggestionToUi(input);
    expect(bullet.id).toBe("sug-1");
    expect(bullet.section).toBe("Experience · Star-Ring");
    expect(bullet.original).toBe(input.originalBullet);
    expect(bullet.rewritten).toBe(input.suggestedBullet);
    expect(bullet.why.length).toBeGreaterThanOrEqual(1);
    expect(bullet.status).toBe("pending");
  });
});
