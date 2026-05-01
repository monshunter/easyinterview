// P0 completion screens: JD parse flow, Onboarding, Report-generating, Settings/Privacy, Retest state machine, Resume diff

// ═══════════════════════════════════════════════════════════════════
// #1 JD PARSE FLOW — loading state + structured preview / confirm
// ═══════════════════════════════════════════════════════════════════
const ParseScreen = ({ T, lang, nav }) => {
  const [stage, setStage] = React.useState("loading"); // loading -> preview
  const [step, setStep] = React.useState(0);

  const steps = lang === "en" ? [
    "Extracting title, level, location",
    "Identifying must-have vs nice-to-have",
    "Building the mock interview context",
    "Matching against your profile",
  ] : [
    "抽取岗位名、职级、地点",
    "识别必需项与加分项",
    "生成模拟面试上下文",
    "对比你的画像",
  ];

  React.useEffect(() => {
    if (stage !== "loading") return;
    const ticks = [600, 900, 800, 700];
    let cancel = false;
    let acc = 0;
    ticks.forEach((t, i) => {
      acc += t;
      setTimeout(() => { if (!cancel) setStep(i + 1); }, acc);
    });
    setTimeout(() => { if (!cancel) setStage("preview"); }, acc + 200);
    return () => { cancel = true; };
  }, [stage]);

  // Mock parsed data — exposed for user confirmation
  const [parsed, setParsed] = React.useState({
    title: "资深前端工程师",
    company: "星环科技",
    level: "P6 / Senior",
    location: "上海 · 混合办公（周 3 天在办）",
    language: "中文",
    source: "https://star-ring.com/careers/frontend-sr",
    fetched: lang === "en" ? "just now" : "刚刚抓取",
    mustHave: [
      { t: "React 18 + TypeScript 生产经验", hit: true },
      { t: "5+ 年前端 / Web 开发", hit: true },
      { t: "性能优化与渲染底层理解", hit: "partial", note: lang === "en" ? "your case lacks quantified impact" : "你的案例缺量化结果" },
      { t: "大型组件库 / Design System 落地", hit: false },
    ],
    niceToHave: [
      { t: "可访问性（WAI-ARIA）经验", hit: true },
      { t: "Node.js / BFF 经验", hit: "partial" },
      { t: "技术写作 / 团队分享", hit: false },
    ],
    hidden: [
      lang === "en" ? "Cross-team influence — JD mentions 'collaborating with design/product/backend' 3x" : "跨团队协作 —— JD 里 3 次提到「与设计/产品/后端协作」",
      lang === "en" ? "Ownership — 'end-to-end delivery' language, expect solo-drive stories" : "Ownership 倾向 —— 出现「端到端交付」，预计会问独立推动的故事",
      lang === "en" ? "Ambiguity tolerance — startup at Series-B stage" : "对模糊性的容忍 —— B 轮阶段",
    ],
    rounds: [
      { r: lang === "en" ? "HR screen · 20m" : "HR 初筛 · 20 分钟", focus: lang === "en" ? "Motivation, timing, comp" : "动机 · 节奏 · 薪资" },
      { r: lang === "en" ? "Tech round 1 · 45m" : "技术一面 · 45 分钟", focus: lang === "en" ? "React internals, perf, TS" : "React 底层 · 性能 · TS" },
      { r: lang === "en" ? "Tech round 2 · 60m" : "技术二面 · 60 分钟", focus: lang === "en" ? "Architecture, trade-offs" : "架构 · 权衡" },
      { r: lang === "en" ? "Hiring manager · 40m" : "经理面 · 40 分钟", focus: lang === "en" ? "Influence, conflict" : "影响力 · 冲突" },
    ],
  });

  if (stage === "loading") {
    return (
      <div className="ei-fadein" style={{ minHeight: "calc(100vh - 58px)", display: "flex", alignItems: "center", justifyContent: "center", padding: 48 }}>
        <div style={{ maxWidth: 520, width: "100%" }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>
            {lang === "en" ? "PARSING · step 01 of 04" : "解析中 · 第 01 / 04 步"}
          </div>
          <div className="ei-serif" style={{ fontSize: 28, color: T.ink, letterSpacing: "-0.015em", lineHeight: 1.3, marginBottom: 32 }}>
            {lang === "en" ? "Turning the JD into mock interview context…" : "正在把这份 JD 变成模拟面试上下文…"}
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
            {steps.map((s, i) => {
              const done = i < step;
              const active = i === step;
              return (
                <div key={i} style={{ display: "flex", gap: 14, alignItems: "center" }}>
                  <div style={{
                    width: 22, height: 22, borderRadius: 11, border: `1.5px solid ${done ? T.ok : active ? T.accent : T.rule}`,
                    background: done ? T.ok : "transparent",
                    display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0,
                  }}>
                    {done && <Icon name="check" size={12} color="#fff" stroke={2.5} />}
                    {active && <div style={{ width: 6, height: 6, borderRadius: 3, background: T.accent }} className="ei-pulse" />}
                  </div>
                  <div style={{ fontSize: 14, color: done ? T.ink3 : active ? T.ink : T.ink4, textDecoration: done ? "line-through" : "none" }}>{s}</div>
                  {active && <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink4, marginLeft: "auto" }}>
                    <span className="ei-pulse">●</span> {lang === "en" ? "working" : "处理中"}
                  </div>}
                </div>
              );
            })}
          </div>
          <div style={{ marginTop: 40, paddingTop: 20, borderTop: `1px dotted ${T.rule}`, fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", lineHeight: 1.6 }}>
            <div>model · claude-haiku-4.5 · zh-CN</div>
            <div>rubric · jd-parse-v1.3 · prompt@a8f2e1</div>
            <div>typical · 3–6s · this one · slightly richer JD</div>
          </div>
        </div>
      </div>
    );
  }

  // Preview / confirm
  const setField = (k, v) => setParsed((p) => ({ ...p, [k]: v }));
  const toggleHit = (section, idx) => setParsed((p) => {
    const copy = { ...p, [section]: [...p[section]] };
    const cur = copy[section][idx];
    const cycle = { true: "partial", partial: false, false: true };
    copy[section][idx] = { ...cur, hit: cycle[String(cur.hit)] };
    return copy;
  });

  const HitDot = ({ hit }) => {
    const color = hit === true ? T.ok : hit === "partial" ? T.warn : T.ink4;
    const label = hit === true ? (lang === "en" ? "hit" : "命中") : hit === "partial" ? (lang === "en" ? "partial" : "部分") : (lang === "en" ? "gap" : "缺口");
    return (
      <div style={{ display: "flex", gap: 5, alignItems: "center", padding: "2px 7px", background: hit === true ? T.okSoft : hit === "partial" ? T.warnSoft : "transparent", border: hit === false ? `1px dashed ${T.rule}` : "none", borderRadius: 2 }}>
        <div style={{ width: 5, height: 5, borderRadius: 3, background: color }} />
        <span style={{ fontSize: 10.5, color, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em", textTransform: "uppercase" }}>{label}</span>
      </div>
    );
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 1200, margin: "0 auto", padding: "32px 48px 96px" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 24 }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
            {lang === "en" ? "STEP 2 OF 2 · REVIEW & CONFIRM" : "第 2 / 2 步 · 核对并确认"}
          </div>
          <h1 className="ei-serif" style={{ fontSize: 32, margin: 0, color: T.ink, letterSpacing: "-0.02em", lineHeight: 1.2 }}>
            {lang === "en" ? "Here's what I read from the JD." : "这是我从 JD 里读出来的内容。"}
          </h1>
          <div style={{ fontSize: 14, color: T.ink3, marginTop: 8, maxWidth: 620, lineHeight: 1.5 }}>
            {lang === "en"
              ? "Fix anything that's off. The generator will use exactly what you confirm here — no silent extras."
              : "看一眼有没有偏的。练习题生成器只会用你在这里确认的内容——不会偷偷加料。"}
          </div>
        </div>
        <div style={{ textAlign: "right" }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{lang === "en" ? "SOURCE" : "来源"}</div>
          <div style={{ fontSize: 12, fontFamily: "var(--ei-mono)", color: T.ink2, maxWidth: 280, wordBreak: "break-all" }}>
            {parsed.source}
          </div>
          <div style={{ fontSize: 11, color: T.ink3, marginTop: 4 }}>{parsed.fetched}</div>
        </div>
      </div>

      {/* Basic fields */}
      <Card T={T} pad={0} style={{ marginBottom: 20 }}>
        <div style={{ padding: "16px 24px", borderBottom: `1px solid ${T.rule}` }}>
          <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "BASICS" : "基础信息"}</div>
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(2, 1fr)", padding: "6px 24px" }}>
          {[
            { k: lang === "en" ? "Title" : "岗位名", v: parsed.title, f: "title" },
            { k: lang === "en" ? "Company" : "公司", v: parsed.company, f: "company" },
            { k: lang === "en" ? "Level" : "职级", v: parsed.level, f: "level" },
            { k: lang === "en" ? "Location" : "地点", v: parsed.location, f: "location" },
            { k: lang === "en" ? "Language" : "语言", v: parsed.language, f: "language" },
          ].map((r, i) => (
            <div key={r.f} style={{ display: "flex", gap: 14, padding: "12px 0", borderBottom: i < 3 ? `1px dotted ${T.rule}` : "none", alignItems: "baseline" }}>
              <div className="ei-label" style={{ color: T.ink3, minWidth: 68, fontSize: 10.5 }}>{r.k}</div>
              <input value={r.v} onChange={(e) => setField(r.f, e.target.value)} style={{ flex: 1, fontSize: 14, color: T.ink, background: "transparent", border: "none", outline: "none", borderBottom: `1px dashed transparent`, padding: "2px 0", fontFamily: "var(--ei-sans)" }} onFocus={(e) => e.target.style.borderBottomColor = T.accent} onBlur={(e) => e.target.style.borderBottomColor = "transparent"} />
            </div>
          ))}
        </div>
      </Card>

      {/* Requirements */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20, marginBottom: 20 }}>
        <RequirementBlock T={T} lang={lang} title={lang === "en" ? "MUST HAVE" : "必需项"} items={parsed.mustHave} onToggle={(i) => toggleHit("mustHave", i)} HitDot={HitDot} />
        <RequirementBlock T={T} lang={lang} title={lang === "en" ? "NICE TO HAVE" : "加分项"} items={parsed.niceToHave} onToggle={(i) => toggleHit("niceToHave", i)} HitDot={HitDot} />
      </div>

      {/* Hidden signals */}
      <Card T={T} style={{ marginBottom: 20, borderColor: T.accent }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14 }}>
          <div>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 4 }}>{lang === "en" ? "HIDDEN SIGNALS · inference" : "隐性关注点 · 推断"}</div>
            <div style={{ fontSize: 13, color: T.ink3 }}>{lang === "en" ? "Not stated, but JD language suggests —" : "JD 没明说，但字里行间暗示的——"}</div>
          </div>
          <Tag T={T} tone="accent">
            <Icon name="sparkle" size={10} style={{ marginRight: 4 }} />
            {lang === "en" ? "confidence · medium" : "置信度 · 中"}
          </Tag>
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {parsed.hidden.map((h, i) => (
            <div key={i} style={{ display: "flex", gap: 10, alignItems: "flex-start", padding: "8px 12px", background: T.bgSoft, borderRadius: 2 }}>
              <Icon name="sparkle" size={12} color={T.accent} style={{ marginTop: 3, flexShrink: 0 }} />
              <div style={{ fontSize: 13.5, color: T.ink, lineHeight: 1.5, flex: 1 }}>{h}</div>
              <button style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", fontSize: 11, fontFamily: "var(--ei-mono)" }}>
                {lang === "en" ? "remove" : "移除"}
              </button>
            </div>
          ))}
        </div>
      </Card>

      {/* Round assumptions */}
      <Card T={T} style={{ marginBottom: 28 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14 }}>
          <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "ROUND ASSUMPTIONS" : "轮次假设"}</div>
          <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
            {lang === "en" ? "editable after creation" : "创建后仍可编辑"}
          </div>
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 10 }}>
          {parsed.rounds.map((r, i) => (
            <div key={i} style={{ padding: "12px 14px", background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2, position: "relative" }}>
              <div style={{ fontFamily: "var(--ei-mono)", fontSize: 10.5, color: T.ink4, marginBottom: 5, letterSpacing: "0.06em" }}>R{i + 1}</div>
              <div style={{ fontSize: 13, color: T.ink, fontWeight: 500, marginBottom: 4 }}>{r.r}</div>
              <div style={{ fontSize: 11.5, color: T.ink3, lineHeight: 1.45 }}>{r.focus}</div>
            </div>
          ))}
        </div>
      </Card>

      {/* Footer actions */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "16px 0", borderTop: `1px solid ${T.rule}` }}>
        <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", lineHeight: 1.6, maxWidth: 420 }}>
          {lang === "en" ? "The interview setup will use these assumptions. You can edit anything later." : "面试前确认会使用上面这些信息，之后随时可改。"}
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          <Btn T={T} variant="ghost" onClick={() => nav("home")}>{lang === "en" ? "Cancel" : "取消"}</Btn>
          <Btn T={T} variant="secondary" icon="edit">{lang === "en" ? "Re-parse" : "重新解析"}</Btn>
          <Btn T={T} variant="accent" iconRight="arrow_right" onClick={() => nav("workspace", { jobId: "tj-1" })}>
            {lang === "en" ? "Confirm & open interview setup" : "确认并进入面试前确认"}
          </Btn>
        </div>
      </div>
    </div>
  );
};

const RequirementBlock = ({ T, lang, title, items, onToggle, HitDot }) => (
  <Card T={T} pad={0}>
    <div style={{ padding: "14px 20px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
      <div className="ei-label" style={{ color: T.ink3 }}>{title}</div>
      <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{items.length}</div>
    </div>
    <div>
      {items.map((item, i) => (
        <div key={i} style={{ padding: "12px 20px", borderBottom: i < items.length - 1 ? `1px dotted ${T.rule}` : "none", display: "flex", gap: 12, alignItems: "flex-start" }}>
          <button onClick={() => onToggle(i)} style={{ background: "transparent", border: "none", padding: 0, cursor: "pointer", marginTop: 2 }}>
            <HitDot hit={item.hit} />
          </button>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 13.5, color: T.ink, lineHeight: 1.45 }}>{item.t}</div>
            {item.note && <div style={{ fontSize: 11.5, color: T.warn, marginTop: 4, fontStyle: "italic" }}>{item.note}</div>}
          </div>
        </div>
      ))}
    </div>
  </Card>
);

// ═══════════════════════════════════════════════════════════════════
// #2 ONBOARDING — M1-lite progressive profile
// ═══════════════════════════════════════════════════════════════════
const OnboardingScreen = ({ T, lang, nav }) => {
  const [step, setStep] = React.useState(0);
  const total = 4;

  const [data, setData] = React.useState({
    targetRole: "",
    years: 5,
    locations: ["上海"],
    languages: ["中文", "English"],
    resumeMode: null, // "upload" | "paste" | "skip"
    resumePasted: "",
    experiences: [],
  });

  const next = () => step < total - 1 ? setStep(step + 1) : nav("home");
  const back = () => step > 0 ? setStep(step - 1) : nav("home");

  const steps = [
    { t: lang === "en" ? "Goal" : "目标", d: lang === "en" ? "Who are you looking to be?" : "你想成为谁？" },
    { t: lang === "en" ? "Context" : "背景", d: lang === "en" ? "A bit about your experience" : "说说你的情况" },
    { t: lang === "en" ? "Resume" : "简历", d: lang === "en" ? "Optional — you can start without one" : "可选——没有也能开始练" },
    { t: lang === "en" ? "Stories" : "故事", d: lang === "en" ? "Stories we found in your resume" : "从你简历里挑出来的故事" },
  ];

  return (
    <div className="ei-fadein" style={{ minHeight: "calc(100vh - 58px)", display: "grid", gridTemplateColumns: "300px 1fr", background: T.bg }}>
      {/* Left rail */}
      <div style={{ background: T.bgSoft, borderRight: `1px solid ${T.rule}`, padding: "40px 28px", display: "flex", flexDirection: "column" }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
          {lang === "en" ? "FIRST 5 MINUTES" : "最初 5 分钟"}
        </div>
        <div className="ei-serif" style={{ fontSize: 22, color: T.ink, letterSpacing: "-0.015em", marginBottom: 6, lineHeight: 1.3 }}>
          {lang === "en" ? "Just enough to start." : "先问够用的那点。"}
        </div>
        <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.55, marginBottom: 32 }}>
          {lang === "en" ? "We'll fill the rest as you practice. No long profile required up-front." : "剩下的在练习里慢慢补。不用先填完所有东西。"}
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: 0, flex: 1 }}>
          {steps.map((s, i) => {
            const done = i < step;
            const active = i === step;
            return (
              <div key={i} style={{ display: "flex", gap: 12, padding: "14px 0", alignItems: "flex-start", borderBottom: i < steps.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
                <div style={{
                  width: 24, height: 24, borderRadius: 12, flexShrink: 0,
                  border: `1.5px solid ${done ? T.ok : active ? T.accent : T.rule}`,
                  background: done ? T.ok : "transparent",
                  display: "flex", alignItems: "center", justifyContent: "center",
                  fontSize: 11, fontFamily: "var(--ei-mono)",
                  color: done ? "#fff" : active ? T.accent : T.ink4,
                  marginTop: 1,
                }}>
                  {done ? <Icon name="check" size={13} color="#fff" stroke={2.5} /> : String(i + 1).padStart(2, "0")}
                </div>
                <div>
                  <div style={{ fontSize: 13.5, color: active ? T.ink : done ? T.ink3 : T.ink4, fontWeight: active ? 500 : 400 }}>{s.t}</div>
                  <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 2 }}>{s.d}</div>
                </div>
              </div>
            );
          })}
        </div>

        <button onClick={() => nav("home")} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 12, textAlign: "left", padding: 0, cursor: "pointer", marginTop: 24 }}>
          {lang === "en" ? "Skip all · go to inbox →" : "全部跳过，直接进收件箱 →"}
        </button>
      </div>

      {/* Right content */}
      <div style={{ padding: "56px 72px", maxWidth: 760, overflowY: "auto" }} className="ei-scroll">
        {step === 0 && <OnbStepGoal T={T} lang={lang} data={data} setData={setData} />}
        {step === 1 && <OnbStepContext T={T} lang={lang} data={data} setData={setData} />}
        {step === 2 && <OnbStepResume T={T} lang={lang} data={data} setData={setData} />}
        {step === 3 && <OnbStepStories T={T} lang={lang} data={data} setData={setData} />}

        <div style={{ display: "flex", justifyContent: "space-between", marginTop: 48, paddingTop: 20, borderTop: `1px solid ${T.rule}` }}>
          <Btn T={T} variant="ghost" onClick={back}>{step === 0 ? (lang === "en" ? "Cancel" : "取消") : (lang === "en" ? "Back" : "上一步")}</Btn>
          <div style={{ display: "flex", gap: 10 }}>
            {step === 2 && data.resumeMode == null && (
              <Btn T={T} variant="secondary" onClick={() => { setData({ ...data, resumeMode: "skip" }); next(); }}>
                {lang === "en" ? "Start without resume →" : "没有简历也开始 →"}
              </Btn>
            )}
            <Btn T={T} variant="accent" iconRight="arrow_right" onClick={next}>
              {step === total - 1 ? (lang === "en" ? "Finish & start" : "完成并开始") : (lang === "en" ? "Continue" : "下一步")}
            </Btn>
          </div>
        </div>
      </div>
    </div>
  );
};

const OnbStepGoal = ({ T, lang, data, setData }) => {
  const roles = lang === "en"
    ? ["Frontend engineer", "Senior frontend", "Full-stack", "Platform / infra", "Product designer", "PM", "Other"]
    : ["前端工程师", "资深前端", "全栈工程师", "平台 / 基础架构", "产品设计师", "产品经理", "其它"];
  return (
    <div>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>01 · {lang === "en" ? "GOAL" : "目标"}</div>
      <h2 className="ei-serif" style={{ fontSize: 30, margin: "0 0 10px", letterSpacing: "-0.02em", color: T.ink, lineHeight: 1.2 }}>
        {lang === "en" ? "What role are you interviewing for?" : "你要面的是什么角色？"}
      </h2>
      <div style={{ fontSize: 14, color: T.ink3, marginBottom: 32, lineHeight: 1.5 }}>
        {lang === "en" ? "Pick one. We'll use this to tone the questions — you can change it per target job later." : "先挑一个。生成题目时会参考它——每个具体岗位都还能单独改。"}
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(2, 1fr)", gap: 10, marginBottom: 32 }}>
        {roles.map((r) => {
          const sel = data.targetRole === r;
          return (
            <button key={r} onClick={() => setData({ ...data, targetRole: r })} style={{
              padding: "14px 16px", textAlign: "left", cursor: "pointer",
              background: sel ? T.accentSoft : T.bgCard,
              border: `1px solid ${sel ? T.accent : T.rule}`, borderRadius: 2,
              fontSize: 14, color: sel ? T.accent : T.ink, fontWeight: sel ? 500 : 400, fontFamily: "var(--ei-sans)",
              display: "flex", justifyContent: "space-between", alignItems: "center",
            }}>
              <span>{r}</span>
              {sel && <Icon name="check" size={14} color={T.accent} stroke={2.5} />}
            </button>
          );
        })}
      </div>

      <div>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
          {lang === "en" ? "OR — describe it in your own words" : "或——用你自己的话说"}
        </div>
        <input
          value={data.targetRole.startsWith("·") ? data.targetRole.slice(2) : ""}
          onChange={(e) => setData({ ...data, targetRole: "· " + e.target.value })}
          placeholder={lang === "en" ? "e.g. Growth engineer at early-stage SaaS" : "比如：早期 SaaS 的增长工程师"}
          style={{ width: "100%", padding: "12px 14px", border: `1px solid ${T.rule}`, borderRadius: 2, fontSize: 14, color: T.ink, background: T.bgCard, fontFamily: "var(--ei-sans)", outline: "none", boxSizing: "border-box" }}
        />
      </div>
    </div>
  );
};

const OnbStepContext = ({ T, lang, data, setData }) => (
  <div>
    <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>02 · {lang === "en" ? "CONTEXT" : "背景"}</div>
    <h2 className="ei-serif" style={{ fontSize: 30, margin: "0 0 10px", letterSpacing: "-0.02em", color: T.ink, lineHeight: 1.2 }}>
      {lang === "en" ? "Where are you in your career?" : "你现在处在什么位置？"}
    </h2>
    <div style={{ fontSize: 14, color: T.ink3, marginBottom: 32, lineHeight: 1.5 }}>
      {lang === "en" ? "Rough signals — helps us calibrate difficulty." : "大概说说——用来校准题目的难度。"}
    </div>

    <div style={{ marginBottom: 32 }}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>
        {lang === "en" ? "YEARS IN THE FIELD" : "工作年限"}
      </div>
      <div style={{ display: "flex", gap: 12, alignItems: "center" }}>
        <input type="range" min="0" max="20" value={data.years} onChange={(e) => setData({ ...data, years: Number(e.target.value) })} style={{ flex: 1, accentColor: T.accent }} />
        <div style={{ fontFamily: "var(--ei-mono)", fontSize: 18, color: T.ink, minWidth: 60 }}>
          {data.years}{data.years >= 20 ? "+" : ""} <span style={{ fontSize: 13, color: T.ink3 }}>{lang === "en" ? "yr" : "年"}</span>
        </div>
      </div>
      <div style={{ display: "flex", justifyContent: "space-between", fontSize: 11, color: T.ink4, fontFamily: "var(--ei-mono)", marginTop: 4 }}>
        <span>{lang === "en" ? "new grad" : "应届"}</span><span>junior</span><span>mid</span><span>senior</span><span>staff+</span>
      </div>
    </div>

    <div style={{ marginBottom: 32 }}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>
        {lang === "en" ? "TARGET LOCATION" : "目标地点"}
      </div>
      <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
        {["上海", "北京", "深圳", "杭州", "Remote · Asia", "Remote · Global", "Singapore", "Berlin"].map((loc) => {
          const sel = data.locations.includes(loc);
          return (
            <button key={loc} onClick={() => setData({ ...data, locations: sel ? data.locations.filter((l) => l !== loc) : [...data.locations, loc] })}
              style={{ padding: "6px 12px", borderRadius: 14, border: `1px solid ${sel ? T.accent : T.rule}`, background: sel ? T.accentSoft : "transparent", color: sel ? T.accent : T.ink2, fontSize: 12.5, cursor: "pointer", fontFamily: "var(--ei-sans)" }}>
              {loc}
            </button>
          );
        })}
      </div>
    </div>

    <div>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>
        {lang === "en" ? "INTERVIEW LANGUAGE(S)" : "面试语言"}
      </div>
      <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
        {["中文", "English", "日本語", "Deutsch", "Français"].map((lg) => {
          const sel = data.languages.includes(lg);
          return (
            <button key={lg} onClick={() => setData({ ...data, languages: sel ? data.languages.filter((l) => l !== lg) : [...data.languages, lg] })}
              style={{ padding: "6px 12px", borderRadius: 14, border: `1px solid ${sel ? T.accent : T.rule}`, background: sel ? T.accentSoft : "transparent", color: sel ? T.accent : T.ink2, fontSize: 12.5, cursor: "pointer", fontFamily: "var(--ei-sans)" }}>
              {lg}
            </button>
          );
        })}
      </div>
    </div>
  </div>
);

const OnbStepResume = ({ T, lang, data, setData }) => (
  <div>
    <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>03 · {lang === "en" ? "RESUME" : "简历"}</div>
    <h2 className="ei-serif" style={{ fontSize: 30, margin: "0 0 10px", letterSpacing: "-0.02em", color: T.ink, lineHeight: 1.2 }}>
      {lang === "en" ? "Give us something to work with — or don't." : "给我点材料——或者不给也行。"}
    </h2>
    <div style={{ fontSize: 14, color: T.ink3, marginBottom: 32, lineHeight: 1.5 }}>
      {lang === "en" ? "Everything you put in here stays private to your account, and we'll pull stories out so you don't have to retype them every session." : "放进来的内容只属于你自己的账号；我们会从中挑出故事，下次练习就不用重打一遍。"}
    </div>

    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14, marginBottom: 24 }}>
      <button onClick={() => setData({ ...data, resumeMode: "upload" })} style={{
        padding: "28px 20px", cursor: "pointer", textAlign: "left",
        background: data.resumeMode === "upload" ? T.accentSoft : T.bgCard,
        border: `1px dashed ${data.resumeMode === "upload" ? T.accent : T.rule}`, borderRadius: 3,
      }}>
        <Icon name="upload" size={22} color={data.resumeMode === "upload" ? T.accent : T.ink3} />
        <div className="ei-serif" style={{ fontSize: 17, color: T.ink, marginTop: 12, marginBottom: 4 }}>
          {lang === "en" ? "Upload file" : "上传文件"}
        </div>
        <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>.pdf · .docx · .md · max 5 MB</div>
      </button>
      <button onClick={() => setData({ ...data, resumeMode: "paste" })} style={{
        padding: "28px 20px", cursor: "pointer", textAlign: "left",
        background: data.resumeMode === "paste" ? T.accentSoft : T.bgCard,
        border: `1px dashed ${data.resumeMode === "paste" ? T.accent : T.rule}`, borderRadius: 3,
      }}>
        <Icon name="edit" size={22} color={data.resumeMode === "paste" ? T.accent : T.ink3} />
        <div className="ei-serif" style={{ fontSize: 17, color: T.ink, marginTop: 12, marginBottom: 4 }}>
          {lang === "en" ? "Paste a summary" : "粘贴履历摘要"}
        </div>
        <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{lang === "en" ? "bullet points, linkedin dump, anything" : "几条经历 · LinkedIn 导出 · 都行"}</div>
      </button>
    </div>

    {data.resumeMode === "paste" && (
      <div style={{ marginBottom: 24 }}>
        <textarea
          value={data.resumePasted}
          onChange={(e) => setData({ ...data, resumePasted: e.target.value })}
          placeholder={lang === "en" ? "• Senior Frontend @ Star-Ring · 2022-now\n  led cart/checkout rewrite, LCP 3.2s → 1.4s\n• Frontend @ Lumen · 2019-2022\n  …" : "• 星环科技 · 资深前端 · 2022 至今\n  主导购物车 / 结账链路重写，LCP 3.2s → 1.4s\n• Lumen · 前端 · 2019-2022\n  …"}
          rows={9}
          style={{ width: "100%", padding: 16, fontFamily: "var(--ei-mono)", fontSize: 12.5, color: T.ink, background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, resize: "vertical", lineHeight: 1.6, outline: "none", boxSizing: "border-box" }}
        />
      </div>
    )}

    {data.resumeMode === "upload" && (
      <div style={{ padding: 20, background: T.okSoft, borderRadius: 2, marginBottom: 24, display: "flex", gap: 12, alignItems: "center" }}>
        <Icon name="check" size={18} color={T.ok} />
        <div>
          <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>lin-zhou-resume-2026.pdf</div>
          <div style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 2 }}>
            247 KB · 2 pages · {lang === "en" ? "ready to parse" : "准备解析"}
          </div>
        </div>
      </div>
    )}

    <div style={{ fontSize: 11.5, color: T.ink3, lineHeight: 1.6, fontFamily: "var(--ei-mono)", padding: "12px 14px", background: T.bgSoft, borderLeft: `2px solid ${T.rule}` }}>
      {lang === "en" ? "⏷ Resume is used only to suggest stories during practice. Delete it anytime in Settings." : "⏷ 简历只在练习时用来推荐故事，随时可在设置里删除。"}
    </div>
  </div>
);

const OnbStepStories = ({ T, lang, data, setData }) => {
  // Mock extracted experience cards
  const [cards, setCards] = React.useState([
    { id: "c1", title: lang === "en" ? "Cart / checkout rewrite" : "购物车/结账链路重写", company: "星环科技 · 2024", s: lang === "en" ? "Legacy stack, LCP 3.2s, ~8% abandon on page 2" : "老栈、LCP 3.2s、第二步流失 ~8%", a: lang === "en" ? "Led migration to RSC + selective hydration" : "主导迁移到 RSC + 选择性注水", r: lang === "en" ? "LCP → 1.4s, abandon → 4.2%" : "LCP → 1.4s，流失 → 4.2%", tags: lang === "en" ? ["perf", "ownership", "trade-offs"] : ["性能", "Ownership", "权衡"], keep: true },
    { id: "c2", title: lang === "en" ? "Design System rollout" : "Design System 在 5 个产品的落地", company: "星环科技 · 2023", s: lang === "en" ? "Fragmented UI across 5 products" : "5 个产品 UI 各自为政", a: lang === "en" ? "Built tokens + primitives, ran adoption workshops" : "搭 tokens + 基础组件 + 多次推广", r: lang === "en" ? "4 / 5 products on v1 in 6 mo" : "6 个月内 4/5 产品迁到 v1", tags: lang === "en" ? ["influence", "cross-team", "design"] : ["影响力", "跨团队", "设计"], keep: true },
    { id: "c3", title: lang === "en" ? "Disagreement with designer" : "与设计师的方案分歧", company: "星环科技 · 2023", s: lang === "en" ? "Designer wanted 40 fields on one screen" : "设计想把 40 个字段放一屏", a: lang === "en" ? "Ran shadow-session with 6 operators, proposed tabs" : "跟 6 个操作员做影子观察，改提标签页", r: lang === "en" ? "Task time -42%, pushed through in 2 wks" : "操作耗时 -42%，两周推完", tags: lang === "en" ? ["conflict", "empathy", "data"] : ["冲突", "共情", "数据"], keep: true },
    { id: "c4", title: lang === "en" ? "Failed A/B on checkout CTA" : "结账 CTA 的失败实验", company: "Lumen · 2021", s: lang === "en" ? "Hypothesized that color would move CVR" : "以为换色能提转化", a: lang === "en" ? "Shipped, ran for 2 wks, no signal" : "上线、跑两周、无显著变化", r: lang === "en" ? "Learned to lead with priors not hunches" : "学会先看先验、再做假设", tags: lang === "en" ? ["failure", "learning"] : ["失败", "反思"], keep: false },
  ]);

  const toggleKeep = (id) => setCards(cards.map((c) => c.id === id ? { ...c, keep: !c.keep } : c));

  return (
    <div>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>04 · {lang === "en" ? "STORIES" : "故事"}</div>
      <h2 className="ei-serif" style={{ fontSize: 30, margin: "0 0 10px", letterSpacing: "-0.02em", color: T.ink, lineHeight: 1.2 }}>
        {lang === "en" ? "Four stories I pulled from your resume." : "从你简历里挑了四个故事。"}
      </h2>
      <div style={{ fontSize: 14, color: T.ink3, marginBottom: 28, lineHeight: 1.5 }}>
        {lang === "en" ? "Keep the ones that feel accurate. I'll use these when a question calls for a concrete example." : "留下那些你觉得没跑偏的。面试题需要具体案例时我会从这里取。"}
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {cards.map((c) => (
          <div key={c.id} style={{
            padding: "16px 20px",
            background: c.keep ? T.bgCard : T.bgSoft,
            border: `1px solid ${c.keep ? T.rule : T.rule}`,
            borderLeft: `3px solid ${c.keep ? T.accent : T.ink4}`,
            borderRadius: 2, opacity: c.keep ? 1 : 0.6,
          }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 10, gap: 12 }}>
              <div>
                <div className="ei-serif" style={{ fontSize: 16, color: T.ink, fontWeight: 500, letterSpacing: "-0.01em" }}>{c.title}</div>
                <div style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 3, letterSpacing: "0.04em" }}>{c.company}</div>
              </div>
              <button onClick={() => toggleKeep(c.id)} style={{
                padding: "4px 10px", fontSize: 11.5, borderRadius: 2, cursor: "pointer",
                border: `1px solid ${c.keep ? T.ok : T.rule}`,
                background: c.keep ? T.okSoft : "transparent", color: c.keep ? T.ok : T.ink3,
                fontFamily: "var(--ei-mono)", letterSpacing: "0.04em",
              }}>
                {c.keep ? (lang === "en" ? "✓ KEEP" : "✓ 保留") : (lang === "en" ? "SKIP" : "跳过")}
              </button>
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 10, marginBottom: 10 }}>
              {[
                { k: "S", v: c.s, label: lang === "en" ? "Situation" : "情境" },
                { k: "A", v: c.a, label: lang === "en" ? "Action" : "行动" },
                { k: "R", v: c.r, label: lang === "en" ? "Result" : "结果" },
              ].map((x) => (
                <div key={x.k} style={{ fontSize: 12.5, color: T.ink2, lineHeight: 1.5 }}>
                  <span style={{ fontFamily: "var(--ei-mono)", fontSize: 10, color: T.ink4, letterSpacing: "0.08em", marginRight: 6 }}>{x.k} · {x.label.toUpperCase()}</span>
                  <div style={{ marginTop: 3 }}>{x.v}</div>
                </div>
              ))}
            </div>
            <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
              {c.tags.map((t) => <Tag key={t} T={T} tone="neutral">{t}</Tag>)}
            </div>
          </div>
        ))}
      </div>

      <div style={{ fontSize: 12, color: T.ink3, marginTop: 20, fontFamily: "var(--ei-mono)", lineHeight: 1.6 }}>
        {lang === "en" ? `kept · ${cards.filter(c => c.keep).length} / ${cards.length} · you can edit or add more from Profile later` : `保留 · ${cards.filter(c => c.keep).length} / ${cards.length} · 之后在画像里还能编辑或添加`}
      </div>
    </div>
  );
};

// ═══════════════════════════════════════════════════════════════════
// #3 REPORT GENERATING — async intermediate screen
// ═══════════════════════════════════════════════════════════════════
const ReportGeneratingScreen = ({ T, lang, nav }) => {
  const [phase, setPhase] = React.useState(0);
  const phases = lang === "en" ? [
    { t: "Transcribing & aligning turns", s: 900, hint: "8 questions · 23 turns" },
    { t: "Extracting evidence per question", s: 1200, hint: "looking for S/A/R + quantification" },
    { t: "Scoring against rubric", s: 900, hint: "rubric · behavior-v2.1 · confidence tagged" },
    { t: "Clustering into mistake entries", s: 700, hint: "3 flagged for mistake book" },
    { t: "Writing recommendations", s: 900, hint: "frameworks + mapped experiences" },
  ] : [
    { t: "转写并对齐对话", s: 900, hint: "8 题 · 23 轮对话" },
    { t: "逐题抽取证据", s: 1200, hint: "寻找 S/A/R + 量化结果" },
    { t: "按 rubric 评分", s: 900, hint: "rubric · behavior-v2.1 · 带置信度" },
    { t: "聚类为错题条目", s: 700, hint: "识别 3 条入本" },
    { t: "生成建议", s: 900, hint: "推荐回答框架 + 映射经历" },
  ];

  React.useEffect(() => {
    let cancel = false;
    let acc = 0;
    phases.forEach((p, i) => {
      acc += p.s;
      setTimeout(() => { if (!cancel) setPhase(i + 1); }, acc);
    });
    setTimeout(() => { if (!cancel) nav("report"); }, acc + 600);
    return () => { cancel = true; };
  }, []);

  const pct = Math.min(100, (phase / phases.length) * 100);

  // Live "evidence" fragments appearing
  const liveSnippets = lang === "en" ? [
    '"LCP from 3.2s to 1.4s" → Q2 · evidence · high confidence',
    '"We — actually I — drove the migration" → Q5 · self-correction observed',
    '"they just agreed" → Q6 · conflict resolution lacks depth · flag',
    'reverse-Q round · 1 question asked, 3 recommended minimum',
  ] : [
    '"LCP 从 3.2s 到 1.4s" → 第 2 题 · 证据 · 高置信度',
    '"我们——其实主要是我——推动迁移" → 第 5 题 · 观察到自我修正',
    '"他就同意了" → 第 6 题 · 冲突解决缺乏深度 · 标记',
    '反问环节 · 实际问了 1 题，推荐至少 3 题',
  ];

  const [shown, setShown] = React.useState([]);
  React.useEffect(() => {
    let cancel = false;
    liveSnippets.forEach((s, i) => {
      setTimeout(() => { if (!cancel) setShown((prev) => [...prev, s]); }, 700 + i * 800);
    });
    return () => { cancel = true; };
  }, []);

  return (
    <div className="ei-fadein" style={{ minHeight: "calc(100vh - 58px)", background: T.bg, display: "flex", alignItems: "center", justifyContent: "center", padding: 48 }}>
      <div style={{ maxWidth: 780, width: "100%" }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 12, letterSpacing: "0.1em" }}>
          {lang === "en" ? "GENERATING REPORT · ASYNC" : "报告生成中 · 异步"}
        </div>
        <h1 className="ei-serif" style={{ fontSize: 34, margin: 0, color: T.ink, letterSpacing: "-0.02em", lineHeight: 1.2, marginBottom: 10 }}>
          {lang === "en" ? "Reading every turn. Evidence first." : "在逐轮读——先找证据，再打分。"}
        </h1>
        <div style={{ fontSize: 14, color: T.ink3, marginBottom: 32, lineHeight: 1.5, maxWidth: 540 }}>
          {lang === "en" ? "Typical: 8-15s. You can close this tab — the report lands in your inbox when it's done." : "通常 8-15 秒。可以关掉这个页面——好了会出现在收件箱里。"}
        </div>

        {/* Progress */}
        <div style={{ marginBottom: 32 }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", marginBottom: 6 }}>
            <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink3, letterSpacing: "0.04em" }}>
              {phase} / {phases.length} · {phase < phases.length ? phases[phase]?.t : (lang === "en" ? "done" : "完成")}
            </div>
            <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink3 }}>{Math.round(pct)}%</div>
          </div>
          <div style={{ height: 2, background: T.rule, overflow: "hidden" }}>
            <div style={{ height: "100%", width: `${pct}%`, background: T.accent, transition: "width .5s ease" }} />
          </div>
        </div>

        {/* Phases */}
        <div style={{ display: "flex", flexDirection: "column", gap: 0, marginBottom: 32 }}>
          {phases.map((p, i) => {
            const done = i < phase;
            const active = i === phase;
            return (
              <div key={i} style={{ display: "flex", gap: 12, padding: "10px 0", borderBottom: i < phases.length - 1 ? `1px dotted ${T.rule}` : "none", alignItems: "center" }}>
                <div style={{
                  width: 18, height: 18, borderRadius: 9, flexShrink: 0,
                  background: done ? T.ok : active ? T.accent : "transparent",
                  border: `1.5px solid ${done ? T.ok : active ? T.accent : T.rule}`,
                  display: "flex", alignItems: "center", justifyContent: "center",
                }}>
                  {done && <Icon name="check" size={11} color="#fff" stroke={2.5} />}
                  {active && <div style={{ width: 5, height: 5, borderRadius: 3, background: "#fff" }} className="ei-pulse" />}
                </div>
                <div style={{ fontSize: 13.5, color: done ? T.ink3 : active ? T.ink : T.ink4, flex: 1, textDecoration: done ? "line-through" : "none" }}>
                  {p.t}
                </div>
                <div style={{ fontSize: 11, color: T.ink4, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>
                  {active ? <span className="ei-pulse">●</span> : ""} {p.hint}
                </div>
              </div>
            );
          })}
        </div>

        {/* Live evidence stream */}
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "LIVE OBSERVATIONS" : "实时观察"}</div>
          <div style={{ padding: "14px 16px", background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2, fontFamily: "var(--ei-mono)", fontSize: 12, lineHeight: 1.75, color: T.ink2, minHeight: 100 }}>
            {shown.map((s, i) => (
              <div key={i} className="ei-fadein" style={{ marginBottom: 4 }}>
                <span style={{ color: T.ink4 }}>›</span> {s}
              </div>
            ))}
            {shown.length < liveSnippets.length && (
              <div style={{ display: "inline-block", width: 8, height: 12, background: T.accent, verticalAlign: "text-bottom" }} className="ei-pulse" />
            )}
          </div>
        </div>

        <div style={{ marginTop: 28, paddingTop: 16, borderTop: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>
            {lang === "en" ? "target p95 · <12s · you can safely leave" : "P95 目标 · <12s · 可以安心离开"}
          </div>
          <Btn T={T} variant="ghost" size="sm" onClick={() => nav("home")}>
            {lang === "en" ? "Notify me when ready →" : "好了通知我 →"}
          </Btn>
        </div>
      </div>
    </div>
  );
};

// ═══════════════════════════════════════════════════════════════════
// #8 SETTINGS / PRIVACY / DATA EXPORT & DELETE
// ═══════════════════════════════════════════════════════════════════
const SettingsScreen = ({ T, lang, nav }) => {
  const [tab, setTab] = React.useState("profile");

  const tabs = lang === "en"
    ? [{ k: "profile", t: "Profile" }, { k: "privacy", t: "Privacy & data" }, { k: "notifications", t: "Notifications" }, { k: "billing", t: "Billing" }]
    : [{ k: "profile", t: "个人资料" }, { k: "privacy", t: "隐私与数据" }, { k: "notifications", t: "通知" }, { k: "billing", t: "订阅" }];

  return (
    <div className="ei-fadein" style={{ maxWidth: 1040, margin: "0 auto", padding: "40px 48px 96px" }}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "SETTINGS" : "设置"}</div>
      <h1 className="ei-serif" style={{ fontSize: 36, margin: 0, color: T.ink, letterSpacing: "-0.02em", marginBottom: 32 }}>
        {lang === "en" ? "Your account & what we keep." : "账号与我们在你这里存的东西。"}
      </h1>

      <div style={{ display: "grid", gridTemplateColumns: "200px 1fr", gap: 40 }}>
        {/* Tab rail */}
        <div style={{ display: "flex", flexDirection: "column", gap: 0 }}>
          {tabs.map((t) => (
            <button key={t.k} onClick={() => setTab(t.k)} style={{
              background: "transparent", border: "none", textAlign: "left",
              padding: "10px 12px", cursor: "pointer", borderRadius: 2, fontFamily: "var(--ei-sans)",
              color: tab === t.k ? T.ink : T.ink3,
              fontWeight: tab === t.k ? 500 : 400, fontSize: 14,
              borderLeft: `2px solid ${tab === t.k ? T.accent : "transparent"}`,
            }}>
              {t.t}
            </button>
          ))}
        </div>

        {/* Content */}
        <div>
          {tab === "privacy" && <SettingsPrivacy T={T} lang={lang} />}
          {tab === "profile" && <SettingsProfile T={T} lang={lang} />}
          {tab === "notifications" && <SettingsNotif T={T} lang={lang} />}
          {tab === "billing" && <SettingsBilling T={T} lang={lang} />}
        </div>
      </div>
    </div>
  );
};

const SettingsPrivacy = ({ T, lang }) => {
  const [toggles, setToggles] = React.useState({ audio: true, transcript: true, resume: true, anon: false, emails: true });
  const Toggle = ({ k }) => (
    <button onClick={() => setToggles({ ...toggles, [k]: !toggles[k] })} style={{
      width: 34, height: 18, borderRadius: 9, border: "none", cursor: "pointer",
      background: toggles[k] ? T.accent : T.rule, position: "relative", flexShrink: 0,
    }}>
      <div style={{ position: "absolute", top: 2, left: toggles[k] ? 18 : 2, width: 14, height: 14, borderRadius: 7, background: "#fff", transition: "left .15s" }} />
    </button>
  );

  return (
    <div>
      {/* Data retention */}
      <section style={{ marginBottom: 40 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "DATA RETENTION" : "数据留存"}</div>
        <div style={{ fontSize: 13.5, color: T.ink3, marginBottom: 18, lineHeight: 1.55 }}>
          {lang === "en" ? "Everything is keep-by-default-you-can-delete — not the other way around. These toggles tell the system what it's allowed to touch." : "默认保留、你随时能删——不是反过来。下面的开关决定系统能碰哪些数据。"}
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 0, background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
          {[
            { k: "audio", t: lang === "en" ? "Keep voice recordings (session only)" : "保留语音录音（仅当次会话）", d: lang === "en" ? "Deleted within 24h of the report. Off → no recording, text only." : "报告生成后 24h 内自动删除。关闭 → 不录音，仅文字。" },
            { k: "transcript", t: lang === "en" ? "Keep text transcripts long-term" : "长期保留文字转写", d: lang === "en" ? "Used to show you your past answers & trend your growth." : "用来让你看回去的回答、统计成长趋势。" },
            { k: "resume", t: lang === "en" ? "Keep uploaded resumes" : "保留上传的简历", d: lang === "en" ? "Used only to suggest stories during practice." : "只在练习时用来推荐故事。" },
            { k: "anon", t: lang === "en" ? "Contribute anonymized samples to improve rubrics" : "匿名贡献样本用于改进评分", d: lang === "en" ? "Off by default. Names, companies, numbers are stripped." : "默认关闭。姓名、公司、数字会被去除。" },
          ].map((r, i, arr) => (
            <div key={r.k} style={{ padding: "16px 20px", display: "flex", gap: 16, alignItems: "flex-start", borderBottom: i < arr.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 14, color: T.ink, marginBottom: 3 }}>{r.t}</div>
                <div style={{ fontSize: 12, color: T.ink3, lineHeight: 1.5 }}>{r.d}</div>
              </div>
              <Toggle k={r.k} />
            </div>
          ))}
        </div>
      </section>

      {/* Data overview */}
      <section style={{ marginBottom: 40 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "WHAT WE HAVE ON YOU" : "我们这边存了你什么"}</div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 10, marginBottom: 14 }}>
          {[
            { k: "4", l: lang === "en" ? "target jobs" : "目标岗位" },
            { k: "18", l: lang === "en" ? "practice sessions" : "练习会话" },
            { k: "2", l: lang === "en" ? "resumes" : "份简历" },
            { k: "0 min", l: lang === "en" ? "audio retained" : "语音留存" },
          ].map((c) => (
            <div key={c.l} style={{ padding: "16px 18px", background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
              <div className="ei-serif" style={{ fontSize: 24, color: T.ink, letterSpacing: "-0.015em" }}>{c.k}</div>
              <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em", marginTop: 4 }}>{c.l}</div>
            </div>
          ))}
        </div>
      </section>

      {/* Export */}
      <section style={{ marginBottom: 40 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "EXPORT" : "导出"}</div>
        <div style={{ padding: "18px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, display: "flex", gap: 16, alignItems: "center" }}>
          <Icon name="download" size={20} color={T.ink2} />
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 14, color: T.ink, marginBottom: 3 }}>
              {lang === "en" ? "Download everything — JSON + PDF reports" : "下载全部 —— JSON + PDF 报告"}
            </div>
            <div style={{ fontSize: 12, color: T.ink3, lineHeight: 1.5 }}>
              {lang === "en" ? "Practice sessions, reports, mistakes, resume versions, experience cards. Link is emailed when ready (<5min)." : "练习会话、报告、错题、简历版本、经历卡片。准备好发到你邮箱（<5 分钟）。"}
            </div>
          </div>
          <Btn T={T} variant="secondary" size="sm">{lang === "en" ? "Request export" : "申请导出"}</Btn>
        </div>
      </section>

      {/* Delete — danger zone */}
      <section>
        <div className="ei-label" style={{ color: T.danger, marginBottom: 14 }}>{lang === "en" ? "DANGER ZONE" : "高危操作"}</div>
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          {[
            { t: lang === "en" ? "Delete a single session" : "删除某一次会话", d: lang === "en" ? "Pick a session — transcript, report, and any audio are removed. Mistakes derived from it stay unless you also remove those." : "挑一次会话，转写、报告、音频全部删除。派生的错题保留，除非你一并删除。", b: lang === "en" ? "Pick" : "选择" },
            { t: lang === "en" ? "Delete all practice data" : "删除所有练习数据", d: lang === "en" ? "Sessions, reports, mistakes, growth. Target jobs & profile stay." : "会话、报告、错题、成长全部删掉。岗位和画像保留。", b: lang === "en" ? "Delete…" : "删除…" },
            { t: lang === "en" ? "Delete my account" : "注销账号", d: lang === "en" ? "Permanent. All data is purged within 30 days per GDPR. Backups rotated within 90." : "永久。30 天内按 GDPR 清理所有数据。备份 90 天内轮换清除。", b: lang === "en" ? "Delete account…" : "注销账号…", danger: true },
          ].map((r) => (
            <div key={r.t} style={{ padding: "16px 20px", background: r.danger ? T.dangerSoft : T.bgCard, border: `1px solid ${r.danger ? T.danger : T.rule}`, borderRadius: 2, display: "flex", gap: 16, alignItems: "flex-start" }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 14, color: T.ink, marginBottom: 3, fontWeight: 500 }}>{r.t}</div>
                <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.55 }}>{r.d}</div>
              </div>
              <button style={{
                padding: "6px 14px", fontSize: 12.5, cursor: "pointer", fontFamily: "var(--ei-sans)",
                background: r.danger ? T.danger : "transparent",
                color: r.danger ? "#fff" : T.danger,
                border: `1px solid ${T.danger}`, borderRadius: 2, whiteSpace: "nowrap",
              }}>{r.b}</button>
            </div>
          ))}
        </div>
        <div style={{ marginTop: 16, fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", lineHeight: 1.6 }}>
          {lang === "en" ? "deletion requests typically complete < 5 min · legal-hold exceptions documented in privacy policy" : "删除请求通常 < 5 分钟完成 · 法律保留例外情况详见隐私政策"}
        </div>
      </section>
    </div>
  );
};

const SettingsProfile = ({ T, lang }) => {
  const identityRows = [
    { k: lang === "en" ? "Display name" : "显示姓名", v: "刘哲" },
    { k: lang === "en" ? "Login email" : "登录邮箱", v: "liuzhe@example.com" },
    { k: lang === "en" ? "Mobile" : "手机号", v: lang === "en" ? "Not connected" : "未绑定" },
    { k: lang === "en" ? "Interface language" : "界面语言", v: lang === "en" ? "Chinese · English" : "中文 · English" },
    { k: lang === "en" ? "Time zone" : "时区", v: "Asia/Shanghai" },
  ];
  const securityRows = [
    { k: lang === "en" ? "Password" : "密码", v: lang === "en" ? "Last updated 18 days ago" : "18 天前更新" },
    { k: lang === "en" ? "Login method" : "登录方式", v: lang === "en" ? "Email code + password" : "邮箱验证码 + 密码" },
    { k: lang === "en" ? "Two-step verification" : "两步验证", v: lang === "en" ? "Off" : "未开启" },
  ];

  const Row = ({ r, last }) => (
    <div style={{ display: "flex", gap: 16, padding: "14px 0", borderBottom: last ? "none" : `1px dotted ${T.rule}`, alignItems: "baseline" }}>
      <div className="ei-label" style={{ color: T.ink3, minWidth: 160, fontSize: 11 }}>{r.k}</div>
      <div style={{ fontSize: 14, color: T.ink, flex: 1 }}>{r.v}</div>
      <button style={{ background: "transparent", border: "none", color: T.accent, fontSize: 12, cursor: "pointer" }}>{lang === "en" ? "Edit" : "编辑"}</button>
    </div>
  );

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 26 }}>
      <section>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "BASIC ACCOUNT INFO" : "账号基础信息"}</div>
        <div style={{ padding: "18px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, display: "flex", gap: 18, alignItems: "center", marginBottom: 14 }}>
          <div style={{ width: 48, height: 48, borderRadius: 24, background: T.ink2, color: T.bg, display: "flex", alignItems: "center", justifyContent: "center", fontFamily: "var(--ei-mono)", fontWeight: 600 }}>LZ</div>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 15, color: T.ink, fontWeight: 600 }}>{lang === "en" ? "Liu Zhe" : "刘哲"}</div>
            <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 4 }}>{lang === "en" ? "This page only stores account identity and product preferences." : "这里仅保存账号身份和产品基础偏好，不保存岗位、年限、目标方向等一对多信息。"}</div>
          </div>
          <Btn T={T} variant="secondary" size="sm" icon="upload">{lang === "en" ? "Change avatar" : "更换头像"}</Btn>
        </div>
        <div>
          {identityRows.map((r, i) => <Row key={r.k} r={r} last={i === identityRows.length - 1} />)}
        </div>
      </section>

      <section>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "SIGN-IN & SECURITY" : "登录与安全"}</div>
        <div>
          {securityRows.map((r, i) => <Row key={r.k} r={r} last={i === securityRows.length - 1} />)}
        </div>
      </section>
    </div>
  );
};

const SettingsNotif = ({ T, lang }) => (
  <div style={{ fontSize: 14, color: T.ink3, padding: 24, background: T.bgSoft, borderRadius: 2 }}>
    {lang === "en" ? "Report ready notifications, practice reminders, and weekly growth digests — coming in P1." : "报告就绪通知、练习提醒、每周成长摘要——P1 上线。"}
  </div>
);

const SettingsBilling = ({ T, lang }) => (
  <div style={{ padding: "24px 28px", background: T.accentSoft, borderRadius: 2 }}>
    <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{lang === "en" ? "CURRENT PLAN" : "当前套餐"}</div>
    <div className="ei-serif" style={{ fontSize: 24, color: T.ink, marginBottom: 6, letterSpacing: "-0.015em" }}>{lang === "en" ? "Free · 2 JD parses / month" : "免费版 · 每月 2 次 JD 解析"}</div>
    <div style={{ fontSize: 13, color: T.ink3, marginBottom: 16 }}>{lang === "en" ? "Upgrade for unlimited practice, real-interview debrief, and the mistake book." : "升级解锁无限练习、真实面试复盘、完整错题本。"}</div>
    <Btn T={T} variant="accent" iconRight="arrow_right">{lang === "en" ? "See plans" : "查看套餐"}</Btn>
  </div>
);

window.ParseScreen = ParseScreen;
window.OnboardingScreen = OnboardingScreen;
window.ReportGeneratingScreen = ReportGeneratingScreen;
window.SettingsScreen = SettingsScreen;
