import type { FC } from "react";
import { useI18n, type MessageKey } from "../../../i18n/messages";

export interface ReportContextStripProps { sessionId: string; targetLabel: string | null; roundLabel: string | null; resumeLabel: string | null; }
export const ReportContextStrip: FC<ReportContextStripProps> = (props) => {
  const { t } = useI18n();
  const fields: Array<{ id: string; label: MessageKey; value: string }> = [
    { id: "session", label: "report.context.session", value: props.sessionId },
    { id: "job", label: "report.context.job", value: props.targetLabel ?? "—" },
    { id: "round", label: "report.context.round", value: props.roundLabel ?? "—" },
    { id: "resume", label: "report.context.resume", value: props.resumeLabel ?? "—" },
  ];
  return <div data-testid="report-context-strip" style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))", border: "1px solid var(--ei-color-rule-soft)", marginBottom: 18 }}>{fields.map((field) => <div key={field.id} data-testid={`report-context-${field.id}`} style={{ padding: "13px 16px", minWidth: 0 }}><div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 4 }}>{t(field.label)}</div><div style={{ whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{field.value}</div></div>)}</div>;
};
