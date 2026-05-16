import type { FC } from "react";

export interface PhaseDefinition {
  /** i18n label key for the phase headline (e.g. generating.phase.1). */
  labelKey: string;
  /** i18n hint key shown on the right edge (e.g. generating.phase.1.hint). */
  hintKey: string;
}

interface PhaseListProps {
  phaseIndex: number;
  phases: PhaseDefinition[];
  resolve: (key: string) => string;
}

/**
 * Source-level mirror of ui-design/src/screens-p0-complete.jsx
 * lines 347-371. Renders the 5-phase column with done / active / pending state
 * circles (line-through done text, pulsing dot for active, soft hint on right).
 */
export const PhaseList: FC<PhaseListProps> = ({ phaseIndex, phases, resolve }) => {
  return (
    <div
      data-testid="generating-phase-list"
      style={{
        display: "flex",
        flexDirection: "column",
        gap: 0,
        marginBottom: 32,
      }}
    >
      {phases.map((p, i) => {
        const done = i < phaseIndex;
        const active = i === phaseIndex;
        const state = done ? "done" : active ? "active" : "pending";
        return (
          <div
            key={p.labelKey}
            data-testid={`generating-phase-${i}`}
            data-state={state}
            style={{
              display: "flex",
              gap: 12,
              padding: "10px 0",
              borderBottom:
                i < phases.length - 1 ? "1px dotted var(--ei-color-rule-soft)" : "none",
              alignItems: "center",
            }}
          >
            <div
              data-testid={`generating-phase-${i}-marker`}
              style={{
                width: 18,
                height: 18,
                borderRadius: 9,
                flexShrink: 0,
                background: done
                  ? "var(--ei-color-ok)"
                  : active
                    ? "var(--ei-color-accent)"
                    : "transparent",
                border: `1.5px solid ${
                  done ? "var(--ei-color-ok)" : active ? "var(--ei-color-accent)" : "var(--ei-color-rule-soft)"
                }`,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              {done ? (
                <span
                  aria-hidden="true"
                  style={{ color: "#fff", fontSize: 11, lineHeight: 1 }}
                >
                  ✓
                </span>
              ) : null}
              {active ? (
                <div
                  className="ei-pulse"
                  style={{
                    width: 5,
                    height: 5,
                    borderRadius: 3,
                    background: "#fff",
                  }}
                />
              ) : null}
            </div>
            <div
              data-testid={`generating-phase-${i}-label`}
              style={{
                fontSize: 13.5,
                color: done
                  ? "var(--ei-color-fg-tertiary)"
                  : active
                    ? "var(--ei-color-fg-primary)"
                    : "var(--ei-ink4, var(--ei-color-fg-tertiary))",
                flex: 1,
                textDecoration: done ? "line-through" : "none",
              }}
            >
              {resolve(p.labelKey)}
            </div>
            <div
              data-testid={`generating-phase-${i}-hint`}
              style={{
                fontSize: 11,
                color: "var(--ei-ink4, var(--ei-color-fg-tertiary))",
                fontFamily: "var(--ei-font-mono)",
                letterSpacing: "0.04em",
              }}
            >
              {active ? <span className="ei-pulse">●</span> : ""}{" "}
              {resolve(p.hintKey)}
            </div>
          </div>
        );
      })}
    </div>
  );
};
