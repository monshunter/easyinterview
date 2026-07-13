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
  <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: 5 }}>
    <button
      data-testid="practice-finish-cta"
      type="button"
      onClick={onFinish}
      disabled={disabled}
      aria-describedby={disabledReason ? disabledReasonId : undefined}
      style={{ padding: "7px 12px", background: disabled ? "var(--ei-color-bg-soft)" : "var(--ei-color-accent)", color: disabled ? "var(--ei-color-fg-muted)" : "#fff", border: disabled ? "1px solid var(--ei-color-rule-strong)" : "none", borderRadius: 2, cursor: disabled ? "not-allowed" : "pointer", fontSize: 12.5, fontWeight: 500, fontFamily: "var(--ei-font-sans)" }}
    >
      {label}
    </button>
    {disabledReason ? (
      <span
        id={disabledReasonId}
        data-testid="practice-finish-disabled-reason"
        style={{ maxWidth: 190, color: "var(--ei-color-fg-tertiary)", fontSize: 11, lineHeight: 1.35, textAlign: "right" }}
      >
        {disabledReason}
      </span>
    ) : null}
  </div>
);
