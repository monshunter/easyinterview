import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type FC,
  type ReactNode,
} from "react";

import {
  EasyInterviewClient,
  type RequestOptions,
} from "../../api/generated/client";
import type { RuntimeConfig, UserContext } from "../../api/generated/types";

export type RuntimeConfigState =
  | { status: "loading" }
  | { status: "ready"; config: RuntimeConfig }
  | { status: "error"; error: Error };

export type AuthState =
  | { status: "loading" }
  | { status: "authenticated"; user: UserContext }
  | { status: "unauthenticated" }
  | { status: "error"; error: Error };

export interface AppRuntimeValue {
  client: EasyInterviewClient;
  runtime: RuntimeConfigState;
  auth: AuthState;
  /** Force a re-fetch of `/me`, or commit a freshly returned user context. */
  refreshAuth: (user?: UserContext) => Promise<AuthState> | void;
}

export interface AppRuntimeProviderProps {
  client: EasyInterviewClient;
  /**
   * Public auth entry routes should not probe `/me` on first paint: an
   * expected signed-out 401 is harmless to state, but Chrome still reports it
   * as a failed resource and obscures real auth-chain errors during debug.
   */
  skipInitialAuthProbe?: boolean;
  /**
   * Per-operation request options. Tests use this to inject `Prefer:
   * example=<scenario>` headers; production bootstrap leaves this undefined
   * and lets the mock transport / real backend resolve scenarios on its own.
   */
  requestOptions?: {
    runtimeConfig?: RequestOptions;
    getMe?: RequestOptions;
  };
  children: ReactNode;
}

/**
 * React context that carries the runtime/auth/client tuple. Exported so that
 * focused hook tests can mount the context directly with a test client
 * (avoiding the network roundtrips that `<AppRuntimeProvider>` would issue
 * for getRuntimeConfig / getMe). Application code should keep using
 * {@link useAppRuntime} / {@link useAppRuntimeOptional} instead of reading
 * this context directly.
 */
export const AppRuntimeContext = createContext<AppRuntimeValue | null>(null);

function wrapError(error: unknown): Error {
  return error instanceof Error ? error : new Error(String(error));
}

function isUnauthenticatedError(error: Error): boolean {
  return /^HTTP 401\b/.test(error.message);
}

function requestOptionsKey(options: RequestOptions | undefined): string {
  if (!options) return "";
  const headers = Array.from(new Headers(options.headers ?? {}).entries()).sort(
    ([left], [right]) => left.localeCompare(right),
  );
  const query = Object.entries(options.query ?? {})
    .filter(([, value]) => value !== undefined)
    .sort(([left], [right]) => left.localeCompare(right));
  return JSON.stringify([options.idempotencyKey ?? "", headers, query]);
}

export const AppRuntimeProvider: FC<AppRuntimeProviderProps> = ({
  client,
  skipInitialAuthProbe = false,
  requestOptions,
  children,
}) => {
  const [runtime, setRuntime] = useState<RuntimeConfigState>({
    status: "loading",
  });
  const [auth, setAuth] = useState<AuthState>(
    skipInitialAuthProbe
      ? { status: "unauthenticated" }
      : { status: "loading" },
  );
  const [authNonce, setAuthNonce] = useState(0);
  const skippedInitialAuthProbeRef = useRef(false);
  const authRefreshResolversRef = useRef<Array<(state: AuthState) => void>>([]);
  const runtimeConfigOptionsRef = useRef(requestOptions?.runtimeConfig);
  const getMeOptionsRef = useRef(requestOptions?.getMe);
  runtimeConfigOptionsRef.current = requestOptions?.runtimeConfig;
  getMeOptionsRef.current = requestOptions?.getMe;
  const runtimeConfigOptionsKey = requestOptionsKey(requestOptions?.runtimeConfig);
  const getMeOptionsKey = requestOptionsKey(requestOptions?.getMe);
  const runtimeConfigSignal = requestOptions?.runtimeConfig?.signal;
  const getMeSignal = requestOptions?.getMe?.signal;

  useEffect(() => {
    let cancelled = false;

    client
      .getRuntimeConfig(runtimeConfigOptionsRef.current)
      .then((config) => {
        if (!cancelled) setRuntime({ status: "ready", config });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        setRuntime({ status: "error", error: wrapError(error) });
      });

    return () => {
      cancelled = true;
    };
  }, [client, runtimeConfigOptionsKey, runtimeConfigSignal]);

  useEffect(() => {
    let cancelled = false;

    if (
      skipInitialAuthProbe &&
      authNonce === 0 &&
      !skippedInitialAuthProbeRef.current
    ) {
      skippedInitialAuthProbeRef.current = true;
      setAuth({ status: "unauthenticated" });
      return () => {
        cancelled = true;
      };
    }

    setAuth({ status: "loading" });

    client
      .getMe(getMeOptionsRef.current)
      .then((user) => {
        if (cancelled) return;
        const next: AuthState = { status: "authenticated", user };
        setAuth(next);
        authRefreshResolversRef.current.splice(0).forEach((resolve) => resolve(next));
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        const wrapped = wrapError(error);
        // Spec D-3: `/me` only drives the user area. A real 401 is the signed
        // out state; fixture or transport wiring errors must stay visible.
        const next: AuthState = isUnauthenticatedError(wrapped)
          ? { status: "unauthenticated" }
          : { status: "error", error: wrapped };
        setAuth(next);
        authRefreshResolversRef.current.splice(0).forEach((resolve) => resolve(next));
      });

    return () => {
      cancelled = true;
    };
  }, [
    authNonce,
    client,
    getMeOptionsKey,
    getMeSignal,
    skipInitialAuthProbe,
  ]);

  const value = useMemo<AppRuntimeValue>(
    () => ({
      client,
      runtime,
      auth,
      refreshAuth: (user?: UserContext) => {
        skippedInitialAuthProbeRef.current = true;
        if (user) {
          const next: AuthState = { status: "authenticated", user };
          setAuth(next);
          return Promise.resolve(next);
        }
        setAuth({ status: "loading" });
        setAuthNonce((n) => n + 1);
        return new Promise<AuthState>((resolve) => {
          authRefreshResolversRef.current.push(resolve);
        });
      },
    }),
    [client, runtime, auth],
  );

  return (
    <AppRuntimeContext.Provider value={value}>
      {children}
    </AppRuntimeContext.Provider>
  );
};

export function useAppRuntime(): AppRuntimeValue {
  const ctx = useContext(AppRuntimeContext);
  if (!ctx) {
    throw new Error(
      "useAppRuntime must be used inside an <AppRuntimeProvider>",
    );
  }
  return ctx;
}

/**
 * Same as {@link useAppRuntime} but returns `null` when no provider is
 * mounted. App shells that may run with or without runtime bootstrap (early
 * tests, isolated component scenarios) consume this variant to stay
 * provider-optional without re-implementing the hook.
 */
export function useAppRuntimeOptional(): AppRuntimeValue | null {
  return useContext(AppRuntimeContext);
}
