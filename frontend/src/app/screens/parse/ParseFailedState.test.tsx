// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";

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

const LOADING_PREVIEW_DELAY = 3200;

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

async function renderReadyParse() {
  vi.useFakeTimers();
  const result = renderParse("ready");

  await act(async () => {
    await vi.advanceTimersByTimeAsync(LOADING_PREVIEW_DELAY);
  });
  vi.useRealTimers();

  return result;
}

afterEach(() => {
  vi.useRealTimers();
});

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
    await renderReadyParse();

    await waitFor(() => {
      expect(screen.getByTestId("parse-basics-title")).toBeInTheDocument();
    });

    expect(screen.queryByTestId("parse-failed-title")).toBeNull();
  });
});
