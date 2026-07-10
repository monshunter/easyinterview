// Resume workshop — flat IA (D-20): a plain list of resume assets.
// Flat resume list. Each resume keeps its source-derived content and opens to
// a read-only detail view that renders PDF sources as a vertical page stack
// and paste / Markdown / text sources with Markdown.

const buildResumeData = (lang) => {
  const isEn = lang === "en";
  return [
    {
      id: "frontend-v3",
      name: isEn ? "Liu Zhe · Frontend Platform v3" : "刘哲 · 前端平台版 v3",
      langTag: isEn ? "CN" : "中",
      sourceType: "upload",
      sourceName: isEn ? "LiuZhe_Frontend_CN_2026.pdf" : "刘哲_前端_中文_2026.pdf",
      createdAt: "2026-04-18",
      updatedAt: "2026-04-20",
      summary: isEn ? "Frontend platform · checkout · design system" : "前端平台 · 结账 · 设计系统",
      parseStatus: "ready",
      text: [
        isEn ? "# Liu Zhe · Senior Frontend Engineer · Shanghai" : "# 刘哲 · 资深前端工程师 · 上海",
        isEn ? "## Star-Ring · Senior Frontend · 2022-now" : "## 星环科技 · 资深前端 · 2022 至今",
        isEn ? "- Worked on checkout performance and complex admin surfaces." : "- 负责结账流程性能改进和复杂后台系统建设。",
        isEn ? "- Built shared UI components for internal products." : "- 为内部产品建设通用 UI 组件。",
      ],
    },
    {
      id: "english-v1",
      name: "Liu Zhe · Frontend Platform EN v1",
      langTag: "EN",
      sourceType: "upload",
      sourceName: "LiuZhe_Frontend_EN_2026.pdf",
      createdAt: "2026-04-15",
      updatedAt: "2026-04-17",
      summary: isEn ? "Frontend platform · TypeScript · APAC/US" : "前端平台 · TypeScript · 跨时区协作",
      parseStatus: "ready",
      text: [
        "# Liu Zhe · Frontend Platform Engineer",
        "## Platform work",
        "- Built platform tooling, design system infrastructure, and TypeScript foundations.",
        "- Worked with distributed teams across APAC and US time zones.",
      ],
    },
    {
      id: "impact-v2",
      name: isEn ? "Liu Zhe · Collaboration Impact v2" : "刘哲 · 协作影响力版 v2",
      langTag: isEn ? "CN" : "中",
      sourceType: "paste",
      sourceName: isEn ? "Pasted text" : "粘贴文本",
      createdAt: "2026-04-18",
      updatedAt: "2026-04-18",
      summary: isEn ? "Cross-team influence · DS rollout · mentoring" : "跨团队推动 · 设计系统落地 · 带教",
      parseStatus: "ready",
      text: [
        isEn ? "# Liu Zhe · Senior Frontend Engineer" : "# 刘哲 · 资深前端工程师",
        isEn ? "- Focus: collaboration impact and rollout stories." : "- 侧重协作影响力与落地故事。",
      ],
    },
  ];
};

const resumeToday = "2026-06-12";

const resumeNotify = (lang, zh, en, opts = {}) => {
  if (window.eiToast) {
    window.eiToast(lang === "en" ? en : zh, { tone: opts.tone || "ok", duration: opts.duration || 2400 });
  }
};

// ─────────────────────────────────────────────────────────────────────
// Top-level screen: branches between list, create flow, and detail
// ─────────────────────────────────────────────────────────────────────
const ResumeWorkshopScreen = ({ T, lang, nav, params = {} }) => {
  const baseData = React.useMemo(() => buildResumeData(lang), [lang]);
  const [createdResumes, setCreatedResumes] = React.useState([]);
  const [deletedIds, setDeletedIds] = React.useState([]);
  const resumes = React.useMemo(
    () => [...baseData, ...createdResumes].filter((resume) => !deletedIds.includes(resume.id)),
    [baseData, createdResumes, deletedIds],
  );

  const [flow, setFlow] = React.useState(params.flow === "create" ? "create" : "list");
  React.useEffect(() => {
    setFlow(params.flow === "create" ? "create" : "list");
  }, [params.flow]);

  const isEn = lang === "en";

  const addCreatedResume = (sourceLabel, rawText = "") => {
    const suffix = Date.now();
    const pendingName = isEn ? "Generating name" : "名称生成中";
    const lines = rawText
      .split(/\n+/)
      .map((line) => line.trim())
      .filter(Boolean);
    const resume = {
      id: `resume-created-${suffix}`,
      name: pendingName,
      langTag: isEn ? "CN" : "中",
      sourceType: rawText ? "paste" : "upload",
      sourceName: rawText ? (isEn ? "Pasted text" : "粘贴文本") : (sourceLabel || (isEn ? "Uploaded file" : "上传文件")),
      createdAt: resumeToday,
      updatedAt: resumeToday,
      summary: isEn ? "Parsing resume detail" : "正在解析简历详情",
      parseStatus: "processing",
      text: [],
    };
    setCreatedResumes((prev) => [...prev, resume]);
    setFlow("list");
    nav("resume_versions", { resumeId: resume.id });
    resumeNotify(lang, "已创建简历 · 正在解析", "Resume created · parsing");
    window.setTimeout(() => {
      setCreatedResumes((prev) => prev.map((item) => {
        if (item.id !== resume.id) return item;
        const readyLines = lines.length > 0 ? lines : [
          isEn ? "# Uploaded resume" : "# 上传的简历",
          isEn ? "- Extracted resume content will render here as Markdown." : "- 文件提取后的简历正文会以 Markdown 渲染。",
        ];
        return {
          ...item,
          parseStatus: "ready",
          name: isEn ? "Parsed resume · Markdown ready" : "已解析简历 · Markdown 就绪",
          summary: isEn ? "Markdown snapshot ready" : "Markdown 正文已就绪",
          text: readyLines,
        };
      }));
    }, 1200);
  };

  const deleteResume = (resumeId) => {
    setDeletedIds((prev) => [...new Set([...prev, resumeId])]);
    resumeNotify(lang, "已删除简历", "Resume deleted");
  };

  const resumeId = params.resumeId;
  const detailResume = resumeId ? resumes.find((r) => r.id === resumeId) : null;

  if (flow === "create") {
    return <ResumeCreateFlow T={T} lang={lang} onBack={() => setFlow("list")} onCreateResume={addCreatedResume} />;
  }

  if (detailResume) {
    return (
      <ResumeDetailView
        T={T}
        lang={lang}
        nav={nav}
        resume={detailResume}
      />
    );
  }

  return (
    <ResumeListView
      T={T}
      lang={lang}
      nav={nav}
      resumes={resumes}
      onCreate={() => setFlow("create")}
      onDelete={deleteResume}
    />
  );
};

// ─────────────────────────────────────────────────────────────────────
// LIST VIEW — one flat table of resume assets
// ─────────────────────────────────────────────────────────────────────
const ResumeListView = ({ T, lang, nav, resumes, onCreate, onDelete }) => {
  const isEn = lang === "en";
  const sorted = [...resumes].sort((a, b) => (b.updatedAt || "").localeCompare(a.updatedAt || ""));
  const openResume = (r) => nav("resume_versions", { resumeId: r.id });

  return (
    <div className="ei-fadein" style={{ maxWidth: 1320, margin: "0 auto", padding: "40px 48px 96px" }}>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginBottom: 28, gap: 32, flexWrap: "wrap" }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
            {isEn ? "RESUME WORKSHOP" : "简历工坊"}
          </div>
          <h1 className="ei-serif" style={{ fontSize: 38, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.15 }}>
            {isEn ? "Your resumes." : "你的简历"}
          </h1>
          <div style={{ fontSize: 14, color: T.ink3, marginTop: 10, maxWidth: 720, lineHeight: 1.55 }}>
            {isEn
              ? "A flat list of resume assets. Open any resume to read the resume content itself."
              : "简历按平铺列表管理。打开任意一份看到的就是简历内容本身。"}
          </div>
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          <Btn T={T} variant="accent" size="sm" icon="plus" onClick={onCreate}>{isEn ? "New resume" : "新建简历"}</Btn>
        </div>
      </div>

      {/* Flat table */}
      <div style={{ border: `1px solid ${T.rule}`, borderRadius: 3, background: T.bgCard, overflow: "hidden" }}>
        <div style={{
          display: "grid",
          gridTemplateColumns: "1.8fr 1.4fr 0.6fr 1fr 132px",
          gap: 14, padding: "11px 18px",
          background: T.bgSoft, borderBottom: `1px solid ${T.rule}`,
          fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.06em", textTransform: "uppercase",
        }}>
          <div>{isEn ? "Resume" : "简历"}</div>
          <div>{isEn ? "Source" : "来源"}</div>
          <div>{isEn ? "Lang" : "语言"}</div>
          <div>{isEn ? "Last edit" : "最近编辑"}</div>
          <div></div>
        </div>
        {sorted.map((r, i) => (
          <div key={r.id} style={{
            display: "grid",
            gridTemplateColumns: "1.8fr 1.4fr 0.6fr 1fr 132px",
            gap: 14, padding: "13px 18px",
            borderBottom: i < sorted.length - 1 ? `1px dotted ${T.rule}` : "none",
            alignItems: "center",
          }}>
            <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
              <Icon name="resume" size={13} color={T.accent} />
              <div style={{ minWidth: 0 }}>
                <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{r.name}</div>
                <div style={{ fontSize: 10.5, color: T.ink3, fontFamily: "var(--ei-mono)", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{r.summary}</div>
              </div>
            </div>
            <div style={{ fontSize: 12.5, color: T.ink2, fontFamily: "var(--ei-mono)", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
              {r.sourceName}
            </div>
            <div>
              <span style={{ fontFamily: "var(--ei-mono)", fontSize: 9, letterSpacing: "0.08em", padding: "1px 6px", borderRadius: 2, background: T.bgSoft, color: T.ink3 }}>{r.langTag}</span>
            </div>
            <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{r.updatedAt}</div>
            <div style={{ display: "flex", justifyContent: "flex-end", gap: 6 }}>
              <button onClick={() => openResume(r)} style={{
                padding: "5px 10px", fontSize: 12, cursor: "pointer",
                background: "transparent", color: T.ink2,
                border: `1px solid ${T.rule}`, borderRadius: 2, fontFamily: "var(--ei-sans)",
              }}>
                {isEn ? "Open" : "打开"}
              </button>
              <button onClick={() => onDelete && onDelete(r.id)} title={isEn ? "Delete resume" : "删除简历"} style={{
                width: 29, height: 29, cursor: "pointer",
                background: "transparent", color: T.ink3,
                border: `1px solid ${T.rule}`, borderRadius: 2,
                display: "inline-flex", alignItems: "center", justifyContent: "center",
              }}>
                <Icon name="trash" size={13} />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// ─────────────────────────────────────────────────────────────────────
// DETAIL VIEW — read-only resume body
// ─────────────────────────────────────────────────────────────────────
const ResumeDetailView = ({ T, lang, nav, resume }) => {
  const isEn = lang === "en";
  const back = () => nav("resume_versions", {});
  const pending = (resume.parseStatus === "queued" || resume.parseStatus === "processing") && (!Array.isArray(resume.text) || resume.text.length === 0);
  const failed = resume.parseStatus === "failed" && (!Array.isArray(resume.text) || resume.text.length === 0);

  if (pending || failed) {
    return <ResumeParseStateView T={T} lang={lang} nav={nav} resume={resume} failed={failed} />;
  }

  return (
    <div className="ei-fadein" style={{ maxWidth: 1320, margin: "0 auto", padding: "32px 48px 96px" }}>
      <button onClick={back} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 14, display: "flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {isEn ? "Back to resume workshop" : "返回简历工坊"}
      </button>

      <div style={{ marginBottom: 20 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>
          <span style={{ color: T.ink3 }}>{isEn ? "Resume workshop" : "简历工坊"}</span>
          <span style={{ margin: "0 8px", color: T.ink4 }}>›</span>
          <span style={{ color: T.ink2 }}>{resume.name}</span>
        </div>
        <h1 className="ei-serif" style={{ fontSize: 32, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.2 }}>
          {resume.name}
        </h1>
        <div style={{ fontSize: 13, color: T.ink3, marginTop: 8, fontFamily: "var(--ei-mono)" }}>
          {resume.sourceName} · {resume.createdAt} · {isEn ? "last edit" : "最近编辑"} {resume.updatedAt}
        </div>
      </div>

      <ResumePreviewTab T={T} lang={lang} resume={resume} />
    </div>
  );
};

// ─── WAITING / FAILURE STATE ─────────────────────────────────────────
const ResumeParseStateView = ({ T, lang, nav, resume, failed }) => {
  const isEn = lang === "en";
  return (
    <div className="ei-fadein" style={{ maxWidth: 960, margin: "0 auto", padding: "72px 48px 96px", textAlign: "center" }}>
      <button onClick={() => nav("resume_versions", {})} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 26, display: "inline-flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {isEn ? "Back to resume workshop" : "返回简历工坊"}
      </button>
      <div style={{ width: 72, height: 72, borderRadius: 36, margin: "0 auto 22px", border: `1px solid ${failed ? T.danger : T.rule}`, color: failed ? T.danger : T.accent, display: "flex", alignItems: "center", justifyContent: "center", boxShadow: "0 18px 44px rgba(30, 22, 15, 0.10)" }}>
        <Icon name={failed ? "info" : "sparkle"} size={28} />
      </div>
      <div className="ei-label" style={{ color: failed ? T.danger : T.accent, marginBottom: 10 }}>
        {failed ? (isEn ? "PARSE FAILED" : "解析失败") : (isEn ? "PARSING RESUME" : "正在解析简历")}
      </div>
      <h1 className="ei-serif" style={{ fontSize: 34, margin: 0, color: T.ink, lineHeight: 1.18 }}>
        {failed ? (isEn ? "We could not prepare this resume." : "暂时无法生成这份简历。") : (isEn ? "Preparing resume detail." : "正在生成简历详情。")}
      </h1>
      <p style={{ fontSize: 14, color: T.ink3, lineHeight: 1.7, maxWidth: 560, margin: "14px auto 0" }}>
        {failed
          ? (isEn ? "Delete this resume and upload another copy, or return to the list." : "你可以删除这份简历后重新上传，或先返回列表。")
          : (isEn ? "PDF sources render as a plain page stack, while paste and Markdown sources render as Markdown." : "PDF 来源会以页面栈展示，粘贴和 Markdown 来源会以 Markdown 渲染。")}
      </p>
      <div style={{ marginTop: 28, fontSize: 12.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
        {resume.sourceName}
      </div>
    </div>
  );
};

// ─── RESUME BODY ─────────────────────────────────────────────────────
const ResumePreviewTab = ({ T, lang, resume }) => {
  const isEn = lang === "en";
  const isPdfSource = resume.sourceType === "upload" && /\.pdf$/i.test(resume.sourceName || "");
  const lines = Array.isArray(resume.text) && resume.text.length > 0
    ? resume.text
    : [isEn ? "No readable resume content yet." : "暂无可读简历正文。"];
  if (isPdfSource) {
    return (
      <div style={{ display: "flex", justifyContent: "center", alignItems: "flex-start" }}>
        <div style={{ width: "min(100%, 860px)", minHeight: 720, background: "#f6f3ee", color: "#222", border: `1px solid ${T.rule}`, boxShadow: "0 18px 50px rgba(30, 22, 15, 0.10)", borderRadius: 3, padding: "28px", fontFamily: "Georgia, serif" }}>
          <div style={{ display: "flex", flexDirection: "column", gap: 22, alignItems: "center" }}>
            {[0, 1].map((pageIndex) => (
              <div key={pageIndex} style={{ width: "min(100%, 720px)", minHeight: pageIndex === 0 ? 640 : 520, background: "#ffffff", border: `1px solid ${T.rule}`, boxShadow: "0 10px 30px rgba(30,22,15,0.08)", padding: "42px 50px" }}>
                {pageIndex === 0 && (
                  <>
                    <div style={{ fontSize: 30, fontWeight: 600, marginBottom: 18 }}>{resume.name}</div>
                    <div style={{ height: 1, background: "#222", marginBottom: 22 }} />
                  </>
                )}
                {lines.map((line, index) => (
                  <div key={`${pageIndex}-${index}`} style={{ height: 12, width: `${88 - ((index + pageIndex) % 4) * 11}%`, background: pageIndex === 0 && index < 2 ? "#222" : "#d6d0c8", marginBottom: 13, opacity: pageIndex === 0 && index < 2 ? 0.92 : 1 }} />
                ))}
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }
  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "flex-start" }}>
      <div style={{ width: "min(100%, 860px)", minHeight: 720, background: "#f6f3ee", color: "#222", border: `1px solid ${T.rule}`, boxShadow: "0 18px 50px rgba(30, 22, 15, 0.10)", borderRadius: 3, padding: "28px", fontFamily: "Georgia, serif" }}>
        <div style={{ width: "min(100%, 720px)", minHeight: 664, margin: "0 auto", background: "#ffffff", border: `1px solid ${T.rule}`, boxShadow: "0 10px 30px rgba(30,22,15,0.08)", padding: "44px 56px" }}>
          <div style={{ fontSize: 14.5, lineHeight: 1.85, color: "#333" }}>
            {lines.map((line, index) => {
              if (/^##\s+/.test(line)) {
                return <h3 key={index} style={{ fontSize: 18, margin: "22px 0 6px" }}>{line.replace(/^##\s+/, "")}</h3>;
              }
              if (/^#\s+/.test(line)) {
                return <h2 key={index} style={{ fontSize: 22, margin: "20px 0 8px" }}>{line.replace(/^#\s+/, "")}</h2>;
              }
              if (/^-\s+/.test(line)) {
                return <div key={index} style={{ display: "grid", gridTemplateColumns: "16px 1fr", gap: 8, margin: "6px 0" }}><span>•</span><span>{line.replace(/^-\s+/, "")}</span></div>;
              }
              return <p key={index} style={{ margin: "8px 0" }}>{line}</p>;
            })}
          </div>
        </div>
      </div>
    </div>
  );
};

// ─── Create flow — upload or paste only (guided Q&A outside current scope, D-20). Successful submit opens the detail route waiting state. ──
const ResumeCreateFlow = ({ T, lang, onBack, onCreateResume }) => {
  const [createMode, setCreateMode] = React.useState("upload");
  const [resumeText, setResumeText] = React.useState("");
  const [pickedFile, setPickedFile] = React.useState(null);

  const sourceLabel = createMode === "upload"
    ? (pickedFile ? pickedFile.name : (lang === "en" ? "Uploaded file" : "上传文件"))
    : (lang === "en" ? "Pasted text" : "粘贴文本");

  const handleSubmit = () => {
    if (onCreateResume) {
      onCreateResume(sourceLabel, createMode === "paste" ? resumeText : "");
      return;
    }
    onBack();
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 1220, margin: "0 auto", padding: "40px 48px 96px" }}>
      <button onClick={onBack} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 22, display: "flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {lang === "en" ? "Back to resume workshop" : "返回简历工坊"}
      </button>

      <div style={{ marginBottom: 26 }}>
        <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{lang === "en" ? "FIRST RESUME" : "创建第一份简历"}</div>
        <h1 className="ei-serif" style={{ fontSize: 38, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.15 }}>
          {lang === "en" ? "Start from a file or pasted text." : "上传文件，或直接粘贴简历内容。"}
        </h1>
        <div style={{ fontSize: 14, color: T.ink2, marginTop: 10, maxWidth: 720, lineHeight: 1.6 }}>
          {lang === "en" ? "We keep the source, parse it, and then render the detail with the matching source viewer." : "系统会保留原始来源，解析完成后按来源格式自动展示详情。"}
        </div>
      </div>

      <div style={{ maxWidth: 860 }}>
        <Card T={T} pad={0}>
          <div style={{ display: "flex", borderBottom: `1px solid ${T.rule}`, overflowX: "auto" }}>
            {[
              { k: "upload", icon: "upload", zh: "上传文件", en: "Upload" },
              { k: "paste", icon: "file", zh: "粘贴内容", en: "Paste" },
            ].map((mode) => (
              <button key={mode.k} onClick={() => setCreateMode(mode.k)} style={{
                padding: "15px 20px", background: createMode === mode.k ? T.bgSoft : "transparent", border: "none",
                borderBottom: `2px solid ${createMode === mode.k ? T.accent : "transparent"}`, color: createMode === mode.k ? T.ink : T.ink3,
                display: "flex", alignItems: "center", gap: 8, cursor: "pointer", fontFamily: "var(--ei-sans)", marginBottom: -1,
              }}>
                <Icon name={mode.icon} size={14} /> {lang === "en" ? mode.en : mode.zh}
              </button>
            ))}
          </div>

          {createMode === "upload" && (
            <div style={{ padding: 24 }}>
              <div style={{
                minHeight: 260, border: `1px dashed ${T.ink4}`, borderRadius: 3, background: T.bgSoft,
                display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", gap: 14, textAlign: "center", padding: 28,
              }}>
                <div style={{ width: 54, height: 54, borderRadius: 27, background: T.bgCard, border: `1px solid ${T.rule}`, color: T.accent, display: "flex", alignItems: "center", justifyContent: "center" }}>
                  <Icon name="upload" size={24} />
                </div>
                <div className="ei-serif" style={{ fontSize: 22, color: T.ink }}>{lang === "en" ? "Drop a PDF / Markdown resume" : "拖入 PDF / Markdown 简历"}</div>
                <div style={{ fontSize: 13, color: T.ink3, maxWidth: 460, lineHeight: 1.55 }}>
                  {lang === "en" ? "PDF / Markdown / TXT, up to 2 MiB. PDFs render as a page stack; Markdown and text render as Markdown." : "支持 PDF / Markdown / TXT，最大 2MiB。PDF 以页面栈展示，Markdown 和文本以 Markdown 渲染。"}
                </div>
                <Btn T={T} variant="accent" icon="upload" onClick={() => {
                  const input = document.createElement("input");
                  input.type = "file";
                  input.accept = ".pdf,.md,.markdown,.txt";
                  input.onchange = (e) => {
                    const f = e.target.files && e.target.files[0];
                    if (f) {
                      setPickedFile(f);
                      if (onCreateResume) {
                        onCreateResume(f.name, "");
                      } else {
                        onBack();
                      }
                    }
                  };
                  input.click();
                }}>{lang === "en" ? "Choose file" : "选择文件"}</Btn>
                {pickedFile && (
                  <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 4 }}>
                    {lang === "en" ? "Selected: " : "已选择："}<span style={{ color: T.ink }}>{pickedFile.name}</span>
                  </div>
                )}
              </div>
            </div>
          )}

          {createMode === "paste" && (
            <div style={{ padding: 24 }}>
              <textarea
                value={resumeText}
                onChange={(e) => setResumeText(e.target.value)}
                placeholder={lang === "en" ? "Paste your resume text here…" : "把你的简历内容粘贴到这里…"}
                style={{
                  width: "100%", minHeight: 260, resize: "vertical", border: `1px solid ${T.rule}`, borderRadius: 2,
                  padding: 16, background: T.bg, color: T.ink, fontSize: 14, lineHeight: 1.65, outline: "none",
                }}
              />
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: 14, gap: 14, flexWrap: "wrap" }}>
                <div style={{ fontSize: 12.5, color: T.ink3 }}>
                  {lang === "en" ? "Raw text is retained and converted into a Markdown snapshot." : "原始文本会保留，并转成 Markdown 快照。"}
                </div>
                <Btn T={T} variant="accent" icon="arrow_right" disabled={!resumeText.trim()} onClick={handleSubmit}>{lang === "en" ? "Save and open" : "保存并打开"}</Btn>
              </div>
            </div>
          )}
        </Card>

      </div>
    </div>
  );
};

window.ResumeVersionsScreen = ResumeWorkshopScreen;
