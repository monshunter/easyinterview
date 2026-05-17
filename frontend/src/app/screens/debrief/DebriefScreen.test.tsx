// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { InterviewContextProvider } from "../../interview-context/InterviewContext";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { DebriefScreen } from "./DebriefScreen";

function renderDebriefScreen(navigate = vi.fn()) {
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider initial={{ lang: "zh" }}>
        <InterviewContextProvider>
          <NavigationProvider value={{ navigate }}>
            <DebriefScreen route={{ name: "debrief", params: {} }} />
          </NavigationProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("DebriefScreen — TestDebriefScreen_DefaultRender", () => {
  it("renders the route-debrief shell with Header + ContextStrip + Stepper + step panel", () => {
    renderDebriefScreen();
    const shell = screen.getByTestId("route-debrief");
    expect(shell).toBeInTheDocument();
    expect(shell).toHaveAttribute("data-route-name", "debrief");
    expect(shell).toHaveAttribute("data-step", "0");
    expect(shell).toHaveAttribute("data-input-mode", "text");
    expect(shell).toHaveAttribute("data-picker-kind", "none");
    expect(screen.getByTestId("debrief-header")).toBeInTheDocument();
    expect(screen.getByTestId("debrief-context-strip")).toBeInTheDocument();
    expect(screen.getByTestId("debrief-stepper")).toBeInTheDocument();
    expect(screen.getByTestId("debrief-step-panel-0")).toBeInTheDocument();
  });

  it("invokes navigate({name:'home'}) when the header back control is clicked", () => {
    const { navigate } = renderDebriefScreen();
    fireEvent.click(screen.getByTestId("debrief-header-back"));
    expect(navigate).toHaveBeenCalledWith({ name: "home" });
  });

  it("flips data-picker-kind when a context-strip card is opened", () => {
    renderDebriefScreen();
    fireEvent.click(screen.getByTestId("debrief-context-card-targetJob-open"));
    const shell = screen.getByTestId("route-debrief");
    expect(shell).toHaveAttribute("data-picker-kind", "targetJob");
  });
});
