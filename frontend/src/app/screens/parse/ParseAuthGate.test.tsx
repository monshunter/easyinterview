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

const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";

function createUnauthClient() {
  const body = (
    getTargetJobFixture.scenarios["default"] as {
      response: { body: Record<string, unknown> };
    }
  ).response.body;
  const readyFixture = {
    operationId: "getTargetJob",
    scenarios: {
      default: {
        response: {
          status: 200,
          body: {
            ...body,
            analysisStatus: "ready" as const,
            resumeId: null,
          },
        },
      },
    },
  };

  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([getRuntimeConfigFixture, getMeFixture, readyFixture]),
  );
  return new EasyInterviewClient({ fetch });
}

function renderUnauth(client: EasyInterviewClient) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider
          client={client}
          requestOptions={{
            getMe: { headers: { Prefer: "example=unauthenticated" } },
          }}
        >
          <NavigationProvider value={{ navigate }}>
            <ParseScreen
              route={{
                name: "workspace",
                params: { targetJobId: TARGET_JOB_ID },
              }}
            />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

async function renderReadyUnauth(client: EasyInterviewClient) {
  return renderUnauth(client);
}

describe("ParseAuthGate — resume-required launch", () => {
  it("keeps start disabled for unauthenticated users without a saved bound resume", async () => {
    const client = createUnauthClient();
    const { navigate } = await renderReadyUnauth(client);
    const updateSpy = vi.spyOn(client, "updateTargetJob");
    const listSpy = vi.spyOn(client, "listResumes");

    const startBtn = await screen.findByTestId("parse-action-start-interview");

    expect(screen.queryByTestId("parse-action-save-plan")).not.toBeInTheDocument();
    expect(startBtn).toBeDisabled();
    expect(screen.getByTestId("parse-resume-missing")).toBeInTheDocument();
    expect(screen.queryByTestId("parse-resume-picker")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-confirm")).not.toBeInTheDocument();

    startBtn.click();

    expect(navigate).not.toHaveBeenCalled();
    expect(updateSpy).not.toHaveBeenCalled();
    expect(listSpy).not.toHaveBeenCalled();
  });
});
