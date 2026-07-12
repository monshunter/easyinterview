import { type FC } from "react";

export interface TranscriptMessage { id: string; role: "user" | "assistant"; text: string; t: string; }
export interface TranscriptProps { messages: TranscriptMessage[]; helperText: string; aiLabel: string; userLabel: string; }

export const Transcript: FC<TranscriptProps> = ({ messages, helperText, aiLabel, userLabel }) => (
  <div data-testid="practice-transcript" style={{ flex: 1, overflowY: "auto", padding: "28px clamp(24px, 8vw, 144px) 20px" }}>
    {messages.map((message, index) => {
      const isAi = message.role === "assistant";
      return (
        <div key={message.id} data-testid={`practice-transcript-message-${index}`} data-role={message.role} style={{ marginBottom: 18, display: "flex", gap: 12 }}>
          <div style={{ width: 28, height: 28, borderRadius: 2, flexShrink: 0, background: isAi ? "var(--ei-color-accent-soft)" : "var(--ei-color-bg-soft)", color: isAi ? "var(--ei-color-accent)" : "var(--ei-color-fg-secondary)", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontFamily: "var(--ei-font-mono)", fontWeight: 500 }}>{isAi ? "AI" : userLabel}</div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 4 }}>
              <span style={{ fontSize: 12, color: "var(--ei-color-fg-secondary)", fontWeight: 500 }}>{isAi ? aiLabel : userLabel}</span>
              <span style={{ fontSize: 11, color: "var(--ei-color-fg-muted)", fontFamily: "var(--ei-font-mono)" }}>{message.t}</span>
            </div>
            <div style={{ fontSize: 14, color: "var(--ei-color-fg-primary)", lineHeight: 1.6 }}>{message.text}</div>
          </div>
        </div>
      );
    })}
    <div data-testid="practice-transcript-helper" style={{ textAlign: "center", marginTop: 12, fontSize: 11, color: "var(--ei-color-fg-tertiary)", fontFamily: "var(--ei-font-mono)" }}>{helperText}</div>
  </div>
);
