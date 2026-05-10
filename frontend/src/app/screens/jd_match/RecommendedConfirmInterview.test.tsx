// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
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

function buildClient(opts?: { signedIn?: boolean }) {
  const fixtures = [
    getJobMatchProfileFixture,
    getAgentScanStatusFixture,
    listJobRecommendationsFixture,
    getMeFixture,
    getRuntimeConfigFixture,
  ];
  const registry = createFixtureRegistry(fixtures);
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const wantsAuthenticated =
        url.includes("/me") &&
        (opts?.signedIn ?? true) === true;
      const wantsUnauthenticated =
        url.includes("/me") &&
        (opts?.signedIn ?? true) === false;
      const inner = createFixtureBackedFetch(registry, undefined);
      const headers = new Headers(init?.headers ?? {});
      if (wantsAuthenticated) headers.set("Prefer", "example=authenticated");
      if (wantsUnauthenticated) headers.set("Prefer", "example=unauthenticated");
      return inner(input, { ...init, headers });
    },
  });
}

function wrap(ui: ReactNode, opts?: { signedIn?: boolean }) {
  const navigate = vi.fn();
  const client = buildClient({ signedIn: opts?.signedIn ?? true });
  const tree = (
    <DisplayPreferencesProvider initial={{ lang: "en" }}>
      <AppRuntimeProvider client={client}>
        <NavigationProvider value={{ navigate }}>{ui}</NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>
  );
  return { navigate, ...render(tree) };
}

describe("RecommendedConfirmInterview integration (item 3.5)", () => {
  it("Confirm interview button navigates to parse with { source, sourceJobMatchId } only", async () => {
    const { navigate } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
    );
    const confirmBtn = await screen.findByTestId(
      "jdmatch-detail-action-confirm",
    );
    fireEvent.click(confirmBtn);
    await waitFor(() => expect(navigate).toHaveBeenCalledTimes(1));
    const call = navigate.mock.calls[0]![0];
    expect(call.name).toBe("parse");
    const params = call.params as Record<string, string>;
    expect(params.source).toBe("jd_match");
    expect(typeof params.sourceJobMatchId).toBe("string");
    expect(params.sourceJobMatchId!.length).toBeGreaterThan(0);
    // Param surface is exactly source + sourceJobMatchId (no other jd_match
    // internal state leaks: query, saved, hidden, etc.)
    expect(Object.keys(params).sort()).toEqual(["source", "sourceJobMatchId"]);
  });

  it("Switching cards updates Confirm interview target so nav carries the newly-selected jobMatchId", async () => {
    const { navigate } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
    );
    // Wait for cards to render then click the second card
    const secondCard = await screen.findByTestId(
      "jdmatch-card-01918fa0-0000-7000-8000-00000000a002",
    );
    fireEvent.click(secondCard);
    fireEvent.click(screen.getByTestId("jdmatch-detail-action-confirm"));
    await waitFor(() => expect(navigate).toHaveBeenCalled());
    const call = navigate.mock.calls.at(-1)![0];
    const params = call.params as Record<string, string>;
    expect(params.sourceJobMatchId).toBe(
      "01918fa0-0000-7000-8000-00000000a002",
    );
  });
});
