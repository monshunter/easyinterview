import type { Resume, ResumeSummary } from "../../../../api/generated/types";

export type ResumeStatus = "active" | "archived";
export type ResumeDetailRenderer = "markdown" | "pdf";

export interface UiResumeSource {
  id: string;
  name: string;
  langTag: string;
  type: string;
  sourceName: string;
  createdAt: string;
  updatedAt: string;
  status: ResumeStatus;
  summary: string;
  text: string[];
}

export interface UiResumeListItem {
  id: string;
  name: string;
  langTag: string;
  sourceName: string;
  updatedAt: string;
  summary: string;
}

const LANG_TAG_MAP: Record<string, string> = {
  zh: "中",
  en: "EN",
};

const SOURCE_TYPE_LABEL: Record<string, string> = {
  upload: "上传文件",
  paste: "粘贴文本",
};

const PENDING_DISPLAY_NAME = "名称生成中";

const GENERIC_RESUME_NAMES = new Set([
  "粘贴的简历",
  "粘帖的简历",
  "上传的简历",
  "pasted resume",
  "uploaded resume",
  "pasted text",
  "uploaded file",
]);

const formatDateOnly = (iso: string): string => iso.slice(0, 10);

const deriveLangTag = (language: string): string => {
  const primary = language.split("-")[0]?.toLowerCase() ?? "";
  const mapped = LANG_TAG_MAP[primary];
  if (mapped) return mapped;
  return primary.toUpperCase();
};

const deriveSourceTypeLabel = (sourceType: Resume["sourceType"]): string => {
  if (!sourceType) return "Unknown";
  return SOURCE_TYPE_LABEL[sourceType] ?? sourceType;
};

const deriveSummary = (parsedSummary: Resume["parsedSummary"]): string => {
  if (!parsedSummary) return "";
  const headline = (parsedSummary as Record<string, unknown>).headline;
  return typeof headline === "string" ? headline : "";
};

const normalizeStatus = (status: Resume["status"]): ResumeStatus =>
  status === "archived" ? "archived" : "active";

const safeString = (value: unknown): string =>
  typeof value === "string" ? value.trim() : "";

const safeStringArray = (value: unknown): string[] =>
  Array.isArray(value)
    ? value
        .map((item) => safeString(item))
        .filter((item) => item.length > 0)
    : [];

const normalizeName = (value: string): string =>
  value
    .replace(/^\s{0,3}#{1,6}\s+/, "")
    .replace(/\s+/g, " ")
    .trim();

const isGenericResumeName = (value: string): boolean => {
  const normalized = normalizeName(value);
  if (!normalized) return true;
  return GENERIC_RESUME_NAMES.has(normalized.toLowerCase());
};

const isSameName = (left: string, right: string): boolean =>
  normalizeName(left).toLowerCase() === normalizeName(right).toLowerCase();

const looksLikeUploadFileName = (value: string): boolean =>
  /\.(pdf|txt|md|markdown)$/i.test(normalizeName(value));

const hasSourceExtension = (value: string, extensions: string[]): boolean => {
  const normalized = normalizeName(value).toLowerCase();
  return extensions.some((ext) => normalized.endsWith(ext));
};

const splitContentLines = (content: string | null | undefined): string[] => {
  if (typeof content !== "string" || content.trim().length === 0) return [];
  return content
    .split(/\r?\n+/)
    .map((line) => line.trim())
    .filter((line) => line.length > 0);
};

const profileRecord = (resume: Resume): Record<string, unknown> =>
  (resume.structuredProfile ?? {}) as Record<string, unknown>;

const parsedSummaryRecord = (resume: Resume): Record<string, unknown> =>
  (resume.parsedSummary ?? {}) as Record<string, unknown>;

const deriveStructuredName = (resume: Resume): string => {
  const profile = profileRecord(resume);
  const basics =
    typeof profile.basics === "object" && profile.basics !== null
      ? (profile.basics as Record<string, unknown>)
      : {};
  const parsed = parsedSummaryRecord(resume);

  const name = safeString(basics.name);
  const headline =
    safeString(basics.headline) ||
    safeString(basics.title) ||
    safeString(profile.headline) ||
    safeString(parsed.headline);

  if (name && headline) return `${name} · ${headline}`;
  if (name) return name;
  if (headline) return headline;

  const skills = safeStringArray(profile.skills);
  if (skills.length > 0) return skills.slice(0, 4).join(" · ");

  return "";
};

const firstNonGeneric = (values: string[], forbidden: string[] = []): string => {
  const value = values.find(
    (candidate) =>
      !isGenericResumeName(candidate) &&
      !forbidden.some((blocked) => isSameName(candidate, blocked)),
  );
  return value ? normalizeName(value) : "";
};

const deriveDisplayName = (resume: Resume): string =>
  firstNonGeneric(
    [safeString(resume.displayName), deriveStructuredName(resume)],
    [safeString(resume.title)].filter(
      (value) => isGenericResumeName(value) || looksLikeUploadFileName(value),
    ),
  ) || PENDING_DISPLAY_NAME;

const deriveSourceName = (resume: Resume): string => {
  if (resume.sourceType === "paste") return deriveSourceTypeLabel("paste");
  return (
    firstNonGeneric([safeString(resume.title), safeString(resume.displayName)]) ||
    deriveSourceTypeLabel(resume.sourceType)
  );
};

const deriveSummaryDisplayName = (resume: ResumeSummary): string =>
  firstNonGeneric(
    [safeString(resume.displayName), safeString(resume.summaryHeadline)],
    [safeString(resume.title)].filter(
      (value) => isGenericResumeName(value) || looksLikeUploadFileName(value),
    ),
  ) || PENDING_DISPLAY_NAME;

const deriveSummarySourceName = (resume: ResumeSummary): string => {
  if (resume.sourceType === "paste") return deriveSourceTypeLabel("paste");
  return (
    firstNonGeneric([safeString(resume.title), safeString(resume.displayName)]) ||
    deriveSourceTypeLabel(resume.sourceType)
  );
};

/**
 * Maps the full OpenAPI `Resume` to the detail-only UI source shape.
 * `displayName` is the LLM-derived label after parse;
 * `title` is the original-source title fallback.
 */
export const mapResumeToUiSource = (resume: Resume): UiResumeSource => ({
  id: resume.id,
  name: deriveDisplayName(resume),
  langTag: deriveLangTag(resume.language),
  type: deriveSourceTypeLabel(resume.sourceType),
  sourceName: deriveSourceName(resume),
  createdAt: formatDateOnly(resume.createdAt),
  updatedAt: formatDateOnly(resume.updatedAt),
  status: normalizeStatus(resume.status),
  summary: deriveSummary(resume.parsedSummary),
  text: buildResumeBodyLines(resume),
});

/** Maps the summary-only list contract without reading detail-only fields. */
export const mapResumeSummaryToUiSource = (
  resume: ResumeSummary,
): UiResumeListItem => ({
  id: resume.id,
  name: deriveSummaryDisplayName(resume),
  langTag: deriveLangTag(resume.language),
  sourceName: deriveSummarySourceName(resume),
  updatedAt: formatDateOnly(resume.updatedAt),
  summary: safeString(resume.summaryHeadline),
});

export const buildResumeBodyLines = (resume: Resume): string[] => {
  return splitContentLines(buildResumeBodyMarkdown(resume));
};

export const buildResumeBodyMarkdown = (resume: Resume): string => {
  const parsedMarkdown = safeString(resume.parsedTextSnapshot);
  if (parsedMarkdown) return parsedMarkdown;

  const rawText = safeString(resume.originalText);
  if (rawText) return rawText;

  const profile = profileRecord(resume);
  const sectionsRaw = profile.sections;
  const sections = Array.isArray(sectionsRaw)
    ? sectionsRaw.flatMap((entry) => {
        if (typeof entry !== "object" || entry === null) return [];
        const record = entry as Record<string, unknown>;
        return [
          {
            title: safeString(record.title),
            bullets: safeStringArray(record.bullets),
          },
        ];
      })
    : [];

  return [
    safeString(profile.headline),
    safeString(profile.summary),
    ...sections.flatMap((section) => [section.title, ...section.bullets]),
    safeStringArray(profile.skills).join(" · "),
  ]
    .filter((line) => line.length > 0)
    .join("\n");
};

export const getResumeDetailRenderer = (
  resume: Resume,
): ResumeDetailRenderer => {
  if (
    resume.sourceType === "upload" &&
    hasSourceExtension(safeString(resume.title), [".pdf"])
  ) {
    return "pdf";
  }
  return "markdown";
};

export const getResumeSourceUrl = (
  resume: Resume,
  baseUrl = "/api/v1",
): string => {
  const normalizedBase = baseUrl.replace(/\/+$/, "");
  return `${normalizedBase}/resumes/${encodeURIComponent(resume.id)}/source`;
};
