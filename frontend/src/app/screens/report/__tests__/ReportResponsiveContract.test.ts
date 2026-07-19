import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const CSS_PATH = resolve(__dirname, "..", "..", "screens.css");

describe("Report ready responsive layout contract", () => {
  it("locks desktop 4/2/2/2/1 grids, equal-height detail cards, and mobile single-column order", () => {
    const source = readFileSync(CSS_PATH, "utf8");

    expect(source).toMatch(
      /\.ei-report-context-grid\s*\{[^}]*grid-template-columns:\s*repeat\(4, minmax\(0, 1fr\)\)/s,
    );
    for (const selector of ["ei-report-summary-grid", "ei-report-detail-grid"]) {
      expect(source).toMatch(
        new RegExp(`\\.${selector}\\s*\\{[^}]*grid-template-columns:\\s*repeat\\(2, minmax\\(0, 1fr\\)\\)`, "s"),
      );
    }
    expect(source).toMatch(
      /@media\s*\(max-width:\s*700px\)[\s\S]*?\.ei-report-context-grid,[\s\S]*?\.ei-report-summary-grid,[\s\S]*?\.ei-report-detail-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\)/,
    );
    expect(source).toMatch(/\.ei-report-panel\s*\{[^}]*display:\s*flex[^}]*min-height:\s*100%/s);
    expect(source).toMatch(/\.ei-report-panel-card\s*\{[^}]*flex:\s*1[^}]*box-sizing:\s*border-box/s);
    expect(source).toMatch(
      /@media\s*\(max-width:\s*700px\)[\s\S]*?\.ei-report-panel\s*\{[^}]*min-height:\s*auto/,
    );
    expect(source).toMatch(
      /@media\s*\(max-width:\s*700px\)[\s\S]*?\.ei-report-header-copy\s*\{[^}]*flex:\s*0 0 auto/,
    );
  });

  it("aligns the ready dashboard to the supplied 1336px reference canvas", () => {
    const source = readFileSync(CSS_PATH, "utf8");
    const dashboard = readFileSync(resolve(__dirname, "../components/ReportDashboard.tsx"), "utf8");
    const header = readFileSync(resolve(__dirname, "../components/ReportHeader.tsx"), "utf8");

    expect(source).toMatch(/\.ei-report-screen\s*\{[^}]*max-width:\s*1336px/s);
    expect(source).toMatch(/\.ei-report-header\s*\{[^}]*margin-bottom:\s*24px/s);
    expect(source).toMatch(/\.ei-report-metric\s*\{[^}]*border-radius:\s*12px/s);
    expect(source).toMatch(/\.ei-report-panel-card\s*\{[^}]*border-radius:\s*12px/s);
    expect(source).toMatch(/\.ei-report-overall\s*\{[^}]*grid-column:\s*1 \/ -1/s);
    expect(dashboard).toContain('className="ei-report-screen ei-fadein"');
    expect(header).toContain('className="ei-report-header"');
  });
});
