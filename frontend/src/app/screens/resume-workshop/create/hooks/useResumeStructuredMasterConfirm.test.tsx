// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../../../api/generated/client";
import type { ResumeVersion } from "../../../../../api/generated/types";
import { AppRuntimeProvider } from "../../../../runtime/AppRuntimeProvider";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../../api/mockTransport";
import {
  useResumeStructuredMasterConfirm,
  parseConfirmError,
} from "./useResumeStructuredMasterConfirm";

import getRuntimeConfigFixture from "../../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../../openapi/fixtures/Auth/getMe.json";

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getRuntimeConfigFixture, getMeFixture]),
      { scenario },
    ),
  });
}

function buildWrapper(client: EasyInterviewClient) {
  return ({ children }: { children: ReactNode }) => (
    <AppRuntimeProvider
      client={client}
      requestOptions={{
        getMe: { headers: { Prefer: "example=authenticated" } },
      }}
    >
      {children}
    </AppRuntimeProvider>
  );
}

const SAVED_VERSION: ResumeVersion = {
  id: "0195f2d0-0001-7000-8000-000000000201",
  resumeAssetId: "01918fa0-0000-7000-8000-000000001000",
  versionType: "structured_master",
  displayName: "Structured master",
  parentVersionId: null,
  targetJobId: null,
  seedStrategy: null,
  focusAngle: null,
  structuredProfile: { headline: "x" },
  suggestions: [],
	provenance: {
		promptVersion: "resume_profile.v1",
		rubricVersion: "not_applicable",
		modelId: "fixture-model",
		language: "zh-CN",
		featureFlag: "none",
		dataSourceVersion: "not_applicable",
	} as ResumeVersion["provenance"],
  modelId: null,
  promptVersion: null,
  provider: null,
  rubricVersion: null,
  matchScore: null,
  createdAt: "2026-05-17T00:00:00Z",
  updatedAt: "2026-05-17T00:00:00Z",
  deletedAt: null,
};

describe("parseConfirmError", () => {
  it("recognises 409 RESUME_STRUCTURED_MASTER_ALREADY_EXISTS from message text", () => {
    expect(
      parseConfirmError(
        new Error("HTTP 409 Conflict: RESUME_STRUCTURED_MASTER_ALREADY_EXISTS"),
      ),
    ).toMatchObject({ kind: "already_exists" });
  });
  it("recognises 422 VALIDATION_FAILED from message text", () => {
    expect(
      parseConfirmError(new Error("HTTP 422 Unprocessable: VALIDATION_FAILED")),
    ).toMatchObject({ kind: "validation" });
  });
  it("returns kind=error for opaque failures", () => {
    expect(parseConfirmError(new Error("network down"))).toMatchObject({
      kind: "error",
    });
  });
});

describe("useResumeStructuredMasterConfirm", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("returns saved + the new ResumeVersion on 201, passing an Idempotency-Key header", async () => {
    const client = buildClient("default");
    const spy = vi
      .spyOn(client, "confirmResumeStructuredMaster")
      .mockResolvedValue(SAVED_VERSION);
    const { result } = renderHook(() => useResumeStructuredMasterConfirm(), {
      wrapper: buildWrapper(client),
    });

    let outcome: unknown;
    await act(async () => {
      outcome = await result.current.confirm({
        resumeAssetId: "01918fa0-0000-7000-8000-000000001000",
        body: {
          displayName: "Structured master",
          language: "zh",
          structuredProfile: { headline: "x" },
        },
      });
    });
    expect(outcome).toMatchObject({ kind: "saved", version: SAVED_VERSION });
    const call = spy.mock.calls[0]!;
    expect(call[2]?.idempotencyKey).toMatch(/^v1\.\d+\.[0-9a-f-]{36}$/);
    expect(call[2]?.headers).toMatchObject({ "Accept-Language": "zh" });
  });

  it("reuses the same idempotency key on retry (replay) for the same asset", async () => {
    const client = buildClient("default");
    const spy = vi
      .spyOn(client, "confirmResumeStructuredMaster")
      .mockResolvedValue(SAVED_VERSION);
    const { result } = renderHook(() => useResumeStructuredMasterConfirm(), {
      wrapper: buildWrapper(client),
    });

    await act(async () => {
      await result.current.confirm({
        resumeAssetId: "asset-1",
        body: {
          displayName: "Master",
          language: "zh",
          structuredProfile: {},
        },
      });
    });
    await act(async () => {
      await result.current.confirm({
        resumeAssetId: "asset-1",
        body: {
          displayName: "Master",
          language: "zh",
          structuredProfile: {},
        },
      });
    });
    expect(spy).toHaveBeenCalledTimes(2);
    const firstKey = spy.mock.calls[0]![2]?.idempotencyKey;
    const secondKey = spy.mock.calls[1]![2]?.idempotencyKey;
    expect(firstKey).toBe(secondKey);
  });

  it("maps 409 already-exists to kind=already_exists and pre-resolves the existing master via listResumeVersions", async () => {
    const client = buildClient("default");
    vi.spyOn(client, "confirmResumeStructuredMaster").mockRejectedValue(
      new Error("HTTP 409 Conflict: RESUME_STRUCTURED_MASTER_ALREADY_EXISTS"),
    );
    vi.spyOn(client, "listResumeVersions").mockResolvedValue({
      items: [
        {
          ...SAVED_VERSION,
          id: "0195f2d0-0001-7000-8000-000000000777",
        },
      ],
			pageInfo: { nextCursor: null, pageSize: 20, hasMore: false },
		});
    const { result } = renderHook(() => useResumeStructuredMasterConfirm(), {
      wrapper: buildWrapper(client),
    });

    let outcome: unknown;
    await act(async () => {
      outcome = await result.current.confirm({
        resumeAssetId: "asset-1",
        body: {
          displayName: "Master",
          language: "zh",
          structuredProfile: {},
        },
      });
    });
    expect(outcome).toMatchObject({
      kind: "already_exists",
      existingMasterId: "0195f2d0-0001-7000-8000-000000000777",
    });
  });

  it("maps 422 VALIDATION_FAILED to kind=validation and resets the idempotency cache", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "confirmResumeStructuredMaster");
    spy.mockRejectedValueOnce(
      new Error("HTTP 422 Unprocessable: VALIDATION_FAILED"),
    );
    spy.mockResolvedValueOnce(SAVED_VERSION);
    const { result } = renderHook(() => useResumeStructuredMasterConfirm(), {
      wrapper: buildWrapper(client),
    });

    let firstOutcome: unknown;
    await act(async () => {
      firstOutcome = await result.current.confirm({
        resumeAssetId: "asset-1",
        body: {
          displayName: "",
          language: "zh",
          structuredProfile: {},
        },
      });
    });
    expect(firstOutcome).toMatchObject({ kind: "validation" });
    await act(async () => {
      await result.current.confirm({
        resumeAssetId: "asset-1",
        body: {
          displayName: "Master",
          language: "zh",
          structuredProfile: {},
        },
      });
    });
    expect(spy).toHaveBeenCalledTimes(2);
    const firstKey = spy.mock.calls[0]![2]?.idempotencyKey;
    const secondKey = spy.mock.calls[1]![2]?.idempotencyKey;
    expect(firstKey).not.toBe(secondKey);
  });
});
