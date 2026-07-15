import type { FC } from "react";

import type { ApiErrorCode } from "../../../../api/generated/types";
import { useI18n } from "../../../i18n/messages";

interface ReportFailureStateProps {
  errorCode: ApiErrorCode | string | null;
  /** True when the failure source is REPORT_NOT_FOUND (HTTP 404 cross-user). */
  notFound?: boolean;
  contractInvalid?: boolean;
  recoverable?: boolean;
  onRetry?: () => void;
  onBack: () => void;
  backDestination: "reports" | "workspace";
}

/**
 * Source-level mirror of formal frontend implementation
 * Internal provider/config error codes select no visible copy. REPORT_NOT_FOUND
 * uses dedicated user-safe state text so cross-user 404s never leak existence.
 */
export const ReportFailureState: FC<ReportFailureStateProps> = ({
  errorCode,
  notFound,
  contractInvalid,
  recoverable,
  onRetry,
  onBack,
  backDestination,
}) => {
  const { t } = useI18n();
  const isNotFound = notFound || errorCode === "REPORT_NOT_FOUND";
  const isOversize = errorCode === "REPORT_CONTEXT_TOO_LARGE";
  const titleKey = contractInvalid
    ? "report.failureState.invalidContract.title"
    : isOversize
      ? "report.failureState.contextTooLarge.title"
      : isNotFound
        ? "report.failureState.notFound.title"
        : "report.failureState.title";
  const descKey = contractInvalid
    ? "report.failureState.invalidContract.desc"
    : isOversize
      ? "report.failureState.contextTooLarge.desc"
      : isNotFound
        ? "report.failureState.notFound.desc"
        : "report.failureState.desc";
  return (
    <div
      data-testid="report-failure-state"
      data-not-found={isNotFound ? "true" : "false"}
      data-contract-invalid={contractInvalid ? "true" : "false"}
      className="ei-fadein"
      style={{
        maxWidth: 820,
        margin: "0 auto",
        padding: "72px 48px",
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
          data-testid={
            isNotFound
              ? "report-failure-state-not-found-eyebrow"
              : "report-failure-eyebrow"
          }
          style={{ color: "var(--ei-color-danger, var(--ei-color-fg-primary))", marginBottom: 10 }}
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
            color: "var(--ei-color-fg-primary)",
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
            color: "var(--ei-color-fg-tertiary)",
            lineHeight: 1.6,
            marginBottom: 18,
          }}
        >
          {t(descKey)}
        </div>
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
          {recoverable && onRetry ? (
            <button
              type="button"
              data-testid="report-failure-retry-cta"
              onClick={onRetry}
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
              {t("report.failureState.retry")}
            </button>
          ) : null}
          <button
            type="button"
            data-testid="report-failure-back-to-workspace"
            data-back-destination={backDestination}
            onClick={onBack}
            style={{
              padding: "10px 16px",
              background: "transparent",
              color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
              border: "1px solid var(--ei-color-rule-soft)",
              borderRadius: 2,
              cursor: "pointer",
              fontFamily: "var(--ei-font-sans)",
              fontSize: 13,
            }}
          >
            {t(
              backDestination === "reports"
                ? "report.failureState.backToReports"
                : "report.failureState.backToWorkspace",
            )}
          </button>
        </div>
      </div>
    </div>
  );
};
