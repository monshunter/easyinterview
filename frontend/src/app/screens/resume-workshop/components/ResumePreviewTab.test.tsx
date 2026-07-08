// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { render, screen, within } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import getResumeFixture from "../../../../../../openapi/fixtures/Resumes/getResume.json";
import type { Resume } from "../../../../api/generated/types";
import { ResumePreviewTab } from "./ResumePreviewTab";

const RESUME = getResumeFixture.scenarios.default.response.body as unknown as Resume;

function renderPreview(resume: Resume) {
  return render(
    <DisplayPreferencesProvider>
      <ResumePreviewTab resume={resume} />
    </DisplayPreferencesProvider>,
  );
}

describe("ResumePreviewTab read-only rendering", () => {
  it("renders the original parsed text snapshot as the resume body", () => {
    renderPreview(RESUME);
    const content = screen.getByTestId("resume-detail-preview-content");
    expect(content).toHaveTextContent("Original resume parsed text snapshot");
    expect(content).toHaveTextContent(
      "Led platform release guardrail work across frontend surfaces",
    );
    expect(content).not.toHaveTextContent(
      "Senior frontend engineer for platform-heavy product teams",
    );
  });

  it("renders parsedTextSnapshot Markdown as headings, lists, and paragraphs", () => {
    renderPreview({
      ...RESUME,
      parsedTextSnapshot:
        "# 谭章毓\n\n## 工作经历\n- **EasyInterview** 平台工程\n- [GitHub](https://github.com/monshunter)\n\n负责端到端训练工作台。",
      originalText: "plain fallback should not render",
    });

    const content = screen.getByTestId("resume-detail-preview-content");
    expect(
      within(content).getByRole("heading", { level: 1, name: "谭章毓" }),
    ).toBeInTheDocument();
    expect(
      within(content).getByRole("heading", { level: 2, name: "工作经历" }),
    ).toBeInTheDocument();
    expect(within(content).getAllByRole("listitem")).toHaveLength(2);
    expect(within(content).getByText("EasyInterview").tagName).toBe("STRONG");
    expect(
      within(content).getByRole("link", { name: "GitHub" }),
    ).toHaveAttribute("href", "https://github.com/monshunter");
    expect(content).toHaveTextContent("负责端到端训练工作台。");
    expect(content).not.toHaveTextContent("plain fallback should not render");
    expect(content).not.toHaveTextContent("**EasyInterview**");
    expect(content).not.toHaveTextContent("[GitHub]");
  });

  it("does not inject the display name or header metadata into Markdown body", () => {
    renderPreview({
      ...RESUME,
      displayName: "Injected Display Name Must Stay In Header",
      title: "source-markdown.md",
      sourceType: "upload",
      parsedTextSnapshot: "# Candidate Body\n\nActual body only.",
      originalText: "plain fallback should not render",
    });

    const content = screen.getByTestId("resume-detail-preview-content");
    const page = screen.getByTestId("resume-detail-markdown-page");
    expect(
      within(page).getByRole("heading", { level: 1, name: "Candidate Body" }),
    ).toBeInTheDocument();
    expect(page).toHaveTextContent("Actual body only.");
    expect(content).not.toHaveTextContent(
      "Injected Display Name Must Stay In Header",
    );
    expect(content.querySelector(".ei-text-title")).toBeNull();
  });

  it("renders Markdown inside the same reading surface model as source pages", () => {
    renderPreview({
      ...RESUME,
      sourceType: "paste",
      parsedTextSnapshot: "# Markdown page\n\n- One",
      originalText: "plain fallback should not render",
    });

    const card = screen
      .getByTestId("resume-detail-preview-content")
      .querySelector(".ei-resume-detail-preview-card");
    expect(card).not.toBeNull();
    expect(card).not.toHaveClass("ei-resume-detail-preview-card--pdf");
    expect(screen.getByTestId("resume-detail-markdown-page")).toHaveClass(
      "ei-resume-detail-markdown-page",
    );
  });

  it("renders upload-backed PDF resumes as a top-to-bottom page stack without native viewer chrome", () => {
    renderPreview({
      ...RESUME,
      sourceType: "upload",
      title: "alice-platform.pdf",
      parsedTextSnapshot: "# Markdown fallback should not render for PDF",
      originalText: "plain fallback should not render for PDF",
    });

    const stack = screen.getByTestId("resume-detail-pdf-preview-stack");
    const card = screen
      .getByTestId("resume-detail-preview-content")
      .querySelector(".ei-resume-detail-preview-card");
    expect(card).not.toHaveClass("ei-resume-detail-preview-card--pdf");
    expect(stack).toHaveAttribute(
      "data-source-url",
      "/api/v1/resumes/01918fa0-0000-7000-8000-000000001000/source",
    );
    expect(document.querySelector("object, iframe, embed")).toBeNull();
    const content = screen.getByTestId("resume-detail-preview-content");
    expect(content).not.toHaveTextContent("Markdown fallback should not render");
    expect(content).not.toHaveTextContent("plain fallback should not render");
  });

  it("does not render secondary actions around the resume body", () => {
    renderPreview(RESUME);
    for (const forbidden of [
      "resume-detail-export-pdf",
      "resume-detail-copy-text",
      "resume-detail-view-original",
    ]) {
      expect(screen.queryByTestId(forbidden)).not.toBeInTheDocument();
    }
  });
});
