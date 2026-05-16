/**
 * @vitest-environment jsdom
 *
 * Phase 2.7 — ReportMissingSessionState focused gate.
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { ReportMissingSessionState } from "../components/ReportMissingSessionState";

describe("ReportMissingSessionState", () => {
  it("renders the eyebrow + title + desc + back CTA and wires the handler (TestReportMissingSessionNavigatesWorkspace)", () => {
    const onBack = vi.fn();
    render(<ReportMissingSessionState onBackToWorkspace={onBack} />);
    expect(screen.getByTestId("report-missing-session-eyebrow")).toBeInTheDocument();
    expect(screen.getByTestId("report-missing-session-title")).toBeInTheDocument();
    expect(screen.getByTestId("report-missing-session-desc")).toBeInTheDocument();
    screen.getByTestId("report-missing-session-cta").click();
    expect(onBack).toHaveBeenCalledTimes(1);
  });
});
