// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, type RenderResult } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";

import { DisplayPreferencesProvider } from "../display/DisplayPreferencesProvider";
import { PRIMARY_NAV_ROUTES } from "../routes";

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

  it("does not render non-current entries (mistakes / growth / voice / drill / debrief / profile)", () => {
    renderInProvider(<TopBar activeRoute="home" onNavigate={() => {}} />);
    for (const nonCurrent of ["mistakes", "growth", "voice", "drill", "welcome", "debrief", "profile"]) {
      expect(
        screen.queryByTestId(`topbar-nav-${nonCurrent}`),
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

  it("renders the signed-in avatar chip, opens the ui-design user dropdown, and dispatches the right routes", async () => {
    const onNavigate = vi.fn();
    renderInProvider(
      <TopBar
        activeRoute="home"
        onNavigate={onNavigate}
        signedIn={true}
        user={{
          displayName: "Alice Example",
          emailMasked: "ali***@example.com",
        }}
      />,
    );
    expect(screen.getByTestId("topbar-user-chip")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-user-avatar")).toHaveTextContent("AE");
    expect(screen.getByTestId("topbar-user-name")).toHaveTextContent(
      "Alice Example",
    );
    expect(screen.queryByTestId("topbar-user-menu")).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("topbar-user-profile"),
    ).not.toBeInTheDocument();
    expect(screen.queryByTestId("topbar-login")).not.toBeInTheDocument();
    expect(screen.queryByTestId("topbar-register")).not.toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-user-chip"));
    expect(screen.getByTestId("topbar-user-menu")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-user-menu-header")).toHaveTextContent(
      "Alice Example",
    );
    expect(screen.getByTestId("topbar-user-email")).toHaveTextContent(
      "ali***@example.com",
    );
    expect(screen.getByTestId("topbar-user-backdrop")).toBeInTheDocument();
    expect(
      screen.queryByTestId("topbar-user-profile"),
    ).not.toBeInTheDocument();

    await user.click(screen.getByTestId("topbar-user-settings"));
    expect(screen.queryByTestId("topbar-user-menu")).not.toBeInTheDocument();
    await user.click(screen.getByTestId("topbar-user-chip"));
    await user.click(screen.getByTestId("topbar-user-logout"));
    expect(onNavigate).toHaveBeenNthCalledWith(1, {
      name: "settings",
      params: {},
    });
    expect(onNavigate).toHaveBeenNthCalledWith(2, {
      name: "auth_logout",
      params: {},
    });
  });

  it("uses neutral signed-in fallbacks instead of prototype sample identity", async () => {
    render(
      <DisplayPreferencesProvider initial={{ lang: "zh" }}>
        <TopBar
          activeRoute="home"
          onNavigate={() => {}}
          signedIn={true}
          user={{ displayName: "", emailMasked: "" }}
        />
      </DisplayPreferencesProvider>,
    );

    expect(screen.getByTestId("topbar-user-name")).toHaveTextContent("候选人");
    expect(screen.getByTestId("topbar-user-name")).not.toHaveTextContent("刘哲");

    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-user-chip"));
    expect(screen.getByTestId("topbar-user-email")).toHaveTextContent(
      "邮箱未提供",
    );
    expect(screen.getByTestId("topbar-user-menu")).not.toHaveTextContent(
      "liuzhe@example.com",
    );
  });
});

describe("TopBar display controls", () => {
  it("exposes theme / dark / lang dropdown controls bound to the display preferences provider", async () => {
    render(
      <DisplayPreferencesProvider initial={{ lang: "zh" }}>
        <TopBar activeRoute="home" onNavigate={() => {}} />
      </DisplayPreferencesProvider>,
    );
    const user = userEvent.setup();

    const themeButton = screen.getByTestId("topbar-theme-button");
    const darkToggle = screen.getByTestId("topbar-dark-toggle");
    const langToggle = screen.getByTestId("topbar-lang-toggle");

    expect(themeButton).toHaveAttribute("aria-expanded", "false");
    expect(darkToggle).toHaveAttribute("aria-pressed", "false");
    expect(langToggle).toHaveAttribute("aria-expanded", "false");
    expect(langToggle).toHaveTextContent("中文");

    await user.click(themeButton);
    expect(themeButton).toHaveAttribute("aria-expanded", "true");
    await user.click(screen.getByTestId("topbar-theme-option-forest"));
    expect(themeButton).toHaveAttribute("aria-expanded", "false");

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
