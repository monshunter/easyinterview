import { useCallback, useMemo, useState, type FC } from "react";

import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useI18n } from "../../../i18n/messages";
import { DebriefContextPickerModal, type PickerOption } from "./DebriefContextPickerModal";
import { usePickerOptions } from "../hooks/usePickerOptions";
import type { ResumeAsset, ResumeVersion } from "../types";

interface ResumePickerProps {
  selectedAssetId: string | null;
  selectedVersionId: string | null;
  onClose: () => void;
  onConfirm: (selection: {
    asset: ResumeAsset | null;
    version: ResumeVersion | null;
  }) => void;
}

/**
 * Phase 2.4 Resume picker. Two-step selection: first the asset, then a
 * version under that asset. Both lists come from the generated client
 * (`listResumes` → `listResumeVersions(assetId)`); only `parseStatus==='ready'`
 * assets and ready resume versions are listed per spec §3.2.
 */
export const ResumePicker: FC<ResumePickerProps> = ({
  selectedAssetId,
  selectedVersionId,
  onClose,
  onConfirm,
}) => {
  const runtime = useAppRuntimeOptional();
  const { t } = useI18n();
  const [assetId, setAssetId] = useState<string | null>(selectedAssetId);
  const [assetMap, setAssetMap] = useState<Record<string, ResumeAsset>>({});

  const loadAssets = useCallback(async () => {
    if (!runtime) return { options: [] as PickerOption<ResumeAsset>[] };
    const res = await runtime.client.listResumes();
    const items = res.items.filter(
      (a) =>
        (a.status === undefined || a.status === "active") &&
        a.parseStatus === "ready",
    );
    const map: Record<string, ResumeAsset> = {};
    items.forEach((a) => (map[a.id] = a));
    setAssetMap(map);
    const options = items.map<PickerOption<ResumeAsset>>((asset) => ({
      id: asset.id,
      title: asset.title,
      meta: `${asset.language}`,
      value: asset,
    }));
    return { options };
  }, [runtime]);

  const loadVersions = useCallback(async () => {
    if (!runtime || !assetId) {
      return { options: [] as PickerOption<ResumeVersion>[] };
    }
    const res = await runtime.client.listResumeVersions(assetId);
    const items = res.items;
    const options = items.map<PickerOption<ResumeVersion>>((version) => ({
      id: version.id,
      title: version.displayName,
      meta: `${version.versionType}${version.focusAngle ? ` · ${version.focusAngle}` : ""}`,
      value: version,
    }));
    return { options };
  }, [runtime, assetId]);

  const assetState = usePickerOptions<ResumeAsset>({
    enabled: Boolean(runtime) && assetId === null,
    load: loadAssets,
  });
  const versionState = usePickerOptions<ResumeVersion>({
    enabled: Boolean(runtime) && assetId !== null,
    load: loadVersions,
  });

  const phase = useMemo<"asset" | "version">(
    () => (assetId === null ? "asset" : "version"),
    [assetId],
  );

  if (phase === "asset") {
    return (
      <DebriefContextPickerModal
        kind="resume"
        options={assetState.options}
        selectedId={selectedAssetId}
        loading={assetState.loading}
        errorMessage={assetState.error}
        banner={
          <span data-testid="debrief-picker-banner-resume-phase">
            {t("debrief.picker.resume.assetPhase")}
          </span>
        }
        onClose={onClose}
        onConfirm={(opt) => {
          if (opt) setAssetId(opt.id);
        }}
      />
    );
  }
  return (
    <DebriefContextPickerModal
      kind="resume"
      options={versionState.options}
      selectedId={selectedVersionId}
      loading={versionState.loading}
      errorMessage={versionState.error}
      banner={
        <span data-testid="debrief-picker-banner-resume-phase">
          {t("debrief.picker.resume.versionPhase")}
        </span>
      }
      onClose={() => {
        setAssetId(null);
        onClose();
      }}
      onConfirm={(opt) => {
        if (!opt) return;
        const asset = assetId ? (assetMap[assetId] ?? null) : null;
        onConfirm({ asset, version: opt.value });
      }}
    />
  );
};
