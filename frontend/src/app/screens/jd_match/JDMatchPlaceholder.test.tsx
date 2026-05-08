// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { JDMatchScreen } from "./JDMatchScreen";

function wrap(ui: React.ReactElement, navigate = vi.fn()) {
  return (
    <NavigationProvider value={{ navigate }}>
      <DisplayPreferencesProvider>{ui}</DisplayPreferencesProvider>
    </NavigationProvider>
  );
}

describe("JDMatchPlaceholder", () => {
  it("renders hero with label, title, sub", () => {
    render(
      wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />),
    );

    expect(screen.getByTestId("jdmatch-hero-label")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-hero-title")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-hero-sub")).toBeInTheDocument();
  });

  it("renders profile snapshot chip", () => {
    render(
      wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />),
    );

    expect(screen.getByTestId("jdmatch-profile-chip")).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-profile-chip-title"),
    ).toBeInTheDocument();
  });

  it("renders three tabs with labels", () => {
    render(
      wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />),
    );

    expect(screen.getByTestId("jdmatch-tab-recommended")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-tab-search")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-tab-watchlist")).toBeInTheDocument();
  });

  it("renders placeholder content in tab area", () => {
    render(
      wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />),
    );

    expect(screen.getByTestId("jdmatch-placeholder")).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-placeholder-cta"),
    ).toBeInTheDocument();
  });

  it("renders shell data attributes", () => {
    render(
      wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />),
    );

    const root = screen.getByTestId("route-jd_match");
    expect(root.getAttribute("data-route-name")).toBe("jd_match");
  });

  it("negative — does not render old prototype business testids", () => {
    render(
      wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />),
    );

    expect(screen.queryByTestId("jdmatch-card-0")).toBeNull();
    expect(screen.queryByTestId("jdmatch-saved-search-0")).toBeNull();
    expect(screen.queryByTestId("jdmatch-watchlist-0")).toBeNull();
    expect(screen.queryByTestId("jdmatch-market-signal-0")).toBeNull();
    expect(screen.queryByTestId("jdmatch-search-bar")).toBeNull();
  });
});
