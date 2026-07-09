// Screen 1: Home / JD 导入 / 最近模拟面试
const HomeScreen = ({ T, lang, nav, role, signedIn = false }) => {
  const D = window.EI_DATA;
  const [input, setInput] = React.useState("");
  const [selectedResumeId, setSelectedResumeId] = React.useState("");
  const [parsing, setParsing] = React.useState(false);
  const [assistOpen, setAssistOpen] = React.useState(null);
  const recentJobs = D.targetJobs || [];
  const recentPreviewJobs = recentJobs.slice(0, 3);
  const hasMoreRecentJobs = recentJobs.length > recentPreviewJobs.length;
  const resumeOptions = window.getWorkspaceResumeOptions ? window.getWorkspaceResumeOptions(lang) : [];
  const selectedResume = resumeOptions.find((resume) => resume.id === selectedResumeId);

  const handleImport = () => {
    if (!input.trim() || !selectedResume) return;
    setParsing(true);
    setTimeout(() => { setParsing(false); nav("parse", { source: "pasted", resumeId: selectedResume.id }); }, 400);
  };

  const L = lang === "en" ? {
    tag: "HOME · MOCK INTERVIEWS",
    title: "Let's win the interview you already care about.",
    ph: "Paste the JD here…",
    pasteSource: "Paste JD",
    uploadSource: "Upload JD file",
    uploadSourceSub: "Supports .pdf / .docx / .md, or import from URL.",
    importBtn: "Start interview now",
    orUpload: "or upload .pdf / .docx / .md",
    active: "Recent mock interviews",
    activeSub: "Sorted by recent preparation. Each card is tied to one target job and interview round.",
    recentMore: "More",
    resumeSelect: "Select existing resume",
    resumeSelectPlaceholder: "Select a resume",
    resumeSelectHint: "Pick the resume this mock interview should use.",
    resumeEmpty: "No ready resume yet",
    selectedResume: "Selected resume",
    resumeCreate: "No resume yet? Create one in 1 minute →",
  } : {
    tag: "首页 · 模拟面试",
    title: "先把你已经拿在手里的那场面试，赢下来。",
    ph: "把 JD 粘贴到这里…",
    pasteSource: "粘贴 JD",
    uploadSource: "上传 JD 文件",
    uploadSourceSub: "支持 .pdf / .docx / .md，也可以从 URL 导入。",
    importBtn: "立即面试",
    orUpload: "也可以上传 .pdf / .docx / .md",
    active: "最近模拟面试",
    activeSub: "按最近准备排序。每张卡片都对应一个目标岗位和一轮面试。",
    recentMore: "更多",
    resumeSelect: "选择已有简历",
    resumeSelectPlaceholder: "请选择简历",
    resumeSelectHint: "选择这场模拟面试要使用的简历。",
    resumeEmpty: "还没有可用简历",
    selectedResume: "已选择简历",
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

        <div style={{ marginTop: 32 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{L.pasteSource}</div>
          <div data-testid="home-jd-input-card" style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3, padding: 20 }}>
            <textarea
              data-testid="home-jd-textarea"
              value={input} onChange={(e) => setInput(e.target.value)}
              placeholder={L.ph}
              style={{
                width: "100%", minHeight: 120, border: "none", outline: "none", resize: "vertical",
                fontSize: 14.5, lineHeight: 1.6, color: T.ink, background: "transparent",
                fontFamily: "var(--ei-sans)",
              }}
            />
            <div data-testid="home-jd-source-controls" style={{ marginTop: 14, paddingTop: 14, borderTop: `1px solid ${T.rule}`, display: "flex", alignItems: "center", justifyContent: "space-between", gap: 12, flexWrap: "wrap" }}>
              <div style={{ color: T.ink3, fontSize: 12.5, lineHeight: 1.5 }}>{L.orUpload}</div>
              <div style={{ display: "flex", alignItems: "center", gap: 10, flexWrap: "wrap" }}>
                <button data-testid="home-upload-trigger" onClick={() => setAssistOpen("upload")} style={{ background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 3, color: T.ink, fontSize: 13, display: "flex", alignItems: "center", justifyContent: "center", minHeight: 34, padding: "0 12px", cursor: "pointer", fontWeight: 500 }}>
                  <Icon name="upload" size={14} /> {L.uploadSource}
                </button>
                <button data-testid="home-url-trigger" onClick={() => setAssistOpen("url")} style={{ background: "transparent", border: "1px solid transparent", color: T.accent, fontSize: 13, display: "flex", alignItems: "center", justifyContent: "center", minHeight: 34, padding: "0 12px", cursor: "pointer", fontWeight: 500 }}>
                  <Icon name="link" size={14} /> URL
                </button>
              </div>
            </div>
          </div>
        </div>

        <div style={{ marginTop: 16 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{L.resumeSelect}</div>
          <div data-testid="home-resume-row" style={{ display: "flex", alignItems: "center", gap: 14, flexWrap: "wrap" }}>
            <select
              data-testid="home-resume-select"
              value={selectedResumeId}
              disabled={!resumeOptions.length}
              onChange={(e) => setSelectedResumeId(e.target.value)}
              style={{ width: 360, maxWidth: "100%", flex: "0 1 360px", boxSizing: "border-box", minHeight: 42, border: `1px solid ${T.rule}`, borderRadius: 3, background: T.bgCard, color: T.ink, fontSize: 13.5, fontFamily: "var(--ei-sans)", padding: "0 12px", outline: "none", cursor: resumeOptions.length ? "pointer" : "not-allowed" }}
            >
              <option value="">{L.resumeSelectPlaceholder}</option>
              {resumeOptions.map((resume) => (
                <option key={resume.id} value={resume.id}>{resume.name} · {resume.meta}</option>
              ))}
            </select>
            <button data-testid="home-resume-create" onClick={() => nav("resume_versions", { flow: "create" })} style={{ background: "transparent", border: "none", color: T.accent, fontSize: 13, padding: 0, cursor: "pointer", fontWeight: 500, minHeight: 42, display: "flex", alignItems: "center" }}>
              {L.resumeCreate}
            </button>
          </div>
          {!resumeOptions.length && (
            <div style={{ marginTop: 8, maxWidth: 360, border: `1px dashed ${T.rule}`, borderRadius: 3, padding: "10px 12px", color: T.ink3, fontSize: 13 }}>{L.resumeEmpty}</div>
          )}
          <div style={{ marginTop: 8, fontSize: 12.5, color: selectedResume ? T.ink2 : T.ink3 }}>
            {selectedResume ? `${L.selectedResume} · ${selectedResume.name}` : L.resumeSelectHint}
          </div>
          <div data-testid="home-submit-row" style={{ marginTop: 14, display: "flex" }}>
            <Btn variant="accent" onClick={handleImport} T={T} iconRight="arrow_right" disabled={!input.trim() || !selectedResume || parsing}>
              {parsing ? (lang === "en" ? "Parsing JD…" : "正在解析 JD…") : L.importBtn}
            </Btn>
          </div>
        </div>
      </div>

      {/* Recent mock interviews */}
      {signedIn && (
        <div style={{ marginBottom: 48 }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", gap: 20, marginBottom: 16 }}>
            <SectionHeader eyebrow={lang === "en" ? "RECENT" : "最近"} title={L.active} sub={L.activeSub} T={T} />
            {hasMoreRecentJobs && (
              <button onClick={() => nav("workspace", {})} style={{ background: "transparent", border: "none", color: T.accent, fontSize: 13, padding: 0, cursor: "pointer", fontWeight: 500, whiteSpace: "nowrap" }}>
                {L.recentMore}
              </button>
            )}
          </div>
          {recentJobs.length ? (
            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(320px, 1fr))", gap: 16 }}>
              {recentPreviewJobs.map((j) => <MockInterviewCard key={j.id} job={j} rounds={D.jdSample.rounds} T={T} onClick={() => nav("workspace", { targetJobId: j.id, jobId: j.id, planId: `plan-${j.id}`, jdId: `jd-${j.id}` })} lang={lang} />)}
            </div>
          ) : (
            <HomeEmptyState T={T} lang={lang} onImport={() => document.querySelector("textarea")?.focus()} />
          )}
        </div>
      )}

      {assistOpen && <JDAssistModal T={T} lang={lang} type={assistOpen} onClose={() => setAssistOpen(null)} onConfirm={() => { if (!selectedResume) return; setAssistOpen(null); nav("parse", { source: assistOpen, resumeId: selectedResume.id }); }} />}
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
              {lang === "en" ? "The system will fetch the JD, then open the interview plan detail for confirmation." : "系统会读取 JD 内容，然后进入面试规划详情确认。"}
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
