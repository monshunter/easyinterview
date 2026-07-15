/**
 * Phase 5.7 — Generating module out-of-scope negative gate.
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
  /window\.EI_DATA/,
  /listTargetJobReports/,
  /\bProgressBar\b/,
  /\bPhaseList\b/,
  /\bLiveEvidenceStream\b/,
  /\bSlaHint\b/,
  /generating-notify-cta/,
  /generating-live-stream/,
  /HANDOFF_PASSTHROUGH_KEYS/,
  /reportStatus/,
];

function* walk(dir: string): Generator<string> {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      if (entry === "__tests__") continue;
      yield* walk(full);
      continue;
    }
    if (/\.(ts|tsx)$/.test(entry)) {
      yield full;
    }
  }
}

describe("frontend-report-dashboard/001 generating module negative grep", () => {
  it("contains no out-of-scope vocabulary in generating implementation", () => {
    const offenders: Array<{ file: string; needle: string }> = [];
    for (const file of walk(SCREEN_DIR)) {
      const text = readFileSync(file, "utf8");
      for (const needle of FORBIDDEN) {
        if (needle.test(text)) {
          offenders.push({ file, needle: String(needle) });
        }
      }
    }
    expect(offenders).toEqual([]);
  });
});
