import { describe, expect, it } from "vitest";

import { MESSAGES } from "./messages";

const PROFILE_REQUIRED_KEYS = [
  "jdMatch.profile.searchingAs",
  "jdMatch.profile.searchingAsLoading",
  "jdMatch.profile.summaryYearsUnit",
  "jdMatch.profile.summaryFallback",
  "jdMatch.profile.sourcesHeading",
  "jdMatch.profile.sourcesUnitResumes",
  "jdMatch.profile.sourcesUnitJds",
  "jdMatch.profile.sourcesUnitMocks",
  "jdMatch.profile.sourcesUnitDebriefs",
  "jdMatch.profile.sourcesEmpty",
];

const AGENT_REQUIRED_KEYS = [
  "jdMatch.agent.statusIdle",
  "jdMatch.agent.statusScanning",
  "jdMatch.agent.statusError",
  "jdMatch.agent.statusLoading",
  "jdMatch.agent.lastScanLabel",
  "jdMatch.agent.nextScanLabel",
];

const RECOMMENDED_REQUIRED_KEYS = [
  "jdMatch.recommended.scoreLabelStrong",
  "jdMatch.recommended.scoreLabelGood",
  "jdMatch.recommended.scoreLabelStretch",
  "jdMatch.recommended.topReasonLabel",
  "jdMatch.recommended.fitMustTemplate",
  "jdMatch.recommended.fitPlusTemplate",
  "jdMatch.recommended.whyMatchesHeading",
  "jdMatch.recommended.whereStretchHeading",
  "jdMatch.recommended.roleSnapshotHeading",
  "jdMatch.recommended.intelHeading",
];

const SEARCH_REQUIRED_KEYS = [
  "jdMatch.search.inputPlaceholder",
  "jdMatch.search.runButton",
  "jdMatch.search.runButtonRunning",
  "jdMatch.search.dataSourcesHeading",
  "jdMatch.search.searchingPanelLabel",
  "jdMatch.search.searchingStep1",
  "jdMatch.search.searchingStep2",
  "jdMatch.search.searchingStep3",
  "jdMatch.search.searchingStep4",
  "jdMatch.search.searchingStep5",
  "jdMatch.search.savedSearchesHeading",
  "jdMatch.search.savedSearchSaveCurrent",
  "jdMatch.search.filterAll",
  "jdMatch.search.filterStrong",
  "jdMatch.search.filterRemote",
  "jdMatch.search.filterUnseen",
];

const WATCHLIST_REQUIRED_KEYS = [
  "jdMatch.watchlist.heading",
  "jdMatch.watchlist.toneOk",
  "jdMatch.watchlist.toneWarn",
  "jdMatch.watchlist.toneMuted",
  "jdMatch.watchlist.addedTimeLabel",
  "jdMatch.watchlist.empty",
  "jdMatch.watchlist.refreshFooter",
  "jdMatch.watchlist.error",
  "jdMatch.watchlist.marketSignalsHeading",
  "jdMatch.watchlist.marketSignalUnavailable",
];

const RETIRED_PLACEHOLDER_KEYS = [
  "jdMatch.placeholderTitle",
  "jdMatch.placeholderCopy",
  "jdMatch.placeholderCta",
];

// UI-design SearchTab dynamic JD numbers must NOT appear in any frontend i18n
// value (per plan §3.5 "UI dynamic-data negative · 动态 JD 数字" row).
const FORBIDDEN_DYNAMIC_DATA_TOKENS = [
  "248",
  "87",
  "unique postings",
  "唯一岗位",
  "248 → 87",
];

describe("D1 jdMatch i18n namespaces (item 2.4)", () => {
  it("zh and en dictionaries share an identical key set", () => {
    const zhKeys = new Set(Object.keys(MESSAGES.zh));
    const enKeys = new Set(Object.keys(MESSAGES.en));
    const onlyInZh = [...zhKeys].filter((k) => !enKeys.has(k));
    const onlyInEn = [...enKeys].filter((k) => !zhKeys.has(k));
    expect(onlyInZh, "keys present in zh but missing in en").toEqual([]);
    expect(onlyInEn, "keys present in en but missing in zh").toEqual([]);
  });

  it("contains jdMatch.profile.* namespace (≥8 required keys) in both locales", () => {
    expect(PROFILE_REQUIRED_KEYS.length).toBeGreaterThanOrEqual(8);
    for (const key of PROFILE_REQUIRED_KEYS) {
      expect(MESSAGES.zh, `zh missing ${key}`).toHaveProperty(key);
      expect(MESSAGES.en, `en missing ${key}`).toHaveProperty(key);
    }
  });

  it("contains jdMatch.agent.* namespace (≥6 required keys) in both locales", () => {
    expect(AGENT_REQUIRED_KEYS.length).toBeGreaterThanOrEqual(6);
    for (const key of AGENT_REQUIRED_KEYS) {
      expect(MESSAGES.zh, `zh missing ${key}`).toHaveProperty(key);
      expect(MESSAGES.en, `en missing ${key}`).toHaveProperty(key);
    }
  });

  it("contains jdMatch.recommended.* namespace (≥10 required keys) in both locales", () => {
    expect(RECOMMENDED_REQUIRED_KEYS.length).toBeGreaterThanOrEqual(10);
    for (const key of RECOMMENDED_REQUIRED_KEYS) {
      expect(MESSAGES.zh, `zh missing ${key}`).toHaveProperty(key);
      expect(MESSAGES.en, `en missing ${key}`).toHaveProperty(key);
    }
  });

  it("contains jdMatch.search.* namespace (≥16 required keys incl. 5 step keys) in both locales", () => {
    expect(SEARCH_REQUIRED_KEYS.length).toBeGreaterThanOrEqual(16);
    for (const key of SEARCH_REQUIRED_KEYS) {
      expect(MESSAGES.zh, `zh missing ${key}`).toHaveProperty(key);
      expect(MESSAGES.en, `en missing ${key}`).toHaveProperty(key);
    }
    // Explicit 5-step coverage gate
    for (let step = 1; step <= 5; step++) {
      expect(MESSAGES.zh).toHaveProperty(`jdMatch.search.searchingStep${step}`);
      expect(MESSAGES.en).toHaveProperty(`jdMatch.search.searchingStep${step}`);
    }
  });

  it("contains jdMatch.watchlist.* namespace (≥10 required keys) in both locales", () => {
    expect(WATCHLIST_REQUIRED_KEYS.length).toBeGreaterThanOrEqual(10);
    for (const key of WATCHLIST_REQUIRED_KEYS) {
      expect(MESSAGES.zh, `zh missing ${key}`).toHaveProperty(key);
      expect(MESSAGES.en, `en missing ${key}`).toHaveProperty(key);
    }
  });

  it("does not register any retired plan-001 placeholder keys", () => {
    for (const key of RETIRED_PLACEHOLDER_KEYS) {
      expect(MESSAGES.zh).not.toHaveProperty(key);
      expect(MESSAGES.en).not.toHaveProperty(key);
    }
  });

  it("does not embed ui-design dynamic JD numbers in any locale value", () => {
    for (const [lang, dict] of Object.entries(MESSAGES)) {
      for (const [key, value] of Object.entries(dict)) {
        for (const token of FORBIDDEN_DYNAMIC_DATA_TOKENS) {
          expect(
            String(value).includes(token),
            `${lang}.${key} contains forbidden dynamic-data token "${token}"`,
          ).toBe(false);
        }
      }
    }
  });
});
