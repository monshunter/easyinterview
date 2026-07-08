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
      ".ei-resume-workshop-table",
      ".ei-resume-workshop-table-head",
      ".ei-resume-workshop-table-row",
      ".ei-resume-workshop-table-delete",
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
    expect(source).toContain("grid-template-columns: 1.8fr 1.4fr 0.6fr 1fr 132px");
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

  it("does not keep non-current tree/version/branch styling after D-20", () => {
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
});
