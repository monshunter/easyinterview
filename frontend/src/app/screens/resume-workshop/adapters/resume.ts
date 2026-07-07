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

export type UiBulletStatus = "pending" | "accepted";

export interface UiBullet {
  id: string;
  section: string;
  original: string;
  rewritten: string;
  why: string[];
  status: UiBulletStatus;
}

export interface ResumeSuggestionInput {
  id: string;
  originalBullet: string;
  suggestedBullet: string;
  reason: string;
  section?: string;
}

const LANG_TAG_MAP: Record<string, string> = {
  zh: "中",
  en: "EN",
};

const SOURCE_TYPE_LABEL: Record<string, string> = {
  upload: "Uploaded file",
  paste: "Pasted text",
};

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

const deriveText = (resume: Resume): string[] => {
  const snapshot = resume.parsedTextSnapshot;
  if (typeof snapshot === "string" && snapshot.trim().length > 0) {
    return snapshot
      .split(/\n+/)
      .map((line) => line.trim())
      .filter((line) => line.length > 0);
  }
  const original = resume.originalText;
  if (typeof original === "string" && original.trim().length > 0) {
    return original
      .split(/\n+/)
      .map((line) => line.trim())
      .filter((line) => line.length > 0);
  }
  return [];
};

const normalizeStatus = (status: Resume["status"]): ResumeStatus =>
  status === "archived" ? "archived" : "active";

/**
 * Maps the flat OpenAPI `Resume` to the UI source shape consumed by the list
 * row and detail header. `displayName` is the editable
 * label; `title` is the original-source title fallback.
 */
export const mapResumeToUiSource = (resume: Resume): UiResumeSource => ({
  id: resume.id,
  name: resume.displayName || resume.title,
  langTag: deriveLangTag(resume.language),
  type: deriveSourceTypeLabel(resume.sourceType),
  sourceName: resume.title,
  createdAt: formatDateOnly(resume.createdAt),
  updatedAt: formatDateOnly(resume.updatedAt),
  status: normalizeStatus(resume.status),
  summary: deriveSummary(resume.parsedSummary),
  text: deriveText(resume),
});

const splitWhy = (reason: string): string[] => {
  if (!reason) return [];
  return reason
    .split(/\s*[|;；]\s*/)
    .map((part) => part.trim())
    .filter((part) => part.length > 0);
};

/**
 * Maps an ephemeral resume-tailor bullet suggestion to the UI bullet shape.
 * D-20: suggestions are accept-only and not persisted server-side until the
 * accepted set is saved via overwrite / save-as-new, so `status` starts
 * `pending` and is tracked client-side.
 */
export const mapBulletSuggestionToUi = (
  input: ResumeSuggestionInput,
): UiBullet => ({
  id: input.id,
  section: input.section ?? "",
  original: input.originalBullet,
  rewritten: input.suggestedBullet,
  why: splitWhy(input.reason),
  status: "pending",
});

const safeString = (value: unknown): string =>
  typeof value === "string" ? value : "";

const safeStringArray = (value: unknown): string[] =>
  Array.isArray(value) ? value.filter((s): s is string => typeof s === "string") : [];

export interface ResumePreviewProjection {
  headline: string;
  summary: string;
  skills: string[];
  sections: { title: string; bullets: string[] }[];
}

export const buildResumePreview = (resume: Resume): ResumePreviewProjection => {
  const profile = (resume.structuredProfile ?? {}) as Record<string, unknown>;
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
  return {
    headline: safeString(profile.headline),
    summary: safeString(profile.summary),
    skills: safeStringArray(profile.skills),
    sections,
  };
};

/**
 * Plain-text projection of the resume preview, suitable for clipboard copy.
 * Reads structuredProfile from the API response so it stays in sync with real
 * data once backend lands.
 */
export const buildResumePlainText = (resume: Resume): string => {
  const projection = buildResumePreview(resume);
  const lines: string[] = [];
  if (projection.headline) lines.push(projection.headline);
  if (projection.summary) lines.push(projection.summary);
  for (const section of projection.sections) {
    if (lines.length > 0) lines.push("");
    if (section.title) lines.push(section.title);
    for (const bullet of section.bullets) {
      lines.push(`- ${bullet}`);
    }
  }
  if (projection.skills.length > 0) {
    if (lines.length > 0) lines.push("");
    lines.push(projection.skills.join(" · "));
  }
  return lines.join("\n");
};
