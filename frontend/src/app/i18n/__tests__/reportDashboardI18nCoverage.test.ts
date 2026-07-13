/**
 * Phase 5.8 — Report-dashboard / Generating i18n namespace coverage.
 *
 * Asserts:
 *   - report.* keys exist in both locales and are byte-identical sets.
 *   - generating.* keys exist in both locales and are byte-identical sets.
 *   - combined report.* + generating.* key count is at least 60.
 *   - report.failureState.errorCode.* covers all B1 AI_* enum values
 *     (TestErrorCodeI18nCoversAllAIErrors).
 *   - REPORT_NOT_FOUND has dedicated failureState.notFound.* keys + an
 *     errorCode.REPORT_NOT_FOUND label that is distinct from AI_* mapping
 *     (TestReportFailureStateErrorCodeCoversReportNotFound).
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

  it("combined report.* + generating.* >= 60 keys (TestI18nKeyCountAtLeast60)", () => {
    const zhCombined = pick("report.", zhKeys).length + pick("generating.", zhKeys).length;
    const enCombined = pick("report.", enKeys).length + pick("generating.", enKeys).length;
    expect(zhCombined).toBeGreaterThanOrEqual(60);
    expect(enCombined).toBeGreaterThanOrEqual(60);
  });

  it("report.failureState.errorCode.* covers all AI_* enum values (TestErrorCodeI18nCoversAllAIErrors)", () => {
    const aiCodes = [
      "AI_PROVIDER_TIMEOUT",
      "AI_PROVIDER_SECRET_MISSING",
      "AI_PROVIDER_CONFIG_INVALID",
      "AI_OUTPUT_INVALID",
      "AI_FALLBACK_EXHAUSTED",
      "AI_UNSUPPORTED_CAPABILITY",
    ];
    for (const code of aiCodes) {
      const key = `report.failureState.errorCode.${code}`;
      expect(zhKeys, `zh missing ${key}`).toContain(key);
      expect(enKeys, `en missing ${key}`).toContain(key);
    }
  });

  it("REPORT_NOT_FOUND uses dedicated keys distinct from AI_* (TestReportFailureStateErrorCodeCoversReportNotFound)", () => {
    expect(zhKeys).toContain("report.failureState.errorCode.REPORT_NOT_FOUND");
    expect(enKeys).toContain("report.failureState.errorCode.REPORT_NOT_FOUND");
    expect(zhKeys).toContain("report.failureState.notFound.title");
    expect(zhKeys).toContain("report.failureState.notFound.desc");
    expect(enKeys).toContain("report.failureState.notFound.title");
    expect(enKeys).toContain("report.failureState.notFound.desc");
    // Distinct from AI_* mapping.
    const zhCopy = (zh as Record<string, string>)["report.failureState.errorCode.REPORT_NOT_FOUND"];
    const enCopy = (en as Record<string, string>)["report.failureState.errorCode.REPORT_NOT_FOUND"];
    const aiCopy = (zh as Record<string, string>)["report.failureState.errorCode.AI_PROVIDER_TIMEOUT"];
    expect(zhCopy).toBeDefined();
    expect(enCopy).toBeDefined();
    expect(zhCopy).not.toBe(aiCopy);
  });
});
