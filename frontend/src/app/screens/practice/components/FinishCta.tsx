import { type FC } from "react";

export interface FinishCtaProps { label: string; disabled?: boolean; onFinish: () => void; }

export const FinishCta: FC<FinishCtaProps> = ({ label, disabled = false, onFinish }) => (
  <button data-testid="practice-finish-cta" type="button" onClick={onFinish} disabled={disabled} style={{ padding: "7px 12px", background: disabled ? "var(--ei-color-fg-muted)" : "var(--ei-color-accent)", color: "#fff", border: "none", borderRadius: 2, cursor: disabled ? "default" : "pointer", fontSize: 12.5, fontWeight: 500, fontFamily: "var(--ei-font-sans)" }}>{label}</button>
);
