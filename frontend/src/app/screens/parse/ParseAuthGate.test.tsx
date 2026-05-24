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

import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";

const LOADING_PREVIEW_DELAY = 3200;

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
          body: { ...body, analysisStatus: "ready" as const },
        },
      },
    },
  };

  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([getMeFixture, readyFixture]),
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
              route={{ name: "parse", params: { targetJobId: "tj-1" } }}
            />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

async function renderReadyUnauth(client: EasyInterviewClient) {
  vi.useFakeTimers();
  const result = renderUnauth(client);

  await act(async () => {
    await vi.advanceTimersByTimeAsync(LOADING_PREVIEW_DELAY);
  });
  vi.useRealTimers();

  return result;
}

afterEach(() => {
  vi.useRealTimers();
});

describe("ParseAuthGate — confirm", () => {
  it("redirects to auth_login when unauthenticated user clicks Confirm", async () => {
    const client = createUnauthClient();
    const { navigate } = await renderReadyUnauth(client);
    const updateSpy = vi.spyOn(client, "updateTargetJob");

    await screen.findByTestId("parse-action-confirm");
    screen.getByTestId("parse-action-confirm").click();

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "auth_login",
          params: expect.objectContaining({
            pendingRoute: "workspace",
            pendingType: "confirm_interview",
            targetJobId: "01918fa0-0000-7000-8000-000000002000",
            jobId: "01918fa0-0000-7000-8000-000000002000",
            jdId: "jd-01918fa0-0000-7000-8000-000000002000",
            planId: "plan-01918fa0-0000-7000-8000-000000002000",
            resumeVersionId: "resume-unbound",
            roundId: "round-technical-1",
            roundName: "Technical Round 1",
          }),
        }),
      );
    });

    // Should not call updateTargetJob when unauthenticated
    expect(updateSpy).not.toHaveBeenCalled();
  });
});
