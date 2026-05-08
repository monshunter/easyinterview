// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
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

import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";

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

describe("ParseAuthGate — confirm", () => {
  it("redirects to auth_login when unauthenticated user clicks Confirm", async () => {
    const client = createUnauthClient();
    const { navigate } = renderUnauth(client);
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
          }),
        }),
      );
    });

    // Should not call updateTargetJob when unauthenticated
    expect(updateSpy).not.toHaveBeenCalled();
  });
});
