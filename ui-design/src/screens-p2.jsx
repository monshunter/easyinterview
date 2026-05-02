// P2 screens: Voice practice in-progress

// ───────────── Shared: animated waveform ─────────────
const WaveformBars = ({ T, bars = 64, active = true, height = 44, accent }) => {
  const [tick, setTick] = React.useState(0);
  React.useEffect(() => {
    if (!active) return;
    const i = setInterval(() => setTick((t) => t + 1), 90);
    return () => clearInterval(i);
  }, [active]);
  const color = accent || T.accent;
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 2, height, flex: 1 }}>
      {Array.from({ length: bars }).map((_, i) => {
        const phase = (tick + i * 3) * 0.18;
        const seed = Math.sin(i * 1.3) * 0.5 + 0.5;
        const wobble = active ? (Math.sin(phase) * 0.5 + Math.cos(phase * 1.7) * 0.35) : 0;
        const h = Math.max(3, (seed * 0.6 + 0.2 + wobble * 0.45) * height);
        const isRecent = i > bars - 8;
        return (
          <div key={i} style={{
            flex: 1, height: h, minWidth: 2, borderRadius: 1,
            background: isRecent && active ? color : T.ink4,
            opacity: isRecent ? 1 : 0.55,
            transition: "height .09s ease",
          }} />
        );
      })}
    </div>
  );
};

// Static waveform with annotations (pause markers, pace shifts, fillers)
const AnnotatedWaveform = ({ T, samples, annotations = [], width = 880, height = 72 }) => {
  const mid = height / 2;
  const step = width / samples.length;
  return (
    <svg width="100%" viewBox={`0 0 ${width} ${height + 28}`} style={{ display: "block" }}>
      {/* baseline */}
      <line x1="0" y1={mid} x2={width} y2={mid} stroke={T.rule} strokeWidth="1" />
      {/* waveform bars */}
      {samples.map((v, i) => {
        const h = Math.max(1, Math.abs(v) * (height / 2 - 4));
        return <rect key={i} x={i * step} y={mid - h} width={Math.max(1, step - 0.5)} height={h * 2} fill={T.ink3} opacity={v < 0 ? 0.45 : 0.85} />;
      })}
      {/* annotations */}
      {annotations.map((a, i) => {
        const x = a.at * width;
        if (a.kind === "pause") {
          const w = (a.dur || 0.02) * width;
          return (
            <g key={i}>
              <rect x={x} y={4} width={w} height={height - 8} fill={T.amberSoft} opacity="0.55" />
              <line x1={x} y1={4} x2={x} y2={height - 4} stroke={T.warn} strokeWidth="1" strokeDasharray="2 2" />
              <line x1={x + w} y1={4} x2={x + w} y2={height - 4} stroke={T.warn} strokeWidth="1" strokeDasharray="2 2" />
              <text x={x + w / 2} y={height + 14} textAnchor="middle" fontSize="9.5" fill={T.warn} fontFamily="var(--ei-mono)" letterSpacing="0.08em">
                PAUSE {a.label}
              </text>
            </g>
          );
        }
        if (a.kind === "filler") {
          return (
            <g key={i}>
              <circle cx={x} cy={mid} r="4" fill={T.danger} />
              <line x1={x} y1={mid + 6} x2={x} y2={height + 2} stroke={T.danger} strokeWidth="1" />
              <text x={x} y={height + 14} textAnchor="middle" fontSize="9.5" fill={T.danger} fontFamily="var(--ei-mono)">
                {a.label}
              </text>
            </g>
          );
        }
        if (a.kind === "pace") {
          return (
            <g key={i}>
              <line x1={x} y1={0} x2={x} y2={height} stroke={T.cool} strokeWidth="1" strokeDasharray="3 3" />
              <rect x={x - 20} y={height + 2} width="40" height="14" fill={T.coolSoft} rx="2" />
              <text x={x} y={height + 13} textAnchor="middle" fontSize="9.5" fill={T.cool} fontFamily="var(--ei-mono)">
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

// ───────────── VOICE PRACTICE SCREEN ─────────────
const VoicePracticeScreen = ({ T, lang, nav }) => {
  const [elapsed, setElapsed] = React.useState(147); // seconds
  const [recording, setRecording] = React.useState(true);
  const [lastUtterance, setLastUtterance] = React.useState(0);

  React.useEffect(() => {
    if (!recording) return;
    const i = setInterval(() => setElapsed((e) => e + 1), 1000);
    return () => clearInterval(i);
  }, [recording]);

  const fmt = (s) => `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;

  // Fake sample waveform — 200 samples
  const samples = React.useMemo(() => {
    const out = [];
    for (let i = 0; i < 200; i++) {
      const env = Math.sin((i / 200) * Math.PI * 3) * 0.5 + 0.5;
      const n = (Math.sin(i * 0.9) + Math.sin(i * 0.3) * 0.7 + Math.random() * 0.3) * 0.5;
      // pause zones
      let v = n * (0.3 + env * 0.7);
      if (i > 58 && i < 70) v *= 0.05; // pause 1
      if (i > 120 && i < 138) v *= 0.08; // long pause
      out.push(v);
    }
    return out;
  }, []);

  const annotations = [
    { at: 0.30, kind: "pause", dur: 0.06, label: "0.8s" },
    { at: 0.60, kind: "pause", dur: 0.09, label: "1.6s · 偏长" },
    { at: 0.44, kind: "filler", label: "嗯…" },
    { at: 0.78, kind: "filler", label: "就是" },
    { at: 0.15, kind: "pace", label: "正常" },
    { at: 0.85, kind: "pace", label: "偏快" },
  ];

  const L = lang === "en" ? {
    round: "Behavior · Voice mode · Round 2",
    q: "Tell me about a time you had a serious disagreement with a designer — how did you pull the team toward alignment?",
    live: "LIVE · 正在转写",
    transcript: "Live transcript",
    speaking: "Lin Zhou",
    metrics: "Expression metrics",
    pause: "Pause",
    done: "End & review",
  } : {
    round: "行为面 · 语音模式 · 第 2 题",
    q: "描述一次你和设计师在实现方案上分歧较大的情况，你是怎么把团队拉到对齐的？",
    live: "录音中 · 正在实时转写",
    transcript: "实时转写",
    speaking: "林舟",
    metrics: "表达层指标",
    pause: "暂停",
    done: "结束并生成报告",
  };

  const metrics = [
    { k: lang === "en" ? "Words / min" : "语速", v: "186", hint: lang === "en" ? "steady 160-200 wpm" : "稳定在 160-200 字/分", tone: "ok", bar: 0.7 },
    { k: lang === "en" ? "Long pauses" : "长停顿", v: "2", hint: lang === "en" ? "2 over 1.5s" : "本题 2 次超过 1.5 秒", tone: "warn", bar: 0.5 },
    { k: lang === "en" ? "Fillers" : "口头禅", v: "4", hint: "「嗯」×2 · 「就是」×2", tone: "danger", bar: 0.6 },
    { k: lang === "en" ? "Volume" : "音量", v: "稳定", hint: lang === "en" ? "no drop-offs" : "没有明显衰减", tone: "ok", bar: 0.78 },
  ];

  const transcript = [
    { t: "00:02:14", role: "ai", text: lang === "en" ? "Take your time. When you're ready, start with the situation." : "别急，你准备好了就从当时的情境开始。" },
    { t: "00:02:18", role: "user", text: lang === "en" ? "OK. So the project was our order repricing system…" : "好。那个项目是我们的订单改价系统——" },
    { t: "00:02:27", role: "user", text: lang === "en" ? "We needed to support 40+ fields, and the designer wanted to keep everything on one screen." : "要支持 40 多个字段，当时设计想把所有字段都放在一屏上。" },
    { t: "00:02:41", role: "user", text: "嗯… 我当时担心的其实是认知负荷，就是…", fillerFlags: ["嗯", "就是"] },
    { t: "00:02:44", role: "user", text: "操作员一天要处理 300 多单，如果每单都要在一屏上扫 40 个字段…", pause: "1.2s" },
    { t: "00:02:49", role: "ai-note", text: lang === "en" ? "Soft pause detected · 1.6s" : "检测到长停顿 · 1.6 秒", tone: "warn" },
  ];

  return (
    <div className="ei-fadein" style={{ height: "100vh", display: "flex", flexDirection: "column", background: T.bg, color: T.ink }}>
      {/* Top strip */}
      <div style={{ padding: "14px 32px", borderBottom: `1px solid ${T.rule}`, display: "flex", alignItems: "center", gap: 20 }}>
        <button onClick={() => nav("workspace", { jobId: "tj-1" })} style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", display: "flex", gap: 6, alignItems: "center", fontSize: 13 }}>
          <Icon name="arrow_left" size={14} /> {lang === "en" ? "Workspace" : "工作台"}
        </button>
        <span style={{ color: T.rule }}>/</span>
        <div className="ei-label" style={{ color: T.ink3 }}>{L.round}</div>
        <div style={{ flex: 1 }} />
        <div style={{ display: "flex", gap: 8, alignItems: "center", padding: "5px 12px", background: T.accentSoft, border: `1px solid ${T.accent}`, borderRadius: 2 }}>
          <div style={{ width: 8, height: 8, borderRadius: 4, background: T.accent }} className="ei-pulse" />
          <div style={{ color: T.accent, fontSize: 12, fontFamily: "var(--ei-mono)", letterSpacing: "0.06em" }}>{L.live}</div>
        </div>
        <div style={{ fontFamily: "var(--ei-mono)", fontSize: 16, color: T.ink, fontVariantNumeric: "tabular-nums" }}>{fmt(elapsed)}</div>
      </div>

      {/* Question */}
      <div style={{ padding: "28px 48px 16px", borderBottom: `1px dotted ${T.rule}`, background: T.bgSoft }}>
        <div style={{ display: "flex", gap: 14, alignItems: "flex-start", maxWidth: 1080, margin: "0 auto" }}>
          <div className="ei-label" style={{ color: T.ink3, marginTop: 6, flexShrink: 0 }}>Q2 ·</div>
          <div className="ei-serif" style={{ fontSize: 22, lineHeight: 1.4, color: T.ink, letterSpacing: "-0.01em" }}>
            {L.q}
          </div>
        </div>
      </div>

      {/* Main grid */}
      <div style={{ flex: 1, display: "grid", gridTemplateColumns: "1fr 340px", minHeight: 0 }}>
        {/* Left: waveform + transcript */}
        <div style={{ padding: "28px 48px", overflowY: "auto", display: "flex", flexDirection: "column", gap: 24 }} className="ei-scroll">
          {/* Live bars */}
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 10 }}>
              <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "NOW SPEAKING · " : "正在说话 · "}{L.speaking}</div>
              <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink3 }}>—12 dB</div>
            </div>
            <div style={{ display: "flex", alignItems: "center", gap: 14, padding: "16px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3 }}>
              <div style={{ width: 38, height: 38, borderRadius: 19, background: T.accent, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
                <Icon name="mic" size={18} />
              </div>
              <WaveformBars T={T} bars={70} active={recording} height={48} />
            </div>
          </div>

          {/* Annotated waveform (the last answer so far) */}
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 10 }}>
              <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "CURRENT ANSWER · annotated" : "本次回答 · 已标注"}</div>
              <div style={{ display: "flex", gap: 14, fontFamily: "var(--ei-mono)", fontSize: 11 }}>
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
            <div style={{ padding: "16px 20px 4px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3 }}>
              <AnnotatedWaveform T={T} samples={samples} annotations={annotations} />
            </div>
          </div>

          {/* Transcript */}
          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{L.transcript}</div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {transcript.map((m, i) => {
                if (m.role === "ai-note") return (
                  <div key={i} style={{ display: "flex", gap: 10, padding: "6px 12px", background: T.warnSoft, borderLeft: `2px solid ${T.warn}`, fontSize: 12, color: T.warn, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>
                    <Icon name="info" size={13} /> {m.text}
                  </div>
                );
                const isAI = m.role === "ai";
                return (
                  <div key={i} style={{ display: "flex", gap: 12 }}>
                    <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink4, width: 58, flexShrink: 0, paddingTop: 2 }}>{m.t}</div>
                    <div style={{
                      fontSize: 14, lineHeight: 1.55,
                      color: isAI ? T.ink3 : T.ink,
                      fontStyle: isAI ? "italic" : "normal",
                      flex: 1,
                    }}>
                      {m.text}
                      {m.pause && <span style={{ display: "inline-block", marginLeft: 8, padding: "1px 6px", background: T.amberSoft, color: T.warn, fontSize: 11, fontFamily: "var(--ei-mono)", borderRadius: 2 }}>⏸ {m.pause}</span>}
                    </div>
                  </div>
                );
              })}
              {/* streaming cursor */}
              <div style={{ display: "flex", gap: 12 }}>
                <div style={{ width: 58, flexShrink: 0 }} />
                <div style={{ width: 8, height: 16, background: T.accent }} className="ei-pulse" />
              </div>
            </div>
          </div>
        </div>

        {/* Right: metrics + controls */}
        <div style={{ borderLeft: `1px solid ${T.rule}`, background: T.bgSoft, display: "flex", flexDirection: "column" }}>
          <div style={{ padding: "28px 24px", flex: 1, overflowY: "auto" }} className="ei-scroll">
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{L.metrics}</div>
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

            <div style={{ marginTop: 28, padding: 14, background: T.bgCard, border: `1px dotted ${T.rule}`, borderRadius: 3 }}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "GENTLE NUDGE" : "现场提示"}</div>
              <div style={{ fontSize: 13, color: T.ink, lineHeight: 1.55 }}>
                <Icon name="sparkle" size={13} color={T.accent} style={{ marginRight: 6 }} />
                {lang === "en"
                  ? "You're midway through the situation. Try to name the concrete action you took before the designer responded."
                  : "当前在讲「情境」。试着先说一句你采取的具体行动，再切到对方的反应。"}
              </div>
            </div>

            <div style={{ marginTop: 18, fontSize: 11, color: T.ink4, lineHeight: 1.5, fontFamily: "var(--ei-mono)" }}>
              {lang === "en"
                ? "audio stays on-device during the session · deleted after report"
                : "音频仅在本次会话缓存 · 报告生成后自动删除"}
            </div>
          </div>

          {/* Controls */}
          <div style={{ padding: "16px 24px", borderTop: `1px solid ${T.rule}`, background: T.bgCard, display: "flex", gap: 10 }}>
            <button
              onClick={() => setRecording(!recording)}
              style={{
                flex: 1, height: 40, background: recording ? T.bgSoft : T.accent,
                color: recording ? T.ink2 : "#fff",
                border: `1px solid ${recording ? T.rule : T.accent}`, borderRadius: 2, cursor: "pointer",
                display: "flex", alignItems: "center", justifyContent: "center", gap: 8, fontSize: 13, fontWeight: 500,
              }}>
              <Icon name={recording ? "pause" : "mic"} size={13} /> {recording ? L.pause : (lang === "en" ? "Resume" : "继续")}
            </button>
            <button
              onClick={() => nav("report")}
              style={{ flex: 1, height: 40, background: T.ink, color: T.bg, border: "none", borderRadius: 2, cursor: "pointer", fontSize: 13, fontWeight: 500 }}>
              {L.done}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

window.VoicePracticeScreen = VoicePracticeScreen;
