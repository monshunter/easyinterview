// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type { ResumeVersion } from "../../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { ResumeEditTab } from "./ResumeEditTab";

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
    summary: "Owns platform delivery end to end.",
  },
  matchScore: null,
  promptVersion: "p",
  rubricVersion: "r",
  modelId: "m",
  provider: "fixture",
  provenance: {
    promptVersion: "p",
    rubricVersion: "r",
    modelId: "m",
    language: "zh-CN",
    featureFlag: "f",
    dataSourceVersion: "d",
  },
  suggestions: [],
  createdAt: "2026-05-12T08:20:00Z",
  updatedAt: "2026-05-12T08:20:00Z",
  deletedAt: null,
};

const baseTargeted: ResumeVersion = {
  ...baseMaster,
  id: "0195f2d0-0001-7000-8000-000000000202",
  parentVersionId: baseMaster.id,
  versionType: "targeted",
  targetJobId: "01918fa0-0000-7000-8000-000000002000",
  displayName: "Northstar Systems frontend target",
  seedStrategy: "copy_master",
  focusAngle: "platform",
  structuredProfile: {
    headline: "Targeted headline",
    summary: "Targeted summary.",
  },
};

const renderEdit = (
  version: ResumeVersion,
  props: Partial<Parameters<typeof ResumeEditTab>[0]> = {},
) =>
  render(
    <DisplayPreferencesProvider>
      <ResumeEditTab version={version} {...props} />
    </DisplayPreferencesProvider>,
  );

describe("ResumeEditTab render baseline (plan 003 Phase 6.1)", () => {
  it("renders the master scope banner + headline + summary + Experience/Skills placeholders + save button (>= 10 testid)", () => {
    renderEdit(baseMaster);
    for (const id of [
      "resume-edit-tab",
      "resume-edit-scope-banner",
      "resume-edit-scope-banner-message",
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
    expect(
      screen.getByTestId("resume-edit-scope-banner"),
    ).toHaveAttribute("data-scope", "master");
  });

  it("renders the targeted scope banner with the version name interpolated", () => {
    renderEdit(baseTargeted);
    const banner = screen.getByTestId("resume-edit-scope-banner");
    expect(banner).toHaveAttribute("data-scope", "targeted");
    expect(banner.textContent).toContain("Northstar Systems frontend target");
  });

  it("pre-fills headline / summary from structuredProfile and disables save while clean", () => {
    renderEdit(baseTargeted);
    expect(
      (screen.getByTestId("resume-edit-headline-input") as HTMLInputElement)
        .value,
    ).toBe("Targeted headline");
    expect(
      (
        screen.getByTestId("resume-edit-summary-textarea") as HTMLTextAreaElement
      ).value,
    ).toBe("Targeted summary.");
    const save = screen.getByTestId("resume-edit-save-button");
    expect(save).toBeDisabled();
    expect(screen.getByTestId("resume-edit-tab")).toHaveAttribute(
      "data-edit-dirty",
      "false",
    );
  });
});

describe("ResumeEditTab save behaviour (plan 003 Phase 6.2-6.4)", () => {
  it("becomes dirty + enables save when the user edits headline; invokes onSave with current draft", async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    renderEdit(baseTargeted, { onSave });
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
      headline: "New headline draft",
      summary: "Targeted summary.",
    });
  });

  it("disables save when saving=true and shows the localized saving copy", () => {
    renderEdit(baseTargeted, { saving: true });
    const save = screen.getByTestId("resume-edit-save-button");
    expect(save).toBeDisabled();
    expect(save.textContent).toMatch(/Saving|保存中/);
  });

  it("renders the in-form error alert when errorMessage is provided", () => {
    renderEdit(baseTargeted, { errorMessage: "Validation failed. Please retry." });
    const alert = screen.getByTestId("resume-edit-error");
    expect(alert).toBeInTheDocument();
    expect(alert).toHaveAttribute("role", "alert");
    expect(alert.textContent).toBe("Validation failed. Please retry.");
  });

  it("Experience / Skills Add buttons fire a 'coming soon' toast and do not call onSave", async () => {
    const eiToast = vi.fn();
    (window as unknown as { eiToast?: typeof eiToast }).eiToast = eiToast;
    const onSave = vi.fn();
    renderEdit(baseTargeted, { onSave });
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

describe("ResumeEditTab privacy (plan 003 Phase 6.6)", () => {
  it("does not append headline / summary text to URL or localStorage on render or typing", async () => {
    const sensitiveHeadline = "Confidential headline 12345";
    const sensitiveSummary = "Sensitive personal summary body";
    const setItemSpy = vi.spyOn(window.localStorage, "setItem");
    const replaceState = vi.spyOn(window.history, "replaceState");
    const version: ResumeVersion = {
      ...baseTargeted,
      structuredProfile: {
        headline: sensitiveHeadline,
        summary: sensitiveSummary,
      },
    };
    renderEdit(version);

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
