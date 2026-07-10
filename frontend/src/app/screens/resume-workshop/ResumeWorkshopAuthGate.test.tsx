// @vitest-environment jsdom
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
import getResumeFixture from "../../../../../openapi/fixtures/Resumes/getResume.json";
import exportResumeFixture from "../../../../../openapi/fixtures/Resumes/exportResume.json";

const RESUME_FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  getResumeFixture,
  exportResumeFixture,
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
) {
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

const RESUME_ID = "01918fa0-0000-7000-8000-000000001000";

describe("ResumeWorkshopScreen auth boundary", () => {
  it("renders the auth gate (and hides list / detail / not-implemented) when the runtime is unauthenticated", async () => {
    const client = buildClient();
    const listSpy = vi.spyOn(client, "listResumes");
    const getSpy = vi.spyOn(client, "getResume");
    const exportSpy = vi.spyOn(client, "exportResume");
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
    expect(getSpy).not.toHaveBeenCalled();
    expect(exportSpy).not.toHaveBeenCalled();
  });

  it("renders the auth gate without calling getResume when resumeId is in the URL but the user is unauthenticated", async () => {
    const client = buildClient();
    const getSpy = vi.spyOn(client, "getResume");
    const nav = vi.fn();

    renderResumeWorkshop(
      client,
      nav,
      {
        name: "resume_versions",
        params: { resumeId: RESUME_ID, tab: "rewrites" },
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

  it("clicking the auth gate CTA navigates to auth_login with a pendingAction that only carries current flat route params — never tab state or raw text", async () => {
    const client = buildClient();
    const nav = vi.fn();

    renderResumeWorkshop(
      client,
      nav,
      {
        name: "resume_versions",
        params: {
          flow: "create",
          createMode: "paste",
          targetJobId: "01918fa0-0000-7000-8000-000000002000",
          resumeId: RESUME_ID,
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
    expect(params.flow).toBe("create");
    expect(params.createMode).toBe("paste");
    expect(params.targetJobId).toBe("01918fa0-0000-7000-8000-000000002000");
    expect(params.resumeId).toBe(RESUME_ID);
    expect(params.tab).toBeUndefined();

    expect(params.rawText).toBeUndefined();
    expect(params.parsedTextSnapshot).toBeUndefined();
    expect(params.parsedSummary).toBeUndefined();
    expect(params.structuredProfile).toBeUndefined();
    expect(params.originalText).toBeUndefined();
    expect(params.suggestion).toBeUndefined();
    // D-20 flat model: out-of-scope branch-draft and version params must NEVER ride along on the
    // pendingAction; only flat route params are restored after sign-in.
    expect(params.name).toBeUndefined();
    expect(params.target).toBeUndefined();
    expect(params.focus).toBeUndefined();
    expect(params.seed).toBeUndefined();
    expect(params.versionId).toBeUndefined();
    expect(params.branchOriginalId).toBeUndefined();
    expect(params.parentVersionId).toBeUndefined();
    expect(params.displayName).toBeUndefined();
    expect(params.focusAngle).toBeUndefined();
    expect(params.seedStrategy).toBeUndefined();
  });

  it("renders the flat list (instead of the auth gate) when the runtime reports authenticated", async () => {
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
