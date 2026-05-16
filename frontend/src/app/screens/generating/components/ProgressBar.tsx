import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

interface ProgressBarProps {
  phaseIndex: number;
  totalPhases: number;
  activePhaseLabel: string;
}

/**
 * Source-level mirror of ui-design/src/screens-p0-complete.jsx
 * lines 334-344 (phase counter + percentage label + thin progress rail).
 */
export const ProgressBar: FC<ProgressBarProps> = ({
  phaseIndex,
  totalPhases,
  activePhaseLabel,
}) => {
  const { t } = useI18n();
  const pct = Math.min(100, Math.max(0, (phaseIndex / totalPhases) * 100));
  const completed = phaseIndex >= totalPhases;
  return (
    <div data-testid="generating-progress" style={{ marginBottom: 32 }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "baseline",
          marginBottom: 6,
        }}
      >
        <div
          data-testid="generating-progress-counter"
          style={{
            fontFamily: "var(--ei-font-mono)",
            fontSize: 11,
            color: "var(--ei-color-fg-tertiary)",
            letterSpacing: "0.04em",
          }}
        >
          {phaseIndex} / {totalPhases} · {completed ? t("generating.progress.done") : activePhaseLabel}
        </div>
        <div
          data-testid="generating-progress-percentage"
          style={{
            fontFamily: "var(--ei-font-mono)",
            fontSize: 11,
            color: "var(--ei-color-fg-tertiary)",
          }}
        >
          {Math.round(pct)}%
        </div>
      </div>
      <div
        data-testid="generating-progress-rail"
        style={{ height: 2, background: "var(--ei-color-rule-soft)", overflow: "hidden" }}
      >
        <div
          data-testid="generating-progress-fill"
          style={{
            height: "100%",
            width: `${pct}%`,
            background: "var(--ei-color-accent)",
            transition: "width .5s ease",
          }}
        />
      </div>
    </div>
  );
};
