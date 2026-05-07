import type { FC } from "react";

import { useI18n } from "../i18n/messages";
import type { Route } from "../routes";

/**
 * D1 placeholder for the User Profile route. Mirrors
 * docs/ui-design/user-profile-and-settings.md §3 sections without rendering
 * full evidence cards — D2-D6 owners populate the AI-summarized content.
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
    >
      <header>
        <h1>{t("profile.title")}</h1>
        <p>{t("profile.subtitle")}</p>
      </header>
      <div data-testid="profile-identity-summary">
        <h2>{t("profile.identity")}</h2>
        <ul>
          {list("profile.identityItems").map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </div>
      <div data-testid="profile-sections">
        <h2>{t("profile.sections")}</h2>
        <ul>
          {list("profile.sectionItems").map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </div>
      <div data-testid="profile-insight-cards">
        <h2>{t("profile.insights")}</h2>
        <p>{t("profile.insightsBody")}</p>
      </div>
      <div data-testid="profile-used-by">
        <h2>{t("profile.usedBy")}</h2>
        <ul>
          {list("profile.usedByItems").map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </div>
      <div data-testid="profile-recent-evidence">
        <h2>{t("profile.recentEvidence")}</h2>
        <p>{t("profile.recentEvidenceBody")}</p>
      </div>
    </section>
  );
};
