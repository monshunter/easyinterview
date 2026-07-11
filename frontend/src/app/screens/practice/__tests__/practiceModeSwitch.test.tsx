/**
 * @vitest-environment jsdom
 *
 * One phone icon enters phone mode. The same top-bar icon and the central hang-up
 * icon leave phone mode while preserving the current session.
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
  it("out-of-scope voice params render the text surface instead of phone mode", () => {
    const voiceRoute: Route = {
      ...ROUTE_BASE,
      params: { ...ROUTE_BASE.params, mode: "voice", modality: "voice" },
    };
    withProviders(<PracticeScreen route={voiceRoute} />);
    expect(screen.queryByTestId("practice-voice-coming-soon")).toBeNull();
    expect(screen.getByTestId("practice-input")).toBeDefined();
    expect(screen.queryByTestId("practice-phone-surface")).toBeNull();
    expect(screen.queryByTestId("practice-phone-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-phone-captions-toggle")).toBeNull();
    expect(screen.queryByTestId("practice-voice-annotated-waveform")).toBeNull();
    expect(screen.queryByTestId("practice-voice-expression-panel")).toBeNull();
  });

  it("clicking the phone icon from out-of-scope voice params navigates with mode='phone'", async () => {
    const voiceRoute: Route = {
      ...ROUTE_BASE,
      params: { ...ROUTE_BASE.params, mode: "voice", modality: "voice" },
    };
    const { nav } = withProviders(<PracticeScreen route={voiceRoute} />);
    const phoneButton = screen.getByTestId("practice-topbar-phone-toggle");
    await userEvent.click(phoneButton);
    expect(nav).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "practice",
        params: expect.objectContaining({ mode: "phone", modality: "phone" }),
      }),
    );
  });

  it("clicking the phone icon navigates with mode='phone'", async () => {
    const { nav } = withProviders(<PracticeScreen route={ROUTE_BASE} />);
    const phoneBtn = screen.getByTestId("practice-topbar-phone-toggle");
    await userEvent.click(phoneBtn);
    expect(nav).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "practice",
        params: expect.objectContaining({ mode: "phone", modality: "phone" }),
      }),
    );
  });

  it.each([
    ["top bar", "practice-topbar-phone-toggle"],
    ["center hang-up", "practice-phone-hangup"],
  ])("%s exit keeps the session and returns to text mode", async (_label, testId) => {
    const phoneRoute: Route = {
      ...ROUTE_BASE,
      params: { ...ROUTE_BASE.params, mode: "phone", modality: "phone" },
    };
    const { nav } = withProviders(<PracticeScreen route={phoneRoute} />);

    await userEvent.click(screen.getByTestId(testId));

    expect(nav).toHaveBeenCalledWith({
      name: "practice",
      params: expect.objectContaining({
        sessionId: ROUTE_BASE.params.sessionId,
        mode: "text",
        modality: "text",
      }),
    });
  });

  it("text mode keeps phone / deleted voice DOM out of the text surface", () => {
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
