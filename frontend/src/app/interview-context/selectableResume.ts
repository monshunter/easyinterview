import type { Resume, ResumeSummary } from "../../api/generated/types";

function hasText(value: unknown): boolean {
  return typeof value === "string" && value.trim().length > 0;
}

function hasStructuredProfile(value: unknown): boolean {
  if (!value || typeof value !== "object" || Array.isArray(value)) return false;
  return Object.keys(value as Record<string, unknown>).length > 0;
}

export function hasReadableResumeEvidence(
  resume: Resume | ResumeSummary,
): boolean {
  if ("hasReadableContent" in resume) return resume.hasReadableContent;
  return (
    hasText(resume.parsedTextSnapshot) ||
    hasText(resume.originalText) ||
    hasStructuredProfile(resume.structuredProfile)
  );
}

export function isSelectableInterviewResume(
  resume: Resume | ResumeSummary,
): boolean {
  if ("status" in resume && resume.status === "archived") return false;
  if ("deletedAt" in resume && resume.deletedAt) return false;
  return resume.parseStatus === "ready" || hasReadableResumeEvidence(resume);
}
