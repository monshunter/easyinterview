import type { FC } from "react";

import { useI18n } from "../i18n/messages";
import type { Route } from "../routes";

/**
 * D1 placeholder for screens whose business surface is owned by D2-D6 follow-on
 * workstreams. Renders only the route name and params via data attributes so
 * route-state tests can assert routing behavior without coupling to screen
 * markup that the follow-on owners will replace.
 */
export const PlaceholderScreen: FC<{ route: Route }> = ({ route }) => {
  const { t } = useI18n();
  const title =
    route.name === "workspace" ? t("placeholder.workspace") : t("placeholder.default");
  return (
    <section
      data-testid={`route-${route.name}`}
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
    >
      <h1>{title}</h1>
    </section>
  );
};
