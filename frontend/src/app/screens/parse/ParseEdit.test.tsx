// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";

import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { EasyInterviewClient } from "../../../api/generated/client";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { ParseScreen } from "./ParseScreen";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import createPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import getPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import startPracticeSessionFixture from "../../../../../openapi/fixtures/PracticeSessions/startPracticeSession.json";

const LOADING_PREVIEW_DELAY = 3200;

function makeReadyFixture() {
  const body = (
    getTargetJobFixture.scenarios.default as {
      response: { body: Record<string, unknown> };
    }
  ).response.body;
  return {
    operationId: "getTargetJob",
    scenarios: {
      default: {
        response: {
          status: 200,
          body: { ...body, analysisStatus: "ready" as const },
        },
      },
    },
  };
}

function createClient() {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getRuntimeConfigFixture,
      getMeFixture,
      makeReadyFixture(),
      listResumesFixture,
      createPracticePlanFixture,
      getPracticePlanFixture,
      startPracticeSessionFixture,
    ]),
    { scenario: "default" },
  );
  return new EasyInterviewClient({ fetch });
}

function renderParse(client: EasyInterviewClient) {
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider client={client}>
        <NavigationProvider value={{ navigate: vi.fn() }}>
          <ParseScreen
            route={{ name: "parse", params: { targetJobId: "tj-1" } }}
          />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

async function renderReadyParse(client: EasyInterviewClient) {
  vi.useFakeTimers();
  const result = renderParse(client);

  await act(async () => {
    await vi.advanceTimersByTimeAsync(LOADING_PREVIEW_DELAY);
  });
  vi.useRealTimers();

  return result;
}

afterEach(() => {
  vi.useRealTimers();
});

describe("ParseEdit — readonly plan receipt", () => {
  it("renders basic fields as read-only text, not editable inputs", async () => {
    const client = createClient();
    await renderReadyParse(client);

    const titleEl = await screen.findByTestId("parse-basics-title");
    const companyEl = await screen.findByTestId("parse-basics-company");
    const locationEl = await screen.findByTestId("parse-basics-location");

    expect(titleEl).toHaveTextContent("Senior Frontend Engineer");
    expect(companyEl).toHaveTextContent("Acme");
    expect(locationEl).toHaveTextContent("Shanghai · Hybrid");
    expect(titleEl.querySelector("input")).toBeNull();
    expect(companyEl.querySelector("input")).toBeNull();
    expect(locationEl.querySelector("input")).toBeNull();
    expect(screen.queryByTestId("parse-basics-notes")).not.toBeInTheDocument();
  });

  it("renders requirements evidence as a non-interactive badge", async () => {
    const client = createClient();
    await renderReadyParse(client);

    const req0 = await screen.findByTestId("parse-requirement-must_have-0");

    expect(req0).toBeInTheDocument();
    expect(
      screen.queryByTestId("parse-requirement-must_have-0-toggle"),
    ).not.toBeInTheDocument();
    expect(req0.querySelector("button")).toBeNull();
    expect(req0.textContent).toMatch(/HIT|命中/);
  });

  it("renders hidden signals and round assumptions without edit controls", async () => {
    const client = createClient();
    await renderReadyParse(client);

    const signal0 = await screen.findByTestId("parse-hidden-signal-0");
    const round0 = await screen.findByTestId("parse-round-0");
    const round1 = await screen.findByTestId("parse-round-1");

    expect(signal0).toBeInTheDocument();
    expect(signal0.querySelector("button")).toBeNull();
    expect(round0).toBeInTheDocument();
    expect(round1).toBeInTheDocument();
    expect(round0.tagName).not.toBe("BUTTON");
    expect(round1.tagName).not.toBe("BUTTON");
  });

  it("does not expose save, cancel, success reparse, or update error controls", async () => {
    const client = createClient();
    await renderReadyParse(client);

    expect(screen.queryByTestId("parse-action-save-plan")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-cancel")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-reparse")).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-action-start-interview")).toBeInTheDocument();
  });
});
