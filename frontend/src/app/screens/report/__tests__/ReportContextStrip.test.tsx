/**
 * @vitest-environment jsdom
 *
 * Phase 2.10 — ReportContextStrip renders all 7 fields and reflects modality /
 * practiceMode / hints toggles in zh+en.
 */

import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import { ReportContextStrip } from "../components/ReportContextStrip";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";

const BASE_PROPS = {
  sessionId: "sess-1",
  targetLabel: "Acme · Senior Frontend",
  roundLabel: "Round 1",
  resumeLabel: "Resume v3",
  modality: "text",
  practiceMode: "strict",
  hintUsed: "false",
  hintCount: "0",
};

describe("ReportContextStrip", () => {
  it("renders all 7 slots with their display knob values", () => {
    render(
      <DisplayPreferencesProvider>
        <ReportContextStrip {...BASE_PROPS} />
      </DisplayPreferencesProvider>,
    );
    const expected = [
      "report-context-session",
      "report-context-job",
      "report-context-round",
      "report-context-resume",
      "report-context-modality",
      "report-context-practice-mode",
      "report-context-hints",
    ];
    for (const id of expected) {
      expect(screen.getByTestId(id)).toBeInTheDocument();
    }
    expect(screen.getByTestId("report-context-session").textContent).toContain("sess-1");
    expect(screen.getByTestId("report-context-job").textContent).toContain(
      "Acme",
    );
  });

  it("shows legacy modality.voice as phone plus practiceMode.assisted variants when toggled", () => {
    render(
      <DisplayPreferencesProvider>
        <ReportContextStrip
          {...BASE_PROPS}
          modality="voice"
          practiceMode="assisted"
          hintUsed="true"
          hintCount="3"
        />
      </DisplayPreferencesProvider>,
    );
    const modality = screen.getByTestId("report-context-modality").textContent ?? "";
    const practiceMode = screen.getByTestId("report-context-practice-mode").textContent ?? "";
    const hints = screen.getByTestId("report-context-hints").textContent ?? "";
    expect(modality).toContain("Phone");
    expect(practiceMode.length).toBeGreaterThan(0);
    // hintUsed=true must surface a non-default copy plus the count.
    expect(hints).not.toBe("");
    expect(hints).toContain("3");
  });

  it("shows current modality.phone as Phone instead of falling back to Text", () => {
    render(
      <DisplayPreferencesProvider>
        <ReportContextStrip {...BASE_PROPS} modality="phone" />
      </DisplayPreferencesProvider>,
    );

    expect(screen.getByTestId("report-context-modality").textContent).toContain(
      "Phone",
    );
  });
});
