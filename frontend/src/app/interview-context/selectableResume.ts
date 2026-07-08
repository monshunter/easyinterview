import type { Resume } from "../../api/generated/types";

function hasText(value: unknown): boolean {
  return typeof value === "string" && value.trim().length > 0;
}

function hasStructuredProfile(value: unknown): boolean {
  if (!value || typeof value !== "object" || Array.isArray(value)) return false;
  return Object.keys(value as Record<string, unknown>).length > 0;
}

export function hasReadableResumeEvidence(resume: Resume): boolean {
  return (
    hasText(resume.parsedTextSnapshot) ||
    hasText(resume.originalText) ||
    hasStructuredProfile(resume.structuredProfile)
  );
}

export function isSelectableInterviewResume(resume: Resume): boolean {
  if (resume.status === "archived" || resume.deletedAt) return false;
  return resume.parseStatus === "ready" || hasReadableResumeEvidence(resume);
}
