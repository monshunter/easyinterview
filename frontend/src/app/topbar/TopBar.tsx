import { useEffect, useState, type FC } from "react";

import { useDisplayPreferences } from "../display/DisplayPreferencesProvider";
import { SUPPORTED_LOCALES } from "../i18n/localeCatalog";
import { translate } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import {
  PRIMARY_NAV_ROUTES,
  resolvePrimaryNavRoute,
  type RouteName,
} from "../routes";

/**
 * Three primary nav entries match docs/ui-design/auth-and-entry.md §4 after
 * product-scope D-22. Report, auth, and settings
 * routes are intentionally NOT promoted to first-level nav.
 *
 * Labels are rendered through the D1 i18n catalog. English RouteName keys stay
 * canonical so route-state tests and URL hashes do not depend on UI locale.
 *
 * D2 visual contract: every text node uses `ei-text-*` className and every layout
 * literal is sourced from
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

const NAV_ICONS: Record<(typeof PRIMARY_NAV_ROUTES)[number], IconName> = {
  home: "home",
  workspace: "play",
  resume_versions: "file",
};

export interface TopBarProps {
  activeRoute: RouteName;
  onNavigate: (route: LooseRoute) => void;
  /**
   * Whether the current user is authenticated. Defaults to `false`. The
   * unauthenticated branch surfaces the single login entry; the authenticated
   * branch surfaces the single Settings entry without an account dropdown.
   */
  signedIn?: boolean;
  /** Authenticated runtime display name used only for the circular initial. */
  userDisplayName?: string;
}

export const TopBar: FC<TopBarProps> = ({
  activeRoute,
  onNavigate,
  signedIn = false,
  userDisplayName,
}) => {
  const prefs = useDisplayPreferences();
  const t = (key: Parameters<typeof translate>[1]) => translate(prefs.lang, key);
  const [langMenuOpen, setLangMenuOpen] = useState<boolean>(false);
  const activePrimaryRoute = resolvePrimaryNavRoute(activeRoute);
  const userInitial = deriveUserInitial(userDisplayName, prefs.lang);
  const currentLocale =
    SUPPORTED_LOCALES.find((locale) => locale.code === prefs.lang) ??
    SUPPORTED_LOCALES.find((locale) => locale.code === "en") ??
    SUPPORTED_LOCALES[0];

  useEffect(() => {
    if (!langMenuOpen) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setLangMenuOpen(false);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [langMenuOpen]);
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
            aria-current={activePrimaryRoute === name ? "page" : undefined}
            className="ei-topbar-nav-button ei-text-body"
            onClick={() => onNavigate({ name, params: {} })}
          >
            <Icon
              name={NAV_ICONS[name]}
              size={19}
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
          <Icon name="sun" size={15} />
          <span
            data-testid="topbar-dark-track"
            className="ei-topbar-dark-track"
            aria-hidden="true"
          >
            <span
              data-testid="topbar-dark-thumb"
              className="ei-topbar-dark-thumb"
            />
          </span>
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
            <span
              data-testid="topbar-lang-caret"
              className="ei-topbar-caret"
              aria-hidden="true"
            >
              <Icon
                name="chevronDown"
                size={14}
                data-testid="topbar-lang-chevron"
              />
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
          <button
            type="button"
            data-testid="topbar-settings"
            aria-label={t("user.settings")}
            title={t("user.settings")}
            className="ei-topbar-settings"
            onClick={() => onNavigate({ name: "settings", params: {} })}
          >
            <span className="ei-topbar-settings-mark" aria-hidden="true">
              {userInitial}
            </span>
          </button>
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

type IconName =
  | "home"
  | "target"
  | "search"
  | "play"
  | "file"
  | "flag"
  | "settings"
  | "check"
  | "moon"
  | "sun"
  | "globe"
  | "chevronDown";

interface IconProps {
  name: IconName;
  size?: number;
  "data-testid"?: string;
}

const Icon: FC<IconProps> = ({ name, size = 13, "data-testid": testId }) => {
  const paths: Record<IconName, JSX.Element> = {
    home: <path d="M3.5 11.2L12 4l8.5 7.2M5.7 9.5V20h4.1v-5.8h4.4V20h4.1V9.5" />,
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
    play: (
      <>
        <circle cx="12" cy="12" r="9" />
        <path d="M10 8.5l5.5 3.5-5.5 3.5v-7z" fill="currentColor" stroke="none" />
      </>
    ),
    file: <path d="M7 3h8l4 4v14H7z M15 3v5h4" />,
    flag: <path d="M5 22V3h13l-3 5 3 5H5" />,
    settings: <path d="M19.14 12.94a7.43 7.43 0 000-1.88l2.03-1.58-1.92-3.32-2.39.96a7.2 7.2 0 00-1.63-.94L14.87 3h-3.84l-.36 2.18a7.2 7.2 0 00-1.63.94l-2.39-.96-1.92 3.32 2.03 1.58a7.43 7.43 0 000 1.88l-2.03 1.58 1.92 3.32 2.39-.96c.5.39 1.04.7 1.63.94l.36 2.18h3.84l.36-2.18c.59-.24 1.13-.55 1.63-.94l2.39.96 1.92-3.32-2.03-1.58zM12.95 15.2a3.2 3.2 0 110-6.4 3.2 3.2 0 010 6.4z" fill="currentColor" stroke="none" />,
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
    chevronDown: <path d="M6.5 9.5 12 15l5.5-5.5" />,
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

function deriveUserInitial(displayName: string | undefined, lang: string): string {
  const normalized = displayName?.trim() ?? "";
  const firstCharacter = Array.from(normalized)[0];
  if (!firstCharacter) return "?";
  return firstCharacter.toLocaleUpperCase(lang === "zh" ? "zh-CN" : "en-US");
}
