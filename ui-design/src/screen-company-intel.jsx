// Company Intel · 轻量岗位情报（spec P2 · 合规来源 · 强制标注抓取时间）

const CompanyIntelScreen = ({ T, lang, nav }) => {
  const intel = mockIntel(lang);
  return (
    <div className="ei-fadein" style={{ maxWidth: 1240, margin: "0 auto", padding: "40px 48px 96px" }}>
      <div style={{ marginBottom: 24 }}>
        <button onClick={() => nav("workspace", { jobId: "tj-1" })} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 12.5, cursor: "pointer", display: "flex", alignItems: "center", gap: 6, padding: 0, marginBottom: 16 }}>
          <Icon name="arrow_left" size={12} /> {lang === "en" ? "Back to interview setup" : "返回面试前确认"}
        </button>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
          {lang === "en" ? "COMPANY INTEL · LIGHT-TOUCH · COMPLIANT SOURCES" : "公司情报 · 轻量版 · 仅合规公开来源"}
        </div>
        <h1 className="ei-serif" style={{ fontSize: 36, color: T.ink, margin: 0, letterSpacing: "-0.022em", lineHeight: 1.15 }}>
          {intel.company}
        </h1>
        <div style={{ fontSize: 14, color: T.ink3, marginTop: 8 }}>
          {intel.tagline}
        </div>
      </div>

      <CompanyIntelBody T={T} lang={lang} intel={intel} />
    </div>
  );
};

const CompanyIntelBody = ({ T, lang, intel, compact = false }) => {
  return (
    <div style={{ display: "grid", gridTemplateColumns: compact ? "1fr" : "1.4fr 1fr", gap: 20 }}>
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Card T={T} pad={0}>
          <div style={{ padding: "18px 22px", borderBottom: `1px solid ${T.rule}` }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "ONE-LINE PROFILE" : "一句话画像"}</div>
            <div className="ei-serif" style={{ fontSize: 17, color: T.ink, lineHeight: 1.5, letterSpacing: "-0.005em" }}>
              {intel.oneLiner}
            </div>
          </div>
          <div style={{ padding: "14px 22px", display: "flex", gap: 18, flexWrap: "wrap", fontSize: 12.5, color: T.ink2 }}>
            {intel.facts.map((f, i) => (
              <div key={i}>
                <span style={{ color: T.ink3, marginRight: 6 }}>{f.k}</span>
                <span style={{ fontFamily: "var(--ei-mono)" }}>{f.v}</span>
              </div>
            ))}
          </div>
        </Card>

        <Card T={T} pad={0}>
          <div style={{ padding: "16px 22px", borderBottom: `1px dotted ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "RECENT PUBLIC SIGNALS · 90 DAYS" : "近 90 天公开动态"}</div>
            <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{intel.signals.length} {lang === "en" ? "items" : "条"}</div>
          </div>
          {intel.signals.map((s, i) => (
            <div key={i} style={{ padding: "14px 22px", borderBottom: i < intel.signals.length - 1 ? `1px dotted ${T.rule}` : "none", display: "flex", gap: 14 }}>
              <div style={{ width: 72, flexShrink: 0, fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{s.date}</div>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 13.5, color: T.ink, lineHeight: 1.45, marginBottom: 4 }}>{s.headline}</div>
                <div style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
                  <Tag T={T} tone={s.toneTag}>{s.kind}</Tag> · {s.source}
                </div>
              </div>
            </div>
          ))}
        </Card>

        <Card T={T} pad={0}>
          <div style={{ padding: "16px 22px", borderBottom: `1px dotted ${T.rule}` }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "INTERVIEW STYLE HINTS · derived from public signals" : "面试风格提示 · 由公开信号推断"}</div>
          </div>
          <div style={{ padding: "16px 22px", display: "grid", gridTemplateColumns: "repeat(2, 1fr)", gap: 14 }}>
            {intel.styleHints.map((h, i) => (
              <div key={i} style={{ padding: "12px 14px", background: T.bgSoft, borderLeft: `2px solid ${T.accent}`, borderRadius: 2 }}>
                <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em", marginBottom: 4 }}>{h.k}</div>
                <div style={{ fontSize: 13, color: T.ink, lineHeight: 1.45 }}>{h.v}</div>
              </div>
            ))}
          </div>
          <div style={{ padding: "10px 22px 14px", fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", lineHeight: 1.5, fontStyle: "italic" }}>
            {lang === "en"
              ? "Generated from public sources only. Treat as orientation, not prediction."
              : "仅基于公开来源推断。用作熟悉方向，不作预测结论。"}
          </div>
        </Card>
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Card T={T} pad={0}>
          <div style={{ padding: "16px 20px", borderBottom: `1px dotted ${T.rule}` }}>
            <div className="ei-label" style={{ color: T.accent }}>{lang === "en" ? "★ REVERSE-QUESTION SUGGESTIONS" : "★ 反问建议"}</div>
            <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 4 }}>
              {lang === "en" ? "Anchored to specific public signals" : "每条都锚定具体公开信号"}
            </div>
          </div>
          {intel.reverseQs.map((q, i) => (
            <div key={i} style={{ padding: "14px 20px", borderBottom: i < intel.reverseQs.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
              <div className="ei-serif" style={{ fontSize: 14, color: T.ink, lineHeight: 1.45, letterSpacing: "-0.005em", marginBottom: 6 }}>
                「{q.q}」
              </div>
              <div style={{ fontSize: 11.5, color: T.ink3, lineHeight: 1.5 }}>
                <span style={{ color: T.ink2 }}>{lang === "en" ? "anchored to · " : "锚定 · "}</span>{q.anchor}
              </div>
            </div>
          ))}
        </Card>

        <Card T={T} pad={0}>
          <div style={{ padding: "16px 20px", borderBottom: `1px dotted ${T.rule}` }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "INTERNAL VOCABULARY" : "内部词汇"}</div>
          </div>
          {intel.glossary.map((g, i) => (
            <div key={i} style={{ padding: "12px 20px", borderBottom: i < intel.glossary.length - 1 ? `1px dotted ${T.rule}` : "none", display: "flex", gap: 12 }}>
              <div style={{ minWidth: 80, fontFamily: "var(--ei-mono)", fontSize: 12, color: T.accent }}>{g.term}</div>
              <div style={{ fontSize: 12.5, color: T.ink2, lineHeight: 1.45 }}>{g.def}</div>
            </div>
          ))}
        </Card>

        <Card T={T} pad={0}>
          <div style={{ padding: "14px 20px", borderBottom: `1px dotted ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "SOURCES & FRESHNESS" : "来源与抓取时间"}</div>
            <span style={{ fontSize: 10.5, color: T.ok, fontFamily: "var(--ei-mono)" }}>
              <span style={{ display: "inline-block", width: 5, height: 5, borderRadius: 3, background: T.ok, marginRight: 5, verticalAlign: "middle" }} />
              {lang === "en" ? "all compliant" : "全部合规"}
            </span>
          </div>
          {intel.sources.map((s, i) => (
            <div key={i} style={{ padding: "10px 20px", borderBottom: i < intel.sources.length - 1 ? `1px dotted ${T.rule}` : "none", display: "flex", justifyContent: "space-between", gap: 10 }}>
              <div style={{ fontSize: 12, color: T.ink, fontFamily: "var(--ei-mono)" }}>{s.url}</div>
              <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", flexShrink: 0 }}>{s.fetched}</div>
            </div>
          ))}
          <div style={{ padding: "12px 20px", fontSize: 11, color: T.ink3, lineHeight: 1.5, fontStyle: "italic" }}>
            {lang === "en"
              ? "We do not aggregate employer reviews or score culture-fit. No private data, no scraping behind logins."
              : "不做雇主评分聚合，不做文化契合度打分。不抓登录后内容、不使用私域数据。"}
          </div>
        </Card>

        <div style={{ padding: "10px 14px", background: T.bgSoft, border: `1px dashed ${T.rule}`, borderRadius: 2, fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <span>{lang === "en" ? "auto-refresh every 24h · last 6h ago" : "每 24 小时自动刷新 · 最近 6 小时前"}</span>
          <button style={{ background: "transparent", border: "none", color: T.accent, fontSize: 11.5, cursor: "pointer" }}>
            {lang === "en" ? "refresh now" : "立即刷新"}
          </button>
        </div>
      </div>
    </div>
  );
};

const CompanyIntelEmbed = ({ T, lang, nav, job }) => {
  const intel = mockIntelForJob(lang, job);
  return (
    <div style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
      <div style={{ padding: "16px 20px", borderBottom: `1px dotted ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 14 }}>
        <div style={{ flex: 1 }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{lang === "en" ? "COMPANY INTEL · LIGHT" : "公司情报 · 轻量"}</div>
          <div className="ei-serif" style={{ fontSize: 16, color: T.ink, lineHeight: 1.4, letterSpacing: "-0.005em" }}>
            {intel.oneLiner}
          </div>
        </div>
        <button onClick={() => nav("company_intel")} style={{ background: "transparent", border: `1px solid ${T.rule}`, padding: "5px 10px", borderRadius: 2, color: T.ink2, fontSize: 11.5, cursor: "pointer", display: "flex", alignItems: "center", gap: 5, flexShrink: 0 }}>
          {lang === "en" ? "Open intel" : "打开情报"} <Icon name="arrow_right" size={11} />
        </button>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 0 }}>
        <div style={{ padding: "14px 20px", borderRight: `1px dotted ${T.rule}` }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "RECENT SIGNALS" : "近期信号"}</div>
          {intel.signals.slice(0, 3).map((s, i) => (
            <div key={i} style={{ padding: "6px 0", borderBottom: i < 2 ? `1px dotted ${T.rule}` : "none", fontSize: 12.5, color: T.ink2, lineHeight: 1.45 }}>
              <span style={{ fontFamily: "var(--ei-mono)", color: T.ink3, fontSize: 10.5, marginRight: 8 }}>{s.date}</span>
              {s.headline}
            </div>
          ))}
        </div>
        <div style={{ padding: "14px 20px" }}>
          <div className="ei-label" style={{ color: T.accent, marginBottom: 8 }}>{lang === "en" ? "★ REVERSE QS" : "★ 反问建议"}</div>
          {intel.reverseQs.slice(0, 2).map((q, i) => (
            <div key={i} className="ei-serif" style={{ fontSize: 13, color: T.ink, lineHeight: 1.45, padding: "6px 0", borderBottom: i < 1 ? `1px dotted ${T.rule}` : "none" }}>
              「{q.q}」
            </div>
          ))}
        </div>
      </div>

      <div style={{ padding: "10px 20px", borderTop: `1px dotted ${T.rule}`, display: "flex", justifyContent: "space-between", fontSize: 10.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
        <span>{lang === "en" ? `${intel.sources.length} compliant sources · last 6h ago` : `${intel.sources.length} 个合规来源 · 6 小时前刷新`}</span>
        <span style={{ color: T.ok }}>
          <span style={{ display: "inline-block", width: 5, height: 5, borderRadius: 3, background: T.ok, marginRight: 5, verticalAlign: "middle" }} />
          {lang === "en" ? "no employer scoring · no private data" : "无雇主评分 · 无私域数据"}
        </span>
      </div>
    </div>
  );
};

// ─── mock data ───
const mockIntelForJob = (lang, job) => {
  if (job?.id === "tj-2") {
    return lang === "en" ? {
      oneLiner: "Remote frontend platform team building internal DX tools — the interview will likely test platform boundaries, async collaboration, and English technical explanation.",
      signals: [
        { date: "2026-04-16", headline: "Lumen Labs published a monorepo migration write-up for its design platform", kind: "Engineering blog", toneTag: "accent" },
        { date: "2026-04-03", headline: "Hiring post emphasizes remote-first rituals and platform ownership", kind: "Hiring signal", toneTag: "ok" },
        { date: "2026-03-27", headline: "DX tooling team opened 2 frontend platform roles for APAC overlap", kind: "Team growth", toneTag: "amber" },
      ],
      reverseQs: [
        { q: "How does the platform team decide what belongs in shared tooling versus product-owned code?" },
        { q: "What async rituals have worked best for keeping platform decisions visible across time zones?" },
      ],
      sources: [{}, {}, {}, {}],
    } : {
      oneLiner: "远程前端平台团队正在建设内部 DX 工具 —— 这场面试更可能考察平台边界、异步协作和英文技术表达。",
      signals: [
        { date: "2026-04-16", headline: "Lumen Labs 发布 Monorepo 迁移文章，重点提到设计平台治理", kind: "技术博客", toneTag: "accent" },
        { date: "2026-04-03", headline: "招聘帖强调 remote-first 协作仪式和平台 owner 机制", kind: "招聘信号", toneTag: "ok" },
        { date: "2026-03-27", headline: "DX tooling 团队新增 2 个 APAC overlap 的前端平台岗位", kind: "团队扩张", toneTag: "amber" },
      ],
      reverseQs: [
        { q: "平台团队如何判断一项能力应该沉淀到 shared tooling，还是留在产品业务代码里？" },
        { q: "跨时区协作里，哪些异步机制最能保证平台决策被大家理解和遵守？" },
      ],
      sources: [{}, {}, {}, {}],
    };
  }
  if (job?.id === "tj-3") {
    return lang === "en" ? {
      oneLiner: "Cloud platform group is hiring a senior web architect — expect architecture trade-offs, governance, and cross-team influence questions.",
      signals: [
        { date: "2026-04-12", headline: "Yunqi Group announced a console unification initiative across three product lines", kind: "Product strategy", toneTag: "accent" },
        { date: "2026-03-30", headline: "Public tech talk focused on micro-frontend governance and architecture review cadence", kind: "Tech talk", toneTag: "ok" },
        { date: "2026-03-18", headline: "Hiring page stresses technical leadership beyond individual delivery", kind: "Hiring signal", toneTag: "amber" },
      ],
      reverseQs: [
        { q: "For the console unification effort, what architecture decision has been hardest to align across teams?" },
        { q: "How do architecture review decisions translate into product-team execution commitments?" },
      ],
      sources: [{}, {}, {}, {}],
    } : {
      oneLiner: "云栖集团正在推进多产品控制台统一 —— 技术专家面更可能关注架构取舍、治理机制和跨团队影响力。",
      signals: [
        { date: "2026-04-12", headline: "云栖集团宣布三条产品线控制台统一计划", kind: "产品战略", toneTag: "accent" },
        { date: "2026-03-30", headline: "公开技术分享聚焦微前端治理和架构评审节奏", kind: "技术分享", toneTag: "ok" },
        { date: "2026-03-18", headline: "招聘页强调技术领导力，不只看个人交付", kind: "招聘信号", toneTag: "amber" },
      ],
      reverseQs: [
        { q: "控制台统一过程中，最难跨团队对齐的是哪类架构决策？" },
        { q: "架构评审形成的决策，如何转化成各产品团队真正执行的承诺？" },
      ],
      sources: [{}, {}, {}, {}],
    };
  }
  return mockIntel(lang);
};

const mockIntel = (lang) => {
  if (lang === "en") {
    return {
      company: "Star-Ring Tech",
      tagline: "Series D · enterprise data infra · ~1,200 people · Shanghai HQ",
      oneLiner: "Data-infra player pivoting to a unified analytics platform — frontend now owns the customer-facing console for the new platform launch later this year.",
      facts: [
        { k: "Stage", v: "Series D · 2024" },
        { k: "Headcount", v: "~1,200" },
        { k: "Eng team", v: "~340" },
        { k: "Founded", v: "2016" },
      ],
      signals: [
        { date: "2026-04-19", headline: "Launched Compass analytics platform — frontend rewrite in RSC mentioned in keynote", kind: "Product launch", toneTag: "accent", source: "starring.com/blog (official)" },
        { date: "2026-04-08", headline: "Hired ex-Vercel DX lead Wei Zhang as Head of FE Platform", kind: "Hire", toneTag: "accent", source: "linkedin.com/posts (public)" },
        { date: "2026-03-22", headline: "Q1 earnings: cloud ARR +38%, on-prem flat — investing in cloud frontend", kind: "Financial", toneTag: "ok", source: "starring.com/ir (official)" },
        { date: "2026-03-11", headline: "AAA accessibility certification announced as 2026 ICL goal", kind: "Goal", toneTag: "amber", source: "company AMA on Zhihu (public)" },
        { date: "2026-02-14", headline: "Acquired small DS team from Northwind (4 designers)", kind: "M&A", toneTag: "neutral", source: "36kr.com (news)" },
      ],
      styleHints: [
        { k: "PACE", v: "Tech rounds run 60min, manager rounds 45min — leave time for reverse questions." },
        { k: "DEPTH", v: "They probe 2–3 levels deep on perf/a11y. Bring numbers (LCP, INP, percent improvements)." },
        { k: "VOCABULARY", v: "Internal terms: Compass, Owner-of-Record, thin client. Use them sparingly." },
        { k: "CULTURE", v: "Direct feedback culture; HR rounds are friendly but real. Don't oversell." },
      ],
      reverseQs: [
        { q: "Compass launched in April — what's the first thing the FE team learned post-launch that surprised you?", anchor: "April 19 product launch" },
        { q: "With Wei Zhang joining as FE platform head, how is the platform/product split evolving?", anchor: "April 8 hire" },
        { q: "AAA a11y certification — is the team aiming for company-wide or just Compass?", anchor: "March 11 ICL goal" },
        { q: "After the Northwind DS acquisition, how is design hand-off changing?", anchor: "Feb 14 M&A" },
      ],
      glossary: [
        { term: "Compass", def: "Their newly-launched unified analytics platform. The role you are interviewing for sits inside this." },
        { term: "ICL goal", def: "Internal Commitment Level — a yearly OKR that leadership has personally signed off on. Stronger than a regular OKR." },
        { term: "Thin client", def: "Their term for the browser-based console (vs the legacy installed desktop app)." },
        { term: "Owner-of-Record", def: "Single accountable engineer per surface — different from tech lead." },
      ],
      sources: [
        { url: "starring.com/blog", fetched: "6h ago" },
        { url: "starring.com/ir/q1-2026", fetched: "6h ago" },
        { url: "linkedin.com/company/star-ring", fetched: "6h ago" },
        { url: "36kr.com/news/star-ring-northwind", fetched: "23h ago" },
        { url: "zhihu.com/star-ring-ama (public)", fetched: "23h ago" },
      ],
    };
  }
  return {
    company: "星环科技",
    tagline: "D 轮 · 企业数据基础设施 · 约 1,200 人 · 上海总部",
    oneLiner: "数据基础设施厂商正在转向统一分析平台 —— 前端目前承担今年下半年新平台 Compass 上线的面向客户的控制台。",
    facts: [
      { k: "阶段", v: "D 轮 · 2024" },
      { k: "总人数", v: "约 1,200" },
      { k: "工程团队", v: "约 340" },
      { k: "成立", v: "2016" },
    ],
    signals: [
      { date: "2026-04-19", headline: "发布 Compass 分析平台 —— 主题演讲明确提到前端用 RSC 重写", kind: "产品发布", toneTag: "accent", source: "starring.com/blog（官方）" },
      { date: "2026-04-08", headline: "前 Vercel DX lead 张伟入职 · 任前端平台负责人", kind: "重要入职", toneTag: "accent", source: "linkedin.com/posts（公开）" },
      { date: "2026-03-22", headline: "Q1 财报：云业务 ARR +38%，本地部署持平 —— 在云端前端持续投入", kind: "财务披露", toneTag: "ok", source: "starring.com/ir（官方）" },
      { date: "2026-03-11", headline: "宣布把 AAA 可访问性认证列为 2026 年 ICL 目标", kind: "公开目标", toneTag: "amber", source: "公司 AMA · 知乎（公开）" },
      { date: "2026-02-14", headline: "收购了 Northwind 的设计系统小团队（4 位设计师）", kind: "并购", toneTag: "neutral", source: "36kr.com（新闻）" },
    ],
    styleHints: [
      { k: "节奏", v: "技术轮 60 分钟、经理轮 45 分钟 —— 留出时间反问。" },
      { k: "深度", v: "性能 / 可访问性会被追问 2-3 层。带具体数字（LCP、INP、优化百分比）。" },
      { k: "词汇", v: "内部术语：Compass、Owner-of-Record、瘦客户端。可以适度使用。" },
      { k: "文化", v: "直接反馈文化；HR 轮看似轻松但实质性提问。不要过度推销自己。" },
    ],
    reverseQs: [
      { q: "Compass 是 4 月刚上线的 —— 上线之后前端团队最让你意外的一个发现是什么？", anchor: "4 月 19 日产品发布" },
      { q: "张伟作为前端平台负责人加入后，平台 / 产品的分工边界在怎么演化？", anchor: "4 月 8 日重要入职" },
      { q: "AAA 可访问性认证 —— 团队是按全公司还是只面向 Compass 在推？", anchor: "3 月 11 日 ICL 目标" },
      { q: "Northwind DS 团队进来之后，设计交付的协作方式有什么变化？", anchor: "2 月 14 日并购" },
    ],
    glossary: [
      { term: "Compass", def: "他们刚发布的统一分析平台。你面的这个岗位就在 Compass 内部。" },
      { term: "ICL goal", def: "Internal Commitment Level，由高管亲自签字背书的年度目标，比普通 OKR 更重。" },
      { term: "瘦客户端", def: "他们对浏览器版控制台的称呼（区别于老的桌面版安装包）。" },
      { term: "Owner-of-Record", def: "每个交付面单一负责的工程师 —— 与「技术 lead」不同。" },
    ],
    sources: [
      { url: "starring.com/blog", fetched: "6 小时前" },
      { url: "starring.com/ir/q1-2026", fetched: "6 小时前" },
      { url: "linkedin.com/company/star-ring", fetched: "6 小时前" },
      { url: "36kr.com/news/star-ring-northwind", fetched: "23 小时前" },
      { url: "zhihu.com/star-ring-ama（公开）", fetched: "23 小时前" },
    ],
  };
};

window.CompanyIntelScreen = CompanyIntelScreen;
window.CompanyIntelEmbed = CompanyIntelEmbed;
