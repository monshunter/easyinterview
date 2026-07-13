// App shell — navigation, tweaks, top bar
const { useState, useEffect } = React;

const TWEAK_DEFAULTS = {
  "theme": "ocean",
  "dark": false,
  "customAccent": null,
  "fontPreset": "editorial",
  "serifFamily": "Noto Serif SC",
  "sansFamily": "Inter"
};

// Seed values used when the user first switches to "Custom" — chosen to match
// each base theme's natural accent so the slider starts on the current colour.
const CUSTOM_ACCENT_SEEDS = {
  ocean:  { h: 255, c: 0.16 },
  plum:   { h: 340, c: 0.15 },
};

const LANGUAGE_OPTIONS = [
  { key: "zh", label: "中文", short: "中", aliases: ["zh", "zh-CN"] },
  { key: "en", label: "English", short: "EN", aliases: ["en", "en-US"] },
];
const DEFAULT_LANGUAGE = "en";

function normalizeLanguage(value) {
  const lower = String(value || "").trim().toLowerCase();
  if (!lower) return null;
  const match = LANGUAGE_OPTIONS.find((item) =>
    item.aliases.map((alias) => alias.toLowerCase()).includes(lower) ||
    item.key === lower ||
    lower.split("-")[0] === item.key
  );
  return match ? match.key : null;
}

function getInitialLanguage() {
  try {
    const saved = normalizeLanguage(localStorage.getItem("ei-lang"));
    if (saved) return saved;
  } catch {}
  const browserLanguages = [...(navigator.languages || []), navigator.language];
  for (const item of browserLanguages) {
    const normalized = normalizeLanguage(item);
    if (normalized) return normalized;
  }
  return DEFAULT_LANGUAGE;
}

const ROUTE_ALIASES = {
  welcome: "home",
  mistakes: "report",
  drill: "practice",
  followup: "practice",
  growth: "home",
  plan: "workspace",
  experiences: "resume_versions",
  star: "resume_versions",
  resume: "resume_versions",
  onboarding: "resume_versions",
  auth_register: "auth_login",
  auth_reset: "auth_login",
  jd_match: "home",
  debrief: "home",
  debrief_full: "home",
  profile: "home",
};

const DEFAULT_INTERVIEW_CONTEXT = {
  planId: "plan-tj-1",
  targetJobId: "tj-1",
  jobId: "tj-1",
  jdId: "jd-tj-1",
  resumeId: "frontend-v3",
  roundId: "round-2-technical",
  roundName: "技术一面 · 45m",
};

const INTERVIEW_CONTEXT_ROUTES = new Set(["practice"]);
const REPORT_LOCATOR_ROUTES = new Set(["generating", "report"]);
const normalizeRouteName = (name) => ROUTE_ALIASES[name] || name;
const shouldCarryInterviewContext = (name) => INTERVIEW_CONTEXT_ROUTES.has(normalizeRouteName(name));
const paramsFromSearch = (params) => {
  const out = {};
  params.forEach((value, key) => {
    if (value != null && value !== "") out[key] = value;
  });
  return out;
};
const stripUndefined = (obj) => Object.fromEntries(Object.entries(obj).filter(([, v]) => v !== undefined && v !== null));
const createInterviewContext = (params = {}, fallback = DEFAULT_INTERVIEW_CONTEXT) => {
  const targetJobId = params.targetJobId || params.jobId || fallback.targetJobId || fallback.jobId || DEFAULT_INTERVIEW_CONTEXT.targetJobId;
  const ctx = {
    ...DEFAULT_INTERVIEW_CONTEXT,
    ...fallback,
    ...params,
    targetJobId,
    jobId: targetJobId,
    planId: params.planId || fallback.planId || `plan-${targetJobId}`,
    jdId: params.jdId || fallback.jdId || `jd-${targetJobId}`,
    resumeId: params.resumeId || fallback.resumeId || DEFAULT_INTERVIEW_CONTEXT.resumeId,
    roundId: params.roundId || fallback.roundId || DEFAULT_INTERVIEW_CONTEXT.roundId,
    roundName: params.roundName || fallback.roundName || DEFAULT_INTERVIEW_CONTEXT.roundName,
  };
  return stripUndefined(ctx);
};
const paramsForRoute = (name, params = {}, fallback = DEFAULT_INTERVIEW_CONTEXT) => {
  const normalizedName = normalizeRouteName(name);
  if (REPORT_LOCATOR_ROUTES.has(normalizedName)) return stripUndefined({ reportId: params.reportId });
  if (shouldCarryInterviewContext(normalizedName)) return createInterviewContext(params, fallback);
  return stripUndefined(params);
};

const eiNormalizeInterviewRounds = (rawRounds = []) => {
  const allowedTypes = new Set(["hr", "technical", "manager", "cross_functional", "culture", "final", "other"]);
  const rounds = rawRounds
    .slice()
    .sort((a, b) => a.sequence - b.sequence);
  if (
    rounds.length < 2 ||
    rounds.length > 5 ||
    rounds.some((round, index) =>
      !Number.isInteger(round.sequence) ||
      round.sequence <= 0 ||
      (index > 0 && round.sequence <= rounds[index - 1].sequence) ||
      !allowedTypes.has(round.type) ||
      !String(round.name || "").trim() ||
      !String(round.focus || "").trim() ||
      !Number.isInteger(round.durationMinutes) ||
      round.durationMinutes < 10 ||
      round.durationMinutes > 180
    )
  ) return [];
  const normalized = rounds.map((round) => ({
    ...round,
    id: `round-${round.sequence}-${round.type}`,
    name: `${round.name || `R${round.sequence}`} · ${round.durationMinutes}m`,
  }));
  return new Set(normalized.map((round) => round.id)).size === normalized.length ? normalized : [];
};

const eiResolveInterviewRoundContext = (rawRounds = [], roundId) => {
  const rounds = eiNormalizeInterviewRounds(rawRounds);
  if (!roundId || new Set(rounds.map((round) => round.id)).size !== rounds.length) {
    return { currentRound: null, nextRound: null };
  }
  const currentIndex = rounds.findIndex((round) => round.id === roundId);
  if (currentIndex < 0) return { currentRound: null, nextRound: null };
  return {
    currentRound: rounds[currentIndex] || null,
    nextRound: rounds[currentIndex + 1] || null,
  };
};

const eiResolvePracticeProgress = (rawRounds = [], progress) => {
  const rounds = eiNormalizeInterviewRounds(rawRounds);
  const ids = rounds.map((round) => round.id);
  const sequences = rounds.map((round) => round.sequence);
  const invalid = { valid: false, rounds, completedCount: 0, currentIndex: null, currentRound: null, completed: false };
  if (!rounds.length || new Set(ids).size !== rounds.length || new Set(sequences).size !== rounds.length || !progress) return invalid;
  const completedRounds = Array.isArray(progress.completedRounds) ? progress.completedRounds : [];
  if (completedRounds.length > rounds.length) return invalid;
  const exact = (ref, round) => !!ref && !!round && ref.roundId === round.id && ref.roundSequence === round.sequence;
  if (completedRounds.some((ref, index) => !exact(ref, rounds[index]))) return invalid;
  if (completedRounds.length === rounds.length) {
    if (progress.status !== "completed" || progress.currentRound !== null) return invalid;
    return { valid: true, rounds, completedCount: rounds.length, currentIndex: rounds.length, currentRound: null, completed: true };
  }
  const expectedStatus = completedRounds.length === 0 ? "not_started" : "in_progress";
  const currentRound = rounds[completedRounds.length];
  if (progress.status !== expectedStatus || !exact(progress.currentRound, currentRound)) return invalid;
  return { valid: true, rounds, completedCount: completedRounds.length, currentIndex: completedRounds.length, currentRound, completed: false };
};

Object.assign(window, {
  EI_DEFAULT_INTERVIEW_CONTEXT: DEFAULT_INTERVIEW_CONTEXT,
  eiCreateInterviewContext: createInterviewContext,
  eiResolveInterviewRoundContext,
  eiResolvePracticeProgress,
});

const App = () => {
  const [route, setRoute] = useState({ name: "home", params: {} });
  const [lang, setLang] = useState(getInitialLanguage);
  const [tweaks, setTweaks] = useState(TWEAK_DEFAULTS);
  const [signedIn, setSignedIn] = useState(false);
  const [profileComplete, setProfileComplete] = useState(false);
  const normalizeRoute = normalizeRouteName;

  // persistence
  useEffect(() => {
    // hash overrides (for iframes in design canvas)
    const hash = window.location.hash.slice(1);
    const params = new URLSearchParams(hash);
    if (params.get("route")) {
      const rawRoute = params.get("route");
      const name = normalizeRoute(rawRoute);
      const parsedParams = paramsFromSearch(params);
      delete parsedParams.route;
      delete parsedParams.lang;
      delete parsedParams.nochrome;
      const nextParams = paramsForRoute(name, parsedParams);
      setRoute({ name, params: nextParams });
    } else if (hash && !hash.includes("=")) {
      const rawRoute = hash;
      const name = normalizeRoute(rawRoute);
      const parsedParams = {};
      setRoute({ name, params: paramsForRoute(name, parsedParams) });
    }
    const savedLang = normalizeLanguage(localStorage.getItem("ei-lang"));
    if (savedLang) setLang(savedLang);
    // tweak overrides
    const overrides = {};
    ["dark","theme","serifFamily","sansFamily","fontPreset"].forEach((k) => {
      const v = params.get(k);
      if (v != null) overrides[k] = (k === "dark") ? (v === "1" || v === "true") : v;
    });
    const ca = params.get("customAccent");
    if (ca != null) {
      if (ca === "" || ca === "null") {
        overrides.customAccent = null;
      } else {
        const [h, c] = ca.split(",").map(Number);
        if (Number.isFinite(h) && Number.isFinite(c)) overrides.customAccent = { h, c };
      }
    }
    if (Object.keys(overrides).length) setTweaks((t) => ({ ...t, ...overrides }));
    const hashLang = normalizeLanguage(params.get("lang"));
    if (hashLang) setLang(hashLang);
    if (params.get("signedIn") === "1") setSignedIn(true);
    if (params.get("nochrome") === "1") document.body.setAttribute("data-nochrome", "1");
  }, []);
  useEffect(() => { localStorage.setItem("ei-lang", lang); }, [lang]);

  const updateTweak = (k, v) => setTweaks((prev) => ({ ...prev, [k]: v }));

  // Apply a font preset atomically (preset key + serif + sans in one update,
  // so the UI doesn't flash with a half-applied pair).
  const setFontPreset = (key) => {
    const p = (window.EI_FONT_PRESETS || []).find((x) => x.key === key);
    if (!p) return;
    const edits = { fontPreset: key, serifFamily: p.serif, sansFamily: p.sans };
    setTweaks((prev) => ({ ...prev, ...edits }));
  };

  const T = React.useMemo(() => {
    const themeKey = window.EI_THEMES[tweaks.theme] ? tweaks.theme : "ocean";
    const isDark = !!tweaks.dark;
    const base = { ...window.EI_THEMES[themeKey][isDark ? "dark" : "light"] };
    const ca = tweaks.customAccent;
    if (ca && typeof ca.h === "number" && typeof ca.c === "number") {
      const h = ((ca.h % 360) + 360) % 360;
      const c = Math.max(0, Math.min(0.28, ca.c));
      const accentL = isDark ? 68 : 58;
      const softL = isDark ? 28 : 92;
      const softC = isDark ? Math.min(c * 0.55, 0.10) : Math.min(c * 0.22, 0.05);
      base.accent = `oklch(${accentL}% ${c.toFixed(3)} ${h.toFixed(1)})`;
      base.accentSoft = `oklch(${softL}% ${softC.toFixed(3)} ${h.toFixed(1)})`;
    }
    return base;
  }, [tweaks.dark, tweaks.theme, tweaks.customAccent]);

  // Apply font CSS vars
  useEffect(() => {
    document.documentElement.style.setProperty("--ei-serif", `"${tweaks.serifFamily}", Georgia, serif`);
    document.documentElement.style.setProperty("--ei-sans", `"${tweaks.sansFamily}", -apple-system, sans-serif`);
    document.body.style.background = T.bg;
    document.body.style.color = T.ink;
  }, [tweaks.serifFamily, tweaks.sansFamily, T.bg, T.ink]);

  const activeRouteName = normalizeRoute(route.name);
  const currentContext = React.useMemo(() => createInterviewContext(route.params || {}), [route.params]);
  const nav = (name, params = {}) => {
    const nextName = normalizeRoute(name);
    const nextParams = paramsForRoute(nextName, params, currentContext);
    setRoute({ name: nextName, params: nextParams });
  };
  const requestAuth = (pendingAction, run) => {
    if (signedIn) {
      run();
      return;
    }
    setRoute({ name: "auth_login", params: { pendingAction } });
  };
  const restoreAfterAuth = (pendingAction) => {
    if (pendingAction?.route) {
      const pendingParams = pendingAction.params || {};
      const pendingRoute = normalizeRoute(pendingAction.route);
      setRoute({
        name: pendingRoute,
        params: paramsForRoute(pendingRoute, pendingParams, currentContext),
      });
      window.eiToast && window.eiToast(
        lang === "en" ? `Continuing: ${pendingAction.label || "pending action"}` : `继续：${pendingAction.label || "刚才的操作"}`,
        { tone: "ok", duration: 2400 }
      );
      return;
    }
    setRoute({ name: "home", params: stripUndefined({}) });
  };
  const completeSignIn = () => {
    setSignedIn(true);
    const pendingAction = route.params?.pendingAction;
    if (!profileComplete) {
      setRoute({ name: "auth_profile_setup", params: { pendingAction } });
      return;
    }
    restoreAfterAuth(pendingAction);
  };
  const completeProfile = (name) => {
    setProfileComplete(true);
    window.eiToast && window.eiToast(
      lang === "en" ? `Profile ready: ${name || "Candidate"}` : `资料已完成：${name || "候选人"}`,
      { tone: "ok", duration: 2400 }
    );
    restoreAfterAuth(route.params?.pendingAction);
  };
  const completeSignOut = () => {
    setSignedIn(false);
  };

  const screens = {
    home: <HomeScreen T={T} lang={lang} nav={nav} signedIn={signedIn} />,
    workspace: <WorkspaceScreen T={T} lang={lang} nav={nav} />,
    practice: <PracticeScreen T={T} lang={lang} nav={nav} params={route.params || {}} jobId={currentContext.targetJobId} mode={route.params.mode} />,
    report: <ReportScreen T={T} lang={lang} nav={nav} params={route.params || {}} requestAuth={requestAuth} />,
    parse: <ParseScreen T={T} lang={lang} nav={nav} requestAuth={requestAuth} />,
    generating: <ReportGeneratingScreen T={T} lang={lang} nav={nav} params={route.params || {}} />,
    settings: <SettingsScreen T={T} lang={lang} fontPreset={tweaks.fontPreset} setFontPreset={setFontPreset} />,
    resume_versions: <ResumeVersionsScreen T={T} lang={lang} nav={nav} params={route.params || {}} />,
    auth_login: <AuthLoginScreen T={T} lang={lang} nav={nav} pendingAction={route.params.pendingAction} />,
    auth_verify: <AuthVerifyScreen T={T} lang={lang} nav={nav} email={route.params.email} onSignIn={completeSignIn} pendingAction={route.params.pendingAction} />,
    auth_profile_setup: <AuthProfileSetupScreen T={T} lang={lang} onCompleteProfile={completeProfile} pendingAction={route.params.pendingAction} />,
    auth_logout: <AuthLogoutScreen T={T} lang={lang} nav={nav} signedIn={signedIn} onSignOut={completeSignOut} />,
  };

  const hideTopBar = activeRouteName === "practice" || activeRouteName === "generating" || document.body.getAttribute("data-nochrome") === "1";

  const effectiveScreen = screens[activeRouteName] || screens.home;

  return (
    <div style={{ minHeight: "100vh", background: T.bg, color: T.ink, fontFamily: "var(--ei-sans)" }} data-screen-label={route.name}>
      {!hideTopBar && <TopBar T={T} route={{ ...route, name: activeRouteName }} nav={nav} lang={lang} setLang={setLang} dark={tweaks.dark} setDark={(v) => updateTweak("dark", v)} theme={tweaks.theme} setTheme={(v) => updateTweak("theme", v)} customAccent={tweaks.customAccent} setCustomAccent={(v) => updateTweak("customAccent", v)} signedIn={signedIn} signOut={() => nav("auth_logout")} />}
      <div key={route.name + (route.params.jobId || "")}>
        {effectiveScreen}
      </div>
    </div>
  );
};

const TopBar = ({ T, route, nav, lang, setLang, dark, setDark, theme, setTheme, customAccent, setCustomAccent, signedIn, signOut }) => {
  const [userMenuOpen, setUserMenuOpen] = React.useState(false);
  const [themeMenuOpen, setThemeMenuOpen] = React.useState(false);
  const [langMenuOpen, setLangMenuOpen] = React.useState(false);
  const [pickerOpen, setPickerOpen] = React.useState(!!customAccent);
  const customActive = !!customAccent;
  const currentLanguage = LANGUAGE_OPTIONS.find((item) => item.key === lang) || LANGUAGE_OPTIONS.find((item) => item.key === DEFAULT_LANGUAGE) || LANGUAGE_OPTIONS[0];
  React.useEffect(() => {
    if (!userMenuOpen && !themeMenuOpen && !langMenuOpen) return;
    const onKey = (e) => {
      if (e.key === "Escape") {
        setUserMenuOpen(false);
        setThemeMenuOpen(false);
        setLangMenuOpen(false);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [userMenuOpen, themeMenuOpen, langMenuOpen]);
  const RAINBOW = "conic-gradient(from 0deg, oklch(60% 0.2 0), oklch(60% 0.2 60), oklch(60% 0.2 120), oklch(60% 0.2 180), oklch(60% 0.2 240), oklch(60% 0.2 300), oklch(60% 0.2 360))";
  const nav_items = [
    { k: "home", labelZh: "首页", labelEn: "Home", icon: "target" },
    { k: "workspace", labelZh: "面试", labelEn: "Interview", icon: "play" },
    { k: "resume_versions", labelZh: "简历", labelEn: "Resume", icon: "file" },
  ];
  return (
    <div className="ei-shell-topbar" style={{
      borderBottom: `1px solid ${T.rule}`, background: T.bg, position: "sticky", top: 0, zIndex: 30,
      padding: "0 32px", height: 58, display: "flex", alignItems: "center", gap: 28,
    }}>
      <div className="ei-topbar-brand" onClick={() => nav("home")} style={{ display: "flex", alignItems: "center", gap: 10, cursor: "pointer" }}>
        <div style={{ width: 26, height: 26, borderRadius: 13, background: T.accent, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", fontFamily: "var(--ei-serif)", fontSize: 15, fontWeight: 600 }}>E</div>
        <div className="ei-topbar-brand-copy">
          <div className="ei-serif" style={{ fontSize: 16, color: T.ink, letterSpacing: "-0.01em", lineHeight: 1 }}>EasyInterview</div>
        </div>
      </div>

      <nav className="ei-topbar-nav" style={{ display: "flex", gap: 4, marginLeft: 20 }}>
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

      <div className="ei-topbar-spacer" style={{ flex: 1 }} />

      <div className="ei-topbar-controls" style={{ display: "flex", alignItems: "center", gap: 8 }}>
      <div style={{ position: "relative" }}>
        <button onClick={() => setThemeMenuOpen((o) => !o)} title={lang === "en" ? "Theme" : "主题色"} style={{
          background: "transparent", border: `1px solid ${T.rule}`, padding: "4px 8px", borderRadius: 2,
          color: T.ink2, fontSize: 12, display: "flex", gap: 6, alignItems: "center", cursor: "pointer",
        }}>
          <span style={{
            display: "inline-block", width: 12, height: 12, borderRadius: 6,
            background: T.accent, border: `1px solid ${T.rule}`,
          }} />
          <span style={{ fontSize: 9, color: T.ink3 }}>▾</span>
        </button>
        {themeMenuOpen && (
          <>
            <div onClick={() => setThemeMenuOpen(false)} style={{ position: "fixed", inset: 0, zIndex: 39 }} />
            <div style={{
              position: "absolute", top: "calc(100% + 6px)", right: 0, width: 240,
              background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3,
              boxShadow: "0 12px 36px rgba(20,15,10,0.16)", padding: 6, zIndex: 40,
            }}>
              <div className="ei-label" style={{ padding: "8px 10px 6px", color: T.ink3 }}>
                {lang === "en" ? "Theme" : "主题色"}
              </div>
              {(window.EI_THEME_LIST || []).map((t) => {
                const selected = theme === t.key && !customActive;
                return (
                  <button key={t.key} onClick={() => {
                    setTheme && setTheme(t.key);
                    setCustomAccent && setCustomAccent(null);
                    setPickerOpen(false);
                    setThemeMenuOpen(false);
                  }} style={{
                    display: "flex", alignItems: "center", gap: 10, width: "100%",
                    background: selected ? T.bgSoft : "transparent",
                    border: "none", textAlign: "left",
                    padding: "8px 10px", borderRadius: 2, cursor: "pointer", color: T.ink2, fontSize: 13,
                  }}
                    onMouseEnter={(e) => { if (!selected) e.currentTarget.style.background = T.bgSoft; }}
                    onMouseLeave={(e) => { if (!selected) e.currentTarget.style.background = "transparent"; }}>
                    <span style={{
                      display: "inline-block", width: 14, height: 14, borderRadius: 7,
                      background: t.swatch, border: `1px solid ${T.rule}`,
                    }} />
                    <span style={{ flex: 1 }}>{lang === "en" ? t.labelEn : t.labelZh}</span>
                    {selected && <Icon name="check" size={12} style={{ color: T.accent }} />}
                  </button>
                );
              })}

              <div style={{ height: 1, background: T.rule, margin: "4px 6px" }} />

              <button onClick={() => {
                if (!customActive) {
                  const seed = CUSTOM_ACCENT_SEEDS[theme] || CUSTOM_ACCENT_SEEDS.ocean;
                  setCustomAccent && setCustomAccent({ ...seed });
                  setPickerOpen(true);
                } else {
                  setPickerOpen((o) => !o);
                }
              }} style={{
                display: "flex", alignItems: "center", gap: 10, width: "100%",
                background: customActive ? T.bgSoft : "transparent",
                border: "none", textAlign: "left",
                padding: "8px 10px", borderRadius: 2, cursor: "pointer", color: T.ink2, fontSize: 13,
              }}
                onMouseEnter={(e) => { if (!customActive) e.currentTarget.style.background = T.bgSoft; }}
                onMouseLeave={(e) => { if (!customActive) e.currentTarget.style.background = "transparent"; }}>
                <span style={{
                  display: "inline-block", width: 14, height: 14, borderRadius: 7,
                  background: customActive ? T.accent : RAINBOW, border: `1px solid ${T.rule}`,
                }} />
                <span style={{ flex: 1 }}>{lang === "en" ? "Custom" : "自定义"}</span>
                {customActive
                  ? <Icon name="check" size={12} style={{ color: T.accent }} />
                  : <span style={{ fontSize: 9, color: T.ink3 }}>{pickerOpen ? "▴" : "▾"}</span>}
              </button>

              {pickerOpen && (
                <AccentPicker
                  T={T}
                  lang={lang}
                  dark={dark}
                  value={customAccent || CUSTOM_ACCENT_SEEDS[theme] || CUSTOM_ACCENT_SEEDS.ocean}
                  active={customActive}
                  onChange={(v) => setCustomAccent && setCustomAccent(v)}
                  onClear={() => { setCustomAccent && setCustomAccent(null); setPickerOpen(false); }}
                />
              )}
            </div>
          </>
        )}
      </div>

      <button className="ei-topbar-dark" onClick={() => setDark && setDark(!dark)} title={dark ? (lang === "en" ? "Switch to light" : "切换到浅色") : (lang === "en" ? "Switch to dark" : "切换到深色")} style={{
        background: "transparent", border: `1px solid ${T.rule}`, padding: "5px 8px", borderRadius: 2,
        color: T.ink2, fontSize: 12, display: "flex", gap: 6, alignItems: "center", cursor: "pointer",
      }}>
        <Icon name={dark ? "sun" : "moon"} size={12} />
      </button>

      <div style={{ position: "relative" }}>
        <button className="ei-topbar-lang" onClick={() => setLangMenuOpen((o) => !o)} aria-expanded={langMenuOpen} style={{
          background: "transparent", border: `1px solid ${T.rule}`, padding: "5px 10px", borderRadius: 2,
          color: T.ink2, fontSize: 12, display: "flex", gap: 6, alignItems: "center", cursor: "pointer",
        }}>
          <Icon name="globe" size={12} /> <span className="ei-topbar-lang-current">{currentLanguage.label}</span>
          <span style={{ fontSize: 9, color: T.ink3 }}>▾</span>
        </button>
        {langMenuOpen && (
          <>
            <div onClick={() => setLangMenuOpen(false)} style={{ position: "fixed", inset: 0, zIndex: 39 }} />
            <div style={{
              position: "absolute", top: "calc(100% + 6px)", right: 0, width: 148,
              background: T.bgCard, border: `1px solid ${T.rule}`, borderRadius: 3,
              boxShadow: "0 12px 36px rgba(20,15,10,0.16)", padding: 6, zIndex: 40,
            }}>
              <div className="ei-label" style={{ padding: "8px 10px 6px", color: T.ink3 }}>
                {lang === "en" ? "Language" : "界面语言"}
              </div>
              {LANGUAGE_OPTIONS.map((item) => {
                const selected = lang === item.key;
                return (
                  <button key={item.key} onClick={() => { setLang(item.key); setLangMenuOpen(false); }} style={{
                    display: "flex", alignItems: "center", gap: 10, width: "100%",
                    background: selected ? T.bgSoft : "transparent",
                    border: "none", textAlign: "left",
                    padding: "8px 10px", borderRadius: 2, cursor: "pointer", color: T.ink2, fontSize: 13,
                  }}
                    onMouseEnter={(e) => { if (!selected) e.currentTarget.style.background = T.bgSoft; }}
                    onMouseLeave={(e) => { if (!selected) e.currentTarget.style.background = "transparent"; }}>
                    <Icon name="globe" size={13} style={{ color: T.ink3 }} />
                    <span style={{ flex: 1 }}>{item.label}</span>
                    {selected ? <Icon name="check" size={12} style={{ color: T.accent }} /> : <span className="ei-label" style={{ color: T.ink3 }}>{item.short}</span>}
                  </button>
                );
              })}
            </div>
          </>
        )}
      </div>
      </div>

      <div className="ei-topbar-user" data-signed-in={signedIn} style={{ display: "flex", gap: 8, alignItems: "center" }}>
      {signedIn ? (
        <div style={{ position: "relative" }}>
          <button onClick={() => setUserMenuOpen((o) => !o)} style={{
            display: "flex", alignItems: "center", gap: 8, background: "transparent", border: `1px solid ${T.rule}`,
            padding: "3px 10px 3px 3px", borderRadius: 18, cursor: "pointer",
          }}>
            <div style={{ width: 26, height: 26, borderRadius: 13, background: T.ink2, color: T.bg, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontWeight: 500, fontFamily: "var(--ei-mono)" }}>AE</div>
            <div style={{ fontSize: 12, color: T.ink2 }}>Alice Example</div>
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
                  <div style={{ fontSize: 13, color: T.ink, fontWeight: 500 }}>Alice Example</div>
                  <div style={{ fontSize: 11.5, color: T.ink3, marginTop: 2, fontFamily: "var(--ei-mono)" }}>ali***@example.com</div>
                </div>
                {[
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
      ) : (
        <div>
          <Btn T={T} variant="ghost" size="sm" onClick={() => nav("auth_login")}>{lang === "en" ? "Sign in" : "登录"}</Btn>
        </div>
      )}
      </div>
    </div>
  );
};

const AccentPicker = ({ T, lang, dark, value, active, onChange, onClear }) => {
  const accentL = dark ? 68 : 58;
  const v = value || { h: 255, c: 0.16 };
  const previewAccent = `oklch(${accentL}% ${v.c} ${v.h})`;

  // Hue track: rainbow at constant L, mid chroma
  const hueStops = [];
  for (let i = 0; i <= 12; i++) {
    const h = (i / 12) * 360;
    hueStops.push(`oklch(${accentL}% 0.18 ${h})`);
  }
  const hueGradient = `linear-gradient(to right, ${hueStops.join(", ")})`;
  // Chroma track: 0 → 0.25 at current hue
  const chromaGradient = `linear-gradient(to right, oklch(${accentL}% 0 ${v.h}), oklch(${accentL}% 0.25 ${v.h}))`;

  const trackWrap = (gradient) => ({
    position: "relative", height: 16, borderRadius: 8,
    background: gradient, border: `1px solid ${T.rule}`,
    opacity: active ? 1 : 0.55,
  });
  const inputStyle = {
    position: "absolute", inset: 0, width: "100%", height: "100%",
    margin: 0, padding: 0, border: "none",
  };

  return (
    <div style={{
      padding: "10px 10px 12px", marginTop: 4,
      borderTop: `1px dotted ${T.rule}`,
      animation: "ei-fadein .18s ease-out",
    }}>
      <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 10 }}>
        <span style={{
          width: 26, height: 26, borderRadius: 13, background: previewAccent,
          border: `1px solid ${T.rule}`, flexShrink: 0,
          opacity: active ? 1 : 0.55,
        }} />
        <div style={{ flex: 1, fontSize: 10.5, color: T.ink3, fontFamily: "var(--ei-mono)", letterSpacing: "0.02em", lineHeight: 1.4 }}>
          oklch({accentL}% {Number(v.c).toFixed(3)} {Math.round(v.h)})
          {!active && <div style={{ color: T.ink3, marginTop: 2 }}>{lang === "en" ? "Drag to apply" : "拖动应用"}</div>}
        </div>
      </div>

      <div style={{ marginBottom: 8 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{lang === "en" ? "Hue" : "色相"}</div>
        <div style={trackWrap(hueGradient)}>
          <input className="ei-slider-overlay" type="range" min={0} max={360} step={1} value={v.h}
            onChange={(e) => onChange({ ...v, h: Number(e.target.value) })}
            style={inputStyle} />
        </div>
      </div>

      <div style={{ marginBottom: 10 }}>
        <div className="ei-label" style={{ color: T.ink3, marginBottom: 4 }}>{lang === "en" ? "Chroma" : "饱和度"}</div>
        <div style={trackWrap(chromaGradient)}>
          <input className="ei-slider-overlay" type="range" min={0} max={0.25} step={0.005} value={v.c}
            onChange={(e) => onChange({ ...v, c: Number(e.target.value) })}
            style={inputStyle} />
        </div>
      </div>

      {active && (
        <button onClick={onClear} className="ei-link" style={{
          background: "transparent", border: "none", padding: 0,
          fontSize: 11.5, color: T.ink3,
        }}>{lang === "en" ? "Reset to theme accent" : "恢复主题默认色"}</button>
      )}
    </div>
  );
};

window.App = App;
