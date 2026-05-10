// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import type { JobMatchRecommendation } from "../../../api/generated/types";

import addToWatchlistFixture from "../../../../../openapi/fixtures/JobMatch/addToWatchlist.json";
import removeFromWatchlistFixture from "../../../../../openapi/fixtures/JobMatch/removeFromWatchlist.json";

import { useToggleWatchlist } from "./useToggleWatchlist";

declare global {
  // eslint-disable-next-line no-var
  var eiToast:
    | ((
        message: string,
        opts?: { tone?: "ok" | "warn" | "danger" | "neutral" },
      ) => void)
    | undefined;
}

function makeRec(
  overrides: Partial<JobMatchRecommendation> = {},
): JobMatchRecommendation {
  return {
    id: "jm-1",
    title: "Senior FE",
    company: "Acme",
    companyTag: null,
    level: null,
    location: "Remote",
    comp: null,
    posted: "2 days ago",
    score: 92,
    fit: { must: 4, total: 5, plus: 3, totalPlus: 4 },
    reasons: ["Reason A"],
    risks: [],
    highlights: [],
    seen: true,
    saved: false,
    sourceUrl: null,
    sourceLabel: null,
    networkNote: null,
    similarInterviewers: null,
    interviewHypotheses: [],
    provenance: {
      promptVersion: "p.v1",
      rubricVersion: "r.v1",
      modelId: "m",
      language: "zh-CN",
      featureFlag: "none",
      dataSourceVersion: "v1",
    },
    ...overrides,
  };
}

function buildClient(opts?: { addScenario?: string; removeScenario?: string }) {
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url = typeof input === "string" ? input : (input as URL | Request).toString();
      const isAdd =
        url.includes("/jd-match/watchlist") &&
        (init?.method ?? "GET") === "POST";
      const isRemove =
        url.includes("/jd-match/watchlist/") &&
        (init?.method ?? "GET") === "DELETE";
      const scenario = isAdd
        ? opts?.addScenario
        : isRemove
          ? opts?.removeScenario
          : undefined;
      const inner = createFixtureBackedFetch(
        createFixtureRegistry([
          addToWatchlistFixture,
          removeFromWatchlistFixture,
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
  globalThis.eiToast = toastSpy;
  (globalThis as unknown as { window?: Window }).window =
    (globalThis as unknown as { window?: Window }).window ?? (globalThis as unknown as Window);
  (window as unknown as { eiToast?: typeof toastSpy }).eiToast = toastSpy;
});

afterEach(() => {
  delete (window as unknown as { eiToast?: unknown }).eiToast;
  globalThis.eiToast = undefined;
  vi.restoreAllMocks();
});

describe("useToggleWatchlist (item 3.3)", () => {
  it("Save flow (saved=false) → optimistic apply true, addToWatchlist with body+IK, success toast", async () => {
    const client = buildClient({ addScenario: "default" });
    const addSpy = vi.spyOn(client, "addToWatchlist");
    const applyOptimistic = vi.fn();
    const { result } = renderHook(
      () => useToggleWatchlist({ applyOptimistic }),
      { wrapper: withRuntime(client) },
    );
    const rec = makeRec({ id: "jm-a", saved: false });
    await act(async () => {
      await result.current.toggleSave(rec);
    });
    expect(applyOptimistic).toHaveBeenCalledWith("jm-a", true);
    expect(addSpy).toHaveBeenCalledTimes(1);
    expect(addSpy.mock.calls[0]![0]).toEqual({ jobMatchId: "jm-a" });
    expect(addSpy.mock.calls[0]![1]?.idempotencyKey).toBeTruthy();
    expect(toastSpy).toHaveBeenCalledTimes(1);
    expect(toastSpy.mock.calls[0]![0]).toMatch(/saved/i);
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("ok");
  });

  it("Unsave flow (saved=true) → optimistic apply false, removeFromWatchlist with id+IK, neutral/ok toast", async () => {
    const client = buildClient({ removeScenario: "default" });
    const removeSpy = vi.spyOn(client, "removeFromWatchlist");
    const applyOptimistic = vi.fn();
    const { result } = renderHook(
      () => useToggleWatchlist({ applyOptimistic }),
      { wrapper: withRuntime(client) },
    );
    const rec = makeRec({ id: "jm-b", saved: true });
    await act(async () => {
      await result.current.toggleSave(rec);
    });
    expect(applyOptimistic).toHaveBeenCalledWith("jm-b", false);
    expect(removeSpy).toHaveBeenCalledTimes(1);
    expect(removeSpy.mock.calls[0]![0]).toBe("jm-b");
    expect(removeSpy.mock.calls[0]![1]?.idempotencyKey).toBeTruthy();
    expect(toastSpy).toHaveBeenCalledTimes(1);
    expect(toastSpy.mock.calls[0]![0].toLowerCase()).toMatch(/removed|watchlist/i);
  });

  it("addToWatchlist 4xx → revert applyOptimistic to original state + error toast", async () => {
    const client = buildClient({ addScenario: "4xx-validation" });
    const applyOptimistic = vi.fn();
    const { result } = renderHook(
      () => useToggleWatchlist({ applyOptimistic }),
      { wrapper: withRuntime(client) },
    );
    const rec = makeRec({ id: "jm-c", saved: false });
    await act(async () => {
      await result.current.toggleSave(rec);
    });
    expect(applyOptimistic).toHaveBeenNthCalledWith(1, "jm-c", true);
    expect(applyOptimistic).toHaveBeenNthCalledWith(2, "jm-c", false);
    expect(toastSpy).toHaveBeenCalledTimes(1);
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("danger");
  });

  it("removeFromWatchlist 4xx → revert + error toast", async () => {
    const client = buildClient({ removeScenario: "4xx-not-found" });
    const applyOptimistic = vi.fn();
    const { result } = renderHook(
      () => useToggleWatchlist({ applyOptimistic }),
      { wrapper: withRuntime(client) },
    );
    const rec = makeRec({ id: "jm-d", saved: true });
    await act(async () => {
      await result.current.toggleSave(rec);
    });
    expect(applyOptimistic).toHaveBeenNthCalledWith(1, "jm-d", false);
    expect(applyOptimistic).toHaveBeenNthCalledWith(2, "jm-d", true);
    expect(toastSpy).toHaveBeenCalledTimes(1);
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("danger");
  });

  it("each call generates a unique Idempotency-Key", async () => {
    const client = buildClient({ addScenario: "default" });
    const addSpy = vi.spyOn(client, "addToWatchlist");
    const { result } = renderHook(
      () => useToggleWatchlist({ applyOptimistic: vi.fn() }),
      { wrapper: withRuntime(client) },
    );
    await act(async () => {
      await result.current.toggleSave(makeRec({ id: "jm-1", saved: false }));
    });
    await act(async () => {
      await result.current.toggleSave(makeRec({ id: "jm-2", saved: false }));
    });
    const ik1 = addSpy.mock.calls[0]![1]?.idempotencyKey;
    const ik2 = addSpy.mock.calls[1]![1]?.idempotencyKey;
    expect(ik1).toBeTruthy();
    expect(ik2).toBeTruthy();
    expect(ik1).not.toBe(ik2);
  });

  it("does not throw when window.eiToast is missing", async () => {
    delete (window as unknown as { eiToast?: unknown }).eiToast;
    globalThis.eiToast = undefined;
    const client = buildClient({ addScenario: "default" });
    const { result } = renderHook(
      () => useToggleWatchlist({ applyOptimistic: vi.fn() }),
      { wrapper: withRuntime(client) },
    );
    await act(async () => {
      await result.current.toggleSave(makeRec({ saved: false }));
    });
    // no throw == passing
    expect(true).toBe(true);
  });

  it("race handling: only latest toggle outcome wins (later success overrides earlier revert)", async () => {
    const client = buildClient({ addScenario: "default" });
    const addSpy = vi
      .spyOn(client, "addToWatchlist")
      .mockImplementationOnce(
        () =>
          new Promise((_, reject) =>
            setTimeout(() => reject(new Error("HTTP 422 — earlier failed")), 50),
          ),
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) =>
            setTimeout(
              () =>
                resolve({
                  id: "wl-2",
                  linkedJobMatchId: "jm-x",
                  title: "x",
                  company: "x",
                  tone: "ok",
                  addedAt: "2026-05-10T00:00:00Z",
                }),
              5,
            ),
          ),
      );
    const applyOptimistic = vi.fn();
    const { result } = renderHook(
      () => useToggleWatchlist({ applyOptimistic }),
      { wrapper: withRuntime(client) },
    );
    const rec = makeRec({ id: "jm-x", saved: false });

    // Fire two consecutive saves on the same id
    await act(async () => {
      const p1 = result.current.toggleSave(rec);
      const p2 = result.current.toggleSave(rec);
      await Promise.all([p1, p2]);
    });

    // The earlier (slow) failure must NOT revert state because seq advanced.
    // Expected calls: optimistic apply (call 1, true), optimistic apply (call 2, true),
    // then ONLY the later success path runs -> no revert.
    const reverts = applyOptimistic.mock.calls.filter(
      ([, savedNext]) => savedNext === false,
    );
    expect(reverts).toHaveLength(0);
    expect(addSpy).toHaveBeenCalledTimes(2);
  });

  it("returns inert toggleSave when no AppRuntimeProvider is mounted", async () => {
    const applyOptimistic = vi.fn();
    const { result } = renderHook(() =>
      useToggleWatchlist({ applyOptimistic }),
    );
    await act(async () => {
      await result.current.toggleSave(makeRec({ saved: false }));
    });
    expect(applyOptimistic).not.toHaveBeenCalled();
  });

  it("does not log jobMatchId or sourceUrl to console (privacy)", async () => {
    const client = buildClient({ addScenario: "default" });
    const logSpy = vi.spyOn(console, "log").mockImplementation(() => undefined);
    const errSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);
    const warnSpy = vi.spyOn(console, "warn").mockImplementation(() => undefined);
    const { result } = renderHook(
      () => useToggleWatchlist({ applyOptimistic: vi.fn() }),
      { wrapper: withRuntime(client) },
    );
    const rec = makeRec({
      id: "jm-secret-id",
      saved: false,
      sourceUrl: "https://secret.example/url",
    });
    await act(async () => {
      await result.current.toggleSave(rec);
    });
    for (const spy of [logSpy, errSpy, warnSpy]) {
      for (const call of spy.mock.calls) {
        const text = call.map((v) => (typeof v === "string" ? v : "")).join(" ");
        expect(text).not.toContain("jm-secret-id");
        expect(text).not.toContain("secret.example");
      }
    }
  });
});
