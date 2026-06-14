import type { Lang } from "../../../../i18n/messages";

export type CreateMode = "upload" | "paste";

const MAX_TITLE_LENGTH = 80;

const FALLBACK_TITLES: Record<Lang, Record<CreateMode, string>> = {
  zh: {
    upload: "上传的简历",
    paste: "粘贴的简历",
  },
  en: {
    upload: "Uploaded resume",
    paste: "Pasted resume",
  },
};

export function deriveDefaultTitle(
  mode: CreateMode,
  lang: Lang,
  pickedFileName?: string | null,
): string {
  if (mode === "upload" && pickedFileName) {
    const trimmed = pickedFileName.trim();
    if (trimmed) {
      return trimmed.length > MAX_TITLE_LENGTH
        ? trimmed.slice(0, MAX_TITLE_LENGTH)
        : trimmed;
    }
  }
  const fallback = FALLBACK_TITLES[lang]?.[mode] ?? FALLBACK_TITLES.zh[mode];
  return fallback;
}
