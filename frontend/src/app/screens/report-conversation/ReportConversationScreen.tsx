import type { FC } from "react";

import type { ReportConversationMessage } from "../../../api/generated/types";
import { useI18n } from "../../i18n/messages";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { PracticeMessageBody } from "../practice/components/PracticeMessageBody";
import { ReportContextStrip } from "../report/components/ReportContextStrip";
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
      className="ei-fadein"
      style={{ maxWidth: 880, margin: "0 auto", padding: "32px clamp(16px, 5vw, 48px) 96px" }}
    >
      {backRoute ? (
        <button
          data-testid="report-conversation-back-button"
          type="button"
          onClick={() => navigate(backRoute)}
          style={{ border: 0, background: "transparent", color: "var(--ei-color-fg-tertiary)", cursor: "pointer", marginBottom: 20, padding: 0 }}
        >
          ← {t("report.conversation.back")}
        </button>
      ) : null}

      <header style={{ marginBottom: 24 }}>
        <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 8 }}>
          {t("report.conversation.eyebrow")}
        </div>
        <h1 className="ei-serif" style={{ margin: 0, fontSize: 36, color: "var(--ei-color-fg-primary)", overflowWrap: "anywhere" }}>
          {validConversation.context.targetJobCompany} · {validConversation.context.targetJobTitle}
        </h1>
        <p style={{ color: "var(--ei-color-fg-secondary)", lineHeight: 1.7, margin: "10px 0 0" }}>
          {t("report.conversation.subtitle")}
        </p>
      </header>

      <div data-testid="report-conversation-context-strip">
        <ReportContextStrip report={{ context: validConversation.context }} />
      </div>

      <section
        data-testid="report-conversation-transcript"
        aria-label={t("report.conversation.transcriptLabel")}
        style={{ minWidth: 0, overflowX: "hidden" }}
      >
        {validConversation.messages.length === 0 ? (
          <div
            data-testid="report-conversation-empty"
            role="status"
            style={{ padding: "20px 0 28px", color: "var(--ei-color-fg-tertiary)", fontSize: 13, lineHeight: 1.65 }}
          >
            <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 8 }}>
              {t("report.conversation.empty.eyebrow")}
            </div>
            <div style={{ color: "var(--ei-color-fg-secondary)" }}>
              {t("report.conversation.empty.body")}
            </div>
          </div>
        ) : (
          <ol style={{ margin: 0, padding: 0, listStyle: "none" }}>
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
      style={{ display: "flex", gap: 12, minWidth: 0, marginBottom: 20 }}
    >
      <div
        style={{
          width: 28,
          height: 28,
          borderRadius: 2,
          flexShrink: 0,
          background: assistant ? "var(--ei-color-accent-soft)" : "var(--ei-color-bg-soft)",
          color: assistant ? "var(--ei-color-accent)" : "var(--ei-color-fg-secondary)",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          fontSize: 11,
          fontFamily: "var(--ei-font-mono)",
          fontWeight: 500,
        }}
      >
        {assistant ? "AI" : t("report.conversation.userBadge")}
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ color: "var(--ei-color-fg-secondary)", fontSize: 12, fontWeight: 500, marginBottom: 5 }}>
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
    <main data-testid="report-conversation-loading" className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px clamp(16px, 5vw, 48px)" }}>
      <section style={{ background: "var(--ei-color-bg-card)", border: "1px solid var(--ei-color-rule-strong)", borderRadius: 3, padding: 20 }}>
        <div role="status" aria-live="polite">
          <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}>{t("report.conversation.loading.eyebrow")}</div>
          <div className="ei-serif" style={{ color: "var(--ei-color-fg-primary)", fontSize: 26, marginBottom: 10 }}>{t("report.conversation.loading.title")}</div>
          <p style={{ color: "var(--ei-color-fg-secondary)", lineHeight: 1.65, margin: "0 0 18px" }}>{t("report.conversation.loading.description")}</p>
          <button
            data-testid="report-conversation-loading-back"
            type="button"
            onClick={onBack}
            style={{ display: "inline-flex", alignItems: "center", justifyContent: "center", height: 38, padding: "0 16px", fontSize: 14, fontWeight: 500, background: "var(--ei-color-bg-canvas)", color: "var(--ei-color-fg-primary)", border: "1px solid var(--ei-color-rule-strong)", borderRadius: 2, cursor: "pointer", fontFamily: "var(--ei-font-sans)", letterSpacing: "-0.005em", transition: "transform .08s ease, opacity .15s" }}
          >
            {t("report.conversation.loading.back")}
          </button>
        </div>
      </section>
    </main>
  );
};

const ReportConversationUnavailable: FC<{ onBack: () => void }> = ({ onBack }) => {
  const { t } = useI18n();
  return (
    <main data-testid="report-conversation-unavailable" className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px clamp(16px, 5vw, 48px)" }}>
      <section style={{ background: "var(--ei-color-bg-card)", border: "1px solid var(--ei-color-rule-strong)", borderRadius: 3, padding: 20 }}>
        <div role="alert">
          <div className="ei-label" style={{ color: "var(--ei-color-danger)", marginBottom: 10 }}>{t("report.conversation.unavailable.eyebrow")}</div>
          <div className="ei-serif" style={{ color: "var(--ei-color-fg-primary)", fontSize: 26, marginBottom: 10 }}>{t("report.conversation.unavailable.title")}</div>
          <p style={{ color: "var(--ei-color-fg-secondary)", lineHeight: 1.65, margin: "0 0 18px" }}>{t("report.conversation.unavailable.description")}</p>
          <button
            data-testid="report-conversation-unavailable-back"
            type="button"
            onClick={onBack}
            style={{ display: "inline-flex", alignItems: "center", justifyContent: "center", height: 38, padding: "0 16px", fontSize: 14, fontWeight: 500, background: "var(--ei-color-bg-canvas)", color: "var(--ei-color-fg-primary)", border: "1px solid var(--ei-color-rule-strong)", borderRadius: 2, cursor: "pointer", fontFamily: "var(--ei-font-sans)" }}
          >
            {t("report.conversation.unavailable.back")}
          </button>
        </div>
      </section>
    </main>
  );
};
