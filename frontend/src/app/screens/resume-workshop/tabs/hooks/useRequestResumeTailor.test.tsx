// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import type { FC, ReactNode } from "react";

import type {
  RequestResumeTailorRequest,
  ResumeTailorRunWithJob,
} from "../../../../../api/generated/types";
import { AppRuntimeContext } from "../../../../runtime/AppRuntimeProvider";
import type { AppRuntimeValue } from "../../../../runtime/AppRuntimeProvider";
import {
  RequestResumeTailorError,
  useRequestResumeTailor,
} from "./useRequestResumeTailor";

import requestFixture from "../../../../../../../openapi/fixtures/ResumeTailor/requestResumeTailor.json";

const baseBody: RequestResumeTailorRequest = {
  resumeId: "01918fa0-0000-7000-8000-000000001000",
  targetJobId: "01918fa0-0000-7000-8000-000000002000",
  mode: "bullet_suggestions",
};

const defaultFixture = requestFixture.scenarios.default.response.body as ResumeTailorRunWithJob;

const buildRuntime = (
  overrides: Partial<AppRuntimeValue["client"]>,
): AppRuntimeValue =>
  ({
    client: {
      requestResumeTailor: vi.fn(),
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
  return renderHook(() => useRequestResumeTailor(), { wrapper: Wrapper });
};

describe("useRequestResumeTailor", () => {
  it("posts request with Idempotency-Key header (default scenario)", async () => {
    const requestResumeTailor = vi.fn().mockResolvedValueOnce(defaultFixture);
    const runtime = buildRuntime({ requestResumeTailor });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await result.current.request(baseBody);
    });
    expect(requestResumeTailor).toHaveBeenCalledTimes(1);
    const [bodyArg, optsArg] = requestResumeTailor.mock.calls[0]!;
    expect(bodyArg).toEqual(baseBody);
    expect(optsArg?.idempotencyKey).toMatch(/^v1\.\d+\..+/);
  });

  it("replays the same Idempotency-Key when the body fingerprint is unchanged", async () => {
    const requestResumeTailor = vi.fn().mockResolvedValue(defaultFixture);
    const runtime = buildRuntime({ requestResumeTailor });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await result.current.request(baseBody);
    });
    await act(async () => {
      await result.current.request(baseBody);
    });
    expect(
      requestResumeTailor.mock.calls[0]![1].idempotencyKey,
    ).toBe(requestResumeTailor.mock.calls[1]![1].idempotencyKey);
  });

  it("rotates the Idempotency-Key when the mode changes between submits", async () => {
    const requestResumeTailor = vi.fn().mockResolvedValue(defaultFixture);
    const runtime = buildRuntime({ requestResumeTailor });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await result.current.request(baseBody);
    });
    await act(async () => {
      await result.current.request({ ...baseBody, mode: "gap_review" });
    });
    expect(
      requestResumeTailor.mock.calls[0]![1].idempotencyKey,
    ).not.toBe(requestResumeTailor.mock.calls[1]![1].idempotencyKey);
  });

  it("maps 422 to RequestResumeTailorError(kind=validation) and clears the IK cache", async () => {
    const requestResumeTailor = vi
      .fn()
      .mockRejectedValueOnce(
        new Error(
          'HTTP 422 Unprocessable: {"error":{"code":"VALIDATION_FAILED"}}',
        ),
      );
    const runtime = buildRuntime({ requestResumeTailor });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await expect(result.current.request(baseBody)).rejects.toBeInstanceOf(
        RequestResumeTailorError,
      );
    });
    expect(result.current.lastError?.kind).toBe("validation");
  });

  it("maps 404 to RequestResumeTailorError(kind=cross_user)", async () => {
    const requestResumeTailor = vi
      .fn()
      .mockRejectedValueOnce(new Error("HTTP 404 Not Found: {}"));
    const runtime = buildRuntime({ requestResumeTailor });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await expect(result.current.request(baseBody)).rejects.toBeInstanceOf(
        RequestResumeTailorError,
      );
    });
    expect(result.current.lastError?.kind).toBe("cross_user");
  });

  it("maps 409 to RequestResumeTailorError(kind=idempotency_conflict)", async () => {
    const requestResumeTailor = vi
      .fn()
      .mockRejectedValueOnce(new Error("HTTP 409 Conflict: {}"));
    const runtime = buildRuntime({ requestResumeTailor });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await expect(result.current.request(baseBody)).rejects.toBeInstanceOf(
        RequestResumeTailorError,
      );
    });
    expect(result.current.lastError?.kind).toBe("idempotency_conflict");
  });

  it("captures the canonical requestResumeTailor fixture scenario keys", () => {
    expect(Object.keys(requestFixture.scenarios).sort()).toEqual(
      ["default", "idempotency-replay"].sort(),
    );
  });
});
