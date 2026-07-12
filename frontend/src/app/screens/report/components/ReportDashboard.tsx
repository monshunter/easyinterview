import type { CSSProperties, FC, ReactNode } from "react";
import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import type { Route } from "../../../routes";
import { useFeedbackReport } from "../hooks/useFeedbackReport";
import { useReportContextData } from "../hooks/useReportContextData";
import { useReplayCtaHandlers } from "../useReplayCtaHandlers";
import { ReportContextStrip } from "./ReportContextStrip";
import { ReportFailureState } from "./ReportFailureState";
import { ReportHeader } from "./ReportHeader";

interface ReportDashboardProps { route: Route; }

export const ReportDashboard: FC<ReportDashboardProps> = ({ route }) => {
  const { t, lang } = useI18n();
  const { navigate } = useNavigation();
  const report = useFeedbackReport(route.params.reportId ?? "");
  const context = useReportContextData({ targetJobId: route.params.targetJobId, resumeId: route.params.resumeId });
  const replay = useReplayCtaHandlers({ route, report: report.data, sessionId: route.params.sessionId ?? "" });
  const goWorkspace = () => navigate({ name: "workspace", params: route.params.targetJobId ? { targetJobId: route.params.targetJobId } : {} });

  if (report.state === "notFound") return <ReportFailureState errorCode="REPORT_NOT_FOUND" notFound onRetry={report.refresh} onBackToWorkspace={goWorkspace} />;
  if (report.state === "error") return <ReportFailureState errorCode={report.errorCode ?? "AI_OUTPUT_INVALID"} onRetry={report.refresh} onBackToWorkspace={goWorkspace} />;
  if (report.state === "loading" || report.state === "idle") return <div data-testid="report-dashboard-loading" style={{ padding: 48 }}>{t("report.loading")}</div>;
  if (!report.data || report.data.status === "failed") return <ReportFailureState errorCode={report.data?.errorCode ?? "AI_OUTPUT_INVALID"} onRetry={report.refresh} onBackToWorkspace={goWorkspace} />;

  const data = report.data;
  const ready = data.status === "ready";
  const dimensions = data.dimensionAssessments ?? [];
  const evidenceCount = (data.highlights?.length ?? 0) + (data.issues?.length ?? 0);
  return (
    <main data-testid="report-dashboard" className="ei-fadein" style={{ maxWidth: 1120, width: "100%", boxSizing: "border-box", margin: "0 auto", padding: "32px clamp(16px, 5vw, 48px) 96px" }}>
      <button type="button" data-testid="report-back-button" onClick={goWorkspace} style={{ border: 0, background: "transparent", color: "var(--ei-color-fg-tertiary)", cursor: "pointer", marginBottom: 20 }}>← {t("report.back")}</button>
      <ReportHeader breadcrumb={lang === "en" ? "Mock interview / Conversation report" : "模拟面试 / 会话报告"} title={t("report.header.title")} subtitle={t("report.header.subtitle")} onReplay={replay.goReplay} onNextRound={replay.goNextRound} disableReplay={!ready} disableNextRound={!ready} />
      <ReportContextStrip sessionId={route.params.sessionId ?? ""} targetLabel={context.targetLabel} roundLabel={route.params.roundName ?? route.params.roundId ?? null} resumeLabel={context.resumeLabel} />
      <section data-testid="report-summary-cards" style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: 14, marginBottom: 20 }}>
        <Metric label={t("report.summary.readiness")} value={data.preparednessLevel ? t(readinessKey(data.preparednessLevel)) : "—"} />
        <Metric label={t("report.summary.dimensions")} value={String(dimensions.length)} />
        <Metric label={t("report.summary.evidence")} value={String(evidenceCount)} />
      </section>
      <section style={{ display: "grid", gridTemplateColumns: "minmax(0, 1.15fr) minmax(280px, .85fr)", gap: 18, alignItems: "start" }}>
        <Panel title={t("report.detail.tab.dimensions")} testId="report-dimensions">
          {dimensions.length ? dimensions.map((item) => <div key={item.dimension} style={{ display: "flex", justifyContent: "space-between", gap: 20, padding: "13px 0", borderBottom: "1px dotted var(--ei-color-rule-soft)" }}><span>{item.dimension}</span><span style={{ color: statusColor(item.status) }}>{item.status} · {item.confidence}</span></div>) : <Empty />}
        </Panel>
        <div style={{ display: "grid", gap: 18 }}>
          <EvidencePanel title={t("report.detail.tab.evidence")} testId="report-highlights" items={data.highlights ?? []} />
          <EvidencePanel title={t("report.body.issues.eyebrow")} testId="report-issues" items={data.issues ?? []} />
          <Panel title={t("report.detail.tab.next")} testId="report-next-actions">{(data.nextActions ?? []).length ? data.nextActions!.map((item) => <p key={`${item.type}-${item.label}`} style={itemStyle}>{item.label}</p>) : <Empty />}</Panel>
        </div>
      </section>
    </main>
  );
};

const Metric: FC<{ label: string; value: string }> = ({ label, value }) => <div style={{ border: "1px solid var(--ei-color-rule-soft)", padding: 20, background: "var(--ei-color-bg-card)" }}><div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}>{label}</div><div className="ei-serif" style={{ fontSize: 25 }}>{value}</div></div>;
const Panel: FC<{ title: string; testId: string; children: ReactNode }> = ({ title, testId, children }) => <section data-testid={testId} style={{ border: "1px solid var(--ei-color-rule-soft)", padding: 20, background: "var(--ei-color-bg-card)" }}><div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 12 }}>{title}</div>{children}</section>;
const EvidencePanel: FC<{ title: string; testId: string; items: Array<{ dimension: string; evidence: string; confidence: string }> }> = ({ title, testId, items }) => <Panel title={title} testId={testId}>{items.length ? items.map((item) => <p key={`${item.dimension}-${item.evidence}`} style={itemStyle}><strong>{item.dimension}</strong> · {item.evidence} <small>({item.confidence})</small></p>) : <Empty />}</Panel>;
const Empty = () => <div style={{ color: "var(--ei-color-fg-tertiary)" }}>—</div>;
const itemStyle: CSSProperties = { margin: "10px 0 0", lineHeight: 1.65, color: "var(--ei-color-fg-secondary)" };
const statusColor = (status: string) => status === "needs_work" ? "var(--ei-color-warning, #a86418)" : "var(--ei-color-success, #287850)";
const readinessKey = (value: string): "report.readiness.tier.wellPrepared" | "report.readiness.tier.basicallyReady" | "report.readiness.tier.needsPractice" | "report.readiness.tier.notReady" => value === "well_prepared" ? "report.readiness.tier.wellPrepared" : value === "basically_ready" ? "report.readiness.tier.basicallyReady" : value === "not_ready" ? "report.readiness.tier.notReady" : "report.readiness.tier.needsPractice";
