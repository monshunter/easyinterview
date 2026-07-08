import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { buildResumeBodyMarkdown, mapResumeToUiSource } from "../adapters/resume";
import { useResumeAsset } from "../hooks/useResumeAsset";
import { NotFoundEmptyState } from "./NotFoundEmptyState";
import { ResumePreviewTab } from "./ResumePreviewTab";
import { ResumeWorkshopIcon } from "./ResumeWorkshopIcon";

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
  const isParsing =
    resume.parseStatus === "queued" || resume.parseStatus === "processing";
  const hasReadableBody = buildResumeBodyMarkdown(resume).trim().length > 0;

  if (isParsing) {
    return (
      <ResumeParseState
        testId="resume-detail-parse-waiting"
        icon="sparkle"
        title={t("resumeWorkshop.detail.waitingTitle")}
        body={t("resumeWorkshop.detail.waitingBody")}
        backLabel={t("resumeWorkshop.detail.back")}
        onBack={onBack}
      />
    );
  }

  if (resume.parseStatus === "failed" && !hasReadableBody) {
    return (
      <ResumeParseState
        testId="resume-detail-parse-failed"
        icon="file"
        title={t("resumeWorkshop.detail.failedTitle")}
        body={t("resumeWorkshop.detail.failedBody")}
        backLabel={t("resumeWorkshop.detail.back")}
        onBack={onBack}
      />
    );
  }

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

interface ResumeParseStateProps {
  testId: string;
  icon: "file" | "sparkle";
  title: string;
  body: string;
  backLabel: string;
  onBack: () => void;
}

const ResumeParseState: FC<ResumeParseStateProps> = ({
  testId,
  icon,
  title,
  body,
  backLabel,
  onBack,
}) => (
  <div data-testid={testId} className="ei-resume-detail-parse-state">
    <div className="ei-resume-detail-parse-icon" aria-hidden="true">
      <ResumeWorkshopIcon name={icon} size={22} />
    </div>
    <h1 className="ei-text-display">{title}</h1>
    <p className="ei-text-body">{body}</p>
    <button
      type="button"
      data-testid="resume-detail-parse-back"
      className="ei-resume-detail-back"
      onClick={onBack}
    >
      {backLabel}
    </button>
  </div>
);
