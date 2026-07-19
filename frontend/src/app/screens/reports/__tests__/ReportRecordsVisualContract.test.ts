import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const CSS_PATH = resolve(__dirname, "..", "..", "screens.css");

describe("report list and conversation visual source contract", () => {
  it("locks the shared desktop canvas and target compositions", () => {
    const source = readFileSync(CSS_PATH, "utf8");

    expect(source).toMatch(
      /\.ei-reports-screen\s*\{[^}]*max-width:\s*1372px/s,
    );
    expect(source).toMatch(
      /\.ei-report-conversation-screen\s*\{[^}]*max-width:\s*1372px/s,
    );
    expect(source).toContain(".ei-report-page-illustration");
    expect(source).toContain(".ei-reports-target-summary");
    expect(source).toContain(".ei-reports-timeline");
    expect(source).toContain(".ei-reports-round-index");
    expect(source).toContain(".ei-reports-round-card");
    expect(source).toContain(".ei-report-conversation-message-user");
    expect(source).toMatch(
      /\.ei-report-conversation-message\s*\{[^}]*background:[^}]*border:\s*1px solid[^}]*border-radius:\s*10px/s,
    );
    expect(source).not.toMatch(
      /\.ei-report-conversation-message-assistant\s*\{[^}]*border-bottom/s,
    );
    expect(source).toMatch(
      /\.ei-report-conversation-message-user \.ei-report-conversation-badge\s*\{[^}]*border-radius:\s*9px/s,
    );
    expect(source).toMatch(
      /\.ei-report-conversation-badge\s*\{[^}]*width:\s*60px[^}]*height:\s*60px/s,
    );
  });

  it("keeps both pages responsive without introducing a parallel runtime", () => {
    const source = readFileSync(CSS_PATH, "utf8");

    expect(source).toMatch(
      /@media\s*\(max-width:\s*700px\)[\s\S]*?\.ei-reports-screen[^}]*width:\s*100%/,
    );
    expect(source).toMatch(
      /@media\s*\(max-width:\s*700px\)[\s\S]*?\.ei-report-conversation-screen[^}]*width:\s*100%/,
    );
    expect(source).not.toContain("max-width: 1120px");
    expect(source).not.toContain("maxWidth: 880");
  });
});
