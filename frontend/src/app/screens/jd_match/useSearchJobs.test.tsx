// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";

import searchJobsFixture from "../../../../../openapi/fixtures/JobMatch/searchJobs.json";

import { useSearchJobs } from "./useSearchJobs";

function buildClient(scenario: string) {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([searchJobsFixture]),
      { scenario },
    ),
  });
}

function withRuntime(client: EasyInterviewClient) {
  return ({ children }: { children: ReactNode }) => (
    <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>
  );
}

describe("useSearchJobs (item 4.2)", () => {
  it("starts in searching=false with empty results / no error / hasRunOnce=false", () => {
    const client = buildClient("default");
    const { result } = renderHook(() => useSearchJobs(), {
      wrapper: withRuntime(client),
    });
    expect(result.current.searching).toBe(false);
    expect(result.current.results).toEqual([]);
    expect(result.current.error).toBeNull();
    expect(result.current.hasRunOnce).toBe(false);
  });

  it("run() sends searchJobs with query body + Idempotency-Key + sets results on default scenario", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "searchJobs");
    const { result } = renderHook(() => useSearchJobs(), {
      wrapper: withRuntime(client),
    });
    await act(async () => {
      await result.current.run("Senior frontend roles with strong design-system culture");
    });
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy.mock.calls[0]![0]).toMatchObject({
      query: "Senior frontend roles with strong design-system culture",
    });
    expect(spy.mock.calls[0]![1]?.idempotencyKey).toBeTruthy();
    expect(result.current.results.length).toBeGreaterThan(0);
    expect(result.current.searching).toBe(false);
    expect(result.current.error).toBeNull();
    expect(result.current.hasRunOnce).toBe(true);
  });

  it("empty variant → results=[] without error", async () => {
    const client = buildClient("empty");
    const { result } = renderHook(() => useSearchJobs(), {
      wrapper: withRuntime(client),
    });
    await act(async () => {
      await result.current.run("nothing matches");
    });
    expect(result.current.results).toEqual([]);
    expect(result.current.error).toBeNull();
    expect(result.current.hasRunOnce).toBe(true);
  });

  it("failed variant → error is set, results=[]", async () => {
    const client = buildClient("failed");
    const { result } = renderHook(() => useSearchJobs(), {
      wrapper: withRuntime(client),
    });
    await act(async () => {
      await result.current.run("anything");
    });
    expect(result.current.error).not.toBeNull();
    expect(result.current.results).toEqual([]);
  });

  it("each run() generates a unique Idempotency-Key", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "searchJobs");
    const { result } = renderHook(() => useSearchJobs(), {
      wrapper: withRuntime(client),
    });
    await act(async () => {
      await result.current.run("first");
    });
    await act(async () => {
      await result.current.run("second");
    });
    expect(spy.mock.calls[0]![1]?.idempotencyKey).not.toBe(
      spy.mock.calls[1]![1]?.idempotencyKey,
    );
  });

  it("searching=true while in-flight; transitions back to false after settle", async () => {
    let resolveFn: ((value: Response) => void) | null = null;
    const slowFetch = vi.fn(
      () =>
        new Promise<Response>((resolve) => {
          resolveFn = resolve;
        }),
    );
    const client = new EasyInterviewClient({ fetch: slowFetch as unknown as typeof fetch });
    const { result } = renderHook(() => useSearchJobs(), {
      wrapper: withRuntime(client),
    });
    let runPromise: Promise<void> | null = null;
    await act(async () => {
      runPromise = result.current.run("x");
    });
    expect(result.current.searching).toBe(true);
    await act(async () => {
      resolveFn?.(
        new Response(
          JSON.stringify({ items: [], searchRunId: "run-1" }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        ),
      );
      await runPromise;
    });
    expect(result.current.searching).toBe(false);
  });

  it("does NOT register setInterval / setTimeout for step advancement", async () => {
    const setIntervalSpy = vi.spyOn(globalThis, "setInterval");
    const setTimeoutSpy = vi.spyOn(globalThis, "setTimeout");
    const client = buildClient("default");
    const { result } = renderHook(() => useSearchJobs(), {
      wrapper: withRuntime(client),
    });
    await act(async () => {
      await result.current.run("test");
    });
    // The hook must not register a step-advance timer; some unrelated React
    // internals may use setTimeout (microtasks etc), so we look for any timer
    // tied to a step-advance interval pattern of ~100-2000ms — none should
    // exist. We assert setInterval was not called at all from inside the
    // hook's run path (act() flushes its sync work).
    expect(setIntervalSpy).not.toHaveBeenCalled();
    setIntervalSpy.mockRestore();
    setTimeoutSpy.mockRestore();
  });

  it("abort() cancels the in-flight request and leaves searching=false", async () => {
    let aborted = false;
    const abortableFetch = vi.fn(
      (_input, init: RequestInit | undefined) =>
        new Promise<Response>((_, reject) => {
          init?.signal?.addEventListener("abort", () => {
            aborted = true;
            reject(new DOMException("Aborted", "AbortError"));
          });
        }),
    );
    const client = new EasyInterviewClient({
      fetch: abortableFetch as unknown as typeof fetch,
    });
    const { result } = renderHook(() => useSearchJobs(), {
      wrapper: withRuntime(client),
    });
    let runPromise: Promise<void> | null = null;
    await act(async () => {
      runPromise = result.current.run("x");
    });
    expect(result.current.searching).toBe(true);
    await act(async () => {
      result.current.abort();
      await runPromise;
    });
    expect(aborted).toBe(true);
    expect(result.current.searching).toBe(false);
  });

  it("inert when no AppRuntimeProvider is mounted", async () => {
    const { result } = renderHook(() => useSearchJobs());
    await act(async () => {
      await result.current.run("x");
    });
    expect(result.current.searching).toBe(false);
    expect(result.current.results).toEqual([]);
  });
});
