// Resume workshop — flat IA (D-20): a plain list of resume assets.
// No version tree, no master/targeted split, no branch flow. Each resume keeps
// its read-only original source plus structured content. Rewrites are
// accept-only; accepted rewrites are saved via an explicit preview step that
// either overwrites this resume or saves a new one.

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

const buildResumePlainText = (lang, resume) => {
  const isEn = lang === "en";
  return [
    isEn ? "Liu Zhe" : "刘哲",
    isEn ? "Senior Frontend Engineer · Shanghai" : "资深前端工程师 · 上海",
    "",
    isEn ? "Experience" : "工作经历",
    isEn ? "Star-Ring · Senior Frontend Engineer · 2022-now" : "星环科技 · 资深前端工程师 · 2022 至今",
    isEn
      ? "Led migration of the checkout surface to RSC + selective hydration, cutting LCP from 3.2s to 1.4s and lifting quarterly GMV by 1.8M."
      : "主导结账链路迁移到 RSC + 选择性注水，LCP 3.2s -> 1.4s，季度 GMV +180 万。",
    isEn
      ? "Drove Design System v1 adoption across 5 products in 6 months and reduced new-dev ramp about 50%."
      : "6 个月内推动 Design System v1 在 5 个产品落地，新人上手时间缩短约 50%。",
    "",
    isEn ? "Resume" : "简历",
    resume ? resume.name : "",
  ].join("\n");
};

const buildBullets = (lang) => {
  const isEn = lang === "en";
  const sectionA = isEn ? "Senior Frontend · Star-Ring · 2022-now" : "资深前端 · 星环科技 · 2022 至今";
  const sectionB = isEn ? "Frontend · Lumen · 2019-2022" : "前端 · Lumen · 2019-2022";
  return [
    {
      id: "b1", section: sectionA,
      original: isEn ? "Worked on checkout performance improvements for the e-commerce team, collaborating closely with backend engineers." : "负责电商团队结账流程的性能改进工作，与后端工程师紧密协作。",
      rewritten: isEn ? "Led migration of the checkout surface to RSC + selective hydration, cutting LCP from 3.2s to 1.4s and lifting quarterly GMV by 1.8M (8% → 4.2% abandon)." : "主导结账链路迁移到 RSC + 选择性注水，LCP 3.2s → 1.4s，流失率 8% → 4.2%，季度 GMV +180 万。",
      why: isEn ? ["Weak → strong ownership verb", "Adds quantified impact", "Names the specific architecture"] : ["动词从弱到强：「负责」→「主导」", "加入量化影响", "具体指出架构选择"],
    },
    {
      id: "b2", section: sectionA,
      original: isEn ? "Rolled out a design system across multiple product teams." : "在多个产品团队推广了设计系统。",
      rewritten: isEn ? "Drove Design System v1 adoption across 5 products in 6 months (4 live, 1 in progress) — ran 3 workshops, paired migrations with 2 pilot teams, reduced new-dev ramp ~50%." : "6 个月内推动 Design System v1 在 5 个产品落地（4 上线、1 进行中）——办 3 次推广会、与 2 个试点团队结对迁移，新人上手时间缩短约 50%。",
      why: isEn ? ["Names the scale (5 products)", "Shows method, not just outcome", "Anchored on developer time saved"] : ["量化范围：5 个产品", "讲方法而不只是结果", "以节省的工时收口"],
    },
    {
      id: "b3", section: sectionB,
      original: isEn ? "Built and shipped various features for the core product." : "为核心产品构建并交付了多个功能。",
      rewritten: isEn ? "Shipped 14 features to the order-management core over 3 years, including a batch-edit surface that became the #2 most-used power-user flow." : "3 年内为订单管理核心交付 14 个功能，其中批量编辑成为重度用户第 2 常用流程。",
      why: isEn ? ["Vague → specific count", "Picks one feature worth name-checking", "Usage data gives credibility"] : ["模糊数量变具体", "挑一个值得点名的功能", "用使用数据建立可信度"],
    },
    {
      id: "b4", section: sectionB,
      original: isEn ? "Participated in code reviews and technical discussions." : "参与代码评审和技术讨论。",
      rewritten: isEn ? "Drove weekly architecture reviews for the order core and mentored 2 engineers through their first production launches." : "主持订单核心的每周架构评审，并带 2 名工程师完成首次生产上线。",
      why: isEn ? ["Generic duty → ownership", "Adds people-impact evidence", "Concrete cadence and scope"] : ["泛化职责变 ownership", "补充带人影响证据", "给出具体节奏和范围"],
    },
  ];
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

  const addCreatedResume = (sourceLabel) => {
    const suffix = Date.now();
    const resume = {
      id: `resume-created-${suffix}`,
      name: sourceLabel || (isEn ? "New resume" : "新建简历"),
      langTag: isEn ? "CN" : "中",
      sourceType: "upload",
      sourceName: sourceLabel || (isEn ? "Uploaded file" : "上传文件"),
      createdAt: resumeToday,
      updatedAt: resumeToday,
      summary: isEn ? "Parsed and confirmed in this prototype" : "本原型中解析并确认",
      text: [
        isEn ? "Liu Zhe · Senior Frontend Engineer · Shanghai" : "刘哲 · 资深前端工程师 · 上海",
        isEn ? "Star-Ring · Senior Frontend · 2022-now" : "星环科技 · 资深前端 · 2022 至今",
        isEn ? "Confirmed from the intake preview." : "已从预览确认页保存。",
      ],
    };
    setCreatedResumes((prev) => [...prev, resume]);
    setFlow("list");
    resumeNotify(lang, `已保存 ${resume.name} · 已加入简历工坊`, `Saved ${resume.name} · added to workshop`);
  };

  // Accepted rewrites land here through the confirm-preview step.
  const saveRewriteResult = (resume, acceptedCount, mode) => {
    if (mode === "new") {
      const suffix = Date.now();
      const created = {
        ...resume,
        id: `resume-rewrite-${suffix}`,
        name: isEn ? `${resume.name} · rewritten` : `${resume.name} · 改写稿`,
        createdAt: resumeToday,
        updatedAt: resumeToday,
      };
      setCreatedResumes((prev) => [...prev, created]);
      resumeNotify(lang,
        `已保存为新简历：${created.name}（${acceptedCount} 条改写）`,
        `Saved as a new resume: ${created.name} (${acceptedCount} rewrites)`);
      nav("resume_versions", { resumeId: created.id, tab: "preview" });
      return;
    }
    resumeNotify(lang,
      `已覆盖 ${resume.name}（${acceptedCount} 条改写写入，原始来源不变）`,
      `Overwrote ${resume.name} (${acceptedCount} rewrites applied; original source unchanged)`);
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
        initialTab={params.tab || "preview"}
        onSaveRewrites={saveRewriteResult}
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
  const openResume = (r, tab = "preview") => nav("resume_versions", { resumeId: r.id, tab });

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
              ? "A flat list of resume assets. Each one keeps its read-only original source. Open any resume to preview, accept rewrites, or edit by hand."
              : "简历按平铺列表管理。每份简历都保留只读的原始来源；打开任意一份进入预览、改写建议或手动编辑。"}
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
// DETAIL VIEW — single resume with tabs
// ─────────────────────────────────────────────────────────────────────
const ResumeDetailView = ({ T, lang, nav, resume, initialTab, onSaveRewrites }) => {
  const isEn = lang === "en";
  const [tab, setTab] = React.useState(initialTab || "preview");
  const [sourcePreviewOpen, setSourcePreviewOpen] = React.useState(false);

  const back = () => nav("resume_versions", {});

  React.useEffect(() => {
    setTab(initialTab || "preview");
  }, [initialTab, resume.id]);

  const exportPdf = () => {
    resumeNotify(lang, `正在生成 ${resume.name} PDF · 准备好后会提示下载`, `Generating ${resume.name} PDF · download prompt will appear when ready`, { duration: 2800 });
  };

  const copyText = () => {
    const text = buildResumePlainText(lang, resume);
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text)
        .then(() => resumeNotify(lang, `已复制 ${resume.name} 的纯文本`, `Copied plain text for ${resume.name}`))
        .catch(() => resumeNotify(lang, "当前环境不支持剪贴板写入", "Clipboard write is unavailable in this environment", { tone: "warn" }));
      return;
    }
    resumeNotify(lang, "当前环境不支持剪贴板写入", "Clipboard write is unavailable in this environment", { tone: "warn" });
  };

  const tabs = [
    { k: "preview",  label: isEn ? "Preview" : "预览",       icon: "file" },
    { k: "rewrites", label: isEn ? "Rewrites" : "改写建议", icon: "sparkle" },
    { k: "edit",     label: isEn ? "Edit" : "手动编辑",     icon: "edit" },
  ];

  return (
    <div className="ei-fadein" style={{ maxWidth: 1320, margin: "0 auto", padding: "32px 48px 96px" }}>
      <button onClick={back} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 14, display: "flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {isEn ? "Back to resume workshop" : "返回简历工坊"}
      </button>

      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", gap: 28, marginBottom: 20, flexWrap: "wrap" }}>
        <div>
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
        <div style={{ display: "flex", gap: 8 }}>
          <Btn T={T} variant="secondary" size="sm" icon="download" onClick={exportPdf}>{isEn ? "Export PDF" : "导出 PDF"}</Btn>
        </div>
      </div>

      {/* Tabs */}
      <div style={{ display: "flex", gap: 0, marginTop: 24, marginBottom: 22, borderBottom: `1px solid ${T.rule}` }}>
        {tabs.map((t) => {
          const active = tab === t.k;
          return (
            <button key={t.k}
              onClick={() => setTab(t.k)}
              style={{
                padding: "13px 20px", background: "transparent", border: "none",
                borderBottom: `2px solid ${active ? T.accent : "transparent"}`,
                color: active ? T.ink : T.ink3,
                cursor: "pointer",
                fontFamily: "var(--ei-sans)", fontSize: 14,
                display: "flex", alignItems: "center", gap: 8, marginBottom: -1,
              }}>
              <Icon name={t.icon} size={13} /> {t.label}
            </button>
          );
        })}
      </div>

      {/* Tab body */}
      {tab === "preview"  && <ResumePreviewTab  T={T} lang={lang} resume={resume} onExport={exportPdf} onCopy={copyText} onPreviewOriginal={() => setSourcePreviewOpen(true)} />}
      {tab === "rewrites" && <ResumeRewritesTab T={T} lang={lang} resume={resume} onSaveRewrites={onSaveRewrites} />}
      {tab === "edit"     && <ResumeEditTab     T={T} lang={lang} resume={resume} />}
      {sourcePreviewOpen && (
        <OriginalResumePreviewModal
          T={T}
          lang={lang}
          resume={resume}
          onClose={() => setSourcePreviewOpen(false)}
        />
      )}
    </div>
  );
};

// ─── PREVIEW TAB ─────────────────────────────────────────────────────
const ResumePreviewTab = ({ T, lang, resume, onExport, onCopy, onPreviewOriginal }) => {
  const isEn = lang === "en";
  return (
    <div style={{ display: "grid", gridTemplateColumns: "1fr 320px", gap: 22, alignItems: "start" }}>
      <div style={{ background: "#fff", color: "#222", border: `1px solid ${T.rule}`, boxShadow: "0 18px 50px rgba(30, 22, 15, 0.10)", borderRadius: 3, padding: "44px 56px", fontFamily: "Georgia, serif", minHeight: 720 }}>
        <div style={{ fontSize: 28, fontWeight: 600, letterSpacing: "-0.02em" }}>{isEn ? "Liu Zhe" : "刘哲"}</div>
        <div style={{ fontSize: 14, color: "#666", marginTop: 4 }}>{isEn ? "Senior Frontend Engineer · Shanghai" : "资深前端工程师 · 上海"}</div>
        <div style={{ fontSize: 12, color: "#888", marginTop: 4, fontFamily: "ui-monospace, monospace" }}>liuzhe@example.com · +86 138 0000 0000</div>
        <div style={{ height: 1, background: "#222", margin: "20px 0 18px" }} />

        <div style={{ fontSize: 11, color: "#888", letterSpacing: "0.14em", textTransform: "uppercase", marginBottom: 10 }}>{isEn ? "Experience" : "工作经历"}</div>
        <div style={{ fontSize: 15, fontWeight: 600, marginBottom: 2 }}>{isEn ? "Star-Ring · Senior Frontend Engineer" : "星环科技 · 资深前端工程师"}</div>
        <div style={{ fontSize: 12, color: "#666", marginBottom: 10, fontFamily: "ui-monospace, monospace" }}>2022 — {isEn ? "now" : "至今"}</div>
        <ul style={{ margin: 0, paddingLeft: 20, fontSize: 13.5, lineHeight: 1.75, color: "#333" }}>
          <li>{isEn
            ? "Led migration of the checkout surface to RSC + selective hydration, cutting LCP from 3.2s to 1.4s and lifting quarterly GMV by 1.8M."
            : "主导结账链路迁移到 RSC + 选择性注水，LCP 3.2s → 1.4s，季度 GMV +180 万。"}</li>
          <li>{isEn
            ? "Drove Design System v1 adoption across 5 products in 6 months — ran 3 workshops and reduced new-dev ramp ~50%."
            : "6 个月内推动 Design System v1 在 5 个产品落地——办 3 次推广会，新人上手时间缩短约 50%。"}</li>
        </ul>

        <div style={{ fontSize: 15, fontWeight: 600, marginTop: 18, marginBottom: 2 }}>{isEn ? "Lumen · Frontend Engineer" : "Lumen · 前端工程师"}</div>
        <div style={{ fontSize: 12, color: "#666", marginBottom: 10, fontFamily: "ui-monospace, monospace" }}>2019 — 2022</div>
        <ul style={{ margin: 0, paddingLeft: 20, fontSize: 13.5, lineHeight: 1.75, color: "#333" }}>
          <li>{isEn
            ? "Shipped 14 features to the order-management core; the batch-edit surface became the #2 most-used power-user flow."
            : "为订单管理核心交付 14 个功能；批量编辑成为重度用户第 2 常用流程。"}</li>
        </ul>

        <div style={{ fontSize: 11, color: "#888", letterSpacing: "0.14em", textTransform: "uppercase", marginTop: 22, marginBottom: 10 }}>{isEn ? "Skills" : "技能"}</div>
        <div style={{ fontSize: 13.5, color: "#444", lineHeight: 1.75 }}>React · TypeScript · Performance · Design System · Platform Engineering</div>
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
        <Card T={T}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{isEn ? "ABOUT THIS VIEW" : "关于这个视图"}</div>
          <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.65 }}>
            {isEn
              ? "Read-only preview of the rendered resume. Switch to Rewrites to accept AI suggestions, or Edit to change fields by hand."
              : "只读预览。切到「改写建议」逐条采纳 AI 建议，或切到「手动编辑」直接改字段。"}
          </div>
          <div style={{ display: "flex", gap: 8, marginTop: 14, flexWrap: "wrap" }}>
            <Btn T={T} variant="secondary" size="sm" icon="download" onClick={onExport}>{isEn ? "Export PDF" : "导出 PDF"}</Btn>
            <Btn T={T} variant="ghost" size="sm" icon="file" onClick={onCopy}>{isEn ? "Copy text" : "复制纯文本"}</Btn>
          </div>
        </Card>

        <Card T={T}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{isEn ? "ORIGINAL SOURCE" : "原始来源"}</div>
          <div style={{ fontSize: 13, color: T.ink2, fontFamily: "var(--ei-mono)", marginBottom: 6, wordBreak: "break-all" }}>{resume.sourceName}</div>
          <div style={{ fontSize: 12, color: T.ink3 }}>{resume.sourceType === "paste" ? (isEn ? "Pasted text" : "粘贴文本") : (isEn ? "Uploaded file" : "上传文件")} · {resume.createdAt}</div>
          <Btn T={T} variant="ghost" size="sm" icon="file" style={{ marginTop: 12 }} onClick={onPreviewOriginal}>{isEn ? "View original" : "查看原件"}</Btn>
        </Card>
      </div>
    </div>
  );
};

// ─── REWRITES TAB — accept-only, with a confirm-preview save step ────
const ResumeRewritesTab = ({ T, lang, resume, onSaveRewrites }) => {
  const isEn = lang === "en";
  const allBullets = React.useMemo(() => buildBullets(lang), [lang]);
  const [acceptedIds, setAcceptedIds] = React.useState({});
  const [selected, setSelected] = React.useState("b1");
  const [confirmOpen, setConfirmOpen] = React.useState(false);

  const bullets = allBullets.map((b) => ({ ...b, accepted: !!acceptedIds[b.id] }));
  const sel = bullets.find((b) => b.id === selected) || bullets[0];
  const accepted = bullets.filter((b) => b.accepted).length;

  const acceptBullet = (id) => {
    if (acceptedIds[id]) return;
    setAcceptedIds({ ...acceptedIds, [id]: true });
    if (window.eiToast) {
      window.eiToast(
        isEn ? "Accepted · choose overwrite or save-as-new when you save" : "已采纳 · 保存时可选择覆盖或另存",
        { tone: "ok", duration: 2400 }
      );
    }
  };

  const confirmSave = (mode) => {
    setConfirmOpen(false);
    onSaveRewrites && onSaveRewrites(resume, accepted, mode);
  };

  return (
    <div>
      {/* Scope banner — accepted rewrites only land after the confirm preview */}
      <div style={{
        display: "flex", justifyContent: "space-between", alignItems: "center", gap: 14,
        padding: "10px 14px", marginBottom: 16,
        background: T.accentSoft, border: `1px solid ${T.accent}`, borderRadius: 2,
        flexWrap: "wrap",
      }}>
        <div style={{ fontSize: 13, color: T.ink2 }}>
          <Icon name="info" size={12} color={T.accent} style={{ marginRight: 6 }} />
          {isEn
            ? <>Each suggestion has a single action: <strong>Accept</strong>. Doing nothing skips it. Accepted rewrites are saved only after the preview step, where you choose <strong>overwrite this resume</strong> or <strong>save as a new resume</strong>.</>
            : <>每条建议只有<strong>「采纳」</strong>一个动作，默认不动作即忽略。采纳的改写要经过确认前预览才会保存：可选<strong>覆盖原简历</strong>或<strong>保存为新简历</strong>。</>
          }
        </div>
        <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
          <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
            {accepted} {isEn ? "accepted" : "已采纳"} · {bullets.length - accepted} {isEn ? "untouched" : "未动作"}
          </div>
          <Btn T={T} variant="accent" size="sm" icon="arrow_right" disabled={accepted === 0} onClick={() => setConfirmOpen(true)}>
            {isEn ? "Preview & save" : "预览并保存"}
          </Btn>
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1.3fr", gap: 20 }}>
        {/* Bullet list */}
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{isEn ? "SUGGESTED REWRITES" : "建议改写"}</div>
          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {bullets.map((b) => {
              const active = b.id === selected;
              const statusC = b.accepted ? T.ok : T.ink3;
              return (
                <button key={b.id} onClick={() => setSelected(b.id)} style={{
                  padding: "14px 16px", textAlign: "left", cursor: "pointer",
                  background: active ? T.bgSoft : T.bgCard,
                  border: `1px solid ${active ? T.accent : T.rule}`,
                  borderRadius: 2, fontFamily: "var(--ei-sans)",
                }}>
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 10, marginBottom: 6 }}>
                    <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>{b.section}</div>
                    <div style={{ display: "flex", gap: 4, alignItems: "center", fontSize: 10.5, color: statusC, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>
                      <div style={{ width: 5, height: 5, borderRadius: 3, background: statusC }} />
                      {b.accepted ? (isEn ? "ACCEPTED" : "已采纳") : (isEn ? "SUGGESTED" : "建议")}
                    </div>
                  </div>
                  <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.5 }}>
                    {b.rewritten.slice(0, 90)}{b.rewritten.length > 90 ? "…" : ""}
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* Diff detail */}
        <div>
          <Card T={T} pad={0}>
            <div style={{ padding: "14px 22px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <div className="ei-label" style={{ color: T.ink3 }}>{sel.section}</div>
              <button onClick={() => acceptBullet(sel.id)} style={{ padding: "5px 14px", fontSize: 12, cursor: sel.accepted ? "default" : "pointer", background: sel.accepted ? T.ok : T.accent, color: "#fff", border: "none", borderRadius: 2, fontFamily: "var(--ei-sans)" }}>
                <Icon name="check" size={11} style={{ marginRight: 4 }} stroke={2.5} /> {sel.accepted ? (isEn ? "Accepted" : "已采纳") : (isEn ? "Accept" : "采纳")}
              </button>
            </div>

            {/* Original */}
            <div style={{ padding: "16px 22px", borderBottom: `1px dotted ${T.rule}` }}>
              <div style={{ display: "flex", gap: 10, alignItems: "center", marginBottom: 8 }}>
                <div style={{ padding: "2px 8px", background: T.dangerSoft, color: T.danger, fontSize: 10.5, fontFamily: "var(--ei-mono)", letterSpacing: "0.08em", borderRadius: 2 }}>
                  − {isEn ? "ORIGINAL" : "原句"}
                </div>
                <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{isEn ? "from this resume" : "来自当前简历"}</div>
              </div>
              <div style={{ fontSize: 14.5, color: T.ink2, lineHeight: 1.65, fontFamily: "var(--ei-serif)", background: T.dangerSoft, padding: "12px 14px", borderRadius: 2, borderLeft: `2px solid ${T.danger}` }}>
                {sel.original}
              </div>
            </div>

            {/* Rewritten */}
            <div style={{ padding: "16px 22px", borderBottom: `1px dotted ${T.rule}` }}>
              <div style={{ display: "flex", gap: 10, alignItems: "center", marginBottom: 8 }}>
                <div style={{ padding: "2px 8px", background: T.okSoft, color: T.ok, fontSize: 10.5, fontFamily: "var(--ei-mono)", letterSpacing: "0.08em", borderRadius: 2 }}>
                  + {isEn ? "REWRITTEN" : "AI 改写"}
                </div>
                <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{isEn ? "confidence · high" : "置信度 · 高"}</div>
              </div>
              <div style={{ fontSize: 14.5, color: T.ink, lineHeight: 1.65, fontFamily: "var(--ei-serif)", background: T.okSoft, padding: "12px 14px", borderRadius: 2, borderLeft: `2px solid ${T.ok}` }}>
                {sel.rewritten}
              </div>
            </div>

            {/* Why */}
            <div style={{ padding: "16px 22px" }}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{isEn ? "WHY THIS CHANGE" : "为什么这么改"}</div>
              <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                {sel.why.map((w, i) => (
                  <div key={i} style={{ display: "flex", gap: 10, fontSize: 13, color: T.ink2, lineHeight: 1.5 }}>
                    <Icon name="sparkle" size={12} color={T.accent} style={{ marginTop: 3, flexShrink: 0 }} />
                    <span>{w}</span>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        </div>
      </div>

      {confirmOpen && (
        <RewriteSaveConfirmModal
          T={T}
          lang={lang}
          resume={resume}
          bullets={bullets.filter((b) => b.accepted)}
          onClose={() => setConfirmOpen(false)}
          onConfirm={confirmSave}
        />
      )}
    </div>
  );
};

// ─── Confirm preview: overwrite this resume, or save as a new one ────
const RewriteSaveConfirmModal = ({ T, lang, resume, bullets, onClose, onConfirm }) => {
  const isEn = lang === "en";
  const [mode, setMode] = React.useState("overwrite"); // "overwrite" | "new"
  return (
    <div style={{ position: "fixed", inset: 0, background: "rgba(24, 20, 16, 0.24)", zIndex: 80, display: "flex", alignItems: "center", justifyContent: "center", padding: 24 }} onClick={onClose}>
      <div className="ei-fadein" onClick={(e) => e.stopPropagation()} style={{ width: "min(760px, 100%)", maxHeight: "88vh", overflow: "auto", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4, boxShadow: "0 24px 70px rgba(30, 22, 15, 0.24)", padding: 24 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 18, marginBottom: 18 }}>
          <div>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 6 }}>{isEn ? "PREVIEW BEFORE SAVING" : "确认前预览"}</div>
            <div className="ei-serif" style={{ fontSize: 23, color: T.ink }}>
              {isEn ? "Review the accepted rewrites, then choose where to save." : "确认采纳的改写结果，并选择保存方式。"}
            </div>
            <div style={{ fontSize: 13, color: T.ink3, marginTop: 6, lineHeight: 1.6 }}>
              {isEn
                ? "The original source stays read-only either way."
                : "无论选择哪种方式，原始来源都保持只读不变。"}
            </div>
          </div>
          <button onClick={onClose} style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", padding: 4 }}>
            <Icon name="x" size={16} />
          </button>
        </div>

        {/* Accepted rewrites preview */}
        <div style={{ border: `1px solid ${T.rule}`, background: T.bgSoft, borderRadius: 3, padding: 16, marginBottom: 16 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>
            {bullets.length} {isEn ? "ACCEPTED REWRITES" : "条已采纳改写"}
          </div>
          {bullets.map((b, i) => (
            <div key={b.id} style={{ padding: "10px 0", borderBottom: i < bullets.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
              <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", marginBottom: 4 }}>{b.section}</div>
              <div style={{ fontSize: 13.5, color: T.ink, lineHeight: 1.6 }}>{b.rewritten}</div>
            </div>
          ))}
        </div>

        {/* Save mode */}
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10, marginBottom: 18 }}>
          {[
            { k: "overwrite", icon: "resume", label: isEn ? "Overwrite this resume" : "覆盖原简历",
              desc: isEn ? `Apply rewrites into ${resume.name}. Original source unchanged.` : `把改写写入「${resume.name}」；原始来源不变。` },
            { k: "new", icon: "plus", label: isEn ? "Save as a new resume" : "保存为新简历",
              desc: isEn ? "Keep this resume as-is and add a rewritten copy to the list." : "保留这份简历不动，把改写结果作为新简历加入列表。" },
          ].map((m) => {
            const on = mode === m.k;
            return (
              <button key={m.k} onClick={() => setMode(m.k)} style={{
                textAlign: "left", padding: "14px 14px",
                background: on ? T.accentSoft : T.bg,
                border: `1px solid ${on ? T.accent : T.rule}`,
                borderRadius: 2, cursor: "pointer",
              }}>
                <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
                  <Icon name={m.icon} size={13} color={on ? T.accent : T.ink3} />
                  <div className="ei-label" style={{ color: on ? T.accent : T.ink3 }}>{m.label}</div>
                </div>
                <div style={{ fontSize: 12, color: T.ink2, lineHeight: 1.5 }}>{m.desc}</div>
              </button>
            );
          })}
        </div>

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 10 }}>
          <Btn T={T} variant="ghost" onClick={onClose}>{isEn ? "Cancel" : "取消"}</Btn>
          <Btn T={T} variant="accent" icon="check" onClick={() => onConfirm(mode)}>
            {mode === "new" ? (isEn ? "Save as new resume" : "保存为新简历") : (isEn ? "Overwrite this resume" : "覆盖原简历")}
          </Btn>
        </div>
      </div>
    </div>
  );
};

// ─── EDIT TAB (manual field edit, lightweight) ───────────────────────
const ResumeEditTab = ({ T, lang, resume }) => {
  const isEn = lang === "en";
  const [headline, setHeadline] = React.useState(isEn ? "Senior Frontend Engineer · Frontend platform & checkout" : "资深前端工程师 · 前端平台 & 结账链路");
  const [summary, setSummary] = React.useState(isEn
    ? "Six years of frontend, last three on platform tooling. Comfortable owning architecture from spec to ship."
    : "六年前端，最近三年在平台工程方向。能从架构方案到上线一手承担。");
  const saveChanges = () => {
    resumeNotify(lang, `已保存 ${resume.name} 的结构化字段`, `Saved structured fields for ${resume.name}`);
  };

  const sections = [
    { id: "exp", title: isEn ? "Experience" : "工作经历", items: [
      { role: isEn ? "Senior Frontend · Star-Ring" : "资深前端 · 星环科技", date: "2022 — now", body: isEn ? "Checkout RSC migration; Design System v1 rollout across 5 products." : "结账 RSC 迁移；Design System v1 覆盖 5 个产品。" },
      { role: isEn ? "Frontend · Lumen" : "前端 · Lumen", date: "2019 — 2022", body: isEn ? "14 features into the order-management core; built the batch-edit surface." : "14 个订单管理功能；建立批量编辑模块。" },
    ] },
    { id: "skills", title: isEn ? "Skills" : "技能", items: [
      { role: "—", date: "", body: "React · TypeScript · Performance · Design System · Platform Engineering" },
    ] },
  ];

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "10px 14px", marginBottom: 16, background: T.bgSoft, border: `1px dotted ${T.rule}`, borderRadius: 2 }}>
        <div style={{ fontSize: 13, color: T.ink3 }}>
          <Icon name="edit" size={12} style={{ marginRight: 6 }} />
          {isEn
            ? <>Editing <strong>{resume.name}</strong>. Changes save into this resume; the original source stays read-only.</>
            : <>正在编辑 <strong>{resume.name}</strong>。改动保存到这份简历；原始来源保持只读。</>
          }
        </div>
        <Btn T={T} variant="accent" size="sm" icon="check" onClick={saveChanges}>{isEn ? "Save changes" : "保存改动"}</Btn>
      </div>

      <Card T={T} pad={0}>
        <div style={{ padding: "20px 24px", borderBottom: `1px solid ${T.rule}` }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{isEn ? "HEADLINE" : "一句话标题"}</div>
          <input value={headline} onChange={(e) => setHeadline(e.target.value)} style={{
            width: "100%", padding: "10px 12px", border: `1px solid ${T.rule}`, borderRadius: 2,
            background: T.bg, color: T.ink, fontSize: 16, fontFamily: "var(--ei-serif)", outline: "none",
          }} />
          <div className="ei-label" style={{ color: T.ink3, margin: "16px 0 8px" }}>{isEn ? "SUMMARY" : "简介"}</div>
          <textarea value={summary} onChange={(e) => setSummary(e.target.value)} style={{
            width: "100%", minHeight: 80, padding: "10px 12px", border: `1px solid ${T.rule}`, borderRadius: 2,
            background: T.bg, color: T.ink, fontSize: 13.5, lineHeight: 1.6, resize: "vertical", outline: "none",
          }} />
        </div>

        {sections.map((sec, i) => (
          <div key={sec.id} style={{ padding: "20px 24px", borderBottom: i < sections.length - 1 ? `1px solid ${T.rule}` : "none" }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
              <div className="ei-label" style={{ color: T.ink3 }}>{sec.title.toUpperCase()}</div>
              <button style={{ background: "transparent", border: `1px solid ${T.rule}`, borderRadius: 2, padding: "4px 10px", fontSize: 12, color: T.ink3, cursor: "pointer", fontFamily: "var(--ei-sans)" }}>
                <Icon name="plus" size={11} style={{ marginRight: 4 }} /> {isEn ? "Add" : "新增"}
              </button>
            </div>
            {sec.items.map((it, idx) => (
              <div key={idx} style={{ padding: "12px 0", borderBottom: idx < sec.items.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 10, marginBottom: 4 }}>
                  <div style={{ fontSize: 14, color: T.ink, fontWeight: 500 }}>{it.role}</div>
                  {it.date && <div style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{it.date}</div>}
                </div>
                <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.6 }}>{it.body}</div>
              </div>
            ))}
          </div>
        ))}
      </Card>
    </div>
  );
};

// ─── Dynamic parse flow — streams agent steps after submit ───────────
const ResumeParseFlow = ({ T, lang, sourceLabel, onDone, onBack }) => {
  const steps = lang === "en" ? [
    { k: "extract", label: "Extracting raw text from source…" },
    { k: "identity", label: "Detecting personal info · name · contact · location" },
    { k: "experience", label: "Parsing work experience · 2 employers · 5 years" },
    { k: "projects", label: "Identifying flagship projects · 4 candidates" },
    { k: "skills", label: "Aggregating skills · stack · evidence links" },
    { k: "education", label: "Extracting education · degrees · certifications" },
    { k: "structure", label: "Drafting structured resume · ready to confirm" },
  ] : [
    { k: "extract", label: "提取原始文本…" },
    { k: "identity", label: "识别个人信息 · 姓名 · 联系方式 · 城市" },
    { k: "experience", label: "解析工作经历 · 2 家公司 · 5 年" },
    { k: "projects", label: "识别代表项目 · 候选 4 项" },
    { k: "skills", label: "聚合技能 · 技术栈 · 证据链接" },
    { k: "education", label: "提取教育背景 · 学位 · 认证" },
    { k: "structure", label: "生成结构化简历 · 准备进入预览" },
  ];

  const [active, setActive] = React.useState(0);

  React.useEffect(() => {
    if (active >= steps.length) {
      const t = setTimeout(() => onDone && onDone(), 600);
      return () => clearTimeout(t);
    }
    const t = setTimeout(() => setActive((a) => a + 1), 700);
    return () => clearTimeout(t);
  }, [active, steps.length, onDone]);

  return (
    <div className="ei-fadein" style={{ maxWidth: 1220, margin: "0 auto", padding: "40px 48px 96px" }}>
      <button onClick={onBack} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 22, display: "flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {lang === "en" ? "Cancel and edit input" : "取消并返回修改"}
      </button>

      <div style={{ marginBottom: 26 }}>
        <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{lang === "en" ? "PARSING SOURCE" : "解析原始内容"}</div>
        <h1 className="ei-serif" style={{ fontSize: 34, margin: 0, color: T.ink, letterSpacing: "-0.02em", lineHeight: 1.2 }}>
          {lang === "en" ? "Reading your source and drafting a structured resume." : "正在阅读你的原始内容，生成结构化简历草稿。"}
        </h1>
        <div style={{ fontSize: 13.5, color: T.ink3, marginTop: 10, fontFamily: "var(--ei-mono)" }}>
          <span style={{ color: T.ink2 }}>source · </span>{sourceLabel}
        </div>
      </div>

      <div style={{ background: T.bgCard, border: `1px solid ${T.accent}`, borderLeft: `3px solid ${T.accent}`, padding: "20px 26px", borderRadius: 2 }}>
        <div className="ei-label" style={{ color: T.accent, marginBottom: 14 }}>
          <span style={{ display: "inline-block", width: 7, height: 7, borderRadius: 4, background: T.accent, marginRight: 8, verticalAlign: "middle", animation: "pulse 1.2s ease-in-out infinite" }} />
          {lang === "en" ? "AGENT PARSING" : "AGENT 解析中"}
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 10, fontSize: 13, fontFamily: "var(--ei-mono)" }}>
          {steps.map((s, i) => {
            const isActive = i === active;
            const isDone = i < active;
            const isPending = i > active;
            return (
              <div key={s.k} style={{
                display: "flex", alignItems: "center", gap: 10,
                color: isPending ? T.ink4 : (isDone ? T.ink3 : T.ink),
                opacity: isPending ? 0.5 : 1,
                transition: "color .25s, opacity .25s",
              }}>
                <span style={{ width: 14, color: isDone ? T.ok : (isActive ? T.accent : T.ink4), display: "inline-flex" }}>
                  {isDone ? <Icon name="check" size={12} /> : (isActive ? "→" : "·")}
                </span>
                <span>{s.label}</span>
              </div>
            );
          })}
        </div>
      </div>

      <style>{`@keyframes pulse { 0%,100% { opacity: 1 } 50% { opacity: .35 } }`}</style>
    </div>
  );
};

// ─── Preview & confirm — shows parsed structured resume ─────────────
const ResumePreviewConfirm = ({ T, lang, sourceLabel, onConfirm, onBack }) => {
  const isEn = lang === "en";
  const draft = {
    name: isEn ? "Liu Zhe" : "刘哲",
    title: isEn ? "Senior Frontend Engineer" : "资深前端工程师",
    location: isEn ? "Shanghai, CN" : "中国 · 上海",
    contact: ["liuzhe@example.com", "+86 13800000000"],
    summary: isEn
      ? "5 years of frontend platform work — checkout flows, design systems, performance. Comfortable owning architecture and mentoring."
      : "5 年前端平台经验，覆盖结账流程、设计系统、性能优化；可独立负责架构方向并带新人。",
    experience: [
      {
        co: isEn ? "Star-Ring Technology" : "星环科技",
        role: isEn ? "Senior Frontend Engineer" : "资深前端工程师",
        period: "2022 — " + (isEn ? "now" : "至今"),
        bullets: [
          isEn ? "Owned the checkout surface migration to RSC + selective hydration." : "主导结账链路迁移到 RSC + 选择性注水。",
          isEn ? "Built shared UI components library used by 7 internal products." : "搭建被 7 个内部产品复用的通用 UI 组件库。",
          isEn ? "Mentored 3 mid-level engineers, ran weekly architecture reviews." : "带教 3 名中级工程师，主持每周架构评审。",
        ],
      },
      {
        co: "Lumen",
        role: isEn ? "Frontend Engineer" : "前端工程师",
        period: "2019 — 2022",
        bullets: [
          isEn ? "Shipped admin tooling for the data-team's labeling pipeline." : "为数据团队标注流水线交付后台工具。",
          isEn ? "Drove migration from Webpack to Vite — build time 84s → 11s." : "推动构建从 Webpack 迁到 Vite，构建耗时 84s → 11s。",
        ],
      },
    ],
    projects: [
      { name: isEn ? "Checkout RSC migration" : "结账 RSC 迁移", note: isEn ? "Reduced LCP 3.2s → 1.4s" : "LCP 3.2s → 1.4s" },
      { name: isEn ? "Design system v3" : "设计系统 v3", note: isEn ? "Adopted by 7 products" : "被 7 个产品采用" },
    ],
    skills: ["React 18 / RSC", "TypeScript", "Vite / Rspack", "Node.js", isEn ? "Design systems" : "设计系统", isEn ? "Performance" : "性能优化"],
    education: [{ school: isEn ? "Tongji University" : "同济大学", degree: isEn ? "B.Eng. Computer Science · 2015–2019" : "计算机科学与技术 学士 · 2015–2019" }],
  };

  const Section = ({ title, children }) => (
    <div style={{ paddingBottom: 18, marginBottom: 18, borderBottom: `1px dotted ${T.rule}` }}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{title}</div>
      {children}
    </div>
  );

  return (
    <div className="ei-fadein" style={{ maxWidth: 1220, margin: "0 auto", padding: "40px 48px 96px" }}>
      <button onClick={onBack} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 22, display: "flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {isEn ? "Re-parse from input" : "回到上一步重新解析"}
      </button>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", gap: 28, marginBottom: 26, flexWrap: "wrap" }}>
        <div>
          <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{isEn ? "PREVIEW · CONFIRM TO SAVE" : "预览 · 确认后保存"}</div>
          <h1 className="ei-serif" style={{ fontSize: 34, margin: 0, color: T.ink, letterSpacing: "-0.02em", lineHeight: 1.2 }}>
            {isEn ? "Here's the structured draft. Confirm to save it as a new resume." : "结构化草稿如下，确认后保存为一份新简历。"}
          </h1>
          <div style={{ fontSize: 13, color: T.ink3, marginTop: 10, fontFamily: "var(--ei-mono)" }}>
            <span style={{ color: T.ink2 }}>source · </span>{sourceLabel} <span style={{ color: T.ink4, margin: "0 8px" }}>·</span> <span style={{ color: T.ok }}>{isEn ? "parsed" : "已解析"}</span>
          </div>
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          <Btn T={T} variant="ghost" onClick={onBack}>{isEn ? "Back" : "返回上一步"}</Btn>
          <Btn T={T} variant="accent" icon="check" onClick={() => onConfirm && onConfirm(sourceLabel)}>{isEn ? "Confirm and save" : "确认并保存"}</Btn>
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 320px", gap: 22, alignItems: "start" }}>
        <Card T={T}>
          {/* Identity header */}
          <div style={{ paddingBottom: 18, marginBottom: 18, borderBottom: `1px solid ${T.rule}` }}>
            <div className="ei-serif" style={{ fontSize: 30, color: T.ink, letterSpacing: "-0.018em" }}>{draft.name}</div>
            <div style={{ fontSize: 14, color: T.ink2, marginTop: 4 }}>{draft.title} · {draft.location}</div>
            <div style={{ fontSize: 12.5, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 6 }}>{draft.contact.join("  ·  ")}</div>
          </div>

          <Section title={isEn ? "SUMMARY" : "个人简介"}>
            <div style={{ fontSize: 14, color: T.ink2, lineHeight: 1.65 }}>{draft.summary}</div>
          </Section>

          <Section title={isEn ? "EXPERIENCE" : "工作经历"}>
            {draft.experience.map((e, i) => (
              <div key={i} style={{ marginBottom: i < draft.experience.length - 1 ? 16 : 0 }}>
                <div style={{ display: "flex", justifyContent: "space-between", gap: 12, marginBottom: 6 }}>
                  <div>
                    <div style={{ fontSize: 14.5, color: T.ink, fontWeight: 500 }}>{e.role}</div>
                    <div style={{ fontSize: 13, color: T.ink2 }}>{e.co}</div>
                  </div>
                  <div style={{ fontSize: 12.5, color: T.ink3, fontFamily: "var(--ei-mono)", whiteSpace: "nowrap" }}>{e.period}</div>
                </div>
                <ul style={{ paddingLeft: 18, margin: "6px 0 0", color: T.ink2, fontSize: 13.5, lineHeight: 1.7 }}>
                  {e.bullets.map((b, j) => <li key={j}>{b}</li>)}
                </ul>
              </div>
            ))}
          </Section>

          <Section title={isEn ? "FLAGSHIP PROJECTS" : "代表项目"}>
            {draft.projects.map((p, i) => (
              <div key={i} style={{ display: "flex", justifyContent: "space-between", padding: "8px 0", borderBottom: i < draft.projects.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
                <div style={{ fontSize: 13.5, color: T.ink }}>{p.name}</div>
                <div style={{ fontSize: 12.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{p.note}</div>
              </div>
            ))}
          </Section>

          <Section title={isEn ? "SKILLS" : "技能"}>
            <div style={{ display: "flex", flexWrap: "wrap", gap: 6 }}>
              {draft.skills.map((s) => (
                <Tag key={s} T={T} tone="neutral">{s}</Tag>
              ))}
            </div>
          </Section>

          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{isEn ? "EDUCATION" : "教育"}</div>
            {draft.education.map((ed, i) => (
              <div key={i} style={{ fontSize: 13.5, color: T.ink2 }}>
                <span style={{ color: T.ink, fontWeight: 500 }}>{ed.school}</span> · {ed.degree}
              </div>
            ))}
          </div>
        </Card>

        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <Card T={T}>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>{isEn ? "WHAT WILL BE SAVED" : "确认后保存什么"}</div>
            {[
              { icon: "file", t: isEn ? "Original source" : "原始来源", b: isEn ? "Kept untouched · always traceable." : "原始内容保持不变，可追溯。" },
              { icon: "resume", t: isEn ? "Structured resume" : "结构化简历", b: isEn ? "The draft above · editable later." : "上面的草稿 · 后续可编辑。" },
            ].map((it, i) => (
              <div key={it.t} style={{ display: "grid", gridTemplateColumns: "26px 1fr", gap: 10, padding: "12px 0", borderBottom: i < 1 ? `1px dotted ${T.rule}` : "none" }}>
                <Icon name={it.icon} size={15} color={T.accent} style={{ marginTop: 2 }} />
                <div>
                  <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>{it.t}</div>
                  <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.5, marginTop: 2 }}>{it.b}</div>
                </div>
              </div>
            ))}
          </Card>

          <Card T={T}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{isEn ? "PARSE NOTES" : "解析备注"}</div>
            <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.6 }}>
              {isEn
                ? "Anything missing or wrong can be fixed in the structured editor after you save."
                : "如有遗漏或不准确的内容，保存后可在结构化编辑器里继续调整。"}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
};

// ─── Create flow — upload or paste only (guided Q&A removed, D-20) ──
const ResumeCreateFlow = ({ T, lang, nav, onBack, onCreateResume }) => {
  const [stage, setStage] = React.useState("input"); // "input" | "parsing" | "preview"
  const [createMode, setCreateMode] = React.useState("upload");
  const [resumeText, setResumeText] = React.useState("");
  const [pickedFile, setPickedFile] = React.useState(null);

  const sourceLabel = createMode === "upload"
    ? (pickedFile ? pickedFile.name : (lang === "en" ? "Uploaded file" : "上传文件"))
    : (lang === "en" ? "Pasted text" : "粘贴文本");

  const handleSubmit = () => setStage("parsing");

  if (stage === "parsing") {
    return <ResumeParseFlow T={T} lang={lang} sourceLabel={sourceLabel} onDone={() => setStage("preview")} onBack={() => setStage("input")} />;
  }
  if (stage === "preview") {
    return <ResumePreviewConfirm T={T} lang={lang} sourceLabel={sourceLabel} onConfirm={(label) => onCreateResume ? onCreateResume(label) : onBack()} onBack={() => setStage("input")} />;
  }

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
          {lang === "en" ? "We keep the original source, parse it into a structured resume, and save both as one resume you can revise later." : "系统会保留原始文件或原始文本，同时解析成结构化简历，并作为一份可回溯的简历保存。"}
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
                  {lang === "en" ? "The source file is stored as the read-only original. Parsed sections become your editable structured resume." : "原始文件会作为只读原件保存，解析出的工作经历、项目、技能和教育经历会进入可编辑结构化简历。"}
                </div>
                <Btn T={T} variant="accent" icon="upload" onClick={() => {
                  const input = document.createElement("input");
                  input.type = "file";
                  input.accept = ".pdf,.docx,.md,.txt";
                  input.onchange = (e) => {
                    const f = e.target.files && e.target.files[0];
                    if (f) {
                      setPickedFile(f);
                      handleSubmit();
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
                  {lang === "en" ? "Raw text is retained, then parsed into structured sections." : "原始文本会保留，并解析成结构化段落。"}
                </div>
                <Btn T={T} variant="accent" icon="sparkle" disabled={!resumeText.trim()} onClick={handleSubmit}>{lang === "en" ? "Parse and save" : "解析并保存"}</Btn>
              </div>
            </div>
          )}
        </Card>

        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <Card T={T}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "WHAT GETS SAVED" : "会保存什么"}</div>
            {[
              { icon: "file", title: lang === "en" ? "Original source" : "原始来源", body: lang === "en" ? "File or pasted text stays traceable." : "文件或粘贴文本都会保留来源。" },
              { icon: "resume", title: lang === "en" ? "Structured resume" : "结构化简历", body: lang === "en" ? "Work, projects, skills, education, and evidence are editable." : "工作经历、项目、技能、教育和证据点可编辑。" },
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
                ? <>After you submit, we'll <span style={{ color: T.ink, fontWeight: 500 }}>parse the source live</span>, then show you a <span style={{ color: T.ink, fontWeight: 500 }}>preview to confirm</span> before saving.</>
                : <>提交之后会<span style={{ color: T.ink, fontWeight: 500 }}>动态解析原始内容</span>，然后进入<span style={{ color: T.ink, fontWeight: 500 }}>预览确认页</span>，确认后才保存。</>}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
};

// ─── Original source preview modal ───────────────────────────────────
const OriginalResumePreviewModal = ({ T, lang, resume, onClose }) => {
  const isEn = lang === "en";
  const [view, setView] = React.useState("file");
  const sourceLines = resume && Array.isArray(resume.text) ? resume.text : [];
  return (
    <div style={{ position: "fixed", inset: 0, background: "rgba(24, 20, 16, 0.24)", zIndex: 80, display: "flex", alignItems: "center", justifyContent: "center", padding: 24 }} onClick={onClose}>
      <div className="ei-fadein" onClick={(e) => e.stopPropagation()} style={{ width: "min(960px, 100%)", maxHeight: "88vh", overflow: "auto", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4, boxShadow: "0 24px 70px rgba(30, 22, 15, 0.24)" }}>
        <div style={{ padding: "20px 24px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 18 }}>
          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{isEn ? "ORIGINAL SOURCE PREVIEW" : "原始来源预览"}</div>
            <div className="ei-serif" style={{ fontSize: 24, color: T.ink }}>{resume.sourceName}</div>
            <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 6 }}>
              {resume.sourceType === "paste" ? (isEn ? "Pasted text" : "粘贴文本") : (isEn ? "Uploaded file" : "上传文件")} · {resume.createdAt} · {resume.summary}
            </div>
          </div>
          <button onClick={onClose} style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", padding: 4 }}>
            <Icon name="x" size={16} />
          </button>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "250px 1fr", minHeight: 520 }}>
          <div style={{ borderRight: `1px solid ${T.rule}`, padding: 18, background: T.bgSoft }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{isEn ? "SOURCE RELATION" : "来源关系"}</div>
            {[
              [isEn ? "Original source" : "原始来源", resume.sourceName],
              [isEn ? "Parsed into" : "解析为", resume.name],
            ].map(([k, v], i) => (
              <div key={k} style={{ padding: "10px 0", borderBottom: i < 1 ? `1px dotted ${T.rule}` : "none" }}>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{k}</div>
                <div style={{ fontSize: 13, color: T.ink, lineHeight: 1.45 }}>{v}</div>
              </div>
            ))}
            <div style={{ marginTop: 16, fontSize: 12.5, color: T.ink3, lineHeight: 1.6 }}>
              {isEn
                ? "The original is read-only. Edits and accepted rewrites change the structured resume, never this source."
                : "原始来源只读保存。编辑和改写采纳只改变结构化简历，不覆盖原件。"}
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 8, marginTop: 18 }}>
              {[
                ["file", isEn ? "Original file" : "原始文件"],
                ["text", isEn ? "Parsed text" : "解析文本"],
              ].map(([k, label]) => (
                <button key={k} onClick={() => setView(k)} style={{
                  textAlign: "left",
                  border: `1px solid ${view === k ? T.accent : T.rule}`,
                  background: view === k ? T.accentSoft : T.bgCard,
                  color: view === k ? T.ink : T.ink2,
                  borderRadius: 2,
                  padding: "9px 10px",
                  cursor: "pointer",
                  fontFamily: "var(--ei-sans)",
                  fontSize: 13,
                }}>
                  {label}
                </button>
              ))}
            </div>
          </div>

          <div style={{ padding: 24, background: T.bg }}>
            {view === "file" ? (
              <div style={{ maxWidth: 560, minHeight: 620, margin: "0 auto", background: "#fff", color: "#222", border: `1px solid ${T.rule}`, boxShadow: "0 18px 50px rgba(30, 22, 15, 0.12)", padding: 36, fontFamily: "Georgia, serif" }}>
                <div style={{ fontSize: 25, fontWeight: 600, marginBottom: 4 }}>{sourceLines[0] || resume.sourceName}</div>
                <div style={{ fontSize: 12, color: "#666", marginBottom: 18 }}>{resume.sourceName} · {resume.createdAt}</div>
                <div style={{ height: 1, background: "#333", marginBottom: 18 }} />
                <div style={{ fontSize: 11, color: "#888", letterSpacing: "0.12em", textTransform: "uppercase", marginBottom: 8 }}>Experience</div>
                {sourceLines.slice(1).map((line, i) => (
                  <div key={i} style={{ fontSize: i === 0 ? 15 : 13.5, fontWeight: i === 0 ? 600 : 400, color: i === 0 ? "#222" : "#444", lineHeight: 1.75, marginBottom: 6 }}>
                    {line}
                  </div>
                ))}
                <div style={{ marginTop: 22, fontSize: 11, color: "#888", letterSpacing: "0.12em", textTransform: "uppercase", marginBottom: 8 }}>Skills</div>
                <div style={{ fontSize: 13.5, color: "#444", lineHeight: 1.75 }}>React · TypeScript · Performance · Design System · Platform Engineering</div>
              </div>
            ) : (
              <div style={{ border: `1px solid ${T.rule}`, background: T.bgCard, borderRadius: 3, padding: 18 }}>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{isEn ? "PARSED TEXT SNAPSHOT" : "解析文本快照"}</div>
                {sourceLines.map((line, i) => (
                  <div key={i} style={{ padding: "10px 0", borderBottom: i < sourceLines.length - 1 ? `1px dotted ${T.rule}` : "none", fontSize: 13.5, color: T.ink2, lineHeight: 1.6 }}>
                    {line}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

window.ResumeVersionsScreen = ResumeWorkshopScreen;
