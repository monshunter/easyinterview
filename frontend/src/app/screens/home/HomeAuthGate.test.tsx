// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { createFixtureBackedFetch, createFixtureRegistry } from "../../../api/mockTransport";
import { EasyInterviewClient } from "../../../api/generated/client";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { HomeScreen } from "./HomeScreen";

import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import importTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/importTargetJob.json";

function createClient() {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([getMeFixture, importTargetJobFixture]),
  );
  return new EasyInterviewClient({ fetch });
}

function renderHome(client: EasyInterviewClient) {
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
            <HomeScreen route={{ name: "home", params: {} }} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("HomeAuthGate — paste import", () => {
  it("redirects to auth_login when unauthenticated user pastes JD", async () => {
    const client = createClient();
    const { navigate } = renderHome(client);
    const importSpy = vi.spyOn(client, "importTargetJob");

    await waitFor(() => {
      expect(screen.getByTestId("home-jd-textarea")).toBeInTheDocument();
    });

    await userEvent.type(
      screen.getByTestId("home-jd-textarea"),
      "Senior Frontend Engineer needed",
    );
    screen.getByTestId("home-jd-submit").click();

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "auth_login",
          params: expect.objectContaining({
            pendingRoute: "parse",
            pendingType: "import_jd",
            source: "paste",
          }),
        }),
      );
    });

    expect(importSpy).not.toHaveBeenCalled();
  });
});

describe("HomeAuthGate — url import", () => {
  it("redirects to auth_login when unauthenticated user imports via URL modal", async () => {
    const client = createClient();
    const { navigate } = renderHome(client);
    const importSpy = vi.spyOn(client, "importTargetJob");

    await waitFor(() => {
      expect(screen.getByText("URL")).toBeInTheDocument();
    });

    screen.getByText("URL").click();
    const urlInput = await screen.findByTestId("home-modal-url-input");
    await userEvent.type(urlInput, "https://acme.example/careers/senior");
    screen.getByTestId("home-modal-url-continue").click();

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "auth_login",
          params: expect.objectContaining({
            pendingRoute: "parse",
            pendingType: "import_jd",
            source: "url",
          }),
        }),
      );
    });

    expect(importSpy).not.toHaveBeenCalled();
  });
});

describe("HomeAuthGate — upload import", () => {
  it("redirects to auth_login when unauthenticated user uploads", async () => {
    const client = createClient();
    const { navigate } = renderHome(client);
    const importSpy = vi.spyOn(client, "importTargetJob");

    await waitFor(() => {
      expect(screen.getByTestId("home-upload-trigger")).toBeInTheDocument();
    });

    screen.getByTestId("home-upload-trigger").click();
    const continueBtn = await screen.findByTestId("home-modal-upload-continue");
    continueBtn.click();

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "auth_login",
          params: expect.objectContaining({
            pendingRoute: "parse",
            pendingType: "import_jd",
            source: "upload",
          }),
        }),
      );
    });

    expect(importSpy).not.toHaveBeenCalled();
  });
});
