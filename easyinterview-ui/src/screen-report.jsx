// Screen 4: Evidence-based Review / Report — dashboard only
const ReportScreen = ({ T, lang, nav }) => {
  const D = window.EI_DATA;
  const r = D.report;
  const job = D.targetJobs[0];
  const context = lang === "en" ? {
    breadcrumb: "Mock interview / Session #24 / Report",
    title: `${job.title} · ${r.round} mock report`,
    subtitle: "This report belongs to one completed mock interview. It is opened from interview history or after a session ends.",
    session: "Mock interview #24",
    target: `${job.company} · ${job.title}`,
    round: r.round,
    nextRound: "Technical round 2",
    resume: "Liu Zhe · resume v3",
    time: "Apr 20 · 15:48",
    duration: r.duration,
    modality: "Text",
  } : {
    breadcrumb: "模拟面试 / 会话 #24 / 面试报告",
    title: `${job.title} · ${r.round}模拟报告`,
    subtitle: "这份报告隶属于一次已完成的模拟面试，只从会话历史或面试结束后进入。",
    session: "模拟面试 #24",
    target: `${job.company} · ${job.title}`,
    round: r.round,
    nextRound: "技术二面",
    resume: "刘哲 · 简历 v3",
    time: "4/20 · 15:48",
    duration: r.duration,
    modality: "文本",
  };

  return <ReportDashboard T={T} lang={lang} nav={nav} r={r} context={context} />;
};

// Dashboard — metric-heavy, at-a-glance
const ReportDashboard = ({ T, lang, nav, r, context }) => {
  const [detail, setDetail] = React.useState("questions");
  const [activeQuestion, setActiveQuestion] = React.useState(r.perQuestion[1]?.qId || r.perQuestion[0]?.qId);

  return (
    <div className="ei-fadein" style={{ maxWidth: 1200, margin: "0 auto", padding: "32px 48px 96px" }}>
      <button onClick={() => nav("workspace", { jobId: "tj-1" })} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 20, display: "flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {lang === "en" ? "Back to interview setup" : "返回面试前确认"}
      </button>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", gap: 24, marginBottom: 18, flexWrap: "wrap" }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{context.breadcrumb}</div>
          <h1 className="ei-serif" style={{ fontSize: 38, color: T.ink, margin: 0, lineHeight: 1.15, letterSpacing: "-0.02em" }}>
            {context.title}
          </h1>
          <div style={{ fontSize: 14, color: T.ink2, marginTop: 8, lineHeight: 1.65, maxWidth: 720 }}>
            {context.subtitle}
          </div>
        </div>
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap", justifyContent: "flex-end" }}>
          <Btn variant="accent" icon="replay" onClick={() => nav("workspace", { jobId: "tj-1" })} T={T}>
            {lang === "en" ? `Replay ${context.round}` : `复练当前轮：${context.round}`}
          </Btn>
          <Btn variant="secondary" icon="arrow_right" onClick={() => nav("workspace", { jobId: "tj-1" })} T={T}>
            {lang === "en" ? `Prepare ${context.nextRound}` : `进入下一轮：${context.nextRound}`}
          </Btn>
        </div>
      </div>

      <ReportContextStrip T={T} lang={lang} context={context} />

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr 1fr", gap: 14, marginBottom: 18 }}>
        <ReportStatButton T={T} active={detail === "readiness"} onClick={() => setDetail("readiness")}>
          <StatCard T={T} label={lang === "en" ? "READINESS" : "准备度"} value={r.readinessLabel} big />
        </ReportStatButton>
        <ReportStatButton T={T} active={detail === "dimensions"} onClick={() => setDetail("dimensions")}>
          <StatCard T={T} label={lang === "en" ? "DIMENSION DETAIL" : "维度详情"} value={lang === "en" ? `${r.dimensions.length} cards` : `${r.dimensions.length} 个维度`} mono />
        </ReportStatButton>
        <ReportStatButton T={T} active={detail === "questions"} onClick={() => setDetail("questions")}>
          <StatCard T={T} label={lang === "en" ? "QUESTION REVIEW" : "题目回顾"} value="5 / 5" mono />
        </ReportStatButton>
        <ReportStatButton T={T} active={detail === "next"} onClick={() => setDetail("next")}>
          <StatCard T={T} label={lang === "en" ? "NEXT ACTION" : "下一动作"} value={lang === "en" ? "Replay current round" : "复练当前轮"} big />
        </ReportStatButton>
      </div>

      <ReportDetailSurface
        T={T}
        lang={lang}
        nav={nav}
        r={r}
        detail={detail}
        setDetail={setDetail}
        context={context}
        activeQuestion={activeQuestion}
        setActiveQuestion={setActiveQuestion}
      />

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20, marginBottom: 24 }}>
        <Card T={T}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14 }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "DIMENSIONS" : "能力维度"}</div>
            <button onClick={() => setDetail("dimensions")} style={{ background: "transparent", border: "none", color: T.accent, fontSize: 12, cursor: "pointer" }}>
              {lang === "en" ? "Open detail" : "查看详情"}
            </button>
          </div>
          {r.dimensions.map((d, i) => <DimRow key={i} d={d} T={T} last={i === r.dimensions.length - 1} lang={lang} />)}
        </Card>
        <Card T={T}>
          <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>▸ {lang === "en" ? "TOP PRIORITY" : "下一次最该改这个"}</div>
          <div className="ei-serif" style={{ fontSize: 18, color: T.ink, lineHeight: 1.45, marginBottom: 16 }}>{r.topPriority}</div>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 10 }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "CURRENT ROUND REPLAY" : "本轮复练重点"}</div>
            <button onClick={() => setDetail("next")} style={{ background: "transparent", border: "none", color: T.accent, fontSize: 12, cursor: "pointer" }}>
              {lang === "en" ? "Open plan" : "查看复练计划"}
            </button>
          </div>
          {r.nextPractice.map((n, i) => (
            <div key={i} style={{ display: "flex", gap: 10, padding: "8px 0", fontSize: 13, color: T.ink2, borderBottom: i < r.nextPractice.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
              <span style={{ color: T.accent, fontFamily: "var(--ei-mono)" }}>{String(i + 1).padStart(2, "0")}</span>
              <span>{n}</span>
            </div>
          ))}
        </Card>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1.2fr .8fr", gap: 20 }}>
        <Card T={T}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
            <div className="ei-label" style={{ color: T.danger }}>▲ {lang === "en" ? "QUESTION REVIEW" : "题目回顾 / 回答分析 / 建议"}</div>
            <button onClick={() => setDetail("questions")} style={{ background: "transparent", border: "none", color: T.accent, fontSize: 12, cursor: "pointer" }}>
              {lang === "en" ? "Open page" : "打开题目回顾页"}
            </button>
          </div>
          {r.perQuestion.map((q, i) => (
            <button key={q.qId} onClick={() => { setActiveQuestion(q.qId); setDetail("questions"); }} style={{
              width: "100%", textAlign: "left", background: "transparent", border: "none", padding: "12px 0",
              borderBottom: i < r.perQuestion.length - 1 ? `1px dotted ${T.rule}` : "none", cursor: "pointer", fontFamily: "var(--ei-sans)",
            }}>
              <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
                <Tag tone={q.state === "强项" ? "ok" : q.state === "待加强" ? "warn" : "cool"} T={T}>{q.state}</Tag>
                <div style={{ flex: 1, fontSize: 14, color: T.ink, fontWeight: 500 }}>{q.qId.toUpperCase()} · {q.topic}</div>
                <Icon name="arrow_right" size={14} color={T.ink3} />
              </div>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14, marginTop: 10, fontSize: 12.5, lineHeight: 1.55 }}>
                <div style={{ color: T.ink3 }}><span style={{ color: T.ok }}>{lang === "en" ? "Worked" : "有效"} · </span>{q.good}</div>
                <div style={{ color: T.ink3 }}><span style={{ color: q.state === "待加强" ? T.danger : T.warn }}>{lang === "en" ? "Gap" : "缺口"} · </span>{q.missing}</div>
              </div>
            </button>
          ))}
        </Card>
        <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
          <Card T={T}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
              <div className="ei-label" style={{ color: T.danger }}>▲ {lang === "en" ? "RISKS" : "问题风险"}</div>
              <button onClick={() => setDetail("evidence")} style={{ background: "transparent", border: "none", color: T.accent, fontSize: 12, cursor: "pointer" }}>
                {lang === "en" ? "Evidence" : "证据详情"}
              </button>
            </div>
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
    </div>
  );
};

const ReportContextStrip = ({ T, lang, context }) => {
  const items = lang === "en" ? [
    ["Session", context.session],
    ["Target", context.target],
    ["Round", context.round],
    ["Resume", context.resume],
    ["Completed", `${context.time} · ${context.duration}`],
    ["Mode", context.modality],
  ] : [
    ["所属会话", context.session],
    ["目标岗位", context.target],
    ["面试轮次", context.round],
    ["绑定简历", context.resume],
    ["完成时间", `${context.time} · ${context.duration}`],
    ["沟通形式", context.modality],
  ];
  return (
    <div style={{
      display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 0,
      border: `1px solid ${T.rule}`, borderRadius: 3, background: T.bgCard,
      marginBottom: 18, overflow: "hidden",
    }}>
      {items.map(([k, v], i) => (
        <div key={k} style={{
          padding: "13px 16px", borderRight: (i + 1) % 3 === 0 ? "none" : `1px dotted ${T.rule}`,
          borderBottom: i < 3 ? `1px dotted ${T.rule}` : "none",
        }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{k}</div>
          <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{v}</div>
        </div>
      ))}
    </div>
  );
};

const ReportStatButton = ({ T, active, onClick, children }) => (
  <button onClick={onClick} style={{
    padding: 0, border: active ? `1px solid ${T.accent}` : "1px solid transparent", borderRadius: 3,
    background: "transparent", cursor: "pointer", textAlign: "left", fontFamily: "var(--ei-sans)",
    boxShadow: active ? `0 0 0 2px ${T.accentSoft}` : "none",
  }}>
    {children}
  </button>
);

const ReportDetailSurface = ({ T, lang, nav, r, detail, setDetail, context, activeQuestion, setActiveQuestion }) => {
  const q = r.perQuestion.find((item) => item.qId === activeQuestion) || r.perQuestion[0];
  const tabs = [
    { k: "readiness", labelZh: "准备度详情", labelEn: "Readiness", icon: "target" },
    { k: "dimensions", labelZh: "维度详情", labelEn: "Dimensions", icon: "chart" },
    { k: "questions", labelZh: "题目回顾页", labelEn: "Questions", icon: "chat" },
    { k: "evidence", labelZh: "证据详情", labelEn: "Evidence", icon: "bookmark" },
    { k: "next", labelZh: "复练计划", labelEn: "Next", icon: "play" },
  ];

  return (
    <Card T={T} pad={0} style={{ marginBottom: 24 }}>
      <div style={{ display: "flex", gap: 0, borderBottom: `1px solid ${T.rule}`, overflowX: "auto" }}>
        {tabs.map((tab) => (
          <button key={tab.k} onClick={() => setDetail(tab.k)} style={{
            padding: "14px 18px", background: detail === tab.k ? T.bgSoft : "transparent", border: "none",
            borderBottom: `2px solid ${detail === tab.k ? T.accent : "transparent"}`, color: detail === tab.k ? T.ink : T.ink3,
            display: "flex", alignItems: "center", gap: 7, cursor: "pointer", fontFamily: "var(--ei-sans)", whiteSpace: "nowrap", marginBottom: -1,
          }}>
            <Icon name={tab.icon} size={13} /> {lang === "en" ? tab.labelEn : tab.labelZh}
          </button>
        ))}
      </div>

      {detail === "readiness" && (
        <div style={{ padding: 24, display: "grid", gridTemplateColumns: "220px 1fr", gap: 28, alignItems: "center" }}>
          <div style={{ display: "flex", justifyContent: "center" }}>
            <ReadinessDial level={r.readiness} label={r.readinessLabel} T={T} size={112} />
          </div>
          <div>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{lang === "en" ? "READINESS DETAIL" : "准备度二级详情"}</div>
            <div className="ei-serif" style={{ fontSize: 24, color: T.ink, lineHeight: 1.35, marginBottom: 14 }}>{r.topPriority}</div>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 12 }}>
              {[
                { k: lang === "en" ? "JD match" : "JD 对齐", v: "78%", body: lang === "en" ? "Most examples map to the role, but performance claims still need numbers." : "大部分经历能对上岗位，但性能故事还缺关键数字。" },
                { k: lang === "en" ? "Evidence density" : "证据密度", v: "Medium", body: lang === "en" ? "Two answers had reusable proof; one stayed qualitative." : "两题有可复用证据，一题停留在定性描述。" },
                { k: lang === "en" ? "Next threshold" : "下一档门槛", v: "+1", body: lang === "en" ? "Quantify the highest-risk story before replay." : "先量化最高风险故事，再进入复练。" },
              ].map((item) => (
                <div key={item.k} style={{ padding: 14, background: T.bgSoft, borderRadius: 2 }}>
                  <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{item.k}</div>
                  <div className="ei-serif" style={{ fontSize: 22, color: T.ink, marginBottom: 6 }}>{item.v}</div>
                  <div style={{ fontSize: 12.5, color: T.ink2, lineHeight: 1.55 }}>{item.body}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {detail === "dimensions" && (
        <div style={{ padding: 24 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "DIMENSION DETAIL CARDS" : "能力维度二级详情"}</div>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(2, 1fr)", gap: 14 }}>
            {r.dimensions.map((d, i) => (
              <div key={d.name} style={{ padding: 16, border: `1px solid ${T.rule}`, borderRadius: 2, background: T.bg }}>
                <div style={{ display: "flex", justifyContent: "space-between", gap: 10, marginBottom: 10 }}>
                  <div style={{ fontSize: 15, color: T.ink, fontWeight: 500 }}>{d.name}</div>
                  <Tag tone={d.state === "强项" ? "ok" : d.state === "待加强" ? "warn" : "cool"} T={T}>{d.state}</Tag>
                </div>
                <DimRow d={d} T={T} last lang={lang} />
                <div style={{ fontSize: 12.5, color: T.ink2, lineHeight: 1.6, marginTop: 10 }}>
                  {[
                    lang === "en" ? "Maps directly to the target JD and interviewer rubric." : "直接映射目标 JD 与本轮面试官评分口径。",
                    lang === "en" ? "Needs a clearer trade-off story before replay." : "复练前需要补一个更清楚的取舍故事。",
                    lang === "en" ? "Strong enough to keep; use it as opening proof." : "当前足够稳定，可作为开场证明点。",
                    lang === "en" ? "Evidence exists, but answer ordering can be tighter." : "证据存在，但回答顺序还可以更紧。",
                  ][i % 4]}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {detail === "questions" && (
        <div style={{ padding: 24, display: "grid", gridTemplateColumns: "320px 1fr", gap: 22 }}>
          <div>
            <div className="ei-label" style={{ color: T.danger, marginBottom: 12 }}>▲ {lang === "en" ? "QUESTION REVIEW PAGE" : "题目回顾页面"}</div>
            <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
              {r.perQuestion.map((item) => (
                <button key={item.qId} onClick={() => setActiveQuestion(item.qId)} style={{
                  padding: "12px 14px", borderRadius: 2, border: `1px solid ${activeQuestion === item.qId ? T.accent : T.rule}`,
                  background: activeQuestion === item.qId ? T.accentSoft : T.bg, textAlign: "left", cursor: "pointer", fontFamily: "var(--ei-sans)",
                }}>
                  <div style={{ display: "flex", justifyContent: "space-between", gap: 8, marginBottom: 5 }}>
                    <div className="ei-mono" style={{ fontSize: 11, color: T.ink3 }}>{item.qId.toUpperCase()}</div>
                    <Tag tone={item.state === "强项" ? "ok" : item.state === "待加强" ? "warn" : "cool"} T={T}>{item.state}</Tag>
                  </div>
                  <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>{item.topic}</div>
                </button>
              ))}
            </div>
          </div>
          <div style={{ minWidth: 0 }}>
            <div style={{ display: "flex", justifyContent: "space-between", gap: 16, alignItems: "flex-start", marginBottom: 16 }}>
              <div>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 5 }}>{q.qId.toUpperCase()} · {lang === "en" ? "ANSWER ANALYSIS" : "回答分析"}</div>
                <div className="ei-serif" style={{ fontSize: 26, color: T.ink, lineHeight: 1.25 }}>{q.topic}</div>
              </div>
              <Btn T={T} variant="secondary" size="sm" icon="replay" onClick={() => nav("workspace", { jobId: "tj-1" })}>{lang === "en" ? "Add to current-round replay" : "加入本轮复练"}</Btn>
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 12, marginBottom: 16 }}>
              <div style={{ background: T.okSoft, padding: 14, borderRadius: 2 }}>
                <div className="ei-label" style={{ color: T.ok, marginBottom: 6 }}>{lang === "en" ? "Answer worked" : "回答有效点"}</div>
                <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.6 }}>{q.good}</div>
              </div>
              <div style={{ background: q.state === "待加强" ? T.dangerSoft : T.bgSoft, padding: 14, borderRadius: 2 }}>
                <div className="ei-label" style={{ color: q.state === "待加强" ? T.danger : T.warn, marginBottom: 6 }}>{lang === "en" ? "Gap" : "缺口"}</div>
                <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.6 }}>{q.missing}</div>
              </div>
              <div style={{ background: T.bgSoft, padding: 14, borderRadius: 2 }}>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{lang === "en" ? "Suggested frame" : "建议框架"}</div>
                <div style={{ fontSize: 12.5, color: T.ink2, lineHeight: 1.6, fontFamily: "var(--ei-mono)" }}>{q.frame}</div>
              </div>
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
              <div style={{ padding: 16, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "Evidence snippet" : "证据片段"}</div>
                <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.65, fontFamily: "var(--ei-serif)" }}>
                  “{lang === "en" ? "I led the checkout migration and improved performance a lot." : "我主导了结账链路迁移，性能有明显提升。"}”
                </div>
                <div style={{ marginTop: 10, fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{lang === "en" ? "Needs metric: LCP / abandonment / GMV" : "需要补指标：LCP / 流失率 / GMV"}</div>
              </div>
              <div style={{ padding: 16, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
                <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{lang === "en" ? "Next interviewer prompt" : "下次面试追问"}</div>
                <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.65 }}>
                  {lang === "en" ? "What was the baseline, how did you isolate the bottleneck, and what changed after launch?" : "上线前的基线是什么？你如何定位瓶颈？上线后具体指标怎么变化？"}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {detail === "evidence" && (
        <div style={{ padding: 24, display: "grid", gridTemplateColumns: "1fr 1fr", gap: 18 }}>
          <div>
            <div className="ei-label" style={{ color: T.danger, marginBottom: 12 }}>▲ {lang === "en" ? "RISK EVIDENCE" : "风险证据详情"}</div>
            {r.issues.map((iss, i) => (
              <div key={i} style={{ padding: "14px 0", borderBottom: i < r.issues.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
                <div style={{ fontSize: 15, color: T.ink, fontWeight: 500 }}>{iss.title}</div>
                <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", margin: "4px 0 8px" }}>{iss.evidence}</div>
                <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.6 }}>{iss.body}</div>
              </div>
            ))}
          </div>
          <div>
            <div className="ei-label" style={{ color: T.ok, marginBottom: 12 }}>● {lang === "en" ? "REUSABLE PROOF" : "可复用亮点证据"}</div>
            {r.highlights.map((h, i) => (
              <div key={i} style={{ padding: "14px 0", borderBottom: i < r.highlights.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
                <div style={{ fontSize: 15, color: T.ink, fontWeight: 500 }}>{h.title}</div>
                <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.6, marginTop: 4 }}>{h.body}</div>
                <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 6 }}>{h.evidence}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {detail === "next" && (
        <div style={{ padding: 24, display: "grid", gridTemplateColumns: "1.15fr .85fr", gap: 18, alignItems: "start" }}>
          <div>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 12 }}>{lang === "en" ? "PATH A · REPLAY CURRENT ROUND" : `路径 A · 复练当前轮（${context.round}）`}</div>
            <div style={{ fontSize: 13, color: T.ink3, lineHeight: 1.65, marginBottom: 12 }}>
              {lang === "en"
                ? "This repeats the same interview round and injects the issues found in this report. Use it when readiness says replay."
                : "这不是推进到下一轮，而是重复当前轮次，把本报告里的错题、证据缺口和追问风险带进去。当前准备度为「建议再练」时，默认走这条。"}
            </div>
            {r.nextPractice.map((n, i) => (
              <div key={i} style={{ display: "grid", gridTemplateColumns: "30px 1fr auto", gap: 12, alignItems: "center", padding: "13px 0", borderBottom: i < r.nextPractice.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
                <div style={{ width: 24, height: 24, borderRadius: 12, background: i === 0 ? T.accent : T.bgSoft, color: i === 0 ? "#fff" : T.ink3, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontFamily: "var(--ei-mono)" }}>{i + 1}</div>
                <div style={{ fontSize: 14, color: T.ink2 }}>{n}</div>
                <Tag tone={i === 0 ? "warn" : "muted"} T={T}>{i === 0 ? (lang === "en" ? "must include" : "必练") : (lang === "en" ? "planned" : "计划")}</Tag>
              </div>
            ))}
            <div style={{ marginTop: 16 }}>
              <Btn T={T} variant="accent" icon="replay" onClick={() => nav("workspace", { jobId: "tj-1" })}>{lang === "en" ? `Replay ${context.round}` : `复练当前轮：${context.round}`}</Btn>
            </div>
          </div>
          <div style={{ padding: 18, background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "PATH B · PREPARE NEXT ROUND" : `路径 B · 进入下一轮（${context.nextRound}）`}</div>
            <div className="ei-serif" style={{ fontSize: 20, color: T.ink, lineHeight: 1.35, marginBottom: 10 }}>
              {lang === "en" ? "Use only when you decide to move forward." : "这是另一种面试规划，不是本报告默认复练。"}
            </div>
            <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.65, marginBottom: 14 }}>
              {lang === "en"
                ? "This creates a mock for the next interview round. It should inherit the same JD and resume, but change the interviewer focus and question mix."
                : "进入下一轮会沿用同一 JD 与简历，但切换到下一轮面试官视角和题目结构，例如从技术一面进入技术二面。"}
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 8, marginBottom: 14 }}>
              {[
                lang === "en" ? "Keep target job and resume unchanged" : "目标岗位与绑定简历不变",
                lang === "en" ? "Change round focus and interviewer style" : "改变轮次重点和面试官风格",
                lang === "en" ? "Carry forward unresolved risks as hidden probes" : "把未解决风险作为隐藏追问带入",
              ].map((item) => (
                <div key={item} style={{ display: "flex", gap: 8, alignItems: "center", fontSize: 12.5, color: T.ink2 }}>
                  <Icon name="check" size={12} color={T.ok} /> {item}
                </div>
              ))}
            </div>
            <Btn T={T} variant="secondary" icon="arrow_right" onClick={() => nav("workspace", { jobId: "tj-1" })}>{lang === "en" ? `Prepare ${context.nextRound}` : `准备下一轮：${context.nextRound}`}</Btn>
          </div>
        </div>
      )}
    </Card>
  );
};

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
          <Btn variant="secondary" size="sm" T={T} icon="play" onClick={() => nav("workspace", { jobId: "tj-1" })}>{L.openDrill}</Btn>
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
