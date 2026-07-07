import type { FC } from "react";

import { useI18n } from "../i18n/messages";
import type { Route } from "../routes";

/**
 * Generic fallback shell for retained route-state tests and unsupported route
 * states. It renders route name / params via data attributes without coupling
 * those tests to a business screen implementation.
 *
 * Visual contract: emits the same ei-screen-shell + ei-screen-card scaffold as
 * live screens, with skeleton stripes derived from
 * `ui-design/src/primitives.jsx::Placeholder`.
 */
export const PlaceholderScreen: FC<{ route: Route }> = ({ route }) => {
  const { t } = useI18n();
  const title =
    route.name === "workspace"
      ? t("placeholder.workspace")
      : t("placeholder.default");
  return (
    <section
      data-testid={`route-${route.name}`}
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
      className="ei-screen-shell"
    >
      <header>
        <h1 className="ei-text-display">{title}</h1>
      </header>
      <div className="ei-screen-card ei-screen-card--placeholder">
        <p className="ei-text-body">
          {/* Generic body line for route-state assertions. */}
          {route.name}
        </p>
        <div className="ei-skeleton-stripe">fallback shell</div>
        <div className="ei-skeleton-line" />
        <div className="ei-skeleton-line" style={{ width: "70%" }} />
      </div>
    </section>
  );
};
