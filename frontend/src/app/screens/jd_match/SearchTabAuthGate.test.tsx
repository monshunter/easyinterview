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

function buildClient() {
  const registry = createFixtureRegistry([
    getJobMatchProfileFixture,
    getAgentScanStatusFixture,
    listJobRecommendationsFixture,
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
      const inner = createFixtureBackedFetch(registry, undefined);
      const headers = new Headers(init?.headers ?? {});
      if (url.includes("/me")) headers.set("Prefer", "example=unauthenticated");
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

describe("SearchTabAuthGate (item 4.5 + 4.8)", () => {
  it("Run while unauthenticated → navigate(auth_login) with action=run_search and tab=search", async () => {
    const { navigate } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
    );
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    fireEvent.change(await screen.findByTestId("jdmatch-search-input"), {
      target: { value: "frontend remote" },
    });
    fireEvent.click(screen.getByTestId("jdmatch-search-run"));
    await waitFor(() => expect(navigate).toHaveBeenCalled());
    const call = navigate.mock.calls[0]![0];
    expect(call.name).toBe("auth_login");
    const params = call.params as Record<string, string>;
    expect(params.pendingType).toBe("jd_match_action");
    expect(params.pendingRoute).toBe("jd_match");
    expect(params.tab).toBe("search");
    expect(params.action).toBe("run_search");
    // pendingAction must NOT carry private fields
    expect(params.query).toBeUndefined();
    expect(params.label).toBeUndefined();
  });

  it("Save current while unauthenticated → action=create_saved_search and label is NOT in params", async () => {
    const { navigate } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
    );
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    fireEvent.change(await screen.findByTestId("jdmatch-search-input"), {
      target: { value: "secret-search-label" },
    });
    fireEvent.click(screen.getByTestId("jdmatch-search-save-current"));
    await waitFor(() => expect(navigate).toHaveBeenCalled());
    const params = navigate.mock.calls[0]![0].params as Record<string, string>;
    expect(params.action).toBe("create_saved_search");
    expect(JSON.stringify(params)).not.toContain("secret-search-label");
  });
});
