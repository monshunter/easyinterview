import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { ReadinessTier } from "../../../../api/generated/types";
import { readinessTierLabel } from "../readiness";

export type SummaryDetailKey =
  | "readiness"
  | "dimensions"
  | "questions"
  | "next";

interface SummaryCardsProps {
  active: SummaryDetailKey;
  onSelect: (next: SummaryDetailKey) => void;
  readinessTier?: ReadinessTier | null;
  dimensionsCount: number;
  questionsRatio: string;
}

const CARD_ORDER: SummaryDetailKey[] = [
  "readiness",
  "dimensions",
  "questions",
  "next",
];

/**
 * Source-level mirror of ui-design/src/screen-report.jsx lines 147-160. Four
 * `ReportStatButton` tiles drive `setDetail` for the DetailSurface tabs.
 */
export const SummaryCards: FC<SummaryCardsProps> = ({
  active,
  onSelect,
  readinessTier,
  dimensionsCount,
  questionsRatio,
}) => {
  const { t } = useI18n();
  const values: Record<SummaryDetailKey, string> = {
    readiness: readinessTier ? t(readinessTierLabel(readinessTier)) : "—",
    dimensions: `${dimensionsCount} ${t("report.summary.dimensions.unit")}`,
    questions: questionsRatio,
    next: t("report.summary.next.value"),
  };
  const labelKey: Record<SummaryDetailKey, "report.summary.readiness" | "report.summary.dimensions" | "report.summary.questions" | "report.summary.next"> = {
    readiness: "report.summary.readiness",
    dimensions: "report.summary.dimensions",
    questions: "report.summary.questions",
    next: "report.summary.next",
  };
  return (
    <div
      data-testid="report-summary-cards"
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(auto-fit, minmax(min(180px, 100%), 1fr))",
        gap: 14,
        marginBottom: 18,
      }}
    >
      {CARD_ORDER.map((key) => {
        const isActive = active === key;
        return (
          <button
            key={key}
            type="button"
            data-testid={`report-summary-${key}`}
            data-active={isActive ? "true" : "false"}
            onClick={() => onSelect(key)}
            style={{
              padding: 0,
              border: isActive
                ? "1px solid var(--ei-color-accent)"
                : "1px solid transparent",
              borderRadius: 3,
              background: "transparent",
              cursor: "pointer",
              textAlign: "left",
              fontFamily: "var(--ei-font-sans)",
              minWidth: 0,
              boxShadow: isActive
                ? "0 0 0 2px var(--ei-color-accent-soft, var(--ei-color-accent))"
                : "none",
            }}
          >
            <div
              style={{
                padding: "18px 20px",
                border: "1px solid var(--ei-color-rule-soft)",
                borderRadius: 2,
                background: "var(--ei-color-bg-card, var(--ei-color-bg-canvas))",
                minWidth: 0,
              }}
            >
              <div
                className="ei-label"
                style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}
              >
                {t(labelKey[key])}
              </div>
              <div
                data-testid={`report-summary-${key}-value`}
                className="ei-serif"
                style={{
                  fontSize: key === "questions" || key === "dimensions" ? 26 : 22,
                  color: "var(--ei-color-fg-primary)",
                  letterSpacing: 0,
                  overflowWrap: "anywhere",
                }}
              >
                {values[key]}
              </div>
            </div>
          </button>
        );
      })}
    </div>
  );
};
