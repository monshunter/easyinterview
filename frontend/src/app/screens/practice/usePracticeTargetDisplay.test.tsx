/**
 * @vitest-environment jsdom
 */

import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import type {
  EasyInterviewClient,
  RequestOptions,
} from "../../../api/generated/client";
import type { TargetJob } from "../../../api/generated/types";
import {
  AppRuntimeContext,
  type AppRuntimeValue,
} from "../../runtime/AppRuntimeProvider";
import { usePracticeTargetDisplay } from "./usePracticeTargetDisplay";

const ROUTE_TARGET_ID = "target-route";
const CONTEXT_TARGET_ID = "target-context";
const SESSION_TARGET_ID = "target-session";

interface Deferred<T> {
  promise: Promise<T>;
  resolve: (value: T) => void;
  reject: (reason: unknown) => void;
}

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<T>((onResolve, onReject) => {
    resolve = onResolve;
    reject = onReject;
  });
  return { promise, resolve, reject };
}

function targetJob(
  id: string,
  companyName: string,
  title: string,
): TargetJob {
  return { id, companyName, title } as TargetJob;
}

function runtimeWrapper(client: EasyInterviewClient) {
  const value: AppRuntimeValue = {
    client,
    runtime: { status: "loading" },
    auth: { status: "unauthenticated" },
    refreshAuth: vi.fn(),
  };

  return function RuntimeWrapper({ children }: { children: ReactNode }) {
    return (
      <AppRuntimeContext.Provider value={value}>
        {children}
      </AppRuntimeContext.Provider>
    );
  };
}

function clientWith(
  getTargetJob: (
    targetJobId: string,
    options?: RequestOptions,
  ) => Promise<TargetJob>,
): EasyInterviewClient {
  return { getTargetJob: vi.fn(getTargetJob) } as unknown as EasyInterviewClient;
}

describe("usePracticeTargetDisplay", () => {
  it("uses route/context only before load, then keeps the server-session target authoritative", async () => {
    const routeResult = deferred<TargetJob>();
    const sessionResult = deferred<TargetJob>();
    const getTargetJob = vi.fn(
      (targetJobId: string): Promise<TargetJob> =>
        targetJobId === SESSION_TARGET_ID
          ? sessionResult.promise
          : routeResult.promise,
    );
    const client = clientWith(getTargetJob);

    const { result, rerender } = renderHook(
      ({
        session,
        routeTargetJobId,
        contextTargetJobId,
      }: {
        session: { targetJobId: string } | null;
        routeTargetJobId?: string;
        contextTargetJobId?: string;
      }) =>
        usePracticeTargetDisplay({
          session,
          routeTargetJobId,
          contextTargetJobId,
        }),
      {
        initialProps: {
          session: null as { targetJobId: string } | null,
          routeTargetJobId: ROUTE_TARGET_ID,
          contextTargetJobId: CONTEXT_TARGET_ID,
        },
        wrapper: runtimeWrapper(client),
      },
    );

    expect(result.current.targetJobId).toBe(ROUTE_TARGET_ID);
    expect(result.current.loading).toBe(true);
    expect(result.current.companyName).toBeNull();
    expect(getTargetJob).toHaveBeenCalledWith(
      ROUTE_TARGET_ID,
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    );

    await act(async () => {
      routeResult.resolve(targetJob(ROUTE_TARGET_ID, "Route Co", "Route Role"));
      await routeResult.promise;
    });
    expect(result.current.companyName).toBe("Route Co");
    expect(result.current.title).toBe("Route Role");

    rerender({
      session: { targetJobId: SESSION_TARGET_ID },
      routeTargetJobId: ROUTE_TARGET_ID,
      contextTargetJobId: CONTEXT_TARGET_ID,
    });
    expect(result.current.targetJobId).toBe(SESSION_TARGET_ID);
    expect(result.current.companyName).toBeNull();
    expect(getTargetJob).toHaveBeenLastCalledWith(
      SESSION_TARGET_ID,
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    );

    await act(async () => {
      sessionResult.resolve(
        targetJob(SESSION_TARGET_ID, "Server Co", "Server Role"),
      );
      await sessionResult.promise;
    });
    expect(result.current.companyName).toBe("Server Co");
    expect(result.current.title).toBe("Server Role");

    rerender({
      session: { targetJobId: SESSION_TARGET_ID },
      routeTargetJobId: "target-route-changed",
      contextTargetJobId: "target-context-changed",
    });
    expect(getTargetJob).toHaveBeenCalledTimes(2);
    expect(result.current.targetJobId).toBe(SESSION_TARGET_ID);
    expect(result.current.companyName).toBe("Server Co");
  });

  it("uses the interview-context target when the session and route target are not loaded", async () => {
    const getTargetJob = vi.fn(async (targetJobId: string) =>
      targetJob(targetJobId, "Context Co", "Context Role"),
    );
    const client = clientWith(getTargetJob);

    const { result } = renderHook(
      () =>
        usePracticeTargetDisplay({
          session: null,
          contextTargetJobId: CONTEXT_TARGET_ID,
        }),
      { wrapper: runtimeWrapper(client) },
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.targetJobId).toBe(CONTEXT_TARGET_ID);
    expect(result.current.companyName).toBe("Context Co");
    expect(result.current.title).toBe("Context Role");
    expect(getTargetJob).toHaveBeenCalledTimes(1);
  });

  it("aborts the fallback request and ignores its stale response after the session target arrives", async () => {
    const routeResult = deferred<TargetJob>();
    const sessionResult = deferred<TargetJob>();
    const requestSignals = new Map<string, AbortSignal>();
    const getTargetJob = vi.fn(
      (targetJobId: string, options?: RequestOptions): Promise<TargetJob> => {
        if (options?.signal) requestSignals.set(targetJobId, options.signal);
        return targetJobId === SESSION_TARGET_ID
          ? sessionResult.promise
          : routeResult.promise;
      },
    );
    const client = clientWith(getTargetJob);

    const { result, rerender } = renderHook(
      ({ session }: { session: { targetJobId: string } | null }) =>
        usePracticeTargetDisplay({
          session,
          routeTargetJobId: ROUTE_TARGET_ID,
        }),
      {
        initialProps: {
          session: null as { targetJobId: string } | null,
        },
        wrapper: runtimeWrapper(client),
      },
    );

    rerender({ session: { targetJobId: SESSION_TARGET_ID } });
    expect(requestSignals.get(ROUTE_TARGET_ID)?.aborted).toBe(true);

    await act(async () => {
      sessionResult.resolve(
        targetJob(SESSION_TARGET_ID, "Server Co", "Server Role"),
      );
      await sessionResult.promise;
    });
    expect(result.current.companyName).toBe("Server Co");

    await act(async () => {
      routeResult.resolve(targetJob(ROUTE_TARGET_ID, "Stale Co", "Stale Role"));
      await routeResult.promise;
    });
    expect(result.current.targetJobId).toBe(SESSION_TARGET_ID);
    expect(result.current.companyName).toBe("Server Co");
    expect(result.current.title).toBe("Server Role");
  });

  it("keeps display fields empty on loading and error instead of inventing fixture labels", async () => {
    const failure = new Error("HTTP 503 unavailable");
    const request = deferred<TargetJob>();
    const client = clientWith(() => request.promise);

    const { result } = renderHook(
      () =>
        usePracticeTargetDisplay({
          session: null,
          routeTargetJobId: ROUTE_TARGET_ID,
        }),
      { wrapper: runtimeWrapper(client) },
    );

    expect(result.current.loading).toBe(true);
    expect(result.current.companyName).toBeNull();
    expect(result.current.title).toBeNull();

    await act(async () => {
      request.reject(failure);
      await expect(request.promise).rejects.toBe(failure);
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe(failure);
    expect(result.current.companyName).toBeNull();
    expect(result.current.title).toBeNull();
  });
});
