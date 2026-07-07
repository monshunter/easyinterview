// Screen 2: Mock Interview Setup
const WorkspaceScreen = ({ T, lang, nav, params = {}, requestAuth }) => {
  const D = window.EI_DATA;
  const initialContext = window.eiCreateInterviewContext ? window.eiCreateInterviewContext(params) : params;
  const [activeJobId, setActiveJobId] = React.useState(initialContext.targetJobId || initialContext.jobId);
  const [selectedResumeId, setSelectedResumeId] = React.useState(initialContext.resumeId || "frontend-v3");
  const [resumePickerOpen, setResumePickerOpen] = React.useState(false);
  const [plannerOpen, setPlannerOpen] = React.useState(false);
  React.useEffect(() => {
    const nextJobId = params.targetJobId || params.jobId;
    if (nextJobId) setActiveJobId(nextJobId);
    if (params.resumeId) setSelectedResumeId(params.resumeId);
  }, [params.targetJobId, params.jobId, params.resumeId]);

  const jobs = D.targetJobs || [];
  const resumeOptions = getWorkspaceResumeOptions(lang);
  const planOptions = getWorkspacePlanOptions(lang, jobs);
  const activePlan = planOptions.find((plan) => plan.jobId === activeJobId) || planOptions[0];
  const job = jobs.find((j) => j.id === activeJobId) || jobs[0];
  const jd = getWorkspaceJDSample(job, D.jdSample);
  const currentRoundIndex = getCurrentRoundIndex(job, jd.rounds);
  const currentRound = jd.rounds[currentRoundIndex] || jd.rounds[0];
  const selectedResume = resumeOptions.find((resume) => resume.id === selectedResumeId);
  const interviewContext = createWorkspaceInterviewContext(activePlan, job, jd, currentRound, selectedResume, params);
  const sessionHistory = getWorkspaceSessionHistory(lang, job, currentRound?.name, interviewContext);

  const L = lang === "en" ? {
    back: "Home",
    overview: "Overview",
    requirements: "Requirements",
    prep: "My preparation",
    practices: "Current plan mock records",
    timeline: "Real progress",
    startCore: "Start interview now",
    launchTitle: "Confirm the context before this mock interview starts.",
    launchSub: "The target job, JD, and resume form the interview context. Text and voice can be switched inside the interview.",
    flow: "Interview rounds",
    roundStatus: "Current round",
    jdBound: "Target job / JD",
    resumeBound: "Bound resume",
    changeResume: "Change",
    prepStatus: "Preparation status",
    jdMatch: "JD match",
    sessionTag: "Completed",
    reportReady: "Report ready",
    planEyebrow: "Current mock plan",
    planSub: "A plan is the JD + resume + interview round context used to generate this mock interview.",
    switchPlan: "Switch plan",
    createPlan: "New plan",
    must: "Must have",
    nice: "Nice to have",
    hidden: "Hidden signals",
    risks: "Risks flagged",
    strongs: "Direct hits",
    lastReport: "Last report",
    gotoReport: "Open full report",
    notePractice: "Every question in this mock interview reads the JD requirements, resume evidence, risks, and previous report signals.",
  } : {
    back: "返回首页",
    overview: "概览",
    requirements: "要求拆解",
    prep: "我的准备",
    practices: "当前规划的模拟面试记录",
    timeline: "面试进展",
    startCore: "立即面试",
    launchTitle: "开始前确认这场模拟面试的上下文。",
    launchSub: "目标岗位、JD 和简历组成这场模拟面试的上下文，文本和语音可在面试过程中切换。",
    flow: "面试轮次",
    roundStatus: "当前轮次",
    jdBound: "目标岗位 / JD",
    resumeBound: "绑定简历",
    changeResume: "更换",
    prepStatus: "准备状态",
    jdMatch: "JD 匹配度",
    sessionTag: "已完成",
    reportReady: "报告已生成",
    planEyebrow: "当前面试规划",
    planSub: "面试规划就是这场模拟面试使用的 JD、简历和目标轮次组合。",
    switchPlan: "切换规划",
    createPlan: "新建规划",
    must: "必需项",
    nice: "加分项",
    hidden: "隐性关注点",
    risks: "风险提示",
    strongs: "直接命中",
    lastReport: "最近一次报告",
    gotoReport: "打开完整报告",
    notePractice: "这场模拟面试中的每一道题都会读取 JD 要求、简历证据、风险提示和过往报告信号。",
  };

  if (!job) return <WorkspaceEmptyState T={T} lang={lang} nav={nav} />;
  if (!selectedResume) return <WorkspaceMissingResumeState T={T} lang={lang} nav={nav} />;

  const startContext = {
    ...interviewContext,
    sessionId: `session-${interviewContext.planId}-${interviewContext.roundId}-new`,
    mode: "text",
    modality: "text",
    practiceMode: params.practiceMode || "strict",
    hintUsed: "false",
  };
  const startInterview = () => {
    const run = () => nav("practice", startContext);
    if (!requestAuth) {
      run();
      return;
    }
    requestAuth({
      type: "create_session",
      label: L.startCore,
      route: "practice",
      params: startContext,
    }, run);
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 1280, margin: "0 auto", padding: "32px 48px 96px" }}>
      {/* crumbs */}
      <button onClick={() => nav("home")} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, display: "flex", alignItems: "center", gap: 6, padding: 0, marginBottom: 20, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {L.back}
      </button>

      <div style={{
        background: T.bgCard,
        border: `1px solid ${T.rule}`,
        borderRadius: 3,
        padding: "14px 16px",
        marginBottom: 24,
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        gap: 18,
        flexWrap: "wrap",
      }}>
        <div style={{ minWidth: 280 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 5 }}>{L.planEyebrow}</div>
          <div style={{ display: "flex", alignItems: "center", gap: 10, flexWrap: "wrap" }}>
            <div className="ei-serif" style={{ fontSize: 18, color: T.ink }}>{job.company} · {job.title}</div>
            <Tag tone={job.statusTone === "neutral" ? "muted" : job.statusTone || "amber"} T={T}>{job.status}</Tag>
          </div>
          <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 5, lineHeight: 1.55 }}>
            {L.planSub} <span style={{ color: T.ink2 }}>{activePlan.round} · {selectedResume.name}</span>
          </div>
        </div>
        <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
          <Btn variant="secondary" icon="layers" onClick={() => setPlannerOpen(true)} T={T}>{L.switchPlan}</Btn>
          <Btn variant="ghost" icon="plus" onClick={() => nav("home")} T={T}>{L.createPlan}</Btn>
        </div>
      </div>

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

        <div style={{ minWidth: 168, textAlign: "right", paddingTop: 4 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{L.prepStatus}</div>
          <div className="ei-serif" style={{ fontSize: 22, color: T.ink, marginBottom: 8 }}>{job.readinessLabel}</div>
          <div style={{ fontSize: 12, color: T.ink3 }}>{L.jdMatch} <b style={{ color: T.ink, fontFamily: "var(--ei-mono)" }}>{job.match}%</b></div>
        </div>
      </div>

      {/* Interview launcher */}
      <div style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 22, marginBottom: 32 }}>
        <InterviewRoundRail T={T} lang={lang} label={L.flow} rounds={jd.rounds} currentIndex={currentRoundIndex} nextRound={job.nextRound} />
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginTop: 22, marginBottom: 18, gap: 20, flexWrap: "wrap" }}>
          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{lang === "en" ? "INTERVIEW SETUP" : "面试前确认"}</div>
            <div className="ei-serif" style={{ fontSize: 21, color: T.ink }}>{L.launchTitle}</div>
            <div style={{ fontSize: 13.5, color: T.ink2, marginTop: 6 }}>
              {L.roundStatus} · <b style={{ color: T.ink }}>{currentRound?.name}</b>
              {job.nextRound && <span style={{ color: T.ink3 }}> · {job.nextRound}</span>}
            </div>
            <div style={{ fontSize: 13.5, color: T.ink3, marginTop: 6, lineHeight: 1.6, maxWidth: 680 }}>{L.launchSub}</div>
          </div>
          <Btn variant="accent" icon="play" onClick={startInterview} T={T}>{L.startCore}</Btn>
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
          <BindingPill T={T} icon="briefcase" label={L.jdBound} title={job.title} meta={`${job.company} · ${job.level} · ${job.match}% ${lang === "en" ? "match" : "匹配"}`} />
          <BindingPill T={T} icon="resume" label={L.resumeBound} title={selectedResume.name} meta={selectedResume.meta} action={L.changeResume} onClick={() => setResumePickerOpen(true)} />
        </div>
        <div style={{ fontSize: 12, color: T.ink3, marginTop: 12, display: "flex", gap: 6, alignItems: "center" }}>
          <Icon name="info" size={12} /> {L.notePractice}
        </div>
      </div>

      {/* 2-column */}
      <div style={{ display: "grid", gridTemplateColumns: "1.4fr 1fr", gap: 24 }}>
        {/* left column */}
        <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
          {/* workspace insight summary */}
          <WorkspaceInsightCard T={T} lang={lang} job={job} />

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

          {/* practice records */}
          <Card T={T} pad={0}>
            <div style={{ padding: "16px 20px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <div>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 2 }}>{lang === "en" ? "SESSIONS" : "会话"}</div>
                <div className="ei-serif" style={{ fontSize: 17, color: T.ink }}>{L.practices}</div>
              </div>
            </div>
            <div>
              {sessionHistory.map((r, i) => (
                <div key={i} style={{ padding: "14px 20px", borderBottom: i < 3 ? `1px dotted ${T.rule}` : "none", display: "grid", gridTemplateColumns: "42px 1fr auto", gap: 12, alignItems: "center", cursor: "pointer" }}
                  onClick={() => nav("report", r.context)} role="button">
                  <div className="ei-mono" style={{ fontSize: 12, color: T.ink3 }}>{r.date}</div>
                  <div>
                    <div style={{ fontSize: 13.5, color: T.ink }}>{r.title}</div>
                    <div style={{ fontSize: 12, color: T.ink3, marginTop: 2 }}>{r.target}</div>
                  </div>
                  <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <Tag tone={i === 0 ? "cool" : "muted"} T={T}>{i === 0 ? L.reportReady : L.sessionTag}</Tag>
                    <Icon name="chevron_right" size={14} color={T.ink3} />
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>
      </div>

      {resumePickerOpen && (
        <ResumePickerModal
          T={T}
          lang={lang}
          resumes={resumeOptions}
          selectedId={selectedResumeId}
          onClose={() => setResumePickerOpen(false)}
          onConfirm={(resumeId) => {
            setSelectedResumeId(resumeId);
            setResumePickerOpen(false);
          }}
        />
      )}

      {plannerOpen && (
        <PlanSwitcherModal
          T={T}
          lang={lang}
          plans={planOptions}
          selectedJobId={activeJobId}
          onClose={() => setPlannerOpen(false)}
          onCreate={() => {
            setPlannerOpen(false);
            nav("home");
          }}
          onConfirm={(plan) => {
            setActiveJobId(plan.jobId);
            setSelectedResumeId(plan.resumeId);
            setPlannerOpen(false);
          }}
        />
      )}
    </div>
  );
};

const WorkspaceEmptyState = ({ T, lang, nav }) => (
  <div className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px 48px" }}>
    <Card T={T}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "NO JD CONTEXT" : "没有 JD 上下文"}</div>
      <div className="ei-serif" style={{ fontSize: 28, color: T.ink, lineHeight: 1.25, marginBottom: 10 }}>
        {lang === "en" ? "Import a target JD before opening a mock plan." : "先导入一个目标 JD，再进入面试规划。"}
      </div>
      <div style={{ fontSize: 14, color: T.ink3, lineHeight: 1.6, marginBottom: 18 }}>
        {lang === "en" ? "This empty state avoids showing a fake job when the plan context is missing." : "上下文缺失时不展示假岗位数据，避免用户误以为已经绑定了真实 JD。"}
      </div>
      <Btn T={T} variant="accent" iconRight="arrow_right" onClick={() => nav("home")}>{lang === "en" ? "Import JD" : "导入 JD"}</Btn>
    </Card>
  </div>
);

const WorkspaceMissingResumeState = ({ T, lang, nav }) => (
  <div className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px 48px" }}>
    <Card T={T}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "NO RESUME BOUND" : "没有绑定简历"}</div>
      <div className="ei-serif" style={{ fontSize: 28, color: T.ink, lineHeight: 1.25, marginBottom: 10 }}>
        {lang === "en" ? "Bind or create a resume before starting this mock interview." : "开始这场模拟面试前，需要先绑定或创建一份简历。"}
      </div>
      <div style={{ fontSize: 14, color: T.ink3, lineHeight: 1.6, marginBottom: 18 }}>
        {lang === "en" ? "The interview generator needs resume evidence; the prototype should not fill in a synthetic resume." : "面试生成需要简历证据；静态稿不能用假简历补位。"}
      </div>
      <Btn T={T} variant="accent" iconRight="arrow_right" onClick={() => nav("resume_versions", { flow: "create" })}>{lang === "en" ? "Create resume" : "创建简历"}</Btn>
    </Card>
  </div>
);

const createWorkspaceInterviewContext = (plan, job, jd, round, resume, params = {}) => {
  const targetJobId = job?.id || params.targetJobId || params.jobId;
  const roundId = getWorkspaceRoundId(round?.name || plan?.round || params.roundName);
  const base = {
    planId: plan?.id,
    targetJobId,
    jobId: targetJobId,
    jdId: plan?.jdId || (targetJobId ? `jd-${targetJobId}` : undefined),
    resumeId: resume?.id || params.resumeId,
    roundId,
    roundName: round?.name || plan?.round || params.roundName,
  };
  return window.eiCreateInterviewContext ? window.eiCreateInterviewContext({ ...params, ...base }, base) : base;
};

const getWorkspaceRoundId = (value) => {
  const text = value || "";
  if (text.includes("HR")) return "round-hr";
  if (text.includes("技术一面") || text.includes("Technical round 1")) return "round-tech-1";
  if (text.includes("技术二面") || text.includes("Technical round 2")) return "round-tech-2";
  if (text.includes("经理面") || text.includes("Manager")) return "round-manager";
  return "round-draft";
};

const getWorkspaceResumeOptions = (lang) => lang === "en" ? [
  {
    id: "frontend-v3",
    name: "Liu Zhe · Frontend Platform v3",
    meta: "78% match · source: Liu-Zhe-Frontend-2026.pdf",
    note: "Highlights React depth, performance work, accessibility, and platform experience.",
  },
  {
    id: "impact-v2",
    name: "Liu Zhe · Collaboration Impact v2",
    meta: "Created from pasted text · 2026-04-18",
    note: "Highlights cross-team influence, Design System rollout, and mentoring examples.",
  },
  {
    id: "english-v1",
    name: "Liu Zhe · Frontend Platform EN v1",
    meta: "English resume · source retained from original upload",
    note: "Used for English HR screens and overseas platform roles.",
  },
] : [
  {
    id: "frontend-v3",
    name: "刘哲 · 前端平台版 v3",
    meta: "匹配 78% · 原件：刘哲-前端-2026.pdf",
    note: "突出 React 深度、性能优化、可访问性与平台工程经验。",
  },
  {
    id: "impact-v2",
    name: "刘哲 · 协作影响力版 v2",
    meta: "粘贴文本创建 · 2026-04-18",
    note: "突出跨团队推动、Design System 落地和新人带教案例。",
  },
  {
    id: "english-v1",
    name: "Liu Zhe · Frontend Platform EN v1",
    meta: "英文简历 · 保留上传原件来源",
    note: "用于英文 HR 初筛和海外平台类岗位。",
  },
];

const getWorkspaceSessionHistory = (lang, job, roundName, context) => {
  const target = getWorkspaceTargetLabel(lang, job);
  const currentRound = getWorkspaceRoundLabel(lang, roundName || job?.nextRound);
  const priorRound = getWorkspaceRoundLabel(lang, "技术一面");
  const nextRound = getWorkspaceRoundLabel(lang, "技术二面");
  const prefix = lang === "en" ? "Mock interview" : "模拟面试";
  const rerun = lang === "en" ? "second run" : "第 2 次";
  const sessionContext = (id, roundLabel, roundId, modality = "text", practiceMode = "strict", hintUsed = "false") => ({
    ...context,
    sessionId: id,
    roundId,
    roundName: roundLabel,
    modality,
    practiceMode,
    hintUsed,
  });

  return [
    { date: "4/20", title: `${prefix} · ${currentRound}`, target, context: sessionContext("session-24", currentRound, context.roundId, "text", "strict", "false") },
    { date: "4/19", title: `${prefix} · ${currentRound}`, target: `${target} · ${rerun}`, context: sessionContext("session-23", currentRound, context.roundId, "voice", "assisted", "true") },
    { date: "4/18", title: `${prefix} · ${priorRound}`, target, context: sessionContext("session-20", priorRound, "round-tech-1") },
    { date: "4/17", title: `${prefix} · ${nextRound}`, target, context: sessionContext("session-19", nextRound, "round-tech-2") },
  ];
};

const getWorkspaceTargetLabel = (lang, job) => {
  if (!job) return lang === "en" ? "Target job" : "目标岗位";
  if (lang !== "en") return `${job.company} · ${job.title}`;
  const labels = {
    "tj-1": "Star-Ring Tech · Senior Frontend Engineer",
    "tj-3": "CloudYun Group · Web Architecture Expert",
  };
  return labels[job.id] || `${job.company} · ${job.title}`;
};

const getWorkspaceRoundLabel = (lang, value) => {
  const text = value || "";
  if (lang !== "en") return text || "目标轮次";
  if (text.includes("HR")) return "HR screen";
  if (text.includes("技术一面")) return "Technical round 1";
  if (text.includes("技术二面")) return "Technical round 2";
  if (text.includes("经理面")) return "Manager round";
  return "Target round";
};

const getWorkspacePlanOptions = (lang, jobs) => {
  const roundNames = lang === "en" ? ["Manager round", "HR screen", "Unscheduled draft"] : ["经理面", "HR 初筛", "未安排"];
  const resumeNames = lang === "en" ? [
    "Liu Zhe · Frontend Platform v3",
    "Liu Zhe · Frontend Platform EN v1",
    "Liu Zhe · Collaboration Impact v2",
  ] : [
    "刘哲 · 前端平台版 v3",
    "Liu Zhe · Frontend Platform EN v1",
    "刘哲 · 协作影响力版 v2",
  ];
  const resumeIds = ["frontend-v3", "english-v1", "impact-v2"];
  return jobs.map((job, i) => ({
    id: `plan-${job.id}`,
    jobId: job.id,
    targetJobId: job.id,
    jdId: `jd-${job.id}`,
    title: `${job.company} · ${job.title}`,
    meta: `${job.level} · ${job.match}% ${lang === "en" ? "match" : "匹配"} · ${job.source}`,
    round: roundNames[i] || job.nextRound || (lang === "en" ? "Next round" : "下一轮"),
    next: job.nextRound || (lang === "en" ? "Not scheduled" : "未安排"),
    status: job.status,
    statusTone: job.statusTone === "neutral" ? "muted" : job.statusTone || "amber",
    resumeId: resumeIds[i] || "frontend-v3",
    resumeName: resumeNames[i] || resumeNames[0],
    updatedAt: job.updatedAt,
  }));
};

const getWorkspaceJDSample = (job, fallback) => {
  if (job?.id === "tj-2") {
    return {
      ...fallback,
      mustHave: [
        "3 年以上前端平台 / 工程效率相关经验",
        "熟悉 TypeScript、Monorepo 与现代构建链路",
        "能用英文清晰解释技术方案、边界和取舍",
      ],
      nice: [
        "有 Design System / 开发者体验平台建设经验",
        "熟悉远程跨时区协作节奏",
        "能把平台能力沉淀成文档、规范和可观测指标",
      ],
      hidden: [
        "岗位更关注平台抽象边界，而不是单点业务页面交付",
        "英文 HR 初筛会关注动机、远程协作和表达节奏",
        "面试官可能追问平台投入如何证明业务价值",
      ],
    };
  }
  if (job?.id === "tj-3") {
    return {
      ...fallback,
      mustHave: [
        "有大型 Web 架构设计和技术决策经验",
        "能推动跨团队技术方案落地",
        "熟悉复杂业务系统的性能、稳定性和演进治理",
      ],
      nice: [
        "有技术委员会 / 架构评审经验",
        "能培养技术骨干并建立工程规范",
        "熟悉平台化、组件化或微前端治理",
      ],
      hidden: [
        "P7 更关注影响力和判断力，而不是单个功能实现",
        "草稿状态下需要先确认 JD、简历和目标轮次",
        "面试官大概率会深挖技术决策背后的组织成本",
      ],
    };
  }
  return fallback;
};

const ResumePickerModal = ({ T, lang, resumes, selectedId, onClose, onConfirm }) => {
  const [draftId, setDraftId] = React.useState(selectedId);
  const selected = resumes.find((resume) => resume.id === draftId) || resumes[0];
  return (
    <div style={{ position: "fixed", inset: 0, background: "rgba(24, 20, 16, 0.24)", zIndex: 80, display: "flex", alignItems: "center", justifyContent: "center", padding: 24 }} onClick={onClose}>
      <div className="ei-fadein" onClick={(e) => e.stopPropagation()} style={{ width: "min(680px, 100%)", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4, boxShadow: "0 24px 70px rgba(30, 22, 15, 0.24)", padding: 24 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 18, marginBottom: 18 }}>
          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{lang === "en" ? "RESUME CONTEXT" : "简历上下文"}</div>
            <div className="ei-serif" style={{ fontSize: 23, color: T.ink }}>
              {lang === "en" ? "Choose the resume for this mock interview" : "选择这场模拟面试使用的简历"}
            </div>
            <div style={{ fontSize: 13, color: T.ink3, marginTop: 6, lineHeight: 1.6 }}>
              {lang === "en" ? "Each resume keeps a name, source, and original content." : "每份简历都会保留名称、来源和原始内容。"}
            </div>
          </div>
          <button onClick={onClose} style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", padding: 4 }}>
            <Icon name="x" size={16} />
          </button>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "1fr", gap: 10 }}>
          {resumes.map((resume) => {
            const active = resume.id === draftId;
            return (
              <button
                key={resume.id}
                onClick={() => setDraftId(resume.id)}
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
                  <span style={{ display: "block", fontSize: 14.5, color: T.ink, fontWeight: 600 }}>{resume.name}</span>
                  <span style={{ display: "block", fontSize: 12.5, color: T.ink3, marginTop: 3 }}>{resume.meta}</span>
                  <span style={{ display: "block", fontSize: 13, color: T.ink2, marginTop: 8, lineHeight: 1.55 }}>{resume.note}</span>
                </span>
              </button>
            );
          })}
        </div>

        <div style={{ border: `1px solid ${T.rule}`, background: T.bgSoft, borderRadius: 3, padding: 14, marginTop: 16 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{lang === "en" ? "WILL BE USED AS" : "将作为"}</div>
          <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.6 }}>
            {lang === "en" ? `${selected.name} will be used as answer evidence for this mock interview.` : `${selected.name} 将作为这场模拟面试的回答证据。`}
          </div>
        </div>

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 10, marginTop: 22 }}>
          <Btn T={T} variant="ghost" onClick={onClose}>{lang === "en" ? "Cancel" : "取消"}</Btn>
          <Btn T={T} variant="accent" iconRight="arrow_right" onClick={() => onConfirm(draftId)}>{lang === "en" ? "Use this resume" : "使用这份简历"}</Btn>
        </div>
      </div>
    </div>
  );
};

const PlanSwitcherModal = ({ T, lang, plans, selectedJobId, onClose, onCreate, onConfirm }) => {
  const [draftJobId, setDraftJobId] = React.useState(selectedJobId);
  const selected = plans.find((plan) => plan.jobId === draftJobId) || plans[0];
  return (
    <div style={{ position: "fixed", inset: 0, background: "rgba(24, 20, 16, 0.24)", zIndex: 80, display: "flex", alignItems: "center", justifyContent: "center", padding: 24 }} onClick={onClose}>
      <div className="ei-fadein" onClick={(e) => e.stopPropagation()} style={{ width: "min(760px, 100%)", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4, boxShadow: "0 24px 70px rgba(30, 22, 15, 0.24)", padding: 24 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 18, marginBottom: 18 }}>
          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{lang === "en" ? "MOCK PLAN" : "面试规划"}</div>
            <div className="ei-serif" style={{ fontSize: 23, color: T.ink }}>
              {lang === "en" ? "Switch or create a mock interview plan" : "切换或创建模拟面试规划"}
            </div>
            <div style={{ fontSize: 13, color: T.ink3, marginTop: 6, lineHeight: 1.6 }}>
              {lang === "en" ? "Each plan keeps a JD, resume, and target round together. Switching only changes the current mock context." : "每个规划都绑定一份 JD、一份简历和目标轮次。切换只改变当前模拟面试上下文。"}
            </div>
          </div>
          <button onClick={onClose} style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", padding: 4 }}>
            <Icon name="x" size={16} />
          </button>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "1fr", gap: 10 }}>
          {plans.map((plan) => {
            const active = plan.jobId === draftJobId;
            return (
              <button
                key={plan.id}
                onClick={() => setDraftJobId(plan.jobId)}
                style={{
                  width: "100%",
                  textAlign: "left",
                  border: `1px solid ${active ? T.accent : T.rule}`,
                  background: active ? T.accentSoft : T.bgSoft,
                  borderRadius: 3,
                  padding: "14px 16px",
                  cursor: "pointer",
                  display: "grid",
                  gridTemplateColumns: "24px 1fr auto",
                  gap: 12,
                  alignItems: "start",
                  fontFamily: "var(--ei-sans)",
                }}
              >
                <span style={{ width: 20, height: 20, borderRadius: 10, border: `1px solid ${active ? T.accent : T.rule}`, background: active ? T.accent : T.bgCard, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", marginTop: 1 }}>
                  {active && <Icon name="check" size={12} stroke={2.2} />}
                </span>
                <span>
                  <span style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap" }}>
                    <span style={{ fontSize: 14.5, color: T.ink, fontWeight: 600 }}>{plan.title}</span>
                    <Tag tone={plan.statusTone} T={T}>{plan.status}</Tag>
                  </span>
                  <span style={{ display: "block", fontSize: 12.5, color: T.ink3, marginTop: 4 }}>{plan.meta}</span>
                  <span style={{ display: "block", fontSize: 13, color: T.ink2, marginTop: 8, lineHeight: 1.55 }}>
                    {lang === "en" ? "Round" : "目标轮次"} · {plan.round} · {plan.resumeName}
                  </span>
                </span>
                <span className="ei-mono" style={{ color: T.ink3, fontSize: 11, marginTop: 2 }}>{plan.updatedAt}</span>
              </button>
            );
          })}
        </div>

        <div style={{ border: `1px solid ${T.rule}`, background: T.bgSoft, borderRadius: 3, padding: 14, marginTop: 16 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{lang === "en" ? "SELECTED CONTEXT" : "已选上下文"}</div>
          <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.6 }}>
            {selected.title} · {selected.round} · {selected.resumeName}
          </div>
        </div>

        <div style={{ display: "flex", justifyContent: "space-between", gap: 10, marginTop: 22, flexWrap: "wrap" }}>
          <Btn T={T} variant="ghost" icon="plus" onClick={onCreate}>{lang === "en" ? "Create from new JD" : "从新 JD 创建规划"}</Btn>
          <div style={{ display: "flex", gap: 10 }}>
            <Btn T={T} variant="ghost" onClick={onClose}>{lang === "en" ? "Cancel" : "取消"}</Btn>
            <Btn T={T} variant="accent" iconRight="arrow_right" onClick={() => onConfirm(selected)}>{lang === "en" ? "Use this plan" : "使用这个规划"}</Btn>
          </div>
        </div>
      </div>
    </div>
  );
};

const getCurrentRoundIndex = (job, rounds) => {
  if (!rounds?.length) return 0;
  const next = job?.nextRound || "";
  const found = rounds.findIndex((round) => next.includes(round.name));
  if (found >= 0) return found;
  if (job?.status === "草稿" || next.includes("未安排")) return 0;
  return Math.min(1, rounds.length - 1);
};

const InterviewRoundRail = ({ T, lang, label, rounds, currentIndex, nextRound }) => {
  const stateLabel = (i) => {
    if (i < currentIndex) return lang === "en" ? "Done" : "已完成";
    if (i === currentIndex) return lang === "en" ? "Current" : "当前";
    return lang === "en" ? "Upcoming" : "未到";
  };
  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", gap: 16, alignItems: "baseline", marginBottom: 12 }}>
        <div className="ei-label" style={{ color: T.ink3 }}>{label}</div>
        <div style={{ fontSize: 12, color: T.ink3 }}>
          {lang === "en" ? "Schedule" : "面试安排"} · <span style={{ color: T.ink2 }}>{nextRound || (lang === "en" ? "Not scheduled" : "未安排")}</span>
        </div>
      </div>
      <div style={{ position: "relative" }}>
        <div style={{ position: "absolute", top: 13, left: 13, right: 13, height: 1, background: T.rule }} />
        <div style={{ display: "grid", gridTemplateColumns: `repeat(${rounds.length}, 1fr)`, alignItems: "start" }}>
        {rounds.map((round, i) => {
          const done = i < currentIndex;
          const current = i === currentIndex;
          return (
            <div key={round.name} style={{ position: "relative", display: "flex", flexDirection: "column", alignItems: i === 0 ? "flex-start" : i === rounds.length - 1 ? "flex-end" : "center", minHeight: 72 }}>
              <div style={{
                width: 26, height: 26, borderRadius: 13,
                border: `1px solid ${done ? T.ok : current ? T.accent : T.rule}`,
                background: done ? T.ok : current ? T.accent : T.bgCard,
                color: done || current ? "#fff" : T.ink3,
                display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1,
                boxShadow: current ? `0 0 0 4px ${T.accentSoft}` : "none",
              }}>
                {done ? <Icon name="check" size={13} stroke={2.2} /> : <span className="ei-mono" style={{ fontSize: 11 }}>{i + 1}</span>}
              </div>
              <div style={{ fontSize: 12.5, color: current ? T.ink : done ? T.ink2 : T.ink3, marginTop: 8, textAlign: i === 0 ? "left" : i === rounds.length - 1 ? "right" : "center", maxWidth: 140, fontWeight: current ? 600 : 400 }}>
                {round.name}
              </div>
              <div style={{ fontSize: 11, color: current ? T.accent : T.ink4, marginTop: 3, textAlign: i === 0 ? "left" : i === rounds.length - 1 ? "right" : "center", maxWidth: 140, lineHeight: 1.35 }}>
                {stateLabel(i)} · {round.focus}
              </div>
            </div>
          );
        })}
        </div>
      </div>
    </div>
  );
};

const BindingPill = ({ T, icon, label, title, meta, action, onClick }) => (
  <div style={{ padding: "14px 16px", background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2, display: "grid", gridTemplateColumns: "32px 1fr auto", gap: 12, alignItems: "center" }}>
    <div style={{ width: 32, height: 32, borderRadius: 16, background: T.bgCard, border: `1px solid ${T.rule}`, color: T.accent, display: "flex", alignItems: "center", justifyContent: "center" }}>
      <Icon name={icon} size={15} />
    </div>
    <div style={{ minWidth: 0 }}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 3 }}>{label}</div>
      <div style={{ fontSize: 14, color: T.ink, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{title}</div>
      <div style={{ fontSize: 12, color: T.ink3, marginTop: 2, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{meta}</div>
    </div>
    {action && (
      <button onClick={onClick} style={{ background: "transparent", border: `1px solid ${T.rule}`, borderRadius: 2, color: T.ink2, padding: "5px 10px", fontSize: 12, cursor: "pointer" }}>
        {action}
      </button>
    )}
  </div>
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

window.WorkspaceScreen = WorkspaceScreen;
Object.assign(window, { ResumePickerModal, BindingPill, getWorkspaceResumeOptions });
