// Screen 4: grounded direct-semantic conversation report.
const DIMENSION_STATUS_LABELS = {
  strong: { zh: "强项", en: "Strong" },
  meets_bar: { zh: "符合预期", en: "Meets expectations" },
  needs_work: { zh: "待加强", en: "Needs work" },
};
const CONFIDENCE_LABELS = {
  high: { zh: "高置信度", en: "High confidence" },
  medium: { zh: "中置信度", en: "Medium confidence" },
  low: { zh: "低置信度", en: "Low confidence" },
};
const READINESS_LABELS = {
  not_ready: { zh: "尚未准备好", en: "Not ready" },
  needs_practice: { zh: "建议再练", en: "Needs practice" },
  basically_ready: { zh: "基本可面", en: "Basically ready" },
  well_prepared: { zh: "准备充分", en: "Well prepared" },
};
const ACTION_LABEL_WIRE_MAX_CODE_POINTS = 200;

const localizeDimensionStatus = (status, lang) => DIMENSION_STATUS_LABELS[status]?.[lang] || null;
const localizeConfidence = (confidence, lang) => CONFIDENCE_LABELS[confidence]?.[lang] || null;
const localizeReadiness = (readiness, lang) => READINESS_LABELS[readiness]?.[lang] || null;
const isNonEmptyText = (value) => typeof value === "string" && value.trim().length > 0;
const isValidActionLabel = (value, language) => {
  if (!isNonEmptyText(value)) return false;
  const codePoints = [...value].length;
  if (codePoints > ACTION_LABEL_WIRE_MAX_CODE_POINTS) return false;
  if (language === "en") return value.trim().split(/\s+/u).length <= 24;
  if (language === "zh-CN") return codePoints <= 64;
  return false;
};

const isValidDirectReport = (report) => {
  const dimensions = Array.isArray(report?.dimensionAssessments) ? report.dimensionAssessments : [];
  const highlights = Array.isArray(report?.highlights) ? report.highlights : [];
  const issues = Array.isArray(report?.issues) ? report.issues : [];
  const actions = Array.isArray(report?.nextActions) ? report.nextActions : [];
  const focus = Array.isArray(report?.retryFocusDimensionCodes) ? report.retryFocusDimensionCodes : null;
  const context = report?.context;
  const codes = new Set(dimensions.map((item) => item.code));
  const issueCodes = new Set(issues.map((item) => item.dimensionCode));
  const actionTypes = new Set(["retry_current_round", "next_round", "review_evidence"]);
  const validEvidence = [...highlights, ...issues].every((item) =>
    codes.has(item.dimensionCode) && isNonEmptyText(item.evidence) && !!CONFIDENCE_LABELS[item.confidence]
  );
  const validFocus = focus && focus.every((code) =>
    dimensions.some((item) => item.code === code && item.status === "needs_work") && issueCodes.has(code)
  );
  return !!report &&
    report.status === "ready" &&
    isNonEmptyText(report.id) &&
    isNonEmptyText(report.sessionId) &&
    isNonEmptyText(report.summary) &&
    !!READINESS_LABELS[report.preparednessLevel] &&
    !!context &&
    isNonEmptyText(context.targetJobTitle) &&
    isNonEmptyText(context.targetJobCompany) &&
    isNonEmptyText(context.resumeDisplayName) &&
    isNonEmptyText(context.roundName) &&
    typeof context.hasNextRound === "boolean" &&
    dimensions.length > 0 &&
    codes.size === dimensions.length &&
    dimensions.every((item) => isNonEmptyText(item.code) && isNonEmptyText(item.label) && !!DIMENSION_STATUS_LABELS[item.status] && !!CONFIDENCE_LABELS[item.confidence]) &&
    highlights.length + issues.length > 0 &&
    validEvidence &&
    actions.length > 0 &&
    actions.length <= 2 &&
    actions.every((item) => actionTypes.has(item.type) && isValidActionLabel(item.label, context.language)) &&
    validFocus;
};

const ReportScreen = ({ T, lang, nav, params = {} }) => {
  const report = window.EI_DATA.report;
  if (!params.reportId || report.id !== params.reportId) return <ReportMissingState T={T} lang={lang} nav={nav} />;
  if (report.status === "queued" || report.status === "generating") return <ReportPendingState T={T} lang={lang} nav={nav} reportId={report.id} targetJobId={report.targetJobId} />;
  if (report.status === "failed") return <ReportFailureState T={T} lang={lang} nav={nav} errorCode={report.errorCode} targetJobId={report.targetJobId} />;
  if (!isValidDirectReport(report)) return <ReportFailureState T={T} lang={lang} nav={nav} errorCode="INVALID_REPORT_CONTRACT" />;
  return <ReportDashboard T={T} lang={lang} nav={nav} report={report} />;
};

const ReportMissingState = ({ T, lang, nav }) => (
  <div className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px clamp(16px, 5vw, 48px)" }}>
    <Card T={T}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "REPORT NOT AVAILABLE" : "报告不可用"}</div>
      <div className="ei-serif" style={{ fontSize: 28, color: T.ink, marginBottom: 16 }}>{lang === "en" ? "Open a report from a completed interview." : "请从已完成的模拟面试打开报告。"}</div>
      <Btn T={T} variant="accent" onClick={() => nav("workspace")}>{lang === "en" ? "Back to interviews" : "返回面试"}</Btn>
    </Card>
  </div>
);

const ReportPendingState = ({ T, lang, nav, reportId, targetJobId }) => (
  <div className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px clamp(16px, 5vw, 48px)" }}>
    <Card T={T}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "REPORT IN PROGRESS" : "报告仍在生成"}</div>
      <div className="ei-serif" style={{ fontSize: 28, color: T.ink, marginBottom: 16 }}>{lang === "en" ? "We are still checking the evidence for this conversation." : "系统仍在核对这场对话的证据。"}</div>
      <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
        <Btn T={T} variant="accent" onClick={() => nav("generating", { reportId })}>{lang === "en" ? "View progress" : "查看生成状态"}</Btn>
        <Btn T={T} variant="secondary" onClick={() => targetJobId ? nav("reports", { targetJobId }) : nav("workspace")}>{targetJobId ? (lang === "en" ? "Back to interview reports" : "返回面试报告") : (lang === "en" ? "Back to interviews" : "返回面试")}</Btn>
      </div>
    </Card>
  </div>
);

const ReportFailureState = ({ T, lang, nav, errorCode, targetJobId }) => {
  const isOversize = errorCode === "REPORT_CONTEXT_TOO_LARGE";
  return (
    <div className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px clamp(16px, 5vw, 48px)" }}>
      <Card T={T}>
        <div className="ei-label" style={{ color: T.danger, marginBottom: 10 }}>{lang === "en" ? "REPORT UNAVAILABLE" : "报告不可用"}</div>
        <div className="ei-serif" style={{ fontSize: 28, color: T.ink, marginBottom: 12 }}>{lang === "en" ? "This report cannot be opened." : "这份报告暂时无法打开。"}</div>
        <p style={{ color: T.ink2, lineHeight: 1.65, margin: "0 0 18px" }}>
          {isOversize
            ? (lang === "en" ? "The source material and conversation were too long. Shorten the input in your interview plan, then start a new session." : "本次材料与对话过长。请返回面试规划，缩短输入后开启一场新会话。")
            : (lang === "en" ? "Return to your interviews and open another completed session." : "请返回面试，打开另一场已完成的会话。")}
        </p>
        <Btn T={T} variant="accent" onClick={() => targetJobId ? nav("reports", { targetJobId }) : nav("workspace")}>{targetJobId ? (lang === "en" ? "Back to interview reports" : "返回面试报告") : (lang === "en" ? "Back to interviews" : "返回面试")}</Btn>
      </Card>
    </div>
  );
};

const ReportConversationScreen = ({ T, lang, nav, params = {} }) => {
  const conversation = window.EI_DATA.reportConversation;
  if (!params.reportId) return <ReportConversationUnavailable T={T} lang={lang} nav={nav} />;
  if (conversation?.state === "loading") return <ReportConversationLoading T={T} lang={lang} nav={nav} />;
  if (conversation?.state === "network_error" || !isValidReportConversation(conversation, params.reportId)) {
    return <ReportConversationUnavailable T={T} lang={lang} nav={nav} />;
  }

  const back = () => {
    if (conversation.reportStatus === "ready") {
      nav("report", { reportId: conversation.reportId });
      return;
    }
    if (REPORT_CONVERSATION_REPORT_STATUSES.has(conversation.reportStatus)) {
      nav("generating", { reportId: conversation.reportId });
      return;
    }
    nav("workspace");
  };

  return (
    <main data-testid="report-conversation-screen" className="ei-fadein" style={{ maxWidth: 880, margin: "0 auto", padding: "32px clamp(16px, 5vw, 48px) 96px" }}>
      <button data-testid="report-conversation-back-button" type="button" onClick={back} style={{ border: 0, background: "transparent", color: T.ink3, cursor: "pointer", marginBottom: 20, padding: 0 }}>
        ← {lang === "en" ? "Back to report" : "返回报告"}
      </button>

      <header style={{ marginBottom: 24 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "INTERVIEW RECORD" : "本次面试记录"}</div>
        <h1 className="ei-serif" style={{ margin: 0, fontSize: 36, color: T.ink, overflowWrap: "anywhere" }}>
          {conversation.context.targetJobCompany} · {conversation.context.targetJobTitle}
        </h1>
        <p style={{ color: T.ink2, lineHeight: 1.7, margin: "10px 0 0" }}>
          {lang === "en" ? "A read-only record of the completed interview conversation." : "按原始顺序查看本场面试的完整对话记录。"}
        </p>
      </header>

      <div data-testid="report-conversation-context-strip">
        <ReportContextStrip T={T} lang={lang} report={{ context: conversation.context }} />
      </div>

      <section data-testid="report-conversation-transcript" aria-label={lang === "en" ? "Interview conversation record" : "面试对话记录"} style={{ minWidth: 0, overflowX: "hidden" }}>
        {conversation.messages.length > 0 ? (
          <ol style={{ margin: 0, padding: 0, listStyle: "none" }}>
            {conversation.messages.map((message) => <ReportConversationMessage key={message.sequence} message={message} T={T} lang={lang} />)}
          </ol>
        ) : <ReportConversationEmpty T={T} lang={lang} />}
      </section>
    </main>
  );
};

const REPORT_CONVERSATION_REPORT_STATUSES = new Set(["queued", "generating", "ready", "failed"]);
const REPORT_CONVERSATION_CONTEXT_KEYS = new Set([
  "sourcePlanId",
  "targetJobTitle",
  "targetJobCompany",
  "resumeId",
  "resumeDisplayName",
  "roundId",
  "roundSequence",
  "roundName",
  "roundType",
  "language",
  "hasNextRound",
]);
const REPORT_CONVERSATION_MESSAGE_KEYS = new Set(["sequence", "role", "content", "createdAt"]);
const REPORT_CONVERSATION_MESSAGE_ROLES = new Set(["user", "assistant"]);

const hasExactConversationKeys = (value, keys) => {
  const actualKeys = Object.keys(value || {});
  return actualKeys.length === keys.size && actualKeys.every((key) => keys.has(key));
};

const isValidReportConversation = (conversation, reportId) => {
  if (!conversation || conversation.state !== "ready" || conversation.reportId !== reportId) return false;
  if (!REPORT_CONVERSATION_REPORT_STATUSES.has(conversation.reportStatus)) return false;
  if (!hasExactConversationKeys(conversation.context, REPORT_CONVERSATION_CONTEXT_KEYS)) return false;
  if (!Array.isArray(conversation.messages)) return false;
  const context = conversation.context;
  if (
    !isNonEmptyText(context.sourcePlanId) ||
    !isNonEmptyText(context.targetJobTitle) ||
    !isNonEmptyText(context.targetJobCompany) ||
    !isNonEmptyText(context.resumeId) ||
    !isNonEmptyText(context.resumeDisplayName) ||
    !isNonEmptyText(context.roundId) ||
    !Number.isInteger(context.roundSequence) ||
    context.roundSequence < 1 ||
    !isNonEmptyText(context.roundName) ||
    !isNonEmptyText(context.roundType) ||
    !isNonEmptyText(context.language) ||
    typeof context.hasNextRound !== "boolean"
  ) return false;

  let previousSequence = 0;
  return conversation.messages.every((message) => {
    if (!hasExactConversationKeys(message, REPORT_CONVERSATION_MESSAGE_KEYS)) return false;
    if (
      !Number.isInteger(message.sequence) ||
      message.sequence < 1 ||
      message.sequence <= previousSequence ||
      !REPORT_CONVERSATION_MESSAGE_ROLES.has(message.role) ||
      !isNonEmptyText(message.content) ||
      !isNonEmptyText(message.createdAt)
    ) return false;
    previousSequence = message.sequence;
    return true;
  });
};

const ReportConversationLoading = ({ T, lang, nav }) => (
  <main data-testid="report-conversation-loading" className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px clamp(16px, 5vw, 48px)" }}>
    <Card T={T}>
      <div role="status" aria-live="polite">
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "LOADING RECORD" : "正在读取记录"}</div>
        <div className="ei-serif" style={{ color: T.ink, fontSize: 26, marginBottom: 10 }}>{lang === "en" ? "Opening this interview record…" : "正在打开本次面试记录…"}</div>
        <p style={{ color: T.ink2, lineHeight: 1.65, margin: "0 0 18px" }}>{lang === "en" ? "The conversation will appear only after its report context is verified." : "仅在报告上下文核验完成后显示对话内容。"}</p>
        <Btn T={T} variant="secondary" onClick={() => nav("workspace")}>{lang === "en" ? "Back to interviews" : "返回面试"}</Btn>
      </div>
    </Card>
  </main>
);

const ReportConversationUnavailable = ({ T, lang, nav }) => (
  <main data-testid="report-conversation-unavailable" className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px clamp(16px, 5vw, 48px)" }}>
    <Card T={T}>
      <div role="alert">
        <div className="ei-label" style={{ color: T.danger, marginBottom: 10 }}>{lang === "en" ? "RECORD UNAVAILABLE" : "记录暂时不可用"}</div>
        <div className="ei-serif" style={{ color: T.ink, fontSize: 26, marginBottom: 10 }}>{lang === "en" ? "This interview record cannot be opened." : "无法打开本次面试记录。"}</div>
        <p style={{ color: T.ink2, lineHeight: 1.65, margin: "0 0 18px" }}>{lang === "en" ? "Return to your interviews and open another verified report." : "请返回面试，打开另一份已核验的报告。"}</p>
        <Btn T={T} variant="secondary" onClick={() => nav("workspace")}>{lang === "en" ? "Back to interviews" : "返回面试"}</Btn>
      </div>
    </Card>
  </main>
);

const ReportConversationEmpty = ({ T, lang }) => (
  <div data-testid="report-conversation-empty" role="status" style={{ padding: "20px 0 28px", color: T.ink3, fontSize: 13, lineHeight: 1.65 }}>
    <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "NO MESSAGES" : "暂无对话消息"}</div>
    <div style={{ color: T.ink2 }}>{lang === "en" ? "This verified interview record does not contain any messages." : "这份已核验的面试记录暂未包含对话消息。"}</div>
  </div>
);

const ReportConversationMessage = ({ message, T, lang }) => {
  const isAssistant = message.role === "assistant";
  return (
    <li data-testid={`report-conversation-message-${message.sequence}`} data-role={message.role} style={{ display: "flex", gap: 12, minWidth: 0, marginBottom: 20 }}>
      <div style={{ width: 28, height: 28, borderRadius: 2, flexShrink: 0, background: isAssistant ? T.accentSoft : T.bgSoft, color: isAssistant ? T.accent : T.ink2, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontFamily: "var(--ei-mono)", fontWeight: 500 }}>
        {isAssistant ? "AI" : (lang === "en" ? "ME" : "我")}
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ color: T.ink2, fontSize: 12, fontWeight: 500, marginBottom: 5 }}>{isAssistant ? (lang === "en" ? "Interviewer" : "面试官") : (lang === "en" ? "You" : "我")}</div>
        <PracticeMessageBody text={message.content} T={T} />
      </div>
    </li>
  );
};

const ReportDashboard = ({ T, lang, nav, report }) => {
  const context = report.context;
  const dimensions = report.dimensionAssessments;
  const highlights = report.highlights || [];
  const issues = report.issues || [];
  const actions = report.nextActions;
  const firstAction = actions[0];
  const labelsByCode = new Map(dimensions.map((item) => [item.code, item.label]));
  const retryRequest = { goal: "retry_current_round", sourceReportId: report.id };
  const nextRequest = { goal: "next_round", sourceReportId: report.id };
  const retryPrimary = firstAction?.type === "retry_current_round";
  const nextPrimary = firstAction?.type === "next_round";

  return (
    <main className="ei-fadein" data-testid="report-dashboard" style={{ maxWidth: 1120, margin: "0 auto", padding: "32px clamp(16px, 5vw, 48px) 96px" }}>
      <button data-testid="report-back-button" onClick={() => report.targetJobId ? nav("reports", { targetJobId: report.targetJobId }) : nav("workspace")} style={{ border: 0, background: "transparent", color: T.ink3, cursor: "pointer", marginBottom: 20 }}>← {report.targetJobId ? (lang === "en" ? "Interview reports" : "面试报告") : (lang === "en" ? "Interviews" : "面试")}</button>

      <header data-testid="report-header" style={{ display: "flex", justifyContent: "space-between", gap: 24, alignItems: "flex-end", flexWrap: "wrap", marginBottom: 24 }}>
        <div style={{ minWidth: 0, flex: "1 1 440px" }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "CONVERSATION REPORT" : "会话报告"}</div>
          <h1 className="ei-serif" style={{ margin: 0, fontSize: 38, color: T.ink, overflowWrap: "anywhere" }}>{context.targetJobCompany} · {context.targetJobTitle}</h1>
          <p style={{ color: T.ink2, lineHeight: 1.7, marginBottom: 0 }}>{lang === "en" ? "Evidence and capability signals from the complete interview conversation." : "基于整场模拟面试对话整理证据、能力表现与下一步。"}</p>
        </div>
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
          <Btn T={T} variant={retryPrimary ? "accent" : "secondary"} onClick={() => nav("practice", retryRequest)}>{lang === "en" ? "Practice again" : "复练当前轮"}</Btn>
          <Btn T={T} variant={nextPrimary ? "accent" : "secondary"} disabled={!report.context.hasNextRound} ariaDescribedby="report-next-disabled-reason" onClick={() => report.context.hasNextRound && nav("practice", nextRequest)}>{lang === "en" ? "Next round" : "进入下一轮"}</Btn>
          {!report.context.hasNextRound && <span id="report-next-disabled-reason" style={{ flexBasis: "100%", color: T.ink3, fontSize: 11, lineHeight: 1.35, textAlign: "right" }}>{lang === "en" ? "There is no next round for this target." : "当前岗位没有下一轮可进入。"}</span>}
        </div>
      </header>

      <ReportContextStrip T={T} lang={lang} report={report} marginBottom={10} />
      <button data-testid="report-conversation-entry" type="button" onClick={() => nav("report-conversation", { reportId: report.id })} style={{ display: "inline-flex", alignItems: "center", gap: 5, margin: "0 0 22px", padding: 0, border: 0, background: "transparent", color: T.ink3, fontSize: 12.5, fontFamily: "var(--ei-sans)", cursor: "pointer" }}>
        <Icon name="arrow_right" size={13} /> {lang === "en" ? "View this interview record" : "查看本次面试记录"}
      </button>

      <section data-testid="report-summary-cards" style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))", gap: 14, marginBottom: 22 }}>
        <ReportMetric T={T} label={lang === "en" ? "READINESS" : "准备度"} value={localizeReadiness(report.preparednessLevel, lang)} description={report.summary} />
        <ReportMetric T={T} label={lang === "en" ? "CAPABILITY DIMENSIONS" : "能力维度"} value={`${dimensions.length}`} />
        <ReportMetric T={T} label={lang === "en" ? "CONVERSATION EVIDENCE" : "会话证据"} value={`${highlights.length + issues.length}`} />
      </section>

      <section data-testid="report-detail-grid" style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 420px), 1fr))", gap: 18 }}>
        <div data-testid="report-dimensions"><Card T={T}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "CAPABILITY ASSESSMENT" : "能力维度评估"}</div>
          {dimensions.map((item, index) => (
            <div className="ei-report-dimension-row" key={item.code} style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", flexWrap: "wrap", gap: "8px 16px", padding: "13px 0", borderBottom: index < dimensions.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
              <span className="ei-report-dimension-label" style={{ color: T.ink, minWidth: 0, flex: "1 1 160px", overflowWrap: "break-word", wordBreak: "normal" }}>{item.label}</span>
              <span className="ei-report-dimension-status" style={{ color: item.status === "needs_work" ? T.warn : T.ok, textAlign: "right", flex: "0 1 auto", maxWidth: "100%", overflowWrap: "break-word", wordBreak: "normal" }}>{localizeDimensionStatus(item.status, lang)} · {localizeConfidence(item.confidence, lang)}</span>
            </div>
          ))}
        </Card></div>
        <div data-testid="report-highlights"><EvidenceCard T={T} title={lang === "en" ? "STRENGTH EVIDENCE" : "优势证据"} items={highlights} labelsByCode={labelsByCode} color={T.ok} lang={lang} /></div>
        <div data-testid="report-issues"><EvidenceCard T={T} title={lang === "en" ? "RISKS" : "待加强证据"} items={issues} labelsByCode={labelsByCode} color={T.warn} lang={lang} /></div>
        <div data-testid="report-actions"><ActionCard T={T} lang={lang} actions={actions} /></div>
      </section>
    </main>
  );
};

const ReportContextStrip = ({ T, lang, report, marginBottom = 22 }) => {
  const context = report.context;
  const items = [
    ["job", lang === "en" ? "TARGET" : "目标岗位", `${context.targetJobCompany} · ${context.targetJobTitle}`],
    ["round", lang === "en" ? "ROUND" : "轮次", context.roundName],
    ["resume", lang === "en" ? "RESUME" : "简历", context.resumeDisplayName],
  ];
  return (
    <section data-testid="report-context-strip" style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))", gap: 1, border: `1px solid ${T.rule}`, background: T.rule, marginBottom }}>
      {items.map(([id, label, value]) => <div key={id} data-testid={`report-context-${id}`} style={{ minWidth: 0, padding: "12px 14px", background: T.bgCard }}><div className="ei-label" style={{ color: T.ink3, marginBottom: 5 }}>{label}</div><div title={value} style={{ color: T.ink2, fontSize: 12.5, lineHeight: 1.5, overflowWrap: "anywhere" }}>{value}</div></div>)}
    </section>
  );
};

const ReportMetric = ({ T, label, value, description }) => <div style={{ padding: 20, border: `1px solid ${T.rule}`, background: T.bgCard, minWidth: 0 }}><div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{label}</div><div className="ei-serif" style={{ color: T.ink, fontSize: 24, overflowWrap: "anywhere" }}>{value}</div>{description && <div style={{ color: T.ink2, fontSize: 13, lineHeight: 1.65, marginTop: 10, overflowWrap: "anywhere" }}>{description}</div>}</div>;

const EvidenceCard = ({ T, title, items, labelsByCode, color, lang }) => <Card T={T}><div className="ei-label" style={{ color, marginBottom: 12 }}>{title}</div>{items.map((item, index) => <div key={`${item.dimensionCode}-${index}`} style={{ color: T.ink2, fontSize: 13, lineHeight: 1.65, marginTop: index ? 14 : 0, overflowWrap: "anywhere" }}><div style={{ color: T.ink, fontWeight: 500, marginBottom: 3 }}>{labelsByCode.get(item.dimensionCode)}</div><div>{item.evidence}</div><div style={{ color: T.ink3, fontSize: 11.5, marginTop: 4 }}>{localizeConfidence(item.confidence, lang)}</div></div>)}</Card>;

const ActionCard = ({ T, lang, actions }) => <Card T={T}><div className="ei-label" style={{ color: T.accent, marginBottom: 12 }}>{lang === "en" ? "NEXT ACTIONS" : "下一步行动"}</div>{actions.map((item, index) => <div className="ei-report-action-row" key={`${item.type}-${index}`} style={{ display: "flex", minWidth: 0, gap: 10, color: T.ink2, fontSize: 13, lineHeight: 1.65, marginTop: index ? 12 : 0, overflowWrap: "anywhere", wordBreak: "normal" }}><span style={{ color: T.accent, fontFamily: "var(--ei-mono)", flexShrink: 0 }}>{String(index + 1).padStart(2, "0")}</span><span className="ei-report-action-label" style={{ minWidth: 0, overflowWrap: "anywhere", wordBreak: "normal" }}>{item.label}</span></div>)}</Card>;

window.ReportScreen = ReportScreen;
