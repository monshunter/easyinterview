// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";

import getJobMatchProfileFixture from "../../../../../openapi/fixtures/JobMatch/getJobMatchProfile.json";
import getAgentScanStatusFixture from "../../../../../openapi/fixtures/JobMatch/getAgentScanStatus.json";
import listJobRecommendationsFixture from "../../../../../openapi/fixtures/JobMatch/listJobRecommendations.json";
import listWatchlistFixture from "../../../../../openapi/fixtures/JobMatch/listWatchlist.json";
import getMarketSignalsFixture from "../../../../../openapi/fixtures/JobMatch/getMarketSignals.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

import { JDMatchScreen } from "./JDMatchScreen";

function buildClient(opts: {
  recommendationsScenario?: string;
  watchlistScenario?: string;
}) {
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const method = init?.method ?? "GET";
      const headers = new Headers(init?.headers ?? {});
      const isRecsList =
        url.includes("/jd-match/recommendations") &&
        method === "GET" &&
        !url.match(/\/jd-match\/recommendations\/[^/?]+/);
      const isWatchlist =
        url.includes("/jd-match/watchlist") && method === "GET";
      if (isRecsList && opts.recommendationsScenario)
        headers.set("Prefer", `example=${opts.recommendationsScenario}`);
      if (isWatchlist && opts.watchlistScenario)
        headers.set("Prefer", `example=${opts.watchlistScenario}`);
      const inner = createFixtureBackedFetch(
        createFixtureRegistry([
          getJobMatchProfileFixture,
          getAgentScanStatusFixture,
          listJobRecommendationsFixture,
          listWatchlistFixture,
          getMarketSignalsFixture,
          getMeFixture,
          getRuntimeConfigFixture,
        ]),
      );
      return inner(input, { ...init, headers });
    },
  });
}

function wrap(
  ui: ReactNode,
  opts: { recommendationsScenario?: string; watchlistScenario?: string } = {},
) {
  const client = buildClient(opts);
  const navigate = vi.fn();
  const tree = (
    <DisplayPreferencesProvider initial={{ lang: "en" }}>
      <AppRuntimeProvider client={client}>
        <NavigationProvider value={{ navigate }}>{ui}</NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>
  );
  return { navigate, ...render(tree) };
}

let toastSpy: ReturnType<typeof vi.fn>;
beforeEach(() => {
  toastSpy = vi.fn();
  (window as unknown as { eiToast?: typeof toastSpy }).eiToast = toastSpy;
});
afterEach(() => {
  delete (window as unknown as { eiToast?: unknown }).eiToast;
  vi.restoreAllMocks();
});

describe("WatchlistChevron integration (item 5.3 + 5.6)", () => {
  it("chevron click on watchlist item with linked recommendation switches to Recommended tab and selects target", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      recommendationsScenario: "default",
      watchlistScenario: "default",
    });
    fireEvent.click(await screen.findByTestId("jdmatch-tab-watchlist"));
    // first watchlist row links to jm a002 (Staff FE Lumen Labs)
    const chevron = await screen.findByTestId(
      "jdmatch-watchlist-item-01918fa0-0000-7000-8000-00000000b002-chevron",
    );
    fireEvent.click(chevron);
    await waitFor(() =>
      expect(screen.getByTestId("jdmatch-recommended-tab")).toBeInTheDocument(),
    );
    expect(screen.getByTestId("jdmatch-detail-header")).toHaveTextContent(
      "Staff",
    );
  });

  it("chevron handoff dispatches a warn toast and falls back to first card when linkedJobMatchId is not in the list", async () => {
    // recommendations.one only has jm-a001; watchlist.few links jm-a001 only.
    // Force a mismatch by using listJobRecommendations.empty (no recommendations) + listWatchlist.few.
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      recommendationsScenario: "empty",
      watchlistScenario: "few",
    });
    fireEvent.click(await screen.findByTestId("jdmatch-tab-watchlist"));
    const chevron = await screen.findByTestId(
      "jdmatch-watchlist-item-01918fa0-0000-7000-8000-00000000b001-chevron",
    );
    fireEvent.click(chevron);
    await waitFor(() => {
      expect(toastSpy).toHaveBeenCalled();
    });
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("warn");
    // Recommended tab is now active even if no card to show
    expect(screen.getByTestId("jdmatch-recommended-tab")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-recommended-empty")).toBeInTheDocument();
  });
});
