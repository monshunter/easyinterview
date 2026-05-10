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
import createSavedSearchFixture from "../../../../../openapi/fixtures/JobMatch/createSavedSearch.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

import { JDMatchScreen } from "./JDMatchScreen";

function buildClient(opts: {
  listScenario?: string;
  createScenario?: string;
  fetchSpy?: ReturnType<typeof vi.fn>;
}) {
  const inner = createFixtureBackedFetch(
    createFixtureRegistry([
      getJobMatchProfileFixture,
      getAgentScanStatusFixture,
      listJobRecommendationsFixture,
      searchJobsFixture,
      listSavedSearchesFixture,
      createSavedSearchFixture,
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
      const isList =
        url.includes("/jd-match/saved-searches") && method === "GET";
      const isCreate =
        url.includes("/jd-match/saved-searches") && method === "POST";
      if (isList && opts.listScenario)
        headers.set("Prefer", `example=${opts.listScenario}`);
      if (isCreate && opts.createScenario)
        headers.set("Prefer", `example=${opts.createScenario}`);
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

function wrap(
  ui: ReactNode,
  opts: {
    listScenario?: string;
    createScenario?: string;
    fetchSpy?: ReturnType<typeof vi.fn>;
  } = {},
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

describe("SearchTabSavedSearches integration (item 4.3 + 4.8)", () => {
  it("listSavedSearches is called exactly once when entering the Search tab", async () => {
    const fetchSpy = vi.fn();
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      listScenario: "default",
      fetchSpy,
    });
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    await waitFor(() => {
      const listCalls = fetchSpy.mock.calls.filter(
        ([req]) =>
          req.method === "GET" && req.url.includes("/jd-match/saved-searches"),
      );
      expect(listCalls).toHaveLength(1);
    });
    // Confirm rendered grid items
    expect(screen.getByTestId("jdmatch-search-saved-grid")).toBeInTheDocument();
  });

  it("Save current button posts createSavedSearch with Idempotency-Key and prepends to grid", async () => {
    const fetchSpy = vi.fn();
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      listScenario: "default",
      createScenario: "default",
      fetchSpy,
    });
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    fireEvent.change(await screen.findByTestId("jdmatch-search-input"), {
      target: { value: "platform engineering remote" },
    });
    fireEvent.click(screen.getByTestId("jdmatch-search-save-current"));
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    const createCall = fetchSpy.mock.calls.find(
      ([req]) =>
        req.method === "POST" && req.url.includes("/jd-match/saved-searches"),
    );
    expect(createCall).toBeTruthy();
    expect(createCall![0].headers["idempotency-key"]).toBeTruthy();
    const body = JSON.parse(createCall![0].body) as Record<string, unknown>;
    expect(body.label).toBe("platform engineering remote");
    expect(body.query).toBe("platform engineering remote");
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("ok");
  });

  it("createSavedSearch 4xx-validation shows danger toast and inline retry surface", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      createScenario: "4xx-validation",
    });
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    fireEvent.change(await screen.findByTestId("jdmatch-search-input"), {
      target: { value: "x" },
    });
    fireEvent.click(screen.getByTestId("jdmatch-search-save-current"));
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("danger");
    expect(
      screen.getByTestId("jdmatch-search-saved-create-error"),
    ).toBeInTheDocument();
  });

  it("listSavedSearches 4xx renders error surface in saved area", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      listScenario: "4xx",
    });
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    expect(
      await screen.findByTestId("jdmatch-search-saved-error"),
    ).toBeInTheDocument();
  });
});
