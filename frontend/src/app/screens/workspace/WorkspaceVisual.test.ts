import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const CSS_PATH = resolve(__dirname, "..", "screens.css");

function css(): string {
  return readFileSync(CSS_PATH, "utf8");
}

describe("Workspace list reference CSS contract", () => {
  it("uses a full-viewport canvas around the 1508px desktop content", () => {
    const source = css();
    expect(source).toMatch(
      /\.ei-workspace-plan-list\s*\{[^}]*width:\s*100%[^}]*min-height:\s*calc\(100vh - 76px\)[^}]*background:/s,
    );
    expect(source).toMatch(
      /\.ei-workspace-plan-inner\s*\{[^}]*width:\s*calc\(100% - 48px\)[^}]*max-width:\s*1508px[^}]*padding:\s*50px 0 78px/s,
    );
    expect(source).not.toMatch(/\.ei-workspace-plan-list\s*\{[^}]*max-width:/s);
  });

  it("keeps a two-column wide-card grid inside the full-width canvas", () => {
    const source = css();
    expect(source).toMatch(
      /\.ei-workspace-plan-header\s*\{[^}]*max-width:\s*1456px/s,
    );
    expect(source).toMatch(
      /\.ei-workspace-plan-grid\s*\{[^}]*max-width:\s*1456px[^}]*grid-template-columns:\s*repeat\(2, minmax\(0, 1fr\)\)[^}]*gap:\s*28px/s,
    );
    expect(source).not.toContain("repeat(auto-fill, minmax(300px, 360px))");
  });

  it("locks the reference card, action and responsive hierarchy", () => {
    const source = css();
    expect(source).toMatch(
      /\.ei-workspace-card\s*\{[^}]*min-height:\s*384px[^}]*border-radius:\s*12px/s,
    );
    expect(source).toMatch(
      /\.ei-workspace-plan-create\s*\{[^}]*height:\s*64px/s,
    );
    expect(source).toMatch(
      /\.ei-workspace-card-primary\s*\{[^}]*min-width:\s*210px[^}]*height:\s*56px/s,
    );
    expect(source).toMatch(
      /@media\s*\(max-width:\s*760px\)[\s\S]*?\.ei-workspace-plan-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\)/,
    );
  });
});
