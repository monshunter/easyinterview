import { describe, expect, it } from "vitest";

import type { Resume, ResumeSummary } from "../../../../api/generated/types";

import {
  buildResumeBodyLines,
  getResumeDetailRenderer,
  getResumeSourceUrl,
  mapResumeSummaryToUiSource,
  mapResumeToUiSource,
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

  it("falls back to LLM structured headline instead of source title when displayName is empty", () => {
    const ui = mapResumeToUiSource({ ...baseResume, displayName: "" });
    expect(ui.name).toBe("Senior frontend engineer");
    expect(ui.name).not.toBe(baseResume.title);
  });

  it("does not surface generic upload/paste names or derive a visible name from raw content", () => {
    const ui = mapResumeToUiSource({
      ...baseResume,
      sourceType: "paste",
      title: "粘贴的简历",
      displayName: "粘贴的简历",
      originalText:
        "张三 · 后端平台工程师\nFerry / reloadr / grayplan - GitOps CI/CD 与配置治理平台",
      parsedTextSnapshot: null,
      parsedSummary: { headline: "平台工程与 GitOps 简历" },
      structuredProfile: {},
    });
    expect(ui.name).toBe("平台工程与 GitOps 简历");
    expect(ui.name).not.toContain("张三");
    expect(ui.name).not.toBe("粘贴的简历");
    expect(ui.sourceName).toBe("粘贴文本");
  });

  it("does not use markdown first line or file name as the visible name", () => {
    const ui = mapResumeToUiSource({
      ...baseResume,
      sourceType: "upload",
      title: "谭章毓简历-后端工程师AI.pdf",
      displayName: "",
      originalText: null,
      parsedTextSnapshot:
        "# 谭章毓 | AI / Infra / DevOps 平台工程师\nservice-registry-operator / korder / ohmykube",
      parsedSummary: { headline: "AI Infra DevOps 平台工程师" },
      structuredProfile: {},
    });
    expect(ui.name).toBe("AI Infra DevOps 平台工程师");
    expect(ui.name).not.toContain("# 谭章毓");
    expect(ui.name).not.toContain(".pdf");
  });

  it("treats a displayName equal to the upload file title as pending, not a resume name", () => {
    const ui = mapResumeToUiSource({
      ...baseResume,
      sourceType: "upload",
      title: "谭章毓简历-后端工程师AI.pdf",
      displayName: "谭章毓简历-后端工程师AI.pdf",
      parseStatus: "failed",
      originalText: null,
      parsedTextSnapshot: null,
      parsedSummary: null,
      structuredProfile: {},
    });
    expect(ui.name).toBe("名称生成中");
    expect(ui.name).not.toContain(".pdf");
    expect(ui.sourceName).toBe("谭章毓简历-后端工程师AI.pdf");
  });

  it("uses a neutral pending name while parsing has no LLM-derived name", () => {
    const ui = mapResumeToUiSource({
      ...baseResume,
      parseStatus: "queued",
      title: "粘贴文本",
      displayName: "",
      originalText: null,
      parsedTextSnapshot: null,
      parsedSummary: null,
      structuredProfile: {},
    });
    expect(ui.name).toBe("名称生成中");
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
    expect(ui.text).toEqual([
      "Senior frontend engineer",
      "Owns complex product surfaces.",
      "Experience",
      "Bullet a.",
      "Bullet b.",
      "React · TypeScript",
    ]);
  });
});

describe("mapResumeSummaryToUiSource", () => {
  it("maps the list row exclusively from ResumeSummary fields", () => {
    const summary: ResumeSummary = {
      id: baseResume.id,
      title: baseResume.title,
      displayName: baseResume.displayName,
      language: baseResume.language,
      sourceType: "upload",
      parseStatus: "ready",
      summaryHeadline: "Senior frontend engineer focused on platform delivery",
      hasReadableContent: true,
      updatedAt: baseResume.updatedAt,
    };

    expect(mapResumeSummaryToUiSource(summary)).toEqual({
      id: summary.id,
      name: summary.displayName,
      langTag: "中",
      sourceName: summary.title,
      updatedAt: "2026-05-12",
      summary: summary.summaryHeadline,
    });
  });
});

describe("buildResumeBodyLines", () => {
  it("uses parsedTextSnapshot / originalText as the read-only resume body before structured fallback", () => {
    expect(buildResumeBodyLines(baseResume)).toEqual([
      "Senior frontend engineer.",
    ]);

    const originalOnly = {
      ...baseResume,
      parsedTextSnapshot: null,
      originalText: "Original line 1\nOriginal line 2",
    };
    expect(buildResumeBodyLines(originalOnly)).toEqual([
      "Original line 1",
      "Original line 2",
    ]);
  });
});

describe("resume source-format renderer selection", () => {
  it("uses the PDF engine for upload-backed PDF resumes", () => {
    const pdfResume = {
      ...baseResume,
      title: "alice-platform.pdf",
      sourceType: "upload" as const,
      fileObjectId: "01918fa0-0000-7000-8000-000000001100",
    };

    expect(getResumeDetailRenderer(pdfResume)).toBe("pdf");
    expect(getResumeSourceUrl(pdfResume, "/api/v1")).toBe(
      "/api/v1/resumes/01918fa0-0000-7000-8000-000000001000/source",
    );
  });

  it("keeps paste, Markdown, and text sources on the Markdown engine", () => {
    expect(getResumeDetailRenderer({ ...baseResume, sourceType: "paste" })).toBe(
      "markdown",
    );
    expect(getResumeDetailRenderer({ ...baseResume, title: "alice.md" })).toBe(
      "markdown",
    );
    expect(getResumeDetailRenderer({ ...baseResume, title: "alice.txt" })).toBe(
      "markdown",
    );
  });
});
