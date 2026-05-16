import type { FC } from "react";

import type {
  DimensionStatus,
  FeedbackReport,
} from "../../../../../api/generated/types";
import { useI18n } from "../../../../i18n/messages";
import { dimensionStatusLabel } from "../../readiness";
import { DimRow } from "./DimRow";

interface DimensionsTabProps {
  report: FeedbackReport;
}

interface DimensionAggregate {
  name: string;
  status: DimensionStatus | null;
  score: number;
  confidence: number;
}

function aggregateDimensions(report: FeedbackReport): DimensionAggregate[] {
  const map = new Map<string, { strong: number; meets: number; needs: number; samples: number; confidence: number }>();
  for (const q of report.questionAssessments ?? []) {
    const dims = q.dimensionResults ?? {};
    for (const [name, raw] of Object.entries(dims)) {
      const cell = (raw ?? {}) as { status?: string; confidence?: string };
      const acc =
        map.get(name) ?? { strong: 0, meets: 0, needs: 0, samples: 0, confidence: 0 };
      if (cell.status === "strong") acc.strong += 1;
      else if (cell.status === "meets_bar") acc.meets += 1;
      else if (cell.status === "needs_work") acc.needs += 1;
      acc.samples += 1;
      acc.confidence += cell.confidence === "high" ? 1 : cell.confidence === "medium" ? 0.65 : 0.3;
      map.set(name, acc);
    }
  }
  const out: DimensionAggregate[] = [];
  for (const [name, acc] of map.entries()) {
    const status: DimensionStatus = acc.strong > acc.needs && acc.strong > acc.meets
      ? "strong"
      : acc.needs > acc.meets
        ? "needs_work"
        : "meets_bar";
    const score = Math.round(
      ((acc.strong + acc.meets * 0.65) / Math.max(acc.samples, 1)) * 100,
    );
    const confidence = acc.samples ? acc.confidence / acc.samples : 0;
    out.push({ name, status, score, confidence });
  }
  return out;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportDetailSurface
 * dimensions branch (lines 360-382). Two-column grid of dimension cards with
 * DimRow score bar + status tag.
 */
export const DimensionsTab: FC<DimensionsTabProps> = ({ report }) => {
  const { t } = useI18n();
  const dimensions = aggregateDimensions(report);
  return (
    <div data-testid="report-dimensions-panel" style={{ padding: 24 }}>
      <div
        className="ei-label"
        style={{ color: "var(--ei-ink3)", marginBottom: 14 }}
      >
        {t("report.dimensions.detail.eyebrow")}
      </div>
      {dimensions.length === 0 ? (
        <div data-testid="report-dimensions-empty" style={{ color: "var(--ei-ink3)" }}>
          {t("report.dimensions.empty")}
        </div>
      ) : (
        <div
          data-testid="report-dimensions-grid"
          style={{ display: "grid", gridTemplateColumns: "repeat(2, 1fr)", gap: 14 }}
        >
          {dimensions.map((d, idx) => (
            <div
              key={d.name}
              data-testid={`report-dim-card-${idx}`}
              data-dim-name={d.name}
              data-dim-status={d.status ?? "unknown"}
              data-dim-score={d.score}
              data-dim-confidence={d.confidence}
              style={{
                padding: 16,
                border: "1px solid var(--ei-rule)",
                borderRadius: 2,
                background: "var(--ei-bg)",
              }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  gap: 10,
                  marginBottom: 10,
                }}
              >
                <div style={{ fontSize: 15, color: "var(--ei-ink)", fontWeight: 500 }}>
                  {d.name}
                </div>
                <div
                  data-testid={`report-dim-card-${idx}-status`}
                  style={{
                    fontSize: 11,
                    fontFamily: "var(--ei-mono)",
                    color:
                      d.status === "strong"
                        ? "var(--ei-ok)"
                        : d.status === "needs_work"
                          ? "var(--ei-danger, var(--ei-ink))"
                          : "var(--ei-ink2, var(--ei-ink))",
                  }}
                >
                  {d.status ? t(dimensionStatusLabel(d.status)) : "—"}
                </div>
              </div>
              <DimRow dim={d} last />
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
