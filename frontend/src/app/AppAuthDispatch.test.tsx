// @vitest-environment jsdom
import { afterEach, describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import completeMyProfileFixture from "../../../openapi/fixtures/Auth/completeMyProfile.json";
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
      completeMyProfileFixture,
      logoutFixture,
      startAuthEmailChallengeFixture,
      verifyAuthEmailChallengeFixture,
    ]),
  );
  return new EasyInterviewClient({ fetch: spy ?? fixtureFetch });
}

function buildRecordingRuntimeClient(options?: { hangGetMe?: boolean }) {
  const calls: Array<{ url: string; method: string }> = [];
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getRuntimeConfigFixture,
      getMeFixture,
      completeMyProfileFixture,
      logoutFixture,
      startAuthEmailChallengeFixture,
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
    if (options?.hangGetMe && url.endsWith("/api/v1/me")) {
      return new Promise<Response>(() => {});
    }
    return fixtureFetch(input, init);
  };
  return { client: new EasyInterviewClient({ fetch: spy }), calls };
}

describe("App auth route dispatch", () => {
  afterEach(() => {
    window.history.replaceState(null, "", "/");
  });

  it("renders the matching auth screen with screen-specific testids for each auth_* route name", () => {
    const cases: Array<{
      name:
        | "auth_login"
        | "auth_verify"
        | "auth_profile_setup"
        | "auth_logout";
      hallmarkTestId: string;
    }> = [
      { name: "auth_login", hallmarkTestId: "auth-login-email" },
      { name: "auth_verify", hallmarkTestId: "auth-verify-code" },
      { name: "auth_profile_setup", hallmarkTestId: "auth-profile-name" },
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

  it("renders the login screen when the out-of-scope auth_reset route is requested (D-16)", () => {
    const { unmount } = render(
      <App
        client={buildClient()}
        initialRoute={{ name: "auth_reset", params: {} }}
      />,
    );
    expect(screen.getByTestId("route-auth_login")).toBeInTheDocument();
    expect(screen.queryByTestId("route-auth_reset")).not.toBeInTheDocument();
    unmount();
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
    expect(startCall?.body).toEqual({
      email: "alice@example.com",
    });
    await screen.findByTestId("route-auth_verify");
    expect(screen.getByTestId("auth-verify-email-hint")).toHaveTextContent(
      "alice@example.com",
    );
  });

  it("does not issue the initial /me probe on public auth entry routes", async () => {
    const cases: Array<"auth_login" | "auth_verify"> = [
      "auth_login",
      "auth_verify",
    ];
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

  it("normalizes out-of-scope auth_register route names to the single login screen", () => {
    const { unmount } = render(
      <App
        client={buildClient()}
        initialRoute={{ name: "auth_register", params: {} }}
      />,
    );
    expect(screen.getByTestId("route-auth_login")).toBeInTheDocument();
    expect(screen.getByTestId("auth-login-email")).toBeInTheDocument();
    expect(screen.queryByTestId("route-auth_register")).not.toBeInTheDocument();
    unmount();
  });

  it("redirects unauthenticated auth_profile_setup visits back to the login entry", async () => {
    render(
      <App
        client={buildClient()}
        initialRoute={{
          name: "auth_profile_setup",
          params: {
            pendingRoute: "practice",
            pendingType: "start_practice",
            pendingLabel: "立即面试",
            planId: "plan-1",
          },
        }}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );

    await screen.findByTestId("route-auth_login");
    expect(screen.getByTestId("auth-login-email")).toBeInTheDocument();
    expect(
      screen.queryByTestId("route-auth_profile_setup"),
    ).not.toBeInTheDocument();
  });

  it("holds protected interview routes behind auth loading instead of mounting business screens", async () => {
    const { client, calls } = buildRecordingRuntimeClient({ hangGetMe: true });

    render(
      <App
        client={client}
        initialRoute={{
          name: "workspace",
          params: { targetJobId: "01918fa0-0000-7000-8000-000000002000" },
        }}
      />,
    );

    expect(await screen.findByTestId("auth-route-gate")).toBeInTheDocument();
    expect(screen.queryByTestId("workspace-crumbs")).not.toBeInTheDocument();
    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(calls.some((c) => c.url.includes("/api/v1/targets"))).toBe(false);
  });

  it("redirects signed-out protected interview routes to auth_login with pendingAction", async () => {
    const cases: Array<{
      name:
        | "parse"
        | "workspace"
        | "resume_versions"
        | "practice"
        | "generating"
        | "report"
        | "settings";
      params: Record<string, string>;
      businessTestId: string;
    }> = [
      { name: "parse", params: { targetJobId: "tj-1" }, businessTestId: "parse-loading-step-0" },
      { name: "workspace", params: { targetJobId: "tj-1" }, businessTestId: "workspace-crumbs" },
      { name: "resume_versions", params: { flow: "create" }, businessTestId: "resume-workshop-screen" },
      { name: "practice", params: { sessionId: "session-1" }, businessTestId: "practice-screen" },
      { name: "generating", params: { sessionId: "session-1", reportId: "report-1" }, businessTestId: "generating-screen" },
      { name: "report", params: { sessionId: "session-1", reportId: "report-1" }, businessTestId: "report-dashboard" },
      { name: "settings", params: {}, businessTestId: "route-settings" },
    ];

    for (const tc of cases) {
      const { client, calls } = buildRecordingRuntimeClient();
      const { unmount } = render(
        <App
          client={client}
          initialRoute={{ name: tc.name, params: tc.params }}
          requestOptions={{
            getMe: { headers: { Prefer: "example=unauthenticated" } },
          }}
        />,
      );

      await screen.findByTestId("route-auth_login");
      expect(screen.getByTestId("auth-side-pending-action")).toBeInTheDocument();
      expect(screen.queryByTestId(tc.businessTestId)).not.toBeInTheDocument();
      expect(
        calls.some(
          (c) =>
            !c.url.endsWith("/api/v1/runtime-config") &&
            !c.url.endsWith("/api/v1/me"),
        ),
      ).toBe(false);
      unmount();
      window.history.replaceState(null, "", "/");
    }
  });

  it("rewrites a direct protected browser URL to the login URL when signed out", async () => {
    window.history.replaceState(
      null,
      "",
      "/workspace?targetJobId=tj-1",
    );
    const { client, calls } = buildRecordingRuntimeClient();

    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      />,
    );

    await screen.findByTestId("route-auth_login");
    expect(screen.getByTestId("auth-side-pending-action")).toBeInTheDocument();
    expect(window.location.href).toContain("/auth/login");
    expect(window.location.href).toContain("pendingRoute=workspace");
    expect(window.location.href).toContain("targetJobId=tj-1");
    expect(calls.some((c) => c.url.includes("/api/v1/targets"))).toBe(false);
  });

  it("wires auth_profile_setup through completeMyProfile and restores the pending route", async () => {
    const calls: Array<{ url: string; method: string; body: unknown }> = [];
    const fixtureFetch = createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        completeMyProfileFixture,
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
        initialRoute={{
          name: "auth_profile_setup",
          params: {
            pendingRoute: "practice",
            pendingType: "start_practice",
            pendingLabel: "立即面试",
            planId: "plan-1",
          },
        }}
        requestOptions={{
          getMe: { headers: { Prefer: "example=profileIncomplete" } },
        }}
      />,
    );
    const user = userEvent.setup();
    await screen.findByTestId("route-auth_profile_setup");
    await user.type(screen.getByTestId("auth-profile-name"), "Alice");
    await user.click(screen.getByTestId("auth-profile-terms"));
    await user.click(screen.getByTestId("auth-profile-submit"));

    await waitFor(() =>
      expect(
        calls.some(
          (c) =>
            c.method === "PATCH" &&
            c.url.endsWith("/api/v1/me"),
        ),
      ).toBe(true),
    );
    const patchCall = calls.find((c) =>
      c.method === "PATCH" && c.url.endsWith("/api/v1/me"),
    );
    expect(patchCall?.body).toEqual({
      displayName: "Alice",
      acceptedTerms: true,
    });
    await screen.findByTestId("practice-session-lost");
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

  it("does not strand the user on auth_verify when post-verify /me refresh fails", async () => {
    const calls: Array<{ url: string; method: string }> = [];
    const fixtureFetch = createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
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
      if (url.endsWith("/api/v1/me")) {
        return new Response(
          JSON.stringify({
            error: {
              code: "AUTH_PROFILE_REFRESH_UNAVAILABLE",
              message: "Profile context refresh timed out.",
              requestId: "req-profile-refresh-failed",
              retryable: true,
            },
          }),
          {
            status: 503,
            headers: { "Content-Type": "application/json" },
          },
        );
      }
      return fixtureFetch(input, init);
    };
    const client = new EasyInterviewClient({ fetch: spy });
    render(
      <App
        client={client}
        initialRoute={{
          name: "auth_verify",
          params: {
            email: "alice@example.com",
            pendingRoute: "workspace",
            pendingType: "open_protected_route",
            pendingLabel: "workspace",
            targetJobId: "tj-1",
          },
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
    await screen.findByTestId("auth-route-gate");
    expect(screen.getByTestId("auth-route-gate")).toHaveAttribute(
      "data-route-name",
      "workspace",
    );
    expect(screen.queryByTestId("route-auth_verify")).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("auth-verify-code-status"),
    ).not.toBeInTheDocument();
    expect(
      calls.filter((c) => c.url.endsWith("/api/v1/me")).length,
    ).toBeGreaterThan(0);
  });

  it("falls back to RouteShellScreen for auth_* routes when no client / runtime is mounted", () => {
    render(<App initialRoute={{ name: "auth_login", params: {} }} />);
    // RouteShellScreen renders the bare data attributes; it does not render
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

    await screen.findByTestId("route-auth_profile_setup");
    await user.type(screen.getByTestId("auth-profile-name"), "Alice Example");
    await user.click(screen.getByTestId("auth-profile-terms"));
    await user.click(screen.getByTestId("auth-profile-submit"));

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
    expect(screen.queryByTestId("topbar-register")).not.toBeInTheDocument();
  });
});
