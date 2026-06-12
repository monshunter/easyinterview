// P1 depth: Real-interview debrief, Experience library, Resume versions + diff

// ═══════════════════════════════════════════════════════════════════
// #9 DEBRIEF (full version)
// ═══════════════════════════════════════════════════════════════════

// Unified source badge — rendered for entries in both text and voice modes
// so the user always sees how each card got into the debrief record.
const getEntrySourceBadge = (source, lang) => {
  switch (source) {
    case "real_recorded":
    case "confirmed":
      return { icon: "✓", label: lang === "en" ? "recorded" : "已记录", tone: "ok" };
    case "text_guided":
    case "ai_confirmed":
      return { icon: "✏️", label: lang === "en" ? "text" : "文本", tone: "accent" };
    case "voice_extracted":
    case "voice_extracted_confirmed":
      return { icon: "🎙", label: lang === "en" ? "voice" : "语音", tone: "accent" };
    case "manual":
      return { icon: "✏️", label: lang === "en" ? "manual" : "手动", tone: "neutral" };
    default:
      return { icon: "·", label: source || "", tone: "neutral" };
  }
};

// Context preamble injected into the AI's first message when entries already exist
// (e.g. user added some via text mode then switched to voice). Avoids the AI re-asking
// what's already been captured.
const buildVoiceContextOpener = (entries, lang) => {
  if (!entries || entries.length === 0) return "";
  const stages = Array.from(new Set(entries.map((e) => e.stage))).slice(0, 3);
  return lang === "en"
    ? `I can see ${entries.length} question${entries.length > 1 ? "s" : ""} you've already recorded (${stages.join(", ")}). Let's keep going on what's missing —`
    : `我看到你已经记了 ${entries.length} 条问题（${stages.join("、")}）。咱们看看还有什么没覆盖到——`;
};

const DebriefFullScreen = ({ T, lang, nav, params = {} }) => {
  const [step, setStep] = React.useState(0); // 0 record, 1 analyze, 2 replay
  const [activeGuide, setActiveGuide] = React.useState(0);
  const [activeCard, setActiveCard] = React.useState("e1");
  const [inputMode, setInputMode] = React.useState(params.debriefInputMode || "text"); // "text" | "voice"

  const steps = lang === "en"
    ? ["Debrief record", "Debrief analysis", "Debrief interview"]
    : ["复盘记录", "复盘分析", "复盘面试"];

  const contextOptions = getDebriefContextOptions(lang);
  const [selectedContext, setSelectedContext] = React.useState({
    targetJob: "tj-1",
    mockSession: "mock-24",
    resume: "resume-v3",
  });
  const [pickerType, setPickerType] = React.useState(null);
  // Pick-one-derive-two: derivable context is auto-filled and stays changeable (D-15).
  const [autoFilled, setAutoFilled] = React.useState({ targetJob: false, mockSession: false, resume: false });
  const applyContextSelection = (kind, id) => {
    if (kind === "mockSession") {
      const mock = contextOptions.mockSessions.find((item) => item.id === id);
      setSelectedContext({
        mockSession: id,
        targetJob: mock?.targetJobId || selectedContext.targetJob,
        resume: mock?.resumeId || selectedContext.resume,
      });
      setAutoFilled({ targetJob: !!mock?.targetJobId, mockSession: false, resume: !!mock?.resumeId });
      return;
    }
    if (kind === "targetJob") {
      const job = contextOptions.targetJobs.find((item) => item.id === id);
      setSelectedContext({
        targetJob: id,
        resume: job?.defaultResumeId || selectedContext.resume,
        mockSession: job?.latestMockId || selectedContext.mockSession,
      });
      setAutoFilled({ targetJob: false, mockSession: !!job?.latestMockId, resume: !!job?.defaultResumeId });
      return;
    }
    setSelectedContext({ ...selectedContext, [kind]: id });
    setAutoFilled((prev) => ({ ...prev, [kind]: false }));
  };
  const selectedTarget = contextOptions.targetJobs.find((item) => item.id === selectedContext.targetJob) || contextOptions.targetJobs[0];
  const selectedMock = contextOptions.mockSessions.find((item) => item.id === selectedContext.mockSession) || contextOptions.mockSessions[0];
  const selectedResume = contextOptions.resumes.find((item) => item.id === selectedContext.resume) || contextOptions.resumes[0];
  const context = {
    target: selectedTarget.title,
    jd: selectedTarget.meta,
    mock: selectedMock.title,
    resume: selectedResume.title,
  };

  const guideQuestions = lang === "en" ? [
    { stage: "Opening", q: "Did they ask you to introduce yourself or walk through your background?", why: "Common first-round opening; compare against the mock self-introduction and JD positioning.", source: "JD + mock #24" },
    { stage: "Project deep dive", q: "Did they ask about the checkout / RSC performance work?", why: "Your resume and the target JD both point to performance and architecture ownership.", source: "resume v3 + JD must-have" },
    { stage: "Influence", q: "Did they ask how you rolled out Design System work across teams?", why: "The JD emphasizes cross-team technical influence and platform adoption.", source: "JD hidden signal" },
    { stage: "Reverse Q", q: "Did you ask them about team priorities or next-round expectations?", why: "Reverse questions reveal whether the conversation ended with a clear next signal.", source: "mock report" },
  ] : [
    { stage: "开场", q: "他们是否让你做自我介绍，或完整讲一遍背景？", why: "这是技术一面的常见开场，需要和模拟面试里的自我介绍、目标 JD 定位对齐。", source: "JD + 模拟面试 #24" },
    { stage: "项目深挖", q: "他们是否问到结账链路 / RSC 性能优化项目？", why: "你的简历和目标 JD 都指向性能优化与架构 ownership。", source: "简历 v3 + JD 必需项" },
    { stage: "影响力", q: "他们是否问到 Design System 如何跨团队推进？", why: "JD 强调跨团队技术影响力和平台落地。", source: "JD 隐性关注点" },
    { stage: "反问", q: "你是否向对方询问团队重点、下一轮预期或当前痛点？", why: "反问能判断这轮面试是否收束到清晰的下一步信号。", source: "模拟报告" },
  ];

  // Mock real Q&A entries
  const [entries, setEntries] = React.useState([
    {
      id: "e1",
      stage: lang === "en" ? "Project deep dive" : "项目深挖",
      q: lang === "en" ? "Walk me through the checkout perf work — what exactly did YOU drive?" : "跟我讲讲结账性能那个项目——具体哪些是你推动的？",
      a: lang === "en" ? "I talked about RSC migration, but fumbled on the 'you specifically' part." : "讲了 RSC 迁移，但「具体你做了什么」那里卡住了。",
      follow: lang === "en" ? "They asked who made the final architecture call." : "对方追问最后架构方案是谁拍板的。",
      reflection: lang === "en" ? "Ownership wording was vague." : "Ownership 表达不够明确。",
      reaction: "probed",
      source: "confirmed",
      tag: lang === "en" ? "ownership" : "Ownership",
    },
    {
      id: "e2",
      stage: lang === "en" ? "Influence" : "影响力",
      q: lang === "en" ? "How would you roll out a Design System across 8 teams?" : "8 个团队的 Design System 你会怎么推？",
      a: lang === "en" ? "Started with tokens, pilot team, workshops. They nodded. Moved on fast." : "讲了 tokens、先挑一个团队试点、办推广会。对方点头，很快切下一题。",
      follow: lang === "en" ? "No follow-up, but they asked about adoption metrics later." : "没有继续追问，但后面问了采用率指标。",
      reflection: lang === "en" ? "Good structure, weak numbers." : "结构可以，数字不够。",
      reaction: "positive",
      source: "confirmed",
      tag: lang === "en" ? "influence" : "影响力",
    },
    {
      id: "e3",
      stage: lang === "en" ? "Reverse Q" : "反问",
      q: lang === "en" ? "Any questions for me?" : "你有什么想问我的吗？",
      a: lang === "en" ? "Asked about engineering culture. Didn't ask about the team's top pain or his personal take." : "问了工程文化。没问团队眼下最头疼的事、也没问他个人看法。",
      follow: lang === "en" ? "They answered politely and ended the call." : "对方礼貌回答后结束了面试。",
      reflection: lang === "en" ? "Ended too generic." : "收尾过于泛。",
      reaction: "neutral",
      source: "confirmed",
      tag: lang === "en" ? "reverse-Q" : "反问",
    },
  ]);

  const reactions = lang === "en"
    ? { positive: "visibly engaged", neutral: "polite", probed: "pushed back / probed", skeptical: "skeptical" }
    : { positive: "明显投入", neutral: "礼貌回应", probed: "追问 / 反推", skeptical: "有保留" };

  const currentGuide = guideQuestions[activeGuide];
  const activeEntry = entries.find((e) => e.id === activeCard) || entries[0];

  return (
    <div className="ei-fadein" style={{ maxWidth: 1200, margin: "0 auto", padding: "32px 48px 96px" }}>
      <button onClick={() => nav("home")} style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", display: "flex", gap: 6, alignItems: "center", fontSize: 13, padding: 0, marginBottom: 16 }}>
        <Icon name="arrow_left" size={13} /> {lang === "en" ? "Back home" : "返回首页"}
      </button>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginBottom: 32, gap: 32 }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
            {lang === "en" ? "DEBRIEF · 星环科技 · 技术一面 · 4/22" : "复盘 · 星环科技 · 技术一面 · 4/22"}
          </div>
          <h1 className="ei-serif" style={{ fontSize: 38, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.15, maxWidth: 780 }}>
            {lang === "en" ? "Reconstruct the interview, then practice the replay." : "像真人复盘一样，还原面试，再进入复盘面试。"}
          </h1>
          <div style={{ fontSize: 14, color: T.ink3, marginTop: 10, maxWidth: 640, lineHeight: 1.5 }}>
            {lang === "en" ? "Pick the target JD and resume, keep or skip AI-suggested questions, add what was asked, then generate a replay interview." : "选择目标岗位和简历，保留或跳过 AI 推测题，补充被问到的问题，然后生成一场复盘面试。"}
          </div>
        </div>
        <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink3, textAlign: "right", lineHeight: 1.7 }}>
          <div>time · 40 min ago</div>
          <div>interviewer · 张哲 · tech lead</div>
          <div>modality · video</div>
        </div>
      </div>

      <DebriefContextStrip T={T} lang={lang} context={context} autoFilled={autoFilled} onOpenPicker={setPickerType} />

      {pickerType && (
        <DebriefContextPickerModal
          T={T}
          lang={lang}
          kind={pickerType}
          options={getDebriefOptionsForKind(contextOptions, pickerType)}
          selectedId={selectedContext[pickerType]}
          onClose={() => setPickerType(null)}
          onConfirm={(id) => {
            applyContextSelection(pickerType, id);
            setPickerType(null);
          }}
        />
      )}

      {/* Stepper */}
      <div style={{ display: "flex", gap: 0, marginBottom: 36, borderBottom: `1px solid ${T.rule}` }}>
        {steps.map((s, i) => (
          <button key={i} onClick={() => setStep(i)} style={{
            padding: "12px 24px", background: "transparent", border: "none",
            borderBottom: `2px solid ${step === i ? T.accent : "transparent"}`,
            color: step === i ? T.ink : T.ink3, cursor: "pointer",
            fontFamily: "var(--ei-sans)", fontSize: 14, fontWeight: step === i ? 500 : 400,
            marginBottom: -1,
          }}>
            <span style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink4, marginRight: 8 }}>{String(i + 1).padStart(2, "0")}</span>
            {s}
          </button>
        ))}
      </div>

      {/* Step 0: Record */}
      {step === 0 && (
        <div>
          {/* Unified summary bar — shared list across both add modes */}
          {(() => {
            const counts = { real: 0, text: 0, voice: 0, manual: 0 };
            entries.forEach((e) => {
              if (e.source === "real_recorded" || e.source === "confirmed") counts.real++;
              else if (e.source === "text_guided" || e.source === "ai_confirmed") counts.text++;
              else if (e.source === "voice_extracted" || e.source === "voice_extracted_confirmed") counts.voice++;
              else if (e.source === "manual") counts.manual++;
            });
            const chips = [
              counts.real > 0 && { icon: "✓", label: lang === "en" ? "recorded" : "已记录", n: counts.real, tone: T.ok },
              counts.text > 0 && { icon: "✏️", label: lang === "en" ? "text" : "文本", n: counts.text, tone: T.accent },
              counts.voice > 0 && { icon: "🎙", label: lang === "en" ? "voice" : "语音", n: counts.voice, tone: T.accent },
              counts.manual > 0 && { icon: "✏️", label: lang === "en" ? "manual" : "手动", n: counts.manual, tone: T.ink2 },
            ].filter(Boolean);
            return (
              <div style={{
                display: "flex", alignItems: "center", justifyContent: "space-between", gap: 16,
                padding: "11px 16px", marginBottom: 14,
                background: T.bgSoft, border: `1px solid ${T.rule}`, borderLeft: `3px solid ${T.accent}`, borderRadius: 2,
              }}>
                <div style={{ display: "flex", alignItems: "center", gap: 14, flexWrap: "wrap" }}>
                  <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "DEBRIEF RECORD" : "复盘记录"}</div>
                  <div style={{ display: "flex", alignItems: "baseline", gap: 6 }}>
                    <span style={{ fontFamily: "var(--ei-mono)", fontSize: 19, color: T.ink, fontWeight: 500, letterSpacing: "-0.01em" }}>{entries.length}</span>
                    <span style={{ fontSize: 12, color: T.ink3 }}>{lang === "en" ? "questions" : "条已记录"}</span>
                  </div>
                  {chips.length > 0 && (
                    <>
                      <div style={{ height: 16, width: 1, background: T.rule }} />
                      <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
                        {chips.map((c) => (
                          <div key={c.label} style={{
                            display: "inline-flex", alignItems: "center", gap: 5,
                            fontFamily: "var(--ei-mono)", fontSize: 11, color: c.tone,
                          }}>
                            <span style={{ fontSize: 11 }}>{c.icon}</span>
                            <span>{c.label}</span>
                            <span style={{ color: T.ink2, fontWeight: 500 }}>{c.n}</span>
                          </div>
                        ))}
                      </div>
                    </>
                  )}
                </div>
                <div style={{ fontSize: 11, color: T.ink4, fontFamily: "var(--ei-mono)", flexShrink: 0 }}>
                  {lang === "en" ? "shared list · safe to switch modes" : "跨模式共享 · 切换不丢数据"}
                </div>
              </div>
            );
          })()}

          {/* Add-mode toggle */}
          <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 18 }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "ADD MODE" : "添加方式"}</div>
            <div style={{ display: "flex", border: `1px solid ${T.rule}`, borderRadius: 2, overflow: "hidden" }}>
              {[
                { k: "text", icon: "✏️", t: lang === "en" ? "Text" : "文本" },
                { k: "voice", icon: "🎙", t: lang === "en" ? "Voice" : "语音" },
              ].map((m, mi) => (
                <button key={m.k} onClick={() => setInputMode(m.k)} style={{
                  padding: "6px 16px", fontSize: 13, cursor: "pointer",
                  border: "none", borderRight: mi === 0 ? `1px solid ${T.rule}` : "none",
                  background: inputMode === m.k ? T.ink : "transparent",
                  color: inputMode === m.k ? T.bg : T.ink3,
                  fontFamily: "var(--ei-sans)", transition: "background .15s",
                }}>{m.icon} {m.t}</button>
              ))}
            </div>
            <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink3 }}>
              {inputMode === "voice"
                ? (lang === "en" ? "Speak naturally · AI extracts cards in real time" : "自由讲述 · AI 实时抽取卡片")
                : (lang === "en" ? "AI walks you through likely questions" : "AI 逐题引导，确认后写入")}
            </div>
          </div>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 320px", gap: 28 }}>
            <div>
              {/* Both modes always mounted — display:none switching preserves
                  voice conversation state when user toggles back to text. */}
              <div style={{ display: inputMode === "text" ? "block" : "none" }}>
                <GuidedDebriefRecord
                  T={T}
                  lang={lang}
                  currentGuide={currentGuide}
                  guideIndex={activeGuide}
                  guideTotal={guideQuestions.length}
                  setActiveGuide={setActiveGuide}
                  entries={entries}
                  setEntries={setEntries}
                  activeCard={activeCard}
                  setActiveCard={setActiveCard}
                  reactions={reactions}
                />
              </div>
              <div style={{ display: inputMode === "voice" ? "block" : "none" }}>
                <VoiceDebriefRecord
                  T={T}
                  lang={lang}
                  entries={entries}
                  setEntries={setEntries}
                  initialState={params.debriefVoiceState || "idle"}
                  active={inputMode === "voice"}
                />
              </div>
            </div>

            <div>
              <Card T={T} pad={18} style={{ position: "sticky", top: 20 }}>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "VIBE CHECK" : "整体感受"}</div>
                <div style={{ display: "flex", flexDirection: "column", gap: 12, fontSize: 13 }}>
                  <div>
                    <div style={{ color: T.ink2, marginBottom: 5 }}>{lang === "en" ? "Overall feeling" : "整体感受"}</div>
                    <div style={{ display: "flex", gap: 6 }}>
                      {["🙁", "😐", "🙂", "😊"].map((e, i) => (
                        <button key={i} style={{ width: 36, height: 36, borderRadius: 2, border: `1px solid ${i === 2 ? T.accent : T.rule}`, background: i === 2 ? T.accentSoft : "transparent", cursor: "pointer", fontSize: 18 }}>{e}</button>
                      ))}
                    </div>
                  </div>
                  <div>
                    <div style={{ color: T.ink2, marginBottom: 5 }}>{lang === "en" ? "What they seemed to like" : "他们似乎认可的"}</div>
                    <textarea rows={2} style={{ width: "100%", padding: "8px 10px", fontSize: 12.5, border: `1px solid ${T.rule}`, borderRadius: 2, background: T.bgSoft, outline: "none", resize: "vertical", boxSizing: "border-box", fontFamily: "var(--ei-sans)", color: T.ink2 }} defaultValue={lang === "en" ? "concrete metrics, trade-off framing" : "具体量化 · 权衡表达"} />
                  </div>
                  <div>
                    <div style={{ color: T.ink2, marginBottom: 5 }}>{lang === "en" ? "Where I stumbled" : "我卡住的地方"}</div>
                    <textarea rows={2} style={{ width: "100%", padding: "8px 10px", fontSize: 12.5, border: `1px solid ${T.rule}`, borderRadius: 2, background: T.bgSoft, outline: "none", resize: "vertical", boxSizing: "border-box", fontFamily: "var(--ei-sans)", color: T.ink2 }} defaultValue={lang === "en" ? "ownership attribution, reverse-Q" : "Ownership 归属 · 反问"} />
                  </div>
                </div>
              </Card>
            </div>
          </div>

          <div style={{ display: "flex", justifyContent: "flex-end", marginTop: 28 }}>
            <Btn T={T} variant="accent" iconRight="arrow_right" onClick={() => setStep(1)}>{lang === "en" ? "Generate debrief analysis" : "生成复盘分析"}</Btn>
          </div>
        </div>
      )}

      {/* Step 1: Analysis */}
      {step === 1 && (
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "INTERVIEW ANALYSIS" : "面试分析"}</div>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14, marginBottom: 28 }}>
            {[
              { k: lang === "en" ? "Questions overlap" : "题目重合度", sim: "6", real: "3", tone: "warn", hint: lang === "en" ? "only 50% of sim Qs came up" : "仅 50% 模拟题命中" },
              { k: lang === "en" ? "Ownership depth" : "Ownership 深度", sim: lang === "en" ? "steady" : "稳", real: lang === "en" ? "stumbled" : "卡住", tone: "danger", hint: lang === "en" ? "main gap to close" : "最核心的差距" },
              { k: lang === "en" ? "Perf quantification" : "性能量化", sim: "✓", real: "✓", tone: "ok", hint: lang === "en" ? "held up" : "扛住了" },
              { k: lang === "en" ? "Reverse-Q count" : "反问数量", sim: "3 planned", real: "1", tone: "warn", hint: lang === "en" ? "panic-cut the list" : "慌里砍掉了" },
            ].map((m) => {
              const c = { ok: T.ok, warn: T.warn, danger: T.danger }[m.tone];
              return (
                <div key={m.k} style={{ padding: "18px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderLeft: `3px solid ${c}`, borderRadius: 2 }}>
                  <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{m.k}</div>
                  <div style={{ display: "flex", alignItems: "baseline", gap: 14 }}>
                    <div><div style={{ fontSize: 11, color: T.ink4, fontFamily: "var(--ei-mono)" }}>SIM</div><div style={{ fontFamily: "var(--ei-mono)", fontSize: 20, color: T.ink2 }}>{m.sim}</div></div>
                    <Icon name="arrow_right" size={14} color={T.ink3} />
                    <div><div style={{ fontSize: 11, color: c, fontFamily: "var(--ei-mono)" }}>REAL</div><div style={{ fontFamily: "var(--ei-mono)", fontSize: 20, color: c, fontWeight: 500 }}>{m.real}</div></div>
                  </div>
                  <div style={{ fontSize: 12, color: T.ink3, marginTop: 8, lineHeight: 1.5 }}>{m.hint}</div>
                </div>
              );
            })}
          </div>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16, marginBottom: 28 }}>
            {[
              { title: lang === "en" ? "Compared with target JD" : "对照目标 JD", body: lang === "en" ? "This interview spent more time on ownership and influence than the mock interview predicted." : "这轮面试比模拟面试更集中在 ownership 与影响力，而不是单纯技术细节。" },
              { title: lang === "en" ? "Compared with resume evidence" : "对照绑定简历", body: lang === "en" ? "Resume v3 has the right stories, but the interview answer did not claim the decision points clearly." : "简历 v3 有对应素材，但面试回答没有把决策点和个人贡献讲清。" },
            ].map((m) => (
              <div key={m.title} style={{ padding: "18px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
                <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{m.title}</div>
                <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.65 }}>{m.body}</div>
              </div>
            ))}
          </div>

          <div style={{ display: "flex", justifyContent: "space-between" }}>
            <Btn T={T} variant="ghost" onClick={() => setStep(0)}>{lang === "en" ? "Back" : "上一步"}</Btn>
            <Btn T={T} variant="accent" iconRight="arrow_right" onClick={() => setStep(2)}>{lang === "en" ? "Generate debrief interview" : "生成复盘面试"}</Btn>
          </div>
        </div>
      )}

      {/* Step 2: Debrief interview */}
      {step === 2 && <DebriefReplayPlan T={T} lang={lang} nav={nav} back={() => setStep(1)} entries={entries} context={context} />}
    </div>
  );
};

const getDebriefContextOptions = (lang) => lang === "en" ? {
  targetJobs: [
    { id: "tj-1", title: "Star-Ring Tech · Senior Frontend Engineer", meta: "P6 · Shanghai · JD match 78%", note: "Current real interview target. Used to anchor debrief questions and replay practice.", defaultResumeId: "resume-v3", latestMockId: "mock-24" },
    { id: "tj-2", title: "Lumen Labs · Frontend Platform Engineer", meta: "Senior · remote · JD match 64%", note: "English HR-screen context. Pick this only when the debrief belongs to that process.", defaultResumeId: "resume-en", latestMockId: null },
    { id: "tj-3", title: "CloudYun Group · Web Architecture Expert", meta: "P7 · Hangzhou · JD match 52%", note: "Draft target. Complete JD details before using it as the debrief baseline.", defaultResumeId: "resume-v3", latestMockId: null },
  ],
  mockSessions: [
    { id: "mock-24", title: "Mock interview #24 · text · 4/20", meta: "Star-Ring Tech · Technical round 1 · report ready", note: "Best baseline for this real technical interview.", targetJobId: "tj-1", resumeId: "resume-v3" },
    { id: "mock-23", title: "Mock interview #23 · voice · 4/19", meta: "Star-Ring Tech · Technical round 1 · second run", note: "Use when comparing against the replay run instead of the first report.", targetJobId: "tj-1", resumeId: "resume-v3" },
    { id: "mock-20", title: "Mock interview #20 · text · 4/17", meta: "Star-Ring Tech · Technical round 2", note: "Useful if the real interview focused on architecture probes.", targetJobId: "tj-1", resumeId: "resume-v3" },
  ],
  resumes: [
    { id: "resume-v3", title: "Liu Zhe · resume v3 · 78% match", meta: "Master version · source retained", note: "Primary evidence source for this debrief analysis." },
    { id: "resume-impact", title: "Liu Zhe · collaboration impact v2", meta: "Guided resume draft · 2026-04-18", note: "Use when the interview focused on influence and rollout stories." },
    { id: "resume-en", title: "Liu Zhe · Frontend Platform EN v1", meta: "English version · source retained", note: "Use for English-language interview debriefs." },
  ],
} : {
  targetJobs: [
    { id: "tj-1", title: "星环科技 · 资深前端工程师", meta: "P6 · 上海 · JD 匹配 78%", note: "当前真实面试目标，用来锚定复盘问题和复盘面试。", defaultResumeId: "resume-v3", latestMockId: "mock-24" },
    { id: "tj-2", title: "Lumen Labs · Frontend Platform Engineer", meta: "Senior · 远程 · JD 匹配 64%", note: "英文 HR 初筛上下文；只有复盘属于这条流程时才选择。", defaultResumeId: "resume-en", latestMockId: null },
    { id: "tj-3", title: "云栖集团 · 技术专家（Web 架构）", meta: "P7 · 杭州 · JD 匹配 52%", note: "草稿目标；用于复盘前应先补全 JD 细节。", defaultResumeId: "resume-v3", latestMockId: null },
  ],
  mockSessions: [
    { id: "mock-24", title: "模拟面试 #24 · 文本 · 4/20", meta: "星环科技 · 技术一面 · 报告已生成", note: "当前真实技术面最合适的对比基线。", targetJobId: "tj-1", resumeId: "resume-v3" },
    { id: "mock-23", title: "模拟面试 #23 · 语音 · 4/19", meta: "星环科技 · 技术一面 · 第 2 次", note: "当用户想和复练后的表现对比时选择。", targetJobId: "tj-1", resumeId: "resume-v3" },
    { id: "mock-20", title: "模拟面试 #20 · 文本 · 4/17", meta: "星环科技 · 技术二面", note: "真实面试偏架构追问时可作为对照。", targetJobId: "tj-1", resumeId: "resume-v3" },
  ],
  resumes: [
    { id: "resume-v3", title: "刘哲 · 简历 v3 · 匹配 78%", meta: "主版本 · 保留原始来源", note: "当前复盘分析的主要证据来源。" },
    { id: "resume-impact", title: "刘哲 · 协作影响力版 v2", meta: "由简历问答生成 · 2026-04-18", note: "真实面试更偏影响力和落地故事时选择。" },
    { id: "resume-en", title: "Liu Zhe · Frontend Platform EN v1", meta: "英文版 · 保留上传原件", note: "用于英文真实面试的复盘。" },
  ],
};

const getDebriefOptionsForKind = (contextOptions, kind) => ({
  targetJob: contextOptions.targetJobs,
  mockSession: contextOptions.mockSessions,
  resume: contextOptions.resumes,
}[kind] || []);

const DebriefContextStrip = ({ T, lang, context, autoFilled = {}, onOpenPicker }) => (
  <div style={{ display: "grid", gridTemplateColumns: "1.2fr 1fr 1fr", gap: 12, marginBottom: 28 }}>
    {[
      { key: "targetJob", icon: "briefcase", label: lang === "en" ? "Target job / JD" : "目标岗位 / JD", title: context.target, meta: context.jd, action: lang === "en" ? "Change" : "更换" },
      { key: "mockSession", icon: "chart", label: lang === "en" ? "Related mock interview" : "关联模拟面试", title: context.mock, meta: lang === "en" ? "used as comparison baseline" : "作为面试分析基线", action: lang === "en" ? "Select" : "选择" },
      { key: "resume", icon: "resume", label: lang === "en" ? "Resume version" : "绑定简历", title: context.resume, meta: lang === "en" ? "used for evidence comparison" : "用于回答证据对比", action: lang === "en" ? "Change" : "更换" },
    ].map((item) => (
      <div key={item.key} style={{ padding: "14px 16px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, display: "grid", gridTemplateColumns: "30px 1fr auto", gap: 10, alignItems: "center" }}>
        <div style={{ width: 30, height: 30, borderRadius: 15, background: T.bgSoft, color: T.accent, display: "flex", alignItems: "center", justifyContent: "center" }}>
          <Icon name={item.icon} size={14} />
        </div>
        <div style={{ minWidth: 0 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 3 }}>{item.label}</div>
          <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{item.title}</div>
          <div style={{ fontSize: 12, color: T.ink3, marginTop: 2, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{item.meta}</div>
          {autoFilled[item.key] && (
            <div style={{ fontSize: 10.5, color: T.accent, fontFamily: "var(--ei-mono)", marginTop: 3, letterSpacing: "0.04em" }}>
              {lang === "en" ? "AUTO-FILLED · CHANGEABLE" : "已自动带入 · 可更换"}
            </div>
          )}
        </div>
        <button onClick={() => onOpenPicker(item.key)} style={{ background: "transparent", border: `1px solid ${T.rule}`, borderRadius: 2, color: T.ink2, fontSize: 12, padding: "5px 9px", cursor: "pointer" }}>{item.action}</button>
      </div>
    ))}
  </div>
);

const DebriefContextPickerModal = ({ T, lang, kind, options, selectedId, onClose, onConfirm }) => {
  const [draftId, setDraftId] = React.useState(selectedId);
  const selected = options.find((item) => item.id === draftId) || options[0];
  const meta = {
    targetJob: {
      eyebrow: lang === "en" ? "TARGET JOB / JD" : "目标岗位 / JD",
      title: lang === "en" ? "Choose the target JD for this debrief" : "选择这次复盘对应的目标岗位 / JD",
      body: lang === "en" ? "Changing it only updates this debrief context. It does not leave the debrief flow." : "更换后只更新本次复盘上下文，不离开复盘流程。",
      confirm: lang === "en" ? "Use this JD" : "使用这个 JD",
    },
    mockSession: {
      eyebrow: lang === "en" ? "RELATED MOCK" : "关联模拟面试",
      title: lang === "en" ? "Choose the mock interview baseline" : "选择作为对比基线的模拟面试",
      body: lang === "en" ? "Pick the completed session whose report should be compared with the real interview." : "选择一场已完成模拟面试，用它的报告和真实面试做对比。",
      confirm: lang === "en" ? "Use this mock" : "关联这场模拟面试",
    },
    resume: {
      eyebrow: lang === "en" ? "RESUME VERSION" : "绑定简历",
      title: lang === "en" ? "Choose the resume evidence source" : "选择这次复盘使用的简历",
      body: lang === "en" ? "The selected resume is used only for evidence comparison in this debrief." : "选择后仅作为本次复盘的回答证据对比来源。",
      confirm: lang === "en" ? "Use this resume" : "绑定这份简历",
    },
  }[kind];

  return (
    <div style={{ position: "fixed", inset: 0, background: "rgba(24, 20, 16, 0.24)", zIndex: 90, display: "flex", alignItems: "center", justifyContent: "center", padding: 24 }} onClick={onClose}>
      <div className="ei-fadein" onClick={(e) => e.stopPropagation()} style={{ width: "min(720px, 100%)", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4, boxShadow: "0 24px 70px rgba(30, 22, 15, 0.24)", padding: 24 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 18, marginBottom: 18 }}>
          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{meta.eyebrow}</div>
            <div className="ei-serif" style={{ fontSize: 23, color: T.ink }}>{meta.title}</div>
            <div style={{ fontSize: 13, color: T.ink3, marginTop: 6, lineHeight: 1.6 }}>{meta.body}</div>
          </div>
          <button onClick={onClose} style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", padding: 4 }}>
            <Icon name="x" size={16} />
          </button>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "1fr", gap: 10 }}>
          {options.map((item) => {
            const active = item.id === draftId;
            return (
              <button
                key={item.id}
                onClick={() => setDraftId(item.id)}
                style={{
                  textAlign: "left",
                  border: `1px solid ${active ? T.accent : T.rule}`,
                  background: active ? T.accentSoft : T.bgSoft,
                  borderRadius: 3,
                  padding: "14px 16px",
                  cursor: "pointer",
                  display: "grid",
                  gridTemplateColumns: "24px 1fr",
                  gap: 12,
                  alignItems: "start",
                }}
              >
                <span style={{ width: 20, height: 20, borderRadius: 10, border: `1px solid ${active ? T.accent : T.rule}`, background: active ? T.accent : T.bgCard, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", marginTop: 1 }}>
                  {active && <Icon name="check" size={12} stroke={2.2} />}
                </span>
                <span>
                  <span style={{ display: "block", fontSize: 14.5, color: T.ink, fontWeight: 600 }}>{item.title}</span>
                  <span style={{ display: "block", fontSize: 12.5, color: T.ink3, marginTop: 3 }}>{item.meta}</span>
                  <span style={{ display: "block", fontSize: 13, color: T.ink2, marginTop: 8, lineHeight: 1.55 }}>{item.note}</span>
                </span>
              </button>
            );
          })}
        </div>

        <div style={{ border: `1px solid ${T.rule}`, background: T.bgSoft, borderRadius: 3, padding: 14, marginTop: 16 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{lang === "en" ? "SELECTED" : "已选择"}</div>
          <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.6 }}>{selected.title} · {selected.meta}</div>
        </div>

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 10, marginTop: 22 }}>
          <Btn T={T} variant="ghost" onClick={onClose}>{lang === "en" ? "Cancel" : "取消"}</Btn>
          <Btn T={T} variant="accent" iconRight="arrow_right" onClick={() => onConfirm(draftId)}>{meta.confirm}</Btn>
        </div>
      </div>
    </div>
  );
};

const GuidedDebriefRecord = ({ T, lang, currentGuide, guideIndex, guideTotal, setActiveGuide, entries, setEntries, activeCard, setActiveCard, reactions }) => {
  const activeEntry = entries.find((e) => e.id === activeCard) || entries[0];
  const addCurrentQuestion = () => {
    const id = `e${entries.length + 1}`;
    const next = {
      id,
      stage: currentGuide.stage,
      q: currentGuide.q,
      a: "",
      follow: "",
      reflection: "",
      reaction: "neutral",
      source: "text_guided",
      tag: currentGuide.stage,
    };
    setEntries([...entries, next]);
    setActiveCard(id);
    setActiveGuide(Math.min(guideTotal - 1, guideIndex + 1));
  };
  return (
    <div>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "AI-GUIDED RECALL" : "AI 逐题复盘记录"}</div>
      <Card T={T} pad={0} style={{ marginBottom: 14 }}>
        <div style={{ padding: "16px 20px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", gap: 14, alignItems: "center" }}>
          <div>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 6 }}>{currentGuide.stage} · {String(guideIndex + 1).padStart(2, "0")} / {String(guideTotal).padStart(2, "0")}</div>
            <div className="ei-serif" style={{ fontSize: 22, color: T.ink, lineHeight: 1.35 }}>{currentGuide.q}</div>
          </div>
          <div style={{ width: 82, height: 82, borderRadius: 41, border: `1px solid ${T.rule}`, background: T.bgSoft, display: "flex", alignItems: "center", justifyContent: "center", color: T.accent, flexShrink: 0 }}>
            <Icon name="chat" size={28} />
          </div>
        </div>
        <div style={{ padding: "16px 20px", display: "grid", gridTemplateColumns: "1fr auto", gap: 18, alignItems: "center" }}>
          <div>
            <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.6 }}>{currentGuide.why}</div>
            <div style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 8 }}>{lang === "en" ? "source" : "来源"} · {currentGuide.source}</div>
          </div>
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap", justifyContent: "flex-end" }}>
            <Btn T={T} variant="secondary" size="sm" icon="check" onClick={addCurrentQuestion}>{lang === "en" ? "Yes, record it" : "遇到过，记录"}</Btn>
            <Btn T={T} variant="ghost" size="sm" onClick={() => setActiveGuide(Math.min(guideTotal - 1, guideIndex + 1))}>{lang === "en" ? "Skip" : "没问到，跳过"}</Btn>
            <Btn T={T} variant="ghost" size="sm" icon="edit" onClick={addCurrentQuestion}>{lang === "en" ? "Edit question" : "改成面试问题"}</Btn>
          </div>
        </div>
      </Card>

      <div style={{ display: "grid", gridTemplateColumns: "270px 1fr", gap: 14 }}>
        <Card T={T} pad={0}>
          <div style={{ padding: "14px 16px", borderBottom: `1px solid ${T.rule}` }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "QUESTION CARDS" : "已形成的问题卡片"}</div>
          </div>
          {entries.map((entry, idx) => (
            <button key={entry.id} onClick={() => setActiveCard(entry.id)} style={{
              width: "100%", padding: "13px 16px", textAlign: "left", border: "none", borderBottom: idx < entries.length - 1 ? `1px dotted ${T.rule}` : "none",
              background: activeCard === entry.id ? T.accentSoft : "transparent", cursor: "pointer", fontFamily: "var(--ei-sans)",
            }}>
              <div style={{ display: "flex", justifyContent: "space-between", gap: 8, marginBottom: 5 }}>
                <span className="ei-mono" style={{ fontSize: 11, color: T.ink3 }}>Q{idx + 1}</span>
                {(() => {
                  const b = getEntrySourceBadge(entry.source, lang);
                  return <Tag tone={b.tone} T={T}>{b.icon} {b.label}</Tag>;
                })()}
              </div>
              <div style={{ fontSize: 13, color: T.ink, lineHeight: 1.45 }}>{entry.stage}</div>
            </button>
          ))}
        </Card>

        {activeEntry && (
          <Card T={T} pad={18}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12, marginBottom: 12 }}>
              <div>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{activeEntry.stage} · {activeEntry.tag}</div>
                <div className="ei-serif" style={{ fontSize: 19, color: T.ink, lineHeight: 1.4 }}>{activeEntry.q}</div>
              </div>
              <div style={{ display: "flex", gap: 6, flexWrap: "wrap", justifyContent: "flex-end" }}>
                {Object.keys(reactions).map((r) => (
                  <button key={r} onClick={() => setEntries(entries.map(x => x.id === activeEntry.id ? { ...x, reaction: r } : x))} style={{
                    padding: "4px 8px", fontSize: 10.5, borderRadius: 2, cursor: "pointer",
                    border: `1px solid ${activeEntry.reaction === r ? T.ink2 : T.rule}`,
                    background: activeEntry.reaction === r ? T.bgSoft : "transparent",
                    color: activeEntry.reaction === r ? T.ink : T.ink3, fontFamily: "var(--ei-mono)",
                  }}>{reactions[r]}</button>
                ))}
              </div>
            </div>
            {[
              { k: "a", label: lang === "en" ? "What I answered" : "我当时怎么回答", value: activeEntry.a, rows: 3 },
              { k: "follow", label: lang === "en" ? "Follow-up / interviewer reaction" : "追问 / 面试官反应", value: activeEntry.follow, rows: 2 },
              { k: "reflection", label: lang === "en" ? "What I missed or want to preserve" : "遗漏点 / 需要保留的信息", value: activeEntry.reflection, rows: 2 },
            ].map((field) => (
              <div key={field.k} style={{ marginTop: 10 }}>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 5 }}>{field.label}</div>
                <textarea defaultValue={field.value} rows={field.rows} style={{ width: "100%", padding: "10px 12px", fontSize: 13.5, color: T.ink2, background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2, resize: "vertical", lineHeight: 1.5, outline: "none", boxSizing: "border-box", fontFamily: "var(--ei-sans)" }} />
              </div>
            ))}
          </Card>
        )}
      </div>
    </div>
  );
};

// ─── Conversational debrief mock data ────────────────────────────────────────
const DEBRIEF_AI_QUESTIONS = (lang) => lang === "en" ? [
  "Let's debrief this interview. How did it go overall — any moments where you felt stuck or pushed hard?",
  "Got it. How did it open — self-intro first, or straight into technical questions?",
  "Any other questions you felt under-prepared for, or where you got pushed back?",
  "Last one — how did the reverse-Q and close go? What did you ask, and how did they react?",
] : [
  "好，我们来复盘这场面试。整体感觉怎么样？有没有明显卡住或被追问的地方？",
  "了解。面试开场是先自我介绍，还是直接进技术题？",
  "还有没有其他你觉得答得不好、或者被追问到的地方？",
  "最后反问那块怎么样？你问了什么，对方有什么反应？",
];

const DEBRIEF_USER_RESPONSES = (lang) => lang === "en" ? [
  "Pretty good overall. The RSC / checkout work got probed hard on who made the final architecture call — I fumbled that part.",
  "Self-intro first, then straight into the checkout performance project.",
  "The Design System question went well — they nodded and moved on. That was mainly the one gap.",
  "Only asked about engineering culture. Pretty generic — should have asked about team priorities or their biggest pain point.",
] : [
  "整体还可以，就是 RSC 那块被追问谁做的架构决定，我有点混。",
  "先自我介绍，然后直接问了结账优化项目。",
  "Design System 那题答得还不错，对方点头就过去了，主要就 ownership 那个缺口。",
  "只问了工程文化，问题太泛了，应该问团队当前重点方向或者最头疼的事。",
];

const DEBRIEF_CONV_CARDS = (lang) => lang === "en" ? [
  { id: "vc1", q: "Walk me through the checkout perf work — what exactly did YOU drive?", summary: "Explained RSC migration, but ownership attribution was unclear under pressure.", followup: "They pushed: who made the final architecture call?", confidence: "high", afterTurn: 0 },
  { id: "vc2", q: "Tell me about yourself / walk me through your background.", summary: "Self-intro as opening, then moved directly into the checkout project.", followup: "No follow-up noted on intro itself.", confidence: "medium", afterTurn: 1 },
  { id: "vc3", q: "Any questions for me?", summary: "Only asked about engineering culture — too generic.", followup: "They answered politely and wrapped up.", confidence: "high", afterTurn: 3 },
] : [
  { id: "vc1", q: "跟我讲讲结账性能那个项目——具体哪些是你推动的？", summary: "讲了 RSC 迁移过程，但「谁做的最终架构决定」那里答得比较混。", followup: "追问：架构方案最后是谁拍板的？", confidence: "high", afterTurn: 0 },
  { id: "vc2", q: "先做个自我介绍，说说你的背景。", summary: "开场自我介绍，之后直接进入结账优化项目。", followup: "无直接追问记录。", confidence: "medium", afterTurn: 1 },
  { id: "vc3", q: "你有什么想问我的吗？", summary: "只问了工程文化，太泛，没问痛点或下一步预期。", followup: "对方礼貌回答后结束面试。", confidence: "high", afterTurn: 3 },
];

const VoiceDebriefRecord = ({ T, lang, entries, setEntries, initialState = "intro", active = true }) => {
  const aiQuestions = DEBRIEF_AI_QUESTIONS(lang);
  const userResponses = DEBRIEF_USER_RESPONSES(lang);
  const convCards = DEBRIEF_CONV_CARDS(lang);

  const seededChat = initialState === "chat";
  const seedMessages = seededChat
    ? [
        { id: "a0", role: "ai", text: aiQuestions[0], streaming: false },
        { id: "u0", role: "user", text: userResponses[0], streaming: false },
        { id: "a1", role: "ai", text: aiQuestions[1], streaming: true },
      ]
    : [];
  const seedCards = initialState === "review"
    ? convCards
    : seededChat
      ? convCards.filter((c) => c.afterTurn === 0)
      : [];

  const [phase, setPhase] = React.useState(
    initialState === "review" ? "review" : seededChat ? "chat" : "intro"
  );
  // voiceState: idle | ai_speaking | listening | user_speaking | processing | ended
  const [voiceState, setVoiceState] = React.useState(seededChat ? "ai_speaking" : "idle");
  const [paused, setPaused] = React.useState(false);
  const [messages, setMessages] = React.useState(seedMessages);
  const [turnIdx, setTurnIdx] = React.useState(seededChat ? 1 : 0);
  const [extractedCards, setExtractedCards] = React.useState(seedCards);
  const [cardStatus, setCardStatus] = React.useState(() => {
    const init = {};
    seedCards.forEach((c) => { init[c.id] = "pending"; });
    return init;
  });
  const [committedCount, setCommittedCount] = React.useState(0);
  const [elapsed, setElapsed] = React.useState(seededChat ? 38 : 0);
  // awaitingUser holds the conversation at "listening" without auto-advancing.
  // Used after "继续补充" so the AI doesn't fabricate a user turn when there's no real input.
  const [awaitingUser, setAwaitingUser] = React.useState(false);
  // per-card edit state in review phase (only one card editable at a time)
  const [editingId, setEditingId] = React.useState(null);
  const [editDraft, setEditDraft] = React.useState({ q: "", summary: "", followup: "" });

  // total elapsed timer — runs while in chat phase, halts on pause or when not active
  // (i.e. user has switched to text mode tab — voice convo persists but freezes)
  React.useEffect(() => {
    if (!active) return;
    if (phase !== "chat" || paused) return;
    const t = setInterval(() => setElapsed((e) => e + 1), 1000);
    return () => clearInterval(t);
  }, [phase, paused, active]);

  // continuous-conversation auto-advance state machine
  React.useEffect(() => {
    if (!active) return;             // user is on text tab — freeze convo, preserve state
    if (phase !== "chat" || paused) return;

    if (voiceState === "idle") {
      setVoiceState("ai_speaking");
      return;
    }
    if (voiceState === "ai_speaking") {
      const id = `a${turnIdx}`;
      let text = aiQuestions[turnIdx];
      // First AI message: if user already has entries (e.g. from text mode),
      // acknowledge them so the AI doesn't repeat-ask what's already captured.
      if (turnIdx === 0 && text && entries.length > 0) {
        const ctx = buildVoiceContextOpener(entries, lang);
        if (ctx) text = ctx + " " + text;
      }
      if (!text) {
        // nothing scripted; in continuation mode the AI prompt was added externally
        // and its own timeout flips us to listening — just hold here.
        if (awaitingUser) return;
        setVoiceState("listening");
        return;
      }
      setMessages((prev) => prev.some((m) => m.id === id) ? prev : [...prev, { id, role: "ai", text, streaming: true }]);
      const dur = Math.max(2400, text.length * 65);
      const t = setTimeout(() => {
        setMessages((prev) => prev.map((m) => (m.id === id ? { ...m, streaming: false } : m)));
        setVoiceState("listening");
      }, dur);
      return () => clearTimeout(t);
    }
    if (voiceState === "listening") {
      // continuation mode: hold here — wait for the user to actually speak (or end)
      if (awaitingUser) return;
      const t = setTimeout(() => setVoiceState("user_speaking"), 800);
      return () => clearTimeout(t);
    }
    if (voiceState === "user_speaking") {
      const id = `u${turnIdx}`;
      const text = userResponses[turnIdx] || (lang === "en" ? "(continuing the conversation…)" : "（继续口述…）");
      setMessages((prev) => prev.some((m) => m.id === id) ? prev : [...prev, { id, role: "user", text, streaming: true }]);
      const dur = Math.max(2600, text.length * 75);
      const t = setTimeout(() => {
        setMessages((prev) => prev.map((m) => (m.id === id ? { ...m, streaming: false } : m)));
        setVoiceState("processing");
      }, dur);
      return () => clearTimeout(t);
    }
    if (voiceState === "processing") {
      const card = convCards.find((c) => c.afterTurn === turnIdx);
      if (card) {
        setExtractedCards((prev) => prev.find((c) => c.id === card.id) ? prev : [...prev, card]);
        setCardStatus((prev) => prev[card.id] ? prev : { ...prev, [card.id]: "pending" });
      }
      const next = turnIdx + 1;
      const t = setTimeout(() => {
        if (next < aiQuestions.length) {
          setTurnIdx(next);
          setVoiceState("ai_speaking");
          return;
        }
        // past the scripted turn list — wrap if first time, then go to review
        const alreadyWrapped = messages.some((m) => m.id === "a-wrap");
        if (alreadyWrapped) {
          setPhase("review");
          return;
        }
        const remaining = convCards.filter((c) => c.afterTurn >= next);
        if (remaining.length) {
          setExtractedCards((prev) => {
            const adds = remaining.filter((c) => !prev.find((x) => x.id === c.id));
            return adds.length ? [...prev, ...adds] : prev;
          });
          setCardStatus((prev) => {
            const out = { ...prev };
            remaining.forEach((c) => { if (!out[c.id]) out[c.id] = "pending"; });
            return out;
          });
        }
        const wrap = lang === "en"
          ? "OK, I think I have enough. I pulled together a few question cards — let's review them."
          : "好，我大概记下来了。整理出几张卡片，我们一起看看。";
        setMessages((prev) => [...prev, { id: "a-wrap", role: "ai", text: wrap, streaming: true }]);
        setVoiceState("ended");
        setTimeout(() => {
          setMessages((prev) => prev.map((m) => (m.id === "a-wrap" ? { ...m, streaming: false } : m)));
          setPhase("review");
        }, 2400);
      }, 1100);
      return () => clearTimeout(t);
    }
  }, [phase, paused, voiceState, turnIdx, lang, awaitingUser, active]);

  // [space] toggles pause while in chat phase — only when voice tab is active
  // (otherwise pressing space in text-mode forms would unintentionally pause the convo)
  React.useEffect(() => {
    if (!active) return;
    if (phase !== "chat") return;
    const onKey = (e) => {
      if (e.code !== "Space") return;
      const tag = (e.target && e.target.tagName) || "";
      if (tag === "INPUT" || tag === "TEXTAREA" || (e.target && e.target.isContentEditable)) return;
      e.preventDefault();
      setPaused((p) => !p);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [phase, active]);

  const scrollRef = React.useRef(null);
  React.useEffect(() => {
    if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
  }, [messages, voiceState]);

  const fmtElapsed = (s) => `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;

  const startChat = () => {
    setPhase("chat");
    setVoiceState("ai_speaking");
    setMessages([]);
    setTurnIdx(0);
    setElapsed(0);
    setPaused(false);
    setAwaitingUser(false);
  };

  const endChatEarly = () => {
    if (extractedCards.length === 0) { setPhase("intro"); return; }
    setPhase("review");
  };

  const setStatus = (id, s) => setCardStatus((prev) => ({ ...prev, [id]: s }));

  const startEdit = (card) => {
    setEditDraft({ q: card.q, summary: card.summary, followup: card.followup });
    setEditingId(card.id);
  };
  const cancelEdit = () => {
    setEditingId(null);
    setEditDraft({ q: "", summary: "", followup: "" });
  };
  const saveEdit = () => {
    if (!editingId) return;
    const editingCard = extractedCards.find((c) => c.id === editingId);
    setExtractedCards((prev) => prev.map((c) => (
      c.id === editingId
        ? { ...c, q: editDraft.q.trim() || c.q, summary: editDraft.summary, followup: editDraft.followup }
        : c
    )));
    // Manual cards auto-confirm on first save (user explicitly typed them — there's
    // no AI extraction to verify). Voice-extracted cards still need an explicit confirm.
    if (editingCard && editingCard.source === "manual" && editDraft.q.trim()) {
      setCardStatus((prev) => ({ ...prev, [editingId]: "confirmed" }));
    }
    setEditingId(null);
    setEditDraft({ q: "", summary: "", followup: "" });
  };

  const commitCards = () => {
    const confirmed = extractedCards.filter((c) => cardStatus[c.id] === "confirmed");
    const newEntries = confirmed.map((c, i) => {
      const isManual = c.source === "manual";
      return {
        id: `vce${entries.length + i + 1}`,
        stage: isManual
          ? (lang === "en" ? "Manual add" : "手动添加")
          : (lang === "en" ? "Voice debrief" : "语音复盘"),
        q: c.q, a: c.summary, follow: c.followup, reflection: "",
        reaction: "neutral",
        source: isManual ? "manual" : "voice_extracted",
        tag: isManual ? (lang === "en" ? "manual" : "手动") : (lang === "en" ? "voice" : "语音复盘"),
      };
    });
    setEntries([...entries, ...newEntries]);
    setCommittedCount(confirmed.length);
    setPhase("committed");
  };

  // Manual add — for the "I want to type this one specifically" case while in voice mode.
  // Pauses the conversation, jumps to the review phase with a fresh empty card already in
  // edit mode so the user can fill it in. They can return to chat afterwards.
  const addManualCard = () => {
    const id = `manual-${Date.now()}`;
    const newCard = {
      id, q: "", summary: "", followup: "",
      confidence: "high", source: "manual", afterTurn: turnIdx,
    };
    setExtractedCards((prev) => [...prev, newCard]);
    setCardStatus((prev) => ({ ...prev, [id]: "pending" }));
    setPaused(true);
    setPhase("review");
    setEditDraft({ q: "", summary: "", followup: "" });
    setEditingId(id);
  };

  const continueDebrief = () => {
    const id = `a-cont-${Date.now()}`;
    const text = lang === "en"
      ? "We've saved those cards. Anything else worth adding — a question we missed, or a detail you remembered?"
      : "已保存这几张卡片。还有什么想补充的吗——漏掉的题目，或者刚想到的细节？";
    setMessages((prev) => [...prev, { id, role: "ai", text, streaming: true }]);
    // hold at listening — don't auto-fabricate a user turn; wait for real input or end
    setTurnIdx(aiQuestions.length);
    setVoiceState("ai_speaking");
    setPhase("chat");
    setPaused(false);
    setAwaitingUser(true);
    setTimeout(() => {
      setMessages((prev) => prev.map((m) => (m.id === id ? { ...m, streaming: false } : m)));
      setVoiceState("listening");
    }, 1800);
  };

  const confidenceColor = { high: T.ok, medium: T.warn, low: T.ink3 };
  const confidenceLabel = lang === "en"
    ? { high: "high confidence", medium: "medium", low: "low" }
    : { high: "高置信", medium: "中", low: "低" };

  const stateLabel = paused
    ? (lang === "en" ? "Paused" : "已暂停")
    : voiceState === "ai_speaking"
      ? (lang === "en" ? "AI speaking" : "AI 正在说")
      : voiceState === "listening"
        ? (awaitingUser ? (lang === "en" ? "Listening · waiting for you" : "聆听中 · 等你开口") : (lang === "en" ? "Listening…" : "聆听中"))
        : voiceState === "user_speaking"
          ? (lang === "en" ? "You're speaking" : "你正在说")
          : voiceState === "processing"
            ? (lang === "en" ? "AI thinking" : "AI 思考中")
            : voiceState === "ended"
              ? (lang === "en" ? "Wrapping up" : "正在收尾")
              : (lang === "en" ? "Ready" : "准备开始");

  const stateColor = paused
    ? T.ink3
    : voiceState === "user_speaking"
      ? T.ok
      : voiceState === "processing"
        ? T.warn
        : voiceState === "listening"
          ? T.ink3
          : T.accent;

  // ── intro ──────────────────────────────────────────────────────────────────
  if (phase === "intro") {
    const topics = lang === "en"
      ? ["Overall feel + stuck moments", "Interview opening", "Questions you didn't handle well", "Reverse-Q and close"]
      : ["整体感受和卡住的地方", "面试开场方式", "没答好或被追问的题", "反问和收尾"];
    return (
      <div style={{ display: "flex", flexDirection: "column", alignItems: "center", padding: "24px 0", gap: 20 }}>
        <div style={{ width: "100%", maxWidth: 520 }}>
          <Card T={T} pad={24}>
            <div style={{ display: "flex", gap: 14, alignItems: "flex-start", marginBottom: 20 }}>
              <div style={{ width: 36, height: 36, borderRadius: 18, background: T.accentSoft, border: `1.5px solid ${T.accent}`, flexShrink: 0, display: "flex", alignItems: "center", justifyContent: "center" }}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={T.accent} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
                  <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
                </svg>
              </div>
              <div>
                <div style={{ fontSize: 14, fontWeight: 500, color: T.ink, marginBottom: 6 }}>
                  {lang === "en" ? "AI Debrief — voice conversation" : "AI 语音复盘 · 连续对话"}
                </div>
                <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.7 }}>
                  {lang === "en"
                    ? "Talk naturally — like a friend asking you to walk through the interview. I'll listen, follow up, and extract question cards as we go. No tap-to-talk; pause anytime with [space]."
                    : "像和朋友复盘一样自然聊就好。我会一边听一边追问，过程中实时提取问题卡片。无需手动点录音，随时可按 [空格] 暂停。"}
                </div>
              </div>
            </div>
            <div style={{ padding: "14px 16px", background: T.bgSoft, borderRadius: 2, marginBottom: 20 }}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "TOPICS WE'LL COVER" : "会聊到的内容"}</div>
              <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                {topics.map((t, i) => (
                  <div key={i} style={{ display: "flex", gap: 10, alignItems: "center" }}>
                    <span style={{ fontFamily: "var(--ei-mono)", fontSize: 10.5, color: T.accent }}>{String(i + 1).padStart(2, "0")}</span>
                    <span style={{ fontSize: 13, color: T.ink2 }}>{t}</span>
                  </div>
                ))}
              </div>
            </div>
            <button onClick={startChat} style={{
              width: "100%", padding: "13px 20px", borderRadius: 2, border: `1.5px solid ${T.accent}`,
              background: T.accent, color: "#fff", cursor: "pointer", fontSize: 14, fontWeight: 500,
              fontFamily: "var(--ei-sans)", display: "flex", alignItems: "center", justifyContent: "center", gap: 8,
            }}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
                <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
              </svg>
              {lang === "en" ? "Start voice conversation" : "开始语音复盘对话"}
            </button>
            <div style={{ marginTop: 10, fontSize: 11, color: T.ink4, fontFamily: "var(--ei-mono)", textAlign: "center" }}>
              {lang === "en" ? "tip · press [space] anytime to pause" : "小提示 · 随时按 [空格] 暂停"}
            </div>
          </Card>
        </div>
      </div>
    );
  }

  // ── chat ───────────────────────────────────────────────────────────────────
  if (phase === "chat") {
    return (
      <div style={{ display: "grid", gridTemplateColumns: "1fr 220px", gap: 18 }}>
        <div style={{ display: "flex", flexDirection: "column", gap: 12, minHeight: 460 }}>
          {/* Status bar — current voice state + total elapsed */}
          <div style={{
            display: "flex", alignItems: "center", justifyContent: "space-between",
            padding: "10px 14px", background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2,
          }}>
            <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
              <div className={paused ? "" : "ei-pulse"} style={{
                width: 8, height: 8, borderRadius: 4, background: stateColor,
                boxShadow: paused ? "none" : `0 0 0 4px ${stateColor}22`,
              }} />
              <div style={{ fontSize: 12.5, color: T.ink2, fontWeight: 500 }}>{stateLabel}</div>
              {voiceState === "user_speaking" && !paused && (
                <div style={{ display: "flex", gap: 2, alignItems: "center", height: 14, marginLeft: 4 }}>
                  {Array.from({ length: 12 }).map((_, i) => (
                    <div key={i} className="ei-pulse" style={{
                      width: 2, background: T.ok, borderRadius: 1,
                      height: `${30 + Math.sin(i * 0.9) * 22}%`,
                      animationDelay: `${(i * 0.06).toFixed(2)}s`,
                      animationDuration: `${0.55 + (i % 3) * 0.1}s`,
                      opacity: 0.75,
                    }} />
                  ))}
                </div>
              )}
            </div>
            <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink3 }}>
              {fmtElapsed(elapsed)}
            </div>
          </div>

          {/* Conversation thread */}
          <div ref={scrollRef} style={{
            flex: 1, minHeight: 320, maxHeight: 420, overflowY: "auto",
            padding: "4px 2px",
            display: "flex", flexDirection: "column", gap: 12,
          }}>
            {messages.map((m) => (
              <div key={m.id} style={{ display: "flex", justifyContent: m.role === "user" ? "flex-end" : "flex-start" }}>
                {m.role === "ai" && (
                  <div style={{ maxWidth: "78%" }}>
                    <div style={{ fontSize: 10, fontFamily: "var(--ei-mono)", color: T.accent, marginBottom: 4 }}>
                      {lang === "en" ? "AI DEBRIEF GUIDE" : "AI 复盘引导"}
                    </div>
                    <div style={{
                      padding: "12px 15px", background: T.accentSoft, border: `1px solid ${T.accent}`,
                      borderRadius: "2px 12px 12px 12px", fontSize: 13.5, color: T.ink, lineHeight: 1.65,
                    }}>
                      {m.text}
                      {m.streaming && !paused && (
                        <span className="ei-pulse" style={{
                          display: "inline-block", width: 6, height: 13, marginLeft: 4, verticalAlign: "-2px",
                          background: T.accent, opacity: 0.7,
                        }} />
                      )}
                    </div>
                  </div>
                )}
                {m.role === "user" && (
                  <div style={{ maxWidth: "74%" }}>
                    <div style={{ fontSize: 10, fontFamily: "var(--ei-mono)", color: T.ink3, marginBottom: 4, textAlign: "right" }}>
                      {lang === "en" ? "YOU · live transcript" : "你 · 实时转写"}
                    </div>
                    <div style={{
                      padding: "12px 15px", background: T.bgCard, border: `1px solid ${T.rule}`,
                      borderRadius: "12px 2px 12px 12px", fontSize: 13.5, color: T.ink2, lineHeight: 1.65,
                    }}>
                      {m.text}
                      {m.streaming && !paused && (
                        <span className="ei-pulse" style={{
                          display: "inline-block", width: 6, height: 13, marginLeft: 4, verticalAlign: "-2px",
                          background: T.ink3, opacity: 0.7,
                        }} />
                      )}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* Voice control bar — pause/resume + end */}
          <div style={{
            padding: "12px 16px",
            background: paused ? "rgba(180,130,0,.06)" : T.bgSoft,
            border: `1px solid ${paused ? T.warn : T.rule}`,
            borderRadius: 2,
            display: "flex", alignItems: "center", justifyContent: "space-between", gap: 12,
          }}>
            <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
              <button
                onClick={() => setPaused((p) => !p)}
                aria-label={paused ? (lang === "en" ? "Resume conversation" : "继续对话") : (lang === "en" ? "Pause conversation" : "暂停对话")}
                style={{
                  width: 40, height: 40, borderRadius: 20,
                  border: `1.5px solid ${paused ? T.accent : T.rule}`,
                  background: paused ? T.accentSoft : "transparent",
                  cursor: "pointer", flexShrink: 0,
                  display: "flex", alignItems: "center", justifyContent: "center",
                }}
              >
                {paused ? (
                  <svg width="14" height="14" viewBox="0 0 24 24" fill={T.accent}><path d="M8 5v14l11-7z"/></svg>
                ) : (
                  <svg width="14" height="14" viewBox="0 0 24 24" fill={T.ink2}><rect x="6" y="5" width="4" height="14" rx="1"/><rect x="14" y="5" width="4" height="14" rx="1"/></svg>
                )}
              </button>
              <div>
                <div style={{ fontSize: 12.5, color: T.ink2, fontWeight: 500 }}>
                  {paused
                    ? (lang === "en" ? "Conversation paused" : "对话已暂停")
                    : awaitingUser
                      ? (lang === "en" ? "Waiting — speak anytime, or wrap up →" : "等你开口 · 也可以直接结束 →")
                      : (lang === "en" ? "Listening continuously" : "持续监听中")}
                </div>
                <div style={{ fontSize: 11, color: T.ink4, fontFamily: "var(--ei-mono)" }}>
                  {paused
                    ? (lang === "en" ? "tap ▶ or [space] to resume" : "点 ▶ 或按 [空格] 继续")
                    : (lang === "en" ? "[space] · pause   ·   no need to tap" : "[空格] 暂停 · 无需手动操作")}
                </div>
              </div>
            </div>
            <button onClick={endChatEarly} style={{
              padding: "11px 18px", borderRadius: 2,
              border: `1.5px solid ${T.accent}`,
              background: T.accent, color: "#fff",
              fontSize: 13, fontWeight: 500, cursor: "pointer",
              fontFamily: "var(--ei-sans)", flexShrink: 0,
              display: "inline-flex", alignItems: "center", gap: 8,
              boxShadow: awaitingUser ? `0 0 0 4px ${T.accentSoft}` : "none",
              transition: "box-shadow .2s",
            }}>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              {lang === "en" ? "End & review" : "结束 · 进入确认"}
            </button>
          </div>
        </div>

        {/* Live extraction panel */}
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>
            {lang === "en" ? "EXTRACTING" : "实时提取"}
            {extractedCards.length > 0 && (
              <span style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.accent, marginLeft: 6 }}>{extractedCards.length}</span>
            )}
          </div>
          {extractedCards.length === 0 ? (
            <div style={{ fontSize: 12, color: T.ink4, fontFamily: "var(--ei-mono)", lineHeight: 1.8 }}>
              {lang === "en" ? "Cards appear\nhere as we talk" : "对话过程中\n卡片会出现在这里"}
            </div>
          ) : (
            <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
              {extractedCards.map((c, i) => (
                <div key={c.id} className="ei-fadein" style={{ padding: "10px 12px", background: T.bgSoft, border: `1px solid ${T.rule}`, borderLeft: `2px solid ${c.source === "manual" ? T.ink3 : confidenceColor[c.confidence]}`, borderRadius: 2 }}>
                  <div style={{ fontFamily: "var(--ei-mono)", fontSize: 9.5, color: T.ink4, marginBottom: 4 }}>
                    Q{i + 1} · {c.source === "manual" ? (lang === "en" ? "manual" : "手动") : confidenceLabel[c.confidence]}
                  </div>
                  <div style={{ fontSize: 12, color: T.ink2, lineHeight: 1.5 }}>
                    {c.q ? (c.q.length > 55 ? c.q.slice(0, 52) + "…" : c.q) : (lang === "en" ? <span style={{ fontStyle: "italic", color: T.ink4 }}>(empty — fill in)</span> : <span style={{ fontStyle: "italic", color: T.ink4 }}>（空白 · 待填写）</span>)}
                  </div>
                </div>
              ))}
            </div>
          )}
          {/* Manual add — for the "I want to type this exact one" case */}
          <button onClick={addManualCard} style={{
            marginTop: 12, width: "100%", padding: "9px 12px", borderRadius: 2,
            border: `1px dashed ${T.rule}`, background: "transparent", color: T.ink3,
            cursor: "pointer", fontSize: 12, fontFamily: "var(--ei-sans)",
            display: "flex", alignItems: "center", justifyContent: "center", gap: 6,
            transition: "border-color .15s, color .15s",
          }}
            onMouseEnter={(e) => { e.currentTarget.style.borderColor = T.accent; e.currentTarget.style.color = T.accent; }}
            onMouseLeave={(e) => { e.currentTarget.style.borderColor = T.rule; e.currentTarget.style.color = T.ink3; }}
          >
            <span style={{ fontSize: 14, lineHeight: 1 }}>+</span>
            {lang === "en" ? "Add one manually" : "手动加一条"}
          </button>
          {extractedCards.length > 0 && (
            <div style={{ marginTop: 10, fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", lineHeight: 1.7 }}>
              {lang === "en" ? "You'll confirm\nthese at the end" : "复盘结束后\n你来确认这些卡片"}
            </div>
          )}
        </div>
      </div>
    );
  }

  // ── review ─────────────────────────────────────────────────────────────────
  if (phase === "review") {
    const totalPending = extractedCards.filter((c) => cardStatus[c.id] === "pending").length;
    const totalConfirmed = extractedCards.filter((c) => cardStatus[c.id] === "confirmed").length;
    const totalDeleted = extractedCards.filter((c) => cardStatus[c.id] === "deleted").length;
    return (
      <div>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14 }}>
          <div>
            <div className="ei-label" style={{ color: T.warn, marginBottom: 4 }}>
              {lang === "en" ? "PENDING CARDS · confirm before saving" : "待确认问题卡片 · 确认后写入复盘记录"}
            </div>
            <div style={{ fontSize: 13, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
              {lang === "en"
                ? `${totalPending} pending · ${totalConfirmed} confirmed · ${totalDeleted} deleted`
                : `${totalPending} 待确认 · ${totalConfirmed} 已确认 · ${totalDeleted} 已删除`}
            </div>
          </div>
          <div style={{ display: "flex", gap: 8 }}>
            <Btn T={T} variant="ghost" size="sm" onClick={() => setPhase("chat")} disabled={!!editingId}>
              {lang === "en" ? "Back to conversation" : "回到对话"}
            </Btn>
            <Btn T={T} variant="accent" size="sm" icon="check" onClick={commitCards}
              disabled={totalConfirmed === 0 || !!editingId} style={{ opacity: (totalConfirmed === 0 || !!editingId) ? 0.45 : 1 }}>
              {lang === "en" ? `Save ${totalConfirmed}` : `写入 ${totalConfirmed} 条`}
            </Btn>
          </div>
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          {extractedCards.map((card, idx) => {
            const status = cardStatus[card.id] || "pending";
            const isDeleted = status === "deleted";
            const isConfirmed = status === "confirmed";
            const isEditing = editingId === card.id;
            const editingDisabled = !!editingId && !isEditing;
            const inputStyle = {
              width: "100%", padding: "8px 10px", borderRadius: 2,
              border: `1px solid ${T.rule}`, background: T.bgCard,
              fontFamily: "var(--ei-sans)", fontSize: 13, color: T.ink,
              lineHeight: 1.5, outline: "none", resize: "vertical",
              boxSizing: "border-box",
            };
            return (
              <div key={card.id} style={{
                padding: "16px 18px",
                background: isEditing ? T.bgCard : isConfirmed ? T.okSoft : isDeleted ? T.bgSoft : T.bgCard,
                border: `1px solid ${isEditing ? T.accent : isConfirmed ? T.ok : isDeleted ? T.rule : T.warn}`,
                borderLeft: `3px solid ${isEditing ? T.accent : isConfirmed ? T.ok : isDeleted ? T.rule : T.warn}`,
                borderRadius: 2, opacity: isDeleted ? 0.5 : editingDisabled ? 0.6 : 1,
                transition: "all .18s",
              }}>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 14, marginBottom: isEditing ? 6 : 10 }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 7 }}>
                      <span style={{ fontFamily: "var(--ei-mono)", fontSize: 10.5, color: T.ink4 }}>Q{idx + 1}</span>
                      <span style={{ fontSize: 10, fontFamily: "var(--ei-mono)", padding: "2px 6px", borderRadius: 2, background: isEditing ? T.accentSoft : isConfirmed ? T.okSoft : isDeleted ? T.bgSoft : "rgba(180,130,0,.08)", color: isEditing ? T.accent : isConfirmed ? T.ok : isDeleted ? T.ink4 : T.warn, border: `1px solid ${isEditing ? T.accent : isConfirmed ? T.ok : isDeleted ? T.rule : T.warn}` }}>
                        {isEditing ? (lang === "en" ? "editing" : "编辑中") : isConfirmed ? (lang === "en" ? "confirmed" : "已确认") : isDeleted ? (lang === "en" ? "deleted" : "已删除") : (lang === "en" ? "pending" : "待确认")}
                      </span>
                      <span style={{ fontSize: 10.5, fontFamily: "var(--ei-mono)", color: confidenceColor[card.confidence] }}>{confidenceLabel[card.confidence]}</span>
                      <span style={{ fontSize: 10, fontFamily: "var(--ei-mono)", color: T.ink3, padding: "1px 6px", borderRadius: 2, background: T.bgSoft, border: `1px solid ${T.rule}` }}>
                        {card.source === "manual" ? (lang === "en" ? "✏️ manual" : "✏️ 手动") : (lang === "en" ? "🎙 voice" : "🎙 语音")}
                      </span>
                    </div>
                    {isEditing ? (
                      <textarea
                        value={editDraft.q}
                        onChange={(e) => setEditDraft((d) => ({ ...d, q: e.target.value }))}
                        rows={2}
                        style={{ ...inputStyle, fontFamily: "var(--ei-serif)", fontSize: 16, lineHeight: 1.4 }}
                        placeholder={lang === "en" ? "Question" : "问题"}
                      />
                    ) : (
                      <div className="ei-serif" style={{ fontSize: 17, color: T.ink, lineHeight: 1.4, marginBottom: 8 }}>{card.q}</div>
                    )}
                  </div>
                  {!isDeleted && !isEditing && (
                    <div style={{ display: "flex", gap: 6, flexShrink: 0 }}>
                      <button onClick={() => setStatus(card.id, isConfirmed ? "pending" : "confirmed")} disabled={editingDisabled} style={{ padding: "5px 12px", fontSize: 12, borderRadius: 2, cursor: editingDisabled ? "not-allowed" : "pointer", border: `1px solid ${isConfirmed ? T.ok : T.rule}`, background: isConfirmed ? T.okSoft : "transparent", color: isConfirmed ? T.ok : T.ink2, fontFamily: "var(--ei-sans)" }}>
                        {isConfirmed ? (lang === "en" ? "✓ confirmed" : "✓ 已确认") : (lang === "en" ? "Confirm" : "确认")}
                      </button>
                      <button onClick={() => startEdit(card)} disabled={editingDisabled} aria-label={lang === "en" ? "Edit card" : "编辑卡片"} style={{ padding: "5px 10px", fontSize: 12, borderRadius: 2, cursor: editingDisabled ? "not-allowed" : "pointer", border: `1px solid ${T.rule}`, background: "transparent", color: T.ink2, fontFamily: "var(--ei-sans)", display: "inline-flex", alignItems: "center", gap: 4 }}>
                        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4Z"/>
                        </svg>
                        {lang === "en" ? "Edit" : "编辑"}
                      </button>
                      <button onClick={() => setStatus(card.id, "deleted")} disabled={editingDisabled} style={{ padding: "5px 10px", fontSize: 12, borderRadius: 2, cursor: editingDisabled ? "not-allowed" : "pointer", border: `1px solid ${T.rule}`, background: "transparent", color: T.ink3, fontFamily: "var(--ei-sans)" }}>
                        {lang === "en" ? "Delete" : "删除"}
                      </button>
                    </div>
                  )}
                  {!isDeleted && isEditing && (
                    <div style={{ display: "flex", gap: 6, flexShrink: 0 }}>
                      <button onClick={saveEdit} style={{ padding: "5px 12px", fontSize: 12, borderRadius: 2, cursor: "pointer", border: `1.5px solid ${T.accent}`, background: T.accent, color: "#fff", fontFamily: "var(--ei-sans)", fontWeight: 500 }}>
                        {lang === "en" ? "Save" : "保存"}
                      </button>
                      <button onClick={cancelEdit} style={{ padding: "5px 10px", fontSize: 12, borderRadius: 2, cursor: "pointer", border: `1px solid ${T.rule}`, background: "transparent", color: T.ink3, fontFamily: "var(--ei-sans)" }}>
                        {lang === "en" ? "Cancel" : "取消"}
                      </button>
                    </div>
                  )}
                  {isDeleted && (
                    <button onClick={() => setStatus(card.id, "pending")} style={{ padding: "5px 10px", fontSize: 12, borderRadius: 2, cursor: "pointer", border: `1px solid ${T.rule}`, background: "transparent", color: T.ink2, fontFamily: "var(--ei-sans)" }}>
                      {lang === "en" ? "Restore" : "恢复"}
                    </button>
                  )}
                </div>
                {!isDeleted && !isEditing && (
                  <div style={{ display: "flex", flexDirection: "column", gap: 7 }}>
                    <div style={{ fontSize: 12.5, color: T.ink2, lineHeight: 1.55 }}>
                      <span style={{ fontFamily: "var(--ei-mono)", fontSize: 10.5, color: T.ink4, marginRight: 6 }}>{lang === "en" ? "ANSWER" : "回答摘要"}</span>
                      {card.summary}
                    </div>
                    <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.55 }}>
                      <span style={{ fontFamily: "var(--ei-mono)", fontSize: 10.5, color: T.ink4, marginRight: 6 }}>{lang === "en" ? "FOLLOW-UP" : "追问/反应"}</span>
                      {card.followup}
                    </div>
                  </div>
                )}
                {isEditing && (
                  <div style={{ display: "flex", flexDirection: "column", gap: 10, marginTop: 10 }}>
                    <div>
                      <div className="ei-label" style={{ color: T.ink4, marginBottom: 5 }}>{lang === "en" ? "ANSWER" : "回答摘要"}</div>
                      <textarea
                        value={editDraft.summary}
                        onChange={(e) => setEditDraft((d) => ({ ...d, summary: e.target.value }))}
                        rows={3}
                        style={inputStyle}
                        placeholder={lang === "en" ? "Summary of how you answered" : "你当时怎么答的"}
                      />
                    </div>
                    <div>
                      <div className="ei-label" style={{ color: T.ink4, marginBottom: 5 }}>{lang === "en" ? "FOLLOW-UP" : "追问/反应"}</div>
                      <textarea
                        value={editDraft.followup}
                        onChange={(e) => setEditDraft((d) => ({ ...d, followup: e.target.value }))}
                        rows={2}
                        style={inputStyle}
                        placeholder={lang === "en" ? "Interviewer follow-up or reaction" : "追问或对方反应"}
                      />
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    );
  }

  // ── committed ──────────────────────────────────────────────────────────────
  return (
    <div style={{ padding: "28px 0", display: "flex", flexDirection: "column", alignItems: "center", gap: 16 }}>
      <div style={{ width: "100%", maxWidth: 480 }}>
        <div style={{ padding: "22px 24px", background: T.okSoft, border: `1px solid ${T.ok}`, borderRadius: 2, marginBottom: 14 }}>
          <div style={{ display: "flex", gap: 10, alignItems: "center", marginBottom: 8 }}>
            <Icon name="check" size={16} color={T.ok} stroke={2.5} />
            <div style={{ fontSize: 14, fontWeight: 500, color: T.ok }}>
              {lang === "en" ? `${committedCount} question cards saved to debrief record` : `${committedCount} 个问题卡片已写入复盘记录`}
            </div>
          </div>
          <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.65 }}>
            {lang === "en"
              ? "They're now part of step 0. Continue the conversation to capture more, or move to analysis."
              : "已进入第一步记录。你可以继续对话补充更多，也可以直接进入复盘分析。"}
          </div>
        </div>
        <button onClick={continueDebrief} style={{
          width: "100%", padding: "12px 20px", borderRadius: 2,
          border: `1.5px solid ${T.rule}`, background: T.bgCard,
          color: T.ink2, cursor: "pointer", fontSize: 13.5, fontFamily: "var(--ei-sans)",
          display: "flex", alignItems: "center", justifyContent: "center", gap: 8,
        }}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
            <path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
          </svg>
          {lang === "en" ? "Continue adding — keep the conversation going" : "继续补充 · 接着复盘对话"}
        </button>
      </div>
    </div>
  );
};

const DebriefReplayPlan = ({ T, lang, nav, back, entries, context }) => (
  <div style={{ display: "grid", gridTemplateColumns: "1fr 340px", gap: 28 }}>
    <div>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "DEBRIEF INTERVIEW CONTENT" : "复盘面试内容"}</div>
      <div style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, padding: "22px 24px", marginBottom: 18 }}>
        <div className="ei-serif" style={{ fontSize: 22, color: T.ink, marginBottom: 10 }}>
          {lang === "en" ? "Use interview questions first, AI probes second." : "先复现面试问题，再由 AI 追问薄弱处。"}
        </div>
        <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.7 }}>
          {lang === "en" ? "The replay interview will prioritize the questions you met, then add adjacent probes from the JD and resume evidence." : "复盘面试会优先使用你遇到的问题，再基于 JD 和简历证据补上相邻追问。"}
        </div>
      </div>
      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {entries.map((e, idx) => (
          <div key={e.id} style={{ padding: "16px 18px", background: T.bgCard, border: `1px solid ${T.rule}`, borderLeft: `3px solid ${idx < 2 ? T.accent : T.rule}`, borderRadius: 2 }}>
            <div style={{ display: "flex", justifyContent: "space-between", gap: 12, marginBottom: 8 }}>
              <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink3 }}>REAL Q{idx + 1}</div>
              <Tag tone={idx < 2 ? "accent" : "muted"} T={T}>{idx < 2 ? (lang === "en" ? "Core replay" : "核心复现") : (lang === "en" ? "Context" : "上下文")}</Tag>
            </div>
            <div className="ei-serif" style={{ fontSize: 17, color: T.ink, lineHeight: 1.45 }}>{e.q}</div>
            <div style={{ fontSize: 13, color: T.ink3, marginTop: 8, lineHeight: 1.55 }}>{e.a}</div>
          </div>
        ))}
      </div>
      <div style={{ display: "flex", justifyContent: "space-between", marginTop: 24 }}>
        <Btn T={T} variant="ghost" onClick={back}>{lang === "en" ? "Back" : "上一步"}</Btn>
        <Btn T={T} variant="accent" icon="play" onClick={() => nav("practice", window.eiCreateInterviewContext ? window.eiCreateInterviewContext({ mode: "text", modality: "text", sessionId: "session-debrief-replay" }) : { mode: "text" })}>{lang === "en" ? "Start debrief interview" : "开始复盘面试"}</Btn>
      </div>
    </div>
    <Card T={T} pad={18} style={{ height: "fit-content", position: "sticky", top: 20 }}>
      <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>{lang === "en" ? "INTERVIEW CONTEXT" : "复盘面试上下文"}</div>
      <div style={{ paddingBottom: 12, borderBottom: `1px dotted ${T.rule}`, marginBottom: 8 }}>
        <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>{context.target}</div>
        <div style={{ fontSize: 12, color: T.ink3, marginTop: 3 }}>{context.resume}</div>
      </div>
      {[
        lang === "en" ? "Ask the recorded questions in the original order where possible." : "尽可能按原顺序重新问一遍。",
        lang === "en" ? "Probe the weak ownership and reverse-question moments." : "重点追问 Ownership 与反问薄弱点。",
        lang === "en" ? "Compare this replay answer with the recorded interview answer." : "把复盘回答与面试记录做对照。",
      ].map((text, i) => (
        <div key={i} style={{ display: "flex", gap: 10, padding: "10px 0", borderBottom: i < 2 ? `1px dotted ${T.rule}` : "none", fontSize: 13.5, color: T.ink2, lineHeight: 1.55 }}>
          <span style={{ fontFamily: "var(--ei-mono)", color: T.accent }}>{String(i + 1).padStart(2, "0")}</span>
          <span>{text}</span>
        </div>
      ))}
    </Card>
  </div>
);


// ═══════════════════════════════════════════════════════════════════
// #12 RESUME VERSIONS + BULLET DIFF
// ═══════════════════════════════════════════════════════════════════
const ResumeVersionsScreen = ({ T, lang, nav, params = {} }) => {
  const [active, setActive] = React.useState("v2");
  const [selectedBullet, setSelectedBullet] = React.useState("b1");
  const [flow, setFlow] = React.useState(params.flow === "create" ? "create" : "versions");
  const [createMode, setCreateMode] = React.useState("upload");
  const [guideStep, setGuideStep] = React.useState(0);
  const [resumeText, setResumeText] = React.useState("");
  const [sourcePreviewOpen, setSourcePreviewOpen] = React.useState(false);

  React.useEffect(() => {
    if (params.flow === "create") setFlow("create");
  }, [params.flow]);

  const versions = [
    { id: "master", name: lang === "en" ? "Master" : "主版本", tag: "MASTER", date: "2026 · 04 · 18", target: lang === "en" ? "General" : "通用", bullets: 24, tone: "neutral", sourceId: "source-upload" },
    { id: "v1", name: lang === "en" ? "Star-Ring · Senior FE" : "星环 · 资深前端", tag: "TARGETED", date: "2026 · 04 · 20", target: "星环科技", bullets: 22, tone: "accent", active: true, sourceId: "source-upload" },
    { id: "v2", name: lang === "en" ? "Lumen · Platform" : "Lumen · 平台", tag: "TARGETED", date: "2026 · 04 · 19", target: "Lumen Labs", bullets: 18, tone: "accent", sourceId: "source-english" },
    { id: "draft", name: lang === "en" ? "Cloud-Yun · draft" : "云栖 · 草稿", tag: "DRAFT", date: "2026 · 04 · 21", target: "云栖集团", bullets: 20, tone: "muted", sourceId: "source-guided" },
  ];
  const originalSources = [
    {
      id: "source-upload",
      name: lang === "en" ? "Liu-Zhe-Frontend-2026.pdf" : "刘哲-前端-2026.pdf",
      type: lang === "en" ? "Uploaded PDF" : "上传 PDF",
      createdAt: "2026-04-18 21:32",
      retained: lang === "en" ? "Original file + parsed text retained" : "保留原始文件 + 解析文本",
      owner: lang === "en" ? "Master v3" : "主版本 v3",
      text: [
        lang === "en" ? "Liu Zhe · Senior Frontend Engineer · Shanghai" : "刘哲 · 资深前端工程师 · 上海",
        lang === "en" ? "Neoshop · Senior Frontend · 2022-now" : "Neoshop · 资深前端 · 2022 至今",
        lang === "en" ? "Worked on checkout performance improvements and complex admin surfaces." : "负责结账流程性能改进和复杂后台系统建设。",
        lang === "en" ? "Built shared UI components for internal products." : "为内部产品建设通用 UI 组件。",
      ],
    },
    {
      id: "source-english",
      name: "Liu-Zhe-Frontend-Platform-EN-v1.docx",
      type: lang === "en" ? "Uploaded DOCX" : "上传 DOCX",
      createdAt: "2026-04-15 10:18",
      retained: lang === "en" ? "Original English file + parsed sections retained" : "保留英文原件 + 解析结构",
      owner: lang === "en" ? "English platform version" : "英文平台版",
      text: [
        "Liu Zhe · Frontend Platform Engineer",
        "Built platform tooling, design system infrastructure, and TypeScript foundations.",
        "Worked with distributed teams across APAC and US time zones.",
      ],
    },
    {
      id: "source-guided",
      name: lang === "en" ? "Guided resume draft · 5 answers" : "问答生成草稿 · 5 个回答",
      type: lang === "en" ? "Guided Q&A" : "轻量问答",
      createdAt: "2026-04-21 09:06",
      retained: lang === "en" ? "Original answers + generated v1 retained" : "保留原始回答 + 生成 v1",
      owner: lang === "en" ? "Cloud-Yun draft" : "云栖草稿",
      text: [
        lang === "en" ? "Most recent role: senior frontend engineer." : "最近职位：资深前端工程师。",
        lang === "en" ? "Main direction: web architecture, design system rollout, frontend performance." : "主要方向：Web 架构、Design System 落地、前端性能。",
        lang === "en" ? "Target role: technical expert / web architect." : "目标岗位：技术专家 / Web 架构。",
      ],
    },
  ];
  const activeVersion = versions.find((v) => v.id === active) || versions[0];
  const activeSource = originalSources.find((s) => s.id === activeVersion.sourceId) || originalSources[0];

  const bullets = [
    {
      id: "b1", section: lang === "en" ? "Senior Frontend · Star-Ring · 2022-now" : "资深前端 · 星环科技 · 2022 至今",
      original: lang === "en" ? "Worked on checkout performance improvements for the e-commerce team, collaborating closely with backend engineers." : "负责电商团队结账流程的性能改进工作，与后端工程师紧密协作。",
      rewritten: lang === "en" ? "Led migration of the checkout surface to RSC + selective hydration, cutting LCP from 3.2s to 1.4s and lifting quarterly GMV by 1.8M (8% → 4.2% abandon)." : "主导结账链路迁移到 RSC + 选择性注水，LCP 3.2s → 1.4s，流失率 8% → 4.2%，季度 GMV +180 万。",
      why: lang === "en" ? ["Weak → strong ownership verb", "Adds quantified impact", "Names the specific architecture"] : ["动词从弱到强：「负责」→「主导」", "加入量化影响", "具体指出架构选择"],
      status: "pending",
    },
    {
      id: "b2", section: lang === "en" ? "Senior Frontend · Star-Ring · 2022-now" : "资深前端 · 星环科技 · 2022 至今",
      original: lang === "en" ? "Rolled out a design system across multiple product teams." : "在多个产品团队推广了设计系统。",
      rewritten: lang === "en" ? "Drove Design System v1 adoption across 5 products in 6 months (4 live, 1 in progress) — ran 3 workshops, paired migrations with 2 pilot teams, reduced new-dev ramp ~50%." : "6 个月内推动 Design System v1 在 5 个产品落地（4 上线、1 进行中）——办 3 次推广会、与 2 个试点团队结对迁移，新人上手时间缩短约 50%。",
      why: lang === "en" ? ["Names the scale (5 products)", "Shows method, not just outcome", "Anchored on developer time saved"] : ["量化范围：5 个产品", "讲方法而不只是结果", "以节省的工时收口"],
      status: "accepted",
    },
    {
      id: "b3", section: lang === "en" ? "Frontend · Lumen · 2019-2022" : "前端 · Lumen · 2019-2022",
      original: lang === "en" ? "Built and shipped various features for the core product." : "为核心产品构建并交付了多个功能。",
      rewritten: lang === "en" ? "Shipped 14 features to the order-management core over 3 years, including a batch-edit surface that became the #2 most-used power-user flow." : "3 年内为订单管理核心交付 14 个功能，其中批量编辑成为重度用户第 2 常用流程。",
      why: lang === "en" ? ["Vague → specific count", "Picks one feature worth name-checking", "Usage data gives credibility"] : ["模糊数量变具体", "挑一个值得点名的功能", "用使用数据建立可信度"],
      status: "pending",
    },
    {
      id: "b4", section: lang === "en" ? "Frontend · Lumen · 2019-2022" : "前端 · Lumen · 2019-2022",
      original: lang === "en" ? "Participated in code reviews and technical discussions." : "参与代码评审和技术讨论。",
      rewritten: lang === "en" ? "(remove — generic, dilutes stronger bullets)" : "（建议删除——太泛，会稀释其它更强的 bullet）",
      why: lang === "en" ? ["Every senior does this", "Takes space from quantifiable wins", "Better implied than stated"] : ["资深都做这个", "占用了可以量化的空间", "隐含就好，别直说"],
      status: "rejected",
    },
  ];

  const sel = bullets.find((b) => b.id === selectedBullet);
  const accepted = bullets.filter((b) => b.status === "accepted").length;
  const pending = bullets.filter((b) => b.status === "pending").length;
  const guideSteps = lang === "en" ? [
    { k: "Last role", q: "Where did you work most recently, and what was your title?", ph: "Company, title, dates…" },
    { k: "Direction", q: "What product or engineering direction did you mainly own?", ph: "Frontend platform, growth, infra, data…" },
    { k: "Proof project", q: "Pick one project that proves your level.", ph: "Problem, action, result…" },
    { k: "Numbers", q: "What measurable result can we attach to that project?", ph: "Performance, revenue, adoption, efficiency…" },
    { k: "Target", q: "What roles are you preparing for now?", ph: "Senior frontend, staff platform, AI infra…" },
  ] : [
    { k: "最近经历", q: "你最近在哪家公司、担任什么职位？", ph: "公司、职位、时间段…" },
    { k: "主要方向", q: "你主要负责什么产品或技术方向？", ph: "前端平台、增长、基础设施、数据、AI 应用…" },
    { k: "代表项目", q: "选一个最能证明你能力的项目。", ph: "背景、你做了什么、最后结果…" },
    { k: "量化结果", q: "这个项目能补哪些数字或可验证结果？", ph: "性能、收入、采用率、效率、成本…" },
    { k: "目标岗位", q: "你现在想准备什么类型的岗位？", ph: "资深前端、平台工程、AI 应用工程…" },
  ];

  if (flow === "create") {
    return (
      <div className="ei-fadein" style={{ maxWidth: 1220, margin: "0 auto", padding: "40px 48px 96px" }}>
        <button onClick={() => setFlow("versions")} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 22, display: "flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
          <Icon name="arrow_left" size={14} /> {lang === "en" ? "Back to resume workshop" : "返回简历工坊"}
        </button>

        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", gap: 28, marginBottom: 26, flexWrap: "wrap" }}>
          <div>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{lang === "en" ? "FIRST RESUME VERSION" : "创建第一版简历"}</div>
            <h1 className="ei-serif" style={{ fontSize: 38, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.15 }}>
              {lang === "en" ? "Start from a file, pasted text, or a five-minute guided draft." : "上传、粘贴，或用 5 分钟问答生成第一版。"}
            </h1>
            <div style={{ fontSize: 14, color: T.ink2, marginTop: 10, maxWidth: 720, lineHeight: 1.6 }}>
              {lang === "en" ? "We keep the original source, parse it into a structured resume, and save both as a version you can revise later." : "系统会保留原始文件或原始文本，同时解析成结构化简历，并作为可回溯版本保存。"}
            </div>
          </div>
          <Btn T={T} variant="secondary" size="sm" icon="briefcase" onClick={() => nav("workspace")}>{lang === "en" ? "Use in current job" : "用于当前岗位"}</Btn>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 340px", gap: 22, alignItems: "start" }}>
          <Card T={T} pad={0}>
            <div style={{ display: "flex", borderBottom: `1px solid ${T.rule}`, overflowX: "auto" }}>
              {[
                { k: "upload", icon: "upload", zh: "上传文件", en: "Upload" },
                { k: "paste", icon: "file", zh: "粘贴内容", en: "Paste" },
                { k: "guided", icon: "chat", zh: "轻量问答", en: "Guided" },
              ].map((mode) => (
                <button key={mode.k} onClick={() => setCreateMode(mode.k)} style={{
                  padding: "15px 20px", background: createMode === mode.k ? T.bgSoft : "transparent", border: "none",
                  borderBottom: `2px solid ${createMode === mode.k ? T.accent : "transparent"}`, color: createMode === mode.k ? T.ink : T.ink3,
                  display: "flex", alignItems: "center", gap: 8, cursor: "pointer", fontFamily: "var(--ei-sans)", marginBottom: -1,
                }}>
                  <Icon name={mode.icon} size={14} /> {lang === "en" ? mode.en : mode.zh}
                </button>
              ))}
            </div>

            {createMode === "upload" && (
              <div style={{ padding: 24 }}>
                <div style={{
                  minHeight: 260, border: `1px dashed ${T.ink4}`, borderRadius: 3, background: T.bgSoft,
                  display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", gap: 14, textAlign: "center", padding: 28,
                }}>
                  <div style={{ width: 54, height: 54, borderRadius: 27, background: T.bgCard, border: `1px solid ${T.rule}`, color: T.accent, display: "flex", alignItems: "center", justifyContent: "center" }}>
                    <Icon name="upload" size={24} />
                  </div>
                  <div className="ei-serif" style={{ fontSize: 22, color: T.ink }}>{lang === "en" ? "Drop a PDF / DOCX / Markdown resume" : "拖入 PDF / DOCX / Markdown 简历"}</div>
                  <div style={{ fontSize: 13, color: T.ink3, maxWidth: 460, lineHeight: 1.55 }}>
                    {lang === "en" ? "The source file is stored as the original version. Parsed sections become your editable structured resume." : "原始文件会作为原始版本保存，解析出的工作经历、项目、技能和教育经历会进入可编辑结构化简历。"}
                  </div>
                  <Btn T={T} variant="accent" icon="upload" onClick={() => {
                    const input = document.createElement("input");
                    input.type = "file";
                    input.accept = ".pdf,.docx,.md,.txt";
                    input.onchange = (e) => {
                      const f = e.target.files && e.target.files[0];
                      if (f) {
                        window.eiToast && window.eiToast(lang === "en" ? `Picked ${f.name} · parsing in prototype is mocked` : `已选择 ${f.name} · 原型仅模拟解析`, { tone: "ok", duration: 2800 });
                      }
                    };
                    input.click();
                  }}>{lang === "en" ? "Choose file" : "选择文件"}</Btn>
                </div>
              </div>
            )}

            {createMode === "paste" && (
              <div style={{ padding: 24 }}>
                <textarea
                  value={resumeText}
                  onChange={(e) => setResumeText(e.target.value)}
                  placeholder={lang === "en" ? "Paste your resume text here…" : "把你的简历内容粘贴到这里…"}
                  style={{
                    width: "100%", minHeight: 260, resize: "vertical", border: `1px solid ${T.rule}`, borderRadius: 2,
                    padding: 16, background: T.bg, color: T.ink, fontSize: 14, lineHeight: 1.65, outline: "none",
                  }}
                />
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: 14, gap: 14, flexWrap: "wrap" }}>
                  <div style={{ fontSize: 12.5, color: T.ink3 }}>
                    {lang === "en" ? "Raw text is retained, then parsed into structured sections." : "原始文本会保留，并解析成结构化段落。"}
                  </div>
                  <Btn T={T} variant="accent" icon="sparkle" disabled={!resumeText.trim()}>{lang === "en" ? "Parse and save v1" : "解析并保存 v1"}</Btn>
                </div>
              </div>
            )}

            {createMode === "guided" && (
              <div style={{ padding: 24 }}>
                <div style={{ display: "grid", gridTemplateColumns: "220px 1fr", gap: 22 }}>
                  <div style={{ borderRight: `1px solid ${T.rule}`, paddingRight: 18 }}>
                    {guideSteps.map((s, i) => (
                      <button key={s.k} onClick={() => setGuideStep(i)} style={{
                        width: "100%", padding: "11px 0", background: "transparent", border: "none", textAlign: "left",
                        display: "flex", gap: 10, alignItems: "center", color: guideStep === i ? T.ink : T.ink3, cursor: "pointer",
                      }}>
                        <span style={{ width: 22, height: 22, borderRadius: 11, background: guideStep === i ? T.accent : T.bgSoft, color: guideStep === i ? "#fff" : T.ink3, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontFamily: "var(--ei-mono)" }}>{i + 1}</span>
                        <span style={{ fontSize: 13 }}>{s.k}</span>
                      </button>
                    ))}
                  </div>
                  <div>
                    <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>{lang === "en" ? "GUIDED DRAFT" : "轻量问答建档"}</div>
                    <div className="ei-serif" style={{ fontSize: 24, color: T.ink, lineHeight: 1.35, marginBottom: 16 }}>{guideSteps[guideStep].q}</div>
                    <textarea
                      placeholder={guideSteps[guideStep].ph}
                      style={{ width: "100%", minHeight: 150, border: `1px solid ${T.rule}`, borderRadius: 2, padding: 14, background: T.bg, color: T.ink, fontSize: 14, lineHeight: 1.6, resize: "vertical", outline: "none" }}
                    />
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: 14 }}>
                      <div style={{ fontSize: 12.5, color: T.ink3 }}>
                        {lang === "en" ? "Answer only the parts you know. The draft can be refined later." : "只回答你现在记得的部分，生成后还可以继续补充。"}
                      </div>
                      <div style={{ display: "flex", gap: 8 }}>
                        <Btn T={T} variant="ghost" size="sm" onClick={() => setGuideStep(Math.max(0, guideStep - 1))}>{lang === "en" ? "Back" : "上一步"}</Btn>
                        <Btn T={T} variant="accent" size="sm" iconRight="arrow_right" onClick={() => guideStep < guideSteps.length - 1 ? setGuideStep(guideStep + 1) : setFlow("versions")}>
                          {guideStep < guideSteps.length - 1 ? (lang === "en" ? "Next" : "下一步") : (lang === "en" ? "Generate v1" : "生成 v1")}
                        </Btn>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </Card>

          <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
            <Card T={T}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "WHAT GETS SAVED" : "会保存什么"}</div>
              {[
                { icon: "file", title: lang === "en" ? "Original source" : "原始版本", body: lang === "en" ? "File, pasted text, or guided answers stay traceable." : "文件、粘贴文本或问答记录都会保留来源。" },
                { icon: "resume", title: lang === "en" ? "Structured resume" : "结构化简历", body: lang === "en" ? "Work, projects, skills, education, and evidence are editable." : "工作经历、项目、技能、教育和证据点可编辑。" },
                { icon: "layers", title: lang === "en" ? "Version baseline" : "版本基线", body: lang === "en" ? "Future JD-specific resumes branch from this v1." : "未来针对不同 JD 的版本从 v1 分叉。" },
              ].map((item, i) => (
                <div key={item.title} style={{ display: "grid", gridTemplateColumns: "26px 1fr", gap: 10, padding: "12px 0", borderBottom: i < 2 ? `1px dotted ${T.rule}` : "none" }}>
                  <Icon name={item.icon} size={15} color={T.accent} style={{ marginTop: 2 }} />
                  <div>
                    <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>{item.title}</div>
                    <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.5, marginTop: 2 }}>{item.body}</div>
                  </div>
                </div>
              ))}
            </Card>
            <Card T={T}>
              <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>{lang === "en" ? "AFTER V1" : "生成之后"}</div>
              <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.65 }}>
                {lang === "en" ? "You can match it against a JD, rewrite bullets, or start a mock interview with this resume as context." : "你可以把它和 JD 做匹配、改写 bullet，或直接作为模拟面试上下文。"}
              </div>
              <div style={{ display: "flex", gap: 8, marginTop: 14, flexWrap: "wrap" }}>
                <Btn T={T} variant="secondary" size="sm" icon="search" onClick={() => nav("jd_match")}>{lang === "en" ? "Job picks" : "岗位推荐"}</Btn>
                <Btn T={T} variant="secondary" size="sm" icon="play" onClick={() => nav("practice", window.eiCreateInterviewContext ? window.eiCreateInterviewContext({ mode: "text", modality: "text", sessionId: "session-resume-preview" }) : { mode: "text" })}>{lang === "en" ? "Mock interview" : "开始面试"}</Btn>
              </div>
            </Card>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="ei-fadein" style={{ maxWidth: 1320, margin: "0 auto", padding: "40px 48px 96px" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginBottom: 28, gap: 32 }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "RESUME WORKSHOP · VERSIONS" : "简历工坊 · 版本"}</div>
          <h1 className="ei-serif" style={{ fontSize: 38, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.15 }}>
            {lang === "en" ? "One master, one version per target." : "一份主版本，每个目标一份定制。"}
          </h1>
          <div style={{ fontSize: 14, color: T.ink3, marginTop: 10, maxWidth: 680, lineHeight: 1.5 }}>
            {lang === "en" ? "Rewrites are suggestions — you accept or reject each bullet. The master stays pristine; targeted versions diverge only where it matters." : "改写是建议——你可以一条一条采纳或拒绝。主版本保持干净，定制版本只在该分的地方分。"}
          </div>
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          <Btn T={T} variant="secondary" size="sm" icon="plus" onClick={() => setFlow("create")}>{lang === "en" ? "New version" : "新版本"}</Btn>
          <Btn T={T} variant="accent" size="sm" icon="download" onClick={() => {
            window.eiToast && window.eiToast(lang === "en" ? "Generating PDF… link will be emailed when ready" : "正在生成 PDF · 准备好后会邮件发送", { tone: "ok", duration: 2800 });
          }}>{lang === "en" ? "Export PDF" : "导出 PDF"}</Btn>
        </div>
      </div>

      <ResumeSourceMap
        T={T}
        lang={lang}
        activeVersion={activeVersion}
        activeSource={activeSource}
        onPreview={() => setSourcePreviewOpen(true)}
        onCreate={() => setFlow("create")}
      />

      {/* Version tabs */}
      <div style={{ display: "flex", gap: 0, marginBottom: 24, overflowX: "auto", borderBottom: `1px solid ${T.rule}` }}>
        {versions.map((v) => (
          <button key={v.id} onClick={() => setActive(v.id)} style={{
            padding: "14px 20px", background: "transparent", border: "none",
            borderBottom: `2px solid ${active === v.id ? T.accent : "transparent"}`,
            color: active === v.id ? T.ink : T.ink3, cursor: "pointer",
            fontFamily: "var(--ei-sans)", textAlign: "left", whiteSpace: "nowrap", marginBottom: -1,
          }}>
            <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 4 }}>
              <span style={{ fontSize: 13.5, fontWeight: active === v.id ? 500 : 400 }}>{v.name}</span>
              <span style={{ fontFamily: "var(--ei-mono)", fontSize: 9, letterSpacing: "0.08em", padding: "1px 6px", borderRadius: 2, background: v.tone === "accent" ? T.accentSoft : T.bgSoft, color: v.tone === "accent" ? T.accent : T.ink3 }}>{v.tag}</span>
            </div>
            <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>{v.date} · {v.bullets} bullets</div>
          </button>
        ))}
      </div>

      {/* Summary bar */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 10, marginBottom: 24 }}>
        {[
          { k: lang === "en" ? "Target" : "目标岗位", v: activeVersion.target },
          { k: lang === "en" ? "Original source" : "原始来源", v: activeSource.name, sub: activeSource.type },
          { k: lang === "en" ? "Bullets rewritten" : "改写 bullet", v: `${bullets.length}`, sub: lang === "en" ? `${accepted} accepted · ${pending} pending` : `${accepted} 已采纳 · ${pending} 待决定` },
          { k: lang === "en" ? "Match delta" : "匹配度变化", v: "+14%", sub: lang === "en" ? "64% → 78% vs JD" : "对 JD 的匹配度 64% → 78%", tone: "ok" },
        ].map((m) => (
          <div key={m.k} style={{ padding: "14px 16px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 5 }}>{m.k}</div>
            <div className="ei-serif" style={{ fontSize: 20, color: m.tone === "ok" ? T.ok : T.ink, letterSpacing: "-0.01em" }}>{m.v}</div>
            {m.sub && <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 3 }}>{m.sub}</div>}
          </div>
        ))}
      </div>

      {/* Diff view */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1.3fr", gap: 20 }}>
        {/* Bullet list */}
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "SUGGESTED REWRITES" : "建议改写"}</div>
          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {bullets.map((b) => {
              const active = b.id === selectedBullet;
              const statusC = b.status === "accepted" ? T.ok : b.status === "rejected" ? T.ink4 : T.warn;
              return (
                <button key={b.id} onClick={() => setSelectedBullet(b.id)} style={{
                  padding: "14px 16px", textAlign: "left", cursor: "pointer",
                  background: active ? T.bgSoft : T.bgCard,
                  border: `1px solid ${active ? T.accent : T.rule}`,
                  borderRadius: 2, fontFamily: "var(--ei-sans)",
                }}>
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 10, marginBottom: 6 }}>
                    <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>{b.section}</div>
                    <div style={{ display: "flex", gap: 4, alignItems: "center", fontSize: 10.5, color: statusC, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>
                      <div style={{ width: 5, height: 5, borderRadius: 3, background: statusC }} />
                      {b.status === "accepted" ? (lang === "en" ? "ACCEPTED" : "已采纳") : b.status === "rejected" ? (lang === "en" ? "REJECTED" : "已拒绝") : (lang === "en" ? "PENDING" : "待决定")}
                    </div>
                  </div>
                  <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.5, textDecoration: b.status === "rejected" ? "line-through" : "none", opacity: b.status === "rejected" ? 0.6 : 1 }}>
                    {b.rewritten.slice(0, 90)}{b.rewritten.length > 90 ? "…" : ""}
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* Diff detail */}
        <div>
          <Card T={T} pad={0}>
            <div style={{ padding: "16px 22px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <div className="ei-label" style={{ color: T.ink3 }}>{sel.section}</div>
              <div style={{ display: "flex", gap: 6 }}>
                <button style={{ padding: "5px 12px", fontSize: 12, cursor: "pointer", background: sel.status === "rejected" ? T.ink2 : "transparent", color: sel.status === "rejected" ? T.bg : T.ink3, border: `1px solid ${T.rule}`, borderRadius: 2, fontFamily: "var(--ei-sans)" }}>
                  <Icon name="x" size={11} style={{ marginRight: 4 }} /> {lang === "en" ? "Reject" : "拒绝"}
                </button>
                <button style={{ padding: "5px 12px", fontSize: 12, cursor: "pointer", background: "transparent", color: T.ink2, border: `1px solid ${T.rule}`, borderRadius: 2, fontFamily: "var(--ei-sans)" }}>
                  <Icon name="edit" size={11} style={{ marginRight: 4 }} /> {lang === "en" ? "Edit" : "编辑"}
                </button>
                <button style={{ padding: "5px 12px", fontSize: 12, cursor: "pointer", background: sel.status === "accepted" ? T.ok : T.accent, color: "#fff", border: "none", borderRadius: 2, fontFamily: "var(--ei-sans)" }}>
                  <Icon name="check" size={11} style={{ marginRight: 4 }} stroke={2.5} /> {sel.status === "accepted" ? (lang === "en" ? "Accepted" : "已采纳") : (lang === "en" ? "Accept" : "采纳")}
                </button>
              </div>
            </div>

            {/* Original */}
            <div style={{ padding: "20px 22px", borderBottom: `1px dotted ${T.rule}` }}>
              <div style={{ display: "flex", gap: 10, alignItems: "center", marginBottom: 10 }}>
                <div style={{ padding: "2px 8px", background: T.dangerSoft, color: T.danger, fontSize: 10.5, fontFamily: "var(--ei-mono)", letterSpacing: "0.08em", borderRadius: 2 }}>
                  − {lang === "en" ? "ORIGINAL" : "原句"}
                </div>
                <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{lang === "en" ? "from master" : "来自主版本"}</div>
              </div>
              <div style={{ fontSize: 14.5, color: T.ink2, lineHeight: 1.65, fontFamily: "var(--ei-serif)", background: T.dangerSoft, padding: "12px 14px", borderRadius: 2, borderLeft: `2px solid ${T.danger}` }}>
                {sel.original}
              </div>
            </div>

            {/* Rewritten */}
            <div style={{ padding: "20px 22px", borderBottom: `1px dotted ${T.rule}` }}>
              <div style={{ display: "flex", gap: 10, alignItems: "center", marginBottom: 10 }}>
                <div style={{ padding: "2px 8px", background: T.okSoft, color: T.ok, fontSize: 10.5, fontFamily: "var(--ei-mono)", letterSpacing: "0.08em", borderRadius: 2 }}>
                  + {lang === "en" ? "REWRITTEN" : "改写"}
                </div>
                <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{lang === "en" ? "confidence · high · pulled from cart-rewrite story" : "置信度 · 高 · 取自「购物车重写」故事"}</div>
              </div>
              <div style={{ fontSize: 14.5, color: T.ink, lineHeight: 1.65, fontFamily: "var(--ei-serif)", background: T.okSoft, padding: "12px 14px", borderRadius: 2, borderLeft: `2px solid ${T.ok}` }}>
                {sel.rewritten}
              </div>
            </div>

            {/* Why */}
            <div style={{ padding: "20px 22px" }}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "WHY THIS CHANGE" : "为什么这么改"}</div>
              <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                {sel.why.map((w, i) => (
                  <div key={i} style={{ display: "flex", gap: 10, fontSize: 13, color: T.ink2, lineHeight: 1.5 }}>
                    <Icon name="sparkle" size={12} color={T.accent} style={{ marginTop: 3, flexShrink: 0 }} />
                    <span>{w}</span>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        </div>
      </div>

      {sourcePreviewOpen && (
        <OriginalResumePreviewModal
          T={T}
          lang={lang}
          source={activeSource}
          activeVersion={activeVersion}
          onClose={() => setSourcePreviewOpen(false)}
        />
      )}
    </div>
  );
};

const ResumeSourceMap = ({ T, lang, activeVersion, activeSource, onPreview, onCreate }) => (
  <div style={{
    display: "grid",
    gridTemplateColumns: "1fr 1fr 1fr",
    gap: 0,
    border: `1px solid ${T.rule}`,
    background: T.bgCard,
    borderRadius: 3,
    marginBottom: 24,
    overflow: "hidden",
  }}>
    {[
      {
        icon: "file",
        label: lang === "en" ? "Original resume" : "原始简历",
        title: activeSource.name,
        body: activeSource.retained,
        action: lang === "en" ? "Preview original" : "预览原件",
        onClick: onPreview,
      },
      {
        icon: "resume",
        label: lang === "en" ? "Structured master" : "结构化主版本",
        title: activeSource.owner,
        body: lang === "en" ? "Parsed into editable sections; does not overwrite the original." : "解析成可编辑字段；不会覆盖原始简历。",
      },
      {
        icon: "briefcase",
        label: lang === "en" ? "Targeted version" : "岗位定制版本",
        title: activeVersion.name,
        body: lang === "en" ? `Branch for ${activeVersion.target}; only accepted changes enter this version.` : `面向 ${activeVersion.target}；只把采纳的改写写入此版本。`,
        action: lang === "en" ? "Import new resume" : "导入新简历",
        onClick: onCreate,
      },
    ].map((item, i) => (
      <div key={item.label} style={{ padding: "16px 18px", borderRight: i < 2 ? `1px dotted ${T.rule}` : "none", minWidth: 0 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 9 }}>
          <Icon name={item.icon} size={14} color={T.accent} />
          <div className="ei-label" style={{ color: T.ink3 }}>{item.label}</div>
        </div>
        <div style={{ fontSize: 14, color: T.ink, fontWeight: 600, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{item.title}</div>
        <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.5, marginTop: 5 }}>{item.body}</div>
        {item.action && (
          <button onClick={item.onClick} style={{ marginTop: 12, border: `1px solid ${T.rule}`, background: "transparent", color: T.ink2, borderRadius: 2, padding: "5px 10px", fontSize: 12, cursor: "pointer" }}>
            {item.action} <Icon name="arrow_right" size={10} style={{ marginLeft: 4 }} />
          </button>
        )}
      </div>
    ))}
  </div>
);

const OriginalResumePreviewModal = ({ T, lang, source, activeVersion, onClose }) => {
  const [view, setView] = React.useState("file");
  return (
    <div style={{ position: "fixed", inset: 0, background: "rgba(24, 20, 16, 0.24)", zIndex: 80, display: "flex", alignItems: "center", justifyContent: "center", padding: 24 }} onClick={onClose}>
      <div className="ei-fadein" onClick={(e) => e.stopPropagation()} style={{ width: "min(960px, 100%)", maxHeight: "88vh", overflow: "auto", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4, boxShadow: "0 24px 70px rgba(30, 22, 15, 0.24)" }}>
        <div style={{ padding: "20px 24px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 18 }}>
          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{lang === "en" ? "ORIGINAL RESUME PREVIEW" : "原始简历预览"}</div>
            <div className="ei-serif" style={{ fontSize: 24, color: T.ink }}>{source.name}</div>
            <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 6 }}>{source.type} · {source.createdAt} · {source.retained}</div>
          </div>
          <button onClick={onClose} style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", padding: 4 }}>
            <Icon name="x" size={16} />
          </button>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "250px 1fr", minHeight: 520 }}>
          <div style={{ borderRight: `1px solid ${T.rule}`, padding: 18, background: T.bgSoft }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "SOURCE RELATION" : "来源关系"}</div>
            {[
              [lang === "en" ? "Original file" : "原始文件", source.name],
              [lang === "en" ? "Parsed into" : "解析为", source.owner],
              [lang === "en" ? "Current version" : "当前版本", activeVersion.name],
            ].map(([k, v], i) => (
              <div key={k} style={{ padding: "10px 0", borderBottom: i < 2 ? `1px dotted ${T.rule}` : "none" }}>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{k}</div>
                <div style={{ fontSize: 13, color: T.ink, lineHeight: 1.45 }}>{v}</div>
              </div>
            ))}
            <div style={{ marginTop: 16, fontSize: 12.5, color: T.ink3, lineHeight: 1.6 }}>
              {lang === "en"
                ? "The original is read-only. Edits and JD-specific rewrites create versions beside it."
                : "原始简历只读保存。编辑和 JD 定制改写会生成旁路版本，不覆盖原件。"}
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 8, marginTop: 18 }}>
              {[
                ["file", lang === "en" ? "Original file" : "原始文件"],
                ["text", lang === "en" ? "Parsed text" : "解析文本"],
              ].map(([k, label]) => (
                <button key={k} onClick={() => setView(k)} style={{
                  textAlign: "left", border: `1px solid ${view === k ? T.accent : T.rule}`, background: view === k ? T.accentSoft : T.bgCard,
                  color: view === k ? T.ink : T.ink2, borderRadius: 2, padding: "9px 10px", cursor: "pointer", fontFamily: "var(--ei-sans)", fontSize: 13,
                }}>
                  {label}
                </button>
              ))}
            </div>
          </div>

          <div style={{ padding: 24, background: T.bg }}>
            {view === "file" ? (
              <div style={{ maxWidth: 560, minHeight: 620, margin: "0 auto", background: "#fff", color: "#222", border: `1px solid ${T.rule}`, boxShadow: "0 18px 50px rgba(30, 22, 15, 0.12)", padding: 36, fontFamily: "Georgia, serif" }}>
                <div style={{ fontSize: 25, fontWeight: 600, marginBottom: 4 }}>{source.text[0]}</div>
                <div style={{ fontSize: 12, color: "#666", marginBottom: 18 }}>{source.name} · {source.createdAt}</div>
                <div style={{ height: 1, background: "#333", marginBottom: 18 }} />
                <div style={{ fontSize: 11, color: "#888", letterSpacing: "0.12em", textTransform: "uppercase", marginBottom: 8 }}>Experience</div>
                {source.text.slice(1).map((line, i) => (
                  <div key={i} style={{ fontSize: i === 0 ? 15 : 13.5, fontWeight: i === 0 ? 600 : 400, color: i === 0 ? "#222" : "#444", lineHeight: 1.75, marginBottom: 6 }}>
                    {line}
                  </div>
                ))}
                <div style={{ marginTop: 22, fontSize: 11, color: "#888", letterSpacing: "0.12em", textTransform: "uppercase", marginBottom: 8 }}>Skills</div>
                <div style={{ fontSize: 13.5, color: "#444", lineHeight: 1.75 }}>React · TypeScript · Performance · Design System · Platform Engineering</div>
              </div>
            ) : (
              <div style={{ border: `1px solid ${T.rule}`, background: T.bgCard, borderRadius: 3, padding: 18 }}>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "PARSED TEXT SNAPSHOT" : "解析文本快照"}</div>
                {source.text.map((line, i) => (
                  <div key={i} style={{ padding: "10px 0", borderBottom: i < source.text.length - 1 ? `1px dotted ${T.rule}` : "none", fontSize: 13.5, color: T.ink2, lineHeight: 1.6 }}>
                    {line}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

window.DebriefFullScreen = DebriefFullScreen;
// Legacy ResumeVersionsScreen is superseded by screen-resume-workshop.jsx (loaded after this file).
// Kept here as dead code to avoid touching a 1300-line file; the new screen overrides window.ResumeVersionsScreen.
window._LegacyResumeVersionsScreen = ResumeVersionsScreen;
