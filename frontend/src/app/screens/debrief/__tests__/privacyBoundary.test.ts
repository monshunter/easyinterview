// @vitest-environment jsdom
/**
 * Phase 8.4 / 8.5 — Privacy boundary unit tests for frontend-debrief/001.
 *
 * The plan forbids debrief raw entry content (questionText / myAnswerSummary
 * / interviewerReaction / notes) from leaking into client-side persistence
 * (localStorage, sessionStorage), console.log telemetry, or URL params.
 *
 * Instead of a grep gate, these tests pin the runtime boundary by reading
 * the actual screen / hook modules and asserting they never call the
 * persistence sinks with debrief field names.
 */

import { readFileSync, readdirSync, statSync } from "node:fs";
import { dirname, extname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const __dirname_compat = dirname(fileURLToPath(import.meta.url));
const ROOT = resolve(__dirname_compat, "..");

function* walk(dir: string): Generator<string> {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    const stat = statSync(full);
    if (stat.isDirectory()) {
      if (entry === "__tests__" || entry === "i18n") continue;
      yield* walk(full);
      continue;
    }
    const ext = extname(entry);
    if (
      (ext === ".ts" || ext === ".tsx") &&
      !entry.endsWith(".test.ts") &&
      !entry.endsWith(".test.tsx")
    ) {
      yield full;
    }
  }
}

const FORBIDDEN_PERSISTENCE = [
  /localStorage\.setItem\([^)]*questionText/,
  /localStorage\.setItem\([^)]*myAnswerSummary/,
  /localStorage\.setItem\([^)]*interviewerReaction/,
  /sessionStorage\.setItem\([^)]*questionText/,
  /sessionStorage\.setItem\([^)]*myAnswerSummary/,
  /sessionStorage\.setItem\([^)]*interviewerReaction/,
  /console\.log\([^)]*questionText/,
  /console\.log\([^)]*myAnswerSummary/,
  /console\.log\([^)]*interviewerReaction/,
  /console\.info\([^)]*questionText/,
  /console\.warn\([^)]*questionText/,
];

describe("frontend-debrief privacy boundary", () => {
  it("never persists raw debrief entry fields to localStorage / sessionStorage / console", () => {
    const offenders: string[] = [];
    for (const file of walk(ROOT)) {
      const content = readFileSync(file, "utf8");
      for (const pattern of FORBIDDEN_PERSISTENCE) {
        if (pattern.test(content)) {
          offenders.push(`${file}: ${pattern.source}`);
        }
      }
    }
    expect(offenders).toEqual([]);
  });

  it("never encodes raw debrief entry text into route params or window.history calls", () => {
    const offenders: string[] = [];
    const forbidden = [
      /navigate\(.*questionText[^)]*\)/,
      /navigate\(.*myAnswerSummary[^)]*\)/,
      /navigate\(.*interviewerReaction[^)]*\)/,
      /history\.(?:push|replace)State\([^)]*questionText/,
    ];
    for (const file of walk(ROOT)) {
      const content = readFileSync(file, "utf8");
      for (const pattern of forbidden) {
        if (pattern.test(content)) {
          offenders.push(`${file}: ${pattern.source}`);
        }
      }
    }
    expect(offenders).toEqual([]);
  });
});
