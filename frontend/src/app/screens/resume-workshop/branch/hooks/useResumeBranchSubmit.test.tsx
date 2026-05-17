// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import type { FC, ReactNode } from "react";

import type {
  BranchResumeVersionAccepted,
  ResumeVersion,
} from "../../../../../api/generated/types";
import { AppRuntimeContext } from "../../../../runtime/AppRuntimeProvider";
import type { AppRuntimeValue } from "../../../../runtime/AppRuntimeProvider";
import {
  useResumeBranchSubmit,
  BranchSubmitError,
  type BranchSubmitOutcome,
} from "./useResumeBranchSubmit";

import branchFixture from "../../../../../../../openapi/fixtures/Resumes/branchResumeVersion.json";

interface RuntimeOverrides {
  client?: Partial<AppRuntimeValue["client"]>;
}

const buildRuntime = (overrides: RuntimeOverrides): AppRuntimeValue => {
  const client = {
    branchResumeVersion: vi.fn(),
    ...overrides.client,
  } as unknown as AppRuntimeValue["client"];
  return {
    client,
    runtime: { status: "ready", config: {} as never },
    auth: { status: "authenticated", user: { id: "u1" } as never },
    refreshAuth: vi.fn(),
  };
};

const renderHookWithRuntime = (runtime: AppRuntimeValue) => {
  const Wrapper: FC<{ children: ReactNode }> = ({ children }) => (
    <AppRuntimeContext.Provider value={runtime}>
      {children}
    </AppRuntimeContext.Provider>
  );
  return renderHook(() => useResumeBranchSubmit(), { wrapper: Wrapper });
};

const PARENT_ID = "0195f2d0-0001-7000-8000-000000000201";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";

const baseDraft = {
  name: "v3 ByteDance",
  target: "ByteDance Frontend Platform",
  focus: "platform" as const,
  seed: "copy_master" as const,
};

const baseCtx = { parentVersionId: PARENT_ID, targetJobId: TARGET_JOB_ID };

const fixtureScenario = (name: keyof typeof branchFixture.scenarios) =>
  branchFixture.scenarios[name].response.body;

describe("useResumeBranchSubmit happy paths", () => {
  it("copy_master 201 returns kind=version + sends Idempotency-Key via opts.idempotencyKey", async () => {
    const branchResumeVersion = vi
      .fn()
      .mockResolvedValueOnce(fixtureScenario("default") as ResumeVersion);
    const runtime = buildRuntime({ client: { branchResumeVersion } });
    const { result } = renderHookWithRuntime(runtime);

    const captured: { outcome: BranchSubmitOutcome | null } = { outcome: null };
    await act(async () => {
      captured.outcome = await result.current.submit(baseDraft, baseCtx);
    });
    expect(captured.outcome).not.toBeNull();
    const outcome = captured.outcome;
    if (!outcome) throw new Error("outcome missing");
    expect(outcome.kind).toBe("version");
    if (outcome.kind !== "version") throw new Error("kind mismatch");
    expect(outcome.version.versionType).toBe("targeted");
    expect(branchResumeVersion).toHaveBeenCalledTimes(1);
    const [bodyArg, optsArg] = branchResumeVersion.mock.calls[0]!;
    expect(bodyArg).toMatchObject({
      parentVersionId: PARENT_ID,
      targetJobId: TARGET_JOB_ID,
      seedStrategy: "copy_master",
      displayName: "v3 ByteDance",
      focusAngle: "platform",
    });
    expect(optsArg?.idempotencyKey).toMatch(/^v1\.\d+\..+/);
    expect(result.current.peekIdempotencyKey()).toBe(optsArg.idempotencyKey);
  });

  it("blank seedStrategy 201 returns kind=version", async () => {
    const branchResumeVersion = vi
      .fn()
      .mockResolvedValueOnce(fixtureScenario("blank-sync") as ResumeVersion);
    const runtime = buildRuntime({ client: { branchResumeVersion } });
    const { result } = renderHookWithRuntime(runtime);

    const captured: { outcome: BranchSubmitOutcome | null } = { outcome: null };
    await act(async () => {
      captured.outcome = await result.current.submit(
        { ...baseDraft, seed: "blank" },
        baseCtx,
      );
    });
    const outcome = captured.outcome;
    if (!outcome) throw new Error("outcome missing");
    expect(outcome.kind).toBe("version");
    if (outcome.kind !== "version") throw new Error("kind mismatch");
    expect(outcome.version.seedStrategy).toBe("blank");
    const [bodyArg] = branchResumeVersion.mock.calls[0]!;
    expect(bodyArg.seedStrategy).toBe("blank");
  });

  it("ai_select 202 returns kind=accepted with job + version", async () => {
    const branchResumeVersion = vi
      .fn()
      .mockResolvedValueOnce(
        fixtureScenario("ai-select-202-with-job") as BranchResumeVersionAccepted,
      );
    const runtime = buildRuntime({ client: { branchResumeVersion } });
    const { result } = renderHookWithRuntime(runtime);

    const captured: { outcome: BranchSubmitOutcome | null } = { outcome: null };
    await act(async () => {
      captured.outcome = await result.current.submit(
        { ...baseDraft, seed: "ai_select" },
        baseCtx,
      );
    });
    const outcome = captured.outcome;
    if (!outcome) throw new Error("outcome missing");
    expect(outcome.kind).toBe("accepted");
    if (outcome.kind !== "accepted") throw new Error("kind mismatch");
    expect(outcome.accepted.resumeVersionId).toBeTruthy();
    expect(outcome.accepted.job.jobType).toBe("resume_tailor");
    expect(outcome.accepted.job.status).toBe("queued");
  });

  it("replays the same Idempotency-Key when the draft fingerprint is unchanged", async () => {
    const branchResumeVersion = vi
      .fn()
      .mockResolvedValue(fixtureScenario("default") as ResumeVersion);
    const runtime = buildRuntime({ client: { branchResumeVersion } });
    const { result } = renderHookWithRuntime(runtime);

    await act(async () => {
      await result.current.submit(baseDraft, baseCtx);
    });
    const firstKey = branchResumeVersion.mock.calls[0]![1].idempotencyKey;
    await act(async () => {
      await result.current.submit(baseDraft, baseCtx);
    });
    const secondKey = branchResumeVersion.mock.calls[1]![1].idempotencyKey;
    expect(secondKey).toBe(firstKey);
    expect(result.current.peekIdempotencyKey()).toBe(firstKey);
  });

  it("rotates the Idempotency-Key when the user changes a form field between retries", async () => {
    const branchResumeVersion = vi
      .fn()
      .mockResolvedValue(fixtureScenario("default") as ResumeVersion);
    const runtime = buildRuntime({ client: { branchResumeVersion } });
    const { result } = renderHookWithRuntime(runtime);

    await act(async () => {
      await result.current.submit(baseDraft, baseCtx);
    });
    const firstKey = branchResumeVersion.mock.calls[0]![1].idempotencyKey;
    await act(async () => {
      await result.current.submit(
        { ...baseDraft, focus: "leadership" },
        baseCtx,
      );
    });
    const secondKey = branchResumeVersion.mock.calls[1]![1].idempotencyKey;
    expect(secondKey).not.toBe(firstKey);
  });
});

describe("useResumeBranchSubmit error mapping", () => {
  const failingClient = (error: Error) => ({
    branchResumeVersion: vi.fn().mockRejectedValue(error),
  });

  it("maps 422 VALIDATION_FAILED to BranchSubmitError(kind=validation) and resets the IK cache", async () => {
    const err = new Error(
      'HTTP 422 Unprocessable: {"error":{"code":"VALIDATION_FAILED","message":"x","details":{"field":"displayName"}}}',
    );
    const runtime = buildRuntime({ client: failingClient(err) });
    const { result } = renderHookWithRuntime(runtime);

    await act(async () => {
      await expect(result.current.submit(baseDraft, baseCtx)).rejects.toBeInstanceOf(
        BranchSubmitError,
      );
    });
    expect(result.current.lastError?.kind).toBe("validation");
    expect(result.current.lastError?.field).toBe("displayName");
    expect(result.current.peekIdempotencyKey()).toBeNull();
  });

  it("maps 404 with details.reason=PARENT_NOT_FOUND to BranchSubmitError(kind=parent_missing)", async () => {
    const err = new Error(
      'HTTP 404 Not Found: {"error":{"code":"NOT_FOUND","details":{"reason":"PARENT_NOT_FOUND"}}}',
    );
    const runtime = buildRuntime({ client: failingClient(err) });
    const { result } = renderHookWithRuntime(runtime);

    await act(async () => {
      await expect(result.current.submit(baseDraft, baseCtx)).rejects.toBeInstanceOf(
        BranchSubmitError,
      );
    });
    expect(result.current.lastError?.kind).toBe("parent_missing");
  });

  it("maps a generic 404 to BranchSubmitError(kind=cross_user)", async () => {
    const err = new Error("HTTP 404 Not Found: {}");
    const runtime = buildRuntime({ client: failingClient(err) });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await expect(result.current.submit(baseDraft, baseCtx)).rejects.toBeInstanceOf(
        BranchSubmitError,
      );
    });
    expect(result.current.lastError?.kind).toBe("cross_user");
  });

  it("maps 409 to BranchSubmitError(kind=idempotency_conflict) without clearing the IK cache", async () => {
    const err = new Error(
      'HTTP 409 Conflict: {"error":{"code":"IDEMPOTENCY_KEY_CONFLICT"}}',
    );
    const runtime = buildRuntime({ client: failingClient(err) });
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await expect(result.current.submit(baseDraft, baseCtx)).rejects.toBeInstanceOf(
        BranchSubmitError,
      );
    });
    expect(result.current.lastError?.kind).toBe("idempotency_conflict");
    expect(result.current.peekIdempotencyKey()).not.toBeNull();
  });

  it("treats a missing runtime client as a generic failure (no submission)", async () => {
    const runtime = {
      client: null,
      runtime: { status: "ready" },
      auth: { status: "authenticated" },
      refreshAuth: vi.fn(),
    } as unknown as AppRuntimeValue;
    const { result } = renderHookWithRuntime(runtime);
    await act(async () => {
      await expect(result.current.submit(baseDraft, baseCtx)).rejects.toBeInstanceOf(
        BranchSubmitError,
      );
    });
    expect(result.current.lastError?.kind).toBe("generic");
  });
});

describe("useResumeBranchSubmit fixture parity", () => {
  it("idempotent-replay fixture echoes the same Idempotency-Key on a second submit", async () => {
    const branchResumeVersion = vi
      .fn()
      .mockResolvedValueOnce(fixtureScenario("default") as ResumeVersion)
      .mockResolvedValueOnce(
        fixtureScenario("idempotent-replay") as ResumeVersion,
      );
    const runtime = buildRuntime({ client: { branchResumeVersion } });
    const { result } = renderHookWithRuntime(runtime);

    await act(async () => {
      await result.current.submit(baseDraft, baseCtx);
    });
    await act(async () => {
      await result.current.submit(baseDraft, baseCtx);
    });
    const firstKey = branchResumeVersion.mock.calls[0]![1].idempotencyKey;
    const secondKey = branchResumeVersion.mock.calls[1]![1].idempotencyKey;
    expect(secondKey).toBe(firstKey);
    // Both responses must have the same canonical resumeVersionId per
    // fixture default vs. idempotent-replay scenarios.
    const firstId = (fixtureScenario("default") as ResumeVersion).id;
    const replayId = (fixtureScenario("idempotent-replay") as ResumeVersion).id;
    expect(replayId).toBe(firstId);
  });

  it("captures the full fixture scenario coverage matrix for branchResumeVersion", () => {
    const expected = [
      "default",
      "copy-master-sync",
      "blank-sync",
      "ai-select-202-with-job",
      "validation-error-422",
      "idempotent-replay",
    ].sort();
    expect(Object.keys(branchFixture.scenarios).sort()).toEqual(expected);
  });
});
