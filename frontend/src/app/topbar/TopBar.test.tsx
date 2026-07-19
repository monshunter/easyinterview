// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, type RenderResult } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";

import { DisplayPreferencesProvider } from "../display/DisplayPreferencesProvider";
import {
  CONTEXT_ROUTES,
  isChromeHidden,
  PRIMARY_NAV_ROUTES,
} from "../routes";

import { TopBar } from "./TopBar";

function renderInProvider(node: ReactElement): RenderResult {
  return render(<DisplayPreferencesProvider>{node}</DisplayPreferencesProvider>);
}

describe("TopBar primary nav", () => {
  it("renders exactly the three primary nav entries (D-22)", () => {
    renderInProvider(<TopBar activeRoute="home" onNavigate={() => {}} />);
    const nav = screen.getByTestId("topbar-primary-nav");
    const items = nav.querySelectorAll("button[data-testid^='topbar-nav-']");
    expect(items).toHaveLength(3);
    const ids = Array.from(items).map((el) =>
      el.getAttribute("data-testid")?.replace("topbar-nav-", ""),
    );
    expect(ids).toEqual([
      "home",
      "workspace",
      "resume_versions",
    ]);
  });

  it("matches the documented PRIMARY_NAV_ROUTES truth source", () => {
    renderInProvider(<TopBar activeRoute="home" onNavigate={() => {}} />);
    for (const name of PRIMARY_NAV_ROUTES) {
      expect(screen.getByTestId(`topbar-nav-${name}`)).toBeInTheDocument();
    }
    expect(PRIMARY_NAV_ROUTES).not.toContain("reports");
    expect(screen.queryByTestId("topbar-nav-reports")).not.toBeInTheDocument();
    expect(CONTEXT_ROUTES).toContain("reports");
    expect(isChromeHidden("reports")).toBe(false);
  });

  it("renders the resume nav label from the ui-design TopBar copy", () => {
    render(
      <DisplayPreferencesProvider initial={{ lang: "zh" }}>
        <TopBar activeRoute="resume_versions" onNavigate={() => {}} />
      </DisplayPreferencesProvider>,
    );
    expect(screen.getByTestId("topbar-nav-resume_versions")).toHaveTextContent(
      "简历",
    );
  });

  it("renders the workspace nav as the concise Interview entry", () => {
    render(
      <DisplayPreferencesProvider initial={{ lang: "zh" }}>
        <TopBar activeRoute="workspace" onNavigate={() => {}} />
      </DisplayPreferencesProvider>,
    );
    expect(screen.getByTestId("topbar-nav-workspace")).toHaveTextContent(
      /^面试$/,
    );
  });

  it("does not render out-of-scope entries (mistakes / growth / voice / drill / debrief / profile)", () => {
    renderInProvider(<TopBar activeRoute="home" onNavigate={() => {}} />);
    for (const outOfScope of ["mistakes", "growth", "voice", "drill", "welcome", "debrief", "profile"]) {
      expect(
        screen.queryByTestId(`topbar-nav-${outOfScope}`),
      ).not.toBeInTheDocument();
    }
  });

  it("marks the active route with aria-current=page and the rest without", () => {
    renderInProvider(<TopBar activeRoute="workspace" onNavigate={() => {}} />);
    expect(screen.getByTestId("topbar-nav-workspace")).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByTestId("topbar-nav-home")).not.toHaveAttribute(
      "aria-current",
    );
  });

  it("invokes onNavigate with the clicked route name and empty params", async () => {
    const onNavigate = vi.fn();
    renderInProvider(<TopBar activeRoute="home" onNavigate={onNavigate} />);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-nav-workspace"));
    expect(onNavigate).toHaveBeenCalledWith({ name: "workspace", params: {} });
  });
});

describe("TopBar user menu", () => {
  it("renders one accessible settings gear for signed-in users without an account menu", async () => {
    const onNavigate = vi.fn();
    renderInProvider(
      <TopBar
        activeRoute="home"
        onNavigate={onNavigate}
        signedIn={true}
      />,
    );

    const settings = screen.getByRole("button", { name: /^设置$|^settings$/i });
    expect(settings).toHaveAttribute("data-testid", "topbar-settings");
    expect(screen.queryByTestId("topbar-user-chip")).not.toBeInTheDocument();
    expect(screen.queryByTestId("topbar-user-menu")).not.toBeInTheDocument();
    expect(screen.queryByTestId("topbar-user-logout")).not.toBeInTheDocument();

    await userEvent.setup().click(settings);
    expect(onNavigate).toHaveBeenCalledWith({ name: "settings", params: {} });
  });

  it("renders the single login entry when signed-out", () => {
    renderInProvider(
      <TopBar activeRoute="home" onNavigate={() => {}} signedIn={false} />,
    );
    expect(screen.getByTestId("topbar-login")).toBeInTheDocument();
    expect(screen.queryByTestId("topbar-register")).not.toBeInTheDocument();
    expect(screen.queryByTestId("topbar-user-menu")).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("topbar-user-profile"),
    ).not.toBeInTheDocument();
  });

});

describe("TopBar display controls", () => {
  it("keeps theme controls out of TopBar while exposing dark and language", async () => {
    render(
      <DisplayPreferencesProvider initial={{ lang: "zh" }}>
        <TopBar activeRoute="home" onNavigate={() => {}} />
      </DisplayPreferencesProvider>,
    );
    const user = userEvent.setup();

    const darkToggle = screen.getByTestId("topbar-dark-toggle");
    const langToggle = screen.getByTestId("topbar-lang-toggle");

    expect(screen.queryByTestId("topbar-theme-button")).not.toBeInTheDocument();
    expect(darkToggle).toHaveAttribute("aria-pressed", "false");
    expect(langToggle).toHaveAttribute("aria-expanded", "false");
    expect(langToggle).toHaveTextContent("中文");

    await user.click(darkToggle);
    expect(darkToggle).toHaveAttribute("aria-pressed", "true");

    await user.click(langToggle);
    expect(langToggle).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByTestId("topbar-lang-menu")).toBeInTheDocument();
    await user.click(screen.getByTestId("topbar-lang-option-en"));
    expect(langToggle).toHaveAttribute("aria-expanded", "false");
    expect(langToggle).toHaveTextContent("English");
  });
});
