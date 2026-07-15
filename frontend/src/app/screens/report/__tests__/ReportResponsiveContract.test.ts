import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const CSS_PATH = resolve(__dirname, "..", "..", "screens.css");

describe("Report ready responsive layout contract", () => {
  it("locks desktop 3/2/2/2/1 grids and mobile single-column order", () => {
    const source = readFileSync(CSS_PATH, "utf8");

    expect(source).toMatch(
      /\.ei-report-context-grid\s*\{[^}]*grid-template-columns:\s*repeat\(3, minmax\(0, 1fr\)\)/s,
    );
    for (const selector of ["ei-report-summary-grid", "ei-report-detail-grid"]) {
      expect(source).toMatch(
        new RegExp(`\\.${selector}\\s*\\{[^}]*grid-template-columns:\\s*repeat\\(2, minmax\\(0, 1fr\\)\\)`, "s"),
      );
    }
    expect(source).toMatch(
      /@media\s*\(max-width:\s*700px\)[\s\S]*?\.ei-report-context-grid,[\s\S]*?\.ei-report-summary-grid,[\s\S]*?\.ei-report-detail-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\)/,
    );
  });
});
