// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

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
