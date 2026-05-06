import {
  useCallback,
  useMemo,
  useState,
  type FC,
  type ReactNode,
} from "react";

import type { EasyInterviewClient } from "../api/generated/client";
import {
  AuthLoginScreen,
  AuthLogoutScreen,
  AuthRegisterScreen,
  AuthResetScreen,
  AuthVerifyScreen,
} from "./auth";
import { DisplayPreferencesProvider } from "./display/DisplayPreferencesProvider";
import { NavigationProvider } from "./navigation/NavigationProvider";
import { normalizeRoute, type LooseRoute } from "./normalizeRoute";
import { DEFAULT_ROUTE, isChromeHidden, type Route } from "./routes";
import {
  AppRuntimeProvider,
  useAppRuntimeOptional,
  type AppRuntimeProviderProps,
  type AppRuntimeValue,
} from "./runtime/AppRuntimeProvider";
import { PlaceholderScreen } from "./screens/PlaceholderScreen";
import { ProfileScreen } from "./screens/ProfileScreen";
import { SettingsScreen } from "./screens/SettingsScreen";
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
  /**
   * Optional probe / harness rendered after the routed screen, inside the
   * navigation + runtime context. Production callers leave this undefined;
   * route-state tests pass a probe that uses {@link useRequestAuth} to
   * trigger pending actions (`立即面试`, `复练当前轮`, etc.).
   */
  children?: ReactNode;
}

function renderRouteScreen(
  route: Route,
  navigate: (next: LooseRoute) => void,
  runtime: AppRuntimeValue | null,
): ReactNode {
  // Profile / Settings shells are pure UI and do not depend on runtime; render
  // them whether or not a client is mounted so D2-D6 owners can iterate
  // without the auth bootstrap.
  if (route.name === "profile") {
    return <ProfileScreen route={route} />;
  }
  if (route.name === "settings") {
    return <SettingsScreen route={route} />;
  }
  if (!runtime) {
    return <PlaceholderScreen route={route} />;
  }
  switch (route.name) {
    case "auth_login":
      return (
        <AuthLoginScreen
          route={route}
          onNavigate={navigate}
          onStartChallenge={async (req) => {
            await runtime.client.startAuthEmailChallenge(req);
          }}
        />
      );
    case "auth_register":
      return (
        <AuthRegisterScreen
          route={route}
          onNavigate={navigate}
          onStartChallenge={async (req) => {
            await runtime.client.startAuthEmailChallenge(req);
          }}
        />
      );
    case "auth_verify":
      return (
        <AuthVerifyScreen
          route={route}
          onNavigate={navigate}
          onVerify={async () => {
            await runtime.client.verifyAuthEmailChallenge();
            runtime.refreshAuth();
          }}
        />
      );
    case "auth_reset":
      return <AuthResetScreen route={route} onNavigate={navigate} />;
    case "auth_logout":
      return (
        <AuthLogoutScreen
          route={route}
          onNavigate={navigate}
          onLogout={async () => {
            await runtime.client.logout();
            runtime.refreshAuth();
          }}
        />
      );
    default:
      return <PlaceholderScreen route={route} />;
  }
}

const AppShell: FC<Pick<AppProps, "initialRoute" | "children">> = ({
  initialRoute,
  children,
}) => {
  const [route, setRoute] = useState<Route>(() =>
    initialRoute ? normalizeRoute(initialRoute) : DEFAULT_ROUTE,
  );
  const navigate = useCallback((next: LooseRoute) => {
    setRoute(normalizeRoute(next));
  }, []);
  const navigationValue = useMemo(() => ({ navigate }), [navigate]);
  const hideChrome = isChromeHidden(route.name);
  const runtime = useAppRuntimeOptional();
  const signedIn = runtime?.auth.status === "authenticated";

  return (
    <NavigationProvider value={navigationValue}>
      <div data-testid="app-root">
        {hideChrome ? null : (
          <TopBar
            activeRoute={route.name}
            onNavigate={navigate}
            signedIn={signedIn}
          />
        )}
        <main>{renderRouteScreen(route, navigate, runtime)}</main>
        {children}
      </div>
    </NavigationProvider>
  );
};

export const App: FC<AppProps> = ({
  initialRoute,
  client,
  requestOptions,
  children,
}) => {
  const inner = (
    <AppShell initialRoute={initialRoute}>{children}</AppShell>
  );
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
