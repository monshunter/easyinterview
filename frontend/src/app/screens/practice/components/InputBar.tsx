import { type FC, type ReactNode } from "react";

export interface InputBarProps {
  value: string;
  onChange: (next: string) => void;
  placeholder: string;
  hintLabel: string;
  sendLabel: string;
  showHintButton: boolean;
  disabled: boolean;
  onHint: () => void;
  onSend: () => void;
  hintBanner: ReactNode | null;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx` lines 218-254
 * (input region + optional hint + send). Text mode does not expose dictation
 * or turn-bypass controls.
 */
export const InputBar: FC<InputBarProps> = ({
  value,
  onChange,
  placeholder,
  hintLabel,
  sendLabel,
  showHintButton,
  disabled,
  onHint,
  onSend,
  hintBanner,
}) => {
  return (
    <div
      data-testid="practice-input"
      style={{
        padding: "16px 40px 24px",
        borderTop: "1px solid var(--ei-color-rule-strong)",
        background: "var(--ei-color-bg-card)",
      }}
    >
      {hintBanner}
      <div
        style={{
          border: "1px solid var(--ei-color-rule-strong)",
          borderRadius: 2,
          padding: 12,
          background: "var(--ei-color-bg-canvas)",
        }}
      >
        <textarea
          data-testid="practice-input-textarea"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          disabled={disabled}
          style={{
            width: "100%",
            minHeight: 70,
            border: "none",
            outline: "none",
            resize: "none",
            fontSize: 14,
            lineHeight: 1.55,
            background: "transparent",
            color: "var(--ei-color-fg-primary)",
            fontFamily: "var(--ei-font-sans)",
          }}
        />
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginTop: 6,
          }}
        >
          <div style={{ display: "flex", gap: 6 }}>
            {showHintButton && (
              <button
                data-testid="practice-input-hint"
                type="button"
                onClick={onHint}
                disabled={disabled}
                style={{
                  background: "transparent",
                  border: "1px solid var(--ei-color-rule-strong)",
                  padding: "6px 10px",
                  borderRadius: 2,
                  fontSize: 12,
                  color: "var(--ei-color-fg-secondary)",
                  cursor: disabled ? "default" : "pointer",
                }}
              >
                {hintLabel}
              </button>
            )}
          </div>
          <div style={{ display: "flex", gap: 8 }}>
            <button
              data-testid="practice-input-send"
              type="button"
              onClick={onSend}
              disabled={disabled}
              style={{
                background: "var(--ei-color-accent)",
                color: "#fff",
                border: "1px solid var(--ei-color-accent)",
                padding: "6px 14px",
                borderRadius: 2,
                fontSize: 12,
                fontWeight: 500,
                cursor: disabled ? "default" : "pointer",
              }}
            >
              {sendLabel}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
