export const SUPPORTED_LOCALES = [
  {
    code: "zh",
    label: "中文",
    shortLabel: "中",
    aliases: ["zh", "zh-cn", "zh-hans", "zh-hans-cn"],
  },
  {
    code: "en",
    label: "English",
    shortLabel: "EN",
    aliases: ["en", "en-us", "en-gb"],
  },
] as const;

export type Lang = (typeof SUPPORTED_LOCALES)[number]["code"];

export const DEFAULT_LANG: Lang = "en";

const LOCALE_ALIASES = new Map<string, Lang>(
  SUPPORTED_LOCALES.flatMap((locale) =>
    locale.aliases.map((alias) => [alias, locale.code] as const),
  ),
);

export function resolveSupportedLocale(tag: string | undefined | null): Lang | null {
  const lower = tag?.trim().toLowerCase();
  if (!lower) return null;
  const exact = LOCALE_ALIASES.get(lower);
  if (exact) return exact;
  const base = lower.split("-")[0];
  return LOCALE_ALIASES.get(base ?? "") ?? null;
}

export function normalizeLocaleTag(tag: string | undefined | null): Lang {
  return resolveSupportedLocale(tag) ?? DEFAULT_LANG;
}
