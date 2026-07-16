// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

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

const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";

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

function renderWorkspaceDetail(client: EasyInterviewClient) {
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider client={client}>
        <NavigationProvider value={{ navigate: vi.fn() }}>
          <ParseScreen
            route={{
              name: "workspace",
              params: { targetJobId: TARGET_JOB_ID },
            }}
          />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

async function renderReadyParse(client: EasyInterviewClient) {
  return renderWorkspaceDetail(client);
}

describe("Workspace detail — readonly plan receipt", () => {
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
    expect(req0.textContent).toMatch(/Match|命中/);
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
    expect(signal0).toHaveTextContent("Hiring team values cross-team RFC ownership");
    expect(round0).toHaveTextContent("Frontend architecture screen · 45m");
    expect(round0).toHaveTextContent("Probe scaling design systems across 10+ teams.");
    expect(round1).toHaveTextContent("Hiring manager impact interview · 50m");
    expect(round1).toHaveTextContent("Assess cross-team RFC ownership and influence.");
    expect(round0).not.toHaveTextContent(/Motivation, timing|动机/);
    expect(round0.tagName).not.toBe("BUTTON");
    expect(round1.tagName).not.toBe("BUTTON");
  });

  it("does not expose save, cancel, success reparse, or update error controls", async () => {
    const client = createClient();
    await renderReadyParse(client);

    await screen.findByTestId("parse-leading-actions");
    expect(screen.queryByTestId("parse-action-save-plan")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-cancel")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-reparse")).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-action-start-interview")).toBeInTheDocument();
  });
});
