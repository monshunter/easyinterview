// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import { App } from "./App";

describe("App shell", () => {
  it("defaults to the home route with App chrome rendered", () => {
    render(<App />);
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(screen.getByTestId("route-home")).toBeInTheDocument();
  });

  it("keeps chrome rendered for context routes (parse / report / company_intel)", () => {
    const contextRoutes = ["parse", "report", "company_intel"] as const;
    for (const name of contextRoutes) {
      const { unmount } = render(<App initialRoute={{ name, params: {} }} />);
      expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
      expect(screen.getByTestId(`route-${name}`)).toBeInTheDocument();
      unmount();
    }
  });

  it("hides chrome for immersive practice / generating routes", () => {
    const immersiveRoutes = ["practice", "generating"] as const;
    for (const name of immersiveRoutes) {
      const { unmount } = render(<App initialRoute={{ name, params: {} }} />);
      expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();
      expect(screen.getByTestId(`route-${name}`)).toBeInTheDocument();
      unmount();
    }
  });

  it("renders HomeScreen on the home route instead of PlaceholderScreen", () => {
    render(<App />);
    expect(screen.getByTestId("home-hero-label")).toBeInTheDocument();
    expect(screen.getByTestId("home-hero-title")).toBeInTheDocument();
    expect(screen.getByTestId("home-jd-textarea")).toBeInTheDocument();
  });

  it("renders ParseScreen on the parse route instead of PlaceholderScreen", () => {
    render(
      <App
        initialRoute={{
          name: "parse",
          params: { targetJobId: "01918fa0-0000-7000-8000-000000002000" },
        }}
      />,
    );
    expect(screen.getByTestId("parse-loading-step-0")).toBeInTheDocument();
    expect(screen.queryByText("D2-D6")).not.toBeInTheDocument();
  });

  it("propagates route params to the rendered route view", () => {
    render(
      <App
        initialRoute={{
          name: "practice",
          params: { mode: "voice", planId: "plan-tj-1" },
        }}
      />,
    );
    const view = screen.getByTestId("route-practice");
    expect(view).toHaveAttribute(
      "data-route-params",
      JSON.stringify({ mode: "voice", planId: "plan-tj-1" }),
    );
  });

  it("renders WorkspaceScreen on workspace route instead of PlaceholderScreen", () => {
    render(
      <App
        initialRoute={{
          name: "workspace",
          params: { targetJobId: "tj-1" },
        }}
      />,
    );
    expect(screen.getByTestId("workspace-crumbs")).toBeInTheDocument();
    expect(screen.queryByTestId("route-workspace")).not.toBeInTheDocument();
  });

  it("practice route still renders PlaceholderScreen", () => {
    render(
      <App
        initialRoute={{
          name: "practice",
          params: { sessionId: "sess-1" },
        }}
      />,
    );
    expect(screen.getByTestId("route-practice")).toBeInTheDocument();
  });

  it("generating route still renders PlaceholderScreen", () => {
    render(
      <App
        initialRoute={{
          name: "generating",
          params: { sessionId: "sess-1", reportId: "rpt-1" },
        }}
      />,
    );
    expect(screen.getByTestId("route-generating")).toBeInTheDocument();
  });
});
