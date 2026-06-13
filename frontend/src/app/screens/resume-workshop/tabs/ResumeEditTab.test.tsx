// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type { Resume } from "../../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { ResumeEditTab } from "./ResumeEditTab";

const baseResume: Resume = {
  id: "01918fa0-0000-7000-8000-000000001000",
  title: "Alice Example — Senior Frontend Engineer",
  displayName: "Alice Example — Senior Frontend Engineer",
  language: "zh-CN",
  parseStatus: "ready",
  status: "active",
  sourceType: "upload",
  fileObjectId: "01918fa0-0000-7000-8000-000000001100",
  structuredProfile: {
    headline: "Senior frontend engineer",
    summary: "Owns platform delivery end to end.",
  },
  createdAt: "2026-05-12T08:20:00Z",
  updatedAt: "2026-05-12T08:20:00Z",
  deletedAt: null,
};

const renderEdit = (
  resume: Resume,
  props: Partial<Parameters<typeof ResumeEditTab>[0]> = {},
) =>
  render(
    <DisplayPreferencesProvider>
      <ResumeEditTab resume={resume} {...props} />
    </DisplayPreferencesProvider>,
  );

describe("ResumeEditTab render baseline (D-20 flat model)", () => {
  it("renders the scope banner + displayName + headline + summary + Experience/Skills placeholders + save button (>= 12 testid)", () => {
    renderEdit(baseResume);
    for (const id of [
      "resume-edit-tab",
      "resume-edit-scope-banner",
      "resume-edit-scope-banner-message",
      "resume-edit-display-name-input",
      "resume-edit-headline-input",
      "resume-edit-summary-textarea",
      "resume-edit-section-experience",
      "resume-edit-section-experience-add",
      "resume-edit-section-experience-placeholder",
      "resume-edit-section-skills",
      "resume-edit-section-skills-add",
      "resume-edit-section-skills-placeholder",
      "resume-edit-save-button",
    ]) {
      expect(screen.getByTestId(id)).toBeInTheDocument();
    }
  });

  it("interpolates the resume displayName into the scope banner", () => {
    renderEdit(baseResume);
    const banner = screen.getByTestId("resume-edit-scope-banner");
    expect(banner.textContent).toContain(
      "Alice Example — Senior Frontend Engineer",
    );
  });

  it("pre-fills displayName / headline / summary and disables save while clean", () => {
    renderEdit(baseResume);
    expect(
      (
        screen.getByTestId("resume-edit-display-name-input") as HTMLInputElement
      ).value,
    ).toBe("Alice Example — Senior Frontend Engineer");
    expect(
      (screen.getByTestId("resume-edit-headline-input") as HTMLInputElement)
        .value,
    ).toBe("Senior frontend engineer");
    expect(
      (
        screen.getByTestId("resume-edit-summary-textarea") as HTMLTextAreaElement
      ).value,
    ).toBe("Owns platform delivery end to end.");
    const save = screen.getByTestId("resume-edit-save-button");
    expect(save).toBeDisabled();
    expect(screen.getByTestId("resume-edit-tab")).toHaveAttribute(
      "data-edit-dirty",
      "false",
    );
  });
});

describe("ResumeEditTab save behaviour (D-20 flat model)", () => {
  it("becomes dirty + enables save when the user edits headline; invokes onSave with displayName + headline + summary", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    renderEdit(baseResume, { onSave });
    const user = userEvent.setup();
    const headline = screen.getByTestId(
      "resume-edit-headline-input",
    ) as HTMLInputElement;
    await user.clear(headline);
    await user.type(headline, "New headline draft");
    expect(screen.getByTestId("resume-edit-tab")).toHaveAttribute(
      "data-edit-dirty",
      "true",
    );
    const save = screen.getByTestId("resume-edit-save-button");
    expect(save).toBeEnabled();
    await user.click(save);
    expect(onSave).toHaveBeenCalledWith({
      displayName: "Alice Example — Senior Frontend Engineer",
      headline: "New headline draft",
      summary: "Owns platform delivery end to end.",
    });
  });

  it("becomes dirty when the user edits the displayName and forwards the new name to onSave", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    renderEdit(baseResume, { onSave });
    const user = userEvent.setup();
    const displayName = screen.getByTestId(
      "resume-edit-display-name-input",
    ) as HTMLInputElement;
    await user.clear(displayName);
    await user.type(displayName, "Renamed resume");
    expect(screen.getByTestId("resume-edit-tab")).toHaveAttribute(
      "data-edit-dirty",
      "true",
    );
    await user.click(screen.getByTestId("resume-edit-save-button"));
    expect(onSave).toHaveBeenCalledWith({
      displayName: "Renamed resume",
      headline: "Senior frontend engineer",
      summary: "Owns platform delivery end to end.",
    });
  });

  it("disables save when saving=true and shows the localized saving copy", () => {
    renderEdit(baseResume, { saving: true });
    const save = screen.getByTestId("resume-edit-save-button");
    expect(save).toBeDisabled();
    expect(save.textContent).toMatch(/Saving|保存中/);
    expect(screen.getByTestId("resume-edit-tab")).toHaveAttribute(
      "data-edit-saving",
      "true",
    );
  });

  it("renders the in-form error alert when errorMessage is provided", () => {
    renderEdit(baseResume, { errorMessage: "Validation failed. Please retry." });
    const alert = screen.getByTestId("resume-edit-error");
    expect(alert).toBeInTheDocument();
    expect(alert).toHaveAttribute("role", "alert");
    expect(alert.textContent).toBe("Validation failed. Please retry.");
  });

  it("Experience / Skills Add buttons fire a 'coming soon' toast and do not call onSave", async () => {
    const eiToast = vi.fn();
    (window as unknown as { eiToast?: typeof eiToast }).eiToast = eiToast;
    const onSave = vi.fn();
    renderEdit(baseResume, { onSave });
    const user = userEvent.setup();
    await user.click(
      screen.getByTestId("resume-edit-section-experience-add"),
    );
    await user.click(screen.getByTestId("resume-edit-section-skills-add"));
    expect(eiToast).toHaveBeenCalledTimes(2);
    expect(onSave).not.toHaveBeenCalled();
    delete (window as unknown as { eiToast?: typeof eiToast }).eiToast;
  });
});

describe("ResumeEditTab privacy (D-20 flat model)", () => {
  it("does not append displayName / headline / summary text to URL or localStorage on render or typing", async () => {
    const sensitiveHeadline = "Confidential headline 12345";
    const sensitiveSummary = "Sensitive personal summary body";
    const setItemSpy = vi.spyOn(window.localStorage, "setItem");
    const replaceState = vi.spyOn(window.history, "replaceState");
    const resume: Resume = {
      ...baseResume,
      structuredProfile: {
        headline: sensitiveHeadline,
        summary: sensitiveSummary,
      },
    };
    renderEdit(resume);

    const user = userEvent.setup();
    await user.type(
      screen.getByTestId("resume-edit-headline-input"),
      "appended",
    );

    for (const call of setItemSpy.mock.calls) {
      expect(call[1]).not.toContain(sensitiveHeadline);
      expect(call[1]).not.toContain(sensitiveSummary);
    }
    for (const call of replaceState.mock.calls) {
      const url = call[2];
      if (typeof url === "string") {
        expect(url).not.toContain(sensitiveHeadline);
        expect(url).not.toContain(sensitiveSummary);
      }
    }
    expect(window.location.href).not.toContain(sensitiveHeadline);

    setItemSpy.mockRestore();
    replaceState.mockRestore();
  });
});
