import { useState, type FC } from "react";

import type { FeedbackReport } from "../../../../../api/generated/types";
import { useI18n } from "../../../../i18n/messages";

interface QuestionsTabProps {
  report: FeedbackReport;
  activeTurnId: string | null;
  onActiveQuestionChange: (turnId: string) => void;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportDetailSurface
 * questions branch (lines 385-442). Left column lists turns, right column
 * renders the analysis for the active question.
 *
 * Per product-scope v2.1 D-19 the per-question "加入本轮复练" control is a
 * local marker only (mirrors the prototype `replayQueued` / `toggleQueued`
 * state): it toggles intent inside the report and never navigates or starts
 * a session — the single start CTA lives in the report Header.
 */
export const QuestionsTab: FC<QuestionsTabProps> = ({
  report,
  activeTurnId,
  onActiveQuestionChange,
}) => {
  const { t } = useI18n();
  const [markedForReplay, setMarkedForReplay] = useState<
    Record<string, boolean>
  >({});
  const questions = report.questionAssessments ?? [];
  if (questions.length === 0) {
    return (
      <div
        data-testid="report-questions-empty"
        style={{ padding: 24, color: "var(--ei-color-fg-tertiary)" }}
      >
        {t("report.questions.empty")}
      </div>
    );
  }
  const active = questions.find((q) => q.turnId === activeTurnId) ?? questions[0]!;
  const activeMarked = !!markedForReplay[active.turnId];
  const toggleActiveMarked = () =>
    setMarkedForReplay((prev) => ({
      ...prev,
      [active.turnId]: !prev[active.turnId],
    }));
  return (
    <div
      data-testid="report-questions-panel"
      style={{
        padding: 24,
        display: "grid",
        gridTemplateColumns: "repeat(auto-fit, minmax(min(320px, 100%), 1fr))",
        gap: 22,
        minWidth: 0,
      }}
    >
      <div
        data-testid="report-questions-list"
        style={{ display: "flex", flexDirection: "column", gap: 8, minWidth: 0 }}
      >
        <div
          className="ei-label"
          style={{ color: "var(--ei-color-danger, var(--ei-color-fg-primary))", marginBottom: 4 }}
        >
          {t("report.questions.list.eyebrow")}
        </div>
        {questions.map((q) => {
          const isActive = q.turnId === active.turnId;
          return (
            <button
              key={q.turnId}
              type="button"
              data-testid={`report-questions-list-item-${q.turnId}`}
              data-active={isActive ? "true" : "false"}
              onClick={() => onActiveQuestionChange(q.turnId)}
              style={{
                padding: "12px 14px",
                borderRadius: 2,
                border: `1px solid ${
                  isActive ? "var(--ei-color-accent)" : "var(--ei-color-rule-soft)"
                }`,
                background: isActive
                  ? "var(--ei-color-accent-soft, var(--ei-color-bg-canvas))"
                  : "var(--ei-color-bg-canvas)",
                textAlign: "left",
                cursor: "pointer",
                fontFamily: "var(--ei-font-sans)",
                minWidth: 0,
              }}
            >
              <div
                className="ei-mono"
                style={{
                  fontSize: 11,
                  color: "var(--ei-color-fg-tertiary)",
                  marginBottom: 4,
                  overflowWrap: "anywhere",
                }}
              >
                {q.turnId}
              </div>
              <div
                style={{
                  fontSize: 13.5,
                  color: "var(--ei-color-fg-primary)",
                  fontWeight: 500,
                  overflowWrap: "anywhere",
                }}
              >
                {q.questionIntent}
              </div>
              <div
                style={{
                  marginTop: 6,
                  fontSize: 11,
                  fontFamily: "var(--ei-font-mono)",
                  color: "var(--ei-color-fg-tertiary)",
                }}
              >
                {t("report.questions.list.review")}: {q.reviewStatus}
              </div>
            </button>
          );
        })}
      </div>
      <div style={{ minWidth: 0 }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            gap: 16,
            alignItems: "flex-start",
            marginBottom: 16,
            flexWrap: "wrap",
          }}
        >
          <div style={{ minWidth: 0, flex: "1 1 220px" }}>
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 5 }}
            >
              {t("report.questions.detail.eyebrow")}
            </div>
            <div
              data-testid="report-questions-detail-topic"
              className="ei-serif"
              style={{
                fontSize: 26,
                color: "var(--ei-color-fg-primary)",
                lineHeight: 1.25,
                overflowWrap: "anywhere",
              }}
            >
              {active.questionIntent}
            </div>
          </div>
          <button
            type="button"
            data-testid="report-questions-add-to-replay"
            data-marked={activeMarked ? "true" : "false"}
            aria-pressed={activeMarked}
            onClick={toggleActiveMarked}
            style={{
              padding: "8px 12px",
              border: `1px solid ${activeMarked ? "var(--ei-color-accent)" : "var(--ei-color-rule-soft)"}`,
              background: activeMarked
                ? "var(--ei-color-accent-soft, transparent)"
                : "transparent",
              color: activeMarked
                ? "var(--ei-color-accent)"
                : "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
              fontFamily: "var(--ei-font-sans)",
              fontSize: 12,
              cursor: "pointer",
              borderRadius: 2,
              flex: "1 1 160px",
            }}
          >
            {activeMarked
              ? t("report.questions.detail.addedToReplay")
              : t("report.questions.detail.addToReplay")}
          </button>
        </div>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(min(150px, 100%), 1fr))",
            gap: 12,
            marginBottom: 16,
          }}
        >
          <div
            data-testid="report-questions-detail-good"
            style={{
              background: "var(--ei-color-ok-soft, var(--ei-color-bg-soft, var(--ei-color-bg-canvas)))",
              padding: 14,
              borderRadius: 2,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-ok)", marginBottom: 6 }}
            >
              {t("report.questions.detail.good")}
            </div>
            <div
              style={{
                fontSize: 13,
                color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
                lineHeight: 1.6,
              }}
            >
              {t("report.questions.detail.good.body")}
            </div>
          </div>
          <div
            data-testid="report-questions-detail-missing"
            style={{
              background: "var(--ei-color-bg-soft, var(--ei-color-bg-canvas))",
              padding: 14,
              borderRadius: 2,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-danger, var(--ei-color-fg-primary))", marginBottom: 6 }}
            >
              {t("report.questions.detail.missing")}
            </div>
            <div
              style={{
                fontSize: 13,
                color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
                lineHeight: 1.6,
              }}
            >
              {t("report.questions.detail.missing.body")}
            </div>
          </div>
          <div
            data-testid="report-questions-detail-frame"
            style={{
              background: "var(--ei-color-bg-soft, var(--ei-color-bg-canvas))",
              padding: 14,
              borderRadius: 2,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 6 }}
            >
              {t("report.questions.detail.frame")}
            </div>
            <div
              style={{
                fontSize: 12.5,
                color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
                lineHeight: 1.6,
                fontFamily: "var(--ei-font-mono)",
              }}
            >
              {t("report.questions.detail.frame.body")}
            </div>
          </div>
        </div>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(min(180px, 100%), 1fr))",
            gap: 14,
          }}
        >
          <div
            data-testid="report-questions-detail-evidence"
            style={{
              padding: 16,
              border: "1px solid var(--ei-color-rule-soft)",
              borderRadius: 2,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 8 }}
            >
              {t("report.questions.detail.evidence")}
            </div>
            <div
              style={{
                fontSize: 12,
                color: "var(--ei-color-fg-tertiary)",
                fontFamily: "var(--ei-font-mono)",
              }}
            >
              {active.turnId}
            </div>
          </div>
          <div
            data-testid="report-questions-detail-follow-up"
            style={{
              padding: 16,
              border: "1px solid var(--ei-color-rule-soft)",
              borderRadius: 2,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-accent)", marginBottom: 8 }}
            >
              {t("report.questions.detail.followUp")}
            </div>
            <div
              style={{
                fontSize: 13.5,
                color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
                lineHeight: 1.65,
              }}
            >
              {t("report.questions.detail.followUp.body")}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
