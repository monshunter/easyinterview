import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

export interface NotFoundEmptyStateProps {
  onBack: () => void;
}

export const NotFoundEmptyState: FC<NotFoundEmptyStateProps> = ({ onBack }) => {
  const { t } = useI18n();
  return (
    <div
      data-testid="resume-detail-not-found"
      className="ei-screen-card ei-resume-detail-not-found"
      role="alert"
    >
      <h2 className="ei-text-title">
        {t("resumeWorkshop.detail.notFoundTitle")}
      </h2>
      <p className="ei-text-body">
        {t("resumeWorkshop.detail.notFoundBody")}
      </p>
      <button
        type="button"
        data-testid="resume-detail-not-found-back"
        className="ei-cta"
        onClick={onBack}
      >
        {t("resumeWorkshop.detail.notFoundCta")}
      </button>
    </div>
  );
};
