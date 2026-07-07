import type { Lang } from "../../../../i18n/messages";

export type CreateMode = "upload" | "paste";

const MAX_TITLE_LENGTH = 80;

const FALLBACK_TITLES: Record<Lang, Record<CreateMode, string>> = {
  zh: {
    upload: "上传文件",
    paste: "粘贴文本",
  },
  en: {
    upload: "Uploaded file",
    paste: "Pasted text",
  },
};

const truncateTitle = (value: string): string =>
  value.length > MAX_TITLE_LENGTH ? value.slice(0, MAX_TITLE_LENGTH) : value;

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
