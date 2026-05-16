import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

interface ReportHeaderProps {
  breadcrumb: string;
  title: string;
  subtitle: string;
  onReplay: () => void;
  onNextRound: () => void;
  disableReplay?: boolean;
  disableNextRound?: boolean;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx ReportDashboard
 * header block (lines 121-143). Carries breadcrumb + title + subtitle + the
 * paired Replay / Next Round CTAs.
 */
export const ReportHeader: FC<ReportHeaderProps> = ({
  breadcrumb,
  title,
  subtitle,
  onReplay,
  onNextRound,
  disableReplay,
  disableNextRound,
}) => {
  const { t } = useI18n();
  return (
    <div
      data-testid="report-header"
      style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "flex-end",
        gap: 24,
        marginBottom: 18,
        flexWrap: "wrap",
      }}
    >
      <div>
        <div
          className="ei-label"
          data-testid="report-header-breadcrumb"
          style={{ color: "var(--ei-ink3)", marginBottom: 8 }}
        >
          {breadcrumb}
        </div>
        <h1
          className="ei-serif"
          data-testid="report-header-title"
          style={{
            fontSize: 38,
            color: "var(--ei-ink)",
            margin: 0,
            lineHeight: 1.15,
            letterSpacing: "-0.02em",
          }}
        >
          {title}
        </h1>
        <div
          data-testid="report-header-subtitle"
          style={{
            fontSize: 14,
            color: "var(--ei-ink2, var(--ei-ink))",
            marginTop: 8,
            lineHeight: 1.65,
            maxWidth: 720,
          }}
        >
          {subtitle}
        </div>
      </div>
      <div
        style={{
          display: "flex",
          gap: 10,
          flexWrap: "wrap",
          justifyContent: "flex-end",
        }}
      >
        <button
          type="button"
          data-testid="report-replay-cta"
          onClick={onReplay}
          disabled={disableReplay}
          aria-disabled={disableReplay}
          style={{
            padding: "10px 16px",
            background: "var(--ei-accent)",
            color: "#fff",
            border: "1px solid var(--ei-accent)",
            borderRadius: 2,
            cursor: disableReplay ? "not-allowed" : "pointer",
            opacity: disableReplay ? 0.5 : 1,
            fontFamily: "var(--ei-sans)",
            fontSize: 13,
          }}
        >
          {t("report.header.cta.replay")}
        </button>
        <button
          type="button"
          data-testid="report-next-cta"
          onClick={onNextRound}
          disabled={disableNextRound}
          aria-disabled={disableNextRound}
          style={{
            padding: "10px 16px",
            background: "transparent",
            color: "var(--ei-ink2, var(--ei-ink))",
            border: "1px solid var(--ei-rule)",
            borderRadius: 2,
            cursor: disableNextRound ? "not-allowed" : "pointer",
            opacity: disableNextRound ? 0.5 : 1,
            fontFamily: "var(--ei-sans)",
            fontSize: 13,
          }}
        >
          {t("report.header.cta.nextRound")}
        </button>
      </div>
    </div>
  );
};
