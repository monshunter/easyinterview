import type { FC } from "react";

import type { ReportConversationMessage } from "../../../api/generated/types";
import { useI18n } from "../../i18n/messages";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { PracticeMessageBody } from "../practice/components/PracticeMessageBody";
import { ReportContextStrip } from "../report/components/ReportContextStrip";
import { ReportPageIllustration } from "../reports/ReportPageIllustration";
import { isValidReportConversation } from "./conversationContract";
import { useFailedConversationBackRoute } from "./hooks/useFailedConversationBackRoute";
import { useReportConversation } from "./hooks/useReportConversation";

interface ReportConversationScreenProps {
  route: Route;
}

/** Report-owned readonly interview record; route authority is reportId only. */
export const ReportConversationScreen: FC<ReportConversationScreenProps> = ({ route }) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const reportId = route.name === "report_conversation" ? route.params.reportId ?? "" : "";
  const conversation = useReportConversation(reportId);
  const validConversation = isValidReportConversation(conversation.data, reportId)
    ? conversation.data
    : null;
  const failedBackRoute = useFailedConversationBackRoute(
    reportId,
    validConversation?.reportStatus === "failed",
  );

  if (conversation.state === "loading") {
    return <ReportConversationLoading onBack={() => navigate({ name: "workspace", params: {} })} />;
  }
  if (!validConversation) {
    return <ReportConversationUnavailable onBack={() => navigate({ name: "workspace", params: {} })} />;
  }

  const backRoute: Route | null = validConversation.reportStatus === "ready"
    ? { name: "report" as const, params: { reportId } }
    : validConversation.reportStatus === "failed"
      ? failedBackRoute.status === "resolved"
        ? failedBackRoute.route
        : null
      : { name: "generating" as const, params: { reportId } };

  return (
    <main
      data-testid="report-conversation-screen"
      className="ei-fadein ei-report-conversation-screen"
    >
      {backRoute ? (
        <button
          data-testid="report-conversation-back-button"
          type="button"
          onClick={() => navigate(backRoute)}
          className="ei-reports-back"
        >
          ← {t("common.back")}
        </button>
      ) : null}

      <header className="ei-report-records-header">
        <div className="ei-report-records-header-copy">
          <div className="ei-report-records-eyebrow">
            {t("report.conversation.eyebrow")}
          </div>
          <h1 className="ei-report-records-title">
            {validConversation.context.targetJobCompany} · {validConversation.context.targetJobTitle}
          </h1>
          <p className="ei-report-records-subtitle">
            {t("report.conversation.subtitle")}
          </p>
        </div>
        <ReportPageIllustration testId="report-conversation-header-illustration" />
      </header>

      <div
        className="ei-report-conversation-context"
        data-testid="report-conversation-context-strip"
      >
        <ReportContextStrip report={{ context: validConversation.context }} />
      </div>

      <section
        data-testid="report-conversation-transcript"
        aria-label={t("report.conversation.transcriptLabel")}
        className="ei-report-conversation-transcript"
      >
        {validConversation.messages.length === 0 ? (
          <div
            data-testid="report-conversation-empty"
            role="status"
            className="ei-report-conversation-empty"
          >
            <div className="ei-report-conversation-empty-title">
              {t("report.conversation.empty.eyebrow")}
            </div>
            <div>
              {t("report.conversation.empty.body")}
            </div>
          </div>
        ) : (
          <ol className="ei-report-conversation-list">
            {validConversation.messages.map((message) => (
              <ConversationMessage key={message.sequence} message={message} />
            ))}
          </ol>
        )}
      </section>
    </main>
  );
};

const ConversationMessage: FC<{ message: ReportConversationMessage }> = ({ message }) => {
  const { t } = useI18n();
  const assistant = message.role === "assistant";
  return (
    <li
      data-testid={`report-conversation-message-${message.sequence}`}
      data-role={message.role}
      className={`ei-report-conversation-message ei-report-conversation-message-${message.role}`}
    >
      <div
        className="ei-report-conversation-badge"
        data-testid={`report-conversation-badge-${message.sequence}`}
      >
        {assistant ? "AI" : t("report.conversation.userBadge")}
      </div>
      <div className="ei-report-conversation-message-copy">
        <div className="ei-report-conversation-role">
          {assistant ? t("report.conversation.assistant") : t("report.conversation.user")}
        </div>
        <PracticeMessageBody text={message.content} />
      </div>
    </li>
  );
};

const ReportConversationLoading: FC<{ onBack: () => void }> = ({ onBack }) => {
  const { t } = useI18n();
  return (
    <main data-testid="report-conversation-loading" className="ei-fadein ei-report-conversation-state">
      <section className="ei-reports-state-card">
        <div role="status" aria-live="polite">
          <div className="ei-reports-state-eyebrow">{t("report.conversation.loading.eyebrow")}</div>
          <div className="ei-reports-state-title">{t("report.conversation.loading.title")}</div>
          <p className="ei-reports-state-description">{t("report.conversation.loading.description")}</p>
          <button
            data-testid="report-conversation-loading-back"
            type="button"
            onClick={onBack}
            className="ei-reports-action ei-reports-action-secondary"
          >
            {t("common.back")}
          </button>
        </div>
      </section>
    </main>
  );
};

const ReportConversationUnavailable: FC<{ onBack: () => void }> = ({ onBack }) => {
  const { t } = useI18n();
  return (
    <main data-testid="report-conversation-unavailable" className="ei-fadein ei-report-conversation-state">
      <section className="ei-reports-state-card">
        <div role="alert">
          <div className="ei-reports-state-eyebrow ei-reports-state-eyebrow-error">{t("report.conversation.unavailable.eyebrow")}</div>
          <div className="ei-reports-state-title">{t("report.conversation.unavailable.title")}</div>
          <p className="ei-reports-state-description">{t("report.conversation.unavailable.description")}</p>
          <button
            data-testid="report-conversation-unavailable-back"
            type="button"
            onClick={onBack}
            className="ei-reports-action ei-reports-action-secondary"
          >
            {t("common.back")}
          </button>
        </div>
      </section>
    </main>
  );
};
