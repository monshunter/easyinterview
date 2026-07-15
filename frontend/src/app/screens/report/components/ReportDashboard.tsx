import type { FC, ReactNode } from "react";

import type { Confidence } from "../../../../api/generated/types";
import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { useFeedbackReport } from "../hooks/useFeedbackReport";
import { confidenceLabel, dimensionStatusLabel, readinessTierLabel } from "../readiness";
import { resolveReportBackRoute } from "../reportBackRoute";
import { isValidFeedbackReport, isValidReadyReport } from "../reportContract";
import { useReplayCtaHandlers } from "../useReplayCtaHandlers";
import { ReportContextStrip } from "./ReportContextStrip";
import { ReportFailureState } from "./ReportFailureState";
import { ReportHeader } from "./ReportHeader";

interface ReportDashboardProps {
  reportId: string;
}

export const ReportDashboard: FC<ReportDashboardProps> = ({ reportId }) => {
  const { t, lang } = useI18n();
  const { navigate } = useNavigation();
  const report = useFeedbackReport(reportId);
  const validReport = isValidFeedbackReport(report.data, reportId)
    ? report.data
    : null;
  const readyReport = validReport?.status === "ready" && isValidReadyReport(validReport)
    ? validReport
    : null;
  const replay = useReplayCtaHandlers({ report: readyReport });
  const backRoute = resolveReportBackRoute(report.data, reportId);
  const backDestination = backRoute.name === "reports" ? "reports" : "workspace";
  const goBack = () => navigate(backRoute);

  if (report.state === "notFound") {
    return <ReportFailureState errorCode="REPORT_NOT_FOUND" notFound onBack={goBack} backDestination={backDestination} />;
  }
  if (report.state === "error") {
    return <ReportFailureState errorCode={report.errorCode} onRetry={report.refresh} recoverable onBack={goBack} backDestination={backDestination} />;
  }
  if (report.state === "loading" || report.state === "idle") {
    return <div data-testid="report-dashboard-loading" style={{ padding: 48 }}>{t("report.loading")}</div>;
  }
  if (!validReport) {
    return <ReportFailureState errorCode="AI_OUTPUT_INVALID" contractInvalid onBack={goBack} backDestination={backDestination} />;
  }
  if (validReport.status === "queued" || validReport.status === "generating") {
    return (
      <div data-testid="report-pending-state" style={{ maxWidth: 820, margin: "0 auto", padding: "72px clamp(16px, 5vw, 48px)" }}>
        <button type="button" data-testid="report-pending-back-button" onClick={goBack} style={{ border: 0, background: "transparent", color: "var(--ei-color-fg-tertiary)", cursor: "pointer", marginBottom: 20 }}>← {t("report.back")}</button>
        <p>{t("report.pending")}</p>
        <button type="button" onClick={() => navigate({ name: "generating", params: { reportId } })}>{t("report.pending.cta")}</button>
      </div>
    );
  }
  if (validReport.status === "failed") {
    return <ReportFailureState errorCode={validReport.errorCode} onBack={goBack} backDestination={backDestination} />;
  }
  if (!readyReport) {
    return <ReportFailureState errorCode="AI_OUTPUT_INVALID" contractInvalid onBack={goBack} backDestination={backDestination} />;
  }

  const data = readyReport;
  const dimensions = data.dimensionAssessments;
  const evidenceCount = data.highlights.length + data.issues.length;
  const firstAction = data.nextActions[0];
  const labelsByCode = new Map(dimensions.map((item) => [item.code, item.label]));
  const nextDisabled = !data.context.hasNextRound || replay.starting;
  const nextDisabledReason = replay.starting
    ? t("report.header.cta.starting")
    : !data.context.hasNextRound
      ? t("report.header.cta.noNextRound")
      : undefined;

  return (
    <main data-testid="report-dashboard" className="ei-fadein" style={{ maxWidth: 1120, width: "100%", boxSizing: "border-box", margin: "0 auto", padding: "32px clamp(16px, 5vw, 48px) 96px" }}>
      <button type="button" data-testid="report-back-button" onClick={goBack} style={{ border: 0, background: "transparent", color: "var(--ei-color-fg-tertiary)", cursor: "pointer", marginBottom: 20 }}>← {t("report.back")}</button>
      <ReportHeader
        breadcrumb={lang === "en" ? "CONVERSATION REPORT" : "会话报告"}
        title={`${data.context.targetJobCompany} · ${data.context.targetJobTitle}`}
        subtitle={t("report.header.subtitle")}
        onReplay={replay.goReplay}
        onNextRound={replay.goNextRound}
        disableReplay={replay.starting}
        disableNextRound={nextDisabled}
        replayVariant={firstAction?.type === "retry_current_round" ? "accent" : "secondary"}
        nextVariant={firstAction?.type === "next_round" ? "accent" : "secondary"}
        nextDisabledReason={nextDisabledReason}
      />
      <ReportContextStrip report={data} conversationReportId={reportId} />
      <section className="ei-report-summary-grid" data-testid="report-summary-cards">
        <Metric label={t("report.summary.dimensions")} value={String(dimensions.length)} />
        <Metric label={t("report.summary.evidence")} value={String(evidenceCount)} />
      </section>
      <section className="ei-report-detail-grid" data-testid="report-detail-grid">
        <Panel title={t("report.detail.dimensions")} titleMarginBottom={14} testId="report-dimensions">
          {dimensions.map((item, index) => (
            <div className="ei-report-dimension-row" key={item.code} style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", flexWrap: "wrap", gap: "8px 16px", padding: "13px 0", borderBottom: index < dimensions.length - 1 ? "1px dotted var(--ei-color-rule-strong)" : "none" }}>
              <span className="ei-report-dimension-label" style={{ color: "var(--ei-color-fg-primary)", minWidth: 0, flex: "1 1 160px", overflowWrap: "break-word", wordBreak: "normal" }}>{item.label}</span>
              <span className="ei-report-dimension-status" style={{ color: statusColor(item.status), textAlign: "right", flex: "0 1 auto", maxWidth: "100%", overflowWrap: "break-word", wordBreak: "normal" }}>{t(dimensionStatusLabel(item.status))} · {t(confidenceLabel(item.confidence))}</span>
            </div>
          ))}
        </Panel>
        <EvidencePanel title={t("report.detail.highlights")} titleColor="var(--ei-color-ok)" testId="report-highlights" items={data.highlights} labelsByCode={labelsByCode} confidenceText={(value) => t(confidenceLabel(value))} />
        <EvidencePanel title={t("report.detail.issues")} titleColor="var(--ei-color-warn)" testId="report-issues" items={data.issues} labelsByCode={labelsByCode} confidenceText={(value) => t(confidenceLabel(value))} />
        <Panel title={t("report.detail.actions")} titleColor="var(--ei-color-accent)" testId="report-actions">
          {data.nextActions.map((item, index) => <div className="ei-report-action-row" key={`${item.type}-${index}`} style={{ display: "flex", minWidth: 0, gap: 10, color: "var(--ei-color-fg-secondary)", fontSize: 13, lineHeight: 1.65, marginTop: index ? 12 : 0, overflowWrap: "anywhere", wordBreak: "normal" }}><span style={{ color: "var(--ei-color-accent)", fontFamily: "var(--ei-font-mono)", flexShrink: 0 }}>{String(index + 1).padStart(2, "0")}</span><span className="ei-report-action-label" style={{ minWidth: 0, overflowWrap: "anywhere", wordBreak: "normal" }}>{item.label}</span></div>)}
        </Panel>
        <section
          aria-labelledby="report-overall-summary-title"
          data-testid="report-overall-summary"
          style={{ gridColumn: "1 / -1", minWidth: 0, border: "1px solid var(--ei-color-rule-strong)", borderRadius: 3, padding: 24, background: "var(--ei-color-bg-card)" }}
        >
          <h2 id="report-overall-summary-title" className="ei-label" style={{ margin: "0 0 14px", color: "var(--ei-color-accent)" }}>{t("report.summary.overall")}</h2>
          <div className="ei-serif" style={{ fontSize: 24, overflowWrap: "anywhere" }}>{t(readinessTierLabel(data.preparednessLevel))}</div>
          <p style={{ margin: "12px 0 0", color: "var(--ei-color-fg-secondary)", fontSize: 13.5, lineHeight: 1.7, overflowWrap: "anywhere" }}>{data.summary}</p>
        </section>
      </section>
    </main>
  );
};

const Metric: FC<{ label: string; value: string }> = ({ label, value }) => <div style={{ border: "1px solid var(--ei-color-rule-strong)", padding: 20, background: "var(--ei-color-bg-card)", minWidth: 0 }}><div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}>{label}</div><div className="ei-serif" style={{ fontSize: 24, overflowWrap: "anywhere" }}>{value}</div></div>;
const Panel: FC<{ title: string; titleColor?: string; titleMarginBottom?: number; testId: string; children: ReactNode }> = ({ title, titleColor = "var(--ei-color-fg-tertiary)", titleMarginBottom = 12, testId, children }) => <div className="ei-report-panel" data-testid={testId}><div className="ei-report-panel-card" style={{ border: "1px solid var(--ei-color-rule-strong)", borderRadius: 3, padding: 20, background: "var(--ei-color-bg-card)", minWidth: 0, cursor: "default", transition: "border-color .15s, transform .15s" }}><div className="ei-label" style={{ color: titleColor, marginBottom: titleMarginBottom }}>{title}</div>{children}</div></div>;
const EvidencePanel: FC<{ title: string; titleColor: string; testId: string; items: Array<{ dimensionCode: string; evidence: string; confidence: Confidence }>; labelsByCode: Map<string, string>; confidenceText: (value: Confidence) => string }> = ({ title, titleColor, testId, items, labelsByCode, confidenceText }) => <Panel title={title} titleColor={titleColor} testId={testId}>{items.map((item, index) => <div key={`${item.dimensionCode}-${index}`} style={{ color: "var(--ei-color-fg-secondary)", fontSize: 13, lineHeight: 1.65, marginTop: index ? 14 : 0, overflowWrap: "anywhere" }}><div style={{ color: "var(--ei-color-fg-primary)", fontWeight: 500, marginBottom: 3 }}>{labelsByCode.get(item.dimensionCode)}</div><div>{item.evidence}</div><div style={{ color: "var(--ei-color-fg-tertiary)", fontSize: 11.5, marginTop: 4 }}>{confidenceText(item.confidence)}</div></div>)}</Panel>;
const statusColor = (status: string) => status === "needs_work" ? "var(--ei-color-warn)" : "var(--ei-color-ok)";
