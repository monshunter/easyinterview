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
  it("renders exactly the five primary nav entries", () => {
    renderInProvider(<TopBar activeRoute="home" onNavigate={() => {}} />);
    const nav = screen.getByTestId("topbar-primary-nav");
    const items = nav.querySelectorAll("[data-testid^='topbar-nav-']");
    expect(items).toHaveLength(5);
    const ids = Array.from(items).map((el) =>
      el.getAttribute("data-testid")?.replace("topbar-nav-", ""),
    );
    expect(ids).toEqual([
      "home",
      "jd_match",
      "workspace",
      "resume_versions",
      "debrief",
    ]);
  });

  it("matches the documented PRIMARY_NAV_ROUTES truth source", () => {
    renderInProvider(<TopBar activeRoute="home" onNavigate={() => {}} />);
    for (const name of PRIMARY_NAV_ROUTES) {
      expect(screen.getByTestId(`topbar-nav-${name}`)).toBeInTheDocument();
    }
  });

  it("does not render legacy / removed entries (mistakes / growth / voice / drill)", () => {
    renderInProvider(<TopBar activeRoute="home" onNavigate={() => {}} />);
    for (const legacy of ["mistakes", "growth", "voice", "drill", "welcome"]) {
      expect(
        screen.queryByTestId(`topbar-nav-${legacy}`),
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
    await user.click(screen.getByTestId("topbar-nav-jd_match"));
    expect(onNavigate).toHaveBeenCalledWith({ name: "jd_match", params: {} });
  });
});

describe("TopBar display controls", () => {
  it("exposes theme / dark / lang controls bound to the display preferences provider", async () => {
    renderInProvider(<TopBar activeRoute="home" onNavigate={() => {}} />);
    const user = userEvent.setup();

    const themeSelect = screen.getByTestId(
      "topbar-theme-select",
    ) as HTMLSelectElement;
    const darkToggle = screen.getByTestId("topbar-dark-toggle");
    const langSelect = screen.getByTestId(
      "topbar-lang-select",
    ) as HTMLSelectElement;

    expect(themeSelect.value).toBe("warm");
    expect(darkToggle).toHaveAttribute("aria-pressed", "false");
    expect(langSelect.value).toBe("zh");

    await user.selectOptions(themeSelect, "forest");
    expect(themeSelect.value).toBe("forest");

    await user.click(darkToggle);
    expect(darkToggle).toHaveAttribute("aria-pressed", "true");

    await user.selectOptions(langSelect, "en");
    expect(langSelect.value).toBe("en");
  });
});
