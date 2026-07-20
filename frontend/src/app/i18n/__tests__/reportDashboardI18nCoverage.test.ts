/**
 * Phase 5.8 — Report-dashboard / Generating i18n namespace coverage.
 *
 * Asserts:
 *   - report.* keys exist in both locales and are byte-identical sets.
 *   - generating.* keys exist in both locales and are byte-identical sets.
 *   - combined report.* + generating.* key count is at least 60.
 *   - Internal error-code/provider copy is absent from the user-facing locale
 *     catalog; REPORT_NOT_FOUND keeps dedicated user-safe state copy.
 */

import { describe, expect, it } from "vitest";

import { en } from "../locales/en";
import { zh } from "../locales/zh";

function pick(prefix: string, keys: ReadonlyArray<string>): ReadonlyArray<string> {
  return keys.filter((key) => key.startsWith(prefix)).sort();
}

const zhKeys = Object.keys(zh);
const enKeys = Object.keys(en);

describe("frontend-report-dashboard/001 i18n coverage", () => {
  it("report.* keys are present in both locales and identical sets (TestReportNamespaceZhEnSync)", () => {
    const zhReport = pick("report.", zhKeys);
    const enReport = pick("report.", enKeys);
    expect(zhReport).toEqual(enReport);
    expect(zhReport.length).toBeGreaterThanOrEqual(40);
  });

  it("generating.* keys are present in both locales and identical sets (TestGeneratingNamespaceZhEnSync)", () => {
    const zhGen = pick("generating.", zhKeys);
    const enGen = pick("generating.", enKeys);
    expect(zhGen).toEqual(enGen);
    expect(zhGen.length).toBeGreaterThanOrEqual(20);
  });

  it("removes fake generating claims and keeps truthful terminal/recoverable copy", () => {
    for (const key of [
      "generating.progress.done",
      "generating.phase.1",
      "generating.evidence.streamLabel",
      "generating.sla.notifyCta",
    ]) {
      expect(zhKeys).not.toContain(key);
      expect(enKeys).not.toContain(key);
    }
    for (const key of [
      "generating.status.queued",
      "generating.status.generating",
      "generating.errors.contextTooLarge.desc",
      "generating.errors.continueCheck",
    ]) {
      expect(zhKeys).toContain(key);
      expect(enKeys).toContain(key);
    }
  });

  it("keeps enum chrome localized and the zero-answer reason synchronized", () => {
    for (const key of [
      "report.confidence.high",
      "report.confidence.medium",
      "report.confidence.low",
      "practice.finishDisabled.zeroAnswer",
    ]) {
      expect(zhKeys).toContain(key);
      expect(enKeys).toContain(key);
    }
  });

  it("does not retain an orphan session-locator label", () => {
    expect(zhKeys).not.toContain("report.context.session");
    expect(enKeys).not.toContain("report.context.session");
  });

  it("combined report.* + generating.* >= 60 keys (TestI18nKeyCountAtLeast60)", () => {
    const zhCombined = pick("report.", zhKeys).length + pick("generating.", zhKeys).length;
    const enCombined = pick("report.", enKeys).length + pick("generating.", enKeys).length;
    expect(zhCombined).toBeGreaterThanOrEqual(60);
    expect(enCombined).toBeGreaterThanOrEqual(60);
  });

  it("does not retain user-facing report error-code copy", () => {
    expect(pick("report.failureState.errorCode.", zhKeys)).toEqual([]);
    expect(pick("report.failureState.errorCode.", enKeys)).toEqual([]);
  });

  it("REPORT_NOT_FOUND uses dedicated user-safe state copy", () => {
    expect(zhKeys).toContain("report.failureState.notFound.title");
    expect(zhKeys).toContain("report.failureState.notFound.desc");
    expect(enKeys).toContain("report.failureState.notFound.title");
    expect(enKeys).toContain("report.failureState.notFound.desc");
  });

  it("distinguishes trusted reports Back from the workspace fallback in both locales", () => {
    expect(zh["generating.backToReports"]).toBe("返回面试报告");
    expect(en["generating.backToReports"]).toBe("Back to interview reports");
    expect(zh["common.back"]).toBe("返回");
    expect(en["common.back"]).toBe("Back");
  });
});
