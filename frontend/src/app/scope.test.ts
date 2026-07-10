import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, resolve } from "node:path";

import { describe, expect, it } from "vitest";

const FRONTEND_SRC = resolve(__dirname, "..");

function* walk(dir: string): Generator<string> {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      yield* walk(full);
    } else if (/\.(ts|tsx)$/.test(entry)) {
      yield full;
    }
  }
}

describe("frontend D1 scope guards", () => {
  it("never imports ui-design/src/data.jsx into the active frontend bundle", () => {
    const importPattern = /from\s+["'][^"']*ui-design\/src\/data/;
    const offenders: string[] = [];
    for (const file of walk(FRONTEND_SRC)) {
      const content = readFileSync(file, "utf8");
      if (importPattern.test(content)) {
        offenders.push(file);
      }
    }
    expect(offenders).toEqual([]);
  });

  it("never ships standalone out-of-scope route screens (voice / growth / mistakes / drill)", () => {
    const FORBIDDEN_FILE_NAMES = [
      /\bVoiceScreen\.(tsx|ts)$/,
      /\bGrowthScreen\.(tsx|ts)$/,
      /\bMistakesScreen\.(tsx|ts)$/,
      /\bDrillScreen\.(tsx|ts)$/,
      /\bFollowupScreen\.(tsx|ts)$/,
      /\bExperiencesScreen\.(tsx|ts)$/,
      /\bWelcomeScreen\.(tsx|ts)$/,
    ];
    const offenders: string[] = [];
    for (const file of walk(FRONTEND_SRC)) {
      for (const pattern of FORBIDDEN_FILE_NAMES) {
        if (pattern.test(file)) offenders.push(file);
      }
    }
    expect(offenders).toEqual([]);
  });

  it("never references out-of-scope route names from active code", () => {
    // The route alias map in normalizeRoute.ts intentionally references the
    // out-of-scope names for compatibility normalization. All other active code
    // must avoid referencing these aliases as live route names.
    const ALIAS_OWNER = "/app/normalizeRoute.ts";
    const FORBIDDEN_LITERALS = [
      `name: "voice"`,
      `name: 'voice'`,
      `name: "growth"`,
      `name: 'growth'`,
      `name: "mistakes"`,
      `name: 'mistakes'`,
      `name: "drill"`,
      `name: 'drill'`,
      `name: "followup"`,
      `name: 'followup'`,
      `name: "welcome"`,
      `name: 'welcome'`,
    ];
    const offenders: Array<{ file: string; needle: string }> = [];
    for (const file of walk(FRONTEND_SRC)) {
      if (file.endsWith(ALIAS_OWNER)) continue;
      // Skip test files: tests legitimately exercise out-of-scope aliases through
      // normalizeRoute and via initialRoute={{ name: "welcome", ... }} loose
      // input. Their job is to prove the gate works.
      if (/\.test\.(ts|tsx)$/.test(file)) continue;
      const content = readFileSync(file, "utf8");
      for (const needle of FORBIDDEN_LITERALS) {
        if (content.includes(needle))
          offenders.push({ file, needle });
      }
    }
    expect(offenders).toEqual([]);
  });

  it("does not keep the out-of-scope voice alias in normalizeRoute", () => {
    const file = join(FRONTEND_SRC, "app", "normalizeRoute.ts");
    const content = readFileSync(file, "utf8");
    expect(content).not.toMatch(/^\s*voice\s*:/m);
  });

  it("does not keep out-of-scope JD Match CSS assets", () => {
    const file = join(FRONTEND_SRC, "app", "theme", "global.css");
    const content = readFileSync(file, "utf8");
    expect(content).not.toMatch(/jdmatch-/);
  });
});
