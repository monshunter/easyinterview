/**
 * @vitest-environment jsdom
 *
 * Item 1.1: PracticeScreen static shell + ≥ 20 practice-* testids + control
 * type assertions. Source-level mirror of `ui-design/src/screen-practice.jsx::
 * PracticeScreen` text branch (lines 184-326). This test asserts the static
 * skeleton; data-driven assertions land in later phases.
 *
 * Truth source: docs/spec/frontend-workspace-and-practice/plans/
 *   002-practice-text-event-loop/plan.md §3.5 + checklist.md item 1.1.
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { InterviewContextProvider } from "../../interview-context/InterviewContext";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { PracticeScreen } from "./PracticeScreen";

const PRACTICE_ROUTE: Route = {
  name: "practice",
  params: {
    sessionId: "01918fa0-0000-7000-8000-000000005000",
    planId: "01918fa0-0000-7000-8000-000000004000",
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    jdId: "jd-1",
    resumeId: "01918fa0-0000-7000-8000-000000001000",
    roundId: "round-tech1",
    mode: "text",
    modality: "text",
    practiceMode: "assisted",
    practiceGoal: "baseline",
    hintUsed: "false",
    hintCount: "0",
  },
};

function withProviders(ui: ReactNode) {
  const nav = vi.fn();
  return {
    nav,
    ...render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <NavigationProvider value={{ navigate: nav }}>{ui}</NavigationProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("PracticeScreen static shell (item 1.1)", () => {
  it("renders TopBar with company/title and required controls", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    expect(screen.getByTestId("practice-topbar")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-company")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-title")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-question")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-timer")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-pause")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-mode-text")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-mode-voice")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-strict")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-role")).toBeDefined();
  });

  it("strict toggle is a switch (role + aria-checked) — not a checkbox / select", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    const strict = screen.getByTestId("practice-topbar-strict");
    expect(strict.getAttribute("role")).toBe("switch");
    // assisted route param → aria-checked="false"
    expect(strict.getAttribute("aria-checked")).toBe("false");
    expect(strict.tagName).not.toBe("SELECT");
    expect(strict.tagName).not.toBe("INPUT");
  });

  it("segmented mode controls are buttons (not <select>)", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    const text = screen.getByTestId("practice-topbar-mode-text");
    const voice = screen.getByTestId("practice-topbar-mode-voice");
    expect(text.tagName).toBe("BUTTON");
    expect(voice.tagName).toBe("BUTTON");
  });

  it("RoleDropdown is a menu trigger button (not a <select>)", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    const role = screen.getByTestId("practice-topbar-role");
    expect(role.tagName).toBe("BUTTON");
  });

  it("renders SessionMap on the left rail with label + at least one item", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    expect(screen.getByTestId("practice-sessionmap")).toBeDefined();
    expect(screen.getByTestId("practice-sessionmap-label")).toBeDefined();
    expect(screen.getByTestId("practice-sessionmap-item-0")).toBeDefined();
  });

  it("renders QuestionCard with badge + topic + prompt skeleton", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    expect(screen.getByTestId("practice-question")).toBeDefined();
    expect(screen.getByTestId("practice-question-badge")).toBeDefined();
    expect(screen.getByTestId("practice-question-prompt")).toBeDefined();
  });

  it("renders Transcript container + helper line", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    expect(screen.getByTestId("practice-transcript")).toBeDefined();
    expect(screen.getByTestId("practice-transcript-helper")).toBeDefined();
  });

  it("renders InputBar with textarea + send / skip / dictate buttons + hint (assisted)", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    const textarea = screen.getByTestId("practice-input-textarea");
    expect(textarea.tagName).toBe("TEXTAREA");
    expect(screen.getByTestId("practice-input-send").tagName).toBe("BUTTON");
    expect(screen.getByTestId("practice-input-skip").tagName).toBe("BUTTON");
    expect(screen.getByTestId("practice-input-dictate").tagName).toBe("BUTTON");
    // assisted route param → hint button rendered
    expect(screen.getByTestId("practice-input-hint")).toBeDefined();
  });

  it("renders RightPanel with JD link card + AI transparency + finish CTA", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    expect(screen.getByTestId("practice-rightpanel")).toBeDefined();
    expect(screen.getByTestId("practice-rightpanel-jd")).toBeDefined();
    expect(screen.getByTestId("practice-rightpanel-ai-transparency")).toBeDefined();
    expect(screen.getByTestId("practice-rightpanel-cta-finish").tagName).toBe(
      "BUTTON",
    );
  });

  it("provides ≥ 20 unique practice-* testids on the static shell", () => {
    const { container } = withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    const nodes = container.querySelectorAll("[data-testid^='practice-']");
    const unique = new Set(
      Array.from(nodes).map((n) => n.getAttribute("data-testid")),
    );
    expect(unique.size).toBeGreaterThanOrEqual(20);
  });

  it("does not render any voice surface DOM in text mode (negative gate)", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    expect(screen.queryByTestId("practice-voice-coming-soon")).toBeNull();
    // Voice waveform / annotated waveform / expression panel must not appear.
    expect(screen.queryByTestId("practice-voice-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-annotated-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-expression-panel")).toBeNull();
  });

  it("does not render legacy prototype testids (negative gate)", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    // Legacy practice prototype surfaces deprecated by the current spec.
    expect(screen.queryByTestId("practice-mode-card-strict")).toBeNull();
    expect(screen.queryByTestId("practice-mode-card-assisted")).toBeNull();
    expect(screen.queryByTestId("growth-summary")).toBeNull();
    expect(screen.queryByTestId("drill-builder-form")).toBeNull();
    expect(screen.queryByTestId("mistakes-queue-list")).toBeNull();
  });

  it("renders the ui-design voice surface anchors when mode='voice'", () => {
    const voiceRoute: Route = {
      ...PRACTICE_ROUTE,
      params: { ...PRACTICE_ROUTE.params, mode: "voice", modality: "voice" },
    };
    withProviders(<PracticeScreen route={voiceRoute} />);
    expect(screen.queryByTestId("practice-voice-coming-soon")).toBeNull();
    expect(screen.getByTestId("practice-voice-surface")).toBeDefined();
    expect(screen.getByTestId("practice-voice-waveform")).toBeDefined();
    expect(screen.getByTestId("practice-voice-annotated-waveform")).toBeDefined();
    expect(screen.getByTestId("practice-voice-live-transcript")).toBeDefined();
    expect(screen.getByTestId("practice-voice-expression-panel")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-live").style.visibility).toBe(
      "visible",
    );
  });

  it("renders PracticeSessionLostState when sessionId is missing", () => {
    const lostRoute: Route = {
      ...PRACTICE_ROUTE,
      params: { ...PRACTICE_ROUTE.params, sessionId: "" },
    };
    withProviders(<PracticeScreen route={lostRoute} />);
    expect(screen.getByTestId("practice-session-lost")).toBeDefined();
    expect(screen.getByTestId("practice-session-lost-cta").tagName).toBe(
      "BUTTON",
    );
  });

  it("hides hint button + LIVE NOTES + experience cards when practiceMode='strict'", () => {
    const strictRoute: Route = {
      ...PRACTICE_ROUTE,
      params: { ...PRACTICE_ROUTE.params, practiceMode: "strict" },
    };
    withProviders(<PracticeScreen route={strictRoute} />);
    expect(screen.queryByTestId("practice-input-hint")).toBeNull();
    expect(screen.queryByTestId("practice-sessionmap-live-notes")).toBeNull();
    expect(screen.queryByTestId("practice-rightpanel-exp-0")).toBeNull();
    expect(screen.getByTestId("practice-rightpanel-strict-banner")).toBeDefined();
    // strict route → strict toggle aria-checked='true'
    const strict = screen.getByTestId("practice-topbar-strict");
    expect(strict.getAttribute("aria-checked")).toBe("true");
  });
});
