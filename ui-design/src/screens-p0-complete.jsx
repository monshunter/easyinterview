// P0 support screens: JD parse flow, report generation, settings/privacy, and upload/deletion states.

// ═══════════════════════════════════════════════════════════════════
// #1 JD PARSE FLOW — loading state + structured preview / confirm
// ═══════════════════════════════════════════════════════════════════
const PlanBindingPill = ({ T, icon, label, title, meta }) => (
  <div style={{ padding: "14px 16px", background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2, display: "grid", gridTemplateColumns: "32px 1fr", gap: 12, alignItems: "center" }}>
    <div style={{ width: 32, height: 32, borderRadius: 16, background: T.bgCard, border: `1px solid ${T.rule}`, color: T.accent, display: "flex", alignItems: "center", justifyContent: "center" }}>
      <Icon name={icon} size={15} />
    </div>
    <div style={{ minWidth: 0 }}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 3 }}>{label}</div>
      <div style={{ fontSize: 14, color: T.ink, fontWeight: 500, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{title}</div>
      <div style={{ fontSize: 12, color: T.ink3, marginTop: 2, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{meta}</div>
    </div>
  </div>
);

const ParseScreen = ({ T, lang, nav, requestAuth }) => {
  const [stage, setStage] = React.useState("loading"); // loading -> preview
  const [step, setStep] = React.useState(0);

  const steps = lang === "en" ? [
    "Extracting title, level, location",
    "Identifying must-have vs nice-to-have",
    "Building the mock interview context",
    "Comparing with your resume context",
  ] : [
    "抽取岗位名、职级、地点",
    "识别必需项与加分项",
    "生成模拟面试上下文",
    "对比你的简历上下文",
  ];

  React.useEffect(() => {
    if (stage !== "loading") return;
    const ticks = [600, 900, 800, 700];
    let cancel = false;
    let acc = 0;
    ticks.forEach((t, i) => {
      acc += t;
      setTimeout(() => { if (!cancel) setStep(i + 1); }, acc);
    });
    setTimeout(() => { if (!cancel) setStage("preview"); }, acc + 200);
    return () => { cancel = true; };
  }, [stage]);

  // Mock parsed data — exposed as a readonly saved interview-plan receipt.
  const parsed = {
    title: "资深前端工程师",
    company: "星环科技",
    level: "P6 / Senior",
    location: "上海 · 混合办公（周 3 天在办）",
    language: "中文",
    source: "https://star-ring.com/careers/frontend-sr",
    fetched: lang === "en" ? "just now" : "刚刚抓取",
    mustHave: [
      { t: "React 18 + TypeScript 生产经验", hit: true },
      { t: "5+ 年前端 / Web 开发", hit: true },
      { t: "性能优化与渲染底层理解", hit: "partial", note: lang === "en" ? "your case lacks quantified impact" : "你的案例缺量化结果" },
      { t: "大型组件库 / Design System 落地", hit: false },
    ],
    niceToHave: [
      { t: "可访问性（WAI-ARIA）经验", hit: true },
      { t: "Node.js / BFF 经验", hit: "partial" },
      { t: "技术写作 / 团队分享", hit: false },
    ],
    hidden: [
      lang === "en" ? "Cross-team influence — JD mentions 'collaborating with design/product/backend' 3x" : "跨团队协作 —— JD 里 3 次提到「与设计/产品/后端协作」",
      lang === "en" ? "Ownership — 'end-to-end delivery' language, expect solo-drive stories" : "Ownership 倾向 —— 出现「端到端交付」，预计会问独立推动的故事",
      lang === "en" ? "Ambiguity tolerance — startup at Series-B stage" : "对模糊性的容忍 —— B 轮阶段",
    ],
    rounds: [
      { sequence: 1, type: "hr", name: lang === "en" ? "HR screen" : "HR 初筛", durationMinutes: 20, focus: lang === "en" ? "HR will probe motivation, timing, and fit with B2B frontend work" : "HR 会围绕求职动机、B 端产品兴趣和节奏稳定性追问" },
      { sequence: 2, type: "technical", name: lang === "en" ? "Technical architecture interview" : "技术一面", durationMinutes: 45, focus: lang === "en" ? "Tech round will focus on design-system rollout, TS types, and performance evidence" : "技术一面会聚焦 Design System 推进、TS 类型设计和性能证据" },
      { sequence: 3, type: "technical", name: lang === "en" ? "System design deep dive" : "技术二面", durationMinutes: 60, focus: lang === "en" ? "Technical deep dive will probe monorepo / micro-frontend trade-offs and collaboration scale" : "技术二面会追问 Monorepo / 微前端架构取舍与大型协作案例" },
      { sequence: 4, type: "manager", name: lang === "en" ? "Hiring manager interview" : "经理面", durationMinutes: 40, focus: lang === "en" ? "Manager round will test influence, conflict handling, and goal alignment" : "经理面会关注跨团队影响力、冲突处理和目标一致性" },
    ],
  };

  // Interview launch uses the saved resume + round snapshot; this page does not edit binding.
  const resumeOptions = window.getWorkspaceResumeOptions ? window.getWorkspaceResumeOptions(lang) : [];
  const selectedResumeId = resumeOptions[0]?.id || "";
  const selectedResume = resumeOptions.find((resume) => resume.id === selectedResumeId);
  const targetJob = window.EI_DATA.targetJobs[0];
  const progress = window.eiResolvePracticeProgress(parsed.rounds, targetJob.practiceProgress);

  const roundLabel = (round) => `${round?.name || ""}${round?.durationMinutes ? ` · ${round.durationMinutes}m` : ""}`;
  const buildParseInterviewContext = () => {
    const currentRound = progress.currentRound;
    const base = {
      resumeId: selectedResumeId,
      roundId: currentRound ? `round-${currentRound.sequence}-${currentRound.type}` : "",
      roundName: roundLabel(currentRound),
    };
    return window.eiCreateInterviewContext ? window.eiCreateInterviewContext(base, base) : base;
  };
  const startInterview = () => {
    if (!selectedResume || !progress.currentRound) return;
    const context = buildParseInterviewContext();
    const startContext = {
      ...context,
      sessionId: `session-${context.planId}-${context.roundId}-new`,
    };
    const run = () => nav("practice", startContext);
    if (!requestAuth) {
      run();
      return;
    }
    requestAuth({
      type: "create_session",
      label: lang === "en" ? "Start interview now" : "立即面试",
      route: "practice",
      params: startContext,
    }, run);
  };

  if (stage === "loading") {
    return (
      <div className="ei-fadein" style={{ minHeight: "calc(100vh - 58px)", display: "flex", alignItems: "center", justifyContent: "center", padding: 48 }}>
        <div style={{ maxWidth: 520, width: "100%" }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>
            {lang === "en" ? "PARSING · step 01 of 04" : "解析中 · 第 01 / 04 步"}
          </div>
          <div className="ei-serif" style={{ fontSize: 28, color: T.ink, letterSpacing: "-0.015em", lineHeight: 1.3, marginBottom: 32 }}>
            {lang === "en" ? "Turning the JD into mock interview context…" : "正在把这份 JD 变成模拟面试上下文…"}
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
            {steps.map((s, i) => {
              const done = i < step;
              const active = i === step;
              return (
                <div key={i} style={{ display: "flex", gap: 14, alignItems: "center" }}>
                  <div style={{
                    width: 22, height: 22, borderRadius: 11, border: `1.5px solid ${done ? T.ok : active ? T.accent : T.rule}`,
                    background: done ? T.ok : "transparent",
                    display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0,
                  }}>
                    {done && <Icon name="check" size={12} color="#fff" stroke={2.5} />}
                    {active && <div style={{ width: 6, height: 6, borderRadius: 3, background: T.accent }} className="ei-pulse" />}
                  </div>
                  <div style={{ fontSize: 14, color: done ? T.ink3 : active ? T.ink : T.ink4, textDecoration: done ? "line-through" : "none" }}>{s}</div>
                  {active && <div style={{ fontFamily: "var(--ei-mono)", fontSize: 11, color: T.ink4, marginLeft: "auto" }}>
                    <span className="ei-pulse">●</span> {lang === "en" ? "working" : "处理中"}
                  </div>}
                </div>
              );
            })}
          </div>
          <div style={{ marginTop: 40, paddingTop: 20, borderTop: `1px dotted ${T.rule}`, fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", lineHeight: 1.6 }}>
            <div>model · claude-haiku-4.5 · zh-CN</div>
            <div>rubric · jd-parse-v1.3 · prompt@a8f2e1</div>
            <div>typical · 3–6s · this one · slightly richer JD</div>
          </div>
        </div>
      </div>
    );
  }

  // Preview / confirm
  const HitDot = ({ hit }) => {
    const color = hit === true ? T.ok : hit === "partial" ? T.warn : T.ink4;
    const label = hit === true ? (lang === "en" ? "hit" : "命中") : hit === "partial" ? (lang === "en" ? "partial" : "部分") : (lang === "en" ? "gap" : "缺口");
    return (
      <div style={{ display: "flex", gap: 5, alignItems: "center", padding: "2px 7px", background: hit === true ? T.okSoft : hit === "partial" ? T.warnSoft : "transparent", border: hit === false ? `1px dashed ${T.rule}` : "none", borderRadius: 2 }}>
        <div style={{ width: 5, height: 5, borderRadius: 3, background: color }} />
        <span style={{ fontSize: 10.5, color, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em", textTransform: "uppercase" }}>{label}</span>
      </div>
    );
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 1200, margin: "0 auto", padding: "32px 48px 96px" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 24 }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>
            {lang === "en" ? "STEP 2 OF 2 · REVIEW & LAUNCH" : "第 2 / 2 步 · 核对并启动"}
          </div>
          <h1 className="ei-serif" style={{ fontSize: 32, margin: 0, color: T.ink, letterSpacing: "-0.02em", lineHeight: 1.2 }}>
            {lang === "en" ? "Interview plan detail" : "面试规划详情"}
          </h1>
          <div style={{ fontSize: 14, color: T.ink3, marginTop: 8, maxWidth: 620, lineHeight: 1.5 }}>
            {lang === "en"
              ? "Review the saved JD, bound resume, and interview round snapshot before starting."
              : "开始前核对已保存的 JD、绑定简历和面试轮次快照。"}
          </div>
        </div>
        <div style={{ textAlign: "right" }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{lang === "en" ? "SOURCE" : "来源"}</div>
          <div style={{ fontSize: 12, fontFamily: "var(--ei-mono)", color: T.ink2, maxWidth: 280, wordBreak: "break-all" }}>
            {parsed.source}
          </div>
          <div style={{ fontSize: 11, color: T.ink3, marginTop: 4 }}>{parsed.fetched}</div>
        </div>
      </div>

      {/* Basic fields */}
      <Card T={T} pad={0} style={{ marginBottom: 20 }}>
        <div style={{ padding: "16px 24px", borderBottom: `1px solid ${T.rule}` }}>
          <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "BASICS" : "基础信息"}</div>
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(2, 1fr)", padding: "6px 24px" }}>
          {[
            { k: lang === "en" ? "Title" : "岗位名", v: parsed.title, f: "title" },
            { k: lang === "en" ? "Company" : "公司", v: parsed.company, f: "company" },
            { k: lang === "en" ? "Level" : "职级", v: parsed.level, f: "level" },
            { k: lang === "en" ? "Location" : "地点", v: parsed.location, f: "location" },
            { k: lang === "en" ? "Language" : "语言", v: parsed.language, f: "language" },
          ].map((r, i) => (
            <div key={r.f} style={{ display: "flex", gap: 14, padding: "12px 0", borderBottom: i < 3 ? `1px dotted ${T.rule}` : "none", alignItems: "baseline" }}>
              <div className="ei-label" style={{ color: T.ink3, minWidth: 68, fontSize: 10.5 }}>{r.k}</div>
              <div style={{ flex: 1, fontSize: 14, color: T.ink, padding: "2px 0", fontFamily: "var(--ei-sans)" }}>{r.v}</div>
            </div>
          ))}
        </div>
      </Card>

      {/* Requirements */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20, marginBottom: 20 }}>
        <RequirementBlock T={T} lang={lang} title={lang === "en" ? "MUST HAVE" : "必需项"} items={parsed.mustHave} HitDot={HitDot} />
        <RequirementBlock T={T} lang={lang} title={lang === "en" ? "NICE TO HAVE" : "加分项"} items={parsed.niceToHave} HitDot={HitDot} />
      </div>

      {/* Hidden signals */}
      <Card T={T} style={{ marginBottom: 20, borderColor: T.accent }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14 }}>
          <div>
            <div className="ei-label" style={{ color: T.accent, marginBottom: 4 }}>{lang === "en" ? "HIDDEN SIGNALS · inference" : "隐性关注点 · 推断"}</div>
            <div style={{ fontSize: 13, color: T.ink3 }}>{lang === "en" ? "Not stated, but JD language suggests —" : "JD 没明说，但字里行间暗示的——"}</div>
          </div>
          <Tag T={T} tone="accent">
            <Icon name="sparkle" size={10} style={{ marginRight: 4 }} />
            {lang === "en" ? "confidence · medium" : "置信度 · 中"}
          </Tag>
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {parsed.hidden.map((h, i) => (
            <div key={i} style={{ display: "flex", gap: 10, alignItems: "flex-start", padding: "8px 12px", background: T.bgSoft, borderRadius: 2 }}>
              <Icon name="sparkle" size={12} color={T.accent} style={{ marginTop: 3, flexShrink: 0 }} />
              <div style={{ fontSize: 13.5, color: T.ink, lineHeight: 1.5, flex: 1 }}>{h}</div>
            </div>
          ))}
        </div>
      </Card>

      {/* Round assumptions — readonly saved context */}
      <Card T={T} style={{ marginBottom: 20 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14 }}>
          <div className="ei-label" style={{ color: T.ink3 }}>{lang === "en" ? "ROUND ASSUMPTIONS" : "轮次假设"}</div>
          <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
            {lang === "en" ? "saved with this plan" : "创建后只读"}
          </div>
        </div>
        <div style={{ display: "grid", gridTemplateColumns: `repeat(${Math.min(parsed.rounds.length || 1, 4)}, 1fr)`, gap: 10 }}>
          {parsed.rounds.map((r, i) => (
              <div key={i} style={{ textAlign: "left", fontFamily: "var(--ei-sans)", padding: "12px 14px", background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2, position: "relative" }}>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 5 }}>
                  <span style={{ fontFamily: "var(--ei-mono)", fontSize: 10.5, color: T.ink4, letterSpacing: "0.06em" }}>R{i + 1}</span>
                </div>
                <div style={{ fontSize: 13, color: T.ink, fontWeight: 500, marginBottom: 4 }}>{roundLabel(r)}</div>
                <div style={{ fontSize: 11.5, color: T.ink3, lineHeight: 1.45 }}>{r.focus}</div>
              </div>
          ))}
        </div>
      </Card>

      {/* Interview launch — readonly binding + start, no second confirm page */}
      <Card T={T} style={{ marginBottom: 28 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{lang === "en" ? "INTERVIEW LAUNCH" : "面试启动"}</div>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
          <PlanBindingPill T={T} icon="briefcase" label={lang === "en" ? "Target job / JD" : "目标岗位 / JD"} title={parsed.title} meta={`${parsed.company} · ${parsed.level}`} />
          {selectedResume ? (
            <PlanBindingPill T={T} icon="resume" label={lang === "en" ? "Bound resume" : "绑定简历"} title={selectedResume.name} meta={selectedResume.meta} />
          ) : (
            <PlanBindingPill T={T} icon="resume" label={lang === "en" ? "Bound resume" : "绑定简历"} title={lang === "en" ? "Missing bound resume" : "缺少绑定简历"} meta={lang === "en" ? "Create a new plan from Home" : "请从首页重新创建规划"} />
          )}
        </div>
        <div style={{ fontSize: 12, color: T.ink3, marginTop: 12, display: "flex", gap: 6, alignItems: "center" }}>
          <Icon name="info" size={12} /> {lang === "en" ? "The interview currently runs as text conversation; phone mode is temporarily unavailable." : "当前面试仅使用文字对话，电话模式暂未开放。"}
        </div>
      </Card>

      {/* Footer actions */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "16px 0", borderTop: `1px solid ${T.rule}`, gap: 16, flexWrap: "wrap" }}>
        <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)", lineHeight: 1.6, maxWidth: 420 }}>
          {lang === "en" ? "This saved plan uses the JD and bound resume shown above." : "面试规划会使用这份 JD 和已绑定简历。"}
        </div>
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
          <Btn T={T} variant="accent" icon="play" disabled={!selectedResume || !progress.currentRound} onClick={startInterview}>{lang === "en" ? "Start interview now" : "立即面试"}</Btn>
        </div>
      </div>
    </div>
  );
};

const RequirementBlock = ({ T, title, items, HitDot }) => (
  <Card T={T} pad={0}>
    <div style={{ padding: "14px 20px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
      <div className="ei-label" style={{ color: T.ink3 }}>{title}</div>
      <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{items.length}</div>
    </div>
    <div>
      {items.map((item, i) => (
        <div key={i} style={{ padding: "12px 20px", borderBottom: i < items.length - 1 ? `1px dotted ${T.rule}` : "none", display: "flex", gap: 12, alignItems: "flex-start" }}>
          <div style={{ marginTop: 2 }}>
            <HitDot hit={item.hit} />
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 13.5, color: T.ink, lineHeight: 1.45 }}>{item.t}</div>
            {item.note && <div style={{ fontSize: 11.5, color: T.warn, marginTop: 4, fontStyle: "italic" }}>{item.note}</div>}
          </div>
        </div>
      ))}
    </div>
  </Card>
);

// ═══════════════════════════════════════════════════════════════════
// #3 REPORT GENERATING — async intermediate screen
// ═══════════════════════════════════════════════════════════════════
const ReportGeneratingScreen = ({ T, lang, nav, params = {} }) => {
  const report = window.EI_DATA.reportGeneration;
  const status = report.status;
  const reportMatches = !!params.reportId && params.reportId === report.id;
  const isWaiting = reportMatches && (status === "queued" || status === "generating");
  const isRecoverable = reportMatches && (status === "timeout" || status === "network_error");
  const isOversize = reportMatches && status === "failed" && report.errorCode === "REPORT_CONTEXT_TOO_LARGE";
  const isTerminal = !reportMatches || status === "failed" || (!isWaiting && !isRecoverable && status !== "ready");
  const continueCheck = () => window.location.reload();

  React.useEffect(() => {
    if (reportMatches && status === "ready") nav("report", { reportId: report.id });
  }, [reportMatches, status, report.id]);

  const terminalCopy = !reportMatches
    ? {
      title: lang === "en" ? "Report ID missing" : "缺少报告 ID",
      description: lang === "en" ? "Open this page from a completed mock interview with a valid report link." : "请从已完成的模拟面试使用有效报告链接打开此页面。",
    }
    : isOversize
      ? {
        title: lang === "en" ? "The source material and conversation were too long" : "本次材料与对话过长",
        description: lang === "en" ? "Return to your interview plan, shorten the input, and start a new session." : "请返回面试规划，缩短输入后开启一场新会话。",
      }
      : report.errorCode === "REPORT_NOT_FOUND"
        ? {
          title: lang === "en" ? "Report not found" : "找不到这份报告",
          description: lang === "en" ? "Make sure this link belongs to a completed mock interview on your account, then return to the workspace." : "请确认链接来自当前账户已完成的模拟面试，然后返回面试规划。",
        }
        : report.errorCode === "AI_OUTPUT_INVALID"
          ? {
            title: lang === "en" ? "Report content failed validation" : "报告内容未通过校验",
            description: lang === "en" ? "Unreliable content was not shown. Return to the workspace and start a new session." : "系统没有展示不可靠内容。请返回面试规划后开启一场新会话。",
          }
          : {
            title: lang === "en" ? "This report cannot continue" : "这份报告无法继续生成",
            description: lang === "en" ? "Return to the workspace and start a new session when ready." : "请返回面试规划，并在准备好后开启一场新会话。",
          };
  const recoverableCopy = status === "network_error"
    ? {
      title: lang === "en" ? "Report status could not connect" : "报告状态连接异常",
      description: lang === "en" ? "This connection did not finish. You can check the same report again or return to the workspace." : "本次连接没有完成。你可以继续检查同一份报告，或返回面试规划。",
    }
    : {
      title: lang === "en" ? "Report status could not be confirmed" : "暂时无法确认报告状态",
      description: lang === "en" ? "This check timed out. You can check the same report again or return to the workspace." : "本次检查超时。你可以继续检查同一份报告，或返回面试规划。",
    };
  const eyebrow = isWaiting
    ? (status === "queued" ? (lang === "en" ? "REPORT QUEUED" : "报告已排队") : (lang === "en" ? "GENERATING REPORT" : "报告生成中"))
    : isRecoverable
      ? (lang === "en" ? "CHECK PAUSED" : "检查已暂停")
      : (lang === "en" ? "REPORT UNAVAILABLE" : "报告不可用");
  const title = isWaiting
    ? (lang === "en" ? "Building an evidence-based report." : "正在生成证据化报告。")
    : isRecoverable
      ? recoverableCopy.title
      : terminalCopy.title;
  const description = isWaiting
    ? (lang === "en" ? "We are checking the completed conversation and forming grounded recommendations. This page will keep checking the same report." : "系统正在核对已完成的对话并形成有依据的建议；本页会继续检查同一份报告。")
    : isRecoverable
      ? recoverableCopy.description
      : terminalCopy.description;

  return (
    <div className="ei-fadein" data-testid="generating-screen" data-report-status={isWaiting ? status : undefined} aria-live="polite" style={{ minHeight: "100vh", background: T.bg, display: "flex", alignItems: "center", justifyContent: "center", padding: "48px clamp(16px, 6vw, 48px)" }}>
      <div style={{ maxWidth: 780, width: "100%" }}>
        <div className="ei-label" data-testid="generating-header-eyebrow" style={{ color: T.ink3, marginBottom: 12, letterSpacing: "0.1em" }}>
          {eyebrow}
        </div>
        <h1 className="ei-serif" data-testid="generating-header-title" style={{ fontSize: 34, margin: 0, color: T.ink, letterSpacing: "-0.02em", lineHeight: 1.2, marginBottom: 10 }}>
          {title}
        </h1>
        <div data-testid="generating-header-subtitle" style={{ fontSize: 14, color: T.ink3, lineHeight: 1.65, maxWidth: 600 }}>
          {description}
        </div>

        {(isRecoverable || isTerminal) && (
          <div style={{ marginTop: 28, paddingTop: 16, borderTop: `1px solid ${T.rule}`, display: "flex", gap: 10, flexWrap: "wrap" }}>
            {isRecoverable && <Btn T={T} variant="accent" size="sm" onClick={continueCheck}>{lang === "en" ? "Check again" : "继续检查"}</Btn>}
            <Btn T={T} variant="secondary" size="sm" onClick={() => nav("workspace")}>{lang === "en" ? "Back to workspace" : "返回面试规划"}</Btn>
          </div>
        )}
      </div>
    </div>
  );
};

// ═══════════════════════════════════════════════════════════════════
// #8 SETTINGS / PRIVACY / DATA EXPORT & DELETE
// ═══════════════════════════════════════════════════════════════════
const SettingsScreen = ({ T, lang, fontPreset, setFontPreset }) => {
  const [tab, setTab] = React.useState("profile");

  const tabs = lang === "en"
    ? [{ k: "profile", t: "Profile" }, { k: "privacy", t: "Privacy & data" }]
    : [{ k: "profile", t: "个人资料" }, { k: "privacy", t: "隐私与数据" }];

  return (
    <div className="ei-fadein" style={{ maxWidth: 1040, margin: "0 auto", padding: "40px 48px 96px" }}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "SETTINGS" : "设置"}</div>
      <h1 className="ei-serif" style={{ fontSize: 36, margin: 0, color: T.ink, letterSpacing: "-0.02em", marginBottom: 32 }}>
        {lang === "en" ? "Your account & what we keep." : "账号与我们在你这里存的东西。"}
      </h1>

      <div style={{ display: "grid", gridTemplateColumns: "200px 1fr", gap: 40 }}>
        {/* Tab rail */}
        <div style={{ display: "flex", flexDirection: "column", gap: 0 }}>
          {tabs.map((t) => (
            <button key={t.k} onClick={() => setTab(t.k)} style={{
              background: "transparent", border: "none", textAlign: "left",
              padding: "10px 12px", cursor: "pointer", borderRadius: 2, fontFamily: "var(--ei-sans)",
              color: tab === t.k ? T.ink : T.ink3,
              fontWeight: tab === t.k ? 500 : 400, fontSize: 14,
              borderLeft: `2px solid ${tab === t.k ? T.accent : "transparent"}`,
            }}>
              {t.t}
            </button>
          ))}
        </div>

        {/* Content */}
        <div>
          {tab === "privacy" && <SettingsPrivacy T={T} lang={lang} />}
          {tab === "profile" && <SettingsProfile T={T} lang={lang} fontPreset={fontPreset} setFontPreset={setFontPreset} />}
        </div>
      </div>
    </div>
  );
};

const SettingsPrivacy = ({ T, lang }) => {
  const [toggles, setToggles] = React.useState({ transcript: true, resume: true, anon: false, emails: true });
  const Toggle = ({ k }) => (
    <button onClick={() => setToggles({ ...toggles, [k]: !toggles[k] })} style={{
      width: 34, height: 18, borderRadius: 9, border: "none", cursor: "pointer",
      background: toggles[k] ? T.accent : T.rule, position: "relative", flexShrink: 0,
    }}>
      <div style={{ position: "absolute", top: 2, left: toggles[k] ? 18 : 2, width: 14, height: 14, borderRadius: 7, background: "#fff", transition: "left .15s" }} />
    </button>
  );

  return (
    <div>
      {/* Data retention */}
      <section style={{ marginBottom: 40 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "DATA RETENTION" : "数据留存"}</div>
        <div style={{ fontSize: 13.5, color: T.ink3, marginBottom: 18, lineHeight: 1.55 }}>
          {lang === "en" ? "Everything is keep-by-default-you-can-delete — not the other way around. These toggles tell the system what it's allowed to touch." : "默认保留、你随时能删——不是反过来。下面的开关决定系统能碰哪些数据。"}
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 0, background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
          {[
            { k: "transcript", t: lang === "en" ? "Keep conversation transcripts long-term" : "长期保留对话记录", d: lang === "en" ? "Used to review past conversations and compare readiness over time." : "用来回看过往对话，并比较准备度变化。" },
            { k: "resume", t: lang === "en" ? "Keep uploaded resumes" : "保留上传的简历", d: lang === "en" ? "Used only to suggest stories during practice." : "只在练习时用来推荐故事。" },
            { k: "anon", t: lang === "en" ? "Contribute anonymized samples to improve rubrics" : "匿名贡献样本用于改进评分", d: lang === "en" ? "Off by default. Names, companies, numbers are stripped." : "默认关闭。姓名、公司、数字会被去除。" },
          ].map((r, i, arr) => (
            <div key={r.k} style={{ padding: "16px 20px", display: "flex", gap: 16, alignItems: "flex-start", borderBottom: i < arr.length - 1 ? `1px dotted ${T.rule}` : "none" }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 14, color: T.ink, marginBottom: 3 }}>{r.t}</div>
                <div style={{ fontSize: 12, color: T.ink3, lineHeight: 1.5 }}>{r.d}</div>
              </div>
              <Toggle k={r.k} />
            </div>
          ))}
        </div>
      </section>

      {/* Data overview */}
      <section style={{ marginBottom: 40 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "WHAT WE HAVE ON YOU" : "我们这边存了你什么"}</div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 10, marginBottom: 14 }}>
          {[
            { k: "4", l: lang === "en" ? "target jobs" : "目标岗位" },
            { k: "18", l: lang === "en" ? "practice sessions" : "练习会话" },
            { k: "2", l: lang === "en" ? "resumes" : "份简历" },
          ].map((c) => (
            <div key={c.l} style={{ padding: "16px 18px", background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
              <div className="ei-serif" style={{ fontSize: 24, color: T.ink, letterSpacing: "-0.015em" }}>{c.k}</div>
              <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em", marginTop: 4 }}>{c.l}</div>
            </div>
          ))}
        </div>
      </section>

      {/* Export */}
      <section style={{ marginBottom: 40 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "EXPORT" : "导出"}</div>
        <div style={{ padding: "18px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, display: "flex", gap: 16, alignItems: "center" }}>
          <Icon name="download" size={20} color={T.ink2} />
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 14, color: T.ink, marginBottom: 3 }}>
              {lang === "en" ? "Download everything — JSON + PDF reports" : "下载全部 —— JSON + PDF 报告"}
            </div>
            <div style={{ fontSize: 12, color: T.ink3, lineHeight: 1.5 }}>
              {lang === "en" ? "Practice conversations, reports, resumes, and saved JDs. Link is emailed when ready (<5min)." : "练习对话、报告、简历和已保存 JD。准备好发到你邮箱（<5 分钟）。"}
            </div>
          </div>
          <Btn T={T} variant="secondary" size="sm" onClick={() => window.eiToast && window.eiToast(lang === "en" ? "Export requested · link emailed to liuzhe@example.com when ready" : "已申请导出 · 准备好后会发到 liuzhe@example.com", { tone: "ok", duration: 3000 })}>{lang === "en" ? "Request export" : "申请导出"}</Btn>
        </div>
      </section>

      {/* Delete — danger zone */}
      <section>
        <div className="ei-label" style={{ color: T.danger, marginBottom: 14 }}>{lang === "en" ? "DANGER ZONE" : "高危操作"}</div>
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          {[
            { t: lang === "en" ? "Delete a single session" : "删除某一次会话", d: lang === "en" ? "Pick a session — its conversation and report are removed together." : "挑一次会话，对话和报告会一起删除。", b: lang === "en" ? "Pick" : "选择" },
            { t: lang === "en" ? "Delete all practice data" : "删除所有练习数据", d: lang === "en" ? "Conversations, reports, and readiness signals are removed. Saved JDs and resumes stay." : "对话、报告和准备度信号全部删掉。已保存 JD 和简历保留。", b: lang === "en" ? "Delete…" : "删除…" },
            { t: lang === "en" ? "Delete my account" : "注销账号", d: lang === "en" ? "Permanent. All data is purged within 30 days per GDPR. Backups rotated within 90." : "永久。30 天内按 GDPR 清理所有数据。备份 90 天内轮换清除。", b: lang === "en" ? "Delete account…" : "注销账号…", danger: true },
          ].map((r) => (
            <div key={r.t} style={{ padding: "16px 20px", background: r.danger ? T.dangerSoft : T.bgCard, border: `1px solid ${r.danger ? T.danger : T.rule}`, borderRadius: 2, display: "flex", gap: 16, alignItems: "flex-start" }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 14, color: T.ink, marginBottom: 3, fontWeight: 500 }}>{r.t}</div>
                <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.55 }}>{r.d}</div>
              </div>
              <button style={{
                padding: "6px 14px", fontSize: 12.5, cursor: "pointer", fontFamily: "var(--ei-sans)",
                background: r.danger ? T.danger : "transparent",
                color: r.danger ? "#fff" : T.danger,
                border: `1px solid ${T.danger}`, borderRadius: 2, whiteSpace: "nowrap",
              }}>{r.b}</button>
            </div>
          ))}
        </div>
        <div style={{ marginTop: 16, fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", lineHeight: 1.6 }}>
          {lang === "en" ? "deletion requests typically complete < 5 min · legal-hold exceptions documented in privacy policy" : "删除请求通常 < 5 分钟完成 · 法律保留例外情况详见隐私政策"}
        </div>
      </section>
    </div>
  );
};

const SettingsProfile = ({ T, lang, fontPreset, setFontPreset }) => {
  const identityRows = [
    { k: lang === "en" ? "Display name" : "显示姓名", v: "刘哲" },
    { k: lang === "en" ? "Login email" : "登录邮箱", v: "liuzhe@example.com" },
    { k: lang === "en" ? "Mobile" : "手机号", v: lang === "en" ? "Not connected" : "未绑定" },
    { k: lang === "en" ? "Interface language" : "界面语言", v: lang === "en" ? "Chinese · English" : "中文 · English" },
    { k: lang === "en" ? "Time zone" : "时区", v: "Asia/Shanghai" },
  ];
  const securityRows = [
    { k: lang === "en" ? "Login method" : "登录方式", v: lang === "en" ? "Email sign-in code · no password" : "邮箱验证码 · 无密码" },
  ];

  const Row = ({ r, last }) => (
    <div style={{ display: "flex", gap: 16, padding: "14px 0", borderBottom: last ? "none" : `1px dotted ${T.rule}`, alignItems: "baseline" }}>
      <div className="ei-label" style={{ color: T.ink3, minWidth: 160, fontSize: 11 }}>{r.k}</div>
      <div style={{ fontSize: 14, color: T.ink, flex: 1 }}>{r.v}</div>
      <button style={{ background: "transparent", border: "none", color: T.accent, fontSize: 12, cursor: "pointer" }}>{lang === "en" ? "Edit" : "编辑"}</button>
    </div>
  );

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 26 }}>
      <section>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "BASIC ACCOUNT INFO" : "账号基础信息"}</div>
        <div style={{ padding: "18px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, display: "flex", gap: 18, alignItems: "center", marginBottom: 14 }}>
          <div style={{ width: 48, height: 48, borderRadius: 24, background: T.ink2, color: T.bg, display: "flex", alignItems: "center", justifyContent: "center", fontFamily: "var(--ei-mono)", fontWeight: 600 }}>LZ</div>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 15, color: T.ink, fontWeight: 600 }}>{lang === "en" ? "Liu Zhe" : "刘哲"}</div>
            <div style={{ fontSize: 12.5, color: T.ink3, marginTop: 4 }}>{lang === "en" ? "This page only stores account identity and product preferences." : "这里仅保存账号身份和产品基础偏好，不保存岗位、年限、目标方向等一对多信息。"}</div>
          </div>
          <Btn T={T} variant="secondary" size="sm" icon="upload" onClick={() => window.eiToast && window.eiToast(lang === "en" ? "Pick an image (≤2 MB) · upload mocked in prototype" : "选择图片（≤2 MB）· 原型仅模拟上传", { tone: "neutral" })}>{lang === "en" ? "Change avatar" : "更换头像"}</Btn>
        </div>
        <div>
          {identityRows.map((r, i) => <Row key={r.k} r={r} last={i === identityRows.length - 1} />)}
        </div>
      </section>

      <section>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "SIGN-IN & SECURITY" : "登录与安全"}</div>
        <div>
          {securityRows.map((r, i) => <Row key={r.k} r={r} last={i === securityRows.length - 1} />)}
        </div>
      </section>

      <section>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "INTERFACE PREFERENCES" : "界面偏好"}</div>
        <div style={{ fontSize: 13, color: T.ink3, marginBottom: 14, lineHeight: 1.55 }}>
          {lang === "en"
            ? "Pick a typography pack — switches both the serif (display) and the sans (body) together so the rhythm stays coherent. Mono never changes."
            : "选一套字体组合 —— serif（标题）和 sans（正文）会成对切换，保证排版节奏不乱。等宽字体不会变。"}
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 12 }}>
          {(window.EI_FONT_PRESETS || []).map((p) => {
            const selected = (fontPreset || "editorial") === p.key;
            return (
              <button
                key={p.key}
                onClick={() => setFontPreset && setFontPreset(p.key)}
                style={{
                  textAlign: "left", padding: "16px 18px",
                  background: selected ? T.accentSoft : T.bgCard,
                  border: `1.5px solid ${selected ? T.accent : T.rule}`,
                  borderRadius: 3, cursor: "pointer",
                  display: "flex", flexDirection: "column", gap: 10,
                  fontFamily: "var(--ei-sans)",
                }}
                onMouseEnter={(e) => { if (!selected) e.currentTarget.style.borderColor = T.ink3; }}
                onMouseLeave={(e) => { if (!selected) e.currentTarget.style.borderColor = T.rule; }}
              >
                <div style={{ display: "flex", alignItems: "baseline", justifyContent: "space-between", gap: 8 }}>
                  <div style={{ fontSize: 13.5, color: T.ink, fontWeight: 600 }}>
                    {lang === "en" ? p.labelEn : p.labelZh}
                  </div>
                  {selected && <Icon name="check" size={13} style={{ color: T.accent }} />}
                </div>
                <div style={{ fontFamily: `"${p.serif}", Georgia, serif`, fontSize: 26, color: T.ink, letterSpacing: "-0.015em", lineHeight: 1.1 }}>
                  Aa 你好
                </div>
                <div style={{ fontFamily: `"${p.sans}", -apple-system, sans-serif`, fontSize: 12.5, color: T.ink2, lineHeight: 1.45 }}>
                  The quick brown fox · 这是正文样式示例
                </div>
                <div style={{ fontSize: 11, color: T.ink3, lineHeight: 1.5, marginTop: 2 }}>
                  {lang === "en" ? p.descEn : p.descZh}
                </div>
                <div style={{ fontSize: 10.5, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.04em", marginTop: 2 }}>
                  {p.serif} · {p.sans}
                </div>
              </button>
            );
          })}
        </div>
      </section>

      <section>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "PRODUCT INFO" : "产品信息"}</div>
        <div style={{ padding: "16px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
          {[
            { k: lang === "en" ? "Product" : "产品名称", v: "EasyInterview" },
            { k: lang === "en" ? "Version" : "版本", v: "v1.0" },
          ].map((r, i, arr) => (
            <div key={r.k} style={{ display: "flex", gap: 16, padding: "10px 0", borderBottom: i < arr.length - 1 ? `1px dotted ${T.rule}` : "none", alignItems: "baseline" }}>
              <div className="ei-label" style={{ color: T.ink3, minWidth: 160, fontSize: 11 }}>{r.k}</div>
              <div style={{ fontSize: 14, color: T.ink, flex: 1, fontFamily: r.v === "v1.0" ? "var(--ei-mono)" : "var(--ei-sans)" }}>{r.v}</div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
};

window.ParseScreen = ParseScreen;
window.ReportGeneratingScreen = ReportGeneratingScreen;
window.SettingsScreen = SettingsScreen;
