import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

describe("Practice reference visual contract", () => {
  it("keeps the active interview on the wide reference grid with rounded cards", () => {
    const css = readFileSync(resolve(__dirname, "../screens.css"), "utf8");

    expect(css).toMatch(/\.ei-practice-screen\s*\{[\s\S]*height:\s*calc\(100dvh - 76px\)/);
    expect(css).toMatch(/\.ei-practice-frame\s*\{[\s\S]*max-width:\s*1706px/);
    expect(css).toMatch(/\.ei-practice-session-header\s*\{[\s\S]*border-radius:\s*12px/);
    expect(css).toMatch(/\.ei-practice-conversation\s*\{[\s\S]*border-radius:\s*12px/);
    expect(css).toMatch(/\.ei-practice-input-shell\s*\{[\s\S]*border-radius:\s*10px/);
    expect(css).toMatch(/\.ei-practice-message-avatar\s*\{[\s\S]*border-radius:\s*7px/);
    expect(css).toMatch(/\.ei-practice-input-helper\s*\{[\s\S]*border-radius:\s*7px/);
  });

  it("uses semantic layout classes instead of page-level inline layout", () => {
    const screen = readFileSync(resolve(__dirname, "PracticeScreen.tsx"), "utf8");
    const topBar = readFileSync(resolve(__dirname, "components/TopBar.tsx"), "utf8");
    const transcript = readFileSync(resolve(__dirname, "components/Transcript.tsx"), "utf8");
    const input = readFileSync(resolve(__dirname, "components/InputBar.tsx"), "utf8");

    expect(screen).toContain('className="ei-practice-screen ei-fadein"');
    expect(screen).toContain('className="ei-practice-frame"');
    expect(screen).toContain('className="ei-practice-conversation"');
    expect(topBar).toContain('className="ei-practice-session-header"');
    expect(transcript).toContain('className="ei-practice-transcript"');
    expect(input).toContain('className="ei-practice-input-shell"');
    expect(transcript).not.toContain("helperText");
    expect(transcript).not.toContain("practice-transcript-helper");
    expect(input).toContain("helperText: string");
    expect(input).toContain('data-testid="practice-input-helper"');
    expect(input.indexOf('data-testid="practice-input-helper"')).toBeLessThan(
      input.indexOf('className="ei-practice-input-shell"'),
    );
    expect(input).toContain("<SparkleIcon />");
  });

  it("keeps the transcript as the only scrolling region and fixes the composer at the conversation bottom", () => {
    const css = readFileSync(resolve(__dirname, "../screens.css"), "utf8");
    const screen = readFileSync(resolve(__dirname, "PracticeScreen.tsx"), "utf8");

    expect(css).toMatch(/\.ei-practice-conversation\s*\{[\s\S]*display:\s*flex;[\s\S]*flex-direction:\s*column;[\s\S]*overflow:\s*hidden;/);
    expect(css).toMatch(/\.ei-practice-transcript\s*\{[\s\S]*flex:\s*1 1 auto;[\s\S]*min-height:\s*0;[\s\S]*overflow-y:\s*auto;/);
    expect(css).toMatch(/\.ei-practice-input\s*\{[\s\S]*flex:\s*0 0 auto;/);
    expect(screen.indexOf("<Transcript")).toBeLessThan(screen.indexOf("<InputBar"));
  });
});
