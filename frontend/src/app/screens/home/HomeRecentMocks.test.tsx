// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import { createFixtureBackedFetch, createFixtureRegistry } from "../../../api/mockTransport";
import { EasyInterviewClient } from "../../../api/generated/client";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { HomeScreen } from "./HomeScreen";

import listTargetJobsFixture from "../../../../../openapi/fixtures/TargetJobs/listTargetJobs.json";

function createClient(scenario?: string) {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([listTargetJobsFixture]),
    scenario ? { scenario } : undefined,
  );
  return new EasyInterviewClient({ fetch });
}

function renderHome(client: EasyInterviewClient) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate }}>
            <HomeScreen route={{ name: "home", params: {} }} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("HomeRecentMocks", () => {
  it("calls listTargetJobs on mount", async () => {
    const client = createClient();
    const spy = vi.spyOn(client, "listTargetJobs");

    renderHome(client);

    await waitFor(() => {
      expect(spy).toHaveBeenCalled();
    });
  });

  it("renders 2 cards from default fixture", async () => {
    const client = createClient("default");
    renderHome(client);

    await waitFor(() => {
      expect(
        screen.getByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-000000002000"),
      ).toBeInTheDocument();
      expect(
        screen.getByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-000000002100"),
      ).toBeInTheDocument();
    });
  });

  it("shows empty state for empty variant", async () => {
    const client = createClient("empty");
    renderHome(client);

    await waitFor(() => {
      expect(screen.queryByTestId(/home-recent-mock-card-/)).not.toBeInTheDocument();
    });
  });

  it("renders single card for one-job variant", async () => {
    const client = createClient("one-job");
    renderHome(client);

    await waitFor(() => {
      expect(
        screen.getByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-000000000001"),
      ).toBeInTheDocument();
    });
  });

  it("renders at most 12 cards for twelve-plus variant", async () => {
    const client = createClient("twelve-plus");
    renderHome(client);

    await waitFor(() => {
      const cards = screen.queryAllByTestId(/home-recent-mock-card-/);
      expect(cards).toHaveLength(12);
      expect(cards[0]?.getAttribute("data-testid")).toBe(
        "home-recent-mock-card-01918fa0-0000-7000-8000-00000000a013",
      );
      expect(
        screen.getByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-00000000a013"),
      ).toBeInTheDocument();
      expect(
        screen.queryByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-00000000a001"),
      ).not.toBeInTheDocument();
    });
  });

  it("navigates to workspace on card click with interviewContext", async () => {
    const client = createClient("one-job");
    const { navigate } = renderHome(client);

    await waitFor(() => {
      expect(
        screen.getByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-000000000001"),
      ).toBeInTheDocument();
    });

    screen.getByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-000000000001").click();

    expect(navigate).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "workspace",
        params: expect.objectContaining({
          targetJobId: "01918fa0-0000-7000-8000-000000000001",
          jobId: "01918fa0-0000-7000-8000-000000000001",
          jdId: "jd-01918fa0-0000-7000-8000-000000000001",
          planId: "plan-01918fa0-0000-7000-8000-000000000001",
          resumeVersionId: "resume-unbound",
          roundId: "round-technical-1",
        }),
      }),
    );
  });
});
