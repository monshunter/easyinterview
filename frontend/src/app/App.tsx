import { useState, type FC } from "react";

import { normalizeRoute, type LooseRoute } from "./normalizeRoute";
import { DEFAULT_ROUTE, isChromeHidden, type Route } from "./routes";
import { PlaceholderScreen } from "./screens/PlaceholderScreen";

export interface AppProps {
  /**
   * Optional initial route. Accepts loose input (legacy alias names, missing
   * params) and runs it through {@link normalizeRoute} before mounting so old
   * URLs / saved state cannot materialize standalone legacy screens. Production
   * bootstrap (Phase 1.3) wires this from URL hash + saved state and falls back
   * to {@link DEFAULT_ROUTE}.
   */
  initialRoute?: LooseRoute;
}

export const App: FC<AppProps> = ({ initialRoute }) => {
  const [route] = useState<Route>(() =>
    initialRoute ? normalizeRoute(initialRoute) : DEFAULT_ROUTE,
  );
  const hideChrome = isChromeHidden(route.name);

  return (
    <div data-testid="app-root">
      {hideChrome ? null : <header data-testid="app-shell-topbar" />}
      <main>
        <PlaceholderScreen route={route} />
      </main>
    </div>
  );
};
