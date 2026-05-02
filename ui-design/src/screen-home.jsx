// Screen 1: Home / JD 导入 / 最近模拟面试
const HomeScreen = ({ T, lang, nav, role }) => {
  const D = window.EI_DATA;
  const [input, setInput] = React.useState("");
  const [parsing, setParsing] = React.useState(false);
  const [assistOpen, setAssistOpen] = React.useState(null);
  const recentJobs = D.targetJobs || [];

  const handleImport = () => {
    if (!input.trim()) return;
    setParsing(true);
    setTimeout(() => { setParsing(false); nav("parse", { source: "pasted" }); }, 400);
  };

  const L = lang === "en" ? {
    tag: "HOME · MOCK INTERVIEWS",
    title: "Let's win the interview you already care about.",
    sub: "Paste a JD or continue a recent mock interview. Practice stays tied to the role, not to generic question banks.",
    ph: "Paste the JD here…",
    importBtn: "Parse & confirm interview",
    orUpload: "or upload .pdf / .docx / .md",
    active: "Recent mock interviews",
    activeSub: "Sorted by recent preparation. Each card is tied to one target job and interview round.",
    startAfter: "Just finished an interview?",
    startAfterSub: "Drop in the questions you were asked and turn it into next-round ammo.",
    startAfterBtn: "Open debrief",
    jobPicks: "Match more JDs from your resume",
    jobPicksSub: "Use your resume and role preference to find JDs worth preparing for, then confirm a mock interview from the match list.",
    jobPicksBtn: "Open job recommendations",
    resumeCreate: "No resume yet? Create one in 1 minute →",
  } : {
    tag: "首页 · 模拟面试",
    title: "先把你已经拿在手里的那场面试，赢下来。",
    sub: "粘贴 JD，或继续最近一次模拟面试。每一次练习都绑定具体岗位，而不是泛用题库。",
    ph: "把 JD 粘贴到这里…",
    importBtn: "解析并确认面试",
    orUpload: "也可以上传 .pdf / .docx / .md",
    active: "最近模拟面试",
    activeSub: "按最近准备排序。每张卡片都对应一个目标岗位和一轮面试。",
    startAfter: "刚面完一轮？",
    startAfterSub: "把面试问过的问题丢进来，沉淀成下一轮的弹药。",
    startAfterBtn: "打开复盘",
    jobPicks: "按简历匹配更多 JD",
    jobPicksSub: "用你的简历和岗位偏好筛出值得准备的 JD，再从匹配结果进入模拟面试前确认。",
    jobPicksBtn: "打开岗位推荐",
    resumeCreate: "还没有简历？1 分钟创建 →",
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 1160, margin: "0 auto", padding: "48px 56px 96px" }}>
      {/* Hero / import */}
      <div style={{ marginBottom: 56 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{L.tag}</div>
        <h1 className="ei-serif" style={{ fontSize: 48, color: T.ink, margin: 0, lineHeight: 1.1, letterSpacing: "-0.025em", maxWidth: 820, textWrap: "balance" }}>
          {L.title}
        </h1>
        <p style={{ fontSize: 15.5, color: T.ink2, maxWidth: 620, marginTop: 16, lineHeight: 1.55 }}>{L.sub}</p>

        <div style={{ marginTop: 32, background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 20 }}>
          <textarea
            value={input} onChange={(e) => setInput(e.target.value)}
            placeholder={L.ph}
            style={{
              width: "100%", minHeight: 120, border: "none", outline: "none", resize: "vertical",
              fontSize: 14.5, lineHeight: 1.6, color: T.ink, background: "transparent",
              fontFamily: "var(--ei-sans)",
            }}
          />
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: 10, paddingTop: 14, borderTop: `1px dotted ${T.rule}` }}>
            <div style={{ display: "flex", gap: 12, alignItems: "center" }}>
              <button onClick={() => setAssistOpen("upload")} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, display: "flex", alignItems: "center", gap: 6, padding: 0, cursor: "pointer" }}>
                <Icon name="upload" size={14} /> {L.orUpload}
              </button>
              <span style={{ color: T.rule }}>·</span>
              <button onClick={() => setAssistOpen("url")} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, display: "flex", alignItems: "center", gap: 6, padding: 0, cursor: "pointer" }}>
                <Icon name="link" size={14} /> URL
              </button>
            </div>
            <Btn variant="accent" onClick={handleImport} T={T} iconRight="arrow_right" disabled={!input.trim() && !parsing}>
              {parsing ? (lang === "en" ? "Parsing JD…" : "正在解析 JD…") : L.importBtn}
            </Btn>
          </div>
        </div>

        <div style={{ marginTop: 16, display: "flex", gap: 16, flexWrap: "wrap" }}>
          <button onClick={() => nav("resume_versions", { flow: "create" })} style={{ background: "transparent", border: "none", color: T.accent, fontSize: 13, padding: 0, cursor: "pointer", fontWeight: 500 }}>
            {L.resumeCreate}
          </button>
        </div>
      </div>

      {/* Recent mock interviews */}
      <div style={{ marginBottom: 48 }}>
        <SectionHeader eyebrow={lang === "en" ? "RECENT" : "最近"} title={L.active} sub={L.activeSub} T={T} />
        {recentJobs.length ? (
          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(320px, 1fr))", gap: 16 }}>
            {recentJobs.map((j) => <MockInterviewCard key={j.id} job={j} rounds={D.jdSample.rounds} T={T} onClick={() => nav("workspace", { targetJobId: j.id, jobId: j.id, planId: `plan-${j.id}`, jdId: `jd-${j.id}` })} lang={lang} />)}
          </div>
        ) : (
          <HomeEmptyState T={T} lang={lang} onImport={() => document.querySelector("textarea")?.focus()} />
        )}
      </div>

      {/* Auxiliary starts */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
        <div style={{
          background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 24,
          display: "flex", justifyContent: "space-between", alignItems: "center", gap: 20, flexWrap: "wrap"
        }}>
          <div style={{ flex: 1, minWidth: 260 }}>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 6 }}>JOB PICKS</div>
            <div className="ei-serif" style={{ fontSize: 20, color: T.ink }}>{L.jobPicks}</div>
            <div style={{ fontSize: 13.5, color: T.ink2, marginTop: 4, lineHeight: 1.55 }}>{L.jobPicksSub}</div>
          </div>
          <Btn variant="secondary" icon="search" onClick={() => nav("jd_match")} T={T} iconRight="arrow_right">{L.jobPicksBtn}</Btn>
        </div>
        <div style={{
          background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 24,
          display: "flex", justifyContent: "space-between", alignItems: "center", gap: 20, flexWrap: "wrap"
        }}>
          <div style={{ flex: 1, minWidth: 260 }}>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 6 }}>POST-INTERVIEW</div>
            <div className="ei-serif" style={{ fontSize: 20, color: T.ink }}>{L.startAfter}</div>
            <div style={{ fontSize: 13.5, color: T.ink2, marginTop: 4 }}>{L.startAfterSub}</div>
          </div>
          <Btn variant="secondary" icon="flag" onClick={() => nav("debrief")} T={T} iconRight="arrow_right">{L.startAfterBtn}</Btn>
        </div>
      </div>

      {assistOpen && <JDAssistModal T={T} lang={lang} type={assistOpen} onClose={() => setAssistOpen(null)} onConfirm={() => { setAssistOpen(null); nav("parse", { source: assistOpen }); }} />}
    </div>
  );
};

const HomeEmptyState = ({ T, lang, onImport }) => (
  <div style={{ padding: 24, border: `1px dashed ${T.rule}`, borderRadius: 3, background: T.bgSoft }}>
    <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "NO RECENT MOCKS" : "暂无最近模拟面试"}</div>
    <div className="ei-serif" style={{ fontSize: 20, color: T.ink, marginBottom: 8 }}>
      {lang === "en" ? "Start from a JD instead of showing placeholder interviews." : "从一份 JD 开始，不展示占位面试数据。"}
    </div>
    <div style={{ fontSize: 13.5, color: T.ink3, lineHeight: 1.6, marginBottom: 14 }}>
      {lang === "en" ? "Paste or upload a target job description to create the first interview context." : "粘贴或上传目标岗位 JD 后，系统会生成第一条面试上下文。"}
    </div>
    <Btn T={T} variant="secondary" icon="arrow_left" onClick={onImport}>{lang === "en" ? "Go to JD input" : "回到 JD 输入"}</Btn>
  </div>
);

const MockInterviewCard = ({ job, rounds, T, onClick, lang }) => {
  const statusMap = {
    amber: { bg: T.amberSoft, fg: T.warn },
    neutral: { bg: T.bgSoft, fg: T.ink2 },
    muted: { bg: "transparent", fg: T.ink3 },
  };
  const s = statusMap[job.statusTone];
  const currentRoundIndex = getHomeRoundIndex(job, rounds);
  return (
    <div onClick={onClick} style={{
      background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 20,
      cursor: "pointer", transition: "border-color .15s, transform .15s", display: "flex", flexDirection: "column", gap: 14,
    }}
      onMouseEnter={(e) => e.currentTarget.style.borderColor = T.ink3}
      onMouseLeave={(e) => e.currentTarget.style.borderColor = T.rule}
    >
      <div style={{ display: "flex", justifyContent: "space-between", gap: 12 }}>
        <div>
          <div style={{ fontSize: 11, fontFamily: "var(--ei-mono)", color: T.ink3, marginBottom: 4 }}>{job.company.toUpperCase()} · {job.level}</div>
          <div className="ei-serif" style={{ fontSize: 19, color: T.ink, letterSpacing: "-0.01em" }}>{job.title}</div>
          <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 4 }}>{job.location}</div>
        </div>
        <div style={{ padding: "3px 8px", height: "fit-content", background: s.bg, color: s.fg, fontSize: 11, fontFamily: "var(--ei-mono)", borderRadius: 2, whiteSpace: "nowrap" }}>
          {job.status}
        </div>
      </div>
      <MiniRoundRail T={T} lang={lang} rounds={rounds} currentIndex={currentRoundIndex} />
    </div>
  );
};

const getHomeRoundIndex = (job, rounds) => {
  if (!rounds?.length) return 0;
  const next = job?.nextRound || "";
  const found = rounds.findIndex((round) => next.includes(round.name));
  if (found >= 0) return found;
  if (job?.status === "草稿" || next.includes("未安排")) return 0;
  return Math.min(1, rounds.length - 1);
};

const MiniRoundRail = ({ T, lang, rounds, currentIndex }) => (
  <div style={{ marginTop: 18 }}>
    <div style={{ position: "relative", height: 34 }}>
      <div style={{ position: "absolute", top: 9, left: 8, right: 8, height: 1, background: T.rule }} />
      <div style={{ display: "grid", gridTemplateColumns: `repeat(${rounds.length}, 1fr)` }}>
        {rounds.map((round, i) => {
          const done = i < currentIndex;
          const current = i === currentIndex;
          return (
            <div key={round.name} style={{ position: "relative", display: "flex", flexDirection: "column", alignItems: i === 0 ? "flex-start" : i === rounds.length - 1 ? "flex-end" : "center" }}>
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
                {round.name}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  </div>
);

const JDAssistModal = ({ T, lang, type, onClose, onConfirm }) => {
  const isUpload = type === "upload";
  return (
    <div style={{ position: "fixed", inset: 0, background: "rgba(24, 20, 16, 0.24)", zIndex: 80, display: "flex", alignItems: "center", justifyContent: "center", padding: 24 }} onClick={onClose}>
      <div className="ei-fadein" onClick={(e) => e.stopPropagation()} style={{ width: "min(520px, 100%)", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4, boxShadow: "0 24px 70px rgba(30, 22, 15, 0.24)", padding: 24 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 18, marginBottom: 18 }}>
          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{lang === "en" ? "JD INPUT" : "JD 输入"}</div>
            <div className="ei-serif" style={{ fontSize: 23, color: T.ink }}>
              {isUpload ? (lang === "en" ? "Upload a JD file" : "上传 JD 文件") : (lang === "en" ? "Import from URL" : "从 URL 导入")}
            </div>
          </div>
          <button onClick={onClose} style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", padding: 4 }}>
            <Icon name="x" size={16} />
          </button>
        </div>

        {isUpload ? (
          <div style={{ border: `1px dashed ${T.rule}`, background: T.bgSoft, borderRadius: 3, padding: "30px 22px", textAlign: "center" }}>
            <Icon name="upload" size={24} color={T.accent} />
            <div style={{ fontSize: 15, color: T.ink, marginTop: 12, fontWeight: 500 }}>
              {lang === "en" ? "Drop a .pdf, .docx, or .md JD file here" : "拖入 .pdf / .docx / .md 格式的 JD 文件"}
            </div>
            <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 6 }}>
              {lang === "en" ? "Prototype state: file picker is represented here." : "静态稿中用这个弹窗表示文件选择流程。"}
            </div>
          </div>
        ) : (
          <div>
            <label className="ei-label" style={{ display: "block", color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "JD URL" : "JD 链接"}</label>
            <input placeholder={lang === "en" ? "https://company.com/careers/job..." : "https://company.com/careers/job..."} style={{ width: "100%", boxSizing: "border-box", border: `1px solid ${T.rule}`, background: T.bgSoft, color: T.ink, borderRadius: 3, padding: "12px 14px", fontSize: 14, outline: "none", fontFamily: "var(--ei-sans)" }} />
            <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 8 }}>
              {lang === "en" ? "The system will fetch the JD, then ask you to confirm the parsed content." : "系统会读取 JD 内容，然后进入解析结果确认。"}
            </div>
          </div>
        )}

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 10, marginTop: 22 }}>
          <Btn T={T} variant="ghost" onClick={onClose}>{lang === "en" ? "Cancel" : "取消"}</Btn>
          <Btn T={T} variant="accent" iconRight="arrow_right" onClick={onConfirm}>{lang === "en" ? "Continue" : "继续解析"}</Btn>
        </div>
      </div>
    </div>
  );
};

window.HomeScreen = HomeScreen;
