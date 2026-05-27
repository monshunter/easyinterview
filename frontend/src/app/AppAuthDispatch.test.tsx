// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import logoutFixture from "../../../openapi/fixtures/Auth/logout.json";
import startAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/startAuthEmailChallenge.json";
import verifyAuthEmailChallengeFixture from "../../../openapi/fixtures/Auth/verifyAuthEmailChallenge.json";
import { EasyInterviewClient } from "../api/generated/client";
import { createDevMockClient } from "../api/devMockClient";
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
  afterEach(() => {
    window.history.replaceState(null, "", "/");
  });

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

  it("wires auth_login submit through startAuthEmailChallenge and treats empty 202 as success", async () => {
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
      if (url.endsWith("/api/v1/auth/email/start")) {
        return new Response(null, { status: 202, statusText: "Accepted" });
      }
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
    await screen.findByTestId("route-auth_verify");
    expect(screen.getByTestId("auth-verify-email-hint")).toHaveTextContent(
      "alice@example.com",
    );
  });

  it("does not issue the initial /me probe on public auth entry routes", async () => {
    const cases: Array<
      "auth_login" | "auth_register" | "auth_verify" | "auth_reset"
    > = ["auth_login", "auth_register", "auth_verify", "auth_reset"];
    for (const name of cases) {
      const calls: Array<{ url: string; method: string }> = [];
      const fixtureFetch = createFixtureBackedFetch(
        createFixtureRegistry([getRuntimeConfigFixture, getMeFixture]),
      );
      const spy: typeof fetch = async (input, init) => {
        const url =
          typeof input === "string"
            ? input
            : input instanceof URL
              ? input.href
              : input.url;
        const method = init?.method ?? "GET";
        calls.push({ url, method });
        return fixtureFetch(input, init);
      };
      const client = new EasyInterviewClient({ fetch: spy });

      const { unmount } = render(
        <App client={client} initialRoute={{ name, params: {} }} />,
      );

      await waitFor(() =>
        expect(
          calls.some((c) => c.url.endsWith("/api/v1/runtime-config")),
        ).toBe(true),
      );
      expect(calls.filter((c) => c.url.endsWith("/api/v1/me"))).toHaveLength(
        0,
      );
      unmount();
    }
  });

  it("wires auth_register submit through startAuthEmailChallenge and treats empty 202 as success", async () => {
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
      if (url.endsWith("/api/v1/auth/email/start")) {
        return new Response(null, { status: 202, statusText: "Accepted" });
      }
      return fixtureFetch(input, init);
    };
    const client = new EasyInterviewClient({ fetch: spy });
    render(
      <App
        client={client}
        initialRoute={{ name: "auth_register", params: {} }}
      />,
    );
    const user = userEvent.setup();
    await user.type(screen.getByTestId("auth-register-name"), "Alice");
    await user.type(
      screen.getByTestId("auth-register-email"),
      "alice@example.com",
    );
    await user.click(screen.getByTestId("auth-register-terms"));
    await user.click(screen.getByTestId("auth-register-submit"));

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
    await screen.findByTestId("route-auth_verify");
    expect(screen.getByTestId("auth-verify-email-hint")).toHaveTextContent(
      "alice@example.com",
    );
  });

  it("wires auth_verify through generated verifyAuthEmailChallenge with the required token query", async () => {
    const calls: Array<{ url: string; method: string }> = [];
    const fixtureFetch = createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        verifyAuthEmailChallengeFixture,
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
      calls.push({ url, method });
      return fixtureFetch(input, init);
    };
    const client = new EasyInterviewClient({ fetch: spy });
    render(
      <App
        client={client}
        initialRoute={{
          name: "auth_verify",
          params: { email: "alice@example.com" },
        }}
      />,
    );

    const user = userEvent.setup();
    await user.type(screen.getByTestId("auth-verify-code"), "654321");
    await user.click(screen.getByTestId("auth-verify-submit"));

    await waitFor(() =>
      expect(
        calls.some(
          (c) =>
            c.method === "GET" &&
            c.url.endsWith("/api/v1/auth/email/verify?token=654321"),
        ),
      ).toBe(true),
    );
  });

  it("auto-consumes auth_verify magic-link token and replaces it out of the URL", async () => {
    window.history.replaceState(
      null,
      "",
      "/auth/verify?token=magic-link-token&pendingRoute=workspace&pendingType=start_practice&pendingLabel=%E7%AB%8B%E5%8D%B3%E9%9D%A2%E8%AF%95&targetJobId=tj-1",
    );
    const calls: Array<{ url: string; method: string }> = [];
    const fixtureFetch = createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        verifyAuthEmailChallengeFixture,
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
      calls.push({ url, method });
      return fixtureFetch(input, init);
    };
    const client = new EasyInterviewClient({ fetch: spy });

    render(<App client={client} />);

    await waitFor(() =>
      expect(
        calls.some(
          (c) =>
            c.method === "GET" &&
            c.url.endsWith("/api/v1/auth/email/verify?token=magic-link-token"),
        ),
      ).toBe(true),
    );
    await waitFor(() => expect(window.location.pathname).toBe("/workspace"));
    await waitFor(() =>
      expect(calls.filter((c) => c.url.endsWith("/api/v1/me"))).toHaveLength(
        1,
      ),
    );
    expect(window.location.search).not.toContain("token=");
    expect(window.location.search).toContain("targetJobId=tj-1");
  });

  it("falls back to PlaceholderScreen for auth_* routes when no client / runtime is mounted", () => {
    render(<App initialRoute={{ name: "auth_login", params: {} }} />);
    // PlaceholderScreen renders the bare data attributes; it does not render
    // the email form testid.
    expect(screen.getByTestId("route-auth_login")).toBeInTheDocument();
    expect(screen.queryByTestId("auth-login-email")).not.toBeInTheDocument();
  });

  it("uses the dev mock session state for sign in and logout in the mounted App", async () => {
    render(<App client={createDevMockClient()} />);
    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "false",
      ),
    );

    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-login"));
    await user.type(screen.getByTestId("auth-login-email"), "alice@example.com");
    await user.click(screen.getByTestId("auth-login-submit-email"));
    await screen.findByTestId("route-auth_verify");
    await user.type(screen.getByTestId("auth-verify-code"), "654321");
    await user.click(screen.getByTestId("auth-verify-submit"));

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "true",
      ),
    );
    await user.click(screen.getByTestId("topbar-user-chip"));
    await user.click(screen.getByTestId("topbar-user-logout"));
    await screen.findByTestId("route-auth_logout");
    await user.click(screen.getByTestId("auth-logout-confirm"));

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "false",
      ),
    );
    expect(screen.getByTestId("topbar-login")).toBeInTheDocument();
    expect(screen.getByTestId("topbar-register")).toBeInTheDocument();
  });
});
