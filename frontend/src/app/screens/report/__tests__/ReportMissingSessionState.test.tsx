/**
 * @vitest-environment jsdom
 *
 * ReportMissingState reportId-only boundary gate.
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { ReportMissingState } from "../components/ReportMissingState";

describe("ReportMissingState", () => {
  it("renders the missing-report state and wires the back handler", () => {
    const onBack = vi.fn();
    render(<ReportMissingState onBackToWorkspace={onBack} />);
    expect(screen.getByTestId("report-missing-report-eyebrow")).toBeInTheDocument();
    expect(screen.getByTestId("report-missing-report-title")).toBeInTheDocument();
    expect(screen.getByTestId("report-missing-report-desc")).toBeInTheDocument();
    screen.getByTestId("report-missing-report-cta").click();
    expect(onBack).toHaveBeenCalledTimes(1);
  });
});
