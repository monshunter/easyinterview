// P1 depth: Real-interview debrief, Experience library, Resume versions + diff

// ═══════════════════════════════════════════════════════════════════
// #9 REAL-INTERVIEW DEBRIEF (full version)
// ═══════════════════════════════════════════════════════════════════
const DebriefFullScreen = ({ T, lang, nav }) => {
  const [step, setStep] = React.useState(0); // 0 record, 1 analyze, 2 replay
  const [activeGuide, setActiveGuide] = React.useState(0);
  const [activeCard, setActiveCard] = React.useState("e1");

  const steps = lang === "en"
    ? ["Debrief record", "Debrief analysis", "Debrief interview"]
    : ["复盘记录", "复盘分析", "复盘面试"];

  const context = lang === "en" ? {
    target: "Star-Ring · Senior Frontend Engineer",
    jd: "P6 · Shanghai · JD match 78%",
    mock: "Mock interview #24 · text · 4/20",
    resume: "Liu Zhe · resume v3 · 78% match",
  } : {
    target: "星环科技 · 资深前端工程师",
    jd: "P6 · 上海 · JD 匹配 78%",
    mock: "模拟面试 #24 · 文本 · 4/20",
    resume: "刘哲 · 简历 v3 · 匹配 78%",
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
    { stage: "反问", q: "你是否向对方询问团队重点、下一轮预期或当前痛点？", why: "反问能判断真实面试是否收束到清晰的下一步信号。", source: "模拟报告" },
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
        <Icon name="arrow_left" size={13} /> {lang === "en" ? "Inbox" : "收件箱"}
      </button>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginBottom: 32, gap: 32 }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
            {lang === "en" ? "REAL-INTERVIEW DEBRIEF · 星环科技 · 技术一面 · 4/22" : "真实面试复盘 · 星环科技 · 技术一面 · 4/22"}
          </div>
          <h1 className="ei-serif" style={{ fontSize: 38, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.15, maxWidth: 780 }}>
            {lang === "en" ? "Reconstruct the real interview, then practice the replay." : "像真人复盘一样，还原真实面试，再进入复盘面试。"}
          </h1>
          <div style={{ fontSize: 14, color: T.ink3, marginTop: 10, maxWidth: 640, lineHeight: 1.5 }}>
            {lang === "en" ? "Pick the target JD and resume, keep or skip AI-suggested questions, add what was actually asked, then generate a replay interview." : "选择目标岗位和简历，保留或跳过 AI 推测题，补充真实被问到的问题，然后生成一场复盘面试。"}
          </div>
        </div>
        <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink3, textAlign: "right", lineHeight: 1.7 }}>
          <div>time · 40 min ago</div>
          <div>interviewer · 张哲 · tech lead</div>
          <div>modality · video</div>
        </div>
      </div>

      <DebriefContextStrip T={T} lang={lang} nav={nav} context={context} />

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
          <div style={{ display: "grid", gridTemplateColumns: "1fr 320px", gap: 28 }}>
            <div>
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
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "REAL INTERVIEW ANALYSIS" : "真实面试分析"}</div>

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
              { title: lang === "en" ? "Compared with target JD" : "对照目标 JD", body: lang === "en" ? "The real interview spent more time on ownership and influence than the mock interview predicted." : "真实面试比模拟面试更集中在 ownership 与影响力，而不是单纯技术细节。" },
              { title: lang === "en" ? "Compared with resume evidence" : "对照绑定简历", body: lang === "en" ? "Resume v3 has the right stories, but the real answer did not claim the decision points clearly." : "简历 v3 有对应素材，但真实回答没有把决策点和个人贡献讲清。" },
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

const DebriefContextStrip = ({ T, lang, nav, context }) => (
  <div style={{ display: "grid", gridTemplateColumns: "1.2fr 1fr 1fr", gap: 12, marginBottom: 28 }}>
    {[
      { icon: "briefcase", label: lang === "en" ? "Target job / JD" : "目标岗位 / JD", title: context.target, meta: context.jd, action: lang === "en" ? "Change" : "更换", onClick: () => nav("workspace", { jobId: "tj-1" }) },
      { icon: "chart", label: lang === "en" ? "Related mock interview" : "关联模拟面试", title: context.mock, meta: lang === "en" ? "used as comparison baseline" : "作为真实面试分析基线", action: lang === "en" ? "Report" : "报告", onClick: () => nav("report") },
      { icon: "resume", label: lang === "en" ? "Resume version" : "绑定简历", title: context.resume, meta: lang === "en" ? "used for evidence comparison" : "用于回答证据对比", action: lang === "en" ? "Change" : "更换", onClick: () => nav("resume_versions") },
    ].map((item) => (
      <div key={item.label} style={{ padding: "14px 16px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, display: "grid", gridTemplateColumns: "30px 1fr auto", gap: 10, alignItems: "center" }}>
        <div style={{ width: 30, height: 30, borderRadius: 15, background: T.bgSoft, color: T.accent, display: "flex", alignItems: "center", justifyContent: "center" }}>
          <Icon name={item.icon} size={14} />
        </div>
        <div style={{ minWidth: 0 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 3 }}>{item.label}</div>
          <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{item.title}</div>
          <div style={{ fontSize: 12, color: T.ink3, marginTop: 2, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{item.meta}</div>
        </div>
        <button onClick={item.onClick} style={{ background: "transparent", border: `1px solid ${T.rule}`, borderRadius: 2, color: T.ink2, fontSize: 12, padding: "5px 9px", cursor: "pointer" }}>{item.action}</button>
      </div>
    ))}
  </div>
);

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
      source: "ai_confirmed",
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
            <Btn T={T} variant="ghost" size="sm" icon="edit" onClick={addCurrentQuestion}>{lang === "en" ? "Edit as real question" : "改成真实问题"}</Btn>
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
                <Tag tone={entry.source === "confirmed" ? "ok" : "accent"} T={T}>{entry.source === "confirmed" ? (lang === "en" ? "real" : "真实遇到") : (lang === "en" ? "confirmed" : "已确认")}</Tag>
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

const DebriefReplayPlan = ({ T, lang, nav, back, entries, context }) => (
  <div style={{ display: "grid", gridTemplateColumns: "1fr 340px", gap: 28 }}>
    <div>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "DEBRIEF INTERVIEW CONTENT" : "复盘面试内容"}</div>
      <div style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, padding: "22px 24px", marginBottom: 18 }}>
        <div className="ei-serif" style={{ fontSize: 22, color: T.ink, marginBottom: 10 }}>
          {lang === "en" ? "Use real questions first, AI probes second." : "先复现真实问题，再由 AI 追问薄弱处。"}
        </div>
        <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.7 }}>
          {lang === "en" ? "The replay interview will prioritize the questions you actually met, then add adjacent probes from the JD and resume evidence." : "复盘面试会优先使用你真实遇到的问题，再基于 JD 和简历证据补上相邻追问。"}
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
        <Btn T={T} variant="accent" icon="play" onClick={() => nav("practice", { jobId: "tj-1", mode: "text" })}>{lang === "en" ? "Start debrief interview" : "开始复盘面试"}</Btn>
      </div>
    </div>
    <Card T={T} pad={18} style={{ height: "fit-content", position: "sticky", top: 20 }}>
      <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>{lang === "en" ? "INTERVIEW CONTEXT" : "复盘面试上下文"}</div>
      <div style={{ paddingBottom: 12, borderBottom: `1px dotted ${T.rule}`, marginBottom: 8 }}>
        <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>{context.target}</div>
        <div style={{ fontSize: 12, color: T.ink3, marginTop: 3 }}>{context.resume}</div>
      </div>
      {[
        lang === "en" ? "Ask your real questions in the original order where possible." : "尽可能按真实顺序重新问一遍。",
        lang === "en" ? "Probe the weak ownership and reverse-question moments." : "重点追问 Ownership 与反问薄弱点。",
        lang === "en" ? "Compare this replay answer with the recorded real answer." : "把复盘回答与真实记录做对照。",
      ].map((text, i) => (
        <div key={i} style={{ display: "flex", gap: 10, padding: "10px 0", borderBottom: i < 2 ? `1px dotted ${T.rule}` : "none", fontSize: 13.5, color: T.ink2, lineHeight: 1.55 }}>
          <span style={{ fontFamily: "var(--ei-mono)", color: T.accent }}>{String(i + 1).padStart(2, "0")}</span>
          <span>{text}</span>
        </div>
      ))}
    </Card>
  </div>
);

const ThankYouLetter = ({ T, lang, back, entries }) => {
  const [tone, setTone] = React.useState("warm"); // warm / concise / formal
  const [variant, setVariant] = React.useState(0);

  const tones = lang === "en"
    ? [{ k: "warm", t: "Warm" }, { k: "concise", t: "Concise" }, { k: "formal", t: "Formal" }]
    : [{ k: "warm", t: "温和" }, { k: "concise", t: "简洁" }, { k: "formal", t: "正式" }];

  const letters = {
    warm: lang === "en"
      ? `Hi 张哲,

Thanks for the 45 minutes this afternoon. The question about who specifically drove the RSC migration stayed with me after the call — the honest answer is that I proposed it, ran the prototype, and pushed the rollout through a skeptical SRE team. I undersold that on the call.

One thing I wanted to circle back on: you mentioned the team is weighing whether to keep the checkout rewrite on RSC or move back to client-rendering for the edit paths. I'd love to share the decision matrix I used at Star-Ring — if it's useful, happy to drop it in as a follow-up.

Either way, the conversation was sharper than most first rounds I've had. Looking forward to the next step.

— 林舟`
      : `哲哥，你好：

感谢今天下午的 45 分钟。你问到 RSC 迁移具体是谁推动的，那个问题面完之后一直在我脑子里——诚实的回答是：方案是我提的、原型是我跑的，也是我顶着 SRE 的质疑把上线推过去的。当时在镜头前我没讲到位。

另外有件事想多说一句：你提到团队正在权衡结账流程的编辑路径要不要从 RSC 回退到客户端渲染。我之前在星环做过一个类似的决策矩阵，如果你们用得上，我可以整理出来发你。

不管结果如何，今天的对话比我过去大多数一面都要锋利，期待下一步。

— 林舟`,
    concise: lang === "en"
      ? `Hi 张哲,

Thanks for today. Quick correction on the RSC migration — I was the one who proposed and drove it. I undersold that on the call.

Happy to share the decision matrix we used for the rollback question if useful.

Looking forward to the next step.

— 林舟`
      : `哲哥：

感谢今天。关于 RSC 迁移一个更正——方案是我提的、我推的。当时没讲到位。

如果你们在 RSC 回退那个问题上用得到，我可以把当时的决策矩阵发给你。

期待下一步。

— 林舟`,
    formal: lang === "en"
      ? `Dear 张哲,

Thank you for the interview this afternoon. I appreciated the depth of the discussion, particularly around the checkout rewrite and Design System rollout.

I'd like to clarify one point: the RSC migration at Star-Ring was proposed, prototyped, and driven through to rollout primarily by me. I did not articulate this clearly during our conversation.

Should it be helpful, I would be glad to share the decision framework we used when considering a partial rollback — a question you raised that I have direct experience with.

Looking forward to hearing from you.

Sincerely,
林舟`
      : `哲哥：

感谢您今天下午的面试。我很珍惜这次深入的交流，尤其是围绕结账链路重写与 Design System 落地的讨论。

有一点想再澄清：星环的 RSC 迁移，从方案提出、原型验证到最终推动上线，主要由我负责。这一点我在面试中未能清晰表达。

您提到关于是否部分回退到客户端渲染的权衡，我之前处理过类似场景，如果对团队当前的决策有帮助，我可以整理一份决策框架供您参考。

期待您的回复。

林舟`,
  };

  const content = letters[tone] || letters.warm;
  const wordCount = content.length;

  return (
    <div>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 300px", gap: 28 }}>
        <div>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14 }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "THANK-YOU NOTE · draft" : "感谢信 · 草稿"}</div>
            <div style={{ display: "flex", gap: 6 }}>
              {tones.map((t) => (
                <button key={t.k} onClick={() => setTone(t.k)} style={{
                  padding: "5px 12px", fontSize: 12, borderRadius: 2, cursor: "pointer",
                  border: `1px solid ${tone === t.k ? T.accent : T.rule}`,
                  background: tone === t.k ? T.accentSoft : "transparent",
                  color: tone === t.k ? T.accent : T.ink2, fontFamily: "var(--ei-sans)",
                }}>{t.t}</button>
              ))}
            </div>
          </div>

          <div style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, padding: "24px 28px" }}>
            <div style={{ paddingBottom: 12, marginBottom: 16, borderBottom: `1px dotted ${T.rule}`, display: "flex", justifyContent: "space-between", fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink3 }}>
              <div>TO · zhang.zhe@star-ring.com</div>
              <div>SUBJECT · {lang === "en" ? "Quick note after today's chat" : "今天面试后的一点补充"}</div>
            </div>
            <textarea
              key={tone}
              defaultValue={content}
              style={{
                width: "100%", minHeight: 420, padding: 0,
                fontFamily: "var(--ei-serif)", fontSize: 15, lineHeight: 1.75, color: T.ink,
                background: "transparent", border: "none", outline: "none", resize: "vertical",
                boxSizing: "border-box", whiteSpace: "pre-wrap",
              }}
            />
          </div>

          <div style={{ marginTop: 14, display: "flex", justifyContent: "space-between", alignItems: "center", fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
            <div>≈ {wordCount} {lang === "en" ? "chars · 45s read" : "字 · 45 秒读完"} · {lang === "en" ? "draft v1 · auto-saved" : "草稿 v1 · 已自动保存"}</div>
            <div>{lang === "en" ? "send within 24h recommended" : "建议 24h 内发出"}</div>
          </div>

          <div style={{ display: "flex", justifyContent: "space-between", marginTop: 24 }}>
            <Btn T={T} variant="ghost" onClick={back}>{lang === "en" ? "Back" : "上一步"}</Btn>
            <div style={{ display: "flex", gap: 10 }}>
              <Btn T={T} variant="secondary" icon="download" onClick={() => {
                if (navigator.clipboard && navigator.clipboard.writeText) {
                  navigator.clipboard.writeText(content);
                  window.eiToast && window.eiToast(lang === "en" ? "Letter copied to clipboard" : "致谢信已复制到剪贴板", { tone: "ok" });
                } else {
                  window.eiToast && window.eiToast(lang === "en" ? "Clipboard unavailable" : "当前环境不支持剪贴板", { tone: "warn" });
                }
              }}>{lang === "en" ? "Copy text" : "复制文本"}</Btn>
              <Btn T={T} variant="accent" icon="send" onClick={() => {
                const subject = encodeURIComponent(lang === "en" ? "Thank you — follow up after our chat" : "今天面试后的一点补充");
                const body = encodeURIComponent(content);
                window.location.href = `mailto:zhang.zhe@star-ring.com?subject=${subject}&body=${body}`;
              }}>{lang === "en" ? "Open in mail client" : "在邮件客户端打开"}</Btn>
            </div>
          </div>
        </div>

        <div>
          <Card T={T} pad={18} style={{ marginBottom: 14 }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "WHAT THIS LETTER DOES" : "这封信在做什么"}</div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10, fontSize: 12.5, lineHeight: 1.55, color: T.ink2 }}>
              {[
                { t: lang === "en" ? "Repairs the Ownership miss" : "修复 Ownership 的失分", c: T.warn },
                { t: lang === "en" ? "Reinforces perf credibility" : "夯实性能可信度", c: T.ok },
                { t: lang === "en" ? "Opens a second touchpoint (decision matrix)" : "开一个二次接触点（决策矩阵）", c: T.accent },
                { t: lang === "en" ? "Keeps tone within what you showed on call" : "语气不超出你面试时的形象", c: T.ink3 },
              ].map((x, i) => (
                <div key={i} style={{ display: "flex", gap: 8, alignItems: "flex-start" }}>
                  <div style={{ width: 5, height: 5, borderRadius: 3, background: x.c, marginTop: 7, flexShrink: 0 }} />
                  <span>{x.t}</span>
                </div>
              ))}
            </div>
          </Card>

          <Card T={T} pad={18}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "FOLLOW-UP PLAN" : "后续动作"}</div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10, fontSize: 12.5, color: T.ink2, lineHeight: 1.5 }}>
              <div><span style={{ fontFamily: "var(--ei-mono)", color: T.ink4, marginRight: 6 }}>T+0</span>{lang === "en" ? "Send this note" : "发这封信"}</div>
              <div><span style={{ fontFamily: "var(--ei-mono)", color: T.ink4, marginRight: 6 }}>T+1</span>{lang === "en" ? "Prep for tech round 2" : "准备技术二面"}</div>
              <div><span style={{ fontFamily: "var(--ei-mono)", color: T.ink4, marginRight: 6 }}>T+4</span>{lang === "en" ? "If silent → light follow-up to HR" : "如无回应，轻量跟 HR 联系"}</div>
              <div><span style={{ fontFamily: "var(--ei-mono)", color: T.ink4, marginRight: 6 }}>T+10</span>{lang === "en" ? "Close loop, move on" : "收口，继续别的"}</div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
};

// ═══════════════════════════════════════════════════════════════════
// #10 EXPERIENCE LIBRARY
// ═══════════════════════════════════════════════════════════════════
const ExperienceLibraryScreen = ({ T, lang, nav }) => {
  const [filter, setFilter] = React.useState("all");
  const [selected, setSelected] = React.useState("c1");

  const cards = [
    { id: "c1", title: lang === "en" ? "Cart / checkout rewrite" : "购物车/结账链路重写", period: "2024 Q1–Q2", co: "星环科技", tags: lang === "en" ? ["perf", "ownership", "trade-offs"] : ["性能", "Ownership", "权衡"], uses: 6, last: lang === "en" ? "used 2 days ago" : "2 天前用过", maturity: "strong",
      s: lang === "en" ? "Checkout LCP 3.2s; ~8% abandon on step 2; legacy class components; no SSR." : "结账 LCP 3.2s；第二步流失约 8%；老的 class 组件栈；无 SSR。",
      a: lang === "en" ? "Proposed RSC + selective hydration. Built 2-week prototype, got buy-in from SRE despite skepticism. Rolled out in 4 waves behind a feature flag." : "提出 RSC + 选择性注水。两周搭出原型，顶着 SRE 的质疑拿到认可。用 feature flag 分 4 波灰度上线。",
      r: lang === "en" ? "LCP 3.2s → 1.4s. Abandon 8% → 4.2%. Revenue +1.8M / quarter." : "LCP 3.2s → 1.4s。流失率 8% → 4.2%。营收季度 +180 万。",
      questions: lang === "en" ? ["Tell me about a technical trade-off", "A time you drove a project end-to-end", "Performance optimization story"] : ["一次技术权衡", "端到端推动的项目", "性能优化故事"],
    },
    { id: "c2", title: lang === "en" ? "Design System rollout across 5 products" : "Design System 在 5 个产品的落地", period: "2023 H2", co: "星环科技", tags: lang === "en" ? ["influence", "cross-team", "design"] : ["影响力", "跨团队", "设计"], uses: 4, last: lang === "en" ? "used 1 week ago" : "1 周前用过", maturity: "strong",
      s: lang === "en" ? "5 products, 4 different UI libraries, 0 shared tokens." : "5 个产品、4 套 UI 库、0 个共享 token。",
      a: lang === "en" ? "Built tokens + primitives, picked one pilot team, ran 3 adoption workshops, offered paired migration." : "搭 tokens + 基础组件、挑一个试点团队、办 3 次推广会、提供结对迁移。",
      r: lang === "en" ? "4 of 5 products on v1 in 6 months. Onboarding new devs ~50% faster per pilot team PM." : "6 个月内 4/5 产品迁到 v1。试点团队 PM 反馈新人上手快 ~50%。",
      questions: lang === "en" ? ["How would you drive an initiative without authority", "Cross-team alignment", "Changing how people work"] : ["无权情况下如何推动", "跨团队对齐", "改变别人的工作方式"],
    },
    { id: "c3", title: lang === "en" ? "Disagreement with designer (40-field form)" : "与设计师的分歧（40 字段表单）", period: "2023 Q3", co: "星环科技", tags: lang === "en" ? ["conflict", "empathy", "data"] : ["冲突", "共情", "数据"], uses: 5, last: lang === "en" ? "used 3 days ago" : "3 天前用过", maturity: "strong",
      s: lang === "en" ? "Designer wanted 40 operational fields on one screen; I thought cognitive load would tank efficiency." : "设计想把 40 个运营字段放一屏；我担心认知负荷会拖垮效率。",
      a: lang === "en" ? "Ran shadow-session with 6 real operators, tracked task time and error rate, proposed tabbed layout with smart defaults." : "跟 6 个真实操作员做影子观察，记录任务耗时和错误率，提出带智能默认的标签页布局。",
      r: lang === "en" ? "Task time -42%. Designer bought in after seeing the data. Shipped in 2 weeks." : "任务耗时 -42%。设计师看完数据认可方案，两周上线。",
      questions: lang === "en" ? ["Conflict with a peer", "Changed someone's mind", "Used data to resolve disagreement"] : ["与同事的冲突", "改变别人的看法", "用数据解决分歧"],
    },
    { id: "c4", title: lang === "en" ? "Failed A/B on checkout CTA color" : "结账 CTA 颜色的失败 A/B", period: "2021 Q4", co: "Lumen", tags: lang === "en" ? ["failure", "learning", "data"] : ["失败", "反思", "数据"], uses: 2, last: lang === "en" ? "used 3 weeks ago" : "3 周前用过", maturity: "dev",
      s: lang === "en" ? "Hypothesized a green CTA would beat the existing orange one." : "以为把 CTA 从橙改绿能提升转化。",
      a: lang === "en" ? "Shipped the test, ran for 2 weeks with enough samples to declare." : "上线实验，跑两周拿到足够样本量。",
      r: lang === "en" ? "No measurable difference. Taught me to lead with priors and mechanism before A/B — not hunches." : "没有显著差异。学会先看先验、先讲机制，而不是凭感觉做 A/B。",
      questions: lang === "en" ? ["A time you failed", "Something that didn't work out"] : ["一次失败", "没做成的事"],
    },
    { id: "c5", title: lang === "en" ? "Onboarded 3 new frontend engineers" : "带起 3 个前端新人", period: "2024 Q3–Q4", co: "星环科技", tags: lang === "en" ? ["mentorship", "leadership"] : ["辅导", "领导力"], uses: 1, last: lang === "en" ? "used 1 month ago" : "1 个月前用过", maturity: "dev",
      s: lang === "en" ? "Team scaled from 4 to 7; no onboarding process existed." : "团队从 4 人扩到 7 人；没有 onboarding 流程。",
      a: lang === "en" ? "Wrote a 2-week ramp doc, paired every new dev with a code owner, instituted weekly show-and-tell." : "写了一份 2 周上手文档，每个新人配一个代码 owner，周会做成果展示。",
      r: lang === "en" ? "Time-to-first-PR: 9 days → 3 days. All 3 retained after 1 year." : "首个 PR 时间 9 天 → 3 天。一年后 3 人全部留存。",
      questions: lang === "en" ? ["Mentorship story", "Building a process from scratch"] : ["辅导他人", "从零建流程"],
    },
  ];

  const filters = lang === "en"
    ? [{ k: "all", t: "All" }, { k: "strong", t: "Battle-tested" }, { k: "dev", t: "Developing" }, { k: "orphan", t: "Unused 30d+" }]
    : [{ k: "all", t: "全部" }, { k: "strong", t: "已稳" }, { k: "dev", t: "成型中" }, { k: "orphan", t: "30 天未用" }];

  const filtered = cards.filter((c) => filter === "all" || (filter === "strong" && c.maturity === "strong") || (filter === "dev" && c.maturity === "dev") || (filter === "orphan" && c.uses <= 1));
  const sel = cards.find((c) => c.id === selected);

  return (
    <div className="ei-fadein" style={{ maxWidth: 1320, margin: "0 auto", padding: "40px 48px 96px" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginBottom: 28, gap: 32 }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "EXPERIENCE LIBRARY" : "经历库"}</div>
          <h1 className="ei-serif" style={{ fontSize: 38, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.15 }}>
            {lang === "en" ? "Your stories, versioned." : "你自己的故事——带版本。"}
          </h1>
          <div style={{ fontSize: 14, color: T.ink3, marginTop: 10, maxWidth: 680, lineHeight: 1.5 }}>
            {lang === "en" ? "Each card is a story you can reach for. The more you use one, the sharper it gets. Unused ones quietly age — that's a hint, not a judgment." : "每张卡片是一个你随手能取的故事。用得越多越锋利。长时间没用的会悄悄变暗——这是提示，不是评判。"}
          </div>
        </div>
        <Btn T={T} variant="accent" icon="plus" onClick={() => nav("star", { storyId: "new" })}>{lang === "en" ? "New story" : "新建故事"}</Btn>
      </div>

      {/* Filters */}
      <div style={{ display: "flex", gap: 8, marginBottom: 20 }}>
        {filters.map((f) => (
          <button key={f.k} onClick={() => setFilter(f.k)} style={{
            padding: "6px 14px", fontSize: 12.5, borderRadius: 14, cursor: "pointer",
            border: `1px solid ${filter === f.k ? T.accent : T.rule}`,
            background: filter === f.k ? T.accentSoft : "transparent",
            color: filter === f.k ? T.accent : T.ink2, fontFamily: "var(--ei-sans)",
          }}>{f.t}</button>
        ))}
        <div style={{ flex: 1 }} />
        <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", alignSelf: "center" }}>
          {filtered.length} / {cards.length}
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1.2fr", gap: 20 }}>
        {/* Card grid */}
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          {filtered.map((c) => {
            const active = c.id === selected;
            return (
              <button key={c.id} onClick={() => setSelected(c.id)} style={{
                padding: "14px 16px", textAlign: "left", cursor: "pointer",
                background: active ? T.accentSoft : T.bgCard,
                border: `1px solid ${active ? T.accent : T.rule}`,
                borderLeft: `3px solid ${c.maturity === "strong" ? T.ok : T.warn}`,
                borderRadius: 2, fontFamily: "var(--ei-sans)",
              }}>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 6, gap: 10 }}>
                  <div className="ei-serif" style={{ fontSize: 15.5, color: T.ink, fontWeight: 500, letterSpacing: "-0.01em", lineHeight: 1.35 }}>{c.title}</div>
                  <div style={{ fontFamily: "var(--ei-mono)", fontSize: 10.5, color: T.ink4, flexShrink: 0, textAlign: "right" }}>{c.period}</div>
                </div>
                <div style={{ fontSize: 11.5, color: T.ink3, marginBottom: 10, fontFamily: "var(--ei-mono)" }}>{c.co}</div>
                <div style={{ display: "flex", gap: 4, flexWrap: "wrap", marginBottom: 10 }}>
                  {c.tags.map((t) => <Tag key={t} T={T} tone="neutral">{t}</Tag>)}
                </div>
                <div style={{ display: "flex", justifyContent: "space-between", fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>
                  <span>↻ {c.uses}× {lang === "en" ? "used" : "用过"}</span>
                  <span>{c.last}</span>
                </div>
              </button>
            );
          })}
        </div>

        {/* Detail */}
        <div>
          <Card T={T} pad={28} style={{ position: "sticky", top: 20 }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 16, gap: 12 }}>
              <div>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 5 }}>{sel.co} · {sel.period}</div>
                <div className="ei-serif" style={{ fontSize: 24, color: T.ink, letterSpacing: "-0.015em", lineHeight: 1.25 }}>{sel.title}</div>
              </div>
              <Tag T={T} tone={sel.maturity === "strong" ? "ok" : "warn"}>
                {sel.maturity === "strong" ? (lang === "en" ? "BATTLE-TESTED" : "已稳") : (lang === "en" ? "DEVELOPING" : "成型中")}
              </Tag>
            </div>

            <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginBottom: 22 }}>
              {sel.tags.map((t) => <Tag key={t} T={T} tone="accent">{t}</Tag>)}
            </div>

            {[
              { k: "S", label: lang === "en" ? "SITUATION" : "情境", v: sel.s },
              { k: "A", label: lang === "en" ? "ACTION" : "行动", v: sel.a },
              { k: "R", label: lang === "en" ? "RESULT" : "结果", v: sel.r },
            ].map((x) => (
              <div key={x.k} style={{ paddingBottom: 16, marginBottom: 16, borderBottom: `1px dotted ${T.rule}` }}>
                <div style={{ fontFamily: "var(--ei-mono)", fontSize: 10.5, color: T.ink4, letterSpacing: "0.08em", marginBottom: 6 }}>
                  {x.k} · {x.label}
                </div>
                <div style={{ fontSize: 14, color: T.ink, lineHeight: 1.6 }}>{x.v}</div>
              </div>
            ))}

            <div style={{ marginBottom: 22 }}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "GOOD FOR QUESTIONS LIKE" : "适合这类问题"}</div>
              <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
                {sel.questions.map((q, i) => (
                  <div key={i} style={{ fontSize: 13, color: T.ink2, fontStyle: "italic", paddingLeft: 10, borderLeft: `2px solid ${T.rule}` }}>"{q}"</div>
                ))}
              </div>
            </div>

            <div style={{ display: "flex", gap: 8 }}>
              <Btn T={T} variant="secondary" size="sm" icon="edit" onClick={() => nav("star", { storyId: sel.id })}>{lang === "en" ? "Restructure (STAR)" : "STAR 重构"}</Btn>
              <Btn T={T} variant="secondary" size="sm" icon="replay" onClick={() => nav("practice", { jobId: "tj-1" })}>{lang === "en" ? "Practice with this" : "用它练一下"}</Btn>
              <div style={{ flex: 1 }} />
              <Btn T={T} variant="ghost" size="sm" icon="trash" onClick={() => window.eiToast && window.eiToast(lang === "en" ? `Archived "${sel.title}" · move out of active library` : `已归档「${sel.title}」· 移出活跃故事库`, { tone: "neutral" })}>{lang === "en" ? "Archive" : "归档"}</Btn>
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
};

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
          <Btn T={T} variant="secondary" size="sm" icon="briefcase" onClick={() => nav("workspace", { jobId: "tj-1" })}>{lang === "en" ? "Use in current job" : "用于当前岗位"}</Btn>
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
                <Btn T={T} variant="secondary" size="sm" icon="play" onClick={() => nav("practice", { jobId: "tj-1", mode: "text" })}>{lang === "en" ? "Mock interview" : "开始面试"}</Btn>
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
window.ExperienceLibraryScreen = ExperienceLibraryScreen;
window.ResumeVersionsScreen = ResumeVersionsScreen;
