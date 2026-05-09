// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";

import getAgentScanStatusFixture from "../../../../../openapi/fixtures/JobMatch/getAgentScanStatus.json";
import getJobMatchProfileFixture from "../../../../../openapi/fixtures/JobMatch/getJobMatchProfile.json";

import { JDMatchScreen } from "./JDMatchScreen";

function buildClient() {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([
        getJobMatchProfileFixture,
        getAgentScanStatusFixture,
      ]),
      { scenario: "default" },
    ),
  });
}

function renderJDMatch(client: EasyInterviewClient) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate }}>
            <JDMatchScreen route={{ name: "jd_match", params: {} }} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("JDMatchScreen fetch behavior on entry and tab switch", () => {
  it("calls getJobMatchProfile and getAgentScanStatus exactly once on entry", async () => {
    const client = buildClient();
    const profileSpy = vi.spyOn(client, "getJobMatchProfile");
    const agentSpy = vi.spyOn(client, "getAgentScanStatus");

    renderJDMatch(client);

    await waitFor(() => {
      expect(profileSpy).toHaveBeenCalledTimes(1);
    });
    await waitFor(() => {
      expect(agentSpy).toHaveBeenCalledTimes(1);
    });
  });

  it("does not refetch profile on tab change; refetches agent only when returning to recommended", async () => {
    const client = buildClient();
    const profileSpy = vi.spyOn(client, "getJobMatchProfile");
    const agentSpy = vi.spyOn(client, "getAgentScanStatus");

    renderJDMatch(client);

    await waitFor(() => {
      expect(profileSpy).toHaveBeenCalledTimes(1);
      expect(agentSpy).toHaveBeenCalledTimes(1);
    });

    screen.getByTestId("jdmatch-tab-search").click();
    await new Promise((r) => setTimeout(r, 30));
    expect(profileSpy).toHaveBeenCalledTimes(1);
    expect(agentSpy).toHaveBeenCalledTimes(1);

    screen.getByTestId("jdmatch-tab-watchlist").click();
    await new Promise((r) => setTimeout(r, 30));
    expect(profileSpy).toHaveBeenCalledTimes(1);
    expect(agentSpy).toHaveBeenCalledTimes(1);

    screen.getByTestId("jdmatch-tab-recommended").click();
    await waitFor(() => {
      expect(agentSpy).toHaveBeenCalledTimes(2);
    });
    expect(profileSpy).toHaveBeenCalledTimes(1);
  });
});
