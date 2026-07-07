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

const RESUME_ID = getResumeFixture.scenarios.default.response.body.id;

describe("resume detail export is not part of the read-only detail contract", () => {
  it("does not render Export PDF and does not call exportResume from detail", async () => {
    const client = new EasyInterviewClient({
      fetch: createFixtureBackedFetch(
        createFixtureRegistry(FIXTURES),
        { scenario: "default" },
      ),
    });
    const exportSpy = vi.spyOn(client, "exportResume");
    const route: Route = {
      name: "resume_versions",
      params: { resumeId: RESUME_ID },
    };

    render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider
          client={client}
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

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    expect(screen.queryByTestId("resume-detail-export-pdf")).not.toBeInTheDocument();
    expect(screen.queryByTestId("resume-detail-header-actions")).not.toBeInTheDocument();
    expect(exportSpy).not.toHaveBeenCalled();
  });
});
