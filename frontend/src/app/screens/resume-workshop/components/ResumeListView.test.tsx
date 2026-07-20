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
import archiveResumeFixture from "../../../../../../openapi/fixtures/Resumes/archiveResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  archiveResumeFixture,
];

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario,
    }),
  });
}

function renderListView(
  route: Route,
  scenario = "default",
  client: EasyInterviewClient = buildClient(scenario),
) {
  const nav = vi.fn();
  const result = render(
    <DisplayPreferencesProvider initial={{ lang: "zh" }}>
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
  return { ...result, nav };
}

const LIST_ROUTE: Route = { name: "resume_versions", params: {} };

const FIRST_ID =
  listResumesFixture.scenarios.default.response.body.items[0]!.id;
const SECOND_ID =
  listResumesFixture.scenarios.default.response.body.items[1]!.id;

describe("ResumeListView default fixture rendering", () => {
  it("renders a semantic card grid with one list item per closed resume summary", async () => {
    renderListView(LIST_ROUTE);

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-card-grid")).toBeInTheDocument();
    });

    const grid = screen.getByRole("list", { name: "你的简历" });
    expect(grid).toHaveAttribute("data-testid", "resume-workshop-card-grid");
    expect(screen.getAllByRole("listitem")).toHaveLength(2);
    expect(screen.getByTestId(`resume-list-card-${FIRST_ID}`)).toHaveTextContent(
      "Senior frontend engineer focused on growth-stage SaaS",
    );
    expect(screen.getByTestId(`resume-list-card-${SECOND_ID}`)).toHaveTextContent(
      "Frontend platform engineer with product systems scope",
    );
    expect(
      screen.getByTestId(`resume-list-card-${FIRST_ID}`).querySelector(
        ".ei-resume-workshop-card-icon",
      ),
    ).not.toBeNull();
    expect(
      screen.getByTestId(`resume-list-card-${FIRST_ID}`).querySelector(
        ".ei-resume-workshop-lang-tag",
      ),
    ).toBeNull();
    expect(
      screen.getByTestId(`resume-list-open-${FIRST_ID}`),
    ).toHaveAccessibleName("打开 Alice Example - Senior Frontend Engineer");
    expect(
      screen.getByTestId(`resume-list-delete-${FIRST_ID}`),
    ).toHaveAccessibleName("删除简历 Alice Example - Senior Frontend Engineer");
    expect(screen.queryByRole("row")).not.toBeInTheDocument();
    expect(screen.queryByRole("columnheader")).not.toBeInTheDocument();
    expect(screen.queryByTestId("resume-workshop-table")).not.toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-create")).toHaveTextContent(
      "新建简历",
    );
    const createIcon = screen
      .getByTestId("resume-workshop-create")
      .querySelector('svg[data-icon="plus"]');
    expect(createIcon).toHaveAttribute("width", "22");
    expect(createIcon?.querySelector("circle")).not.toBeNull();
    expect(screen.getByTestId("resume-workshop-screen")).toHaveClass(
      "ei-resume-workshop-screen",
    );
  });

  it("opens a resume detail via the row Open button with only the resumeId", async () => {
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
      params: { resumeId: FIRST_ID },
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

  it("does not render the duplicate upload-or-paste CTA below the card grid", async () => {
    renderListView(LIST_ROUTE);

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-card-grid")).toBeInTheDocument();
    });

    expect(screen.queryByTestId("resume-workshop-upload-cta")).not.toBeInTheDocument();
    expect(screen.queryByText("上传或粘贴另一份简历")).not.toBeInTheDocument();
  });

  it("archives a resume from the row delete action and hides it from the list", async () => {
    const client = buildClient("default");
    const archiveSpy = vi.spyOn(client, "archiveResume");
    const user = userEvent.setup();

    renderListView(LIST_ROUTE, "default", client);
    await waitFor(() => {
      expect(screen.getByTestId(`resume-list-card-${FIRST_ID}`)).toBeInTheDocument();
    });

    await user.click(screen.getByTestId(`resume-list-delete-${FIRST_ID}`));

    expect(archiveSpy).not.toHaveBeenCalled();
    expect(
      screen.getByRole("dialog", { name: "确认删除这份简历？" }),
    ).toHaveTextContent("删除后，这份简历会从简历列表中移除。此操作当前无法撤销。");
    await user.click(screen.getByRole("button", { name: "取消" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(archiveSpy).not.toHaveBeenCalled();
    expect(screen.getByTestId(`resume-list-delete-${FIRST_ID}`)).toHaveFocus();

    await user.click(screen.getByTestId(`resume-list-delete-${FIRST_ID}`));
    await user.click(screen.getByRole("button", { name: "确认删除" }));

    await waitFor(() => {
      expect(screen.queryByTestId(`resume-list-card-${FIRST_ID}`)).not.toBeInTheDocument();
    });
    expect(archiveSpy).toHaveBeenCalledWith(
      FIRST_ID,
      expect.objectContaining({
        idempotencyKey: expect.stringMatching(/^v1\.\d+\.[0-9a-f-]{36}$/),
      }),
    );
  });

  it("keeps the row visible and shows an error when archiveResume fails", async () => {
    const client = buildClient("default");
    const archiveSpy = vi
      .spyOn(client, "archiveResume")
      .mockRejectedValueOnce(new Error("HTTP 500 archive failed"))
      .mockResolvedValueOnce(
        archiveResumeFixture.scenarios.default.response.body as Awaited<
          ReturnType<EasyInterviewClient["archiveResume"]>
        >,
      );
    const user = userEvent.setup();

    renderListView(LIST_ROUTE, "default", client);
    await waitFor(() => {
      expect(screen.getByTestId(`resume-list-card-${FIRST_ID}`)).toBeInTheDocument();
    });

    await user.click(screen.getByTestId(`resume-list-delete-${FIRST_ID}`));
    expect(archiveSpy).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "确认删除" }));

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-delete-error")).toBeInTheDocument();
    });
    expect(screen.getByTestId(`resume-list-card-${FIRST_ID}`)).toBeInTheDocument();
    expect(screen.getByRole("dialog", { name: "确认删除这份简历？" })).toBeInTheDocument();
    const firstKey = archiveSpy.mock.calls[0]?.[1]?.idempotencyKey;
    await user.click(screen.getByRole("button", { name: "重试" }));
    expect(archiveSpy.mock.calls[1]?.[1]?.idempotencyKey).toBe(firstKey);
    await waitFor(() => {
      expect(screen.queryByTestId(`resume-list-card-${FIRST_ID}`)).not.toBeInTheDocument();
    });
  });

  it("shows the empty state when listResumes returns no items", async () => {
    renderListView(LIST_ROUTE, "empty");

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-list-empty"),
      ).toBeInTheDocument();
    });
    expect(screen.queryByTestId("resume-workshop-card-grid")).not.toBeInTheDocument();
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
