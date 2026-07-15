// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { StrictMode, type FC } from "react";

import getMeFixture from "../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import { EasyInterviewClient } from "../../api/generated/client";
import type { UserContext } from "../../api/generated/types";
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

const CommitProbe: FC<{ user: UserContext }> = ({ user }) => {
  const { auth, refreshAuth } = useAppRuntime();
  return (
    <div>
      <div data-testid="auth-status">{auth.status}</div>
      {auth.status === "authenticated" ? (
        <div data-testid="auth-user-name">{auth.user.displayName}</div>
      ) : null}
      <button
        data-testid="commit-auth-user"
        type="button"
        onClick={() => refreshAuth(user)}
      >
        commit
      </button>
    </div>
  );
};

describe("AppRuntimeProvider", () => {
	it("issues one underlying runtime/auth GET per key under StrictMode", async () => {
		const calls: string[] = [];
		const fixtureFetch = createFixtureBackedFetch(
			createFixtureRegistry([getRuntimeConfigFixture, getMeFixture]),
		);
		const client = new EasyInterviewClient({
			fetch: async (input, init) => {
				calls.push(String(input));
				return fixtureFetch(input, init);
			},
		});

		render(
			<StrictMode>
				<AppRuntimeProvider client={client}>
					<Probe />
				</AppRuntimeProvider>
			</StrictMode>,
		);

		await waitFor(() =>
			expect(screen.getByTestId("runtime-status")).toHaveTextContent("ready"),
		);
		expect(calls.filter((url) => url.endsWith("/runtime-config"))).toHaveLength(1);
		expect(calls.filter((url) => url.endsWith("/me"))).toHaveLength(1);
	});

	it("does not refetch when a parent recreates semantically identical option wrappers", async () => {
		const client = buildFixtureClient();
		const runtimeSpy = vi.spyOn(client, "getRuntimeConfig");
		const meSpy = vi.spyOn(client, "getMe");
		const view = () => (
			<AppRuntimeProvider
				client={client}
				requestOptions={{
					runtimeConfig: { headers: { Prefer: "example=default" } },
					getMe: { headers: { Prefer: "example=authenticated" } },
				}}
			>
				<Probe />
			</AppRuntimeProvider>
		);
		const { rerender } = render(view());

		await waitFor(() =>
			expect(screen.getByTestId("auth-status")).toHaveTextContent("authenticated"),
		);
		expect(runtimeSpy).toHaveBeenCalledTimes(1);
		expect(meSpy).toHaveBeenCalledTimes(1);

		rerender(view());
		await act(async () => {
			await Promise.resolve();
			await Promise.resolve();
		});
		expect(runtimeSpy).toHaveBeenCalledTimes(1);
		expect(meSpy).toHaveBeenCalledTimes(1);
	});

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

  it("does not reuse the public auth initial skip after directly committing a verified user", async () => {
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
      calls.push({ url, method: init?.method ?? "GET" });
      return fixtureFetch(input, init);
    };
    const client = new EasyInterviewClient({ fetch: spy });
    const committedUser: UserContext = {
      id: "01918fa0-0000-7000-8000-00000000feed",
      email: "verified@example.com",
      displayName: "Verified User",
      profileCompletionRequired: false,
    };
    const view = (lang: "zh" | "en") => (
      <AppRuntimeProvider
        client={client}
        skipInitialAuthProbe
        requestOptions={{
          getMe: { headers: { "Accept-Language": lang } },
        }}
      >
        <CommitProbe user={committedUser} />
      </AppRuntimeProvider>
    );

    const { rerender } = render(view("zh"));

    await waitFor(() =>
      expect(screen.getByTestId("auth-status")).toHaveTextContent(
        "unauthenticated",
      ),
    );
    expect(calls.filter((c) => c.url.endsWith("/api/v1/me"))).toHaveLength(0);

    fireEvent.click(screen.getByTestId("commit-auth-user"));
    await waitFor(() =>
      expect(screen.getByTestId("auth-user-name")).toHaveTextContent(
        "Verified User",
      ),
    );

    const meCallsBeforeLanguageSwitch = calls.filter((c) =>
      c.url.endsWith("/api/v1/me"),
    ).length;
    rerender(view("en"));

    await waitFor(() =>
      expect(
        calls.filter((c) => c.url.endsWith("/api/v1/me")).length,
      ).toBeGreaterThan(meCallsBeforeLanguageSwitch),
    );
    await waitFor(() =>
      expect(screen.getByTestId("auth-user-name")).toHaveTextContent(
        "Alice Example",
      ),
    );
  });
});
