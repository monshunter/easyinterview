// User Profile · AI-inferred candidate profile with user corrections

const UserProfileScreen = ({ T, lang, nav }) => {
  const [activeSection, setActiveSection] = React.useState("positioning");
  const [editor, setEditor] = React.useState(null);
  const [patches, setPatches] = React.useState({});
  const [draft, setDraft] = React.useState("");

  const data = getUserProfileData(lang);
  const active = data.sections.find((section) => section.id === activeSection) || data.sections[0];

  const openEditor = (item) => {
    setEditor(item);
    setDraft(patches[item.id] || item.value || item.title);
  };

  const savePatch = () => {
    if (!editor) return;
    setPatches((next) => ({ ...next, [editor.id]: draft }));
    setEditor(null);
    setDraft("");
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 1280, margin: "0 auto", padding: "42px 48px 96px" }}>
      <button onClick={() => nav("home")} style={{
        background: "transparent", border: "none", color: T.ink3, fontSize: 13,
        display: "inline-flex", alignItems: "center", gap: 6, padding: 0, marginBottom: 24,
      }}>
        <Icon name="arrow_left" size={13} /> {lang === "en" ? "Back home" : "返回首页"}
      </button>

      <div style={{ display: "grid", gridTemplateColumns: "1.15fr 0.85fr", gap: 32, alignItems: "end", marginBottom: 28 }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
            {lang === "en" ? "USER PROFILE · AI SUMMARY" : "用户画像 · AI 汇总"}
          </div>
          <h1 className="ei-serif" style={{ margin: 0, fontSize: 40, lineHeight: 1.12, letterSpacing: "-0.024em", color: T.ink }}>
            {lang === "en" ? "The profile the system uses to understand you." : "系统理解你的那份画像。"}
          </h1>
          <div style={{ marginTop: 12, maxWidth: 720, fontSize: 14, lineHeight: 1.65, color: T.ink3 }}>
            {lang === "en"
              ? "This is inferred from resumes, target jobs, mock interviews, real debriefs, and your corrections. It is not a required onboarding form; every item can be inspected and corrected."
              : "这不是一个需要用户先填完的表单，而是系统根据简历、目标岗位、模拟面试、复盘和用户纠偏持续沉淀出的结构化画像。每一项都应该能看来源、能修正、能决定是否用于推荐和面试。"}
          </div>
        </div>

        <Card T={T} pad={18} style={{ background: T.bgSoft }}>
          <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
            <div style={{ width: 46, height: 46, borderRadius: 23, background: T.accentSoft, color: T.accent, display: "flex", alignItems: "center", justifyContent: "center", fontFamily: "var(--ei-serif)", fontSize: 18, fontWeight: 600 }}>LZ</div>
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: 14, color: T.ink, fontWeight: 600 }}>{data.identity.name}</div>
              <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 2 }}>{data.identity.summary}</div>
            </div>
            <Tag T={T} tone="ok">{data.identity.confidence}</Tag>
          </div>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 8, marginTop: 16 }}>
            {data.sources.map((source) => (
              <div key={source.label} style={{ padding: "10px 8px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, textAlign: "center" }}>
                <div className="ei-serif" style={{ fontSize: 22, color: T.ink, lineHeight: 1 }}>{source.count}</div>
                <div style={{ fontSize: 10.5, color: T.ink3, marginTop: 5 }}>{source.label}</div>
              </div>
            ))}
          </div>
        </Card>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "260px 1fr", gap: 22, alignItems: "start" }}>
        <aside style={{ position: "sticky", top: 86 }}>
          <Card T={T} pad={0}>
            {data.sections.map((section) => {
              const active = section.id === activeSection;
              return (
                <button key={section.id} onClick={() => setActiveSection(section.id)} style={{
                  width: "100%", border: "none", borderBottom: `1px solid ${T.rule}`,
                  background: active ? T.accentSoft : "transparent",
                  color: active ? T.ink : T.ink2, textAlign: "left", padding: "14px 16px",
                  display: "flex", alignItems: "center", gap: 10,
                }}>
                  <Icon name={section.icon} size={14} color={active ? T.accent : T.ink3} />
                  <div>
                    <div style={{ fontSize: 13.5, fontWeight: active ? 600 : 500 }}>{section.title}</div>
                    <div style={{ fontSize: 11.5, color: active ? T.accent : T.ink3, marginTop: 2 }}>{section.summary}</div>
                  </div>
                </button>
              );
            })}
          </Card>

          <Card T={T} pad={16} style={{ marginTop: 14 }}>
            <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "USED BY" : "被哪些模块使用"}</div>
            <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
              {data.usage.map((item) => (
                <UsageRow key={item.label} T={T} item={item} />
              ))}
            </div>
          </Card>
        </aside>

        <main style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <Card T={T} pad={0}>
            <div style={{ padding: "18px 22px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", gap: 16 }}>
              <div>
                <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{active.kicker}</div>
                <div className="ei-serif" style={{ fontSize: 24, color: T.ink }}>{active.title}</div>
              </div>
              <Btn T={T} variant="secondary" size="sm" icon="plus" onClick={() => openEditor({ id: `${active.id}-new`, title: lang === "en" ? "Add correction" : "新增补充", value: "" })}>
                {lang === "en" ? "Add note" : "补充一条"}
              </Btn>
            </div>

            <div style={{ padding: "20px 22px", display: "grid", gridTemplateColumns: active.layout === "wide" ? "1fr" : "1fr 1fr", gap: 12 }}>
              {active.items.map((item) => (
                <ProfileInsightCard key={item.id} T={T} lang={lang} item={item} patch={patches[item.id]} onEdit={() => openEditor(item)} />
              ))}
            </div>
          </Card>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
            <Card T={T} pad={18}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "RECENT EVIDENCE" : "最近证据来源"}</div>
              <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                {data.evidence.map((item) => (
                  <div key={item.title} style={{ paddingBottom: 10, borderBottom: `1px dotted ${T.rule}` }}>
                    <div style={{ display: "flex", justifyContent: "space-between", gap: 10 }}>
                      <div style={{ fontSize: 13, color: T.ink, fontWeight: 500 }}>{item.title}</div>
                      <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{item.time}</div>
                    </div>
                    <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.55, marginTop: 4 }}>{item.body}</div>
                  </div>
                ))}
              </div>
            </Card>

            <Card T={T} pad={18}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "CORRECTION RULE" : "纠偏原则"}</div>
              <div className="ei-serif" style={{ fontSize: 21, lineHeight: 1.35, color: T.ink, marginBottom: 12 }}>
                {lang === "en" ? "User corrections outrank AI inference." : "用户修正优先级高于 AI 推断。"}
              </div>
              <div style={{ fontSize: 13, color: T.ink3, lineHeight: 1.65 }}>
                {lang === "en"
                  ? "When you correct a field, the system keeps the original evidence, stores the correction as a higher-priority layer, and uses it in mock interview planning, report rubrics, and debrief analysis."
                  : "用户修正不会删除原始证据，而是作为更高优先级的一层覆盖推断结果。模拟面试规划、报告分析维度和复盘分析都会优先读取这层修正。"}
              </div>
            </Card>
          </div>
        </main>
      </div>

      {editor && (
        <ProfileCorrectionModal
          T={T}
          lang={lang}
          item={editor}
          draft={draft}
          setDraft={setDraft}
          onClose={() => setEditor(null)}
          onSave={savePatch}
        />
      )}
    </div>
  );
};

const UsageRow = ({ T, item }) => (
  <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 10 }}>
    <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
      <span style={{ width: 7, height: 7, borderRadius: 4, background: item.on ? T.ok : T.ink4 }} />
      <span style={{ fontSize: 12.5, color: T.ink2 }}>{item.label}</span>
    </div>
    <span style={{ fontSize: 11, color: item.on ? T.ok : T.ink3, fontFamily: "var(--ei-mono)" }}>{item.on ? "ON" : "OFF"}</span>
  </div>
);

const ProfileInsightCard = ({ T, lang, item, patch, onEdit }) => {
  const value = patch || item.value;
  return (
    <div style={{
      border: `1px solid ${patch ? T.accent : T.rule}`, background: patch ? T.accentSoft : T.bgCard,
      borderRadius: 3, padding: 16, minHeight: 150, display: "flex", flexDirection: "column",
    }}>
      <div style={{ display: "flex", justifyContent: "space-between", gap: 12, alignItems: "flex-start" }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{item.kicker}</div>
          <div className="ei-serif" style={{ fontSize: 19, color: T.ink, lineHeight: 1.25 }}>{item.title}</div>
        </div>
        <button onClick={onEdit} title={lang === "en" ? "Correct" : "修正"} style={{
          width: 30, height: 30, borderRadius: 2, border: `1px solid ${T.rule}`,
          background: T.bg, color: T.ink2, display: "flex", alignItems: "center", justifyContent: "center",
        }}>
          <Icon name="edit" size={13} />
        </button>
      </div>
      <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.58, marginTop: 12, flex: 1 }}>{value}</div>
      <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginTop: 12, alignItems: "center" }}>
        {patch && <Tag T={T} tone="accent">{lang === "en" ? "USER CORRECTED" : "用户修正"}</Tag>}
        <Tag T={T} tone={item.confidenceTone || "neutral"}>{item.confidence}</Tag>
        <span style={{ fontSize: 11.5, color: T.ink3 }}>{item.source}</span>
      </div>
    </div>
  );
};

const ProfileCorrectionModal = ({ T, lang, item, draft, setDraft, onClose, onSave }) => (
  <div style={{ position: "fixed", inset: 0, zIndex: 90, background: "rgba(20,15,10,0.34)", display: "flex", justifyContent: "flex-end" }}>
    <div onClick={onClose} style={{ position: "absolute", inset: 0 }} />
    <div style={{ position: "relative", width: 460, maxWidth: "100%", minHeight: "100%", background: T.bgCard, borderLeft: `1px solid ${T.rule}`, boxShadow: "-20px 0 46px rgba(20,15,10,0.18)", padding: 28 }}>
      <div style={{ display: "flex", justifyContent: "space-between", gap: 16, alignItems: "flex-start", marginBottom: 24 }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "CORRECT PROFILE ITEM" : "修正画像信息"}</div>
          <div className="ei-serif" style={{ fontSize: 24, color: T.ink }}>{item.title}</div>
        </div>
        <button onClick={onClose} style={{ border: `1px solid ${T.rule}`, background: T.bg, width: 32, height: 32, borderRadius: 2, color: T.ink2 }}>
          <Icon name="x" size={14} />
        </button>
      </div>

      <div style={{ marginBottom: 18 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "AI INFERENCE" : "AI 当前推断"}</div>
        <div style={{ padding: 14, background: T.bgSoft, border: `1px solid ${T.rule}`, color: T.ink2, fontSize: 13, lineHeight: 1.6 }}>
          {item.value || (lang === "en" ? "No value yet." : "还没有内容。")}
        </div>
      </div>

      <div style={{ marginBottom: 18 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "YOUR CORRECTION" : "你的修正或补充"}</div>
        <textarea value={draft} onChange={(e) => setDraft(e.target.value)} rows={8} style={{
          width: "100%", resize: "vertical", border: `1px solid ${T.rule}`, background: T.bg,
          color: T.ink, padding: 12, fontSize: 13.5, lineHeight: 1.55, outline: "none",
        }} />
      </div>

      <div style={{ padding: 14, background: T.bgSoft, border: `1px solid ${T.rule}`, fontSize: 12.5, color: T.ink3, lineHeight: 1.6, marginBottom: 20 }}>
        {lang === "en"
          ? "Saving creates a higher-priority user correction. The original evidence remains available for audit."
          : "保存后会生成一条高优先级的用户修正。原始证据仍然保留，用于回溯和审计。"}
      </div>

      <div style={{ display: "flex", gap: 10 }}>
        <Btn T={T} variant="accent" icon="check" onClick={onSave}>{lang === "en" ? "Save correction" : "保存修正"}</Btn>
        <Btn T={T} variant="secondary" onClick={onClose}>{lang === "en" ? "Cancel" : "取消"}</Btn>
      </div>
    </div>
  </div>
);

const getUserProfileData = (lang) => ({
  identity: {
    name: lang === "en" ? "Liu Zhe" : "刘哲",
    summary: lang === "en" ? "Senior frontend · platform / e-commerce · Shanghai / hybrid" : "资深前端 · 平台 / 电商 · 上海 / 混合办公",
    confidence: lang === "en" ? "HIGH CONFIDENCE" : "高置信",
  },
  sources: [
    { label: lang === "en" ? "Resumes" : "简历", count: 4 },
    { label: "JD", count: 12 },
    { label: lang === "en" ? "Mocks" : "模拟", count: 8 },
    { label: lang === "en" ? "Debriefs" : "复盘", count: 2 },
  ],
  usage: [
    { label: lang === "en" ? "Mock interview planning" : "模拟面试规划", on: true },
    { label: lang === "en" ? "Report rubrics" : "报告分析维度", on: true },
    { label: lang === "en" ? "Public sharing" : "公开分享", on: false },
  ],
  sections: [
    {
      id: "positioning",
      icon: "target",
      title: lang === "en" ? "Career positioning" : "职业定位",
      summary: lang === "en" ? "Role · level · market" : "方向 · 级别 · 市场",
      kicker: lang === "en" ? "POSITIONING" : "POSITIONING",
      layout: "two",
      items: [
        {
          id: "role-target",
          kicker: lang === "en" ? "TARGET ROLE" : "目标角色",
          title: lang === "en" ? "Senior Frontend Engineer" : "资深前端工程师",
          value: lang === "en" ? "Best fit is senior IC or small platform lead roles, especially platform UI, e-commerce checkout, design systems, and performance-sensitive product surfaces." : "最适合资深 IC 或小型平台前端负责人角色，重点是平台 UI、电商结账链路、Design System 和性能敏感型业务界面。",
          confidence: lang === "en" ? "HIGH" : "高置信",
          confidenceTone: "ok",
          source: lang === "en" ? "from 4 resumes + 12 JDs" : "来自 4 份简历 + 12 个 JD",
        },
        {
          id: "level-band",
          kicker: lang === "en" ? "LEVEL BAND" : "级别区间",
          title: "P6 / P7",
          value: lang === "en" ? "Current evidence supports P6 strongly and P7 selectively when the role values platform influence, cross-team rollout, and performance ownership." : "当前证据稳定支撑 P6；当岗位重视平台影响力、跨团队落地和性能 ownership 时，可以选择性冲 P7。",
          confidence: lang === "en" ? "MEDIUM" : "中置信",
          confidenceTone: "amber",
          source: lang === "en" ? "from job matches + reports" : "来自岗位匹配 + 面试报告",
        },
        {
          id: "market-pref",
          kicker: lang === "en" ? "MARKET PREFERENCE" : "市场偏好",
          title: lang === "en" ? "Shanghai · hybrid / remote" : "上海 · 混合 / 远程优先",
          value: lang === "en" ? "Hybrid is acceptable. Full on-site roles should be treated as a trade-off unless compensation or scope is unusually strong." : "接受混合办公；全职到岗岗位默认应作为取舍项，除非薪资或职责范围明显更强。",
          confidence: lang === "en" ? "USER SIGNAL" : "用户信号",
          confidenceTone: "accent",
          source: lang === "en" ? "from user corrections" : "来自用户修正",
        },
        {
          id: "comp-floor",
          kicker: lang === "en" ? "COMPENSATION" : "薪资约束",
          title: lang === "en" ? "¥50K floor · equity friendly" : "底线 5 万/月 · 接受期权",
          value: lang === "en" ? "Roles below the floor should be treated as trade-offs unless they offer exceptional scope, remote flexibility, or strong strategic upside." : "低于底线的岗位默认视为取舍项，除非职责范围、远程灵活性或战略机会明显更强。",
          confidence: lang === "en" ? "MEDIUM" : "中置信",
          confidenceTone: "amber",
          source: lang === "en" ? "from imported JDs" : "来自导入的 JD",
        },
      ],
    },
    {
      id: "skills",
      icon: "spark",
      title: lang === "en" ? "Skills and depth" : "技能与深度",
      summary: lang === "en" ? "Stack · proof · gaps" : "栈 · 证据 · 缺口",
      kicker: lang === "en" ? "SKILLS" : "SKILLS",
      layout: "two",
      items: [
        {
          id: "react-depth",
          kicker: "React / RSC",
          title: lang === "en" ? "Strong production evidence" : "有强生产证据",
          value: lang === "en" ? "RSC migration, selective hydration, and performance impact are backed by concrete business and metric evidence." : "RSC 迁移、选择性注水和性能收益都有具体业务和指标证据支撑。",
          confidence: lang === "en" ? "HIGH" : "高置信",
          confidenceTone: "ok",
          source: lang === "en" ? "from resume bullets + reports" : "来自简历 bullet + 报告",
        },
        {
          id: "design-system",
          kicker: "Design System",
          title: lang === "en" ? "Rollout story needs sharper ownership" : "落地故事需要更清晰 ownership",
          value: lang === "en" ? "There is evidence of adoption across products, but interviews still need clearer ownership, conflict, and measurable adoption metrics." : "已有跨产品落地证据，但面试表达里还需要更清晰的 ownership、冲突处理和采用指标。",
          confidence: lang === "en" ? "MEDIUM" : "中置信",
          confidenceTone: "amber",
          source: lang === "en" ? "from mock interview gaps" : "来自模拟面试缺口",
        },
        {
          id: "a11y",
          kicker: "a11y",
          title: lang === "en" ? "Useful differentiator, not core identity" : "可作为差异化，不是主身份",
          value: lang === "en" ? "Accessibility is a useful matching boost for platform roles, but the profile should not over-index on it unless the JD names it." : "可访问性可作为平台岗位的加分项，但除非 JD 明确要求，否则不应成为匹配判断的主权重。",
          confidence: lang === "en" ? "MEDIUM" : "中置信",
          confidenceTone: "amber",
          source: lang === "en" ? "from JD matching history" : "来自 JD 匹配历史",
        },
        {
          id: "risk-skill",
          kicker: lang === "en" ? "RISK" : "风险",
          title: lang === "en" ? "Management evidence is thin" : "正式管理证据较弱",
          value: lang === "en" ? "Mentoring and cross-team influence are visible, but formal people management should be treated as a stretch." : "辅导新人和跨团队影响力可见，但正式带人管理仍应被视为挑战项。",
          confidence: lang === "en" ? "HIGH" : "高置信",
          confidenceTone: "danger",
          source: lang === "en" ? "from reports + debrief" : "来自报告 + 复盘",
        },
      ],
    },
    {
      id: "evidence",
      icon: "book",
      title: lang === "en" ? "Experience evidence" : "经历证据",
      summary: lang === "en" ? "Stories to reuse" : "可复用故事",
      kicker: lang === "en" ? "STORIES" : "STORIES",
      layout: "wide",
      items: [
        {
          id: "story-checkout",
          kicker: lang === "en" ? "PRIMARY STORY" : "主故事",
          title: lang === "en" ? "Checkout performance rewrite" : "结账链路性能重写",
          value: lang === "en" ? "Strongest reusable story: migration to RSC, performance baseline and outcome, cross-functional coordination, and business impact. Use for performance, architecture, and ownership questions." : "最强可复用故事：RSC 迁移、性能基线与结果、跨职能协作、业务影响。适用于性能、架构和 ownership 类问题。",
          confidence: lang === "en" ? "HIGH" : "高置信",
          confidenceTone: "ok",
          source: lang === "en" ? "from resume v3 + 3 mocks" : "来自简历 v3 + 3 次模拟",
        },
        {
          id: "story-ds",
          kicker: lang === "en" ? "SECONDARY STORY" : "次级故事",
          title: lang === "en" ? "Design System rollout" : "Design System 落地",
          value: lang === "en" ? "Useful story for influence and platform roles. Needs a sharper before/after and stakeholder conflict frame." : "适合回答影响力和平台建设类问题，但还需要更明确的前后对比和 stakeholder 冲突框架。",
          confidence: lang === "en" ? "MEDIUM" : "中置信",
          confidenceTone: "amber",
          source: lang === "en" ? "from resume workshop" : "来自简历工坊",
        },
      ],
    },
    {
      id: "interview",
      icon: "chat",
      title: lang === "en" ? "Interview behavior" : "面试表现",
      summary: lang === "en" ? "Strengths · habits" : "优势 · 习惯",
      kicker: lang === "en" ? "INTERVIEW SIGNALS" : "INTERVIEW SIGNALS",
      layout: "two",
      items: [
        {
          id: "answer-style",
          kicker: lang === "en" ? "ANSWER STYLE" : "回答风格",
          title: lang === "en" ? "Clear context, late metrics" : "上下文清楚，指标偏晚",
          value: lang === "en" ? "Answers usually start with useful context, but metrics and trade-offs often arrive too late. Reports recommend leading with baseline, action, outcome." : "回答通常能给出上下文，但指标和取舍经常出现得太晚。报告建议先给基线、动作和结果。",
          confidence: lang === "en" ? "HIGH" : "高置信",
          confidenceTone: "ok",
          source: lang === "en" ? "from 8 mock reports" : "来自 8 份模拟报告",
        },
        {
          id: "followup-risk",
          kicker: lang === "en" ? "FOLLOW-UP RISK" : "追问风险",
          title: lang === "en" ? "Weak on boundary conditions" : "边界条件容易被追问",
          value: lang === "en" ? "When asked about architecture choices, follow-ups often expose incomplete comparison between alternatives and rollback plans." : "被追问架构选择时，容易暴露方案对比和回滚计划不完整。",
          confidence: lang === "en" ? "HIGH" : "高置信",
          confidenceTone: "danger",
          source: lang === "en" ? "from report question review" : "来自报告题目回顾",
        },
      ],
    },
  ],
  evidence: [
    {
      title: lang === "en" ? "Resume v3 parsed" : "简历 v3 解析",
      time: "4/18",
      body: lang === "en" ? "Added checkout RSC migration, DS rollout, Lumen platform experience." : "补充结账 RSC 迁移、Design System 落地和 Lumen 平台经历。",
    },
    {
      title: lang === "en" ? "Mock report · performance optimization" : "模拟报告 · 性能优化",
      time: "4/20",
      body: lang === "en" ? "Marked as useful story, but missing quantitative baseline in first answer." : "标记为可复用故事，但首答缺少量化基线。",
    },
    {
      title: lang === "en" ? "Debrief · Star interview" : "复盘 · 星环面试",
      time: "4/22",
      body: lang === "en" ? "User corrected that the interviewer cared more about ownership than React API details." : "用户纠正：面试官更关注 ownership，而不是 React API 细节。",
    },
  ],
});

window.UserProfileScreen = UserProfileScreen;
