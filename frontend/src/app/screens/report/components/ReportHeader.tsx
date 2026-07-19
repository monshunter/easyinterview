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
  replayVariant?: "accent" | "secondary";
  nextVariant?: "accent" | "secondary";
  nextDisabledReason?: string;
}

/**
 * Source-level mirror of formal frontend implementation ReportDashboard
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
  replayVariant = "secondary",
  nextVariant = "secondary",
  nextDisabledReason,
}) => {
  const { t } = useI18n();
  return (
    <header data-testid="report-header" className="ei-report-header">
      <div className="ei-report-header-copy">
        <div
          className="ei-report-header-breadcrumb"
          data-testid="report-header-breadcrumb"
        >
          {breadcrumb}
        </div>
        <h1
          className="ei-report-header-title"
          data-testid="report-header-title"
        >
          {title}
        </h1>
        <p
          data-testid="report-header-subtitle"
          className="ei-report-header-subtitle"
        >
          {subtitle}
        </p>
      </div>
      <div className="ei-report-header-actions">
        <button
          type="button"
          data-testid="report-replay-cta"
          onClick={onReplay}
          disabled={disableReplay}
          aria-disabled={disableReplay}
          aria-describedby={disableReplay && nextDisabledReason ? "report-next-disabled-reason" : undefined}
          data-variant={replayVariant}
          className="ei-report-header-cta"
        >
          <span data-testid="report-replay-icon" className="ei-report-header-cta-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M4 12a8 8 0 1 0 2.34-5.66L4 8.68" /><path d="M4 4v4.68h4.68" /></svg>
          </span>
          {t("report.header.cta.replay")}
        </button>
        <button
          type="button"
          data-testid="report-next-cta"
          onClick={onNextRound}
          disabled={disableNextRound}
          aria-disabled={disableNextRound}
          aria-describedby={disableNextRound && nextDisabledReason ? "report-next-disabled-reason" : undefined}
          data-variant={nextVariant}
          className="ei-report-header-cta"
        >
          {t("report.header.cta.nextRound")}
          <span data-testid="report-next-icon" className="ei-report-header-cta-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h13" /><path d="m14 7 5 5-5 5" /></svg>
          </span>
        </button>
        {disableNextRound && nextDisabledReason ? (
          <span id="report-next-disabled-reason" data-testid="report-next-disabled-reason" className="ei-report-header-disabled-reason">
            {nextDisabledReason}
          </span>
        ) : null}
      </div>
    </header>
  );
};
