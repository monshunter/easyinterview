import {
  createContext,
  useContext,
  useEffect,
  useMemo,
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
  /** Force a re-fetch of `/me`. Used by auth screens after verify / logout. */
  refreshAuth: () => void;
}

export interface AppRuntimeProviderProps {
  client: EasyInterviewClient;
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
 * focused hook tests can mount the context directly with a stubbed client
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

export const AppRuntimeProvider: FC<AppRuntimeProviderProps> = ({
  client,
  requestOptions,
  children,
}) => {
  const [runtime, setRuntime] = useState<RuntimeConfigState>({
    status: "loading",
  });
  const [auth, setAuth] = useState<AuthState>({ status: "loading" });
  const [authNonce, setAuthNonce] = useState(0);

  useEffect(() => {
    let cancelled = false;

    client
      .getRuntimeConfig(requestOptions?.runtimeConfig)
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
  }, [client, requestOptions]);

  useEffect(() => {
    let cancelled = false;
    setAuth({ status: "loading" });

    client
      .getMe(requestOptions?.getMe)
      .then((user) => {
        if (!cancelled) setAuth({ status: "authenticated", user });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        const wrapped = wrapError(error);
        // Spec D-3: `/me` only drives the user area. A real 401 is the signed
        // out state; fixture or transport wiring errors must stay visible.
        setAuth(
          isUnauthenticatedError(wrapped)
            ? { status: "unauthenticated" }
            : { status: "error", error: wrapped },
        );
      });

    return () => {
      cancelled = true;
    };
  }, [client, requestOptions, authNonce]);

  const value = useMemo<AppRuntimeValue>(
    () => ({
      client,
      runtime,
      auth,
      refreshAuth: () => setAuthNonce((n) => n + 1),
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
