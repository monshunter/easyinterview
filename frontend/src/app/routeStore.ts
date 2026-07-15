/**
 * Browser-aware route store for the formal frontend App shell.
 *
 * Truth source: docs/spec/frontend-shell/spec.md §4 / C-11 / C-13,
 * docs/spec/frontend-shell/plans/004-url-addressable-routing/plan.md §6
 * Phase 2 + Phase 4.
 *
 * Owns initial route bootstrap, History `pushState` / `replaceState`,
 * `popstate` subscription and a small URL equality cache so the App can
 * avoid double-push when navigation lands on the same canonical URL.
 *
 * Initial route priority (plan §6.2.1):
 *   1. Explicit `initialRoute` prop (test override, also used by tests in
 *      App.test.tsx that pre-stage a route without touching window.location).
 *   2. `window.__EASYINTERVIEW_INITIAL_ROUTE__` test harness override.
 *   3. Canonical path + query parsed via {@link parseUrlToRoute}.
 *   4. `DEFAULT_ROUTE` (home).
 */
import { useCallback, useEffect, useRef, useState } from "react";

import { normalizeRoute, type LooseRoute } from "./normalizeRoute";
import { DEFAULT_ROUTE, type Route } from "./routes";
import { formatRouteUrl, parseUrlToRoute } from "./routeUrl";

declare global {
  interface Window {
    __EASYINTERVIEW_INITIAL_ROUTE__?: LooseRoute;
  }
}

export interface ResolveInitialRouteOptions {
  initialRoute?: LooseRoute;
  windowRef?: Window;
}

function getWindow(opts: ResolveInitialRouteOptions): Window | undefined {
  if (opts.windowRef) return opts.windowRef;
  if (typeof window !== "undefined") return window;
  return undefined;
}

function isMeaningfulCanonicalLocation(loc: Location): boolean {
  const path = loc.pathname || "";
  const search = loc.search || "";
  return (path && path !== "/") || search !== "";
}

export function resolveInitialRoute(
  opts: ResolveInitialRouteOptions = {},
): Route {
  if (opts.initialRoute) return normalizeRoute(opts.initialRoute);
  const w = getWindow(opts);
  if (!w) return DEFAULT_ROUTE;
  if (w.__EASYINTERVIEW_INITIAL_ROUTE__) {
    return normalizeRoute(w.__EASYINTERVIEW_INITIAL_ROUTE__);
  }
  const loc = w.location;
  if (loc && isMeaningfulCanonicalLocation(loc)) {
    return parseUrlToRoute(`${loc.pathname || "/"}${loc.search || ""}`);
  }
  return DEFAULT_ROUTE;
}

export interface UseBrowserRouteOptions {
  /** Optional bootstrap override; takes priority over window.location. */
  initialRoute?: LooseRoute;
  /** Optional window reference; defaults to global window for tests. */
  windowRef?: Window;
  /**
   * When true (default in production), the store keeps the browser URL in
   * sync with the active route via `pushState` / `replaceState`. When false
   * (default when an explicit `initialRoute` is passed, typical for unit
   * tests using App harness), the store still listens to `popstate` events
   * but does not rewrite `window.location` so unrelated tests do not leak
   * URL state across files.
   */
  syncUrl?: boolean;
}

export interface BrowserRouteApi {
  route: Route;
  /** Push a new history entry for `next` (no-op if URL is unchanged). */
  navigate: (next: LooseRoute) => void;
  /** Replace the current history entry with `next`. */
  replaceRoute: (next: LooseRoute) => void;
}

function sameParams(
  a: Record<string, string>,
  b: Record<string, string>,
): boolean {
  const aKeys = Object.keys(a);
  const bKeys = Object.keys(b);
  if (aKeys.length !== bKeys.length) return false;
  for (const k of aKeys) {
    if (a[k] !== b[k]) return false;
  }
  return true;
}

export function useBrowserRoute(
  opts: UseBrowserRouteOptions = {},
): BrowserRouteApi {
  const windowRef = opts.windowRef ?? (typeof window !== "undefined" ? window : undefined);
  const syncUrl = opts.syncUrl ?? opts.initialRoute === undefined;

  // The route store always holds a canonical Route (= URL round-tripped
  // through `parseUrlToRoute`) so route.params can never carry raw payload
  // markers even if a caller passes them. This is part of the Plan 004
  // privacy redline (history.state / React state must not leak raw text).
  const [route, setRoute] = useState<Route>(() =>
    canonicalize(resolveInitialRoute({
      initialRoute: opts.initialRoute,
      windowRef,
    })),
  );

  const lastUrlRef = useRef<string | null>(null);
  if (lastUrlRef.current === null) {
    lastUrlRef.current = formatRouteUrl(route);
  }

  // On mount, write canonical URL back to the browser when sync is enabled.
  // This removes unsafe params that survived the initial GET and strips any
  // leftover fragment (canonical addresses are path + query only).
  useEffect(() => {
    if (!syncUrl) return;
    if (!windowRef?.history || !windowRef?.location) return;
    const canonicalUrl = formatRouteUrl(route);
    // A child screen can issue a fail-closed replace during its mount effect
    // before this parent mount effect runs. Do not overwrite that newer route
    // with the stale bootstrap URL.
    if (lastUrlRef.current !== canonicalUrl) return;
    const currentUrl = `${windowRef.location.pathname || ""}${windowRef.location.search || ""}${windowRef.location.hash || ""}`;
    if (currentUrl !== canonicalUrl) {
      windowRef.history.replaceState(null, "", canonicalUrl);
    }
    lastUrlRef.current = canonicalUrl;
    // The effect intentionally runs only on mount; subsequent navigate /
    // replaceRoute / popstate updates manage history directly.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Subscribe to popstate so back/forward updates React state.
  useEffect(() => {
    if (!windowRef?.addEventListener) return;
    const handler = (): void => {
      const next = parseUrlToRoute(
        `${windowRef.location?.pathname || "/"}${windowRef.location?.search || ""}`,
      );
      const canonicalUrl = formatRouteUrl(next);
      const currentUrl = `${windowRef.location?.pathname || ""}${windowRef.location?.search || ""}${windowRef.location?.hash || ""}`;
      if (windowRef.history && currentUrl !== canonicalUrl) {
        windowRef.history.replaceState(null, "", canonicalUrl);
      }
      lastUrlRef.current = canonicalUrl;
      setRoute(next);
    };
    windowRef.addEventListener("popstate", handler);
    return () => windowRef.removeEventListener("popstate", handler);
  }, [windowRef]);

  const navigate = useCallback(
    (next: LooseRoute) => {
      const normalized = normalizeRoute(next);
      const nextUrl = formatRouteUrl(normalized);
      const canonical = parseUrlToRoute(nextUrl);
      if (syncUrl && windowRef?.history) {
        if (lastUrlRef.current !== nextUrl) {
          windowRef.history.pushState(null, "", nextUrl);
        }
      }
      lastUrlRef.current = nextUrl;
      setRoute((prev) =>
        prev.name === canonical.name && sameParams(prev.params, canonical.params)
          ? prev
          : canonical,
      );
    },
    [syncUrl, windowRef],
  );

  const replaceRoute = useCallback(
    (next: LooseRoute) => {
      const normalized = normalizeRoute(next);
      const nextUrl = formatRouteUrl(normalized);
      const canonical = parseUrlToRoute(nextUrl);
      if (syncUrl && windowRef?.history) {
        windowRef.history.replaceState(null, "", nextUrl);
      }
      lastUrlRef.current = nextUrl;
      setRoute((prev) =>
        prev.name === canonical.name && sameParams(prev.params, canonical.params)
          ? prev
          : canonical,
      );
    },
    [syncUrl, windowRef],
  );

  return { route, navigate, replaceRoute };
}

/** Round-trip a Route through the canonical URL codec to strip unsafe params. */
function canonicalize(route: Route): Route {
  return parseUrlToRoute(formatRouteUrl(route));
}
