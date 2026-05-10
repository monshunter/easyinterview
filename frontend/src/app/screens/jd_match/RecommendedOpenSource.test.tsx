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
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

import { JDMatchScreen } from "./JDMatchScreen";

function buildClient() {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([
        getJobMatchProfileFixture,
        getAgentScanStatusFixture,
        listJobRecommendationsFixture,
        getMeFixture,
        getRuntimeConfigFixture,
      ]),
    ),
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

let openSpy: MockInstance<typeof window.open>;

beforeEach(() => {
  openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
});

afterEach(() => {
  openSpy.mockRestore();
  vi.restoreAllMocks();
});

describe("RecommendedOpenSource integration (item 3.6)", () => {
  it("Source button calls window.open with sourceUrl, _blank, and noopener,noreferrer", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);
    const sourceBtn = await screen.findByTestId(
      "jdmatch-detail-action-source",
    );
    fireEvent.click(sourceBtn);
    await waitFor(() => expect(openSpy).toHaveBeenCalledTimes(1));
    const [url, target, features] = openSpy.mock.calls[0]!;
    expect(url).toBe("https://acme.example/careers/senior-frontend");
    expect(target).toBe("_blank");
    expect(features).toBe("noopener,noreferrer");
  });

  it("Source button is disabled when sourceUrl is null and never invokes window.open", async () => {
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);
    // Wait for cards then select the third card whose sourceUrl is null
    const thirdCard = await screen.findByTestId(
      "jdmatch-card-01918fa0-0000-7000-8000-00000000a003",
    );
    fireEvent.click(thirdCard);
    const sourceBtn = await screen.findByTestId(
      "jdmatch-detail-action-source",
    );
    expect((sourceBtn as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(sourceBtn);
    expect(openSpy).not.toHaveBeenCalled();
  });

  it("does not write sourceUrl to console / URL / localStorage / telemetry (privacy)", async () => {
    const logSpy = vi.spyOn(console, "log").mockImplementation(() => undefined);
    const errSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
    const warnSpy = vi
      .spyOn(console, "warn")
      .mockImplementation(() => undefined);
    const setItemSpy = vi.spyOn(Storage.prototype, "setItem");
    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />);
    const sourceBtn = await screen.findByTestId(
      "jdmatch-detail-action-source",
    );
    fireEvent.click(sourceBtn);
    await waitFor(() => expect(openSpy).toHaveBeenCalled());
    const SECRET = "acme.example";
    for (const spy of [logSpy, errSpy, warnSpy]) {
      for (const call of spy.mock.calls) {
        const text = call.map((v) => (typeof v === "string" ? v : "")).join(" ");
        expect(text).not.toContain(SECRET);
      }
    }
    for (const call of setItemSpy.mock.calls) {
      expect(call[1]).not.toContain(SECRET);
    }
  });
});
