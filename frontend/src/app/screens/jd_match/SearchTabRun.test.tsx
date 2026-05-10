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
import searchJobsFixture from "../../../../../openapi/fixtures/JobMatch/searchJobs.json";
import listSavedSearchesFixture from "../../../../../openapi/fixtures/JobMatch/listSavedSearches.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

import { JDMatchScreen } from "./JDMatchScreen";

function buildClient(opts: {
  searchScenario?: string;
  fetchSpy?: ReturnType<typeof vi.fn>;
}) {
  const inner = createFixtureBackedFetch(
    createFixtureRegistry([
      getJobMatchProfileFixture,
      getAgentScanStatusFixture,
      listJobRecommendationsFixture,
      searchJobsFixture,
      listSavedSearchesFixture,
      getMeFixture,
      getRuntimeConfigFixture,
    ]),
  );
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const method = init?.method ?? "GET";
      const headers = new Headers(init?.headers ?? {});
      const isSearch =
        url.includes("/jd-match/search") && method === "POST";
      if (isSearch && opts.searchScenario)
        headers.set("Prefer", `example=${opts.searchScenario}`);
      let bodyText: string | null = null;
      if (init?.body && typeof init.body === "string") bodyText = init.body;
      opts.fetchSpy?.({
        url,
        method,
        headers: Object.fromEntries(headers),
        body: bodyText,
      });
      return inner(input, { ...init, headers });
    },
  });
}

function wrap(ui: ReactNode, opts: typeof defaultOpts = defaultOpts) {
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

const defaultOpts: { searchScenario?: string; fetchSpy?: ReturnType<typeof vi.fn> } = {};

beforeEach(() => {
  (window as unknown as { eiToast?: () => void }).eiToast = vi.fn();
});
afterEach(() => {
  delete (window as unknown as { eiToast?: unknown }).eiToast;
  vi.restoreAllMocks();
});

describe("SearchTabRun integration (item 4.2 + 4.8)", () => {
  it("Run dispatch sends searchJobs body with query + Idempotency-Key", async () => {
    const fetchSpy = vi.fn();
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      searchScenario: "default",
      fetchSpy,
    });
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    const input = await screen.findByTestId("jdmatch-search-input");
    fireEvent.change(input, {
      target: {
        value: "Senior frontend roles with strong design-system culture",
      },
    });
    fireEvent.click(screen.getByTestId("jdmatch-search-run"));
    await waitFor(() =>
      expect(screen.getByTestId("jdmatch-search-results")).toBeInTheDocument(),
    );
    const searchCall = fetchSpy.mock.calls.find(
      ([req]) =>
        req.method === "POST" && req.url.includes("/jd-match/search"),
    );
    expect(searchCall).toBeTruthy();
    expect(searchCall![0].headers["idempotency-key"]).toBeTruthy();
    const body = JSON.parse(searchCall![0].body) as Record<string, unknown>;
    expect(body.query).toBe(
      "Senior frontend roles with strong design-system culture",
    );
  });

  it("AGENT scanning panel renders 5-step DOM while a slow search is in-flight", async () => {
    const resolveRef: { current: ((res: Response) => void) | null } = {
      current: null,
    };
    const slowFetch = vi.fn(
      () =>
        new Promise<Response>((r) => {
          resolveRef.current = r;
        }),
    );
    const innerFetch = createFixtureBackedFetch(
      createFixtureRegistry([
        getJobMatchProfileFixture,
        getAgentScanStatusFixture,
        listJobRecommendationsFixture,
        listSavedSearchesFixture,
        getMeFixture,
        getRuntimeConfigFixture,
      ]),
    );
    const client = new EasyInterviewClient({
      fetch: async (input, init) => {
        const url =
          typeof input === "string"
            ? input
            : (input as URL | Request).toString();
        if (
          url.includes("/jd-match/search") &&
          (init?.method ?? "GET") === "POST"
        ) {
          return slowFetch();
        }
        return innerFetch(input, init);
      },
    });
    const navigate = vi.fn();
    render(
      <DisplayPreferencesProvider initial={{ lang: "en" }}>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate }}>
            <JDMatchScreen route={{ name: "jd_match", params: {} }} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    );
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    const input = await screen.findByTestId("jdmatch-search-input");
    fireEvent.change(input, { target: { value: "platform" } });
    fireEvent.click(screen.getByTestId("jdmatch-search-run"));
    expect(
      await screen.findByTestId("jdmatch-search-searching-panel"),
    ).toBeInTheDocument();
    for (let i = 1; i <= 5; i++) {
      expect(
        screen.getByTestId(`jdmatch-search-searching-step-${i}`),
      ).toBeInTheDocument();
    }
    // Resolve to flush effects
    resolveRef.current?.(
      new Response(JSON.stringify({ items: [], searchRunId: "x" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
  });

  it("Run is disabled while searching=true and Enter key is a no-op", async () => {
    const resolveRef: { current: ((res: Response) => void) | null } = {
      current: null,
    };
    const slowFetch = vi.fn(
      () =>
        new Promise<Response>((r) => {
          resolveRef.current = r;
        }),
    );
    const innerFetch = createFixtureBackedFetch(
      createFixtureRegistry([
        getJobMatchProfileFixture,
        getAgentScanStatusFixture,
        listJobRecommendationsFixture,
        listSavedSearchesFixture,
        getMeFixture,
        getRuntimeConfigFixture,
      ]),
    );
    const client = new EasyInterviewClient({
      fetch: async (input, init) => {
        const url =
          typeof input === "string"
            ? input
            : (input as URL | Request).toString();
        if (
          url.includes("/jd-match/search") &&
          (init?.method ?? "GET") === "POST"
        ) {
          return slowFetch();
        }
        return innerFetch(input, init);
      },
    });
    const navigate = vi.fn();
    render(
      <DisplayPreferencesProvider initial={{ lang: "en" }}>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate }}>
            <JDMatchScreen route={{ name: "jd_match", params: {} }} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    );
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    const input = await screen.findByTestId("jdmatch-search-input");
    fireEvent.change(input, { target: { value: "test" } });
    fireEvent.click(screen.getByTestId("jdmatch-search-run"));
    await waitFor(() =>
      expect(
        (screen.getByTestId("jdmatch-search-run") as HTMLButtonElement).disabled,
      ).toBe(true),
    );
    // Enter while searching should not trigger another call (search is gated by searching=true).
    fireEvent.keyDown(input, { key: "Enter", code: "Enter" });
    // Resolve to flush effects
    resolveRef.current?.(
      new Response(
        JSON.stringify({ items: [], searchRunId: "run-x" }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );
    await waitFor(() =>
      expect(
        (screen.getByTestId("jdmatch-search-run") as HTMLButtonElement).disabled,
      ).toBe(false),
    );
    expect(slowFetch).toHaveBeenCalledTimes(1);
  });

  it("Switching to another tab clears query and aborts in-flight search", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      searchScenario: "slow-response",
    });
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    const input = await screen.findByTestId("jdmatch-search-input");
    fireEvent.change(input, { target: { value: "to-be-cleared" } });
    fireEvent.click(screen.getByTestId("jdmatch-search-run"));
    expect(
      await screen.findByTestId("jdmatch-search-searching-panel"),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByTestId("jdmatch-tab-recommended"));
    fireEvent.click(screen.getByTestId("jdmatch-tab-search"));
    const inputAgain = await screen.findByTestId("jdmatch-search-input");
    expect((inputAgain as HTMLInputElement).value).toBe("");
    expect(
      screen.queryByTestId("jdmatch-search-searching-panel"),
    ).not.toBeInTheDocument();
  });
});
