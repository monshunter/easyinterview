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
import type { UserContext } from "../../api/generated/types";

/**
 * Three primary nav entries match docs/ui-design/auth-and-entry.md §4 after
 * product-scope D-22. Report, auth, and settings
 * routes are intentionally NOT promoted to first-level nav.
 *
 * Labels are rendered through the D1 i18n catalog. English RouteName keys stay
 * canonical so route-state tests and URL hashes do not depend on UI locale.
 *
 * D2 visual contract: every text node uses `ei-text-*` className, every layout
 * literal (height 58, padding 0 32, gap 28, etc.) is sourced from
 * `topbar.css`, and the custom-accent control surfaces hue / chroma sliders
 * mirroring `formal frontend implementation` `AccentPicker`. Language selection is a
 * TopBar dropdown, not a binary toggle, so future locale options can be added
 * without changing the control shape. D1 testids and the
 * `aria-current` / `aria-pressed` contract remain unchanged.
 */
const NAV_LABEL_KEYS: Record<(typeof PRIMARY_NAV_ROUTES)[number], Parameters<typeof translate>[1]> = {
  home: "nav.home",
  workspace: "nav.workspace",
  resume_versions: "nav.resume_versions",
};

const THEME_OPTIONS = ["ocean", "plum"] as const satisfies readonly Theme[];
const THEME_LABEL_KEYS: Record<(typeof THEME_OPTIONS)[number], Parameters<typeof translate>[1]> = {
  ocean: "theme.ocean",
  plum: "theme.plum",
};
const CUSTOM_ACCENT_SEEDS: Record<Theme, CustomAccent> = {
  ocean: { h: 255, c: 0.16 },
  plum: { h: 340, c: 0.15 },
};

const NAV_ICONS: Record<(typeof PRIMARY_NAV_ROUTES)[number], IconName> = {
  home: "target",
  workspace: "play",
  resume_versions: "file",
};

export interface TopBarProps {
  activeRoute: RouteName;
  onNavigate: (route: LooseRoute) => void;
  /**
   * Whether the current user is authenticated. Defaults to `false`. The
   * unauthenticated branch surfaces the single login entry; the authenticated
   * branch surfaces the avatar chip + dropdown from
   * formal frontend implementation
   */
  signedIn?: boolean;
  user?: Pick<UserContext, "displayName" | "emailMasked">;
}

export const TopBar: FC<TopBarProps> = ({
  activeRoute,
  onNavigate,
  signedIn = false,
  user,
}) => {
  const prefs = useDisplayPreferences();
  const t = (key: Parameters<typeof translate>[1]) => translate(prefs.lang, key);
  const customActive = prefs.customAccent != null;
  const [pickerOpen, setPickerOpen] = useState<boolean>(customActive);
  const [themeMenuOpen, setThemeMenuOpen] = useState<boolean>(false);
  const [langMenuOpen, setLangMenuOpen] = useState<boolean>(false);
  const [userMenuOpen, setUserMenuOpen] = useState<boolean>(false);

  const seed = CUSTOM_ACCENT_SEEDS[prefs.theme] ?? CUSTOM_ACCENT_SEEDS.ocean;
  const accentValue: CustomAccent = prefs.customAccent ?? seed;
  const swatchOklch = customActive
    ? `oklch(${prefs.dark ? 68 : 58}% ${accentValue.c.toFixed(3)} ${accentValue.h.toFixed(1)})`
    : "";
  const currentLocale =
    SUPPORTED_LOCALES.find((locale) => locale.code === prefs.lang) ??
    SUPPORTED_LOCALES.find((locale) => locale.code === "en") ??
    SUPPORTED_LOCALES[0];

  useEffect(() => {
    if (!userMenuOpen && !themeMenuOpen && !langMenuOpen) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setUserMenuOpen(false);
        setThemeMenuOpen(false);
        setLangMenuOpen(false);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [userMenuOpen, themeMenuOpen, langMenuOpen]);

  const handleAccentChange = useCallback(
    (next: Partial<CustomAccent>) => {
      prefs.setCustomAccent({ ...accentValue, ...next });
    },
    [accentValue, prefs],
  );
  const userName =
    user?.displayName?.trim() || (prefs.lang === "en" ? "Candidate" : "候选人");
  const userEmail =
    user?.emailMasked?.trim() ||
    (prefs.lang === "en" ? "Email unavailable" : "邮箱未提供");
  const userInitials = getInitials(userName);

  const navigateFromUserMenu = (route: LooseRoute) => {
    setUserMenuOpen(false);
    onNavigate(route);
  };

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
                    onChange={handleAccentChange}
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
          <div className="ei-topbar-user-wrap">
            <button
              type="button"
              data-testid="topbar-user-chip"
              aria-label={prefs.lang === "en" ? "User menu" : "用户菜单"}
              aria-expanded={userMenuOpen}
              className="ei-topbar-user-chip"
              onClick={() => setUserMenuOpen((open) => !open)}
            >
              <span data-testid="topbar-user-avatar" className="ei-topbar-user-avatar">
                {userInitials}
              </span>
              <span data-testid="topbar-user-name" className="ei-topbar-user-name">
                {userName}
              </span>
              <span className="ei-topbar-caret" aria-hidden="true">
                ▾
              </span>
            </button>
            {userMenuOpen && (
              <>
                <button
                  type="button"
                  data-testid="topbar-user-backdrop"
                  className="ei-topbar-menu-backdrop"
                  aria-label={prefs.lang === "en" ? "Close user menu" : "关闭用户菜单"}
                  onClick={() => setUserMenuOpen(false)}
                />
                <nav
                  data-testid="topbar-user-menu"
                  aria-label="user"
                  className="ei-topbar-user-menu"
                >
                  <div
                    data-testid="topbar-user-menu-header"
                    className="ei-topbar-user-menu-header"
                  >
                    <div className="ei-topbar-user-menu-name">{userName}</div>
                    <div
                      data-testid="topbar-user-email"
                      className="ei-topbar-user-menu-email"
                    >
                      {userEmail}
                    </div>
                  </div>
                  <button
                    type="button"
                    data-testid="topbar-user-settings"
                    className="ei-topbar-user-button ei-text-body"
                    onClick={() => navigateFromUserMenu({ name: "settings", params: {} })}
                  >
                    <Icon name="settings" size={13} />
                    {t("user.settings")}
                  </button>
                  <div className="ei-topbar-user-separator" />
                  <button
                    type="button"
                    data-testid="topbar-user-logout"
                    className="ei-topbar-user-button ei-topbar-user-button--logout ei-text-body"
                    onClick={() => navigateFromUserMenu({ name: "auth_logout", params: {} })}
                  >
                    <Icon name="logout" size={13} />
                    {t("user.logout")}
                  </button>
                </nav>
              </>
            )}
          </div>
        ) : (
          <button
            type="button"
            data-testid="topbar-login"
            className="ei-topbar-auth-login"
            onClick={() => onNavigate({ name: "auth_login", params: {} })}
          >
            {t("auth.login")}
          </button>
        )}
      </div>
    </header>
  );
};

interface CustomAccentPickerProps {
  accent: CustomAccent;
  onChange: (next: Partial<CustomAccent>) => void;
  lang: Lang;
  dark: boolean;
}

const CustomAccentPicker: FC<CustomAccentPickerProps> = ({
  accent,
  onChange,
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
        <div
          className="ei-topbar-custom-accent-track"
          style={{ background: hueGradient }}
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
          style={{ background: chromaGradient }}
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
    </div>
  );
};

type IconName =
  | "target"
  | "search"
  | "play"
  | "file"
  | "flag"
  | "user"
  | "settings"
  | "logout"
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
    user: (
      <>
        <circle cx="12" cy="8" r="4" />
        <path d="M5 21a7 7 0 0114 0" />
      </>
    ),
    settings: (
      <>
        <circle cx="12" cy="12" r="3" />
        <path d="M12 2v3M12 19v3M2 12h3M19 12h3M4.5 4.5l2 2M17.5 17.5l2 2M4.5 19.5l2-2M17.5 6.5l2-2" />
      </>
    ),
    logout: <path d="M9 4H5a2 2 0 00-2 2v12a2 2 0 002 2h4M16 8l4 4-4 4M20 12h-9" />,
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

function getInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length >= 2) {
    return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
  }
  return (parts[0]?.slice(0, 2) || "U").toUpperCase();
}
