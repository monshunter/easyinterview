// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { buildResumePlainText } from "../adapters/resume";
import getResumeFixture from "../../../../../../openapi/fixtures/Resumes/getResume.json";
import type { Resume } from "../../../../api/generated/types";
import { ResumePreviewTab } from "./ResumePreviewTab";

const RESUME = getResumeFixture.scenarios.default.response.body as unknown as Resume;

function renderPreview(resume: Resume) {
  return render(
    <DisplayPreferencesProvider>
      <ResumePreviewTab
        resume={resume}
        onExport={vi.fn()}
        onCopy={vi.fn()}
        onViewOriginal={vi.fn()}
      />
    </DisplayPreferencesProvider>,
  );
}

describe("ResumePreviewTab content rendering and copy-to-clipboard", () => {
  it("renders the headline, summary, and skills derived from structuredProfile", () => {
    renderPreview(RESUME);
    const card = screen.getByTestId("resume-detail-preview-content");
    expect(card).toHaveTextContent(
      "Senior frontend engineer for platform-heavy product teams",
    );
    expect(card).toHaveTextContent(
      "Highlights reliability, release quality, and ownership evidence.",
    );
    expect(card).toHaveTextContent("React");
    expect(card).toHaveTextContent("TypeScript");
    expect(card).toHaveTextContent("Observability");
  });

  it("invokes onCopy when Copy text is clicked", async () => {
    const onCopy = vi.fn();
    render(
      <DisplayPreferencesProvider>
        <ResumePreviewTab
          resume={RESUME}
          onExport={vi.fn()}
          onCopy={onCopy}
          onViewOriginal={vi.fn()}
        />
      </DisplayPreferencesProvider>,
    );

    await userEvent.setup().click(screen.getByTestId("resume-detail-copy-text"));
    expect(onCopy).toHaveBeenCalledTimes(1);
  });

  it("buildResumePlainText projects headline, summary, and skills", () => {
    const text = buildResumePlainText(RESUME);
    expect(text).toContain(
      "Senior frontend engineer for platform-heavy product teams",
    );
    expect(text).toContain(
      "Highlights reliability, release quality, and ownership evidence.",
    );
    expect(text).toContain("React");
  });

  it("invokes onExport when Export PDF is clicked and onViewOriginal when View original is clicked", async () => {
    const onExport = vi.fn();
    const onViewOriginal = vi.fn();
    render(
      <DisplayPreferencesProvider>
        <ResumePreviewTab
          resume={RESUME}
          onExport={onExport}
          onCopy={vi.fn()}
          onViewOriginal={onViewOriginal}
        />
      </DisplayPreferencesProvider>,
    );

    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-detail-export-pdf"));
    await user.click(screen.getByTestId("resume-detail-view-original"));
    expect(onExport).toHaveBeenCalledTimes(1);
    expect(onViewOriginal).toHaveBeenCalledTimes(1);
  });
});
