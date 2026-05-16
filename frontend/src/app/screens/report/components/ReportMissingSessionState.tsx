import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

interface ReportMissingSessionStateProps {
  kind?: "missingSession" | "missingReport";
  onBackToWorkspace: () => void;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportMissingSessionState
 * (lines 46-59). Renders when the route enters /report without a sessionId — we
 * never invent placeholder report data, only direct the user back to history.
 */
export const ReportMissingSessionState: FC<ReportMissingSessionStateProps> = ({
  kind = "missingSession",
  onBackToWorkspace,
}) => {
  const { t } = useI18n();
  const prefix =
    kind === "missingReport" ? "report.missingReport" : "report.missingSession";
  const testIdPrefix =
    kind === "missingReport" ? "report-missing-report" : "report-missing-session";
  return (
    <div
      data-testid={testIdPrefix}
      className="ei-fadein"
      style={{ maxWidth: 820, margin: "0 auto", padding: "72px 48px" }}
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
          data-testid={`${testIdPrefix}-eyebrow`}
          style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}
        >
          {t(`${prefix}.eyebrow` as never)}
        </div>
        <div
          className="ei-serif"
          data-testid={`${testIdPrefix}-title`}
          style={{
            fontSize: 28,
            color: "var(--ei-color-fg-primary)",
            lineHeight: 1.25,
            marginBottom: 10,
          }}
        >
          {t(`${prefix}.title` as never)}
        </div>
        <div
          data-testid={`${testIdPrefix}-desc`}
          style={{
            fontSize: 14,
            color: "var(--ei-color-fg-tertiary)",
            lineHeight: 1.6,
            marginBottom: 18,
          }}
        >
          {t(`${prefix}.desc` as never)}
        </div>
        <button
          type="button"
          data-testid={`${testIdPrefix}-cta`}
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
          {t(`${prefix}.cta` as never)}
        </button>
      </div>
    </div>
  );
};
