import type { FC } from "react";

import { useI18n } from "../i18n/messages";
import type { Route } from "../routes";

/**
 * D1 placeholder for screens whose business surface is owned by D2-D6 follow-on
 * workstreams. Renders only the route name and params via data attributes so
 * route-state tests can assert routing behavior without coupling to screen
 * markup that the follow-on owners will replace.
 *
 * D2 visual contract: emits the same ei-screen-shell + ei-screen-card scaffold
 * as the live screens so D2-D6 owners can grow their content inside an already
 * styled card without re-doing the shell. The card body shows three skeleton
 * stripes derived from `ui-design/src/primitives.jsx::Placeholder`.
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
          {/* Generic body line; D2-D6 owners replace the placeholder body with
              their domain content. */}
          {route.name}
        </p>
        <div className="ei-skeleton-stripe">D2-D6</div>
        <div className="ei-skeleton-line" />
        <div className="ei-skeleton-line" style={{ width: "70%" }} />
      </div>
    </section>
  );
};
