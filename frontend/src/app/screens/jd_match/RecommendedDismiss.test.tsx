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
import markJobNotRelevantFixture from "../../../../../openapi/fixtures/JobMatch/markJobNotRelevant.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

import { JDMatchScreen } from "./JDMatchScreen";

function buildClient(opts: {
  dismissScenario?: string;
  fetchSpy?: ReturnType<typeof vi.fn>;
}) {
  const inner = createFixtureBackedFetch(
    createFixtureRegistry([
      getJobMatchProfileFixture,
      getAgentScanStatusFixture,
      listJobRecommendationsFixture,
      markJobNotRelevantFixture,
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
      const isDismiss =
        url.includes("/jd-match/recommendations/") &&
        url.endsWith("/dismiss") &&
        method === "POST";
      if (isDismiss && opts.dismissScenario)
        headers.set("Prefer", `example=${opts.dismissScenario}`);
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
    dismissScenario?: string;
    lang?: "zh" | "en";
    fetchSpy?: ReturnType<typeof vi.fn>;
  } = {},
) {
  const client = buildClient(opts);
  const navigate = vi.fn();
  const tree = (
    <DisplayPreferencesProvider initial={{ lang: opts.lang ?? "en" }}>
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

describe("RecommendedDismiss integration (item 3.4 + 3.9)", () => {
  it("Mark not relevant click → markJobNotRelevant body { reason: not_relevant } + IK + neutral toast (en)", async () => {
    const fetchSpy = vi.fn();
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { dismissScenario: "default", lang: "en", fetchSpy },
    );
    fireEvent.click(await screen.findByTestId("jdmatch-detail-action-dismiss"));
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    expect(toastSpy.mock.calls[0]![0]).toMatch(/not relevant/i);
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("neutral");
    const dismissCall = fetchSpy.mock.calls.find(
      ([req]) =>
        req.method === "POST" &&
        req.url.includes("/jd-match/recommendations/") &&
        req.url.endsWith("/dismiss"),
    );
    expect(dismissCall).toBeTruthy();
    expect(dismissCall![0].headers["idempotency-key"]).toBeTruthy();
    expect(JSON.parse(dismissCall![0].body)).toEqual({ reason: "not_relevant" });
  });

  it("Mark not relevant request body must NOT include freeNote", async () => {
    const fetchSpy = vi.fn();
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { dismissScenario: "default", fetchSpy },
    );
    fireEvent.click(await screen.findByTestId("jdmatch-detail-action-dismiss"));
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    const dismissCall = fetchSpy.mock.calls.find(
      ([req]) =>
        req.method === "POST" &&
        req.url.endsWith("/dismiss"),
    );
    const body = JSON.parse(dismissCall![0].body) as Record<string, unknown>;
    expect("freeNote" in body).toBe(false);
  });

  it("Optimistic dismiss hides the card and auto-selects the next visible item", async () => {
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { dismissScenario: "default" },
    );
    const firstCardId = "01918fa0-0000-7000-8000-00000000a001";
    await screen.findByTestId(`jdmatch-card-${firstCardId}`);
    fireEvent.click(screen.getByTestId("jdmatch-detail-action-dismiss"));
    await waitFor(() =>
      expect(screen.queryByTestId(`jdmatch-card-${firstCardId}`)).toBeNull(),
    );
    // Detail should show the next-card (jm-2)
    const detail = screen.getByTestId("jdmatch-detail-header");
    expect(detail).toHaveTextContent("Staff");
  });

  it("Mark not relevant 4xx → revert (card returns) + danger toast", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      dismissScenario: "4xx",
    });
    const firstCardId = "01918fa0-0000-7000-8000-00000000a001";
    await screen.findByTestId(`jdmatch-card-${firstCardId}`);
    fireEvent.click(screen.getByTestId("jdmatch-detail-action-dismiss"));
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("danger");
    expect(
      screen.getByTestId(`jdmatch-card-${firstCardId}`),
    ).toBeInTheDocument();
  });

  it("zh dismiss toast", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, {
      dismissScenario: "default",
      lang: "zh",
    });
    fireEvent.click(await screen.findByTestId("jdmatch-detail-action-dismiss"));
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    expect(toastSpy.mock.calls[0]![0]).toMatch(/已标记不相关/);
  });
});
