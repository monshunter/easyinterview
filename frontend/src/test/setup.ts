import "@testing-library/jest-dom/vitest";
import { afterEach, beforeEach } from "vitest";

// Reset jsdom-only window state between tests so the URL-addressable route
// store (see frontend/src/app/routeStore.ts) does not leak browser history
// across files. Guards against tests that navigate to `/practice` (chrome
// hidden) being followed by tests that expect the default home shell.
beforeEach(() => {
  if (typeof window !== "undefined") {
    const w = window as { __EASYINTERVIEW_INITIAL_ROUTE__?: unknown };
    delete w.__EASYINTERVIEW_INITIAL_ROUTE__;
    if (window.history?.replaceState) {
      window.history.replaceState(null, "", "/");
    }
  }
});

afterEach(() => {
  if (typeof window !== "undefined") {
    const w = window as { __EASYINTERVIEW_INITIAL_ROUTE__?: unknown };
    delete w.__EASYINTERVIEW_INITIAL_ROUTE__;
    if (window.history?.replaceState) {
      window.history.replaceState(null, "", "/");
    }
  }
});
