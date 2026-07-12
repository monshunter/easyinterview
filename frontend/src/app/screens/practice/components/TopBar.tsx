import { type FC, type ReactNode } from "react";

import { PhoneIcon } from "./PhoneIcon";

export interface TopBarProps {
  company: string;
  title: string;
  elapsed: string;
  budget: string;
  paused: boolean;
  pauseLabel: string;
  resumeLabel: string;
  onTogglePause: () => void;
  interviewerLabel: string;
  phoneDisabledLabel: string;
  finishCta: ReactNode;
}

export const TopBar: FC<TopBarProps> = ({
  company,
  title,
  elapsed,
  budget,
  paused,
  pauseLabel,
  resumeLabel,
  onTogglePause,
  interviewerLabel,
  phoneDisabledLabel,
  finishCta,
}) => (
  <div data-testid="practice-topbar" style={{ padding: "14px 28px", borderBottom: "1px solid var(--ei-color-rule-strong)", display: "flex", alignItems: "center", flexWrap: "wrap", gap: 16, background: "var(--ei-color-bg-card)" }}>
    <div style={{ minWidth: 0 }}>
      <div data-testid="practice-topbar-company" style={{ fontSize: 12, color: "var(--ei-color-fg-tertiary)", fontFamily: "var(--ei-font-mono)", textTransform: "uppercase" }}>{company}</div>
      <div data-testid="practice-topbar-title" style={{ fontSize: 14, color: "var(--ei-color-fg-primary)", fontWeight: 500 }}>{title}</div>
    </div>
    <div style={{ flex: "1 1 80px" }} />
    <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap", justifyContent: "flex-end" }}>
      <span data-testid="practice-topbar-interviewer" className="ei-mono" style={{ display: "inline-flex", alignItems: "center", padding: "3px 8px", borderRadius: 3, fontSize: 11.5, background: "var(--ei-color-bg-soft)", color: "var(--ei-color-fg-tertiary)", border: "1px solid var(--ei-color-rule-strong)" }}>{interviewerLabel}</span>
      <span data-testid="practice-topbar-timer" className="ei-mono" style={{ display: "inline-flex", alignItems: "center", padding: "3px 8px", borderRadius: 3, fontSize: 11.5, background: "var(--ei-color-bg-soft)", color: "var(--ei-color-fg-tertiary)" }}>{elapsed} / {budget}</span>
      <button data-testid="practice-topbar-pause" type="button" onClick={onTogglePause} aria-pressed={paused} style={{ background: "transparent", border: "1px solid var(--ei-color-rule-strong)", padding: "6px 10px", borderRadius: 2, display: "flex", gap: 6, alignItems: "center", color: "var(--ei-color-fg-secondary)", fontSize: 12, cursor: "pointer" }}>{paused ? "▶" : "❚❚"} {paused ? resumeLabel : pauseLabel}</button>
      <div style={{ height: 18, width: 1, background: "var(--ei-color-rule-strong)" }} />
      <button data-testid="practice-topbar-phone-toggle" type="button" disabled aria-disabled="true" aria-label={phoneDisabledLabel} title={phoneDisabledLabel} style={{ width: 34, height: 34, padding: 0, borderRadius: 17, border: "1px solid var(--ei-color-rule-strong)", background: "var(--ei-color-bg-soft)", color: "var(--ei-color-fg-muted)", display: "inline-flex", alignItems: "center", justifyContent: "center", cursor: "not-allowed", opacity: 0.58 }}><PhoneIcon size={15} /></button>
      <div style={{ height: 18, width: 1, background: "var(--ei-color-rule-strong)" }} />
      {finishCta}
    </div>
  </div>
);
