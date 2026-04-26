// Screens 5-8: mistakes, debrief, resume, growth
const MistakesScreen = ({ T, lang, nav }) => {
  const D = window.EI_DATA;
  const [filter, setFilter] = React.useState("all");
  const items = D.mistakes.filter((m) => filter === "all" ? true : filter === "open" ? m.status !== "已攻克" : m.status === "已攻克");

  const L = lang === "en" ? {
    title: "Mistake book",
    sub: "Every question you've struggled on, sorted by how badly it's still hurting you.",
    all: "All", open: "Open", solved: "Solved",
    ability: "Ability", lastTry: "Last attempt", attempts: "Tries",
    retry: "Retry", addNote: "Add note",
  } : {
    title: "错题本",
    sub: "所有让你卡壳过的题，按对你当前的杀伤力排序。",
    all: "全部", open: "未攻克", solved: "已攻克",
    ability: "能力点", lastTry: "最近复练", attempts: "尝试次数",
    retry: "复练", addNote: "加批注",
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 1100, margin: "0 auto", padding: "40px 48px 96px" }}>
      <div style={{ marginBottom: 28, display: "flex", alignItems: "flex-end", justifyContent: "space-between", gap: 24 }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "ARCHIVE · PERSONAL" : "档案 · 私人"}</div>
          <h1 className="ei-serif" style={{ fontSize: 38, color: T.ink, margin: 0, letterSpacing: "-0.02em" }}>{L.title}</h1>
          <div style={{ fontSize: 14.5, color: T.ink2, marginTop: 8, maxWidth: 580 }}>{L.sub}</div>
        </div>
        <Btn T={T} variant="primary" icon="target" onClick={() => nav("drill")}>
          {lang === "en" ? "Build a targeted drill →" : "针对性复练 →"}
        </Btn>
      </div>

      <div style={{ display: "flex", gap: 8, marginBottom: 20 }}>
        {[["all", L.all, D.mistakes.length], ["open", L.open, D.mistakes.filter(m => m.status !== "已攻克").length], ["solved", L.solved, D.mistakes.filter(m => m.status === "已攻克").length]].map(([k, label, n]) => (
          <button key={k} onClick={() => setFilter(k)} style={{
            padding: "8px 14px", background: filter === k ? T.ink : "transparent",
            color: filter === k ? T.bg : T.ink2, border: `1px solid ${filter === k ? T.ink : T.rule}`,
            borderRadius: 2, fontSize: 13, cursor: "pointer", display: "flex", gap: 6,
          }}>
            {label} <span style={{ opacity: 0.6, fontFamily: "var(--ei-mono)" }}>{n}</span>
          </button>
        ))}
      </div>

      <div style={{ border: `1px solid ${T.rule}`, borderRadius: 2, background: T.bgCard }}>
        <div style={{ display: "grid", gridTemplateColumns: "40px 2fr 1.5fr 100px 80px 90px 120px", padding: "12px 18px", borderBottom: `1px solid ${T.rule}`, background: T.bgSoft }}>
          {["#", L.ability, lang === "en" ? "Question" : "题面", L.lastTry, L.attempts, lang === "en" ? "Status" : "状态", ""].map((h, i) => (
            <div key={i} className="ei-label" style={{ color: T.ink3 }}>{h}</div>
          ))}
        </div>
        {items.map((m, i) => {
          const statusTone = m.status === "已攻克" ? "ok" : m.status === "改善中" ? "amber" : "danger";
          return (
            <div key={m.id} style={{ display: "grid", gridTemplateColumns: "40px 2fr 1.5fr 100px 80px 90px 120px", padding: "16px 18px", borderBottom: i < items.length - 1 ? `1px dotted ${T.rule}` : "none", alignItems: "center", gap: 10 }}>
              <div className="ei-mono" style={{ color: T.ink3, fontSize: 12 }}>{String(i + 1).padStart(2, "0")}</div>
              <div>
                <div style={{ fontSize: 14, color: T.ink, fontWeight: 500 }}>{m.ability}</div>
                <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 2, fontFamily: "var(--ei-mono)" }}>{m.jd} · {lang === "en" ? "prio" : "优先级"} {m.priority}</div>
              </div>
              <div style={{ fontSize: 13, color: T.ink2, lineHeight: 1.45 }}>{m.question}</div>
              <div style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{m.lastAttempt}</div>
              <div style={{ fontSize: 12, color: T.ink2, fontFamily: "var(--ei-mono)" }}>{m.attempts}×</div>
              <Tag tone={statusTone} T={T}>{m.status}</Tag>
              <Btn variant="secondary" size="sm" T={T} icon="replay" onClick={() => nav("practice", { jobId: "tj-1", mode: "drill" })}>{L.retry}</Btn>
            </div>
          );
        })}
      </div>
    </div>
  );
};

const DebriefScreen = ({ T, lang, nav }) => {
  const D = window.EI_DATA;
  const [step, setStep] = React.useState(1);

  const L = lang === "en" ? {
    title: "Real-interview debrief",
    sub: "You just walked out of a room. Dump what happened — we'll turn it into the prep for next round.",
    step1: "Which role?", step2: "Questions they asked", step3: "What did you feel?",
    continue: "Continue", finish: "Turn into prep",
  } : {
    title: "真实面试复盘",
    sub: "刚从会议室出来？把刚才经历的记下来，我们帮你变成下一轮的准备。",
    step1: "是哪场面试？", step2: "他们问了哪些问题？", step3: "你当时的感受？",
    continue: "继续", finish: "沉淀为下一轮准备",
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 880, margin: "0 auto", padding: "40px 48px 96px" }}>
      <button onClick={() => nav("home")} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, display: "flex", gap: 6, alignItems: "center", marginBottom: 20 }}>
        <Icon name="arrow_left" size={14} /> {lang === "en" ? "Back" : "返回"}
      </button>

      <div style={{ marginBottom: 28 }}>
        <div className="ei-label" style={{ color: T.accent, marginBottom: 10 }}>{lang === "en" ? "POST-INTERVIEW" : "面后立即"}</div>
        <h1 className="ei-serif" style={{ fontSize: 36, color: T.ink, margin: 0, letterSpacing: "-0.02em" }}>{L.title}</h1>
        <div style={{ fontSize: 14.5, color: T.ink2, marginTop: 8 }}>{L.sub}</div>
      </div>

      {/* Step indicator */}
      <div style={{ display: "flex", gap: 10, marginBottom: 28 }}>
        {[1, 2, 3].map((n) => (
          <div key={n} style={{ flex: 1, height: 3, borderRadius: 2, background: n <= step ? T.accent : T.rule }} />
        ))}
      </div>

      <div style={{ background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2, padding: 28, minHeight: 400 }}>
        {step === 1 && (
          <div>
            <div className="ei-serif" style={{ fontSize: 22, color: T.ink, marginBottom: 20 }}>{L.step1}</div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {D.targetJobs.map((j) => (
                <label key={j.id} style={{ display: "flex", gap: 12, padding: "12px 14px", border: `1px solid ${T.rule}`, borderRadius: 2, cursor: "pointer", alignItems: "center" }}>
                  <input type="radio" name="job" defaultChecked={j.id === "tj-1"} />
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 14, color: T.ink, fontWeight: 500 }}>{j.title}</div>
                    <div style={{ fontSize: 12, color: T.ink3 }}>{j.company} · {j.location}</div>
                  </div>
                  <Tag tone="muted" T={T}>{j.status}</Tag>
                </label>
              ))}
              <button style={{ padding: "12px 14px", border: `1px dashed ${T.rule}`, borderRadius: 2, background: "transparent", color: T.ink3, fontSize: 13, display: "flex", gap: 8, alignItems: "center" }}>
                <Icon name="plus" size={14} /> {lang === "en" ? "Or add a new target job" : "或新建一个目标岗位"}
              </button>
            </div>
            <div style={{ marginTop: 20, display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
              <select style={{ padding: "10px 12px", border: `1px solid ${T.rule}`, borderRadius: 2, fontSize: 13, background: T.bg, color: T.ink }}>
                <option>{lang === "en" ? "Round · Technical" : "轮次 · 技术一面"}</option>
                <option>{lang === "en" ? "Round · HR" : "轮次 · HR"}</option>
                <option>{lang === "en" ? "Round · Manager" : "轮次 · 经理面"}</option>
              </select>
              <input type="date" defaultValue="2026-04-21" style={{ padding: "10px 12px", border: `1px solid ${T.rule}`, borderRadius: 2, fontSize: 13, background: T.bg, color: T.ink, fontFamily: "var(--ei-mono)" }} />
            </div>
          </div>
        )}

        {step === 2 && (
          <div>
            <div className="ei-serif" style={{ fontSize: 22, color: T.ink, marginBottom: 6 }}>{L.step2}</div>
            <div style={{ fontSize: 13, color: T.ink3, marginBottom: 16 }}>{lang === "en" ? "One per line. Roughly in the order they asked. We'll auto-tag the topics." : "一行一题，尽量按被问的顺序写。我们会自动打标签。"}</div>
            <textarea
              defaultValue={"你为什么要离开现在的公司？\n讲一个你最近做的性能优化项目，具体数字是多少？\n如果后端给你的接口返回慢，你怎么处理？\n你最大的短板是什么？\n你有什么想问我们的？"}
              style={{ width: "100%", minHeight: 280, padding: 16, border: `1px solid ${T.rule}`, borderRadius: 2, fontSize: 14, lineHeight: 1.7, fontFamily: "var(--ei-mono)", background: T.bg, color: T.ink, resize: "vertical", outline: "none" }}
            />
            <div style={{ fontSize: 12, color: T.ink3, marginTop: 10, display: "flex", gap: 6, alignItems: "center" }}>
              <Icon name="info" size={12} /> {lang === "en" ? "5 questions detected · 3 likely behavioral · 1 technical · 1 reverse-Q" : "识别到 5 道题 · 3 道行为题 · 1 道技术题 · 1 道反问"}
            </div>
          </div>
        )}

        {step === 3 && (
          <div>
            <div className="ei-serif" style={{ fontSize: 22, color: T.ink, marginBottom: 6 }}>{L.step3}</div>
            <div style={{ fontSize: 13, color: T.ink3, marginBottom: 16 }}>{lang === "en" ? "Don't polish — this feeds the next-round plan, not a report card." : "不用润色——这份是给下一轮准备用的，不是成绩单。"}</div>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12, marginBottom: 16 }}>
              {[
                { k: "confident", label: lang === "en" ? "Felt confident" : "感觉答得稳" },
                { k: "blank", label: lang === "en" ? "Went blank on a Q" : "有一题卡住了" },
                { k: "followed_up", label: lang === "en" ? "Got follow-ups" : "被追问多" },
                { k: "rushed", label: lang === "en" ? "Ran out of time" : "时间不够" },
              ].map((x) => (
                <label key={x.k} style={{ padding: "10px 12px", border: `1px solid ${T.rule}`, borderRadius: 2, fontSize: 13, color: T.ink2, display: "flex", gap: 8, alignItems: "center" }}>
                  <input type="checkbox" /> {x.label}
                </label>
              ))}
            </div>
            <textarea placeholder={lang === "en" ? "Anything else — something they said, a vibe, who was in the room…" : "还想记的——面试官的话、气氛、谁在场…"}
              style={{ width: "100%", minHeight: 140, padding: 14, border: `1px solid ${T.rule}`, borderRadius: 2, fontSize: 14, lineHeight: 1.6, background: T.bg, color: T.ink, resize: "vertical", outline: "none" }} />

            <div style={{ marginTop: 18, padding: 16, background: T.accentSoft, borderLeft: `3px solid ${T.accent}` }}>
              <div className="ei-label" style={{ color: T.accent, marginBottom: 6 }}>{lang === "en" ? "PREVIEW · WHAT YOU'LL GET" : "预览 · 你将得到"}</div>
              <ul style={{ margin: 0, paddingLeft: 20, fontSize: 13, color: T.ink2, lineHeight: 1.7 }}>
                <li>{lang === "en" ? "Better answers for each Q (with frameworks)" : "每题的更优答法（含框架）"}</li>
                <li>{lang === "en" ? "Thank-you note draft" : "感谢信草稿"}</li>
                <li>{lang === "en" ? "Next-round prep checklist" : "下一轮准备清单"}</li>
                <li>{lang === "en" ? "Weak points written to mistake book" : "薄弱点自动写入错题本"}</li>
              </ul>
            </div>
          </div>
        )}
      </div>

      <div style={{ display: "flex", justifyContent: "flex-end", gap: 10, marginTop: 20 }}>
        {step > 1 && <Btn variant="ghost" T={T} onClick={() => setStep(step - 1)} icon="arrow_left">{lang === "en" ? "Back" : "上一步"}</Btn>}
        <Btn variant="accent" T={T} onClick={() => step < 3 ? setStep(step + 1) : nav("report")} iconRight="arrow_right">
          {step < 3 ? L.continue : L.finish}
        </Btn>
      </div>

      {/* past debriefs */}
      <div style={{ marginTop: 48 }}>
        <SectionHeader eyebrow={lang === "en" ? "PAST DEBRIEFS" : "过往复盘"} title={lang === "en" ? "Also on the record" : "已沉淀"} T={T} />
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(320px, 1fr))", gap: 14 }}>
          {D.debriefs.map((d) => (
            <Card T={T} key={d.id} interactive>
              <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 8 }}>
                <div style={{ fontSize: 14, color: T.ink, fontWeight: 500 }}>{d.company} · {d.round}</div>
                <Tag tone={d.outcome === "通过" ? "ok" : "accent"} T={T}>{d.outcome}</Tag>
              </div>
              <div className="ei-mono" style={{ fontSize: 11, color: T.ink3, marginBottom: 10 }}>{d.date}</div>
              <div style={{ fontSize: 12.5, color: T.ink2, lineHeight: 1.5 }}>{lang === "en" ? "Hot:" : "高频："} {d.hotQuestions.join(" · ")}</div>
              <div style={{ fontSize: 12.5, color: T.accent, marginTop: 8 }}>→ {d.nextAction}</div>
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
};

const ResumeScreen = ({ T, lang, nav }) => {
  const D = window.EI_DATA;
  const r = D.resume;
  const L = lang === "en" ? {
    title: "Resume workshop",
    sub: "Not a layout tool — a mirror that shows what this specific JD wants you to say differently.",
    match: "JD match", yourResume: "Your resume · v3", suggestions: "What to change",
    before: "Now reads", after: "Try instead", why: "Why",
  } : {
    title: "简历工坊",
    sub: "不是排版工具——是一面镜子，告诉你这份 JD 希望你怎么说。",
    match: "JD 匹配度", yourResume: "你的简历 · v3", suggestions: "改哪几处",
    before: "现在的表达", after: "建议改成", why: "为什么",
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 1200, margin: "0 auto", padding: "40px 48px 96px" }}>
      <button onClick={() => nav("workspace", { jobId: "tj-1" })} style={{ background: "transparent", border: "none", color: T.ink3, fontSize: 13, display: "flex", gap: 6, marginBottom: 20 }}>
        <Icon name="arrow_left" size={14} /> {lang === "en" ? "Back to workspace" : "返回工作台"}
      </button>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginBottom: 32, gap: 24, flexWrap: "wrap" }}>
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "TAILORED · 星环科技 · 资深前端" : "已定制 · 星环科技 · 资深前端"}</div>
          <h1 className="ei-serif" style={{ fontSize: 36, color: T.ink, margin: 0, letterSpacing: "-0.02em" }}>{L.title}</h1>
          <div style={{ fontSize: 14.5, color: T.ink2, marginTop: 8, maxWidth: 580 }}>{L.sub}</div>
        </div>
        <div style={{ textAlign: "right" }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 6 }}>{L.match}</div>
          <div className="ei-mono" style={{ fontSize: 32, color: T.accent, fontWeight: 600 }}>{r.match}<span style={{ fontSize: 18, color: T.ink3, marginLeft: 2 }}>%</span></div>
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20 }}>
        {/* preview */}
        <div style={{ background: "#fff", border: `1px solid ${T.rule}`, borderRadius: 2, padding: 32, fontFamily: "Georgia, serif", minHeight: 600, color: "#222" }}>
          <div style={{ fontSize: 22, fontWeight: 600, marginBottom: 2 }}>林舟 · Lin Zhou</div>
          <div style={{ fontSize: 12, color: "#666" }}>{lang === "en" ? "Senior Frontend Engineer · 5 yrs · Shanghai" : "高级前端工程师 · 5 年 · 上海"}</div>
          <div style={{ height: 1, background: "#333", margin: "14px 0" }} />

          <div style={{ fontSize: 11, color: "#888", letterSpacing: "0.1em", marginBottom: 8, textTransform: "uppercase" }}>Experience</div>
          <div style={{ fontSize: 14, fontWeight: 600, marginBottom: 2 }}>Neoshop · Senior Frontend</div>
          <div style={{ fontSize: 12, color: "#666", marginBottom: 8 }}>2022 – Now · Shanghai</div>
          <ul style={{ fontSize: 13, margin: 0, paddingLeft: 18, lineHeight: 1.7, color: "#444" }}>
            <li style={{ background: "#fee5d6", padding: "2px 4px", borderRadius: 2, marginBottom: 6 }}>
              <span style={{ textDecoration: "line-through", opacity: 0.6 }}>负责订单系统前端开发，提升了用户体验。</span>
            </li>
            <li>重构复杂表单组件库，服务 20+ 内部产品线。</li>
            <li style={{ opacity: 0.5 }}>熟悉 Vue、Angular、React。</li>
            <li>推动前端监控体系搭建，覆盖 85% 核心页面。</li>
          </ul>

          <div style={{ marginTop: 16, fontSize: 11, color: "#888", letterSpacing: "0.1em", marginBottom: 6, textTransform: "uppercase" }}>Skills</div>
          <div style={{ fontSize: 13, lineHeight: 1.7, color: "#444" }}>TypeScript · React · Webpack · 微前端 · Design System</div>
        </div>

        {/* suggestions */}
        <div>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>▸ {L.suggestions}</div>
          {r.suggestions.map((s, i) => (
            <div key={i} style={{ marginBottom: 16, padding: 16, border: `1px solid ${T.rule}`, borderRadius: 2, background: T.bgCard }}>
              <div style={{ display: "flex", gap: 8, marginBottom: 10 }}>
                <Tag tone={s.type === "rewrite" ? "accent" : s.type === "add" ? "ok" : "warn"} T={T}>
                  {s.type === "rewrite" ? (lang === "en" ? "Rewrite" : "改写") : s.type === "add" ? (lang === "en" ? "Add" : "补充") : (lang === "en" ? "Trim" : "删除")}
                </Tag>
              </div>
              {s.before && (
                <div style={{ padding: "10px 12px", background: T.dangerSoft, color: T.ink2, fontSize: 13, lineHeight: 1.55, borderRadius: 2, marginBottom: 8, textDecoration: "line-through", opacity: 0.7 }}>
                  {s.before}
                </div>
              )}
              {s.after && (
                <div style={{ padding: "10px 12px", background: T.okSoft, color: T.ink, fontSize: 13, lineHeight: 1.55, borderRadius: 2, marginBottom: 8 }}>
                  {s.after}
                </div>
              )}
              {s.text && <div style={{ padding: "10px 12px", background: T.bgSoft, fontSize: 13, color: T.ink, borderRadius: 2, marginBottom: 8 }}>{s.text}</div>}
              <div style={{ fontSize: 12, color: T.ink3, lineHeight: 1.55 }}><b style={{ color: T.ink2 }}>{L.why}:</b> {s.reason}</div>
              <div style={{ marginTop: 10, display: "flex", gap: 8 }}>
                <Btn variant="secondary" size="sm" T={T} icon="check">{lang === "en" ? "Apply" : "采纳"}</Btn>
                <Btn variant="ghost" size="sm" T={T}>{lang === "en" ? "Skip" : "跳过"}</Btn>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

const GrowthScreen = ({ T, lang, nav }) => {
  const D = window.EI_DATA;
  const g = D.growth;
  const L = lang === "en" ? {
    title: "Growth board",
    sub: "Are you actually getting better? The only honest answer is drawn from 30 days of real sessions.",
    last30: "Last 30 days", practices: "Sessions", minutes: "Minutes", solved: "Mistakes solved",
    dims: "Dimension trends (last 7 weeks)", recent: "Recent sessions",
  } : {
    title: "成长看板",
    sub: "你是真的在变好吗？只看最近 30 天真实练习给出的答案。",
    last30: "近 30 天", practices: "练习场次", minutes: "总时长（分钟）", solved: "已攻克错题",
    dims: "能力维度趋势（最近 7 周）", recent: "近期会话",
  };

  return (
    <div className="ei-fadein" style={{ maxWidth: 1200, margin: "0 auto", padding: "40px 48px 96px" }}>
      <div style={{ marginBottom: 28 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{lang === "en" ? "GROWTH · BEYOND ONE JD" : "成长 · 跨岗位"}</div>
        <h1 className="ei-serif" style={{ fontSize: 38, color: T.ink, margin: 0, letterSpacing: "-0.02em" }}>{L.title}</h1>
        <div style={{ fontSize: 14.5, color: T.ink2, marginTop: 8, maxWidth: 600 }}>{L.sub}</div>
      </div>

      {/* top stats */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 14, marginBottom: 28 }}>
        <StatBlock T={T} label={L.practices} value={g.last30d.practices} delta="+4" positive />
        <StatBlock T={T} label={L.minutes} value={g.last30d.minutes} delta="+82" positive />
        <StatBlock T={T} label={L.solved} value={`${g.last30d.mistakesResolved}/${g.last30d.mistakesTotal}`} delta="50%" />
        <StatBlock T={T} label={lang === "en" ? "Avg readiness" : "平均准备度"} value={lang === "en" ? "Ready-ish" : "建议再练"} />
      </div>

      {/* dim trends */}
      <Card T={T} style={{ marginBottom: 28 }} pad={0}>
        <div style={{ padding: "18px 24px", borderBottom: `1px solid ${T.rule}`, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <div className="ei-serif" style={{ fontSize: 18, color: T.ink }}>{L.dims}</div>
          <div className="ei-mono" style={{ fontSize: 11, color: T.ink3 }}>{g.weeks[0]} — {g.weeks[g.weeks.length-1]}</div>
        </div>
        <div style={{ padding: 24 }}>
          <DimTrendChart T={T} trends={g.trendDim} weeks={g.weeks} />
        </div>
      </Card>

      {/* recent sessions + cross-job */}
      <div style={{ display: "grid", gridTemplateColumns: "1.4fr 1fr", gap: 20 }}>
        <Card T={T} pad={0}>
          <div style={{ padding: "16px 20px", borderBottom: `1px solid ${T.rule}` }}>
            <div className="ei-serif" style={{ fontSize: 17, color: T.ink }}>{L.recent}</div>
          </div>
          <div>
            {g.recent.map((r, i) => (
              <div key={i} onClick={() => nav("report")} style={{ padding: "14px 20px", borderBottom: i < g.recent.length - 1 ? `1px dotted ${T.rule}` : "none", display: "flex", alignItems: "center", gap: 14, cursor: "pointer" }}>
                <div className="ei-mono" style={{ fontSize: 12, color: T.ink3, width: 36 }}>{r.date}</div>
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: 13.5, color: T.ink }}>{r.mode}</div>
                  <div style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 2 }}>{r.job}</div>
                </div>
                <ReadinessDial level={r.readiness} T={T} size={32} />
                <Icon name="chevron_right" size={14} color={T.ink3} />
              </div>
            ))}
          </div>
        </Card>

        <Card T={T}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{lang === "en" ? "CROSS-JOB · WHERE YOU STAND" : "跨岗位 · 你现在的位置"}</div>
          {D.targetJobs.filter(j => j.status !== "草稿").map((j) => (
            <div key={j.id} style={{ marginBottom: 16 }}>
              <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 4 }}>
                <div style={{ fontSize: 13, color: T.ink, fontWeight: 500 }}>{j.company}</div>
                <div className="ei-mono" style={{ fontSize: 11, color: T.ink3 }}>{j.readinessLabel}</div>
              </div>
              <div style={{ height: 4, background: T.bgSoft, borderRadius: 2, overflow: "hidden" }}>
                <div style={{ width: `${(j.readiness + 1) * 25}%`, height: "100%", background: j.readiness >= 2 ? T.ok : T.amber }} />
              </div>
            </div>
          ))}
          <div style={{ padding: 12, background: T.bgSoft, borderRadius: 2, marginTop: 16, fontSize: 12.5, color: T.ink2, lineHeight: 1.55 }}>
            <b style={{ color: T.ink, display: "block", marginBottom: 4 }}>{lang === "en" ? "Suggestion" : "建议"}</b>
            {lang === "en" ? "Star-ring (星环) is close; push there first. Lumen Labs needs English-rhythm drills — schedule 2 this week." :
              "星环已接近可面，先冲。Lumen Labs 需要英文节奏练习——这周安排两次。"}
          </div>
        </Card>
      </div>
    </div>
  );
};

const StatBlock = ({ T, label, value, delta, positive }) => (
  <div style={{ padding: "18px 20px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
    <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{label}</div>
    <div style={{ display: "flex", alignItems: "baseline", gap: 8 }}>
      <div className="ei-serif" style={{ fontSize: 30, color: T.ink, letterSpacing: "-0.02em", fontWeight: 500 }}>{value}</div>
      {delta && <div style={{ fontSize: 12, color: positive ? T.ok : T.ink3, fontFamily: "var(--ei-mono)" }}>{delta}</div>}
    </div>
  </div>
);

const DimTrendChart = ({ T, trends, weeks }) => {
  const W = 1000, H = 260, pad = { t: 16, r: 16, b: 30, l: 40 };
  const iw = W - pad.l - pad.r, ih = H - pad.t - pad.b;
  const colors = [T.accent, T.cool, T.amber, T.ok, T.warn];
  return (
    <svg viewBox={`0 0 ${W} ${H}`} style={{ width: "100%", display: "block" }}>
      {[0, 25, 50, 75, 100].map((y) => {
        const yy = pad.t + ih - (y / 100) * ih;
        return (
          <g key={y}>
            <line x1={pad.l} y1={yy} x2={W - pad.r} y2={yy} stroke={T.rule} strokeWidth="1" strokeDasharray={y === 0 || y === 100 ? "" : "2 4"} />
            <text x={pad.l - 8} y={yy + 3} fontSize="10" fill={T.ink3} textAnchor="end" fontFamily="var(--ei-mono)">{y}</text>
          </g>
        );
      })}
      {weeks.map((w, i) => {
        const x = pad.l + (i / (weeks.length - 1)) * iw;
        return <text key={i} x={x} y={H - 12} fontSize="10" fill={T.ink3} textAnchor="middle" fontFamily="var(--ei-mono)">{w}</text>;
      })}
      {trends.map((tr, ti) => {
        const pts = tr.values.map((v, i) => {
          const x = pad.l + (i / (tr.values.length - 1)) * iw;
          const y = pad.t + ih - (v / 100) * ih;
          return `${x},${y}`;
        }).join(" ");
        return (
          <g key={ti}>
            <polyline points={pts} fill="none" stroke={colors[ti]} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
            {tr.values.map((v, i) => {
              const x = pad.l + (i / (tr.values.length - 1)) * iw;
              const y = pad.t + ih - (v / 100) * ih;
              return <circle key={i} cx={x} cy={y} r="2.5" fill={colors[ti]} />;
            })}
            <text x={W - pad.r + 4} y={pad.t + ih - (tr.values[tr.values.length-1] / 100) * ih + 3} fontSize="11" fill={colors[ti]} fontFamily="var(--ei-sans)" fontWeight="500">{tr.name}</text>
          </g>
        );
      })}
    </svg>
  );
};

Object.assign(window, { MistakesScreen, DebriefScreen, ResumeScreen, GrowthScreen });
