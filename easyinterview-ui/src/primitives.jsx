// Shared primitives — design tokens, icons, small UI atoms
// EasyInterview · warm editorial aesthetic

// Theme palettes — orthogonal to dark/light mode.
// Structure: EI_THEMES[themeKey][mode]  where mode is "light" | "dark".
// Each theme is a coordinated bg/ink/rule/accent set; the dark/light toggle
// just flips the mode within the chosen theme.
window.EI_THEMES = {
  warm: {
    light: {
      bg: "#fdfcf8", bgSoft: "#f7f3ea", bgCard: "#ffffff",
      ink: "#1c1917", ink2: "#44403c", ink3: "#78716c", ink4: "#a8a29e",
      rule: "#e7e2d6", ruleSoft: "#efeadc",
      accent: "#c96442", accentSoft: "#fbe8dc",
      amber: "#d9893a", amberSoft: "#fbe9ce",
      ok: "#5a7a4a", okSoft: "#e7efd9",
      warn: "#b8813a", warnSoft: "#f6ead0",
      danger: "#a8452a", dangerSoft: "#f6dcd0",
      cool: "#4a6670", coolSoft: "#dce6ea",
    },
    dark: {
      bg: "#16130e", bgSoft: "#1f1b15", bgCard: "#1a1611",
      ink: "#f5f0e4", ink2: "#d6cdb8", ink3: "#968d7a", ink4: "#6b6455",
      rule: "#2d2820", ruleSoft: "#24201a",
      accent: "#e08061", accentSoft: "#3a2318",
      amber: "#e6a25a", amberSoft: "#3b2a16",
      ok: "#8fae7c", okSoft: "#24301a",
      warn: "#d9a868", warnSoft: "#362812",
      danger: "#d4694a", dangerSoft: "#3a1e14",
      cool: "#89a4ae", coolSoft: "#1c2830",
    },
  },
  forest: {
    light: {
      bg: "#f9faf3", bgSoft: "#eef2e3", bgCard: "#ffffff",
      ink: "#181d14", ink2: "#3c4434", ink3: "#6f7565", ink4: "#a3a895",
      rule: "#dde3ce", ruleSoft: "#e8ecdc",
      accent: "#5a7d3a", accentSoft: "#dfe9c9",
      amber: "#b8813a", amberSoft: "#f3e4c4",
      ok: "#5a7a4a", okSoft: "#dde7cd",
      warn: "#a87832", warnSoft: "#f0e0bf",
      danger: "#a8452a", dangerSoft: "#f3d9cd",
      cool: "#4a6670", coolSoft: "#d6e1e5",
    },
    dark: {
      bg: "#0e120a", bgSoft: "#161b10", bgCard: "#11160c",
      ink: "#eaeed8", ink2: "#cbcfb0", ink3: "#888c70", ink4: "#5e6250",
      rule: "#252a1c", ruleSoft: "#1c2014",
      accent: "#8fae60", accentSoft: "#1f2a14",
      amber: "#d9a868", amberSoft: "#352712",
      ok: "#9ab78a", okSoft: "#212c16",
      warn: "#cda06a", warnSoft: "#332512",
      danger: "#d4694a", dangerSoft: "#361d12",
      cool: "#89a4ae", coolSoft: "#1a242a",
    },
  },
  ocean: {
    light: {
      bg: "#f8fafd", bgSoft: "#eef2f7", bgCard: "#ffffff",
      ink: "#141821", ink2: "#363c4a", ink3: "#6b7280", ink4: "#a0a8b3",
      rule: "#dde2ec", ruleSoft: "#e7ebf2",
      accent: "#3a5fc4", accentSoft: "#dde6f7",
      amber: "#c98730", amberSoft: "#f3e1c0",
      ok: "#3f8367", okSoft: "#d4ebde",
      warn: "#a87832", warnSoft: "#f0e0bf",
      danger: "#b3402b", dangerSoft: "#f4d6cc",
      cool: "#4a6670", coolSoft: "#d6e1e5",
    },
    dark: {
      bg: "#0c0f17", bgSoft: "#13182a", bgCard: "#0f1320",
      ink: "#e8edf6", ink2: "#c4cad8", ink3: "#8389a0", ink4: "#5d627a",
      rule: "#212740", ruleSoft: "#181d2e",
      accent: "#7493d4", accentSoft: "#1c2540",
      amber: "#e6a25a", amberSoft: "#322411",
      ok: "#74b08c", okSoft: "#1a2c20",
      warn: "#d9a868", warnSoft: "#332512",
      danger: "#d4694a", dangerSoft: "#361d12",
      cool: "#89a4ae", coolSoft: "#1a242a",
    },
  },
  plum: {
    light: {
      bg: "#fcf8fa", bgSoft: "#f4ebef", bgCard: "#ffffff",
      ink: "#1f161b", ink2: "#4a3a43", ink3: "#7c6c75", ink4: "#a8a0a4",
      rule: "#e9dde2", ruleSoft: "#f0e6ea",
      accent: "#9c3a5c", accentSoft: "#f4dde6",
      amber: "#c98730", amberSoft: "#f3e1c0",
      ok: "#5a7a4a", okSoft: "#dde7cd",
      warn: "#a87832", warnSoft: "#f0e0bf",
      danger: "#a8452a", dangerSoft: "#f3d9cd",
      cool: "#5e6480", coolSoft: "#dde0eb",
    },
    dark: {
      bg: "#15101a", bgSoft: "#1d1620", bgCard: "#171120",
      ink: "#f0e6ed", ink2: "#d2c5cd", ink3: "#988b94", ink4: "#6a5e66",
      rule: "#2c2230", ruleSoft: "#211826",
      accent: "#c4709a", accentSoft: "#3a1f30",
      amber: "#e6a25a", amberSoft: "#3b2a16",
      ok: "#8fae7c", okSoft: "#24301a",
      warn: "#d9a868", warnSoft: "#362812",
      danger: "#d4694a", dangerSoft: "#3a1e14",
      cool: "#9da4c0", coolSoft: "#1f2240",
    },
  },
};

// Theme metadata — drives the topbar swatches and tweaks panel.
window.EI_THEME_LIST = [
  { key: "warm",   labelZh: "暖陶",   labelEn: "Warm",   swatch: "#c96442" },
  { key: "forest", labelZh: "苔林",   labelEn: "Forest", swatch: "#5a7d3a" },
  { key: "ocean",  labelZh: "深海",   labelEn: "Ocean",  swatch: "#3a5fc4" },
  { key: "plum",   labelZh: "梅子",   labelEn: "Plum",   swatch: "#9c3a5c" },
];

// Font preset packs — switching a preset changes both serif + sans atomically.
// Mono never changes (numbers/labels are anchored).
window.EI_FONT_PRESETS = [
  {
    key: "editorial",
    labelZh: "编辑级",   labelEn: "Editorial",
    descZh: "默认 · 中文衬线 + Inter，沉静呼吸感",
    descEn: "Default · Chinese serif + Inter, quietly editorial",
    serif: "Noto Serif SC",
    sans:  "Inter",
  },
  {
    key: "modern",
    labelZh: "现代",     labelEn: "Modern",
    descZh: "西文衬线 + 现代无衬线，更接近 SaaS 产品感",
    descEn: "Western serif + modern sans, closer to a SaaS product",
    serif: "Source Serif Pro",
    sans:  "Geist",
  },
  {
    key: "magazine",
    labelZh: "杂志",     labelEn: "Magazine",
    descZh: "瘦长 Garamond + IBM Plex，偏 print 杂志风",
    descEn: "Slim Garamond + IBM Plex, leans print magazine",
    serif: "Cormorant Garamond",
    sans:  "IBM Plex Sans",
  },
];

// Inject global styles once
if (!document.getElementById("ei-global")) {
  const s = document.createElement("style");
  s.id = "ei-global";
  s.textContent = `
    :root {
      --ei-serif: "Noto Serif SC", "Source Serif Pro", Georgia, serif;
      --ei-sans: "Inter", "PingFang SC", -apple-system, sans-serif;
      --ei-mono: "JetBrains Mono", "SF Mono", "Consolas", monospace;
    }
    * { box-sizing: border-box; }
    body { margin: 0; font-family: var(--ei-sans); font-feature-settings: "ss01", "cv11"; -webkit-font-smoothing: antialiased; }
    .ei-serif { font-family: var(--ei-serif); font-weight: 500; letter-spacing: -0.01em; }
    .ei-mono { font-family: var(--ei-mono); font-feature-settings: "tnum"; }
    .ei-label { font-family: var(--ei-mono); font-size: 11px; letter-spacing: 0.08em; text-transform: uppercase; }
    .ei-scroll { scrollbar-width: thin; scrollbar-color: rgba(0,0,0,0.12) transparent; }
    .ei-scroll::-webkit-scrollbar { width: 8px; height: 8px; }
    .ei-scroll::-webkit-scrollbar-thumb { background: rgba(0,0,0,0.12); border-radius: 4px; }
    .ei-scroll::-webkit-scrollbar-track { background: transparent; }
    @keyframes ei-fadein { from { opacity: 0.01; transform: translateY(4px);} to { opacity: 1; transform: none;} }
    .ei-fadein { animation: ei-fadein .28s ease-out; animation-fill-mode: forwards; }
    @keyframes ei-pulse { 0%, 100% { opacity: 1;} 50% { opacity: 0.4;} }
    .ei-pulse { animation: ei-pulse 1.6s ease-in-out infinite; }
    button { font-family: inherit; cursor: pointer; }
    .ei-link { color: inherit; text-decoration: underline; text-decoration-thickness: 1px; text-underline-offset: 3px; text-decoration-color: currentColor; opacity: 0.8; cursor: pointer; }
    .ei-link:hover { opacity: 1; }
    input, textarea { font-family: inherit; }
    .ei-slider-overlay { -webkit-appearance: none; appearance: none; background: transparent; cursor: pointer; }
    .ei-slider-overlay::-webkit-slider-runnable-track { background: transparent; height: 100%; }
    .ei-slider-overlay::-moz-range-track { background: transparent; height: 100%; border: none; }
    .ei-slider-overlay::-webkit-slider-thumb { -webkit-appearance: none; width: 14px; height: 14px; border-radius: 50%; background: #fff; border: 2px solid rgba(0,0,0,0.85); box-shadow: 0 1px 3px rgba(0,0,0,0.25); cursor: pointer; margin-top: 0; }
    .ei-slider-overlay::-moz-range-thumb { width: 14px; height: 14px; border-radius: 50%; background: #fff; border: 2px solid rgba(0,0,0,0.85); box-shadow: 0 1px 3px rgba(0,0,0,0.25); cursor: pointer; }
    @keyframes ei-toast-in { from { opacity: 0; transform: translateY(8px);} to { opacity: 1; transform: translateY(0);} }
    .ei-toast-stack { position: fixed; left: 50%; bottom: 28px; transform: translateX(-50%); display: flex; flex-direction: column; align-items: center; gap: 8px; z-index: 9999; pointer-events: none; }
    .ei-toast { pointer-events: auto; padding: 10px 16px; border-radius: 4px; font-family: var(--ei-sans); font-size: 13px; line-height: 1.4; max-width: 360px; box-shadow: 0 8px 28px rgba(0,0,0,0.18); animation: ei-toast-in .22s ease-out both; display: flex; gap: 8px; align-items: center; }
  `;
  document.head.appendChild(s);
}

// ─────── Toast (global, prop-drill-free) ───────
// Usage: window.eiToast("Saved"), window.eiToast("Failed", { tone: "danger" })
if (!window.eiToast) {
  window.eiToast = (message, opts = {}) => {
    const tone = opts.tone || "neutral";
    let stack = document.getElementById("ei-toast-stack");
    if (!stack) {
      stack = document.createElement("div");
      stack.id = "ei-toast-stack";
      stack.className = "ei-toast-stack";
      document.body.appendChild(stack);
    }
    const palette = {
      neutral: { bg: "#1c1917", fg: "#fafaf7" },
      ok: { bg: "#3f6234", fg: "#fff" },
      warn: { bg: "#a87832", fg: "#fff" },
      danger: { bg: "#a8452a", fg: "#fff" },
    }[tone] || { bg: "#1c1917", fg: "#fafaf7" };
    const node = document.createElement("div");
    node.className = "ei-toast";
    node.style.background = palette.bg;
    node.style.color = palette.fg;
    node.textContent = message;
    stack.appendChild(node);
    setTimeout(() => {
      node.style.transition = "opacity .25s ease, transform .25s ease";
      node.style.opacity = "0";
      node.style.transform = "translateY(-4px)";
      setTimeout(() => node.remove(), 280);
    }, opts.duration || 2200);
  };
}

// ─────── Icons ───────
const Icon = ({ name, size = 16, stroke = 1.5, style = {}, color = "currentColor" }) => {
  const paths = {
    arrow_right: <path d="M5 12h14M13 6l6 6-6 6" />,
    arrow_left: <path d="M19 12H5M11 18l-6-6 6-6" />,
    plus: <path d="M12 5v14M5 12h14" />,
    check: <path d="M5 12l5 5L20 7" />,
    x: <path d="M6 6l12 12M6 18L18 6" />,
    play: <path d="M7 5l12 7-12 7V5z" fill={color} stroke="none" />,
    pause: <><path d="M6 5h4v14H6zM14 5h4v14h-4z" fill={color} stroke="none" /></>,
    mic: <><path d="M12 2a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3z"/><path d="M19 10v1a7 7 0 0 1-14 0v-1M12 18v3"/></>,
    chat: <path d="M4 5h16v11H9l-5 4V5z" />,
    target: <><circle cx="12" cy="12" r="9"/><circle cx="12" cy="12" r="5"/><circle cx="12" cy="12" r="1.4" fill={color} stroke="none"/></>,
    book: <path d="M4 5a2 2 0 0 1 2-2h13v17H7a2 2 0 0 0-2 2V5zM19 3v17" />,
    chart: <path d="M4 20h16M6 16l4-5 4 3 5-7" />,
    resume: <path d="M7 3h8l4 4v14H7V3zM15 3v5h5M9 12h8M9 16h6M9 8h3" />,
    replay: <path d="M3 12a9 9 0 1 0 3-6.7M3 4v5h5" />,
    spark: <path d="M12 3v4M12 17v4M3 12h4M17 12h4M6 6l2.5 2.5M15.5 15.5L18 18M6 18l2.5-2.5M15.5 8.5L18 6" />,
    search: <><circle cx="11" cy="11" r="7"/><path d="M20 20l-4-4"/></>,
    upload: <path d="M12 16V4M6 10l6-6 6 6M4 20h16" />,
    link: <path d="M10 14a4 4 0 0 0 5.66 0l3-3a4 4 0 0 0-5.66-5.66l-1 1M14 10a4 4 0 0 0-5.66 0l-3 3a4 4 0 0 0 5.66 5.66l1-1" />,
    sparkle: <path d="M12 2l1.8 5.5L19 9l-5.2 1.5L12 16l-1.8-5.5L5 9l5.2-1.5z" />,
    chevron_down: <path d="M6 9l6 6 6-6" />,
    chevron_right: <path d="M9 6l6 6-6 6" />,
    dot: <circle cx="12" cy="12" r="4" fill={color} stroke="none" />,
    globe: <><circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3c3 3 3 15 0 18M12 3c-3 3-3 15 0 18"/></>,
    filter: <path d="M4 5h16l-6 8v5l-4 2v-7L4 5z" />,
    calendar: <><rect x="4" y="5" width="16" height="16" rx="2"/><path d="M4 9h16M9 3v4M15 3v4"/></>,
    send: <path d="M4 12l16-8-5 17-4-7-7-2z" />,
    pin: <path d="M12 2l3 6 6 1-4.5 4 1 7L12 16l-5.5 4 1-7L3 9l6-1z" />,
    flag: <path d="M5 22V3h13l-3 5 3 5H5" />,
    clock: <><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></>,
    settings: <><circle cx="12" cy="12" r="3"/><path d="M12 2v3M12 19v3M2 12h3M19 12h3M4.5 4.5l2 2M17.5 17.5l2 2M4.5 19.5l2-2M17.5 6.5l2-2"/></>,
    download: <path d="M12 4v12M6 10l6 6 6-6M4 20h16" />,
    menu: <path d="M4 7h16M4 12h16M4 17h16" />,
    trash: <path d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13M10 11v6M14 11v6" />,
    info: <><circle cx="12" cy="12" r="9"/><path d="M12 8v.01M11 12h1v5h1"/></>,
    edit: <path d="M4 20h4L20 8l-4-4L4 16v4zM14 6l4 4" />,
    sun: <><circle cx="12" cy="12" r="4"/><path d="M12 3v2M12 19v2M5 12H3M21 12h-2M5.6 5.6l1.4 1.4M17 17l1.4 1.4M5.6 18.4L7 17M17 7l1.4-1.4"/></>,
    moon: <path d="M21 13.5A8.5 8.5 0 1110.5 3a7 7 0 0010.5 10.5z" />,
    logout: <path d="M9 4H5a2 2 0 00-2 2v12a2 2 0 002 2h4M16 8l4 4-4 4M20 12h-9" />,
    user: <><circle cx="12" cy="8" r="4"/><path d="M4 21c1.5-4 4.5-6 8-6s6.5 2 8 6"/></>,
    more: <><circle cx="5" cy="12" r="1.4"/><circle cx="12" cy="12" r="1.4"/><circle cx="19" cy="12" r="1.4"/></>,
    building: <path d="M4 21V7l8-4 8 4v14M4 21h16M9 11h0M9 14h0M9 17h0M14 11h0M14 14h0M14 17h0" />,
    star: <path d="M12 3l2.9 6 6.6 1-4.8 4.6 1.2 6.6L12 18l-5.9 3.2L7.3 14.6 2.5 10l6.6-1z" />,
    lang: <path d="M4 7h7M7 4v3M5 7c0 4 2 8 6 10M11 11c-2 4-5 5-7 5M13 20l4-10 4 10M14.5 17h5"/>,
    briefcase: <><rect x="3" y="7" width="18" height="13" rx="1"/><path d="M8 7V4h8v3M3 12h18"/></>,
    layers: <path d="M12 3l9 5-9 5-9-5 9-5zM3 13l9 5 9-5M3 18l9 5 9-5" />,
    file: <path d="M7 3h8l4 4v14H7z M15 3v5h4" />,
    bookmark: <path d="M6 3h12v18l-6-4-6 4z" />,
  };
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth={stroke} strokeLinecap="round" strokeLinejoin="round" style={{ flexShrink: 0, display: "inline-block", verticalAlign: "middle", ...style }}>
      {paths[name] || null}
    </svg>
  );
};

// ─────── Common atoms ───────

const Tag = ({ children, tone = "neutral", T }) => {
  const map = {
    neutral: { bg: T.bgSoft, fg: T.ink2, bd: T.rule },
    accent: { bg: T.accentSoft, fg: T.accent, bd: "transparent" },
    amber: { bg: T.amberSoft, fg: T.warn, bd: "transparent" },
    ok: { bg: T.okSoft, fg: T.ok, bd: "transparent" },
    warn: { bg: T.warnSoft, fg: T.warn, bd: "transparent" },
    danger: { bg: T.dangerSoft, fg: T.danger, bd: "transparent" },
    cool: { bg: T.coolSoft, fg: T.cool, bd: "transparent" },
    muted: { bg: "transparent", fg: T.ink3, bd: T.rule },
  };
  const c = map[tone] || map.neutral;
  return (
    <span style={{
      display: "inline-flex", alignItems: "center", gap: 4,
      padding: "3px 8px", borderRadius: 3, fontSize: 11.5,
      fontFamily: "var(--ei-mono)", letterSpacing: "0.04em",
      background: c.bg, color: c.fg, border: `1px solid ${c.bd}`,
      whiteSpace: "nowrap",
    }}>{children}</span>
  );
};

const Btn = ({ children, onClick, variant = "primary", size = "md", icon, iconRight, T, style = {}, disabled }) => {
  const sizes = {
    sm: { px: 12, h: 30, fs: 13 },
    md: { px: 16, h: 38, fs: 14 },
    lg: { px: 22, h: 46, fs: 15 },
  };
  const variants = {
    primary: { bg: T.ink, fg: T.bg, bd: T.ink },
    secondary: { bg: T.bg, fg: T.ink, bd: T.rule },
    ghost: { bg: "transparent", fg: T.ink2, bd: "transparent" },
    accent: { bg: T.accent, fg: "#fff", bd: T.accent },
    danger: { bg: "transparent", fg: T.danger, bd: T.rule },
  };
  const s = sizes[size]; const v = variants[variant];
  return (
    <button onClick={disabled ? undefined : onClick} disabled={disabled} style={{
      display: "inline-flex", alignItems: "center", justifyContent: "center", gap: 8,
      height: s.h, padding: `0 ${s.px}px`, fontSize: s.fs, fontWeight: 500,
      background: v.bg, color: v.fg, border: `1px solid ${v.bd}`,
      borderRadius: 2, cursor: disabled ? "not-allowed" : "pointer",
      opacity: disabled ? 0.5 : 1,
      fontFamily: "var(--ei-sans)", letterSpacing: "-0.005em",
      transition: "transform .08s ease, opacity .15s", ...style,
    }}
    onMouseDown={(e) => e.currentTarget.style.transform = "translateY(0.5px)"}
    onMouseUp={(e) => e.currentTarget.style.transform = ""}
    onMouseLeave={(e) => e.currentTarget.style.transform = ""}
    >
      {icon && <Icon name={icon} size={s.fs + 2} />}
      {children}
      {iconRight && <Icon name={iconRight} size={s.fs + 2} />}
    </button>
  );
};

const Card = ({ children, T, style = {}, pad = 20, interactive, onClick }) => (
  <div onClick={onClick} style={{
    background: T.bgCard,
    border: `1px solid ${T.rule}`,
    borderRadius: 3,
    padding: pad,
    cursor: interactive ? "pointer" : "default",
    transition: "border-color .15s, transform .15s",
    ...style,
  }}
  onMouseEnter={interactive ? (e) => e.currentTarget.style.borderColor = T.ink3 : undefined}
  onMouseLeave={interactive ? (e) => e.currentTarget.style.borderColor = T.rule : undefined}
  >{children}</div>
);

const SectionHeader = ({ eyebrow, title, sub, T, right }) => (
  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginBottom: 16, gap: 20 }}>
    <div>
      {eyebrow && <div className="ei-label" style={{ color: T.ink3, marginBottom: 8 }}>{eyebrow}</div>}
      <div className="ei-serif" style={{ fontSize: 22, color: T.ink, letterSpacing: "-0.02em" }}>{title}</div>
      {sub && <div style={{ fontSize: 13, color: T.ink3, marginTop: 4 }}>{sub}</div>}
    </div>
    {right}
  </div>
);

// Readiness status — 4-step: 未就绪 / 基本可面 / 建议再练 / 较为充分
const ReadinessDial = ({ level, label, T, size = 56 }) => {
  const states = ["未就绪", "基本可面", "建议再练", "较为充分"];
  const colors = [T.danger, T.warn, T.amber, T.ok];
  const c = colors[level] || T.ink3;
  const state = label || states[level] || states[0];
  const compact = size <= 40;
  return (
    <div style={{ display: "inline-flex", alignItems: "center", gap: 8 }}>
      <span style={{ width: 7, height: 7, borderRadius: 4, background: c, boxShadow: `0 0 0 3px ${T.bgSoft}` }} />
      <span style={{
        border: `1px solid ${T.rule}`,
        background: T.bgSoft,
        color: T.ink,
        borderRadius: 2,
        padding: compact ? "3px 7px" : "5px 10px",
        fontSize: compact ? 11.5 : 12.5,
        fontWeight: 500,
        whiteSpace: "nowrap",
      }}>
        {state}
      </span>
    </div>
  );
};

// Mini sparkline
const Sparkline = ({ values, color, width = 80, height = 24 }) => {
  const min = Math.min(...values), max = Math.max(...values);
  const range = max - min || 1;
  const pts = values.map((v, i) => {
    const x = (i / (values.length - 1)) * width;
    const y = height - ((v - min) / range) * (height - 4) - 2;
    return `${x},${y}`;
  }).join(" ");
  return (
    <svg width={width} height={height} style={{ display: "block" }}>
      <polyline points={pts} fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
};

// Striped placeholder
const Placeholder = ({ label, height = 120, T }) => (
  <div style={{
    height, borderRadius: 2,
    background: `repeating-linear-gradient(135deg, ${T.bgSoft} 0 8px, ${T.ruleSoft} 8px 9px)`,
    border: `1px solid ${T.rule}`, display: "flex", alignItems: "center", justifyContent: "center",
    color: T.ink3, fontFamily: "var(--ei-mono)", fontSize: 11, letterSpacing: "0.06em", textTransform: "uppercase",
  }}>{label}</div>
);

// Key/value label stack (editorial list)
const KV = ({ k, v, T, mono }) => (
  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", padding: "8px 0", borderBottom: `1px dotted ${T.rule}`, gap: 16 }}>
    <div className="ei-label" style={{ color: T.ink3 }}>{k}</div>
    <div style={{ fontSize: 13, color: T.ink, fontFamily: mono ? "var(--ei-mono)" : "inherit", textAlign: "right" }}>{v}</div>
  </div>
);

Object.assign(window, { Icon, Tag, Btn, Card, SectionHeader, ReadinessDial, Sparkline, Placeholder, KV });
