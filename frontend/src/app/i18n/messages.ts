import {
  useDisplayPreferencesOptional,
  type Lang,
} from "../display/DisplayPreferencesProvider";
import { en } from "./locales/en";
import { zh, type LocaleMessages, type MessageKey } from "./locales/zh";

export const DEFAULT_LANG: Lang = "en";

export function normalizeLocaleTag(tag: string | undefined | null): Lang {
  const lower = tag?.trim().toLowerCase();
  if (!lower) return DEFAULT_LANG;
  if (lower === "en" || lower.startsWith("en-")) return "en";
  if (lower === "zh" || lower.startsWith("zh-")) return "zh";
  return DEFAULT_LANG;
}

export const MESSAGES = {
  zh,
  en,
} satisfies Record<Lang, LocaleMessages>;

export type { MessageKey };

export function translate(lang: Lang, key: MessageKey): string {
  return MESSAGES[lang][key] ?? MESSAGES[DEFAULT_LANG][key];
}

export function splitMessageList(lang: Lang, key: MessageKey): string[] {
  return translate(lang, key).split("|");
}

export function useI18n(): {
  lang: Lang;
  t: (key: MessageKey) => string;
  list: (key: MessageKey) => string[];
} {
  const prefs = useDisplayPreferencesOptional();
  const lang = prefs?.lang ?? DEFAULT_LANG;
  return {
    lang,
    t: (key) => translate(lang, key),
    list: (key) => splitMessageList(lang, key),
  };
}
