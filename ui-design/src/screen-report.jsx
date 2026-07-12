// Screen 4: conversation-level evidence report.
const ReportScreen = ({ T, lang, nav, params = {} }) => {
  const D = window.EI_DATA;
  const report = D.report;
  const context = window.eiCreateInterviewContext ? window.eiCreateInterviewContext(params) : params;
  const job = D.targetJobs.find((item) => item.id === context.targetJobId) || D.targetJobs[0];
  if (params.reportStatus === "failed") return <ReportFailureState T={T} lang={lang} nav={nav} context={context} />;
  if (!params.sessionId) return <ReportMissingSessionState T={T} lang={lang} nav={nav} context={context} />;
  return <ReportDashboard T={T} lang={lang} nav={nav} report={report} job={job} params={context} />;
};

const ReportMissingSessionState = ({ T, lang, nav, context }) => (
  <div className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px 48px" }}>
    <Card T={T}>
      <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "REPORT NEEDS A SESSION" : "报告缺少会话"}</div>
      <div className="ei-serif" style={{ fontSize: 28, color: T.ink, marginBottom: 16 }}>{lang === "en" ? "Open a report from a completed interview." : "请从已完成的模拟面试打开报告。"}</div>
      <Btn T={T} variant="accent" onClick={() => nav("workspace", context)}>{lang === "en" ? "Back to records" : "返回面试记录"}</Btn>
    </Card>
  </div>
);

const ReportFailureState = ({ T, lang, nav, context }) => (
  <div className="ei-fadein" style={{ maxWidth: 820, margin: "0 auto", padding: "72px 48px" }}>
    <Card T={T}>
      <div className="ei-label" style={{ color: T.danger, marginBottom: 10 }}>{lang === "en" ? "REPORT FAILED" : "报告生成失败"}</div>
      <div className="ei-serif" style={{ fontSize: 28, color: T.ink, marginBottom: 16 }}>{lang === "en" ? "Evidence could not be generated for this conversation." : "暂时无法为这场对话生成证据报告。"}</div>
      <div style={{ display: "flex", gap: 10 }}><Btn T={T} variant="accent" onClick={() => nav("generating", context)}>{lang === "en" ? "Retry" : "重新生成"}</Btn><Btn T={T} variant="secondary" onClick={() => nav("workspace", context)}>{lang === "en" ? "Back" : "返回记录"}</Btn></div>
    </Card>
  </div>
);

const ReportDashboard = ({ T, lang, nav, report, job, params }) => {
  const { nextRound } = window.eiResolveInterviewRoundContext(window.EI_DATA.jdSample.interviewRounds, params.roundId);
  const dimensions = report.dimensions || [];
  const highlights = report.highlights || [];
  const issues = report.issues || [];
  const actions = report.nextPractice || [];
  return (
    <main className="ei-fadein" data-testid="report-dashboard" style={{ maxWidth: 1120, margin: "0 auto", padding: "32px clamp(16px, 5vw, 48px) 96px" }}>
      <button onClick={() => nav("workspace", params)} style={{ border: 0, background: "transparent", color: T.ink3, cursor: "pointer", marginBottom: 20 }}>← {lang === "en" ? "Records" : "面试记录"}</button>
      <header style={{ display: "flex", justifyContent: "space-between", gap: 24, alignItems: "flex-end", flexWrap: "wrap", marginBottom: 24 }}>
        <div><div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{lang === "en" ? "CONVERSATION REPORT" : "会话报告"}</div><h1 className="ei-serif" style={{ margin: 0, fontSize: 38, color: T.ink }}>{job.company} · {job.title}</h1><p style={{ color: T.ink2, lineHeight: 1.7 }}>{lang === "en" ? "Evidence and capability signals from the complete interview conversation." : "基于整场模拟面试对话提取证据与能力信号。"}</p></div>
        <div style={{ display: "flex", gap: 10 }}><Btn T={T} variant="accent" onClick={() => nav("practice", { ...params, practiceGoal: "retry_current_round" })}>{lang === "en" ? "Practice again" : "复练当前轮"}</Btn><Btn T={T} variant="secondary" disabled={!nextRound} onClick={() => nextRound && nav("practice", { ...params, practiceGoal: "next_round", roundId: nextRound.id, roundName: nextRound.name })}>{lang === "en" ? "Next round" : "进入下一轮"}</Btn></div>
      </header>
      <section style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))", gap: 14, marginBottom: 22 }}>
        <ReportMetric T={T} label={lang === "en" ? "READINESS" : "准备度"} value={report.readiness || (lang === "en" ? "Needs practice" : "建议再练")} />
        <ReportMetric T={T} label={lang === "en" ? "CAPABILITY DIMENSIONS" : "能力维度"} value={`${dimensions.length || 4}`} />
        <ReportMetric T={T} label={lang === "en" ? "EVIDENCE SIGNALS" : "证据信号"} value={`${highlights.length + issues.length || 5}`} />
      </section>
      <section style={{ display: "grid", gridTemplateColumns: "minmax(0, 1.2fr) minmax(280px, .8fr)", gap: 18 }}>
        <Card T={T}><div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "CAPABILITY ASSESSMENT" : "能力维度评估"}</div>{dimensions.map((item) => <div key={item.name} style={{ display: "flex", justifyContent: "space-between", gap: 16, padding: "13px 0", borderBottom: `1px dotted ${T.rule}` }}><span style={{ color: T.ink }}>{item.name}</span><span style={{ color: item.state === "待加强" ? T.warn : T.ok }}>{item.state}</span></div>)}</Card>
        <div style={{ display: "grid", gap: 18 }}><EvidenceCard T={T} title={lang === "en" ? "STRENGTH EVIDENCE" : "优势证据"} items={highlights} color={T.ok} /><EvidenceCard T={T} title={lang === "en" ? "RISKS" : "待加强证据"} items={issues} color={T.warn} /><EvidenceCard T={T} title={lang === "en" ? "NEXT ACTIONS" : "下一步行动"} items={actions} color={T.accent} /></div>
      </section>
    </main>
  );
};

const ReportMetric = ({ T, label, value }) => <div style={{ padding: 20, border: `1px solid ${T.rule}`, background: T.bgCard }}><div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{label}</div><div className="ei-serif" style={{ color: T.ink, fontSize: 24 }}>{value}</div></div>;
const EvidenceCard = ({ T, title, items, color }) => <Card T={T}><div className="ei-label" style={{ color, marginBottom: 12 }}>{title}</div>{items.length ? items.map((item, index) => <div key={item.id || item.title || index} style={{ color: T.ink2, fontSize: 13, lineHeight: 1.65, marginTop: index ? 10 : 0 }}>{item.evidence || item.title || item.label || String(item)}</div>) : <div style={{ color: T.ink3 }}>—</div>}</Card>;

window.ReportScreen = ReportScreen;
