import { useEffect, useRef, useState } from "react";

import type { ResumeVersion, TargetJob } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export interface ReportContextLabels {
  /** Human-readable target line `${company} · ${title}`. Falls back to id. */
  targetLabel: string | null;
  /** Human-readable resume label. Falls back to id. */
  resumeLabel: string | null;
  /** True while any of the two underlying operations is still resolving. */
  loading: boolean;
}

interface UseReportContextDataOptions {
  targetJobId?: string;
  resumeVersionId?: string;
}

/**
 * Fetches the two label-only operations (`getTargetJob` + `getResumeVersion`)
 * required by ReportContextStrip. Each failure independently falls back to the
 * id (per plan §3.7 mapping table) so a single broken upstream never blocks
 * the report body.
 *
 * Privacy red line: this hook reads display labels only — title / companyName
 * for the target and displayName for the resume. Raw resume / JD / parsed body
 * fields must not flow into UI from here.
 */
export function useReportContextData(
  options: UseReportContextDataOptions,
): ReportContextLabels {
  const { targetJobId, resumeVersionId } = options;
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;

  const [targetLabel, setTargetLabel] = useState<string | null>(
    targetJobId ?? null,
  );
  const [resumeLabel, setResumeLabel] = useState<string | null>(
    resumeVersionId ?? null,
  );
  const [targetLoading, setTargetLoading] = useState<boolean>(
    Boolean(client && targetJobId),
  );
  const [resumeLoading, setResumeLoading] = useState<boolean>(
    Boolean(client && resumeVersionId),
  );

  const targetSeqRef = useRef(0);
  const resumeSeqRef = useRef(0);

  useEffect(() => {
    if (!client || !targetJobId) {
      setTargetLabel(targetJobId ?? null);
      setTargetLoading(false);
      return;
    }
    const seq = targetSeqRef.current + 1;
    targetSeqRef.current = seq;
    setTargetLoading(true);
    setTargetLabel(targetJobId);
    const controller = new AbortController();
    client
      .getTargetJob(targetJobId, { signal: controller.signal })
      .then((job: TargetJob) => {
        if (targetSeqRef.current !== seq) return;
        const label = buildTargetLabel(job, targetJobId);
        setTargetLabel(label);
        setTargetLoading(false);
      })
      .catch((err: unknown) => {
        if (targetSeqRef.current !== seq) return;
        if (isAbortError(err)) return;
        setTargetLabel(targetJobId);
        setTargetLoading(false);
        // Always fall back to the ID. ContextStrip is decorative — a broken
        // upstream must not bubble past the hook boundary.
      });
    return () => {
      controller.abort();
    };
  }, [client, targetJobId]);

  useEffect(() => {
    if (!client || !resumeVersionId) {
      setResumeLabel(resumeVersionId ?? null);
      setResumeLoading(false);
      return;
    }
    const seq = resumeSeqRef.current + 1;
    resumeSeqRef.current = seq;
    setResumeLoading(true);
    setResumeLabel(resumeVersionId);
    const controller = new AbortController();
    client
      .getResumeVersion(resumeVersionId, { signal: controller.signal })
      .then((resume: ResumeVersion) => {
        if (resumeSeqRef.current !== seq) return;
        const label = buildResumeLabel(resume, resumeVersionId);
        setResumeLabel(label);
        setResumeLoading(false);
      })
      .catch((err: unknown) => {
        if (resumeSeqRef.current !== seq) return;
        if (isAbortError(err)) return;
        setResumeLabel(resumeVersionId);
        setResumeLoading(false);
      });
    return () => {
      controller.abort();
    };
  }, [client, resumeVersionId]);

  return {
    targetLabel,
    resumeLabel,
    loading: targetLoading || resumeLoading,
  };
}

function buildTargetLabel(job: TargetJob, fallback: string): string {
  const bag = job as unknown as Record<string, unknown>;
  const title = readString(bag, ["title", "roleTitle"]);
  const company = readString(bag, [
    "companyName",
    "company",
    "company_name",
  ]);
  if (title && company) return `${company} · ${title}`;
  if (title) return title;
  if (company) return company;
  return fallback;
}

function buildResumeLabel(resume: ResumeVersion, fallback: string): string {
  const bag = resume as unknown as Record<string, unknown>;
  const display = readString(bag, ["displayName", "label", "name"]);
  if (display) return display;
  const versionLabel = readString(bag, ["versionLabel"]);
  if (versionLabel) return versionLabel;
  return fallback;
}

function readString(
  source: Record<string, unknown>,
  keys: readonly string[],
): string | null {
  for (const key of keys) {
    const value = source[key];
    if (typeof value === "string" && value.length > 0) return value;
  }
  return null;
}

function isAbortError(err: unknown): boolean {
  if (!err) return false;
  if (typeof err === "object" && "name" in err) {
    return (err as { name?: string }).name === "AbortError";
  }
  return false;
}
