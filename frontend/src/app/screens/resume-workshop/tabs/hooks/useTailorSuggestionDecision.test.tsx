// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import type { FC, ReactNode } from "react";

import type { ResumeVersion } from "../../../../../api/generated/types";
import { AppRuntimeContext } from "../../../../runtime/AppRuntimeProvider";
import type { AppRuntimeValue } from "../../../../runtime/AppRuntimeProvider";
import { useAcceptResumeTailorSuggestion } from "./useAcceptResumeTailorSuggestion";
import { useRejectResumeTailorSuggestion } from "./useRejectResumeTailorSuggestion";
import {
  SuggestionDecisionError,
  parseSuggestionDecisionError,
} from "./useTailorSuggestionDecision";

import acceptFixture from "../../../../../../../openapi/fixtures/Resumes/acceptResumeTailorSuggestion.json";
import rejectFixture from "../../../../../../../openapi/fixtures/Resumes/rejectResumeTailorSuggestion.json";

const VERSION_ID = "0195f2d0-0001-7000-8000-000000000202";
const SUG_ID = "0195f2d0-0001-7000-8000-000000000302";

const buildRuntime = (
  overrides: Partial<AppRuntimeValue["client"]>,
): AppRuntimeValue =>
  ({
    client: {
      acceptResumeTailorSuggestion: vi.fn(),
      rejectResumeTailorSuggestion: vi.fn(),
      ...overrides,
    },
    runtime: { status: "ready", config: {} as never },
    auth: { status: "authenticated", user: { id: "u1" } as never },
    refreshAuth: vi.fn(),
  } as unknown as AppRuntimeValue);

const wrap = (runtime: AppRuntimeValue): FC<{ children: ReactNode }> =>
  ({ children }) => (
    <AppRuntimeContext.Provider value={runtime}>
      {children}
    </AppRuntimeContext.Provider>
  );

const renderAccept = (runtime: AppRuntimeValue) =>
  renderHook(() => useAcceptResumeTailorSuggestion(), { wrapper: wrap(runtime) });

const renderReject = (runtime: AppRuntimeValue) =>
  renderHook(() => useRejectResumeTailorSuggestion(), { wrapper: wrap(runtime) });

const fixtureAcceptDefault = acceptFixture.scenarios.default.response.body as ResumeVersion;
const fixtureRejectDefault = rejectFixture.scenarios.default.response.body as ResumeVersion;

describe("useAcceptResumeTailorSuggestion", () => {
  it("accepts a suggestion bodyless and forwards an Idempotency-Key", async () => {
    const acceptResumeTailorSuggestion = vi
      .fn()
      .mockResolvedValueOnce(fixtureAcceptDefault);
    const runtime = buildRuntime({ acceptResumeTailorSuggestion });
    const { result } = renderAccept(runtime);
    await act(async () => {
      await result.current.decide(VERSION_ID, SUG_ID);
    });
    expect(acceptResumeTailorSuggestion).toHaveBeenCalledTimes(1);
    const [versionArg, sugArg, optsArg] =
      acceptResumeTailorSuggestion.mock.calls[0]!;
    expect(versionArg).toBe(VERSION_ID);
    expect(sugArg).toBe(SUG_ID);
    expect(optsArg?.idempotencyKey).toMatch(/^v1\.\d+\..+/);
    // The generated method signature is bodyless (signature only accepts opts),
    // so there should be no third argument that looks like a body record.
    expect(acceptResumeTailorSuggestion.mock.calls[0]!.length).toBe(3);
    expect(typeof acceptResumeTailorSuggestion.mock.calls[0]![2]).toBe(
      "object",
    );
    expect(
      acceptResumeTailorSuggestion.mock.calls[0]![2].manualEditText,
    ).toBeUndefined();
  });

  it("replays the same Idempotency-Key when accept is retried for the same (versionId, suggestionId)", async () => {
    const acceptResumeTailorSuggestion = vi
      .fn()
      .mockResolvedValue(fixtureAcceptDefault);
    const runtime = buildRuntime({ acceptResumeTailorSuggestion });
    const { result } = renderAccept(runtime);
    await act(async () => {
      await result.current.decide(VERSION_ID, SUG_ID);
    });
    await act(async () => {
      await result.current.decide(VERSION_ID, SUG_ID);
    });
    const firstKey =
      acceptResumeTailorSuggestion.mock.calls[0]![2].idempotencyKey;
    const secondKey =
      acceptResumeTailorSuggestion.mock.calls[1]![2].idempotencyKey;
    expect(secondKey).toBe(firstKey);
  });

  it("rotates the Idempotency-Key when a different suggestion is decided", async () => {
    const acceptResumeTailorSuggestion = vi
      .fn()
      .mockResolvedValue(fixtureAcceptDefault);
    const runtime = buildRuntime({ acceptResumeTailorSuggestion });
    const { result } = renderAccept(runtime);
    await act(async () => {
      await result.current.decide(VERSION_ID, SUG_ID);
    });
    await act(async () => {
      await result.current.decide(VERSION_ID, "other-sug");
    });
    const firstKey =
      acceptResumeTailorSuggestion.mock.calls[0]![2].idempotencyKey;
    const secondKey =
      acceptResumeTailorSuggestion.mock.calls[1]![2].idempotencyKey;
    expect(secondKey).not.toBe(firstKey);
  });

  it("maps the already-decided-409 fixture envelope to SuggestionDecisionError(kind=already_decided)", async () => {
    const fixture = acceptFixture.scenarios["already-decided-409"].response;
    const error = new Error(
      `HTTP ${fixture.status} Conflict: ${JSON.stringify(fixture.body)}`,
    );
    const acceptResumeTailorSuggestion = vi.fn().mockRejectedValue(error);
    const runtime = buildRuntime({ acceptResumeTailorSuggestion });
    const { result } = renderAccept(runtime);
    await act(async () => {
      await expect(
        result.current.decide(VERSION_ID, SUG_ID),
      ).rejects.toBeInstanceOf(SuggestionDecisionError);
    });
    expect(result.current.lastError?.kind).toBe("already_decided");
  });

  it("maps a 422 envelope to validation and clears the cached IK", async () => {
    const acceptResumeTailorSuggestion = vi
      .fn()
      .mockRejectedValue(
        new Error(
          'HTTP 422 Unprocessable: {"error":{"code":"VALIDATION_FAILED"}}',
        ),
      );
    const runtime = buildRuntime({ acceptResumeTailorSuggestion });
    const { result } = renderAccept(runtime);
    await act(async () => {
      await expect(
        result.current.decide(VERSION_ID, SUG_ID),
      ).rejects.toBeInstanceOf(SuggestionDecisionError);
    });
    expect(result.current.lastError?.kind).toBe("validation");
    expect(result.current.peekIdempotencyKey(VERSION_ID, SUG_ID)).toBeNull();
  });

  it("maps 404 to cross_user without leaking ownership info", async () => {
    const acceptResumeTailorSuggestion = vi
      .fn()
      .mockRejectedValue(new Error("HTTP 404 Not Found: {}"));
    const runtime = buildRuntime({ acceptResumeTailorSuggestion });
    const { result } = renderAccept(runtime);
    await act(async () => {
      await expect(
        result.current.decide(VERSION_ID, SUG_ID),
      ).rejects.toBeInstanceOf(SuggestionDecisionError);
    });
    expect(result.current.lastError?.kind).toBe("cross_user");
  });

  it("captures the canonical accept fixture scenario keys", () => {
    expect(Object.keys(acceptFixture.scenarios).sort()).toEqual(
      ["already-decided-409", "default", "idempotency-replay"].sort(),
    );
  });
});

describe("useRejectResumeTailorSuggestion", () => {
  it("rejects a suggestion bodyless with an Idempotency-Key header", async () => {
    const rejectResumeTailorSuggestion = vi
      .fn()
      .mockResolvedValueOnce(fixtureRejectDefault);
    const runtime = buildRuntime({ rejectResumeTailorSuggestion });
    const { result } = renderReject(runtime);
    await act(async () => {
      await result.current.decide(VERSION_ID, SUG_ID);
    });
    expect(rejectResumeTailorSuggestion).toHaveBeenCalledTimes(1);
    expect(
      rejectResumeTailorSuggestion.mock.calls[0]![2].idempotencyKey,
    ).toMatch(/^v1\.\d+\..+/);
  });

  it("captures the canonical reject fixture scenario keys", () => {
    expect(Object.keys(rejectFixture.scenarios).sort()).toEqual(
      ["already-decided-409", "default", "idempotency-replay"].sort(),
    );
  });

  it("maps reject 409 SUGGESTION_ALREADY_DECIDED to kind=already_decided", async () => {
    const rejectResumeTailorSuggestion = vi
      .fn()
      .mockRejectedValue(
        new Error(
          'HTTP 409 Conflict: {"error":{"code":"VALIDATION_FAILED","details":{"reason":"SUGGESTION_ALREADY_DECIDED"}}}',
        ),
      );
    const runtime = buildRuntime({ rejectResumeTailorSuggestion });
    const { result } = renderReject(runtime);
    await act(async () => {
      await expect(
        result.current.decide(VERSION_ID, SUG_ID),
      ).rejects.toBeInstanceOf(SuggestionDecisionError);
    });
    expect(result.current.lastError?.kind).toBe("already_decided");
  });
});

describe("parseSuggestionDecisionError", () => {
  it("returns generic SuggestionDecisionError for non-Error inputs", () => {
    const parsed = parseSuggestionDecisionError("oops");
    expect(parsed).toBeInstanceOf(SuggestionDecisionError);
    expect(parsed.kind).toBe("generic");
  });
});
