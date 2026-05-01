// JD Match · web-search driven job recommendations matched to user profile
// Section of EasyInterview · v1.0

const JDMatchScreen = ({ T, lang, nav }) => {
  const [tab, setTab] = React.useState("recommended"); // recommended / search / watchlist
  const [selected, setSelected] = React.useState("jm-2");
  const [searching, setSearching] = React.useState(false);
  const [query, setQuery] = React.useState(lang === "en" ? "Senior frontend, React + design systems, Shanghai / remote" : "资深前端，React + 设计系统，上海/远程");

  // Profile snapshot — what drives the matching
  const profile = {
    title: lang === "en" ? "Senior Frontend Engineer" : "资深前端工程师",
    years: 5,
    skills: ["React", "TypeScript", "Next.js / RSC", lang === "en" ? "Design Systems" : "Design Systems", lang === "en" ? "Perf" : "性能优化", "a11y"],
    strengths: lang === "en" ? ["quantified perf wins", "cross-team rollout", "mentorship"] : ["量化性能收益", "跨团队推动", "辅导新人"],
    gaps: lang === "en" ? ["large-scale collab story", "formal DS leadership"] : ["大型协作叙事", "DS 正式 leadership"],
    location: lang === "en" ? "Shanghai · open to hybrid / remote" : "上海 · 接受混合 / 远程",
    comp: lang === "en" ? "¥50-70K / month · equity friendly" : "5-7 万/月 · 接受期权",
  };

  // Recommended JDs — surfaced because they fit profile well
  const jobs = [
    {
      id: "jm-1",
      title: lang === "en" ? "Senior Frontend · Platform" : "资深前端 · 平台方向",
      company: "Lumen Labs",
      companyTag: lang === "en" ? "Series C · infra" : "C 轮 · 基础设施",
      location: lang === "en" ? "Shanghai · hybrid 3d" : "上海 · 每周 3 天到岗",
      posted: lang === "en" ? "2 days ago" : "2 天前",
      score: 92,
      seen: false,
      saved: false,
      comp: "¥55-75K × 15",
      source: "lumen.co/careers",
      level: "P6-P7",
      fit: { must: 5, total: 5, plus: 3, totalPlus: 4 },
      reasons: lang === "en"
        ? ["RSC migration story (star cart rewrite)", "Design-system rollout @ 5 products", "a11y deep dive — AAA audit matches their ICL priority"]
        : ["RSC 迁移故事（星环购物车重写）", "5 个产品的 Design System 落地", "a11y 深度 — 与他们的 AAA 审计方向吻合"],
      risks: lang === "en" ? ["Prefers English-speaking team (B-level ok)", "Needs 1+ years TS strict mode — covered"] : ["偏好英文工作（B-level 即可）", "要求 1 年以上 TS strict — 已覆盖"],
      highlights: lang === "en"
        ? ["Own the platform UI SDK used by 40+ internal apps", "Works directly with DX lead (ex-Vercel)", "Perf budget culture — LCP is an OKR"]
        : ["主导平台 UI SDK，40+ 内部应用使用", "与 DX lead（前 Vercel）直接协作", "性能预算写进 OKR"],
      similarInterviewers: 2,
      networkNote: lang === "en" ? "2 similar-background candidates interviewed here before" : "2 位相似背景候选人在这里面试过",
    },
    {
      id: "jm-2",
      title: lang === "en" ? "Staff Frontend · E-commerce Core" : "资深前端 · 电商核心",
      company: "CloudYun · 云栖",
      companyTag: lang === "en" ? "Pre-IPO · e-commerce" : "Pre-IPO · 电商",
      location: lang === "en" ? "Shanghai · on-site" : "上海 · 全职到岗",
      posted: lang === "en" ? "5 days ago" : "5 天前",
      score: 88,
      seen: true,
      saved: true,
      comp: "¥60-85K × 16",
      source: "maimai.cn/job/cy-3021",
      level: "P7",
      fit: { must: 5, total: 5, plus: 2, totalPlus: 4 },
      reasons: lang === "en"
        ? ["Checkout performance rewrite is their #1 OKR — direct hit", "They ship with RSC (rare in ecomm @ scale)", "Hiring manager ex-Star-Ring — shared vocabulary"]
        : ["结账性能重写是他们第 1 OKR——正中红心", "大规模电商中少有的 RSC 用户", "用人经理是前星环同事——词汇一致"],
      risks: lang === "en" ? ["On-site 5d — your preference was hybrid", "Team velocity is aggressive (nightly releases)"] : ["每周 5 天到岗 — 你偏好混合", "节奏激进（每晚发版）"],
      highlights: lang === "en"
        ? ["Drive checkout from P75 LCP 2.8s → <1.5s", "Team of 6 → planned 12 in 2 quarters", "GMV 2024 ≈ ¥18B"]
        : ["带队把结账 P75 LCP 2.8s → <1.5s", "团队 6 人 → 2 季度内扩到 12", "2024 GMV ≈ 180 亿"],
      similarInterviewers: 3,
      networkNote: lang === "en" ? "Hiring manager Ma Xiaoyuan was at Star-Ring 2021-2023" : "用人经理马晓沅 2021-2023 在星环",
    },
    {
      id: "jm-3",
      title: lang === "en" ? "Frontend Tech Lead · Fintech" : "前端技术负责人 · Fintech",
      company: "Beacon Pay",
      companyTag: lang === "en" ? "Series B · payments" : "B 轮 · 支付",
      location: lang === "en" ? "Remote · China timezone" : "全远程 · 中国时区",
      posted: lang === "en" ? "1 week ago" : "1 周前",
      score: 79,
      seen: true,
      saved: false,
      comp: "¥65-90K × 14",
      source: "beaconpay.com/jobs",
      level: "P7 / Tech Lead",
      fit: { must: 4, total: 5, plus: 3, totalPlus: 4 },
      reasons: lang === "en"
        ? ["Tech-lead track fits your mentorship signal", "Remote matches preference", "TS + React + a11y — full overlap"]
        : ["TL 轨道契合你的辅导信号", "远程符合偏好", "TS + React + a11y 全对齐"],
      risks: lang === "en"
        ? ["Formal people-management — you've led but not managed", "Compliance-heavy domain (new to you)", "On-call rotation"]
        : ["正式带人——你带过但没 manage 过", "合规密集领域（对你是新的）", "有 on-call 轮值"],
      highlights: lang === "en"
        ? ["Lead 8-person frontend, hire 2 more", "Cross-functional with risk + compliance", "Fully async team"]
        : ["管 8 人前端团队，再招 2 人", "与风控 + 合规跨职能协作", "全异步团队"],
      similarInterviewers: 0,
      networkNote: null,
    },
    {
      id: "jm-4",
      title: lang === "en" ? "Senior Frontend · AI product" : "资深前端 · AI 产品",
      company: "Monolith AI",
      companyTag: lang === "en" ? "seed+ · agents" : "种子轮+ · Agent",
      location: lang === "en" ? "Shanghai / Beijing · hybrid" : "上海 / 北京 · 混合",
      posted: lang === "en" ? "3 days ago" : "3 天前",
      score: 74,
      seen: false,
      saved: false,
      comp: "¥48-68K × 13",
      source: "linkedin.com/jobs/ml-9922",
      level: "P6",
      fit: { must: 4, total: 5, plus: 1, totalPlus: 4 },
      reasons: lang === "en"
        ? ["Your prototype-to-production ratio fits early-stage", "React + RSC aligned", "Product-minded eng — matches your debrief notes"]
        : ["你原型到生产的节奏适合早期", "React + RSC 对齐", "产品感——你复盘里反复出现的信号"],
      risks: lang === "en"
        ? ["Smaller comp band — below floor", "Stack includes Python glue (bootcamp-level ok)", "No formal design team yet"]
        : ["薪资带略低于底线", "技术栈含 Python 粘合（基础够）", "尚无正式设计团队"],
      highlights: lang === "en"
        ? ["Ship the UI for a multi-agent canvas product", "Small team (12), high ownership", "Founders ex-ByteDance"]
        : ["做多 Agent 画布产品的前端", "12 人小团队，高 Ownership", "创始人前字节"],
      similarInterviewers: 1,
      networkNote: null,
    },
    {
      id: "jm-5",
      title: lang === "en" ? "Frontend · Growth pod" : "前端 · 增长组",
      company: "Meridian",
      companyTag: lang === "en" ? "Series A · social" : "A 轮 · 社交",
      location: lang === "en" ? "Shanghai · hybrid 2d" : "上海 · 每周 2 天到岗",
      posted: lang === "en" ? "2 weeks ago" : "2 周前",
      score: 61,
      seen: false,
      saved: false,
      comp: "¥35-50K × 14",
      source: "boss.zhipin.com/j/42991",
      level: "P5-P6",
      fit: { must: 4, total: 5, plus: 1, totalPlus: 4 },
      reasons: lang === "en"
        ? ["React + experimentation is in your wheelhouse", "Smaller team — more autonomy"]
        : ["React + 实验文化正是你的主场", "小团队 — 更多自主权"],
      risks: lang === "en"
        ? ["Comp is below your floor", "Level is below current", "Heavy A/B culture — you wrote about a failed one"]
        : ["薪资低于你的下限", "级别低于当前", "重 A/B 文化 — 你写过一次失败的"],
      highlights: lang === "en"
        ? ["Growth-focused pod, ships 3-5 experiments/week", "PM-lite environment", "Product-facing role"]
        : ["增长向小组，每周 3-5 次实验", "弱 PM 环境", "面向用户的角色"],
      similarInterviewers: 0,
      networkNote: null,
    },
  ];

  const watchlist = [
    { id: "w1", title: lang === "en" ? "Senior Frontend · Platform" : "资深前端 · 平台", co: "Lumen Labs", added: lang === "en" ? "2d ago" : "2 天前", change: lang === "en" ? "+1 similar role posted" : "+1 条相似岗位", tone: "ok" },
    { id: "w2", title: lang === "en" ? "Staff Frontend · E-commerce" : "资深前端 · 电商核心", co: "CloudYun", added: lang === "en" ? "5d ago" : "5 天前", change: lang === "en" ? "JD updated · check diff" : "JD 有更新 · 查看 diff", tone: "warn" },
    { id: "w3", title: lang === "en" ? "Frontend TL · Payments" : "前端 TL · 支付", co: "Beacon Pay", added: lang === "en" ? "1w ago" : "1 周前", change: lang === "en" ? "unchanged" : "无变化", tone: "muted" },
  ];

  const savedSearches = [
    { id: "s1", label: lang === "en" ? "Senior FE · RSC · Shanghai" : "资深前端 · RSC · 上海", newJobs: 3, last: lang === "en" ? "12h ago" : "12 小时前" },
    { id: "s2", label: lang === "en" ? "Tech Lead · Remote · Fintech / SaaS" : "TL · 远程 · Fintech / SaaS", newJobs: 1, last: lang === "en" ? "2d ago" : "2 天前" },
    { id: "s3", label: lang === "en" ? "AI product · Early stage" : "AI 产品 · 早期", newJobs: 5, last: lang === "en" ? "6h ago" : "6 小时前" },
  ];

  const sel = jobs.find((j) => j.id === selected);

  const runSearch = () => {
    setSearching(true);
    setTimeout(() => setSearching(false), 900);
  };

  const tabs = lang === "en"
    ? [{ k: "recommended", t: "Recommended for you", n: jobs.length }, { k: "search", t: "Search the web", n: null }, { k: "watchlist", t: "Watchlist", n: watchlist.length }]
    : [{ k: "recommended", t: "为你推荐", n: jobs.length }, { k: "search", t: "联网搜索", n: null }, { k: "watchlist", t: "关注列表", n: watchlist.length }];

  return (
    <div className="ei-fadein" style={{ maxWidth: 1320, margin: "0 auto", padding: "40px 48px 96px" }}>
      {/* Hero */}
      <div style={{ marginBottom: 28 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
          {lang === "en" ? "JD MATCH · WEB-SEARCH AGENT" : "JD 智能匹配 · 联网代理"}
        </div>
        <h1 className="ei-serif" style={{ fontSize: 38, margin: 0, color: T.ink, letterSpacing: "-0.022em", lineHeight: 1.15, maxWidth: 820 }}>
          {lang === "en" ? "We read the market — you interview the ones worth it." : "我们读市场，你只面值得的。"}
        </h1>
        <div style={{ fontSize: 14, color: T.ink3, marginTop: 10, maxWidth: 720, lineHeight: 1.5 }}>
          {lang === "en"
            ? "A background agent scans the postings that match your profile (skills · stories · comp · geography) and shortlists the ones where your sharpest evidence will actually land. Click any card to see why we think it's a fit — and where it isn't."
            : "一个后台代理持续扫描符合你画像（技能 · 故事 · 薪资 · 地域）的岗位，把「你最锋利的证据正好派得上用场」的那几个端出来。点任一卡片可以看到我们为什么判断它匹配——以及哪里不匹配。"}
        </div>
      </div>

      {/* Profile snapshot chip */}
      <div style={{ padding: "14px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderLeft: `3px solid ${T.accent}`, borderRadius: 2, marginBottom: 24, display: "flex", alignItems: "center", gap: 20, flexWrap: "wrap" }}>
        <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
          <div style={{ width: 32, height: 32, borderRadius: 16, background: T.accentSoft, color: T.accent, display: "flex", alignItems: "center", justifyContent: "center", fontFamily: "var(--ei-serif)", fontWeight: 600, fontSize: 14 }}>LZ</div>
          <div>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 2 }}>{lang === "en" ? "SEARCHING AS" : "以此画像搜索"}</div>
            <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500 }}>{profile.title} · {profile.years} {lang === "en" ? "yrs" : "年"} · {profile.location}</div>
          </div>
        </div>
        <div style={{ height: 32, width: 1, background: T.rule }} />
        <div style={{ display: "flex", gap: 5, flexWrap: "wrap" }}>
          {profile.skills.map((s) => <Tag key={s} T={T} tone="neutral">{s}</Tag>)}
        </div>
        <div style={{ flex: 1, minWidth: 40 }} />
        <div style={{ textAlign: "right", color: T.ink3, fontSize: 11.5, lineHeight: 1.45 }}>
          <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "PROFILE SOURCES" : "画像来源"}</div>
          <div>{lang === "en" ? "4 resumes · 12 JDs · 8 mocks · 2 debriefs" : "4 份简历 · 12 个 JD · 8 次模拟 · 2 次复盘"}</div>
        </div>
      </div>

      {/* Tabs */}
      <div style={{ display: "flex", gap: 0, marginBottom: 20, borderBottom: `1px solid ${T.rule}` }}>
        {tabs.map((t) => (
          <button key={t.k} onClick={() => setTab(t.k)} style={{
            padding: "12px 22px", background: "transparent", border: "none",
            borderBottom: `2px solid ${tab === t.k ? T.accent : "transparent"}`,
            color: tab === t.k ? T.ink : T.ink3, cursor: "pointer",
            fontFamily: "var(--ei-sans)", fontSize: 13.5, fontWeight: tab === t.k ? 500 : 400,
            marginBottom: -1, display: "flex", gap: 8, alignItems: "center",
          }}>
            {t.t}
            {t.n != null && <span style={{ fontFamily: "var(--ei-mono)", fontSize: 10.5, padding: "1px 6px", borderRadius: 10, background: tab === t.k ? T.accentSoft : T.bgSoft, color: tab === t.k ? T.accent : T.ink3 }}>{t.n}</span>}
          </button>
        ))}
        <div style={{ flex: 1 }} />
        <div style={{ alignSelf: "center", fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em", paddingRight: 6 }}>
          <span style={{ display: "inline-block", width: 6, height: 6, borderRadius: 3, background: T.ok, marginRight: 6, verticalAlign: "middle" }} />
          {lang === "en" ? "AGENT ACTIVE · last scan 38m ago" : "AGENT 运行中 · 38 分钟前扫过"}
        </div>
      </div>

      {/* === Recommended === */}
      {tab === "recommended" && (
        <div style={{ display: "grid", gridTemplateColumns: "1.1fr 1.4fr", gap: 20 }}>
          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {jobs.map((j) => (
              <JobMatchCard key={j.id} job={j} T={T} lang={lang} active={selected === j.id} onClick={() => setSelected(j.id)} />
            ))}
            <div style={{ padding: "16px 20px", textAlign: "center", fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", background: T.bgSoft, border: `1px dashed ${T.rule}`, borderRadius: 2 }}>
              {lang === "en" ? "agent checks every 4h · next scan in 2h 14m" : "agent 每 4 小时扫一次 · 下次 2 小时 14 分钟后"}
            </div>
          </div>
          <JDDetail job={sel} T={T} lang={lang} nav={nav} />
        </div>
      )}

      {/* === Search === */}
      {tab === "search" && (
        <SearchTab T={T} lang={lang} query={query} setQuery={setQuery} searching={searching} runSearch={runSearch} savedSearches={savedSearches} jobs={jobs} />
      )}

      {/* === Watchlist === */}
      {tab === "watchlist" && (
        <WatchlistTab T={T} lang={lang} watchlist={watchlist} jobs={jobs} />
      )}
    </div>
  );
};

// ─── Job match card ───
const JobMatchCard = ({ job, T, lang, active, onClick }) => {
  const scoreC = job.score >= 85 ? T.ok : job.score >= 70 ? T.warn : T.ink3;
  const scoreLabel = job.score >= 85 ? (lang === "en" ? "STRONG FIT" : "强匹配") : job.score >= 70 ? (lang === "en" ? "GOOD FIT" : "较合适") : (lang === "en" ? "STRETCH" : "挑战型");
  return (
    <button onClick={onClick} style={{
      padding: "18px 20px", textAlign: "left", cursor: "pointer",
      background: active ? T.accentSoft : T.bgCard,
      border: `1px solid ${active ? T.accent : T.rule}`,
      borderLeft: `3px solid ${scoreC}`,
      borderRadius: 2, fontFamily: "var(--ei-sans)",
    }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 14, marginBottom: 8 }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 4 }}>
            <div className="ei-serif" style={{ fontSize: 16.5, color: T.ink, letterSpacing: "-0.01em", fontWeight: 500 }}>{job.title}</div>
            {!job.seen && <div style={{ width: 6, height: 6, borderRadius: 3, background: T.accent }} />}
            {job.saved && <Icon name="pin" size={11} color={T.accent} />}
          </div>
          <div style={{ fontSize: 12.5, color: T.ink2, marginBottom: 2 }}>{job.company} <span style={{ color: T.ink3 }}>· {job.companyTag}</span></div>
          <div style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.02em" }}>{job.location} · {job.comp}</div>
        </div>
        <div style={{ textAlign: "right", flexShrink: 0 }}>
          <div className="ei-serif" style={{ fontSize: 32, color: scoreC, fontWeight: 500, lineHeight: 1, letterSpacing: "-0.02em" }}>{job.score}</div>
          <div style={{ fontSize: 9.5, color: scoreC, fontFamily: "var(--ei-mono)", letterSpacing: "0.08em", marginTop: 3 }}>{scoreLabel}</div>
        </div>
      </div>
      {/* Top reason */}
      <div style={{ marginTop: 10, padding: "8px 10px", background: active ? T.bg : T.bgSoft, borderLeft: `2px solid ${T.accent}`, fontSize: 12.5, color: T.ink2, lineHeight: 1.45 }}>
        <span style={{ color: T.accent, fontWeight: 500 }}>{lang === "en" ? "Top reason · " : "首要原因 · "}</span>
        {job.reasons[0]}
      </div>
      <div style={{ marginTop: 10, display: "flex", gap: 14, fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>
        <span>{lang === "en" ? "must" : "必需"} {job.fit.must}/{job.fit.total}</span>
        <span>+ {lang === "en" ? "plus" : "加分"} {job.fit.plus}/{job.fit.totalPlus}</span>
        <span>{job.posted}</span>
        <span style={{ flex: 1, textAlign: "right", color: T.ink4 }}>{job.source}</span>
      </div>
    </button>
  );
};

// ─── Detail drawer ───
const JDDetail = ({ job, T, lang, nav }) => {
  if (!job) return null;
  const scoreC = job.score >= 85 ? T.ok : job.score >= 70 ? T.warn : T.ink3;

  return (
    <div style={{ position: "sticky", top: 20, alignSelf: "start" }}>
      <Card T={T} pad={0}>
        {/* Header */}
        <div style={{ padding: "20px 24px 18px", borderBottom: `1px solid ${T.rule}` }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 14 }}>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{job.company} · {job.companyTag}</div>
              <div className="ei-serif" style={{ fontSize: 24, color: T.ink, letterSpacing: "-0.015em", lineHeight: 1.2, marginBottom: 8 }}>{job.title}</div>
              <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
                <Tag T={T} tone="neutral">{job.level}</Tag>
                <Tag T={T} tone="neutral">{job.location}</Tag>
                <Tag T={T} tone="accent">{job.comp}</Tag>
              </div>
            </div>
            <div style={{ textAlign: "center", padding: "6px 14px", background: T.bgSoft, borderRadius: 2, border: `1px solid ${T.rule}` }}>
              <div className="ei-serif" style={{ fontSize: 36, color: scoreC, fontWeight: 500, lineHeight: 1, letterSpacing: "-0.02em" }}>{job.score}</div>
              <div style={{ fontSize: 9.5, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.08em", marginTop: 4 }}>/ 100</div>
            </div>
          </div>
        </div>

        {/* Why it matches */}
        <div style={{ padding: "18px 24px", borderBottom: `1px dotted ${T.rule}` }}>
          <div className="ei-label" style={{ color: T.ok, marginBottom: 10 }}>{lang === "en" ? "+ WHY IT MATCHES" : "+ 为什么匹配"}</div>
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            {job.reasons.map((r, i) => (
              <div key={i} style={{ display: "flex", gap: 10, fontSize: 13, color: T.ink, lineHeight: 1.5 }}>
                <Icon name="check" size={12} color={T.ok} stroke={2.5} style={{ marginTop: 3, flexShrink: 0 }} />
                <span>{r}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Risks */}
        <div style={{ padding: "18px 24px", borderBottom: `1px dotted ${T.rule}` }}>
          <div className="ei-label" style={{ color: T.warn, marginBottom: 10 }}>{lang === "en" ? "⚠ WHERE IT'S A STRETCH" : "⚠ 需要留意的地方"}</div>
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            {job.risks.map((r, i) => (
              <div key={i} style={{ display: "flex", gap: 10, fontSize: 13, color: T.ink2, lineHeight: 1.5 }}>
                <Icon name="info" size={12} color={T.warn} style={{ marginTop: 3, flexShrink: 0 }} />
                <span>{r}</span>
              </div>
            ))}
          </div>
        </div>

        {/* JD highlights */}
        <div style={{ padding: "18px 24px", borderBottom: `1px dotted ${T.rule}` }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "ROLE SNAPSHOT" : "岗位快照"}</div>
          <ul style={{ margin: 0, paddingLeft: 18, fontSize: 13, color: T.ink2, lineHeight: 1.7 }}>
            {job.highlights.map((h, i) => <li key={i}>{h}</li>)}
          </ul>
        </div>

        {/* Network / intel */}
        {(job.similarInterviewers > 0 || job.networkNote) && (
          <div style={{ padding: "18px 24px", borderBottom: `1px dotted ${T.rule}`, background: T.bgSoft }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "INTEL" : "情报"}</div>
            <div style={{ display: "flex", flexDirection: "column", gap: 8, fontSize: 12.5, color: T.ink2, lineHeight: 1.5 }}>
              {job.networkNote && (
                <div style={{ display: "flex", gap: 8 }}>
                  <Icon name="target" size={12} color={T.accent} style={{ marginTop: 3, flexShrink: 0 }} />
                  <span>{job.networkNote}</span>
                </div>
              )}
              {job.similarInterviewers > 0 && (
                <div style={{ display: "flex", gap: 8 }}>
                  <Icon name="book" size={12} color={T.accent} style={{ marginTop: 3, flexShrink: 0 }} />
                  <span>{lang === "en" ? `${job.similarInterviewers} interviewer profiles surfaced from public sources — glassdoor + 知乎` : `从公开来源（Glassdoor + 知乎）抓到 ${job.similarInterviewers} 位面试官公开信息`}</span>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Action bar */}
        <div style={{ padding: "16px 24px", display: "flex", gap: 10, flexWrap: "wrap" }}>
          <Btn T={T} variant="accent" icon="arrow_right" onClick={() => nav("parse")}>
            {lang === "en" ? "Confirm interview" : "确认面试"}
          </Btn>
          <Btn T={T} variant="secondary" size="sm" icon={job.saved ? "pin" : "plus"}>
            {job.saved ? (lang === "en" ? "Saved" : "已关注") : (lang === "en" ? "Save to watchlist" : "加入关注")}
          </Btn>
          <div style={{ flex: 1 }} />
          <Btn T={T} variant="ghost" size="sm" icon="link">
            {lang === "en" ? "Source" : "原文"}
          </Btn>
          <Btn T={T} variant="ghost" size="sm" icon="x">
            {lang === "en" ? "Not relevant" : "不相关"}
          </Btn>
        </div>
      </Card>
    </div>
  );
};

// ─── Search tab ───
const SearchTab = ({ T, lang, query, setQuery, searching, runSearch, savedSearches, jobs }) => {
  const sources = lang === "en"
    ? [{ k: "li", t: "LinkedIn", n: 42 }, { k: "boss", t: "Boss 直聘", n: 128 }, { k: "maimai", t: "脉脉", n: 36 }, { k: "lagou", t: "拉勾", n: 24 }, { k: "company", t: lang === "en" ? "Company sites" : "公司官网", n: 18 }]
    : [{ k: "li", t: "LinkedIn", n: 42 }, { k: "boss", t: "Boss 直聘", n: 128 }, { k: "maimai", t: "脉脉", n: 36 }, { k: "lagou", t: "拉勾", n: 24 }, { k: "company", t: "公司官网", n: 18 }];

  return (
    <div>
      {/* Search bar */}
      <div style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, padding: "20px 24px", marginBottom: 20 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "NATURAL LANGUAGE SEARCH" : "自然语言搜索"}</div>
        <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
          <div style={{ flex: 1, position: "relative" }}>
            <Icon name="search" size={14} color={T.ink3} style={{ position: "absolute", left: 12, top: "50%", transform: "translateY(-50%)" }} />
            <input value={query} onChange={(e) => setQuery(e.target.value)} style={{
              width: "100%", padding: "12px 14px 12px 36px", fontSize: 14, color: T.ink,
              background: T.bg, border: `1px solid ${T.rule}`, borderRadius: 2,
              fontFamily: "var(--ei-sans)", outline: "none", boxSizing: "border-box",
            }} />
          </div>
          <Btn T={T} variant="accent" icon={searching ? "" : "search"} onClick={runSearch} disabled={searching}>
            {searching ? (lang === "en" ? "Scanning…" : "扫描中…") : (lang === "en" ? "Run web search" : "联网搜索")}
          </Btn>
        </div>
        {/* Source chips */}
        <div style={{ display: "flex", gap: 6, marginTop: 14, flexWrap: "wrap", alignItems: "center" }}>
          <span style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.06em", marginRight: 4 }}>{lang === "en" ? "SOURCES" : "数据源"}</span>
          {sources.map((s) => (
            <Tag key={s.k} T={T} tone="neutral">{s.t} <span style={{ color: T.ink3, marginLeft: 4 }}>{s.n}</span></Tag>
          ))}
        </div>
      </div>

      {/* Live agent status while searching */}
      {searching && (
        <div style={{ background: T.bgCard, border: `1px solid ${T.accent}`, borderLeft: `3px solid ${T.accent}`, padding: "18px 22px", marginBottom: 20, borderRadius: 2 }}>
          <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>{lang === "en" ? "● AGENT SCANNING" : "● AGENT 扫描中"}</div>
          <div style={{ display: "flex", flexDirection: "column", gap: 6, fontSize: 12.5, fontFamily: "var(--ei-mono)", color: T.ink2 }}>
            {[
              lang === "en" ? "→ Embedding your profile…" : "→ 向量化你的画像…",
              lang === "en" ? "→ Querying LinkedIn · Boss · Maimai · Lagou in parallel…" : "→ 并行查询 LinkedIn · Boss · 脉脉 · 拉勾…",
              lang === "en" ? "→ Dedup by JD hash · 248 → 87 unique postings" : "→ 按 JD 哈希去重 · 248 → 87 条唯一岗位",
              lang === "en" ? "→ Scoring each against profile · strengths · gaps · prefs" : "→ 按画像/优势/盲点/偏好逐条打分",
              lang === "en" ? "→ Drafting match reasoning…" : "→ 起草匹配理由…",
            ].map((l, i) => (
              <div key={i} style={{ opacity: i <= 2 ? 1 : 0.4 }}>{l}</div>
            ))}
          </div>
        </div>
      )}

      {/* Saved searches */}
      <div style={{ marginBottom: 24 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 10 }}>
          <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "SAVED SEARCHES · runs in background" : "已保存搜索 · 后台自动运行"}</div>
          <button style={{ background: "transparent", border: "none", color: T.accent, fontSize: 12.5, cursor: "pointer" }}>
            <Icon name="plus" size={11} style={{ marginRight: 4 }} /> {lang === "en" ? "Save current as watch" : "保存为自动搜索"}
          </button>
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 10 }}>
          {savedSearches.map((s) => (
            <div key={s.id} style={{ padding: "14px 16px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 8, marginBottom: 8 }}>
                <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 500, lineHeight: 1.3 }}>{s.label}</div>
                {s.newJobs > 0 && <div style={{ padding: "2px 7px", background: T.accentSoft, color: T.accent, fontSize: 10.5, fontFamily: "var(--ei-mono)", borderRadius: 2 }}>+{s.newJobs}</div>}
              </div>
              <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>{lang === "en" ? "last · " : "上次 · "}{s.last}</div>
            </div>
          ))}
        </div>
      </div>

      {/* Results */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
        <div className="ei-label" style={{ color: T.ink3 }}>
          {lang === "en" ? `RESULTS · ${jobs.length} ranked by fit` : `结果 · 按匹配度排序 · ${jobs.length} 条`}
        </div>
        <div style={{ display: "flex", gap: 6 }}>
          {(lang === "en" ? ["All", "Strong fit (85+)", "Remote-friendly", "Unseen"] : ["全部", "强匹配 (85+)", "支持远程", "未看过"]).map((f, i) => (
            <button key={f} style={{
              padding: "4px 10px", fontSize: 11.5, borderRadius: 12, cursor: "pointer",
              border: `1px solid ${i === 0 ? T.accent : T.rule}`,
              background: i === 0 ? T.accentSoft : "transparent",
              color: i === 0 ? T.accent : T.ink3, fontFamily: "var(--ei-sans)",
            }}>{f}</button>
          ))}
        </div>
      </div>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10 }}>
        {jobs.slice(0, 4).map((j) => <JobMatchCard key={j.id} job={j} T={T} lang={lang} active={false} onClick={() => {}} />)}
      </div>
    </div>
  );
};

// ─── Watchlist tab ───
const WatchlistTab = ({ T, lang, watchlist, jobs }) => {
  return (
    <div>
      <div style={{ marginBottom: 28 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "JOBS YOU'RE TRACKING" : "你正在跟踪的岗位"}</div>
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          {watchlist.map((w) => {
            const toneC = { ok: T.ok, warn: T.warn, muted: T.ink3 }[w.tone];
            return (
              <div key={w.id} style={{ padding: "16px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderLeft: `3px solid ${toneC}`, borderRadius: 2, display: "flex", alignItems: "center", gap: 16 }}>
                <div style={{ flex: 1 }}>
                  <div className="ei-serif" style={{ fontSize: 16, color: T.ink, letterSpacing: "-0.01em", marginBottom: 4 }}>{w.title}</div>
                  <div style={{ fontSize: 12.5, color: T.ink3 }}>{w.co} · {lang === "en" ? "added " : "加入于 "}{w.added}</div>
                </div>
                <div style={{ fontSize: 12.5, color: toneC, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em" }}>
                  <span style={{ display: "inline-block", width: 6, height: 6, borderRadius: 3, background: toneC, marginRight: 8, verticalAlign: "middle" }} />
                  {w.change}
                </div>
                <Btn T={T} variant="ghost" size="sm" icon="chevron_right" />
              </div>
            );
          })}
        </div>
      </div>

      {/* Market signals */}
      <div>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "MARKET SIGNALS · last 7 days" : "市场信号 · 过去 7 天"}</div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 10, marginBottom: 20 }}>
          {[
            { k: lang === "en" ? "New postings matching profile" : "符合画像的新岗位", v: "23", d: "+6", tone: "ok" },
            { k: lang === "en" ? "Avg. salary band (same level)" : "同级薪资带中位", v: "¥58K", d: "+3%", tone: "ok" },
            { k: lang === "en" ? "Median response time" : "平均回复时间", v: "4.2d", d: "-0.8", tone: "ok" },
            { k: lang === "en" ? "RSC mentions in JDs" : "JD 提 RSC 的比例", v: "34%", d: "+12pp", tone: "accent" },
          ].map((m) => {
            const c = { ok: T.ok, accent: T.accent }[m.tone];
            return (
              <div key={m.k} style={{ padding: "14px 16px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{m.k}</div>
                <div style={{ display: "flex", alignItems: "baseline", gap: 8 }}>
                  <div className="ei-serif" style={{ fontSize: 24, color: T.ink, letterSpacing: "-0.015em" }}>{m.v}</div>
                  <div style={{ fontSize: 11.5, color: c, fontFamily: "var(--ei-mono)" }}>{m.d}</div>
                </div>
              </div>
            );
          })}
        </div>
        <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", textAlign: "center", padding: "10px", background: T.bgSoft, borderRadius: 2 }}>
          {lang === "en" ? "signals computed from ~340 postings scanned this week · refreshed every 4h" : "基于本周扫描的 ~340 条岗位计算 · 每 4 小时刷新"}
        </div>
      </div>
    </div>
  );
};

window.JDMatchScreen = JDMatchScreen;
