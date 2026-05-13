import { useCallback, useMemo, useState, type FC } from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import {
  mapResumeAssetToUiSource,
  mapResumeVersionToUi,
  type UiResumeSource,
  type UiResumeVersion,
} from "../adapters/resume";
import { useResumeAssets } from "../hooks/useResumeAssets";
import { useResumeVersions } from "../hooks/useResumeVersions";
import { ResumeFlatView } from "./ResumeFlatView";
import { ResumeWorkshopIcon } from "./ResumeWorkshopIcon";
import { ResumeTreeView } from "./ResumeTreeView";

type GroupBy = "tree" | "flat";

export const ResumeListView: FC = () => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const assetsQuery = useResumeAssets();
  const primerAssetId = assetsQuery.data?.items[0]?.id ?? null;
  const versionsQuery = useResumeVersions(primerAssetId);
  const [groupBy, setGroupBy] = useState<GroupBy>("tree");
  const [selectedTreeId, setSelectedTreeId] = useState<string | null>(null);

  const sources = useMemo<UiResumeSource[]>(
    () => assetsQuery.data?.items.map(mapResumeAssetToUiSource) ?? [],
    [assetsQuery.data],
  );
  const versions = useMemo<UiResumeVersion[]>(
    () => versionsQuery.data?.items.map(mapResumeVersionToUi) ?? [],
    [versionsQuery.data],
  );

  const versionsByAsset = useMemo(() => {
    const map = new Map<string, UiResumeVersion[]>();
    for (const version of versions) {
      const list = map.get(version.originalId) ?? [];
      list.push(version);
      map.set(version.originalId, list);
    }
    return map;
  }, [versions]);

  const stats = useMemo(() => {
    const totalOriginals = sources.length;
    const activeOriginals = sources.filter((source) => source.status === "active").length;
    const totalVersions = versions.length;
    const targetedVersions = versions.filter((v) => v.tag === "TARGETED").length;
    const matches = versions
      .map((v) => v.match)
      .filter((value): value is number => value !== null);
    const topMatch = matches.length > 0 ? Math.max(...matches) : null;
    const topMatchVersion =
      topMatch === null ? null : versions.find((v) => v.match === topMatch) ?? null;
    const recentVersion = [...versions].sort((a, b) =>
      b.date.localeCompare(a.date),
    )[0] ?? null;
    return {
      activeOriginals,
      totalOriginals,
      totalVersions,
      targetedVersions,
      topMatch,
      topMatchLabel: topMatchVersion?.target ?? topMatchVersion?.name ?? null,
      recent: recentVersion?.date ?? null,
      recentLabel:
        recentVersion === null
          ? null
          : `v · ${recentVersion.target ?? recentVersion.name}`,
    };
  }, [sources, versions]);

  const selectedTree = useMemo(
    () => sources.find((source) => source.id === selectedTreeId) ?? null,
    [selectedTreeId, sources],
  );

  const onOpenVersion = useCallback(
    (version: UiResumeVersion) => {
      navigate({
        name: "resume_versions",
        params: {
          versionId: version.id,
          tab: version.tag === "TARGETED" ? "rewrites" : "preview",
        },
      });
    },
    [navigate],
  );
  const onCreate = useCallback(() => {
    navigate({
      name: "resume_versions",
      params: { flow: "create" },
    });
  }, [navigate]);

  const onBranch = useCallback(
    (sourceId: string) => {
      navigate({
        name: "resume_versions",
        params: { flow: "branch", branchOriginalId: sourceId },
      });
    },
    [navigate],
  );

  if (assetsQuery.loading) {
    return (
      <div data-testid="resume-workshop-list" className="ei-screen-card">
        <span className="ei-text-body" role="status">
          {t("resumeWorkshop.list.loading")}
        </span>
      </div>
    );
  }

  if (assetsQuery.error) {
    return (
      <div data-testid="resume-workshop-list" className="ei-screen-card">
        <p className="ei-text-body" role="alert">
          {t("resumeWorkshop.list.error")}
        </p>
        <button
          type="button"
          className="ei-cta"
          data-testid="resume-workshop-list-retry"
          onClick={assetsQuery.retry}
        >
          {t("workspace.errors.retry")}
        </button>
      </div>
    );
  }

  const versionsPending =
    sources.length > 0 &&
    primerAssetId !== null &&
    !versionsQuery.data &&
    !versionsQuery.error;

  if (versionsPending || (versionsQuery.loading && sources.length > 0)) {
    return (
      <div data-testid="resume-workshop-list" className="ei-screen-card">
        <span className="ei-text-body" role="status">
          {t("resumeWorkshop.list.loading")}
        </span>
      </div>
    );
  }

  if (versionsQuery.error) {
    return (
      <div
        data-testid="resume-workshop-versions-error"
        className="ei-screen-card"
      >
        <p className="ei-text-body" role="alert">
          {t("resumeWorkshop.list.versionsError")}
        </p>
        <button
          type="button"
          className="ei-cta"
          data-testid="resume-workshop-versions-retry"
          onClick={versionsQuery.retry}
        >
          {t("workspace.errors.retry")}
        </button>
      </div>
    );
  }

  return (
    <div
      data-testid="resume-workshop-list"
      data-group-by={groupBy}
      className="ei-resume-workshop-list"
    >
      <header className="ei-resume-workshop-list-header">
        <div>
          <span className="ei-text-label">{t("resumeWorkshop.eyebrow")}</span>
          <h1 className="ei-text-display">{t("resumeWorkshop.title")}</h1>
          <p className="ei-text-body">{t("resumeWorkshop.subtitle")}</p>
        </div>
        <button
          type="button"
          className="ei-resume-workshop-create"
          data-testid="resume-workshop-create"
          onClick={onCreate}
        >
          <ResumeWorkshopIcon name="plus" size={14} />
          {t("resumeWorkshop.create")}
        </button>
      </header>

      <StatsStrip
        activeOriginals={stats.activeOriginals}
        totalOriginals={stats.totalOriginals}
        totalVersions={stats.totalVersions}
        targetedVersions={stats.targetedVersions}
        topMatch={stats.topMatch}
        topMatchLabel={stats.topMatchLabel}
        recent={stats.recent}
        recentLabel={stats.recentLabel}
        translate={t}
      />

      <div className="ei-resume-workshop-view-row">
        <ViewSwitcher value={groupBy} onChange={setGroupBy} translate={t} />
        <div
          className="ei-resume-workshop-view-context"
          data-testid="resume-workshop-view-context"
        >
          {groupBy === "tree"
            ? t("resumeWorkshop.viewContext.tree")
            : t("resumeWorkshop.viewContext.flat")}
        </div>
      </div>

      {groupBy === "tree" && selectedTree ? (
        <div
          className="ei-resume-workshop-selected-tree"
          data-testid="resume-workshop-selected-tree-helper"
        >
          <div className="ei-resume-workshop-selected-tree-copy">
            <ResumeWorkshopIcon name="check" size={12} />
            <span>{t("resumeWorkshop.tree.selectedHelper")}</span>
            <span className="ei-resume-workshop-selected-tree-name">
              {selectedTree.name}
            </span>
          </div>
          <div className="ei-resume-workshop-selected-tree-actions">
            <button
              type="button"
              data-testid="resume-workshop-selected-tree-clear"
              onClick={() => setSelectedTreeId(null)}
            >
              {t("resumeWorkshop.tree.clearSelection")}
            </button>
            <button
              type="button"
              data-testid="resume-workshop-selected-tree-branch"
              onClick={() => onBranch(selectedTree.id)}
            >
              <ResumeWorkshopIcon name="plus" size={12} />
              {t("resumeWorkshop.tree.newVersion")}
            </button>
          </div>
        </div>
      ) : null}

      {sources.length === 0 ? (
        <div
          data-testid="resume-workshop-list-empty"
          className="ei-screen-card"
        >
          <p className="ei-text-body">{t("resumeWorkshop.list.empty")}</p>
        </div>
      ) : groupBy === "tree" ? (
        <ResumeTreeView
          sources={sources}
          versionsByAsset={versionsByAsset}
          onOpenVersion={onOpenVersion}
          selectedSourceId={selectedTreeId}
          onSelectSource={setSelectedTreeId}
          onCreate={onCreate}
        />
      ) : (
        <ResumeFlatView
          sources={sources}
          versions={versions}
          onOpenVersion={onOpenVersion}
        />
      )}

      {assetsQuery.data?.pageInfo.hasMore ? (
        <div
          data-testid="resume-workshop-list-paginated"
          className="ei-text-body"
        >
          {t("resumeWorkshop.list.paginated")}
        </div>
      ) : null}
    </div>
  );
};

interface StatsStripProps {
  activeOriginals: number;
  totalOriginals: number;
  totalVersions: number;
  targetedVersions: number;
  topMatch: number | null;
  topMatchLabel: string | null;
  recent: string | null;
  recentLabel: string | null;
  translate: (key: MessageKey) => string;
}

const StatsStrip: FC<StatsStripProps> = ({
  activeOriginals,
  totalOriginals,
  totalVersions,
  targetedVersions,
  topMatch,
  topMatchLabel,
  recent,
  recentLabel,
  translate,
}) => (
  <ul className="ei-resume-workshop-stats">
    <li
      data-testid="resume-workshop-stats-originals"
      className="ei-resume-workshop-stat"
    >
      <span className="ei-text-label">
        {translate("resumeWorkshop.stats.originals")}
      </span>
      <span className="ei-text-title">{`${activeOriginals} / ${totalOriginals}`}</span>
      <span className="ei-resume-workshop-stat-sub">
        {translate("resumeWorkshop.stats.originalsSub")}
      </span>
    </li>
    <li
      data-testid="resume-workshop-stats-versions"
      className="ei-resume-workshop-stat"
    >
      <span className="ei-text-label">
        {translate("resumeWorkshop.stats.versions")}
      </span>
      <span className="ei-text-title">{totalVersions}</span>
      <span className="ei-resume-workshop-stat-sub">
        {translate("resumeWorkshop.stats.versionsSub").replace(
          "{count}",
          String(targetedVersions),
        )}
      </span>
    </li>
    <li
      data-testid="resume-workshop-stats-top-match"
      className="ei-resume-workshop-stat"
    >
      <span className="ei-text-label">
        {translate("resumeWorkshop.stats.topMatch")}
      </span>
      <span className="ei-text-title">
        {topMatch !== null
          ? `${topMatch}%`
          : translate("resumeWorkshop.stats.empty")}
      </span>
      <span className="ei-resume-workshop-stat-sub">
        {topMatchLabel ?? translate("resumeWorkshop.stats.empty")}
      </span>
    </li>
    <li
      data-testid="resume-workshop-stats-recent"
      className="ei-resume-workshop-stat"
    >
      <span className="ei-text-label">
        {translate("resumeWorkshop.stats.recent")}
      </span>
      <span className="ei-text-title">
        {recent ?? translate("resumeWorkshop.stats.empty")}
      </span>
      <span className="ei-resume-workshop-stat-sub">
        {recentLabel ?? translate("resumeWorkshop.stats.empty")}
      </span>
    </li>
  </ul>
);

interface ViewSwitcherProps {
  value: GroupBy;
  onChange: (next: GroupBy) => void;
  translate: (key: MessageKey) => string;
}

const ViewSwitcher: FC<ViewSwitcherProps> = ({ value, onChange, translate }) => (
  <div
    role="tablist"
    aria-label={translate("resumeWorkshop.eyebrow")}
    className="ei-resume-workshop-view-switcher"
  >
    <button
      type="button"
      data-testid="resume-workshop-view-switcher-tree"
      data-active={value === "tree" ? "true" : "false"}
      role="tab"
      aria-selected={value === "tree"}
      onClick={() => onChange("tree")}
    >
      <ResumeWorkshopIcon name="file" size={12} />
      {translate("resumeWorkshop.viewSwitcher.tree")}
    </button>
    <button
      type="button"
      data-testid="resume-workshop-view-switcher-flat"
      data-active={value === "flat" ? "true" : "false"}
      role="tab"
      aria-selected={value === "flat"}
      onClick={() => onChange("flat")}
    >
      <ResumeWorkshopIcon name="layers" size={12} />
      {translate("resumeWorkshop.viewSwitcher.flat")}
    </button>
  </div>
);
