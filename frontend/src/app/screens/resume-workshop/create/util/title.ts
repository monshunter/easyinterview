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

const GENERIC_LINE_PATTERN =
  /^(resume|curriculum\s+vitae|cv|个人简历|简历)$/i;

const normalizeTitle = (value: string): string =>
  value.replace(/\s+/g, " ").trim();

const truncateTitle = (value: string): string =>
  value.length > MAX_TITLE_LENGTH ? value.slice(0, MAX_TITLE_LENGTH) : value;

const firstMeaningfulLine = (rawText: string): string | null => {
  const lines = rawText
    .split(/\r?\n+/)
    .map(normalizeTitle)
    .filter((line) => line.length > 0);

  const line = lines.find((candidate) => !GENERIC_LINE_PATTERN.test(candidate));
  return line ? truncateTitle(line) : null;
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

export function derivePasteTitle(rawText: string, lang: Lang): string {
  return firstMeaningfulLine(rawText) ?? deriveDefaultTitle("paste", lang, null);
}
