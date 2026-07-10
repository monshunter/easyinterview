import { useState, type FC } from "react";

import { useI18n } from "../i18n/messages";
import type { Route } from "../routes";

/**
 * Settings & Privacy route. Mirrors docs/ui-design/auth-and-entry.md §4 / §8 and
 * ui-design/src/screens-p0-complete.jsx::SettingsScreen.
 *
 * product-scope D-21: settings keeps exactly two tabs — profile and
 * privacy & data. The profile tab holds account basics, sign-in & security
 * (email code · no password, per D-16), font preset, and product info.
 * Notifications / subscription tabs are outside current scope.
 */
type SettingsTab = "profile" | "privacy";

export const SettingsScreen: FC<{ route: Route }> = ({ route }) => {
  const { t, list } = useI18n();
  const [tab, setTab] = useState<SettingsTab>("profile");

  const tabs: Array<{ key: SettingsTab; label: string }> = [
    { key: "profile", label: t("settings.tab.profile") },
    { key: "privacy", label: t("settings.privacy") },
  ];

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
      <nav data-testid="settings-tabs" className="ei-settings-tab-rail">
        {tabs.map(({ key, label }) => (
          <button
            key={key}
            type="button"
            data-testid={`settings-tab-${key}`}
            className="ei-settings-tab"
            data-active={tab === key ? "true" : "false"}
            onClick={() => setTab(key)}
          >
            {label}
          </button>
        ))}
      </nav>
      {tab === "profile" && (
        <>
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
        </>
      )}
      {tab === "privacy" && (
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
      )}
    </section>
  );
};
