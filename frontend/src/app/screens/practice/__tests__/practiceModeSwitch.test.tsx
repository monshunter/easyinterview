/**
 * @vitest-environment jsdom
 *
 * Current segmented control exposes text / phone. Legacy voice route params
 * normalize into the phone surface without rendering the retired voice UI.
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

describe("PracticeScreen mode switch", () => {
  it("legacy voice params render the current phone surface", () => {
    const voiceRoute: Route = {
      ...ROUTE_BASE,
      params: { ...ROUTE_BASE.params, mode: "voice", modality: "voice" },
    };
    withProviders(<PracticeScreen route={voiceRoute} />);
    expect(screen.queryByTestId("practice-voice-coming-soon")).toBeNull();
    expect(screen.getByTestId("practice-phone-surface")).toBeDefined();
    expect(screen.getByTestId("practice-phone-waveform")).toBeDefined();
    expect(screen.getByTestId("practice-phone-captions-toggle")).toBeDefined();
    expect(screen.queryByTestId("practice-voice-annotated-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-expression-panel")).toBeNull();
  });

  it("clicking the text segmented control from voice mode navigates back to mode='text'", async () => {
    const voiceRoute: Route = {
      ...ROUTE_BASE,
      params: { ...ROUTE_BASE.params, mode: "voice", modality: "voice" },
    };
    const { nav } = withProviders(<PracticeScreen route={voiceRoute} />);
    const textButton = screen.getByTestId("practice-topbar-mode-text");
    await userEvent.click(textButton);
    expect(nav).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "practice",
        params: expect.objectContaining({ mode: "text", modality: "text" }),
      }),
    );
  });

  it("clicking the phone segmented control navigates with mode='phone'", async () => {
    const { nav } = withProviders(<PracticeScreen route={ROUTE_BASE} />);
    const phoneBtn = screen.getByTestId("practice-topbar-mode-phone");
    await userEvent.click(phoneBtn);
    expect(nav).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "practice",
        params: expect.objectContaining({ mode: "phone", modality: "phone" }),
      }),
    );
  });

  it("text mode keeps phone / retired voice DOM out of the text surface", () => {
    const voiceRoute: Route = {
      ...ROUTE_BASE,
      params: { ...ROUTE_BASE.params, mode: "text", modality: "text" },
    };
    withProviders(<PracticeScreen route={voiceRoute} />);
    expect(screen.queryByTestId("practice-voice-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-annotated-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-expression-panel")).toBeNull();
    expect(screen.queryByTestId("practice-phone-surface")).toBeNull();
  });
});
