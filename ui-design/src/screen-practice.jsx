// Screen 3: continuous text interview conversation
const PracticeScreen = ({ T, lang, nav, params = {}, jobId }) => {
  const D = window.EI_DATA;
  const context = window.eiCreateInterviewContext ? window.eiCreateInterviewContext(params) : params;
  const job = D.targetJobs.find((item) => item.id === (context.targetJobId || jobId)) || D.targetJobs[0];
  const [input, setInput] = React.useState("");
  const [paused, setPaused] = React.useState(false);
  const [elapsed, setElapsed] = React.useState(502);
  const [messages, setMessages] = React.useState(D.sessionTranscript);

  React.useEffect(() => {
    if (paused) return;
    const timer = setInterval(() => setElapsed((value) => value + 1), 1000);
    return () => clearInterval(timer);
  }, [paused]);

  const formatElapsed = (seconds) => `${String(Math.floor(seconds / 60)).padStart(2, "0")}:${String(seconds % 60).padStart(2, "0")}`;
  const interviewerRole = context.roundName || (lang === "en" ? "Manager round" : "经理面");
  const interviewerLabel = lang === "en" ? `${interviewerRole} interviewer` : `${interviewerRole}面试官`;
  const send = () => {
    const text = input.trim();
    if (!text || paused) return;
    setMessages((current) => [...current, { role: "user", text, t: formatElapsed(elapsed) }]);
    setInput("");
    setTimeout(() => setMessages((current) => [...current, {
      role: "ai",
      text: lang === "en" ? "Interesting. Walk me through one concrete tradeoff and its impact." : "有意思。请继续说说其中一个具体取舍，以及它带来的影响。",
      t: formatElapsed(elapsed + 20),
    }]), 500);
  };

  return (
    <div className="ei-fadein" style={{ height: "100vh", display: "flex", flexDirection: "column", background: T.bg }}>
      <div style={{ padding: "14px 28px", borderBottom: `1px solid ${T.rule}`, display: "flex", alignItems: "center", gap: 16, background: T.bgCard }}>
        <div>
          <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{job.company.toUpperCase()}</div>
          <div style={{ fontSize: 14, color: T.ink, fontWeight: 500 }}>{job.title}</div>
        </div>
        <div style={{ flex: 1 }} />
        <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap", justifyContent: "flex-end" }}>
          <Tag tone="muted" T={T}><Icon name="briefcase" size={11} style={{ marginRight: 4 }} />{interviewerLabel}</Tag>
          <Tag tone="muted" T={T}><Icon name="clock" size={11} style={{ marginRight: 4 }} />{formatElapsed(elapsed)} / 25:00</Tag>
          <button onClick={() => setPaused((value) => !value)} style={{ background: "transparent", border: `1px solid ${T.rule}`, padding: "6px 10px", borderRadius: 2, display: "flex", gap: 6, alignItems: "center", color: T.ink2, fontSize: 12 }}>
            <Icon name={paused ? "play" : "pause"} size={12} /> {paused ? (lang === "en" ? "Resume" : "继续") : (lang === "en" ? "Pause" : "暂停")}
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
            <Icon name="phone" size={15} />
          </button>
          <div style={{ height: 18, width: 1, background: T.rule }} />
          <button onClick={() => nav("generating", { ...context })} style={{ padding: "7px 12px", background: T.accent, color: "#fff", border: "none", borderRadius: 2, cursor: "pointer", fontSize: 12.5, fontWeight: 500, display: "flex", alignItems: "center", gap: 6 }}>
            <Icon name="check" size={13} />{lang === "en" ? "Finish report" : "结束并生成报告"}
          </button>
        </div>
      </div>

      <main data-testid="practice-conversation" style={{ flex: 1, minHeight: 0, display: "flex", flexDirection: "column", width: "100%" }}>
        <div data-testid="practice-transcript" className="ei-scroll" style={{ flex: 1, overflowY: "auto", padding: "28px clamp(24px, 8vw, 144px) 20px" }}>
          {messages.map((message, index) => <TranscriptMsg key={index} msg={message} T={T} lang={lang} />)}
          <div style={{ textAlign: "center", marginTop: 12, fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
            {lang === "en" ? "— continue naturally, or finish the interview when ready —" : "— 自然继续对话，准备好后可结束面试 —"}
          </div>
        </div>
        <div style={{ padding: "16px clamp(24px, 8vw, 144px) 24px", borderTop: `1px solid ${T.rule}`, background: T.bgCard }}>
          <div style={{ border: `1px solid ${T.rule}`, borderRadius: 2, padding: 12, background: T.bg }}>
            <textarea value={input} onChange={(event) => setInput(event.target.value)} disabled={paused} placeholder={lang === "en" ? "Type your message here." : "在这里输入消息。"} style={{ width: "100%", minHeight: 74, border: "none", outline: "none", resize: "none", fontSize: 14, lineHeight: 1.55, background: "transparent", color: T.ink, fontFamily: "var(--ei-sans)" }} />
            <div style={{ display: "flex", justifyContent: "flex-end", marginTop: 6 }}>
              <Btn variant="accent" size="sm" T={T} icon="send" onClick={send}>{lang === "en" ? "Send" : "发送"}</Btn>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};

const TranscriptMsg = ({ msg, T, lang }) => {
  const isAI = msg.role === "ai" || msg.role === "assistant";
  return (
    <div style={{ marginBottom: 18, display: "flex", gap: 12 }}>
      <div style={{ width: 28, height: 28, borderRadius: 2, flexShrink: 0, background: isAI ? T.accentSoft : T.bgSoft, color: isAI ? T.accent : T.ink2, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontFamily: "var(--ei-mono)", fontWeight: 500 }}>{isAI ? "AI" : (lang === "en" ? "ME" : "我")}</div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 4 }}>
          <span style={{ fontSize: 12, color: T.ink2, fontWeight: 500 }}>{isAI ? (lang === "en" ? "Interviewer" : "面试官") : (lang === "en" ? "You" : "我")}</span>
          <span style={{ fontSize: 11, color: T.ink4, fontFamily: "var(--ei-mono)" }}>{msg.t}</span>
        </div>
        <div style={{ fontSize: 14, color: T.ink, lineHeight: 1.6 }}>{msg.text}</div>
      </div>
    </div>
  );
};

window.PracticeScreen = PracticeScreen;
