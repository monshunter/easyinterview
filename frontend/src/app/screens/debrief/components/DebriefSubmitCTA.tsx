import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

interface DebriefSubmitCTAProps {
  entriesCount: number;
  entriesReady: boolean;
  targetJobSelected: boolean;
  submitting: boolean;
  onSubmit: () => void;
}

export const DebriefSubmitCTA: FC<DebriefSubmitCTAProps> = ({
  entriesCount,
  entriesReady,
  targetJobSelected,
  submitting,
  onSubmit,
}) => {
  const { t } = useI18n();
  const disabled =
    submitting || entriesCount === 0 || !entriesReady || !targetJobSelected;
  const reason = !targetJobSelected
    ? t("debrief.record.submit.disabledNoTargetJob")
    : entriesCount === 0
      ? t("debrief.record.submit.disabledNoEntries")
      : !entriesReady
        ? t("debrief.record.submit.disabledMissingAnswers")
      : null;
  return (
    <div className="ei-debrief-submit" data-testid="debrief-submit-bar">
      {reason && (
        <div className="ei-debrief-submit__reason" data-testid="debrief-submit-reason">
          {reason}
        </div>
      )}
      <button
        type="button"
        className="ei-debrief-btn ei-debrief-btn--accent"
        data-testid="debrief-submit-btn"
        disabled={disabled}
        aria-disabled={disabled}
        onClick={onSubmit}
      >
        {t("debrief.record.submit.cta")}
      </button>
    </div>
  );
};
