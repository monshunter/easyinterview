/**
 * @vitest-environment jsdom
 *
 * Phase 2.7 — ReportFailureState focused gates.
 *  - Internal provider/config error details never render in visible or accessible UI.
 *  - REPORT_NOT_FOUND routes through dedicated failureState.notFound.* copy.
 *  - Terminal report failures do not offer a fake regeneration action.
 *  - Only recoverable transport errors expose a reload action.
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
  it("hides the AI_* enum matrix behind one user-safe terminal state", () => {
    for (const code of AI_ENUM) {
      const { unmount } = render(
        <ReportFailureState
          errorCode={code}
          onRetry={() => undefined}
          onBack={() => undefined}
          backDestination="workspace"
        />,
      );
      expect(screen.queryByTestId("report-failure-error-code")).not.toBeInTheDocument();
      expect(document.body).not.toHaveTextContent(code);
      unmount();
    }
  });

  it("REPORT_NOT_FOUND uses the dedicated notFound copy and hides the retry CTA (TestReportFailureStateRendersNotFoundCopy)", () => {
    const onRetry = vi.fn();
    const onBack = vi.fn();
    render(
      <ReportFailureState
        errorCode="REPORT_NOT_FOUND"
        onRetry={onRetry}
        onBack={onBack}
        backDestination="workspace"
      />,
    );
    const root = screen.getByTestId("report-failure-state");
    expect(root.getAttribute("data-not-found")).toBe("true");
    expect(screen.getByTestId("report-failure-state-not-found-title")).toBeInTheDocument();
    expect(screen.queryByTestId("report-failure-title")).toBeNull();
    expect(screen.queryByTestId("report-failure-retry-cta")).toBeNull();
    expect(screen.getByTestId("report-failure-back-to-workspace")).toBeInTheDocument();
  });

  it("terminal report failures hide reload and keep the back action", () => {
    const onRetry = vi.fn();
    const onBack = vi.fn();
    render(
      <ReportFailureState
        errorCode="AI_PROVIDER_TIMEOUT"
        onRetry={onRetry}
        onBack={onBack}
        backDestination="reports"
      />,
    );
    expect(screen.queryByTestId("report-failure-retry-cta")).not.toBeInTheDocument();
    const back = screen.getByTestId("report-failure-back-to-workspace");
    expect(back).toHaveAttribute("data-back-destination", "reports");
    expect(back).toHaveTextContent("Back");
    back.click();
    expect(onBack).toHaveBeenCalledTimes(1);
  });

  it("recoverable transport failure reloads exactly once", () => {
    const onRetry = vi.fn();
    render(
      <ReportFailureState
        errorCode={null}
        recoverable
        onRetry={onRetry}
        onBack={() => undefined}
        backDestination="workspace"
      />,
    );
    screen.getByTestId("report-failure-retry-cta").click();
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it("uses generic copy without exposing an unknown error-code panel", () => {
    render(
      <ReportFailureState
        errorCode={null}
        onRetry={() => undefined}
        onBack={() => undefined}
        backDestination="workspace"
      />,
    );
    expect(screen.queryByTestId("report-failure-error-code")).not.toBeInTheDocument();
  });
});
