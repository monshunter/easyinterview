import { useMemo } from "react";

import type { SessionStatus } from "../../../../api/generated/types";

export type ErrorMode = "none" | "failed" | "cancelled";

export interface UsePracticeSessionResult {
  status: SessionStatus | null;
  inputDisabled: boolean;
  showWaitingForFirstQuestion: boolean;
  showCompletingNotice: boolean;
  completionCtaDisabled: boolean;
  shouldNavigateGenerating: boolean;
  showSessionLost: boolean;
  showBackToWorkspace: boolean;
  showRetry: boolean;
  errorMode: ErrorMode;
}

/**
 * Item 2.3 — derive PracticeScreen UI flags from `SessionStatus` 7 values.
 *
 *  - queued                → 占位等待首题 + input disabled
 *  - running               → 主交互；input enabled
 *  - waiting_user_input    → 输入禁用 + 暂停 / 等待状态
 *  - completing            → 输入禁用 + 「正在生成报告…」notice
 *  - completed             → 自动 nav generating（caller 防抖）
 *  - failed                → ErrorState + retry / back-to-workspace
 *  - cancelled             → PracticeSessionLost + back-to-workspace
 *
 * Negative gate: this hook does NOT recognise the deprecated `draft` /
 * `archived` values. Their union member is removed from `SessionStatus`,
 * so call sites should fail typecheck rather than reach runtime.
 */
export function usePracticeSession(
  status: SessionStatus | null,
): UsePracticeSessionResult {
  return useMemo<UsePracticeSessionResult>(() => {
    switch (status) {
      case "queued":
        return base({
          status,
          inputDisabled: true,
          showWaitingForFirstQuestion: true,
        });
      case "running":
        return base({
          status,
          inputDisabled: false,
        });
      case "waiting_user_input":
        return base({
          status,
          inputDisabled: true,
        });
      case "completing":
        return base({
          status,
          inputDisabled: true,
          showCompletingNotice: true,
          completionCtaDisabled: true,
        });
      case "completed":
        return base({
          status,
          inputDisabled: true,
          shouldNavigateGenerating: true,
        });
      case "failed":
        return base({
          status,
          inputDisabled: true,
          showRetry: true,
          showBackToWorkspace: true,
          errorMode: "failed",
        });
      case "cancelled":
        return base({
          status,
          inputDisabled: true,
          showSessionLost: true,
          showBackToWorkspace: true,
          errorMode: "cancelled",
        });
      default:
        // Unknown / null statuses fall through to a safe default that locks
        // the input. This protects against deprecated `draft` / `archived`
        // values reaching runtime through an untyped boundary.
        return base({
          status: status ?? null,
          inputDisabled: true,
        });
    }
  }, [status]);
}

function base(
  partial: Partial<UsePracticeSessionResult> & { status: SessionStatus | null },
): UsePracticeSessionResult {
  return {
    status: partial.status,
    inputDisabled: partial.inputDisabled ?? false,
    showWaitingForFirstQuestion: partial.showWaitingForFirstQuestion ?? false,
    showCompletingNotice: partial.showCompletingNotice ?? false,
    completionCtaDisabled: partial.completionCtaDisabled ?? false,
    shouldNavigateGenerating: partial.shouldNavigateGenerating ?? false,
    showSessionLost: partial.showSessionLost ?? false,
    showBackToWorkspace: partial.showBackToWorkspace ?? false,
    showRetry: partial.showRetry ?? false,
    errorMode: partial.errorMode ?? "none",
  };
}
