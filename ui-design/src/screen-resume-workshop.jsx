// Resume workshop — redesigned IA: list (tree-grouped originals) → detail (preview / rewrites / edit)
// Replaces the legacy ResumeVersionsScreen exported from screens-p1-depth.jsx.

const buildResumeData = (lang) => {
  const isEn = lang === "en";
  const originals = [
    {
      id: "src-cn",
      name: isEn ? "LiuZhe_Frontend_CN_2026.pdf" : "刘哲_前端_中文_2026.pdf",
      langTag: isEn ? "CN" : "中",
      type: isEn ? "Uploaded PDF" : "上传 PDF",
      createdAt: "2026-04-18",
      status: "active",
      summary: isEn ? "Frontend platform · checkout · design system" : "前端平台 · 结账 · 设计系统",
      text: [
        isEn ? "Liu Zhe · Senior Frontend Engineer · Shanghai" : "刘哲 · 资深前端工程师 · 上海",
        isEn ? "Star-Ring · Senior Frontend · 2022-now" : "星环科技 · 资深前端 · 2022 至今",
        isEn ? "Worked on checkout performance and complex admin surfaces." : "负责结账流程性能改进和复杂后台系统建设。",
        isEn ? "Built shared UI components for internal products." : "为内部产品建设通用 UI 组件。",
      ],
    },
    {
      id: "src-en",
      name: "LiuZhe_Frontend_EN_2026.pdf",
      langTag: "EN",
      type: isEn ? "Uploaded PDF" : "上传 PDF",
      createdAt: "2026-04-15",
      status: "active",
      summary: isEn ? "Frontend platform · TypeScript · APAC/US" : "前端平台 · TypeScript · 跨时区协作",
      text: [
        "Liu Zhe · Frontend Platform Engineer",
        "Built platform tooling, design system infrastructure, and TypeScript foundations.",
        "Worked with distributed teams across APAC and US time zones.",
      ],
    },
    {
      id: "src-fullstack",
      name: isEn ? "LiuZhe_Fullstack_2024.pdf" : "刘哲_全栈方向_2024.pdf",
      langTag: isEn ? "CN" : "中",
      type: isEn ? "Uploaded PDF" : "上传 PDF",
      createdAt: "2024-08-12",
      status: "archived",
      summary: isEn ? "Earlier draft · Node + React fullstack" : "早期版本 · Node + React 全栈方向",
      text: [
        isEn ? "Earlier fullstack-direction resume — kept for history." : "更早的全栈方向简历，保留用于回溯。",
      ],
    },
  ];

  const versions = [
    // src-cn tree
    { id: "v-cn-master", originalId: "src-cn", name: isEn ? "Master" : "主版本", tag: "MASTER", date: "2026-04-18", target: isEn ? "General" : "通用", bullets: 24, accepted: 0 },
    { id: "v-cn-bd",     originalId: "src-cn", name: isEn ? "ByteDance · Frontend Platform" : "字节跳动 · 前端平台", tag: "TARGETED", date: "2026-04-20", target: isEn ? "ByteDance" : "字节跳动",   bullets: 22, accepted: 4, match: 82 },
    { id: "v-cn-mt",     originalId: "src-cn", name: isEn ? "Meituan · Web Infra" : "美团 · Web 基建",       tag: "TARGETED", date: "2026-04-19", target: isEn ? "Meituan" : "美团",     bullets: 20, accepted: 6, match: 76 },
    // src-en tree
    { id: "v-en-master", originalId: "src-en", name: isEn ? "Master" : "主版本", tag: "MASTER", date: "2026-04-15", target: isEn ? "General" : "通用", bullets: 22, accepted: 0 },
    { id: "v-en-stripe", originalId: "src-en", name: "Stripe · SWE",          tag: "TARGETED", date: "2026-04-17", target: "Stripe",  bullets: 21, accepted: 5, match: 71 },
    // src-fullstack tree (archived, no customizations)
    { id: "v-fs-master", originalId: "src-fullstack", name: isEn ? "Master (archived)" : "主版本（已归档）", tag: "MASTER", date: "2024-08-12", target: isEn ? "General" : "通用", bullets: 18, accepted: 0, archived: true },
  ];

  return { originals, versions };
};

const buildBullets = (lang, version) => {
  const isEn = lang === "en";
  const sectionA = isEn ? "Senior Frontend · Star-Ring · 2022-now" : "资深前端 · 星环科技 · 2022 至今";
  const sectionB = isEn ? "Frontend · Lumen · 2019-2022" : "前端 · Lumen · 2019-2022";
  // version-specific opening bullet so different versions feel distinct
  const opener = version && version.target ? (isEn ? `Targeted for ${version.target}.` : `面向「${version.target}」的定制建议。`) : "";
  void opener;
  return [
    {
      id: "b1", section: sectionA,
      original: isEn ? "Worked on checkout performance improvements for the e-commerce team, collaborating closely with backend engineers." : "负责电商团队结账流程的性能改进工作，与后端工程师紧密协作。",
      rewritten: isEn ? "Led migration of the checkout surface to RSC + selective hydration, cutting LCP from 3.2s to 1.4s and lifting quarterly GMV by 1.8M (8% → 4.2% abandon)." : "主导结账链路迁移到 RSC + 选择性注水，LCP 3.2s → 1.4s，流失率 8% → 4.2%，季度 GMV +180 万。",
      why: isEn ? ["Weak → strong ownership verb", "Adds quantified impact", "Names the specific architecture"] : ["动词从弱到强：「负责」→「主导」", "加入量化影响", "具体指出架构选择"],
      status: "pending",
    },
    {
      id: "b2", section: sectionA,
      original: isEn ? "Rolled out a design system across multiple product teams." : "在多个产品团队推广了设计系统。",
      rewritten: isEn ? "Drove Design System v1 adoption across 5 products in 6 months (4 live, 1 in progress) — ran 3 workshops, paired migrations with 2 pilot teams, reduced new-dev ramp ~50%." : "6 个月内推动 Design System v1 在 5 个产品落地（4 上线、1 进行中）——办 3 次推广会、与 2 个试点团队结对迁移，新人上手时间缩短约 50%。",
      why: isEn ? ["Names the scale (5 products)", "Shows method, not just outcome", "Anchored on developer time saved"] : ["量化范围：5 个产品", "讲方法而不只是结果", "以节省的工时收口"],
      status: "accepted",
    },
    {
      id: "b3", section: sectionB,
      original: isEn ? "Built and shipped various features for the core product." : "为核心产品构建并交付了多个功能。",
      rewritten: isEn ? "Shipped 14 features to the order-management core over 3 years, including a batch-edit surface that became the #2 most-used power-user flow." : "3 年内为订单管理核心交付 14 个功能，其中批量编辑成为重度用户第 2 常用流程。",
      why: isEn ? ["Vague → specific count", "Picks one feature worth name-checking", "Usage data gives credibility"] : ["模糊数量变具体", "挑一个值得点名的功能", "用使用数据建立可信度"],
      status: "pending",
    },
    {
      id: "b4", section: sectionB,
      original: isEn ? "Participated in code reviews and technical discussions." : "参与代码评审和技术讨论。",
      rewritten: isEn ? "(remove — generic, dilutes stronger bullets)" : "（建议删除——太泛，会稀释其它更强的 bullet）",
      why: isEn ? ["Every senior does this", "Takes space from quantifiable wins", "Better implied than stated"] : ["资深都做这个", "占用了可以量化的空间", "隐含就好，别直说"],
      status: "rejected",
    },
  ];
};

// ─────────────────────────────────────────────────────────────────────
// Top-level screen: branches between list and detail
// ─────────────────────────────────────────────────────────────────────
const ResumeWorkshopScreen = ({ T, lang, nav, params = {} }) => {
  const isEn = lang === "en";
  const { originals, versions } = React.useMemo(() => buildResumeData(lang), [lang]);

  // create-flow / branch-flow stay inside this screen — driven by params + local state
  const [flow, setFlow] = React.useState(
    params.flow === "create" ? "create" :
    params.flow === "branch" ? "branch" : "list"
  );
  const [branchOriginalId, setBranchOriginalId] = React.useState(params.branchOriginalId || null);
  React.useEffect(() => {
    if (params.flow === "create") setFlow("create");
    if (params.flow === "branch") { setFlow("branch"); setBranchOriginalId(params.branchOriginalId || null); }
  }, [params.flow, params.branchOriginalId]);

  // detail navigation — versionId in params drives detail view
  const versionId = params.versionId;
  const detailVersion = versionId ? versions.find((v) => v.id === versionId) : null;

  if (flow === "create") {
    return <ResumeCreateFlow T={T} lang={lang} nav={nav} onBack={() => setFlow("list")} originals={originals} />;
  }

  if (flow === "branch") {
    const sourceOriginal = originals.find((o) => o.id === branchOriginalId) || originals[0];
    const sourceMaster = versions.find((v) => v.originalId === sourceOriginal.id && v.tag === "MASTER");
    return (
      <ResumeBranchFlow
        T={T} lang={lang}
        original={sourceOriginal}
        master={sourceMaster}
        onBack={() => setFlow("list")}
      />
    );
  }

  if (detailVersion) {
    return (
      <ResumeDetailView
        T={T}
        lang={lang}
        nav={nav}
        version={detailVersion}
        original={originals.find((o) => o.id === detailVersion.originalId)}
        siblings={versions.filter((v) => v.originalId === detailVersion.originalId)}
        initialTab={params.tab || "preview"}
      />
    );
  }

  return (
    <ResumeListView
      T={T}
      lang={lang}
      nav={nav}
      originals={originals}
      versions={versions}
      onCreate={() => setFlow("create")}
      onBranch={(originalId) => { setBranchOriginalId(originalId); setFlow("branch"); }}
    />
  );
};

// ─────────────────────────────────────────────────────────────────────
// LIST VIEW — answers: "what resume versions do I have?"
// ─────────────────────────────────────────────────────────────────────
const ResumeListView = ({ T, lang, nav, originals, versions, onCreate, onBranch }) => {
  const isEn = lang === "en";
  const [groupBy, setGroupBy] = React.useState("original"); // "original" | "flat"
  const [collapsed, setCollapsed] = React.useState({});
  const [selectedTreeId, setSelectedTreeId] = React.useState(null);

  const totals = {
    originals: originals.length,
    activeOriginals: originals.filter((o) => o.status === "active").length,
    versions: versions.length,
    targeted: versions.filter((v) => v.tag === "TARGETED").length,
  };

  const openVersion = (v, tab = "preview") => nav("resume_versions", { versionId: v.id, tab });

  return (
    <div className="ei-fadein" style={{ maxWidth: 1320, margin: "0 auto", padding: "40px 48px 96px" }}>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginBottom: 28, gap: 32, flexWrap: "wrap" }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
            {isEn ? "RESUME WORKSHOP" : "简历工坊"}
          </div>
          <h1 className="ei-serif" style={{ fontSize: 38, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.15 }}>
            {isEn ? "Your resumes — originals, masters, and targeted versions." : "你的简历 · 原始、主版本与岗位定制"}
          </h1>
          <div style={{ fontSize: 14, color: T.ink3, marginTop: 10, maxWidth: 720, lineHeight: 1.55 }}>
            {isEn
              ? "Each original file becomes one tree: a master parsed from it, plus targeted versions branched off for specific jobs. Open any version to preview, accept rewrites, or edit by hand."
              : "每份原始简历是一棵独立的树：解析得到一份主版本，再为不同岗位分叉出定制版本。打开任意版本进入预览、改写建议或手动编辑。"}
          </div>
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          <Btn T={T} variant="accent" size="sm" icon="plus" onClick={onCreate}>{isEn ? "New resume" : "新建简历"}</Btn>
        </div>
      </div>

      {/* Top stats strip */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 10, marginBottom: 22 }}>
        {[
          { k: isEn ? "Originals" : "原始简历", v: `${totals.activeOriginals} / ${totals.originals}`, sub: isEn ? "active / total" : "在用 / 总数" },
          { k: isEn ? "Versions" : "全部版本", v: `${totals.versions}`, sub: isEn ? `${totals.targeted} targeted` : `${totals.targeted} 个定制` },
          { k: isEn ? "Best match" : "最高匹配", v: "82%", sub: isEn ? "ByteDance · FE Platform" : "字节跳动 · 前端平台", tone: "ok" },
          { k: isEn ? "Last edit" : "最近编辑", v: "2026 · 04 · 20", sub: isEn ? "v · ByteDance" : "v · 字节跳动" },
        ].map((m) => (
          <div key={m.k} style={{ padding: "14px 16px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 5 }}>{m.k}</div>
            <div className="ei-serif" style={{ fontSize: 20, color: m.tone === "ok" ? T.ok : T.ink, letterSpacing: "-0.01em" }}>{m.v}</div>
            <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 3 }}>{m.sub}</div>
          </div>
        ))}
      </div>

      {/* View switcher */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 16, gap: 14, flexWrap: "wrap" }}>
        <div style={{ display: "flex", gap: 0, border: `1px solid ${T.rule}`, borderRadius: 2, overflow: "hidden", background: T.bgCard }}>
          {[
            { k: "original", label: isEn ? "Group by original" : "按原始分组", icon: "file" },
            { k: "flat",     label: isEn ? "Flat by version" : "按版本平铺",  icon: "layers" },
          ].map((m) => (
            <button key={m.k} onClick={() => setGroupBy(m.k)} style={{
              padding: "8px 14px", background: groupBy === m.k ? T.bgSoft : "transparent",
              color: groupBy === m.k ? T.ink : T.ink3, border: "none",
              borderRight: m.k === "original" ? `1px solid ${T.rule}` : "none",
              cursor: "pointer", fontFamily: "var(--ei-sans)", fontSize: 13,
              display: "flex", alignItems: "center", gap: 6,
            }}>
              <Icon name={m.icon} size={12} /> {m.label}
            </button>
          ))}
        </div>
        <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
          {groupBy === "original"
            ? (isEn ? "Manage your source files" : "管理底稿与分叉关系")
            : (isEn ? "Pick the right one to send" : "挑一份去投递")}
        </div>
      </div>

      {/* If a tree is selected, surface the inline create-version helper */}
      {groupBy === "original" && selectedTreeId && (
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "10px 14px", marginBottom: 12, background: T.accentSoft, border: `1px solid ${T.accent}`, borderRadius: 2 }}>
          <div style={{ fontSize: 13, color: T.ink2 }}>
            <Icon name="check" size={12} color={T.accent} style={{ marginRight: 6 }} />
            {isEn ? "Selected as the base for the next version" : "已选为下一个新版本的底稿"} ·{" "}
            <span style={{ fontFamily: "var(--ei-mono)", color: T.ink3 }}>
              {(originals.find((o) => o.id === selectedTreeId) || {}).name}
            </span>
          </div>
          <div style={{ display: "flex", gap: 8 }}>
            <button onClick={() => setSelectedTreeId(null)} style={{ background: "transparent", border: "none", color: T.ink3, cursor: "pointer", fontSize: 12 }}>
              {isEn ? "Clear" : "取消"}
            </button>
            <Btn T={T} variant="accent" size="sm" icon="plus" onClick={() => onBranch(selectedTreeId)}>{isEn ? "New version from this tree" : "基于这棵树新建版本"}</Btn>
          </div>
        </div>
      )}

      {groupBy === "original" ? (
        <ResumeTreeView
          T={T} lang={lang}
          originals={originals}
          versions={versions}
          collapsed={collapsed}
          setCollapsed={setCollapsed}
          selectedTreeId={selectedTreeId}
          setSelectedTreeId={setSelectedTreeId}
          onOpen={openVersion}
          onCreate={onCreate}
        />
      ) : (
        <ResumeFlatView
          T={T} lang={lang}
          originals={originals}
          versions={versions}
          onOpen={openVersion}
        />
      )}
    </div>
  );
};

// ─────────────────────────────────────────────────────────────────────
// TREE VIEW — by original
// ─────────────────────────────────────────────────────────────────────
const ResumeTreeView = ({ T, lang, originals, versions, collapsed, setCollapsed, selectedTreeId, setSelectedTreeId, onOpen, onCreate }) => {
  const isEn = lang === "en";
  const toggle = (id) => setCollapsed({ ...collapsed, [id]: !collapsed[id] });

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
      {originals.map((o) => {
        const tree = versions.filter((v) => v.originalId === o.id);
        const master = tree.find((v) => v.tag === "MASTER");
        const targeted = tree.filter((v) => v.tag === "TARGETED");
        const isCollapsed = !!collapsed[o.id];
        const isSelected = selectedTreeId === o.id;
        const archived = o.status === "archived";

        return (
          <div key={o.id} style={{
            border: `1px solid ${isSelected ? T.accent : T.rule}`,
            borderRadius: 3,
            background: T.bgCard,
            opacity: archived ? 0.7 : 1,
          }}>
            {/* Tree header — original file row */}
            <div style={{
              display: "grid",
              gridTemplateColumns: "20px 1fr auto",
              gap: 12,
              alignItems: "center",
              padding: "16px 18px",
              borderBottom: isCollapsed ? "none" : `1px solid ${T.rule}`,
              cursor: "pointer",
            }} onClick={() => toggle(o.id)}>
              <Icon name={isCollapsed ? "chevron_right" : "chevron_down"} size={12} color={T.ink3} />
              <div style={{ minWidth: 0 }}>
                <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4 }}>
                  <Icon name="file" size={14} color={T.accent} />
                  <span style={{ fontSize: 15, color: T.ink, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{o.name}</span>
                  <span style={{ fontFamily: "var(--ei-mono)", fontSize: 9, letterSpacing: "0.08em", padding: "1px 6px", borderRadius: 2, background: T.bgSoft, color: T.ink3 }}>{o.langTag}</span>
                  {archived && (
                    <span style={{ fontFamily: "var(--ei-mono)", fontSize: 9, letterSpacing: "0.08em", padding: "1px 6px", borderRadius: 2, background: T.bgSoft, color: T.ink4 }}>{isEn ? "ARCHIVED" : "已归档"}</span>
                  )}
                </div>
                <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
                  {o.type} · {o.createdAt} · <span style={{ color: T.ink2 }}>{o.summary}</span> · {tree.length} {isEn ? "versions" : "个版本"}
                </div>
              </div>
              <div style={{ display: "flex", gap: 6 }} onClick={(e) => e.stopPropagation()}>
                <button onClick={() => setSelectedTreeId(isSelected ? null : o.id)} style={{
                  padding: "5px 10px", fontSize: 12, cursor: "pointer",
                  background: isSelected ? T.accent : "transparent",
                  color: isSelected ? "#fff" : T.ink2,
                  border: `1px solid ${isSelected ? T.accent : T.rule}`, borderRadius: 2, fontFamily: "var(--ei-sans)",
                }}>
                  <Icon name={isSelected ? "check" : "plus"} size={11} style={{ marginRight: 4 }} />
                  {isSelected ? (isEn ? "Selected" : "已选") : (isEn ? "Use as base" : "选为底稿")}
                </button>
              </div>
            </div>

            {/* Tree body — versions */}
            {!isCollapsed && (
              <div style={{ padding: "6px 0" }}>
                {master && <ResumeVersionRow T={T} lang={lang} v={master} onOpen={onOpen} indent />}
                {targeted.map((v) => (
                  <ResumeVersionRow key={v.id} T={T} lang={lang} v={v} onOpen={onOpen} indent />
                ))}
              </div>
            )}
          </div>
        );
      })}

      {/* New original CTA */}
      <button onClick={onCreate} style={{
        padding: "18px 18px", background: "transparent",
        border: `1px dashed ${T.rule}`, borderRadius: 3, color: T.ink3,
        cursor: "pointer", fontFamily: "var(--ei-sans)", fontSize: 13.5,
        display: "flex", alignItems: "center", justifyContent: "center", gap: 8,
      }}>
        <Icon name="upload" size={14} /> {isEn ? "Upload another original (new tree)" : "上传另一份原始简历（新建一棵树）"}
      </button>
    </div>
  );
};

const ResumeVersionRow = ({ T, lang, v, onOpen, indent }) => {
  const isEn = lang === "en";
  const isMaster = v.tag === "MASTER";
  return (
    <div style={{
      display: "grid",
      gridTemplateColumns: indent ? "32px 1fr auto" : "1fr auto",
      gap: 12,
      alignItems: "center",
      padding: "11px 18px 11px 18px",
      borderBottom: `1px dotted ${T.rule}`,
    }}>
      {indent && (
        <div style={{ display: "flex", justifyContent: "center", color: T.ink4, fontFamily: "var(--ei-mono)", fontSize: 14 }}>
          └
        </div>
      )}
      <div style={{ display: "flex", alignItems: "center", gap: 10, minWidth: 0 }}>
        <Icon name={isMaster ? "resume" : "briefcase"} size={13} color={isMaster ? T.ink3 : T.accent} />
        <span style={{ fontSize: 14, color: T.ink, fontWeight: isMaster ? 400 : 500 }}>{v.name}</span>
        <span style={{ fontFamily: "var(--ei-mono)", fontSize: 9, letterSpacing: "0.08em", padding: "1px 6px", borderRadius: 2, background: isMaster ? T.bgSoft : T.accentSoft, color: isMaster ? T.ink3 : T.accent }}>{v.tag}</span>
        <span style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
          {v.date} · {v.bullets} {isEn ? "bullets" : "条 bullet"}
          {v.accepted > 0 && ` · ${v.accepted} ${isEn ? "accepted" : "已采纳"}`}
        </span>
      </div>
      <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
        {typeof v.match === "number" && (
          <span style={{
            fontFamily: "var(--ei-mono)", fontSize: 11, letterSpacing: "0.04em",
            padding: "2px 8px", borderRadius: 2,
            background: v.match >= 80 ? T.okSoft : v.match >= 70 ? T.accentSoft : T.bgSoft,
            color: v.match >= 80 ? T.ok : v.match >= 70 ? T.accent : T.ink3,
          }}>
            {isEn ? "MATCH" : "匹配"} {v.match}%
          </span>
        )}
        <button onClick={() => onOpen(v, "preview")} style={{
          padding: "5px 12px", fontSize: 12, cursor: "pointer",
          background: "transparent", color: T.ink2,
          border: `1px solid ${T.rule}`, borderRadius: 2, fontFamily: "var(--ei-sans)",
        }}>
          {isEn ? "Open" : "打开"} <Icon name="arrow_right" size={10} style={{ marginLeft: 4 }} />
        </button>
      </div>
    </div>
  );
};

// ─────────────────────────────────────────────────────────────────────
// FLAT VIEW — by version, sorted by match desc / date desc
// ─────────────────────────────────────────────────────────────────────
const ResumeFlatView = ({ T, lang, originals, versions, onOpen }) => {
  const isEn = lang === "en";
  const sorted = [...versions].sort((a, b) => {
    const aM = typeof a.match === "number" ? a.match : -1;
    const bM = typeof b.match === "number" ? b.match : -1;
    if (aM !== bM) return bM - aM;
    return (b.date || "").localeCompare(a.date || "");
  });
  return (
    <div style={{ border: `1px solid ${T.rule}`, borderRadius: 3, background: T.bgCard, overflow: "hidden" }}>
      <div style={{
        display: "grid",
        gridTemplateColumns: "1.6fr 1.4fr 1fr 0.8fr 1fr 100px",
        gap: 14, padding: "11px 18px",
        background: T.bgSoft, borderBottom: `1px solid ${T.rule}`,
        fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.06em", textTransform: "uppercase",
      }}>
        <div>{isEn ? "Version" : "版本"}</div>
        <div>{isEn ? "From original" : "来源原始"}</div>
        <div>{isEn ? "Target" : "目标岗位"}</div>
        <div>{isEn ? "Match" : "匹配"}</div>
        <div>{isEn ? "Last edit" : "最近编辑"}</div>
        <div></div>
      </div>
      {sorted.map((v, i) => {
        const o = originals.find((x) => x.id === v.originalId);
        const isMaster = v.tag === "MASTER";
        return (
          <div key={v.id} style={{
            display: "grid",
            gridTemplateColumns: "1.6fr 1.4fr 1fr 0.8fr 1fr 100px",
            gap: 14, padding: "13px 18px",
            borderBottom: i < sorted.length - 1 ? `1px dotted ${T.rule}` : "none",
            alignItems: "center",
            opacity: v.archived ? 0.6 : 1,
          }}>
            <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
              <Icon name={isMaster ? "resume" : "briefcase"} size={13} color={isMaster ? T.ink3 : T.accent} />
              <div style={{ minWidth: 0 }}>
                <div style={{ fontSize: 13.5, color: T.ink, fontWeight: isMaster ? 400 : 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{v.name}</div>
                <div style={{ fontSize: 10.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{v.tag}</div>
              </div>
            </div>
            <div style={{ fontSize: 12.5, color: T.ink2, fontFamily: "var(--ei-mono)", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
              {o ? o.name : "—"}
            </div>
            <div style={{ fontSize: 13, color: T.ink2 }}>{v.target}</div>
            <div style={{ fontSize: 12, fontFamily: "var(--ei-mono)", color: typeof v.match === "number" ? (v.match >= 80 ? T.ok : v.match >= 70 ? T.accent : T.ink3) : T.ink4 }}>
              {typeof v.match === "number" ? `${v.match}%` : "—"}
            </div>
            <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{v.date}</div>
            <button onClick={() => onOpen(v, "preview")} style={{
              padding: "5px 12px", fontSize: 12, cursor: "pointer",
              background: "transparent", color: T.ink2,
              border: `1px solid ${T.rule}`, borderRadius: 2, fontFamily: "var(--ei-sans)",
            }}>
              {isEn ? "Open" : "打开"}
            </button>
          </div>
        );
      })}
    </div>
  );
};

// ─────────────────────────────────────────────────────────────────────
// DETAIL VIEW — single version with tabs
// ─────────────────────────────────────────────────────────────────────
const ResumeDetailView = ({ T, lang, nav, version, original, siblings, initialTab }) => {
  const isEn = lang === "en";
  const [tab, setTab] = React.useState(initialTab || "preview");
  const isMaster = version.tag === "MASTER";

  const back = () => nav("resume_versions", {});

  const tabs = [
    { k: "preview",  label: isEn ? "Preview" : "预览",       icon: "file" },
    { k: "rewrites", label: isEn ? "Rewrites" : "改写建议", icon: "sparkle", disabled: isMaster, hint: isEn ? "Master stays pristine" : "主版本不参与改写" },
    { k: "edit",     label: isEn ? "Edit" : "手动编辑",     icon: "edit" },
  ];

  return (
    <div className="ei-fadein" style={{ maxWidth: 1320, margin: "0 auto", padding: "32px 48px 96px" }}>
      {/* Back + breadcrumb (this version's tree only) */}
      <button onClick={back} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 14, display: "flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {isEn ? "Back to resume workshop" : "返回简历工坊"}
      </button>

      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", gap: 28, marginBottom: 20, flexWrap: "wrap" }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>
            <span style={{ color: T.ink4 }}>{original ? original.name : "—"}</span>
            <span style={{ margin: "0 8px", color: T.ink4 }}>›</span>
            <span style={{ color: T.ink3 }}>{isEn ? "Versions" : "版本"}</span>
            <span style={{ margin: "0 8px", color: T.ink4 }}>›</span>
            <span style={{ color: T.ink2 }}>{version.name}</span>
          </div>
          <h1 className="ei-serif" style={{ fontSize: 32, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.2 }}>
            {version.name}
            <span style={{ fontFamily: "var(--ei-mono)", fontSize: 10, letterSpacing: "0.08em", padding: "3px 8px", marginLeft: 12, borderRadius: 2, background: isMaster ? T.bgSoft : T.accentSoft, color: isMaster ? T.ink3 : T.accent, verticalAlign: "middle" }}>{version.tag}</span>
          </h1>
          <div style={{ fontSize: 13, color: T.ink3, marginTop: 8, fontFamily: "var(--ei-mono)" }}>
            {isEn ? "Target" : "目标"}: {version.target} · {version.bullets} {isEn ? "bullets" : "条 bullet"}
            {typeof version.match === "number" && (
              <> · <span style={{ color: version.match >= 80 ? T.ok : T.accent }}>{isEn ? "Match" : "匹配"} {version.match}%</span></>
            )}
            {" · "}{version.date}
          </div>
        </div>
        <div style={{ display: "flex", gap: 8 }}>
          <Btn T={T} variant="secondary" size="sm" icon="file" onClick={() => setTab("preview")}>{isEn ? "Export PDF" : "导出 PDF"}</Btn>
          {!isMaster && (
            <Btn T={T} variant="ghost" size="sm" icon="layers">{isEn ? "Duplicate" : "复制为新版本"}</Btn>
          )}
        </div>
      </div>

      {/* Tree-scoped source map: only THIS branch */}
      <ResumeBranchMap T={T} lang={lang} original={original} siblings={siblings} version={version} />

      {/* Tabs */}
      <div style={{ display: "flex", gap: 0, marginTop: 24, marginBottom: 22, borderBottom: `1px solid ${T.rule}` }}>
        {tabs.map((t) => {
          const active = tab === t.k;
          return (
            <button key={t.k}
              onClick={() => !t.disabled && setTab(t.k)}
              title={t.disabled ? t.hint : ""}
              style={{
                padding: "13px 20px", background: "transparent", border: "none",
                borderBottom: `2px solid ${active ? T.accent : "transparent"}`,
                color: t.disabled ? T.ink4 : active ? T.ink : T.ink3,
                cursor: t.disabled ? "not-allowed" : "pointer",
                fontFamily: "var(--ei-sans)", fontSize: 14,
                display: "flex", alignItems: "center", gap: 8, marginBottom: -1,
              }}>
              <Icon name={t.icon} size={13} /> {t.label}
              {t.disabled && <span style={{ fontSize: 10.5, color: T.ink4, fontFamily: "var(--ei-mono)" }}>· {t.hint}</span>}
            </button>
          );
        })}
      </div>

      {/* Tab body */}
      {tab === "preview"  && <ResumePreviewTab  T={T} lang={lang} version={version} original={original} />}
      {tab === "rewrites" && !isMaster && <ResumeRewritesTab T={T} lang={lang} version={version} />}
      {tab === "edit"     && <ResumeEditTab     T={T} lang={lang} version={version} isMaster={isMaster} />}
    </div>
  );
};

// ─── PREVIEW TAB ─────────────────────────────────────────────────────
const ResumePreviewTab = ({ T, lang, version, original }) => {
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
            <Btn T={T} variant="secondary" size="sm" icon="download">{isEn ? "Export PDF" : "导出 PDF"}</Btn>
            <Btn T={T} variant="ghost" size="sm" icon="file">{isEn ? "Copy text" : "复制纯文本"}</Btn>
          </div>
        </Card>

        {original && (
          <Card T={T}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{isEn ? "ORIGINAL FILE" : "原始文件"}</div>
            <div style={{ fontSize: 13, color: T.ink2, fontFamily: "var(--ei-mono)", marginBottom: 6, wordBreak: "break-all" }}>{original.name}</div>
            <div style={{ fontSize: 12, color: T.ink3 }}>{original.type} · {original.createdAt}</div>
            <Btn T={T} variant="ghost" size="sm" icon="file" style={{ marginTop: 12 }}>{isEn ? "View original" : "查看原件"}</Btn>
          </Card>
        )}
      </div>
    </div>
  );
};

// ─── REWRITES TAB ────────────────────────────────────────────────────
const ResumeRewritesTab = ({ T, lang, version }) => {
  const isEn = lang === "en";
  const allBullets = React.useMemo(() => buildBullets(lang, version), [lang, version]);
  const [bulletState, setBulletState] = React.useState(() => allBullets.reduce((acc, b) => (acc[b.id] = b.status, acc), {}));
  const [selected, setSelected] = React.useState("b1");
  const [editing, setEditing] = React.useState(false);
  const [editText, setEditText] = React.useState("");

  const bullets = allBullets.map((b) => ({ ...b, status: bulletState[b.id] || b.status }));
  const sel = bullets.find((b) => b.id === selected) || bullets[0];
  const accepted = bullets.filter((b) => b.status === "accepted").length;
  const pending  = bullets.filter((b) => b.status === "pending").length;
  const rejected = bullets.filter((b) => b.status === "rejected").length;

  const setStatus = (id, status) => {
    setBulletState({ ...bulletState, [id]: status });
    setEditing(false);
    if (window.eiToast) {
      const msg = status === "accepted" ? (isEn ? "Accepted into this version" : "已写入当前版本") :
                  status === "rejected" ? (isEn ? "Rejected — original kept" : "已拒绝 · 保留原句") :
                                          (isEn ? "Saved manual edit" : "人工改写已保存");
      window.eiToast(`${msg} · ${version.name}`, { tone: "ok", duration: 2400 });
    }
  };

  return (
    <div>
      {/* Scope banner — answers question #1 */}
      <div style={{
        display: "flex", justifyContent: "space-between", alignItems: "center", gap: 14,
        padding: "10px 14px", marginBottom: 16,
        background: T.accentSoft, border: `1px solid ${T.accent}`, borderRadius: 2,
        flexWrap: "wrap",
      }}>
        <div style={{ fontSize: 13, color: T.ink2 }}>
          <Icon name="info" size={12} color={T.accent} style={{ marginRight: 6 }} />
          {isEn
            ? <>Reject / Edit / Accept apply to <strong>this version only</strong>: <span style={{ fontFamily: "var(--ei-mono)", color: T.ink }}>{version.name}</span>. The master and the original file are not changed.</>
            : <>「拒绝 / 编辑 / 采纳」只作用于<strong>当前版本</strong>：<span style={{ fontFamily: "var(--ei-mono)", color: T.ink }}>{version.name}</span>。主版本与原始文件保持不变。</>
          }
        </div>
        <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
          {accepted} {isEn ? "accepted" : "已采纳"} · {pending} {isEn ? "pending" : "待决定"} · {rejected} {isEn ? "rejected" : "已拒绝"}
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1.3fr", gap: 20 }}>
        {/* Bullet list */}
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{isEn ? "SUGGESTED REWRITES" : "建议改写"}</div>
          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {bullets.map((b) => {
              const active = b.id === selected;
              const statusC = b.status === "accepted" ? T.ok : b.status === "rejected" ? T.ink4 : T.warn;
              return (
                <button key={b.id} onClick={() => { setSelected(b.id); setEditing(false); }} style={{
                  padding: "14px 16px", textAlign: "left", cursor: "pointer",
                  background: active ? T.bgSoft : T.bgCard,
                  border: `1px solid ${active ? T.accent : T.rule}`,
                  borderRadius: 2, fontFamily: "var(--ei-sans)",
                }}>
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 10, marginBottom: 6 }}>
                    <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>{b.section}</div>
                    <div style={{ display: "flex", gap: 4, alignItems: "center", fontSize: 10.5, color: statusC, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>
                      <div style={{ width: 5, height: 5, borderRadius: 3, background: statusC }} />
                      {b.status === "accepted" ? (isEn ? "ACCEPTED" : "已采纳") : b.status === "rejected" ? (isEn ? "REJECTED" : "已拒绝") : (isEn ? "PENDING" : "待决定")}
                    </div>
                  </div>
                  <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.5, textDecoration: b.status === "rejected" ? "line-through" : "none", opacity: b.status === "rejected" ? 0.6 : 1 }}>
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
              <div style={{ display: "flex", gap: 6 }}>
                <button onClick={() => setStatus(sel.id, "rejected")} style={{ padding: "5px 12px", fontSize: 12, cursor: "pointer", background: sel.status === "rejected" ? T.ink2 : "transparent", color: sel.status === "rejected" ? T.bg : T.ink3, border: `1px solid ${T.rule}`, borderRadius: 2, fontFamily: "var(--ei-sans)" }}>
                  <Icon name="x" size={11} style={{ marginRight: 4 }} /> {isEn ? "Reject" : "拒绝"}
                </button>
                <button onClick={() => { setEditing(true); setEditText(sel.rewritten); }} style={{ padding: "5px 12px", fontSize: 12, cursor: "pointer", background: editing ? T.bgSoft : "transparent", color: T.ink2, border: `1px solid ${editing ? T.accent : T.rule}`, borderRadius: 2, fontFamily: "var(--ei-sans)" }}>
                  <Icon name="edit" size={11} style={{ marginRight: 4 }} /> {isEn ? "Edit" : "编辑"}
                </button>
                <button onClick={() => setStatus(sel.id, "accepted")} style={{ padding: "5px 12px", fontSize: 12, cursor: "pointer", background: sel.status === "accepted" ? T.ok : T.accent, color: "#fff", border: "none", borderRadius: 2, fontFamily: "var(--ei-sans)" }}>
                  <Icon name="check" size={11} style={{ marginRight: 4 }} stroke={2.5} /> {sel.status === "accepted" ? (isEn ? "Accepted" : "已采纳") : (isEn ? "Accept" : "采纳")}
                </button>
              </div>
            </div>

            {/* Original */}
            <div style={{ padding: "16px 22px", borderBottom: `1px dotted ${T.rule}` }}>
              <div style={{ display: "flex", gap: 10, alignItems: "center", marginBottom: 8 }}>
                <div style={{ padding: "2px 8px", background: T.dangerSoft, color: T.danger, fontSize: 10.5, fontFamily: "var(--ei-mono)", letterSpacing: "0.08em", borderRadius: 2 }}>
                  − {isEn ? "ORIGINAL" : "原句"}
                </div>
                <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{isEn ? "from master" : "来自主版本"}</div>
              </div>
              <div style={{ fontSize: 14.5, color: T.ink2, lineHeight: 1.65, fontFamily: "var(--ei-serif)", background: T.dangerSoft, padding: "12px 14px", borderRadius: 2, borderLeft: `2px solid ${T.danger}` }}>
                {sel.original}
              </div>
            </div>

            {/* Rewritten or manual edit */}
            <div style={{ padding: "16px 22px", borderBottom: `1px dotted ${T.rule}` }}>
              <div style={{ display: "flex", gap: 10, alignItems: "center", marginBottom: 8 }}>
                <div style={{ padding: "2px 8px", background: editing ? T.accentSoft : T.okSoft, color: editing ? T.accent : T.ok, fontSize: 10.5, fontFamily: "var(--ei-mono)", letterSpacing: "0.08em", borderRadius: 2 }}>
                  + {editing ? (isEn ? "MANUAL EDIT" : "人工改写") : (isEn ? "REWRITTEN" : "AI 改写")}
                </div>
                <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
                  {editing
                    ? (isEn ? `will save into ${version.name}` : `将保存到「${version.name}」`)
                    : (isEn ? "confidence · high" : "置信度 · 高")}
                </div>
              </div>
              {editing ? (
                <div>
                  <textarea
                    value={editText}
                    onChange={(e) => setEditText(e.target.value)}
                    style={{ width: "100%", minHeight: 110, padding: "12px 14px", border: `1px solid ${T.accent}`, background: T.accentSoft, color: T.ink, borderRadius: 2, fontFamily: "var(--ei-serif)", fontSize: 14.5, lineHeight: 1.65, resize: "vertical", outline: "none" }}
                  />
                  <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 10 }}>
                    <Btn T={T} variant="ghost" size="sm" onClick={() => setEditing(false)}>{isEn ? "Cancel" : "取消"}</Btn>
                    <Btn T={T} variant="accent" size="sm" icon="check" onClick={() => setStatus(sel.id, "accepted")}>{isEn ? "Save manual edit" : "保存人工改写"}</Btn>
                  </div>
                </div>
              ) : (
                <div style={{ fontSize: 14.5, color: T.ink, lineHeight: 1.65, fontFamily: "var(--ei-serif)", background: T.okSoft, padding: "12px 14px", borderRadius: 2, borderLeft: `2px solid ${T.ok}` }}>
                  {sel.rewritten}
                </div>
              )}
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
    </div>
  );
};

// ─── EDIT TAB (manual field edit, lightweight) ───────────────────────
const ResumeEditTab = ({ T, lang, version, isMaster }) => {
  const isEn = lang === "en";
  const [headline, setHeadline] = React.useState(isEn ? "Senior Frontend Engineer · Frontend platform & checkout" : "资深前端工程师 · 前端平台 & 结账链路");
  const [summary, setSummary] = React.useState(isEn
    ? "Six years of frontend, last three on platform tooling. Comfortable owning architecture from spec to ship."
    : "六年前端，最近三年在平台工程方向。能从架构方案到上线一手承担。");

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
          {isMaster
            ? (isEn ? <>Editing the <strong>master</strong>. Targeted versions branched from this master will keep their existing accepted bullets, but unedited fields will follow.</> : <>正在编辑<strong>主版本</strong>。已分叉的定制版本会保留已采纳的句子，未改的字段跟随主版本。</>)
            : (isEn ? <>Editing <strong>{version.name}</strong>. Changes here do not affect the master or other versions.</> : <>正在编辑 <strong>{version.name}</strong>。改动不会影响主版本或其它定制版本。</>)
          }
        </div>
        <Btn T={T} variant="accent" size="sm" icon="check">{isEn ? "Save changes" : "保存改动"}</Btn>
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

// ─── Branch flow — create a targeted version off an existing tree ────
//   Distinct from ResumeCreateFlow: no upload, no parsing. The source tree
//   already exists; this is purely "configure the branch": target JD, name,
//   focus, and how to seed the bullets.
const ResumeBranchFlow = ({ T, lang, original, master, onBack }) => {
  const isEn = lang === "en";
  const [name, setName] = React.useState("");
  const [target, setTarget] = React.useState("");
  const [focus, setFocus] = React.useState("platform");
  const [seed, setSeed] = React.useState("copy_master");

  const focusOptions = [
    { k: "platform",     label: isEn ? "Platform / infra angle"      : "前端平台 / 基建方向" },
    { k: "collaboration",label: isEn ? "Collaboration & impact"      : "协作影响力" },
    { k: "fullstack",    label: isEn ? "Full-stack breadth"          : "全栈广度" },
    { k: "leadership",   label: isEn ? "Tech leadership"             : "技术 Lead 视角" },
    { k: "custom",       label: isEn ? "Custom — I'll write it"      : "自定义" },
  ];
  const seedOptions = [
    { k: "copy_master", icon: "layers", label: isEn ? "Start from master" : "从主版本复制",
      desc: isEn ? "All bullets copied from the master. AI rewrites suggested per JD." : "全量复制主版本 bullet，按岗位生成改写建议。" },
    { k: "blank",       icon: "file",      label: isEn ? "Start blank"       : "空白起步",
      desc: isEn ? "Empty version. You write or accept rewrites bullet by bullet." : "空白版本，逐条人工撰写或采纳建议。" },
    { k: "ai_select",   icon: "sparkle",   label: isEn ? "AI picks bullets"  : "AI 选 bullet",
      desc: isEn ? "Match the JD against the master, keep top-N most relevant." : "用 JD 对齐主版本，保留最相关的 N 条。" },
  ];

  const canSubmit = name.trim().length > 0 && target.trim().length > 0;

  return (
    <div className="ei-fadein" style={{ maxWidth: 980, margin: "0 auto", padding: "40px 48px 96px" }}>
      <button onClick={onBack} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 16, display: "flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {isEn ? "Back to resume workshop" : "返回简历工坊"}
      </button>
      <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{isEn ? "NEW TARGETED VERSION" : "新建定制版本"}</div>
      <h1 className="ei-serif" style={{ fontSize: 32, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.2 }}>
        {isEn ? "Branch a targeted version from this tree." : "从这棵树分叉一个面向岗位的定制版本。"}
      </h1>
      <div style={{ fontSize: 14, color: T.ink3, marginTop: 10, maxWidth: 720, lineHeight: 1.55 }}>
        {isEn
          ? "The original file and the master stay untouched. Only this new version receives accepted rewrites."
          : "原始文件与主版本保持不变；后续采纳的改写只写入这个新版本。"}
      </div>

      {/* Source tree context */}
      <div style={{ marginTop: 22 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{isEn ? "BRANCHING FROM" : "底稿来源"}</div>
        <Card T={T} pad={0}>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 0 }}>
            <div style={{ padding: "16px 20px", borderRight: `1px dotted ${T.rule}` }}>
              <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
                <Icon name="file" size={13} color={T.ink3} />
                <div className="ei-label" style={{ color: T.ink3 }}>{isEn ? "Original" : "原始简历"}</div>
              </div>
              <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                {original ? original.name : (isEn ? "—" : "—")}
              </div>
              <div style={{ fontSize: 12, color: T.ink3, marginTop: 3 }}>
                {original ? `${original.type} · ${original.createdAt}` : ""}
              </div>
            </div>
            <div style={{ padding: "16px 20px", background: T.bgSoft }}>
              <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
                <Icon name="resume" size={13} color={T.ink3} />
                <div className="ei-label" style={{ color: T.ink3 }}>{isEn ? "Master (parsed)" : "主版本（解析自原始）"}</div>
              </div>
              <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                {master ? master.name : (isEn ? "—" : "—")}
              </div>
              <div style={{ fontSize: 12, color: T.ink3, marginTop: 3 }}>
                {isEn ? "Stays pristine. Read-only source for the branch." : "保持干净，作为分叉的只读来源。"}
              </div>
            </div>
          </div>
        </Card>
      </div>

      {/* Form */}
      <div style={{ marginTop: 22, display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
        <Card T={T} pad={20}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{isEn ? "VERSION NAME" : "版本名称"}</div>
          <input
            value={name} onChange={(e) => setName(e.target.value)}
            placeholder={isEn ? "e.g. v3 ByteDance · FE Platform" : "例如：v3 字节-前端平台"}
            style={{
              width: "100%", boxSizing: "border-box", padding: "10px 12px",
              border: `1px solid ${T.rule}`, borderRadius: 2, background: T.bg,
              fontFamily: "var(--ei-sans)", fontSize: 14, color: T.ink,
            }}
          />
          <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 6 }}>
            {isEn ? "Shown in the version list and on JD-match results." : "会显示在版本列表和岗位匹配结果中。"}
          </div>
        </Card>
        <Card T={T} pad={20}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{isEn ? "TARGET JOB / COMPANY" : "目标岗位 / 公司"}</div>
          <input
            value={target} onChange={(e) => setTarget(e.target.value)}
            placeholder={isEn ? "e.g. ByteDance · Frontend Platform" : "例如：字节跳动 · 前端平台"}
            style={{
              width: "100%", boxSizing: "border-box", padding: "10px 12px",
              border: `1px solid ${T.rule}`, borderRadius: 2, background: T.bg,
              fontFamily: "var(--ei-sans)", fontSize: 14, color: T.ink,
            }}
          />
          <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 6 }}>
            {isEn ? "Anchors the JD-match score and the rewrite suggestions." : "用于锚定 JD 匹配分数与后续改写建议。"}
          </div>
        </Card>
      </div>

      <div style={{ marginTop: 14 }}>
        <Card T={T} pad={20}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{isEn ? "FOCUS / ANGLE" : "侧重方向"}</div>
          <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
            {focusOptions.map((f) => {
              const on = focus === f.k;
              return (
                <button key={f.k} onClick={() => setFocus(f.k)} style={{
                  padding: "8px 14px",
                  background: on ? T.accentSoft : "transparent",
                  border: `1px solid ${on ? T.accent : T.rule}`,
                  borderRadius: 2,
                  color: on ? T.accent : T.ink2,
                  fontFamily: "var(--ei-sans)", fontSize: 13,
                  cursor: "pointer",
                }}>{f.label}</button>
              );
            })}
          </div>
          <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 10 }}>
            {isEn ? "Used to bias rewrite suggestions for this version." : "用于影响这个版本的改写建议倾向。"}
          </div>
        </Card>
      </div>

      <div style={{ marginTop: 14 }}>
        <Card T={T} pad={20}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{isEn ? "HOW TO SEED BULLETS" : "Bullet 初始化方式"}</div>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 10 }}>
            {seedOptions.map((s) => {
              const on = seed === s.k;
              return (
                <button key={s.k} onClick={() => setSeed(s.k)} style={{
                  textAlign: "left", padding: "14px 14px",
                  background: on ? T.accentSoft : T.bg,
                  border: `1px solid ${on ? T.accent : T.rule}`,
                  borderRadius: 2, cursor: "pointer",
                }}>
                  <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
                    <Icon name={s.icon} size={13} color={on ? T.accent : T.ink3} />
                    <div className="ei-label" style={{ color: on ? T.accent : T.ink3 }}>{s.label}</div>
                  </div>
                  <div style={{ fontSize: 12, color: T.ink2, lineHeight: 1.5 }}>{s.desc}</div>
                </button>
              );
            })}
          </div>
        </Card>
      </div>

      {/* Actions */}
      <div style={{ marginTop: 22, display: "flex", justifyContent: "space-between", alignItems: "center", gap: 12 }}>
        <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
          {canSubmit
            ? (isEn ? "Ready to branch." : "准备就绪，可以分叉。")
            : (isEn ? "Fill in version name and target to continue." : "填写版本名称与目标岗位后继续。")}
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          <Btn T={T} variant="ghost"  size="sm" onClick={onBack}>{isEn ? "Cancel" : "取消"}</Btn>
          <Btn T={T} variant="accent" size="sm" icon="check" disabled={!canSubmit} onClick={onBack}>
            {isEn ? "Create version" : "创建版本"}
          </Btn>
        </div>
      </div>
    </div>
  );
};

// ─── Original create/upload flow ─────────────────────────────────────
// Restored from the legacy ResumeVersionsScreen (screens-p1-depth.jsx).
// Three modes — upload / paste / guided — each with distinct content and
// a sidebar explaining what gets saved and what to do after v1.
// ─────────────────────────────────────────────────────────
// Dynamic parse flow — streams agent steps after submit, mirrors
// the JD scanning pattern: a card with mono step lines that "tick on"
// in sequence, then transitions to preview.
// ─────────────────────────────────────────────────────────
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

// ─────────────────────────────────────────────────────────
// Preview & confirm — shows parsed structured resume after
// the dynamic flow. User confirms (saves as v1) or goes back.
// ─────────────────────────────────────────────────────────
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
          <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{isEn ? "PREVIEW · CONFIRM TO SAVE AS V1" : "预览 · 确认后保存为 v1"}</div>
          <h1 className="ei-serif" style={{ fontSize: 34, margin: 0, color: T.ink, letterSpacing: "-0.02em", lineHeight: 1.2 }}>
            {isEn ? "Here's the structured draft. Confirm to save it as a new original." : "结构化草稿如下，确认后保存为一份新的原始简历。"}
          </h1>
          <div style={{ fontSize: 13, color: T.ink3, marginTop: 10, fontFamily: "var(--ei-mono)" }}>
            <span style={{ color: T.ink2 }}>source · </span>{sourceLabel} <span style={{ color: T.ink4, margin: "0 8px" }}>·</span> <span style={{ color: T.ok }}>{isEn ? "parsed" : "已解析"}</span>
          </div>
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          <Btn T={T} variant="ghost" onClick={onBack}>{isEn ? "Back" : "返回上一步"}</Btn>
          <Btn T={T} variant="accent" icon="check" onClick={() => {
            window.eiToast && window.eiToast(isEn ? "Saved as v1 · added to workshop" : "已保存为 v1· 已加入简历工坊", { tone: "ok", duration: 2400 });
            onConfirm && onConfirm();
          }}>{isEn ? "Confirm and save v1" : "确认并保存 v1"}</Btn>
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
              { icon: "file", t: isEn ? "Original source" : "原始版本", b: isEn ? "Kept untouched · always traceable." : "原始内容保持不变，可追溯。" },
              { icon: "resume", t: isEn ? "Structured resume" : "结构化简历", b: isEn ? "The draft above · editable later." : "上面的草稿 · 后续可编辑。" },
              { icon: "layers", t: isEn ? "Master version (v1)" : "主版本（v1）", b: isEn ? "Future targeted versions branch from here." : "未来针对岗位的版本从这里分叉。" },
            ].map((it, i) => (
              <div key={it.t} style={{ display: "grid", gridTemplateColumns: "26px 1fr", gap: 10, padding: "12px 0", borderBottom: i < 2 ? `1px dotted ${T.rule}` : "none" }}>
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

const ResumeCreateFlow = ({ T, lang, nav, onBack }) => {
  const [stage, setStage] = React.useState("input"); // "input" | "parsing" | "preview"
  const [createMode, setCreateMode] = React.useState("upload");
  const [resumeText, setResumeText] = React.useState("");
  const [guideStep, setGuideStep] = React.useState(0);
  const [pickedFile, setPickedFile] = React.useState(null);
  const [guideAnswers, setGuideAnswers] = React.useState(["", "", "", "", ""]);

  const guideSteps = lang === "en" ? [
    { k: "Last role", q: "Where did you work most recently, and what was your title?", ph: "Company, title, dates…" },
    { k: "Direction", q: "What product or engineering direction did you mainly own?", ph: "Frontend platform, growth, infra, data…" },
    { k: "Proof project", q: "Pick one project that proves your level.", ph: "Problem, action, result…" },
    { k: "Numbers", q: "What measurable result can we attach to that project?", ph: "Performance, revenue, adoption, efficiency…" },
    { k: "Target", q: "What roles are you preparing for now?", ph: "Senior frontend, staff platform, AI infra…" },
  ] : [
    { k: "最近经历", q: "你最近在哪家公司、担任什么职位？", ph: "公司、职位、时间段…" },
    { k: "主要方向", q: "你主要负责什么产品或技术方向？", ph: "前端平台、增长、基础设施、数据、AI 应用…" },
    { k: "代表项目", q: "选一个最能证明你能力的项目。", ph: "背景、你做了什么、最后结果…" },
    { k: "量化结果", q: "这个项目能补哪些数字或可验证结果？", ph: "性能、收入、采用率、效率、成本…" },
    { k: "目标岗位", q: "你现在想准备什么类型的岗位？", ph: "资深前端、平台工程、AI 应用工程…" },
  ];

  const sourceLabel = createMode === "upload"
    ? (pickedFile ? pickedFile.name : (lang === "en" ? "Uploaded file" : "上传文件"))
    : createMode === "paste"
      ? (lang === "en" ? "Pasted text" : "粘贴文本")
      : (lang === "en" ? "Guided answers" : "轻量问答");

  const handleSubmit = () => setStage("parsing");

  if (stage === "parsing") {
    return <ResumeParseFlow T={T} lang={lang} sourceLabel={sourceLabel} onDone={() => setStage("preview")} onBack={() => setStage("input")} />;
  }
  if (stage === "preview") {
    return <ResumePreviewConfirm T={T} lang={lang} sourceLabel={sourceLabel} onConfirm={onBack} onBack={() => setStage("input")} />;
  }

  return (
    <div className="ei-fadein" style={{ maxWidth: 1220, margin: "0 auto", padding: "40px 48px 96px" }}>
      <button onClick={onBack} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, marginBottom: 22, display: "flex", alignItems: "center", gap: 6, cursor: "pointer" }}>
        <Icon name="arrow_left" size={14} /> {lang === "en" ? "Back to resume workshop" : "返回简历工坊"}
      </button>

      <div style={{ marginBottom: 26 }}>
        <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{lang === "en" ? "FIRST RESUME VERSION" : "创建第一版简历"}</div>
        <h1 className="ei-serif" style={{ fontSize: 38, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.15 }}>
          {lang === "en" ? "Start from a file, pasted text, or a five-minute guided draft." : "上传、粘贴，或用 5 分钟问答生成第一版。"}
        </h1>
        <div style={{ fontSize: 14, color: T.ink2, marginTop: 10, maxWidth: 720, lineHeight: 1.6 }}>
          {lang === "en" ? "We keep the original source, parse it into a structured resume, and save both as a version you can revise later." : "系统会保留原始文件或原始文本，同时解析成结构化简历，并作为可回溯版本保存。"}
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 340px", gap: 22, alignItems: "start" }}>
        <Card T={T} pad={0}>
          <div style={{ display: "flex", borderBottom: `1px solid ${T.rule}`, overflowX: "auto" }}>
            {[
              { k: "upload", icon: "upload", zh: "上传文件", en: "Upload" },
              { k: "paste", icon: "file", zh: "粘贴内容", en: "Paste" },
              { k: "guided", icon: "chat", zh: "轻量问答", en: "Guided" },
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
                  {lang === "en" ? "The source file is stored as the original version. Parsed sections become your editable structured resume." : "原始文件会作为原始版本保存，解析出的工作经历、项目、技能和教育经历会进入可编辑结构化简历。"}
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
                <Btn T={T} variant="accent" icon="sparkle" disabled={!resumeText.trim()} onClick={handleSubmit}>{lang === "en" ? "Parse and save v1" : "解析并保存 v1"}</Btn>
              </div>
            </div>
          )}

          {createMode === "guided" && (
            <div style={{ padding: 24 }}>
              <div style={{ display: "grid", gridTemplateColumns: "220px 1fr", gap: 22 }}>
                <div style={{ borderRight: `1px solid ${T.rule}`, paddingRight: 18 }}>
                  {guideSteps.map((s, i) => (
                    <button key={s.k} onClick={() => setGuideStep(i)} style={{
                      width: "100%", padding: "11px 0", background: "transparent", border: "none", textAlign: "left",
                      display: "flex", gap: 10, alignItems: "center", color: guideStep === i ? T.ink : T.ink3, cursor: "pointer",
                    }}>
                      <span style={{ width: 22, height: 22, borderRadius: 11, background: guideStep === i ? T.accent : T.bgSoft, color: guideStep === i ? "#fff" : T.ink3, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontFamily: "var(--ei-mono)" }}>{i + 1}</span>
                      <span style={{ fontSize: 13 }}>{s.k}</span>
                    </button>
                  ))}
                </div>
                <div>
                  <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>{lang === "en" ? "GUIDED DRAFT" : "轻量问答建档"}</div>
                  <div className="ei-serif" style={{ fontSize: 24, color: T.ink, lineHeight: 1.35, marginBottom: 16 }}>{guideSteps[guideStep].q}</div>
                  <textarea
                    value={guideAnswers[guideStep]}
                    onChange={(e) => {
                      const next = guideAnswers.slice();
                      next[guideStep] = e.target.value;
                      setGuideAnswers(next);
                    }}
                    placeholder={guideSteps[guideStep].ph}
                    style={{ width: "100%", minHeight: 150, border: `1px solid ${T.rule}`, borderRadius: 2, padding: 14, background: T.bg, color: T.ink, fontSize: 14, lineHeight: 1.6, resize: "vertical", outline: "none" }}
                  />
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: 14 }}>
                    <div style={{ fontSize: 12.5, color: T.ink3 }}>
                      {lang === "en" ? "Answer only the parts you know. The draft can be refined later." : "只回答你现在记得的部分，生成后还可以继续补充。"}
                    </div>
                    <div style={{ display: "flex", gap: 8 }}>
                      <Btn T={T} variant="ghost" size="sm" onClick={() => setGuideStep(Math.max(0, guideStep - 1))}>{lang === "en" ? "Back" : "上一步"}</Btn>
                      <Btn T={T} variant="accent" size="sm" iconRight="arrow_right" onClick={() => guideStep < guideSteps.length - 1 ? setGuideStep(guideStep + 1) : handleSubmit()}>
                        {guideStep < guideSteps.length - 1 ? (lang === "en" ? "Next" : "下一步") : (lang === "en" ? "Generate v1" : "生成 v1")}
                      </Btn>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </Card>

        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <Card T={T}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "WHAT GETS SAVED" : "会保存什么"}</div>
            {[
              { icon: "file", title: lang === "en" ? "Original source" : "原始版本", body: lang === "en" ? "File, pasted text, or guided answers stay traceable." : "文件、粘贴文本或问答记录都会保留来源。" },
              { icon: "resume", title: lang === "en" ? "Structured resume" : "结构化简历", body: lang === "en" ? "Work, projects, skills, education, and evidence are editable." : "工作经历、项目、技能、教育和证据点可编辑。" },
              { icon: "layers", title: lang === "en" ? "Version baseline" : "版本基线", body: lang === "en" ? "Future JD-specific resumes branch from this v1." : "未来针对不同 JD 的版本从 v1 分叉。" },
            ].map((item, i) => (
              <div key={item.title} style={{ display: "grid", gridTemplateColumns: "26px 1fr", gap: 10, padding: "12px 0", borderBottom: i < 2 ? `1px dotted ${T.rule}` : "none" }}>
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
                ? <>After you submit, we'll <span style={{ color: T.ink, fontWeight: 500 }}>parse the source live</span>, then show you a <span style={{ color: T.ink, fontWeight: 500 }}>preview to confirm</span> before saving as v1.</>
                : <>提交之后会<span style={{ color: T.ink, fontWeight: 500 }}>动态解析原始内容</span>，然后进入<span style={{ color: T.ink, fontWeight: 500 }}>预览确认页</span>，确认后才保存为 v1。</>}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
};

window.ResumeVersionsScreen = ResumeWorkshopScreen;

const ResumeBranchMap = ({ T, lang, original, siblings, version }) => {
  const isEn = lang === "en";
  const master = (siblings || []).find((v) => v.tag === "MASTER");
  const isMaster = version.tag === "MASTER";
  return (
    <div style={{
      display: "grid",
      gridTemplateColumns: isMaster ? "1fr 1fr" : "1fr 1fr 1fr",
      gap: 0,
      border: `1px solid ${T.rule}`,
      background: T.bgCard,
      borderRadius: 3,
      overflow: "hidden",
    }}>
      {[
        original ? {
          icon: "file", label: isEn ? "Original" : "原始简历",
          title: original.name,
          body: `${original.type} · ${original.createdAt}`,
        } : null,
        master ? {
          icon: "resume", label: isEn ? "Master (parsed)" : "主版本（解析自原始）",
          title: master.name,
          body: isEn ? "Editable structured fields. The master stays pristine." : "解析为结构化字段；主版本保持干净。",
          highlight: isMaster,
        } : null,
        !isMaster ? {
          icon: "briefcase", label: isEn ? "This version" : "当前版本",
          title: version.name,
          body: isEn ? `Branch for ${version.target}. Only accepted rewrites land here.` : `面向「${version.target}」分叉；只把采纳的改写写入此版本。`,
          highlight: true,
        } : null,
      ].filter(Boolean).map((item, i, arr) => (
        <div key={item.label} style={{
          padding: "14px 18px",
          borderRight: i < arr.length - 1 ? `1px dotted ${T.rule}` : "none",
          background: item.highlight ? T.accentSoft : "transparent",
          minWidth: 0,
        }}>
          <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
            <Icon name={item.icon} size={13} color={item.highlight ? T.accent : T.ink3} />
            <div className="ei-label" style={{ color: item.highlight ? T.accent : T.ink3 }}>{item.label}</div>
          </div>
          <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{item.title}</div>
          <div style={{ fontSize: 12, color: T.ink3, lineHeight: 1.5, marginTop: 3 }}>{item.body}</div>
        </div>
      ))}
    </div>
  );
};
