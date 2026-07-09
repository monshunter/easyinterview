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
  const [callState, setCallState] = React.useState("connected");

  React.useEffect(() => {
    if (paused) return;
    const t = setInterval(() => setElapsed((e) => e + 1), 1000);
    return () => clearInterval(t);
  }, [paused]);

  const fmt = (s) => `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;
  const currentQ = D.questions[qIdx];
  const requestedMode = params.modality || params.mode || mode;
  const activeMode = requestedMode === "phone" || requestedMode === "voice" ? "phone" : "text";
  const isPhone = activeMode === "phone";
  const interviewerRole = context.roundName || (lang === "en" ? "Manager round" : "经理面");
  const interviewerLabel = lang === "en" ? `${interviewerRole} interviewer` : `${interviewerRole}面试官`;
  const modes = lang === "en"
    ? [
      { k: "text", label: "Text", icon: "chat" },
      { k: "phone", label: "Phone", icon: "mic" },
    ]
    : [
      { k: "text", label: "文本", icon: "chat" },
      { k: "phone", label: "电话模式", icon: "mic" },
    ];
  const onSwitchMode = (k) => {
    setCallState(k === "phone" ? "connected" : callState);
    nav("practice", { ...context, mode: k, modality: k });
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
          <div style={{ display: "flex", background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 2, gap: 2 }}>
            {modes.map((m) => {
              const on = activeMode === m.k;
              return (
                <button key={m.k} onClick={() => onSwitchMode(m.k)} style={{
                  background: on ? T.bgCard : "transparent",
                  border: `1px solid ${on ? T.rule : "transparent"}`,
                  boxShadow: on ? "0 1px 2px rgba(0,0,0,0.06)" : "none",
                  color: on ? T.ink : T.ink3,
                  padding: "4px 9px",
                  borderRadius: 2,
                  cursor: "pointer",
                  display: "flex", gap: 5, alignItems: "center",
                  fontSize: 12, fontWeight: on ? 500 : 400,
                  fontFamily: "var(--ei-sans)",
                }}>
                  <Icon name={m.icon} size={12} />
                  {m.label}
                </button>
              );
            })}
          </div>
          {isPhone && (
            <div style={{ display: "flex", gap: 5, alignItems: "center", padding: "4px 8px", background: paused || callState === "ended" ? T.bgSoft : T.accentSoft, border: `1px solid ${paused || callState === "ended" ? T.rule : T.accent}`, borderRadius: 2 }}>
              <span className={!paused && callState === "connected" ? "ei-pulse" : ""} style={{ width: 6, height: 6, borderRadius: 3, background: paused || callState === "ended" ? T.ink4 : T.accent, display: "inline-block" }} />
              <span style={{ fontSize: 11, color: paused || callState === "ended" ? T.ink3 : T.accent, fontFamily: "var(--ei-mono)" }}>
                {callState === "ended" ? (lang === "en" ? "ended" : "已切断") : paused ? (lang === "en" ? "paused" : "已暂停") : (lang === "en" ? "live" : "通话中")}
              </span>
            </div>
          )}
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
              active={!paused && callState === "connected"}
              callState={callState}
              captionsShown={captionsShown}
              onToggleCaptions={() => setCaptionsShown((v) => !v)}
              onHangUp={() => setCallState("ended")}
              onRestart={() => { setCallState("connected"); setCaptionsShown(false); }}
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

const PhoneSessionSurface = ({ T, lang, active, callState, captionsShown, onToggleCaptions, onHangUp, onRestart, transcript }) => {
  const ended = callState === "ended";
  return (
    <>
      <div style={{ flex: 1, minHeight: 0, display: "flex", flexDirection: "column", background: T.bg }}>
        <div style={{ flex: captionsShown ? "0 0 auto" : 1, padding: "46px 56px 28px", display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", gap: 22 }}>
          <div style={{
            width: 96, height: 96, borderRadius: 48,
            background: ended ? T.bgSoft : T.accentSoft,
            border: `1px solid ${ended ? T.rule : T.accent}`,
            display: "flex", alignItems: "center", justifyContent: "center",
            color: ended ? T.ink3 : T.accent,
          }}>
            <Icon name={ended ? "pause" : "mic"} size={34} stroke={1.7} />
          </div>
          <div style={{ textAlign: "center" }}>
            <div className="ei-serif" style={{ fontSize: 28, color: T.ink, letterSpacing: 0 }}>
              {ended ? (lang === "en" ? "Call ended" : "通话已切断") : (lang === "en" ? "Phone interview in progress" : "电话模式进行中")}
            </div>
            <div style={{ marginTop: 8, fontSize: 13, color: T.ink3, lineHeight: 1.5 }}>
              {ended
                ? (lang === "en" ? "Restart when you are ready to continue this round." : "准备好后可重新开始本轮通话。")
                : (lang === "en" ? "Listen and answer naturally. Captions are optional." : "像真实电话一样听题并回答；字幕可按需显示。")}
            </div>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 16, width: "min(720px, 100%)", padding: "18px 22px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4 }}>
            <div className={!ended && active ? "ei-pulse" : ""} style={{ width: 34, height: 34, borderRadius: 17, background: ended ? T.ink4 : T.accent, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
              <Icon name={ended ? "pause" : "mic"} size={15} />
            </div>
            <PracticeWaveformBars T={T} bars={66} active={!ended && active} height={58} />
          </div>
          <div style={{ display: "flex", gap: 10, flexWrap: "wrap", justifyContent: "center" }}>
            <button onClick={onToggleCaptions} style={{ background: captionsShown ? T.accentSoft : "transparent", border: `1px solid ${captionsShown ? T.accent : T.rule}`, color: captionsShown ? T.accent : T.ink2, padding: "8px 12px", borderRadius: 2, fontSize: 12.5, display: "flex", alignItems: "center", gap: 6 }}>
              <Icon name="chat" size={13} /> {captionsShown ? (lang === "en" ? "Hide captions" : "隐藏字幕") : (lang === "en" ? "Show captions" : "显示字幕")}
            </button>
            <button onClick={onHangUp} disabled={ended} style={{ background: "transparent", border: `1px solid ${ended ? T.rule : T.danger}`, color: ended ? T.ink4 : T.danger, padding: "8px 12px", borderRadius: 2, fontSize: 12.5, display: "flex", alignItems: "center", gap: 6, cursor: ended ? "default" : "pointer" }}>
              <Icon name="x" size={13} /> {lang === "en" ? "Hang up" : "切断"}
            </button>
            <button onClick={onRestart} style={{ background: "transparent", border: `1px solid ${T.rule}`, color: T.ink2, padding: "8px 12px", borderRadius: 2, fontSize: 12.5, display: "flex", alignItems: "center", gap: 6 }}>
              <Icon name="replay" size={13} /> {lang === "en" ? "Restart" : "重新开始"}
            </button>
          </div>
        </div>
        {captionsShown && (
          <div style={{ borderTop: `1px solid ${T.rule}`, background: T.bgCard, minHeight: 220, maxHeight: "42vh", display: "flex", flexDirection: "column" }}>
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
