import type { FC, ReactNode } from "react";
import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { useFeedbackReport } from "../hooks/useFeedbackReport";
import { confidenceLabel, dimensionStatusLabel, readinessTierLabel } from "../readiness";
import { resolveReportBackRoute } from "../reportBackRoute";
import { isValidFeedbackReport, isValidReadyReport } from "../reportContract";
import { useReplayCtaHandlers } from "../useReplayCtaHandlers";
import { ReportContextStrip } from "./ReportContextStrip";
import { PracticeLaunchTransition } from "../../../interview-context/PracticeLaunchTransition";
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
    return <div data-testid="report-dashboard-loading" className="ei-report-loading-state">{t("report.loading")}</div>;
  }
  if (!validReport) {
    return <ReportFailureState errorCode="AI_OUTPUT_INVALID" contractInvalid onBack={goBack} backDestination={backDestination} />;
  }
  if (validReport.status === "queued" || validReport.status === "generating") {
    return (
      <div data-testid="report-pending-state" className="ei-report-pending-state">
        <button type="button" data-testid="report-pending-back-button" onClick={goBack} className="ei-report-pending-back">← {t("common.back")}</button>
        <p>{t("report.pending")}</p>
        <button type="button" onClick={() => navigate({ name: "generating", params: { reportId } })} className="ei-report-pending-cta">{t("report.pending.cta")}</button>
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
  const labelsByCode = new Map(dimensions.map((item) => [item.code, item.label]));
  const nextDisabled = !data.context.hasNextRound || replay.starting;
  const nextDisabledReason = replay.starting
    ? t("report.header.cta.starting")
    : !data.context.hasNextRound
      ? t("report.header.cta.noNextRound")
      : undefined;

  return (
    <>
      {replay.starting ? <PracticeLaunchTransition /> : null}
      <main data-testid="report-dashboard" className="ei-report-screen ei-fadein">
      <button type="button" data-testid="report-back-button" onClick={goBack} className="ei-report-back">← {t("common.back")}</button>
      <ReportHeader
        breadcrumb={lang === "en" ? "CONVERSATION REPORT" : "会话报告"}
        title={`${data.context.targetJobCompany} · ${data.context.targetJobTitle}`}
        subtitle={t("report.header.subtitle")}
        onReplay={replay.goReplay}
        onNextRound={replay.goNextRound}
        disableReplay={replay.starting}
        disableNextRound={nextDisabled}
        replayVariant="accent"
        nextVariant="secondary"
        nextDisabledReason={nextDisabledReason}
      />
      {replay.startError ? (
        <p data-testid="report-practice-start-error" role="alert" className="ei-report-start-error">
          {t("report.header.cta.startError")}
        </p>
      ) : null}
      <ReportContextStrip report={data} conversationReportId={reportId} />
      <section className="ei-report-summary-grid" data-testid="report-summary-cards">
        <Metric label={t("report.summary.dimensions")} value={String(dimensions.length)} icon="dimensions" />
        <Metric label={t("report.summary.evidence")} value={String(evidenceCount)} icon="evidence" />
      </section>
      <section className="ei-report-detail-grid" data-testid="report-detail-grid">
        <Panel title={t("report.detail.dimensions")} icon="dimensions" testId="report-dimensions">
          <div className="ei-report-dimension-list">
          {dimensions.map((item, index) => (
            <div className="ei-report-dimension-row" key={item.code} data-status={item.status} data-last={index === dimensions.length - 1}>
              <span className="ei-report-dimension-label">{item.label}</span>
              <span className="ei-report-dimension-status">{t(dimensionStatusLabel(item.status))} · {t(confidenceLabel(item.confidence))}</span>
            </div>
          ))}
          </div>
        </Panel>
        <EvidencePanel title={t("report.detail.highlights")} icon="highlights" testId="report-highlights" items={data.highlights} labelsByCode={labelsByCode} />
        <EvidencePanel title={t("report.detail.issues")} icon="issues" testId="report-issues" items={data.issues} labelsByCode={labelsByCode} />
        <Panel title={t("report.detail.actions")} icon="actions" testId="report-actions">
          {data.nextActions.map((item, index) => <div className="ei-report-action-row" key={`${item.type}-${index}`} data-first={index === 0}><span className="ei-report-action-index">{String(index + 1).padStart(2, "0")}</span><span className="ei-report-action-label">{item.label}</span></div>)}
        </Panel>
        <section
          aria-labelledby="report-overall-summary-title"
          data-testid="report-overall-summary"
          className="ei-report-overall"
        >
          <span data-testid="report-overall-icon" className="ei-report-overall-icon"><DetailGlyph kind="overall" /></span>
          <div className="ei-report-overall-content">
            <h2 id="report-overall-summary-title" className="ei-report-overall-label">{t("report.summary.overall")}</h2>
            <div className="ei-report-overall-tier">{t(readinessTierLabel(data.preparednessLevel))}</div>
            <p className="ei-report-overall-copy">{data.summary}</p>
          </div>
        </section>
      </section>
      </main>
    </>
  );
};

type DetailIconKind = "dimensions" | "highlights" | "issues" | "actions" | "overall";

const Metric: FC<{ label: string; value: string; icon: "dimensions" | "evidence" }> = ({ label, value, icon }) => <div className="ei-report-metric" data-kind={icon}><span className="ei-report-metric-icon"><ReportGlyph kind={icon} /></span><div><div className="ei-report-metric-label">{label}</div><div className="ei-report-metric-value">{value}</div></div></div>;
const Panel: FC<{ title: string; icon: Exclude<DetailIconKind, "overall">; testId: string; children: ReactNode }> = ({ title, icon, testId, children }) => <div className="ei-report-panel" data-testid={testId} data-icon={icon}><div className="ei-report-panel-card"><span data-testid="report-detail-card-icon" className="ei-report-detail-card-icon"><DetailGlyph kind={icon} /></span><div className="ei-report-panel-content"><div className="ei-report-panel-title">{title}</div>{children}</div></div></div>;
const EvidencePanel: FC<{ title: string; icon: "highlights" | "issues"; testId: string; items: Array<{ dimensionCode: string; evidence: string }>; labelsByCode: Map<string, string> }> = ({ title, icon, testId, items, labelsByCode }) => <Panel title={title} icon={icon} testId={testId}>{items.map((item, index) => <div className="ei-report-evidence" key={`${item.dimensionCode}-${index}`} data-first={index === 0}><div className="ei-report-evidence-label">{labelsByCode.get(item.dimensionCode)}</div><div className="ei-report-evidence-copy">{item.evidence}</div></div>)}</Panel>;

const ReportGlyph: FC<{ kind: "dimensions" | "evidence" }> = ({ kind }) => kind === "dimensions" ? (
  <svg aria-hidden="true" viewBox="0 0 24 24" width="23" height="23" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round"><path d="M6 19v-6M12 19V5M18 19v-9" /></svg>
) : (
  <svg aria-hidden="true" viewBox="0 0 24 24" width="23" height="23" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M8 5h8M9 3h6v4H9z" /><path d="M7 5H5v16h14V5h-2M8 11h8M8 15h8" /></svg>
);

const DetailGlyph: FC<{ kind: DetailIconKind }> = ({ kind }) => {
  if (kind === "dimensions") return <svg aria-hidden="true" viewBox="0 0 24 24" width="23" height="23" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round"><path d="m12 3 8 4v5c0 5-3.4 8-8 9-4.6-1-8-4-8-9V7z" /><path d="m8.5 12 2.2 2.2 4.8-5" /></svg>;
  if (kind === "highlights") return <svg aria-hidden="true" viewBox="0 0 24 24" width="23" height="23" fill="currentColor"><path d="m12 2.8 2.76 5.6 6.18.9-4.47 4.35 1.05 6.15L12 16.9l-5.52 2.9 1.05-6.15L3.06 9.3l6.18-.9z" /></svg>;
  if (kind === "issues") return <svg aria-hidden="true" viewBox="0 0 24 24" width="23" height="23" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"><circle cx="12" cy="12" r="9" /><path d="M12 7v6M12 17h.01" /></svg>;
  if (kind === "actions") return <svg aria-hidden="true" viewBox="0 0 24 24" width="23" height="23" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="m4 17 5-5 3 3 7-8" /><path d="M14 7h5v5" /></svg>;
  return <svg aria-hidden="true" viewBox="0 0 24 24" width="23" height="23" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M8 4h8v5a4 4 0 0 1-8 0zM8 6H5v2a4 4 0 0 0 4 4M16 6h3v2a4 4 0 0 1-4 4M12 13v4M8 21h8M9 17h6" /></svg>;
};
