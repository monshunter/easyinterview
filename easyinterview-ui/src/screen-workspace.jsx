// Screen 2: Target Job Workspace
const WorkspaceScreen = ({ T, lang, nav, jobId }) => {
  const D = window.EI_DATA;
  const job = D.targetJobs.find((j) => j.id === jobId) || D.targetJobs[0];
  const jd = D.jdSample;

  const L = lang === "en" ? {
    back: "All jobs",
    overview: "Overview",
    requirements: "Requirements",
    prep: "My preparation",
    practices: "Practice history",
    materials: "Materials",
    timeline: "Real progress",
    startCore: "Start JD-tailored core",
    startWarm: "Quick warm-up",
    startReal: "Real-interview replay",
    startAsk: "Reverse-Q drill",
    must: "Must have",
    nice: "Nice to have",
    hidden: "Hidden signals",
    rounds: "Round assumptions",
    risks: "Risks flagged",
    strongs: "Direct hits",
    nextPractice: "What to practice next",
    lastReport: "Last report",
    gotoReport: "Open full report",
    notePractice: "Every session here is tied to this JD — the question generator reads Must Have / Nice to Have / Hidden Signals above.",
  } : {
    back: "返回岗位列表",
    overview: "概览",
    requirements: "要求拆解",
    prep: "我的准备",
    practices: "练习历史",
    materials: "材料工坊",
    timeline: "真实进展",
    startCore: "岗位定制核心面",
    startWarm: "快速热身",
    startReal: "真实面试复现",
    startAsk: "反问环节专练",
    must: "必需项",
    nice: "加分项",
    hidden: "隐性关注点",
    rounds: "轮次假设",
    risks: "风险提示",
    strongs: "直接命中",
    nextPractice: "下一步练什么",
    lastReport: "最近一次报告",
    gotoReport: "打开完整报告",
    notePractice: "这里的每一次练习都绑定这份 JD——题目生成器会直接读取上方的必需项 / 加分项 / 隐性关注点。",
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 1280, margin: "0 auto", padding: "32px 48px 96px" }}>
      {/* crumbs */}
      <button onClick={() => nav("home")} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, display: "flex", alignItems: "center", gap: 6, padding: 0, marginBottom: 20, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {L.back}
      </button>

      {/* header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 24, flexWrap: "wrap", marginBottom: 32 }}>
        <div style={{ flex: 1, minWidth: 320 }}>
          <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 10 }}>
            <Tag tone="amber" T={T}>{job.status}</Tag>
            <Tag tone="muted" T={T}>{job.level}</Tag>
            <span style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{lang === "en" ? "UPDATED" : "更新于"} {job.updatedAt}</span>
          </div>
          <h1 className="ei-serif" style={{ fontSize: 38, color: T.ink, margin: 0, letterSpacing: "-0.02em", lineHeight: 1.15 }}>
            {job.title}
          </h1>
          <div style={{ fontSize: 15, color: T.ink2, marginTop: 6 }}>
            {job.company} · {job.location} · <span style={{ color: T.ink3 }}>{job.source}</span>
          </div>
        </div>

        <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: 10 }}>
          <ReadinessDial level={job.readiness} label={job.readinessLabel} T={T} size={56} />
          <div style={{ fontSize: 12, color: T.ink3 }}>{lang === "en" ? "JD match" : "JD 匹配度"} <b style={{ color: T.ink, fontFamily: "var(--ei-mono)" }}>{job.match}%</b></div>
        </div>
      </div>

      {/* Practice launcher */}
      <div style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 22, marginBottom: 32 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14, gap: 16, flexWrap: "wrap" }}>
          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{lang === "en" ? "PRACTICE · JD-LOCKED" : "练习 · 已锁定此 JD"}</div>
            <div className="ei-serif" style={{ fontSize: 20, color: T.ink }}>{lang === "en" ? "Pick a mode and go. Every answer feeds this job's mistake book." : "挑一个模式直接开始。每一道回答都会沉淀到这个岗位的错题本。"}</div>
          </div>
          <Btn variant="accent" icon="play" onClick={() => nav("practice", { jobId })} T={T}>{L.startCore}</Btn>
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: 10 }}>
          <ModeCard T={T} icon="spark" title={L.startWarm} sub={lang === "en" ? "5–8 min · icebreaker" : "5–8 分钟 · 破冰"} onClick={() => nav("practice", { jobId, mode: "warm" })} />
          <ModeCard T={T} icon="replay" title={L.startReal} sub={lang === "en" ? "Paste real questions" : "粘贴真实问过的题"} onClick={() => nav("debrief")} />
          <ModeCard T={T} icon="chat" title={L.startAsk} sub={lang === "en" ? "Just reverse-Q" : "只练反问环节"} onClick={() => nav("practice", { jobId, mode: "reverse" })} />
          <ModeCard T={T} icon="mic" title={lang === "en" ? "Voice mode · P2" : "语音模式 · P2"} sub={lang === "en" ? "Waveform + pace feedback" : "波形 + 语速停顿反馈"} onClick={() => nav("voice")} />
          <ModeCard T={T} icon="layers" title={lang === "en" ? "Multi-round plan" : "多轮计划编排"} sub={lang === "en" ? "HR → tech → manager chain" : "HR → 技术 → 经理串联"} onClick={() => nav("plan")} />
          <ModeCard T={T} icon="target" title={lang === "en" ? "Single-question drill" : "单题深钻"} sub={lang === "en" ? "Fix one weak point" : "专攻一道错题"} onClick={() => nav("practice", { jobId, mode: "drill" })} />
        </div>
        <div style={{ fontSize: 12, color: T.ink3, marginTop: 12, display: "flex", gap: 6, alignItems: "center" }}>
          <Icon name="info" size={12} /> {L.notePractice}
        </div>
      </div>

      {/* 2-column */}
      <div style={{ display: "grid", gridTemplateColumns: "1.4fr 1fr", gap: 24 }}>
        {/* left column */}
        <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
          {/* company intel — embed */}
          <CompanyIntelEmbed T={T} lang={lang} nav={nav} />

          {/* requirements */}
          <Card T={T} pad={0}>
            <div style={{ padding: "16px 20px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <div>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 2 }}>{lang === "en" ? "PARSED JD" : "JD 拆解"}</div>
                <div className="ei-serif" style={{ fontSize: 17, color: T.ink }}>{L.requirements}</div>
              </div>
              <span className="ei-mono" style={{ fontSize: 11, color: T.ink3 }}>parsed · {job.updatedAt}</span>
            </div>
            <div style={{ padding: 20 }}>
              <ReqBlock T={T} title={L.must} items={jd.mustHave} tone="accent" hits={job.hits} />
              <ReqBlock T={T} title={L.nice} items={jd.nice} tone="amber" hits={job.hits} />
              <ReqBlock T={T} title={L.hidden} items={jd.hidden} tone="cool" />
            </div>
          </Card>

          {/* rounds */}
          <Card T={T} pad={0}>
            <div style={{ padding: "16px 20px", borderBottom: `1px solid ${T.rule}` }}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 2 }}>{lang === "en" ? "INTERVIEW LOOP" : "面试流程假设"}</div>
              <div className="ei-serif" style={{ fontSize: 17, color: T.ink }}>{L.rounds}</div>
            </div>
            <div style={{ padding: "8px 0" }}>
              {jd.rounds.map((r, i) => (
                <div key={i} style={{ padding: "14px 20px", borderBottom: i < jd.rounds.length - 1 ? `1px dotted ${T.rule}` : "none", display: "flex", alignItems: "center", gap: 16 }}>
                  <div style={{ width: 26, height: 26, borderRadius: 13, border: `1px solid ${i === 1 ? T.accent : T.rule}`, background: i === 1 ? T.accentSoft : "transparent", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 12, color: i === 1 ? T.accent : T.ink3, fontFamily: "var(--ei-mono)" }}>
                    {i + 1}
                  </div>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 14, color: T.ink, fontWeight: 500 }}>{r.name}</div>
                    <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 2 }}>{r.focus}</div>
                  </div>
                  {i === 1 && <Tag tone="accent" T={T}>{lang === "en" ? "Current" : "当前"}</Tag>}
                  {i > 1 && <Tag tone="muted" T={T}>{lang === "en" ? "Future" : "后续"}</Tag>}
                  {i === 0 && <Tag tone="ok" T={T}>{lang === "en" ? "Cleared" : "已通过"}</Tag>}
                </div>
              ))}
            </div>
          </Card>

          {/* practice history */}
          <Card T={T} pad={0}>
            <div style={{ padding: "16px 20px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <div>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 2 }}>{lang === "en" ? "SESSIONS" : "会话"}</div>
                <div className="ei-serif" style={{ fontSize: 17, color: T.ink }}>{L.practices}</div>
              </div>
            </div>
            <div>
              {D.growth.recent.slice(0, 4).map((r, i) => (
                <div key={i} style={{ padding: "14px 20px", borderBottom: i < 3 ? `1px dotted ${T.rule}` : "none", display: "flex", alignItems: "center", gap: 16 }}
                  onClick={() => nav("report")} role="button">
                  <div className="ei-mono" style={{ fontSize: 12, color: T.ink3, width: 40 }}>{r.date}</div>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 14, color: T.ink }}>{r.mode}</div>
                    <div style={{ fontSize: 12, color: T.ink3 }}>{r.job}</div>
                  </div>
                  <ReadinessDial level={r.readiness} T={T} size={34} />
                  <Icon name="chevron_right" size={14} color={T.ink3} />
                </div>
              ))}
            </div>
          </Card>
        </div>

        {/* right column */}
        <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
          {/* risks & strengths */}
          <Card T={T}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{L.prep}</div>
            <div style={{ marginBottom: 14 }}>
              <div style={{ fontSize: 12.5, color: T.ok, fontWeight: 500, marginBottom: 6 }}>● {L.strongs}</div>
              {job.hits.map((h, i) => (
                <div key={i} style={{ fontSize: 13, color: T.ink2, padding: "4px 0" }}>{h}</div>
              ))}
            </div>
            <div>
              <div style={{ fontSize: 12.5, color: T.danger, fontWeight: 500, marginBottom: 6 }}>● {L.risks}</div>
              {job.gaps.map((g, i) => (
                <div key={i} style={{ fontSize: 13, color: T.ink2, padding: "4px 0" }}>{g}</div>
              ))}
            </div>
          </Card>

          {/* materials */}
          <Card T={T} pad={0}>
            <div style={{ padding: "16px 20px", borderBottom: `1px solid ${T.rule}` }}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 2 }}>{lang === "en" ? "MATERIALS" : "材料"}</div>
              <div className="ei-serif" style={{ fontSize: 17, color: T.ink }}>{L.materials}</div>
            </div>
            <div>
              <MaterialRow T={T} icon="resume" title={lang === "en" ? "Resume · v3 tailored" : "简历 · v3 定制版"} meta={lang === "en" ? "78% match" : "匹配 78%"} onClick={() => nav("resume")} />
              <MaterialRow T={T} icon="chat" title={lang === "en" ? "Reverse-Q draft" : "反问问题草稿"} meta={lang === "en" ? "3 questions" : "3 条"} />
              <MaterialRow T={T} icon="book" title={lang === "en" ? "Mistake book (JD-scoped)" : "错题本（本岗位）"} meta={lang === "en" ? "3 open" : "3 条未攻克"} onClick={() => nav("mistakes")} />
            </div>
          </Card>

          {/* next */}
          <Card T={T} style={{ background: T.accentSoft, borderColor: "transparent" }}>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{L.nextPractice}</div>
            <div style={{ fontSize: 14, color: T.ink2, lineHeight: 1.55 }}>
              {lang === "en" ?
                "Nail the quantified performance story (Q2), then add one role-specific reverse-Q." :
                "把性能优化故事补上量化结果（Q2），再加一个针对本公司的反问。"}
            </div>
            <Btn variant="primary" size="sm" icon="play" onClick={() => nav("practice", { jobId })} T={T} style={{ marginTop: 14 }}>
              {lang === "en" ? "Start 15-min drill" : "开始 15 分钟深钻"}
            </Btn>
          </Card>
        </div>
      </div>
    </div>
  );
};

const ModeCard = ({ T, icon, title, sub, onClick }) => (
  <button onClick={onClick} style={{
    textAlign: "left", background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2,
    padding: "12px 14px", display: "flex", gap: 12, alignItems: "center", cursor: "pointer", transition: "border-color .15s",
  }}
    onMouseEnter={(e) => e.currentTarget.style.borderColor = T.ink3}
    onMouseLeave={(e) => e.currentTarget.style.borderColor = T.rule}
  >
    <div style={{ width: 32, height: 32, borderRadius: 2, background: T.bgCard, border: `1px solid ${T.rule}`, display: "flex", alignItems: "center", justifyContent: "center", color: T.ink2 }}>
      <Icon name={icon} size={16} />
    </div>
    <div>
      <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>{title}</div>
      <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 2 }}>{sub}</div>
    </div>
  </button>
);

const ReqBlock = ({ T, title, items, tone, hits = [] }) => (
  <div style={{ marginBottom: 18 }}>
    <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 8 }}>
      <Tag tone={tone} T={T}>{title}</Tag>
    </div>
    <div>
      {items.map((t, i) => {
        const hit = hits.some((h) => t.includes(h) || h.includes(t.slice(0, 3)));
        return (
          <div key={i} style={{ padding: "7px 0", fontSize: 13.5, color: T.ink2, display: "flex", gap: 10, lineHeight: 1.5 }}>
            <span style={{ color: hit ? T.ok : T.ink4, flexShrink: 0, marginTop: 2 }}>{hit ? "●" : "○"}</span>
            <span>{t}</span>
          </div>
        );
      })}
    </div>
  </div>
);

const MaterialRow = ({ T, icon, title, meta, onClick }) => (
  <div onClick={onClick} style={{ padding: "14px 20px", display: "flex", alignItems: "center", gap: 12, borderBottom: `1px dotted ${T.rule}`, cursor: onClick ? "pointer" : "default" }}>
    <Icon name={icon} size={16} color={T.ink3} />
    <div style={{ flex: 1, fontSize: 13.5, color: T.ink }}>{title}</div>
    <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{meta}</div>
    {onClick && <Icon name="chevron_right" size={14} color={T.ink3} />}
  </div>
);

window.WorkspaceScreen = WorkspaceScreen;
