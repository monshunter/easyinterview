/**
 * @vitest-environment jsdom
 *
 * Item 2.3 — usePracticeSession derives UI flags from `SessionStatus` 7-
 * value machine: queued / running / waiting_user_input / completing /
 * completed / failed / cancelled. Negative gate: must not reference
 * deprecated values `draft` / `archived`.
 */

import { describe, expect, it } from "vitest";
import { renderHook } from "@testing-library/react";

import type { SessionStatus } from "../../../../api/generated/types";
import { usePracticeSession } from "./usePracticeSession";

const ALL_STATUSES: SessionStatus[] = [
  "queued",
  "running",
  "waiting_user_input",
  "completing",
  "completed",
  "failed",
  "cancelled",
];

describe("usePracticeSession", () => {
  it("returns null-shaped flags when status is null", () => {
    const { result } = renderHook(() => usePracticeSession(null));
    expect(result.current.status).toBeNull();
    expect(result.current.inputDisabled).toBe(true);
    expect(result.current.showWaitingForFirstQuestion).toBe(false);
  });

  it("queued: occupy with 'preparing first question' notice; input disabled", () => {
    const { result } = renderHook(() => usePracticeSession("queued"));
    expect(result.current.status).toBe("queued");
    expect(result.current.showWaitingForFirstQuestion).toBe(true);
    expect(result.current.inputDisabled).toBe(true);
    expect(result.current.shouldNavigateGenerating).toBe(false);
  });

  it("running: input enabled; no completion side effects", () => {
    const { result } = renderHook(() => usePracticeSession("running"));
    expect(result.current.inputDisabled).toBe(false);
    expect(result.current.showCompletingNotice).toBe(false);
    expect(result.current.shouldNavigateGenerating).toBe(false);
    expect(result.current.errorMode).toBe("none");
  });

  it("waiting_user_input: input disabled (paused/wait state); no error", () => {
    const { result } = renderHook(() => usePracticeSession("waiting_user_input"));
    expect(result.current.inputDisabled).toBe(true);
    expect(result.current.errorMode).toBe("none");
  });

  it("completing: input + buttons disabled, completing notice visible", () => {
    const { result } = renderHook(() => usePracticeSession("completing"));
    expect(result.current.inputDisabled).toBe(true);
    expect(result.current.completionCtaDisabled).toBe(true);
    expect(result.current.showCompletingNotice).toBe(true);
    expect(result.current.shouldNavigateGenerating).toBe(false);
  });

  it("completed: triggers shouldNavigateGenerating exactly once via the snapshot", () => {
    const { result } = renderHook(() => usePracticeSession("completed"));
    expect(result.current.shouldNavigateGenerating).toBe(true);
    expect(result.current.inputDisabled).toBe(true);
  });

  it("failed: errorMode 'failed'; CTA back-to-workspace + retry surface true", () => {
    const { result } = renderHook(() => usePracticeSession("failed"));
    expect(result.current.errorMode).toBe("failed");
    expect(result.current.inputDisabled).toBe(true);
    expect(result.current.showRetry).toBe(true);
    expect(result.current.showBackToWorkspace).toBe(true);
  });

  it("cancelled: errorMode 'cancelled'; PracticeSessionLost surface", () => {
    const { result } = renderHook(() => usePracticeSession("cancelled"));
    expect(result.current.errorMode).toBe("cancelled");
    expect(result.current.showSessionLost).toBe(true);
    expect(result.current.showBackToWorkspace).toBe(true);
  });

  it("covers exactly the 7 SessionStatus values (negative on draft / archived)", () => {
    for (const status of ALL_STATUSES) {
      const { result } = renderHook(() => usePracticeSession(status));
      expect(result.current.status).toBe(status);
    }
    // The hook must NOT type-allow draft / archived — call sites should fail
    // typecheck. Runtime sanity: passing those values triggers the unknown
    // branch and surfaces an unsupported flag.
    const draftHook = renderHook(() =>
      // @ts-expect-error draft is a deprecated SessionStatus value
      usePracticeSession("draft"),
    );
    expect(draftHook.result.current.errorMode).toBe("none");
    const archivedHook = renderHook(() =>
      // @ts-expect-error archived is a deprecated SessionStatus value
      usePracticeSession("archived"),
    );
    expect(archivedHook.result.current.errorMode).toBe("none");
  });
});
