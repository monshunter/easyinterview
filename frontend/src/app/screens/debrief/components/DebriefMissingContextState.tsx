import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

interface DebriefMissingContextStateProps {
  onPickTargetJob: () => void;
}

export const DebriefMissingContextState: FC<DebriefMissingContextStateProps> = ({
  onPickTargetJob,
}) => {
  const { t } = useI18n();
  return (
    <section
      className="ei-debrief-state-card ei-debrief-state-card--missing"
      data-testid="debrief-missing-context-state"
    >
      <div className="ei-label">{t("debrief.missing.eyebrow")}</div>
      <h2 className="ei-serif">{t("debrief.missing.title")}</h2>
      <p>{t("debrief.missing.body")}</p>
      <div className="ei-debrief-state-card__actions">
        <button
          type="button"
          className="ei-debrief-btn ei-debrief-btn--accent"
          data-testid="debrief-missing-cta"
          onClick={onPickTargetJob}
        >
          {t("debrief.missing.cta")}
        </button>
      </div>
    </section>
  );
};
