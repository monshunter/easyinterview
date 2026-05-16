import type { FC } from "react";

import type { DimensionStatus } from "../../../../../api/generated/types";
import { useI18n } from "../../../../i18n/messages";
import { dimensionStatusLabel } from "../../readiness";

export interface DimRowDimension {
  name: string;
  status: DimensionStatus | null;
  score: number;
  confidence: number;
}

interface DimRowProps {
  dim: DimRowDimension;
  last?: boolean;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::DimRow
 * (lines 565-577). Name + filled rail + status tag + confidence indicator.
 */
export const DimRow: FC<DimRowProps> = ({ dim, last }) => {
  const { t } = useI18n();
  const tone =
    dim.status === "strong"
      ? "var(--ei-color-ok)"
      : dim.status === "needs_work"
        ? "var(--ei-color-warn, var(--ei-color-danger, var(--ei-color-fg-primary)))"
        : "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))";
  return (
    <div
      data-testid="report-dim-row"
      data-dim-name={dim.name}
      data-dim-status={dim.status ?? "unknown"}
      data-dim-score={dim.score}
      style={{
        padding: "14px 0",
        display: "flex",
        alignItems: "center",
        gap: 16,
        borderBottom: last ? "none" : "1px dotted var(--ei-color-rule-soft)",
      }}
    >
      <div
        data-testid="report-dim-row-name"
        style={{ width: 110, fontSize: 13, color: "var(--ei-color-fg-primary)", fontWeight: 500 }}
      >
        {dim.name}
      </div>
      <div
        data-testid="report-dim-row-score"
        style={{
          flex: 1,
          height: 4,
          background: "var(--ei-color-bg-soft, var(--ei-color-bg-canvas))",
          borderRadius: 2,
          position: "relative",
          overflow: "hidden",
        }}
      >
        <div
          style={{
            position: "absolute",
            inset: 0,
            width: `${Math.max(0, Math.min(100, dim.score))}%`,
            background: tone,
          }}
        />
      </div>
      <div
        data-testid="report-dim-row-state"
        style={{
          width: 80,
          fontSize: 12,
          color: tone,
          fontWeight: 500,
          textAlign: "right",
        }}
      >
        {dim.status ? t(dimensionStatusLabel(dim.status)) : "—"}
      </div>
      <div
        data-testid="report-dim-row-confidence"
        style={{
          width: 70,
          fontSize: 11,
          color: "var(--ei-color-fg-tertiary)",
          fontFamily: "var(--ei-font-mono)",
          textAlign: "right",
        }}
      >
        {t("report.dimension.confidence.short")}: {dim.confidence.toFixed(2)}
      </div>
    </div>
  );
};
