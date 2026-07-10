import type { FC } from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";

type MissingStateKind = "missingSession" | "missingReport";

interface ReportMissingSessionStateProps {
  kind?: MissingStateKind;
  onBackToWorkspace: () => void;
}

const MISSING_STATE_COPY = {
  missingSession: {
    eyebrow: "report.missingSession.eyebrow",
    title: "report.missingSession.title",
    desc: "report.missingSession.desc",
    cta: "report.missingSession.cta",
    testIdPrefix: "report-missing-session",
  },
  missingReport: {
    eyebrow: "report.missingReport.eyebrow",
    title: "report.missingReport.title",
    desc: "report.missingReport.desc",
    cta: "report.missingReport.cta",
    testIdPrefix: "report-missing-report",
  },
} as const satisfies Record<
  MissingStateKind,
  {
    eyebrow: MessageKey;
    title: MessageKey;
    desc: MessageKey;
    cta: MessageKey;
    testIdPrefix: string;
  }
>;

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportMissingSessionState
 * (lines 46-59). Renders when the route enters /report without a sessionId - we
 * never invent synthetic report data, only direct the user back to history.
 */
export const ReportMissingSessionState: FC<ReportMissingSessionStateProps> = ({
  kind = "missingSession",
  onBackToWorkspace,
}) => {
  const { t } = useI18n();
  const copy = MISSING_STATE_COPY[kind];
  const { testIdPrefix } = copy;
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
          {t(copy.eyebrow)}
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
          {t(copy.title)}
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
          {t(copy.desc)}
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
          {t(copy.cta)}
        </button>
      </div>
    </div>
  );
};
