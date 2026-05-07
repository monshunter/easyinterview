import type { FC } from "react";

import { useI18n } from "../i18n/messages";
import type { Route } from "../routes";

/**
 * D1 placeholder for the User Profile route. Mirrors
 * docs/ui-design/user-profile-and-settings.md §3 sections without rendering
 * full evidence cards — D2-D6 owners populate the AI-summarized content.
 *
 * D2 visual contract: shell uses ei-screen-shell + ei-screen-card cadence so
 * the placeholder hands off the visual scaffold for D2-D6 to extend.
 *
 * Critically: this shell never restores the retired Growth / Experiences /
 * Mistakes / Drill modules. The settings page handles字体预设; the profile
 * page does NOT carry job preferences, target role, or skill tags.
 */
export const ProfileScreen: FC<{ route: Route }> = ({ route }) => {
  const { t, list } = useI18n();
  return (
    <section
      data-testid={`route-${route.name}`}
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
      className="ei-screen-shell"
    >
      <header>
        <h1 className="ei-text-display">{t("profile.title")}</h1>
        <p className="ei-text-body">{t("profile.subtitle")}</p>
      </header>
      <div data-testid="profile-identity-summary" className="ei-screen-card">
        <h2 className="ei-text-title">{t("profile.identity")}</h2>
        <ul>
          {list("profile.identityItems").map((item) => (
            <li key={item} className="ei-text-body">
              {item}
            </li>
          ))}
        </ul>
      </div>
      <div data-testid="profile-sections" className="ei-screen-card">
        <h2 className="ei-text-title">{t("profile.sections")}</h2>
        <ul>
          {list("profile.sectionItems").map((item) => (
            <li key={item} className="ei-text-body">
              {item}
            </li>
          ))}
        </ul>
      </div>
      <div data-testid="profile-insight-cards" className="ei-screen-card">
        <h2 className="ei-text-title">{t("profile.insights")}</h2>
        <p className="ei-text-body">{t("profile.insightsBody")}</p>
        <div className="ei-skeleton-stripe">D2-D6</div>
      </div>
      <div data-testid="profile-used-by" className="ei-screen-card">
        <h2 className="ei-text-title">{t("profile.usedBy")}</h2>
        <ul>
          {list("profile.usedByItems").map((item) => (
            <li key={item} className="ei-text-body">
              {item}
            </li>
          ))}
        </ul>
      </div>
      <div data-testid="profile-recent-evidence" className="ei-screen-card">
        <h2 className="ei-text-title">{t("profile.recentEvidence")}</h2>
        <p className="ei-text-body">{t("profile.recentEvidenceBody")}</p>
        <div className="ei-skeleton-stripe">D2-D6</div>
      </div>
    </section>
  );
};
