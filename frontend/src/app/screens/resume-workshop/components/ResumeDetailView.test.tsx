// @vitest-environment jsdom
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

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

function renderDetail(
  scenario: string,
  route: Route,
  nav: ReturnType<typeof vi.fn> = vi.fn(),
): ReactNode {
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider
        client={buildClient(scenario)}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      >
        <NavigationProvider value={{ navigate: nav }}>
          <ResumeWorkshopScreen route={route} />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

const TARGETED_VERSION_ID =
  getResumeVersionFixture.scenarios.default.response.body.id;
const MASTER_VERSION_ID =
  getResumeVersionFixture.scenarios["master-default"].response.body.id;

describe("ResumeDetailView container (Phase 3.1)", () => {
  it("renders breadcrumb, branch graph, and three tabs once the version loads", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { versionId: TARGETED_VERSION_ID, tab: "preview" },
    });

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

  it("MASTER version (master-default scenario) defaults the active tab to preview", async () => {
    renderDetail("master-default", {
      name: "resume_versions",
      params: { versionId: MASTER_VERSION_ID },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-tab-preview")).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });
    expect(
      screen.getByTestId("resume-detail-preview-content"),
    ).toBeInTheDocument();
  });

  it("TARGETED version with explicit tab=rewrites in the URL keeps rewrites active and renders the ComingSoonTab placeholder", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { versionId: TARGETED_VERSION_ID, tab: "rewrites" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-tab-rewrites"),
      ).toBeInTheDocument();
    });
    expect(screen.getByTestId("resume-detail-tab-rewrites")).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(
      screen.getByTestId("resume-detail-tab-content-coming-soon-rewrites"),
    ).toBeInTheDocument();
  });

  it("clicking a tab updates the active selection and shows that tab's content (preview from default route)", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { versionId: TARGETED_VERSION_ID, tab: "rewrites" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-tab-preview"),
      ).toBeInTheDocument();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-detail-tab-preview"));

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-tab-preview")).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });
    expect(
      screen.getByTestId("resume-detail-preview-content"),
    ).toBeInTheDocument();
  });
});
