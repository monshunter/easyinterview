import type { FC } from "react";

import {
  useDisplayPreferences,
  type Lang,
  type Theme,
} from "../display/DisplayPreferencesProvider";
import { translate } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import { PRIMARY_NAV_ROUTES, type RouteName } from "../routes";

/**
 * Five primary nav entries match docs/spec/frontend-shell/spec.md §2.1 and
 * docs/ui-design/auth-and-entry.md §4. Reports / company-intel / auth /
 * profile / settings are intentionally NOT promoted to first-level nav.
 *
 * Labels are rendered through the D1 i18n catalog. English RouteName keys stay
 * canonical so route-state tests and URL hashes do not depend on UI locale.
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

export interface TopBarProps {
  activeRoute: RouteName;
  onNavigate: (route: LooseRoute) => void;
  /**
   * Whether the current user is authenticated. Defaults to `false`. The
   * unauthenticated branch surfaces login / register entries; the
   * authenticated branch is expanded by Phase 4.1 (avatar menu with profile /
   * settings / logout).
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
  return (
    <header data-testid="app-shell-topbar">
      <nav data-testid="topbar-primary-nav" aria-label="primary">
        {PRIMARY_NAV_ROUTES.map((name) => (
          <button
            key={name}
            type="button"
            data-testid={`topbar-nav-${name}`}
            aria-current={activeRoute === name ? "page" : undefined}
            onClick={() => onNavigate({ name, params: {} })}
          >
            {t(NAV_LABEL_KEYS[name])}
          </button>
        ))}
      </nav>
      <div data-testid="topbar-display-controls">
        <label>
          <span className="visually-hidden">{t("display.theme")}</span>
          <select
            data-testid="topbar-theme-select"
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
        <button
          type="button"
          data-testid="topbar-dark-toggle"
          aria-pressed={prefs.dark}
          onClick={() => prefs.setDark(!prefs.dark)}
        >
          {prefs.dark ? t("display.dark") : t("display.light")}
        </button>
        <label>
          <span className="visually-hidden">{t("display.language")}</span>
          <select
            data-testid="topbar-lang-select"
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
      >
        {signedIn ? (
          <nav data-testid="topbar-user-menu" aria-label="user">
            <button
              type="button"
              data-testid="topbar-user-profile"
              onClick={() => onNavigate({ name: "profile", params: {} })}
            >
              {t("user.profile")}
            </button>
            <button
              type="button"
              data-testid="topbar-user-settings"
              onClick={() => onNavigate({ name: "settings", params: {} })}
            >
              {t("user.settings")}
            </button>
            <button
              type="button"
              data-testid="topbar-user-logout"
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
              onClick={() => onNavigate({ name: "auth_login", params: {} })}
            >
              {t("auth.login")}
            </button>
            <button
              type="button"
              data-testid="topbar-register"
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
