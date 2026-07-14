// Screen 2: Interview plan list
const WorkspaceScreen = ({ T, lang, nav }) => {
  const jobs = window.EI_DATA?.targetJobs || [];
  return <WorkspacePlanList T={T} lang={lang} nav={nav} jobs={jobs} />;
};

const WorkspacePlanList = ({ T, lang, nav, jobs = [] }) => {
  const [hiddenJobIds, setHiddenJobIds] = React.useState([]);
  const L = lang === "en" ? {
    eyebrow: "INTERVIEW PLANS",
    title: "Choose an interview plan to continue.",
    subtitle: "Saved target JDs and interview plans live here. Click a card to review the plan detail, or start the interview directly.",
    create: "Import JD",
    emptyTitle: "No interview plans yet",
    emptyDesc: "Import a target JD from Home first, then continue the plan here.",
    start: "Start interview now",
    delete: "Delete",
  } : {
    eyebrow: "面试规划",
    title: "选择要继续的面试规划。",
    subtitle: "这里展示已保存的目标 JD / 面试规划。点击卡片进入规划详情，也可以直接开始面试。",
    create: "导入 JD",
    emptyTitle: "还没有面试规划",
    emptyDesc: "先从首页导入一个目标 JD，再回来继续面试规划。",
    start: "立即面试",
    delete: "删除",
  };
  const openPlan = (job) => nav("workspace", {
    targetJobId: job.id,
  });
  const startInterview = (job) => {
    const currentRound = window.eiResolvePracticeProgress(window.EI_DATA?.jdSample?.rounds || [], job.practiceProgress).currentRound;
    if (!currentRound) return;
    nav("practice", {
      targetJobId: job.id,
      roundId: currentRound.id,
      roundName: currentRound.name,
      sessionId: `session-${job.id}-${currentRound.id}-new`,
    });
  };
  const deletePlan = (job) => setHiddenJobIds((prev) => [...new Set([...prev, job.id])]);
  const visibleJobs = jobs.filter((job) => !hiddenJobIds.includes(job.id));
  return (
    <div data-testid="workspace-plan-list" className="ei-fadein" style={{ maxWidth: 1120, margin: "0 auto", padding: "48px 48px 96px" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 24, flexWrap: "wrap", marginBottom: 28 }}>
        <div style={{ maxWidth: 640 }}>
          <div data-testid="workspace-plan-list-eyebrow" className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{L.eyebrow}</div>
          <h1 data-testid="workspace-plan-list-title" className="ei-serif" style={{ fontSize: 40, color: T.ink, margin: 0, lineHeight: 1.14 }}>{L.title}</h1>
          <div data-testid="workspace-plan-list-subtitle" style={{ fontSize: 14, color: T.ink2, marginTop: 10, lineHeight: 1.6 }}>{L.subtitle}</div>
        </div>
        <Btn variant="primary" icon="plus" onClick={() => nav("home")} T={T}>{L.create}</Btn>
      </div>
      {visibleJobs.length === 0 ? (
        <div data-testid="workspace-plan-list-empty" style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 32, textAlign: "center" }}>
          <div className="ei-serif" style={{ fontSize: 18, color: T.ink, marginBottom: 10 }}>{L.emptyTitle}</div>
          <div style={{ fontSize: 13, color: T.ink3, lineHeight: 1.55 }}>{L.emptyDesc}</div>
        </div>
      ) : (
        <div data-testid="workspace-plan-list-grid" style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(300px, 360px))", justifyContent: "start", gap: 16, alignItems: "stretch" }}>
          {visibleJobs.map((job) => {
            const rounds = window.EI_DATA?.jdSample?.rounds || [];
            const progress = window.eiResolvePracticeProgress(rounds, job.practiceProgress);
            const currentRoundIndex = progress.currentIndex;
            const statusMap = {
              amber: { bg: T.amberSoft, fg: T.warn },
              neutral: { bg: T.bgSoft, fg: T.ink2 },
              muted: { bg: "transparent", fg: T.ink3 },
            };
            const s = statusMap[job.statusTone] || statusMap.amber;
            return (
              <article key={job.id} data-testid={`workspace-plan-list-card-${job.id}`} onClick={() => openPlan(job)} style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 20, cursor: "pointer", display: "flex", flexDirection: "column", gap: 14, position: "relative" }}>
                <button
                  aria-label={L.delete}
                  title={L.delete}
                  data-testid={`workspace-plan-list-delete-${job.id}`}
                  onClick={(event) => { event.stopPropagation(); deletePlan(job); }}
                  style={{ position: "absolute", top: 20, right: 20, zIndex: 1, display: "inline-flex", alignItems: "center", justifyContent: "center", width: 28, height: 28, color: T.ink3, background: "transparent", border: `1px solid ${T.rule}`, borderRadius: 2, cursor: "pointer" }}
                >
                  <Icon name="trash" size={13} />
                </button>
                <div data-testid={`workspace-plan-list-card-body-${job.id}`} style={{ display: "flex", justifyContent: "space-between", gap: 12, paddingRight: 48 }}>
                  <div>
                    <div style={{ fontSize: 11, fontFamily: "var(--ei-mono)", color: T.ink3, marginBottom: 4 }}>{String(job.company || "").toUpperCase()} · {job.status}</div>
                    <div className="ei-serif" style={{ fontSize: 19, color: T.ink, letterSpacing: "-0.01em" }}>{job.title}</div>
                    <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 4 }}>{job.location || "Location not set"}</div>
                  </div>
                  <div style={{ padding: "3px 8px", height: "fit-content", background: s.bg, color: s.fg, fontSize: 11, fontFamily: "var(--ei-mono)", borderRadius: 2, whiteSpace: "nowrap" }}>
                    {job.status}
                  </div>
                </div>
                <div data-testid={`workspace-plan-list-rail-${job.id}`}>
                  <WorkspaceMiniRoundRail T={T} rounds={rounds} currentIndex={currentRoundIndex} />
                </div>
                <div data-testid={`workspace-plan-list-card-footer-${job.id}`} style={{ borderTop: `1px solid ${T.rule}`, paddingTop: 14, background: T.bgCard, display: "flex", justifyContent: "flex-end", alignItems: "center", gap: 12 }}>
                  <button
                    disabled={!progress.currentRound}
                    onClick={(event) => { event.stopPropagation(); if (progress.currentRound) startInterview(job); }}
                    style={{ flex: "0 0 auto", height: 32, padding: "0 12px", fontSize: 13, fontWeight: 500, background: T.accent, color: "#fff", border: `1px solid ${T.accent}`, borderRadius: 2, cursor: progress.currentRound ? "pointer" : "not-allowed", opacity: progress.currentRound ? 1 : 0.58, fontFamily: "var(--ei-sans)" }}
                  >
                    {L.start}
                  </button>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </div>
  );
};

const WorkspaceMiniRoundRail = ({ T, rounds, currentIndex }) => (
  <div style={{ marginTop: 18 }}>
    <div style={{ position: "relative", height: 34 }}>
      <div style={{ position: "absolute", top: 9, left: 8, right: 8, height: 1, background: T.rule }} />
      <div style={{ display: "grid", gridTemplateColumns: `repeat(${rounds.length}, 1fr)` }}>
        {rounds.map((round, i) => {
          const done = currentIndex !== null && i < currentIndex;
          const current = i === currentIndex;
          return (
            <div key={`round-${round.sequence}-${round.type}`} data-round-state={done ? "done" : current ? "current" : "pending"} style={{ position: "relative", display: "flex", flexDirection: "column", alignItems: i === 0 ? "flex-start" : i === rounds.length - 1 ? "flex-end" : "center" }}>
              <div style={{
                width: 18, height: 18, borderRadius: 9,
                border: `1px solid ${done ? T.ok : current ? T.accent : T.rule}`,
                background: done ? T.ok : current ? T.accent : T.bgCard,
                color: done || current ? "#fff" : T.ink3,
                display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1,
              }}>
                {done ? <Icon name="check" size={10} stroke={2.2} /> : <span className="ei-mono" style={{ fontSize: 9 }}>{i + 1}</span>}
              </div>
              <div style={{ marginTop: 6, fontSize: 10.5, color: current ? T.ink : T.ink3, maxWidth: 68, textAlign: i === 0 ? "left" : i === rounds.length - 1 ? "right" : "center", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                {round.name}{round.durationMinutes ? ` · ${round.durationMinutes}m` : ""}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  </div>
);

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

window.WorkspaceScreen = WorkspaceScreen;
Object.assign(window, { getWorkspaceResumeOptions });
