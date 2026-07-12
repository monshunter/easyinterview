// @vitest-environment jsdom
/**
 * Plan 004 Phase 2.2 + 2.3 integration tests — Browser History routing.
 *
 * Validates that the formal frontend App shell uses Browser History as the
 * canonical route source (push / replace / popstate), keeps the out-of-scope
 * `navigate(next)` API for screens, and preserves TopBar active state +
 * chrome hidden behavior under back / forward navigation.
 */
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { FC } from "react";

import { App } from "./App";
import { useNavigation } from "./navigation/NavigationProvider";

function resetWindow(): void {
  delete (window as { __EASYINTERVIEW_INITIAL_ROUTE__?: unknown })
    .__EASYINTERVIEW_INITIAL_ROUTE__;
  window.history.replaceState(null, "", "/");
}

beforeEach(resetWindow);
afterEach(resetWindow);

const NavTrigger: FC<{
  testid: string;
  to: { name: string; params: Record<string, string> };
}> = ({ testid, to }) => {
  const { navigate } = useNavigation();
  return (
    <button type="button" data-testid={testid} onClick={() => navigate(to)}>
      go
    </button>
  );
};

describe("App browser-aware routing — Phase 2.2 navigate via History", () => {
  it("bootstraps from canonical window.location on mount (workspace empty fallback)", () => {
    window.history.replaceState(null, "", "/workspace?targetJobId=tj-bootstrap");
    render(<App />);
    expect(screen.getByTestId("workspace-plan-list")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/workspace");
    expect(window.location.search).toBe("");
  });

  it("navigate(next) updates URL via pushState and TopBar aria-current", async () => {
    render(
      <App>
        <NavTrigger
          testid="go-workspace"
          to={{ name: "workspace", params: { targetJobId: "tj-1" } }}
        />
      </App>,
    );
    const startLength = window.history.length;
    const user = userEvent.setup();
    await user.click(screen.getByTestId("go-workspace"));
    await waitFor(() =>
      expect(screen.getByTestId("topbar-nav-workspace")).toHaveAttribute(
        "aria-current",
        "page",
      ),
    );
    expect(window.location.pathname + window.location.search).toBe("/workspace");
    expect(window.history.length).toBe(startLength + 1);
  });

  it("does not double-push when navigate(next) targets the same canonical URL", async () => {
    render(
      <App>
        <NavTrigger
          testid="go-workspace"
          to={{ name: "workspace", params: { targetJobId: "tj-1" } }}
        />
      </App>,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("go-workspace"));
    const lenAfterFirst = window.history.length;
    await user.click(screen.getByTestId("go-workspace"));
    expect(window.history.length).toBe(lenAfterFirst);
  });

  it("drops unsafe params (rawText / query / label) from canonical URL even when supplied by caller", async () => {
    render(
      <App>
        <NavTrigger
          testid="go-workspace-leak"
          to={{
            name: "workspace",
            params: {
              targetJobId: "tj-1",
              rawText: "raw jd body",
              query: "secret",
            },
          }}
        />
      </App>,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("go-workspace-leak"));
    await waitFor(() =>
      expect(screen.getByTestId("workspace-plan-list")).toBeInTheDocument(),
    );
    const url = window.location.pathname + window.location.search;
    expect(url).toBe("/workspace");
    expect(url.includes("rawText")).toBe(false);
    expect(url.includes("query=secret")).toBe(false);
  });

  it("preserves TopBar active state under aria-current after navigation", async () => {
    render(
      <App>
        <NavTrigger
          testid="go-resume"
          to={{ name: "resume_versions", params: {} }}
        />
      </App>,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("go-resume"));
    await waitFor(() =>
      expect(screen.getByTestId("topbar-nav-resume_versions")).toHaveAttribute(
        "aria-current",
        "page",
      ),
    );
    expect(
      screen.getByTestId("topbar-nav-home").getAttribute("aria-current"),
    ).toBe(null);
  });

  it("canonicalizes #route=workspace bootstrap into canonical /workspace URL on mount", () => {
    window.history.replaceState(null, "", "/#route=workspace&targetJobId=tj-1");
    render(<App />);
    // The route-store mount effect must rewrite the URL to canonical form.
    expect(window.location.pathname).toBe("/workspace");
    expect(window.location.search).toBe("");
    expect(window.location.hash).toBe("");
  });
});

describe("App browser-aware routing — Phase 2.3 popstate / chrome parity", () => {
  it("popstate from a legacy phone URL renders chat with voice disabled", async () => {
    render(<App />);
    act(() => {
      window.history.pushState(
        null,
        "",
        "/practice?sessionId=01918fa0-0000-7000-8000-000000005000&mode=phone&modality=phone&planId=plan-1",
      );
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
    await waitFor(() =>
      expect(screen.getByTestId("practice-conversation")).toBeInTheDocument(),
    );
    expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-topbar-phone-toggle")).toBeDisabled();
  });

  it("popstate back to /workspace restores chrome + workspace empty fallback", async () => {
    render(<App />);
    act(() => {
      window.history.pushState(null, "", "/workspace?targetJobId=tj-1");
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
    await waitFor(() =>
      expect(screen.getByTestId("workspace-plan-list")).toBeInTheDocument(),
    );
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/workspace");
  });

  it("popstate hydrates InterviewContext / chrome-hidden state for generating route", async () => {
    render(<App />);
    act(() => {
      window.history.pushState(
        null,
        "",
        "/generating?sessionId=s-1&reportId=rpt-1",
      );
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
    await waitFor(() =>
      expect(screen.getByTestId("generating-screen")).toBeInTheDocument(),
    );
    expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();
  });

  it("popstate to unknown / out-of-scope path falls back to home", async () => {
    render(<App />);
    act(() => {
      window.history.pushState(null, "", "/totally-unknown");
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
    await waitFor(() =>
      expect(screen.getByTestId("home-hero-label")).toBeInTheDocument(),
    );
  });
});
