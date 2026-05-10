// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";

import listWatchlistFixture from "../../../../../openapi/fixtures/JobMatch/listWatchlist.json";
import getMarketSignalsFixture from "../../../../../openapi/fixtures/JobMatch/getMarketSignals.json";

import { useWatchlist, useMarketSignals } from "./useWatchlist";

function buildClient(opts: {
  watchlistScenario?: string;
  signalsScenario?: string;
}) {
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const isWatchlist = url.includes("/jd-match/watchlist");
      const isSignals = url.includes("/jd-match/market-signals");
      const scenario = isWatchlist
        ? opts.watchlistScenario
        : isSignals
          ? opts.signalsScenario
          : undefined;
      const inner = createFixtureBackedFetch(
        createFixtureRegistry([
          listWatchlistFixture,
          getMarketSignalsFixture,
        ]),
        scenario ? { scenario } : undefined,
      );
      return inner(input, init);
    },
  });
}

function withRuntime(client: EasyInterviewClient) {
  return ({ children }: { children: ReactNode }) => (
    <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>
  );
}

describe("useWatchlist (item 5.2)", () => {
  it("calls listWatchlist exactly once when active=true", async () => {
    const client = buildClient({ watchlistScenario: "default" });
    const spy = vi.spyOn(client, "listWatchlist");
    const { result } = renderHook(() => useWatchlist(true), {
      wrapper: withRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(spy).toHaveBeenCalledTimes(1);
    expect(result.current.items.length).toBeGreaterThan(0);
  });

  it("does NOT call listWatchlist when active=false", async () => {
    const client = buildClient({ watchlistScenario: "default" });
    const spy = vi.spyOn(client, "listWatchlist");
    renderHook(() => useWatchlist(false), { wrapper: withRuntime(client) });
    await new Promise((r) => setTimeout(r, 5));
    expect(spy).not.toHaveBeenCalled();
  });

  it("variant=empty → items=[] without error", async () => {
    const client = buildClient({ watchlistScenario: "empty" });
    const { result } = renderHook(() => useWatchlist(true), {
      wrapper: withRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.items).toEqual([]);
    expect(result.current.error).toBeNull();
  });

  it("variant=4xx → error is set, items=[]", async () => {
    const client = buildClient({ watchlistScenario: "4xx" });
    const { result } = renderHook(() => useWatchlist(true), {
      wrapper: withRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).not.toBeNull();
    expect(result.current.items).toEqual([]);
  });

  it("inert when no AppRuntimeProvider is mounted", () => {
    const { result } = renderHook(() => useWatchlist(true));
    expect(result.current.loading).toBe(false);
    expect(result.current.items).toEqual([]);
  });
});

describe("useMarketSignals (item 5.2)", () => {
  it("calls getMarketSignals exactly once when active=true", async () => {
    const client = buildClient({ signalsScenario: "default" });
    const spy = vi.spyOn(client, "getMarketSignals");
    const { result } = renderHook(() => useMarketSignals(true), {
      wrapper: withRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(spy).toHaveBeenCalledTimes(1);
    expect(result.current.signals.length).toBeGreaterThan(0);
  });

  it("variant=partial-data → renders partial signals", async () => {
    const client = buildClient({ signalsScenario: "partial-data" });
    const { result } = renderHook(() => useMarketSignals(true), {
      wrapper: withRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.signals.length).toBeGreaterThan(0);
    expect(result.current.error).toBeNull();
  });

  it("variant=failed → error is set, signals=[]", async () => {
    const client = buildClient({ signalsScenario: "failed" });
    const { result } = renderHook(() => useMarketSignals(true), {
      wrapper: withRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).not.toBeNull();
    expect(result.current.signals).toEqual([]);
  });
});
