import { type FC } from "react";

export interface FinishCtaProps {
  label: string;
  disabled?: boolean;
  disabledReason?: string;
  disabledReasonId?: string;
  onFinish: () => void;
}

export const FinishCta: FC<FinishCtaProps> = ({
  label,
  disabled = false,
  disabledReason,
  disabledReasonId = "practice-finish-disabled-reason",
  onFinish,
}) => (
  <div className="ei-practice-finish">
    <button
      data-testid="practice-finish-cta"
      type="button"
      onClick={onFinish}
      disabled={disabled}
      aria-describedby={disabledReason ? disabledReasonId : undefined}
      className="ei-practice-finish-button"
    >
      {label}
    </button>
    {disabledReason ? (
      <span
        id={disabledReasonId}
        data-testid="practice-finish-disabled-reason"
        className="ei-practice-finish-reason"
      >
        {disabledReason}
      </span>
    ) : null}
  </div>
);
