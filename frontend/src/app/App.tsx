import { useEffect, useMemo, type FC, type ReactNode } from "react";

import type {
  EasyInterviewClient,
  RequestOptions,
} from "../api/generated/client";
import {
  AuthLoginScreen,
  AuthLogoutScreen,
  AuthProfileSetupScreen,
  AuthVerifyScreen,
} from "./auth";
import { encodePendingAction } from "./auth/pendingAction";
import { buildResumeRoute } from "./auth/resumeRoute";
import {
  DisplayPreferencesProvider,
  useDisplayPreferences,
  type Lang,
} from "./display/DisplayPreferencesProvider";
import { NavigationProvider } from "./navigation/NavigationProvider";
import { normalizeRoute, type LooseRoute } from "./normalizeRoute";
import {
  isChromeHidden,
  shouldCarryInterviewContext,
  type Route,
  type RouteName,
} from "./routes";
import { resolveInitialRoute, useBrowserRoute } from "./routeStore";
import {
  AppRuntimeProvider,
  useAppRuntimeOptional,
  type AppRuntimeProviderProps,
  type AppRuntimeValue,
} from "./runtime/AppRuntimeProvider";
import { GeneratingScreen } from "./screens/generating/GeneratingScreen";
import { HomeScreen } from "./screens/home/HomeScreen";
import { ParseScreen } from "./screens/parse/ParseScreen";
import { PracticeScreen } from "./screens/practice/PracticeScreen";
import { ReportScreen } from "./screens/report/ReportScreen";
import { ReportConversationScreen } from "./screens/report-conversation/ReportConversationScreen";
import { ReportsScreen } from "./screens/reports/ReportsScreen";
import { ResumeWorkshopScreen } from "./screens/resume-workshop/ResumeWorkshopScreen";
import { RouteShellScreen } from "./screens/RouteShellScreen";
import { SettingsScreen } from "./screens/SettingsScreen";
import { WorkspaceScreen } from "./screens/workspace/WorkspaceScreen";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "./interview-context/InterviewContext";
import { TopBar } from "./topbar/TopBar";

export interface AppProps {
  /**
   * Optional initial route. Accepts loose input (out-of-scope alias names, missing
   * params) and runs it through {@link normalizeRoute} before mounting so out-of-scope
   * URLs / saved state cannot materialize standalone out-of-scope screens. Production
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
  // Settings / Home shells are pure UI and do not depend on runtime; render
  // them whether or not a client is mounted so feature owners can iterate
  // without the auth bootstrap.
  if (route.name === "settings") {
    return <SettingsScreen route={route} />;
  }
  if (route.name === "home") {
    return <HomeScreen route={route} />;
  }
  if (route.name === "parse") {
    return <ParseScreen route={route} />;
  }
  if (route.name === "workspace") {
    return route.params.targetJobId ? (
      <ParseScreen route={route} />
    ) : (
      <WorkspaceScreen route={route} />
    );
  }
  if (route.name === "resume_versions") {
    return <ResumeWorkshopScreen route={route} />;
  }
  if (route.name === "practice") {
    return <PracticeScreen route={route} />;
  }
  if (route.name === "reports") {
    return <ReportsScreen route={route} />;
  }
  if (route.name === "generating") {
    return <GeneratingScreen route={route} />;
  }
  if (route.name === "report") {
    return <ReportScreen route={route} />;
  }
  if (route.name === "report_conversation") {
    return <ReportConversationScreen route={route} />;
  }
  if (!runtime) {
    return <RouteShellScreen route={route} />;
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
    case "auth_verify":
      return (
        <AuthVerifyScreen
          route={route}
          onNavigate={navigate}
          onVerify={async (req) => {
            await runtime.client.verifyAuthEmailChallenge(
              withLocaleHeader(lang, { query: { token: req.token } }),
            );
            try {
              const user = await runtime.client.getMe(withLocaleHeader(lang));
              runtime.refreshAuth(user);
              return {
                profileCompletionRequired: user.profileCompletionRequired,
              };
            } catch {
              runtime.refreshAuth();
              return { profileCompletionRequired: false };
            }
          }}
        />
      );
    case "auth_profile_setup":
      return (
        <AuthProfileSetupScreen
          route={route}
          onNavigate={navigate}
          onCompleteProfile={async (req) => {
            const user = await runtime.client.completeMyProfile(
              req,
              withLocaleHeader(lang),
            );
            runtime.refreshAuth(user);
            return {
              profileCompletionRequired: user.profileCompletionRequired,
            };
          }}
        />
      );
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
      return <RouteShellScreen route={route} />;
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
  const { route, navigate, replaceRoute } = useBrowserRoute({ initialRoute });
  const navigationValue = useMemo(
    () => ({ navigate, replaceRoute }),
    [navigate, replaceRoute],
  );
  const runtime = useAppRuntimeOptional();
  const signedIn = runtime?.auth.status === "authenticated";
  const routeForRender = useMemo<Route>(() => {
    if (
      runtime?.auth.status === "unauthenticated" &&
      route.name === "auth_profile_setup"
    ) {
      return normalizeRoute({
        name: "auth_login",
        params: { ...route.params },
      });
    }
    if (
      runtime?.auth.status === "unauthenticated" &&
      isProtectedRouteName(route.name)
    ) {
      return normalizeRoute(buildAuthLoginRoute(route));
    }
    if (
      runtime?.auth.status === "authenticated" &&
      runtime.auth.user.profileCompletionRequired &&
      route.name !== "auth_profile_setup" &&
      route.name !== "auth_logout"
    ) {
      return normalizeRoute(buildProfileSetupRoute(route));
    }
    return route;
  }, [route, runtime?.auth]);
  const hideChrome = isChromeHidden(routeForRender.name);
  const prefs = useDisplayPreferences();
  const authRouteGate =
    runtime &&
    isProtectedRouteName(route.name) &&
    (runtime.auth.status === "loading" || runtime.auth.status === "error")
      ? runtime.auth.status
      : null;

  useEffect(() => {
    if (runtime?.auth.status === "unauthenticated") {
      if (route.name === "auth_profile_setup") {
        navigate({ name: "auth_login", params: { ...route.params } });
      } else if (isProtectedRouteName(route.name)) {
        navigate(buildAuthLoginRoute(route));
      }
      return;
    }
    if (runtime?.auth.status !== "authenticated") return;
    const { user } = runtime.auth;
    if (user.profileCompletionRequired) {
      if (route.name === "auth_profile_setup" || route.name === "auth_logout")
        return;
      navigate(buildProfileSetupRoute(route));
      return;
    }
    if (route.name === "auth_profile_setup") {
      navigate(buildResumeRoute(route.params));
    }
  }, [navigate, route, runtime?.auth]);

  return (
    <NavigationProvider value={navigationValue}>
      <div data-testid="app-root">
        {hideChrome ? null : (
          <TopBar
            activeRoute={routeForRender.name}
            onNavigate={navigate}
            signedIn={signedIn}
            user={
              runtime?.auth.status === "authenticated"
                ? runtime.auth.user
                : undefined
            }
          />
        )}
        <main>
          <InterviewContextProvider>
            {authRouteGate ? (
              <AuthRouteGate status={authRouteGate} route={route} />
            ) : (
              <>
                <InterviewContextRouteSync route={routeForRender} />
                {renderRouteScreen(
                  routeForRender,
                  navigate,
                  runtime,
                  prefs.lang,
                )}
              </>
            )}
          </InterviewContextProvider>
        </main>
        {children}
      </div>
    </NavigationProvider>
  );
};

function buildAuthLoginRoute(route: Route): LooseRoute {
  return {
    name: "auth_login",
    params: encodePendingAction({
      route: route.name,
      type: "open_protected_route",
      label: route.name,
      params: route.params,
    }),
  };
}

function buildProfileSetupRoute(route: Route): LooseRoute {
  if (isAuthRouteName(route.name)) {
    return { name: "auth_profile_setup", params: { ...route.params } };
  }
  return {
    name: "auth_profile_setup",
    params: encodePendingAction({
      route: route.name,
      type: "complete_profile_resume",
      label: route.name,
      params: route.params,
    }),
  };
}

function isAuthRouteName(name: RouteName): boolean {
  return name.startsWith("auth_");
}

function isProtectedRouteName(name: RouteName): boolean {
  return name !== "home" && !isAuthRouteName(name);
}

const AuthRouteGate: FC<{ status: "loading" | "error"; route: Route }> = ({
  status,
  route,
}) => (
  <section
    data-testid="auth-route-gate"
    data-route-name={route.name}
    className="ei-screen-shell"
    style={{ padding: "48px 56px 96px" }}
  >
    <div
      style={{
        maxWidth: 520,
        background: "var(--ei-color-bg-card)",
        border: "1px solid var(--ei-color-rule-strong)",
        borderRadius: 3,
        padding: 24,
      }}
    >
      <div
        style={{
          color: "var(--ei-color-fg-tertiary)",
          marginBottom: 8,
          fontSize: 11,
          fontWeight: 500,
          letterSpacing: "0.08em",
          textTransform: "uppercase",
          fontFamily: "var(--ei-font-mono)",
        }}
      >
        AUTH
      </div>
      <div
        style={{
          fontSize: 22,
          color: "var(--ei-color-fg-primary)",
          fontFamily: "var(--ei-font-serif)",
          letterSpacing: "-0.02em",
        }}
      >
        {status === "loading" ? "Checking sign-in" : "Sign-in required"}
      </div>
      <p
        style={{
          color: "var(--ei-color-fg-secondary)",
          fontSize: 14,
          lineHeight: 1.55,
          margin: "8px 0 0",
        }}
      >
        {status === "loading"
          ? "Please wait while we verify this session."
          : "Please sign in before opening interview workspace routes."}
      </p>
    </div>
  </section>
);

export const InterviewContextRouteSync: FC<{ route: Route }> = ({ route }) => {
  const { dispatch } = useInterviewContext();

  useEffect(() => {
    if (route.name === "workspace") {
      dispatch({ type: "CLEAR" });
      if (route.params.targetJobId) {
        dispatch({ type: "HYDRATE_FROM_ROUTE", params: route.params });
      }
      return;
    }
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
  skipInitialAuthProbe?: boolean;
  children: ReactNode;
}> = ({ client, requestOptions, skipInitialAuthProbe = false, children }) => {
  const prefs = useDisplayPreferences();
  const localizedRequestOptions = useMemo(
    () => withRuntimeLocaleHeaders(prefs.lang, requestOptions),
    [prefs.lang, requestOptions],
  );
  return (
    <AppRuntimeProvider
      client={client}
      skipInitialAuthProbe={skipInitialAuthProbe}
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
  const skipInitialAuthProbe = useMemo(
    () => shouldSkipInitialAuthProbe(initialRoute),
    [initialRoute],
  );
  const inner = (
    <AppShell initialRoute={initialRoute}>{children}</AppShell>
  );
  return (
    <DisplayPreferencesProvider>
      {client ? (
        <AppRuntimeShell
          client={client}
          requestOptions={requestOptions}
          skipInitialAuthProbe={skipInitialAuthProbe}
        >
          {inner}
        </AppRuntimeShell>
      ) : (
        inner
      )}
    </DisplayPreferencesProvider>
  );
};

function shouldSkipInitialAuthProbe(initialRoute?: LooseRoute): boolean {
  const route = resolveInitialRoute({ initialRoute });
  return route.name === "auth_login" || route.name === "auth_verify";
}
