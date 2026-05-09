// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { TopBar } from "../../topbar/TopBar";

import getAgentScanStatusFixture from "../../../../../openapi/fixtures/JobMatch/getAgentScanStatus.json";
import getJobMatchProfileFixture from "../../../../../openapi/fixtures/JobMatch/getJobMatchProfile.json";

import { JDMatchScreen } from "./JDMatchScreen";

function buildClient(opts?: {
  profileScenario?: string;
  agentScenario?: string;
}) {
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const isAgent = url.includes("/jd-match/agent-status");
      const scenario = isAgent ? opts?.agentScenario : opts?.profileScenario;
      const inner = createFixtureBackedFetch(
        createFixtureRegistry([
          getJobMatchProfileFixture,
          getAgentScanStatusFixture,
        ]),
        scenario ? { scenario } : undefined,
      );
      return inner(input, init);
    },
  });
}

function wrap(ui: ReactNode, opts?: { client?: EasyInterviewClient; lang?: "zh" | "en" }) {
  const navigate = vi.fn();
  const tree = (
    <DisplayPreferencesProvider initial={opts?.lang ? { lang: opts.lang } : undefined}>
      {opts?.client ? (
        <AppRuntimeProvider client={opts.client}>
          <NavigationProvider value={{ navigate }}>{ui}</NavigationProvider>
        </AppRuntimeProvider>
      ) : (
        <NavigationProvider value={{ navigate }}>{ui}</NavigationProvider>
      )}
    </DisplayPreferencesProvider>
  );
  return { navigate, ...render(tree) };
}

describe("JDMatchScreen — screen-level shell + data-driven contract (item 2.5)", () => {
  it("renders Hero label / title / sub testids", () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);

    expect(screen.getByTestId("jdmatch-hero-label")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-hero-title")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-hero-sub")).toBeInTheDocument();
  });

  it("renders Profile snapshot chip with avatar / searching-as / skills / sources slots", () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);

    expect(screen.getByTestId("jdmatch-profile-chip")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-profile-chip-avatar")).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-profile-chip-searching-as"),
    ).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-profile-chip-skills")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-profile-chip-sources")).toBeInTheDocument();
  });

  it("renders three tab testids (recommended / search / watchlist)", () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);

    expect(screen.getByTestId("jdmatch-tab-recommended")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-tab-search")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-tab-watchlist")).toBeInTheDocument();
  });

  it("exposes route shell data attributes for topbar / nav resolution", () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: { source: "home" } }} />);

    const root = screen.getByTestId("route-jd_match");
    expect(root.getAttribute("data-route-name")).toBe("jd_match");
    expect(root.getAttribute("data-route-params")).toBe(
      JSON.stringify({ source: "home" }),
    );
  });

  it("renders zero plan-001 placeholder testids in DOM (negative regression)", () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);

    expect(screen.queryByTestId("jdmatch-placeholder")).toBeNull();
    expect(screen.queryByTestId("jdmatch-placeholder-cta")).toBeNull();
  });

  it("renders zero legacy prototype business testids (jdmatch-card-* / saved-search-* / watchlist-* / market-signal-* / search-bar)", () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);

    expect(screen.queryByTestId("jdmatch-card-0")).toBeNull();
    expect(screen.queryByTestId("jdmatch-saved-search-0")).toBeNull();
    expect(screen.queryByTestId("jdmatch-watchlist-0")).toBeNull();
    expect(screen.queryByTestId("jdmatch-market-signal-0")).toBeNull();
    expect(screen.queryByTestId("jdmatch-search-bar")).toBeNull();
  });

  it("Profile chip pulls live data from getJobMatchProfile (default fixture: avatar img + 6 skills + sources counts)", async () => {
    const client = buildClient({
      profileScenario: "default",
      agentScenario: "default",
    });
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, { client });

    const avatar = await screen.findByTestId("jdmatch-profile-chip-avatar");
    await waitFor(() => {
      expect(avatar.querySelector("img")).not.toBeNull();
    });
    const skills = screen.getByTestId("jdmatch-profile-chip-skills");
    const skillTags = skills.querySelectorAll("[data-testid^='jdmatch-profile-chip-skill-']");
    expect(skillTags.length).toBe(6);
    const sources = screen.getByTestId("jdmatch-profile-chip-sources");
    expect(sources.textContent ?? "").toMatch(/2.*5.*4.*1/);
  });

  it("Profile chip falls back to initials when avatarUrl is null (partial-profile fixture)", async () => {
    const client = buildClient({
      profileScenario: "partial-profile",
      agentScenario: "default",
    });
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, { client });

    const avatar = await screen.findByTestId("jdmatch-profile-chip-avatar");
    await waitFor(() => {
      expect(avatar.querySelector("img")).toBeNull();
      expect(avatar.textContent ?? "").toMatch(/AE/);
    });
  });

  it("AGENT badge renders idle / scanning / error tones from fixture variants", async () => {
    for (const scenario of ["idle", "scanning", "error"] as const) {
      const client = buildClient({
        profileScenario: "default",
        agentScenario: scenario,
      });
      const { unmount } = wrap(
        <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
        { client },
      );

      const badge = await screen.findByTestId("jdmatch-agent-status-badge");
      await waitFor(() => {
        expect(badge.getAttribute("data-tone")).toBe(scenario);
      });
      if (scenario === "scanning") {
        expect(screen.queryByTestId("jdmatch-agent-status-next-scan")).toBeNull();
      }
      unmount();
    }
  });

  it("switches Hero text between zh and en when DisplayPreferences.lang changes", () => {
    const { unmount } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { lang: "zh" },
    );
    const zhTitle = screen.getByTestId("jdmatch-hero-title").textContent ?? "";
    expect(zhTitle).toMatch(/我们读市场/);
    unmount();

    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      lang: "en",
    });
    const enTitle = screen.getByTestId("jdmatch-hero-title").textContent ?? "";
    expect(enTitle).toMatch(/We read the market/);
  });

  it("topbar-nav-jd_match becomes aria-current=page when active route is jd_match", () => {
    render(
      <DisplayPreferencesProvider>
        <TopBar
          activeRoute="jd_match"
          onNavigate={() => {}}
          signedIn={false}
        />
      </DisplayPreferencesProvider>,
    );

    expect(screen.getByTestId("topbar-nav-jd_match")).toHaveAttribute(
      "aria-current",
      "page",
    );
  });

  it("does NOT register setInterval / EventSource / WebSocket on mount (D-10)", async () => {
    const client = buildClient({
      profileScenario: "default",
      agentScenario: "default",
    });
    const intervalSpy = vi.spyOn(window, "setInterval");
    const eventSourceSpy = vi.fn();
    const webSocketSpy = vi.fn();
    const originalES = (globalThis as { EventSource?: unknown }).EventSource;
    const originalWS = (globalThis as { WebSocket?: unknown }).WebSocket;
    Object.defineProperty(globalThis, "EventSource", {
      configurable: true,
      writable: true,
      value: eventSourceSpy,
    });
    Object.defineProperty(globalThis, "WebSocket", {
      configurable: true,
      writable: true,
      value: webSocketSpy,
    });

    try {
      wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, { client });
      // settle the hook effect chain without using waitFor (which polls via setInterval)
      await new Promise((r) => setTimeout(r, 50));
      expect(intervalSpy).not.toHaveBeenCalled();
      expect(eventSourceSpy).not.toHaveBeenCalled();
      expect(webSocketSpy).not.toHaveBeenCalled();
    } finally {
      intervalSpy.mockRestore();
      Object.defineProperty(globalThis, "EventSource", {
        configurable: true,
        writable: true,
        value: originalES,
      });
      Object.defineProperty(globalThis, "WebSocket", {
        configurable: true,
        writable: true,
        value: originalWS,
      });
    }
  });
});
