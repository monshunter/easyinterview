// Screen 1: Home / JD 导入 / 目标岗位列表
const HomeScreen = ({ T, lang, nav, role }) => {
  const D = window.EI_DATA;
  const [input, setInput] = React.useState("");
  const [parsing, setParsing] = React.useState(false);

  const handleImport = () => {
    if (!input.trim()) return;
    setParsing(true);
    setTimeout(() => { setParsing(false); nav("parse", { source: input.length > 200 ? "pasted" : "link" }); }, 400);
  };

  const L = lang === "en" ? {
    tag: "INBOX · TARGET JOBS",
    title: "Let's win the interview you already care about.",
    sub: "Paste a JD, drop a link, or continue one of your open jobs. Practice stays tied to the role, not to generic question banks.",
    ph: "Paste the JD here, or a jobs.example.com URL…",
    importBtn: "Parse & open workspace",
    orUpload: "or upload .pdf / .docx / .md",
    active: "Active jobs",
    activeSub: "Sorted by last touch — your most urgent preparation is on top.",
    startAfter: "Just finished a real interview?",
    startAfterSub: "Drop in the questions you were asked and turn it into next-round ammo.",
    startAfterBtn: "Open real-interview debrief",
    newJob: "New target job",
    noJd: "No JD yet? Start a blank workspace →",
  } : {
    tag: "收件箱 · 目标岗位",
    title: "先把你已经拿在手里的那场面试，赢下来。",
    sub: "粘贴 JD、丢一个链接，或继续之前的目标岗位。每一次练习都绑定具体岗位，而不是泛用题库。",
    ph: "把 JD 粘贴到这里，或者丢一个岗位链接…",
    importBtn: "解析并进入工作台",
    orUpload: "也可以上传 .pdf / .docx / .md",
    active: "进行中的岗位",
    activeSub: "按最近动作排序——最紧要的准备永远在最上面。",
    startAfter: "刚面完一轮？",
    startAfterSub: "把真实问过的问题丢进来，沉淀成下一轮的弹药。",
    startAfterBtn: "打开真实面试复盘",
    newJob: "新建目标岗位",
    noJd: "还没有 JD？先开一个空白工作台 →",
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
              <button style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, display: "flex", alignItems: "center", gap: 6, padding: 0 }}>
                <Icon name="upload" size={14} /> {L.orUpload}
              </button>
              <span style={{ color: T.rule }}>·</span>
              <button style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, display: "flex", alignItems: "center", gap: 6, padding: 0 }}>
                <Icon name="link" size={14} /> URL
              </button>
            </div>
            <Btn variant="accent" onClick={handleImport} T={T} iconRight="arrow_right" disabled={!input.trim() && !parsing}>
              {parsing ? (lang === "en" ? "Parsing JD…" : "正在解析 JD…") : L.importBtn}
            </Btn>
          </div>
        </div>

        <div style={{ marginTop: 16, display: "flex", gap: 16, flexWrap: "wrap" }}>
          <button onClick={() => nav("onboarding")} style={{ background: "transparent", border: "none", color: T.accent, fontSize: 13, padding: 0, cursor: "pointer", fontWeight: 500 }}>
            {lang === "en" ? "First time? Do the 5-min setup →" : "第一次？5 分钟画像引导 →"}
          </button>
          <button onClick={() => nav("workspace", { jobId: "tj-3" })} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, padding: 0, cursor: "pointer" }}>{L.noJd}</button>
          <button onClick={() => nav("jd_match")} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, padding: 0, cursor: "pointer", display: "flex", alignItems: "center", gap: 6 }}>
            <Icon name="search" size={12} /> {lang === "en" ? "Or let us match JDs to your profile →" : "或让我们按画像为你匹配岗位 →"}
          </button>
        </div>
      </div>

      {/* Active jobs */}
      <div style={{ marginBottom: 48 }}>
        <SectionHeader eyebrow={lang === "en" ? "ACTIVE" : "进行中"} title={L.active} sub={L.activeSub} T={T}
          right={<Btn variant="secondary" size="sm" icon="plus" T={T}>{L.newJob}</Btn>} />
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(320px, 1fr))", gap: 16 }}>
          {D.targetJobs.map((j) => <JobCard key={j.id} job={j} T={T} onClick={() => nav("workspace", { jobId: j.id })} lang={lang} />)}
        </div>
      </div>

      {/* Post-interview prompt */}
      <div style={{
        background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 24,
        display: "flex", justifyContent: "space-between", alignItems: "center", gap: 24, flexWrap: "wrap"
      }}>
        <div style={{ flex: 1, minWidth: 260 }}>
          <div className="ei-label" style={{ color: T.accent, marginBottom: 6 }}>POST-INTERVIEW</div>
          <div className="ei-serif" style={{ fontSize: 20, color: T.ink }}>{L.startAfter}</div>
          <div style={{ fontSize: 13.5, color: T.ink2, marginTop: 4 }}>{L.startAfterSub}</div>
        </div>
        <Btn variant="secondary" icon="flag" onClick={() => nav("debrief")} T={T} iconRight="arrow_right">{L.startAfterBtn}</Btn>
      </div>
    </div>
  );
};

const JobCard = ({ job, T, onClick, lang }) => {
  const statusMap = {
    amber: { bg: T.amberSoft, fg: T.warn },
    neutral: { bg: T.bgSoft, fg: T.ink2 },
    muted: { bg: "transparent", fg: T.ink3 },
  };
  const s = statusMap[job.statusTone];
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

      <div style={{ borderTop: `1px dotted ${T.rule}`, paddingTop: 14, display: "flex", gap: 16, alignItems: "center" }}>
        <ReadinessDial level={job.readiness} T={T} size={44} />
        <div style={{ fontSize: 12, color: T.ink3, lineHeight: 1.5 }}>
          <div style={{ color: T.ink2, fontWeight: 500, fontSize: 13 }}>{job.readinessLabel}</div>
          <div>{job.practices} {lang === "en" ? "sessions" : "次练习"} · {job.mistakes} {lang === "en" ? "mistakes" : "道错题"}</div>
        </div>
      </div>

      {/* Status machine — compact pipeline pill row */}
      <div onClick={(e) => e.stopPropagation()}>
        <window.JobStatusPipeline T={T} lang={lang}
          currentKey={job.statusKey || (job.status?.includes("面试") || job.status?.toLowerCase().includes("interview") ? "interviewing" : job.status?.includes("准备") || job.status?.toLowerCase().includes("prepar") ? "preparing" : "draft")}
          compact />
      </div>

      <div style={{ fontSize: 12, color: T.ink2, background: T.bgSoft, padding: "8px 10px", borderRadius: 2 }}>
        <span style={{ color: T.ink3 }}>{lang === "en" ? "Next" : "下一步"} · </span>{job.nextRound}
      </div>
    </div>
  );
};

window.HomeScreen = HomeScreen;
