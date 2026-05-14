/**
 * @vitest-environment jsdom
 *
 * Phase 1 test-checklist mapping — VoiceSurfaceComingSoon placeholder + back
 * to text. Phase 3.6 extends this file with the segmented control click
 * handler (nav with mode='voice' / mode='text') and negative voice DOM
 * assertions.
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";

import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { InterviewContextProvider } from "../../../interview-context/InterviewContext";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import type { Route } from "../../../routes";
import { PracticeScreen } from "../PracticeScreen";

const ROUTE_BASE: Route = {
  name: "practice",
  params: {
    sessionId: "01918fa0-0000-7000-8000-000000005000",
    planId: "01918fa0-0000-7000-8000-000000004000",
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    jdId: "jd-1",
    resumeVersionId: "01918fa0-0000-7000-8000-000000001000",
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

describe("PracticeScreen mode switch (Phase 1 + 3.6 contract)", () => {
  it("voice mode renders the scoped VoiceSurfaceComingSoon placeholder", () => {
    const voiceRoute: Route = {
      ...ROUTE_BASE,
      params: { ...ROUTE_BASE.params, mode: "voice", modality: "voice" },
    };
    withProviders(<PracticeScreen route={voiceRoute} />);
    expect(screen.getByTestId("practice-voice-coming-soon")).toBeDefined();
    expect(screen.getByTestId("practice-voice-coming-soon-icon")).toBeDefined();
    expect(screen.getByTestId("practice-voice-coming-soon-title")).toBeDefined();
    expect(screen.getByTestId("practice-voice-coming-soon-desc")).toBeDefined();
    expect(screen.getByTestId("practice-voice-coming-soon-back-to-text"))
      .toBeDefined();
  });

  it("clicking VoiceSurfaceComingSoon back-to-text navigates back to mode='text'", async () => {
    const voiceRoute: Route = {
      ...ROUTE_BASE,
      params: { ...ROUTE_BASE.params, mode: "voice", modality: "voice" },
    };
    const { nav } = withProviders(<PracticeScreen route={voiceRoute} />);
    const back = screen.getByTestId("practice-voice-coming-soon-back-to-text");
    await userEvent.click(back);
    expect(nav).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "practice",
        params: expect.objectContaining({ mode: "text", modality: "text" }),
      }),
    );
  });

  it("clicking the voice segmented control nav-s with mode='voice'", async () => {
    const { nav } = withProviders(<PracticeScreen route={ROUTE_BASE} />);
    const voiceBtn = screen.getByTestId("practice-topbar-mode-voice");
    await userEvent.click(voiceBtn);
    expect(nav).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "practice",
        params: expect.objectContaining({ mode: "voice", modality: "voice" }),
      }),
    );
  });

  it("voice mode does not leak voice surface DOM (waveform / annotated / expression)", () => {
    const voiceRoute: Route = {
      ...ROUTE_BASE,
      params: { ...ROUTE_BASE.params, mode: "voice", modality: "voice" },
    };
    withProviders(<PracticeScreen route={voiceRoute} />);
    expect(screen.queryByTestId("practice-voice-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-annotated-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-expression-panel")).toBeNull();
  });
});
