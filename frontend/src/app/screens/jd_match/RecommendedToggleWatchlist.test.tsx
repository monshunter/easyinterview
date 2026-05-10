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
import addToWatchlistFixture from "../../../../../openapi/fixtures/JobMatch/addToWatchlist.json";
import removeFromWatchlistFixture from "../../../../../openapi/fixtures/JobMatch/removeFromWatchlist.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

import { JDMatchScreen } from "./JDMatchScreen";

function buildClient(opts: {
  addScenario?: string;
  removeScenario?: string;
  fetchSpy?: ReturnType<typeof vi.fn>;
}) {
  const inner = createFixtureBackedFetch(
    createFixtureRegistry([
      getJobMatchProfileFixture,
      getAgentScanStatusFixture,
      listJobRecommendationsFixture,
      addToWatchlistFixture,
      removeFromWatchlistFixture,
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
      const isAdd =
        url.includes("/jd-match/watchlist") &&
        !url.match(/\/jd-match\/watchlist\/[^/?]+/) &&
        method === "POST";
      const isRemove =
        url.match(/\/jd-match\/watchlist\/[^/?]+/) && method === "DELETE";
      if (isAdd && opts.addScenario)
        headers.set("Prefer", `example=${opts.addScenario}`);
      if (isRemove && opts.removeScenario)
        headers.set("Prefer", `example=${opts.removeScenario}`);
      opts.fetchSpy?.({ url, method, headers: Object.fromEntries(headers) });
      return inner(input, { ...init, headers });
    },
  });
}

function wrap(
  ui: ReactNode,
  opts: {
    addScenario?: string;
    removeScenario?: string;
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

describe("RecommendedToggleWatchlist integration (item 3.3 + 3.9)", () => {
  it("Save click on unsaved card → addToWatchlist body { jobMatchId } + IK header + ok toast (en)", async () => {
    const fetchSpy = vi.fn();
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { addScenario: "default", lang: "en", fetchSpy },
    );
    const saveBtn = await screen.findByTestId("jdmatch-detail-action-save");
    expect(saveBtn.textContent?.toLowerCase()).not.toContain("saved");
    fireEvent.click(saveBtn);
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    expect(toastSpy.mock.calls[0]![0]).toMatch(/saved to watchlist/i);
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("ok");
    // Verify the underlying fetch carried Idempotency-Key on the POST
    const addCall = fetchSpy.mock.calls.find(
      ([req]) =>
        req.method === "POST" &&
        req.url.includes("/jd-match/watchlist") &&
        !req.url.match(/\/jd-match\/watchlist\/[^/?]+/),
    );
    expect(addCall).toBeTruthy();
    expect(addCall![0].headers["idempotency-key"]).toBeTruthy();
  });

  it("Save toast in zh dictionary", async () => {
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { addScenario: "default", lang: "zh" },
    );
    fireEvent.click(await screen.findByTestId("jdmatch-detail-action-save"));
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    expect(toastSpy.mock.calls[0]![0]).toMatch(/已加入关注列表/);
  });

  it("Save 4xx → revert label back to Save and dispatch danger toast", async () => {
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { addScenario: "4xx-validation", lang: "en" },
    );
    const saveBtn = await screen.findByTestId("jdmatch-detail-action-save");
    fireEvent.click(saveBtn);
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("danger");
    // After revert, the save button label should still read "Save" (not "Saved")
    expect(saveBtn.textContent?.toLowerCase()).toContain("save");
    expect(saveBtn.textContent?.toLowerCase()).not.toContain("saved");
  });

  it("Unsave click on saved card → removeFromWatchlist with id + IK + ok toast", async () => {
    const fetchSpy = vi.fn();
    wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { removeScenario: "default", lang: "en", fetchSpy },
    );
    // Switch to second card (saved=true in default fixture)
    fireEvent.click(
      await screen.findByTestId(
        "jdmatch-card-01918fa0-0000-7000-8000-00000000a002",
      ),
    );
    const saveBtn = screen.getByTestId("jdmatch-detail-action-save");
    expect(saveBtn.textContent?.toLowerCase()).toContain("saved");
    fireEvent.click(saveBtn);
    await waitFor(() => expect(toastSpy).toHaveBeenCalled());
    expect(toastSpy.mock.calls[0]![0]).toMatch(/removed from watchlist/i);
    const removeCall = fetchSpy.mock.calls.find(
      ([req]) =>
        req.method === "DELETE" &&
        req.url.match(/\/jd-match\/watchlist\/[^/?]+/),
    );
    expect(removeCall).toBeTruthy();
    expect(removeCall![0].headers["idempotency-key"]).toBeTruthy();
    expect(removeCall![0].url).toContain(
      "01918fa0-0000-7000-8000-00000000a002",
    );
  });

  it("Optimistic Save flips the button label to Saved before server resolution", async () => {
    const resolveRef: { current: ((value: Response) => void) | null } = {
      current: null,
    };
    const slowFetch = vi.fn(
      () =>
        new Promise<Response>((r) => {
          resolveRef.current = r;
        }),
    );
    const client = new EasyInterviewClient({
      fetch: async (input, init) => {
        const url =
          typeof input === "string"
            ? input
            : (input as URL | Request).toString();
        if (
          url.includes("/jd-match/watchlist") &&
          (init?.method ?? "GET") === "POST"
        ) {
          return slowFetch();
        }
        const inner = createFixtureBackedFetch(
          createFixtureRegistry([
            getJobMatchProfileFixture,
            getAgentScanStatusFixture,
            listJobRecommendationsFixture,
            getMeFixture,
            getRuntimeConfigFixture,
          ]),
        );
        return inner(input, init);
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
    const saveBtn = await screen.findByTestId("jdmatch-detail-action-save");
    fireEvent.click(saveBtn);
    // optimistic: label transitions to "Saved" before the slow fetch settles
    await waitFor(() => {
      expect(
        screen
          .getByTestId("jdmatch-detail-action-save")
          .textContent?.toLowerCase(),
      ).toContain("saved");
    });
    // Resolve the slow request so React effects flush
    resolveRef.current?.(
      new Response(
        JSON.stringify({
          id: "wl-1",
          linkedJobMatchId: "01918fa0-0000-7000-8000-00000000a001",
          title: "x",
          company: "x",
          tone: "ok",
          addedAt: "2026-05-10T00:00:00Z",
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );
  });
});
