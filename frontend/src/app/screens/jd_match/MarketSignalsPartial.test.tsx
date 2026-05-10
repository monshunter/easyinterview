// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
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

function buildClient() {
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const headers = new Headers(init?.headers ?? {});
      if (url.includes("/jd-match/market-signals")) {
        headers.set("Prefer", "example=partial-data");
      }
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

function wrap(ui: ReactNode) {
  const client = buildClient();
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

beforeEach(() => {
  (window as unknown as { eiToast?: () => void }).eiToast = vi.fn();
});
afterEach(() => {
  delete (window as unknown as { eiToast?: unknown }).eiToast;
  vi.restoreAllMocks();
});

describe("MarketSignalsPartial integration (item 5.4 + 5.6)", () => {
  it("partial-data variant renders the signals it has and a fallback dash for missing d", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);
    fireEvent.click(await screen.findByTestId("jdmatch-tab-watchlist"));
    // Wait for the first card to render
    expect(
      await screen.findByTestId("jdmatch-market-signal-0"),
    ).toBeInTheDocument();
    // partial-data fixture has at least 2 signals; verify each rendered card
    // missing `d` shows a fallback dash testid.
    const fallbacks = screen.queryAllByTestId(/jdmatch-market-signal-\d+-fallback/);
    expect(fallbacks.length).toBeGreaterThan(0);
  });
});
