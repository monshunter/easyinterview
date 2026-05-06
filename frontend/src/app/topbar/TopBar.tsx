import type { FC } from "react";

import {
  useDisplayPreferences,
  type Lang,
  type Theme,
} from "../display/DisplayPreferencesProvider";
import type { LooseRoute } from "../normalizeRoute";
import { PRIMARY_NAV_ROUTES, type RouteName } from "../routes";

/**
 * Five primary nav entries match docs/spec/frontend-shell/spec.md §2.1 and
 * docs/ui-design/auth-and-entry.md §4. Reports / company-intel / auth /
 * profile / settings are intentionally NOT promoted to first-level nav.
 *
 * Labels are user-facing Chinese; English keys remain the canonical RouteName
 * so route-state tests and URL hashes do not depend on i18n state.
 */
const NAV_LABELS: Record<(typeof PRIMARY_NAV_ROUTES)[number], string> = {
  home: "首页",
  jd_match: "岗位推荐",
  workspace: "面试规划",
  resume_versions: "简历版本",
  debrief: "复盘",
};

const THEME_LABELS: Record<Theme, string> = {
  warm: "暖陶",
  forest: "苔林",
  ocean: "深海",
  plum: "梅子",
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
            {NAV_LABELS[name]}
          </button>
        ))}
      </nav>
      <div data-testid="topbar-display-controls">
        <label>
          <span className="visually-hidden">主题</span>
          <select
            data-testid="topbar-theme-select"
            value={prefs.theme}
            onChange={(e) => prefs.setTheme(e.target.value as Theme)}
          >
            {THEME_OPTIONS.map((theme) => (
              <option key={theme} value={theme}>
                {THEME_LABELS[theme]}
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
          {prefs.dark ? "暗色" : "亮色"}
        </button>
        <label>
          <span className="visually-hidden">语言</span>
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
          // Phase 4.1 expands this into the avatar menu with 用户画像 /
          // 设置与隐私 / 退出登录.
          <span data-testid="topbar-user-menu-placeholder" />
        ) : (
          <>
            <button
              type="button"
              data-testid="topbar-login"
              onClick={() => onNavigate({ name: "auth_login", params: {} })}
            >
              登录
            </button>
            <button
              type="button"
              data-testid="topbar-register"
              onClick={() => onNavigate({ name: "auth_register", params: {} })}
            >
              注册
            </button>
          </>
        )}
      </div>
    </header>
  );
};
