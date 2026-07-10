// @vitest-environment jsdom
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { HomeScreen } from "./HomeScreen";

const SCENARIO_DIR = resolve(
  __dirname,
  "../../../../../test/scenarios/e2e/p0-014-home-default-render",
);

function wrap(ui: React.ReactElement, navigate = vi.fn()) {
  return (
    <NavigationProvider value={{ navigate }}>
      <DisplayPreferencesProvider>{ui}</DisplayPreferencesProvider>
    </NavigationProvider>
  );
}

describe("HomeScreen", () => {
  it("keeps P0.014 scenario claims bounded to its Vitest runner evidence", () => {
    const readme = readFileSync(resolve(SCENARIO_DIR, "README.md"), "utf8");
    const seed = readFileSync(resolve(SCENARIO_DIR, "data/seed-input.md"), "utf8");
    const expected = readFileSync(
      resolve(SCENARIO_DIR, "data/expected-outcome.md"),
      "utf8",
    );
    const trigger = readFileSync(resolve(SCENARIO_DIR, "scripts/trigger.sh"), "utf8");
    const assets = `${readme}\n${seed}\n${expected}`;

    for (const testFile of [
      "targetJob.realApiMode.test.ts",
      "HomeScreen.test.tsx",
      "HomeLayout.test.tsx",
      "HomeResumeSelection.test.tsx",
      "HomeRecentMocks.test.tsx",
      "MockInterviewCard.test.tsx",
    ]) {
      expect(trigger).toContain(testFile);
      expect(expected).toContain(testFile);
    }
    expect(readme).toContain("stub fetch");
    expect(assets).not.toMatch(
      /Playwright|chromium|frontend dist|TopBar highlights|Theme switching|warm\/dark|Mobile responsive|UI-design golden|Real Backend Overlay|live backend/,
    );
  });

  it("renders the home shell with required testids", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    expect(screen.getByTestId("home-hero-label")).toBeInTheDocument();
    expect(screen.getByTestId("home-hero-title")).toBeInTheDocument();
    expect(screen.queryByTestId("home-hero-sub")).not.toBeInTheDocument();
    expect(screen.getByTestId("home-jd-textarea")).toBeInTheDocument();
    expect(screen.getByTestId("home-jd-submit")).toBeInTheDocument();
    expect(screen.getByTestId("home-resume-select")).toBeInTheDocument();
    expect(screen.queryByTestId("home-aux-jobpicks")).not.toBeInTheDocument();
    expect(screen.queryByTestId("home-aux-debrief")).not.toBeInTheDocument();
  });

  it("renders correct control types", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const textarea = screen.getByTestId("home-jd-textarea");
    expect(textarea.tagName).toBe("TEXTAREA");

    const submitBtn = screen.getByTestId("home-jd-submit");
    expect(submitBtn.tagName).toBe("BUTTON");

    const resumeSelect = screen.getByTestId("home-resume-select");
    expect(resumeSelect.tagName).toBe("SELECT");
  });

  it("renders shell data attributes", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const root = screen.getByTestId("route-home");
    expect(root).toBeInTheDocument();
    expect(root.getAttribute("data-route-name")).toBe("home");
    expect(root.className).toMatch(/\bei-screen-shell\b/);
  });

  it("submit button is disabled when textarea is empty", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const submitBtn = screen.getByTestId("home-jd-submit");
    expect(submitBtn).toBeDisabled();
  });

  it("submit button remains disabled without a selected resume", async () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const textarea = screen.getByTestId("home-jd-textarea");
    const submitBtn = screen.getByTestId("home-jd-submit");

    await userEvent.type(textarea, "Software Engineer JD");

    expect(submitBtn).toBeDisabled();
  });

  it("navigates to resume_versions on resume create CTA click", () => {
    const navigate = vi.fn();
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />, navigate));

    const cta = screen.getByTestId("home-resume-create");
    fireEvent.click(cta);

    expect(navigate).toHaveBeenCalledWith({
      name: "resume_versions",
      params: { flow: "create" },
    });
  });

  it("renders i18n content in English", () => {
    render(
      <NavigationProvider value={{ navigate: vi.fn() }}>
        <DisplayPreferencesProvider initial={{ lang: "en" }}>
          <HomeScreen route={{ name: "home", params: {} }} />
        </DisplayPreferencesProvider>
      </NavigationProvider>,
    );

    expect(screen.getByTestId("home-hero-label")).toHaveTextContent(
      "HOME · MOCK INTERVIEWS",
    );
    expect(screen.getByTestId("home-hero-title")).toHaveTextContent(
      "Let's win the interview you already care about.",
    );
    expect(screen.getByTestId("home-jd-submit")).toHaveTextContent(
      "Start interview now",
    );
  });

  it("does not surface out-of-scope prototype testids", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    expect(screen.queryByTestId("home-pasted-success")).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("home-mocked-recent"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("home-recent-mock-card-default"),
    ).not.toBeInTheDocument();
  });
});
