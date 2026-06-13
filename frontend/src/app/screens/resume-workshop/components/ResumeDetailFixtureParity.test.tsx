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
import getResumeFixture from "../../../../../../openapi/fixtures/Resumes/getResume.json";
import exportResumeFixture from "../../../../../../openapi/fixtures/Resumes/exportResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  getResumeFixture,
  exportResumeFixture,
];

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario,
    }),
  });
}

function renderDetail(scenario: string, resumeId: string) {
  const route: Route = {
    name: "resume_versions",
    params: { resumeId },
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

const RESUME_ID = getResumeFixture.scenarios.default.response.body.id;

describe("getResume fixture parity (Phase 3.6)", () => {
  it("renders the detail container with crumb + tabs for the default scenario", async () => {
    renderDetail("default", RESUME_ID);

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    expect(screen.getByTestId("resume-detail-tab-preview")).toBeInTheDocument();
    expect(screen.getByTestId("resume-detail-tab-rewrites")).toBeInTheDocument();
    expect(screen.getByTestId("resume-detail-tab-edit")).toBeInTheDocument();
  });

  it("renders NotFoundEmptyState when getResume returns 404 (UI copy is independent of fixture error.code spelling)", async () => {
    renderDetail("not-found", RESUME_ID);

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-not-found"),
      ).toBeInTheDocument();
    });
    // The fixture's error.code is a wire code; the UI must surface a generic
    // not-found message rather than echoing the wire code verbatim.
    const card = screen.getByTestId("resume-detail-not-found");
    expect(card).not.toHaveTextContent("RESOURCE_NOT_FOUND");
    expect(card).toHaveTextContent(/未找到|not found/i);
    expect(
      screen.getByTestId("resume-detail-not-found-back"),
    ).toBeInTheDocument();
  });
});
