// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import logoutFixture from "../../../openapi/fixtures/Auth/logout.json";
import startAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/startAuthEmailChallenge.json";
import verifyAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/verifyAuthEmailChallenge.json";
import { EasyInterviewClient } from "../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../api/mockTransport";
import { App } from "./App";

function buildClient(spy?: typeof fetch) {
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getRuntimeConfigFixture,
      getMeFixture,
      logoutFixture,
      startAuthEmailChallengeFixture,
      verifyAuthEmailChallengeFixture,
    ]),
  );
  return new EasyInterviewClient({ fetch: spy ?? fixtureFetch });
}

describe("App auth route dispatch", () => {
  it("renders the matching auth screen with screen-specific testids for each auth_* route name", () => {
    const cases: Array<{
      name:
        | "auth_login"
        | "auth_register"
        | "auth_verify"
        | "auth_reset"
        | "auth_logout";
      hallmarkTestId: string;
    }> = [
      { name: "auth_login", hallmarkTestId: "auth-login-email" },
      { name: "auth_register", hallmarkTestId: "auth-register-email" },
      { name: "auth_verify", hallmarkTestId: "auth-verify-code" },
      { name: "auth_reset", hallmarkTestId: "auth-reset-email" },
      { name: "auth_logout", hallmarkTestId: "auth-logout-confirm" },
    ];
    for (const { name, hallmarkTestId } of cases) {
      const { unmount } = render(
        <App client={buildClient()} initialRoute={{ name, params: {} }} />,
      );
      expect(screen.getByTestId(`route-${name}`)).toBeInTheDocument();
      expect(screen.getByTestId(hallmarkTestId)).toBeInTheDocument();
      unmount();
    }
  });

  it("wires the auth_login submit through the generated startAuthEmailChallenge operation", async () => {
    const calls: Array<{ url: string; method: string; body: unknown }> = [];
    const fixtureFetch = createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        startAuthEmailChallengeFixture,
      ]),
    );
    const spy: typeof fetch = async (input, init) => {
      const url =
        typeof input === "string"
          ? input
          : input instanceof URL
            ? input.href
            : input.url;
      const method = init?.method ?? "GET";
      let body: unknown = undefined;
      if (typeof init?.body === "string") {
        try {
          body = JSON.parse(init.body);
        } catch {
          body = init.body;
        }
      }
      calls.push({ url, method, body });
      return fixtureFetch(input, init);
    };
    const client = new EasyInterviewClient({ fetch: spy });
    render(
      <App
        client={client}
        initialRoute={{ name: "auth_login", params: {} }}
      />,
    );
    const user = userEvent.setup();
    await user.type(
      screen.getByTestId("auth-login-email"),
      "alice@example.com",
    );
    await user.click(screen.getByTestId("auth-login-submit-email"));
    await waitFor(() =>
      expect(
        calls.some(
          (c) =>
            c.method === "POST" &&
            c.url.endsWith("/api/v1/auth/email/start"),
        ),
      ).toBe(true),
    );
    const startCall = calls.find((c) =>
      c.url.endsWith("/api/v1/auth/email/start"),
    );
    expect(startCall?.body).toEqual({ email: "alice@example.com" });
  });

  it("falls back to PlaceholderScreen for auth_* routes when no client / runtime is mounted", () => {
    render(<App initialRoute={{ name: "auth_login", params: {} }} />);
    // PlaceholderScreen renders the bare data attributes; it does not render
    // the email form testid.
    expect(screen.getByTestId("route-auth_login")).toBeInTheDocument();
    expect(screen.queryByTestId("auth-login-email")).not.toBeInTheDocument();
  });
});
