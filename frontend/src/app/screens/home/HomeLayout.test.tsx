// @vitest-environment jsdom
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { HomeScreen } from "./HomeScreen";

function wrap(ui: React.ReactElement) {
  return (
    <NavigationProvider value={{ navigate: vi.fn() }}>
      <DisplayPreferencesProvider>{ui}</DisplayPreferencesProvider>
    </NavigationProvider>
  );
}

describe("Home layout", () => {
  it("keeps the JD input card paste-only", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const inputCard = screen.getByTestId("home-jd-input-card");
    const textarea = screen.getByTestId("home-jd-textarea");

    expect(screen.queryByTestId("home-source-layout")).not.toBeInTheDocument();
    expect(screen.queryByTestId("home-upload-source-panel")).not.toBeInTheDocument();
    expect(inputCard).toContainElement(textarea);
    expect(screen.queryByTestId("home-jd-source-controls")).not.toBeInTheDocument();
    expect(screen.queryByTestId("home-upload-trigger")).not.toBeInTheDocument();
    expect(screen.queryByTestId("home-url-trigger")).not.toBeInTheDocument();
    expect(screen.queryByTestId("home-modal-upload-backdrop")).not.toBeInTheDocument();
    expect(screen.queryByTestId("home-modal-url-backdrop")).not.toBeInTheDocument();
  });

  it("keeps resume selection compact with the create CTA on the same row", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const resumeRow = screen.getByTestId("home-resume-row");
    const resumeSelect = screen.getByTestId("home-resume-select");
    const createCta = screen.getByTestId("home-resume-create");

    expect(resumeRow).toContainElement(resumeSelect);
    expect(resumeRow).toContainElement(createCta);
    expect(resumeRow.className).toMatch(/\bei-home-resume-row\b/);
    expect(resumeSelect.className).toMatch(/\bei-home-resume-select\b/);
  });

  it("groups JD, resume, submit and privacy into one intake card", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const inputCard = screen.getByTestId("home-jd-input-card");
    const resumeRow = screen.getByTestId("home-resume-row");
    const submitRow = screen.getByTestId("home-submit-row");
    const submitButton = screen.getByTestId("home-jd-submit");

    expect(submitRow).toContainElement(submitButton);
    expect(inputCard).toContainElement(resumeRow);
    expect(inputCard).toContainElement(submitButton);
    expect(inputCard).toContainElement(screen.getByTestId("home-privacy-notice"));
    expect(
      resumeRow.compareDocumentPosition(submitRow) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });

  it("defines the screenshot-aligned desktop hierarchy and mobile stacking in scoped CSS", () => {
    const css = readFileSync(resolve(__dirname, "..", "screens.css"), "utf8");

    expect(css).toMatch(/\.ei-home-screen\s*\{[^}]*max-width:\s*none/);
    expect(css).toMatch(/\.ei-home-content\s*\{[^}]*max-width:\s*1400px/);
    expect(css).toMatch(/\.ei-home-intake-card\s*\{[^}]*border-radius:\s*16px/);
    expect(css).toMatch(/\.ei-home-jd-textarea\s*\{[^}]*width:\s*100%/);
    expect(css).toMatch(/\.ei-home-jd-textarea\s*\{[^}]*min-height:\s*212px/);
    expect(css).toMatch(/\.ei-home-jd-textarea\s*\{[^}]*overflow-y:\s*hidden/);
    expect(css).toMatch(/\.ei-home-recent-grid\s*\{[^}]*grid-template-columns:\s*1fr/);
    expect(css).toMatch(/@media \(max-width: 760px\)[\s\S]*\.ei-home-resume-action-row\s*\{[^}]*grid-template-columns:\s*1fr/);
  });

  it("remeasures the JD textarea when pasted content grows and shrinks", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const textarea = screen.getByTestId("home-jd-textarea");
    let scrollHeight = 420;
    Object.defineProperty(textarea, "scrollHeight", {
      configurable: true,
      get: () => scrollHeight,
    });

    fireEvent.change(textarea, {
      target: { value: "第一行\n第二行\n第三行\n第四行" },
    });
    expect(textarea).toHaveStyle({ height: "420px" });

    scrollHeight = 224;
    fireEvent.change(textarea, { target: { value: "缩短后的 JD" } });
    expect(textarea).toHaveStyle({ height: "224px" });
  });
});
