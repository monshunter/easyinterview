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
import { ResumeTreeView } from "./ResumeTreeView";

type GroupBy = "tree" | "flat";

export const ResumeListView: FC = () => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const assetsQuery = useResumeAssets();
  const primerAssetId = assetsQuery.data?.items[0]?.id ?? null;
  const versionsQuery = useResumeVersions(primerAssetId);
  const [groupBy, setGroupBy] = useState<GroupBy>("tree");

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
    const totalVersions = versions.length;
    const matches = versions
      .map((v) => v.match)
      .filter((value): value is number => value !== null);
    const topMatch = matches.length > 0 ? Math.max(...matches) : null;
    const recent = versions
      .map((v) => v.date)
      .sort((a, b) => b.localeCompare(a))[0];
    return {
      totalOriginals,
      totalVersions,
      topMatch,
      recent: recent ?? null,
    };
  }, [sources, versions]);

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
      </header>

      <StatsStrip
        totalOriginals={stats.totalOriginals}
        totalVersions={stats.totalVersions}
        topMatch={stats.topMatch}
        recent={stats.recent}
        translate={t}
      />

      <ViewSwitcher value={groupBy} onChange={setGroupBy} translate={t} />

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
  totalOriginals: number;
  totalVersions: number;
  topMatch: number | null;
  recent: string | null;
  translate: (key: MessageKey) => string;
}

const StatsStrip: FC<StatsStripProps> = ({
  totalOriginals,
  totalVersions,
  topMatch,
  recent,
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
      <span className="ei-text-title">{totalOriginals}</span>
    </li>
    <li
      data-testid="resume-workshop-stats-versions"
      className="ei-resume-workshop-stat"
    >
      <span className="ei-text-label">
        {translate("resumeWorkshop.stats.versions")}
      </span>
      <span className="ei-text-title">{totalVersions}</span>
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
      {translate("resumeWorkshop.viewSwitcher.flat")}
    </button>
  </div>
);
