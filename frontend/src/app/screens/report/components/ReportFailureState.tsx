import type { FC } from "react";

import type { ApiErrorCode } from "../../../../api/generated/types";
import { useI18n } from "../../../i18n/messages";
import { failureErrorCodeKey } from "../readiness";

interface ReportFailureStateProps {
  errorCode: ApiErrorCode | string | null;
  /** True when the failure source is REPORT_NOT_FOUND (HTTP 404 cross-user). */
  notFound?: boolean;
  onRetry: () => void;
  onBackToWorkspace: () => void;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportFailureState
 * (lines 61-77). Two distinct copy decks:
 *  - AI_* enum failures map through failureErrorCodeKey() to error-specific text.
 *  - REPORT_NOT_FOUND uses dedicated `failureState.notFound.{title,desc}` keys
 *    so cross-user 404s never leak existence of another user's report.
 */
export const ReportFailureState: FC<ReportFailureStateProps> = ({
  errorCode,
  notFound,
  onRetry,
  onBackToWorkspace,
}) => {
  const { t } = useI18n();
  const isNotFound = notFound || errorCode === "REPORT_NOT_FOUND";
  const titleKey = isNotFound
    ? "report.failureState.notFound.title"
    : "report.failureState.title";
  const descKey = isNotFound
    ? "report.failureState.notFound.desc"
    : "report.failureState.desc";
  const codeLabel = isNotFound
    ? t("report.failureState.errorCode.REPORT_NOT_FOUND")
    : t(failureErrorCodeKey(errorCode));
  return (
    <div
      data-testid="report-failure-state"
      data-not-found={isNotFound ? "true" : "false"}
      className="ei-fadein"
      style={{
        maxWidth: 820,
        margin: "0 auto",
        padding: "72px 48px",
      }}
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
          data-testid={
            isNotFound
              ? "report-failure-state-not-found-eyebrow"
              : "report-failure-eyebrow"
          }
          style={{ color: "var(--ei-danger, var(--ei-ink))", marginBottom: 10 }}
        >
          {t(
            isNotFound
              ? "report.failureState.notFound.eyebrow"
              : "report.failureState.eyebrow",
          )}
        </div>
        <div
          className="ei-serif"
          data-testid={
            isNotFound
              ? "report-failure-state-not-found-title"
              : "report-failure-title"
          }
          style={{
            fontSize: 28,
            color: "var(--ei-ink)",
            lineHeight: 1.25,
            marginBottom: 10,
          }}
        >
          {t(titleKey)}
        </div>
        <div
          data-testid={
            isNotFound
              ? "report-failure-state-not-found-desc"
              : "report-failure-desc"
          }
          style={{
            fontSize: 14,
            color: "var(--ei-ink3)",
            lineHeight: 1.6,
            marginBottom: 18,
          }}
        >
          {t(descKey)}
        </div>
        <div
          data-testid="report-failure-error-code"
          style={{
            fontSize: 12,
            color: "var(--ei-ink3)",
            fontFamily: "var(--ei-mono)",
            marginBottom: 18,
            letterSpacing: "0.04em",
          }}
        >
          {codeLabel}
        </div>
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
          {isNotFound ? null : (
            <button
              type="button"
              data-testid="report-failure-retry-cta"
              onClick={onRetry}
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
              {t("report.failureState.retry")}
            </button>
          )}
          <button
            type="button"
            data-testid="report-failure-back-to-workspace"
            onClick={onBackToWorkspace}
            style={{
              padding: "10px 16px",
              background: "transparent",
              color: "var(--ei-ink2, var(--ei-ink))",
              border: "1px solid var(--ei-rule)",
              borderRadius: 2,
              cursor: "pointer",
              fontFamily: "var(--ei-sans)",
              fontSize: 13,
            }}
          >
            {t("report.failureState.backToWorkspace")}
          </button>
        </div>
      </div>
    </div>
  );
};
