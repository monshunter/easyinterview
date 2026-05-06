import { useCallback, useState, type FC } from "react";

import type { EasyInterviewClient } from "../api/generated/client";
import { DisplayPreferencesProvider } from "./display/DisplayPreferencesProvider";
import { normalizeRoute, type LooseRoute } from "./normalizeRoute";
import { DEFAULT_ROUTE, isChromeHidden, type Route } from "./routes";
import {
  AppRuntimeProvider,
  useAppRuntimeOptional,
  type AppRuntimeProviderProps,
} from "./runtime/AppRuntimeProvider";
import { PlaceholderScreen } from "./screens/PlaceholderScreen";
import { TopBar } from "./topbar/TopBar";

export interface AppProps {
  /**
   * Optional initial route. Accepts loose input (legacy alias names, missing
   * params) and runs it through {@link normalizeRoute} before mounting so old
   * URLs / saved state cannot materialize standalone legacy screens. Production
   * bootstrap (Phase 1.3) wires this from URL hash + saved state and falls back
   * to {@link DEFAULT_ROUTE}.
   */
  initialRoute?: LooseRoute;
  /**
   * Optional generated API client. When provided, App mounts an
   * {@link AppRuntimeProvider} and the TopBar derives signed-in state from
   * `/me`. Tests pass a fixture-backed client; production bootstrap will pass
   * a client built from `globalThis.fetch`.
   */
  client?: EasyInterviewClient;
  /** Per-operation request options, forwarded to the runtime provider. */
  requestOptions?: AppRuntimeProviderProps["requestOptions"];
}

const AppShell: FC<Pick<AppProps, "initialRoute">> = ({ initialRoute }) => {
  const [route, setRoute] = useState<Route>(() =>
    initialRoute ? normalizeRoute(initialRoute) : DEFAULT_ROUTE,
  );
  const navigate = useCallback((next: LooseRoute) => {
    setRoute(normalizeRoute(next));
  }, []);
  const hideChrome = isChromeHidden(route.name);
  const runtime = useAppRuntimeOptional();
  const signedIn = runtime?.auth.status === "authenticated";

  return (
    <div data-testid="app-root">
      {hideChrome ? null : (
        <TopBar
          activeRoute={route.name}
          onNavigate={navigate}
          signedIn={signedIn}
        />
      )}
      <main>
        <PlaceholderScreen route={route} />
      </main>
    </div>
  );
};

export const App: FC<AppProps> = ({ initialRoute, client, requestOptions }) => {
  const inner = <AppShell initialRoute={initialRoute} />;
  return (
    <DisplayPreferencesProvider>
      {client ? (
        <AppRuntimeProvider client={client} requestOptions={requestOptions}>
          {inner}
        </AppRuntimeProvider>
      ) : (
        inner
      )}
    </DisplayPreferencesProvider>
  );
};
