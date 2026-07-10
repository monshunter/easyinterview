import type { FC } from "react";

import { useI18n } from "../i18n/messages";
import type { Route } from "../routes";

/**
 * Generic route shell for unsupported route states. It renders route name /
 * params via data attributes without coupling those states to a business screen
 * implementation.
 *
 * Visual contract: emits the same ei-screen-shell + ei-screen-card scaffold as
 * live screens, with skeleton stripes from the shared screen CSS primitives.
 */
export const RouteShellScreen: FC<{ route: Route }> = ({ route }) => {
  const { t } = useI18n();
  const title =
    route.name === "workspace"
      ? t("routeShell.workspace")
      : t("routeShell.default");
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
      <div className="ei-screen-card ei-screen-card--route-shell">
        <p className="ei-text-body">
          {/* Generic body line for route-state assertions. */}
          {route.name}
        </p>
        <div className="ei-skeleton-stripe">route shell</div>
        <div className="ei-skeleton-line" />
        <div className="ei-skeleton-line" style={{ width: "70%" }} />
      </div>
    </section>
  );
};
