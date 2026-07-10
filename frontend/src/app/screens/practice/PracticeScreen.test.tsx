/**
 * @vitest-environment jsdom
 *
 * PracticeScreen current real-interview shell. Source-level mirror of
 * `ui-design/src/screen-practice.jsx::PracticeScreen` after the phone-mode
 * simplification.
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
    expect(screen.getByTestId("practice-topbar-question")).toHaveTextContent(
      "Question",
    );
    expect(screen.getByTestId("practice-topbar-timer")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-pause")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-pause")).toHaveTextContent(
      "Pause",
    );
    expect(screen.getByTestId("practice-topbar-mode-text")).toBeDefined();
    expect(screen.getByTestId("practice-topbar-mode-phone")).toBeDefined();
    expect(screen.getByTestId("practice-finish-cta").tagName).toBe("BUTTON");
    expect(screen.queryByTestId("practice-topbar-mode-voice")).toBeNull();
    expect(screen.queryByTestId("practice-topbar-strict")).toBeNull();
    expect(screen.queryByTestId("practice-topbar-role")).toBeNull();
  });

  it("does not render strict or role controls inside the session", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    expect(screen.queryByTestId("practice-topbar-strict")).toBeNull();
    expect(screen.queryByTestId("practice-topbar-role")).toBeNull();
    expect(screen.queryByTestId("practice-strict-locked-toast")).toBeNull();
  });

  it("segmented mode controls are buttons (not <select>)", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    const text = screen.getByTestId("practice-topbar-mode-text");
    const phone = screen.getByTestId("practice-topbar-mode-phone");
    expect(text.tagName).toBe("BUTTON");
    expect(phone.tagName).toBe("BUTTON");
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

  it("renders InputBar with textarea + send + optional hint only", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    const textarea = screen.getByTestId("practice-input-textarea");
    expect(textarea.tagName).toBe("TEXTAREA");
    expect(screen.getByTestId("practice-input-send").tagName).toBe("BUTTON");
    expect(screen.getByTestId("practice-input-hint")).toBeDefined();
    expect(screen.queryByTestId("practice-input-skip")).toBeNull();
    expect(screen.queryByTestId("practice-input-dictate")).toBeNull();
  });

  it("renders the current two-column session layout in text mode", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    const main = screen.getByTestId("practice-main");
    const center = screen.getByTestId("practice-center");
    expect(main.contains(screen.getByTestId("practice-sessionmap"))).toBe(
      true,
    );
    expect(main.contains(center)).toBe(true);
    expect(center.contains(screen.getByTestId("practice-question"))).toBe(
      true,
    );
    expect(center.contains(screen.getByTestId("practice-transcript"))).toBe(
      true,
    );
    expect(center.contains(screen.getByTestId("practice-input"))).toBe(true);
    expect(
      screen
        .getByTestId("practice-finish-cta-wrap")
        .closest("[data-testid='practice-topbar']"),
    ).not.toBeNull();
  });

  it("provides ≥ 20 unique practice-* testids on the static shell", () => {
    const { container } = withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    const nodes = container.querySelectorAll("[data-testid^='practice-']");
    const unique = new Set(
      Array.from(nodes).map((n) => n.getAttribute("data-testid")),
    );
    expect(unique.size).toBeGreaterThanOrEqual(20);
  });

  it("does not render any phone surface DOM in text mode (negative gate)", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    expect(screen.queryByTestId("practice-voice-coming-soon")).toBeNull();
    expect(screen.queryByTestId("practice-voice-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-annotated-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-expression-panel")).toBeNull();
    expect(screen.queryByTestId("practice-phone-surface")).toBeNull();
  });

  it("does not render out-of-scope prototype testids (negative gate)", () => {
    withProviders(<PracticeScreen route={PRACTICE_ROUTE} />);
    // Out-of-scope practice prototype surfaces excluded by the current spec.
    expect(screen.queryByTestId("practice-mode-card-strict")).toBeNull();
    expect(screen.queryByTestId("practice-mode-card-assisted")).toBeNull();
    expect(screen.queryByTestId("growth-summary")).toBeNull();
    expect(screen.queryByTestId("drill-builder-form")).toBeNull();
    expect(screen.queryByTestId("mistakes-queue-list")).toBeNull();
  });

  it("renders the phone surface anchors when mode='phone'", () => {
    const phoneRoute: Route = {
      ...PRACTICE_ROUTE,
      params: { ...PRACTICE_ROUTE.params, mode: "phone", modality: "phone" },
    };
    withProviders(<PracticeScreen route={phoneRoute} />);
    expect(screen.queryByTestId("practice-voice-coming-soon")).toBeNull();
    expect(screen.getByTestId("practice-phone-surface")).toBeDefined();
    expect(screen.getByTestId("practice-phone-waveform")).toBeDefined();
    expect(screen.getByTestId("practice-phone-captions-toggle")).toBeDefined();
    expect(screen.getByTestId("practice-phone-hangup")).toBeDefined();
    expect(screen.getByTestId("practice-phone-restart")).toBeDefined();
    expect(screen.queryByTestId("practice-voice-annotated-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-expression-panel")).toBeNull();
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

  it("keeps hint available even when out-of-scope practiceMode='strict'", () => {
    const strictRoute: Route = {
      ...PRACTICE_ROUTE,
      params: { ...PRACTICE_ROUTE.params, practiceMode: "strict" },
    };
    withProviders(<PracticeScreen route={strictRoute} />);
    expect(screen.getByTestId("practice-input-hint")).toBeDefined();
    expect(screen.queryByTestId("practice-topbar-strict")).toBeNull();
  });
});
