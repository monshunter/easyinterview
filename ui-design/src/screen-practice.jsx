// Screen 3: Interview session in progress
const PracticeScreen = ({ T, lang, nav, params = {}, jobId, mode }) => {
  const D = window.EI_DATA;
  const context = window.eiCreateInterviewContext ? window.eiCreateInterviewContext(params) : params;
  const job = D.targetJobs.find((j) => j.id === (context.targetJobId || jobId)) || D.targetJobs[0];
  const [qIdx, setQIdx] = React.useState(1);
  const [input, setInput] = React.useState("");
  const [paused, setPaused] = React.useState(false);
  const [showHint, setShowHint] = React.useState(false);
  const [hintCount, setHintCount] = React.useState(0);
  const [elapsed, setElapsed] = React.useState(502);
  const [transcript, setTranscript] = React.useState(D.sessionTranscript);
  const [captionsShown, setCaptionsShown] = React.useState(false);

  React.useEffect(() => {
    if (paused) return;
    const t = setInterval(() => setElapsed((e) => e + 1), 1000);
    return () => clearInterval(t);
  }, [paused]);

  const fmt = (s) => `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;
  const currentQ = D.questions[qIdx];
  const requestedMode = params.modality || params.mode || mode;
  const activeMode = requestedMode === "phone" ? "phone" : "text";
  const isPhone = activeMode === "phone";
  const interviewerRole = context.roundName || (lang === "en" ? "Manager round" : "经理面");
  const interviewerLabel = lang === "en" ? `${interviewerRole} interviewer` : `${interviewerRole}面试官`;
  const enterPhoneMode = () => {
    setCaptionsShown(false);
    nav("practice", { ...context, mode: "phone", modality: "phone" });
  };
  const exitPhoneMode = () => {
    setCaptionsShown(false);
    nav("practice", { ...context, mode: "text", modality: "text" });
  };
  const finishAndGenerate = () => nav("generating", {
    ...context,
    mode: activeMode,
    modality: activeMode,
    hintUsed: hintCount > 0 ? "true" : "false",
    hintCount: String(hintCount),
  });
  const send = () => {
    if (!input.trim()) return;
    setTranscript((t) => [...t, { role: "user", text: input, t: fmt(elapsed) }]);
    setInput("");
    setTimeout(() => {
      setTranscript((t) => [...t, { role: "ai", text: lang === "en" ? "Interesting. Could you add one concrete impact number before we move on?" : "有意思。进入下一问前，你能补一个具体影响数字吗？", t: fmt(elapsed + 20), followUp: true }]);
    }, 700);
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
          <Tag tone="accent" T={T}>{lang === "en" ? "Question" : "题"} {qIdx + 1}/5</Tag>
          <Tag tone="muted" T={T}><Icon name="clock" size={11} style={{ marginRight: 4 }} />{fmt(elapsed)} / 25:00</Tag>
          <button onClick={() => setPaused((p) => !p)} style={{ background: "transparent", border: `1px solid ${T.rule}`, padding: "6px 10px", borderRadius: 2, display: "flex", gap: 6, alignItems: "center", color: T.ink2, fontSize: 12 }}>
            <Icon name={paused ? "play" : "pause"} size={12} /> {paused ? (lang === "en" ? "Resume" : "继续") : (lang === "en" ? "Pause" : "暂停")}
          </button>
          <div style={{ height: 18, width: 1, background: T.rule }} />
          <button
            type="button"
            data-testid="practice-topbar-phone-toggle"
            aria-pressed={isPhone}
            aria-label={isPhone
              ? (lang === "en" ? "Hang up and return to text mode" : "挂断并返回文本模式")
              : (lang === "en" ? "Enter phone mode" : "进入电话模式")}
            title={isPhone
              ? (lang === "en" ? "Return to text" : "返回文本模式")
              : (lang === "en" ? "Phone mode" : "电话模式")}
            onClick={isPhone ? exitPhoneMode : enterPhoneMode}
            style={{
              width: 34, height: 34, padding: 0, borderRadius: 17,
              border: `1px solid ${isPhone ? T.accent : T.rule}`,
              background: isPhone ? T.accentSoft : "transparent",
              color: isPhone ? T.accent : T.ink2,
              display: "inline-flex", alignItems: "center", justifyContent: "center",
            }}
          >
            <Icon name="phone" size={15} />
          </button>
          <button onClick={finishAndGenerate} style={{
            padding: "7px 12px",
            background: T.accent, color: "#fff",
            border: "none", borderRadius: 2, cursor: "pointer",
            fontSize: 12.5, fontWeight: 500, fontFamily: "var(--ei-sans)",
            display: "flex", alignItems: "center", justifyContent: "center", gap: 6,
          }}>
            <Icon name="check" size={13} />
            {lang === "en" ? "Finish report" : "结束并生成报告"}
          </button>
        </div>
      </div>

      <div style={{ flex: 1, display: "grid", gridTemplateColumns: "260px minmax(0, 1fr)", minHeight: 0 }}>
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
          {hintCount > 0 && (
            <div style={{ borderTop: `1px dotted ${T.rule}`, marginTop: 14, paddingTop: 14 }}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{lang === "en" ? "SESSION NOTE" : "会话标记"}</div>
              <div style={{ fontSize: 12, color: T.ink2, lineHeight: 1.5, padding: "8px 10px", background: T.bgCard, borderRadius: 2, border: `1px solid ${T.rule}` }}>
                {lang === "en" ? `${hintCount} hint${hintCount > 1 ? "s" : ""} used in this session.` : `本场已使用 ${hintCount} 次提示。`}
              </div>
            </div>
          )}
        </div>

        <div style={{ display: "flex", flexDirection: "column", minHeight: 0 }}>
          {isPhone ? (
            <PhoneSessionSurface
              T={T}
              lang={lang}
              active={!paused}
              captionsShown={captionsShown}
              onToggleCaptions={() => setCaptionsShown((v) => !v)}
              onHangUp={exitPhoneMode}
              transcript={transcript}
            />
          ) : (
            <>
              <QuestionHeader T={T} lang={lang} currentQ={currentQ} qIdx={qIdx} />
              <TranscriptPane T={T} lang={lang} transcript={transcript} />
              <div style={{ padding: "16px 40px 24px", borderTop: `1px solid ${T.rule}`, background: T.bgCard }}>
                {showHint && (
                  <div style={{ marginBottom: 10, padding: "10px 12px", background: T.amberSoft, borderRadius: 2, fontSize: 13, color: T.warn }}>
                    <b>{lang === "en" ? "Hint:" : "提示："}</b> {lang === "en" ? "Try STAR + numbers. Open with the baseline metric, then state your action." : "尝试 STAR + 数字结构。开头给基线指标，再说你的具体行动。"}
                  </div>
                )}
                <div style={{ border: `1px solid ${T.rule}`, borderRadius: 2, padding: 12, background: T.bg }}>
                  <textarea
                    value={input} onChange={(e) => setInput(e.target.value)}
                    placeholder={lang === "en" ? "Type your answer here." : "在这里输入回答。"}
                    style={{ width: "100%", minHeight: 74, border: "none", outline: "none", resize: "none", fontSize: 14, lineHeight: 1.55, background: "transparent", color: T.ink, fontFamily: "var(--ei-sans)" }}
                  />
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: 6 }}>
                    <button onClick={() => { if (!showHint) setHintCount((c) => c + 1); setShowHint((h) => !h); }} style={{ background: showHint ? T.amberSoft : "transparent", border: `1px solid ${showHint ? T.warn : T.rule}`, padding: "6px 10px", borderRadius: 2, fontSize: 12, color: showHint ? T.warn : T.ink2, display: "flex", gap: 4, alignItems: "center" }}>
                      <Icon name="sparkle" size={12} /> {lang === "en" ? "Hint" : "提示"}
                    </button>
                    <Btn variant="accent" size="sm" T={T} icon="send" onClick={send}>{lang === "en" ? "Send" : "提交"}</Btn>
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

const QuestionHeader = ({ T, lang, currentQ, qIdx }) => (
  <div style={{ padding: "28px 40px 20px", borderBottom: `1px solid ${T.rule}`, background: T.bgCard }}>
    <div style={{ display: "flex", gap: 8, marginBottom: 10, flexWrap: "wrap" }}>
      <Tag tone="accent" T={T}>{lang === "en" ? `Q${qIdx+1} · ${currentQ.topic}` : `第 ${qIdx+1} 题 · ${currentQ.topic}`}</Tag>
      {currentQ.tags.map((t) => <Tag key={t} tone="muted" T={T}>{t}</Tag>)}
    </div>
    <div className="ei-serif" style={{ fontSize: 22, color: T.ink, lineHeight: 1.35, letterSpacing: 0 }}>
      {currentQ.prompt}
    </div>
  </div>
);

const TranscriptPane = ({ T, lang, transcript, compact = false }) => (
  <div style={{ flex: 1, overflowY: "auto", padding: compact ? "14px 22px" : "20px 40px" }} className="ei-scroll">
    {transcript.map((m, i) => (
      <TranscriptMsg key={i} msg={m} T={T} lang={lang} />
    ))}
    {!compact && (
      <div style={{ textAlign: "center", marginTop: 12, fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
        {lang === "en" ? "— you can pause, ask for a hint, or finish the interview —" : "— 可以暂停、请求提示，或结束面试 —"}
      </div>
    )}
  </div>
);

const PracticeWaveformBars = ({ T, bars = 70, active = true, height = 58 }) => {
  const [tick, setTick] = React.useState(0);
  React.useEffect(() => {
    if (!active) return;
    const i = setInterval(() => setTick((t) => t + 1), 90);
    return () => clearInterval(i);
  }, [active]);
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 3, height, flex: 1, minWidth: 0 }}>
      {Array.from({ length: bars }).map((_, i) => {
        const phase = (tick + i * 3) * 0.18;
        const seed = Math.sin(i * 1.3) * 0.5 + 0.5;
        const wobble = active ? (Math.sin(phase) * 0.5 + Math.cos(phase * 1.7) * 0.35) : 0;
        const h = Math.max(5, (seed * 0.55 + 0.24 + wobble * 0.38) * height);
        const recent = i > bars - 10;
        return (
          <div key={i} style={{
            flex: 1,
            height: h,
            minWidth: 2,
            borderRadius: 1,
            background: recent && active ? T.accent : T.ink4,
            opacity: recent ? 1 : 0.5,
            transition: "height .09s ease",
          }} />
        );
      })}
    </div>
  );
};

const PhoneSessionSurface = ({ T, lang, active, captionsShown, onToggleCaptions, onHangUp, transcript }) => {
  return (
    <>
      <div data-testid="practice-phone-surface" style={{ flex: 1, minHeight: 0, display: "flex", flexDirection: "column", background: T.bg }}>
        <div style={{ flex: captionsShown ? "0 0 auto" : 1, padding: "46px 56px 28px", display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", gap: 22 }}>
          <div style={{
            width: 96, height: 96, borderRadius: 48,
            background: T.accentSoft,
            border: `1px solid ${T.accent}`,
            display: "flex", alignItems: "center", justifyContent: "center",
            color: T.accent,
          }}>
            <Icon name="mic" size={34} stroke={1.7} />
          </div>
          <div data-testid="practice-phone-call-state" style={{ textAlign: "center" }}>
            <div className="ei-serif" style={{ fontSize: 28, color: T.ink, letterSpacing: 0 }}>
              {lang === "en" ? "Phone interview in progress" : "电话模式进行中"}
            </div>
            <div style={{ marginTop: 8, fontSize: 13, color: T.ink3, lineHeight: 1.5 }}>
              {lang === "en" ? "Listen and answer naturally. Captions are optional." : "像真实电话一样听题并回答；字幕可按需显示。"}
            </div>
          </div>
          <div data-testid="practice-phone-waveform" style={{ display: "flex", alignItems: "center", gap: 16, width: "min(720px, 100%)", padding: "18px 22px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4 }}>
            <div className={active ? "ei-pulse" : ""} style={{ width: 34, height: 34, borderRadius: 17, background: active ? T.accent : T.ink4, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
              <Icon name={active ? "mic" : "pause"} size={15} />
            </div>
            <PracticeWaveformBars T={T} bars={66} active={active} height={58} />
          </div>
          <div style={{ display: "flex", gap: 10, flexWrap: "wrap", justifyContent: "center" }}>
            <button type="button" data-testid="practice-phone-captions-toggle" onClick={onToggleCaptions} style={{ background: captionsShown ? T.accentSoft : "transparent", border: `1px solid ${captionsShown ? T.accent : T.rule}`, color: captionsShown ? T.accent : T.ink2, padding: "8px 12px", borderRadius: 2, fontSize: 12.5, display: "flex", alignItems: "center", gap: 6 }}>
              <Icon name="chat" size={13} /> {captionsShown ? (lang === "en" ? "Hide captions" : "隐藏字幕") : (lang === "en" ? "Show captions" : "显示字幕")}
            </button>
            <button
              type="button"
              data-testid="practice-phone-hangup"
              aria-label={lang === "en" ? "Hang up and return to text mode" : "挂断并返回文本模式"}
              title={lang === "en" ? "Hang up" : "挂断"}
              onClick={onHangUp}
              style={{ width: 56, height: 56, padding: 0, background: T.danger, border: "none", color: "#fff", borderRadius: 28, display: "inline-flex", alignItems: "center", justifyContent: "center", boxShadow: "0 6px 18px rgba(179,64,43,0.24)" }}
            >
              <Icon name="phone" size={22} stroke={1.8} style={{ transform: "rotate(135deg)" }} />
            </button>
          </div>
        </div>
        {captionsShown && (
          <div data-testid="practice-phone-captions" style={{ borderTop: `1px solid ${T.rule}`, background: T.bgCard, minHeight: 220, maxHeight: "42vh", display: "flex", flexDirection: "column" }}>
            <div style={{ padding: "12px 22px 0", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "CAPTIONS" : "字幕"}</div>
              <Tag tone="muted" T={T}>{lang === "en" ? "same session transcript" : "同一会话记录"}</Tag>
            </div>
            <TranscriptPane T={T} lang={lang} transcript={transcript} compact />
          </div>
        )}
      </div>
    </>
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

window.PracticeScreen = PracticeScreen;
