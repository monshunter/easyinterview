// Resume workshop — flat IA (D-20): a plain list of resume assets.
// Flat resume list. Each resume keeps its source-derived content and opens to
// a read-only detail view.

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
      text: [
        isEn ? "Liu Zhe · Senior Frontend Engineer · Shanghai" : "刘哲 · 资深前端工程师 · 上海",
        isEn ? "Star-Ring · Senior Frontend · 2022-now" : "星环科技 · 资深前端 · 2022 至今",
        isEn ? "Worked on checkout performance and complex admin surfaces." : "负责结账流程性能改进和复杂后台系统建设。",
        isEn ? "Built shared UI components for internal products." : "为内部产品建设通用 UI 组件。",
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
      text: [
        "Liu Zhe · Frontend Platform Engineer",
        "Built platform tooling, design system infrastructure, and TypeScript foundations.",
        "Worked with distributed teams across APAC and US time zones.",
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
      text: [
        isEn ? "Liu Zhe · Senior Frontend Engineer" : "刘哲 · 资深前端工程师",
        isEn ? "Focus: collaboration impact and rollout stories." : "侧重协作影响力与落地故事。",
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
  const resumes = React.useMemo(() => [...baseData, ...createdResumes], [baseData, createdResumes]);

  const [flow, setFlow] = React.useState(params.flow === "create" ? "create" : "list");
  React.useEffect(() => {
    if (params.flow === "create") setFlow("create");
  }, [params.flow]);

  const isEn = lang === "en";

  const addCreatedResume = (sourceLabel, rawText = "") => {
    const suffix = Date.now();
    const lines = rawText
      .split(/\n+/)
      .map((line) => line.trim())
      .filter(Boolean);
    const resume = {
      id: `resume-created-${suffix}`,
      name: lines[0] || sourceLabel || (isEn ? "New resume" : "新建简历"),
      langTag: isEn ? "CN" : "中",
      sourceType: rawText ? "paste" : "upload",
      sourceName: rawText ? (isEn ? "Pasted text" : "粘贴文本") : (sourceLabel || (isEn ? "Uploaded file" : "上传文件")),
      createdAt: resumeToday,
      updatedAt: resumeToday,
      summary: isEn ? "Registered and opened directly" : "已注册并直接打开",
      text: lines.length > 0 ? lines : [
        isEn ? "Original file content will appear here after extraction." : "文件提取后的原始简历正文会显示在这里。",
      ],
    };
    setCreatedResumes((prev) => [...prev, resume]);
    nav("resume_versions", { resumeId: resume.id });
    resumeNotify(lang, `已创建 ${resume.name} · 正在打开详情`, `Created ${resume.name} · opening detail`);
  };

  const resumeId = params.resumeId;
  const detailResume = resumeId ? resumes.find((r) => r.id === resumeId) : null;

  if (flow === "create") {
    return <ResumeCreateFlow T={T} lang={lang} nav={nav} onBack={() => setFlow("list")} onCreateResume={addCreatedResume} />;
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
    />
  );
};

// ─────────────────────────────────────────────────────────────────────
// LIST VIEW — one flat table of resume assets
// ─────────────────────────────────────────────────────────────────────
const ResumeListView = ({ T, lang, nav, resumes, onCreate }) => {
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
          gridTemplateColumns: "1.8fr 1.4fr 0.6fr 1fr 100px",
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
            gridTemplateColumns: "1.8fr 1.4fr 0.6fr 1fr 100px",
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
            <button onClick={() => openResume(r)} style={{
              padding: "5px 12px", fontSize: 12, cursor: "pointer",
              background: "transparent", color: T.ink2,
              border: `1px solid ${T.rule}`, borderRadius: 2, fontFamily: "var(--ei-sans)",
            }}>
              {isEn ? "Open" : "打开"}
            </button>
          </div>
        ))}
      </div>

      {/* New resume CTA */}
      <button onClick={onCreate} style={{
        marginTop: 14, width: "100%",
        padding: "18px 18px", background: "transparent",
        border: `1px dashed ${T.rule}`, borderRadius: 3, color: T.ink3,
        cursor: "pointer", fontFamily: "var(--ei-sans)", fontSize: 13.5,
        display: "flex", alignItems: "center", justifyContent: "center", gap: 8,
      }}>
        <Icon name="upload" size={14} /> {isEn ? "Upload or paste another resume" : "上传或粘贴另一份简历"}
      </button>
    </div>
  );
};

// ─────────────────────────────────────────────────────────────────────
// DETAIL VIEW — read-only resume body
// ─────────────────────────────────────────────────────────────────────
const ResumeDetailView = ({ T, lang, nav, resume }) => {
  const isEn = lang === "en";
  const back = () => nav("resume_versions", {});

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

// ─── RESUME BODY ─────────────────────────────────────────────────────
const ResumePreviewTab = ({ T, lang, resume }) => {
  const isEn = lang === "en";
  const lines = Array.isArray(resume.text) && resume.text.length > 0
    ? resume.text
    : [isEn ? "No readable resume content yet." : "暂无可读简历正文。"];
  return (
    <div style={{ display: "flex", justifyContent: "center", alignItems: "flex-start" }}>
      <div style={{ width: "min(100%, 860px)", background: "#fff", color: "#222", border: `1px solid ${T.rule}`, boxShadow: "0 18px 50px rgba(30, 22, 15, 0.10)", borderRadius: 3, padding: "44px 56px", fontFamily: "Georgia, serif", minHeight: 720 }}>
        <div style={{ fontSize: 28, fontWeight: 600, letterSpacing: "-0.02em" }}>{resume.name}</div>
        <div style={{ fontSize: 14, color: "#666", marginTop: 4 }}>{resume.summary}</div>
        <div style={{ height: 1, background: "#222", margin: "20px 0 18px" }} />
        <div style={{ whiteSpace: "pre-wrap", fontSize: 14.5, lineHeight: 1.85, color: "#333" }}>
          {lines.join("\n")}
        </div>
      </div>
    </div>
  );
};

// ─── Create flow — upload or paste only (guided Q&A outside current scope, D-20). Successful submit opens detail directly. ──
const ResumeCreateFlow = ({ T, lang, nav, onBack, onCreateResume }) => {
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
          {lang === "en" ? "We keep the original source and open the read-only resume detail as soon as it is registered." : "系统会保留原始文件或原始文本，注册成功后直接打开只读简历详情。"}
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 340px", gap: 22, alignItems: "start" }}>
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
                <div className="ei-serif" style={{ fontSize: 22, color: T.ink }}>{lang === "en" ? "Drop a PDF / DOCX / Markdown resume" : "拖入 PDF / DOCX / Markdown 简历"}</div>
                <div style={{ fontSize: 13, color: T.ink3, maxWidth: 460, lineHeight: 1.55 }}>
                  {lang === "en" ? "The source file is parsed into the resume content shown later in the read-only detail view." : "原始文件会被解析成后续只读详情页展示的简历内容。"}
                </div>
                <Btn T={T} variant="accent" icon="upload" onClick={() => {
                  const input = document.createElement("input");
                  input.type = "file";
                  input.accept = ".pdf,.docx,.md,.txt";
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
                  {lang === "en" ? "Raw text is retained and opened directly in the read-only detail view." : "原始文本会保留，并直接在只读详情页展示。"}
                </div>
                <Btn T={T} variant="accent" icon="arrow_right" disabled={!resumeText.trim()} onClick={handleSubmit}>{lang === "en" ? "Save and open" : "保存并打开"}</Btn>
              </div>
            </div>
          )}
        </Card>

        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <Card T={T}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "WHAT GETS SAVED" : "会保存什么"}</div>
            {[
              { icon: "file", title: lang === "en" ? "Original source" : "原始来源", body: lang === "en" ? "File or pasted text stays traceable." : "文件或粘贴文本都会保留来源。" },
              { icon: "resume", title: lang === "en" ? "Read-only detail" : "只读详情", body: lang === "en" ? "After registration, the detail view opens immediately with the resume content itself." : "注册成功后立即打开详情，看到的就是简历内容本身。" },
            ].map((item, i) => (
              <div key={item.title} style={{ display: "grid", gridTemplateColumns: "26px 1fr", gap: 10, padding: "12px 0", borderBottom: i < 1 ? `1px dotted ${T.rule}` : "none" }}>
                <Icon name={item.icon} size={15} color={T.accent} style={{ marginTop: 2 }} />
                <div>
                  <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>{item.title}</div>
                  <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.5, marginTop: 2 }}>{item.body}</div>
                </div>
              </div>
            ))}
          </Card>
          <Card T={T}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "WHAT HAPPENS NEXT" : "接下来"}</div>
            <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.7 }}>
              {lang === "en"
                ? <>After you submit, we'll <span style={{ color: T.ink, fontWeight: 500 }}>open the resume detail directly</span>.</>
                : <>提交之后会<span style={{ color: T.ink, fontWeight: 500 }}>直接打开简历详情</span>。</>}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
};

window.ResumeVersionsScreen = ResumeWorkshopScreen;
