// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import type { FC, ReactNode } from "react";

import type { ResumeVersion } from "../../../../../api/generated/types";
import { AppRuntimeContext } from "../../../../runtime/AppRuntimeProvider";
import type { AppRuntimeValue } from "../../../../runtime/AppRuntimeProvider";
import {
  UpdateResumeVersionError,
  filterUpdateResumeVersionPayload,
  useUpdateResumeVersion,
} from "./useUpdateResumeVersion";

import updateFixture from "../../../../../../../openapi/fixtures/Resumes/updateResumeVersion.json";

const VERSION_ID = "0195f2d0-0001-7000-8000-000000000202";

const buildRuntime = (
  overrides: Partial<AppRuntimeValue["client"]>,
): AppRuntimeValue =>
  ({
    client: {
      updateResumeVersion: vi.fn(),
      ...overrides,
    },
    runtime: { status: "ready", config: {} as never },
    auth: { status: "authenticated", user: { id: "u1" } as never },
    refreshAuth: vi.fn(),
  } as unknown as AppRuntimeValue);

const renderHookWithRuntime = (runtime: AppRuntimeValue) => {
  const Wrapper: FC<{ children: ReactNode }> = ({ children }) => (
    <AppRuntimeContext.Provider value={runtime}>
      {children}
    </AppRuntimeContext.Provider>
  );
  return renderHook(() => useUpdateResumeVersion(), { wrapper: Wrapper });
};

const defaultFixture = updateFixture.scenarios.default.response.body as ResumeVersion;

describe("filterUpdateResumeVersionPayload", () => {
  it("preserves allowed fields", () => {
    const result = filterUpdateResumeVersionPayload({
      displayName: "name",
      focusAngle: "leadership",
      matchScore: 0.84,
      structuredProfile: { headline: "h" },
    });
    expect(result).toEqual({
      displayName: "name",
      focusAngle: "leadership",
      matchScore: 0.84,
      structuredProfile: { headline: "h" },
    });
  });

  it("throws on disallowed fields", () => {
    for (const key of [
      "versionType",
      "resumeAssetId",
      "parentVersionId",
      "targetJobId",
      "seedStrategy",
    ]) {
      expect(() => filterUpdateResumeVersionPayload({ [key]: "v" })).toThrow(
        new RegExp(key),
      );
    }
  });
});

describe("useUpdateResumeVersion", () => {
  it("updates with an Idempotency-Key and resolves with the new ResumeVersion", async () => {
    const updateResumeVersion = vi.fn().mockResolvedValueOnce(defaultFixture);
    const runtime = buildRuntime({ updateResumeVersion });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await result.current.update({
        versionId: VERSION_ID,
        payload: {
          structuredProfile: { headline: "new headline", summary: "new" },
        },
      });
    });
    expect(updateResumeVersion).toHaveBeenCalledTimes(1);
    const [, body, opts] = updateResumeVersion.mock.calls[0]!;
    expect(body).toEqual({
      structuredProfile: { headline: "new headline", summary: "new" },
    });
    expect(opts?.idempotencyKey).toMatch(/^v1\.\d+\..+/);
  });

  it("replays the same Idempotency-Key when payload + versionId match", async () => {
    const updateResumeVersion = vi.fn().mockResolvedValue(defaultFixture);
    const runtime = buildRuntime({ updateResumeVersion });
    const { result } = renderHookWithRuntime(runtime);
    const payload = { displayName: "v3 trimmed" };
    await act(async () => {
      await result.current.update({ versionId: VERSION_ID, payload });
    });
    await act(async () => {
      await result.current.update({ versionId: VERSION_ID, payload });
    });
    const first = updateResumeVersion.mock.calls[0]![2].idempotencyKey;
    const second = updateResumeVersion.mock.calls[1]![2].idempotencyKey;
    expect(second).toBe(first);
  });

  it("maps 422 to validation and clears IK cache", async () => {
    const updateResumeVersion = vi
      .fn()
      .mockRejectedValue(
        new Error(
          'HTTP 422 Unprocessable: {"error":{"code":"VALIDATION_FAILED","details":{"field":"structuredProfile.headline"}}}',
        ),
      );
    const runtime = buildRuntime({ updateResumeVersion });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await expect(
        result.current.update({
          versionId: VERSION_ID,
          payload: { displayName: "x" },
        }),
      ).rejects.toBeInstanceOf(UpdateResumeVersionError);
    });
    expect(result.current.lastError?.kind).toBe("validation");
    expect(result.current.lastError?.field).toBe(
      "structuredProfile.headline",
    );
    expect(
      result.current.peekIdempotencyKey(VERSION_ID, { displayName: "x" }),
    ).toBeNull();
  });

  it("maps 409 to idempotency_conflict and keeps the IK cache (allow user to revise the payload)", async () => {
    const updateResumeVersion = vi
      .fn()
      .mockRejectedValue(new Error("HTTP 409 Conflict: {}"));
    const runtime = buildRuntime({ updateResumeVersion });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await expect(
        result.current.update({
          versionId: VERSION_ID,
          payload: { displayName: "x" },
        }),
      ).rejects.toBeInstanceOf(UpdateResumeVersionError);
    });
    expect(result.current.lastError?.kind).toBe("idempotency_conflict");
    expect(
      result.current.peekIdempotencyKey(VERSION_ID, { displayName: "x" }),
    ).not.toBeNull();
  });

  it("maps 404 to cross_user", async () => {
    const updateResumeVersion = vi
      .fn()
      .mockRejectedValue(new Error("HTTP 404 Not Found: {}"));
    const runtime = buildRuntime({ updateResumeVersion });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await expect(
        result.current.update({
          versionId: VERSION_ID,
          payload: { displayName: "x" },
        }),
      ).rejects.toBeInstanceOf(UpdateResumeVersionError);
    });
    expect(result.current.lastError?.kind).toBe("cross_user");
  });

  it("captures the canonical updateResumeVersion fixture scenario keys", () => {
    expect(Object.keys(updateFixture.scenarios).sort()).toEqual(
      ["default", "idempotency-replay", "validation-error-422"].sort(),
    );
  });
});
