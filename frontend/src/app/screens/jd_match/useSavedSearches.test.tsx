// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";

import listSavedSearchesFixture from "../../../../../openapi/fixtures/JobMatch/listSavedSearches.json";
import createSavedSearchFixture from "../../../../../openapi/fixtures/JobMatch/createSavedSearch.json";

import {
  useCreateSavedSearch,
  useSavedSearches,
} from "./useSavedSearches";

function buildClient(opts: {
  listScenario?: string;
  createScenario?: string;
}) {
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const method = init?.method ?? "GET";
      const isList =
        url.includes("/jd-match/saved-searches") && method === "GET";
      const isCreate =
        url.includes("/jd-match/saved-searches") && method === "POST";
      const scenario = isList
        ? opts.listScenario
        : isCreate
          ? opts.createScenario
          : undefined;
      const inner = createFixtureBackedFetch(
        createFixtureRegistry([
          listSavedSearchesFixture,
          createSavedSearchFixture,
        ]),
        scenario ? { scenario } : undefined,
      );
      return inner(input, init);
    },
  });
}

function withRuntime(client: EasyInterviewClient) {
  return ({ children }: { children: ReactNode }) => (
    <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>
  );
}

let toastSpy: ReturnType<typeof vi.fn>;
beforeEach(() => {
  toastSpy = vi.fn();
  (window as unknown as { eiToast?: typeof toastSpy }).eiToast = toastSpy;
});
afterEach(() => {
  delete (window as unknown as { eiToast?: unknown }).eiToast;
  vi.restoreAllMocks();
});

describe("useSavedSearches (item 4.3)", () => {
  it("calls listSavedSearches exactly once on mount when active=true", async () => {
    const client = buildClient({ listScenario: "default" });
    const spy = vi.spyOn(client, "listSavedSearches");
    const { result } = renderHook(() => useSavedSearches(true), {
      wrapper: withRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(spy).toHaveBeenCalledTimes(1);
    expect(result.current.items.length).toBeGreaterThan(0);
    expect(result.current.error).toBeNull();
  });

  it("does NOT call listSavedSearches when active=false (deferred until tab opens)", async () => {
    const client = buildClient({ listScenario: "default" });
    const spy = vi.spyOn(client, "listSavedSearches");
    renderHook(() => useSavedSearches(false), {
      wrapper: withRuntime(client),
    });
    // Wait a tick to be sure the hook's effect didn't enqueue a call
    await new Promise((r) => setTimeout(r, 5));
    expect(spy).not.toHaveBeenCalled();
  });

  it("4xx variant → error is set, items=[]", async () => {
    const client = buildClient({ listScenario: "4xx" });
    const { result } = renderHook(() => useSavedSearches(true), {
      wrapper: withRuntime(client),
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).not.toBeNull();
    expect(result.current.items).toEqual([]);
  });

  it("inert when no AppRuntimeProvider is mounted", () => {
    const { result } = renderHook(() => useSavedSearches(true));
    expect(result.current.loading).toBe(false);
    expect(result.current.items).toEqual([]);
    expect(result.current.error).toBeNull();
  });
});

describe("useCreateSavedSearch (item 4.3)", () => {
  it("create() sends label+query body + Idempotency-Key, prepends to applied list, dispatches ok toast", async () => {
    const client = buildClient({ createScenario: "default" });
    const spy = vi.spyOn(client, "createSavedSearch");
    const onCreated = vi.fn();
    const { result } = renderHook(
      () => useCreateSavedSearch({ onCreated }),
      { wrapper: withRuntime(client) },
    );
    await act(async () => {
      await result.current.create({
        label: "Frontend remote",
        query: "frontend remote",
      });
    });
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy.mock.calls[0]![0]).toEqual({
      label: "Frontend remote",
      query: "frontend remote",
    });
    expect(spy.mock.calls[0]![1]?.idempotencyKey).toBeTruthy();
    expect(onCreated).toHaveBeenCalledTimes(1);
    expect(toastSpy).toHaveBeenCalledTimes(1);
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("ok");
  });

  it("4xx-validation variant → error is set, onCreated NOT called, danger toast", async () => {
    const client = buildClient({ createScenario: "4xx-validation" });
    const onCreated = vi.fn();
    const { result } = renderHook(
      () => useCreateSavedSearch({ onCreated }),
      { wrapper: withRuntime(client) },
    );
    await act(async () => {
      await result.current.create({ label: "bad", query: "bad" });
    });
    expect(onCreated).not.toHaveBeenCalled();
    expect(result.current.error).not.toBeNull();
    expect(toastSpy).toHaveBeenCalledTimes(1);
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("danger");
  });

  it("does not log label / query to console (privacy)", async () => {
    const client = buildClient({ createScenario: "default" });
    const logSpy = vi.spyOn(console, "log").mockImplementation(() => undefined);
    const errSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
    const warnSpy = vi
      .spyOn(console, "warn")
      .mockImplementation(() => undefined);
    const { result } = renderHook(() => useCreateSavedSearch({}), {
      wrapper: withRuntime(client),
    });
    await act(async () => {
      await result.current.create({
        label: "secret-saved-search-label-xyz",
        query: "secret-search-query-xyz",
      });
    });
    for (const spy of [logSpy, errSpy, warnSpy]) {
      for (const call of spy.mock.calls) {
        const text = call.map((v) => (typeof v === "string" ? v : "")).join(" ");
        expect(text).not.toContain("secret-saved-search-label-xyz");
        expect(text).not.toContain("secret-search-query-xyz");
      }
    }
  });
});
