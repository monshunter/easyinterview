// @vitest-environment jsdom
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
const LOADING_PREVIEW_DELAY = 3200;

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
          body: { ...fixtureBody(), analysisStatus },
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

function renderParse(client: EasyInterviewClient) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate }}>
            <ParseScreen
              route={{ name: "parse", params: { targetJobId: "tj-1" } }}
            />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
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

  it("keeps the ui-design loading demo before preview when analysisStatus is ready", async () => {
    vi.useFakeTimers();
    const client = createClientFromFixtures([makeFixture("ready")]);

    act(() => {
      renderParse(client);
    });

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });

    expect(screen.getByTestId("parse-loading-step-0")).toBeInTheDocument();
    expect(screen.queryByTestId("parse-basics-title")).not.toBeInTheDocument();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(600);
    });
    expect(screen.getByTestId("parse-loading-step-0")).toHaveTextContent(
      "Extracting title, level, location",
    );

    await act(async () => {
      await vi.advanceTimersByTimeAsync(LOADING_PREVIEW_DELAY - 600);
    });

    expect(screen.getByTestId("parse-basics-title")).toBeInTheDocument();
  });

  it("shows failed state when analysisStatus is failed", async () => {
    const client = createClientFromFixtures([makeFixture("failed")]);

    renderParse(client);

    await waitFor(() => {
      expect(screen.getByTestId("parse-failed-title")).toBeInTheDocument();
    });
  });

  it("polls getTargetJob when status is queued", async () => {
    vi.useFakeTimers();

    const queuedFixture = makeFixture("queued");
    const client = createClientFromFixtures([queuedFixture]);
    const spy = vi.spyOn(client, "getTargetJob");

    act(() => {
      renderParse(client);
    });

    // Let initial render + effect run
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });

    const initialCalls = spy.mock.calls.length;
    expect(initialCalls).toBeGreaterThanOrEqual(1);

    // Advance poll interval — should trigger another call
    await act(async () => {
      await vi.advanceTimersByTimeAsync(POLL_INTERVAL);
    });

    expect(spy.mock.calls.length).toBeGreaterThanOrEqual(initialCalls + 1);

    // Advance again
    await act(async () => {
      await vi.advanceTimersByTimeAsync(POLL_INTERVAL);
    });

    expect(spy.mock.calls.length).toBeGreaterThanOrEqual(initialCalls + 2);

    // Still in loading state since fixture returns queued
    expect(screen.getByTestId("parse-loading-step-0")).toBeInTheDocument();
  });

  it("transitions from queued to preview when status returns ready", async () => {
    vi.useFakeTimers();
    const client = createClientFromFixtures([makeFixture("ready")]);

    act(() => {
      renderParse(client);
    });

    await act(async () => {
      await vi.advanceTimersByTimeAsync(LOADING_PREVIEW_DELAY);
    });

    expect(screen.getByTestId("parse-basics-title")).toBeInTheDocument();
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
