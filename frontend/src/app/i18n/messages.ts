import {
  useDisplayPreferencesOptional,
} from "../display/DisplayPreferencesProvider";
import {
  DEFAULT_LANG,
  type Lang,
} from "./localeCatalog";
import { en } from "./locales/en";
import { zh, type LocaleMessages, type MessageKey } from "./locales/zh";

export {
  DEFAULT_LANG,
  SUPPORTED_LOCALES,
  normalizeLocaleTag,
} from "./localeCatalog";
export type { Lang } from "./localeCatalog";

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
