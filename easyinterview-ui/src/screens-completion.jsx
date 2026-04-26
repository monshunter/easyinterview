// EasyInterview · final-completion screens
// Welcome / Sign-in · Job status machine · Mistakes→Drill · Follow-up tree · STAR editor · Trust panel
// Self-contained: every component only depends on window primitives + EI_DATA.

(() => {
  const D = window.EI_DATA;

  // ────────────────────────────────────────────────────────────
  // 0 · Tiny shared helpers
  // ────────────────────────────────────────────────────────────

  const Stack = ({ gap = 12, children, style = {} }) => (
    <div style={{ display: "flex", flexDirection: "column", gap, ...style }}>{children}</div>
  );
  const Row = ({ gap = 12, align = "center", children, style = {} }) => (
    <div style={{ display: "flex", alignItems: align, gap, ...style }}>{children}</div>
  );
  const Hairline = ({ T, style = {} }) => (
    <div style={{ height: 1, background: T.rule, ...style }} />
  );
  const Dot = ({ color, size = 6 }) => (
    <span style={{ width: size, height: size, borderRadius: size, background: color, display: "inline-block" }} />
  );

  // ────────────────────────────────────────────────────────────
  // 1 · TRUST PANEL — universal "show evidence" component
  //    Wraps any AI-inferred claim. Click → reveals sources, model, confidence.
  // ────────────────────────────────────────────────────────────

  const TrustChip = ({ T, confidence = 0.78, lang = "zh", evidenceCount = 3 }) => (
    <span style={{
      display: "inline-flex", alignItems: "center", gap: 5, padding: "1px 7px 1px 5px",
      background: T.bgSoft, border: `1px solid ${T.rule}`, borderRadius: 999,
      fontSize: 10.5, color: T.ink3, fontFamily: "var(--ei-mono)", lineHeight: 1.4,
      cursor: "help", marginLeft: 6, verticalAlign: "1px",
    }}>
      <Dot color={confidence > 0.75 ? T.ok : confidence > 0.5 ? T.warn : T.danger} size={5} />
      {Math.round(confidence * 100)}% · {evidenceCount}{lang === "en" ? " refs" : " 处"}
    </span>
  );

  // The full panel — opens on click (rendered inline, not as portal, to keep simple)
  const TrustPanel = ({ T, lang = "zh", title, claim, confidence = 0.78,
                       evidence = [], model = "ei-model-2026.04", elapsed = "1.4s",
                       inputs = [], onClose, anchor = "right" }) => {
    const L = lang === "en"
      ? { why: "Why we said this", srcLabel: "Sources", confLabel: "Confidence",
          modelLabel: "Model", elapsedLabel: "Generated in", inputsLabel: "Inputs we used",
          override: "Disagree — override", helpful: "Mark accurate", close: "Close" }
      : { why: "为什么我们这么说", srcLabel: "证据", confLabel: "置信度",
          modelLabel: "模型", elapsedLabel: "用时", inputsLabel: "我们用了什么",
          override: "不同意 · 改写", helpful: "认为准确", close: "关闭" };

    return (
      <div style={{
        position: "absolute", [anchor === "right" ? "right" : "left"]: 0, top: "calc(100% + 8px)",
        width: 360, background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4,
        boxShadow: "0 12px 40px rgba(20,15,10,0.18)", padding: 18, zIndex: 50,
        fontFamily: "var(--ei-sans)",
      }}>
        <Row style={{ marginBottom: 10, justifyContent: "space-between" }}>
          <Row gap={6}>
            <Icon name="info" size={13} color={T.ink3} />
            <span className="ei-label" style={{ color: T.ink3 }}>{L.why}</span>
          </Row>
          <button onClick={onClose} style={{ background: "transparent", border: "none", cursor: "pointer", color: T.ink3, padding: 0 }}>
            <Icon name="x" size={12} />
          </button>
        </Row>

        <div className="ei-serif" style={{ fontSize: 15, color: T.ink, lineHeight: 1.5, marginBottom: 14 }}>
          "{claim}"
        </div>

        {/* Confidence bar */}
        <div style={{ marginBottom: 14 }}>
          <Row style={{ justifyContent: "space-between", marginBottom: 5 }}>
            <span style={{ fontSize: 11, color: T.ink3 }}>{L.confLabel}</span>
            <span style={{ fontSize: 11, color: T.ink2, fontFamily: "var(--ei-mono)" }}>
              {Math.round(confidence * 100)}%
            </span>
          </Row>
          <div style={{ height: 4, background: T.bgSoft, borderRadius: 2, overflow: "hidden" }}>
            <div style={{
              width: `${confidence * 100}%`, height: "100%",
              background: confidence > 0.75 ? T.ok : confidence > 0.5 ? T.warn : T.danger,
            }} />
          </div>
        </div>

        {/* Sources */}
        <div style={{ marginBottom: 14 }}>
          <span className="ei-label" style={{ color: T.ink3 }}>{L.srcLabel} · {evidence.length}</span>
          <Stack gap={8} style={{ marginTop: 8 }}>
            {evidence.map((e, i) => (
              <div key={i} style={{
                padding: 10, background: T.bgSoft, borderLeft: `2px solid ${T.accent}`,
                borderRadius: "0 2px 2px 0",
              }}>
                <Row style={{ justifyContent: "space-between", marginBottom: 4 }}>
                  <span style={{ fontSize: 10.5, color: T.ink3, fontFamily: "var(--ei-mono)", textTransform: "uppercase", letterSpacing: 0.5 }}>{e.tag}</span>
                  {e.locator && <span style={{ fontSize: 10.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{e.locator}</span>}
                </Row>
                <div style={{ fontSize: 12.5, color: T.ink, lineHeight: 1.55 }}>{e.text}</div>
              </div>
            ))}
          </Stack>
        </div>

        {/* Inputs */}
        {inputs.length > 0 && (
          <div style={{ marginBottom: 14 }}>
            <span className="ei-label" style={{ color: T.ink3 }}>{L.inputsLabel}</span>
            <Row gap={5} style={{ marginTop: 6, flexWrap: "wrap" }}>
              {inputs.map((inp, i) => <Tag key={i} tone="muted" T={T}>{inp}</Tag>)}
            </Row>
          </div>
        )}

        {/* Meta */}
        <Row style={{ paddingTop: 10, borderTop: `1px dotted ${T.rule}`, justifyContent: "space-between", color: T.ink3, fontSize: 10.5, fontFamily: "var(--ei-mono)" }}>
          <span>{L.modelLabel}: {model}</span>
          <span>{L.elapsedLabel}: {elapsed}</span>
        </Row>

        <Row style={{ marginTop: 12, justifyContent: "space-between" }}>
          <button style={{
            background: "transparent", border: `1px solid ${T.rule}`, padding: "5px 10px",
            borderRadius: 2, color: T.ink2, fontSize: 11.5, cursor: "pointer",
          }}>
            <Icon name="edit" size={11} /> {L.override}
          </button>
          <button style={{
            background: "transparent", border: "none", padding: "5px 4px",
            color: T.ink3, fontSize: 11.5, cursor: "pointer",
          }}>
            <Icon name="check" size={11} /> {L.helpful}
          </button>
        </Row>
      </div>
    );
  };

  // Wrapping helper — turns any inline text into a trust-attached span.
  const TrustClaim = ({ T, lang = "zh", children, ...panelProps }) => {
    const [open, setOpen] = React.useState(false);
    return (
      <span style={{ position: "relative", display: "inline-block" }}>
        <span
          onClick={() => setOpen(!open)}
          style={{
            cursor: "pointer",
            borderBottom: `1px dashed ${T.ink3}`,
            paddingBottom: 1,
          }}>
          {children}
          <TrustChip T={T} confidence={panelProps.confidence} lang={lang} evidenceCount={panelProps.evidence?.length || 0} />
        </span>
        {open && <TrustPanel T={T} lang={lang} {...panelProps} onClose={() => setOpen(false)} />}
      </span>
    );
  };

  window.TrustChip = TrustChip;
  window.TrustPanel = TrustPanel;
  window.TrustClaim = TrustClaim;

  // ────────────────────────────────────────────────────────────
  // 2 · M0 · WELCOME / SIGN-IN
  // ────────────────────────────────────────────────────────────

  const WelcomeScreen = ({ T, lang, nav, onSignIn }) => {
    const enter = onSignIn || (() => nav("home"));
    const [mode, setMode] = React.useState("welcome"); // welcome | signin | signup

    const L = lang === "en" ? {
      eyebrow: "EASYINTERVIEW · v1.0",
      h1: "Win the interview",
      h1b: "you already care about.",
      sub: "A practice loop that's tied to a specific job, a specific company, a specific you. Not generic question banks — the prep your next round actually needs.",
      cta: "Start with a JD",
      cta2: "I have an account",
      tag1: "Tied to one role",
      tag1Sub: "Each session is rooted in a real JD — practice, mistakes, resume edits all stay in that job's workspace.",
      tag2: "Evidence over scores",
      tag2Sub: "Every claim links back to a transcript timestamp. No black-box ratings.",
      tag3: "Round-by-round",
      tag3Sub: "HR → tech → manager. Each round inherits the signal from the last.",
      tag4: "Quiet by design",
      tag4Sub: "Local-first by default. You decide what we keep, what we forget, and when.",
      builtFor: "Built for people who interview for a living, not in theory.",
      footer: "© 2026 EasyInterview · 隐私优先 · privacy first",
    } : {
      eyebrow: "EASYINTERVIEW · v1.0",
      h1: "把你已经在乎的那场面试，",
      h1b: "稳稳地赢下来。",
      sub: "一个绑定具体岗位、具体公司、具体的你的练习闭环 —— 不是泛用题库，而是为下一轮真正用得上的弹药。",
      cta: "从一份 JD 开始",
      cta2: "我已经有账号",
      tag1: "绑定一个岗位",
      tag1Sub: "每一次练习都根植于一份真实 JD；错题、简历修改、复盘都留在这个岗位的工作台里。",
      tag2: "证据优先于打分",
      tag2Sub: "每一个判断都能跳回原话时间戳。没有黑箱评分。",
      tag3: "面到第几轮，准备到第几轮",
      tag3Sub: "HR → 技术 → 经理。每一轮都继承前一轮留下的信号。",
      tag4: "克制是默认值",
      tag4Sub: "默认本地优先。你决定我们记什么、忘什么、什么时候忘。",
      builtFor: "为那些把面试当作日常工作的人，而不是理论上谈论它的人。",
      footer: "© 2026 EasyInterview · 隐私优先",
    };

    if (mode !== "welcome") return <SignInScreen T={T} lang={lang} nav={nav} mode={mode} setMode={setMode} onSignIn={enter} />;

    return (
      <div style={{ minHeight: "100vh", background: T.bg, display: "flex", flexDirection: "column" }}>
        {/* Top bar */}
        <Row style={{ padding: "20px 56px", justifyContent: "space-between" }}>
          <Row gap={10}>
            <div style={{ width: 28, height: 28, borderRadius: 14, background: T.accent, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", fontFamily: "var(--ei-serif)", fontSize: 16, fontWeight: 600 }}>E</div>
            <div>
              <div className="ei-serif" style={{ fontSize: 17, color: T.ink, lineHeight: 1, letterSpacing: "-0.01em" }}>EasyInterview</div>
              <div className="ei-label" style={{ color: T.ink3, fontSize: 9, marginTop: 2 }}>面试训练器</div>
            </div>
          </Row>
          <Row gap={20}>
            <button onClick={() => setMode("signin")} style={{
              background: "transparent", border: "none", color: T.ink2, fontSize: 13.5, cursor: "pointer",
              fontFamily: "var(--ei-sans)",
            }}>
              {lang === "en" ? "Sign in" : "登录"}
            </button>
            <Btn T={T} variant="primary" size="sm" onClick={() => enter()}>
              {lang === "en" ? "Open app →" : "打开应用 →"}
            </Btn>
          </Row>
        </Row>

        {/* Hero */}
        <div style={{ flex: 1, display: "grid", gridTemplateColumns: "1.1fr 1fr", gap: 80, padding: "60px 80px 40px", maxWidth: 1440, margin: "0 auto", width: "100%", boxSizing: "border-box" }}>
          {/* Left: copy */}
          <div className="ei-fadein">
            <div className="ei-label" style={{ color: T.accent, marginBottom: 24 }}>{L.eyebrow}</div>
            <h1 className="ei-serif" style={{
              fontSize: 64, lineHeight: 1.05, letterSpacing: "-0.03em",
              color: T.ink, margin: 0, fontWeight: 500, textWrap: "balance",
            }}>
              {L.h1}<br />
              <span style={{ color: T.accent, fontStyle: "italic", fontWeight: 400 }}>{L.h1b}</span>
            </h1>
            <p style={{
              fontSize: 18, lineHeight: 1.6, color: T.ink2, marginTop: 28, maxWidth: 540,
              fontFamily: "var(--ei-serif)", textWrap: "pretty",
            }}>
              {L.sub}
            </p>
            <Row gap={14} style={{ marginTop: 36 }}>
              <Btn T={T} variant="primary" onClick={() => enter()} icon="arrow_right">{L.cta}</Btn>
              <Btn T={T} variant="ghost" onClick={() => setMode("signin")}>{L.cta2}</Btn>
            </Row>
            <div style={{
              marginTop: 60, paddingTop: 24, borderTop: `1px solid ${T.rule}`,
              fontSize: 13, color: T.ink3, fontStyle: "italic", fontFamily: "var(--ei-serif)", maxWidth: 480,
            }}>
              — {L.builtFor}
            </div>
          </div>

          {/* Right: editorial preview tile */}
          <div style={{ position: "relative" }}>
            <PreviewMontage T={T} lang={lang} />
          </div>
        </div>

        {/* Tag rows */}
        <div style={{ borderTop: `1px solid ${T.rule}`, padding: "44px 80px", maxWidth: 1440, margin: "0 auto", width: "100%", boxSizing: "border-box" }}>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 36 }}>
            {[[L.tag1, L.tag1Sub, "target"], [L.tag2, L.tag2Sub, "info"], [L.tag3, L.tag3Sub, "layers"], [L.tag4, L.tag4Sub, "settings"]].map(([t, s, ic], i) => (
              <div key={i}>
                <Row gap={8} style={{ marginBottom: 10 }}>
                  <Icon name={ic} size={14} color={T.accent} />
                  <span className="ei-label" style={{ color: T.ink3 }}>{`0${i + 1}`}</span>
                </Row>
                <div className="ei-serif" style={{ fontSize: 19, color: T.ink, marginBottom: 8, lineHeight: 1.3 }}>{t}</div>
                <div style={{ fontSize: 13.5, color: T.ink2, lineHeight: 1.55, textWrap: "pretty" }}>{s}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Footer */}
        <Row style={{ padding: "24px 56px", borderTop: `1px solid ${T.rule}`, justifyContent: "space-between" }}>
          <span style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{L.footer}</span>
          <Row gap={20}>
            <a style={{ fontSize: 11.5, color: T.ink3 }}>{lang === "en" ? "Privacy" : "隐私协议"}</a>
            <a style={{ fontSize: 11.5, color: T.ink3 }}>{lang === "en" ? "Terms" : "条款"}</a>
            <button onClick={() => {/* lang toggle */}} style={{
              background: "transparent", border: "none", color: T.ink3, fontSize: 11.5,
              cursor: "pointer", fontFamily: "var(--ei-mono)",
            }}>
              <Icon name="globe" size={11} /> 中 · EN
            </button>
          </Row>
        </Row>
      </div>
    );
  };

  // Editorial montage — preview of what's inside the app
  const PreviewMontage = ({ T, lang }) => {
    const labels = lang === "en"
      ? { tag: "TODAY · LIN ZHOU", title: "Manager round · Apr 24", sub: "Star Ring Tech · Senior Frontend",
          stat1: "Readiness", stat1v: "Get one more drill in",
          stat2: "Active jobs", stat2v: "3", stat3: "Open mistakes", stat3v: "5",
          line: "What I'd do for the manager round →",
          quote: "I'd push hard on the Design System rollout story — that's the one we haven't told yet, and it's exactly the influence-and-cross-team signal a manager round looks for. Let's build it now." }
      : { tag: "今天 · 林舟", title: "经理面 · 4月24日", sub: "星环科技 · 资深前端",
          stat1: "就绪度", stat1v: "再练一题", stat2: "进行中岗位", stat2v: "3", stat3: "未解决错题", stat3v: "5",
          line: "为经理面我会做什么 →",
          quote: "我会把 Design System 落地的故事补上 —— 这是目前还没讲过的故事，恰好对应经理面要听的「影响力 + 跨团队」信号。现在就把它建起来。" };
    return (
      <div style={{
        background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 4,
        boxShadow: "0 24px 64px rgba(20,15,10,0.10)", padding: 32,
        transform: "rotate(-0.5deg)", position: "relative",
      }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>{labels.tag}</div>
        <div className="ei-serif" style={{ fontSize: 26, color: T.ink, lineHeight: 1.2, marginBottom: 4, letterSpacing: "-0.015em" }}>
          {labels.title}
        </div>
        <div style={{ fontSize: 13, color: T.ink3, marginBottom: 22 }}>{labels.sub}</div>

        <div style={{ display: "grid", gridTemplateColumns: "1.4fr 1fr 1fr", gap: 14, marginBottom: 22 }}>
          {[[labels.stat1, labels.stat1v], [labels.stat2, labels.stat2v], [labels.stat3, labels.stat3v]].map(([k, v], i) => (
            <div key={i} style={{ padding: "12px 14px", background: i === 0 ? T.accentSoft : T.bgSoft, borderRadius: 2 }}>
              <div className="ei-label" style={{ color: i === 0 ? T.accent : T.ink3, marginBottom: 6, fontSize: 9 }}>{k}</div>
              <div className="ei-serif" style={{ fontSize: i === 0 ? 17 : 24, color: T.ink, lineHeight: 1.1, fontWeight: 500 }}>
                {v}
              </div>
            </div>
          ))}
        </div>

        <div style={{
          padding: "16px 18px", borderLeft: `2px solid ${T.accent}`, background: T.bgSoft,
          fontFamily: "var(--ei-serif)", fontSize: 14, color: T.ink2, lineHeight: 1.6, textWrap: "pretty",
        }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8, fontSize: 9 }}>{labels.line}</div>
          {labels.quote}
        </div>

        <Row style={{ marginTop: 18, justifyContent: "space-between" }}>
          <Row gap={6}>
            <Dot color={T.ok} />
            <span style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>SESSION 04 · 22 min</span>
          </Row>
          <span style={{ fontSize: 11, color: T.accent, fontFamily: "var(--ei-mono)" }}>→ {lang === "en" ? "open in app" : "进入工作台"}</span>
        </Row>
      </div>
    );
  };

  const SignInScreen = ({ T, lang, nav, mode, setMode, onSignIn }) => {
    const enter = onSignIn || (() => nav("home"));
    const isSignup = mode === "signup";
    const L = lang === "en" ? {
      tag: isSignup ? "CREATE ACCOUNT" : "WELCOME BACK",
      h1: isSignup ? "Start your prep loop." : "Open your workspace.",
      sub: isSignup ? "We'll keep your prep tied to real jobs — not generic banks." : "Pick up where you left off.",
      email: "Email", pw: "Password", name: "Your name",
      btn: isSignup ? "Create account" : "Sign in",
      sso: "Continue with",
      switchHint: isSignup ? "Already have an account?" : "First time here?",
      switchCta: isSignup ? "Sign in" : "Create one",
      back: "← Back",
    } : {
      tag: isSignup ? "创建账号" : "欢迎回来",
      h1: isSignup ? "开始你的准备闭环。" : "回到你的工作台。",
      sub: isSignup ? "我们让每一次准备都绑定真实岗位，而不是泛用题库。" : "从你上次停下的地方继续。",
      email: "邮箱", pw: "密码", name: "称呼",
      btn: isSignup ? "创建账号" : "登录",
      sso: "或使用",
      switchHint: isSignup ? "已经有账号？" : "第一次来？",
      switchCta: isSignup ? "登录" : "创建一个",
      back: "← 返回",
    };

    const inp = {
      width: "100%", padding: "12px 14px", border: `1px solid ${T.rule}`, borderRadius: 2,
      background: T.bg, color: T.ink, fontSize: 14, fontFamily: "var(--ei-sans)",
      outline: "none", boxSizing: "border-box",
    };
    const lbl = { fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", textTransform: "uppercase", letterSpacing: 0.4, marginBottom: 6, display: "block" };

    return (
      <div style={{ minHeight: "100vh", background: T.bg, display: "grid", gridTemplateColumns: "1fr 1.2fr" }}>
        {/* Left: pitch panel */}
        <div style={{ background: T.accent, color: "#fffaf3", padding: "60px 56px", display: "flex", flexDirection: "column", justifyContent: "space-between" }}>
          <Row gap={10}>
            <div style={{ width: 30, height: 30, borderRadius: 15, background: "rgba(255,255,255,0.2)", border: "1px solid rgba(255,255,255,0.3)", display: "flex", alignItems: "center", justifyContent: "center", color: "#fffaf3", fontFamily: "var(--ei-serif)", fontSize: 16, fontWeight: 600 }}>E</div>
            <div className="ei-serif" style={{ fontSize: 17 }}>EasyInterview</div>
          </Row>
          <div>
            <div className="ei-serif" style={{ fontSize: 38, lineHeight: 1.1, letterSpacing: "-0.02em", marginBottom: 20, textWrap: "balance" }}>
              {lang === "en"
                ? <>Practice the round in front of you, <em style={{ fontStyle: "italic", opacity: 0.85 }}>not the abstract one in your head.</em></>
                : <>练你眼前那一场面试，<em style={{ fontStyle: "italic", opacity: 0.85 }}>而不是脑子里那一场抽象的。</em></>}
            </div>
            <div style={{ fontSize: 14, opacity: 0.85, lineHeight: 1.6, fontFamily: "var(--ei-serif)" }}>
              {lang === "en" ? "Every session is rooted in a real JD. Every claim links back to a transcript timestamp. Local-first by default."
                            : "每一次练习都根植于一份真实 JD；每一条判断都能跳回原话；默认本地优先。"}
            </div>
          </div>
          <Row gap={16} style={{ fontSize: 11, fontFamily: "var(--ei-mono)", opacity: 0.65 }}>
            <span>· SOC 2 prep</span>
            <span>· GDPR friendly</span>
            <span>· EU + APAC</span>
          </Row>
        </div>

        {/* Right: form */}
        <div style={{ padding: "60px 80px", display: "flex", flexDirection: "column", justifyContent: "center", maxWidth: 560 }}>
          <button onClick={() => setMode("welcome")} style={{
            background: "transparent", border: "none", color: T.ink3, fontSize: 12.5,
            cursor: "pointer", marginBottom: 32, alignSelf: "flex-start", padding: 0,
            fontFamily: "var(--ei-sans)",
          }}>{L.back}</button>

          <div className="ei-label" style={{ color: T.accent, marginBottom: 12 }}>{L.tag}</div>
          <h1 className="ei-serif" style={{ fontSize: 36, color: T.ink, margin: 0, lineHeight: 1.1, letterSpacing: "-0.02em", marginBottom: 12 }}>
            {L.h1}
          </h1>
          <div style={{ fontSize: 14.5, color: T.ink2, marginBottom: 32, lineHeight: 1.5 }}>{L.sub}</div>

          <Stack gap={16}>
            {isSignup && (
              <div>
                <label style={lbl}>{L.name}</label>
                <input style={inp} placeholder={lang === "en" ? "Lin Zhou" : "林舟"} defaultValue="" />
              </div>
            )}
            <div>
              <label style={lbl}>{L.email}</label>
              <input style={inp} placeholder="lin.zhou@example.com" type="email" defaultValue="lin.zhou@example.com" />
            </div>
            <div>
              <label style={lbl}>{L.pw}</label>
              <input style={inp} type="password" defaultValue="••••••••••" />
            </div>
            <Btn T={T} variant="primary" onClick={() => enter()} style={{ width: "100%", justifyContent: "center", marginTop: 4 }}>
              {L.btn}
            </Btn>
          </Stack>

          <Row style={{ margin: "32px 0", gap: 12 }}>
            <Hairline T={T} style={{ flex: 1 }} />
            <span style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)", textTransform: "uppercase" }}>{L.sso}</span>
            <Hairline T={T} style={{ flex: 1 }} />
          </Row>

          <Row gap={10}>
            {[
              { name: "Google", glyph: "G" },
              { name: "Apple", glyph: "" },
              { name: "WeChat", glyph: "微" },
            ].map((p) => (
              <button key={p.name} onClick={() => enter()} style={{
                flex: 1, padding: "11px", border: `1px solid ${T.rule}`, borderRadius: 2,
                background: T.bgCard, color: T.ink2, fontSize: 13, cursor: "pointer",
                display: "flex", alignItems: "center", justifyContent: "center", gap: 8,
                fontFamily: "var(--ei-sans)",
              }}>
                <span style={{ fontFamily: "var(--ei-serif)", fontSize: 14, color: T.ink }}>{p.glyph}</span>
                {p.name}
              </button>
            ))}
          </Row>

          <div style={{ marginTop: 32, fontSize: 12.5, color: T.ink3, textAlign: "center" }}>
            {L.switchHint}{" "}
            <button onClick={() => setMode(isSignup ? "signin" : "signup")} style={{
              background: "transparent", border: "none", color: T.accent, fontSize: 12.5,
              cursor: "pointer", padding: 0, textDecoration: "underline", fontFamily: "var(--ei-sans)",
            }}>{L.switchCta}</button>
          </div>
        </div>
      </div>
    );
  };

  // ────────────────────────────────────────────────────────────
  // 3 · JOB STATUS MACHINE
  //    A 5-state pipeline component — drop-in for any job card / workspace header.
  // ────────────────────────────────────────────────────────────

  const JOB_STATES_ZH = [
    { k: "draft",      label: "草稿",       sub: "JD 已存，未开始" },
    { k: "preparing",  label: "准备中",     sub: "练习与简历调整中" },
    { k: "interviewing", label: "面试中",   sub: "已进入面试轮次" },
    { k: "offer",      label: "已发 offer", sub: "等待签字 / 谈判" },
    { k: "closed",     label: "已结束",     sub: "通过 / 婉拒 / 终止" },
  ];
  const JOB_STATES_EN = [
    { k: "draft",      label: "Draft",       sub: "JD saved, not started" },
    { k: "preparing",  label: "Preparing",   sub: "Practice + resume edits" },
    { k: "interviewing", label: "Interviewing", sub: "In an active loop" },
    { k: "offer",      label: "Offer",       sub: "Awaiting sign / negotiation" },
    { k: "closed",     label: "Closed",      sub: "Won / declined / ended" },
  ];

  const JobStatusPipeline = ({ T, lang, currentKey = "interviewing", onChange, compact = false }) => {
    const states = lang === "en" ? JOB_STATES_EN : JOB_STATES_ZH;
    const idx = states.findIndex((s) => s.k === currentKey);

    if (compact) {
      // Single-line condensed pill row
      return (
        <Row gap={0} style={{ background: T.bgSoft, padding: 4, borderRadius: 2, border: `1px solid ${T.rule}` }}>
          {states.map((s, i) => {
            const done = i < idx, active = i === idx, future = i > idx;
            return (
              <button key={s.k} onClick={() => onChange?.(s.k)} title={s.sub} style={{
                flex: 1, padding: "5px 8px", border: "none", borderRadius: 2,
                background: active ? T.accent : "transparent",
                color: active ? "#fff" : done ? T.ink2 : T.ink3,
                fontSize: 11.5, fontWeight: active ? 500 : 400, cursor: "pointer",
                fontFamily: "var(--ei-sans)",
                opacity: future ? 0.5 : 1,
              }}>
                {s.label}
              </button>
            );
          })}
        </Row>
      );
    }

    // Full editorial pipeline
    return (
      <div style={{ position: "relative" }}>
        <div style={{ position: "absolute", top: 11, left: 14, right: 14, height: 1, background: T.rule, zIndex: 0 }} />
        <div style={{ position: "absolute", top: 11, left: 14, height: 1, background: T.accent, zIndex: 1, width: `calc(${(idx / (states.length - 1)) * 100}% - 14px + 14px)`, transition: "width .2s" }} />
        <div style={{ display: "grid", gridTemplateColumns: `repeat(${states.length}, 1fr)`, position: "relative", zIndex: 2 }}>
          {states.map((s, i) => {
            const done = i < idx, active = i === idx;
            return (
              <button key={s.k} onClick={() => onChange?.(s.k)} style={{
                background: "transparent", border: "none", padding: "0 8px", cursor: onChange ? "pointer" : "default",
                display: "flex", flexDirection: "column", alignItems: "flex-start", textAlign: "left",
                fontFamily: "var(--ei-sans)",
              }}>
                <div style={{
                  width: 22, height: 22, borderRadius: 11,
                  background: done ? T.accent : active ? T.accent : T.bg,
                  border: `2px solid ${done || active ? T.accent : T.rule}`,
                  display: "flex", alignItems: "center", justifyContent: "center",
                  marginBottom: 8,
                  boxShadow: active ? `0 0 0 4px ${T.accentSoft}` : "none",
                }}>
                  {done && <Icon name="check" size={11} color="#fff" stroke={2.5} />}
                  {active && <div style={{ width: 6, height: 6, borderRadius: 3, background: "#fff" }} />}
                </div>
                <div style={{
                  fontSize: 12, fontWeight: active ? 500 : 400,
                  color: active ? T.ink : done ? T.ink2 : T.ink3,
                }}>{s.label}</div>
                <div style={{ fontSize: 10.5, color: T.ink3, marginTop: 3, lineHeight: 1.4, maxWidth: 130, fontFamily: "var(--ei-mono)" }}>
                  {s.sub}
                </div>
              </button>
            );
          })}
        </div>
      </div>
    );
  };

  // Status-detail card — what to do next at each state
  const StatusDetailPanel = ({ T, lang, jobStateKey = "interviewing" }) => {
    const advice = lang === "en" ? {
      draft: { next: "Open the workspace and confirm the JD parse before practicing.", actions: ["Confirm JD parse", "Start onboarding"] },
      preparing: { next: "Run two more practice sessions, then commit a resume bullet edit.", actions: ["Schedule a drill", "Open resume diff"] },
      interviewing: { next: "Manager round in 2 days — drill cross-team influence story.", actions: ["Drill mistakes (5)", "Open multi-round plan"] },
      offer: { next: "Use the comp negotiation playbook before signing.", actions: ["Open negotiation guide", "Mark as won"] },
      closed: { next: "Run debrief — what to carry into the next loop?", actions: ["Open debrief", "Archive"] },
    } : {
      draft: { next: "进入工作台，确认 JD 解析结果，然后开始练习。", actions: ["确认 JD 解析", "进入 onboarding"] },
      preparing: { next: "再练 2 场，然后改一条简历 bullet 提交。", actions: ["安排一场练习", "打开简历 diff"] },
      interviewing: { next: "经理面 2 天后 —— 把跨团队影响力的故事再练 1 次。", actions: ["复练错题 (5)", "打开多轮计划"] },
      offer: { next: "签字前过一遍薪酬谈判清单。", actions: ["打开谈判清单", "标记为成功"] },
      closed: { next: "做一次复盘 —— 把信号带去下一份准备里。", actions: ["打开复盘", "归档"] },
    };
    const a = advice[jobStateKey];
    return (
      <div style={{ padding: 14, background: T.bgSoft, borderRadius: 2, borderLeft: `2px solid ${T.accent}` }}>
        <div className="ei-label" style={{ color: T.accent, marginBottom: 6 }}>
          {lang === "en" ? "WHAT'S NEXT" : "下一步"}
        </div>
        <div style={{ fontSize: 13.5, color: T.ink, lineHeight: 1.55, marginBottom: 10, fontFamily: "var(--ei-serif)" }}>
          {a.next}
        </div>
        <Row gap={6} style={{ flexWrap: "wrap" }}>
          {a.actions.map((act, i) => (
            <button key={i} style={{
              padding: "5px 10px", background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2,
              color: T.ink2, fontSize: 11.5, cursor: "pointer", fontFamily: "var(--ei-sans)",
            }}>
              {act}
            </button>
          ))}
        </Row>
      </div>
    );
  };

  window.JobStatusPipeline = JobStatusPipeline;
  window.StatusDetailPanel = StatusDetailPanel;

  // ────────────────────────────────────────────────────────────
  // 4 · MISTAKES → DRILL FLOW
  //    From the mistake list, build a custom drill — pick which to retry,
  //    pick depth (light / focused / deep), pick mode (typed / voice).
  // ────────────────────────────────────────────────────────────

  const DrillBuilderScreen = ({ T, lang, nav }) => {
    const [selected, setSelected] = React.useState(() => new Set(D.mistakes.slice(0, 3).map((m) => m.id)));
    const [depth, setDepth] = React.useState("focused");
    const [mode, setMode] = React.useState("typed");
    const [running, setRunning] = React.useState(false);

    const L = lang === "en" ? {
      tag: "TARGETED DRILL · BUILD A SESSION",
      h1: "Drill the questions you actually missed.",
      sub: "Pick which mistakes go into this round, how deep we go, and how you want to deliver — typed or spoken. Each drill writes back to your mistake book.",
      step1: "1 · Pick your mistakes",
      step1Sub: "Open mistakes only by default. Tap to include / exclude.",
      onlyOpen: "Only unsolved",
      step2: "2 · How deep?",
      step3: "3 · How will you answer?",
      summaryTitle: "Your drill",
      summary: "questions",
      depth: { light: "Light · 2 min each", focused: "Focused · 4 min each + one follow-up", deep: "Deep · 6 min + 3 follow-ups" },
      mode: { typed: "Typed", voice: "Voice (live transcription + pace feedback)" },
      start: "Start drill",
      empty: "Nothing selected — pick at least one to start.",
    } : {
      tag: "针对性复练 · 自定义一场",
      h1: "只练你真正卡住的那几题。",
      sub: "挑题、挑深度、挑形式 —— 文字或语音。每一场复练都会回写到错题本里，保留改善轨迹。",
      step1: "1 · 选你要练的题",
      step1Sub: "默认只显示未解决的。点一下加入或移出。",
      onlyOpen: "只看未解决",
      step2: "2 · 想练多深？",
      step3: "3 · 怎么回答？",
      summaryTitle: "本次复练",
      summary: "道",
      depth: { light: "轻量 · 每题 2 分钟", focused: "聚焦 · 每题 4 分钟 + 1 次追问", deep: "深度 · 每题 6 分钟 + 3 次追问" },
      mode: { typed: "文字", voice: "语音（实时转写 + 语速反馈）" },
      start: "开始复练",
      empty: "还没选题 —— 至少挑一个再开始。",
    };

    const toggle = (id) => {
      const next = new Set(selected);
      next.has(id) ? next.delete(id) : next.add(id);
      setSelected(next);
    };

    if (running) {
      return <DrillRunningScreen T={T} lang={lang} nav={nav}
        ids={[...selected]} depth={depth} mode={mode}
        onExit={() => setRunning(false)} />;
    }

    return (
      <div style={{ maxWidth: 1100, margin: "0 auto", padding: "40px 56px 80px" }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{L.tag}</div>
        <h1 className="ei-serif" style={{ fontSize: 38, color: T.ink, margin: 0, letterSpacing: "-0.02em", lineHeight: 1.15, marginBottom: 12 }}>
          {L.h1}
        </h1>
        <p style={{ fontSize: 14.5, color: T.ink2, lineHeight: 1.6, maxWidth: 720, marginBottom: 36, fontFamily: "var(--ei-serif)" }}>
          {L.sub}
        </p>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 320px", gap: 32 }}>
          {/* LEFT */}
          <Stack gap={32}>
            {/* Step 1: pick mistakes */}
            <div>
              <Row style={{ justifyContent: "space-between", marginBottom: 8 }}>
                <h2 className="ei-serif" style={{ fontSize: 18, color: T.ink, margin: 0, fontWeight: 500 }}>{L.step1}</h2>
                <span className="ei-label" style={{ color: T.ink3 }}>{L.onlyOpen}</span>
              </Row>
              <div style={{ fontSize: 13, color: T.ink3, marginBottom: 14 }}>{L.step1Sub}</div>

              <Stack gap={8}>
                {D.mistakes.map((m) => {
                  const on = selected.has(m.id);
                  const tone = m.status === "已攻克" ? "ok" : m.status === "改善中" ? "amber" : "danger";
                  return (
                    <button key={m.id} onClick={() => toggle(m.id)} style={{
                      textAlign: "left", padding: "14px 16px", borderRadius: 2,
                      background: on ? T.bgCard : T.bg,
                      border: `1px solid ${on ? T.accent : T.rule}`,
                      cursor: "pointer", display: "grid", gridTemplateColumns: "20px 1fr auto auto",
                      gap: 14, alignItems: "center", fontFamily: "var(--ei-sans)",
                      boxShadow: on ? `0 0 0 3px ${T.accentSoft}` : "none",
                    }}>
                      <div style={{
                        width: 16, height: 16, borderRadius: 2, border: `1.5px solid ${on ? T.accent : T.rule}`,
                        background: on ? T.accent : "transparent", display: "flex", alignItems: "center", justifyContent: "center",
                      }}>
                        {on && <Icon name="check" size={10} color="#fff" stroke={3} />}
                      </div>
                      <div>
                        <div style={{ fontSize: 14, color: T.ink, marginBottom: 4, fontWeight: 500 }}>{m.ability}</div>
                        <div style={{ fontSize: 12.5, color: T.ink3, lineHeight: 1.5 }}>{m.question}</div>
                      </div>
                      <div style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{m.attempts}× · {m.lastAttempt}</div>
                      <Tag tone={tone} T={T}>{m.status}</Tag>
                    </button>
                  );
                })}
              </Stack>
            </div>

            {/* Step 2: depth */}
            <div>
              <h2 className="ei-serif" style={{ fontSize: 18, color: T.ink, margin: "0 0 14px", fontWeight: 500 }}>{L.step2}</h2>
              <Row gap={10}>
                {["light", "focused", "deep"].map((d) => (
                  <button key={d} onClick={() => setDepth(d)} style={{
                    flex: 1, padding: 14, borderRadius: 2,
                    border: `1px solid ${depth === d ? T.accent : T.rule}`,
                    background: depth === d ? T.accentSoft : T.bg,
                    cursor: "pointer", textAlign: "left", fontFamily: "var(--ei-sans)",
                  }}>
                    <div style={{ fontSize: 13, color: T.ink, fontWeight: 500, marginBottom: 6 }}>
                      {L.depth[d]}
                    </div>
                    <div style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
                      {d === "light" ? "≈10 min" : d === "focused" ? "≈18 min" : "≈30 min"} · {selected.size} {lang === "en" ? "Q" : "题"}
                    </div>
                  </button>
                ))}
              </Row>
            </div>

            {/* Step 3: mode */}
            <div>
              <h2 className="ei-serif" style={{ fontSize: 18, color: T.ink, margin: "0 0 14px", fontWeight: 500 }}>{L.step3}</h2>
              <Row gap={10}>
                {["typed", "voice"].map((mo) => (
                  <button key={mo} onClick={() => setMode(mo)} style={{
                    flex: 1, padding: 14, borderRadius: 2,
                    border: `1px solid ${mode === mo ? T.accent : T.rule}`,
                    background: mode === mo ? T.accentSoft : T.bg,
                    cursor: "pointer", textAlign: "left", fontFamily: "var(--ei-sans)",
                    display: "flex", alignItems: "center", gap: 10,
                  }}>
                    <Icon name={mo === "typed" ? "chat" : "mic"} size={16} color={mode === mo ? T.accent : T.ink2} />
                    <div>
                      <div style={{ fontSize: 13, color: T.ink, fontWeight: 500 }}>{L.mode[mo]}</div>
                    </div>
                  </button>
                ))}
              </Row>
            </div>
          </Stack>

          {/* RIGHT — sticky summary */}
          <div style={{ position: "sticky", top: 80, alignSelf: "flex-start" }}>
            <Card T={T} style={{ padding: 22 }}>
              <div className="ei-label" style={{ color: T.ink3, marginBottom: 10 }}>{L.summaryTitle}</div>
              <div className="ei-serif" style={{ fontSize: 38, color: T.ink, lineHeight: 1, marginBottom: 4, fontWeight: 500 }}>
                {selected.size}
              </div>
              <div style={{ fontSize: 13, color: T.ink3, marginBottom: 18 }}>{L.summary}</div>

              <Stack gap={10} style={{ paddingTop: 14, borderTop: `1px dotted ${T.rule}` }}>
                <Row style={{ justifyContent: "space-between" }}>
                  <span style={{ fontSize: 12, color: T.ink3 }}>{lang === "en" ? "Depth" : "深度"}</span>
                  <span style={{ fontSize: 12.5, color: T.ink, fontWeight: 500 }}>{L.depth[depth].split(" · ")[0]}</span>
                </Row>
                <Row style={{ justifyContent: "space-between" }}>
                  <span style={{ fontSize: 12, color: T.ink3 }}>{lang === "en" ? "Mode" : "形式"}</span>
                  <span style={{ fontSize: 12.5, color: T.ink, fontWeight: 500 }}>{L.mode[mode].split(" (")[0]}</span>
                </Row>
                <Row style={{ justifyContent: "space-between" }}>
                  <span style={{ fontSize: 12, color: T.ink3 }}>{lang === "en" ? "Est. time" : "预计用时"}</span>
                  <span style={{ fontSize: 12.5, color: T.ink, fontFamily: "var(--ei-mono)" }}>
                    ≈ {selected.size * (depth === "light" ? 2 : depth === "focused" ? 4 : 6)} min
                  </span>
                </Row>
              </Stack>

              <Btn T={T} variant="primary" disabled={selected.size === 0}
                onClick={() => setRunning(true)}
                style={{ width: "100%", justifyContent: "center", marginTop: 18,
                  opacity: selected.size === 0 ? 0.4 : 1 }}>
                <Icon name="play" size={11} /> {L.start}
              </Btn>
              {selected.size === 0 && (
                <div style={{ fontSize: 11, color: T.warn, marginTop: 8, textAlign: "center" }}>{L.empty}</div>
              )}
            </Card>

            <div style={{ marginTop: 14, fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)", lineHeight: 1.5 }}>
              {lang === "en"
                ? "→ Drill writes back to mistakes book."
                : "→ 复练结果会自动回写错题本。"}
            </div>
          </div>
        </div>
      </div>
    );
  };

  const DrillRunningScreen = ({ T, lang, nav, ids, depth, mode, onExit }) => {
    const [idx, setIdx] = React.useState(0);
    const items = D.mistakes.filter((m) => ids.includes(m.id));
    const cur = items[idx];

    const L = lang === "en" ? {
      tag: "DRILL", of: "of", followUp: "Possible follow-up",
      starHint: "Try a STAR-shaped answer.", next: "Next question →", done: "Finish drill",
      placeholder: "Type your answer here…",
    } : {
      tag: "复练中", of: "/", followUp: "可能的追问",
      starHint: "试着用 STAR 结构作答。", next: "下一题 →", done: "完成本次复练",
      placeholder: "在这里写下你的回答…",
    };

    const advance = () => {
      if (idx < items.length - 1) setIdx(idx + 1);
      else { onExit(); nav("mistakes"); }
    };

    return (
      <div style={{ minHeight: "100vh", background: T.bg, display: "flex", flexDirection: "column" }}>
        {/* Top */}
        <Row style={{ padding: "14px 28px", borderBottom: `1px solid ${T.rule}`, justifyContent: "space-between" }}>
          <Row gap={14}>
            <button onClick={onExit} style={{ background: "transparent", border: "none", cursor: "pointer", color: T.ink3, padding: 0 }}>
              <Icon name="x" size={16} />
            </button>
            <span className="ei-label" style={{ color: T.accent }}>{L.tag} · {cur.ability}</span>
            <span style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>{idx + 1} {L.of} {items.length}</span>
          </Row>
          <Row gap={6}>
            {items.map((_, i) => (
              <Dot key={i} color={i < idx ? T.ok : i === idx ? T.accent : T.rule} size={6} />
            ))}
          </Row>
        </Row>

        {/* Body */}
        <div style={{ flex: 1, maxWidth: 880, margin: "0 auto", padding: "60px 56px", width: "100%", boxSizing: "border-box" }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 14 }}>
            {lang === "en" ? "QUESTION" : "题面"} · {cur.attempts}× {lang === "en" ? "before" : "之前练过"}
          </div>
          <h1 className="ei-serif" style={{
            fontSize: 32, color: T.ink, margin: 0, lineHeight: 1.3, letterSpacing: "-0.015em", marginBottom: 24,
            textWrap: "pretty",
          }}>
            {cur.question}
          </h1>

          <div style={{ padding: "12px 16px", background: T.bgSoft, borderLeft: `2px solid ${T.accent}`, borderRadius: 2, marginBottom: 24 }}>
            <div style={{ fontSize: 12, color: T.ink3, marginBottom: 4 }}>
              <Icon name="info" size={12} /> {lang === "en" ? "Last time you got stuck on:" : "上次你卡在了:"}
            </div>
            <div style={{ fontSize: 13.5, color: T.ink, lineHeight: 1.5, fontFamily: "var(--ei-serif)" }}>
              {cur.lastStuck || (lang === "en"
                ? "no concrete numbers; the result was not measurable."
                : "答案缺乏量化指标，结果难以验证。")}
            </div>
          </div>

          {mode === "typed" ? (
            <textarea
              placeholder={L.placeholder}
              style={{
                width: "100%", minHeight: 220, padding: 18, border: `1px solid ${T.rule}`, borderRadius: 2,
                background: T.bgCard, color: T.ink, fontSize: 15, lineHeight: 1.6, fontFamily: "var(--ei-serif)",
                outline: "none", resize: "vertical", boxSizing: "border-box",
              }}
            />
          ) : (
            <VoiceMockPanel T={T} lang={lang} />
          )}

          <Row style={{ justifyContent: "space-between", marginTop: 24 }}>
            <span style={{ fontSize: 12.5, color: T.ink3, fontStyle: "italic", fontFamily: "var(--ei-serif)" }}>
              💡 {L.starHint}
            </span>
            <Btn T={T} variant="primary" onClick={advance}>
              {idx < items.length - 1 ? L.next : L.done}
            </Btn>
          </Row>
        </div>
      </div>
    );
  };

  const VoiceMockPanel = ({ T, lang }) => {
    const bars = Array.from({ length: 60 }, (_, i) => 4 + Math.abs(Math.sin(i * 0.7) + Math.cos(i * 0.3)) * 16);
    return (
      <div style={{ padding: 24, background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2 }}>
        <Row style={{ justifyContent: "space-between", marginBottom: 18 }}>
          <Row gap={8}>
            <Dot color={T.danger} size={8} />
            <span style={{ fontSize: 12.5, color: T.ink, fontFamily: "var(--ei-mono)" }}>
              {lang === "en" ? "Recording — speak naturally" : "录音中 · 自然作答即可"}
            </span>
          </Row>
          <span style={{ fontSize: 12, color: T.ink3, fontFamily: "var(--ei-mono)" }}>00:42 / 04:00</span>
        </Row>
        <div style={{ display: "flex", gap: 2, alignItems: "center", height: 40, marginBottom: 16 }}>
          {bars.map((h, i) => (
            <div key={i} style={{ flex: 1, height: h, background: i < 28 ? T.ink2 : T.accent, borderRadius: 1 }} />
          ))}
        </div>
        <Row style={{ justifyContent: "space-between", fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
          <span>~ 175 wpm</span>
          <span>2 long pauses</span>
          <span>0 fillers</span>
        </Row>
      </div>
    );
  };

  window.DrillBuilderScreen = DrillBuilderScreen;

  // ────────────────────────────────────────────────────────────
  // 5 · FOLLOW-UP TREE
  //    For any practice question, surface 3 possible follow-up angles
  //    + how each one tests something different.
  // ────────────────────────────────────────────────────────────

  const FollowUpTreeScreen = ({ T, lang, nav, embedded }) => {
    const baseQ = D.questions.find((q) => q.id === "q2") || D.questions[1];

    const branches = lang === "en" ? [
      { angle: "Numbers angle", tests: "Quantification + measurement rigor",
        question: "What was your starting metric and the final number? How did you instrument it?",
        why: "Tests whether you can connect engineering work to a measurable business / UX outcome — managers care most.",
        risk: "If you only have qualitative results, this is the angle that exposes it.", weight: 0.85, color: T.accent },
      { angle: "Trade-off angle", tests: "Architectural judgment",
        question: "What did you NOT do? What would the team have shipped instead?",
        why: "Tests whether you actually weighed alternatives, or just rationalized one path.",
        risk: "Common follow-up from staff-level interviewers.", weight: 0.72, color: T.amber },
      { angle: "Influence angle", tests: "Cross-team collaboration",
        question: "Who pushed back, and how did you handle the disagreement?",
        why: "Tests soft signals — Ownership, conflict resolution, persuasion.",
        risk: "Manager round almost always asks a version of this.", weight: 0.78, color: T.cool },
    ] : [
      { angle: "数字角度", tests: "量化能力 + 度量严谨",
        question: "起点指标是多少？最终落到什么数字？你是怎么埋点和度量的？",
        why: "测试你能否把工程工作连接到可验证的业务 / 体验指标 —— 经理面最在意这条线。",
        risk: "如果你只有定性结果，这一问会立刻暴露。", weight: 0.85, color: T.accent },
      { angle: "权衡角度", tests: "架构判断",
        question: "你 *没* 选什么方案？团队本来可能会做什么？",
        why: "测试你是否真的权衡过备选项，还是只是事后合理化了一条路径。",
        risk: "P7+ / Staff 级面试官的高频追问。", weight: 0.72, color: T.amber },
      { angle: "影响力角度", tests: "跨团队协作",
        question: "谁反对过你？你是怎么处理这次分歧的？",
        why: "测试软信号 —— Ownership、冲突处理、说服力。",
        risk: "经理面几乎一定会出这一类问题。", weight: 0.78, color: T.cool },
    ];

    const L = lang === "en" ? {
      tag: "FOLLOW-UP TREE",
      h1: "One question, three places it can go.",
      sub: "AI doesn't ask the same follow-up the same way. We surface three plausible branches and what each one is testing — so you can prepare answers across, not just deeper.",
      base: "BASE QUESTION",
      legend: "Likelihood (× this round)",
      tries: "What it tests", why: "Why they'd ask", risk: "Where you might trip",
      drill: "Drill this branch",
      cta: "Start a drill across all 3",
    } : {
      tag: "追问树",
      h1: "同一题，三种走向。",
      sub: "AI 不会用同一种方式追同一道题。我们把可能的三条分支拿出来，让你横向准备 —— 不只是把答案做深，而是把准备做宽。",
      base: "母题",
      legend: "本轮出现概率",
      tries: "测试什么", why: "为什么会问", risk: "你可能在哪卡住",
      drill: "练这条分支",
      cta: "横向练完三条",
    };

    return (
      <div style={{ maxWidth: 1280, margin: "0 auto", padding: embedded ? "24px 40px 60px" : "40px 56px 80px" }}>
        {!embedded && <div className="ei-label" style={{ color: T.ink3, marginBottom: 12 }}>{L.tag}</div>}
        {!embedded && (<h1 className="ei-serif" style={{ fontSize: 38, color: T.ink, margin: 0, letterSpacing: "-0.02em", lineHeight: 1.15, marginBottom: 14 }}>
          {L.h1}
        </h1>)}
        {!embedded && (<p style={{ fontSize: 14.5, color: T.ink2, lineHeight: 1.6, maxWidth: 800, marginBottom: 32, fontFamily: "var(--ei-serif)", textWrap: "pretty" }}>
          {L.sub}
        </p>)}

        {/* Base question */}
        <div style={{
          padding: "24px 28px", background: T.bgCard, border: `1px solid ${T.rule}`,
          borderRadius: 2, marginBottom: 36, position: "relative",
        }}>
          <Row style={{ justifyContent: "space-between", marginBottom: 10 }}>
            <span className="ei-label" style={{ color: T.accent }}>{L.base} · Q{baseQ.n}</span>
            <Tag tone="muted" T={T}>{baseQ.topic}</Tag>
          </Row>
          <div className="ei-serif" style={{ fontSize: 22, color: T.ink, lineHeight: 1.35, letterSpacing: "-0.01em", textWrap: "pretty" }}>
            {baseQ.prompt}
          </div>
        </div>

        {/* Tree */}
        <div style={{ position: "relative" }}>
          {/* Connector spine */}
          <svg style={{ position: "absolute", top: -36, left: "50%", width: 800, height: 36, transform: "translateX(-50%)", pointerEvents: "none" }}>
            <line x1="400" y1="0" x2="400" y2="20" stroke={T.rule} strokeWidth="1" />
            <line x1="100" y1="20" x2="700" y2="20" stroke={T.rule} strokeWidth="1" />
            <line x1="100" y1="20" x2="100" y2="36" stroke={T.rule} strokeWidth="1" />
            <line x1="400" y1="20" x2="400" y2="36" stroke={T.rule} strokeWidth="1" />
            <line x1="700" y1="20" x2="700" y2="36" stroke={T.rule} strokeWidth="1" />
          </svg>

          <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 20 }}>
            {branches.map((b, i) => (
              <div key={i} style={{
                background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2,
                borderTop: `3px solid ${b.color}`, padding: 22, display: "flex", flexDirection: "column",
              }}>
                <Row style={{ justifyContent: "space-between", marginBottom: 12 }}>
                  <div className="ei-label" style={{ color: b.color }}>
                    {String(i + 1).padStart(2, "0")} · {b.angle}
                  </div>
                  <Row gap={4}>
                    {[0.25, 0.5, 0.75, 1].map((threshold) => (
                      <div key={threshold} style={{
                        width: 12, height: 4,
                        background: b.weight >= threshold ? b.color : T.rule,
                      }} />
                    ))}
                  </Row>
                </Row>

                <div style={{
                  fontSize: 11, color: T.ink3, marginBottom: 12, fontFamily: "var(--ei-mono)",
                  textTransform: "uppercase", letterSpacing: 0.5,
                }}>
                  {L.tries} → {b.tests}
                </div>

                <div className="ei-serif" style={{
                  fontSize: 17, color: T.ink, lineHeight: 1.4, marginBottom: 20,
                  fontStyle: "italic", textWrap: "pretty",
                }}>
                  "{b.question}"
                </div>

                <Stack gap={12} style={{ flex: 1, marginBottom: 18 }}>
                  <div>
                    <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{L.why}</div>
                    <div style={{ fontSize: 12.5, color: T.ink2, lineHeight: 1.55, textWrap: "pretty" }}>{b.why}</div>
                  </div>
                  <div>
                    <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{L.risk}</div>
                    <div style={{ fontSize: 12.5, color: T.ink2, lineHeight: 1.55, textWrap: "pretty" }}>{b.risk}</div>
                  </div>
                </Stack>

                <Row style={{ justifyContent: "space-between", paddingTop: 12, borderTop: `1px dotted ${T.rule}` }}>
                  <span style={{ fontSize: 11, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
                    P = {Math.round(b.weight * 100)}%
                  </span>
                  <button onClick={() => nav("practice", { jobId: "tj-1", mode: "drill" })} style={{
                    background: "transparent", border: "none", color: b.color, fontSize: 12.5,
                    cursor: "pointer", fontFamily: "var(--ei-sans)", padding: 0,
                  }}>
                    {L.drill} →
                  </button>
                </Row>
              </div>
            ))}
          </div>
        </div>

        {/* Bottom CTA */}
        <Row style={{ marginTop: 36, justifyContent: "center", flexDirection: "column", gap: 12 }}>
          <Btn T={T} variant="primary" onClick={() => nav("practice", { jobId: "tj-1", mode: "drill" })} icon="play">
            {L.cta}
          </Btn>
          <span style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
            {lang === "en" ? "≈ 14 min · all three branches in one sitting" : "≈ 14 分钟 · 三条分支一气练完"}
          </span>
        </Row>
      </div>
    );
  };

  window.FollowUpTreeScreen = FollowUpTreeScreen;

  // ────────────────────────────────────────────────────────────
  // 6 · STAR EDITOR
  //    Four-column editor (Situation / Task / Action / Result) + AI rewrite suggestions per cell.
  // ────────────────────────────────────────────────────────────

  const STAR_INITIAL_ZH = {
    headline: "性能优化：把 B 端订单系统的 LCP 从 4.2s 压到 1.6s",
    s: "去年我们的 B 端订单管理系统，运营每天要在上面看几百个订单详情，单页 LCP 长期在 4–5 秒，运营投诉频次明显上升，电话客服那边接到不少抱怨。",
    t: "团队希望我作为前端 owner，在 6 周内把核心两个页面的 LCP 拉回到 2 秒以内，同时不能牺牲数据完整度（运营拒绝任何「分页加载」型妥协）。",
    a: "我做了三件事：第一，把首屏数据接口拆分成 critical + 后续，critical 走 SSR；第二，把图表换成 viewport-aware 的 lazy + canvas 实现；第三，把全表搜索从前端转移到服务端，并加了 1 秒输入防抖。",
    r: "六周后两个核心页面的 LCP 中位数降到 1.6 秒，P95 落到 2.4 秒；运营平均每个订单的处理时间从 2 分 18 秒降到 1 分 32 秒，客服侧关于「系统卡」的工单量下降了 40%。",
  };
  const STAR_INITIAL_EN = {
    headline: "Performance: brought the order admin LCP from 4.2s → 1.6s",
    s: "Our B2B order admin app was the daily workhorse for ~50 ops staff. Page LCP sat at 4–5s; complaints to support were spiking week over week.",
    t: "I owned the frontend. Goal: bring the two core pages under 2s LCP within six weeks, without giving up data completeness (ops vetoed any pagination-style compromise).",
    a: "Three moves: split the initial fetch into critical + follow-up (critical SSR'd); replaced heavy charts with viewport-aware lazy + canvas; moved full-table search server-side with 1s debounce.",
    r: "Median LCP fell to 1.6s, P95 to 2.4s. Per-order handling time fell 2'18\" → 1'32\". \"System feels slow\" tickets dropped 40%.",
  };

  const STAR_TIPS_ZH = {
    s: { hint: "舞台 / 上下文。1–2 句把规模、时间、为什么重要交代清楚。",
         issues: ["运营痛点没有量化（「投诉很多」不算）", "缺时间线 —— 是去年 Q3 还是哪个迭代？"] },
    t: { hint: "你 *被* 要求做什么 / 你 *主动* 接下了什么。要写得能看到约束。",
         issues: ["缺约束条件 —— 团队规模？工期？"] },
    a: { hint: "**你** 做了什么。不是「我们」，是「我」。三个动作以内最佳。",
         issues: ["第一点是技术决策但没说 *为什么* 这么决策"] },
    r: { hint: "结尾 = 数字 + 影响范围。Manager 听这一段。",
         issues: ["「下降 40%」很好；但口径要交代清楚（按周还是按月？）"] },
  };
  const STAR_TIPS_EN = {
    s: { hint: "Stage. One or two lines on scale, time, why it matters.",
         issues: ["Pain isn't quantified (\"complaints up\" isn't a number)", "Timeline missing — Q3 last year? Which sprint?"] },
    t: { hint: "What you were *asked* / what you *picked up*. Show the constraints.",
         issues: ["Constraints missing — team size? timeline?"] },
    a: { hint: "**You**, not we. Three actions, max.",
         issues: ["Action 1 is a technical call but doesn't say *why* you chose it"] },
    r: { hint: "End = numbers + scope. This is what manager rounds remember.",
         issues: ["\"40% drop\" is good — but specify: weekly basis? monthly?"] },
  };

  const STAR_REWRITES_ZH = {
    s: "去年 Q3，我们的 B 端订单管理系统每天承载约 50 名运营 700+ 单的处理。核心两个页面 LCP 长期 4.2 秒，过去一个月的客服工单里有 18% 提到「系统慢」。",
    a: "我做了三件事，按重要性：(1) 把首屏接口拆成 critical / 后续，critical 走 SSR —— 因为 60% 的数据其实运营首屏看不到；(2) 替换图表为 viewport-aware lazy + canvas，把首屏 JS 从 1.2 MB 降到 380 KB；(3) 把全表搜索推到服务端，加 1 秒防抖，减少不必要的 query。",
  };
  const STAR_REWRITES_EN = {
    s: "Q3 last year. ~50 ops users, 700+ orders/day. Two core pages sat at 4.2s LCP. 18% of last month's support tickets mentioned \"system feels slow.\"",
    a: "Three moves, in priority order: (1) split the initial fetch into critical + follow-up — critical went SSR — because 60% of payload was below the fold; (2) viewport-aware lazy + canvas charts dropped initial JS from 1.2 MB to 380 KB; (3) full-table search moved server-side with 1s debounce, killing redundant queries.",
  };

  const StarEditorScreen = ({ T, lang, nav }) => {
    const initial = lang === "en" ? STAR_INITIAL_EN : STAR_INITIAL_ZH;
    const tips = lang === "en" ? STAR_TIPS_EN : STAR_TIPS_ZH;
    const rewrites = lang === "en" ? STAR_REWRITES_EN : STAR_REWRITES_ZH;
    const [data, setData] = React.useState(initial);
    const [active, setActive] = React.useState("a");
    const [showRewrite, setShowRewrite] = React.useState({ s: false, t: false, a: false, r: false });
    const [accepted, setAccepted] = React.useState({ s: false, t: false, a: false, r: false });

    const cells = lang === "en"
      ? [
          { k: "s", letter: "S", title: "Situation", color: T.cool },
          { k: "t", letter: "T", title: "Task", color: T.amber },
          { k: "a", letter: "A", title: "Action", color: T.accent },
          { k: "r", letter: "R", title: "Result", color: T.ok },
        ]
      : [
          { k: "s", letter: "S", title: "情境", color: T.cool },
          { k: "t", letter: "T", title: "任务", color: T.amber },
          { k: "a", letter: "A", title: "行动", color: T.accent },
          { k: "r", letter: "R", title: "结果", color: T.ok },
        ];

    const wordCount = (s) => (lang === "en" ? s.trim().split(/\s+/).filter(Boolean).length : s.trim().length);
    const totalWc = wordCount(data.s) + wordCount(data.t) + wordCount(data.a) + wordCount(data.r);

    const L = lang === "en" ? {
      tag: "STAR EDITOR",
      h1: "Restructure your story before you tell it.",
      sub: "Map your answer onto Situation / Task / Action / Result. Each cell gets feedback in place — and one click swaps in an AI rewrite that keeps your voice.",
      header: "Story headline",
      headerPh: "One line — what's this story about?",
      issues: "Issues we found",
      rewrite: "AI rewrite",
      acceptRewrite: "Accept", undoRewrite: "Revert",
      saveStory: "Save to Stories library",
      attached: "Attached to: 'Performance' · q2",
      wc: "words",
      noRewrite: "Looks tight — no rewrite suggested.",
      delivery: "Practice this story",
    } : {
      tag: "STAR 重构",
      h1: "先把故事重新组装，再讲出来。",
      sub: "把答案铺到 情境 / 任务 / 行动 / 结果 四格里。每格就地反馈 —— 一键替换为保留你说话方式的 AI 改写。",
      header: "故事标题",
      headerPh: "一句话 —— 这个故事讲的是什么？",
      issues: "我们看到的问题",
      rewrite: "AI 改写",
      acceptRewrite: "采纳", undoRewrite: "撤回",
      saveStory: "存入经历库",
      attached: "关联：「性能优化」· q2",
      wc: "字",
      noRewrite: "已经够紧凑 —— 不需要改写。",
      delivery: "用这个故事练一次",
    };

    return (
      <div style={{ maxWidth: 1320, margin: "0 auto", padding: "40px 56px 80px" }}>
        <Row style={{ justifyContent: "space-between", marginBottom: 12 }}>
          <div className="ei-label" style={{ color: T.ink3 }}>{L.tag}</div>
          <span className="ei-label" style={{ color: T.ink3 }}>{L.attached}</span>
        </Row>

        <h1 className="ei-serif" style={{ fontSize: 36, color: T.ink, margin: 0, letterSpacing: "-0.02em", lineHeight: 1.15, marginBottom: 12 }}>
          {L.h1}
        </h1>
        <p style={{ fontSize: 14.5, color: T.ink2, lineHeight: 1.6, maxWidth: 760, marginBottom: 24, fontFamily: "var(--ei-serif)", textWrap: "pretty" }}>
          {L.sub}
        </p>

        {/* Headline + meta */}
        <div style={{
          marginBottom: 28, padding: 18, background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 2,
        }}>
          <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{L.header}</div>
          <input
            value={data.headline}
            onChange={(e) => setData({ ...data, headline: e.target.value })}
            placeholder={L.headerPh}
            style={{
              width: "100%", border: "none", outline: "none", background: "transparent",
              fontFamily: "var(--ei-serif)", fontSize: 22, color: T.ink, letterSpacing: "-0.01em",
              padding: 0, lineHeight: 1.3,
            }}
          />
          <Row style={{ marginTop: 12, justifyContent: "space-between", paddingTop: 10, borderTop: `1px dotted ${T.rule}` }}>
            <Row gap={14} style={{ fontSize: 11.5, color: T.ink3, fontFamily: "var(--ei-mono)" }}>
              <span>≈ {totalWc} {L.wc}</span>
              <span>· ≈ {Math.max(45, Math.round(totalWc / (lang === "en" ? 130 : 220) * 60))}s {lang === "en" ? "spoken" : "口述"}</span>
              <span>· {Object.values(accepted).filter(Boolean).length}/4 {lang === "en" ? "rewrites in" : "处改写已采纳"}</span>
            </Row>
            <Row gap={8}>
              <Btn T={T} variant="ghost" size="sm" icon="book">{L.saveStory}</Btn>
              <Btn T={T} variant="primary" size="sm" icon="play" onClick={() => nav("practice", { jobId: "tj-1" })}>
                {L.delivery}
              </Btn>
            </Row>
          </Row>
        </div>

        {/* Four columns */}
        <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 14 }}>
          {cells.map((c) => {
            const isActive = active === c.k;
            const tip = tips[c.k];
            const hasRewrite = !!rewrites[c.k];
            const wc = wordCount(data[c.k]);
            return (
              <div key={c.k} onClick={() => setActive(c.k)} style={{
                background: T.bgCard, border: `1px solid ${isActive ? c.color : T.rule}`,
                borderRadius: 2, padding: 18, display: "flex", flexDirection: "column", minHeight: 460,
                boxShadow: isActive ? `0 0 0 3px ${c.color}1a` : "none",
                transition: "box-shadow .15s, border-color .15s",
              }}>
                {/* Cell header */}
                <Row style={{ marginBottom: 12, justifyContent: "space-between" }}>
                  <Row gap={10}>
                    <div style={{
                      width: 28, height: 28, borderRadius: 14, background: c.color, color: "#fff",
                      display: "flex", alignItems: "center", justifyContent: "center",
                      fontFamily: "var(--ei-serif)", fontSize: 14, fontWeight: 600,
                    }}>{c.letter}</div>
                    <div>
                      <div className="ei-serif" style={{ fontSize: 16, color: T.ink, lineHeight: 1, fontWeight: 500 }}>{c.title}</div>
                      <div style={{ fontSize: 10, color: T.ink3, fontFamily: "var(--ei-mono)", marginTop: 2 }}>{wc} {L.wc}</div>
                    </div>
                  </Row>
                  {accepted[c.k] && <Tag tone="ok" T={T}>{lang === "en" ? "rewritten" : "已改写"}</Tag>}
                </Row>

                <div style={{ fontSize: 11.5, color: T.ink3, marginBottom: 10, lineHeight: 1.5, fontStyle: "italic", fontFamily: "var(--ei-serif)" }}
                     dangerouslySetInnerHTML={{ __html: tip.hint.replace(/\*\*(.+?)\*\*/g, `<b style="color:${T.ink2}">$1</b>`) }} />

                <textarea
                  value={data[c.k]}
                  onChange={(e) => setData({ ...data, [c.k]: e.target.value })}
                  style={{
                    width: "100%", flex: 1, minHeight: 140, padding: 12, border: `1px solid ${T.rule}`, borderRadius: 2,
                    background: T.bg, color: T.ink, fontSize: 13.5, lineHeight: 1.6, fontFamily: "var(--ei-serif)",
                    outline: "none", resize: "none", boxSizing: "border-box", marginBottom: 12,
                  }}
                />

                {/* Issues */}
                <div style={{
                  padding: 10, background: T.warnSoft, borderRadius: 2, marginBottom: 10,
                  borderLeft: `2px solid ${T.warn}`,
                }}>
                  <div className="ei-label" style={{ color: T.warn, marginBottom: 6 }}>
                    <Icon name="info" size={10} /> {L.issues}
                  </div>
                  {tip.issues.length === 0 ? (
                    <div style={{ fontSize: 11.5, color: T.ink2 }}>{lang === "en" ? "Nothing flagged." : "没问题。"}</div>
                  ) : (
                    <ul style={{ margin: 0, padding: 0, listStyle: "none" }}>
                      {tip.issues.map((iss, i) => (
                        <li key={i} style={{ fontSize: 11.5, color: T.ink2, lineHeight: 1.5, marginBottom: 4, paddingLeft: 10, position: "relative" }}>
                          <span style={{ position: "absolute", left: 0, color: T.warn }}>·</span>
                          {iss}
                        </li>
                      ))}
                    </ul>
                  )}
                </div>

                {/* AI rewrite block */}
                {hasRewrite ? (
                  showRewrite[c.k] ? (
                    <div style={{ padding: 10, background: T.bgSoft, borderRadius: 2, borderLeft: `2px solid ${c.color}` }}>
                      <Row style={{ justifyContent: "space-between", marginBottom: 6 }}>
                        <span className="ei-label" style={{ color: c.color }}>{L.rewrite}</span>
                        <TrustChip T={T} confidence={0.82} lang={lang} evidenceCount={2} />
                      </Row>
                      <div style={{ fontSize: 12, color: T.ink, lineHeight: 1.5, marginBottom: 10, fontFamily: "var(--ei-serif)", fontStyle: "italic" }}>
                        "{rewrites[c.k]}"
                      </div>
                      <Row gap={6}>
                        <button onClick={(e) => {
                          e.stopPropagation();
                          setData({ ...data, [c.k]: rewrites[c.k] });
                          setAccepted({ ...accepted, [c.k]: true });
                          setShowRewrite({ ...showRewrite, [c.k]: false });
                        }} style={{
                          flex: 1, padding: "5px 10px", background: c.color, color: "#fff", border: "none",
                          borderRadius: 2, fontSize: 11.5, cursor: "pointer", fontFamily: "var(--ei-sans)",
                        }}>
                          <Icon name="check" size={10} /> {L.acceptRewrite}
                        </button>
                        <button onClick={(e) => {
                          e.stopPropagation();
                          setShowRewrite({ ...showRewrite, [c.k]: false });
                        }} style={{
                          padding: "5px 10px", background: "transparent", color: T.ink3, border: `1px solid ${T.rule}`,
                          borderRadius: 2, fontSize: 11.5, cursor: "pointer", fontFamily: "var(--ei-sans)",
                        }}>
                          {L.undoRewrite}
                        </button>
                      </Row>
                    </div>
                  ) : (
                    <button onClick={(e) => { e.stopPropagation(); setShowRewrite({ ...showRewrite, [c.k]: true }); }} style={{
                      padding: "8px 10px", background: "transparent", border: `1px dashed ${c.color}`, borderRadius: 2,
                      color: c.color, fontSize: 12, cursor: "pointer", fontFamily: "var(--ei-sans)",
                      display: "flex", alignItems: "center", justifyContent: "center", gap: 6,
                    }}>
                      <Icon name="sparkle" size={12} /> {L.rewrite}
                    </button>
                  )
                ) : (
                  <div style={{ padding: "8px 10px", fontSize: 11.5, color: T.ink3, textAlign: "center", fontFamily: "var(--ei-mono)" }}>
                    {L.noRewrite}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    );
  };

  window.WelcomeScreen = WelcomeScreen;
  window.StarEditorScreen = StarEditorScreen;
})();
