import { describe, expect, it } from "vitest";

import type { ResumeAsset } from "../../../../../api/generated/types";
import {
  buildStructuredProfilePayload,
  mapParsedSummaryToStructuredProfileDraft,
} from "./mapParsedSummaryToStructuredProfileDraft";

const baseAsset: ResumeAsset = {
  id: "01918fa0-0000-7000-8000-000000001000",
  title: "alice.pdf",
  language: "zh",
  parseStatus: "ready",
  createdAt: "2026-05-17T00:00:00Z",
  updatedAt: "2026-05-17T00:00:00Z",
};

describe("mapParsedSummaryToStructuredProfileDraft", () => {
  it("maps identity, summary, experience, projects, skills, education from parsedSummary", () => {
    const draft = mapParsedSummaryToStructuredProfileDraft({
      ...baseAsset,
      parsedSummary: {
        identity: {
          name: "Alice",
          title: "Senior FE",
          location: "上海",
          contact: ["alice@example.com", "+86 138"],
        },
        summary: "Five years of platform work.",
        experience: [
          {
            co: "Star-Ring",
            role: "Senior FE",
            period: "2022 — now",
            bullets: ["Owned X"],
          },
        ],
        projects: [{ name: "RSC", note: "LCP 3.2→1.4" }],
        skills: ["React"],
        education: [{ school: "Tongji", degree: "BSc CS" }],
      },
    });
    expect(draft).toEqual({
      name: "Alice",
      title: "Senior FE",
      location: "上海",
      contact: ["alice@example.com", "+86 138"],
      summary: "Five years of platform work.",
      experience: [
        {
          co: "Star-Ring",
          role: "Senior FE",
          period: "2022 — now",
          bullets: ["Owned X"],
        },
      ],
      projects: [{ name: "RSC", note: "LCP 3.2→1.4" }],
      skills: ["React"],
      education: [{ school: "Tongji", degree: "BSc CS" }],
    });
  });

  it("falls back to asset.title when parsedSummary.identity.name is missing", () => {
    const draft = mapParsedSummaryToStructuredProfileDraft({
      ...baseAsset,
      title: "粘贴的简历",
      parsedSummary: { summary: "" },
    });
    expect(draft.name).toBe("粘贴的简历");
    expect(draft.experience).toEqual([]);
    expect(draft.skills).toEqual([]);
  });

  it("normalises legacy alias keys (company → co, title → role, dates → period, highlights → bullets)", () => {
    const draft = mapParsedSummaryToStructuredProfileDraft({
      ...baseAsset,
      parsedSummary: {
        experience: [
          {
            company: "Foo",
            title: "FE",
            dates: "2020 — 2022",
            highlights: ["A", "B"],
          },
        ],
      },
    });
    expect(draft.experience).toEqual([
      { co: "Foo", role: "FE", period: "2020 — 2022", bullets: ["A", "B"] },
    ]);
  });

  it("normalises projects.title / projects.impact aliases", () => {
    const draft = mapParsedSummaryToStructuredProfileDraft({
      ...baseAsset,
      parsedSummary: {
        projects: [{ title: "X", impact: "Y" }],
      },
    });
    expect(draft.projects).toEqual([{ name: "X", note: "Y" }]);
  });

  it("collects email + phone into contact list when contact field is missing", () => {
    const draft = mapParsedSummaryToStructuredProfileDraft({
      ...baseAsset,
      parsedSummary: {
        identity: { email: "a@b.c", phone: "+1" },
      },
    });
    expect(draft.contact).toEqual(["a@b.c", "+1"]);
  });

  it("renders empty arrays when parsedSummary is null/undefined", () => {
    const draft = mapParsedSummaryToStructuredProfileDraft({
      ...baseAsset,
      parsedSummary: null,
    });
    expect(draft.experience).toEqual([]);
    expect(draft.projects).toEqual([]);
    expect(draft.skills).toEqual([]);
    expect(draft.education).toEqual([]);
  });
});

describe("buildStructuredProfilePayload", () => {
  it("builds a confirmResumeStructuredMaster body shape from parsedSummary", () => {
    const payload = buildStructuredProfilePayload({
      ...baseAsset,
      parsedSummary: {
        identity: { name: "Alice", title: "Senior FE", location: "上海" },
        summary: "Owns the surface.",
        skills: ["React"],
      },
    });
    expect(payload).toMatchObject({
      headline: "Senior FE",
      summary: "Owns the surface.",
      identity: {
        name: "Alice",
        title: "Senior FE",
        location: "上海",
      },
      skills: ["React"],
      sections: [],
      provenance: {
        promptVersion: "resume_profile.v1",
        rubricVersion: "not_applicable",
        modelId: "resume-profile.confirmed.v1",
        language: "zh",
        featureFlag: "resume-workshop-additive",
        dataSourceVersion: "resume_asset.v1",
      },
    });
  });
});
