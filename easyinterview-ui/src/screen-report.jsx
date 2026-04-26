// Screen 4: Evidence-based Review / Report — with 3 layout variants
const ReportScreen = ({ T, lang, nav, reportLayout, setReportLayout }) => {
  const D = window.EI_DATA;
  const r = D.report;

  const layouts = lang === "en"
    ? [{ k: "editorial", label: "Editorial", icon: "book", desc: "Long-form, evidence-led" }, { k: "timeline", label: "Timeline", icon: "clock", desc: "Question-by-question" }, { k: "dashboard", label: "Dashboard", icon: "chart", desc: "Metrics-first" }]
    : [{ k: "editorial", label: "刊物式", icon: "book", desc: "长文 · 证据驱动" }, { k: "timeline", label: "时间线", icon: "clock", desc: "按题目展开" }, { k: "dashboard", label: "仪表盘", icon: "chart", desc: "指标优先" }];

  const switcher = setReportLayout ? (
    <div style={{ display: "flex", alignItems: "center", gap: 10, padding: "12px 48px", borderBottom: `1px solid ${T.rule}`, background: T.bgSoft, position: "sticky", top: 58, zIndex: 5 }}>
      <span className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "VIEW" : "视图"}</span>
      {layouts.map((L) => {
        const on = (reportLayout || "editorial") === L.k;
        return (
          <button key={L.k} onClick={() => setReportLayout(L.k)} title={L.desc} style={{
            background: on ? T.bgCard : "transparent", border: `1px solid ${on ? T.rule : "transparent"}`,
            color: on ? T.ink : T.ink3, padding: "5px 11px", borderRadius: 2,
            fontSize: 12.5, cursor: "pointer", display: "flex", gap: 6, alignItems: "center", fontFamily: "var(--ei-sans)",
          }}>
            <Icon name={L.icon} size={12} /> {L.label}
          </button>
        );
      })}
      <span style={{ flex: 1 }} />
      <span style={{ fontSize: 11.5, color: T.ink3 }}>{layouts.find((x) => x.k === (reportLayout || "editorial"))?.desc}</span>
    </div>
  ) : null;

  let body;
  if (reportLayout === "timeline") body = <ReportTimeline T={T} lang={lang} nav={nav} r={r} />;
  else if (reportLayout === "dashboard") body = <ReportDashboard T={T} lang={lang} nav={nav} r={r} />;
  else body = <ReportEditorial T={T} lang={lang} nav={nav} r={r} />;

  return (
    <>
      {switcher}
      {body}
    </>
  );
};

// Variant 1: Editorial — document-style, priority at top
const ReportEditorial = ({ T, lang, nav, r }) => {
  const D = window.EI_DATA;
  const L = lang === "en" ? {
    back: "Back to workspace", session: "SESSION REPORT", round: "Round",
    priority: "WHAT TO FIX NEXT", highlights: "What went well", issues: "What to fix",
    perQ: "Per-question review", dims: "Dimension snapshot", next: "Plan next practice",
    evidence: "Evidence", suggestion: "Suggestion", good: "Good", missing: "Missing", frame: "Frame",
    openDrill: "Open as drill", addToMistakes: "Send to mistake book", addedToMistakes: "Added",
  } : {
    back: "返回工作台", session: "会话报告", round: "轮次",
    priority: "下一次最该改这个", highlights: "做得好的点", issues: "需要改的点",
    perQ: "逐题复盘", dims: "能力维度快照", next: "规划下一轮练习",
    evidence: "证据", suggestion: "建议", good: "做得好", missing: "缺的点", frame: "推荐框架",
    openDrill: "作为单题深钻", addToMistakes: "写入错题本", addedToMistakes: "已加入",
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 900, margin: "0 auto", padding: "32px 48px 120px" }}>
      <button onClick={() => nav("workspace", { jobId: "tj-1" })} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, display: "flex", alignItems: "center", gap: 6, padding: 0, marginBottom: 28, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {L.back}
      </button>

      {/* Masthead */}
      <div style={{ borderBottom: `2px solid ${T.ink}`, paddingBottom: 18, marginBottom: 28 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
          <div className="ei-label" style={{ color: T.ink }}>{L.session} · Nº 024</div>
          <div className="ei-mono" style={{ fontSize: 11, color: T.ink3 }}>{r.completedAt}</div>
        </div>
        <h1 className="ei-serif" style={{ fontSize: 40, color: T.ink, margin: 0, letterSpacing: "-0.025em", lineHeight: 1.1 }}>
          {lang === "en" ? "Close, but the performance story still has no numbers." : "离过关还差一口气——性能故事缺了数字。"}
        </h1>
        <div style={{ display: "flex", gap: 24, marginTop: 18, alignItems: "center" }}>
          <ReadinessDial level={r.readiness} label={r.readinessLabel} T={T} size={52} />
          <div style={{ height: 40, width: 1, background: T.rule }} />
          <KVInline T={T} k={lang === "en" ? "Round" : "轮次"} v={r.round} />
          <KVInline T={T} k={lang === "en" ? "Duration" : "时长"} v={r.duration} />
          <KVInline T={T} k={lang === "en" ? "Target" : "目标"} v="星环 · 资深前端" />
        </div>
      </div>

      {/* Priority callout */}
      <div style={{ background: T.accentSoft, padding: 24, marginBottom: 40, borderLeft: `3px solid ${T.accent}` }}>
        <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>▸ {L.priority}</div>
        <div className="ei-serif" style={{ fontSize: 22, color: T.ink, lineHeight: 1.4 }}>{r.topPriority}</div>
      </div>

      {/* Issues first (priority) */}
      <section style={{ marginBottom: 44 }}>
        <h2 className="ei-serif" style={{ fontSize: 24, color: T.ink, margin: "0 0 16px", letterSpacing: "-0.015em" }}>{L.issues}</h2>
        {r.issues.map((iss, i) => <IssueRow key={i} iss={iss} T={T} L={L} lang={lang} nav={nav} />)}
      </section>

      {/* Highlights */}
      <section style={{ marginBottom: 44 }}>
        <h2 className="ei-serif" style={{ fontSize: 24, color: T.ink, margin: "0 0 16px", letterSpacing: "-0.015em" }}>{L.highlights}</h2>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(240px, 1fr))", gap: 14 }}>
          {r.highlights.map((h, i) => (
            <div key={i} style={{ padding: 18, background: T.okSoft, borderRadius: 2 }}>
              <div style={{ color: T.ok, fontSize: 14, fontWeight: 500, marginBottom: 6 }}>● {h.title}</div>
              <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.55 }}>{h.body}</div>
              <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 8 }}>{L.evidence} · {h.evidence}</div>
            </div>
          ))}
        </div>
      </section>

      {/* Dimensions */}
      <section style={{ marginBottom: 44 }}>
        <h2 className="ei-serif" style={{ fontSize: 24, color: T.ink, margin: "0 0 16px", letterSpacing: "-0.015em" }}>{L.dims}</h2>
        <div style={{ border: `1px solid ${T.rule}`, borderRadius: 2, overflow: "hidden" }}>
          {r.dimensions.map((d, i) => <DimRow key={i} d={d} T={T} last={i === r.dimensions.length - 1} lang={lang} />)}
        </div>
        <div style={{ fontSize: 12, color: T.ink3, marginTop: 10, display: "flex", gap: 6, alignItems: "flex-start" }}>
          <Icon name="info" size={12} style={{ marginTop: 2 }} />
          {lang === "en" ? "Scores guide your prep — they're not a pass/fail. Confidence shows how certain we are given the sample." :
            "分数是导航，不是结论。置信度说明了我们对这个判断的把握程度。"}
        </div>
      </section>

      {/* Per question */}
      <section style={{ marginBottom: 44 }}>
        <h2 className="ei-serif" style={{ fontSize: 24, color: T.ink, margin: "0 0 16px", letterSpacing: "-0.015em" }}>{L.perQ}</h2>
        {r.perQuestion.map((q, i) => <PerQBlock key={i} q={q} T={T} L={L} />)}
      </section>

      {/* Next */}
      <section style={{ padding: 24, background: T.bgSoft, borderRadius: 2 }}>
        <h2 className="ei-serif" style={{ fontSize: 22, color: T.ink, margin: "0 0 14px" }}>{L.next}</h2>
        {r.nextPractice.map((n, i) => (
          <div key={i} style={{ padding: "12px 0", display: "flex", gap: 12, alignItems: "center", borderBottom: i < r.nextPractice.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
            <div style={{ width: 22, height: 22, borderRadius: 11, background: T.accent, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontFamily: "var(--ei-mono)", fontWeight: 500 }}>{i + 1}</div>
            <div style={{ flex: 1, fontSize: 14, color: T.ink }}>{n}</div>
            <Btn variant="secondary" size="sm" T={T} icon="play">{lang === "en" ? "Start" : "开始"}</Btn>
          </div>
        ))}
      </section>
    </div>
  );
};

// Variant 2: Timeline — vertical session timeline with markers
const ReportTimeline = ({ T, lang, nav, r }) => (
  <div className="ei-fadein" style={{ maxWidth: 1000, margin: "0 auto", padding: "32px 48px 120px" }}>
    <button onClick={() => nav("workspace", { jobId: "tj-1" })} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, display: "flex", alignItems: "center", gap: 6, padding: 0, marginBottom: 24 }}>
      <Icon name="arrow_left" size={14} /> {lang === "en" ? "Back" : "返回"}
    </button>
    <div style={{ marginBottom: 28 }}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "TIMELINE VIEW" : "时间轴视图"}</div>
      <h1 className="ei-serif" style={{ fontSize: 36, color: T.ink, margin: 0, letterSpacing: "-0.02em" }}>
        {lang === "en" ? "Walk through every moment that mattered." : "走一遍每个关键时刻。"}
      </h1>
      <div style={{ display: "flex", gap: 20, marginTop: 16, alignItems: "center" }}>
        <ReadinessDial level={r.readiness} label={r.readinessLabel} T={T} size={48} />
        <KVInline T={T} k="Duration" v={r.duration} />
        <KVInline T={T} k="Round" v={r.round} />
      </div>
    </div>

    <div style={{ position: "relative", paddingLeft: 40 }}>
      <div style={{ position: "absolute", left: 10, top: 0, bottom: 0, width: 1, background: T.rule }} />
      {[
        { t: "00:00", title: lang === "en" ? "Session start" : "会话开始", tone: "muted", body: lang === "en" ? "General interviewer · 25 min budget" : "综合面试官 · 25 分钟预算" },
        { t: "00:18", title: "Q1 · " + (lang === "en" ? "Self intro" : "自我介绍"), tone: "ok", body: lang === "en" ? "Clean opening, maps to JD early." : "开场结构清晰，早早映射到 JD。" },
        { t: "02:14", title: "Q2 · " + (lang === "en" ? "Performance story" : "性能优化"), tone: "danger", body: lang === "en" ? "Missing quantified result." : "缺可量化结果。" },
        { t: "08:22", title: lang === "en" ? "Follow-up triggered" : "触发追问", tone: "amber", body: lang === "en" ? "AI asked for numbers — answer stayed qualitative." : "AI 追问具体数字——回答仍然停留在定性层面。" },
        { t: "14:02", title: "Q4 · Design System", tone: "amber", body: lang === "en" ? "Restated plan instead of comparing alternatives." : "复述了方案，没给替代方案权衡。" },
        { t: "20:41", title: "Q5 · " + (lang === "en" ? "Reverse questions" : "反问"), tone: "warn", body: lang === "en" ? "Generic; no company-specific question." : "偏通用，没有贴合公司近况的问题。" },
        { t: "22:14", title: lang === "en" ? "Session complete" : "会话结束", tone: "accent", body: r.topPriority },
      ].map((ev, i) => (
        <div key={i} style={{ position: "relative", marginBottom: 22, paddingLeft: 8 }}>
          <div style={{ position: "absolute", left: -36, top: 4, width: 14, height: 14, borderRadius: 7,
            background: T[ev.tone] || T.ink3, border: `3px solid ${T.bg}` }} />
          <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{ev.t}</div>
          <div style={{ fontSize: 15, color: T.ink, fontWeight: 500, marginTop: 2 }}>{ev.title}</div>
          <div style={{ fontSize: 13, color: T.ink2, marginTop: 4, lineHeight: 1.55, maxWidth: 580 }}>{ev.body}</div>
        </div>
      ))}
    </div>

    <div style={{ marginTop: 40, padding: 24, background: T.accentSoft, borderLeft: `3px solid ${T.accent}` }}>
      <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{lang === "en" ? "TOP PRIORITY" : "下一次最该改这个"}</div>
      <div className="ei-serif" style={{ fontSize: 20, color: T.ink, lineHeight: 1.4 }}>{r.topPriority}</div>
    </div>
  </div>
);

// Variant 3: Dashboard — metric-heavy, at-a-glance
const ReportDashboard = ({ T, lang, nav, r }) => (
  <div className="ei-fadein" style={{ maxWidth: 1200, margin: "0 auto", padding: "32px 48px 96px" }}>
    <button onClick={() => nav("workspace", { jobId: "tj-1" })} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 20, display: "flex", alignItems: "center", gap: 6 }}>
      <Icon name="arrow_left" size={14} /> {lang === "en" ? "Back" : "返回"}
    </button>

    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr 1fr", gap: 14, marginBottom: 24 }}>
      <StatCard T={T} label={lang === "en" ? "READINESS" : "准备度"} value={r.readinessLabel} big />
      <StatCard T={T} label={lang === "en" ? "DURATION" : "时长"} value={r.duration} mono />
      <StatCard T={T} label={lang === "en" ? "QUESTIONS" : "题目数"} value="5 / 5" mono />
      <StatCard T={T} label={lang === "en" ? "FOLLOW-UPS" : "追问次数"} value="7" mono />
    </div>

    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20, marginBottom: 24 }}>
      <Card T={T}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "DIMENSIONS" : "能力维度"}</div>
        {r.dimensions.map((d, i) => <DimRow key={i} d={d} T={T} last={i === r.dimensions.length - 1} lang={lang} />)}
      </Card>
      <Card T={T}>
        <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>▸ {lang === "en" ? "TOP PRIORITY" : "下一次最该改这个"}</div>
        <div className="ei-serif" style={{ fontSize: 18, color: T.ink, lineHeight: 1.45, marginBottom: 16 }}>{r.topPriority}</div>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "NEXT PRACTICE" : "下一步"}</div>
        {r.nextPractice.map((n, i) => (
          <div key={i} style={{ display: "flex", gap: 10, padding: "8px 0", fontSize: 13, color: T.ink2, borderBottom: i < r.nextPractice.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
            <span style={{ color: T.accent, fontFamily: "var(--ei-mono)" }}>{String(i + 1).padStart(2, "0")}</span>
            <span>{n}</span>
          </div>
        ))}
      </Card>
    </div>

    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20 }}>
      <Card T={T}>
        <div className="ei-label" style={{ color: T.danger, marginBottom: 12 }}>▲ {lang === "en" ? "ISSUES" : "问题"}</div>
        {r.issues.map((iss, i) => (
          <div key={i} style={{ padding: "10px 0", borderBottom: i < r.issues.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
            <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>{iss.title}</div>
            <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 2 }}>{iss.evidence}</div>
          </div>
        ))}
      </Card>
      <Card T={T}>
        <div className="ei-label" style={{ color: T.ok, marginBottom: 12 }}>● {lang === "en" ? "HIGHLIGHTS" : "亮点"}</div>
        {r.highlights.map((h, i) => (
          <div key={i} style={{ padding: "10px 0", borderBottom: i < r.highlights.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
            <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>{h.title}</div>
            <div style={{ fontSize: 12, color: T.ink3, marginTop: 2 }}>{h.body}</div>
          </div>
        ))}
      </Card>
    </div>
  </div>
);

const IssueRow = ({ iss, T, L, lang, nav }) => {
  const toneMap = { high: T.danger, medium: T.warn, low: T.ink3 };
  const [added, setAdded] = React.useState(false);
  return (
    <div style={{ padding: "18px 0", borderBottom: `1px dotted ${T.rule}`, display: "grid", gridTemplateColumns: "auto 1fr", gap: 16 }}>
      <div style={{ color: toneMap[iss.severity], fontSize: 18, lineHeight: 1, marginTop: 2 }}>▲</div>
      <div>
        <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 6 }}>
          <div style={{ fontSize: 16, color: T.ink, fontWeight: 500 }}>{iss.title}</div>
          <Tag tone={iss.severity === "high" ? "danger" : iss.severity === "medium" ? "warn" : "muted"} T={T}>
            {iss.severity === "high" ? (lang === "en" ? "High" : "高优先级") : iss.severity === "medium" ? (lang === "en" ? "Medium" : "中") : (lang === "en" ? "Low" : "低")}
          </Tag>
          <span style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{iss.evidence}</span>
        </div>
        <div style={{ fontSize: 14, color: T.ink2, lineHeight: 1.6, marginBottom: 8 }}>{iss.body}</div>
        <div style={{ fontSize: 13, color: T.ink2, padding: "10px 12px", background: T.bgSoft, borderRadius: 2, lineHeight: 1.55 }}>
          <b style={{ color: T.ink, fontSize: 11.5, fontFamily: "var(--ei-mono)", letterSpacing: "0.06em" }}>▸ {L.suggestion}</b><br/>{iss.suggestion}
        </div>
        <div style={{ display: "flex", gap: 8, marginTop: 10 }}>
          <Btn variant="secondary" size="sm" T={T} icon="play" onClick={() => nav("practice", { jobId: "tj-1", mode: "drill" })}>{L.openDrill}</Btn>
          <Btn variant="ghost" size="sm" T={T} icon={added ? "check" : "plus"} onClick={() => setAdded(true)}>
            {added ? L.addedToMistakes : L.addToMistakes}
          </Btn>
        </div>
      </div>
    </div>
  );
};

const PerQBlock = ({ q, T, L }) => {
  const toneMap = { 强项: T.ok, 达标: T.cool, 待加强: T.warn };
  return (
    <div style={{ padding: "18px 0", borderBottom: `1px dotted ${T.rule}` }}>
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 10 }}>
        <div style={{ fontSize: 15, color: T.ink, fontWeight: 500 }}>{q.qId.toUpperCase()} · {q.topic}</div>
        <Tag tone={q.state === "强项" ? "ok" : q.state === "待加强" ? "warn" : "cool"} T={T}>{q.state}</Tag>
      </div>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 16, fontSize: 13, lineHeight: 1.55 }}>
        <div><div className="ei-label" style={{ color: T.ok, marginBottom: 4 }}>{L.good}</div><div style={{ color: T.ink2 }}>{q.good}</div></div>
        <div><div className="ei-label" style={{ color: T.danger, marginBottom: 4 }}>{L.missing}</div><div style={{ color: T.ink2 }}>{q.missing}</div></div>
        <div><div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{L.frame}</div><div style={{ color: T.ink2, fontFamily: "var(--ei-mono)", fontSize: 12 }}>{q.frame}</div></div>
      </div>
    </div>
  );
};

const DimRow = ({ d, T, last, lang }) => {
  const toneMap = { 强项: T.ok, 达标: T.ink2, 待加强: T.warn };
  return (
    <div style={{ padding: "14px 16px", display: "flex", alignItems: "center", gap: 16, borderBottom: last ? "none" : `1px dotted ${T.rule}` }}>
      <div style={{ width: 110, fontSize: 13, color: T.ink, fontWeight: 500 }}>{d.name}</div>
      <div style={{ flex: 1, height: 4, background: T.bgSoft, borderRadius: 2, position: "relative", overflow: "hidden" }}>
        <div style={{ position: "absolute", left: 0, top: 0, bottom: 0, width: `${d.score}%`, background: toneMap[d.state] || T.ink2 }} />
      </div>
      <div style={{ width: 60, fontSize: 12, color: toneMap[d.state] || T.ink2, fontWeight: 500, textAlign: "right" }}>{d.state}</div>
      <div style={{ width: 50, fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", textAlign: "right" }}>{lang === "en" ? "conf:" : "置信:"} {d.confidence}</div>
    </div>
  );
};

const StatCard = ({ T, label, value, mono, big }) => (
  <div style={{ padding: "18px 20px", border: `1px solid ${T.rule}`, borderRadius: 2, background: T.bgCard }}>
    <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{label}</div>
    <div className={mono ? "ei-mono" : "ei-serif"} style={{ fontSize: big ? 22 : 26, color: T.ink, letterSpacing: big ? "-0.01em" : 0 }}>{value}</div>
  </div>
);

const KVInline = ({ T, k, v }) => (
  <div>
    <div className="ei-label" style={{ color: T.ink3, marginBottom: 2 }}>{k}</div>
    <div style={{ fontSize: 14, color: T.ink, fontWeight: 500 }}>{v}</div>
  </div>
);

window.ReportScreen = ReportScreen;
