import type { Resume } from "../../../../api/generated/types";

export type ResumeStatus = "active" | "archived";

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

const LANG_TAG_MAP: Record<string, string> = {
  zh: "中",
  en: "EN",
};

const SOURCE_TYPE_LABEL: Record<string, string> = {
  upload: "上传文件",
  paste: "粘贴文本",
};

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
  value.replace(/\s+/g, " ").trim();

const isGenericResumeName = (value: string): boolean => {
  const normalized = normalizeName(value);
  if (!normalized) return true;
  return GENERIC_RESUME_NAMES.has(normalized.toLowerCase());
};

const splitContentLines = (content: string | null | undefined): string[] => {
  if (typeof content !== "string" || content.trim().length === 0) return [];
  return content
    .split(/\r?\n+/)
    .map((line) => line.trim())
    .filter((line) => line.length > 0);
};

const firstContentLine = (resume: Resume): string => {
  const line = [
    ...splitContentLines(resume.parsedTextSnapshot),
    ...splitContentLines(resume.originalText),
  ].find((candidate) => !isGenericResumeName(candidate));
  return line ?? "";
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

const firstNonGeneric = (values: string[]): string => {
  const value = values.find((candidate) => !isGenericResumeName(candidate));
  return value ? normalizeName(value) : "";
};

const deriveDisplayName = (resume: Resume): string =>
  firstNonGeneric([
    safeString(resume.displayName),
    safeString(resume.title),
    firstContentLine(resume),
    deriveStructuredName(resume),
  ]) || deriveSourceTypeLabel(resume.sourceType);

const deriveSourceName = (resume: Resume): string => {
  if (resume.sourceType === "paste") return deriveSourceTypeLabel("paste");
  return (
    firstNonGeneric([safeString(resume.title), safeString(resume.displayName)]) ||
    deriveSourceTypeLabel(resume.sourceType)
  );
};

/**
 * Maps the flat OpenAPI `Resume` to the UI source shape consumed by the list
 * row and detail header. `displayName` is the LLM-derived label after parse;
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

export const buildResumeBodyLines = (resume: Resume): string[] => {
  const originalLines = splitContentLines(resume.parsedTextSnapshot);
  if (originalLines.length > 0) return originalLines;

  const rawLines = splitContentLines(resume.originalText);
  if (rawLines.length > 0) return rawLines;

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
  ].filter((line) => line.length > 0);
};
