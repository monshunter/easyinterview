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
import searchJobsFixture from "../../../../../openapi/fixtures/JobMatch/searchJobs.json";
import listSavedSearchesFixture from "../../../../../openapi/fixtures/JobMatch/listSavedSearches.json";
import createSavedSearchFixture from "../../../../../openapi/fixtures/JobMatch/createSavedSearch.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

import { JDMatchScreen } from "./JDMatchScreen";

const SECRET_QUERY = "ultra-private-search-query-do-not-leak";
const SECRET_SAVED_LABEL = "ultra-private-saved-label-do-not-leak";

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
      const inner = createFixtureBackedFetch(registry, undefined);
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

function expectNoLeak(...tokens: string[]): void {
  for (const spy of [logSpy, errSpy, warnSpy]) {
    for (const call of spy.mock.calls) {
      const text = call.map((v) => (typeof v === "string" ? v : "")).join(" ");
      for (const token of tokens) {
        expect(text.includes(token), `console leaked ${token}: ${text}`).toBe(
          false,
        );
      }
    }
  }
  for (const call of setItemSpy.mock.calls) {
    const value = String(call[1]);
    for (const token of tokens) {
      expect(
        value.includes(token),
        `localStorage.setItem leaked ${token}`,
      ).toBe(false);
    }
  }
  const url =
    window.location.href +
    (window.location.search ?? "") +
    (window.location.hash ?? "");
  for (const token of tokens) {
    expect(url.includes(token), `URL leaked ${token}`).toBe(false);
  }
}

describe("SearchTabPrivacy (item 4.6 + 4.8)", () => {
  it("Run search does NOT leak query through console / URL / localStorage", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    fireEvent.change(await screen.findByTestId("jdmatch-search-input"), {
      target: { value: SECRET_QUERY },
    });
    fireEvent.click(screen.getByTestId("jdmatch-search-run"));
    await waitFor(() =>
      expect(
        (screen.getByTestId("jdmatch-search-run") as HTMLButtonElement).disabled,
      ).toBe(false),
    );
    expectNoLeak(SECRET_QUERY);
  });

  it("Save current does NOT leak label / query through console / URL / localStorage", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    fireEvent.change(await screen.findByTestId("jdmatch-search-input"), {
      target: { value: SECRET_SAVED_LABEL },
    });
    fireEvent.click(screen.getByTestId("jdmatch-search-save-current"));
    await waitFor(() => expect(setItemSpy.mock.calls.length >= 0).toBe(true));
    expectNoLeak(SECRET_SAVED_LABEL);
  });

  it("Switching tab away from Search clears query so it does not survive in the input box", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);
    fireEvent.click(await screen.findByTestId("jdmatch-tab-search"));
    fireEvent.change(await screen.findByTestId("jdmatch-search-input"), {
      target: { value: SECRET_QUERY },
    });
    fireEvent.click(screen.getByTestId("jdmatch-tab-recommended"));
    fireEvent.click(screen.getByTestId("jdmatch-tab-search"));
    const input = await screen.findByTestId("jdmatch-search-input");
    expect((input as HTMLInputElement).value).toBe("");
    expectNoLeak(SECRET_QUERY);
  });
});
