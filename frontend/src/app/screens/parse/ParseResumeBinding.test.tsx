// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import {
  createFixtureBackedFetch,
  createFixtureRegistry,
  type OperationFixture,
} from "../../../api/mockTransport";
import { EasyInterviewClient } from "../../../api/generated/client";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { ParseScreen } from "./ParseScreen";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import createPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import getPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import startPracticeSessionFixture from "../../../../../openapi/fixtures/PracticeSessions/startPracticeSession.json";

const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";

function makeReadyFixture(overrides: Record<string, unknown> = {}) {
  const body = (
    getTargetJobFixture.scenarios.default as {
      response: { body: Record<string, unknown> };
    }
  ).response.body;
  return {
    operationId: "getTargetJob",
    scenarios: {
      default: {
        response: {
          status: 200,
          body: { ...body, analysisStatus: "ready" as const, ...overrides },
        },
      },
    },
  };
}

function createClient(
  fixtures: readonly OperationFixture[] = [listResumesFixture],
  targetOverrides: Record<string, unknown> = {},
) {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getRuntimeConfigFixture,
      getMeFixture,
      makeReadyFixture(targetOverrides),
      createPracticePlanFixture,
      getPracticePlanFixture,
      startPracticeSessionFixture,
      ...fixtures,
    ]),
    { scenario: "default" },
  );
  return new EasyInterviewClient({ fetch });
}

function renderParse(
  client: EasyInterviewClient,
  routeParams: Record<string, string> = {},
) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate }}>
            <ParseScreen
              route={{
                name: "workspace",
                params: { targetJobId: TARGET_JOB_ID, ...routeParams },
              }}
            />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

async function renderReadyParse(
  client: EasyInterviewClient,
  routeParams?: Record<string, string>,
) {
  return renderParse(client, routeParams);
}

describe("ParseResumeBinding", () => {
  it("shows the saved bound resume as readonly context", async () => {
    const client = createClient();
    const listSpy = vi.spyOn(client, "listResumes");

    await renderReadyParse(client);

    expect(await screen.findByTestId("parse-launch")).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByTestId("parse-resume-binding")).toHaveTextContent(
        "Resume saved with this interview plan",
      );
    });
    expect(screen.queryByTestId("parse-resume-picker-toggle")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-resume-picker")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-resume-create")).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-action-start-interview")).toBeEnabled();
    expect(listSpy).not.toHaveBeenCalled();
  });

  it("does not inherit route resumeId when the saved TargetJob lacks one", async () => {
    const client = createClient([listResumesFixture], {
      currentPracticePlanId: null,
      resumeId: null,
    });

    await renderReadyParse(client, {
      resumeId: "01918fa0-0000-7000-8000-000000001000",
    });

    await waitFor(() => {
      expect(screen.getByTestId("parse-resume-empty")).toBeInTheDocument();
    });
    expect(screen.getByTestId("parse-action-start-interview")).toBeDisabled();
    expect(screen.queryByTestId("parse-resume-option-01918fa0-0000-7000-8000-000000001000")).not.toBeInTheDocument();
  });

  it("blocks start when the saved plan has no bound resume without offering in-place binding", async () => {
    const client = createClient([listResumesFixture], {
      currentPracticePlanId: null,
      resumeId: null,
    });
    const listSpy = vi.spyOn(client, "listResumes");

    await renderReadyParse(client);

    expect(await screen.findByTestId("parse-launch")).toBeInTheDocument();
    expect(screen.getByTestId("parse-resume-binding")).toBeInTheDocument();
    expect(screen.getByTestId("parse-resume-empty")).toBeInTheDocument();
    expect(screen.queryByTestId("parse-resume-picker-toggle")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-resume-picker")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-resume-create")).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(
        "parse-resume-option-01918fa0-0000-7000-8000-000000001000",
      ),
    ).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-action-start-interview")).toBeDisabled();
    expect(listSpy).not.toHaveBeenCalled();
  });

  it("starts interview directly from workspace detail with the saved resumeId and no target patch", async () => {
    const client = createClient();
    const updateSpy = vi.spyOn(client, "updateTargetJob");
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const getPlanSpy = vi.spyOn(client, "getPracticePlan");
    const startSpy = vi.spyOn(client, "startPracticeSession");
    const { navigate } = await renderReadyParse(client);

    fireEvent.click(await screen.findByTestId("parse-action-start-interview"));

    await waitFor(() => {
      expect(startSpy).toHaveBeenCalledTimes(1);
    });

    expect(updateSpy).not.toHaveBeenCalled();
    expect(getPlanSpy).toHaveBeenCalledWith(
      "01918fa0-0000-7000-8000-000000004000",
    );
    expect(createSpy).not.toHaveBeenCalled();
    expect(startSpy).toHaveBeenCalledWith(
      {
        planId: "01918fa0-0000-7000-8000-000000004000",
      },
      expect.objectContaining({
        idempotencyKey: expect.stringMatching(/^v1\./),
      }),
    );
    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith({
        name: "practice",
        params: expect.objectContaining({
          targetJobId: "01918fa0-0000-7000-8000-000000002000",
          resumeId: "01918fa0-0000-7000-8000-000000001000",
          sessionId: "01918fa0-0000-7000-8000-000000005000",
          planId: "01918fa0-0000-7000-8000-000000004000",
        }),
      });
    });
    const params = navigate.mock.calls[0]?.[0].params as Record<string, string>;
    expect(params.autoStartPractice).toBeUndefined();
    expect(JSON.stringify(params)).not.toContain("resume-unbound");
  });

  it("does not request resume list or expose its failure from readonly detail", async () => {
    const client = createClient();
    const listSpy = vi.spyOn(client, "listResumes").mockRejectedValue(
      new Error("HTTP 503 RESUME_STORE_UNAVAILABLE"),
    );

    await renderReadyParse(client);

    expect(await screen.findByTestId("parse-launch")).toBeInTheDocument();
    expect(listSpy).not.toHaveBeenCalled();
    expect(screen.queryByText("HTTP 503 RESUME_STORE_UNAVAILABLE")).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-action-start-interview")).toBeEnabled();
  });

  it("does not expose a backend error when starting the interview fails", async () => {
    const client = createClient();
    vi.spyOn(client, "getPracticePlan").mockRejectedValue(
      new Error("HTTP 503 PRACTICE_STORE_UNAVAILABLE"),
    );
    const { navigate } = await renderReadyParse(client);

    fireEvent.click(await screen.findByTestId("parse-action-start-interview"));

    expect(await screen.findByTestId("parse-confirm-error")).toHaveTextContent(
      "The interview cannot be started right now. Please try again.",
    );
    expect(screen.queryByText("HTTP 503 PRACTICE_STORE_UNAVAILABLE")).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-action-start-interview")).toBeEnabled();
    expect(navigate).not.toHaveBeenCalledWith(expect.objectContaining({ name: "practice" }));
  });

  it("disables start for final backend progress with zero plan/session calls", async () => {
    const client = createClient([listResumesFixture], {
      currentPracticePlanId: null,
      practiceProgress: {
        status: "completed",
        completedRounds: [
          { roundId: "round-1-technical", roundSequence: 1 },
          { roundId: "round-2-manager", roundSequence: 2 },
          { roundId: "round-3-culture", roundSequence: 3 },
        ],
        currentRound: null,
      },
    });
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const startSpy = vi.spyOn(client, "startPracticeSession");

    await renderReadyParse(client);

    expect(await screen.findByTestId("parse-action-start-interview")).toBeDisabled();
    expect(createSpy).not.toHaveBeenCalled();
    expect(startSpy).not.toHaveBeenCalled();
  });
});
