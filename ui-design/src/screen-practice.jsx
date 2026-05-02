// Screen 3: Mock Interview in progress
const PracticeScreen = ({ T, lang, nav, params = {}, jobId, mode, role, setRole }) => {
  const D = window.EI_DATA;
  const context = window.eiCreateInterviewContext ? window.eiCreateInterviewContext(params) : params;
  const job = D.targetJobs.find((j) => j.id === (context.targetJobId || jobId)) || D.targetJobs[0];
  const [qIdx, setQIdx] = React.useState(1);
  const [input, setInput] = React.useState("");
  const [paused, setPaused] = React.useState(false);
  const [showHint, setShowHint] = React.useState(false);
  const [dictating, setDictating] = React.useState(false);
  const [transcriptFailed, setTranscriptFailed] = React.useState(params.transcriptStatus === "failed");
  const [elapsed, setElapsed] = React.useState(502); // 08:22
  const [transcript, setTranscript] = React.useState(D.sessionTranscript);

  React.useEffect(() => {
    if (paused) return;
    const t = setInterval(() => setElapsed((e) => e + 1), 1000);
    return () => clearInterval(t);
  }, [paused]);

  const fmt = (s) => `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;
  const currentQ = D.questions[qIdx];
  const activeMode = (params.modality || mode) === "voice" ? "voice" : "text";
  const practiceMode = params.practiceMode || "strict";
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
    nav("practice", { ...context, mode: k, modality: k });
  };
  const finishAndGenerate = () => nav("generating", {
    ...context,
    mode: activeMode,
    modality: activeMode,
    practiceMode,
    hintUsed: showHint ? "true" : (params.hintUsed || "false"),
  });
  const toggleDictation = () => {
    if (dictating) {
      const sample = lang === "en"
        ? "I led the checkout performance rewrite. The starting point was a P75 LCP around 3.2 seconds, and the goal was to reduce it below 1.5 seconds without breaking conversion."
        : "我主导过一次结账链路性能优化。起点是 P75 LCP 大约 3.2 秒，目标是在不影响转化的前提下降到 1.5 秒以内。";
      setInput((v) => (v.trim() ? `${v.trim()}\n${sample}` : sample));
      setDictating(false);
      setTranscriptFailed(false);
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
        <button onClick={finishAndGenerate} style={{ background: "transparent", border: "none", color: T.ink3, display: "flex", alignItems: "center", gap: 6, cursor: "pointer", fontSize: 13 }}>
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
        {activeMode === "voice" ? (
          <div style={{ display: "flex", gap: 8, alignItems: "center", padding: "5px 10px", background: paused ? T.bgSoft : T.accentSoft, border: `1px solid ${paused ? T.rule : T.accent}`, borderRadius: 2 }}>
            <span className={paused ? "" : "ei-pulse"} style={{ width: 7, height: 7, borderRadius: 4, background: paused ? T.ink4 : T.accent, display: "inline-block" }} />
            <span style={{ fontSize: 11.5, color: paused ? T.ink3 : T.accent, fontFamily: "var(--ei-mono)" }}>
              {paused ? (lang === "en" ? "voice paused" : "语音已暂停") : (lang === "en" ? "recording · live transcript" : "录音中 · 正在实时转写")}
            </span>
          </div>
        ) : (
          <div style={{ fontSize: 11.5, color: T.ink3 }}>
            {lang === "en" ? "Choose how the interview itself runs." : "这里决定整场面试如何进行。"}
          </div>
        )}
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

        {/* middle: interview surface */}
        <div style={{ display: "flex", flexDirection: "column", minHeight: 0 }}>
          {activeMode === "voice" ? (
            <VoiceSessionSurface
              T={T}
              lang={lang}
              currentQ={currentQ}
              qIdx={qIdx}
              recording={!paused}
              transcriptFailed={transcriptFailed}
              onRetryTranscript={() => setTranscriptFailed(false)}
            />
          ) : (
            <>
              {/* current question */}
              <div style={{ padding: "28px 40px 20px", borderBottom: `1px solid ${T.rule}`, background: T.bgCard }}>
                <div style={{ display: "flex", gap: 8, marginBottom: 10, flexWrap: "wrap" }}>
                  <Tag tone="accent" T={T}>{lang === "en" ? `Q${qIdx+1} · ${currentQ.topic}` : `第 ${qIdx+1} 题 · ${currentQ.topic}`}</Tag>
                  {currentQ.tags.map((t) => <Tag key={t} tone="muted" T={T}>{t}</Tag>)}
                </div>
                <div className="ei-serif" style={{ fontSize: 22, color: T.ink, lineHeight: 1.35, letterSpacing: 0 }}>
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
                  {transcriptFailed && <VoiceTranscriptionFailure T={T} lang={lang} onRetry={() => setTranscriptFailed(false)} />}
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
            </>
          )}
        </div>

        {/* right: context panel */}
        <div style={{ borderLeft: `1px solid ${T.rule}`, padding: "20px 18px", overflowY: "auto", background: T.bgSoft }} className="ei-scroll">
          {activeMode === "voice" ? (
            <VoiceExpressionPanel T={T} lang={lang} />
          ) : (
            <>
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
            </>
          )}
        </div>
      </div>
    </div>
  );
};

const PracticeWaveformBars = ({ T, bars = 70, active = true, height = 48 }) => {
  const [tick, setTick] = React.useState(0);
  React.useEffect(() => {
    if (!active) return;
    const i = setInterval(() => setTick((t) => t + 1), 90);
    return () => clearInterval(i);
  }, [active]);
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 2, height, flex: 1, minWidth: 0 }}>
      {Array.from({ length: bars }).map((_, i) => {
        const phase = (tick + i * 3) * 0.18;
        const seed = Math.sin(i * 1.3) * 0.5 + 0.5;
        const wobble = active ? (Math.sin(phase) * 0.5 + Math.cos(phase * 1.7) * 0.35) : 0;
        const h = Math.max(3, (seed * 0.6 + 0.2 + wobble * 0.45) * height);
        const recent = i > bars - 8;
        return (
          <div key={i} style={{
            flex: 1,
            height: h,
            minWidth: 2,
            borderRadius: 1,
            background: recent && active ? T.accent : T.ink4,
            opacity: recent ? 1 : 0.55,
            transition: "height .09s ease",
          }} />
        );
      })}
    </div>
  );
};

const PracticeAnnotatedWaveform = ({ T, samples, annotations = [], width = 880, height = 72 }) => {
  const mid = height / 2;
  const step = width / samples.length;
  return (
    <svg width="100%" viewBox={`0 0 ${width} ${height + 30}`} style={{ display: "block" }}>
      <line x1="0" y1={mid} x2={width} y2={mid} stroke={T.rule} strokeWidth="1" />
      {samples.map((v, i) => {
        const h = Math.max(1, Math.abs(v) * (height / 2 - 4));
        return <rect key={i} x={i * step} y={mid - h} width={Math.max(1, step - 0.5)} height={h * 2} fill={T.ink3} opacity={v < 0 ? 0.45 : 0.85} />;
      })}
      {annotations.map((a, i) => {
        const x = a.at * width;
        if (a.kind === "pause") {
          const w = (a.dur || 0.02) * width;
          return (
            <g key={i}>
              <rect x={x} y={4} width={w} height={height - 8} fill={T.amberSoft} opacity="0.55" />
              <line x1={x} y1={4} x2={x} y2={height - 4} stroke={T.warn} strokeWidth="1" strokeDasharray="2 2" />
              <line x1={x + w} y1={4} x2={x + w} y2={height - 4} stroke={T.warn} strokeWidth="1" strokeDasharray="2 2" />
              <text x={x + w / 2} y={height + 15} textAnchor="middle" fontSize="9.5" fill={T.warn} fontFamily="var(--ei-mono)" letterSpacing="0">
                PAUSE {a.label}
              </text>
            </g>
          );
        }
        if (a.kind === "filler") {
          return (
            <g key={i}>
              <circle cx={x} cy={mid} r="4" fill={T.danger} />
              <line x1={x} y1={mid + 6} x2={x} y2={height + 3} stroke={T.danger} strokeWidth="1" />
              <text x={x} y={height + 15} textAnchor="middle" fontSize="9.5" fill={T.danger} fontFamily="var(--ei-mono)">
                {a.label}
              </text>
            </g>
          );
        }
        if (a.kind === "pace") {
          return (
            <g key={i}>
              <line x1={x} y1={0} x2={x} y2={height} stroke={T.cool} strokeWidth="1" strokeDasharray="3 3" />
              <rect x={x - 20} y={height + 3} width="40" height="14" fill={T.coolSoft} rx="2" />
              <text x={x} y={height + 14} textAnchor="middle" fontSize="9.5" fill={T.cool} fontFamily="var(--ei-mono)">
                {a.label}
              </text>
            </g>
          );
        }
        return null;
      })}
    </svg>
  );
};

const VoiceSessionSurface = ({ T, lang, currentQ, qIdx, recording, transcriptFailed, onRetryTranscript }) => {
  const samples = React.useMemo(() => {
    const out = [];
    for (let i = 0; i < 200; i++) {
      const env = Math.sin((i / 200) * Math.PI * 3) * 0.5 + 0.5;
      const signal = (Math.sin(i * 0.9) + Math.sin(i * 0.3) * 0.7 + Math.sin(i * 2.1) * 0.18) * 0.5;
      let v = signal * (0.3 + env * 0.7);
      if (i > 58 && i < 70) v *= 0.05;
      if (i > 120 && i < 138) v *= 0.08;
      out.push(v);
    }
    return out;
  }, []);
  const annotations = [
    { at: 0.30, kind: "pause", dur: 0.06, label: "0.8s" },
    { at: 0.60, kind: "pause", dur: 0.09, label: lang === "en" ? "1.6s · long" : "1.6s · 偏长" },
    { at: 0.44, kind: "filler", label: lang === "en" ? "um..." : "嗯..." },
    { at: 0.78, kind: "filler", label: lang === "en" ? "basically" : "就是" },
    { at: 0.15, kind: "pace", label: lang === "en" ? "steady" : "正常" },
    { at: 0.85, kind: "pace", label: lang === "en" ? "fast" : "偏快" },
  ];
  const transcript = lang === "en" ? [
    { t: "00:02:14", role: "ai", text: "Take your time. When you're ready, start with the situation." },
    { t: "00:02:18", role: "user", text: "OK. The project was our order repricing system..." },
    { t: "00:02:27", role: "user", text: "We needed to support 40+ fields, and the designer wanted to keep everything on one screen." },
    { t: "00:02:41", role: "user", text: "Um... I was worried about cognitive load, basically..." },
    { t: "00:02:44", role: "user", text: "Operators handled more than 300 orders a day, so scanning 40 fields on every order...", pause: "1.2s" },
    { t: "00:02:49", role: "note", text: "Long pause detected · 1.6s" },
  ] : [
    { t: "00:02:14", role: "ai", text: "别急，你准备好了就从当时的情境开始。" },
    { t: "00:02:18", role: "user", text: "好。那个项目是我们的订单改价系统——" },
    { t: "00:02:27", role: "user", text: "要支持 40 多个字段，当时设计想把所有字段都放在一屏上。" },
    { t: "00:02:41", role: "user", text: "嗯... 我当时担心的其实是认知负荷，就是..." },
    { t: "00:02:44", role: "user", text: "操作员一天要处理 300 多单，如果每单都要在一屏上扫 40 个字段...", pause: "1.2s" },
    { t: "00:02:49", role: "note", text: "检测到长停顿 · 1.6 秒" },
  ];

  return (
    <>
      <div style={{ padding: "24px 34px 18px", borderBottom: `1px solid ${T.rule}`, background: T.bgCard }}>
        <div style={{ display: "flex", gap: 8, marginBottom: 10, flexWrap: "wrap", alignItems: "center" }}>
          <Tag tone="accent" T={T}>{lang === "en" ? `Q${qIdx+1} · ${currentQ.topic}` : `第 ${qIdx+1} 题 · ${currentQ.topic}`}</Tag>
          {currentQ.tags.map((t) => <Tag key={t} tone="muted" T={T}>{t}</Tag>)}
        </div>
        <div className="ei-serif" style={{ fontSize: 22, color: T.ink, lineHeight: 1.35, letterSpacing: 0 }}>
          {currentQ.prompt}
        </div>
      </div>

      <div style={{ flex: 1, overflowY: "auto", padding: "22px 34px", display: "flex", flexDirection: "column", gap: 20 }} className="ei-scroll">
        <div>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 10 }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "NOW SPEAKING · Lin Zhou" : "正在说话 · 林舟"}</div>
            <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink3 }}>-12 dB</div>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 14, padding: "16px 18px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3 }}>
            <div style={{ width: 38, height: 38, borderRadius: 19, background: T.accent, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
              <Icon name="mic" size={18} />
            </div>
            <PracticeWaveformBars T={T} bars={70} active={recording} height={48} />
          </div>
        </div>

        <div>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 10 }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "CURRENT ANSWER · annotated" : "本次回答 · 已标注"}</div>
            <div style={{ display: "flex", gap: 12, fontFamily: "var(--ei-mono)", fontSize: 10.5, flexWrap: "wrap", justifyContent: "flex-end" }}>
              <span style={{ color: T.warn, display: "flex", gap: 4, alignItems: "center" }}>
                <span style={{ width: 8, height: 8, background: T.amberSoft, border: `1px solid ${T.warn}` }} />{lang === "en" ? "pause" : "停顿"}
              </span>
              <span style={{ color: T.danger, display: "flex", gap: 4, alignItems: "center" }}>
                <span style={{ width: 8, height: 8, borderRadius: 4, background: T.danger }} />{lang === "en" ? "filler" : "口头禅"}
              </span>
              <span style={{ color: T.cool, display: "flex", gap: 4, alignItems: "center" }}>
                <span style={{ width: 8, height: 8, background: T.coolSoft, border: `1px solid ${T.cool}` }} />{lang === "en" ? "pace" : "语速"}
              </span>
            </div>
          </div>
          <div style={{ padding: "16px 18px 4px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3 }}>
            <PracticeAnnotatedWaveform T={T} samples={samples} annotations={annotations} />
          </div>
        </div>

        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "Live transcript" : "实时转写"}</div>
          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {transcript.map((m, i) => {
              if (m.role === "note") return (
                <div key={i} style={{ display: "flex", gap: 10, padding: "6px 12px", background: T.warnSoft, borderLeft: `2px solid ${T.warn}`, fontSize: 12, color: T.warn, fontFamily: "var(--ei-mono)" }}>
                  <Icon name="info" size={13} /> {m.text}
                </div>
              );
              const isAI = m.role === "ai";
              return (
                <div key={i} style={{ display: "flex", gap: 12 }}>
                  <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink4, width: 58, flexShrink: 0, paddingTop: 2 }}>{m.t}</div>
                  <div style={{ fontSize: 14, lineHeight: 1.55, color: isAI ? T.ink3 : T.ink, fontStyle: isAI ? "italic" : "normal", flex: 1 }}>
                    {m.text}
                    {m.pause && <span style={{ display: "inline-block", marginLeft: 8, padding: "1px 6px", background: T.amberSoft, color: T.warn, fontSize: 11, fontFamily: "var(--ei-mono)", borderRadius: 2 }}>II {m.pause}</span>}
                  </div>
                </div>
              );
            })}
            <div style={{ display: "flex", gap: 12 }}>
              <div style={{ width: 58, flexShrink: 0 }} />
              <div style={{ width: 8, height: 16, background: recording ? T.accent : T.ink4 }} className={recording ? "ei-pulse" : ""} />
            </div>
          </div>
          {transcriptFailed && <VoiceTranscriptionFailure T={T} lang={lang} onRetry={onRetryTranscript} />}
        </div>
      </div>
    </>
  );
};

const VoiceExpressionPanel = ({ T, lang }) => {
  const metrics = [
    { k: lang === "en" ? "Words / min" : "语速", v: "186", hint: lang === "en" ? "steady 160-200 wpm" : "稳定在 160-200 字/分", tone: "ok", bar: 0.7 },
    { k: lang === "en" ? "Long pauses" : "长停顿", v: "2", hint: lang === "en" ? "2 over 1.5s" : "本题 2 次超过 1.5 秒", tone: "warn", bar: 0.5 },
    { k: lang === "en" ? "Fillers" : "口头禅", v: "4", hint: "「嗯」×2 · 「就是」×2", tone: "danger", bar: 0.6 },
    { k: lang === "en" ? "Volume" : "音量", v: lang === "en" ? "stable" : "稳定", hint: lang === "en" ? "no drop-offs" : "没有明显衰减", tone: "ok", bar: 0.78 },
  ];
  return (
    <>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "Expression metrics" : "表达层指标"}</div>
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        {metrics.map((m) => {
          const toneColor = { ok: T.ok, warn: T.warn, danger: T.danger }[m.tone] || T.ink2;
          return (
            <div key={m.k}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", marginBottom: 5 }}>
                <div style={{ fontSize: 13, color: T.ink2 }}>{m.k}</div>
                <div style={{ fontFamily: "var(--ei-mono)", fontSize: 15, color: toneColor, fontWeight: 500 }}>{m.v}</div>
              </div>
              <div style={{ height: 3, background: T.rule, borderRadius: 2, overflow: "hidden" }}>
                <div style={{ width: `${m.bar * 100}%`, height: "100%", background: toneColor }} />
              </div>
              <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 4, fontFamily: "var(--ei-mono)" }}>{m.hint}</div>
            </div>
          );
        })}
      </div>

      <div style={{ marginTop: 26, padding: 14, background: T.bgCard, border: `1px dotted ${T.rule}`, borderRadius: 3 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "GENTLE NUDGE" : "现场提示"}</div>
        <div style={{ fontSize: 13, color: T.ink, lineHeight: 1.55 }}>
          <Icon name="sparkle" size={13} color={T.accent} style={{ marginRight: 6 }} />
          {lang === "en"
            ? "You're midway through the situation. Name the concrete action you took before moving to the response."
            : "当前在讲「情境」。试着先说一句你采取的具体行动，再切到对方的反应。"}
        </div>
      </div>

      <div style={{ borderTop: `1px dotted ${T.rule}`, marginTop: 16, paddingTop: 14 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "AI TRANSPARENCY" : "AI 透明度"}</div>
        <div style={{ fontSize: 11.5, color: T.ink3, lineHeight: 1.55, fontFamily: "var(--ei-mono)" }}>
          prompt v1.0.4<br/>rubric v0.9<br/>model · haiku-4.5<br/>lang · {lang}
        </div>
      </div>

      <div style={{ marginTop: 16, fontSize: 11, color: T.ink4, lineHeight: 1.5, fontFamily: "var(--ei-mono)" }}>
        {lang === "en"
          ? "audio stays on-device during the session · deleted after report"
          : "音频仅在本次会话缓存 · 报告生成后自动删除"}
      </div>
    </>
  );
};

const VoiceTranscriptionFailure = ({ T, lang, onRetry }) => (
  <div style={{ marginTop: 8, padding: "8px 10px", background: T.dangerSoft, border: `1px solid ${T.danger}`, color: T.danger, fontSize: 12, borderRadius: 2, display: "flex", alignItems: "center", justifyContent: "space-between", gap: 10 }}>
    <span>{lang === "en" ? "Speech-to-text failed. Your typed answer is preserved; retry or keep typing." : "语音转文字失败。已输入内容会保留，可以重试或继续手动输入。"}</span>
    <button onClick={onRetry} style={{ background: "transparent", border: `1px solid ${T.danger}`, color: T.danger, borderRadius: 2, padding: "3px 8px", fontSize: 11 }}>
      {lang === "en" ? "Retry" : "重试"}
    </button>
  </div>
);

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
