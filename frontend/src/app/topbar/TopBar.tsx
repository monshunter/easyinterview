import { useCallback, useEffect, useState, type FC } from "react";

import {
  useDisplayPreferences,
  type CustomAccent,
  type Lang,
  type Theme,
} from "../display/DisplayPreferencesProvider";
import { SUPPORTED_LOCALES } from "../i18n/localeCatalog";
import { translate } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import { PRIMARY_NAV_ROUTES, type RouteName } from "../routes";
import { THEME_METADATA } from "../theme/themes.data";

/**
 * Five primary nav entries match docs/spec/frontend-shell/spec.md §2.1 and
 * docs/ui-design/auth-and-entry.md §4. Reports / company-intel / auth /
 * profile / settings are intentionally NOT promoted to first-level nav.
 *
 * Labels are rendered through the D1 i18n catalog. English RouteName keys stay
 * canonical so route-state tests and URL hashes do not depend on UI locale.
 *
 * D2 visual contract: every text node uses `ei-text-*` className, every layout
 * literal (height 58, padding 0 32, gap 28, etc.) is sourced from
 * `topbar.css`, and the custom-accent control surfaces hue / chroma sliders
 * mirroring `ui-design/src/app.jsx` `AccentPicker`. Language selection is a
 * TopBar dropdown, not a binary toggle, so future locale options can be added
 * without changing the control shape. D1 testids and the
 * `aria-current` / `aria-pressed` contract remain unchanged.
 */
const NAV_LABEL_KEYS: Record<(typeof PRIMARY_NAV_ROUTES)[number], Parameters<typeof translate>[1]> = {
  home: "nav.home",
  jd_match: "nav.jd_match",
  workspace: "nav.workspace",
  resume_versions: "nav.resume_versions",
  debrief: "nav.debrief",
};

const THEME_LABEL_KEYS: Record<Theme, Parameters<typeof translate>[1]> = {
  warm: "theme.warm",
  forest: "theme.forest",
  ocean: "theme.ocean",
  plum: "theme.plum",
};

const THEME_OPTIONS: readonly Theme[] = ["warm", "forest", "ocean", "plum"];
const CUSTOM_ACCENT_SEEDS: Record<Theme, CustomAccent> = {
  warm: { h: 30, c: 0.16 },
  forest: { h: 130, c: 0.13 },
  ocean: { h: 255, c: 0.16 },
  plum: { h: 340, c: 0.15 },
};

const NAV_ICONS: Record<(typeof PRIMARY_NAV_ROUTES)[number], IconName> = {
  home: "target",
  jd_match: "search",
  workspace: "play",
  resume_versions: "file",
  debrief: "flag",
};

export interface TopBarProps {
  activeRoute: RouteName;
  onNavigate: (route: LooseRoute) => void;
  /**
   * Whether the current user is authenticated. Defaults to `false`. The
   * unauthenticated branch surfaces login / register entries; the
   * authenticated branch surfaces the inline `topbar-user-menu` row.
   */
  signedIn?: boolean;
}

export const TopBar: FC<TopBarProps> = ({
  activeRoute,
  onNavigate,
  signedIn = false,
}) => {
  const prefs = useDisplayPreferences();
  const t = (key: Parameters<typeof translate>[1]) => translate(prefs.lang, key);
  const customActive = prefs.customAccent != null;
  const [pickerOpen, setPickerOpen] = useState<boolean>(customActive);
  const [themeMenuOpen, setThemeMenuOpen] = useState<boolean>(false);
  const [langMenuOpen, setLangMenuOpen] = useState<boolean>(false);

  const seed = CUSTOM_ACCENT_SEEDS[prefs.theme] ?? CUSTOM_ACCENT_SEEDS.warm;
  const accentValue: CustomAccent = prefs.customAccent ?? seed;
  const swatchOklch = customActive
    ? `oklch(${prefs.dark ? 68 : 58}% ${accentValue.c.toFixed(3)} ${accentValue.h.toFixed(1)})`
    : "";
  const currentLocale =
    SUPPORTED_LOCALES.find((locale) => locale.code === prefs.lang) ??
    SUPPORTED_LOCALES.find((locale) => locale.code === "en") ??
    SUPPORTED_LOCALES[0];

  useEffect(() => {
    if (!themeMenuOpen && !langMenuOpen) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setThemeMenuOpen(false);
        setLangMenuOpen(false);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [themeMenuOpen, langMenuOpen]);

  const handleAccentChange = useCallback(
    (next: Partial<CustomAccent>) => {
      prefs.setCustomAccent({ ...accentValue, ...next });
    },
    [accentValue, prefs],
  );

  return (
    <header data-testid="app-shell-topbar" className="ei-shell-topbar">
      <div className="ei-topbar-brand">
        <span className="ei-topbar-brand-mark" aria-hidden="true">
          E
        </span>
        <span className="ei-topbar-brand-copy">
          <span className="ei-text-subtitle">EasyInterview</span>
        </span>
      </div>
      <nav
        data-testid="topbar-primary-nav"
        aria-label="primary"
        className="ei-topbar-nav"
      >
        {PRIMARY_NAV_ROUTES.map((name) => (
          <button
            key={name}
            type="button"
            data-testid={`topbar-nav-${name}`}
            aria-current={activeRoute === name ? "page" : undefined}
            className="ei-topbar-nav-button ei-text-body"
            onClick={() => onNavigate({ name, params: {} })}
          >
            <Icon
              name={NAV_ICONS[name]}
              size={13}
              data-testid={`topbar-nav-icon-${name}`}
            />
            {t(NAV_LABEL_KEYS[name])}
          </button>
        ))}
      </nav>
      <div className="ei-topbar-spacer" />
      <div
        data-testid="topbar-display-controls"
        className="ei-topbar-controls"
      >
        <span className="ei-topbar-theme-wrap">
          <button
            type="button"
            data-testid="topbar-theme-button"
            title={prefs.lang === "en" ? "Theme" : "主题色"}
            aria-label={prefs.lang === "en" ? "Theme" : "主题色"}
            aria-expanded={themeMenuOpen}
            className="ei-topbar-control ei-topbar-theme-button"
            onClick={() => setThemeMenuOpen((open) => !open)}
          >
            <span
              className="ei-topbar-theme-swatch"
              data-testid="topbar-theme-swatch"
            />
            <span className="ei-topbar-caret" aria-hidden="true">
              ▾
            </span>
          </button>
          {themeMenuOpen && (
            <>
              <button
                type="button"
                className="ei-topbar-menu-backdrop"
                aria-label={prefs.lang === "en" ? "Close theme menu" : "关闭主题菜单"}
                onClick={() => setThemeMenuOpen(false)}
              />
              <div
                data-testid="topbar-theme-menu"
                className="ei-topbar-theme-menu"
              >
                <div className="ei-text-label ei-topbar-theme-menu-label">
                  {prefs.lang === "en" ? "Theme" : "主题色"}
                </div>
                {THEME_OPTIONS.map((theme) => {
                  const selected = prefs.theme === theme && !customActive;
                  const metadata = THEME_METADATA.find((item) => item.key === theme);
                  return (
                    <button
                      key={theme}
                      type="button"
                      data-testid={`topbar-theme-option-${theme}`}
                      aria-pressed={selected}
                      className={
                        selected
                          ? "ei-topbar-theme-option ei-topbar-theme-option--selected"
                          : "ei-topbar-theme-option"
                      }
                      onClick={() => {
                        prefs.setTheme(theme);
                        prefs.setCustomAccent(null);
                        setPickerOpen(false);
                        setThemeMenuOpen(false);
                      }}
                    >
                      <span
                        className="ei-topbar-theme-option-swatch"
                        style={{ background: metadata?.swatch }}
                        aria-hidden="true"
                      />
                      <span>{t(THEME_LABEL_KEYS[theme])}</span>
                      {selected && <Icon name="check" size={12} />}
                    </button>
                  );
                })}
                <div className="ei-topbar-theme-separator" />
                <button
                  type="button"
                  data-testid="topbar-theme-custom-option"
                  aria-pressed={customActive}
                  className={
                    customActive
                      ? "ei-topbar-theme-option ei-topbar-theme-option--selected"
                      : "ei-topbar-theme-option"
                  }
                  onClick={() => {
                    if (!customActive) {
                      prefs.setCustomAccent({ ...seed });
                      setPickerOpen(true);
                    } else {
                      setPickerOpen((open) => !open);
                    }
                  }}
                >
                  <span
                    data-testid="topbar-custom-accent-swatch"
                    className="ei-topbar-custom-accent-rainbow"
                    style={customActive ? { background: swatchOklch } : undefined}
                  />
                  <span>{prefs.lang === "en" ? "Custom" : "自定义"}</span>
                  {customActive ? (
                    <Icon name="check" size={12} />
                  ) : (
                    <span className="ei-topbar-caret" aria-hidden="true">
                      {pickerOpen ? "▴" : "▾"}
                    </span>
                  )}
                </button>
                {pickerOpen && (
                  <CustomAccentPicker
                    accent={accentValue}
                    active={customActive}
                    onChange={handleAccentChange}
                    onClear={() => {
                      prefs.setCustomAccent(null);
                      setPickerOpen(false);
                    }}
                    lang={prefs.lang}
                    dark={prefs.dark}
                  />
                )}
              </div>
            </>
          )}
        </span>
        <button
          type="button"
          data-testid="topbar-dark-toggle"
          aria-pressed={prefs.dark}
          aria-label={
            prefs.dark
              ? prefs.lang === "en"
                ? "Switch to light"
                : "切换到浅色"
              : prefs.lang === "en"
                ? "Switch to dark"
                : "切换到暗色"
          }
          title={
            prefs.dark
              ? prefs.lang === "en"
                ? "Switch to light"
                : "切换到浅色"
              : prefs.lang === "en"
                ? "Switch to dark"
                : "切换到暗色"
          }
          className="ei-topbar-control ei-topbar-dark"
          onClick={() => prefs.setDark(!prefs.dark)}
        >
          <Icon name={prefs.dark ? "sun" : "moon"} size={12} />
        </button>
        <span className="ei-topbar-lang-wrap">
          <button
            type="button"
            data-testid="topbar-lang-toggle"
            aria-label={`${t("display.language")}: ${currentLocale.label}`}
            aria-expanded={langMenuOpen}
            className="ei-topbar-control ei-topbar-lang"
            onClick={() => setLangMenuOpen((open) => !open)}
          >
            <Icon name="globe" size={12} />
            <span className="ei-topbar-lang-current">{currentLocale.label}</span>
            <span className="ei-topbar-caret" aria-hidden="true">
              ▾
            </span>
          </button>
          {langMenuOpen && (
            <>
              <button
                type="button"
                className="ei-topbar-menu-backdrop"
                aria-label={prefs.lang === "en" ? "Close language menu" : "关闭语言菜单"}
                onClick={() => setLangMenuOpen(false)}
              />
              <div
                data-testid="topbar-lang-menu"
                className="ei-topbar-lang-menu"
              >
                <div className="ei-text-label ei-topbar-lang-menu-label">
                  {prefs.lang === "en" ? "Language" : "界面语言"}
                </div>
                {SUPPORTED_LOCALES.map((locale) => {
                  const selected = prefs.lang === locale.code;
                  return (
                    <button
                      key={locale.code}
                      type="button"
                      data-testid={`topbar-lang-option-${locale.code}`}
                      aria-pressed={selected}
                      className={
                        selected
                          ? "ei-topbar-lang-option ei-topbar-lang-option--selected"
                          : "ei-topbar-lang-option"
                      }
                      onClick={() => {
                        prefs.setLang(locale.code);
                        setLangMenuOpen(false);
                      }}
                    >
                      <Icon name="globe" size={13} />
                      <span>{locale.label}</span>
                      {selected ? (
                        <Icon name="check" size={12} />
                      ) : (
                        <span className="ei-text-label ei-topbar-lang-short">
                          {locale.shortLabel}
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            </>
          )}
        </span>
      </div>
      <div
        data-testid="topbar-user-area"
        data-signed-in={signedIn ? "true" : "false"}
        className="ei-topbar-user"
      >
        {signedIn ? (
          <nav
            data-testid="topbar-user-menu"
            aria-label="user"
            className="ei-topbar-user-menu"
          >
            <button
              type="button"
              data-testid="topbar-user-profile"
              className="ei-topbar-user-button ei-text-body"
              onClick={() => onNavigate({ name: "profile", params: {} })}
            >
              {t("user.profile")}
            </button>
            <button
              type="button"
              data-testid="topbar-user-settings"
              className="ei-topbar-user-button ei-text-body"
              onClick={() => onNavigate({ name: "settings", params: {} })}
            >
              {t("user.settings")}
            </button>
            <button
              type="button"
              data-testid="topbar-user-logout"
              className="ei-topbar-user-button ei-topbar-user-button--logout ei-text-body"
              onClick={() => onNavigate({ name: "auth_logout", params: {} })}
            >
              {t("user.logout")}
            </button>
          </nav>
        ) : (
          <>
            <button
              type="button"
              data-testid="topbar-login"
              className="ei-topbar-auth-login"
              onClick={() => onNavigate({ name: "auth_login", params: {} })}
            >
              {t("auth.login")}
            </button>
            <button
              type="button"
              data-testid="topbar-register"
              className="ei-topbar-auth-register"
              onClick={() => onNavigate({ name: "auth_register", params: {} })}
            >
              {t("auth.register")}
            </button>
          </>
        )}
      </div>
    </header>
  );
};

interface CustomAccentPickerProps {
  accent: CustomAccent;
  active: boolean;
  onChange: (next: Partial<CustomAccent>) => void;
  onClear: () => void;
  lang: Lang;
  dark: boolean;
}

const CustomAccentPicker: FC<CustomAccentPickerProps> = ({
  accent,
  active,
  onChange,
  onClear,
  lang,
  dark,
}) => {
  const accentL = dark ? 68 : 58;
  const normalizedHue = ((accent.h % 360) + 360) % 360;
  const hueStops = Array.from({ length: 13 }, (_, index) => {
    const h = (index / 12) * 360;
    return `oklch(${accentL}% 0.18 ${h})`;
  }).join(", ");
  const hueGradient = `linear-gradient(to right, ${hueStops})`;
  const chromaGradient = `linear-gradient(to right, oklch(${accentL}% 0 ${normalizedHue}), oklch(${accentL}% 0.25 ${normalizedHue}))`;
  const previewAccent = `oklch(${accentL}% ${accent.c.toFixed(3)} ${normalizedHue.toFixed(1)})`;

  return (
    <div
      data-testid="topbar-custom-accent-picker"
      className="ei-topbar-custom-accent-picker"
      role="group"
      aria-label={lang === "en" ? "Custom accent picker" : "自定义主题色"}
    >
      <div className="ei-topbar-custom-accent-preview">
        <span
          className="ei-topbar-custom-accent-preview-swatch"
          style={{ background: previewAccent, opacity: active ? 1 : 0.55 }}
          aria-hidden="true"
        />
        <span className="ei-topbar-custom-accent-value">
          oklch({accentL}% {accent.c.toFixed(3)} {Math.round(normalizedHue)})
        </span>
      </div>
      <div className="ei-topbar-custom-accent-row">
        <span className="ei-text-label">
          {lang === "en" ? "Hue" : "色相"}
        </span>
        <div
          className="ei-topbar-custom-accent-track"
          style={{ background: hueGradient, opacity: active ? 1 : 0.55 }}
        >
          <input
            data-testid="topbar-custom-accent-hue"
            className="ei-topbar-custom-accent-input"
            type="range"
            min={0}
            max={360}
            step={1}
            aria-label={lang === "en" ? "Custom accent hue" : "自定义主题色色相"}
            value={accent.h}
            onChange={(e) => onChange({ h: Number(e.target.value) })}
          />
        </div>
      </div>
      <div className="ei-topbar-custom-accent-row">
        <span className="ei-text-label">
          {lang === "en" ? "Chroma" : "饱和度"}
        </span>
        <div
          className="ei-topbar-custom-accent-track"
          style={{ background: chromaGradient, opacity: active ? 1 : 0.55 }}
        >
          <input
            data-testid="topbar-custom-accent-chroma"
            className="ei-topbar-custom-accent-input"
            type="range"
            min={0}
            max={0.25}
            step={0.005}
            aria-label={
              lang === "en" ? "Custom accent chroma" : "自定义主题色饱和度"
            }
            value={accent.c}
            onChange={(e) => onChange({ c: Number(e.target.value) })}
          />
        </div>
      </div>
      <button
        type="button"
        data-testid="topbar-custom-accent-clear"
        className="ei-link"
        onClick={onClear}
      >
        {lang === "en" ? "Reset to theme accent" : "恢复主题默认色"}
      </button>
    </div>
  );
};

type IconName =
  | "target"
  | "search"
  | "play"
  | "file"
  | "flag"
  | "check"
  | "moon"
  | "sun"
  | "globe";

interface IconProps {
  name: IconName;
  size?: number;
  "data-testid"?: string;
}

const Icon: FC<IconProps> = ({ name, size = 13, "data-testid": testId }) => {
  const paths: Record<IconName, JSX.Element> = {
    target: (
      <>
        <circle cx="12" cy="12" r="9" />
        <circle cx="12" cy="12" r="5" />
        <circle cx="12" cy="12" r="1.4" fill="currentColor" stroke="none" />
      </>
    ),
    search: (
      <>
        <circle cx="11" cy="11" r="7" />
        <path d="M20 20l-4-4" />
      </>
    ),
    play: <path d="M7 5l12 7-12 7V5z" fill="currentColor" stroke="none" />,
    file: <path d="M7 3h8l4 4v14H7z M15 3v5h4" />,
    flag: <path d="M5 22V3h13l-3 5 3 5H5" />,
    check: <path d="M5 12l5 5L20 7" />,
    moon: <path d="M21 13.5A8.5 8.5 0 1110.5 3a7 7 0 0010.5 10.5z" />,
    sun: (
      <>
        <circle cx="12" cy="12" r="4" />
        <path d="M12 3v2M12 19v2M5 12H3M21 12h-2M5.6 5.6l1.4 1.4M17 17l1.4 1.4M5.6 18.4L7 17M17 7l1.4-1.4" />
      </>
    ),
    globe: (
      <>
        <circle cx="12" cy="12" r="9" />
        <path d="M3 12h18M12 3c3 3 3 15 0 18M12 3c-3 3-3 15 0 18" />
      </>
    ),
  };
  return (
    <svg
      data-testid={testId}
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.8}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      className="ei-topbar-icon"
    >
      {paths[name]}
    </svg>
  );
};
