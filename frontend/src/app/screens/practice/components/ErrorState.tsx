import { type FC } from "react";

export interface ErrorStateProps {
  message: string | null;
  retryLabel?: string;
  onRetry?: () => void;
}

/**
 * Inline error banner anchored above InputBar (Phase 4 wires it for
 * append AI 502 / network / 5xx). Phase 1 renders nothing when message is
 * null so the static shell test does not need to special-case it.
 */
export const ErrorState: FC<ErrorStateProps> = ({
  message,
  retryLabel,
  onRetry,
}) => {
  if (!message) return null;
  return (
    <div
      data-testid="practice-error-state"
      style={{
        margin: "8px 40px",
        padding: "10px 12px",
        background: "var(--ei-color-danger-soft)",
        border: "1px solid var(--ei-color-danger)",
        color: "var(--ei-color-danger)",
        borderRadius: 2,
        fontSize: 12,
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        gap: 12,
      }}
    >
      <span data-testid="practice-error-state-message">{message}</span>
      {onRetry && retryLabel && (
        <button
          data-testid="practice-error-state-retry"
          type="button"
          onClick={onRetry}
          style={{
            background: "transparent",
            border: "1px solid var(--ei-color-danger)",
            color: "var(--ei-color-danger)",
            padding: "4px 10px",
            borderRadius: 2,
            cursor: "pointer",
            fontSize: 11,
          }}
        >
          {retryLabel}
        </button>
      )}
    </div>
  );
};
