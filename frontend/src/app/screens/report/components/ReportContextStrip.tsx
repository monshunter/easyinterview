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
        <div key={field.id} data-testid={`report-context-${field.id}`} className="ei-report-context-item">
          <span className="ei-report-context-icon"><ContextIcon kind={field.id} /></span>
          <div className="ei-report-context-copy">
            <div className="ei-report-context-label">{t(field.label)}</div>
            {field.href && field.route ? (
            <a
              data-testid={`report-context-${field.id}-link`}
              href={field.href}
              onClick={(event) => navigateFromAnchor(event, field.route!, navigate)}
              title={field.value}
              className="ei-report-context-value"
            >
              {field.value}
            </a>
          ) : field.route ? (
            <button
              data-testid={`report-context-${field.id}-action`}
              type="button"
              onClick={() => navigate(field.route!)}
              title={field.value}
              className="ei-report-context-value ei-report-context-action"
            >
              {field.value}
            </button>
            ) : (
              <div title={field.value} className="ei-report-context-value">{field.value}</div>
            )}
          </div>
        </div>
      ))}
    </section>
  );
};

const ContextIcon: FC<{ kind: string }> = ({ kind }) => {
  if (kind === "job") return <svg aria-hidden="true" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.7"><path d="M5 7h14v12H5zM9 7V4h6v3M5 11h14" /></svg>;
  if (kind === "round") return <svg aria-hidden="true" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.7"><circle cx="12" cy="12" r="8" /><path d="M12 8v4l3 2" /></svg>;
  if (kind === "resume") return <svg aria-hidden="true" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.7"><path d="M7 3h8l3 3v15H7zM15 3v4h3M10 11h5M10 15h5" /></svg>;
  return <svg aria-hidden="true" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.7"><path d="M5 5h14v11H9l-4 3z" /><path d="M9 9h6M9 12h4" /></svg>;
};

function navigateFromAnchor(event: MouseEvent<HTMLAnchorElement>, route: LooseRoute, navigate: (route: LooseRoute) => void) {
  if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
  event.preventDefault();
  navigate(route);
}
