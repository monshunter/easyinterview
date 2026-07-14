// @vitest-environment jsdom
import { StrictMode } from "react";
import { describe, expect, it, vi, afterEach } from "vitest";
import { render, screen, waitFor, act } from "@testing-library/react";

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

import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";

const POLL_INTERVAL = 610;

function fixtureBody() {
  return (
    getTargetJobFixture.scenarios["default"] as {
      response: { body: Record<string, unknown> };
    }
  ).response.body;
}

function makeFixture(
  analysisStatus: "queued" | "processing" | "ready" | "failed",
): OperationFixture {
  return {
    operationId: "getTargetJob",
    scenarios: {
      default: {
        response: {
          status: 200,
          body: { ...fixtureBody(), id: "tj-1", analysisStatus },
        },
      },
    },
  };
}

function createClientFromFixtures(fixtures: OperationFixture[]) {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry(fixtures),
    { scenario: "default" },
  );
  return new EasyInterviewClient({ fetch });
}

function createCountingClientFromFixtures(fixtures: OperationFixture[]) {
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry(fixtures),
    { scenario: "default" },
  );
  const fetch = vi.fn<typeof globalThis.fetch>((input, init) =>
    fixtureFetch(input, init),
  );
  return { client: new EasyInterviewClient({ fetch }), fetch };
}

function targetJobTransportCount(fetch: ReturnType<typeof vi.fn>): number {
  return fetch.mock.calls.filter(([input, init]) => {
    const method = (init?.method ?? "GET").toUpperCase();
    const path = new URL(String(input), "http://fixture.local").pathname;
    return method === "GET" && path === "/api/v1/targets/tj-1";
  }).length;
}

function createClientForTargetSwitch() {
  const base = fixtureBody();
  const fetch = vi.fn(async (input: RequestInfo | URL) => {
    const url = new URL(String(input), "http://fixture.local");
    if (url.pathname === "/api/v1/runtime-config") {
      return new Response(JSON.stringify({}), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }
    if (url.pathname === "/api/v1/me") {
      return new Response(JSON.stringify({ id: "user-1", email: "user@example.com" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }
    const targetId = url.pathname.replace("/api/v1/targets/", "");
    const title =
      targetId === "target-b"
        ? "Backend Platform Engineer"
        : "Senior Frontend Engineer";
    return new Response(
      JSON.stringify({
        ...base,
        id: targetId,
        title,
        companyName: targetId === "target-b" ? "Boreal Systems" : "Acme",
        locationText: targetId === "target-b" ? "Remote" : "Shanghai",
        analysisStatus: "ready",
      }),
      {
        status: 200,
        headers: { "Content-Type": "application/json" },
      },
    );
  });
  return new EasyInterviewClient({ fetch });
}

function renderParse(client: EasyInterviewClient, options?: { strict?: boolean }) {
  const navigate = vi.fn();
  const replaceRoute = vi.fn();
  const parse = (
    <DisplayPreferencesProvider>
      <AppRuntimeProvider client={client}>
        <NavigationProvider value={{ navigate, replaceRoute }}>
          <ParseScreen
            route={{ name: "parse", params: { targetJobId: "tj-1" } }}
          />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>
  );
  return {
    navigate,
    replaceRoute,
    ...render(options?.strict ? <StrictMode>{parse}</StrictMode> : parse),
  };
}

function renderWorkspaceDetail(client: EasyInterviewClient, targetJobId: string) {
  return (
    <DisplayPreferencesProvider>
      <AppRuntimeProvider client={client}>
        <NavigationProvider value={{ navigate: vi.fn() }}>
          <ParseScreen route={{ name: "workspace", params: { targetJobId } }} />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>
  );
}

describe("ParseFlow — analysisStatus polling", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("calls getTargetJob on mount with correct targetJobId", async () => {
    const client = createClientFromFixtures([makeFixture("ready")]);
    const spy = vi.spyOn(client, "getTargetJob");

    renderParse(client);

    await waitFor(() => {
      expect(spy).toHaveBeenCalledWith("tj-1");
    });
  });

  it("immediately replaces ready Parse state with workspace detail", async () => {
    const { client, fetch } = createCountingClientFromFixtures([
      makeFixture("ready"),
    ]);
    const { replaceRoute } = renderParse(client, { strict: true });

    await waitFor(() => {
      expect(replaceRoute).toHaveBeenCalledWith({
        name: "workspace",
        params: { targetJobId: "tj-1" },
      });
    });
    expect(targetJobTransportCount(fetch)).toBe(1);
    expect(screen.queryByTestId("parse-basics-title")).not.toBeInTheDocument();
  });

  it("shows failed state when analysisStatus is failed", async () => {
    const client = createClientFromFixtures([makeFixture("failed")]);

    renderParse(client);

    await waitFor(() => {
      expect(screen.getByTestId("parse-failed-title")).toBeInTheDocument();
    });
  });

  it("issues one queued transport on StrictMode mount and one per scheduler tick", async () => {
    vi.useFakeTimers();

    const queuedFixture = makeFixture("queued");
    const { client, fetch } = createCountingClientFromFixtures([queuedFixture]);

    act(() => {
      renderParse(client, { strict: true });
    });

    // Let initial render + effect run
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });

    expect(targetJobTransportCount(fetch)).toBe(1);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(POLL_INTERVAL - 11);
    });
    expect(targetJobTransportCount(fetch)).toBe(1);

    // Advance poll interval — should trigger another call
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1);
    });

    expect(targetJobTransportCount(fetch)).toBe(2);

    // Advance again
    await act(async () => {
      await vi.advanceTimersByTimeAsync(POLL_INTERVAL);
    });

    expect(targetJobTransportCount(fetch)).toBe(3);

    // Still in loading state since fixture returns queued
    expect(screen.getByTestId("parse-loading-step-0")).toBeInTheDocument();
    console.info(
      "E2E.P0.015 Parse StrictMode transport PASS initial=1 tick1=2 tick2=3",
    );
  });

  it("polls only after the scheduler interval and replaces when status becomes ready", async () => {
    vi.useFakeTimers();
    const client = createClientFromFixtures([makeFixture("queued")]);
    const ready = {
      ...fixtureBody(),
      id: "tj-1",
      analysisStatus: "ready" as const,
    } as Awaited<ReturnType<EasyInterviewClient["getTargetJob"]>>;
    const queued = { ...ready, analysisStatus: "queued" as const };
    const getTargetJob = vi
      .spyOn(client, "getTargetJob")
      .mockResolvedValueOnce(queued)
      .mockResolvedValueOnce(ready);
    let replaceRoute: ReturnType<typeof vi.fn>;

    act(() => {
      ({ replaceRoute } = renderParse(client));
    });

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });
    expect(getTargetJob).toHaveBeenCalledTimes(1);
    expect(replaceRoute!).not.toHaveBeenCalled();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(POLL_INTERVAL);
    });
    expect(getTargetJob).toHaveBeenCalledTimes(2);
    expect(replaceRoute!).toHaveBeenCalledWith({
      name: "workspace",
      params: { targetJobId: "tj-1" },
    });
  });

  it("loads direct workspace detail without Parse progress and reloads on target switch", async () => {
    const client = createClientForTargetSwitch();
    const getTargetJob = vi.spyOn(client, "getTargetJob");
    const { rerender } = render(renderWorkspaceDetail(client, "target-a"));

    await waitFor(() => {
      expect(screen.getByTestId("parse-basics-title")).toHaveTextContent(
        "Senior Frontend Engineer",
      );
    });
    expect(getTargetJob).toHaveBeenCalledTimes(1);
    expect(screen.queryByTestId("parse-loading-step-0")).not.toBeInTheDocument();

    rerender(renderWorkspaceDetail(client, "target-b"));

    expect(screen.getByTestId("workspace-detail-loading")).toBeInTheDocument();
    expect(screen.queryByText("Senior Frontend Engineer")).not.toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByTestId("parse-basics-title")).toHaveTextContent(
        "Backend Platform Engineer",
      );
    });
    expect(getTargetJob).toHaveBeenCalledTimes(2);
  });

  it("cleans up polling on unmount", async () => {
    vi.useFakeTimers();

    const client = createClientFromFixtures([makeFixture("queued")]);
    const spy = vi.spyOn(client, "getTargetJob");

    const { unmount } = render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate: vi.fn() }}>
            <ParseScreen
              route={{ name: "parse", params: { targetJobId: "tj-1" } }}
            />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    );

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });

    const callsAfterMount = spy.mock.calls.length;
    expect(callsAfterMount).toBeGreaterThanOrEqual(1);

    unmount();

    // Advance well past several poll intervals
    await act(async () => {
      await vi.advanceTimersByTimeAsync(POLL_INTERVAL * 5);
    });

    expect(spy.mock.calls.length).toBe(callsAfterMount);
  });
});
