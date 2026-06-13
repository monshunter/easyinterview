import type { Resume } from "../../../../../api/generated/types";
import type { PreviewDraft } from "../ResumePreviewConfirm";

interface ParsedSummaryShape {
  identity?: {
    name?: string;
    title?: string;
    headline?: string;
    location?: string;
    contact?: string[] | string;
    email?: string;
    phone?: string;
  };
  summary?: string;
  experience?: Array<{
    co?: string;
    company?: string;
    role?: string;
    title?: string;
    period?: string;
    dates?: string;
    bullets?: string[];
    highlights?: string[];
  }>;
  projects?: Array<{
    name?: string;
    title?: string;
    note?: string;
    impact?: string;
  }>;
  skills?: string[];
  education?: Array<{
    school?: string;
    institution?: string;
    degree?: string;
    summary?: string;
  }>;
}

function safeArray<T>(value: T[] | undefined | null): T[] {
  return Array.isArray(value) ? value : [];
}

function toContactList(value: unknown): string[] {
  if (Array.isArray(value)) return value.filter((v): v is string => typeof v === "string");
  if (typeof value === "string") return [value];
  return [];
}

export function mapParsedSummaryToStructuredProfileDraft(
  asset: Resume,
): PreviewDraft {
  const parsed = (asset.parsedSummary ?? {}) as ParsedSummaryShape;
  const identity = parsed.identity ?? {};
  const contactCandidates: unknown[] = [identity.contact, identity.email, identity.phone];
  const contact: string[] = [];
  for (const value of contactCandidates) {
    for (const entry of toContactList(value)) {
      if (entry && !contact.includes(entry)) {
        contact.push(entry);
      }
    }
  }

  return {
    name: identity.name ?? asset.title,
    title: identity.title ?? identity.headline,
    location: identity.location,
    contact,
    summary: parsed.summary,
    experience: safeArray(parsed.experience).map((entry) => ({
      co: entry.co ?? entry.company ?? "",
      role: entry.role ?? entry.title ?? "",
      period: entry.period ?? entry.dates ?? "",
      bullets: safeArray(entry.bullets ?? entry.highlights),
    })),
    projects: safeArray(parsed.projects).map((entry) => ({
      name: entry.name ?? entry.title ?? "",
      note: entry.note ?? entry.impact,
    })),
    skills: safeArray(parsed.skills),
    education: safeArray(parsed.education).map((entry) => ({
      school: entry.school ?? entry.institution ?? "",
      degree: entry.degree ?? entry.summary ?? "",
    })),
  };
}

export function buildStructuredProfilePayload(
  asset: Resume,
): Record<string, unknown> {
  const draft = mapParsedSummaryToStructuredProfileDraft(asset);
  const language = asset.language?.trim() || "en";
  return {
    headline: draft.title ?? draft.name,
    summary: draft.summary ?? "",
    identity: {
      name: draft.name,
      title: draft.title,
      location: draft.location,
      contact: draft.contact,
    },
    experience: draft.experience,
    projects: draft.projects,
    skills: draft.skills,
    education: draft.education,
    sections: [],
    provenance: {
      promptVersion: "resume_profile.v1",
      rubricVersion: "not_applicable",
      modelId: "resume-profile.confirmed.v1",
      language,
      featureFlag: "resume-workshop-additive",
      dataSourceVersion: "resume_asset.v1",
    },
  };
}
