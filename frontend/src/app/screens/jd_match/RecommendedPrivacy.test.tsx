// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { MockInstance } from "vitest";
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
import addToWatchlistFixture from "../../../../../openapi/fixtures/JobMatch/addToWatchlist.json";
import removeFromWatchlistFixture from "../../../../../openapi/fixtures/JobMatch/removeFromWatchlist.json";
import markJobNotRelevantFixture from "../../../../../openapi/fixtures/JobMatch/markJobNotRelevant.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

import { JDMatchScreen } from "./JDMatchScreen";

const SECRET_TOKENS = [
  "01918fa0-0000-7000-8000-00000000a001",
  "https://acme.example",
  "acme.example/careers/senior-frontend",
];

function buildClient(transportSpy: ReturnType<typeof vi.fn>) {
  const inner = createFixtureBackedFetch(
    createFixtureRegistry([
      getJobMatchProfileFixture,
      getAgentScanStatusFixture,
      listJobRecommendationsFixture,
      addToWatchlistFixture,
      removeFromWatchlistFixture,
      markJobNotRelevantFixture,
      getMeFixture,
      getRuntimeConfigFixture,
    ]),
  );
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      transportSpy({ url, method: init?.method ?? "GET" });
      return inner(input, init);
    },
  });
}

function wrap(ui: ReactNode, transportSpy: ReturnType<typeof vi.fn>) {
  const client = buildClient(transportSpy);
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

describe("Recommended privacy negative-grep (item 3.8)", () => {
  let logSpy: MockInstance<typeof console.log>;
  let errSpy: MockInstance<typeof console.error>;
  let warnSpy: MockInstance<typeof console.warn>;
  let setItemSpy: MockInstance<typeof Storage.prototype.setItem>;
  let openSpy: MockInstance<typeof window.open>;
  let toastSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    logSpy = vi.spyOn(console, "log").mockImplementation(() => undefined);
    errSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);
    warnSpy = vi.spyOn(console, "warn").mockImplementation(() => undefined);
    setItemSpy = vi.spyOn(Storage.prototype, "setItem");
    openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
    toastSpy = vi.fn();
    (window as unknown as { eiToast?: typeof toastSpy }).eiToast = toastSpy;
  });

  afterEach(() => {
    delete (window as unknown as { eiToast?: unknown }).eiToast;
    vi.restoreAllMocks();
  });

  function expectNoLeak(): void {
    for (const spy of [logSpy, errSpy, warnSpy]) {
      for (const call of spy.mock.calls) {
        const text = call.map((v) => (typeof v === "string" ? v : "")).join(" ");
        for (const token of SECRET_TOKENS) {
          expect(
            text.includes(token),
            `console call leaked ${token}: ${text}`,
          ).toBe(false);
        }
      }
    }
    for (const call of setItemSpy.mock.calls) {
      const value = String(call[1]);
      for (const token of SECRET_TOKENS) {
        expect(
          value.includes(token),
          `localStorage.setItem leaked ${token}`,
        ).toBe(false);
      }
    }
    // URL must not contain any secret tokens (no jobMatchId / sourceUrl in URL)
    const urlState =
      window.location.href + (window.location.search ?? "") + (window.location.hash ?? "");
    for (const token of SECRET_TOKENS) {
      expect(urlState.includes(token), `URL leaked ${token}`).toBe(false);
    }
  }

  it("toggleSave does not leak jobMatchId / sourceUrl through console / URL / localStorage", async () => {
    const transportSpy = vi.fn();
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      transportSpy,
    );
    fireEvent.click(await screen.findByTestId("jdmatch-detail-action-save"));
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    expectNoLeak();
  });

  it("dismiss does not leak jobMatchId / freeNote through console / URL / localStorage", async () => {
    const transportSpy = vi.fn();
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      transportSpy,
    );
    fireEvent.click(await screen.findByTestId("jdmatch-detail-action-dismiss"));
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    expectNoLeak();
  });

  it("openSource does not leak sourceUrl through console / URL / localStorage", async () => {
    const transportSpy = vi.fn();
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      transportSpy,
    );
    fireEvent.click(await screen.findByTestId("jdmatch-detail-action-source"));
    await waitFor(() => expect(openSpy).toHaveBeenCalled());
    expectNoLeak();
  });

  it("transport spy logs only url + method, never request body or response body", async () => {
    const transportSpy = vi.fn();
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      transportSpy,
    );
    fireEvent.click(await screen.findByTestId("jdmatch-detail-action-save"));
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    for (const call of transportSpy.mock.calls) {
      const arg = call[0] as { url: string; method: string };
      // Only url + method are recorded; no body field on the spy contract
      expect(Object.keys(arg).sort()).toEqual(["method", "url"]);
      // jobMatchId may legitimately appear in URL path for DELETE
      // /jd-match/watchlist/{id}, but body+IK headers must NOT be in the spy's
      // recorded shape. The shape itself omits body so this is enforced by
      // construction.
    }
  });
});
