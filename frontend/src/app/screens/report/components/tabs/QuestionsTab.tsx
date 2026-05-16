import type { FC } from "react";

import type { FeedbackReport } from "../../../../../api/generated/types";
import { useI18n } from "../../../../i18n/messages";

interface QuestionsTabProps {
  report: FeedbackReport;
  activeTurnId: string | null;
  onActiveQuestionChange: (turnId: string) => void;
  onAddToReplay: () => void;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportDetailSurface
 * questions branch (lines 385-442). Left column lists turns, right column
 * renders the analysis for the active question.
 */
export const QuestionsTab: FC<QuestionsTabProps> = ({
  report,
  activeTurnId,
  onActiveQuestionChange,
  onAddToReplay,
}) => {
  const { t } = useI18n();
  const questions = report.questionAssessments ?? [];
  if (questions.length === 0) {
    return (
      <div
        data-testid="report-questions-empty"
        style={{ padding: 24, color: "var(--ei-ink3)" }}
      >
        {t("report.questions.empty")}
      </div>
    );
  }
  const active = questions.find((q) => q.turnId === activeTurnId) ?? questions[0]!;
  return (
    <div
      data-testid="report-questions-panel"
      style={{
        padding: 24,
        display: "grid",
        gridTemplateColumns: "320px 1fr",
        gap: 22,
      }}
    >
      <div
        data-testid="report-questions-list"
        style={{ display: "flex", flexDirection: "column", gap: 8 }}
      >
        <div
          className="ei-label"
          style={{ color: "var(--ei-danger, var(--ei-ink))", marginBottom: 4 }}
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
                  isActive ? "var(--ei-accent)" : "var(--ei-rule)"
                }`,
                background: isActive
                  ? "var(--ei-accent-soft, var(--ei-bg))"
                  : "var(--ei-bg)",
                textAlign: "left",
                cursor: "pointer",
                fontFamily: "var(--ei-sans)",
              }}
            >
              <div
                className="ei-mono"
                style={{ fontSize: 11, color: "var(--ei-ink3)", marginBottom: 4 }}
              >
                {q.turnId}
              </div>
              <div style={{ fontSize: 13.5, color: "var(--ei-ink)", fontWeight: 500 }}>
                {q.questionIntent}
              </div>
              <div
                style={{
                  marginTop: 6,
                  fontSize: 11,
                  fontFamily: "var(--ei-mono)",
                  color: "var(--ei-ink3)",
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
          }}
        >
          <div>
            <div
              className="ei-label"
              style={{ color: "var(--ei-ink3)", marginBottom: 5 }}
            >
              {t("report.questions.detail.eyebrow")}
            </div>
            <div
              data-testid="report-questions-detail-topic"
              className="ei-serif"
              style={{ fontSize: 26, color: "var(--ei-ink)", lineHeight: 1.25 }}
            >
              {active.questionIntent}
            </div>
          </div>
          <button
            type="button"
            data-testid="report-questions-add-to-replay"
            onClick={onAddToReplay}
            style={{
              padding: "8px 12px",
              border: "1px solid var(--ei-rule)",
              background: "transparent",
              color: "var(--ei-ink2, var(--ei-ink))",
              fontFamily: "var(--ei-sans)",
              fontSize: 12,
              cursor: "pointer",
              borderRadius: 2,
            }}
          >
            {t("report.questions.detail.addToReplay")}
          </button>
        </div>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr 1fr",
            gap: 12,
            marginBottom: 16,
          }}
        >
          <div
            data-testid="report-questions-detail-good"
            style={{
              background: "var(--ei-ok-soft, var(--ei-bg-soft, var(--ei-bg)))",
              padding: 14,
              borderRadius: 2,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-ok)", marginBottom: 6 }}
            >
              {t("report.questions.detail.good")}
            </div>
            <div
              style={{
                fontSize: 13,
                color: "var(--ei-ink2, var(--ei-ink))",
                lineHeight: 1.6,
              }}
            >
              {t("report.questions.detail.good.body")}
            </div>
          </div>
          <div
            data-testid="report-questions-detail-missing"
            style={{
              background: "var(--ei-bg-soft, var(--ei-bg))",
              padding: 14,
              borderRadius: 2,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-danger, var(--ei-ink))", marginBottom: 6 }}
            >
              {t("report.questions.detail.missing")}
            </div>
            <div
              style={{
                fontSize: 13,
                color: "var(--ei-ink2, var(--ei-ink))",
                lineHeight: 1.6,
              }}
            >
              {t("report.questions.detail.missing.body")}
            </div>
          </div>
          <div
            data-testid="report-questions-detail-frame"
            style={{
              background: "var(--ei-bg-soft, var(--ei-bg))",
              padding: 14,
              borderRadius: 2,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-ink3)", marginBottom: 6 }}
            >
              {t("report.questions.detail.frame")}
            </div>
            <div
              style={{
                fontSize: 12.5,
                color: "var(--ei-ink2, var(--ei-ink))",
                lineHeight: 1.6,
                fontFamily: "var(--ei-mono)",
              }}
            >
              {t("report.questions.detail.frame.body")}
            </div>
          </div>
        </div>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            gap: 14,
          }}
        >
          <div
            data-testid="report-questions-detail-evidence"
            style={{
              padding: 16,
              border: "1px solid var(--ei-rule)",
              borderRadius: 2,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-ink3)", marginBottom: 8 }}
            >
              {t("report.questions.detail.evidence")}
            </div>
            <div
              style={{
                fontSize: 12,
                color: "var(--ei-ink3)",
                fontFamily: "var(--ei-mono)",
              }}
            >
              {active.turnId}
            </div>
          </div>
          <div
            data-testid="report-questions-detail-follow-up"
            style={{
              padding: 16,
              border: "1px solid var(--ei-rule)",
              borderRadius: 2,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-accent)", marginBottom: 8 }}
            >
              {t("report.questions.detail.followUp")}
            </div>
            <div
              style={{
                fontSize: 13.5,
                color: "var(--ei-ink2, var(--ei-ink))",
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
