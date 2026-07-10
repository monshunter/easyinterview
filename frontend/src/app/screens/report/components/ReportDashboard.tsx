import React, { useState, type FC } from "react";

import type { FeedbackReport } from "../../../../api/generated/types";
import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import type { Route } from "../../../routes";
import { useFeedbackReport } from "../hooks/useFeedbackReport";
import { useReportContextData } from "../hooks/useReportContextData";
import { ReportContextStrip } from "./ReportContextStrip";
import { ReportFailureState } from "./ReportFailureState";
import { ReportHeader } from "./ReportHeader";
import {
  SummaryCards,
  type SummaryDetailKey,
} from "./SummaryCards";
import { DetailSurface } from "./DetailSurface";
import { useReplayCtaHandlers } from "../useReplayCtaHandlers";

interface ReportDashboardProps {
  route: Route;
}

/**
 * Top-level container behind `<ReportScreen>` for the success-path dashboard.
 * Owns:
 *   - `useFeedbackReport(reportId)` (single-shot read; 404 → notFound branch)
 *   - `useReportContextData({targetJobId, resumeId})` for the strip
 *   - default detail tab = `questions`
 *   - Replay / Next round CTA wire (see useReplayCtaHandlers).
 */
export const ReportDashboard: FC<ReportDashboardProps> = ({ route }) => {
  const { t, lang } = useI18n();
  const { navigate } = useNavigation();
  const reportId = route.params.reportId ?? "";
  const sessionId = route.params.sessionId ?? "";

  const report = useFeedbackReport(reportId);
  const ctxData = useReportContextData({
    targetJobId: route.params.targetJobId,
    resumeId: route.params.resumeId,
  });

  const [detail, setDetail] = useState<SummaryDetailKey | "evidence">("questions");
  const [activeQuestionTurn, setActiveQuestionTurn] = useState<string | null>(
    null,
  );

  const onSelectSummary = (next: SummaryDetailKey | "evidence") => setDetail(next);

  const dataNotReady =
    report.state !== "data" ||
    (report.data?.status && report.data.status !== "ready");

  const replayHandlers = useReplayCtaHandlers({
    route,
    report: report.data,
    sessionId,
  });

  if (report.state === "notFound") {
    return (
      <ReportFailureState
        errorCode="REPORT_NOT_FOUND"
        notFound
        onRetry={() => navigate({ name: "generating", params: route.params })}
        onBackToWorkspace={() =>
          navigate({ name: "workspace", params: stripContextParams(route.params) })
        }
      />
    );
  }

  if (report.state === "error") {
    return (
      <ReportFailureState
        errorCode={report.errorCode ?? "AI_OUTPUT_INVALID"}
        onRetry={() => report.refresh()}
        onBackToWorkspace={() =>
          navigate({ name: "workspace", params: stripContextParams(route.params) })
        }
      />
    );
  }

  if (report.state === "loading" || report.state === "idle") {
    return (
      <div
        data-testid="report-dashboard-loading"
        style={{
          maxWidth: 1200,
          margin: "0 auto",
          padding: "32px 48px",
          color: "var(--ei-color-fg-tertiary)",
        }}
      >
        {t("report.loading")}
      </div>
    );
  }

  if (report.data && report.data.status === "failed") {
    return (
      <ReportFailureState
        errorCode={report.data.errorCode ?? "AI_OUTPUT_INVALID"}
        onRetry={() => navigate({ name: "generating", params: route.params })}
        onBackToWorkspace={() =>
          navigate({ name: "workspace", params: stripContextParams(route.params) })
        }
      />
    );
  }

  const data = report.data!;
  const dimensionsCount = countDimensions(data);
  const questionsRatio = `${(data.questionAssessments ?? []).length} / ${
    (data.questionAssessments ?? []).length || 0
  }`;
  const roundLabel = route.params.roundName ?? route.params.roundId ?? null;
  const targetTitle = extractTargetTitle(ctxData.targetLabel);

  const breadcrumb = lang === "en"
    ? `Mock interview / ${sessionId || "session"} / Report`
    : `模拟面试 / ${sessionId || "会话"} / 面试报告`;
  const titleSubject = [targetTitle, roundLabel].filter(Boolean).join(" · ");
  const title = titleSubject
    ? lang === "en"
      ? `${titleSubject} mock report`
      : `${titleSubject} 模拟报告`
    : t("report.header.title");
  const subtitle = t("report.header.subtitle");

  return (
    <div
      data-testid="report-dashboard"
      className="ei-fadein"
      style={{
        maxWidth: 1200,
        width: "100%",
        boxSizing: "border-box",
        margin: "0 auto",
        padding: "32px clamp(16px, 5vw, 48px) 96px",
        overflowX: "clip",
      }}
    >
      <button
        type="button"
        data-testid="report-back-button"
        onClick={() =>
          navigate({ name: "workspace", params: stripContextParams(route.params) })
        }
        style={{
          background: "transparent",
          border: "none",
          color: "var(--ei-color-fg-tertiary)",
          fontSize: 13,
          marginBottom: 20,
          cursor: "pointer",
        }}
      >
        ← {t("report.back")}
      </button>
      <ReportHeader
        breadcrumb={breadcrumb}
        title={title}
        subtitle={subtitle}
        onReplay={replayHandlers.goReplay}
        onNextRound={replayHandlers.goNextRound}
        disableReplay={dataNotReady}
        disableNextRound={dataNotReady}
      />
      <ReportContextStrip
        sessionId={sessionId}
        targetLabel={ctxData.targetLabel}
        roundLabel={roundLabel}
        resumeLabel={ctxData.resumeLabel}
        modality={route.params.modality ?? "text"}
        practiceMode={route.params.practiceMode ?? "strict"}
        hintUsed={route.params.hintUsed ?? "false"}
        hintCount={route.params.hintCount ?? "0"}
      />
      <SummaryCards
        active={detail === "evidence" ? "next" : detail}
        onSelect={onSelectSummary}
        readinessTier={data.preparednessLevel ?? null}
        dimensionsCount={dimensionsCount}
        questionsRatio={questionsRatio}
      />
      <DetailSurface
        detail={detail}
        onSelect={onSelectSummary}
        report={data}
        activeQuestionTurn={activeQuestionTurn}
        onActiveQuestionChange={setActiveQuestionTurn}
      />
      <ReportDashboardBody
        report={data}
        onOpenDimensions={() => onSelectSummary("dimensions")}
        onOpenNext={() => onSelectSummary("next")}
        onOpenQuestions={(turnId) => {
          setActiveQuestionTurn(turnId);
          onSelectSummary("questions");
        }}
        onOpenEvidence={() => onSelectSummary("evidence")}
      />
    </div>
  );
};

interface ReportDashboardBodyProps {
  report: FeedbackReport;
  onOpenDimensions: () => void;
  onOpenNext: () => void;
  onOpenQuestions: (turnId: string) => void;
  onOpenEvidence: () => void;
}

function ReportDashboardBody({
  report,
  onOpenDimensions,
  onOpenNext,
  onOpenQuestions,
  onOpenEvidence,
}: ReportDashboardBodyProps) {
  const { t } = useI18n();
  const dimensionAggregate = aggregateDimensions(report);
  const topPriority =
    report.issues?.[0]?.evidence ??
    report.highlights?.[0]?.evidence ??
    t("report.body.topPriority.empty");
  const nextPractice = (report.nextActions ?? []).slice(0, 3);
  const perQuestion = report.questionAssessments ?? [];
  const issues = report.issues ?? [];
  const highlights = report.highlights ?? [];
  return (
    <>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(min(260px, 100%), 1fr))",
          gap: 20,
          marginBottom: 24,
        }}
      >
        <section data-testid="report-body-dimensions-card" style={cardStyle()}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: 14,
            }}
          >
            <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)" }}>
              {t("report.body.dimensions.eyebrow")}
            </div>
            <button
              type="button"
              onClick={onOpenDimensions}
              data-testid="report-body-dimensions-open"
              style={{
                background: "transparent",
                border: "none",
                color: "var(--ei-color-accent)",
                fontSize: 12,
                cursor: "pointer",
              }}
            >
              {t("report.body.dimensions.open")}
            </button>
          </div>
          {dimensionAggregate.length === 0 ? (
            <div data-testid="report-body-dimensions-empty" style={{ color: "var(--ei-color-fg-tertiary)" }}>
              {t("report.dimensions.empty")}
            </div>
          ) : (
            dimensionAggregate.map((d, i) => (
              <div
                key={d.name}
                data-testid={`report-dim-row-${i}`}
                data-dim-name={d.name}
                style={{
                  padding: "12px 0",
                  borderBottom:
                    i < dimensionAggregate.length - 1
                      ? "1px dotted var(--ei-color-rule-soft)"
                      : "none",
                  fontSize: 13,
                  color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
                  display: "flex",
                  justifyContent: "space-between",
                  gap: 12,
                  minWidth: 0,
                }}
              >
                <span style={{ minWidth: 0, overflowWrap: "anywhere" }}>{d.name}</span>
                <span
                  style={{
                    fontFamily: "var(--ei-font-mono)",
                    fontSize: 11,
                    color: "var(--ei-color-fg-tertiary)",
                    flex: "0 0 auto",
                  }}
                >
                  {d.status ?? "—"} · {d.score}%
                </span>
              </div>
            ))
          )}
        </section>
        <section data-testid="report-body-priority-card" style={cardStyle()}>
          <div
            className="ei-label"
            data-testid="report-top-priority"
            style={{ color: "var(--ei-color-accent)", marginBottom: 10 }}
          >
            {t("report.body.topPriority.eyebrow")}
          </div>
          <div
            className="ei-serif"
            style={{
              fontSize: 18,
              color: "var(--ei-color-fg-primary)",
              lineHeight: 1.45,
              marginBottom: 16,
            }}
          >
            {topPriority}
          </div>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: 10,
            }}
          >
            <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)" }}>
              {t("report.body.nextPractice.eyebrow")}
            </div>
            <button
              type="button"
              data-testid="report-body-next-practice-open"
              onClick={onOpenNext}
              style={{
                background: "transparent",
                border: "none",
                color: "var(--ei-color-accent)",
                fontSize: 12,
                cursor: "pointer",
              }}
            >
              {t("report.body.nextPractice.open")}
            </button>
          </div>
          {nextPractice.length === 0 ? (
            <div
              data-testid="report-body-next-practice-empty"
              style={{ color: "var(--ei-color-fg-tertiary)" }}
            >
              {t("report.next.empty")}
            </div>
          ) : (
            nextPractice.map((entry, i) => (
              <div
                key={i}
                data-testid={`report-next-practice-${i}`}
                style={{
                  display: "flex",
                  gap: 10,
                  padding: "8px 0",
                  fontSize: 13,
                  color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
                  borderBottom:
                    i < nextPractice.length - 1
                      ? "1px dotted var(--ei-color-rule-soft)"
                      : "none",
                }}
              >
                <span
                  style={{
                    color: "var(--ei-color-accent)",
                    fontFamily: "var(--ei-font-mono)",
                  }}
                >
                  {String(i + 1).padStart(2, "0")}
                </span>
                <span>{entry.label}</span>
              </div>
            ))
          )}
        </section>
      </div>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(min(280px, 100%), 1fr))",
          gap: 20,
          marginBottom: 24,
        }}
      >
        <section data-testid="report-body-questions-card" style={cardStyle()}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: 12,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-danger, var(--ei-color-fg-primary))" }}
            >
              {t("report.body.questions.eyebrow")}
            </div>
            <button
              type="button"
              data-testid="report-body-questions-open"
              onClick={() => onOpenQuestions(perQuestion[0]?.turnId ?? "")}
              style={{
                background: "transparent",
                border: "none",
                color: "var(--ei-color-accent)",
                fontSize: 12,
                cursor: "pointer",
              }}
            >
              {t("report.body.questions.open")}
            </button>
          </div>
          {perQuestion.length === 0 ? (
            <div data-testid="report-body-questions-empty" style={{ color: "var(--ei-color-fg-tertiary)" }}>
              {t("report.questions.empty")}
            </div>
          ) : (
            perQuestion.map((q, i) => (
              <button
                key={q.turnId}
                type="button"
                data-testid={`report-perq-${i}`}
                onClick={() => onOpenQuestions(q.turnId)}
                style={{
                  width: "100%",
                  textAlign: "left",
                  background: "transparent",
                  border: "none",
                  padding: "12px 0",
                  borderBottom:
                    i < perQuestion.length - 1
                      ? "1px dotted var(--ei-color-rule-soft)"
                      : "none",
                  cursor: "pointer",
                  fontFamily: "var(--ei-font-sans)",
                }}
              >
                <div
                  style={{
                  display: "flex",
                  gap: 10,
                  alignItems: "center",
                  minWidth: 0,
                }}
              >
                  <span
                    style={{
                      fontFamily: "var(--ei-font-mono)",
                      fontSize: 11,
                      color: "var(--ei-color-fg-tertiary)",
                      minWidth: 0,
                      overflowWrap: "anywhere",
                    }}
                  >
                    {q.turnId}
                  </span>
                  <span
                    style={{
                      fontSize: 14,
                      color: "var(--ei-color-fg-primary)",
                      fontWeight: 500,
                      minWidth: 0,
                      overflowWrap: "anywhere",
                    }}
                  >
                    {q.questionIntent}
                  </span>
                </div>
              </button>
            ))
          )}
        </section>
        <div
          style={{ display: "flex", flexDirection: "column", gap: 20, minWidth: 0 }}
        >
          <section data-testid="report-body-issues-card" style={cardStyle()}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: 12,
              }}
            >
              <div
                className="ei-label"
                style={{ color: "var(--ei-color-danger, var(--ei-color-fg-primary))" }}
              >
                {t("report.body.issues.eyebrow")}
              </div>
              <button
                type="button"
                data-testid="report-body-issues-open"
                onClick={onOpenEvidence}
                style={{
                  background: "transparent",
                  border: "none",
                  color: "var(--ei-color-accent)",
                  fontSize: 12,
                  cursor: "pointer",
                }}
              >
                {t("report.body.issues.open")}
              </button>
            </div>
            {issues.length === 0 ? (
              <div data-testid="report-body-issues-empty" style={{ color: "var(--ei-color-fg-tertiary)" }}>
                {t("report.evidence.risks.empty")}
              </div>
            ) : (
              issues.map((issue, i) => (
                <div
                  key={i}
                  data-testid={`report-issue-${i}`}
                  style={{
                    padding: "10px 0",
                    borderBottom:
                      i < issues.length - 1
                        ? "1px dotted var(--ei-color-rule-soft)"
                        : "none",
                  }}
                >
                  <div
                    style={{
                      fontSize: 13.5,
                      color: "var(--ei-color-fg-primary)",
                      fontWeight: 500,
                    }}
                  >
                    {issue.dimension}
                  </div>
                  <div
                    style={{
                      fontSize: 12,
                      color: "var(--ei-color-fg-tertiary)",
                      fontFamily: "var(--ei-font-mono)",
                      marginTop: 2,
                    }}
                  >
                    {issue.evidence}
                  </div>
                </div>
              ))
            )}
          </section>
          <section data-testid="report-body-highlights-card" style={cardStyle()}>
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-ok)", marginBottom: 12 }}
            >
              {t("report.body.highlights.eyebrow")}
            </div>
            {highlights.length === 0 ? (
              <div
                data-testid="report-body-highlights-empty"
                style={{ color: "var(--ei-color-fg-tertiary)" }}
              >
                {t("report.evidence.highlights.empty")}
              </div>
            ) : (
              highlights.map((h, i) => (
                <div
                  key={i}
                  data-testid={`report-highlight-${i}`}
                  style={{
                    padding: "10px 0",
                    borderBottom:
                      i < highlights.length - 1
                        ? "1px dotted var(--ei-color-rule-soft)"
                        : "none",
                  }}
                >
                  <div
                    style={{
                      fontSize: 13.5,
                      color: "var(--ei-color-fg-primary)",
                      fontWeight: 500,
                    }}
                  >
                    {h.dimension}
                  </div>
                  <div
                    style={{
                      fontSize: 12,
                      color: "var(--ei-color-fg-tertiary)",
                      marginTop: 2,
                    }}
                  >
                    {h.evidence}
                  </div>
                </div>
              ))
            )}
          </section>
        </div>
      </div>
    </>
  );
}

function cardStyle(): React.CSSProperties {
  return {
    padding: 18,
    border: "1px solid var(--ei-color-rule-soft)",
    borderRadius: 3,
    background: "var(--ei-color-bg-card, var(--ei-color-bg-canvas))",
    minWidth: 0,
    overflowWrap: "anywhere",
  };
}

interface DimensionAggregateRow {
  name: string;
  status: string | null;
  score: number;
}

function aggregateDimensions(report: FeedbackReport): DimensionAggregateRow[] {
  const map = new Map<string, { strong: number; meets: number; needs: number; samples: number }>();
  for (const q of report.questionAssessments ?? []) {
    const dims = q.dimensionResults ?? {};
    for (const [name, raw] of Object.entries(dims)) {
      const cell = (raw ?? {}) as { status?: string };
      const acc =
        map.get(name) ?? { strong: 0, meets: 0, needs: 0, samples: 0 };
      if (cell.status === "strong") acc.strong += 1;
      else if (cell.status === "meets_bar") acc.meets += 1;
      else if (cell.status === "needs_work") acc.needs += 1;
      acc.samples += 1;
      map.set(name, acc);
    }
  }
  const out: DimensionAggregateRow[] = [];
  for (const [name, acc] of map.entries()) {
    const status: string =
      acc.strong > acc.needs && acc.strong > acc.meets
        ? "strong"
        : acc.needs > acc.meets
          ? "needs_work"
          : "meets_bar";
    const score = Math.round(
      ((acc.strong + acc.meets * 0.65) / Math.max(acc.samples, 1)) * 100,
    );
    out.push({ name, status, score });
  }
  return out;
}

function stripContextParams(params: Record<string, string>): Record<string, string> {
  const next: Record<string, string> = {};
  const allowed: ReadonlyArray<string> = [
    "planId",
    "targetJobId",
    "jdId",
    "resumeId",
    "roundId",
  ];
  for (const key of allowed) {
    if (params[key]) next[key] = params[key]!;
  }
  return next;
}

function countDimensions(report: FeedbackReport): number {
  const set = new Set<string>();
  for (const q of report.questionAssessments ?? []) {
    for (const key of Object.keys(q.dimensionResults ?? {})) {
      set.add(key);
    }
  }
  return set.size;
}

function extractTargetTitle(targetLabel: string | null): string | null {
  if (!targetLabel) return null;
  const parts = targetLabel
    .split(" · ")
    .map((part) => part.trim())
    .filter(Boolean);
  return parts.length > 0 ? parts[parts.length - 1]! : targetLabel;
}
