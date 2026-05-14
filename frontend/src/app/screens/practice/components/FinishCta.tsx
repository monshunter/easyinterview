import { type FC } from "react";

export interface FinishCtaProps {
  label: string;
  hintCount: number;
  hintUsageNote: string;
  disabled?: boolean;
  onFinish: () => void;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx` lines 303-321
 * (pinned bottom CTA + hint usage note). Wired to `completePracticeSession`
 * + `Idempotency-Key` in Phase 4.1.
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
      data-testid="practice-rightpanel-cta-finish-wrap"
      style={{
        padding: "14px 18px",
        borderTop: "1px solid var(--ei-color-rule)",
        background: "var(--ei-color-bgCard)",
        flexShrink: 0,
      }}
    >
      <button
        data-testid="practice-rightpanel-cta-finish"
        type="button"
        onClick={onFinish}
        disabled={disabled}
        style={{
          width: "100%",
          padding: "11px 14px",
          background: disabled
            ? "var(--ei-color-ink4)"
            : "var(--ei-color-accent)",
          color: "#fff",
          border: "none",
          borderRadius: 2,
          cursor: disabled ? "default" : "pointer",
          fontSize: 13.5,
          fontWeight: 500,
          fontFamily: "var(--ei-sans)",
        }}
      >
        {label}
      </button>
      {hintCount > 0 && (
        <div
          data-testid="practice-rightpanel-hint-count"
          style={{
            fontSize: 11,
            color: "var(--ei-color-ink3)",
            textAlign: "center",
            marginTop: 6,
            fontFamily: "var(--ei-mono)",
          }}
        >
          {hintUsageNote.replace("{count}", String(hintCount))}
        </div>
      )}
    </div>
  );
};
