import { type FC } from "react";

export interface TerminalRecoveryProps {
  title: string;
  description: string;
  ctaLabel: string;
  onBackToPlan: () => void;
}

export const TerminalRecovery: FC<TerminalRecoveryProps> = ({
  title,
  description,
  ctaLabel,
  onBackToPlan,
}) => (
  <div
    data-testid="practice-terminal-recovery"
    role="alert"
    style={{
      marginBottom: 12,
      padding: "12px 14px",
      border: "1px solid var(--ei-color-rule-strong)",
      borderRadius: 2,
      background: "var(--ei-color-bg-soft)",
      display: "flex",
      alignItems: "center",
      justifyContent: "space-between",
      gap: 14,
      flexWrap: "wrap",
    }}
  >
    <div style={{ minWidth: 0 }}>
      <div style={{ color: "var(--ei-color-fg-primary)", fontSize: 13.5, fontWeight: 500, marginBottom: 3 }}>
        {title}
      </div>
      <div style={{ color: "var(--ei-color-fg-tertiary)", fontSize: 12, lineHeight: 1.5 }}>
        {description}
      </div>
    </div>
    <button
      type="button"
      data-testid="practice-terminal-recovery-cta"
      onClick={onBackToPlan}
      style={{
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        gap: 8,
        height: 30,
        padding: "0 12px",
        fontSize: 13,
        fontWeight: 500,
        background: "var(--ei-color-bg-canvas)",
        color: "var(--ei-color-fg-primary)",
        border: "1px solid var(--ei-color-rule-strong)",
        borderRadius: 2,
        cursor: "pointer",
        opacity: 1,
        fontFamily: "var(--ei-font-sans)",
        letterSpacing: "-0.005em",
        transition: "transform .08s ease, opacity .15s",
      }}
      onMouseDown={(event) => { event.currentTarget.style.transform = "translateY(0.5px)"; }}
      onMouseUp={(event) => { event.currentTarget.style.transform = ""; }}
      onMouseLeave={(event) => { event.currentTarget.style.transform = ""; }}
    >
      <ArrowLeftIcon />
      {ctaLabel}
    </button>
  </div>
);

const ArrowLeftIcon: FC = () => (
  <svg
    aria-hidden="true"
    viewBox="0 0 24 24"
    width="15"
    height="15"
    fill="none"
    stroke="currentColor"
    strokeWidth="1.5"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="M19 12H5M11 18l-6-6 6-6" />
  </svg>
);
