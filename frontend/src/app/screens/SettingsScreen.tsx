import type { FC } from "react";

import { useI18n } from "../i18n/messages";
import type { Route } from "../routes";

/**
 * D1 placeholder for the Settings & Privacy route. Mirrors
 * docs/ui-design/user-profile-and-settings.md §5.
 *
 * Settings holds account basics, login security, font preset, and the
 * privacy / data control pane. It does NOT restore Growth / Experiences /
 * Mistakes / Drill modules or any job-target / skill-tag metadata.
 *
 * Notifications + Subscription are P1 placeholders.
 *
 * D2 visual contract: shell uses ei-screen-shell + ei-screen-card cadence to
 * match the ui-design profile page rhythm.
 */
export const SettingsScreen: FC<{ route: Route }> = ({ route }) => {
  const { t, list } = useI18n();
  return (
    <section
      data-testid={`route-${route.name}`}
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
      className="ei-screen-shell"
    >
      <header>
        <h1 className="ei-text-display">{t("settings.title")}</h1>
      </header>
      <div data-testid="settings-account" className="ei-screen-card">
        <h2 className="ei-text-title">{t("settings.account")}</h2>
        <ul>
          {list("settings.accountItems").map((item) => (
            <li key={item} className="ei-text-body">
              {item}
            </li>
          ))}
        </ul>
      </div>
      <div data-testid="settings-login-security" className="ei-screen-card">
        <h2 className="ei-text-title">{t("settings.security")}</h2>
        <ul>
          {list("settings.securityItems").map((item) => (
            <li key={item} className="ei-text-body">
              {item}
            </li>
          ))}
        </ul>
      </div>
      <div data-testid="settings-font-preset" className="ei-screen-card">
        <h2 className="ei-text-title">{t("settings.fontPreset")}</h2>
        <ul>
          {list("settings.fontPresetItems").map((item) => (
            <li key={item} className="ei-text-body">
              {item}
            </li>
          ))}
        </ul>
      </div>
      <div data-testid="settings-privacy" className="ei-screen-card">
        <h2 className="ei-text-title">{t("settings.privacy")}</h2>
        <ul>
          {list("settings.privacyItems").map((item) => (
            <li key={item} className="ei-text-body">
              {item}
            </li>
          ))}
        </ul>
      </div>
      <div
        data-testid="settings-notifications-placeholder"
        className="ei-screen-card"
      >
        <h2 className="ei-text-title">{t("settings.notifications")}</h2>
        <div className="ei-skeleton-stripe">P1</div>
      </div>
      <div
        data-testid="settings-subscription-placeholder"
        className="ei-screen-card"
      >
        <h2 className="ei-text-title">{t("settings.subscription")}</h2>
        <div className="ei-skeleton-stripe">P1</div>
      </div>
      <div data-testid="settings-app-info" className="ei-screen-card">
        <h2 className="ei-text-title">{t("settings.appInfo")}</h2>
        <ul>
          {list("settings.appInfoItems").map((item) => (
            <li key={item} className="ei-text-body">
              {item}
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
};
