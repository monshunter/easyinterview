import { useEffect, useState, type FC } from "react";

import type { Resume } from "../../../../api/generated/types";
import { generateIdempotencyKey } from "../../../../lib/conventions/idempotency";
import { useDisplayPreferencesOptional } from "../../../display/DisplayPreferencesProvider";
import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import {
  buildResumePlainText,
  mapResumeToUiSource,
} from "../adapters/resume";
import { useResumeAsset } from "../hooks/useResumeAsset";
import { useResumeSave, ResumeSaveError } from "../hooks/useResumeSave";
import type { ResumeDetailTab } from "../params";
import { NotFoundEmptyState } from "./NotFoundEmptyState";
import { OriginalResumePreviewModal } from "./OriginalResumePreviewModal";
import { ResumePreviewTab } from "./ResumePreviewTab";
import { ResumeEditTab } from "../tabs/ResumeEditTab";
import { ResumeRewritesTab } from "../tabs/ResumeRewritesTab";
import {
  RequestResumeTailorError,
  useRequestResumeTailor,
} from "../tabs/hooks/useRequestResumeTailor";
import { useResumeTailorRunPolling } from "../tabs/hooks/useResumeTailorRunPolling";
import type { ReactPollingBanner } from "../tabs/ResumeRewritesTab";
import { fireResumeWorkshopToast } from "./toast";

export interface ResumeDetailViewProps {
  resumeId: string;
  initialTab: ResumeDetailTab | null;
  initialTailorRunId?: string | null;
}

export const ResumeDetailView: FC<ResumeDetailViewProps> = ({
  resumeId,
  initialTab,
  initialTailorRunId = null,
}) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const runtime = useAppRuntimeOptional();
  const lang = useDisplayPreferencesOptional()?.lang ?? "zh";
  const resumeQuery = useResumeAsset(resumeId);
  const [activeTab, setActiveTab] = useState<ResumeDetailTab>(
    initialTab ?? "preview",
  );
  const [originalOpen, setOriginalOpen] = useState(false);

  useEffect(() => {
    setActiveTab(initialTab ?? "preview");
    setOriginalOpen(false);
  }, [resumeId, initialTab]);

  const onBack = () => navigate({ name: "resume_versions", params: {} });

  const onExport = async () => {
    if (!runtime?.client || !resumeQuery.data) return;
    try {
      const idempotencyKey = generateIdempotencyKey();
      await runtime.client.exportResume(resumeQuery.data.id, {
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

  const onCopy = async () => {
    if (!resumeQuery.data) return;
    const text = buildResumePlainText(resumeQuery.data);
    if (window.navigator.clipboard?.writeText) {
      try {
        await window.navigator.clipboard.writeText(text);
        fireResumeWorkshopToast(t("resumeWorkshop.detail.copySuccess"), "ok");
        return;
      } catch {
        // fall through to the same unavailable toast used by Preview.
      }
    }
    fireResumeWorkshopToast(t("resumeWorkshop.detail.copyUnavailable"), "warn");
  };

  if (resumeQuery.notFound) {
    return (
      <div data-testid="resume-detail-container">
        <NotFoundEmptyState onBack={onBack} />
      </div>
    );
  }

  if (resumeQuery.error) {
    return (
      <div data-testid="resume-detail-error" className="ei-screen-card">
        <p className="ei-text-body" role="alert">
          {t("resumeWorkshop.detail.error")}
        </p>
        <button
          type="button"
          className="ei-cta"
          data-testid="resume-detail-retry"
          onClick={resumeQuery.retry}
        >
          {t("workspace.errors.retry")}
        </button>
      </div>
    );
  }

  if (resumeQuery.loading || !resumeQuery.data) {
    return (
      <div data-testid="resume-detail-container" className="ei-screen-card">
        <span className="ei-text-body" role="status">
          {t("resumeWorkshop.detail.loading")}
        </span>
      </div>
    );
  }

  const resume = resumeQuery.data;
  const ui = mapResumeToUiSource(resume);

  const tabs: ResumeDetailTab[] = ["preview", "rewrites", "edit"];
  const tabLabels: Record<ResumeDetailTab, string> = {
    preview: t("resumeWorkshop.detail.tabPreview"),
    rewrites: t("resumeWorkshop.detail.tabRewrites"),
    edit: t("resumeWorkshop.detail.tabEdit"),
  };

  const originalText =
    ui.text.length > 0
      ? ui.text
      : (() => {
          const profile = resume.structuredProfile as Record<string, unknown>;
          const headline =
            typeof profile.headline === "string" ? profile.headline : "";
          const summary =
            typeof profile.summary === "string" ? profile.summary : "";
          const lines: string[] = [];
          if (headline) lines.push(headline);
          if (summary) lines.push(summary);
          return lines;
        })();

  return (
    <div data-testid="resume-detail-container" className="ei-resume-detail">
      <button
        type="button"
        data-testid="resume-detail-back"
        className="ei-resume-detail-back"
        onClick={onBack}
      >
        ← {t("resumeWorkshop.detail.back")}
      </button>

      <header className="ei-resume-detail-header">
        <div className="ei-resume-detail-crumb" data-testid="resume-detail-crumb">
          <span className="ei-text-label">{t("resumeWorkshop.eyebrow")}</span>
          <span className="ei-text-label">›</span>
          <span className="ei-text-label">{ui.name}</span>
        </div>
        <h1 className="ei-text-display">{ui.name}</h1>
        <div
          className="ei-resume-detail-meta"
          data-testid="resume-detail-meta"
        >
          {ui.sourceName} · {ui.createdAt} · {t("resumeWorkshop.detail.lastEdit")}{" "}
          {ui.updatedAt}
        </div>
        <div
          className="ei-resume-detail-header-actions"
          data-testid="resume-detail-header-actions"
        >
          <button
            type="button"
            data-testid="resume-detail-export-pdf"
            onClick={onExport}
          >
            {t("resumeWorkshop.detail.exportPdf")}
          </button>
        </div>
      </header>

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
            aria-selected={activeTab === tab}
            onClick={() => setActiveTab(tab)}
          >
            {tabLabels[tab]}
          </button>
        ))}
      </div>

      <div className="ei-resume-detail-tab-content">
        {activeTab === "preview" ? (
          <ResumePreviewTab
            resume={resume}
            onExport={onExport}
            onCopy={onCopy}
            onViewOriginal={() => setOriginalOpen(true)}
          />
        ) : activeTab === "rewrites" ? (
          <ResumeRewritesTabContainer
            resume={resume}
            initialTailorRunId={initialTailorRunId}
            onResumeRefreshed={resumeQuery.retry}
          />
        ) : (
          <ResumeEditTabContainer
            resume={resume}
            onResumeRefreshed={resumeQuery.retry}
          />
        )}
      </div>

      <OriginalResumePreviewModal
        open={originalOpen}
        onClose={() => setOriginalOpen(false)}
        originalText={originalText}
        contentState="ready"
        title={ui.sourceName}
      />
    </div>
  );
};

interface ResumeRewritesTabContainerProps {
  resume: Resume;
  initialTailorRunId: string | null;
  onResumeRefreshed: () => void;
}

const ResumeRewritesTabContainer: FC<ResumeRewritesTabContainerProps> = ({
  resume,
  initialTailorRunId,
  onResumeRefreshed,
}) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const save = useResumeSave();

  const [activeTailorRunId, setActiveTailorRunId] = useState<string | null>(
    initialTailorRunId,
  );
  useEffect(() => {
    setActiveTailorRunId(initialTailorRunId);
  }, [initialTailorRunId, resume.id]);

  const tailorRequest = useRequestResumeTailor();
  const polling = useResumeTailorRunPolling(activeTailorRunId);
  const suggestions = polling.run?.suggestions ?? [];

  const handleRerun = async (mode: "bullet_suggestions" | "gap_review") => {
    if (!resume.id) return;
    try {
      const result = await tailorRequest.request({
        resumeId: resume.id,
        mode,
      });
      setActiveTailorRunId(result.tailorRunId);
      fireResumeWorkshopToast(
        t("resumeWorkshop.rewrites.toast.rerunRequested"),
        "ok",
      );
    } catch (err) {
      if (
        err instanceof RequestResumeTailorError &&
        err.kind === "validation"
      ) {
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
      : polling.phase === "failed" ||
          polling.phase === "timeout" ||
          polling.phase === "error"
        ? {
            kind: "danger",
            message:
              polling.phase === "timeout"
                ? t("resumeWorkshop.rewrites.polling.timeout")
                : t("resumeWorkshop.rewrites.polling.failed"),
            onRetry: () => polling.retry(),
          }
        : null;

  const showSaveError = (err: unknown) => {
    if (err instanceof ResumeSaveError) {
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
    fireResumeWorkshopToast(t("resumeWorkshop.rewrites.error.generic"), "danger");
  };

  // Merge accepted rewrites into the structured profile sections so the saved
  // resume reflects the chosen bullets.
  const buildStructuredProfileWith = (
    acceptedRewrites: { original: string; rewritten: string }[],
  ): Record<string, unknown> => {
    const profile = {
      ...((resume.structuredProfile ?? {}) as Record<string, unknown>),
    };
    const rewrites = new Map(
      acceptedRewrites.map((rewrite) => [rewrite.original, rewrite.rewritten]),
    );
    const sections = profile.sections;
    if (Array.isArray(sections)) {
      profile.sections = sections.map((section) => {
        if (typeof section !== "object" || section === null) return section;
        const record = section as Record<string, unknown>;
        if (!Array.isArray(record.bullets)) return section;
        return {
          ...record,
          bullets: record.bullets.map((bullet) =>
            typeof bullet === "string" ? rewrites.get(bullet) ?? bullet : bullet,
          ),
        };
      });
    }
    return profile;
  };

  const handleOverwrite = async (
    acceptedRewrites: { original: string; rewritten: string }[],
  ) => {
    try {
      await save.overwrite(resume.id, {
        structuredProfile: buildStructuredProfileWith(acceptedRewrites),
      });
      fireResumeWorkshopToast(
        t("resumeWorkshop.rewrites.toast.overwritten").replace(
          "{resumeName}",
          resume.displayName,
        ),
        "ok",
      );
      onResumeRefreshed();
    } catch (err) {
      showSaveError(err);
    }
  };

  const handleSaveAsNew = async (
    acceptedRewrites: { original: string; rewritten: string }[],
  ) => {
    try {
      const created = await save.saveAsNew(resume.id, {
        displayName: t("resumeWorkshop.rewrites.save.newNameSuffix").replace(
          "{resumeName}",
          resume.displayName,
        ),
        structuredProfile: buildStructuredProfileWith(acceptedRewrites),
      });
      fireResumeWorkshopToast(
        t("resumeWorkshop.rewrites.toast.savedAsNew").replace(
          "{resumeName}",
          created.displayName,
        ),
        "ok",
      );
      navigate({
        name: "resume_versions",
        params: { resumeId: created.id, tab: "preview" },
      });
    } catch (err) {
      showSaveError(err);
    }
  };

  return (
    <ResumeRewritesTab
      resume={resume}
      suggestions={suggestions}
      onRequestRerun={handleRerun}
      onOverwrite={handleOverwrite}
      onSaveAsNew={handleSaveAsNew}
      saving={save.pending}
      pollingBanner={pollingBanner}
    />
  );
};

interface ResumeEditTabContainerProps {
  resume: Resume;
  onResumeRefreshed: () => void;
}

const ResumeEditTabContainer: FC<ResumeEditTabContainerProps> = ({
  resume,
  onResumeRefreshed,
}) => {
  const { t } = useI18n();
  const save = useResumeSave();
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const onSave = async ({
    displayName,
    headline,
    summary,
  }: {
    displayName: string;
    headline: string;
    summary: string;
  }) => {
    setErrorMessage(null);
    const existing = (resume.structuredProfile ?? {}) as Record<string, unknown>;
    const nextProfile = {
      ...existing,
      headline,
      summary,
    };
    try {
      await save.overwrite(resume.id, {
        displayName,
        structuredProfile: nextProfile,
      });
      fireResumeWorkshopToast(
        t("resumeWorkshop.edit.toast.saved").replace(
          "{resumeName}",
          displayName || resume.displayName,
        ),
        "ok",
      );
      onResumeRefreshed();
    } catch (err) {
      if (err instanceof ResumeSaveError) {
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
      resume={resume}
      onSave={onSave}
      saving={save.pending}
      errorMessage={errorMessage}
    />
  );
};
