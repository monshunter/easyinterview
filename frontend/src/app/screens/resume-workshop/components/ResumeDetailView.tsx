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
import { ResumeEditTab } from "../tabs/ResumeEditTab";
import { ResumeRewritesTab } from "../tabs/ResumeRewritesTab";
import { useResumeRewritesActions } from "../tabs/hooks/useResumeRewritesActions";
import { useUpdateResumeVersion } from "../tabs/hooks/useUpdateResumeVersion";
import {
  SuggestionDecisionError,
} from "../tabs/hooks/useTailorSuggestionDecision";
import { UpdateResumeVersionError } from "../tabs/hooks/useUpdateResumeVersion";
import {
  RequestResumeTailorError,
  useRequestResumeTailor,
} from "../tabs/hooks/useRequestResumeTailor";
import { useResumeTailorRunPolling } from "../tabs/hooks/useResumeTailorRunPolling";
import type { ReactPollingBanner } from "../tabs/ResumeRewritesTab";
import { fireResumeWorkshopToast } from "./toast";

export interface ResumeDetailViewProps {
  versionId: string;
  initialTab: ResumeDetailTab | null;
  initialTailorRunId?: string | null;
}

const defaultTabFor = (version: ResumeVersion): ResumeDetailTab =>
  version.versionType === "structured_master" ? "preview" : "rewrites";

export const ResumeDetailView: FC<ResumeDetailViewProps> = ({
  versionId,
  initialTab,
  initialTailorRunId = null,
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
          <ResumeRewritesTabContainer
            version={version}
            initialTailorRunId={initialTailorRunId}
            onVersionRefreshed={versionQuery.retry}
          />
        ) : (
          <ResumeEditTabContainer
            version={version}
            onVersionRefreshed={versionQuery.retry}
          />
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

interface ResumeRewritesTabContainerProps {
  version: ResumeVersion;
  initialTailorRunId: string | null;
  onVersionRefreshed: () => void;
}

const ResumeRewritesTabContainer: FC<ResumeRewritesTabContainerProps> = ({
  version,
  initialTailorRunId,
  onVersionRefreshed,
}) => {
  const { t } = useI18n();
  const actions = useResumeRewritesActions({
    version,
    onVersionRefreshed,
  });

  const [activeTailorRunId, setActiveTailorRunId] = useState<string | null>(
    initialTailorRunId,
  );
  useEffect(() => {
    setActiveTailorRunId(initialTailorRunId);
  }, [initialTailorRunId, version.id]);

  const tailorRequest = useRequestResumeTailor();
  const polling = useResumeTailorRunPolling(activeTailorRunId, {
    onReady: () => onVersionRefreshed(),
  });

  const handleRerun = async (
    mode: "bullet_suggestions" | "gap_review",
  ) => {
    if (!version.targetJobId) {
      fireResumeWorkshopToast(
        t("resumeWorkshop.rewrites.error.generic"),
        "warn",
      );
      return;
    }
    try {
      const result = await tailorRequest.request({
        resumeAssetId: version.resumeAssetId,
        targetJobId: version.targetJobId,
        mode,
      });
      setActiveTailorRunId(result.tailorRunId);
      fireResumeWorkshopToast(
        t("resumeWorkshop.rewrites.toast.rerunRequested"),
        "ok",
      );
    } catch (err) {
      if (err instanceof RequestResumeTailorError && err.kind === "validation") {
        fireResumeWorkshopToast(
          t("resumeWorkshop.rewrites.error.validation"),
          "warn",
        );
      } else if (
        err instanceof RequestResumeTailorError &&
        err.kind === "cross_user"
      ) {
        fireResumeWorkshopToast(
          t("resumeWorkshop.rewrites.error.crossUser"),
          "danger",
        );
      } else {
        fireResumeWorkshopToast(
          t("resumeWorkshop.rewrites.error.generic"),
          "danger",
        );
      }
    }
  };

  const pollingBanner: ReactPollingBanner | null =
    polling.phase === "polling"
      ? { kind: "info", message: t("resumeWorkshop.rewrites.polling.banner") }
      : polling.phase === "failed"
        ? {
            kind: "danger",
            message: t("resumeWorkshop.rewrites.polling.failed"),
            onRetry: () => polling.retry(),
          }
        : polling.phase === "timeout"
          ? {
              kind: "danger",
              message: t("resumeWorkshop.rewrites.polling.timeout"),
              onRetry: () => polling.retry(),
            }
          : polling.phase === "error"
            ? {
                kind: "danger",
                message: t("resumeWorkshop.rewrites.polling.failed"),
                onRetry: () => polling.retry(),
              }
            : null;

  const showError = (err: unknown) => {
    if (err instanceof SuggestionDecisionError) {
      if (err.kind === "already_decided") {
        fireResumeWorkshopToast(
          t("resumeWorkshop.rewrites.toast.alreadyDecided"),
          "warn",
        );
        return;
      }
      if (err.kind === "cross_user") {
        fireResumeWorkshopToast(
          t("resumeWorkshop.rewrites.error.crossUser"),
          "danger",
        );
        return;
      }
      if (err.kind === "validation") {
        fireResumeWorkshopToast(
          t("resumeWorkshop.rewrites.error.validation"),
          "warn",
        );
        return;
      }
    }
    if (err instanceof UpdateResumeVersionError) {
      if (err.kind === "validation") {
        fireResumeWorkshopToast(
          t("resumeWorkshop.rewrites.error.validation"),
          "warn",
        );
        return;
      }
      if (err.kind === "cross_user") {
        fireResumeWorkshopToast(
          t("resumeWorkshop.rewrites.error.crossUser"),
          "danger",
        );
        return;
      }
    }
    fireResumeWorkshopToast(
      t("resumeWorkshop.rewrites.error.generic"),
      "danger",
    );
  };

  const handleAccept = async (id: string) => {
    try {
      await actions.onAccept(id);
      fireResumeWorkshopToast(
        t("resumeWorkshop.rewrites.toast.accept").replace(
          "{versionName}",
          version.displayName,
        ),
        "ok",
      );
    } catch (err) {
      showError(err);
    }
  };

  const handleReject = async (id: string) => {
    try {
      await actions.onReject(id);
      fireResumeWorkshopToast(
        t("resumeWorkshop.rewrites.toast.reject").replace(
          "{versionName}",
          version.displayName,
        ),
        "ok",
      );
    } catch (err) {
      showError(err);
    }
  };

  const handleManual = async (id: string, text: string) => {
    try {
      await actions.onSaveManualEdit(id, text);
      fireResumeWorkshopToast(
        t("resumeWorkshop.rewrites.toast.manualSaved").replace(
          "{versionName}",
          version.displayName,
        ),
        "ok",
      );
    } catch (err) {
      showError(err);
    }
  };

  return (
    <ResumeRewritesTab
      version={version}
      onAccept={handleAccept}
      onReject={handleReject}
      onSaveManualEdit={handleManual}
      onRequestRerun={handleRerun}
      manualEditPendingFor={actions.manualPendingFor}
      pollingBanner={pollingBanner}
    />
  );
};

interface ResumeEditTabContainerProps {
  version: ResumeVersion;
  onVersionRefreshed: () => void;
}

const ResumeEditTabContainer: FC<ResumeEditTabContainerProps> = ({
  version,
  onVersionRefreshed,
}) => {
  const { t } = useI18n();
  const updater = useUpdateResumeVersion();
  const saving = !!updater.pendingFor[version.id];
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const onSave = async ({
    headline,
    summary,
  }: {
    headline: string;
    summary: string;
  }) => {
    setErrorMessage(null);
    const existing = (version.structuredProfile ?? {}) as Record<string, unknown>;
    const nextProfile = {
      ...existing,
      headline,
      summary,
    };
    try {
      await updater.update({
        versionId: version.id,
        payload: { structuredProfile: nextProfile },
      });
      fireResumeWorkshopToast(
        t("resumeWorkshop.edit.toast.saved").replace(
          "{versionName}",
          version.displayName,
        ),
        "ok",
      );
      onVersionRefreshed();
    } catch (err) {
      if (err instanceof UpdateResumeVersionError) {
        if (err.kind === "validation") {
          setErrorMessage(t("resumeWorkshop.edit.error.validation"));
          return;
        }
        if (err.kind === "idempotency_conflict") {
          setErrorMessage(t("resumeWorkshop.edit.error.idempotency"));
          return;
        }
        if (err.kind === "cross_user") {
          setErrorMessage(t("resumeWorkshop.edit.error.crossUser"));
          return;
        }
      }
      setErrorMessage(t("resumeWorkshop.edit.error.generic"));
    }
  };

  return (
    <ResumeEditTab
      version={version}
      onSave={onSave}
      saving={saving}
      errorMessage={errorMessage}
    />
  );
};
