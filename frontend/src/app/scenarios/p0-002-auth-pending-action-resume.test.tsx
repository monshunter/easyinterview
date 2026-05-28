// @vitest-environment jsdom
/**
 * E2E.P0.002 — Auth pending-action resume scenario.
 *
 * Truth source: docs/spec/frontend-shell/plans/001-app-shell-auth-settings/bdd-plan.md
 *               + bdd-checklist.md.
 *
 * Given a user without a session in a workspace plan context, clicking
 * `立即面试` while signed-out must redirect to `auth_login`. After
 * `verifyAuthEmailChallenge` succeeds, App must restore the `practice` route
 * with all five interview-context params (planId / targetJobId / jdId /
 * resumeVersionId / roundId) intact.
 */
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
import type { PendingAction } from "../auth";
import { useRequestAuth } from "../auth";

const PRACTICE_PENDING_ACTION: PendingAction = {
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

const WorkspaceTrigger: FC = () => {
  const requestAuth = useRequestAuth();
  return (
    <button
      type="button"
      data-testid="workspace-start-practice"
      onClick={() => requestAuth(PRACTICE_PENDING_ACTION)}
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

describe("E2E.P0.002 auth pending-action resume", () => {
  it("redirects to auth_login when signed-out user clicks 立即面试 and restores practice context after verify", async () => {
    render(
      <App
        client={buildClient()}
        initialRoute={{ name: "home", params: {} }}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      >
        <WorkspaceTrigger />
      </App>,
    );

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "false",
      ),
    );

    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-start-practice"));
    expect(screen.getByTestId("route-auth_login")).toBeInTheDocument();

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

    // After practice route restore, PracticeScreen mounts (sessionId in
    // pending action). PracticeScreen exposes route param echo via data-*
    // attrs on its root for E2E inspection.
    const practice = await screen.findByTestId("practice-screen");
    expect(practice.getAttribute("data-plan-id")).toBe(
      PRACTICE_PENDING_ACTION.params.planId,
    );
    expect(practice.getAttribute("data-target-job-id")).toBe(
      PRACTICE_PENDING_ACTION.params.targetJobId,
    );
    expect(practice.getAttribute("data-jd-id")).toBe(
      PRACTICE_PENDING_ACTION.params.jdId,
    );
    expect(practice.getAttribute("data-resume-version-id")).toBe(
      PRACTICE_PENDING_ACTION.params.resumeVersionId,
    );
    expect(practice.getAttribute("data-round-id")).toBe(
      PRACTICE_PENDING_ACTION.params.roundId,
    );
    expect(practice.getAttribute("data-session-id")).toBe(
      PRACTICE_PENDING_ACTION.params.sessionId,
    );
  });
});
