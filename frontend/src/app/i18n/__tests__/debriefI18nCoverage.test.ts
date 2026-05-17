/**
 * Phase 7.1 — frontend-debrief i18n namespace coverage.
 *
 * Asserts (TestI18n_DebriefNamespaceComplete):
 *   - debrief.* keys are byte-identical sets across zh + en.
 *   - debrief.* total key count is large enough to cover header / stepper /
 *     contextStrip / picker / record (text+voice) / failure / missing /
 *     timeout / analysis / replay / severity slices.
 *   - failure.code.* covers every B1 canonical AI error code surfaced by
 *     useSuggestDebriefQuestions / useSubmitDebrief plus the
 *     IDEMPOTENCY_KEY_MISMATCH retry copy and VALIDATION_FAILED copy.
 *   - picker / mode toggle / step buttons all have both zh + en entries.
 */

import { describe, expect, it } from "vitest";

import { en } from "../locales/en";
import { zh } from "../locales/zh";

function pick(prefix: string, keys: ReadonlyArray<string>): ReadonlyArray<string> {
  return keys.filter((key) => key.startsWith(prefix)).sort();
}

const zhKeys = Object.keys(zh);
const enKeys = Object.keys(en);

describe("frontend-debrief/001 i18n coverage", () => {
  it("debrief.* keys are present in both locales and identical sets", () => {
    const zhDebrief = pick("debrief.", zhKeys);
    const enDebrief = pick("debrief.", enKeys);
    expect(zhDebrief).toEqual(enDebrief);
    // The debrief surface spans header (~9) + stepper (~4) + contextStrip
    // (~7) + picker (~22) + record (~26) + failure (~12) + missing (~4) +
    // timeout (~5) + analysis (~10) + replay (~5) + severity (~3). The
    // floor is set lower than the actual count to leave room for future
    // additions without locking exact key shapes.
    expect(zhDebrief.length).toBeGreaterThanOrEqual(80);
  });

  it("debrief.failure.code.* covers AI_* canonical codes + idempotency mismatch + validation failure + DEBRIEF_NOT_FOUND + unknown", () => {
    const required = [
      "AI_PROVIDER_TIMEOUT",
      "AI_PROVIDER_SECRET_MISSING",
      "AI_PROVIDER_CONFIG_INVALID",
      "AI_OUTPUT_INVALID",
      "AI_FALLBACK_EXHAUSTED",
      "DEBRIEF_NOT_FOUND",
      "IDEMPOTENCY_KEY_MISMATCH",
      "VALIDATION_FAILED",
      "UNKNOWN",
    ];
    for (const code of required) {
      const key = `debrief.failure.code.${code}`;
      expect(zhKeys, `zh missing ${key}`).toContain(key);
      expect(enKeys, `en missing ${key}`).toContain(key);
    }
  });

  it("debrief stepper / contextStrip / picker / record / submit / analysis / replay slices have keys", () => {
    const must = [
      "debrief.stepper.step0",
      "debrief.stepper.step1",
      "debrief.stepper.step2",
      "debrief.contextStrip.targetJobLabel",
      "debrief.contextStrip.mockSessionLabel",
      "debrief.contextStrip.resumeLabel",
      "debrief.picker.targetJob.title",
      "debrief.picker.mockSession.title",
      "debrief.picker.resume.title",
      "debrief.record.summary.eyebrow",
      "debrief.record.mode.text",
      "debrief.record.mode.voice",
      "debrief.record.submit.cta",
      "debrief.analysis.eyebrow",
      "debrief.analysis.dimensionsEyebrow",
      "debrief.analysis.provenance.title",
      "debrief.replay.eyebrow",
      "debrief.replay.cta",
      "debrief.severity.low",
      "debrief.severity.medium",
      "debrief.severity.high",
    ];
    for (const key of must) {
      expect(zhKeys, `zh missing ${key}`).toContain(key);
      expect(enKeys, `en missing ${key}`).toContain(key);
    }
  });
});
