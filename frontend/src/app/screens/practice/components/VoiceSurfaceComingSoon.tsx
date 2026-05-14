import { type FC } from "react";

export interface VoiceSurfaceComingSoonProps {
  title: string;
  desc: string;
  backLabel: string;
  onBackToText: () => void;
}

/**
 * Scoped temporary placeholder for the voice surface. Lands as part of
 * plan 002 to give text-mode users a clear signal that voice is deferred to
 * future plan 003 (`practice-voice-mvp`). NOT a source parity row — voice
 * surface DOM (waveform, annotated waveform, expression panel) MUST NOT be
 * imported here.
 */
export const VoiceSurfaceComingSoon: FC<VoiceSurfaceComingSoonProps> = ({
  title,
  desc,
  backLabel,
  onBackToText,
}) => {
  return (
    <div
      data-testid="practice-voice-coming-soon"
      style={{
        flex: 1,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: 12,
        padding: 32,
        background: "var(--ei-color-bgCard)",
      }}
    >
      <div
        data-testid="practice-voice-coming-soon-icon"
        aria-hidden="true"
        style={{
          width: 48,
          height: 48,
          borderRadius: 24,
          background: "var(--ei-color-accentSoft)",
          color: "var(--ei-color-accent)",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          fontSize: 22,
        }}
      >
        ●
      </div>
      <div
        data-testid="practice-voice-coming-soon-title"
        className="ei-serif"
        style={{ fontSize: 18, color: "var(--ei-color-ink)" }}
      >
        {title}
      </div>
      <div
        data-testid="practice-voice-coming-soon-desc"
        style={{
          fontSize: 13,
          color: "var(--ei-color-ink3)",
          textAlign: "center",
          maxWidth: 360,
          lineHeight: 1.6,
        }}
      >
        {desc}
      </div>
      <button
        data-testid="practice-voice-coming-soon-back-to-text"
        type="button"
        onClick={onBackToText}
        style={{
          marginTop: 8,
          padding: "8px 14px",
          fontSize: 13,
          background: "transparent",
          border: "1px solid var(--ei-color-rule)",
          color: "var(--ei-color-ink2)",
          borderRadius: 2,
          cursor: "pointer",
        }}
      >
        {backLabel}
      </button>
    </div>
  );
};
