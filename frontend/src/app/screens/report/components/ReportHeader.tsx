import type { CSSProperties, FC } from "react";

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
    <header
      data-testid="report-header"
      style={{
        display: "flex",
        justifyContent: "space-between",
        gap: 24,
        alignItems: "flex-end",
        flexWrap: "wrap",
        marginBottom: 24,
      }}
    >
      <div style={{ minWidth: 0, flex: "1 1 440px" }}>
        <div
          className="ei-label"
          data-testid="report-header-breadcrumb"
          style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 8 }}
        >
          {breadcrumb}
        </div>
        <h1
          className="ei-serif"
          data-testid="report-header-title"
          style={{
            fontSize: 38,
            color: "var(--ei-color-fg-primary)",
            margin: 0,
            overflowWrap: "anywhere",
          }}
        >
          {title}
        </h1>
        <p
          data-testid="report-header-subtitle"
          style={{
            color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
            lineHeight: 1.7,
            marginBottom: 0,
          }}
        >
          {subtitle}
        </p>
      </div>
      <div
        style={{
          display: "flex",
          gap: 10,
          flexWrap: "wrap",
        }}
      >
        <button
          type="button"
          data-testid="report-replay-cta"
          onClick={onReplay}
          disabled={disableReplay}
          aria-disabled={disableReplay}
          aria-describedby={disableReplay && nextDisabledReason ? "report-next-disabled-reason" : undefined}
          data-variant={replayVariant}
          style={buttonStyle(replayVariant, Boolean(disableReplay))}
        >
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
          style={buttonStyle(nextVariant, Boolean(disableNextRound))}
        >
          {t("report.header.cta.nextRound")}
        </button>
        {disableNextRound && nextDisabledReason ? (
          <span id="report-next-disabled-reason" data-testid="report-next-disabled-reason" style={{ flexBasis: "100%", color: "var(--ei-color-fg-tertiary)", fontSize: 11, lineHeight: 1.35, textAlign: "right" }}>
            {nextDisabledReason}
          </span>
        ) : null}
      </div>
    </header>
  );
};

function buttonStyle(
  variant: "accent" | "secondary",
  disabled: boolean,
): CSSProperties {
  const accent = variant === "accent";
  return {
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
    height: 38,
    padding: "0 16px",
    fontSize: 14,
    fontWeight: 500,
    background: accent ? "var(--ei-color-accent)" : "var(--ei-color-bg-canvas)",
    color: accent ? "#fff" : "var(--ei-color-fg-primary)",
    border: accent ? "1px solid var(--ei-color-accent)" : "1px solid var(--ei-color-rule-strong)",
    borderRadius: 2,
    cursor: disabled ? "not-allowed" : "pointer",
    opacity: disabled ? 0.5 : 1,
    fontFamily: "var(--ei-font-sans)",
    letterSpacing: "-0.005em",
    transition: "transform .08s ease, opacity .15s",
  };
}
