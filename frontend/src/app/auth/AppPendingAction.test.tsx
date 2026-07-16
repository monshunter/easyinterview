// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it } from "vitest";
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

const TriggerProbe: FC = () => {
  const requestAuth = useRequestAuth();
  return (
    <button
      type="button"
      data-testid="probe-start-practice"
      onClick={() => requestAuth(SAMPLE_ACTION)}
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

beforeEach(() => window.history.replaceState(null, "", "/"));
afterEach(() => window.history.replaceState(null, "", "/"));

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

  it("restores practice with planId / targetJobId / jdId / resumeId / roundId after verify success", async () => {
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

    await screen.findByTestId("practice-screen");
    const restoredParams = new URLSearchParams(window.location.search);
    for (const [key, value] of Object.entries(SAMPLE_ACTION.params)) {
      expect(restoredParams.get(key), key).toBe(value);
    }
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
    await screen.findByTestId("practice-screen");
    const restoredParams = new URLSearchParams(window.location.search);
    expect(restoredParams.get("planId")).toBe("plan-tj-1");
    expect(restoredParams.get("targetJobId")).toBe("tj-1");
    expect(restoredParams.get("sessionId")).toBe(SAMPLE_ACTION.params.sessionId);
  });

  it("restores an unauthenticated Reports deep link with targetJobId only", async () => {
    window.history.replaceState(
      { rawText: "private-history" },
      "",
      "/reports?targetJobId=01918fa0-0000-7000-8000-000000002000&section=reports&reportId=rpt-hostile&status=ready&roundId=round-hostile&rawText=private-query",
    );
    const { unmount } = render(
      <App
        client={buildClient()}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );

    await waitFor(() => screen.getByTestId("route-auth_login"));
    expect(window.location.pathname).toBe("/auth/login");
    const pending = new URLSearchParams(window.location.search);
    expect(pending.get("pendingRoute")).toBe("reports");
    expect(pending.get("targetJobId")).toBe(
      "01918fa0-0000-7000-8000-000000002000",
    );
    for (const hostile of [
      "section",
      "reportId",
      "status",
      "roundId",
      "rawText",
    ]) {
      expect(pending.has(hostile), hostile).toBe(false);
    }
    expect(window.history.state).toBeNull();

    const user = userEvent.setup();
    await user.type(screen.getByTestId("auth-login-email"), "alice@example.com");
    await user.click(screen.getByTestId("auth-login-submit-email"));
    await waitFor(() => screen.getByTestId("route-auth_verify"));
    await user.type(screen.getByTestId("auth-verify-code"), "654321");
    await user.click(screen.getByTestId("auth-verify-submit"));

    await waitFor(() => screen.getByTestId("reports-screen"));
    expect(window.location.pathname + window.location.search).toBe(
      "/reports?targetJobId=01918fa0-0000-7000-8000-000000002000",
    );
    unmount();
  });
});
