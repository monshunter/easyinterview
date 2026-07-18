// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { createFixtureBackedFetch, createFixtureRegistry } from "../../../api/mockTransport";
import { EasyInterviewClient } from "../../../api/generated/client";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { HomeScreen } from "./HomeScreen";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import listTargetJobsFixture from "../../../../../openapi/fixtures/TargetJobs/listTargetJobs.json";
import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import createPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import getPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import startPracticeSessionFixture from "../../../../../openapi/fixtures/PracticeSessions/startPracticeSession.json";

type ListResumesResponse = Awaited<ReturnType<EasyInterviewClient["listResumes"]>>;
type ListTargetJobsResponse = Awaited<
  ReturnType<EasyInterviewClient["listTargetJobs"]>
>;

const defaultListResumesResponse = listResumesFixture.scenarios.default.response
  .body as ListResumesResponse;
const defaultListTargetJobsResponse = listTargetJobsFixture.scenarios.default
  .response.body as ListTargetJobsResponse;

function createClient(scenario?: string) {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getRuntimeConfigFixture,
      getMeFixture,
      listResumesFixture,
      listTargetJobsFixture,
      getTargetJobFixture,
      createPracticePlanFixture,
    ]),
    scenario ? { scenario } : undefined,
  );
  const client = new EasyInterviewClient({ fetch });
  vi.spyOn(client, "listResumes").mockResolvedValue(defaultListResumesResponse);
  return client;
}

function renderHome(
  client: EasyInterviewClient,
  options?: {
    getMeScenario?: "authenticated" | "unauthenticated";
  },
) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider
          client={client}
          requestOptions={{
            getMe: {
              headers: {
                Prefer: `example=${options?.getMeScenario ?? "authenticated"}`,
              },
            },
          }}
        >
          <NavigationProvider value={{ navigate }}>
            <HomeScreen route={{ name: "home", params: {} }} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("HomeRecentMocks", () => {
  it("does not render recent mocks or call listTargetJobs when signed out", async () => {
    const client = createClient();
    const getMeSpy = vi.spyOn(client, "getMe");
    const listSpy = vi.spyOn(client, "listTargetJobs");

    renderHome(client, { getMeScenario: "unauthenticated" });

    await waitFor(() => {
      expect(getMeSpy).toHaveBeenCalled();
    });
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(screen.queryByText("Recent mock interviews")).not.toBeInTheDocument();
    expect(screen.queryByTestId("home-recent-mocks")).not.toBeInTheDocument();
    expect(screen.queryByText(/AUTH_UNAUTHORIZED|Unauthorized|authentication required/i)).not.toBeInTheDocument();
    expect(listSpy).not.toHaveBeenCalled();
  });

  it("calls listTargetJobs on mount", async () => {
    const client = createClient();
    const spy = vi.spyOn(client, "listTargetJobs");

    renderHome(client);

    await waitFor(() => {
      expect(spy).toHaveBeenCalledWith({
        query: {
          analysisStatus: "ready",
          pageSize: "12",
        },
      });
    });
  });

  it("renders a user-safe recent-plan load failure", async () => {
    const client = createClient();
    vi.spyOn(client, "listTargetJobs").mockRejectedValue(
      new Error("HTTP 503 TARGET_JOB_STORE_UNAVAILABLE"),
    );

    renderHome(client);

    expect(await screen.findByText("We couldn't load your recent interview plans. Try again in a moment.")).toBeInTheDocument();
    expect(screen.queryByText("HTTP 503 TARGET_JOB_STORE_UNAVAILABLE")).not.toBeInTheDocument();
  });

  it("filters non-ready or blank-title recent jobs", async () => {
    const client = createClient();
    const readyJob = defaultListTargetJobsResponse.items[0]!;
    const queuedJob: ListTargetJobsResponse["items"][number] = {
      ...readyJob,
      id: "01918fa0-0000-7000-8000-000000009991",
      title: "Queued JD",
      analysisStatus: "processing",
    };
    const blankJob: ListTargetJobsResponse["items"][number] = {
      ...readyJob,
      id: "01918fa0-0000-7000-8000-000000009992",
      title: "   ",
      analysisStatus: "ready",
    };
    vi.spyOn(client, "listTargetJobs").mockResolvedValue({
      ...defaultListTargetJobsResponse,
      items: [queuedJob, blankJob, readyJob],
    });

    renderHome(client);

    await waitFor(() => {
      expect(
        screen.getByTestId(`home-recent-mock-card-${readyJob.id}`),
      ).toBeInTheDocument();
    });

    expect(screen.queryByText("Queued JD")).toBeNull();
    expect(
      screen.queryByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-000000009991"),
    ).toBeNull();
    expect(
      screen.queryByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-000000009992"),
    ).toBeNull();
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
    const grid = screen.getByTestId("home-recent-mock-grid");
    expect(grid.style.gridTemplateColumns).toContain("360px");
    expect(grid.style.gridTemplateColumns).not.toContain("1fr");
    expect(grid.style.justifyContent).toBe("start");
  });

  it("renders at most 3 cards for twelve-plus variant and exposes More navigation", async () => {
    const client = createClient("twelve-plus");
    const { navigate } = renderHome(client);

    await waitFor(() => {
      const cards = screen.queryAllByTestId(/home-recent-mock-card-/);
      expect(cards).toHaveLength(3);
      expect(cards[0]?.getAttribute("data-testid")).toBe(
        "home-recent-mock-card-01918fa0-0000-7000-8000-00000000a013",
      );
      expect(
        screen.getByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-00000000a013"),
      ).toBeInTheDocument();
      expect(
        screen.queryByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-00000000a010"),
      ).not.toBeInTheDocument();
    });

    await userEvent.click(screen.getByTestId("home-recent-more"));

    expect(navigate).toHaveBeenCalledWith({
      name: "workspace",
      params: {},
    });
  });

  it("opens workspace detail on card click and shows quick-start without delete", async () => {
    const client = createClient("default");
    const { navigate } = renderHome(client);

    await waitFor(() => {
      expect(
        screen.getByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-000000002000"),
      ).toBeInTheDocument();
    });

    expect(
      screen.getByTestId("home-recent-mock-start-01918fa0-0000-7000-8000-000000002000"),
    ).toHaveTextContent("Start a mock interview");
    expect(
      screen.queryByTestId("home-recent-mock-delete-01918fa0-0000-7000-8000-000000002000"),
    ).toBeNull();

    screen.getByTestId("home-recent-mock-card-01918fa0-0000-7000-8000-000000002000").click();

    expect(navigate).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "workspace",
        params: {
          targetJobId: "01918fa0-0000-7000-8000-000000002000",
        },
      }),
    );
  });

  it("quick-starts a recent mock without opening parse detail", async () => {
    const user = userEvent.setup();
    const client = createClient("default");
    const getPlanSpy = vi
      .spyOn(client, "getPracticePlan")
      .mockResolvedValue(
        getPracticePlanFixture.scenarios.default.response.body as Awaited<
          ReturnType<EasyInterviewClient["getPracticePlan"]>
        >,
      );
    const createPlanSpy = vi.spyOn(client, "createPracticePlan");
    const startSpy = vi
      .spyOn(client, "startPracticeSession")
      .mockResolvedValue(
        startPracticeSessionFixture.scenarios.default.response.body as Awaited<
          ReturnType<EasyInterviewClient["startPracticeSession"]>
        >,
      );
    const { navigate } = renderHome(client);

    await waitFor(() => {
      expect(
        screen.getByTestId("home-recent-mock-start-01918fa0-0000-7000-8000-000000002000"),
      ).toBeInTheDocument();
    });

    await user.click(
      screen.getByTestId("home-recent-mock-start-01918fa0-0000-7000-8000-000000002000"),
    );

    await waitFor(() => {
      expect(startSpy).toHaveBeenCalled();
    });

    expect(getPlanSpy).toHaveBeenCalledWith("01918fa0-0000-7000-8000-000000004000");
    expect(createPlanSpy).not.toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith({
      name: "practice",
      params: expect.objectContaining({
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
        planId: "01918fa0-0000-7000-8000-000000004000",
        resumeId: "01918fa0-0000-7000-8000-000000001000",
        roundId: "round-2-manager",
        roundName: "Hiring manager impact interview · 50m",
        sessionId: "01918fa0-0000-7000-8000-000000005000",
      }),
    });
    expect(navigate).not.toHaveBeenCalledWith(
      expect.objectContaining({ name: "parse" }),
    );
  });

  it("shows the shared launch transition while a recent mock waits for its opening message", async () => {
    const user = userEvent.setup();
    const client = createClient("default");
    vi.spyOn(client, "getPracticePlan").mockResolvedValue(
      getPracticePlanFixture.scenarios.default.response.body as Awaited<
        ReturnType<EasyInterviewClient["getPracticePlan"]>
      >,
    );
    let resolveStart!: (value: Awaited<ReturnType<EasyInterviewClient["startPracticeSession"]>>) => void;
    const startSpy = vi.spyOn(client, "startPracticeSession").mockImplementation(
      () => new Promise((resolve) => {
        resolveStart = resolve;
      }),
    );
    const { navigate } = renderHome(client);
    const start = await screen.findByTestId(
      "home-recent-mock-start-01918fa0-0000-7000-8000-000000002000",
    );

    await user.click(start);
    await waitFor(() => expect(startSpy).toHaveBeenCalledTimes(1));

    const transition = screen.getByTestId("practice-launch-transition");
    expect(transition).toHaveAttribute("role", "status");
    expect(transition).toHaveAttribute("aria-busy", "true");
    expect(transition).toHaveTextContent("Preparing your interview");
    expect(transition).not.toHaveTextContent(/%|opening message/i);
    expect(start).toBeDisabled();
    expect(navigate).not.toHaveBeenCalledWith(expect.objectContaining({ name: "practice" }));

    await act(async () => {
      resolveStart(
        startPracticeSessionFixture.scenarios.default.response.body as Awaited<
          ReturnType<EasyInterviewClient["startPracticeSession"]>
        >,
      );
    });
    await waitFor(() => expect(screen.queryByTestId("practice-launch-transition")).toBeNull());
    expect(navigate).toHaveBeenCalledWith(expect.objectContaining({ name: "practice" }));
  });

  it("does not expose a backend error when recent quick-start fails", async () => {
    const user = userEvent.setup();
    const client = createClient("default");
    vi.spyOn(client, "getPracticePlan").mockRejectedValue(
      new Error("HTTP 503 PRACTICE_STORE_UNAVAILABLE"),
    );
    const { navigate } = renderHome(client);

    const start = await screen.findByTestId(
      "home-recent-mock-start-01918fa0-0000-7000-8000-000000002000",
    );
    await user.click(start);

    expect(await screen.findByTestId("home-recent-start-error")).toHaveTextContent(
      "We couldn't start the mock interview. Try again in a moment.",
    );
    expect(screen.queryByText("HTTP 503 PRACTICE_STORE_UNAVAILABLE")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-launch-transition")).not.toBeInTheDocument();
    expect(start).toBeEnabled();
    expect(navigate).not.toHaveBeenCalledWith(expect.objectContaining({ name: "practice" }));
  });

  it("disables quick-start after final backend progress with zero plan/session calls", async () => {
    const client = createClient("default");
    const finished = {
      ...defaultListTargetJobsResponse.items[0]!,
      currentPracticePlanId: null,
      practiceProgress: {
        status: "completed" as const,
        completedRounds: [
          { roundId: "round-1-technical", roundSequence: 1 },
          { roundId: "round-2-manager", roundSequence: 2 },
          { roundId: "round-3-culture", roundSequence: 3 },
        ],
        currentRound: null,
      },
    };
    vi.spyOn(client, "listTargetJobs").mockResolvedValue({
      ...defaultListTargetJobsResponse,
      items: [finished],
    });
    const getTargetSpy = vi.spyOn(client, "getTargetJob");
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const startSpy = vi.spyOn(client, "startPracticeSession");

    renderHome(client);

    const button = await screen.findByTestId(`home-recent-mock-start-${finished.id}`);
    expect(button).toBeDisabled();
    expect(getTargetSpy).not.toHaveBeenCalled();
    expect(createSpy).not.toHaveBeenCalled();
    expect(startSpy).not.toHaveBeenCalled();
  });
});
