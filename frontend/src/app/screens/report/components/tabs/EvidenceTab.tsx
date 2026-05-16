import type { FC } from "react";

import type { FeedbackReport } from "../../../../../api/generated/types";
import { useI18n } from "../../../../i18n/messages";

interface EvidenceTabProps {
  report: FeedbackReport;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportDetailSurface
 * evidence branch (lines 445-467). Two columns: RISK EVIDENCE on the left,
 * REUSABLE PROOF on the right. Empty arrays surface an EmptyHint.
 */
export const EvidenceTab: FC<EvidenceTabProps> = ({ report }) => {
  const { t } = useI18n();
  const risks = report.issues ?? [];
  const highlights = report.highlights ?? [];
  return (
    <div
      data-testid="report-evidence-panel"
      style={{
        padding: 24,
        display: "grid",
        gridTemplateColumns: "1fr 1fr",
        gap: 18,
      }}
    >
      <div data-testid="report-evidence-risk-column">
        <div
          className="ei-label"
          style={{ color: "var(--ei-danger, var(--ei-ink))", marginBottom: 12 }}
        >
          {t("report.evidence.risks.eyebrow")}
        </div>
        {risks.length === 0 ? (
          <div data-testid="report-evidence-risk-empty" style={{ color: "var(--ei-ink3)" }}>
            {t("report.evidence.risks.empty")}
          </div>
        ) : (
          risks.map((issue, idx) => (
            <div
              key={idx}
              data-testid={`report-evidence-risk-${idx}`}
              style={{
                padding: "14px 0",
                borderBottom:
                  idx < risks.length - 1
                    ? "1px dotted var(--ei-rule)"
                    : "none",
              }}
            >
              <div style={{ fontSize: 15, color: "var(--ei-ink)", fontWeight: 500 }}>
                {issue.dimension}
              </div>
              <div
                style={{
                  fontSize: 12,
                  color: "var(--ei-ink3)",
                  fontFamily: "var(--ei-mono)",
                  margin: "4px 0 8px",
                }}
              >
                conf · {issue.confidence ?? "—"}
              </div>
              <div
                style={{
                  fontSize: 13,
                  color: "var(--ei-ink2, var(--ei-ink))",
                  lineHeight: 1.6,
                }}
              >
                {issue.evidence}
              </div>
            </div>
          ))
        )}
      </div>
      <div data-testid="report-evidence-highlight-column">
        <div
          className="ei-label"
          style={{ color: "var(--ei-ok)", marginBottom: 12 }}
        >
          {t("report.evidence.highlights.eyebrow")}
        </div>
        {highlights.length === 0 ? (
          <div
            data-testid="report-evidence-highlight-empty"
            style={{ color: "var(--ei-ink3)" }}
          >
            {t("report.evidence.highlights.empty")}
          </div>
        ) : (
          highlights.map((highlight, idx) => (
            <div
              key={idx}
              data-testid={`report-evidence-highlight-${idx}`}
              style={{
                padding: "14px 0",
                borderBottom:
                  idx < highlights.length - 1
                    ? "1px dotted var(--ei-rule)"
                    : "none",
              }}
            >
              <div style={{ fontSize: 15, color: "var(--ei-ink)", fontWeight: 500 }}>
                {highlight.dimension}
              </div>
              <div
                style={{
                  fontSize: 13,
                  color: "var(--ei-ink2, var(--ei-ink))",
                  lineHeight: 1.6,
                  marginTop: 4,
                }}
              >
                {highlight.evidence}
              </div>
              <div
                style={{
                  fontSize: 12,
                  color: "var(--ei-ink3)",
                  fontFamily: "var(--ei-mono)",
                  marginTop: 6,
                }}
              >
                conf · {highlight.confidence ?? "—"}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};
