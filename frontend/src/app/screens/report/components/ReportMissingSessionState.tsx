import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

interface ReportMissingSessionStateProps {
  onBackToWorkspace: () => void;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportMissingSessionState
 * (lines 46-59). Renders when the route enters /report without a sessionId — we
 * never invent placeholder report data, only direct the user back to history.
 */
export const ReportMissingSessionState: FC<ReportMissingSessionStateProps> = ({
  onBackToWorkspace,
}) => {
  const { t } = useI18n();
  return (
    <div
      data-testid="report-missing-session"
      className="ei-fadein"
      style={{ maxWidth: 820, margin: "0 auto", padding: "72px 48px" }}
    >
      <div
        style={{
          border: "1px solid var(--ei-rule)",
          borderRadius: 3,
          padding: 28,
          background: "var(--ei-bg-card, var(--ei-bg))",
        }}
      >
        <div
          className="ei-label"
          data-testid="report-missing-session-eyebrow"
          style={{ color: "var(--ei-ink3)", marginBottom: 10 }}
        >
          {t("report.missingSession.eyebrow")}
        </div>
        <div
          className="ei-serif"
          data-testid="report-missing-session-title"
          style={{
            fontSize: 28,
            color: "var(--ei-ink)",
            lineHeight: 1.25,
            marginBottom: 10,
          }}
        >
          {t("report.missingSession.title")}
        </div>
        <div
          data-testid="report-missing-session-desc"
          style={{
            fontSize: 14,
            color: "var(--ei-ink3)",
            lineHeight: 1.6,
            marginBottom: 18,
          }}
        >
          {t("report.missingSession.desc")}
        </div>
        <button
          type="button"
          data-testid="report-missing-session-cta"
          onClick={onBackToWorkspace}
          style={{
            padding: "10px 16px",
            background: "var(--ei-accent)",
            color: "#fff",
            border: "1px solid var(--ei-accent)",
            borderRadius: 2,
            cursor: "pointer",
            fontFamily: "var(--ei-sans)",
            fontSize: 13,
          }}
        >
          {t("report.missingSession.cta")}
        </button>
      </div>
    </div>
  );
};
