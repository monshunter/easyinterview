// @vitest-environment jsdom
import type { ReactNode } from "react";
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
import listResumesFixture from "../../../../../../openapi/fixtures/Resumes/listResumes.json";
import listResumeVersionsFixture from "../../../../../../openapi/fixtures/Resumes/listResumeVersions.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  listResumeVersionsFixture,
];

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry(FIXTURES),
      { scenario: "default" },
    ),
  });
}

function renderListView(route: Route): ReactNode {
  const client = buildClient();
  const nav = vi.fn();
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider
        client={client}
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

const LIST_ROUTE: Route = { name: "resume_versions", params: {} };

describe("ResumeListView default fixture rendering", () => {
  it("renders StatsStrip + ViewSwitcher + dispatched tree view (≥ 8 anchored testids on default)", async () => {
    renderListView(LIST_ROUTE);

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-stats-originals"),
      ).toBeInTheDocument();
    });

    expect(
      screen.getByTestId("resume-workshop-stats-originals"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-workshop-stats-versions"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-workshop-stats-top-match"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-workshop-stats-recent"),
    ).toBeInTheDocument();

    expect(
      screen.getByTestId("resume-workshop-view-switcher-tree"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-workshop-view-switcher-flat"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-workshop-view-switcher-tree"),
    ).toHaveAttribute("data-active", "true");

    // At least one tree row (asset entry) is rendered for the default fixture
    const firstAssetId =
      listResumesFixture.scenarios.default.response.body.items[0]?.id ?? "";
    expect(firstAssetId).toBeTruthy();
    expect(
      screen.getByTestId(`resume-tree-row-${firstAssetId}`),
    ).toBeInTheDocument();

    // The asset that has matching versions in the default scenario must surface
    // at least one resume-version-row
    const matchedVersionId =
      listResumeVersionsFixture.scenarios.default.response.body.items[0]?.id ??
      "";
    expect(matchedVersionId).toBeTruthy();
    await waitFor(() => {
      expect(
        screen.getByTestId(`resume-version-row-${matchedVersionId}`),
      ).toBeInTheDocument();
    });
  });

  it("derives stats counts from fixture body, not from static prototype", async () => {
    renderListView(LIST_ROUTE);

    const expectedOriginalCount =
      listResumesFixture.scenarios.default.response.body.items.length;
    const expectedVersionCount =
      listResumeVersionsFixture.scenarios.default.response.body.items.length;

    await waitFor(() => {
      const originals = screen.getByTestId("resume-workshop-stats-originals");
      const versions = screen.getByTestId("resume-workshop-stats-versions");
      expect(originals).toHaveTextContent(String(expectedOriginalCount));
      expect(versions).toHaveTextContent(String(expectedVersionCount));
    });
  });
});
