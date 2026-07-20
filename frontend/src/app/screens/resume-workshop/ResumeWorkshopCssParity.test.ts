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

  it("locks key values transcribed from the supplied resume-list reference", () => {
    const source = css();
    expect(source).toMatch(
      /\.ei-screen-shell\[data-testid="resume-workshop-screen"\]\[data-flow="list"\]\s*\{[^}]*max-width:\s*1408px[^}]*padding:\s*50px 0 96px/s,
    );
    expect(source).toMatch(
      /\.ei-resume-workshop-card-grid\s*\{[^}]*grid-template-columns:\s*repeat\(2, minmax\(0, 1fr\)\)[^}]*gap:\s*28px/s,
    );
    expect(source).toMatch(
      /\.ei-resume-workshop-card\s*\{[^}]*min-height:\s*384px[^}]*padding:\s*28px[^}]*border-radius:\s*12px/s,
    );
    expect(source).toMatch(
      /\.ei-resume-workshop-create\s*\{[^}]*height:\s*64px/s,
    );
    expect(source).toMatch(
      /\.ei-resume-workshop-card-open\s*\{[^}]*min-width:\s*106px[^}]*height:\s*56px/s,
    );
    expect(source).not.toContain("repeat(auto-fill, minmax(300px, 360px))");
    expect(source).not.toContain("grid-template-columns: minmax(0, 918px)");
    expect(source).not.toContain(".ei-resume-workshop-lang-tag");
    expect(source).toMatch(
      /data-flow="detail"\]\s*\{[^}]*max-width:\s*1512px/s,
    );
    expect(source).toContain("padding: 52px 68px");
    expect(source).toContain("gap: 22px");
    const markdownPageRules =
      source.match(/\.ei-resume-detail-markdown-page\s*\{[^}]*\}/gs) ?? [];
    expect(markdownPageRules.length).toBeGreaterThan(0);
    expect(markdownPageRules[0]).toMatch(/width:\s*min\(100%, 794px\)/);
    for (const rule of markdownPageRules) {
      expect(rule).not.toMatch(/aspect-ratio:/);
      expect(rule).not.toMatch(/(?:min-)?height:/);
    }
    expect(source).toMatch(
      /\.ei-resume-detail-pdf-page\s*\{[^}]*width:\s*min\(100%, 794px\)[^}]*aspect-ratio:\s*210 \/ 297/s,
    );
    expect(source).not.toContain("width: min(100%, 1150px)");
    expect(source).toMatch(
      /\.ei-resume-detail-preview\s*>\s*article\s*\{[^}]*width:\s*100%/s,
    );
    expect(source).not.toContain("width: min(100%, 860px)");
    expect(source).not.toContain(".ei-resume-detail-preview-card");
  });

  it("uses two equal desktop card columns, a full-width mobile column and no table selectors", () => {
    const source = css();
    expect(source).toMatch(
      /\.ei-resume-workshop-card-grid\s*\{[^}]*display:\s*grid[^}]*grid-template-columns:\s*repeat\(2, minmax\(0, 1fr\)\)/s,
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

  it("uses the active theme color for the card Open action", () => {
    const source = css();
    expect(source).toMatch(
      /\.ei-resume-workshop-card-open\s*\{[^}]*color:\s*#fff[^}]*background:\s*var\(--ei-color-accent\)[^}]*border:\s*1px solid var\(--ei-color-accent\)/s,
    );
  });

  it("keeps the parse waiting animation free of geometry-changing transforms", () => {
    const source = css();
    const iconRule = source.match(
      /\.ei-resume-detail-parse-icon\s*\{[^}]*\}/s,
    )?.[0];
    const keyframes = source.match(
      /@keyframes\s+ei-resume-parse-pulse\s*\{[\s\S]*?\n\}/,
    )?.[0];

    expect(iconRule).toMatch(/width:\s*56px/);
    expect(iconRule).toMatch(/height:\s*56px/);
    expect(keyframes).toMatch(/box-shadow:/);
    expect(keyframes).not.toMatch(/transform:\s*(?:scale|translate)/);
    expect(source).toMatch(
      /@media\s*\(prefers-reduced-motion:\s*reduce\)[\s\S]*?\.ei-resume-detail-parse-icon\s*\{[^}]*animation:\s*none/,
    );
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
