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
      ".ei-resume-workshop-stats",
      ".ei-resume-workshop-view-switcher",
      ".ei-resume-workshop-tree-row",
      ".ei-resume-workshop-version-row",
      ".ei-resume-workshop-flat",
      ".ei-resume-detail",
      ".ei-resume-detail-preview",
      ".ei-resume-detail-modal",
    ]) {
      expect(source, `${selector} missing from screens.css`).toContain(selector);
    }
  });

  it("locks key values transcribed from the resume workshop prototype source", () => {
    const source = css();
    expect(source).toContain("max-width: 1320px");
    expect(source).toContain("padding: 40px 48px 96px");
    expect(source).toContain("grid-template-columns: repeat(4, minmax(0, 1fr))");
    expect(source).toContain("border-radius: 2px");
    expect(source).toContain("grid-template-columns: minmax(0, 1fr) 320px");
    expect(source).toContain("min-height: 720px");
    expect(source).toContain("box-shadow: 0 18px 50px rgba(30, 22, 15, 0.10)");
  });
});
