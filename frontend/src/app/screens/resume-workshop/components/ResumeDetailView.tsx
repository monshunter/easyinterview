import { useEffect, useMemo, useState, type FC } from "react";

import type { ResumeVersion } from "../../../../api/generated/types";
import { generateIdempotencyKey } from "../../../../lib/conventions/idempotency";
import { useDisplayPreferencesOptional } from "../../../display/DisplayPreferencesProvider";
import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { mapResumeAssetToUiSource, mapResumeVersionToUi } from "../adapters/resume";
import { useResumeAsset } from "../hooks/useResumeAsset";
import { useResumeVersion } from "../hooks/useResumeVersion";
import type { ResumeDetailTab } from "../params";
import { ComingSoonTab } from "./ComingSoonTab";
import { NotFoundEmptyState } from "./NotFoundEmptyState";
import { OriginalResumePreviewModal } from "./OriginalResumePreviewModal";
import { ResumePreviewTab } from "./ResumePreviewTab";
import { fireResumeWorkshopToast } from "./toast";

export interface ResumeDetailViewProps {
  versionId: string;
  initialTab: ResumeDetailTab | null;
}

const defaultTabFor = (version: ResumeVersion): ResumeDetailTab =>
  version.versionType === "structured_master" ? "preview" : "rewrites";

export const ResumeDetailView: FC<ResumeDetailViewProps> = ({
  versionId,
  initialTab,
}) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const runtime = useAppRuntimeOptional();
  const lang = useDisplayPreferencesOptional()?.lang ?? "zh";
  const versionQuery = useResumeVersion(versionId);
  const [activeTab, setActiveTab] = useState<ResumeDetailTab | null>(initialTab);
  const [originalOpen, setOriginalOpen] = useState(false);
  const sourceQuery = useResumeAsset(
    originalOpen && versionQuery.data ? versionQuery.data.resumeAssetId : null,
  );

  // When the URL omits `tab` and the version loads, fall back to the
  // versionType-based default. Reset on versionId change.
  useEffect(() => {
    setActiveTab(initialTab);
    setOriginalOpen(false);
  }, [versionId, initialTab]);

  useEffect(() => {
    if (activeTab === null && versionQuery.data) {
      setActiveTab(defaultTabFor(versionQuery.data));
    }
  }, [activeTab, versionQuery.data]);

  const onBack = () =>
    navigate({ name: "resume_versions", params: {} });

  const onExport = async () => {
    if (!runtime?.client || !versionQuery.data) return;
    try {
      const idempotencyKey = generateIdempotencyKey();
      await runtime.client.exportResumeVersion(versionQuery.data.id, {
        idempotencyKey,
        headers: { "Accept-Language": lang },
      });
      // P0 always returns 501 RESUME_EXPORT_NOT_AVAILABLE through generated
      // client (typed as ApiErrorResponse). Surface the friendly toast
      // regardless of which code is returned.
      fireResumeWorkshopToast(
        t("resumeWorkshop.detail.exportNotAvailable"),
        "warn",
      );
    } catch {
      fireResumeWorkshopToast(
        t("resumeWorkshop.detail.exportNotAvailable"),
        "warn",
      );
    }
  };

  if (versionQuery.notFound) {
    return (
      <div data-testid="resume-detail-container">
        <NotFoundEmptyState onBack={onBack} />
      </div>
    );
  }

  if (versionQuery.error) {
    return (
      <div data-testid="resume-detail-error" className="ei-screen-card">
        <p className="ei-text-body" role="alert">
          {t("resumeWorkshop.detail.error")}
        </p>
        <button
          type="button"
          className="ei-cta"
          data-testid="resume-detail-retry"
          onClick={versionQuery.retry}
        >
          {t("workspace.errors.retry")}
        </button>
      </div>
    );
  }

  if (versionQuery.loading || !versionQuery.data) {
    return (
      <div data-testid="resume-detail-container" className="ei-screen-card">
        <span className="ei-text-body" role="status">
          {t("resumeWorkshop.detail.loading")}
        </span>
      </div>
    );
  }

  const version = versionQuery.data;
  const ui = mapResumeVersionToUi(version);
  const resolvedTab: ResumeDetailTab = activeTab ?? defaultTabFor(version);
  const masterId = version.parentVersionId ?? null;

  const tabs: ResumeDetailTab[] = ["preview", "rewrites", "edit"];
  const tabLabels: Record<ResumeDetailTab, string> = {
    preview: t("resumeWorkshop.detail.tabPreview"),
    rewrites: t("resumeWorkshop.detail.tabRewrites"),
    edit: t("resumeWorkshop.detail.tabEdit"),
  };

  const versionFallbackOriginalText = (() => {
    const profile = version.structuredProfile as Record<string, unknown>;
    const summary = typeof profile.summary === "string" ? profile.summary : "";
    const headline = typeof profile.headline === "string" ? profile.headline : "";
    const lines: string[] = [];
    if (headline) lines.push(headline);
    if (summary) lines.push(summary);
    return lines;
  })();
  const originalSource = sourceQuery.data
    ? mapResumeAssetToUiSource(sourceQuery.data)
    : null;
  const originalText = sourceQuery.data
    ? originalSource && originalSource.text.length > 0
      ? originalSource.text
      : versionFallbackOriginalText
    : [];
  const originalModalState =
    originalOpen && sourceQuery.error
      ? "error"
      : originalOpen && !sourceQuery.data
        ? "loading"
        : "ready";

  return (
    <div
      data-testid="resume-detail-container"
      data-version-tag={ui.tag}
      className="ei-resume-detail"
    >
      <button
        type="button"
        data-testid="resume-detail-back"
        className="ei-resume-detail-back"
        onClick={onBack}
      >
        ← {t("resumeWorkshop.detail.back")}
      </button>

      <nav
        aria-label={t("resumeWorkshop.detail.crumbVersions")}
        data-testid="resume-detail-breadcrumb"
        className="ei-resume-detail-breadcrumb"
      >
        <span className="ei-text-label">{t("resumeWorkshop.eyebrow")}</span>
        <span className="ei-text-label">›</span>
        <span className="ei-text-label">
          {t("resumeWorkshop.detail.crumbVersions")}
        </span>
        <span className="ei-text-label">›</span>
        <span className="ei-text-label">{ui.name}</span>
      </nav>

      <header className="ei-resume-detail-header">
        <h1 className="ei-text-display">{ui.name}</h1>
        <span className="ei-text-label" data-testid="resume-detail-version-tag">
          {ui.tag}
        </span>
      </header>

      <section
        data-testid="resume-detail-branch-graph"
        className="ei-resume-detail-branch-graph"
        aria-label="version branch graph"
      >
        <div className="ei-resume-detail-branch-node">
          <span className="ei-text-label">
            {t("resumeWorkshop.detail.branchOriginal")}
          </span>
          <span className="ei-text-body">{ui.originalId}</span>
        </div>
        {masterId ? (
          <div className="ei-resume-detail-branch-node">
            <span className="ei-text-label">
              {t("resumeWorkshop.detail.branchMaster")}
            </span>
            <span className="ei-text-body">{masterId}</span>
          </div>
        ) : null}
        <div className="ei-resume-detail-branch-node ei-resume-detail-branch-node--current">
          <span className="ei-text-label">
            {t("resumeWorkshop.detail.branchCurrent")}
          </span>
          <span className="ei-text-body">{ui.name}</span>
        </div>
      </section>

      <div
        role="tablist"
        aria-label={t("resumeWorkshop.title")}
        className="ei-resume-detail-tabs"
      >
        {tabs.map((tab) => (
          <button
            key={tab}
            type="button"
            role="tab"
            data-testid={`resume-detail-tab-${tab}`}
            aria-selected={resolvedTab === tab}
            onClick={() => setActiveTab(tab)}
          >
            {tabLabels[tab]}
          </button>
        ))}
      </div>

      <div className="ei-resume-detail-tab-content">
        {resolvedTab === "preview" ? (
          <ResumePreviewTab
            version={version}
            onExport={onExport}
            onViewOriginal={() => setOriginalOpen(true)}
          />
        ) : resolvedTab === "rewrites" ? (
          <ComingSoonTab variant="rewrites" />
        ) : (
          <ComingSoonTab variant="edit" />
        )}
      </div>

      <OriginalResumePreviewModal
        open={originalOpen}
        onClose={() => setOriginalOpen(false)}
        originalText={originalText}
        contentState={originalModalState}
        onRetry={sourceQuery.retry}
        title={originalSource?.name ?? ui.name}
      />
    </div>
  );
};
