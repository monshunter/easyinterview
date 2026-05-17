/**
 * frontend-resume-workshop/002 Phase 6.11 + 6.12 — negative grep gates.
 *
 * (a) The create-flow source tree must NOT reference retired module names that
 *     are explicitly out of scope per spec D-7. The test scans the actual
 *     source files and asserts zero hits (excluding this very file).
 *
 * (b) The create-flow source tree must NOT import the static UI prototype
 *     (ui-design/src/data.jsx, ui-design/src/screen-resume-workshop.jsx) as a
 *     runtime data / component source. The test scans for those import paths.
 */
import { readdirSync, readFileSync, statSync } from "node:fs";
import { extname, join, resolve } from "node:path";

import { describe, expect, it } from "vitest";

const CREATE_DIR = resolve(
  __dirname,
  "../",
  "create",
);

function* walk(dir: string): Generator<string> {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      yield* walk(full);
      continue;
    }
    const ext = extname(entry);
    if (ext === ".ts" || ext === ".tsx") yield full;
  }
}

function isTestFile(file: string): boolean {
  return /\.test\.(ts|tsx)$/.test(file);
}

describe("frontend-resume-workshop/002 — retired module negative grep", () => {
  const retiredPattern =
    /\b(welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true)\b/;

  it("create-flow source files (non-test) contain zero retired-module references", () => {
    const offenders: Array<{ file: string; match: string }> = [];
    for (const file of walk(CREATE_DIR)) {
      if (isTestFile(file)) continue;
      const content = readFileSync(file, "utf8");
      const match = content.match(retiredPattern);
      if (match) offenders.push({ file, match: match[0]! });
    }
    expect(offenders).toEqual([]);
  });
});

describe("frontend-resume-workshop/002 — prototype runtime import negative grep", () => {
  const prototypeImportPattern =
    /from\s+["'][^"']*ui-design\/src\/(data|screen-resume-workshop)/;
  it("create-flow source files (non-test) do not import ui-design/src/data or screen-resume-workshop as a runtime dependency", () => {
    const offenders: Array<{ file: string }> = [];
    for (const file of walk(CREATE_DIR)) {
      if (isTestFile(file)) continue;
      const content = readFileSync(file, "utf8");
      if (prototypeImportPattern.test(content)) offenders.push({ file });
    }
    expect(offenders).toEqual([]);
  });
});
