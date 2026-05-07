// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import type { FC } from "react";

import getMeFixture from "../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import { EasyInterviewClient } from "../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../api/mockTransport";

import { AppRuntimeProvider, useAppRuntime } from "./AppRuntimeProvider";

function buildFixtureClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getRuntimeConfigFixture, getMeFixture]),
    ),
  });
}

const Probe: FC = () => {
  const { runtime, auth } = useAppRuntime();
  return (
    <div>
      <div data-testid="runtime-status">{runtime.status}</div>
      {runtime.status === "ready" ? (
        <div data-testid="runtime-app-version">{runtime.config.appVersion}</div>
      ) : null}
      {runtime.status === "error" ? (
        <div data-testid="runtime-error">{runtime.error.message}</div>
      ) : null}
      <div data-testid="auth-status">{auth.status}</div>
      {auth.status === "authenticated" ? (
        <div data-testid="auth-user-name">{auth.user.displayName}</div>
      ) : null}
      {auth.status === "error" ? (
        <div data-testid="auth-error">{auth.error.message}</div>
      ) : null}
    </div>
  );
};

describe("AppRuntimeProvider", () => {
  it("loads runtime config via the generated client + fixture-backed transport", async () => {
    const client = buildFixtureClient();
    render(
      <AppRuntimeProvider client={client}>
        <Probe />
      </AppRuntimeProvider>,
    );
    await waitFor(() =>
      expect(screen.getByTestId("runtime-status")).toHaveTextContent("ready"),
    );
    expect(screen.getByTestId("runtime-app-version")).toHaveTextContent(
      "1.0.0+dev.0428",
    );
  });

  it("treats /me authenticated scenario as signed-in and exposes user context", async () => {
    const client = buildFixtureClient();
    render(
      <AppRuntimeProvider
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      >
        <Probe />
      </AppRuntimeProvider>,
    );
    await waitFor(() =>
      expect(screen.getByTestId("auth-status")).toHaveTextContent(
        "authenticated",
      ),
    );
    expect(screen.getByTestId("auth-user-name")).toHaveTextContent(
      "Alice Example",
    );
  });

  it("treats /me unauthenticated 401 as signed-out without blocking runtime", async () => {
    const client = buildFixtureClient();
    render(
      <AppRuntimeProvider
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      >
        <Probe />
      </AppRuntimeProvider>,
    );
    await waitFor(() =>
      expect(screen.getByTestId("auth-status")).toHaveTextContent(
        "unauthenticated",
      ),
    );
    expect(screen.queryByTestId("auth-user-name")).not.toBeInTheDocument();
    // Runtime resolves independently of auth state.
    await waitFor(() =>
      expect(screen.getByTestId("runtime-status")).toHaveTextContent("ready"),
    );
  });

  it("fails loudly when an unknown fixture scenario is requested", async () => {
    const client = buildFixtureClient();
    render(
      <AppRuntimeProvider
        client={client}
        requestOptions={{
          runtimeConfig: { headers: { Prefer: "example=does-not-exist" } },
        }}
      >
        <Probe />
      </AppRuntimeProvider>,
    );
    await waitFor(() =>
      expect(screen.getByTestId("runtime-status")).toHaveTextContent("error"),
    );
    expect(screen.getByTestId("runtime-error").textContent).toMatch(
      /unknown fixture scenario/i,
    );
  });

  it("fails loudly when /me uses an unknown fixture scenario", async () => {
    const client = buildFixtureClient();
    render(
      <AppRuntimeProvider
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=does-not-exist" } },
        }}
      >
        <Probe />
      </AppRuntimeProvider>,
    );
    await waitFor(() =>
      expect(screen.getByTestId("auth-status")).toHaveTextContent("error"),
    );
    expect(screen.getByTestId("auth-error").textContent).toMatch(
      /unknown fixture scenario/i,
    );
  });
});
