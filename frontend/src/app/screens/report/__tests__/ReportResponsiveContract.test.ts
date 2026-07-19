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

  it("rebuilds the ready dashboard as the supplied 1432px target composition", () => {
    const source = readFileSync(CSS_PATH, "utf8");
    const dashboard = readFileSync(resolve(__dirname, "../components/ReportDashboard.tsx"), "utf8");
    const header = readFileSync(resolve(__dirname, "../components/ReportHeader.tsx"), "utf8");
    const context = readFileSync(resolve(__dirname, "../components/ReportContextStrip.tsx"), "utf8");

    expect(source).toMatch(/\.ei-report-screen\s*\{[^}]*max-width:\s*1432px/s);
    expect(source).toMatch(/\.ei-report-context-grid\s*\{[^}]*gap:\s*0[^}]*background:[^}]*border:[^}]*border-radius:\s*12px/s);
    expect(source).toMatch(/\.ei-report-context-item:not\(:last-child\)::after\s*\{/s);
    expect(source).toMatch(/\.ei-report-detail-card-icon\s*\{/s);
    expect(source).toMatch(/\.ei-report-overall-icon\s*\{/s);
    expect(source).toMatch(/\.ei-report-metric\s*\{[^}]*border-radius:\s*12px/s);
    expect(source).toMatch(/\.ei-report-panel-card\s*\{[^}]*border-radius:\s*12px/s);
    expect(source).toMatch(/\.ei-report-overall\s*\{[^}]*grid-column:\s*1 \/ -1/s);
    expect(dashboard).toContain('className="ei-report-screen ei-fadein"');
    expect(dashboard).not.toContain("style={{");
    expect(context).not.toContain("style={{");
    expect(header).toContain('className="ei-report-header"');
    expect(header).toContain('data-testid="report-replay-icon"');
    expect(header).toContain('data-testid="report-next-icon"');
  });
});
