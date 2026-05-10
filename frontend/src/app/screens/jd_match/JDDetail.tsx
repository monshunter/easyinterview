import type { FC } from "react";

import { useI18n } from "../../i18n/messages";
import type { JobMatchRecommendation } from "../../../api/generated/types";

import { resolveScoreBand } from "./JobMatchCard";

function scoreBandColor(score: number): string {
  const band = resolveScoreBand(score);
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

export interface JDDetailProps {
  recommendation: JobMatchRecommendation | null;
  onConfirmInterview: () => void;
  onToggleSave: () => void;
  onOpenSource: () => void;
  onMarkNotRelevant: () => void;
}

export const JDDetail: FC<JDDetailProps> = ({
  recommendation,
  onConfirmInterview,
  onToggleSave,
  onOpenSource,
  onMarkNotRelevant,
}) => {
  const { t } = useI18n();

  if (!recommendation) return null;

  const scoreColor = scoreBandColor(recommendation.score);
  const sourceUrl = recommendation.sourceUrl ?? null;
  const showIntel =
    Boolean(recommendation.networkNote) ||
    (recommendation.similarInterviewers != null &&
      recommendation.similarInterviewers > 0);

  const tagStyle = {
    padding: "2px 8px",
    fontSize: 11,
    fontFamily: "var(--ei-font-mono)",
    background: "var(--ei-color-bg-soft)",
    border: "1px solid var(--ei-color-rule-strong)",
    borderRadius: "var(--ei-radius-pill)",
    color: "var(--ei-color-fg-secondary)",
  } as const;

  const sectionStyle = {
    padding: "18px 24px",
    borderBottom: "1px dotted var(--ei-color-rule-strong)",
  } as const;

  return (
    <div
      data-testid="jdmatch-detail"
      className="jdmatch-detail-panel"
      style={{ position: "sticky", top: 20, alignSelf: "start" }}
    >
      <div
        style={{
          background: "var(--ei-color-bg-card)",
          border: "1px solid var(--ei-color-rule-strong)",
          borderRadius: "var(--ei-radius-sm)",
          overflow: "hidden",
        }}
      >
        {/* Header */}
        <div
          data-testid="jdmatch-detail-header"
          style={{
            padding: "20px 24px 18px",
            borderBottom: "1px solid var(--ei-color-rule-strong)",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "flex-start",
            gap: 14,
          }}
        >
          <div style={{ flex: 1, minWidth: 0 }}>
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 6 }}
            >
              {recommendation.company}
              {recommendation.companyTag
                ? ` · ${recommendation.companyTag}`
                : ""}
            </div>
            <div
              className="ei-serif"
              style={{
                fontSize: 24,
                color: "var(--ei-color-fg-primary)",
                letterSpacing: "-0.015em",
                lineHeight: 1.2,
                marginBottom: 8,
              }}
            >
              {recommendation.title}
            </div>
            <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
              {recommendation.level ? (
                <span style={tagStyle}>{recommendation.level}</span>
              ) : null}
              {recommendation.location ? (
                <span style={tagStyle}>{recommendation.location}</span>
              ) : null}
              {recommendation.comp ? (
                <span
                  style={{
                    ...tagStyle,
                    background: "var(--ei-color-accent-soft)",
                    color: "var(--ei-color-accent)",
                  }}
                >
                  {recommendation.comp}
                </span>
              ) : null}
            </div>
          </div>
          <div
            style={{
              textAlign: "center",
              padding: "6px 14px",
              background: "var(--ei-color-bg-soft)",
              borderRadius: "var(--ei-radius-sm)",
              border: "1px solid var(--ei-color-rule-strong)",
            }}
          >
            <div
              className="ei-serif"
              style={{
                fontSize: 36,
                color: scoreColor,
                fontWeight: 500,
                lineHeight: 1,
                letterSpacing: "-0.02em",
              }}
            >
              {recommendation.score}
            </div>
            <div
              style={{
                fontSize: 9.5,
                color: "var(--ei-color-fg-tertiary)",
                fontFamily: "var(--ei-font-mono)",
                letterSpacing: "0.08em",
                marginTop: 4,
              }}
            >
              {t("jdMatch.recommended.detailHeaderScoreSuffix")}
            </div>
          </div>
        </div>

        {/* Why it matches */}
        <div data-testid="jdmatch-detail-why" style={sectionStyle}>
          <div
            className="ei-label"
            style={{ color: "var(--ei-color-ok)", marginBottom: 10 }}
          >
            + {t("jdMatch.recommended.whyMatchesHeading")}
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            {recommendation.reasons.map((reason, i) => (
              <div
                key={`${reason}-${i}`}
                data-testid={`jdmatch-detail-why-item-${i}`}
                style={{
                  display: "flex",
                  gap: 10,
                  fontSize: 13,
                  color: "var(--ei-color-fg-primary)",
                  lineHeight: 1.5,
                }}
              >
                <span
                  aria-hidden
                  style={{ color: "var(--ei-color-ok)", fontWeight: 600 }}
                >
                  ✓
                </span>
                <span>{reason}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Where it stretches */}
        <div data-testid="jdmatch-detail-risk" style={sectionStyle}>
          <div
            className="ei-label"
            style={{ color: "var(--ei-color-warn)", marginBottom: 10 }}
          >
            ⚠ {t("jdMatch.recommended.whereStretchHeading")}
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            {recommendation.risks.map((risk, i) => (
              <div
                key={`${risk}-${i}`}
                data-testid={`jdmatch-detail-risk-item-${i}`}
                style={{
                  display: "flex",
                  gap: 10,
                  fontSize: 13,
                  color: "var(--ei-color-fg-secondary)",
                  lineHeight: 1.5,
                }}
              >
                <span aria-hidden style={{ color: "var(--ei-color-warn)" }}>
                  ⓘ
                </span>
                <span>{risk}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Role snapshot */}
        <div data-testid="jdmatch-detail-snapshot" style={sectionStyle}>
          <div
            className="ei-label"
            style={{
              color: "var(--ei-color-fg-tertiary)",
              marginBottom: 10,
            }}
          >
            {t("jdMatch.recommended.roleSnapshotHeading")}
          </div>
          <ul
            style={{
              margin: 0,
              paddingLeft: 18,
              fontSize: 13,
              color: "var(--ei-color-fg-secondary)",
              lineHeight: 1.7,
            }}
          >
            {recommendation.highlights.map((highlight, i) => (
              <li key={`${highlight}-${i}`}>{highlight}</li>
            ))}
          </ul>
        </div>

        {/* Intel */}
        {showIntel ? (
          <div
            data-testid="jdmatch-detail-intel"
            style={{
              ...sectionStyle,
              background: "var(--ei-color-bg-soft)",
            }}
          >
            <div
              className="ei-label"
              style={{
                color: "var(--ei-color-fg-tertiary)",
                marginBottom: 10,
              }}
            >
              {t("jdMatch.recommended.intelHeading")}
            </div>
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: 8,
                fontSize: 12.5,
                color: "var(--ei-color-fg-secondary)",
                lineHeight: 1.5,
              }}
            >
              {recommendation.networkNote ? (
                <div style={{ display: "flex", gap: 8 }}>
                  <span aria-hidden style={{ color: "var(--ei-color-accent)" }}>
                    ◎
                  </span>
                  <span>{recommendation.networkNote}</span>
                </div>
              ) : null}
              {recommendation.similarInterviewers != null &&
              recommendation.similarInterviewers > 0 ? (
                <div style={{ display: "flex", gap: 8 }}>
                  <span aria-hidden style={{ color: "var(--ei-color-accent)" }}>
                    📖
                  </span>
                  <span>
                    {recommendation.similarInterviewers}
                    {" — "}
                    {t("jdMatch.recommended.intelDisclaimer")}
                  </span>
                </div>
              ) : null}
            </div>
          </div>
        ) : null}

        {/* Action bar */}
        <div
          data-testid="jdmatch-detail-action-bar"
          style={{
            padding: "16px 24px",
            display: "flex",
            gap: 10,
            flexWrap: "wrap",
          }}
        >
          <button
            type="button"
            data-testid="jdmatch-detail-action-confirm"
            onClick={onConfirmInterview}
            style={{
              padding: "10px 18px",
              fontSize: 13,
              fontFamily: "var(--ei-font-sans)",
              background: "var(--ei-color-accent)",
              color: "var(--ei-color-on-accent)",
              border: "none",
              borderRadius: "var(--ei-radius-sm)",
              cursor: "pointer",
            }}
          >
            {t("jdMatch.recommended.actionConfirm")}
          </button>
          <button
            type="button"
            data-testid="jdmatch-detail-action-save"
            onClick={onToggleSave}
            style={{
              padding: "8px 14px",
              fontSize: 12.5,
              fontFamily: "var(--ei-font-sans)",
              background: "transparent",
              color: "var(--ei-color-fg-secondary)",
              border: "1px solid var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              cursor: "pointer",
            }}
          >
            {recommendation.saved
              ? t("jdMatch.recommended.actionSaved")
              : t("jdMatch.recommended.actionSave")}
          </button>
          <div style={{ flex: 1 }} />
          <button
            type="button"
            data-testid="jdmatch-detail-action-source"
            onClick={onOpenSource}
            disabled={!sourceUrl}
            style={{
              padding: "8px 14px",
              fontSize: 12.5,
              fontFamily: "var(--ei-font-sans)",
              background: "transparent",
              color: sourceUrl
                ? "var(--ei-color-fg-tertiary)"
                : "var(--ei-color-fg-muted)",
              border: "1px solid var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              cursor: sourceUrl ? "pointer" : "not-allowed",
              opacity: sourceUrl ? 1 : 0.5,
            }}
          >
            {t("jdMatch.recommended.actionSource")}
          </button>
          <button
            type="button"
            data-testid="jdmatch-detail-action-dismiss"
            onClick={onMarkNotRelevant}
            style={{
              padding: "8px 14px",
              fontSize: 12.5,
              fontFamily: "var(--ei-font-sans)",
              background: "transparent",
              color: "var(--ei-color-fg-tertiary)",
              border: "1px solid var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              cursor: "pointer",
            }}
          >
            {t("jdMatch.recommended.actionDismiss")}
          </button>
        </div>
      </div>
    </div>
  );
};
