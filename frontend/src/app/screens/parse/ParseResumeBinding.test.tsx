// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

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
import updateTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/updateTargetJob.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";

const LOADING_PREVIEW_DELAY = 3200;

function makeReadyFixture() {
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
          body: { ...body, analysisStatus: "ready" as const },
        },
      },
    },
  };
}

function createClient(
  fixtures: readonly OperationFixture[] = [listResumesFixture],
) {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getRuntimeConfigFixture,
      getMeFixture,
      makeReadyFixture(),
      updateTargetJobFixture,
      ...fixtures,
    ]),
    { scenario: "default" },
  );
  return new EasyInterviewClient({ fetch });
}

function emptyListResumesFixture(): OperationFixture {
  const emptyScenario = listResumesFixture.scenarios.empty;
  return {
    operationId: "listResumes",
    scenarios: {
      default: emptyScenario,
    },
  };
}

function renderParse(client: EasyInterviewClient) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate }}>
            <ParseScreen
              route={{ name: "parse", params: { targetJobId: "tj-1" } }}
            />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

async function renderReadyParse(client: EasyInterviewClient) {
  vi.useFakeTimers();
  const result = renderParse(client);

  await act(async () => {
    await vi.advanceTimersByTimeAsync(LOADING_PREVIEW_DELAY);
  });
  vi.useRealTimers();

  return result;
}

afterEach(() => {
  vi.useRealTimers();
});

describe("ParseResumeBinding", () => {
  it("loads ready resumes but requires an explicit resume selection before any handoff", async () => {
    const client = createClient();
    const listSpy = vi.spyOn(client, "listResumes");

    await renderReadyParse(client);

    await waitFor(() => {
      expect(listSpy).toHaveBeenCalledTimes(1);
    });

    expect(await screen.findByTestId("parse-launch")).toBeInTheDocument();
    expect(screen.getByTestId("parse-resume-binding")).toHaveTextContent(
      "Choose the resume for this interview",
    );
    expect(screen.getByTestId("parse-resume-required")).toBeInTheDocument();
    expect(screen.queryByTestId("parse-resume-picker-toggle")).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-action-save-plan")).toBeDisabled();
    expect(screen.getByTestId("parse-action-start-interview")).toBeDisabled();
    expect(
      screen.getByTestId(
        "parse-resume-option-01918fa0-0000-7000-8000-000000001000",
      ),
    ).toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-confirm")).not.toBeInTheDocument();
  });

  it("enables launch actions only after the user chooses a ready resume", async () => {
    const client = createClient();
    await renderReadyParse(client);

    expect(await screen.findByTestId("parse-resume-required")).toBeInTheDocument();
    expect(screen.getByTestId("parse-action-save-plan")).toBeDisabled();

    fireEvent.click(
      screen.getByTestId(
        "parse-resume-option-0195f2d0-0000-7000-8000-000000001010",
      ),
    );

    expect(screen.getByTestId("parse-resume-binding")).toHaveTextContent(
      "Alice Example - Product Platform Resume",
    );
    expect(screen.getByTestId("parse-action-save-plan")).toBeEnabled();
    expect(screen.getByTestId("parse-action-start-interview")).toBeEnabled();
  });

  it("blocks save and start when no ready resume exists and routes to resume creation", async () => {
    const client = createClient([emptyListResumesFixture()]);
    const { navigate } = await renderReadyParse(client);

    expect(await screen.findByTestId("parse-resume-empty")).toBeInTheDocument();
    expect(screen.getByTestId("parse-action-save-plan")).toBeDisabled();
    expect(screen.getByTestId("parse-action-start-interview")).toBeDisabled();

    fireEvent.click(screen.getByTestId("parse-resume-create"));

    expect(navigate).toHaveBeenCalledWith({
      name: "resume_versions",
      params: {
        flow: "create",
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
      },
    });
  });

  it("starts interview through workspace auto-start with a real resumeId", async () => {
    const client = createClient();
    const updateSpy = vi.spyOn(client, "updateTargetJob");
    const { navigate } = await renderReadyParse(client);

    fireEvent.click(
      await screen.findByTestId(
        "parse-resume-option-0195f2d0-0000-7000-8000-000000001010",
      ),
    );
    fireEvent.click(await screen.findByTestId("parse-action-start-interview"));

    await waitFor(() => {
      expect(updateSpy).toHaveBeenCalledTimes(1);
    });

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith({
        name: "workspace",
        params: expect.objectContaining({
          targetJobId: "01918fa0-0000-7000-8000-000000002000",
          resumeId: "0195f2d0-0000-7000-8000-000000001010",
          autoStartPractice: "1",
          practiceMode: "strict",
          mode: "text",
          modality: "text",
        }),
      });
    });
    const params = navigate.mock.calls[0]?.[0].params as Record<string, string>;
    expect(JSON.stringify(params)).not.toContain("resume-unbound");
  });
});
