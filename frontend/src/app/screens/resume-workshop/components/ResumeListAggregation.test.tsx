// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import { EasyInterviewClient } from "../../../../api/generated/client";
import type { PaginatedResumeVersion } from "../../../../api/generated/types";
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

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario,
    }),
  });
}

function renderListView(scenario: string) {
  return renderListViewWithClient(buildClient(scenario));
}

function renderListViewWithClient(client: EasyInterviewClient) {
  const route: Route = { name: "resume_versions", params: {} };
  return render(
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
}

const FIRST_ASSET_ID =
  listResumesFixture.scenarios.default.response.body.items[0]?.id ?? "";
const SECOND_ASSET_ID =
  listResumesFixture.scenarios.default.response.body.items[1]?.id ?? "";

describe("Resume list version aggregation boundaries", () => {
  it("empty scenario renders the list-empty placeholder and does not request listResumeVersions", async () => {
    // listResumeVersions has no `empty` scenario; absence of the call proves
    // the hook never fires when the asset list is empty.
    const client = buildClient("empty");
    const versionsSpy = vi.spyOn(client, "listResumeVersions");
    const route: Route = { name: "resume_versions", params: {} };
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
      expect(
        screen.getByTestId("resume-workshop-list-empty"),
      ).toBeInTheDocument();
    });
    expect(versionsSpy).not.toHaveBeenCalled();
    expect(
      screen.queryByTestId("resume-workshop-list-paginated"),
    ).not.toBeInTheDocument();
  });

  it("paginated scenario renders the continue-loading hint when hasMore=true", async () => {
    const client = buildClient("paginated");
    vi.spyOn(client, "listResumeVersions").mockResolvedValue({
      items: [],
      pageInfo: { nextCursor: null, pageSize: 50, hasMore: false },
    } satisfies PaginatedResumeVersion);
    renderListViewWithClient(client);
    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-list-paginated"),
      ).toBeInTheDocument();
    });
  });

  it("default scenario shows a no-versions indicator for the second asset (which has no matching version)", async () => {
    renderListView("default");
    await waitFor(() => {
      expect(
        screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}`),
      ).toBeInTheDocument();
    });
    // First asset has matching versions
    expect(
      screen.getByTestId(`resume-tree-row-${SECOND_ASSET_ID}`),
    ).toBeInTheDocument();
    // Second asset must surface a no-versions / partial state (not fabricated rows)
    expect(
      screen.getByTestId(`resume-tree-row-${SECOND_ASSET_ID}-no-versions`),
    ).toBeInTheDocument();
  });
});
