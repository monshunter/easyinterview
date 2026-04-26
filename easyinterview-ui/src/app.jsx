// App shell — navigation, tweaks, top bar
const { useState, useEffect } = React;

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "theme": "warm",
  "dark": false,
  "serifFamily": "Noto Serif SC",
  "sansFamily": "Inter",
  "accentHue": 22,
  "reportLayout": "editorial",
  "role": "general"
}/*EDITMODE-END*/;

const App = () => {
  const [route, setRoute] = useState({ name: "welcome", params: {} });
  const [lang, setLang] = useState("zh");
  const [tweaks, setTweaks] = useState(TWEAK_DEFAULTS);
  const [tweaksOpen, setTweaksOpen] = useState(false);
  const [tweaksAvailable, setTweaksAvailable] = useState(false);
  const [signedIn, setSignedIn] = useState(() => {
    const v = localStorage.getItem("ei-signed-in");
    return v === "1";
  });

  // persistence
  useEffect(() => {
    // hash overrides (for iframes in design canvas)
    const hash = window.location.hash.slice(1);
    const params = new URLSearchParams(hash);
    if (params.get("route")) {
      setRoute({ name: params.get("route"), params: { jobId: params.get("jobId") || "tj-1", mode: params.get("mode") } });
    } else {
      const saved = localStorage.getItem("ei-route");
      if (saved) { try { setRoute(JSON.parse(saved)); } catch {} }
    }
    const savedLang = localStorage.getItem("ei-lang");
    if (savedLang) setLang(savedLang);
    // tweak overrides
    const overrides = {};
    ["dark","reportLayout","role","accentHue","serifFamily","sansFamily"].forEach((k) => {
      const v = params.get(k);
      if (v != null) overrides[k] = (k === "dark") ? (v === "1" || v === "true") : (k === "accentHue" ? Number(v) : v);
    });
    if (Object.keys(overrides).length) setTweaks((t) => ({ ...t, ...overrides }));
    if (params.get("lang")) setLang(params.get("lang"));
    if (params.get("nochrome") === "1") document.body.setAttribute("data-nochrome", "1");
  }, []);
  useEffect(() => { if (!window.location.hash) localStorage.setItem("ei-route", JSON.stringify(route)); }, [route]);
  useEffect(() => { localStorage.setItem("ei-lang", lang); }, [lang]);

  // Tweaks protocol
  useEffect(() => {
    const onMsg = (e) => {
      if (e.data?.type === "__activate_edit_mode") setTweaksOpen(true);
      if (e.data?.type === "__deactivate_edit_mode") setTweaksOpen(false);
    };
    window.addEventListener("message", onMsg);
    window.parent.postMessage({ type: "__edit_mode_available" }, "*");
    setTweaksAvailable(true);
    return () => window.removeEventListener("message", onMsg);
  }, []);

  const updateTweak = (k, v) => {
    const next = { ...tweaks, [k]: v };
    setTweaks(next);
    window.parent.postMessage({ type: "__edit_mode_set_keys", edits: { [k]: v } }, "*");
  };

  const T = React.useMemo(() => {
    const base = window.EI_THEMES[tweaks.dark ? "dark" : "warm"];
    // swap accent hue
    const hue = tweaks.accentHue;
    const accent = `oklch(0.62 0.14 ${hue})`;
    const accentSoft = tweaks.dark ? `oklch(0.28 0.06 ${hue})` : `oklch(0.92 0.05 ${hue})`;
    return { ...base, accent, accentSoft };
  }, [tweaks.dark, tweaks.accentHue]);

  // Apply font CSS vars
  useEffect(() => {
    document.documentElement.style.setProperty("--ei-serif", `"${tweaks.serifFamily}", Georgia, serif`);
    document.documentElement.style.setProperty("--ei-sans", `"${tweaks.sansFamily}", -apple-system, sans-serif`);
    document.body.style.background = T.bg;
    document.body.style.color = T.ink;
  }, [tweaks.serifFamily, tweaks.sansFamily, T.bg, T.ink]);

  const nav = (name, params = {}) => setRoute({ name, params });

  const screens = {
    home: <HomeScreen T={T} lang={lang} nav={nav} role={tweaks.role} />,
    workspace: <WorkspaceScreen T={T} lang={lang} nav={nav} jobId={route.params.jobId || "tj-1"} />,
    practice: <PracticeScreen T={T} lang={lang} nav={nav} jobId={route.params.jobId || "tj-1"} mode={route.params.mode} role={tweaks.role} setRole={(r) => updateTweak("role", r)} />,
    report: <ReportScreen T={T} lang={lang} nav={nav} reportLayout={tweaks.reportLayout} setReportLayout={(v) => updateTweak("reportLayout", v)} />,
    mistakes: <MistakesScreen T={T} lang={lang} nav={nav} />,
    debrief: <DebriefFullScreen T={T} lang={lang} nav={nav} />,
    resume: <ResumeScreen T={T} lang={lang} nav={nav} />,
    growth: <GrowthScreen T={T} lang={lang} nav={nav} />,
    voice: <VoicePracticeScreen T={T} lang={lang} nav={nav} />,
    plan: <PlanScreen T={T} lang={lang} nav={nav} />,
    parse: <ParseScreen T={T} lang={lang} nav={nav} />,
    onboarding: <OnboardingScreen T={T} lang={lang} nav={nav} />,
    generating: <ReportGeneratingScreen T={T} lang={lang} nav={nav} />,
    settings: <SettingsScreen T={T} lang={lang} nav={nav} />,
    debrief_full: <DebriefFullScreen T={T} lang={lang} nav={nav} />,
    experiences: <ExperienceLibraryScreen T={T} lang={lang} nav={nav} />,
    resume_versions: <ResumeVersionsScreen T={T} lang={lang} nav={nav} />,
    jd_match: <JDMatchScreen T={T} lang={lang} nav={nav} />,
    company_intel: <CompanyIntelScreen T={T} lang={lang} nav={nav} />,
    welcome: <WelcomeScreen T={T} lang={lang} nav={nav} onSignIn={() => { setSignedIn(true); localStorage.setItem("ei-signed-in", "1"); setRoute({ name: "home", params: {} }); }} />,
    drill: <DrillBuilderScreen T={T} lang={lang} nav={nav} />,
    followup: <FollowUpTreeScreen T={T} lang={lang} nav={nav} />,
    star: <StarEditorScreen T={T} lang={lang} nav={nav} />,
  };

  const hideTopBar = !signedIn || route.name === "practice" || route.name === "voice" || route.name === "onboarding" || route.name === "generating" || route.name === "welcome" || document.body.getAttribute("data-nochrome") === "1";

  const isCanvasIframe = document.body.getAttribute("data-nochrome") === "1" || window.location.hash.includes("nochrome=1");

  // Sync route when auth state flips (signed-in users shouldn't be parked on welcome)
  useEffect(() => {
    if (signedIn && route.name === "welcome" && !isCanvasIframe) {
      setRoute({ name: "home", params: {} });
    }
  }, [signedIn]);

  // Auth gate — when not signed in, force welcome (allow ?route=...nochrome=1 override for canvas iframes)
  let effectiveScreen;
  if (!signedIn && !isCanvasIframe) {
    effectiveScreen = screens.welcome;
  } else if (signedIn && route.name === "welcome" && !isCanvasIframe) {
    // Reverse gate: signed-in users shouldn't sit on welcome
    effectiveScreen = screens.home;
  } else {
    effectiveScreen = screens[route.name] || screens.home;
  }

  return (
    <div style={{ minHeight: "100vh", background: T.bg, color: T.ink, fontFamily: "var(--ei-sans)" }} data-screen-label={route.name}>
      {!hideTopBar && <TopBar T={T} route={route} nav={nav} lang={lang} setLang={setLang} dark={tweaks.dark} setDark={(v) => updateTweak("dark", v)} signOut={() => { setSignedIn(false); localStorage.removeItem("ei-signed-in"); setRoute({ name: "welcome", params: {} }); }} />}
      <div key={route.name + (route.params.jobId || "")}>
        {effectiveScreen}
      </div>
      {tweaksOpen && <TweaksPanel T={T} tweaks={tweaks} updateTweak={updateTweak} onClose={() => setTweaksOpen(false)} />}
    </div>
  );
};

const TopBar = ({ T, route, nav, lang, setLang, dark, setDark, signOut }) => {
  const [userMenuOpen, setUserMenuOpen] = React.useState(false);
  const nav_items = [
    { k: "home", labelZh: "收件箱", labelEn: "Inbox", icon: "target" },
    { k: "jd_match", labelZh: "JD 匹配", labelEn: "JD Match", icon: "search" },
    { k: "workspace", labelZh: "工作台", labelEn: "Workspace", icon: "briefcase" },
    { k: "experiences", labelZh: "经历库", labelEn: "Stories", icon: "book" },
    { k: "resume_versions", labelZh: "简历版本", labelEn: "Resumes", icon: "file" },
    { k: "mistakes", labelZh: "错题本", labelEn: "Mistakes", icon: "bookmark" },
    { k: "growth", labelZh: "成长", labelEn: "Growth", icon: "chart" },
    { k: "plan", labelZh: "多轮计划", labelEn: "Plan", icon: "layers" },
    { k: "debrief", labelZh: "复盘", labelEn: "Debrief", icon: "flag" },
  ];
  return (
    <div style={{
      borderBottom: `1px solid ${T.rule}`, background: T.bg, position: "sticky", top: 0, zIndex: 30,
      padding: "0 32px", height: 58, display: "flex", alignItems: "center", gap: 28,
    }}>
      <div onClick={() => nav("home")} style={{ display: "flex", alignItems: "center", gap: 10, cursor: "pointer" }}>
        <div style={{ width: 26, height: 26, borderRadius: 13, background: T.accent, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", fontFamily: "var(--ei-serif)", fontSize: 15, fontWeight: 600 }}>E</div>
        <div>
          <div className="ei-serif" style={{ fontSize: 16, color: T.ink, letterSpacing: "-0.01em", lineHeight: 1 }}>EasyInterview</div>
          <div className="ei-label" style={{ color: T.ink3, fontSize: 9, marginTop: 2 }}>面试训练器 · v1.0</div>
        </div>
      </div>

      <nav style={{ display: "flex", gap: 4, marginLeft: 20 }}>
        {nav_items.map((n) => (
          <button key={n.k} onClick={() => nav(n.k)} style={{
            background: route.name === n.k ? T.bgSoft : "transparent",
            border: "none", padding: "6px 12px", borderRadius: 2,
            color: route.name === n.k ? T.ink : T.ink3,
            fontSize: 13.5, fontWeight: route.name === n.k ? 500 : 400,
            display: "flex", gap: 6, alignItems: "center", cursor: "pointer", fontFamily: "var(--ei-sans)",
          }}>
            <Icon name={n.icon} size={13} /> {lang === "en" ? n.labelEn : n.labelZh}
          </button>
        ))}
      </nav>

      <div style={{ flex: 1 }} />

      <button onClick={() => setDark && setDark(!dark)} title={dark ? (lang === "en" ? "Switch to light" : "切换到浅色") : (lang === "en" ? "Switch to dark" : "切换到深色")} style={{
        background: "transparent", border: `1px solid ${T.rule}`, padding: "5px 8px", borderRadius: 2,
        color: T.ink2, fontSize: 12, display: "flex", gap: 6, alignItems: "center", cursor: "pointer",
      }}>
        <Icon name={dark ? "sun" : "moon"} size={12} />
      </button>

      <button onClick={() => setLang(lang === "zh" ? "en" : "zh")} style={{
        background: "transparent", border: `1px solid ${T.rule}`, padding: "5px 10px", borderRadius: 2,
        color: T.ink2, fontSize: 12, display: "flex", gap: 6, alignItems: "center", cursor: "pointer",
      }}>
        <Icon name="globe" size={12} /> {lang === "zh" ? "中 · EN" : "EN · 中"}
      </button>

      <div style={{ position: "relative" }}>
        <button onClick={() => setUserMenuOpen((o) => !o)} style={{
          display: "flex", alignItems: "center", gap: 8, background: "transparent", border: `1px solid ${T.rule}`,
          padding: "3px 10px 3px 3px", borderRadius: 18, cursor: "pointer",
        }}>
          <div style={{ width: 26, height: 26, borderRadius: 13, background: T.ink2, color: T.bg, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontWeight: 500, fontFamily: "var(--ei-mono)" }}>LZ</div>
          <div style={{ fontSize: 12, color: T.ink2 }}>{lang === "en" ? "Liu Zhe" : "刘哲"}</div>
          <span style={{ fontSize: 9, color: T.ink3, marginRight: 2 }}>▾</span>
        </button>
        {userMenuOpen && (
          <>
            <div onClick={() => setUserMenuOpen(false)} style={{ position: "fixed", inset: 0, zIndex: 39 }} />
            <div style={{
              position: "absolute", top: "calc(100% + 6px)", right: 0, minWidth: 220,
              background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3,
              boxShadow: "0 12px 36px rgba(20,15,10,0.16)", padding: 6, zIndex: 40,
            }}>
              <div style={{ padding: "10px 12px 8px", borderBottom: `1px solid ${T.rule}`, marginBottom: 6 }}>
                <div style={{ fontSize: 13, color: T.ink, fontWeight: 500 }}>{lang === "en" ? "Liu Zhe" : "刘哲"}</div>
                <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 2, fontFamily: "var(--ei-mono)" }}>liuzhe@example.com</div>
              </div>
              {[
                { k: "profile", icon: "user", labelZh: "个人资料 / 画像", labelEn: "Profile / preferences", action: () => nav("onboarding") },
                { k: "settings", icon: "settings", labelZh: "设置与隐私", labelEn: "Settings & privacy", action: () => nav("settings") },
              ].map((item) => (
                <button key={item.k} onClick={() => { item.action(); setUserMenuOpen(false); }} style={{
                  display: "flex", alignItems: "center", gap: 10, width: "100%",
                  background: "transparent", border: "none", textAlign: "left",
                  padding: "8px 12px", borderRadius: 2, cursor: "pointer", color: T.ink2, fontSize: 13,
                }}
                  onMouseEnter={(e) => e.currentTarget.style.background = T.bgSoft}
                  onMouseLeave={(e) => e.currentTarget.style.background = "transparent"}>
                  <Icon name={item.icon} size={13} style={{ color: T.ink3 }} />
                  {lang === "en" ? item.labelEn : item.labelZh}
                </button>
              ))}
              <div style={{ height: 1, background: T.rule, margin: "6px 0" }} />
              <button onClick={() => { setUserMenuOpen(false); signOut && signOut(); }} style={{
                display: "flex", alignItems: "center", gap: 10, width: "100%",
                background: "transparent", border: "none", textAlign: "left",
                padding: "8px 12px", borderRadius: 2, cursor: "pointer", color: T.danger, fontSize: 13,
              }}
                onMouseEnter={(e) => e.currentTarget.style.background = T.dangerSoft || T.bgSoft}
                onMouseLeave={(e) => e.currentTarget.style.background = "transparent"}>
                <Icon name="logout" size={13} />
                {lang === "en" ? "Sign out" : "退出登录"}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

const TweaksPanel = ({ T, tweaks, updateTweak, onClose }) => {
  return (
    <div style={{
      position: "fixed", right: 20, bottom: 20, width: 320, background: T.bgCard, border: `1px solid ${T.rule}`,
      borderRadius: 3, boxShadow: "0 8px 32px rgba(0,0,0,0.12)", padding: 18, zIndex: 200, fontFamily: "var(--ei-sans)",
    }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 14 }}>
        <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
          <Icon name="settings" size={14} color={T.ink2} />
          <div className="ei-serif" style={{ fontSize: 15, color: T.ink, fontWeight: 500 }}>Tweaks</div>
        </div>
        <button onClick={onClose} style={{ background: "transparent", border: "none", cursor: "pointer", color: T.ink3 }}>
          <Icon name="x" size={14} />
        </button>
      </div>

      <TweakRow T={T} label="Accent hue">
        <div style={{ display: "flex", gap: 6 }}>
          {[22, 45, 160, 220].map((h) => (
            <button key={h} onClick={() => updateTweak("accentHue", h)} style={{
              width: 22, height: 22, borderRadius: 11, border: tweaks.accentHue === h ? `2px solid ${T.ink}` : `1px solid ${T.rule}`,
              background: `oklch(0.62 0.14 ${h})`, cursor: "pointer", padding: 0,
            }} />
          ))}
        </div>
      </TweakRow>

      <TweakRow T={T} label="Serif font">
        <select value={tweaks.serifFamily} onChange={(e) => updateTweak("serifFamily", e.target.value)} style={selectStyle(T)}>
          <option>Noto Serif SC</option>
          <option>Source Serif Pro</option>
          <option>Georgia</option>
          <option>Cormorant Garamond</option>
        </select>
      </TweakRow>

      <TweakRow T={T} label="Sans font">
        <select value={tweaks.sansFamily} onChange={(e) => updateTweak("sansFamily", e.target.value)} style={selectStyle(T)}>
          <option>Inter</option>
          <option>Geist</option>
          <option>IBM Plex Sans</option>
          <option>PingFang SC</option>
        </select>
      </TweakRow>

      <TweakRow T={T} label="Interviewer role">
        <select value={tweaks.role} onChange={(e) => updateTweak("role", e.target.value)} style={selectStyle(T)}>
          <option value="general">综合面试官</option>
          <option value="hr">HR</option>
          <option value="manager">用人经理</option>
        </select>
      </TweakRow>

      <div style={{ marginTop: 10, fontSize: 11, color: T.ink3, lineHeight: 1.5, fontFamily: "var(--ei-mono)" }}>
        changes persist to disk via edit mode bridge.
      </div>
    </div>
  );
};

const selectStyle = (T) => ({
  padding: "4px 8px", fontSize: 12, borderRadius: 2, border: `1px solid ${T.rule}`,
  background: T.bg, color: T.ink, fontFamily: "var(--ei-sans)", minWidth: 130,
});

const TweakRow = ({ T, label, children }) => (
  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "10px 0", borderBottom: `1px dotted ${T.rule}` }}>
    <div style={{ fontSize: 12.5, color: T.ink2 }}>{label}</div>
    {children}
  </div>
);

window.App = App;
