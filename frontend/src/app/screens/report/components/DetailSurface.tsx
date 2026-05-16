import { useEffect, useMemo, type FC } from "react";

import type { FeedbackReport } from "../../../../api/generated/types";
import { useI18n } from "../../../i18n/messages";
import type { ReplayCtaHandlers } from "../useReplayCtaHandlers";
import type { SummaryDetailKey } from "./SummaryCards";
import { ReadinessTab } from "./tabs/ReadinessTab";
import { DimensionsTab } from "./tabs/DimensionsTab";
import { QuestionsTab } from "./tabs/QuestionsTab";
import { EvidenceTab } from "./tabs/EvidenceTab";
import { NextTab } from "./tabs/NextTab";

interface DetailSurfaceProps {
  detail: SummaryDetailKey | "evidence";
  onSelect: (next: SummaryDetailKey | "evidence") => void;
  report: FeedbackReport;
  replayHandlers: ReplayCtaHandlers;
  activeQuestionTurn: string | null;
  onActiveQuestionChange: (turnId: string) => void;
}

type DetailKey = SummaryDetailKey | "evidence";

const TAB_ORDER: DetailKey[] = [
  "readiness",
  "dimensions",
  "questions",
  "evidence",
  "next",
];

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportDetailSurface
 * (lines 311-516). Tab bar + 5 panels with ARIA tablist semantics. Default
 * tab is `questions` per spec D-?; user can switch freely.
 */
export const DetailSurface: FC<DetailSurfaceProps> = ({
  detail,
  onSelect,
  report,
  replayHandlers,
  activeQuestionTurn,
  onActiveQuestionChange,
}) => {
  const { t } = useI18n();

  const defaultTurn = useMemo(
    () => report.questionAssessments?.[0]?.turnId ?? null,
    [report.questionAssessments],
  );
  useEffect(() => {
    if (!activeQuestionTurn && defaultTurn) {
      onActiveQuestionChange(defaultTurn);
    }
  }, [activeQuestionTurn, defaultTurn, onActiveQuestionChange]);

  const renderPanel = (key: DetailKey) => {
    switch (key) {
      case "readiness":
        return <ReadinessTab report={report} />;
      case "dimensions":
        return <DimensionsTab report={report} />;
      case "questions":
        return (
          <QuestionsTab
            report={report}
            activeTurnId={activeQuestionTurn}
            onActiveQuestionChange={onActiveQuestionChange}
            onAddToReplay={replayHandlers.goReplay}
          />
        );
      case "evidence":
        return <EvidenceTab report={report} />;
      case "next":
        return (
          <NextTab
            report={report}
            onReplay={replayHandlers.goReplay}
            onNextRound={replayHandlers.goNextRound}
          />
        );
    }
  };

  return (
    <section
      data-testid="report-detail-surface"
      style={{
        border: "1px solid var(--ei-color-rule-soft)",
        borderRadius: 3,
        background: "var(--ei-color-bg-card, var(--ei-color-bg-canvas))",
        marginBottom: 24,
        minWidth: 0,
      }}
    >
      <div
        role="tablist"
        aria-label={t("report.detail.tablistLabel")}
        style={{
          display: "flex",
          gap: 0,
          borderBottom: "1px solid var(--ei-color-rule-soft)",
          overflowX: "auto",
        }}
      >
        {TAB_ORDER.map((key) => {
          const active = detail === key;
          return (
            <button
              key={key}
              type="button"
              role="tab"
              id={`report-detail-tab-${key}`}
              data-testid={`report-detail-tab-${key}`}
              aria-selected={active}
              aria-controls={`report-detail-panel-${key}`}
              tabIndex={active ? 0 : -1}
              onClick={() => onSelect(key)}
              style={{
                padding: "14px 18px",
                background: active ? "var(--ei-color-bg-soft, var(--ei-color-bg-canvas))" : "transparent",
                border: "none",
                borderBottom: `2px solid ${active ? "var(--ei-color-accent)" : "transparent"}`,
                color: active ? "var(--ei-color-fg-primary)" : "var(--ei-color-fg-tertiary)",
                cursor: "pointer",
                fontFamily: "var(--ei-font-sans)",
                whiteSpace: "nowrap",
                marginBottom: -1,
              }}
            >
              {t((`report.detail.tab.${key}` as const) as never)}
            </button>
          );
        })}
      </div>
      {TAB_ORDER.map((key) => {
        const active = detail === key;
        return (
          <div
            key={key}
            role="tabpanel"
            id={`report-detail-panel-${key}`}
            data-testid={`report-detail-panel-${key}`}
            aria-labelledby={`report-detail-tab-${key}`}
            hidden={!active}
          >
            {active ? renderPanel(key) : null}
          </div>
        );
      })}
    </section>
  );
};
