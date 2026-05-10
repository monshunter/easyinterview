// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import type { MockInstance } from "vitest";
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

const SECRET_TOKENS = [
  "01918fa0-0000-7000-8000-00000000b001",
  "01918fa0-0000-7000-8000-00000000b002",
  "01918fa0-0000-7000-8000-00000000a001",
  "01918fa0-0000-7000-8000-00000000a002",
];

function buildClient() {
  return new EasyInterviewClient({
    fetch: async (input, init) => {
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
      return inner(input, init);
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

let logSpy: MockInstance<typeof console.log>;
let errSpy: MockInstance<typeof console.error>;
let warnSpy: MockInstance<typeof console.warn>;
let setItemSpy: MockInstance<typeof Storage.prototype.setItem>;

beforeEach(() => {
  logSpy = vi.spyOn(console, "log").mockImplementation(() => undefined);
  errSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);
  warnSpy = vi.spyOn(console, "warn").mockImplementation(() => undefined);
  setItemSpy = vi.spyOn(Storage.prototype, "setItem");
  (window as unknown as { eiToast?: () => void }).eiToast = vi.fn();
});

afterEach(() => {
  delete (window as unknown as { eiToast?: unknown }).eiToast;
  vi.restoreAllMocks();
});

describe("WatchlistPrivacy (item 5.5 + 5.6)", () => {
  it("Watchlist tab + chevron handoff does not leak linkedJobMatchId / labels through console / URL / localStorage", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);
    fireEvent.click(await screen.findByTestId("jdmatch-tab-watchlist"));
    const chevron = await screen.findByTestId(
      "jdmatch-watchlist-item-01918fa0-0000-7000-8000-00000000b002-chevron",
    );
    fireEvent.click(chevron);
    await waitFor(() =>
      expect(screen.getByTestId("jdmatch-recommended-tab")).toBeInTheDocument(),
    );
    for (const spy of [logSpy, errSpy, warnSpy]) {
      for (const call of spy.mock.calls) {
        const text = call.map((v) => (typeof v === "string" ? v : "")).join(" ");
        for (const token of SECRET_TOKENS) {
          expect(text.includes(token), `console leaked ${token}`).toBe(false);
        }
      }
    }
    for (const call of setItemSpy.mock.calls) {
      for (const token of SECRET_TOKENS) {
        expect(
          String(call[1]).includes(token),
          `localStorage.setItem leaked ${token}`,
        ).toBe(false);
      }
    }
    const url =
      window.location.href +
      (window.location.search ?? "") +
      (window.location.hash ?? "");
    for (const token of SECRET_TOKENS) {
      expect(url.includes(token), `URL leaked ${token}`).toBe(false);
    }
  });
});
