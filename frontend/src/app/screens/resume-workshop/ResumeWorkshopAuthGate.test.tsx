// @vitest-environment jsdom
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { ResumeWorkshopScreen } from "./ResumeWorkshopScreen";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import listResumeVersionsFixture from "../../../../../openapi/fixtures/Resumes/listResumeVersions.json";
import getResumeVersionFixture from "../../../../../openapi/fixtures/Resumes/getResumeVersion.json";
import exportResumeVersionFixture from "../../../../../openapi/fixtures/Resumes/exportResumeVersion.json";

const RESUME_FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  listResumeVersionsFixture,
  getResumeVersionFixture,
  exportResumeVersionFixture,
];

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry(RESUME_FIXTURES),
      { scenario: "default" },
    ),
  });
}

function renderResumeWorkshop(
  client: EasyInterviewClient,
  nav: ReturnType<typeof vi.fn>,
  route: Route,
  authMode: "authenticated" | "unauthenticated",
): ReactNode {
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: `example=${authMode}` } },
        }}
      >
        <NavigationProvider value={{ navigate: nav }}>
          <ResumeWorkshopScreen route={route} />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

const VERSION_ID = "0195f2d0-0001-7000-8000-000000000201";

describe("ResumeWorkshopScreen auth boundary", () => {
  it("renders the auth gate (and hides list / detail / not-implemented) when the runtime is unauthenticated", async () => {
    const client = buildClient();
    const listSpy = vi.spyOn(client, "listResumes");
    const versionsSpy = vi.spyOn(client, "listResumeVersions");
    const getSpy = vi.spyOn(client, "getResumeVersion");
    const exportSpy = vi.spyOn(client, "exportResumeVersion");
    const nav = vi.fn();

    renderResumeWorkshop(
      client,
      nav,
      { name: "resume_versions", params: {} },
      "unauthenticated",
    );

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-auth-gate")).toBeInTheDocument();
    });
    expect(screen.queryByTestId("resume-workshop-list")).not.toBeInTheDocument();
    expect(screen.queryByTestId("resume-workshop-detail")).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-not-implemented"),
    ).not.toBeInTheDocument();

    expect(listSpy).not.toHaveBeenCalled();
    expect(versionsSpy).not.toHaveBeenCalled();
    expect(getSpy).not.toHaveBeenCalled();
    expect(exportSpy).not.toHaveBeenCalled();
  });

  it("renders the auth gate without calling getResumeVersion when versionId is in the URL but the user is unauthenticated", async () => {
    const client = buildClient();
    const getSpy = vi.spyOn(client, "getResumeVersion");
    const nav = vi.fn();

    renderResumeWorkshop(
      client,
      nav,
      {
        name: "resume_versions",
        params: { versionId: VERSION_ID, tab: "rewrites" },
      },
      "unauthenticated",
    );

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-auth-gate")).toBeInTheDocument();
    });
    expect(getSpy).not.toHaveBeenCalled();
  });

  it("renders the auth gate even for flow=create (CreateFlow itself requires auth)", async () => {
    const client = buildClient();
    const nav = vi.fn();

    renderResumeWorkshop(
      client,
      nav,
      { name: "resume_versions", params: { flow: "create" } },
      "unauthenticated",
    );

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-auth-gate")).toBeInTheDocument();
    });
    expect(
      screen.queryByTestId("resume-workshop-not-implemented"),
    ).not.toBeInTheDocument();
  });

  it("clicking the auth gate CTA navigates to auth_login with a pendingAction that only carries route params (flow, versionId, tab, branchOriginalId) — never raw text", async () => {
    const client = buildClient();
    const nav = vi.fn();

    renderResumeWorkshop(
      client,
      nav,
      {
        name: "resume_versions",
        params: {
          flow: "branch",
          branchOriginalId: "01918fa0-0000-7000-8000-000000001000",
          versionId: VERSION_ID,
          tab: "rewrites",
        },
      },
      "unauthenticated",
    );

    const cta = await screen.findByTestId("resume-workshop-auth-cta");
    await userEvent.setup().click(cta);

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    expect(navCall.name).toBe("auth_login");
    const params = navCall.params as Record<string, string>;

    expect(params.pendingRoute).toBe("resume_versions");
    expect(params.pendingType).toBe("open_resume_workshop");
    expect(params.pendingLabel).toBeTruthy();
    expect(params.flow).toBe("branch");
    expect(params.branchOriginalId).toBe("01918fa0-0000-7000-8000-000000001000");
    expect(params.versionId).toBe(VERSION_ID);
    expect(params.tab).toBe("rewrites");

    expect(params.rawText).toBeUndefined();
    expect(params.parsedTextSnapshot).toBeUndefined();
    expect(params.parsedSummary).toBeUndefined();
    expect(params.structuredProfile).toBeUndefined();
    expect(params.originalText).toBeUndefined();
    expect(params.suggestion).toBeUndefined();
  });

  it("renders the placeholder list (instead of the auth gate) when the runtime reports authenticated", async () => {
    const client = buildClient();
    const nav = vi.fn();

    renderResumeWorkshop(
      client,
      nav,
      { name: "resume_versions", params: {} },
      "authenticated",
    );

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-list")).toBeInTheDocument();
    });
    expect(
      screen.queryByTestId("resume-workshop-auth-gate"),
    ).not.toBeInTheDocument();
  });
});
