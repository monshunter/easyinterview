import { type FC } from "react";

export type TranscriptMessageStatus = "pending" | "retrying" | "retryable_failed" | "terminal_failed" | "complete";
export interface TranscriptMessage {
  id: string;
  role: "user" | "assistant";
  text: string;
  t: string;
  clientMessageId?: string;
  status?: TranscriptMessageStatus;
}
export interface TranscriptProps {
  messages: TranscriptMessage[];
  helperText: string;
  aiLabel: string;
  userLabel: string;
  thinking: boolean;
  thinkingLabel: string;
  retryLabel: string;
  onRetry: (message: TranscriptMessage) => void;
}

export const Transcript: FC<TranscriptProps> = ({ messages, helperText, aiLabel, userLabel, thinking, thinkingLabel, retryLabel, onRetry }) => (
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
            {!isAi && message.status === "retryable_failed" ? (
              <button
                type="button"
                data-testid="practice-message-retry"
                aria-label={retryLabel}
                title={retryLabel}
                onClick={() => onRetry(message)}
                style={{ marginTop: 7, width: 28, height: 28, display: "inline-flex", alignItems: "center", justifyContent: "center", border: "1px solid var(--ei-color-rule-strong)", borderRadius: 2, background: "var(--ei-color-bg-card)", color: "var(--ei-color-accent)", padding: 0 }}
              >
                <RefreshIcon />
              </button>
            ) : null}
          </div>
        </div>
      );
    })}
    {thinking ? (
      <div data-testid="practice-interviewer-thinking" role="status" aria-live="polite" style={{ marginBottom: 18, display: "flex", gap: 12 }}>
        <div style={{ width: 28, height: 28, borderRadius: 2, flexShrink: 0, background: "var(--ei-color-accent-soft)", color: "var(--ei-color-accent)", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontFamily: "var(--ei-font-mono)", fontWeight: 500 }}>AI</div>
        <div style={{ minWidth: 0, display: "flex", alignItems: "center", gap: 8, color: "var(--ei-color-fg-tertiary)", fontSize: 13.5 }}>
          <span>{thinkingLabel}</span>
          <span className="ei-pulse" aria-hidden="true">● ● ●</span>
        </div>
      </div>
    ) : null}
    <div data-testid="practice-transcript-helper" style={{ textAlign: "center", marginTop: 12, fontSize: 11, color: "var(--ei-color-fg-tertiary)", fontFamily: "var(--ei-font-mono)" }}>{helperText}</div>
  </div>
);

const RefreshIcon: FC = () => (
  <svg aria-hidden="true" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="M20 11a8 8 0 1 0 2 5" />
    <path d="M20 4v7h-7" />
  </svg>
);
