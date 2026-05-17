// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import type { FC, ReactNode } from "react";

import type { ResumeTailorRun } from "../../../../../api/generated/types";
import { AppRuntimeContext } from "../../../../runtime/AppRuntimeProvider";
import type { AppRuntimeValue } from "../../../../runtime/AppRuntimeProvider";
import { useResumeTailorRunPolling } from "./useResumeTailorRunPolling";

import tailorRunFixture from "../../../../../../../openapi/fixtures/ResumeTailor/getResumeTailorRun.json";

const RUN_ID = "01918fa0-0000-7000-8000-000000009000";

const fx = (key: keyof typeof tailorRunFixture.scenarios): ResumeTailorRun =>
  tailorRunFixture.scenarios[key].response.body as ResumeTailorRun;

const buildRuntime = (
  overrides: Partial<AppRuntimeValue["client"]>,
): AppRuntimeValue =>
  ({
    client: {
      getResumeTailorRun: vi.fn(),
      ...overrides,
    },
    runtime: { status: "ready", config: {} as never },
    auth: { status: "authenticated", user: { id: "u1" } as never },
    refreshAuth: vi.fn(),
  } as unknown as AppRuntimeValue);

const renderHookWithRuntime = <T,>(
  runtime: AppRuntimeValue,
  callback: () => T,
) => {
  const Wrapper: FC<{ children: ReactNode }> = ({ children }) => (
    <AppRuntimeContext.Provider value={runtime}>
      {children}
    </AppRuntimeContext.Provider>
  );
  return renderHook(callback, { wrapper: Wrapper });
};

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

const advance = async (ms: number) => {
  await act(async () => {
    await vi.advanceTimersByTimeAsync(ms);
  });
};

describe("useResumeTailorRunPolling", () => {
  it("returns idle when tailorRunId is null and does not invoke getResumeTailorRun", () => {
    const getResumeTailorRun = vi.fn();
    const runtime = buildRuntime({ getResumeTailorRun });
    const { result } = renderHookWithRuntime(runtime, () =>
      useResumeTailorRunPolling(null),
    );
    expect(result.current.phase).toBe("idle");
    expect(getResumeTailorRun).not.toHaveBeenCalled();
  });

  it("transitions polling -> ready on terminal ready status and fires onReady", async () => {
    const onReady = vi.fn();
    const ready = fx("default");
    const getResumeTailorRun = vi.fn().mockResolvedValueOnce(ready);
    const runtime = buildRuntime({ getResumeTailorRun });
    const { result } = renderHookWithRuntime(runtime, () =>
      useResumeTailorRunPolling(RUN_ID, {
        initialDelayMs: 1000,
        onReady,
      }),
    );
    expect(result.current.phase).toBe("polling");
    await advance(1100);
    expect(result.current.phase).toBe("ready");
    expect(result.current.run?.status).toBe("ready");
    expect(onReady).toHaveBeenCalledTimes(1);
  });

  it("transitions queued -> generating -> ready across polling cycles", async () => {
    const queued = fx("queued");
    const generating = fx("generating");
    const ready = fx("default");
    const getResumeTailorRun = vi
      .fn()
      .mockResolvedValueOnce(queued)
      .mockResolvedValueOnce(generating)
      .mockResolvedValueOnce(ready);
    const runtime = buildRuntime({ getResumeTailorRun });
    const { result } = renderHookWithRuntime(runtime, () =>
      useResumeTailorRunPolling(RUN_ID, {
        initialDelayMs: 50,
        // Constant delay keeps the test deterministic — backoff is verified by
        // the hook spec/audit, not by the timing assertions here.
        backoffFactor: 1.0,
        maxAttempts: 10,
      }),
    );
    // 1st cycle: initial delay 50ms -> queued
    await advance(60);
    expect(result.current.run?.status).toBe("queued");
    // 2nd cycle: +50ms -> generating
    await advance(60);
    expect(result.current.run?.status).toBe("generating");
    // 3rd cycle: +50ms -> ready
    await advance(60);
    expect(result.current.phase).toBe("ready");
    expect(getResumeTailorRun).toHaveBeenCalledTimes(3);
  });

  it("transitions to failed when status=failed", async () => {
    const failed = fx("failed");
    const getResumeTailorRun = vi.fn().mockResolvedValueOnce(failed);
    const onFailure = vi.fn();
    const runtime = buildRuntime({ getResumeTailorRun });
    const { result } = renderHookWithRuntime(runtime, () =>
      useResumeTailorRunPolling(RUN_ID, {
        initialDelayMs: 50,
        onFailure,
      }),
    );
    await advance(100);
    expect(result.current.phase).toBe("failed");
    expect(onFailure).toHaveBeenCalledTimes(1);
  });

  it("times out when status never becomes terminal within maxAttempts", async () => {
    const queued = fx("queued");
    const getResumeTailorRun = vi.fn().mockResolvedValue(queued);
    const onFailure = vi.fn();
    const runtime = buildRuntime({ getResumeTailorRun });
    const { result } = renderHookWithRuntime(runtime, () =>
      useResumeTailorRunPolling(RUN_ID, {
        initialDelayMs: 10,
        backoffFactor: 1.0, // constant delay for deterministic timing
        maxAttempts: 3,
        onFailure,
      }),
    );
    // Initial delay 10ms + 3 attempts at 10ms each = 40ms, then timeout fires
    // on the 4th tick when attempt >= maxAttempts.
    await advance(200);
    expect(getResumeTailorRun).toHaveBeenCalledTimes(3);
    expect(result.current.phase).toBe("timeout");
    expect(onFailure).toHaveBeenCalledTimes(1);
  });

  it("never passes an Idempotency-Key to getResumeTailorRun (D-12: read operations are not idempotent-keyed)", async () => {
    const getResumeTailorRun = vi.fn().mockResolvedValueOnce(fx("default"));
    const runtime = buildRuntime({ getResumeTailorRun });
    renderHookWithRuntime(runtime, () =>
      useResumeTailorRunPolling(RUN_ID, { initialDelayMs: 10 }),
    );
    await advance(50);
    expect(getResumeTailorRun).toHaveBeenCalledTimes(1);
    // The hook calls getResumeTailorRun(tailorRunId) with no opts argument.
    const args = getResumeTailorRun.mock.calls[0]!;
    expect(args).toHaveLength(1);
    expect(args[0]).toBe(RUN_ID);
  });

  it("clears the polling timer on unmount (no further calls)", async () => {
    const queued = fx("queued");
    const getResumeTailorRun = vi.fn().mockResolvedValue(queued);
    const runtime = buildRuntime({ getResumeTailorRun });
    const { unmount } = renderHookWithRuntime(runtime, () =>
      useResumeTailorRunPolling(RUN_ID, {
        initialDelayMs: 50,
        backoffFactor: 1.0,
      }),
    );
    // Fire only the first tick (50ms) - the second is scheduled at 100ms.
    await advance(60);
    expect(getResumeTailorRun).toHaveBeenCalledTimes(1);
    unmount();
    // Advance past the would-be second tick; the cleanup must cancel it.
    await advance(500);
    expect(getResumeTailorRun).toHaveBeenCalledTimes(1);
  });

  it("captures the canonical getResumeTailorRun fixture status variants", () => {
    expect(Object.keys(tailorRunFixture.scenarios).sort()).toEqual(
      ["default", "failed", "generating", "queued"].sort(),
    );
  });
});
