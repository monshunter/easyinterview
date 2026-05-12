// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { buildResumePlainText } from "../adapters/resume";
import getResumeVersionFixture from "../../../../../../openapi/fixtures/Resumes/getResumeVersion.json";
import type { ResumeVersion } from "../../../../api/generated/types";
import { ResumePreviewTab } from "./ResumePreviewTab";

const TARGETED_VERSION =
  getResumeVersionFixture.scenarios.default.response.body as unknown as ResumeVersion;

function renderPreview(version: ResumeVersion) {
  return render(
    <DisplayPreferencesProvider>
      <ResumePreviewTab
        version={version}
        onExport={vi.fn()}
        onViewOriginal={vi.fn()}
      />
    </DisplayPreferencesProvider>,
  );
}

describe("ResumePreviewTab content rendering and copy-to-clipboard", () => {
  it("renders the headline, summary, sections, and skills derived from structuredProfile", () => {
    renderPreview(TARGETED_VERSION);
    const card = screen.getByTestId("resume-detail-preview-content");
    expect(card).toHaveTextContent("Senior frontend engineer");
    expect(card).toHaveTextContent("Highlights reliability");
    expect(card).toHaveTextContent("Selected evidence");
    expect(card).toHaveTextContent("Reduced checkout incident follow-ups");
    expect(card).toHaveTextContent("React");
    expect(card).toHaveTextContent("TypeScript");
  });

  it("fires a toast when Copy is clicked (copy success when clipboard exists, warn otherwise)", async () => {
    const toastCalls: { message: string; tone?: string }[] = [];
    (
      window as unknown as {
        eiToast?: (msg: string, opts?: { tone?: string }) => void;
      }
    ).eiToast = (message, opts) => {
      toastCalls.push({ message, tone: opts?.tone });
    };

    try {
      renderPreview(TARGETED_VERSION);
      const button = screen.getByTestId("resume-detail-copy-text");
      await userEvent.setup().click(button);

      await waitFor(() => {
        expect(toastCalls.length).toBeGreaterThan(0);
      });
      const last = toastCalls[toastCalls.length - 1]!;
      expect(last.message).toMatch(/复制|copy|copi|剪贴板|clipboard/i);
    } finally {
      delete (
        window as unknown as {
          eiToast?: (msg: string, opts?: { tone?: string }) => void;
        }
      ).eiToast;
    }
  });

  it("buildResumePlainText projects headline, summary, sections, and skills", () => {
    const text = buildResumePlainText(TARGETED_VERSION);
    expect(text).toContain("Senior frontend engineer");
    expect(text).toContain("Highlights reliability");
    expect(text).toContain("Selected evidence");
    expect(text).toContain("- Reduced checkout incident follow-ups");
    expect(text).toContain("React");
  });

  it("invokes onExport when Export PDF is clicked and onViewOriginal when View original is clicked", async () => {
    const onExport = vi.fn();
    const onViewOriginal = vi.fn();
    render(
      <DisplayPreferencesProvider>
        <ResumePreviewTab
          version={TARGETED_VERSION}
          onExport={onExport}
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
