import type { FC } from "react";

import { useI18n } from "../../i18n/messages";
import type { JobMatchRecommendation } from "../../../api/generated/types";

import { JobMatchCard } from "./JobMatchCard";
import { JDDetail } from "./JDDetail";

export interface RecommendedTabProps {
  recommendations: JobMatchRecommendation[];
  loading: boolean;
  error: Error | null;
  selectedId: string | null;
  onSelect: (jobMatchId: string) => void;
  onConfirmInterview: (rec: JobMatchRecommendation) => void;
  onToggleSave: (rec: JobMatchRecommendation) => void;
  onOpenSource: (rec: JobMatchRecommendation) => void;
  onMarkNotRelevant: (rec: JobMatchRecommendation) => void;
  onRetry?: () => void;
}

export const RecommendedTab: FC<RecommendedTabProps> = ({
  recommendations,
  loading,
  error,
  selectedId,
  onSelect,
  onConfirmInterview,
  onToggleSave,
  onOpenSource,
  onMarkNotRelevant,
  onRetry,
}) => {
  const { t } = useI18n();

  const selected =
    recommendations.find((r) => r.id === selectedId) ??
    recommendations[0] ??
    null;

  return (
    <div
      data-testid="jdmatch-recommended-tab"
      style={{
        display: "grid",
        gridTemplateColumns: "1.1fr 1.4fr",
        gap: 20,
      }}
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
        {error ? (
          <div
            data-testid="jdmatch-recommended-error"
            style={{
              padding: "20px 24px",
              background: "var(--ei-color-bg-soft)",
              border: "1px dashed var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              color: "var(--ei-color-warn)",
              fontSize: 13,
              lineHeight: 1.5,
            }}
          >
            <div style={{ marginBottom: 10 }}>
              {t("jdMatch.recommended.errorTitle")}
            </div>
            {onRetry ? (
              <button
                type="button"
                data-testid="jdmatch-recommended-error-retry"
                onClick={onRetry}
                style={{
                  padding: "6px 14px",
                  fontSize: 12,
                  fontFamily: "var(--ei-font-sans)",
                  background: "transparent",
                  color: "var(--ei-color-accent)",
                  border: "1px solid var(--ei-color-accent)",
                  borderRadius: "var(--ei-radius-sm)",
                  cursor: "pointer",
                }}
              >
                {t("jdMatch.recommended.errorRetry")}
              </button>
            ) : null}
          </div>
        ) : null}
        {!error && loading && recommendations.length === 0 ? (
          <div
            data-testid="jdmatch-recommended-loading"
            style={{
              padding: "32px 20px",
              textAlign: "center",
              fontSize: 12,
              color: "var(--ei-color-fg-tertiary)",
              fontFamily: "var(--ei-font-mono)",
              background: "var(--ei-color-bg-soft)",
              border: "1px dashed var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
            }}
          >
            …
          </div>
        ) : null}
        {!error && !loading && recommendations.length === 0 ? (
          <div
            data-testid="jdmatch-recommended-empty"
            style={{
              padding: "32px 20px",
              textAlign: "center",
              fontSize: 13,
              color: "var(--ei-color-fg-tertiary)",
              fontFamily: "var(--ei-font-sans)",
              background: "var(--ei-color-bg-soft)",
              border: "1px dashed var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              lineHeight: 1.5,
            }}
          >
            <div
              style={{
                fontSize: 14,
                color: "var(--ei-color-fg-primary)",
                marginBottom: 6,
              }}
            >
              {t("jdMatch.recommended.emptyTitle")}
            </div>
            <div>{t("jdMatch.recommended.emptyCopy")}</div>
          </div>
        ) : null}
        {!error
          ? recommendations.map((rec) => (
              <JobMatchCard
                key={rec.id}
                recommendation={rec}
                active={selected?.id === rec.id}
                onClick={() => onSelect(rec.id)}
              />
            ))
          : null}
      </div>
      <JDDetail
        recommendation={selected}
        onConfirmInterview={() => selected && onConfirmInterview(selected)}
        onToggleSave={() => selected && onToggleSave(selected)}
        onOpenSource={() => selected && onOpenSource(selected)}
        onMarkNotRelevant={() => selected && onMarkNotRelevant(selected)}
      />
    </div>
  );
};
