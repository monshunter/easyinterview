import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const CSS_PATH = resolve(__dirname, "..", "screens.css");

function css(): string {
  return readFileSync(CSS_PATH, "utf8");
}

describe("Resume Workshop source-level CSS parity", () => {
  it("defines the list/detail layout selectors used by the implemented DOM", () => {
    const source = css();
    for (const selector of [
      ".ei-screen-shell[data-testid=\"resume-workshop-screen\"]",
      ".ei-resume-workshop-list",
      ".ei-resume-workshop-card-grid",
      ".ei-resume-workshop-card",
      ".ei-resume-workshop-card-delete",
      ".ei-resume-workshop-card-open",
      ".ei-resume-detail",
      ".ei-resume-detail-preview",
      ".ei-resume-detail-markdown-page",
      ".ei-resume-detail-pdf-stack",
      ".ei-resume-detail-pdf-page",
      ".ei-resume-detail-pdf-canvas",
      ".ei-resume-detail-parse-state",
    ]) {
      expect(source, `${selector} missing from screens.css`).toContain(selector);
    }
  });

  it("locks key values transcribed from the resume workshop prototype source", () => {
    const source = css();
    expect(source).toContain("max-width: 1320px");
    expect(source).toContain("padding: 40px 48px 96px");
    expect(source).toContain(
      "grid-template-columns: repeat(auto-fill, minmax(300px, 360px))",
    );
    expect(source).toContain("justify-content: start");
    expect(source).toContain("border-radius: 2px");
    expect(source).toContain("grid-template-columns: minmax(0, 860px)");
    expect(source).toContain("min-height: 720px");
    expect(source).toContain("box-shadow: 0 18px 50px rgba(30, 22, 15, 0.10)");
    expect(source).toContain("background: #f6f3ee");
    expect(source).toContain("padding: 28px");
    expect(source).toContain("padding: 44px 56px");
    expect(source).toContain("gap: 22px");
    expect(source).toContain("width: min(100%, 720px)");
    expect(source).not.toContain(".ei-resume-detail-preview-card--pdf");
  });

  it("uses a fixed-width desktop grid, a single mobile column and no table selectors", () => {
    const source = css();
    expect(source).toMatch(
      /\.ei-resume-workshop-card-grid\s*\{[^}]*display:\s*grid[^}]*grid-template-columns:\s*repeat\(auto-fill, minmax\(300px, 360px\)\)[^}]*justify-content:\s*start/s,
    );
    expect(source).toMatch(
      /@media\s*\(max-width:\s*700px\)[\s\S]*?\.ei-resume-workshop-card-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\)/,
    );
    for (const selector of [
      ".ei-resume-workshop-table",
      ".ei-resume-workshop-table-head",
      ".ei-resume-workshop-table-row",
    ]) {
      expect(source, `${selector} should be removed`).not.toContain(selector);
    }
  });

  it("does not keep out-of-scope tree/version/branch styling after D-20", () => {
    const source = css();
    for (const selector of [
      ".ei-resume-workshop-stats",
      ".ei-resume-workshop-view-switcher",
      ".ei-resume-workshop-selected-tree",
      ".ei-resume-workshop-tree",
      ".ei-resume-workshop-version-row",
      ".ei-resume-workshop-flat",
      ".ei-resume-detail-branch-graph",
      ".ei-resume-branch-flow",
    ]) {
      expect(source, `${selector} should remain absent`).not.toContain(selector);
    }
  });

  it("does not keep detail styles without a current DOM or prototype consumer", () => {
    const source = css();
    for (const selector of [
      ".ei-resume-detail-breadcrumb",
      ".ei-resume-detail-preview-actions",
      ".ei-resume-detail-preview-section",
      ".ei-resume-detail-preview-skills",
      ".ei-resume-detail-modal-overlay",
      ".ei-resume-detail-modal-header",
      ".ei-resume-detail-modal-desc",
      ".ei-resume-detail-modal-content",
      ".ei-resume-detail-modal",
    ]) {
      expect(source, `${selector} should remain absent`).not.toContain(selector);
    }
  });

  it("keeps one effective detail-back rule and no grid declaration on the flex preview", () => {
    const source = css();
    const backRules = source.match(/\.ei-resume-detail-back\s*\{[^}]*\}/g) ?? [];
    expect(backRules).toHaveLength(1);
    expect(backRules[0]).toMatch(/display:\s*inline-flex/);
    expect(backRules[0]).toMatch(/padding:\s*0/);
    expect(backRules[0]).toMatch(/border:\s*0/);
    expect(backRules[0]).toMatch(/border-radius:\s*2px/);
    expect(backRules[0]).toMatch(/font-family:\s*var\(--ei-font-sans\)/);
    expect(backRules[0]).toMatch(/font-size:\s*13px/);
    expect(source).not.toMatch(
      /\.ei-resume-detail-preview\s*\{[^}]*grid-template-columns/,
    );
  });
});
