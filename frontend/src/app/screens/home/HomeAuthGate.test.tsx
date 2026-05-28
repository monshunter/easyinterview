// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { createFixtureBackedFetch, createFixtureRegistry } from "../../../api/mockTransport";
import { EasyInterviewClient } from "../../../api/generated/client";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { HomeScreen } from "./HomeScreen";
import {
  clearPendingImportSourcesForTests,
  storePendingImportSource,
} from "./pendingImportState";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import importTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/importTargetJob.json";

function createClient(scenario?: string) {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getRuntimeConfigFixture,
      getMeFixture,
      importTargetJobFixture,
    ]),
    scenario ? { scenario } : undefined,
  );
  return new EasyInterviewClient({ fetch });
}

function renderHome(
  client: EasyInterviewClient,
  options?: {
    routeParams?: Record<string, string>;
    getMeScenario?: "authenticated" | "unauthenticated";
  },
) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider
          client={client}
          requestOptions={{
            getMe: {
              headers: {
                Prefer: `example=${options?.getMeScenario ?? "unauthenticated"}`,
              },
            },
          }}
        >
          <NavigationProvider value={{ navigate }}>
            <HomeScreen
              route={{ name: "home", params: options?.routeParams ?? {} }}
            />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("HomeAuthGate — paste import", () => {
  afterEach(() => {
    clearPendingImportSourcesForTests();
  });

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
            pendingRoute: "home",
            pendingType: "import_jd",
            source: "paste",
            pendingImportId: expect.any(String),
          }),
        }),
      );
    });

    const serialized = JSON.stringify(navigate.mock.calls[0]?.[0]);
    expect(serialized).not.toContain("Senior Frontend Engineer needed");
    expect(importSpy).not.toHaveBeenCalled();
  });

  it("restores pending paste import after login without carrying raw JD in route params", async () => {
    const jdText = "Senior Frontend Engineer needed";
    const pendingImportId = storePendingImportSource({
      source: "paste",
      rawText: jdText,
    });
    const client = createClient("manual-text-primary");
    const importSpy = vi.spyOn(client, "importTargetJob");
    const { navigate } = renderHome(client, {
      getMeScenario: "authenticated",
      routeParams: { pendingImportId, source: "paste" },
    });

    await waitFor(() => {
      expect(importSpy).toHaveBeenCalledTimes(1);
    });

    expect(importSpy.mock.calls[0]?.[0]).toMatchObject({
      source: { type: "manual_text", rawText: jdText },
      targetLanguage: "en",
    });
    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "parse",
          params: expect.objectContaining({
            targetJobId: "01918fa0-0000-7000-8000-000000002001",
            source: "paste",
          }),
        }),
      );
    });
    expect(JSON.stringify({ pendingImportId, source: "paste" })).not.toContain(
      jdText,
    );
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
            pendingRoute: "home",
            pendingType: "import_jd",
            source: "url",
            pendingImportId: expect.any(String),
          }),
        }),
      );
    });

    const serialized = JSON.stringify(navigate.mock.calls[0]?.[0]);
    expect(serialized).not.toContain("https://acme.example/careers/senior");
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
            pendingRoute: "home",
            pendingType: "import_jd",
            source: "upload",
            pendingImportId: expect.any(String),
          }),
        }),
      );
    });

    expect(importSpy).not.toHaveBeenCalled();
  });
});

describe("HomeAuthGate — protected entry CTAs", () => {
  it("redirects to auth_login before opening job picks, resume workshop, or debrief", async () => {
    const client = createClient();
    const { navigate } = renderHome(client);

    await waitFor(() => {
      expect(screen.getByText("Open job recommendations")).toBeInTheDocument();
    });

    screen.getByText("Open job recommendations").click();
    screen.getByTestId("home-resume-create").click();
    screen.getByText("Open debrief").click();

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "auth_login",
          params: expect.objectContaining({
            pendingRoute: "jd_match",
            pendingType: "open_protected_route",
          }),
        }),
      );
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "auth_login",
          params: expect.objectContaining({
            pendingRoute: "resume_versions",
            pendingType: "open_protected_route",
            flow: "create",
          }),
        }),
      );
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "auth_login",
          params: expect.objectContaining({
            pendingRoute: "debrief",
            pendingType: "open_protected_route",
          }),
        }),
      );
    });
  });
});
