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
});
