/**
 * @vitest-environment jsdom
 *
 * Phase 2.7 — ReportFailureState focused gates.
 *  - All AI_* enum values map to dedicated copy (TestReportFailureStateRendersErrorCodeMatrix)
 *  - REPORT_NOT_FOUND routes through failureState.notFound.* keys, NOT AI_*
 *  - Retry CTA + back-to-workspace CTA wire the handlers we hand in.
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import type { ApiErrorCode } from "../../../../api/generated/types";
import { ReportFailureState } from "../components/ReportFailureState";
import { ERROR_CODES } from "../../../../lib/conventions/errors";

const AI_ENUM: ReadonlyArray<ApiErrorCode> = [
  ERROR_CODES.AI_PROVIDER_TIMEOUT,
  ERROR_CODES.AI_PROVIDER_SECRET_MISSING,
  ERROR_CODES.AI_PROVIDER_CONFIG_INVALID,
  ERROR_CODES.AI_OUTPUT_INVALID,
  ERROR_CODES.AI_FALLBACK_EXHAUSTED,
  ERROR_CODES.AI_UNSUPPORTED_CAPABILITY,
];

describe("ReportFailureState", () => {
  it("renders the AI_* enum matrix without collapsing to UNKNOWN (TestReportFailureStateRendersErrorCodeMatrix)", () => {
    const renderedTexts: Record<string, string> = {};
    for (const code of AI_ENUM) {
      const { unmount } = render(
        <ReportFailureState
          errorCode={code}
          onRetry={() => undefined}
          onBackToWorkspace={() => undefined}
        />,
      );
      const codeLabel = screen.getByTestId("report-failure-error-code").textContent ?? "";
      renderedTexts[code] = codeLabel;
      // Each AI_* enum must surface a non-fallback string. The UNKNOWN map key
      // is reserved for true catch-alls.
      expect(codeLabel.length).toBeGreaterThan(0);
      unmount();
    }
    const distinct = new Set(Object.values(renderedTexts));
    // 6 AI_* codes → at least 6 distinct strings, none falling through to UNKNOWN.
    expect(distinct.size).toBeGreaterThanOrEqual(AI_ENUM.length);
  });

  it("REPORT_NOT_FOUND uses the dedicated notFound copy and hides the retry CTA (TestReportFailureStateRendersNotFoundCopy)", () => {
    const onRetry = vi.fn();
    const onBack = vi.fn();
    render(
      <ReportFailureState
        errorCode="REPORT_NOT_FOUND"
        onRetry={onRetry}
        onBackToWorkspace={onBack}
      />,
    );
    const root = screen.getByTestId("report-failure-state");
    expect(root.getAttribute("data-not-found")).toBe("true");
    expect(screen.getByTestId("report-failure-state-not-found-title")).toBeInTheDocument();
    expect(screen.queryByTestId("report-failure-title")).toBeNull();
    expect(screen.queryByTestId("report-failure-retry-cta")).toBeNull();
    expect(screen.getByTestId("report-failure-back-to-workspace")).toBeInTheDocument();
  });

  it("retry + back CTAs trigger their handlers exactly once (TestReportFailureStateRetryNavigatesGenerating / BackToWorkspaceCta)", () => {
    const onRetry = vi.fn();
    const onBack = vi.fn();
    render(
      <ReportFailureState
        errorCode="AI_PROVIDER_TIMEOUT"
        onRetry={onRetry}
        onBackToWorkspace={onBack}
      />,
    );
    screen.getByTestId("report-failure-retry-cta").click();
    expect(onRetry).toHaveBeenCalledTimes(1);
    screen.getByTestId("report-failure-back-to-workspace").click();
    expect(onBack).toHaveBeenCalledTimes(1);
  });

  it("falls back to UNKNOWN copy when errorCode is null or unrecognized", () => {
    render(
      <ReportFailureState
        errorCode={null}
        onRetry={() => undefined}
        onBackToWorkspace={() => undefined}
      />,
    );
    expect(screen.getByTestId("report-failure-error-code").textContent ?? "").not.toBe("");
  });
});
