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

function buildClient(scenario: "empty" | "4xx" | "default") {
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const headers = new Headers(init?.headers ?? {});
      if (
        url.includes("/jd-match/watchlist") &&
        (init?.method ?? "GET") === "GET"
      ) {
        headers.set("Prefer", `example=${scenario}`);
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

function wrap(ui: ReactNode, scenario: "empty" | "4xx" | "default") {
  const client = buildClient(scenario);
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

describe("WatchlistEmpty + WatchlistError integration (item 5.6)", () => {
  it("variant=empty renders the empty CTA", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, "empty");
    fireEvent.click(await screen.findByTestId("jdmatch-tab-watchlist"));
    expect(
      await screen.findByTestId("jdmatch-watchlist-empty"),
    ).toBeInTheDocument();
  });

  it("variant=4xx renders the watchlist error surface", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, "4xx");
    fireEvent.click(await screen.findByTestId("jdmatch-tab-watchlist"));
    expect(
      await screen.findByTestId("jdmatch-watchlist-error"),
    ).toBeInTheDocument();
  });

  it("variant=default renders watchlist rows + market signals grid + refresh footer", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, "default");
    fireEvent.click(await screen.findByTestId("jdmatch-tab-watchlist"));
    expect(
      await screen.findByTestId(
        "jdmatch-watchlist-item-01918fa0-0000-7000-8000-00000000b001",
      ),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-market-signals-grid"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-watchlist-refresh-footer"),
    ).toBeInTheDocument();
  });
});
