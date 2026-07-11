import { type FC, type ReactNode } from "react";

import { PhoneIcon } from "./PhoneIcon";

export interface TopBarProps {
  company: string;
  title: string;
  questionIndex: number;
  questionTotal: number;
  questionLabel: string;
  elapsed: string;
  budget: string;
  paused: boolean;
  pauseLabel: string;
  resumeLabel: string;
  onTogglePause: () => void;
  activeMode: "text" | "phone";
  onTogglePhone: () => void;
  interviewerLabel: string;
  phoneToggleLabel: string;
  phoneToggleTitle: string;
  finishCta: ReactNode;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx::PracticeScreen`.
 * The interviewer identity is display-only and comes from the round plan.
 */
export const TopBar: FC<TopBarProps> = ({
  company,
  title,
  questionIndex,
  questionTotal,
  questionLabel,
  elapsed,
  budget,
  paused,
  pauseLabel,
  resumeLabel,
  onTogglePause,
  activeMode,
  onTogglePhone,
  interviewerLabel,
  phoneToggleLabel,
  phoneToggleTitle,
  finishCta,
}) => {
  return (
    <div
      data-testid="practice-topbar"
      style={{
        padding: "14px 28px",
        borderBottom: "1px solid var(--ei-color-rule-strong)",
        display: "flex",
        alignItems: "center",
        flexWrap: "wrap",
        gap: 16,
        background: "var(--ei-color-bg-card)",
      }}
    >
      <div style={{ minWidth: 0 }}>
        <div
          data-testid="practice-topbar-company"
          style={{
            fontSize: 12,
            color: "var(--ei-color-fg-tertiary)",
            fontFamily: "var(--ei-font-mono)",
            textTransform: "uppercase",
          }}
        >
          {company}
        </div>
        <div
          data-testid="practice-topbar-title"
          style={{
            fontSize: 14,
            color: "var(--ei-color-fg-primary)",
            fontWeight: 500,
          }}
        >
          {title}
        </div>
      </div>
      <div style={{ flex: "1 1 80px" }} />
      <div
        style={{
          display: "flex",
          gap: 8,
          alignItems: "center",
          flexWrap: "wrap",
          justifyContent: "flex-end",
          minWidth: 0,
        }}
      >
        <span
          data-testid="practice-topbar-interviewer"
          className="ei-mono"
          style={{
            display: "inline-flex",
            alignItems: "center",
            padding: "3px 8px",
            borderRadius: 3,
            fontSize: 11.5,
            background: "var(--ei-color-bg-soft)",
            color: "var(--ei-color-fg-tertiary)",
            border: "1px solid var(--ei-color-rule-strong)",
          }}
        >
          {interviewerLabel}
        </span>
        <span
          data-testid="practice-topbar-question"
          className="ei-mono"
          style={{
            display: "inline-flex",
            alignItems: "center",
            padding: "3px 8px",
            borderRadius: 3,
            fontSize: 11.5,
            background: "var(--ei-color-accent-soft)",
            color: "var(--ei-color-accent)",
          }}
        >
          {questionLabel} {questionIndex}/{questionTotal}
        </span>
        <span
          data-testid="practice-topbar-timer"
          className="ei-mono"
          style={{
            display: "inline-flex",
            alignItems: "center",
            padding: "3px 8px",
            borderRadius: 3,
            fontSize: 11.5,
            background: "var(--ei-color-bg-soft)",
            color: "var(--ei-color-fg-tertiary)",
          }}
        >
          {elapsed} / {budget}
        </span>
        <button
          data-testid="practice-topbar-pause"
          type="button"
          onClick={onTogglePause}
          aria-pressed={paused}
          style={{
            background: "transparent",
            border: "1px solid var(--ei-color-rule-strong)",
            padding: "6px 10px",
            borderRadius: 2,
            display: "flex",
            gap: 6,
            alignItems: "center",
            color: "var(--ei-color-fg-secondary)",
            fontSize: 12,
            cursor: "pointer",
          }}
        >
          {paused ? "▶" : "❚❚"} {paused ? resumeLabel : pauseLabel}
        </button>
        <div
          style={{
            height: 18,
            width: 1,
            background: "var(--ei-color-rule-strong)",
          }}
        />
        <button
          data-testid="practice-topbar-phone-toggle"
          type="button"
          aria-pressed={activeMode === "phone"}
          aria-label={phoneToggleLabel}
          title={phoneToggleTitle}
          onClick={onTogglePhone}
          style={{
            width: 34,
            height: 34,
            padding: 0,
            borderRadius: 17,
            border: `1px solid ${
              activeMode === "phone"
                ? "var(--ei-color-accent)"
                : "var(--ei-color-rule-strong)"
            }`,
            background:
              activeMode === "phone"
                ? "var(--ei-color-accent-soft)"
                : "transparent",
            color:
              activeMode === "phone"
                ? "var(--ei-color-accent)"
                : "var(--ei-color-fg-secondary)",
            display: "inline-flex",
            alignItems: "center",
            justifyContent: "center",
            cursor: "pointer",
          }}
        >
          <PhoneIcon size={15} />
        </button>
        <div
          style={{
            height: 18,
            width: 1,
            background: "var(--ei-color-rule-strong)",
          }}
        />
        {finishCta}
      </div>
    </div>
  );
};
