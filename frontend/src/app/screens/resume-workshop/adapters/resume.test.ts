import { describe, expect, it } from "vitest";

import type { Resume } from "../../../../api/generated/types";

import {
  buildResumePlainText,
  buildResumePreview,
  mapBulletSuggestionToUi,
  mapResumeToUiSource,
  type ResumeSuggestionInput,
} from "./resume";

const baseResume: Resume = {
  id: "01918fa0-0000-7000-8000-000000001000",
  title: "Alice Example - Senior Frontend Engineer",
  displayName: "Alice Example - Senior Frontend Engineer",
  language: "zh-CN",
  parseStatus: "ready",
  status: "active",
  sourceType: "upload",
  fileObjectId: "01918fa0-0000-7000-8000-000000001100",
  originalText: null,
  parsedTextSnapshot: "Senior frontend engineer.",
  parsedSummary: { headline: "Senior frontend engineer focused on platform delivery" },
  structuredProfile: {
    headline: "Senior frontend engineer",
    summary: "Owns complex product surfaces.",
    skills: ["React", "TypeScript"],
    sections: [{ title: "Experience", bullets: ["Bullet a.", "Bullet b."] }],
  },
  createdAt: "2026-04-22T09:30:00Z",
  updatedAt: "2026-05-12T08:00:00Z",
  deletedAt: null,
};

describe("mapResumeToUiSource", () => {
  it("maps the basic resume fields to the UI ResumeSource shape", () => {
    const ui = mapResumeToUiSource(baseResume);
    expect(ui.id).toBe(baseResume.id);
    expect(ui.name).toBe(baseResume.displayName);
    expect(ui.createdAt).toBe("2026-04-22");
    expect(ui.status).toBe("active");
    expect(ui.summary).toBe("Senior frontend engineer focused on platform delivery");
    expect(ui.text).toContain("Senior frontend engineer.");
  });

  it("falls back to title when displayName is empty", () => {
    const ui = mapResumeToUiSource({ ...baseResume, displayName: "" });
    expect(ui.name).toBe(baseResume.title);
  });

  it("derives langTag from BCP-47 language tag (zh-CN → 中, en-US → EN)", () => {
    expect(mapResumeToUiSource({ ...baseResume, language: "zh-CN" }).langTag).toBe(
      "中",
    );
    expect(mapResumeToUiSource({ ...baseResume, language: "en-US" }).langTag).toBe(
      "EN",
    );
    expect(mapResumeToUiSource({ ...baseResume, language: "fr" }).langTag).toBe(
      "FR",
    );
  });

  it("propagates archived status", () => {
    const archived = { ...baseResume, status: "archived" as const };
    expect(mapResumeToUiSource(archived).status).toBe("archived");
  });

  it("handles null parsedSummary / parsedTextSnapshot without throwing", () => {
    const sparse: Resume = {
      ...baseResume,
      parsedSummary: null,
      parsedTextSnapshot: null,
      originalText: null,
    };
    const ui = mapResumeToUiSource(sparse);
    expect(ui.summary).toBe("");
    expect(ui.text).toEqual([]);
  });
});

describe("buildResumePreview / buildResumePlainText", () => {
  it("projects the structured profile into the preview shape", () => {
    const preview = buildResumePreview(baseResume);
    expect(preview.headline).toBe("Senior frontend engineer");
    expect(preview.summary).toBe("Owns complex product surfaces.");
    expect(preview.skills).toEqual(["React", "TypeScript"]);
    expect(preview.sections).toEqual([
      { title: "Experience", bullets: ["Bullet a.", "Bullet b."] },
    ]);
  });

  it("renders a plain-text projection suitable for clipboard copy", () => {
    const text = buildResumePlainText(baseResume);
    expect(text).toContain("Senior frontend engineer");
    expect(text).toContain("- Bullet a.");
    expect(text).toContain("React · TypeScript");
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
      section: "Experience · Star-Ring",
    };
    const bullet = mapBulletSuggestionToUi(input);
    expect(bullet.id).toBe("sug-1");
    expect(bullet.section).toBe("Experience · Star-Ring");
    expect(bullet.original).toBe(input.originalBullet);
    expect(bullet.rewritten).toBe(input.suggestedBullet);
    expect(bullet.why.length).toBeGreaterThanOrEqual(1);
    // D-20: suggestions are accept-only and client-side ephemeral; the UI
    // bullet starts pending with no decision provenance fields.
    expect(bullet.status).toBe("pending");
  });

  it("defaults section to an empty string when not provided", () => {
    const bullet = mapBulletSuggestionToUi({
      id: "sug-3",
      originalBullet: "o",
      suggestedBullet: "s",
      reason: "r",
    });
    expect(bullet.section).toBe("");
    expect(bullet.status).toBe("pending");
  });
});
