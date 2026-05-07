import { existsSync, readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

describe("D1 shell i18n locale file structure", () => {
  it("keeps each supported locale in an independent locale file", () => {
    const zhFile = new URL("./locales/zh.ts", import.meta.url);
    const enFile = new URL("./locales/en.ts", import.meta.url);
    const messagesFile = new URL("./messages.ts", import.meta.url);

    expect(existsSync(zhFile)).toBe(true);
    expect(existsSync(enFile)).toBe(true);

    const zhSource = readFileSync(zhFile, "utf8");
    const enSource = readFileSync(enFile, "utf8");
    const messagesSource = readFileSync(messagesFile, "utf8");

    expect(zhSource).toContain('export const zh');
    expect(enSource).toContain('export const en');
    expect(messagesSource).not.toMatch(/\bzh:\s*\{/);
    expect(messagesSource).not.toMatch(/\ben:\s*\{/);
  });
});
