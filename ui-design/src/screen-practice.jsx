// Screen 3: continuous text interview conversation
const PRACTICE_REPLY_STATE_DEMOS = {
  "immediate-pending": [
    { id: "demo-opening", role: "assistant", text: "你好，我们直接开始。先聊聊你最近最有代表性的项目。", t: "08:02" },
    { id: "demo-complete-user", role: "user", text: "我主导过一次跨团队设计系统迁移。", t: "08:03", status: "complete" },
    { id: "demo-follow-up", role: "assistant", text: "当时最难协调的分歧是什么？", t: "08:03" },
    { id: "demo-immediate-pending", role: "user", text: "我会先建立可回滚基线，再逐步放量。", t: "00:00", status: "pending" },
  ],
  "persisted-pending": [
    { id: "demo-risk-opening", role: "assistant", text: "请说明你如何处理一次高风险发布。", t: "08:02" },
    { id: "demo-persisted-pending", role: "user", text: "我先把风险拆成三类。", t: "08:03", status: "pending" },
  ],
  "retryable-failed": [
    { id: "demo-risk-opening", role: "assistant", text: "请说明你如何处理一次高风险发布。", t: "08:02" },
    { id: "demo-retryable-failed", role: "user", text: "我先把风险拆成三类。", t: "08:03", status: "retryable_failed" },
  ],
  "terminal-failed": [
    { id: "demo-risk-opening", role: "assistant", text: "请说明你如何处理一次高风险发布。", t: "08:02" },
    { id: "demo-terminal-failed", role: "user", text: "我先把风险拆成三类。", t: "08:03", status: "terminal_failed" },
  ],
};

const PracticeScreen = ({ T, lang, nav, params = {}, jobId }) => {
  const D = window.EI_DATA;
  const context = window.eiCreateInterviewContext ? window.eiCreateInterviewContext(params) : params;
  const job = D.targetJobs.find((item) => item.id === (context.targetJobId || jobId)) || D.targetJobs[0];
  const { currentRound } = window.eiResolveInterviewRoundContext(D.jdSample.interviewRounds, context.roundId);
  const replyStateDemo = PRACTICE_REPLY_STATE_DEMOS[params.replyState] || null;
  const [input, setInput] = React.useState("");
  const [paused, setPaused] = React.useState(false);
  const [elapsed, setElapsed] = React.useState(replyStateDemo ? 0 : 502);
  const [messages, setMessages] = React.useState(() => (replyStateDemo || D.sessionTranscript).map((message, index) => ({
    ...message,
    id: message.id || `prototype-message-${index + 1}`,
    status: message.role === "user" ? (message.status || "complete") : undefined,
  })));
  const [pendingMessageId, setPendingMessageId] = React.useState(null);

  React.useEffect(() => {
    if (paused) return;
    const timer = setInterval(() => setElapsed((value) => value + 1), 1000);
    return () => clearInterval(timer);
  }, [paused]);

  const formatElapsed = (seconds) => `${String(Math.floor(seconds / 60)).padStart(2, "0")}:${String(seconds % 60).padStart(2, "0")}`;
  const budget = replyStateDemo ? "50:00" : currentRound ? formatElapsed(currentRound.durationMinutes * 60) : "--:--";
  const companyName = replyStateDemo ? "Acme" : job.company;
  const jobTitle = replyStateDemo ? "Senior Frontend Engineer" : job.title;
  const interviewerRole = context.roundName || (lang === "en" ? "Manager round" : "经理面");
  const interviewerLabel = interviewerRole;
  const hasCommittedCandidateMessage = messages.some((message) => message.role === "user");
  const hasFailedCandidateMessage = messages.some((message) => message.role === "user" && message.status !== "complete");
  const hasTerminalFailedMessage = messages.some((message) => message.role === "user" && message.status === "terminal_failed");
  const isThinking = pendingMessageId !== null || messages.some((message) => message.role === "user" && ["pending", "retrying"].includes(message.status));
  const canFinishInterview = hasCommittedCandidateMessage && !isThinking && !hasFailedCandidateMessage;
  const finishReasonId = "practice-finish-disabled-reason";
  const finishDisabledReason = isThinking ? (lang === "en" ? "Wait for the interviewer reply." : "请等待面试官回复。") : hasFailedCandidateMessage ? (lang === "en" ? "Resolve the unfinished reply first." : "请先恢复这条未完成回复的消息。") : (lang === "en" ? "Complete at least one answer first." : "请先完成至少一次回答。");
  const requestReply = (message) => {
    setPendingMessageId(message.id);
    return new Promise((resolve) => {
      setTimeout(resolve, 500);
    })
      .then(() => {
        setMessages((current) => {
          const completed = current.map((item) => item.id === message.id ? { ...item, status: "complete" } : item);
          if (completed.some((item) => item.replyToMessageId === message.id)) return completed;
          return [...completed, {
            id: `prototype-reply-${message.id}`,
            role: "ai",
            replyToMessageId: message.id,
            text: lang === "en" ? "Interesting. Walk me through one concrete tradeoff and its impact." : "有意思。请继续说说其中一个具体取舍，以及它带来的影响。",
            t: formatElapsed(elapsed + 20),
          }];
        });
      })
      .catch(() => {
        setMessages((current) => current.map((item) => item.id === message.id ? { ...item, status: "retryable_failed" } : item));
      })
      .finally(() => {
        setPendingMessageId((current) => current === message.id ? null : current);
      });
  };
  const retryFailedMessage = (message) => {
    if (pendingMessageId !== null || message.status !== "retryable_failed") return;
    setMessages((current) => current.map((item) => item.id === message.id ? { ...item, status: "pending" } : item));
    void requestReply(message);
  };
  const send = () => {
    const text = input.trim();
    if (!text || paused || isThinking || hasFailedCandidateMessage) return;
    const message = { id: `prototype-message-${Date.now()}`, role: "user", text, t: formatElapsed(elapsed), status: "pending" };
    setMessages((current) => [...current, message]);
    setInput("");
    void requestReply(message);
  };

  return (
    <div data-testid="practice-screen" className="ei-fadein" style={{ height: "100vh", display: "flex", flexDirection: "column", background: T.bg, overflow: "hidden" }}>
      <div data-testid="practice-topbar" style={{ padding: "14px 28px", borderBottom: `1px solid ${T.rule}`, display: "flex", alignItems: "center", flexWrap: "wrap", gap: 16, background: T.bgCard }}>
        <div style={{ minWidth: 0 }}>
          <div data-testid="practice-topbar-company" style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", textTransform: "uppercase" }}>{companyName}</div>
          <div data-testid="practice-topbar-title" style={{ fontSize: 14, color: T.ink, fontWeight: 500 }}>{jobTitle}</div>
        </div>
        <div style={{ flex: "1 1 80px" }} />
        <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap", justifyContent: "flex-end" }}>
          <span data-testid="practice-topbar-interviewer" className="ei-mono" style={{ display: "inline-flex", alignItems: "center", padding: "3px 8px", borderRadius: 3, fontSize: 11.5, background: T.bgSoft, color: T.ink3, border: `1px solid ${T.rule}` }}>{interviewerLabel}</span>
          <span data-testid="practice-topbar-timer" className="ei-mono" style={{ display: "inline-flex", alignItems: "center", padding: "3px 8px", borderRadius: 3, fontSize: 11.5, background: T.bgSoft, color: T.ink3 }}>{formatElapsed(elapsed)} / {budget}</span>
          <button data-testid="practice-topbar-pause" type="button" onClick={() => setPaused((value) => !value)} aria-pressed={paused} style={{ background: "transparent", border: `1px solid ${T.rule}`, padding: "6px 10px", borderRadius: 2, display: "flex", gap: 6, alignItems: "center", color: T.ink2, fontSize: 12, cursor: "pointer" }}>
            {paused ? "▶" : "❚❚"} {paused ? (lang === "en" ? "Resume" : "继续") : (lang === "en" ? "Pause" : "暂停")}
          </button>
          <div style={{ height: 18, width: 1, background: T.rule }} />
          <button
            type="button"
            data-testid="practice-topbar-phone-toggle"
            disabled
            aria-disabled="true"
            aria-label={lang === "en" ? "Phone mode is temporarily unavailable" : "电话模式暂未开放"}
            title={lang === "en" ? "Phone mode is temporarily unavailable" : "电话模式暂未开放"}
            style={{ width: 34, height: 34, padding: 0, borderRadius: 17, border: `1px solid ${T.rule}`, background: T.bgSoft, color: T.ink4, display: "inline-flex", alignItems: "center", justifyContent: "center", cursor: "not-allowed", opacity: 0.58 }}
          >
            <svg aria-hidden="true" viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
              <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6A19.79 19.79 0 0 1 2.12 4.18 2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.12.9.33 1.78.62 2.63a2 2 0 0 1-.45 2.11L8 9.73a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.85.29 1.73.5 2.63.62A2 2 0 0 1 22 16.92z" />
            </svg>
          </button>
          <div style={{ height: 18, width: 1, background: T.rule }} />
          <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: 5 }}>
            <button
              type="button"
              data-testid="practice-finish-cta"
              disabled={!canFinishInterview}
              aria-describedby={!canFinishInterview ? finishReasonId : undefined}
              onClick={() => canFinishInterview && nav("generating", { reportId: D.report.id })}
              style={{ padding: "7px 12px", background: canFinishInterview ? T.accent : T.bgSoft, color: canFinishInterview ? "#fff" : T.ink4, border: canFinishInterview ? "none" : `1px solid ${T.rule}`, borderRadius: 2, cursor: canFinishInterview ? "pointer" : "not-allowed", fontSize: 12.5, fontWeight: 500, fontFamily: "var(--ei-sans)" }}
            >
              {lang === "en" ? "Finish report" : "结束并生成报告"}
            </button>
            {!canFinishInterview && <span id={finishReasonId} data-testid="practice-finish-disabled-reason" style={{ maxWidth: 190, color: T.ink3, fontSize: 11, lineHeight: 1.35, textAlign: "right" }}>{finishDisabledReason}</span>}
          </div>
        </div>
      </div>

      <main data-testid="practice-conversation" style={{ flex: 1, minHeight: 0, display: "flex", flexDirection: "column", width: "100%" }}>
        <div data-testid="practice-transcript" style={{ flex: 1, overflowY: "auto", padding: "28px clamp(24px, 8vw, 144px) 20px" }}>
          {messages.map((message, index) => <TranscriptMsg key={message.id} index={index} msg={message} T={T} lang={lang} retryFailedMessage={retryFailedMessage} />)}
          {isThinking && (
            <div data-testid="practice-interviewer-thinking" role="status" aria-live="polite" style={{ marginBottom: 18, display: "flex", gap: 12 }}>
              <div style={{ width: 28, height: 28, borderRadius: 2, flexShrink: 0, background: T.accentSoft, color: T.accent, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontFamily: "var(--ei-mono)", fontWeight: 500 }}>AI</div>
              <div style={{ minWidth: 0, display: "flex", alignItems: "center", gap: 8, color: T.ink3, fontSize: 13.5 }}>
                <span>{lang === "en" ? "The interviewer is thinking" : "面试官正在思考"}</span>
                <span className="ei-pulse" aria-hidden="true">● ● ●</span>
              </div>
            </div>
          )}
          <div data-testid="practice-transcript-helper" style={{ textAlign: "center", marginTop: 12, fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
            {lang === "en" ? "— continue naturally, or finish the interview when ready —" : "— 自然继续对话，准备好后可结束面试 —"}
          </div>
        </div>
        <div data-testid="practice-input" style={{ padding: "16px clamp(24px, 8vw, 144px) 24px", borderTop: `1px solid ${T.rule}`, background: T.bgCard }}>
          {hasTerminalFailedMessage && (
            <div data-testid="practice-terminal-recovery" role="alert" style={{ marginBottom: 12, padding: "12px 14px", border: `1px solid ${T.rule}`, borderRadius: 2, background: T.bgSoft, display: "flex", alignItems: "center", justifyContent: "space-between", gap: 14, flexWrap: "wrap" }}>
              <div style={{ minWidth: 0 }}>
                <div style={{ color: T.ink, fontSize: 13.5, fontWeight: 500, marginBottom: 3 }}>{lang === "en" ? "This reply could not be completed." : "本次回复未能完成。"}</div>
                <div style={{ color: T.ink3, fontSize: 12, lineHeight: 1.5 }}>{lang === "en" ? "Return to this interview plan, then start a new session when you are ready." : "请返回当前面试规划，准备好后重新开始一场面试。"}</div>
              </div>
              <Btn variant="secondary" size="sm" T={T} icon="arrow_left" data-testid="practice-terminal-recovery-cta" onClick={() => nav("parse", { targetJobId: job.id })}>
                {lang === "en" ? "Return to this interview plan" : "返回当前面试规划"}
              </Btn>
            </div>
          )}
          <div style={{ border: `1px solid ${T.rule}`, borderRadius: 2, padding: 12, background: T.bg }}>
            <textarea data-testid="practice-input-textarea" value={input} onChange={(event) => setInput(event.target.value)} disabled={paused || isThinking || hasTerminalFailedMessage} placeholder={isThinking ? (lang === "en" ? "The interviewer is thinking…" : "面试官正在思考…") : hasTerminalFailedMessage ? (lang === "en" ? "Return to the interview plan to continue." : "请返回面试规划后继续。") : (lang === "en" ? "Type your message here." : "在这里输入消息。")} style={{ width: "100%", minHeight: 74, border: "none", outline: "none", resize: "none", fontSize: 14, lineHeight: 1.55, background: "transparent", color: T.ink, fontFamily: "var(--ei-sans)" }} />
            <div style={{ display: "flex", justifyContent: "flex-end", marginTop: 6 }}>
              <button data-testid="practice-input-send" type="button" disabled={paused || isThinking || hasFailedCandidateMessage || !input.trim()} onClick={send} style={{ background: T.accent, color: "#fff", border: `1px solid ${T.accent}`, padding: "6px 14px", borderRadius: 2, fontSize: 12, fontWeight: 500, cursor: paused || isThinking || hasFailedCandidateMessage || !input.trim() ? "not-allowed" : "pointer", opacity: paused || isThinking || hasFailedCandidateMessage || !input.trim() ? 0.5 : 1 }}>{lang === "en" ? "Send" : "发送"}</button>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};

const TranscriptMsg = ({ index, msg, T, lang, retryFailedMessage }) => {
  const isAI = msg.role === "ai" || msg.role === "assistant";
  return (
    <div data-testid={`practice-transcript-message-${index}`} data-role={isAI ? "assistant" : "user"} style={{ marginBottom: 18, display: "flex", gap: 12 }}>
      <div style={{ width: 28, height: 28, borderRadius: 2, flexShrink: 0, background: isAI ? T.accentSoft : T.bgSoft, color: isAI ? T.accent : T.ink2, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontFamily: "var(--ei-mono)", fontWeight: 500 }}>{isAI ? "AI" : (lang === "en" ? "ME" : "我")}</div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 4 }}>
          <span style={{ fontSize: 12, color: T.ink2, fontWeight: 500 }}>{isAI ? (lang === "en" ? "Interviewer" : "面试官") : (lang === "en" ? "You" : "我")}</span>
          <span style={{ fontSize: 11, color: T.ink4, fontFamily: "var(--ei-mono)" }}>{msg.t}</span>
        </div>
        <div style={{ fontSize: 14, color: T.ink, lineHeight: 1.6 }}>{msg.text}</div>
        {!isAI && msg.status === "retryable_failed" && (
          <button type="button" data-testid="practice-message-retry" aria-label={lang === "en" ? "Retry message" : "重试这条消息"} title={lang === "en" ? "Retry message" : "重试这条消息"} onClick={() => retryFailedMessage(msg)} style={{ marginTop: 7, width: 28, height: 28, display: "inline-flex", alignItems: "center", justifyContent: "center", border: `1px solid ${T.rule}`, borderRadius: 2, background: T.bgCard, color: T.accent, padding: 0 }}>
            <svg aria-hidden="true" viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M20 11a8 8 0 1 0 2 5" /><path d="M20 4v7h-7" /></svg>
          </button>
        )}
      </div>
    </div>
  );
};

window.PracticeScreen = PracticeScreen;
