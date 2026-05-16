import type { FC } from "react";

import type { FeedbackReport } from "../../../../../api/generated/types";
import { useI18n, type MessageKey } from "../../../../i18n/messages";
import { readinessTierLabel } from "../../readiness";

interface ReadinessTabProps {
  report: FeedbackReport;
}

const DETAIL_ROWS: ReadonlyArray<{
  testId: string;
  labelKey: MessageKey;
  valueKey: MessageKey;
  bodyKey: MessageKey;
}> = [
  {
    testId: "report-readiness-jd-align",
    labelKey: "report.readiness.detail.jdAlign.label",
    valueKey: "report.readiness.detail.jdAlign.value",
    bodyKey: "report.readiness.detail.jdAlign.body",
  },
  {
    testId: "report-readiness-evidence-density",
    labelKey: "report.readiness.detail.evidenceDensity.label",
    valueKey: "report.readiness.detail.evidenceDensity.value",
    bodyKey: "report.readiness.detail.evidenceDensity.body",
  },
  {
    testId: "report-readiness-next-threshold",
    labelKey: "report.readiness.detail.nextThreshold.label",
    valueKey: "report.readiness.detail.nextThreshold.value",
    bodyKey: "report.readiness.detail.nextThreshold.body",
  },
];

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportDetailSurface
 * readiness branch (lines 335-357). Dial + 3 detail cards.
 */
export const ReadinessTab: FC<ReadinessTabProps> = ({ report }) => {
  const { t } = useI18n();
  const tier = report.preparednessLevel;
  return (
    <div
      data-testid="report-readiness-panel"
      style={{
        padding: 24,
        display: "grid",
        gridTemplateColumns: "220px 1fr",
        gap: 28,
        alignItems: "center",
      }}
    >
      <div
        data-testid="report-readiness-dial"
        data-tier={tier ?? "unknown"}
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: 8,
        }}
      >
        <div
          aria-hidden="true"
          style={{
            width: 112,
            height: 112,
            borderRadius: "50%",
            border: "6px solid var(--ei-color-accent)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontFamily: "var(--ei-serif)",
            fontSize: 18,
            color: "var(--ei-color-fg-primary)",
          }}
        >
          {tier ? t(readinessTierLabel(tier)) : "—"}
        </div>
        <div
          data-testid="report-readiness-dial-label"
          style={{
            fontFamily: "var(--ei-font-mono)",
            fontSize: 11,
            color: "var(--ei-color-fg-tertiary)",
          }}
        >
          {tier ?? "—"}
        </div>
      </div>
      <div>
        <div
          className="ei-label"
          style={{ color: "var(--ei-color-accent)", marginBottom: 8 }}
        >
          {t("report.readiness.detail.eyebrow")}
        </div>
        <div
          className="ei-serif"
          style={{
            fontSize: 24,
            color: "var(--ei-color-fg-primary)",
            lineHeight: 1.35,
            marginBottom: 14,
          }}
        >
          {t("report.readiness.detail.title")}
        </div>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(3, 1fr)",
            gap: 12,
          }}
        >
          {DETAIL_ROWS.map((row) => (
            <div
              key={row.testId}
              data-testid={row.testId}
              style={{
                padding: 14,
                background: "var(--ei-color-bg-soft, var(--ei-color-bg-canvas))",
                borderRadius: 2,
              }}
            >
              <div
                className="ei-label"
                style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 6 }}
              >
                {t(row.labelKey)}
              </div>
              <div
                className="ei-serif"
                style={{
                  fontSize: 22,
                  color: "var(--ei-color-fg-primary)",
                  marginBottom: 6,
                }}
              >
                {t(row.valueKey)}
              </div>
              <div
                style={{
                  fontSize: 12.5,
                  color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
                  lineHeight: 1.55,
                }}
              >
                {t(row.bodyKey)}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
