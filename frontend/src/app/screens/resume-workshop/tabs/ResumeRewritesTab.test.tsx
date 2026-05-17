// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type { ResumeVersion } from "../../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { ResumeRewritesTab } from "./ResumeRewritesTab";

const baseVersion: ResumeVersion = {
  id: "0195f2d0-0001-7000-8000-000000000202",
  resumeAssetId: "01918fa0-0000-7000-8000-000000001000",
  parentVersionId: "0195f2d0-0001-7000-8000-000000000201",
  versionType: "targeted",
  targetJobId: "01918fa0-0000-7000-8000-000000002000",
  displayName: "Northstar Systems frontend target",
  seedStrategy: "copy_master",
  focusAngle: "platform",
  structuredProfile: {},
  matchScore: 0.84,
  promptVersion: "resume_tailor.v2",
  rubricVersion: "not_applicable",
  modelId: "fixture-model:resume-tailor",
  provider: "fixture-provider",
  provenance: {
    promptVersion: "resume_tailor.v2",
    rubricVersion: "not_applicable",
    modelId: "fixture-model:resume-tailor",
    language: "zh-CN",
    featureFlag: "resume-workshop-additive",
    dataSourceVersion: "resume_version.v1",
  },
  suggestions: [],
  createdAt: "2026-05-12T08:24:00Z",
  updatedAt: "2026-05-12T08:24:00Z",
  deletedAt: null,
};

const buildVersionWithSuggestions = (
  rows: Array<{
    id: string;
    status: "pending" | "accepted" | "rejected";
    rewritten?: string;
    section?: string;
  }>,
): ResumeVersion => ({
  ...baseVersion,
  suggestions: rows.map((row) => ({
    id: row.id,
    originalBullet: `Original for ${row.id}`,
    suggestedBullet: row.rewritten ?? `Rewritten for ${row.id}`,
    reason: "Adds quantified ownership. | Clarifies the surface.",
    status: row.status,
    section: row.section ?? "Experience",
    decidedAt: row.status === "pending" ? null : "2026-05-12T09:00:00Z",
    tailorRunId: "01918fa0-0000-7000-8000-000000009000",
  })),
});

const renderRewrites = (version: ResumeVersion, props: Partial<Parameters<typeof ResumeRewritesTab>[0]> = {}) =>
  render(
    <DisplayPreferencesProvider>
      <ResumeRewritesTab version={version} {...props} />
    </DisplayPreferencesProvider>,
  );

describe("ResumeRewritesTab render baseline", () => {
  it("renders the scope banner with the version name and derived counts", () => {
    renderRewrites(
      buildVersionWithSuggestions([
        { id: "b1", status: "pending" },
        { id: "b2", status: "accepted" },
        { id: "b3", status: "rejected" },
        { id: "b4", status: "pending" },
      ]),
    );
    const banner = screen.getByTestId("resume-rewrites-scope-banner");
    expect(banner).toBeInTheDocument();
    expect(banner).toHaveAttribute("role", "status");
    expect(banner).toHaveAttribute("aria-live", "polite");
    expect(banner).toHaveTextContent("Northstar Systems frontend target");
    const counts = screen.getByTestId("resume-rewrites-counts");
    expect(counts.textContent).toMatch(/1.*已采纳|1.*accepted/);
    expect(counts.textContent).toMatch(/2.*待决定|2.*pending/);
    expect(counts.textContent).toMatch(/1.*已拒绝|1.*rejected/);
  });

  it("renders the bullet list with status chips and pre-selects the first bullet", () => {
    renderRewrites(
      buildVersionWithSuggestions([
        { id: "b1", status: "pending", rewritten: "Pending bullet detail" },
        { id: "b2", status: "accepted" },
      ]),
    );
    expect(screen.getByTestId("resume-rewrites-bullet-list")).toBeInTheDocument();
    expect(screen.getByTestId("resume-rewrites-bullet-row-b1")).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByTestId("resume-rewrites-bullet-row-b2")).toHaveAttribute(
      "aria-selected",
      "false",
    );
    expect(
      screen.getByTestId("resume-rewrites-status-chip-pending-b1"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-rewrites-status-chip-accepted-b2"),
    ).toBeInTheDocument();
  });

  it("derives accepted / pending / rejected counts from the suggestions list (not hard-coded)", () => {
    renderRewrites(
      buildVersionWithSuggestions([
        { id: "b1", status: "accepted" },
        { id: "b2", status: "accepted" },
        { id: "b3", status: "rejected" },
      ]),
    );
    const tab = screen.getByTestId("resume-rewrites-tab");
    expect(tab.getAttribute("data-accepted-count")).toBe("2");
    expect(tab.getAttribute("data-pending-count")).toBe("0");
    expect(tab.getAttribute("data-rejected-count")).toBe("1");
    expect(tab.getAttribute("data-bullet-count")).toBe("3");
  });

  it("shows the empty state when suggestions[] is empty and surfaces a rerun CTA when a handler is provided", async () => {
    const onRequestRerun = vi.fn();
    renderRewrites(buildVersionWithSuggestions([]), {
      onRequestRerun,
    });
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
    renderRewrites(
      buildVersionWithSuggestions([
        { id: "b1", status: "pending", rewritten: longBullet },
      ]),
    );
    const row = screen.getByTestId("resume-rewrites-bullet-row-b1");
    expect(row.textContent ?? "").toMatch(/x{90}…/);
  });

  it("renders the full rewritten text only inside the diff card", () => {
    const longBullet = "y".repeat(200);
    renderRewrites(
      buildVersionWithSuggestions([
        { id: "b1", status: "pending", rewritten: longBullet },
      ]),
    );
    const rewritten = screen.getByTestId("resume-rewrites-rewritten-text");
    expect(rewritten.textContent).toBe(longBullet);
  });

  it("changes the selected bullet when another row is clicked and resets editing state", async () => {
    renderRewrites(
      buildVersionWithSuggestions([
        { id: "b1", status: "pending" },
        { id: "b2", status: "pending" },
      ]),
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-rewrites-action-edit"));
    expect(screen.getByTestId("resume-rewrites-edit-textarea")).toBeInTheDocument();

    await user.click(screen.getByTestId("resume-rewrites-bullet-row-b2"));
    expect(
      screen.queryByTestId("resume-rewrites-edit-textarea"),
    ).not.toBeInTheDocument();
    expect(
      screen.getByTestId("resume-rewrites-bullet-row-b2"),
    ).toHaveAttribute("aria-selected", "true");
  });
});

describe("ResumeRewritesTab action callbacks (Phase 3 wiring)", () => {
  it("invokes onAccept / onReject with the selected bullet id", async () => {
    const onAccept = vi.fn().mockResolvedValue(undefined);
    const onReject = vi.fn().mockResolvedValue(undefined);
    renderRewrites(
      buildVersionWithSuggestions([{ id: "b1", status: "pending" }]),
      { onAccept, onReject },
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-rewrites-action-accept"));
    expect(onAccept).toHaveBeenCalledWith("b1");
    await user.click(screen.getByTestId("resume-rewrites-action-reject"));
    expect(onReject).toHaveBeenCalledWith("b1");
  });

  it("renders the rerun CTA on the diff card when onRequestRerun is provided", async () => {
    const onRequestRerun = vi.fn();
    renderRewrites(
      buildVersionWithSuggestions([{ id: "b1", status: "pending" }]),
      { onRequestRerun },
    );
    await userEvent.setup().click(
      screen.getByTestId("resume-rewrites-rerun-tailor"),
    );
    expect(onRequestRerun).toHaveBeenCalledWith("bullet_suggestions");
  });
});

describe("ResumeRewritesTab polling banner (plan 003 Phase 5)", () => {
  it("renders the info polling banner above the bullet list when polling is active", () => {
    renderRewrites(
      buildVersionWithSuggestions([{ id: "b1", status: "pending" }]),
      {
        pollingBanner: {
          kind: "info",
          message: "AI is generating bullets",
        },
      },
    );
    const banner = screen.getByTestId("resume-rewrites-polling-banner");
    expect(banner).toBeInTheDocument();
    expect(banner.textContent).toBe("AI is generating bullets");
    expect(banner).toHaveAttribute("role", "status");
  });

  it("renders the danger banner with a retry button when polling fails", async () => {
    const onRetry = vi.fn();
    renderRewrites(buildVersionWithSuggestions([]), {
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

    const version = buildVersionWithSuggestions([
      { id: "b1", status: "pending", rewritten },
    ]);
    version.suggestions[0]!.originalBullet = original;

    renderRewrites(version);

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
