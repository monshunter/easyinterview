import type { FC } from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";

export type GeneratingErrorKind =
  | "missingReportId"
  | "timeout"
  | "loadFailed"
  | "failed"
  | "notFound"
  | "invalidReport"
  | "contextTooLarge";

interface GeneratingErrorStateProps {
  kind: GeneratingErrorKind;
  onRetry?: () => void;
  onBack: () => void;
  backDestination: "reports" | "workspace";
}

const TITLE_KEY: Record<GeneratingErrorKind, MessageKey> = {
  missingReportId: "generating.errors.missingReportId.title",
  timeout: "generating.errors.timeout.title",
  loadFailed: "generating.errors.loadFailed.title",
  failed: "generating.errors.failed.title",
  notFound: "generating.errors.notFound.title",
  invalidReport: "generating.errors.invalidReport.title",
  contextTooLarge: "generating.errors.contextTooLarge.title",
};

const DESC_KEY: Record<GeneratingErrorKind, MessageKey> = {
  missingReportId: "generating.errors.missingReportId.desc",
  timeout: "generating.errors.timeout.desc",
  loadFailed: "generating.errors.loadFailed.desc",
  failed: "generating.errors.failed.desc",
  notFound: "generating.errors.notFound.desc",
  invalidReport: "generating.errors.invalidReport.desc",
  contextTooLarge: "generating.errors.contextTooLarge.desc",
};

/**
 * Shared error surface for GeneratingScreen. Replaces the success layout when:
 * - reportId is missing in route params (missingReportId)
 * - max attempts hit without a `ready` / `failed` status (timeout)
 * - one-off internal errors (loadFailed)
 */
export const GeneratingErrorState: FC<GeneratingErrorStateProps> = ({
  kind,
  onRetry,
  onBack,
  backDestination,
}) => {
  const { t } = useI18n();
  const showRetry = (kind === "timeout" || kind === "loadFailed") && onRetry !== undefined;
  return (
    <div
      data-testid="generating-screen"
      aria-live="polite"
      className="ei-fadein"
      style={{
        minHeight: "100vh",
        background: "var(--ei-color-bg-canvas)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: "48px clamp(16px, 6vw, 48px)",
      }}
    >
      <div
        data-testid="generating-error-state"
        data-error-kind={kind}
        style={{ maxWidth: 780, width: "100%" }}
      >
        <div
          className="ei-label"
          data-testid="generating-header-eyebrow"
          style={{
            color: "var(--ei-color-fg-tertiary)",
            marginBottom: 12,
            letterSpacing: "0.1em",
          }}
        >
          <span data-testid="generating-error-eyebrow">
            {t(showRetry ? "generating.errors.checkPaused" : "generating.errors.eyebrow")}
          </span>
        </div>
        <h1
          className="ei-serif"
          data-testid="generating-header-title"
          style={{
            fontSize: 34,
            margin: 0,
            color: "var(--ei-color-fg-primary)",
            letterSpacing: "-0.02em",
            lineHeight: 1.2,
            marginBottom: 10,
          }}
        >
          <span data-testid="generating-error-title">{t(TITLE_KEY[kind])}</span>
        </h1>
        <div
          data-testid="generating-header-subtitle"
          style={{
            fontSize: 14,
            color: "var(--ei-color-fg-tertiary)",
            lineHeight: 1.65,
            maxWidth: 600,
          }}
        >
          <span data-testid="generating-error-desc">{t(DESC_KEY[kind])}</span>
        </div>
        <div
          style={{
            marginTop: 28,
            paddingTop: 16,
            borderTop: "1px solid var(--ei-color-rule-strong)",
            display: "flex",
            gap: 10,
            flexWrap: "wrap",
          }}
        >
          {showRetry ? (
            <button
              type="button"
              data-testid="generating-error-retry"
              onClick={onRetry}
              style={buttonStyle("accent")}
            >
              {t("generating.errors.continueCheck")}
            </button>
          ) : null}
          <span data-testid="generating-back-button">
            <button
              type="button"
              data-testid="generating-error-back-to-workspace"
              data-back-destination={backDestination}
              onClick={onBack}
              style={buttonStyle("secondary")}
            >
              {t("common.back")}
            </button>
          </span>
        </div>
      </div>
    </div>
  );
};

function buttonStyle(variant: "accent" | "secondary") {
  const accent = variant === "accent";
  return {
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
    height: 30,
    padding: "0 12px",
    fontSize: 13,
    fontWeight: 500,
    background: accent ? "var(--ei-color-accent)" : "var(--ei-color-bg-canvas)",
    color: accent ? "#fff" : "var(--ei-color-fg-primary)",
    border: accent
      ? "1px solid var(--ei-color-accent)"
      : "1px solid var(--ei-color-rule-strong)",
    borderRadius: "var(--ei-radius-control)",
    cursor: "pointer",
    fontFamily: "var(--ei-font-sans)",
    letterSpacing: "-0.005em",
    transition: "transform .08s ease, opacity .15s",
  } as const;
}
