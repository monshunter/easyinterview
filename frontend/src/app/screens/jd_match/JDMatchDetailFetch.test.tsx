// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import type { JobMatchRecommendation } from "../../../api/generated/types";
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

const DETAIL_ONLY_TITLE = "Detail endpoint title — Staff Frontend Platform";
const DETAIL_ONLY_REASON = "Detail endpoint reason that is absent from list";

function buildClient() {
  const registry = createFixtureRegistry([
    getJobMatchProfileFixture,
    getAgentScanStatusFixture,
    listJobRecommendationsFixture,
    getMeFixture,
    getRuntimeConfigFixture,
  ]);
  const client = new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const headers = new Headers(init?.headers ?? {});
      if (url.includes("/me")) headers.set("Prefer", "example=authenticated");
      const inner = createFixtureBackedFetch(registry, undefined);
      return inner(input, { ...init, headers });
    },
  });
  return client;
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

describe("JDMatchScreen detail fetch (item 3.11)", () => {
  it("calls getJobRecommendation for the selected card instead of using list summary as detail", async () => {
    const client = buildClient();
    const detailSpy = vi
      .spyOn(client, "getJobRecommendation")
      .mockImplementation(async (jobMatchId: string) => {
        const list = await client.listJobRecommendations();
        const summary = list.items.find((item) => item.id === jobMatchId);
        if (!summary) throw new Error("missing summary");
        return {
          ...summary,
          title: DETAIL_ONLY_TITLE,
          reasons: [DETAIL_ONLY_REASON],
        } satisfies JobMatchRecommendation;
      });

    wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />, client);

    const secondCard = await screen.findByTestId(
      "jdmatch-card-01918fa0-0000-7000-8000-00000000a002",
    );
    fireEvent.click(secondCard);

    await waitFor(() =>
      expect(detailSpy).toHaveBeenCalledWith(
        "01918fa0-0000-7000-8000-00000000a002",
      ),
    );
    await waitFor(() => {
      expect(screen.getByTestId("jdmatch-detail-header")).toHaveTextContent(
        DETAIL_ONLY_TITLE,
      );
      expect(screen.getByTestId("jdmatch-detail-why")).toHaveTextContent(
        DETAIL_ONLY_REASON,
      );
    });
  });
});
