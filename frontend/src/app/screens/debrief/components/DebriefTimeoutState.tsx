import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

interface DebriefTimeoutStateProps {
  onRetry: () => void;
  onBackToEdit: () => void;
}

export const DebriefTimeoutState: FC<DebriefTimeoutStateProps> = ({
  onRetry,
  onBackToEdit,
}) => {
  const { t } = useI18n();
  return (
    <section
      className="ei-debrief-state-card ei-debrief-state-card--timeout"
      data-testid="debrief-timeout-state"
    >
      <div className="ei-label">{t("debrief.timeout.eyebrow")}</div>
      <h2 className="ei-serif">{t("debrief.timeout.title")}</h2>
      <p>{t("debrief.timeout.body")}</p>
      <div className="ei-debrief-state-card__actions">
        <button
          type="button"
          className="ei-debrief-btn ei-debrief-btn--ghost"
          data-testid="debrief-timeout-back"
          onClick={onBackToEdit}
        >
          {t("debrief.timeout.back")}
        </button>
        <button
          type="button"
          className="ei-debrief-btn ei-debrief-btn--accent"
          data-testid="debrief-timeout-retry"
          onClick={onRetry}
        >
          {t("debrief.timeout.retry")}
        </button>
      </div>
    </section>
  );
};
