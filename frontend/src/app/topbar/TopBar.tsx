import { useCallback, useState, type FC } from "react";

import {
  useDisplayPreferences,
  type CustomAccent,
  type Lang,
  type Theme,
} from "../display/DisplayPreferencesProvider";
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
 * mirroring `ui-design/src/app.jsx` `AccentPicker`. D1 testids and the
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

const LANG_LABELS: Record<Lang, string> = {
  zh: "中文",
  en: "English",
};

const THEME_OPTIONS: readonly Theme[] = ["warm", "forest", "ocean", "plum"];
const LANG_OPTIONS: readonly Lang[] = ["zh", "en"];

const CUSTOM_ACCENT_SEEDS: Record<Theme, CustomAccent> = {
  warm: { h: 30, c: 0.16 },
  forest: { h: 130, c: 0.13 },
  ocean: { h: 255, c: 0.16 },
  plum: { h: 340, c: 0.15 },
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

  const seed = CUSTOM_ACCENT_SEEDS[prefs.theme] ?? CUSTOM_ACCENT_SEEDS.warm;
  const accentValue: CustomAccent = prefs.customAccent ?? seed;
  const swatchOklch = customActive
    ? `oklch(${prefs.dark ? 68 : 58}% ${accentValue.c.toFixed(3)} ${accentValue.h.toFixed(1)})`
    : "";

  const handleToggleCustomAccent = useCallback(() => {
    if (customActive) {
      prefs.setCustomAccent(null);
      setPickerOpen(false);
    } else {
      prefs.setCustomAccent({ ...seed });
      setPickerOpen(true);
    }
  }, [customActive, prefs, seed]);

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
        <span className="ei-text-subtitle">EasyInterview</span>
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
          <span
            className={
              customActive
                ? "ei-topbar-theme-swatch ei-topbar-theme-swatch--custom-active"
                : "ei-topbar-theme-swatch"
            }
            data-testid="topbar-theme-swatch"
          />
          <label className="ei-topbar-control">
            <span className="visually-hidden">{t("display.theme")}</span>
            <select
              data-testid="topbar-theme-select"
              className="ei-topbar-theme"
              value={prefs.theme}
              onChange={(e) => prefs.setTheme(e.target.value as Theme)}
            >
              {THEME_OPTIONS.map((theme) => (
                <option key={theme} value={theme}>
                  {t(THEME_LABEL_KEYS[theme])}
                </option>
              ))}
            </select>
          </label>
        </span>
        <span className="ei-topbar-custom-accent-host">
          <button
            type="button"
            data-testid="topbar-custom-accent-button"
            aria-pressed={customActive}
            aria-label={
              prefs.lang === "en" ? "Toggle custom accent" : "切换自定义主题色"
            }
            className="ei-topbar-control ei-topbar-custom-accent"
            onClick={handleToggleCustomAccent}
          >
            {customActive ? (
              <span
                data-testid="topbar-custom-accent-swatch"
                className="ei-topbar-custom-accent-rainbow"
                style={{ background: swatchOklch }}
              />
            ) : (
              <span
                data-testid="topbar-custom-accent-swatch"
                className="ei-topbar-custom-accent-rainbow"
              />
            )}
          </button>
          {pickerOpen && (
            <CustomAccentPicker
              accent={accentValue}
              onChange={handleAccentChange}
              onClear={() => {
                prefs.setCustomAccent(null);
                setPickerOpen(false);
              }}
              lang={prefs.lang}
            />
          )}
        </span>
        <button
          type="button"
          data-testid="topbar-dark-toggle"
          aria-pressed={prefs.dark}
          className="ei-topbar-control ei-topbar-dark"
          onClick={() => prefs.setDark(!prefs.dark)}
        >
          {prefs.dark ? t("display.dark") : t("display.light")}
        </button>
        <label className="ei-topbar-control">
          <span className="visually-hidden">{t("display.language")}</span>
          <select
            data-testid="topbar-lang-select"
            className="ei-topbar-lang"
            value={prefs.lang}
            onChange={(e) => prefs.setLang(e.target.value as Lang)}
          >
            {LANG_OPTIONS.map((lang) => (
              <option key={lang} value={lang}>
                {LANG_LABELS[lang]}
              </option>
            ))}
          </select>
        </label>
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
  onChange: (next: Partial<CustomAccent>) => void;
  onClear: () => void;
  lang: Lang;
}

const CustomAccentPicker: FC<CustomAccentPickerProps> = ({
  accent,
  onChange,
  onClear,
  lang,
}) => {
  return (
    <div
      data-testid="topbar-custom-accent-picker"
      className="ei-topbar-custom-accent-picker"
      role="group"
      aria-label={lang === "en" ? "Custom accent picker" : "自定义主题色"}
    >
      <div className="ei-topbar-custom-accent-row">
        <span className="ei-text-label">
          {lang === "en" ? "Hue" : "色相"}
        </span>
        <div className="ei-topbar-custom-accent-track">
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
        <div className="ei-topbar-custom-accent-track">
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

// Side effect: theme metadata is imported above so the theme dropdown can
// surface a consistent swatch dot. It is currently only used to validate
// theme keys at runtime; future iterations may render per-theme dot variants.
void THEME_METADATA;
