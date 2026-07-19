import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const HERE = resolve(__dirname);
const SOURCE = readFileSync(resolve(HERE, "ParseScreen.tsx"), "utf8");
const CSS = readFileSync(resolve(HERE, "..", "screens.css"), "utf8");

describe("Workspace plan-detail screenshot composition", () => {
  it("uses one page-scoped 1250px detail shell instead of inline max-width layout", () => {
    expect(SOURCE).toContain('className="ei-fadein ei-plan-detail-screen"');
    expect(SOURCE).not.toMatch(/maxWidth:\s*1200/);
    expect(CSS).toMatch(/\.ei-plan-detail-screen\s*\{[^}]*max-width:\s*1250px/);
  });

  it("keeps the title cluster and Start/Reports actions in one header grid", () => {
    expect(SOURCE).toContain('className="ei-plan-detail-header"');
    expect(SOURCE).toContain('className="ei-plan-detail-heading"');
    expect(SOURCE).toContain('className="ei-plan-detail-actions"');
    expect(CSS).toMatch(/\.ei-plan-detail-header\s*\{[^}]*display:\s*grid/);
    expect(CSS).toMatch(/\.ei-plan-detail-header\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s+auto/);
  });

  it("defines four card layers and a mobile single-column contract", () => {
    for (const className of [
      "ei-plan-detail-basics",
      "ei-plan-detail-requirements",
      "ei-plan-detail-hidden",
      "ei-plan-detail-rounds",
    ]) {
      expect(SOURCE).toContain(className);
    }
    expect(CSS).toMatch(/@media \(max-width:\s*720px\)[\s\S]*\.ei-plan-detail-header\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)/);
    expect(CSS).toMatch(/@media \(max-width:\s*720px\)[\s\S]*\.ei-plan-detail-round-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)/);
  });
});
