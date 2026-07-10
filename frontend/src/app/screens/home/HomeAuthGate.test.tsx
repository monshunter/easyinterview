// @vitest-environment jsdom
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { createFixtureBackedFetch, createFixtureRegistry } from "../../../api/mockTransport";
import { EasyInterviewClient } from "../../../api/generated/client";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { HomeScreen } from "./HomeScreen";
import { storePendingImportSource } from "./pendingImportState";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import importTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/importTargetJob.json";

type ListResumesResponse = Awaited<ReturnType<EasyInterviewClient["listResumes"]>>;

const defaultListResumesResponse = listResumesFixture.scenarios.default.response
  .body as ListResumesResponse;

function createClient(scenario?: string) {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getRuntimeConfigFixture,
      getMeFixture,
      listResumesFixture,
      importTargetJobFixture,
    ]),
    scenario ? { scenario } : undefined,
  );
  const client = new EasyInterviewClient({ fetch });
  vi.spyOn(client, "listResumes").mockResolvedValue(defaultListResumesResponse);
  return client;
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

describe("pending import production boundary", () => {
  it("does not expose a test-only reset API", () => {
    const resetApi = ["clearPendingImportSources", "ForTests"].join("");
    const source = readFileSync(
      resolve(__dirname, "pendingImportState.ts"),
      "utf8",
    );

    expect(source).not.toContain(resetApi);
  });
});

describe("HomeAuthGate — paste import", () => {
  it("does not create import pending action before a resume is selected", async () => {
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
    await userEvent.click(screen.getByTestId("home-jd-submit"));

    expect(screen.getByTestId("home-jd-submit")).toBeDisabled();
    expect(navigate).not.toHaveBeenCalled();
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
      routeParams: {
        pendingImportId,
        source: "paste",
        resumeId: "01918fa0-0000-7000-8000-000000001000",
      },
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
            resumeId: "01918fa0-0000-7000-8000-000000001000",
          }),
        }),
      );
    });
    expect(
      JSON.stringify({
        pendingImportId,
        source: "paste",
        resumeId: "01918fa0-0000-7000-8000-000000001000",
      }),
    ).not.toContain(jdText);
  });
});

describe("HomeAuthGate — url import", () => {
  it("does not create URL pending action before a resume is selected", async () => {
    const client = createClient();
    const { navigate } = renderHome(client);
    const importSpy = vi.spyOn(client, "importTargetJob");

    await waitFor(() => {
      expect(screen.getByText("URL")).toBeInTheDocument();
    });

    await userEvent.click(screen.getByText("URL"));
    const urlInput = await screen.findByTestId("home-modal-url-input");
    await userEvent.type(urlInput, "https://acme.example/careers/senior");
    await userEvent.click(screen.getByTestId("home-modal-url-continue"));

    expect(navigate).not.toHaveBeenCalled();
    expect(importSpy).not.toHaveBeenCalled();
  });
});

describe("HomeAuthGate — upload import", () => {
  it("does not create upload pending action before a resume is selected", async () => {
    const client = createClient();
    const { navigate } = renderHome(client);
    const importSpy = vi.spyOn(client, "importTargetJob");

    await waitFor(() => {
      expect(screen.getByTestId("home-upload-trigger")).toBeInTheDocument();
    });

    await userEvent.click(screen.getByTestId("home-upload-trigger"));
    const continueBtn = await screen.findByTestId("home-modal-upload-continue");
    await userEvent.click(continueBtn);

    expect(navigate).not.toHaveBeenCalled();
    expect(importSpy).not.toHaveBeenCalled();
  });
});

describe("HomeAuthGate — protected entry CTAs", () => {
  it("redirects to auth_login before opening resume workshop and does not expose debrief", async () => {
    const client = createClient();
    const { navigate } = renderHome(client);

    await waitFor(() => {
      expect(screen.getByTestId("home-resume-create")).toBeInTheDocument();
    });
    expect(screen.queryByText("Open debrief")).not.toBeInTheDocument();

    await userEvent.click(screen.getByTestId("home-resume-create"));

    await waitFor(() => {
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
    });
  });
});
