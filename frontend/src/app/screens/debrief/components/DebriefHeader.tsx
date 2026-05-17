import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { DebriefSelectedContext } from "../types";

interface DebriefHeaderProps {
  selectedContext: DebriefSelectedContext;
  onBack: () => void;
  interviewerLabel?: string;
  modalityLabel?: string;
  capturedAtLabel?: string;
}

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx::DebriefFullScreen
 * lines 122-144. The eyebrow / title / sub-copy / right-rail meta layout is
 * lifted from the prototype; copy keys are localized via the central catalog
 * to keep i18n auditable.
 */
export const DebriefHeader: FC<DebriefHeaderProps> = ({
  selectedContext,
  onBack,
  interviewerLabel,
  modalityLabel,
  capturedAtLabel,
}) => {
  const { t } = useI18n();
  const targetJob = selectedContext.targetJob;
  const eyebrowSegments = [
    t("debrief.header.eyebrowPrefix"),
    targetJob?.companyName ?? t("debrief.header.eyebrowFallbackCompany"),
    targetJob?.title ?? t("debrief.header.eyebrowFallbackRound"),
  ];

  return (
    <header
      className="ei-debrief-header"
      data-testid="debrief-header"
    >
      <button
        type="button"
        className="ei-debrief-header__back"
        data-testid="debrief-header-back"
        onClick={onBack}
      >
        <span aria-hidden="true">←</span> {t("debrief.header.backHome")}
      </button>
      <div className="ei-debrief-header__row">
        <div className="ei-debrief-header__primary">
          <div
            className="ei-label ei-debrief-header__eyebrow"
            data-testid="debrief-header-eyebrow"
          >
            {eyebrowSegments.join(" · ")}
          </div>
          <h1
            className="ei-serif ei-debrief-header__title"
            data-testid="debrief-header-title"
          >
            {t("debrief.header.title")}
          </h1>
          <p
            className="ei-debrief-header__subcopy"
            data-testid="debrief-header-subcopy"
          >
            {t("debrief.header.subcopy")}
          </p>
        </div>
        <dl
          className="ei-debrief-header__meta"
          data-testid="debrief-header-meta"
        >
          <div>
            <dt>{t("debrief.header.metaCapturedAt")}</dt>
            <dd>{capturedAtLabel ?? t("debrief.header.metaFallback")}</dd>
          </div>
          <div>
            <dt>{t("debrief.header.metaInterviewer")}</dt>
            <dd>{interviewerLabel ?? t("debrief.header.metaFallback")}</dd>
          </div>
          <div>
            <dt>{t("debrief.header.metaModality")}</dt>
            <dd>{modalityLabel ?? t("debrief.header.metaFallback")}</dd>
          </div>
        </dl>
      </div>
    </header>
  );
};
