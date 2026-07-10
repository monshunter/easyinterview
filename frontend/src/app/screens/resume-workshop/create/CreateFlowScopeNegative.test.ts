/**
 * frontend-resume-workshop/002 Phase 6.11 + 6.12 — negative grep gates.
 *
 * (a) The create-flow source tree must NOT reference out-of-scope module names that
 *     are explicitly out of scope per spec D-7. The test scans the actual
 *     source files and asserts zero hits (excluding this very file).
 *
 * (b) The create-flow source tree must NOT import the static UI prototype
 *     runtime JSX/data files as a
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

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

describe("frontend-resume-workshop/002 — out-of-scope module negative grep", () => {
  const outOfScopeTerms = [
    ["wel", "come"].join(""),
    ["mis", "take"].join(""),
    ["gro", "wth"].join(""),
    ["dr", "ill"].join(""),
    ["follow", "up"].join(""),
    ["ST", "AR"].join(""),
    ["exper", "iences"].join(""),
    ["vo", "ice"].join(""),
    ["Onboarding", "Screen"].join(""),
    ["onboarding", "=true"].join(""),
    ["Resume", "Parse", "Flow"].join(""),
    ["Parsing", "Stage"].join(""),
    ["Preview", "Stage"].join(""),
    ["Resume", "Preview", "Confirm"].join(""),
    ["resume", "parse", "flow"].join("-"),
    ["resume", "preview", "confirm"].join("-"),
  ];
  const outOfScopePattern = new RegExp(
    `\\b(${outOfScopeTerms.map(escapeRegExp).join("|")})\\b`,
  );

  it("create-flow source files (non-test) contain zero out-of-scope module references", () => {
    const offenders: Array<{ file: string; match: string }> = [];
    for (const file of walk(CREATE_DIR)) {
      if (isTestFile(file)) continue;
      const content = readFileSync(file, "utf8");
      const match = content.match(outOfScopePattern);
      if (match) offenders.push({ file, match: match[0]! });
    }
    expect(offenders).toEqual([]);
  });
});

describe("frontend-resume-workshop/002 — prototype runtime import negative grep", () => {
  const prototypeRoot = ["ui-design", "src"].join("/");
  const prototypeFiles = [
    ["da", "ta"].join(""),
    ["screen-resume", "workshop"].join("-"),
  ];
  const prototypeImportPattern = new RegExp(
    `from\\s+["'][^"']*${escapeRegExp(prototypeRoot)}\\/(${prototypeFiles
      .map(escapeRegExp)
      .join("|")})`,
  );
  it("create-flow source files (non-test) do not import static prototype files as a runtime dependency", () => {
    const offenders: Array<{ file: string }> = [];
    for (const file of walk(CREATE_DIR)) {
      if (isTestFile(file)) continue;
      const content = readFileSync(file, "utf8");
      if (prototypeImportPattern.test(content)) offenders.push({ file });
    }
    expect(offenders).toEqual([]);
  });
});

describe("frontend-resume-workshop/002 — create stage ownership", () => {
  const standaloneStageType = ["Create", "Stage"].join("");

  it("keeps the input DOM marker without a standalone stage type", () => {
    const stageTypePattern = new RegExp(
      `\\b${escapeRegExp(standaloneStageType)}\\b`,
    );
    const offenders: Array<{ file: string }> = [];

    for (const file of walk(CREATE_DIR)) {
      if (isTestFile(file)) continue;
      if (stageTypePattern.test(readFileSync(file, "utf8"))) {
        offenders.push({ file });
      }
    }

    expect(offenders).toEqual([]);
    expect(
      readFileSync(join(CREATE_DIR, "ResumeCreateFlow.tsx"), "utf8"),
    ).toContain('data-stage="input"');
  });
});
