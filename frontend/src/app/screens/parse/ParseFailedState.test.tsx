// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { EasyInterviewClient } from "../../../api/generated/client";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { ParseScreen } from "./ParseScreen";

import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";

function makeFixture(analysisStatus: "failed" | "ready") {
  const body = (
    getTargetJobFixture.scenarios["default"] as {
      response: { body: Record<string, unknown> };
    }
  ).response.body;
  return {
    operationId: "getTargetJob",
    scenarios: {
      default: {
        response: {
          status: 200,
          body: { ...body, analysisStatus },
        },
      },
    },
  };
}

function renderParse(analysisStatus: "failed" | "ready") {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([makeFixture(analysisStatus)]),
    { scenario: "default" },
  );
  const client = new EasyInterviewClient({ fetch });
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider client={client}>
        <NavigationProvider value={{ navigate: () => {} }}>
          <ParseScreen
            route={{ name: "parse", params: { targetJobId: "tj-1" } }}
          />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

describe("ParseFailedState", () => {
  it("shows failed UI when analysisStatus is failed", async () => {
    renderParse("failed");

    await waitFor(() => {
      expect(screen.getByTestId("parse-failed-title")).toBeInTheDocument();
    });

    expect(screen.getByTestId("parse-failed-message")).toBeInTheDocument();
    expect(screen.getByTestId("parse-failed-reparse")).toBeInTheDocument();
    expect(screen.getByTestId("parse-failed-home")).toBeInTheDocument();
  });

  it("shows preview (not failed UI) when analysisStatus is ready", async () => {
    renderParse("ready");

    await waitFor(() => {
      expect(screen.getByTestId("parse-basics-title")).toBeInTheDocument();
    });

    expect(screen.queryByTestId("parse-failed-title")).toBeNull();
  });
});
