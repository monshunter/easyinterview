import type {
  ResumeAsset,
  ResumeVersion as ApiResumeVersion,
} from "../../../../api/generated/types";

export type ResumeAssetStatus = "active" | "archived";

export interface UiResumeSource {
  id: string;
  name: string;
  langTag: string;
  type: string;
  createdAt: string;
  status: ResumeAssetStatus;
  summary: string;
  text: string[];
}

export type UiResumeVersionTag = "MASTER" | "TARGETED";

export interface UiResumeVersion {
  id: string;
  originalId: string;
  parentVersionId: string | null;
  name: string;
  tag: UiResumeVersionTag;
  date: string;
  target: string | null;
  bullets: number;
  accepted: number;
  match: number | null;
  archived: boolean;
}

export type UiBulletStatus = "pending" | "accepted" | "rejected";

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
  status?: string;
  section?: string;
}

const LANG_TAG_MAP: Record<string, string> = {
  zh: "中",
  en: "EN",
};

const SOURCE_TYPE_LABEL: Record<string, string> = {
  upload: "Uploaded",
  paste: "Pasted text",
  guided: "Guided answers",
};

const formatDateOnly = (iso: string): string => iso.slice(0, 10);

const deriveLangTag = (language: string): string => {
  const primary = language.split("-")[0]?.toLowerCase() ?? "";
  if (primary in LANG_TAG_MAP) return LANG_TAG_MAP[primary];
  return primary.toUpperCase();
};

const deriveSourceTypeLabel = (
  sourceType: ResumeAsset["sourceType"],
): string => {
  if (!sourceType) return "Unknown";
  return SOURCE_TYPE_LABEL[sourceType] ?? sourceType;
};

const deriveSummary = (parsedSummary: ResumeAsset["parsedSummary"]): string => {
  if (!parsedSummary) return "";
  const headline = (parsedSummary as Record<string, unknown>).headline;
  return typeof headline === "string" ? headline : "";
};

const deriveText = (asset: ResumeAsset): string[] => {
  const snapshot = asset.parsedTextSnapshot;
  if (typeof snapshot === "string" && snapshot.trim().length > 0) {
    return snapshot
      .split(/\n+/)
      .map((line) => line.trim())
      .filter((line) => line.length > 0);
  }
  const original = asset.originalText;
  if (typeof original === "string" && original.trim().length > 0) {
    return original
      .split(/\n+/)
      .map((line) => line.trim())
      .filter((line) => line.length > 0);
  }
  return [];
};

const normalizeStatus = (
  status: ResumeAsset["status"],
): ResumeAssetStatus => (status === "archived" ? "archived" : "active");

export const mapResumeAssetToUiSource = (asset: ResumeAsset): UiResumeSource => ({
  id: asset.id,
  name: asset.title,
  langTag: deriveLangTag(asset.language),
  type: deriveSourceTypeLabel(asset.sourceType),
  createdAt: formatDateOnly(asset.createdAt),
  status: normalizeStatus(asset.status),
  summary: deriveSummary(asset.parsedSummary),
  text: deriveText(asset),
});

const tagFromVersionType = (
  versionType: ApiResumeVersion["versionType"],
): UiResumeVersionTag => (versionType === "structured_master" ? "MASTER" : "TARGETED");

const formatMatchScore = (matchScore: number | null | undefined): number | null => {
  if (matchScore === null || matchScore === undefined) return null;
  if (matchScore <= 1) return Math.round(matchScore * 100);
  return Math.round(matchScore);
};

const safeStringField = (
  raw: unknown,
  field: string,
): string | undefined => {
  if (typeof raw !== "object" || raw === null) return undefined;
  const value = (raw as Record<string, unknown>)[field];
  return typeof value === "string" ? value : undefined;
};

const countSectionBullets = (
  structuredProfile: ApiResumeVersion["structuredProfile"],
): number => {
  const sections = (structuredProfile as Record<string, unknown>).sections;
  if (!Array.isArray(sections)) return 0;
  return sections.reduce<number>((total, section) => {
    if (typeof section !== "object" || section === null) return total;
    const bullets = (section as Record<string, unknown>).bullets;
    return Array.isArray(bullets) ? total + bullets.length : total;
  }, 0);
};

const countAcceptedSuggestions = (
  suggestions: ApiResumeVersion["suggestions"],
): number => {
  if (!Array.isArray(suggestions)) return 0;
  return suggestions.filter(
    (suggestion) => safeStringField(suggestion, "status") === "accepted",
  ).length;
};

export const mapResumeVersionToUi = (
  version: ApiResumeVersion,
): UiResumeVersion => {
  const sectionBullets = countSectionBullets(version.structuredProfile);
  const suggestionCount = Array.isArray(version.suggestions)
    ? version.suggestions.length
    : 0;
  return {
    id: version.id,
    originalId: version.resumeAssetId,
    parentVersionId: version.parentVersionId ?? null,
    name: version.displayName,
    tag: tagFromVersionType(version.versionType),
    date: formatDateOnly(version.updatedAt),
    target: version.focusAngle ?? null,
    bullets: Math.max(sectionBullets, suggestionCount),
    accepted: countAcceptedSuggestions(version.suggestions),
    match: formatMatchScore(version.matchScore),
    archived: version.deletedAt !== null && version.deletedAt !== undefined,
  };
};

const splitWhy = (reason: string): string[] => {
  if (!reason) return [];
  return reason
    .split(/\s*[|;；]\s*/)
    .map((part) => part.trim())
    .filter((part) => part.length > 0);
};

const normalizeBulletStatus = (status?: string): UiBulletStatus => {
  if (status === "accepted" || status === "rejected") return status;
  return "pending";
};

export const mapBulletSuggestionToUi = (
  input: ResumeSuggestionInput,
): UiBullet => ({
  id: input.id,
  section: input.section ?? "",
  original: input.originalBullet,
  rewritten: input.suggestedBullet,
  why: splitWhy(input.reason),
  status: normalizeBulletStatus(input.status),
});
