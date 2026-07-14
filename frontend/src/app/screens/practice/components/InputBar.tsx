import { type FC, type ReactNode } from "react";

export interface InputBarProps {
  value: string;
  onChange: (next: string) => void;
  placeholder: string;
  sendLabel: string;
  disabled: boolean;
  sendDisabled?: boolean;
  recovery?: ReactNode;
  onSend: () => void;
}

export const InputBar: FC<InputBarProps> = ({ value, onChange, placeholder, sendLabel, disabled, sendDisabled = false, recovery, onSend }) => {
  const isSendDisabled = disabled || sendDisabled || !value.trim();
  return (
    <div data-testid="practice-input" style={{ padding: "16px clamp(24px, 8vw, 144px) 24px", borderTop: "1px solid var(--ei-color-rule-strong)", background: "var(--ei-color-bg-card)" }}>
      {recovery}
      <div style={{ border: "1px solid var(--ei-color-rule-strong)", borderRadius: 2, padding: 12, background: "var(--ei-color-bg-canvas)" }}>
        <textarea data-testid="practice-input-textarea" value={value} onChange={(event) => onChange(event.target.value)} onKeyDown={(event) => { if (!isSendDisabled && (event.metaKey || event.ctrlKey) && event.key === "Enter") onSend(); }} placeholder={placeholder} disabled={disabled} style={{ width: "100%", minHeight: 74, border: "none", outline: "none", resize: "none", fontSize: 14, lineHeight: 1.55, background: "transparent", color: "var(--ei-color-fg-primary)", fontFamily: "var(--ei-font-sans)" }} />
        <div style={{ display: "flex", justifyContent: "flex-end", marginTop: 6 }}>
          <button data-testid="practice-input-send" type="button" onClick={onSend} disabled={isSendDisabled} style={{ background: "var(--ei-color-accent)", color: "#fff", border: "1px solid var(--ei-color-accent)", padding: "6px 14px", borderRadius: 2, fontSize: 12, fontWeight: 500, cursor: isSendDisabled ? "not-allowed" : "pointer", opacity: isSendDisabled ? 0.5 : 1 }}>{sendLabel}</button>
        </div>
      </div>
    </div>
  );
};
