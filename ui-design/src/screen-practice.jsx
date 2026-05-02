// Screen 3: Mock Interview in progress
const PracticeScreen = ({ T, lang, nav, jobId, mode, role, setRole }) => {
  const D = window.EI_DATA;
  const job = D.targetJobs.find((j) => j.id === jobId) || D.targetJobs[0];
  const [qIdx, setQIdx] = React.useState(1);
  const [input, setInput] = React.useState("");
  const [paused, setPaused] = React.useState(false);
  const [showHint, setShowHint] = React.useState(false);
  const [dictating, setDictating] = React.useState(false);
  const [elapsed, setElapsed] = React.useState(502); // 08:22
  const [transcript, setTranscript] = React.useState(D.sessionTranscript);

  React.useEffect(() => {
    if (paused) return;
    const t = setInterval(() => setElapsed((e) => e + 1), 1000);
    return () => clearInterval(t);
  }, [paused]);

  const fmt = (s) => `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;
  const currentQ = D.questions[qIdx];
  const activeMode = mode === "voice" ? "voice" : "text";
  const modes = lang === "en"
    ? [
      { k: "text", label: "Text interview", sub: "type answers", icon: "chat" },
      { k: "voice", label: "Voice interview", sub: "live spoken conversation", icon: "mic" },
    ]
    : [
      { k: "text", label: "文本面试", sub: "打字回答", icon: "chat" },
      { k: "voice", label: "语音面试", sub: "实时语音对话", icon: "mic" },
    ];
  const onSwitchMode = (k) => {
    if (k === "voice") nav("voice");
    else nav("practice", { jobId, mode: "text" });
  };
  const toggleDictation = () => {
    if (dictating) {
      const sample = lang === "en"
        ? "I led the checkout performance rewrite. The starting point was a P75 LCP around 3.2 seconds, and the goal was to reduce it below 1.5 seconds without breaking conversion."
        : "我主导过一次结账链路性能优化。起点是 P75 LCP 大约 3.2 秒，目标是在不影响转化的前提下降到 1.5 秒以内。";
      setInput((v) => (v.trim() ? `${v.trim()}\n${sample}` : sample));
      setDictating(false);
      return;
    }
    setDictating(true);
  };

  const send = () => {
    if (!input.trim()) return;
    setTranscript((t) => [...t, { role: "user", text: input, t: fmt(elapsed) }]);
    setInput("");
    setTimeout(() => {
      setTranscript((t) => [...t, { role: "ai", text: lang === "en" ? "Interesting — could you put a number on the impact? Any latency, error rate, or revenue signals you tracked?" : "有意思——能给一个具体的量化吗？比如延迟、错误率或业务指标上的变化？", t: fmt(elapsed + 20), followUp: true }]);
    }, 700);
  };

  const roleMap = {
    general: { name: lang === "en" ? "General interviewer" : "综合面试官", tone: lang === "en" ? "Neutral · balanced" : "中性 · 综合考察" },
    hr: { name: lang === "en" ? "HR screener" : "HR 面试官", tone: lang === "en" ? "Warm · behavioral" : "友好 · 偏行为题" },
    manager: { name: lang === "en" ? "Hiring manager" : "用人经理", tone: lang === "en" ? "Direct · bar-raiser" : "直接 · 抓决策" },
  };

  return (
    <div className="ei-fadein" style={{ height: "100vh", display: "flex", flexDirection: "column", background: T.bg }}>
      {/* Top bar */}
      <div style={{ padding: "14px 28px", borderBottom: `1px solid ${T.rule}`, display: "flex", alignItems: "center", gap: 16, background: T.bgCard }}>
        <button onClick={() => nav("generating")} style={{ background: "transparent", border: "none", color: T.ink3, display: "flex", alignItems: "center", gap: 6, cursor: "pointer", fontSize: 13 }}>
          <Icon name="check" size={14} /> {lang === "en" ? "Finish & generate report" : "结束并生成报告"}
        </button>
        <div style={{ height: 18, width: 1, background: T.rule }} />
        <div>
          <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{job.company.toUpperCase()}</div>
          <div style={{ fontSize: 14, color: T.ink, fontWeight: 500 }}>{job.title}</div>
        </div>
        <div style={{ flex: 1 }} />
        <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
          <RoleDropdown T={T} role={role} setRole={setRole} roleMap={roleMap} lang={lang} />
          <Tag tone="accent" T={T}>{lang === "en" ? "Question" : "题"} {qIdx + 1}/5</Tag>
          <Tag tone="muted" T={T}><Icon name="clock" size={11} style={{ marginRight: 4 }} />{fmt(elapsed)} / 25:00</Tag>
          <button onClick={() => setPaused((p) => !p)} style={{ background: "transparent", border: `1px solid ${T.rule}`, padding: "6px 10px", borderRadius: 2, display: "flex", gap: 6, alignItems: "center", color: T.ink2, fontSize: 12 }}>
            <Icon name={paused ? "play" : "pause"} size={12} /> {paused ? (lang === "en" ? "Resume" : "继续") : (lang === "en" ? "Pause" : "暂停")}
          </button>
        </div>
      </div>

      {/* Interview modality */}
      <div style={{ padding: "8px 28px", borderBottom: `1px solid ${T.rule}`, background: T.bg, display: "flex", gap: 8, alignItems: "center" }}>
        <span className="ei-label" style={{ color: T.ink3, marginRight: 8 }}>{lang === "en" ? "INTERVIEW MODE" : "面试形式"}</span>
        {modes.map((m) => {
          const on = activeMode === m.k;
          return (
            <button key={m.k} onClick={() => onSwitchMode(m.k)} style={{
              background: on ? T.bgSoft : "transparent",
              border: `1px solid ${on ? T.rule : "transparent"}`,
              color: on ? T.ink : T.ink3, padding: "7px 11px", borderRadius: 2,
              cursor: "pointer", display: "flex", gap: 8, alignItems: "center", fontFamily: "var(--ei-sans)",
            }}>
              <Icon name={m.icon} size={13} />
              <span style={{ display: "flex", flexDirection: "column", alignItems: "flex-start", lineHeight: 1.15 }}>
                <span style={{ fontSize: 12.5, fontWeight: on ? 600 : 500 }}>{m.label}</span>
                <span style={{ fontSize: 10.5, color: on ? T.ink3 : T.ink4, marginTop: 2 }}>{m.sub}</span>
              </span>
            </button>
          );
        })}
        <div style={{ flex: 1 }} />
        <div style={{ fontSize: 11.5, color: T.ink3 }}>
          {lang === "en" ? "Choose how the interview itself runs." : "这里决定整场面试如何进行。"}
        </div>
      </div>

      {/* Main */}
      <div style={{ flex: 1, display: "grid", gridTemplateColumns: "260px 1fr 280px", minHeight: 0 }}>
        {/* left: question map */}
        <div style={{ borderRight: `1px solid ${T.rule}`, padding: "20px 18px", overflowY: "auto", background: T.bgSoft }} className="ei-scroll">
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "SESSION MAP" : "本轮题目"}</div>
          {D.questions.map((q, i) => {
            const active = i === qIdx;
            const done = i < qIdx;
            return (
              <div key={q.id} onClick={() => setQIdx(i)} style={{
                padding: "10px 12px", marginBottom: 6, borderRadius: 2, cursor: "pointer",
                background: active ? T.bgCard : "transparent",
                border: `1px solid ${active ? T.rule : "transparent"}`,
                display: "flex", gap: 10, alignItems: "flex-start",
              }}>
                <div style={{
                  width: 22, height: 22, borderRadius: 11, flexShrink: 0,
                  border: `1px solid ${active ? T.accent : T.rule}`,
                  background: done ? T.ok : active ? T.accentSoft : "transparent",
                  color: done ? "#fff" : active ? T.accent : T.ink3,
                  display: "flex", alignItems: "center", justifyContent: "center",
                  fontSize: 11, fontFamily: "var(--ei-mono)",
                }}>{done ? "✓" : i + 1}</div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 12.5, color: active ? T.ink : T.ink2, fontWeight: active ? 500 : 400 }}>
                    {q.topic}
                  </div>
                  <div style={{ fontSize: 11, color: T.ink3, marginTop: 2, fontFamily: "var(--ei-mono)" }}>
                    {q.duration}
                  </div>
                </div>
              </div>
            );
          })}
          <div style={{ borderTop: `1px dotted ${T.rule}`, marginTop: 14, paddingTop: 14 }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{lang === "en" ? "LIVE NOTES" : "实时观察"}</div>
            <div style={{ fontSize: 12, color: T.ink2, lineHeight: 1.5, padding: "8px 10px", background: T.bgCard, borderRadius: 2, border: `1px solid ${T.rule}` }}>
              <div style={{ color: T.ok }}>● {lang === "en" ? "Clear opening structure" : "开场结构清晰"}</div>
              <div style={{ color: T.warn, marginTop: 4 }}>● {lang === "en" ? "Missing quantified impact" : "缺可量化结果"}</div>
              <div style={{ color: T.ink3, marginTop: 4, fontSize: 11 }}>{lang === "en" ? "Notes are written per question, not after the session." : "每题结束即写入，不等整轮结束。"}</div>
            </div>
          </div>
        </div>

        {/* middle: transcript */}
        <div style={{ display: "flex", flexDirection: "column", minHeight: 0 }}>
          {/* current question */}
          <div style={{ padding: "28px 40px 20px", borderBottom: `1px solid ${T.rule}`, background: T.bgCard }}>
            <div style={{ display: "flex", gap: 8, marginBottom: 10, flexWrap: "wrap" }}>
              <Tag tone="accent" T={T}>{lang === "en" ? `Q${qIdx+1} · ${currentQ.topic}` : `第 ${qIdx+1} 题 · ${currentQ.topic}`}</Tag>
              {currentQ.tags.map((t) => <Tag key={t} tone="muted" T={T}>{t}</Tag>)}
            </div>
            <div className="ei-serif" style={{ fontSize: 22, color: T.ink, lineHeight: 1.35, letterSpacing: "-0.01em" }}>
              {currentQ.prompt}
            </div>
          </div>

          {/* transcript */}
          <div style={{ flex: 1, overflowY: "auto", padding: "20px 40px" }} className="ei-scroll">
            {transcript.map((m, i) => (
              <TranscriptMsg key={i} msg={m} T={T} lang={lang} />
            ))}
            <div style={{ textAlign: "center", marginTop: 12, fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
              {lang === "en" ? "— you can pause, ask for a hint, or move on —" : "— 可以暂停、请求提示，或跳过 —"}
            </div>
          </div>

          {/* input bar */}
          <div style={{ padding: "16px 40px 24px", borderTop: `1px solid ${T.rule}`, background: T.bgCard }}>
            {showHint && (
              <div style={{ marginBottom: 10, padding: "10px 12px", background: T.amberSoft, borderRadius: 2, fontSize: 13, color: T.warn }}>
                <b>{lang === "en" ? "Hint:" : "提示："}</b> {lang === "en" ? "Try STAR + numbers. Open with the baseline metric (e.g. LCP = 3.2s)." : "尝试 STAR + 数字结构。开头给一个基线指标（例如 LCP = 3.2s）。"}
              </div>
            )}
            <div style={{ border: `1px solid ${T.rule}`, borderRadius: 2, padding: 12, background: T.bg }}>
              <textarea
                value={input} onChange={(e) => setInput(e.target.value)}
                placeholder={lang === "en" ? "Type your answer here. You may also use speech-to-text." : "在这里输入回答；也可以用语音转文字填入。"}
                style={{ width: "100%", minHeight: 70, border: "none", outline: "none", resize: "none", fontSize: 14, lineHeight: 1.55, background: "transparent", color: T.ink, fontFamily: "var(--ei-sans)" }}
              />
              {dictating && (
                <div style={{ marginTop: 6, padding: "7px 9px", background: T.coolSoft, border: `1px solid ${T.cool}`, color: T.cool, fontSize: 12, borderRadius: 2, display: "flex", alignItems: "center", gap: 6 }}>
                  <span className="ei-pulse" style={{ width: 6, height: 6, borderRadius: 3, background: T.cool, display: "inline-block" }} />
                  {lang === "en" ? "Speech-to-text is listening. The transcript will be inserted into this text answer." : "语音转文字正在听写，识别结果会填入这个文本回答框。"}
                </div>
              )}
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: 6 }}>
                <div style={{ display: "flex", gap: 6 }}>
                  <button onClick={() => setShowHint((h) => !h)} style={{ background: "transparent", border: `1px solid ${T.rule}`, padding: "6px 10px", borderRadius: 2, fontSize: 12, color: T.ink2, display: "flex", gap: 4, alignItems: "center" }}>
                    <Icon name="sparkle" size={12} /> {lang === "en" ? "Hint" : "提示"}
                  </button>
                  <button onClick={toggleDictation} style={{ background: dictating ? T.coolSoft : "transparent", border: `1px solid ${dictating ? T.cool : T.rule}`, padding: "6px 10px", borderRadius: 2, fontSize: 12, color: dictating ? T.cool : T.ink2, display: "flex", gap: 4, alignItems: "center" }}>
                    <Icon name="mic" size={12} /> {dictating ? (lang === "en" ? "Insert transcript" : "插入转写") : (lang === "en" ? "Speech-to-text" : "语音转文字")}
                  </button>
                </div>
                <div style={{ display: "flex", gap: 8 }}>
                  <Btn variant="secondary" size="sm" T={T} onClick={() => setQIdx((i) => Math.min(4, i + 1))}>{lang === "en" ? "Skip →" : "跳过 →"}</Btn>
                  <Btn variant="accent" size="sm" T={T} icon="send" onClick={send}>{lang === "en" ? "Send" : "提交"}</Btn>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* right: context panel */}
        <div style={{ borderLeft: `1px solid ${T.rule}`, padding: "20px 18px", overflowY: "auto", background: T.bgSoft }} className="ei-scroll">
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "JD LINK" : "与 JD 的关联"}</div>
          <div style={{ padding: 12, background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, marginBottom: 14 }}>
            <div style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", marginBottom: 4 }}>{lang === "en" ? "THIS QUESTION PROBES" : "本题考察"}</div>
            <div style={{ fontSize: 13, color: T.ink, lineHeight: 1.55 }}>
              {qIdx === 1 ? (lang === "en" ? "Must-have · Performance optimization with measurable outcomes" : "必需项 · 性能优化 & 可量化结果") :
                qIdx === 3 ? (lang === "en" ? "Nice-to-have · Design System rollout experience" : "加分项 · Design System 落地经验") :
                (lang === "en" ? "Motivation & role fit" : "动机与岗位匹配")}
            </div>
          </div>

          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "RELEVANT EXPERIENCE" : "可调用的经历"}</div>
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            <ExpCard T={T} title={lang === "en" ? "Order re-pricing refactor" : "订单改价重构"} meta="40+ fields · 2025 Q3" hot />
            <ExpCard T={T} title={lang === "en" ? "Dashboard virtualization" : "仪表盘虚拟列表改造"} meta="2024 Q4" />
            <ExpCard T={T} title={lang === "en" ? "Component library adoption" : "组件库统一落地"} meta="2025 Q1 · partial" />
          </div>

          <div style={{ borderTop: `1px dotted ${T.rule}`, marginTop: 16, paddingTop: 14 }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "AI TRANSPARENCY" : "AI 透明度"}</div>
            <div style={{ fontSize: 11.5, color: T.ink3, lineHeight: 1.55, fontFamily: "var(--ei-mono)" }}>
              prompt v1.0.4<br/>rubric v0.9<br/>model · haiku-4.5<br/>lang · {lang}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

const TranscriptMsg = ({ msg, T, lang }) => {
  const isAI = msg.role === "ai";
  return (
    <div style={{ marginBottom: 18, display: "flex", gap: 12 }}>
      <div style={{
        width: 28, height: 28, borderRadius: 2, flexShrink: 0,
        background: isAI ? T.accentSoft : T.bgSoft,
        color: isAI ? T.accent : T.ink2,
        display: "flex", alignItems: "center", justifyContent: "center",
        fontSize: 11, fontFamily: "var(--ei-mono)", fontWeight: 500,
      }}>{isAI ? "AI" : "我"}</div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 4 }}>
          <span style={{ fontSize: 12, color: T.ink2, fontWeight: 500 }}>
            {isAI ? (lang === "en" ? "Interviewer" : "面试官") : (lang === "en" ? "You" : "我")}
          </span>
          {msg.followUp && <Tag tone="amber" T={T}>{lang === "en" ? "Follow-up" : "追问"}</Tag>}
          <span style={{ fontSize: 11, color: T.ink4, fontFamily: "var(--ei-mono)" }}>{msg.t}</span>
        </div>
        <div style={{ fontSize: 14, color: T.ink, lineHeight: 1.6 }}>{msg.text}</div>
      </div>
    </div>
  );
};

const RoleDropdown = ({ T, role, setRole, roleMap, lang }) => {
  const [open, setOpen] = React.useState(false);
  const cur = roleMap[role] || roleMap.general;
  return (
    <div style={{ position: "relative" }}>
      <button onClick={() => setOpen((o) => !o)} style={{ background: "transparent", border: `1px solid ${T.rule}`, padding: "6px 10px", borderRadius: 2, display: "flex", gap: 6, alignItems: "center", color: T.ink2, fontSize: 12 }}>
        <Icon name="briefcase" size={12} /> {cur.name}
        <Icon name="chevron_down" size={10} />
      </button>
      {open && (
        <div style={{ position: "absolute", top: "100%", right: 0, marginTop: 4, background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, minWidth: 200, zIndex: 20, boxShadow: "0 4px 16px rgba(0,0,0,0.08)" }}>
          {Object.entries(roleMap).map(([k, v]) => (
            <button key={k} onClick={() => { setRole(k); setOpen(false); }} style={{
              display: "block", width: "100%", textAlign: "left", padding: "10px 12px",
              background: role === k ? T.bgSoft : "transparent", border: "none", cursor: "pointer",
            }}>
              <div style={{ fontSize: 13, color: T.ink, fontWeight: role === k ? 500 : 400 }}>{v.name}</div>
              <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 2 }}>{v.tone}</div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

const ExpCard = ({ T, title, meta, hot }) => (
  <div style={{ padding: 10, background: T.bgCard, border: `1px solid ${hot ? T.accent : T.rule}`, borderRadius: 2 }}>
    <div style={{ display: "flex", justifyContent: "space-between", gap: 8, alignItems: "center" }}>
      <div style={{ fontSize: 12.5, color: T.ink, fontWeight: 500 }}>{title}</div>
      {hot && <Tag tone="accent" T={T}>推荐</Tag>}
    </div>
    <div style={{ fontSize: 11, color: T.ink3, marginTop: 4, fontFamily: "var(--ei-mono)" }}>{meta}</div>
  </div>
);

window.PracticeScreen = PracticeScreen;
