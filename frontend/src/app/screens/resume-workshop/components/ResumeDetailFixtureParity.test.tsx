// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import { EasyInterviewClient } from "../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import type { Route } from "../../../routes";
import { ResumeWorkshopScreen } from "../ResumeWorkshopScreen";

import getRuntimeConfigFixture from "../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../openapi/fixtures/Auth/getMe.json";
import getResumeVersionFixture from "../../../../../../openapi/fixtures/Resumes/getResumeVersion.json";
import exportResumeVersionFixture from "../../../../../../openapi/fixtures/Resumes/exportResumeVersion.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  getResumeVersionFixture,
  exportResumeVersionFixture,
];

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario,
    }),
  });
}

function renderDetail(scenario: string, versionId: string) {
  const route: Route = {
    name: "resume_versions",
    params: { versionId },
  };
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider
        client={buildClient(scenario)}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      >
        <NavigationProvider value={{ navigate: vi.fn() }}>
          <ResumeWorkshopScreen route={route} />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

const SCENARIOS = ["default", "master-default", "targeted-with-suggestions"];

describe("getResumeVersion fixture parity (Phase 3.6)", () => {
  for (const scenarioName of SCENARIOS) {
    it(`renders the detail container with breadcrumb + branch graph + tabs for the ${scenarioName} scenario`, async () => {
      const versionId = (
        getResumeVersionFixture.scenarios as Record<
          string,
          { response: { body: { id: string } } }
        >
      )[scenarioName]!.response.body.id;
      renderDetail(scenarioName, versionId);

      await waitFor(() => {
        expect(
          screen.getByTestId("resume-detail-breadcrumb"),
        ).toBeInTheDocument();
      });
      expect(
        screen.getByTestId("resume-detail-branch-graph"),
      ).toBeInTheDocument();
      expect(screen.getByTestId("resume-detail-tab-preview")).toBeInTheDocument();
      expect(screen.getByTestId("resume-detail-tab-rewrites")).toBeInTheDocument();
      expect(screen.getByTestId("resume-detail-tab-edit")).toBeInTheDocument();
    });
  }

  it("renders NotFoundEmptyState when getResumeVersion returns 404 (UI copy is independent of fixture error.code spelling)", async () => {
    renderDetail("not-found-404", "ffffffff-0000-7000-8000-000000009000");

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-not-found"),
      ).toBeInTheDocument();
    });
    // The fixture's error.code is `TARGET_JOB_NOT_FOUND` (a copy gap), but
    // the UI must surface a generic not-found message rather than echoing
    // the wire code verbatim.
    const card = screen.getByTestId("resume-detail-not-found");
    expect(card).not.toHaveTextContent("TARGET_JOB_NOT_FOUND");
    expect(card).toHaveTextContent(/未找到|not found/i);
    expect(
      screen.getByTestId("resume-detail-not-found-back"),
    ).toBeInTheDocument();
  });
});
