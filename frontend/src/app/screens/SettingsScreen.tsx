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
 */
export const SettingsScreen: FC<{ route: Route }> = ({ route }) => {
  const { t, list } = useI18n();
  return (
    <section
      data-testid={`route-${route.name}`}
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
    >
      <header>
        <h1>{t("settings.title")}</h1>
      </header>
      <div data-testid="settings-account">
        <h2>{t("settings.account")}</h2>
        <ul>
          {list("settings.accountItems").map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </div>
      <div data-testid="settings-login-security">
        <h2>{t("settings.security")}</h2>
        <ul>
          {list("settings.securityItems").map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </div>
      <div data-testid="settings-font-preset">
        <h2>{t("settings.fontPreset")}</h2>
        <ul>
          {list("settings.fontPresetItems").map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </div>
      <div data-testid="settings-privacy">
        <h2>{t("settings.privacy")}</h2>
        <ul>
          {list("settings.privacyItems").map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </div>
      <div data-testid="settings-notifications-placeholder">
        <h2>{t("settings.notifications")}</h2>
      </div>
      <div data-testid="settings-subscription-placeholder">
        <h2>{t("settings.subscription")}</h2>
      </div>
    </section>
  );
};
