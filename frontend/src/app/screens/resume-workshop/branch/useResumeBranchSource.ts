import { useMemo } from "react";

import type {
  ResumeAsset,
  ResumeVersion,
} from "../../../../api/generated/types";
import { useResumeAssets } from "../hooks/useResumeAssets";
import { useResumeVersions } from "../hooks/useResumeVersions";

export type ResumeBranchSourceStatus =
  | "missing-id"
  | "loading"
  | "ready"
  | "not-found"
  | "error";

export interface ResumeBranchSource {
  status: ResumeBranchSourceStatus;
  original: ResumeAsset | null;
  master: ResumeVersion | null;
  error: Error | null;
  retry: () => void;
}

const findMaster = (
  versions: ResumeVersion[],
  originalId: string,
): ResumeVersion | null => {
  for (const version of versions) {
    if (version.resumeAssetId !== originalId) continue;
    // `structured_master` is the canonical MASTER version per shared
    // conventions §5.14 (ResumeVersionType). Adapter `tagFromVersionType`
    // already maps this enum to the UI `MASTER` tag.
    if (version.versionType === "structured_master") return version;
  }
  return null;
};

export function useResumeBranchSource(
  branchOriginalId: string | null,
): ResumeBranchSource {
  const assetsQuery = useResumeAssets();
  const versionsQuery = useResumeVersions(branchOriginalId);

  return useMemo<ResumeBranchSource>(() => {
    const retry = () => {
      assetsQuery.retry();
      versionsQuery.retry();
    };

    if (!branchOriginalId) {
      return {
        status: "missing-id",
        original: null,
        master: null,
        error: null,
        retry,
      };
    }

    if (assetsQuery.error) {
      return {
        status: "error",
        original: null,
        master: null,
        error: assetsQuery.error,
        retry,
      };
    }

    if (versionsQuery.error) {
      return {
        status: "error",
        original: null,
        master: null,
        error: versionsQuery.error,
        retry,
      };
    }

    if (assetsQuery.loading || versionsQuery.loading) {
      return {
        status: "loading",
        original: null,
        master: null,
        error: null,
        retry,
      };
    }

    const original =
      assetsQuery.data?.items.find((asset) => asset.id === branchOriginalId) ??
      null;
    if (!original) {
      return {
        status: "not-found",
        original: null,
        master: null,
        error: null,
        retry,
      };
    }

    const master = findMaster(
      versionsQuery.data?.items ?? [],
      branchOriginalId,
    );
    if (!master) {
      return {
        status: "not-found",
        original,
        master: null,
        error: null,
        retry,
      };
    }

    return {
      status: "ready",
      original,
      master,
      error: null,
      retry,
    };
  }, [
    branchOriginalId,
    assetsQuery.data,
    assetsQuery.loading,
    assetsQuery.error,
    assetsQuery.retry,
    versionsQuery.data,
    versionsQuery.loading,
    versionsQuery.error,
    versionsQuery.retry,
  ]);
}
