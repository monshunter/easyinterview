import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { mapResumeToUiSource } from "../adapters/resume";
import { useResumeAsset } from "../hooks/useResumeAsset";
import { NotFoundEmptyState } from "./NotFoundEmptyState";
import { ResumePreviewTab } from "./ResumePreviewTab";

export interface ResumeDetailViewProps {
  resumeId: string;
}

export const ResumeDetailView: FC<ResumeDetailViewProps> = ({ resumeId }) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const resumeQuery = useResumeAsset(resumeId);

  const onBack = () => navigate({ name: "resume_versions", params: {} });

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
      </header>

      <ResumePreviewTab resume={resume} />
    </div>
  );
};
