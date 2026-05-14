// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { FC } from "react";

import getMeFixture from "../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import startAuthEmailChallengeFixture from "../../../../openapi/fixtures/Auth/startAuthEmailChallenge.json";
import verifyAuthEmailChallengeFixture from "../../../../openapi/fixtures/Auth/verifyAuthEmailChallenge.json";
import { EasyInterviewClient } from "../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../api/mockTransport";
import { App } from "../App";
import type { PendingAction } from "./pendingAction";
import { useRequestAuth } from "./useRequestAuth";

const SAMPLE_ACTION: PendingAction = {
  type: "start_practice",
  label: "立即面试",
  route: "practice",
  params: {
    sessionId: "01918fa0-0000-7000-8000-000000005000",
    planId: "plan-tj-1",
    targetJobId: "tj-1",
    jdId: "jd-tj-1",
    resumeVersionId: "frontend-v3",
    roundId: "round-manager",
  },
};

const TriggerProbe: FC = () => {
  const requestAuth = useRequestAuth();
  return (
    <button
      type="button"
      data-testid="probe-start-practice"
      onClick={() => requestAuth(SAMPLE_ACTION)}
    >
      立即面试
    </button>
  );
};

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        startAuthEmailChallengeFixture,
        verifyAuthEmailChallengeFixture,
      ]),
    ),
  });
}

describe("requestAuth pending-action flow", () => {
  it("redirects to auth_login carrying the encoded pending action when not signed in", async () => {
    render(
      <App
        client={buildClient()}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      >
        <TriggerProbe />
      </App>,
    );

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "false",
      ),
    );

    const user = userEvent.setup();
    await user.click(screen.getByTestId("probe-start-practice"));
    expect(screen.getByTestId("route-auth_login")).toBeInTheDocument();
  });

  it("restores practice with planId / targetJobId / jdId / resumeVersionId / roundId after verify success", async () => {
    render(
      <App
        client={buildClient()}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      >
        <TriggerProbe />
      </App>,
    );

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "false",
      ),
    );

    const user = userEvent.setup();
    await user.click(screen.getByTestId("probe-start-practice"));

    await user.type(
      screen.getByTestId("auth-login-email"),
      "alice@example.com",
    );
    await user.click(screen.getByTestId("auth-login-submit-email"));

    await waitFor(() =>
      expect(screen.getByTestId("route-auth_verify")).toBeInTheDocument(),
    );

    await user.type(screen.getByTestId("auth-verify-code"), "654321");
    await user.click(screen.getByTestId("auth-verify-submit"));

    // PracticeScreen exposes route param echo as data-* attrs on its root.
    const practice = await screen.findByTestId("practice-screen");
    expect(practice.getAttribute("data-plan-id")).toBe(
      SAMPLE_ACTION.params.planId,
    );
    expect(practice.getAttribute("data-target-job-id")).toBe(
      SAMPLE_ACTION.params.targetJobId,
    );
    expect(practice.getAttribute("data-jd-id")).toBe(SAMPLE_ACTION.params.jdId);
    expect(practice.getAttribute("data-resume-version-id")).toBe(
      SAMPLE_ACTION.params.resumeVersionId,
    );
    expect(practice.getAttribute("data-round-id")).toBe(
      SAMPLE_ACTION.params.roundId,
    );
    expect(practice.getAttribute("data-session-id")).toBe(
      SAMPLE_ACTION.params.sessionId,
    );
  });

  it("navigates straight to the action route when the user is already signed in", async () => {
    render(
      <App
        client={buildClient()}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      >
        <TriggerProbe />
      </App>,
    );

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "true",
      ),
    );

    const user = userEvent.setup();
    await user.click(screen.getByTestId("probe-start-practice"));
    const practice = await screen.findByTestId("practice-screen");
    expect(practice.getAttribute("data-plan-id")).toBe("plan-tj-1");
    expect(practice.getAttribute("data-target-job-id")).toBe("tj-1");
  });
});
