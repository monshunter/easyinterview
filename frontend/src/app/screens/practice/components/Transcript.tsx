import { type FC } from "react";

export interface TranscriptMessage {
  role: "user" | "ai";
  text: string;
  t: string;
  followUp?: boolean;
}

export interface TranscriptProps {
  messages: TranscriptMessage[];
  helperText: string;
  aiLabel: string;
  userLabel: string;
  followUpLabel: string;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx` lines 208-215
 * (transcript scroll area) + lines 592-615 (TranscriptMsg). Phase 1 renders
 * helper-only state; Phase 2 wires React state with `appendSessionEvent`.
 */
export const Transcript: FC<TranscriptProps> = ({
  messages,
  helperText,
  aiLabel,
  userLabel,
  followUpLabel,
}) => {
  return (
    <div
      data-testid="practice-transcript"
      style={{
        flex: 1,
        overflowY: "auto",
        padding: "20px 40px",
      }}
    >
      {messages.map((m, idx) => {
        const isAi = m.role === "ai";
        return (
          <div
            key={idx}
            data-testid={`practice-transcript-message-${idx}`}
            data-role={m.role}
            style={{ marginBottom: 18, display: "flex", gap: 12 }}
          >
            <div
              style={{
                width: 28,
                height: 28,
                borderRadius: 2,
                flexShrink: 0,
                background: isAi
                  ? "var(--ei-color-accent-soft)"
                  : "var(--ei-color-bg-soft)",
                color: isAi
                  ? "var(--ei-color-accent)"
                  : "var(--ei-color-fg-secondary)",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontSize: 11,
                fontFamily: "var(--ei-font-mono)",
                fontWeight: 500,
              }}
            >
              {isAi ? "AI" : userLabel}
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div
                style={{
                  display: "flex",
                  gap: 8,
                  alignItems: "center",
                  marginBottom: 4,
                }}
              >
                <span
                  style={{
                    fontSize: 12,
                    color: "var(--ei-color-fg-secondary)",
                    fontWeight: 500,
                  }}
                >
                  {isAi ? aiLabel : userLabel}
                </span>
                {m.followUp && (
                  <span
                    data-testid={`practice-transcript-follow-up-badge-${idx}`}
                    className="ei-mono"
                    style={{
                      padding: "1px 6px",
                      borderRadius: 2,
                      fontSize: 11,
                      background: "var(--ei-color-amber-soft)",
                      color: "var(--ei-color-warn)",
                    }}
                  >
                    {followUpLabel}
                  </span>
                )}
                <span
                  style={{
                    fontSize: 11,
                    color: "var(--ei-color-fg-muted)",
                    fontFamily: "var(--ei-font-mono)",
                  }}
                >
                  {m.t}
                </span>
              </div>
              <div
                style={{
                  fontSize: 14,
                  color: "var(--ei-color-fg-primary)",
                  lineHeight: 1.6,
                }}
              >
                {m.text}
              </div>
            </div>
          </div>
        );
      })}
      <div
        data-testid="practice-transcript-helper"
        style={{
          textAlign: "center",
          marginTop: 12,
          fontSize: 11,
          color: "var(--ei-color-fg-tertiary)",
          fontFamily: "var(--ei-font-mono)",
        }}
      >
        {helperText}
      </div>
    </div>
  );
};
