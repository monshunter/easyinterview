// @vitest-environment jsdom
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
import listResumesFixture from "../../../../../../openapi/fixtures/Resumes/listResumes.json";

const FIXTURES = [getRuntimeConfigFixture, getMeFixture, listResumesFixture];

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario,
    }),
  });
}

function renderListView(route: Route, scenario = "default") {
  const nav = vi.fn();
  const result = render(
    <DisplayPreferencesProvider initial={{ lang: "zh" }}>
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
  return { ...result, nav };
}

const LIST_ROUTE: Route = { name: "resume_versions", params: {} };

const FIRST_ID =
  listResumesFixture.scenarios.default.response.body.items[0]!.id;
const SECOND_ID =
  listResumesFixture.scenarios.default.response.body.items[1]!.id;

describe("ResumeListView default fixture rendering", () => {
  it("renders the flat table with one row per resume from the default fixture", async () => {
    renderListView(LIST_ROUTE);

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-table")).toBeInTheDocument();
    });

    expect(
      screen.getByTestId(`resume-list-row-${FIRST_ID}`),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId(`resume-list-row-${SECOND_ID}`),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId(`resume-list-open-${FIRST_ID}`),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId(`resume-list-open-${SECOND_ID}`),
    ).toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-create")).toHaveTextContent(
      "新建简历",
    );
  });

  it("opens a resume detail via the row Open button with the resumeId + preview tab", async () => {
    const { nav } = renderListView(LIST_ROUTE);

    await waitFor(() => {
      expect(
        screen.getByTestId(`resume-list-open-${FIRST_ID}`),
      ).toBeInTheDocument();
    });

    await userEvent
      .setup()
      .click(screen.getByTestId(`resume-list-open-${FIRST_ID}`));
    expect(nav).toHaveBeenCalledWith({
      name: "resume_versions",
      params: { resumeId: FIRST_ID, tab: "preview" },
    });
  });

  it("navigates to the create flow from the header button", async () => {
    const { nav } = renderListView(LIST_ROUTE);

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-create")).toBeInTheDocument();
    });

    await userEvent.setup().click(screen.getByTestId("resume-workshop-create"));
    expect(nav).toHaveBeenCalledWith({
      name: "resume_versions",
      params: { flow: "create" },
    });
  });

  it("navigates to the create flow from the upload-another CTA", async () => {
    const { nav } = renderListView(LIST_ROUTE);

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-upload-cta"),
      ).toBeInTheDocument();
    });

    await userEvent
      .setup()
      .click(screen.getByTestId("resume-workshop-upload-cta"));
    expect(nav).toHaveBeenCalledWith({
      name: "resume_versions",
      params: { flow: "create" },
    });
  });

  it("shows the empty state when listResumes returns no items", async () => {
    renderListView(LIST_ROUTE, "empty");

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-list-empty"),
      ).toBeInTheDocument();
    });
    expect(screen.queryByTestId("resume-workshop-table")).not.toBeInTheDocument();
  });

  it("surfaces the paginated affordance when the page reports more results", async () => {
    renderListView(LIST_ROUTE, "paginated");

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-list-paginated"),
      ).toBeInTheDocument();
    });
  });
});
