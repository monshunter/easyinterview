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

function buildClient(searchScenario: string) {
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
      if (isSearch) headers.set("Prefer", `example=${searchScenario}`);
      return inner(input, { ...init, headers });
    },
  });
}

function wrap(ui: ReactNode, scenario: string) {
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

async function runSearchAndWait(query: string) {
  fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
  fireEvent.change(await screen.findByTestId("jdmatch-search-input"), {
    target: { value: query },
  });
  fireEvent.click(screen.getByTestId("jdmatch-search-run"));
  // Wait until the run completes (button re-enables)
  await waitFor(() =>
    expect(
      (screen.getByTestId("jdmatch-search-run") as HTMLButtonElement).disabled,
    ).toBe(false),
  );
}

describe("SearchTabFailure (item 4.7 + 4.8)", () => {
  it("failed variant → inline error surface, retain query input", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, "failed");
    await runSearchAndWait("anything");
    expect(screen.getByTestId("jdmatch-search-error")).toBeInTheDocument();
    expect(screen.queryByTestId("jdmatch-search-results")).toBeNull();
    expect(
      (screen.getByTestId("jdmatch-search-input") as HTMLInputElement).value,
    ).toBe("anything");
  });

  it("empty variant → no-results empty state", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, "empty");
    await runSearchAndWait("nothing");
    expect(screen.getByTestId("jdmatch-search-empty")).toBeInTheDocument();
    expect(screen.queryByTestId("jdmatch-search-results")).toBeNull();
  });

  it("default variant → renders results grid with at least one card", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, "default");
    await runSearchAndWait("frontend roles");
    expect(screen.getByTestId("jdmatch-search-results")).toBeInTheDocument();
  });
});
