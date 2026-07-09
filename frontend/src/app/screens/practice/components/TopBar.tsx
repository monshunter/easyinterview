import { type FC, type ReactNode } from "react";

export interface TopBarProps {
  company: string;
  title: string;
  questionIndex: number;
  questionTotal: number;
  elapsed: string;
  budget: string;
  paused: boolean;
  onTogglePause: () => void;
  activeMode: "text" | "phone";
  onSwitchMode: (mode: "text" | "phone") => void;
  interviewerLabel: string;
  textModeLabel: string;
  phoneModeLabel: string;
  phoneLiveLabel: string;
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
  elapsed,
  budget,
  paused,
  onTogglePause,
  activeMode,
  onSwitchMode,
  interviewerLabel,
  textModeLabel,
  phoneModeLabel,
  phoneLiveLabel,
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
          {questionIndex}/{questionTotal}
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
          {paused ? "▶" : "❚❚"}
        </button>
        <div
          style={{
            height: 18,
            width: 1,
            background: "var(--ei-color-rule-strong)",
          }}
        />
        <div
          data-testid="practice-topbar-mode-segment"
          style={{
            display: "flex",
            background: "var(--ei-color-bg-soft)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 3,
            padding: 2,
            gap: 2,
          }}
        >
          {([
            ["text", textModeLabel],
            ["phone", phoneModeLabel],
          ] as const).map(([mode, label]) => {
            const on = activeMode === mode;
            return (
              <button
                key={mode}
                data-testid={`practice-topbar-mode-${mode}`}
                type="button"
                aria-pressed={on}
                onClick={() => onSwitchMode(mode)}
                style={{
                  background: on ? "var(--ei-color-bg-card)" : "transparent",
                  border: `1px solid ${
                    on ? "var(--ei-color-rule-strong)" : "transparent"
                  }`,
                  color: on
                    ? "var(--ei-color-fg-primary)"
                    : "var(--ei-color-fg-tertiary)",
                  padding: "4px 9px",
                  borderRadius: 2,
                  cursor: "pointer",
                  fontSize: 12,
                  fontWeight: on ? 500 : 400,
                  fontFamily: "var(--ei-font-sans)",
                }}
              >
                {label}
              </button>
            );
          })}
        </div>
        <div
          data-testid="practice-topbar-live"
          style={{
            display: "flex",
            gap: 5,
            alignItems: "center",
            padding: "4px 8px",
            background: "var(--ei-color-accent-soft)",
            border: "1px solid var(--ei-color-accent)",
            borderRadius: 2,
            visibility: activeMode === "phone" ? "visible" : "hidden",
          }}
        >
          <span
            style={{
              width: 6,
              height: 6,
              borderRadius: 3,
              background: "var(--ei-color-accent)",
              display: "inline-block",
            }}
          />
          <span
            style={{
              fontSize: 11,
              color: "var(--ei-color-accent)",
              fontFamily: "var(--ei-font-mono)",
            }}
          >
            {phoneLiveLabel}
          </span>
        </div>
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
