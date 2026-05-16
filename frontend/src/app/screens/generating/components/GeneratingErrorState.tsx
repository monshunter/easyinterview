import type { FC } from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";

export type GeneratingErrorKind = "missingReportId" | "timeout" | "loadFailed";

interface GeneratingErrorStateProps {
  kind: GeneratingErrorKind;
  onRetry?: () => void;
  onBackToWorkspace?: () => void;
}

const TITLE_KEY: Record<GeneratingErrorKind, MessageKey> = {
  missingReportId: "generating.errors.missingReportId.title",
  timeout: "generating.errors.timeout.title",
  loadFailed: "generating.errors.loadFailed.title",
};

const DESC_KEY: Record<GeneratingErrorKind, MessageKey> = {
  missingReportId: "generating.errors.missingReportId.desc",
  timeout: "generating.errors.timeout.desc",
  loadFailed: "generating.errors.loadFailed.desc",
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
  onBackToWorkspace,
}) => {
  const { t } = useI18n();
  const showRetry = kind !== "missingReportId" && onRetry !== undefined;
  return (
    <div
      data-testid="generating-error-state"
      data-error-kind={kind}
      className="ei-fadein"
      style={{
        minHeight: "calc(100vh - 58px)",
        background: "var(--ei-color-bg-canvas)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: 48,
      }}
    >
      <div style={{ maxWidth: 540, width: "100%" }}>
        <div
          className="ei-label"
          data-testid="generating-error-eyebrow"
          style={{
            color: "var(--ei-color-danger, var(--ei-color-fg-primary))",
            marginBottom: 10,
            letterSpacing: "0.1em",
          }}
        >
          {t("generating.errors.eyebrow")}
        </div>
        <h1
          className="ei-serif"
          data-testid="generating-error-title"
          style={{
            fontSize: 28,
            margin: 0,
            color: "var(--ei-color-fg-primary)",
            lineHeight: 1.25,
            marginBottom: 10,
          }}
        >
          {t(TITLE_KEY[kind])}
        </h1>
        <div
          data-testid="generating-error-desc"
          style={{
            fontSize: 14,
            color: "var(--ei-color-fg-tertiary)",
            lineHeight: 1.6,
            marginBottom: 22,
          }}
        >
          {t(DESC_KEY[kind])}
        </div>
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
          {showRetry ? (
            <button
              type="button"
              data-testid="generating-error-retry"
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
              {t("generating.errors.retry")}
            </button>
          ) : null}
          <button
            type="button"
            data-testid="generating-error-back-to-workspace"
            onClick={onBackToWorkspace}
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
            {t("generating.errors.backToWorkspace")}
          </button>
        </div>
      </div>
    </div>
  );
};
