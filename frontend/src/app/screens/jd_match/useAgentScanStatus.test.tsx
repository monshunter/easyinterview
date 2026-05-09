// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";

import getAgentScanStatusFixture from "../../../../../openapi/fixtures/JobMatch/getAgentScanStatus.json";

import { useAgentScanStatus } from "./useAgentScanStatus";

function buildClient(scenario?: string) {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getAgentScanStatusFixture]),
      scenario ? { scenario } : undefined,
    ),
  });
}

function wrapWithRuntime(client: EasyInterviewClient) {
  return ({ children }: { children: ReactNode }) => (
    <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>
  );
}

describe("useAgentScanStatus", () => {
  it("calls getAgentScanStatus once on mount when activeTab is recommended", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "getAgentScanStatus");

    const { result } = renderHook(
      ({ tab }: { tab: string }) => useAgentScanStatus(tab),
      {
        wrapper: wrapWithRuntime(client),
        initialProps: { tab: "recommended" },
      },
    );

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(spy).toHaveBeenCalledTimes(1);
    expect(result.current.data).not.toBeNull();
    expect(result.current.error).toBeNull();
  });

  it("does NOT call getAgentScanStatus on mount when activeTab is search or watchlist", async () => {
    const clientSearch = buildClient("default");
    const spySearch = vi.spyOn(clientSearch, "getAgentScanStatus");

    renderHook(
      ({ tab }: { tab: string }) => useAgentScanStatus(tab),
      {
        wrapper: wrapWithRuntime(clientSearch),
        initialProps: { tab: "search" },
      },
    );

    await new Promise((r) => setTimeout(r, 30));
    expect(spySearch).not.toHaveBeenCalled();

    const clientWatch = buildClient("default");
    const spyWatch = vi.spyOn(clientWatch, "getAgentScanStatus");
    renderHook(
      ({ tab }: { tab: string }) => useAgentScanStatus(tab),
      {
        wrapper: wrapWithRuntime(clientWatch),
        initialProps: { tab: "watchlist" },
      },
    );
    await new Promise((r) => setTimeout(r, 30));
    expect(spyWatch).not.toHaveBeenCalled();
  });

  it("re-fetches when activeTab transitions back to recommended", async () => {
    const client = buildClient("default");
    const spy = vi.spyOn(client, "getAgentScanStatus");

    const { result, rerender } = renderHook(
      ({ tab }: { tab: string }) => useAgentScanStatus(tab),
      {
        wrapper: wrapWithRuntime(client),
        initialProps: { tab: "recommended" },
      },
    );

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(spy).toHaveBeenCalledTimes(1);

    rerender({ tab: "search" });
    await new Promise((r) => setTimeout(r, 30));
    expect(spy).toHaveBeenCalledTimes(1);

    rerender({ tab: "recommended" });
    await waitFor(() => {
      expect(spy).toHaveBeenCalledTimes(2);
    });

    rerender({ tab: "watchlist" });
    await new Promise((r) => setTimeout(r, 30));
    expect(spy).toHaveBeenCalledTimes(2);

    rerender({ tab: "recommended" });
    await waitFor(() => {
      expect(spy).toHaveBeenCalledTimes(3);
    });
  });

  it("does NOT register setInterval / EventSource / WebSocket from hook effect", async () => {
    const client = buildClient("default");
    const eventSourceSpy = vi.fn();
    const webSocketSpy = vi.fn();

    const originalES =
      "EventSource" in globalThis
        ? (globalThis as { EventSource?: unknown }).EventSource
        : undefined;
    const originalWS =
      "WebSocket" in globalThis
        ? (globalThis as { WebSocket?: unknown }).WebSocket
        : undefined;
    Object.defineProperty(globalThis, "EventSource", {
      configurable: true,
      writable: true,
      value: eventSourceSpy,
    });
    Object.defineProperty(globalThis, "WebSocket", {
      configurable: true,
      writable: true,
      value: webSocketSpy,
    });

    const intervalSpy = vi.spyOn(window, "setInterval");

    try {
      const { result } = renderHook(
        ({ tab }: { tab: string }) => useAgentScanStatus(tab),
        {
          wrapper: wrapWithRuntime(client),
          initialProps: { tab: "recommended" },
        },
      );

      // Use bare setTimeout (NOT waitFor, which polls via setInterval) so the
      // hook's effect chain settles without polluting the interval spy.
      await new Promise((r) => setTimeout(r, 50));

      expect(result.current.loading).toBe(false);
      expect(intervalSpy).not.toHaveBeenCalled();
      expect(eventSourceSpy).not.toHaveBeenCalled();
      expect(webSocketSpy).not.toHaveBeenCalled();
    } finally {
      intervalSpy.mockRestore();
      Object.defineProperty(globalThis, "EventSource", {
        configurable: true,
        writable: true,
        value: originalES,
      });
      Object.defineProperty(globalThis, "WebSocket", {
        configurable: true,
        writable: true,
        value: originalWS,
      });
    }
  });

  it("surfaces error when getAgentScanStatus rejects", async () => {
    const client = buildClient("default");
    vi.spyOn(client, "getAgentScanStatus").mockRejectedValue(
      new Error("HTTP 500 — boom"),
    );

    const { result } = renderHook(
      ({ tab }: { tab: string }) => useAgentScanStatus(tab),
      {
        wrapper: wrapWithRuntime(client),
        initialProps: { tab: "recommended" },
      },
    );

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).not.toBeNull();
    expect(result.current.error?.message).toMatch(/HTTP 500/);
    expect(result.current.data).toBeNull();
  });
});
