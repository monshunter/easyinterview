import type { FC } from "react";

import type { ReportContextSnapshot } from "../../../../api/generated/types";
import { useI18n, type MessageKey } from "../../../i18n/messages";

export const ReportContextStrip: FC<{
  report: { context: ReportContextSnapshot };
  marginBottom?: number;
}> = ({ report, marginBottom = 22 }) => {
  const { t } = useI18n();
  const context = report.context;
  const fields: Array<{ id: string; label: MessageKey; value: string }> = [
    { id: "job", label: "report.context.job", value: `${context.targetJobCompany} · ${context.targetJobTitle}` },
    { id: "round", label: "report.context.round", value: context.roundName },
    { id: "resume", label: "report.context.resume", value: context.resumeDisplayName },
  ];
  return (
    <section className="ei-report-context-grid" data-testid="report-context-strip" style={{ marginBottom }}>
      {fields.map((field) => (
        <div key={field.id} data-testid={`report-context-${field.id}`} style={{ padding: "12px 14px", minWidth: 0, background: "var(--ei-color-bg-card)" }}>
          <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 5 }}>{t(field.label)}</div>
          <div title={field.value} style={{ color: "var(--ei-color-fg-secondary)", fontSize: 12.5, lineHeight: 1.5, overflowWrap: "anywhere" }}>{field.value}</div>
        </div>
      ))}
    </section>
  );
};
