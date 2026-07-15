import type { FC, MouseEvent } from "react";

import type { ReportContextSnapshot } from "../../../../api/generated/types";
import { useI18n, type MessageKey } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import type { LooseRoute } from "../../../normalizeRoute";
import { formatRouteUrl } from "../../../routeUrl";

export const ReportContextStrip: FC<{
  report: { context: ReportContextSnapshot };
  conversationReportId?: string;
  marginBottom?: number;
}> = ({ report, conversationReportId, marginBottom = 22 }) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const context = report.context;
  const resumeRoute: LooseRoute = { name: "resume_versions", params: { resumeId: context.resumeId } };
  const fields: Array<{ id: string; label: MessageKey; value: string; href?: string; route?: LooseRoute }> = [
    { id: "job", label: "report.context.job", value: `${context.targetJobCompany} · ${context.targetJobTitle}` },
    { id: "round", label: "report.context.round", value: context.roundName },
    { id: "resume", label: "report.context.resume", value: context.resumeDisplayName, href: formatRouteUrl(resumeRoute), route: resumeRoute },
  ];
  if (conversationReportId) {
    const route: LooseRoute = { name: "report_conversation", params: { reportId: conversationReportId } };
    fields.push({ id: "conversation", label: "report.context.conversation", value: t("report.conversation.entry"), route });
  }
  return (
    <section className="ei-report-context-grid" data-columns={fields.length} data-testid="report-context-strip" style={{ marginBottom }}>
      {fields.map((field) => (
        <div key={field.id} data-testid={`report-context-${field.id}`} style={{ padding: "12px 14px", minWidth: 0, background: "var(--ei-color-bg-card)" }}>
          <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 5 }}>{t(field.label)}</div>
          {field.href && field.route ? (
            <a
              data-testid={`report-context-${field.id}-link`}
              href={field.href}
              onClick={(event) => navigateFromAnchor(event, field.route!, navigate)}
              title={field.value}
              style={{ color: "var(--ei-color-fg-secondary)", fontSize: 12.5, lineHeight: 1.5, overflowWrap: "anywhere", textUnderlineOffset: 3 }}
            >
              {field.value}
            </a>
          ) : field.route ? (
            <button
              data-testid={`report-context-${field.id}-action`}
              type="button"
              onClick={() => navigate(field.route!)}
              title={field.value}
              style={{ padding: 0, border: 0, background: "transparent", color: "var(--ei-color-fg-secondary)", fontSize: 12.5, lineHeight: 1.5, fontFamily: "var(--ei-font-sans)", textDecoration: "underline", textUnderlineOffset: 3, cursor: "pointer" }}
            >
              {field.value}
            </button>
          ) : (
            <div title={field.value} style={{ color: "var(--ei-color-fg-secondary)", fontSize: 12.5, lineHeight: 1.5, overflowWrap: "anywhere" }}>{field.value}</div>
          )}
        </div>
      ))}
    </section>
  );
};

function navigateFromAnchor(event: MouseEvent<HTMLAnchorElement>, route: LooseRoute, navigate: (route: LooseRoute) => void) {
  if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
  event.preventDefault();
  navigate(route);
}
