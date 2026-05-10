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

import markJobNotRelevantFixture from "../../../../../openapi/fixtures/JobMatch/markJobNotRelevant.json";

import { useDismissRecommendation } from "./useDismissRecommendation";

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

function buildClient(scenario: string) {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([markJobNotRelevantFixture]),
      { scenario },
    ),
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

describe("useDismissRecommendation (item 3.4)", () => {
  it("invokes applyOptimisticHide and calls markJobNotRelevant with reason+IK", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "markJobNotRelevant");
    const applyOptimisticHide = vi.fn<
      (rec: JobMatchRecommendation) => () => void
    >(() => () => undefined);
    const { result } = renderHook(
      () => useDismissRecommendation({ applyOptimisticHide }),
      { wrapper: withRuntime(client) },
    );
    const rec = makeRec({ id: "jm-x" });
    await act(async () => {
      await result.current.dismiss(rec);
    });
    expect(applyOptimisticHide).toHaveBeenCalledTimes(1);
    expect(applyOptimisticHide.mock.calls[0]![0]).toEqual(
      expect.objectContaining({ id: "jm-x" }),
    );
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy.mock.calls[0]![0]).toBe("jm-x");
    expect(spy.mock.calls[0]![1]).toEqual({ reason: "not_relevant" });
    expect(spy.mock.calls[0]![2]?.idempotencyKey).toBeTruthy();
  });

  it("does not include freeNote in the request body (D-12 / privacy)", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "markJobNotRelevant");
    const { result } = renderHook(
      () =>
        useDismissRecommendation({ applyOptimisticHide: () => () => undefined }),
      { wrapper: withRuntime(client) },
    );
    await act(async () => {
      await result.current.dismiss(makeRec({ id: "jm-y" }));
    });
    const body = spy.mock.calls[0]![1] as unknown as Record<string, unknown>;
    expect("freeNote" in body).toBe(false);
  });

  it("dispatches success toast on default scenario (toastDismissed)", async () => {
    const client = buildClient("default");
    const { result } = renderHook(
      () =>
        useDismissRecommendation({ applyOptimisticHide: () => () => undefined }),
      { wrapper: withRuntime(client) },
    );
    await act(async () => {
      await result.current.dismiss(makeRec({ id: "jm-z" }));
    });
    expect(toastSpy).toHaveBeenCalledTimes(1);
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("neutral");
  });

  it("4xx scenario → invokes the revert callback and shows error toast", async () => {
    const client = buildClient("4xx");
    const revert = vi.fn();
    const applyOptimisticHide = vi.fn(() => revert);
    const { result } = renderHook(
      () => useDismissRecommendation({ applyOptimisticHide }),
      { wrapper: withRuntime(client) },
    );
    await act(async () => {
      await result.current.dismiss(makeRec({ id: "jm-bad" }));
    });
    expect(applyOptimisticHide).toHaveBeenCalledTimes(1);
    expect(revert).toHaveBeenCalledTimes(1);
    expect(toastSpy).toHaveBeenCalledTimes(1);
    expect(toastSpy.mock.calls[0]![1]?.tone).toBe("danger");
  });

  it("each call generates a unique Idempotency-Key", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "markJobNotRelevant");
    const { result } = renderHook(
      () =>
        useDismissRecommendation({ applyOptimisticHide: () => () => undefined }),
      { wrapper: withRuntime(client) },
    );
    await act(async () => {
      await result.current.dismiss(makeRec({ id: "jm-a" }));
    });
    await act(async () => {
      await result.current.dismiss(makeRec({ id: "jm-b" }));
    });
    expect(spy.mock.calls[0]![2]?.idempotencyKey).not.toBe(
      spy.mock.calls[1]![2]?.idempotencyKey,
    );
  });

  it("inert when no AppRuntimeProvider is mounted", async () => {
    const apply = vi.fn();
    const { result } = renderHook(() =>
      useDismissRecommendation({ applyOptimisticHide: apply }),
    );
    await act(async () => {
      await result.current.dismiss(makeRec());
    });
    expect(apply).not.toHaveBeenCalled();
  });

  it("does not log jobMatchId or freeNote to console (privacy)", async () => {
    const client = buildClient("default");
    const logSpy = vi.spyOn(console, "log").mockImplementation(() => undefined);
    const errSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
    const warnSpy = vi
      .spyOn(console, "warn")
      .mockImplementation(() => undefined);
    const { result } = renderHook(
      () =>
        useDismissRecommendation({ applyOptimisticHide: () => () => undefined }),
      { wrapper: withRuntime(client) },
    );
    await act(async () => {
      await result.current.dismiss(makeRec({ id: "jm-secret-dismiss-id" }));
    });
    for (const spy of [logSpy, errSpy, warnSpy]) {
      for (const call of spy.mock.calls) {
        const text = call.map((v) => (typeof v === "string" ? v : "")).join(" ");
        expect(text).not.toContain("jm-secret-dismiss-id");
        expect(text).not.toContain("freeNote");
      }
    }
  });
});
