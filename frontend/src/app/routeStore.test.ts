// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import { resolveInitialRoute } from "./routeStore";

function setLocation(url: string): void {
  window.history.replaceState(null, "", url);
}

beforeEach(() => {
  delete (window as { __EASYINTERVIEW_INITIAL_ROUTE__?: unknown })
    .__EASYINTERVIEW_INITIAL_ROUTE__;
  setLocation("/");
});

afterEach(() => {
  delete (window as { __EASYINTERVIEW_INITIAL_ROUTE__?: unknown })
    .__EASYINTERVIEW_INITIAL_ROUTE__;
  setLocation("/");
});

describe("resolveInitialRoute priority", () => {
  it("honors explicit initialRoute override above everything else", () => {
    window.__EASYINTERVIEW_INITIAL_ROUTE__ = {
      name: "workspace",
      params: { targetJobId: "tj-window" },
    };
    setLocation("/report?sessionId=s-1");
    expect(
      resolveInitialRoute({
        initialRoute: { name: "practice", params: { sessionId: "s-prop" } },
      }),
    ).toEqual({ name: "practice", params: { sessionId: "s-prop" } });
  });

  it("uses window.__EASYINTERVIEW_INITIAL_ROUTE__ next", () => {
    window.__EASYINTERVIEW_INITIAL_ROUTE__ = {
      name: "workspace",
      params: { targetJobId: "tj-window" },
    };
    setLocation("/report?sessionId=s-1");
    expect(resolveInitialRoute()).toEqual({
      name: "workspace",
      params: { targetJobId: "tj-window" },
    });
  });

  it("falls back to canonical workspace path and strips legacy context params", () => {
    setLocation(
      "/workspace?targetJobId=tj-1&resumeId=rv-1&planId=plan-1",
    );
    expect(resolveInitialRoute()).toEqual({
      name: "workspace",
      params: {},
    });
  });

  it("drops unsafe canonical params during initial resolution", () => {
    setLocation("/workspace?targetJobId=tj-1&rawText=raw");
    expect(resolveInitialRoute()).toEqual({
      name: "workspace",
      params: {},
    });
  });

  it("falls back to hash adapter when path is bare `/` and hash carries #route=", () => {
    setLocation("/#route=workspace&targetJobId=tj-1");
    expect(resolveInitialRoute()).toEqual({
      name: "workspace",
      params: { targetJobId: "tj-1" },
    });
  });

  it("normalizes non-current hash aliases through normalizeRoute", () => {
    setLocation("/#route=voice");
    expect(resolveInitialRoute()).toEqual({ name: "home", params: {} });
  });

  it("returns DEFAULT_ROUTE (home) when nothing matches", () => {
    setLocation("/");
    expect(resolveInitialRoute()).toEqual({ name: "home", params: {} });
  });

  it("treats home with safe query params as canonical", () => {
    setLocation("/?pendingImportId=imp-1&source=paste");
    expect(resolveInitialRoute()).toEqual({
      name: "home",
      params: { pendingImportId: "imp-1", source: "paste" },
    });
  });

  it("treats unknown canonical path as home with empty params", () => {
    setLocation("/totally-unknown");
    expect(resolveInitialRoute()).toEqual({ name: "home", params: {} });
  });
});
