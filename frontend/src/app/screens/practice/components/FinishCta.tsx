import { type FC } from "react";

export interface FinishCtaProps {
  label: string;
  hintCount: number;
  hintUsageNote: string;
  disabled?: boolean;
  onFinish: () => void;
}

/**
 * Global finish action for the practice top bar. Wired to
 * `completePracticeSession` + `Idempotency-Key` in Phase 4.1.
 */
export const FinishCta: FC<FinishCtaProps> = ({
  label,
  hintCount,
  hintUsageNote,
  disabled = false,
  onFinish,
}) => {
  return (
    <div
      data-testid="practice-finish-cta-wrap"
      style={{
        flexShrink: 0,
      }}
    >
      <button
        data-testid="practice-finish-cta"
        type="button"
        onClick={onFinish}
        disabled={disabled}
        style={{
          padding: "7px 12px",
          background: disabled
            ? "var(--ei-color-fg-muted)"
            : "var(--ei-color-accent)",
          color: "#fff",
          border: "none",
          borderRadius: 2,
          cursor: disabled ? "default" : "pointer",
          fontSize: 12.5,
          fontWeight: 500,
          fontFamily: "var(--ei-font-sans)",
        }}
      >
        {label}
      </button>
      {hintCount > 0 && (
        <div
          data-testid="practice-finish-hint-count"
          style={{
            fontSize: 11,
            color: "var(--ei-color-fg-tertiary)",
            textAlign: "center",
            marginTop: 6,
            fontFamily: "var(--ei-font-mono)",
          }}
        >
          {hintUsageNote.replace("{count}", String(hintCount))}
        </div>
      )}
    </div>
  );
};
