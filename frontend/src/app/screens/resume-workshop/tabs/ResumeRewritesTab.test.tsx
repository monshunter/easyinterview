// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type {
  Resume,
  ResumeTailorBulletSuggestion,
} from "../../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { ResumeRewritesTab } from "./ResumeRewritesTab";

const baseResume: Resume = {
  id: "01918fa0-0000-7000-8000-000000001000",
  title: "Alice Example — Senior Frontend Engineer",
  displayName: "Alice Example — Senior Frontend Engineer",
  language: "zh-CN",
  parseStatus: "ready",
  status: "active",
  sourceType: "upload",
  fileObjectId: "01918fa0-0000-7000-8000-000000001100",
  structuredProfile: {},
  createdAt: "2026-05-12T08:24:00Z",
  updatedAt: "2026-05-12T08:24:00Z",
  deletedAt: null,
};

// D-20: suggestions are ephemeral and accept-only. Each maps to a UI bullet
// keyed by index (`bullet-${index}`); there is no server-side status / reject.
const buildSuggestions = (
  rows: Array<{ rewritten?: string; original?: string; reason?: string }>,
): ResumeTailorBulletSuggestion[] =>
  rows.map((row, index) => ({
    originalBullet: row.original ?? `Original for bullet ${index}`,
    suggestedBullet: row.rewritten ?? `Rewritten for bullet ${index}`,
    reason: row.reason ?? "Adds quantified ownership. | Clarifies the surface.",
  }));

const renderRewrites = (
  suggestions: ResumeTailorBulletSuggestion[],
  props: Partial<Parameters<typeof ResumeRewritesTab>[0]> = {},
) =>
  render(
    <DisplayPreferencesProvider>
      <ResumeRewritesTab
        resume={baseResume}
        suggestions={suggestions}
        {...props}
      />
    </DisplayPreferencesProvider>,
  );

describe("ResumeRewritesTab render baseline", () => {
  it("renders the scope banner and counts with the resume id on the tab", () => {
    renderRewrites(buildSuggestions([{}, {}, {}, {}]));
    const banner = screen.getByTestId("resume-rewrites-scope-banner");
    expect(banner).toBeInTheDocument();
    expect(banner).toHaveAttribute("role", "status");
    expect(banner).toHaveAttribute("aria-live", "polite");
    expect(screen.getByTestId("resume-rewrites-counts")).toBeInTheDocument();
    const tab = screen.getByTestId("resume-rewrites-tab");
    expect(tab).toHaveAttribute("data-resume-id", baseResume.id);
    expect(tab).toHaveAttribute("data-bullet-count", "4");
    expect(tab).toHaveAttribute("data-accepted-count", "0");
  });

  it("renders the bullet list with pending status chips and pre-selects the first bullet", () => {
    renderRewrites(
      buildSuggestions([
        { rewritten: "Pending bullet detail" },
        { rewritten: "Second bullet detail" },
      ]),
    );
    expect(
      screen.getByTestId("resume-rewrites-bullet-list"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-rewrites-bullet-row-bullet-0"),
    ).toHaveAttribute("aria-selected", "true");
    expect(
      screen.getByTestId("resume-rewrites-bullet-row-bullet-1"),
    ).toHaveAttribute("aria-selected", "false");
    expect(
      screen.getByTestId("resume-rewrites-status-chip-pending-bullet-0"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-rewrites-status-chip-pending-bullet-1"),
    ).toBeInTheDocument();
  });

  it("derives bullet + accepted counts from the suggestions list (not hard-coded)", () => {
    renderRewrites(buildSuggestions([{}, {}, {}]));
    const tab = screen.getByTestId("resume-rewrites-tab");
    expect(tab.getAttribute("data-bullet-count")).toBe("3");
    expect(tab.getAttribute("data-accepted-count")).toBe("0");
  });

  it("shows the empty state when suggestions[] is empty and surfaces a rerun CTA when a handler is provided", async () => {
    const onRequestRerun = vi.fn();
    renderRewrites([], { onRequestRerun });
    expect(screen.getByTestId("resume-rewrites-empty")).toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-rewrites-diff-card"),
    ).not.toBeInTheDocument();
    await userEvent.setup().click(
      screen.getByTestId("resume-rewrites-rerun-tailor"),
    );
    expect(onRequestRerun).toHaveBeenCalledWith("bullet_suggestions");
  });
});

describe("ResumeRewritesTab selection + truncation", () => {
  it("truncates the row preview to 90 characters with an ellipsis", () => {
    const longBullet = "x".repeat(200);
    renderRewrites(buildSuggestions([{ rewritten: longBullet }]));
    const row = screen.getByTestId("resume-rewrites-bullet-row-bullet-0");
    expect(row.textContent ?? "").toMatch(/x{90}…/);
  });

  it("renders the full rewritten text only inside the diff card", () => {
    const longBullet = "y".repeat(200);
    renderRewrites(buildSuggestions([{ rewritten: longBullet }]));
    const rewritten = screen.getByTestId("resume-rewrites-rewritten-text");
    expect(rewritten.textContent).toBe(longBullet);
  });

  it("splits the suggestion reason into the why list", () => {
    renderRewrites(
      buildSuggestions([
        { reason: "Stronger ownership verb | Adds quantified impact" },
      ]),
    );
    expect(screen.getByTestId("resume-rewrites-why-0").textContent).toContain(
      "Stronger ownership verb",
    );
    expect(screen.getByTestId("resume-rewrites-why-1").textContent).toContain(
      "Adds quantified impact",
    );
  });

  it("changes the selected bullet when another row is clicked", async () => {
    renderRewrites(buildSuggestions([{}, {}]));
    const user = userEvent.setup();
    await user.click(
      screen.getByTestId("resume-rewrites-bullet-row-bullet-1"),
    );
    expect(
      screen.getByTestId("resume-rewrites-bullet-row-bullet-1"),
    ).toHaveAttribute("aria-selected", "true");
    expect(screen.getByTestId("resume-rewrites-tab")).toHaveAttribute(
      "data-selected-bullet-id",
      "bullet-1",
    );
  });
});

describe("ResumeRewritesTab accept → preview → save (D-20 accept-only flow)", () => {
  it("accepts the selected bullet, enables Preview & save, and marks the row accepted", async () => {
    renderRewrites(buildSuggestions([{}, {}]));
    const user = userEvent.setup();
    // Preview & save is disabled until at least one bullet is accepted.
    expect(screen.getByTestId("resume-rewrites-preview-save")).toBeDisabled();

    await user.click(screen.getByTestId("resume-rewrites-action-accept"));

    expect(screen.getByTestId("resume-rewrites-tab")).toHaveAttribute(
      "data-accepted-count",
      "1",
    );
    expect(
      screen.getByTestId("resume-rewrites-status-chip-accepted-bullet-0"),
    ).toBeInTheDocument();
    expect(screen.getByTestId("resume-rewrites-preview-save")).toBeEnabled();
  });

  it("opens the save modal and calls onOverwrite with the accepted rewrites when overwrite is confirmed", async () => {
    const onOverwrite = vi.fn().mockResolvedValue(undefined);
    const onSaveAsNew = vi.fn().mockResolvedValue(undefined);
    renderRewrites(
      buildSuggestions([
        { original: "Original A", rewritten: "Rewritten A" },
        { original: "Original B", rewritten: "Rewritten B" },
      ]),
      { onOverwrite, onSaveAsNew },
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-rewrites-action-accept"));
    await user.click(screen.getByTestId("resume-rewrites-preview-save"));

    const modal = screen.getByTestId("resume-rewrites-save-modal");
    expect(modal).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-rewrites-save-modal-item-bullet-0"),
    ).toBeInTheDocument();
    // Overwrite is the default save mode.
    expect(
      screen.getByTestId("resume-rewrites-save-mode-overwrite"),
    ).toHaveAttribute("data-active", "true");

    await user.click(screen.getByTestId("resume-rewrites-save-confirm"));
    expect(onOverwrite).toHaveBeenCalledTimes(1);
    expect(onOverwrite).toHaveBeenCalledWith([
      { original: "Original A", rewritten: "Rewritten A" },
    ]);
    expect(onSaveAsNew).not.toHaveBeenCalled();
  });

  it("calls onSaveAsNew when the save-as-new mode is selected before confirm", async () => {
    const onOverwrite = vi.fn().mockResolvedValue(undefined);
    const onSaveAsNew = vi.fn().mockResolvedValue(undefined);
    renderRewrites(
      buildSuggestions([{ original: "Original A", rewritten: "Rewritten A" }]),
      { onOverwrite, onSaveAsNew },
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-rewrites-action-accept"));
    await user.click(screen.getByTestId("resume-rewrites-preview-save"));
    await user.click(screen.getByTestId("resume-rewrites-save-mode-new"));
    await user.click(screen.getByTestId("resume-rewrites-save-confirm"));

    expect(onSaveAsNew).toHaveBeenCalledTimes(1);
    expect(onSaveAsNew).toHaveBeenCalledWith([
      { original: "Original A", rewritten: "Rewritten A" },
    ]);
    expect(onOverwrite).not.toHaveBeenCalled();
  });

  it("closes the save modal via cancel without invoking any save handler", async () => {
    const onOverwrite = vi.fn();
    const onSaveAsNew = vi.fn();
    renderRewrites(buildSuggestions([{}]), { onOverwrite, onSaveAsNew });
    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-rewrites-action-accept"));
    await user.click(screen.getByTestId("resume-rewrites-preview-save"));
    await user.click(screen.getByTestId("resume-rewrites-save-cancel"));
    expect(
      screen.queryByTestId("resume-rewrites-save-modal"),
    ).not.toBeInTheDocument();
    expect(onOverwrite).not.toHaveBeenCalled();
    expect(onSaveAsNew).not.toHaveBeenCalled();
  });
});

describe("ResumeRewritesTab polling banner", () => {
  it("renders the info polling banner above the bullet list when polling is active", () => {
    renderRewrites(buildSuggestions([{}]), {
      pollingBanner: {
        kind: "info",
        message: "AI is generating bullets",
      },
    });
    const banner = screen.getByTestId("resume-rewrites-polling-banner");
    expect(banner).toBeInTheDocument();
    expect(banner.textContent).toBe("AI is generating bullets");
    expect(banner).toHaveAttribute("role", "status");
  });

  it("renders the danger banner with a retry button when polling fails", async () => {
    const onRetry = vi.fn();
    renderRewrites([], {
      pollingBanner: {
        kind: "danger",
        message: "AI generation failed",
        onRetry,
      },
    });
    const banner = screen.getByTestId("resume-rewrites-failed-banner");
    expect(banner).toBeInTheDocument();
    expect(banner).toHaveAttribute("role", "alert");
    await userEvent.setup().click(
      screen.getByTestId("resume-rewrites-polling-retry"),
    );
    expect(onRetry).toHaveBeenCalledTimes(1);
  });
});

describe("ResumeRewritesTab privacy guard", () => {
  it("does not append originalBullet / rewritten text to URL, localStorage, or fetch transport log", () => {
    const original = "Sensitive original bullet body";
    const rewritten = "Sensitive rewritten suggestion body";
    const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(null, { status: 200 }) as unknown as Response,
    );
    const setItemSpy = vi.spyOn(window.localStorage, "setItem");
    const replaceState = vi.spyOn(window.history, "replaceState");

    renderRewrites(buildSuggestions([{ original, rewritten }]));

    expect(window.location.href).not.toContain(original);
    expect(window.location.href).not.toContain(rewritten);

    for (const call of setItemSpy.mock.calls) {
      expect(call[1]).not.toContain(original);
      expect(call[1]).not.toContain(rewritten);
    }
    for (const call of replaceState.mock.calls) {
      const url = call[2];
      if (typeof url === "string") {
        expect(url).not.toContain(original);
        expect(url).not.toContain(rewritten);
      }
    }
    expect(fetchSpy).not.toHaveBeenCalled();

    fetchSpy.mockRestore();
    setItemSpy.mockRestore();
    replaceState.mockRestore();
  });
});
