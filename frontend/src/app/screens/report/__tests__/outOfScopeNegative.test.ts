/**
 * Phase 5.7 — Report module out-of-scope negative gate.
 *
 * Asserts:
 *   - Implementation files under frontend/src/app/screens/report/ do NOT
 *     import ui-design/src/data*, window.EI_DATA, prototype helpers,
 *     practice DOM, or out-of-scope report tab vocabulary.
 *   - Implementation does not call practice operations, voice operations, or workspace insight APIs
 *     operations.
 *   - listTargetJobReports is not referenced from report or generating
 *     implementation code (dashboard-only D-7).
 */

import { readdirSync, readFileSync, statSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { join, resolve } from "node:path";

import { describe, expect, it } from "vitest";

const SCREEN_DIR = resolve(
  fileURLToPath(new URL(".", import.meta.url)),
  "..",
);

const FORBIDDEN = [
  /reportLayout/,
  /report_layout/,
  /readinessScore/,
  /readiness_score/,
  /\bfully_prepared\b/,
  /mistakesQueue/,
  /mistakes_queue/,
  /drillBuilder/,
  /drill_builder/,
  /growthCenter/,
  /growth_center/,
  /reportTimeline/,
  /report_timeline/,
  /reportForm/,
  /report_form/,
  /createPracticeVoiceTurn/,
  /getCompany[A-Za-z]*Insight/,
  /getDebrief/,
  /from\s+["'][^"']*ui-design\/src\/screen-practice/,
  /window\.EI_DATA/,
  /ui-design\/src\/data/,
];

const PRACTICE_OPERATIONS_FORBIDDEN = [
  /getPracticeSession\b/,
  /appendSessionEvent\b/,
  /completePracticeSession\b/,
  /startPracticeSession\b/,
  /createPracticePlan\b/,
  /listTargetJobReports/,
];

function* walk(dir: string): Generator<string> {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      // Skip the __tests__ directory — tests are allowed to write negative
      // assertions that name forbidden terms.
      if (entry === "__tests__") continue;
      yield* walk(full);
      continue;
    }
    if (/\.(ts|tsx)$/.test(entry)) {
      yield full;
    }
  }
}

function survey(): Array<{ file: string; needle: string }> {
  const offenders: Array<{ file: string; needle: string }> = [];
  for (const file of walk(SCREEN_DIR)) {
    const text = readFileSync(file, "utf8");
    for (const needle of FORBIDDEN) {
      if (needle.test(text)) {
        offenders.push({ file, needle: String(needle) });
      }
    }
    for (const needle of PRACTICE_OPERATIONS_FORBIDDEN) {
      if (needle.test(text)) {
        offenders.push({ file, needle: String(needle) });
      }
    }
  }
  return offenders;
}

function surveyUnconsumedHelpers(): Array<{ file: string; name: string }> {
  const names = [
    ["isAi", "ErrorCode"].join(""),
    ["FAILURE_AI", "_ERROR_KEYS"].join(""),
  ];
  const offenders: Array<{ file: string; name: string }> = [];
  for (const file of walk(SCREEN_DIR)) {
    const text = readFileSync(file, "utf8");
    for (const name of names) {
      if (text.includes(name)) offenders.push({ file, name });
    }
  }
  return offenders;
}

describe("frontend-report-dashboard/001 report module negative grep", () => {
  it("contains no out-of-scope report-layout / readiness / mistakes / voice / insight API / data prototype imports", () => {
    expect(survey()).toEqual([]);
  });

  it("contains no report error helpers without repository consumers", () => {
    expect(surveyUnconsumedHelpers()).toEqual([]);
  });
});
