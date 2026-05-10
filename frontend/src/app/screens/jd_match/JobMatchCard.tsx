import type { FC } from "react";

import { useI18n } from "../../i18n/messages";
import type { JobMatchRecommendation } from "../../../api/generated/types";

export type ScoreBand = "strong" | "good" | "stretch";

export function resolveScoreBand(score: number): ScoreBand {
  if (score >= 85) return "strong";
  if (score >= 70) return "good";
  return "stretch";
}

function scoreBandColor(band: ScoreBand): string {
  switch (band) {
    case "strong":
      return "var(--ei-color-ok)";
    case "good":
      return "var(--ei-color-warn)";
    case "stretch":
    default:
      return "var(--ei-color-fg-tertiary)";
  }
}

export interface JobMatchCardProps {
  recommendation: JobMatchRecommendation;
  active: boolean;
  onClick: () => void;
}

export const JobMatchCard: FC<JobMatchCardProps> = ({
  recommendation,
  active,
  onClick,
}) => {
  const { t } = useI18n();
  const band = resolveScoreBand(recommendation.score);
  const scoreColor = scoreBandColor(band);
  const scoreLabel = (() => {
    switch (band) {
      case "strong":
        return t("jdMatch.recommended.scoreLabelStrong");
      case "good":
        return t("jdMatch.recommended.scoreLabelGood");
      case "stretch":
      default:
        return t("jdMatch.recommended.scoreLabelStretch");
    }
  })();
  const topReason = recommendation.reasons[0] ?? "";

  return (
    <button
      type="button"
      data-testid={`jdmatch-card-${recommendation.id}`}
      data-active={active ? "true" : "false"}
      data-score-band={band}
      onClick={onClick}
      style={{
        padding: "18px 20px",
        textAlign: "left",
        cursor: "pointer",
        background: active
          ? "var(--ei-color-accent-soft)"
          : "var(--ei-color-bg-card)",
        borderTop: `1px solid ${
          active ? "var(--ei-color-accent)" : "var(--ei-color-rule-strong)"
        }`,
        borderRight: `1px solid ${
          active ? "var(--ei-color-accent)" : "var(--ei-color-rule-strong)"
        }`,
        borderBottom: `1px solid ${
          active ? "var(--ei-color-accent)" : "var(--ei-color-rule-strong)"
        }`,
        borderLeft: `3px solid ${scoreColor}`,
        borderRadius: "var(--ei-radius-sm)",
        fontFamily: "var(--ei-font-sans)",
        width: "100%",
        boxSizing: "border-box",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          gap: 14,
          marginBottom: 8,
        }}
      >
        <div style={{ flex: 1, minWidth: 0 }}>
          <div
            style={{
              display: "flex",
              gap: 8,
              alignItems: "center",
              marginBottom: 4,
            }}
          >
            <div
              data-testid={`jdmatch-card-${recommendation.id}-title`}
              className="ei-serif"
              style={{
                fontSize: 16.5,
                color: "var(--ei-color-fg-primary)",
                letterSpacing: "-0.01em",
                fontWeight: 500,
              }}
            >
              {recommendation.title}
            </div>
            {!recommendation.seen ? (
              <span
                data-testid={`jdmatch-card-${recommendation.id}-unseen-dot`}
                aria-label={t("jdMatch.recommended.unseenDot")}
                style={{
                  width: 6,
                  height: 6,
                  borderRadius: 3,
                  background: "var(--ei-color-accent)",
                  display: "inline-block",
                }}
              />
            ) : null}
            {recommendation.saved ? (
              <span
                data-testid={`jdmatch-card-${recommendation.id}-saved-pin`}
                aria-label={t("jdMatch.recommended.savedPin")}
                style={{
                  fontSize: 11,
                  color: "var(--ei-color-accent)",
                }}
              >
                ●
              </span>
            ) : null}
          </div>
          <div
            data-testid={`jdmatch-card-${recommendation.id}-company`}
            style={{
              fontSize: 12.5,
              color: "var(--ei-color-fg-secondary)",
              marginBottom: 2,
            }}
          >
            {recommendation.company}
            {recommendation.companyTag ? (
              <span style={{ color: "var(--ei-color-fg-tertiary)" }}>
                {" · "}
                {recommendation.companyTag}
              </span>
            ) : null}
          </div>
          <div
            data-testid={`jdmatch-card-${recommendation.id}-meta`}
            style={{
              fontSize: 11.5,
              color: "var(--ei-color-fg-tertiary)",
              fontFamily: "var(--ei-font-mono)",
              letterSpacing: "0.02em",
            }}
          >
            {recommendation.location}
            {recommendation.comp ? ` · ${recommendation.comp}` : ""}
          </div>
        </div>
        <div style={{ textAlign: "right", flexShrink: 0 }}>
          <div
            data-testid={`jdmatch-card-${recommendation.id}-score`}
            className="ei-serif"
            style={{
              fontSize: 32,
              color: scoreColor,
              fontWeight: 500,
              lineHeight: 1,
              letterSpacing: "-0.02em",
            }}
          >
            {recommendation.score}
          </div>
          <div
            data-testid={`jdmatch-card-${recommendation.id}-label`}
            style={{
              fontSize: 9.5,
              color: scoreColor,
              fontFamily: "var(--ei-font-mono)",
              letterSpacing: "0.08em",
              marginTop: 3,
            }}
          >
            {scoreLabel}
          </div>
        </div>
      </div>
      <div
        data-testid={`jdmatch-card-${recommendation.id}-top-reason`}
        style={{
          marginTop: 10,
          padding: "8px 10px",
          background: active ? "var(--ei-color-bg)" : "var(--ei-color-bg-soft)",
          borderLeft: "2px solid var(--ei-color-accent)",
          fontSize: 12.5,
          color: "var(--ei-color-fg-secondary)",
          lineHeight: 1.45,
        }}
      >
        <span style={{ color: "var(--ei-color-accent)", fontWeight: 500 }}>
          {t("jdMatch.recommended.topReasonLabel")}
          {" · "}
        </span>
        {topReason}
      </div>
      <div
        data-testid={`jdmatch-card-${recommendation.id}-fit-footer`}
        style={{
          marginTop: 10,
          display: "flex",
          gap: 14,
          fontSize: 11,
          color: "var(--ei-color-fg-tertiary)",
          fontFamily: "var(--ei-font-mono)",
          letterSpacing: "0.04em",
        }}
      >
        <span>
          {t("jdMatch.recommended.fitMustTemplate")}{" "}
          {recommendation.fit.must}/{recommendation.fit.total}
        </span>
        <span>
          + {t("jdMatch.recommended.fitPlusTemplate")}{" "}
          {recommendation.fit.plus}/{recommendation.fit.totalPlus}
        </span>
        <span>{recommendation.posted}</span>
        <span
          style={{
            flex: 1,
            textAlign: "right",
            color: "var(--ei-color-fg-muted)",
          }}
        >
          {recommendation.sourceLabel ?? ""}
        </span>
      </div>
    </button>
  );
};
