// @vitest-environment jsdom
/**
 * Code-level auth pending-action resume regression.
 *
 * Truth source: docs/spec/frontend-shell/plans/001-app-shell-auth-settings/bdd-plan.md
 *               + bdd-checklist.md.
 *
 * Given a user without a session in a workspace plan context, clicking
 * `开始模拟面试` while signed-out must redirect to `auth_login`. After
 * `verifyAuthEmailChallenge` succeeds, App must restore the `practice` route
 * with all five interview-context params (planId / targetJobId / jdId /
 * resumeId / roundId) intact.
 */
import { describe, expect, it, vi } from "vitest";
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
  label: "开始模拟面试",
  route: "practice",
  params: {
    sessionId: "01918fa0-0000-7000-8000-000000005000",
    planId: "plan-tj-1",
    targetJobId: "tj-1",
    jdId: "jd-tj-1",
    resumeId: "frontend-v3",
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
      开始模拟面试
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

describe("auth pending-action resume", () => {
  it("redirects to auth_login when signed-out user clicks 开始模拟面试 and restores practice context after verify", async () => {
    window.history.replaceState(null, "", "/");
    const client = buildClient();
    const getPracticeSession = vi.spyOn(client, "getPracticeSession");
    render(
      <App
        client={client}
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

    await screen.findByTestId("practice-screen");
    const restoredParams = new URLSearchParams(window.location.search);
    for (const [key, value] of Object.entries(PRACTICE_PENDING_ACTION.params)) {
      expect(restoredParams.get(key), key).toBe(value);
    }
    await waitFor(() => {
      expect(getPracticeSession).toHaveBeenCalledWith(
        PRACTICE_PENDING_ACTION.params.sessionId,
      );
    });
  });
});
