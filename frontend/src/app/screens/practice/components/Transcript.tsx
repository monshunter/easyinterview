import { type FC } from "react";

import { PracticeMessageBody } from "./PracticeMessageBody";

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
  aiLabel: string;
  userLabel: string;
  thinking: boolean;
  thinkingLabel: string;
  retryLabel: string;
  onRetry: (message: TranscriptMessage) => void;
}

export const Transcript: FC<TranscriptProps> = ({ messages, aiLabel, userLabel, thinking, thinkingLabel, retryLabel, onRetry }) => (
  <div data-testid="practice-transcript" className="ei-practice-transcript">
    {messages.map((message, index) => {
      const isAi = message.role === "assistant";
      return (
        <div key={message.id} data-testid={`practice-transcript-message-${index}`} data-role={message.role} className="ei-practice-message-row">
          <div className="ei-practice-message-avatar">{isAi ? "AI" : userLabel}</div>
          <div className="ei-practice-message-content">
            <div className="ei-practice-message-meta">
              <span>{isAi ? aiLabel : userLabel}</span>
              <time>{message.t}</time>
            </div>
            <PracticeMessageBody text={message.text} />
            {!isAi && message.status === "retryable_failed" ? (
              <button
                type="button"
                data-testid="practice-message-retry"
                aria-label={retryLabel}
                title={retryLabel}
                onClick={() => onRetry(message)}
                className="ei-practice-message-retry"
              >
                <RefreshIcon />
              </button>
            ) : null}
          </div>
        </div>
      );
    })}
    {thinking ? (
      <div data-testid="practice-interviewer-thinking" role="status" aria-live="polite" className="ei-practice-message-row">
        <div className="ei-practice-message-avatar">AI</div>
        <div className="ei-practice-thinking">
          <span>{thinkingLabel}</span>
          <span className="ei-pulse" aria-hidden="true">● ● ●</span>
        </div>
      </div>
    ) : null}
  </div>
);

const RefreshIcon: FC = () => (
  <svg aria-hidden="true" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
    <path d="M20 11a8 8 0 1 0 2 5" />
    <path d="M20 4v7h-7" />
  </svg>
);
