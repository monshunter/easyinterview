// @vitest-environment jsdom
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { StrictMode } from "react";

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
  consumePendingImportIntent,
  storePendingImportIntent,
} from "./pendingImportState";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import importTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/importTargetJob.json";

type ListResumesResponse = Awaited<ReturnType<EasyInterviewClient["listResumes"]>>;

const defaultListResumesResponse = listResumesFixture.scenarios.default.response
  .body as ListResumesResponse;

afterEach(() => {
  vi.restoreAllMocks();
});

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
    strict?: boolean;
  },
) {
  const navigate = vi.fn();
  const screen = (
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
    </DisplayPreferencesProvider>
  );
  return {
    navigate,
    ...render(options?.strict ? <StrictMode>{screen}</StrictMode> : screen),
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
    expect(source).not.toContain("PendingImportSource");
    expect(source).not.toContain('source: "upload"');
    expect(source).not.toContain('source: "url"');
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

  it("restores one exact pending paste import after login under StrictMode", async () => {
    const jdText = "Senior Frontend Engineer needed";
    const resumeId = "01918fa0-0000-7000-8000-000000001000";
    const originalIdempotencyKey = "ik-original-login-intent";
    const opaquePendingImportId = storePendingImportIntent({
      rawText: jdText,
      targetLanguage: "zh-CN",
      resumeId,
      idempotencyKey: originalIdempotencyKey,
    });
    const client = createClient("paste-primary");
    const importSpy = vi.spyOn(client, "importTargetJob");
    const { navigate } = renderHome(client, {
      getMeScenario: "authenticated",
      routeParams: {
        opaquePendingImportId,
      },
      strict: true,
    });

    await waitFor(() => {
      expect(importSpy).toHaveBeenCalledTimes(1);
    });

    expect(importSpy.mock.calls[0]?.[0]).toEqual({
      rawText: jdText,
      targetLanguage: "zh-CN",
      resumeId,
    });
    expect(importSpy.mock.calls[0]?.[1]?.idempotencyKey).toBe(
      originalIdempotencyKey,
    );
    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith({
        name: "parse",
        params: {
          targetJobId: "01918fa0-0000-7000-8000-000000002001",
        },
      });
    });
    expect(JSON.stringify({ opaquePendingImportId })).not.toContain(jdText);
    expect(consumePendingImportIntent(opaquePendingImportId)).toBeNull();
  });

  it("fails closed when the opaque pending import is missing", async () => {
    const client = createClient("paste-primary");
    const importSpy = vi.spyOn(client, "importTargetJob");
    const { navigate } = renderHome(client, {
      getMeScenario: "authenticated",
      routeParams: { opaquePendingImportId: "pending-import-missing" },
    });

    expect(await screen.findByTestId("home-import-error")).toHaveTextContent(
      "Paste the JD and select a resume again",
    );
    expect(importSpy).not.toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith({ name: "home", params: {} });
  });

  it("fails closed when the pending import has expired", async () => {
    const now = vi.spyOn(Date, "now").mockReturnValue(0);
    const opaquePendingImportId = storePendingImportIntent({
      rawText: "Expired JD",
      targetLanguage: "en",
      resumeId: "01918fa0-0000-7000-8000-000000001000",
      idempotencyKey: "ik-expired",
    });
    now.mockReturnValue(Number.MAX_SAFE_INTEGER);
    const client = createClient("paste-primary");
    const importSpy = vi.spyOn(client, "importTargetJob");
    const { navigate } = renderHome(client, {
      getMeScenario: "authenticated",
      routeParams: { opaquePendingImportId },
    });

    expect(await screen.findByTestId("home-import-error")).toBeInTheDocument();
    expect(importSpy).not.toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith({ name: "home", params: {} });
    now.mockRestore();
  });

  it("fails closed when the pending import was already consumed", async () => {
    const opaquePendingImportId = storePendingImportIntent({
      rawText: "Already consumed JD",
      targetLanguage: "en",
      resumeId: "01918fa0-0000-7000-8000-000000001000",
      idempotencyKey: "ik-consumed",
    });
    expect(consumePendingImportIntent(opaquePendingImportId)).not.toBeNull();
    const client = createClient("paste-primary");
    const importSpy = vi.spyOn(client, "importTargetJob");
    const { navigate } = renderHome(client, {
      getMeScenario: "authenticated",
      routeParams: { opaquePendingImportId },
    });

    expect(await screen.findByTestId("home-import-error")).toBeInTheDocument();
    expect(importSpy).not.toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith({ name: "home", params: {} });
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
