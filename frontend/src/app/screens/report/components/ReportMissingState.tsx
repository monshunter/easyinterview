import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

interface ReportMissingStateProps {
  onBackToWorkspace: () => void;
}

/** Source-level mirror of formal frontend implementation */
export const ReportMissingState: FC<ReportMissingStateProps> = ({
  onBackToWorkspace,
}) => {
  const { t } = useI18n();
  return (
    <div
      data-testid="report-missing-report"
      className="ei-fadein"
      style={{
        maxWidth: 820,
        margin: "0 auto",
        padding: "72px clamp(16px, 5vw, 48px)",
      }}
    >
      <div
        style={{
          border: "1px solid var(--ei-color-rule-soft)",
          borderRadius: 3,
          padding: 28,
          background: "var(--ei-color-bg-card, var(--ei-color-bg-canvas))",
        }}
      >
        <div
          className="ei-label"
          data-testid="report-missing-report-eyebrow"
          style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}
        >
          {t("report.missingReport.eyebrow")}
        </div>
        <div
          className="ei-serif"
          data-testid="report-missing-report-title"
          style={{
            fontSize: 28,
            color: "var(--ei-color-fg-primary)",
            lineHeight: 1.25,
            marginBottom: 10,
          }}
        >
          {t("report.missingReport.title")}
        </div>
        <div
          data-testid="report-missing-report-desc"
          style={{
            fontSize: 14,
            color: "var(--ei-color-fg-tertiary)",
            lineHeight: 1.6,
            marginBottom: 18,
          }}
        >
          {t("report.missingReport.desc")}
        </div>
        <button
          type="button"
          data-testid="report-missing-report-cta"
          onClick={onBackToWorkspace}
          style={{
            padding: "10px 16px",
            background: "var(--ei-color-accent)",
            color: "#fff",
            border: "1px solid var(--ei-color-accent)",
            borderRadius: 2,
            cursor: "pointer",
            fontFamily: "var(--ei-font-sans)",
            fontSize: 13,
          }}
        >
          {t("report.missingReport.cta")}
        </button>
      </div>
    </div>
  );
};
