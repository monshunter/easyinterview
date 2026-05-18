import { useEffect, useMemo, type FC, type ReactNode } from "react";

import type {
  EasyInterviewClient,
  RequestOptions,
} from "../api/generated/client";
import {
  AuthLoginScreen,
  AuthLogoutScreen,
  AuthRegisterScreen,
  AuthResetScreen,
  AuthVerifyScreen,
} from "./auth";
import {
  DisplayPreferencesProvider,
  useDisplayPreferences,
  type Lang,
} from "./display/DisplayPreferencesProvider";
import { NavigationProvider } from "./navigation/NavigationProvider";
import { type LooseRoute } from "./normalizeRoute";
import {
  isChromeHidden,
  shouldCarryInterviewContext,
  type Route,
} from "./routes";
import { useBrowserRoute } from "./routeStore";
import {
  AppRuntimeProvider,
  useAppRuntimeOptional,
  type AppRuntimeProviderProps,
  type AppRuntimeValue,
} from "./runtime/AppRuntimeProvider";
import { DebriefScreen } from "./screens/debrief/DebriefScreen";
import { GeneratingScreen } from "./screens/generating/GeneratingScreen";
import { HomeScreen } from "./screens/home/HomeScreen";
import { JDMatchScreen } from "./screens/jd_match/JDMatchScreen";
import { ParseScreen } from "./screens/parse/ParseScreen";
import { PlaceholderScreen } from "./screens/PlaceholderScreen";
import { PracticeScreen } from "./screens/practice/PracticeScreen";
import { ProfileScreen } from "./screens/ProfileScreen";
import { ReportScreen } from "./screens/report/ReportScreen";
import { ResumeWorkshopScreen } from "./screens/resume-workshop/ResumeWorkshopScreen";
import { SettingsScreen } from "./screens/SettingsScreen";
import { WorkspaceScreen } from "./screens/workspace/WorkspaceScreen";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "./interview-context/InterviewContext";
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
  lang: Lang,
): ReactNode {
  // Profile / Settings / Home shells are pure UI and do not depend on
  // runtime; render them whether or not a client is mounted so D2-D6 owners
  // can iterate without the auth bootstrap.
  if (route.name === "profile") {
    return <ProfileScreen route={route} />;
  }
  if (route.name === "settings") {
    return <SettingsScreen route={route} />;
  }
  if (route.name === "home") {
    return <HomeScreen route={route} />;
  }
  if (route.name === "jd_match") {
    return <JDMatchScreen route={route} />;
  }
  if (route.name === "parse") {
    return <ParseScreen route={route} />;
  }
  if (route.name === "workspace") {
    return <WorkspaceScreen route={route} />;
  }
  if (route.name === "resume_versions") {
    return <ResumeWorkshopScreen route={route} />;
  }
  if (route.name === "practice") {
    return <PracticeScreen route={route} />;
  }
  if (route.name === "generating") {
    return <GeneratingScreen route={route} />;
  }
  if (route.name === "report") {
    return <ReportScreen route={route} />;
  }
  if (route.name === "debrief") {
    return <DebriefScreen route={route} />;
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
            await runtime.client.startAuthEmailChallenge(
              req,
              withLocaleHeader(lang),
            );
          }}
        />
      );
    case "auth_register":
      return (
        <AuthRegisterScreen
          route={route}
          onNavigate={navigate}
          onStartChallenge={async (req) => {
            await runtime.client.startAuthEmailChallenge(
              req,
              withLocaleHeader(lang),
            );
          }}
        />
      );
    case "auth_verify":
      return (
        <AuthVerifyScreen
          route={route}
          onNavigate={navigate}
          onVerify={async (req) => {
            await runtime.client.verifyAuthEmailChallenge(
              withLocaleHeader(lang, { query: { token: req.token } }),
            );
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
            await runtime.client.logout(withLocaleHeader(lang));
            runtime.refreshAuth();
          }}
        />
      );
    default:
      return <PlaceholderScreen route={route} />;
  }
}

function withLocaleHeader(lang: Lang, opts?: RequestOptions): RequestOptions {
  return {
    ...opts,
    headers: {
      ...(opts?.headers ?? {}),
      "Accept-Language": lang,
    },
  };
}

function withRuntimeLocaleHeaders(
  lang: Lang,
  requestOptions?: AppRuntimeProviderProps["requestOptions"],
): AppRuntimeProviderProps["requestOptions"] {
  return {
    runtimeConfig: withLocaleHeader(lang, requestOptions?.runtimeConfig),
    getMe: withLocaleHeader(lang, requestOptions?.getMe),
  };
}

const AppShell: FC<Pick<AppProps, "initialRoute" | "children">> = ({
  initialRoute,
  children,
}) => {
  const { route, navigate } = useBrowserRoute({ initialRoute });
  const navigationValue = useMemo(() => ({ navigate }), [navigate]);
  const hideChrome = isChromeHidden(route.name);
  const runtime = useAppRuntimeOptional();
  const signedIn = runtime?.auth.status === "authenticated";
  const prefs = useDisplayPreferences();

  return (
    <NavigationProvider value={navigationValue}>
      <div data-testid="app-root">
        {hideChrome ? null : (
          <TopBar
            activeRoute={route.name}
            onNavigate={navigate}
            signedIn={signedIn}
            user={runtime?.auth.status === "authenticated" ? runtime.auth.user : undefined}
          />
        )}
        <main>
          <InterviewContextProvider>
            <InterviewContextRouteSync route={route} />
            {renderRouteScreen(route, navigate, runtime, prefs.lang)}
          </InterviewContextProvider>
        </main>
        {children}
      </div>
    </NavigationProvider>
  );
};

const InterviewContextRouteSync: FC<{ route: Route }> = ({ route }) => {
  const { dispatch } = useInterviewContext();

  useEffect(() => {
    if (shouldCarryInterviewContext(route.name)) {
      dispatch({ type: "HYDRATE_FROM_ROUTE", params: route.params });
      return;
    }
    dispatch({ type: "CLEAR" });
  }, [dispatch, route.name, route.params]);

  return null;
};

const AppRuntimeShell: FC<{
  client: EasyInterviewClient;
  requestOptions?: AppRuntimeProviderProps["requestOptions"];
  children: ReactNode;
}> = ({ client, requestOptions, children }) => {
  const prefs = useDisplayPreferences();
  const localizedRequestOptions = useMemo(
    () => withRuntimeLocaleHeaders(prefs.lang, requestOptions),
    [prefs.lang, requestOptions],
  );
  return (
    <AppRuntimeProvider
      client={client}
      requestOptions={localizedRequestOptions}
    >
      {children}
    </AppRuntimeProvider>
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
        <AppRuntimeShell client={client} requestOptions={requestOptions}>
          {inner}
        </AppRuntimeShell>
      ) : (
        inner
      )}
    </DisplayPreferencesProvider>
  );
};
