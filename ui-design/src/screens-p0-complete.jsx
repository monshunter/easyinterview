// P0 support screens: JD parse flow, report generation, settings/privacy, and upload/deletion states.

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
          <Btn T={T} variant="secondary" icon="edit" onClick={() => { setStep(0); setStage("loading"); window.scrollTo({ top: 0, behavior: "smooth" }); }}>{lang === "en" ? "Re-parse" : "重新解析"}</Btn>
          <Btn T={T} variant="accent" iconRight="arrow_right" onClick={() => nav("workspace", window.eiCreateInterviewContext ? window.eiCreateInterviewContext() : {})}>
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
// #3 REPORT GENERATING — async intermediate screen
// ═══════════════════════════════════════════════════════════════════
const ReportGeneratingScreen = ({ T, lang, nav, params = {} }) => {
  const [phase, setPhase] = React.useState(0);
  const phases = lang === "en" ? [
    { t: "Transcribing & aligning turns", s: 900, hint: "8 questions · 23 turns" },
    { t: "Extracting evidence per question", s: 1200, hint: "looking for S/A/R + quantification" },
    { t: "Scoring against rubric", s: 900, hint: "rubric · behavior-v2.1 · confidence tagged" },
    { t: "Clustering question review signals", s: 700, hint: "3 items marked for current-round replay" },
    { t: "Writing recommendations", s: 900, hint: "frameworks + mapped resume evidence" },
  ] : [
    { t: "转写并对齐对话", s: 900, hint: "8 题 · 23 轮对话" },
    { t: "逐题抽取证据", s: 1200, hint: "寻找 S/A/R + 量化结果" },
    { t: "按 rubric 评分", s: 900, hint: "rubric · behavior-v2.1 · 带置信度" },
    { t: "聚类题目回顾信号", s: 700, hint: "标记 3 条本轮复练线索" },
    { t: "生成建议", s: 900, hint: "推荐回答框架 + 映射简历证据" },
  ];

  React.useEffect(() => {
    let cancel = false;
    let acc = 0;
    phases.forEach((p, i) => {
      acc += p.s;
      setTimeout(() => { if (!cancel) setPhase(i + 1); }, acc);
    });
    setTimeout(() => { if (!cancel) nav("report", { ...params, sessionId: params.sessionId || "session-24" }); }, acc + 600);
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
          {lang === "en" ? "Typical: 8-15s. You can close this tab; the report opens from this session history when it's done." : "通常 8-15 秒。可以关掉这个页面；报告生成后可从本场会话历史打开。"}
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
const SettingsScreen = ({ T, lang, nav, fontPreset, setFontPreset }) => {
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
          {tab === "profile" && <SettingsProfile T={T} lang={lang} fontPreset={fontPreset} setFontPreset={setFontPreset} />}
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
            { k: "transcript", t: lang === "en" ? "Keep text transcripts long-term" : "长期保留文字转写", d: lang === "en" ? "Used to review past answers and compare readiness over time." : "用来回看历史回答，并比较准备度变化。" },
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
              {lang === "en" ? "Practice sessions, reports, question reviews, resume versions, and profile evidence. Link is emailed when ready (<5min)." : "练习会话、报告、题目回顾、简历版本和画像证据。准备好发到你邮箱（<5 分钟）。"}
            </div>
          </div>
          <Btn T={T} variant="secondary" size="sm" onClick={() => window.eiToast && window.eiToast(lang === "en" ? "Export requested · link emailed to liuzhe@example.com when ready" : "已申请导出 · 准备好后会发到 liuzhe@example.com", { tone: "ok", duration: 3000 })}>{lang === "en" ? "Request export" : "申请导出"}</Btn>
        </div>
      </section>

      {/* Delete — danger zone */}
      <section>
        <div className="ei-label" style={{ color: T.danger, marginBottom: 14 }}>{lang === "en" ? "DANGER ZONE" : "高危操作"}</div>
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          {[
            { t: lang === "en" ? "Delete a single session" : "删除某一次会话", d: lang === "en" ? "Pick a session — transcript, report, question reviews, and any audio are removed together." : "挑一次会话，转写、报告、题目回顾和音频会一起删除。", b: lang === "en" ? "Pick" : "选择" },
            { t: lang === "en" ? "Delete all practice data" : "删除所有练习数据", d: lang === "en" ? "Sessions, reports, question reviews, and readiness signals are removed. Target jobs & profile stay." : "会话、报告、题目回顾和准备度信号全部删掉。岗位和画像保留。", b: lang === "en" ? "Delete…" : "删除…" },
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

const SettingsProfile = ({ T, lang, fontPreset, setFontPreset }) => {
  const identityRows = [
    { k: lang === "en" ? "Display name" : "显示姓名", v: "刘哲" },
    { k: lang === "en" ? "Login email" : "登录邮箱", v: "liuzhe@example.com" },
    { k: lang === "en" ? "Mobile" : "手机号", v: lang === "en" ? "Not connected" : "未绑定" },
    { k: lang === "en" ? "Interface language" : "界面语言", v: lang === "en" ? "Chinese · English" : "中文 · English" },
    { k: lang === "en" ? "Time zone" : "时区", v: "Asia/Shanghai" },
  ];
  const securityRows = [
    { k: lang === "en" ? "Password" : "密码", v: lang === "en" ? "Last updated 18 days ago" : "18 天前更新" },
    { k: lang === "en" ? "Login method" : "登录方式", v: lang === "en" ? "Email link + password" : "邮箱链接 + 密码" },
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
          <Btn T={T} variant="secondary" size="sm" icon="upload" onClick={() => window.eiToast && window.eiToast(lang === "en" ? "Pick an image (≤2 MB) · upload mocked in prototype" : "选择图片（≤2 MB）· 原型仅模拟上传", { tone: "neutral" })}>{lang === "en" ? "Change avatar" : "更换头像"}</Btn>
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

      <section>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "INTERFACE PREFERENCES" : "界面偏好"}</div>
        <div style={{ fontSize: 13, color: T.ink3, marginBottom: 14, lineHeight: 1.55 }}>
          {lang === "en"
            ? "Pick a typography pack — switches both the serif (display) and the sans (body) together so the rhythm stays coherent. Mono never changes."
            : "选一套字体组合 —— serif（标题）和 sans（正文）会成对切换，保证排版节奏不乱。等宽字体不会变。"}
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 12 }}>
          {(window.EI_FONT_PRESETS || []).map((p) => {
            const selected = (fontPreset || "editorial") === p.key;
            return (
              <button
                key={p.key}
                onClick={() => setFontPreset && setFontPreset(p.key)}
                style={{
                  textAlign: "left", padding: "16px 18px",
                  background: selected ? T.accentSoft : T.bgCard,
                  border: `1.5px solid ${selected ? T.accent : T.rule}`,
                  borderRadius: 3, cursor: "pointer",
                  display: "flex", flexDirection: "column", gap: 10,
                  fontFamily: "var(--ei-sans)",
                }}
                onMouseEnter={(e) => { if (!selected) e.currentTarget.style.borderColor = T.ink3; }}
                onMouseLeave={(e) => { if (!selected) e.currentTarget.style.borderColor = T.rule; }}
              >
                <div style={{ display: "flex", alignItems: "baseline", justifyContent: "space-between", gap: 8 }}>
                  <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 600 }}>
                    {lang === "en" ? p.labelEn : p.labelZh}
                  </div>
                  {selected && <Icon name="check" size={13} style={{ color: T.accent }} />}
                </div>
                <div style={{ fontFamily: `"${p.serif}", Georgia, serif`, fontSize: 26, color: T.ink, letterSpacing: "-0.015em", lineHeight: 1.1 }}>
                  Aa 你好
                </div>
                <div style={{ fontFamily: `"${p.sans}", -apple-system, sans-serif`, fontSize: 12.5, color: T.ink2, lineHeight: 1.45 }}>
                  The quick brown fox · 这是正文样式示例
                </div>
                <div style={{ fontSize: 11, color: T.ink3, lineHeight: 1.5, marginTop: 2 }}>
                  {lang === "en" ? p.descEn : p.descZh}
                </div>
                <div style={{ fontSize: 10.5, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em", marginTop: 2 }}>
                  {p.serif} · {p.sans}
                </div>
              </button>
            );
          })}
        </div>
      </section>

      <section>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "PRODUCT INFO" : "产品信息"}</div>
        <div style={{ padding: "16px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
          {[
            { k: lang === "en" ? "Product" : "产品名称", v: "EasyInterview" },
            { k: lang === "en" ? "Version" : "版本", v: "v1.0" },
          ].map((r, i, arr) => (
            <div key={r.k} style={{ display: "flex", gap: 16, padding: "10px 0", borderBottom: i < arr.length - 1 ? `1px dotted ${T.rule}` : "none", alignItems: "baseline" }}>
              <div className="ei-label" style={{ color: T.ink3, minWidth: 160, fontSize: 11 }}>{r.k}</div>
              <div style={{ fontSize: 14, color: T.ink, flex: 1, fontFamily: r.v === "v1.0" ? "var(--ei-mono)" : "var(--ei-sans)" }}>{r.v}</div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
};

const SettingsNotif = ({ T, lang }) => (
  <div style={{ fontSize: 14, color: T.ink3, padding: 24, background: T.bgSoft, borderRadius: 2 }}>
    {lang === "en" ? "Report ready notifications, practice reminders, and weekly readiness summaries — coming in P1." : "报告就绪通知、练习提醒、每周准备度摘要——P1 上线。"}
  </div>
);

const SettingsBilling = ({ T, lang }) => (
  <div style={{ padding: "24px 28px", background: T.accentSoft, borderRadius: 2 }}>
    <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{lang === "en" ? "CURRENT PLAN" : "当前套餐"}</div>
    <div className="ei-serif" style={{ fontSize: 24, color: T.ink, marginBottom: 6, letterSpacing: "-0.015em" }}>{lang === "en" ? "Free · 2 JD parses / month" : "免费版 · 每月 2 次 JD 解析"}</div>
    <div style={{ fontSize: 13, color: T.ink3, marginBottom: 16 }}>{lang === "en" ? "Upgrade for unlimited practice, debrief, and deeper evidence review." : "升级解锁无限练习、复盘和更完整的证据复盘。"}</div>
    <Btn T={T} variant="accent" iconRight="arrow_right" onClick={() => window.eiToast && window.eiToast(lang === "en" ? "Plans page lands in P1 · pricing TBD" : "套餐页 P1 上线 · 定价待定", { tone: "neutral" })}>{lang === "en" ? "See plans" : "查看套餐"}</Btn>
  </div>
);

window.ParseScreen = ParseScreen;
window.ReportGeneratingScreen = ReportGeneratingScreen;
window.SettingsScreen = SettingsScreen;
