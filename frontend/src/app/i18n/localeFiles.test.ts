import { existsSync, readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

describe("D1 shell i18n locale file structure", () => {
  it("keeps each supported locale in an independent locale file", () => {
    const zhFile = new URL("./locales/zh.ts", import.meta.url);
    const enFile = new URL("./locales/en.ts", import.meta.url);
    const messagesFile = new URL("./messages.ts", import.meta.url);
    const catalogFile = new URL("./localeCatalog.ts", import.meta.url);

    expect(existsSync(zhFile)).toBe(true);
    expect(existsSync(enFile)).toBe(true);
    expect(existsSync(catalogFile)).toBe(true);

    const zhSource = readFileSync(zhFile, "utf8");
    const enSource = readFileSync(enFile, "utf8");
    const messagesSource = readFileSync(messagesFile, "utf8");
    const catalogSource = readFileSync(catalogFile, "utf8");

    expect(zhSource).toContain('export const zh');
    expect(enSource).toContain('export const en');
    expect(messagesSource).not.toMatch(/\bzh:\s*\{/);
    expect(messagesSource).not.toMatch(/\ben:\s*\{/);
    expect(catalogSource).toContain("SUPPORTED_LOCALES");
    expect(catalogSource).toContain("aliases");
    expect(catalogSource).toContain("shortLabel");
  });

  it("contains home.* namespace keys in both zh and en (≥14 keys)", () => {
    const zhSource = readFileSync(
      new URL("./locales/zh.ts", import.meta.url),
      "utf8",
    );
    const enSource = readFileSync(
      new URL("./locales/en.ts", import.meta.url),
      "utf8",
    );

    const requiredKeys = [
      "home.heroLabel",
      "home.heroTitle",
      "home.heroSub",
      "home.jdPlaceholder",
      "home.importBtn",
      "home.orUpload",
      "home.recentSection",
      "home.recentSectionSub",
      "home.debriefTitle",
      "home.debriefSub",
      "home.debriefBtn",
      "home.resumeCreateLink",
    ];

    // product-scope v2.1 D-17 removed the 3 home.jobPicks* keys with the
    // jd_match module, trimming the home namespace lower bound to 12.
    expect(requiredKeys.length).toBeGreaterThanOrEqual(12);

    for (const key of requiredKeys) {
      expect(zhSource).toContain(`"${key}"`);
      expect(enSource).toContain(`"${key}"`);
    }
  });

  it("does not keep retired Resume Workshop per-row tree toast keys", () => {
    const zhSource = readFileSync(
      new URL("./locales/zh.ts", import.meta.url),
      "utf8",
    );
    const enSource = readFileSync(
      new URL("./locales/en.ts", import.meta.url),
      "utf8",
    );

    for (const source of [zhSource, enSource]) {
      expect(source).not.toContain("resumeWorkshop.tree.toastSelect");
      expect(source).not.toContain("resumeWorkshop.tree.toastBranch");
    }
  });

  it("does not keep retired Resume Workshop guided create keys", () => {
    const zhSource = readFileSync(
      new URL("./locales/zh.ts", import.meta.url),
      "utf8",
    );
    const enSource = readFileSync(
      new URL("./locales/en.ts", import.meta.url),
      "utf8",
    );

    for (const source of [zhSource, enSource]) {
      expect(source).not.toContain("resumeWorkshop.create.tabs.guided");
      expect(source).not.toContain("resumeWorkshop.create.guided.");
      expect(source).not.toContain("resumeWorkshop.sourceType.guided");
    }
  });
});
