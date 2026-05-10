// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
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

import getJobMatchProfileFixture from "../../../../../openapi/fixtures/JobMatch/getJobMatchProfile.json";
import getAgentScanStatusFixture from "../../../../../openapi/fixtures/JobMatch/getAgentScanStatus.json";
import listJobRecommendationsFixture from "../../../../../openapi/fixtures/JobMatch/listJobRecommendations.json";
import addToWatchlistFixture from "../../../../../openapi/fixtures/JobMatch/addToWatchlist.json";
import searchJobsFixture from "../../../../../openapi/fixtures/JobMatch/searchJobs.json";
import listSavedSearchesFixture from "../../../../../openapi/fixtures/JobMatch/listSavedSearches.json";
import createSavedSearchFixture from "../../../../../openapi/fixtures/JobMatch/createSavedSearch.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

import {
  clearPendingJdMatchActionsForTests,
  storePendingJdMatchAction,
} from "./pendingJdMatchActionState";
import { JDMatchScreen } from "./JDMatchScreen";

const SELECTED_ID = "01918fa0-0000-7000-8000-00000000a001";

function buildClient(opts: { signedIn: boolean }) {
  const registry = createFixtureRegistry([
    getJobMatchProfileFixture,
    getAgentScanStatusFixture,
    listJobRecommendationsFixture,
    addToWatchlistFixture,
    searchJobsFixture,
    listSavedSearchesFixture,
    createSavedSearchFixture,
    getMeFixture,
    getRuntimeConfigFixture,
  ]);
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const headers = new Headers(init?.headers ?? {});
      if (url.includes("/me")) {
        headers.set(
          "Prefer",
          opts.signedIn ? "example=authenticated" : "example=unauthenticated",
        );
      }
      const inner = createFixtureBackedFetch(registry, undefined);
      return inner(input, { ...init, headers });
    },
  });
}

function wrap(ui: ReactNode, client: EasyInterviewClient) {
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
  clearPendingJdMatchActionsForTests();
  (window as unknown as { eiToast?: () => void }).eiToast = vi.fn();
});

afterEach(() => {
  clearPendingJdMatchActionsForTests();
  delete (window as unknown as { eiToast?: unknown }).eiToast;
  vi.restoreAllMocks();
});

describe("JDMatchScreen pending action auto-resume (item 4.10)", () => {
  it("authenticated route params restore Recommended selection and auto-run save once", async () => {
    const client = buildClient({ signedIn: true });
    const addSpy = vi.spyOn(client, "addToWatchlist");

    wrap(
      <JDMatchScreen
        route={{
          name: "jd_match",
          params: {
            tab: "recommended",
            selectedJobMatchId: SELECTED_ID,
            action: "save",
          },
        }}
      />,
      client,
    );

    await waitFor(() =>
      expect(addSpy).toHaveBeenCalledWith(
        { jobMatchId: SELECTED_ID },
        expect.objectContaining({ idempotencyKey: expect.any(String) }),
      ),
    );
    expect(screen.getByTestId(`jdmatch-card-${SELECTED_ID}`)).toHaveAttribute(
      "data-active",
      "true",
    );
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(addSpy).toHaveBeenCalledTimes(1);
  });

  it("authenticated Search route consumes opaque pending payload and auto-runs without leaking query in params or storage", async () => {
    const secretQuery = "secret frontend remote";
    const pendingJdMatchActionId = storePendingJdMatchAction({
      action: "run_search",
      query: secretQuery,
    });
    const client = buildClient({ signedIn: true });
    const searchSpy = vi.spyOn(client, "searchJobs");

    wrap(
      <JDMatchScreen
        route={{
          name: "jd_match",
          params: {
            tab: "search",
            action: "run_search",
            pendingJdMatchActionId,
          },
        }}
      />,
      client,
    );

    await waitFor(() =>
      expect(searchSpy).toHaveBeenCalledWith(
        { query: secretQuery, filters: undefined },
        expect.objectContaining({ idempotencyKey: expect.any(String) }),
      ),
    );
    expect(screen.getByTestId("jdmatch-search-input")).toHaveValue(secretQuery);
    expect(window.location.href).not.toContain(secretQuery);
    expect(JSON.stringify(window.localStorage)).not.toContain(secretQuery);
  });
});
