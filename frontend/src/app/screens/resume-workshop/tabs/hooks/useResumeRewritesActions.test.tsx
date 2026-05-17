// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import type { FC, ReactNode } from "react";

import type { ResumeVersion } from "../../../../../api/generated/types";
import { AppRuntimeContext } from "../../../../runtime/AppRuntimeProvider";
import type { AppRuntimeValue } from "../../../../runtime/AppRuntimeProvider";
import { SuggestionDecisionError } from "./useTailorSuggestionDecision";
import { useResumeRewritesActions } from "./useResumeRewritesActions";

const VERSION_ID = "0195f2d0-0001-7000-8000-000000000202";
const SUG_ID = "0195f2d0-0001-7000-8000-000000000302";

const baseVersion: ResumeVersion = {
  id: VERSION_ID,
  resumeAssetId: "01918fa0-0000-7000-8000-000000001000",
  parentVersionId: "0195f2d0-0001-7000-8000-000000000201",
  versionType: "targeted",
  targetJobId: "01918fa0-0000-7000-8000-000000002000",
  displayName: "Northstar Systems frontend target",
  seedStrategy: "copy_master",
  focusAngle: "platform",
  structuredProfile: {
    headline: "Senior frontend engineer",
    summary: "Owns platform.",
    sections: [],
    manualEdits: [],
  },
  matchScore: 0.84,
  promptVersion: "p",
  rubricVersion: "r",
  modelId: "m",
  provider: "fixture",
  provenance: {
    promptVersion: "p",
    rubricVersion: "r",
    modelId: "m",
    language: "zh-CN",
    featureFlag: "f",
    dataSourceVersion: "d",
  },
  suggestions: [],
  createdAt: "2026-05-12T08:24:00Z",
  updatedAt: "2026-05-12T08:24:00Z",
  deletedAt: null,
};

const buildRuntime = (
  overrides: Partial<AppRuntimeValue["client"]>,
): AppRuntimeValue =>
  ({
    client: {
      acceptResumeTailorSuggestion: vi.fn(),
      rejectResumeTailorSuggestion: vi.fn(),
      updateResumeVersion: vi.fn(),
      ...overrides,
    },
    runtime: { status: "ready", config: {} as never },
    auth: { status: "authenticated", user: { id: "u1" } as never },
    refreshAuth: vi.fn(),
  } as unknown as AppRuntimeValue);

const renderActions = (
  runtime: AppRuntimeValue,
  options: {
    version?: ResumeVersion;
    onVersionRefreshed?: () => void;
    now?: Date;
  } = {},
) => {
  const Wrapper: FC<{ children: ReactNode }> = ({ children }) => (
    <AppRuntimeContext.Provider value={runtime}>
      {children}
    </AppRuntimeContext.Provider>
  );
  return renderHook(
    () =>
      useResumeRewritesActions({
        version: options.version ?? baseVersion,
        onVersionRefreshed: options.onVersionRefreshed,
        nowProvider: options.now ? () => options.now! : undefined,
      }),
    { wrapper: Wrapper },
  );
};

describe("useResumeRewritesActions accept / reject", () => {
  it("onAccept calls the bodyless accept endpoint and triggers refetch", async () => {
    const acceptResumeTailorSuggestion = vi.fn().mockResolvedValue(baseVersion);
    const onVersionRefreshed = vi.fn();
    const runtime = buildRuntime({ acceptResumeTailorSuggestion });
    const { result } = renderActions(runtime, { onVersionRefreshed });
    await act(async () => {
      await result.current.onAccept(SUG_ID);
    });
    expect(acceptResumeTailorSuggestion).toHaveBeenCalledWith(
      VERSION_ID,
      SUG_ID,
      expect.objectContaining({ idempotencyKey: expect.stringMatching(/^v1\./) }),
    );
    expect(onVersionRefreshed).toHaveBeenCalledTimes(1);
  });

  it("onReject calls the bodyless reject endpoint and triggers refetch", async () => {
    const rejectResumeTailorSuggestion = vi.fn().mockResolvedValue(baseVersion);
    const onVersionRefreshed = vi.fn();
    const runtime = buildRuntime({ rejectResumeTailorSuggestion });
    const { result } = renderActions(runtime, { onVersionRefreshed });
    await act(async () => {
      await result.current.onReject(SUG_ID);
    });
    expect(rejectResumeTailorSuggestion).toHaveBeenCalledTimes(1);
    expect(onVersionRefreshed).toHaveBeenCalledTimes(1);
  });

  it("does not call updateResumeVersion when only accept / reject paths run (D-12: no structured_profile mutation)", async () => {
    const updateResumeVersion = vi.fn();
    const acceptResumeTailorSuggestion = vi.fn().mockResolvedValue(baseVersion);
    const runtime = buildRuntime({
      updateResumeVersion,
      acceptResumeTailorSuggestion,
    });
    const { result } = renderActions(runtime);
    await act(async () => {
      await result.current.onAccept(SUG_ID);
    });
    expect(updateResumeVersion).not.toHaveBeenCalled();
  });
});

describe("useResumeRewritesActions manual edit (update -> accept)", () => {
  const NOW = new Date("2026-05-18T10:00:00Z");

  it("first updates structuredProfile.manualEdits[] then calls bodyless accept", async () => {
    const updateResumeVersion = vi.fn().mockResolvedValue(baseVersion);
    const acceptResumeTailorSuggestion = vi.fn().mockResolvedValue(baseVersion);
    const onVersionRefreshed = vi.fn();
    const runtime = buildRuntime({
      updateResumeVersion,
      acceptResumeTailorSuggestion,
    });
    const { result } = renderActions(runtime, {
      onVersionRefreshed,
      now: NOW,
    });

    await act(async () => {
      await result.current.onSaveManualEdit(SUG_ID, "edited text");
    });

    // 1) update call carries the merged manualEdits[] entry
    expect(updateResumeVersion).toHaveBeenCalledTimes(1);
    const [versionArg, body, opts] = updateResumeVersion.mock.calls[0]!;
    expect(versionArg).toBe(VERSION_ID);
    expect(body.structuredProfile.manualEdits).toEqual([
      {
        suggestionId: SUG_ID,
        text: "edited text",
        savedAt: NOW.toISOString(),
      },
    ]);
    expect(opts?.idempotencyKey).toMatch(/^v1\./);

    // 2) accept fires second, bodyless, with its own IK
    expect(acceptResumeTailorSuggestion).toHaveBeenCalledTimes(1);
    const [aVersion, aSug, aOpts] =
      acceptResumeTailorSuggestion.mock.calls[0]!;
    expect(aVersion).toBe(VERSION_ID);
    expect(aSug).toBe(SUG_ID);
    expect(aOpts?.idempotencyKey).toMatch(/^v1\./);

    expect(result.current.manualPendingFor).toBeNull();
    expect(onVersionRefreshed).toHaveBeenCalledTimes(1);
  });

  it("merges into an existing manualEdits[] array by replacing the same suggestionId entry", async () => {
    const previousEntry = {
      suggestionId: SUG_ID,
      text: "earlier draft",
      savedAt: "2026-05-18T09:00:00Z",
    };
    const otherEntry = {
      suggestionId: "other-sug",
      text: "kept",
      savedAt: "2026-05-18T09:30:00Z",
    };
    const versionWithPrior: ResumeVersion = {
      ...baseVersion,
      structuredProfile: {
        ...baseVersion.structuredProfile,
        manualEdits: [previousEntry, otherEntry],
      },
    };
    const updateResumeVersion = vi.fn().mockResolvedValue(versionWithPrior);
    const acceptResumeTailorSuggestion = vi
      .fn()
      .mockResolvedValue(versionWithPrior);
    const runtime = buildRuntime({
      updateResumeVersion,
      acceptResumeTailorSuggestion,
    });
    const { result } = renderActions(runtime, {
      version: versionWithPrior,
      now: NOW,
    });
    await act(async () => {
      await result.current.onSaveManualEdit(SUG_ID, "new draft");
    });
    const [, body] = updateResumeVersion.mock.calls[0]!;
    expect(body.structuredProfile.manualEdits).toEqual([
      otherEntry,
      {
        suggestionId: SUG_ID,
        text: "new draft",
        savedAt: NOW.toISOString(),
      },
    ]);
  });

  it("when update succeeds but accept fails, surfaces manualPendingFor and does not double-write the edit on retry", async () => {
    const updateResumeVersion = vi.fn().mockResolvedValue(baseVersion);
    const acceptResumeTailorSuggestion = vi
      .fn()
      .mockRejectedValueOnce(
        new Error(
          'HTTP 409 Conflict: {"error":{"code":"VALIDATION_FAILED","details":{"reason":"SUGGESTION_ALREADY_DECIDED"}}}',
        ),
      )
      .mockResolvedValueOnce(baseVersion);
    const runtime = buildRuntime({
      updateResumeVersion,
      acceptResumeTailorSuggestion,
    });
    const { result } = renderActions(runtime, { now: NOW });

    await act(async () => {
      await expect(
        result.current.onSaveManualEdit(SUG_ID, "first attempt"),
      ).rejects.toBeInstanceOf(SuggestionDecisionError);
    });
    expect(result.current.manualPendingFor).toBe(SUG_ID);
    expect(updateResumeVersion).toHaveBeenCalledTimes(1);
    expect(acceptResumeTailorSuggestion).toHaveBeenCalledTimes(1);

    // Retry the manual-edit save with the same text. We expect update to run
    // (it does not have an IK-replay cache miss because payload+versionId is
    // unchanged so the same IK is reused) and accept to be invoked again.
    await act(async () => {
      await result.current.onSaveManualEdit(SUG_ID, "first attempt");
    });
    expect(acceptResumeTailorSuggestion).toHaveBeenCalledTimes(2);
    expect(result.current.manualPendingFor).toBeNull();
  });

  it("does NOT call accept when the manual edit update fails (422)", async () => {
    const updateResumeVersion = vi
      .fn()
      .mockRejectedValueOnce(
        new Error(
          'HTTP 422 Unprocessable: {"error":{"code":"VALIDATION_FAILED"}}',
        ),
      );
    const acceptResumeTailorSuggestion = vi.fn();
    const runtime = buildRuntime({
      updateResumeVersion,
      acceptResumeTailorSuggestion,
    });
    const { result } = renderActions(runtime, { now: NOW });

    await act(async () => {
      await expect(
        result.current.onSaveManualEdit(SUG_ID, "txt"),
      ).rejects.toBeTruthy();
    });
    expect(acceptResumeTailorSuggestion).not.toHaveBeenCalled();
    expect(result.current.manualPendingFor).toBeNull();
  });
});
