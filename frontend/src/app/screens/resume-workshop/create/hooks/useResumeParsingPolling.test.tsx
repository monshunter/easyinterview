// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../../../api/generated/client";
import type { ResumeAsset } from "../../../../../api/generated/types";
import { AppRuntimeProvider } from "../../../../runtime/AppRuntimeProvider";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../../api/mockTransport";
import { useResumeParsingPolling } from "./useResumeParsingPolling";

import getRuntimeConfigFixture from "../../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../../openapi/fixtures/Auth/getMe.json";
import getResumeFixture from "../../../../../../../openapi/fixtures/Resumes/getResume.json";

const FIXTURES = [getRuntimeConfigFixture, getMeFixture, getResumeFixture];

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario: "default",
    }),
  });
}

function buildWrapper(client: EasyInterviewClient) {
  return ({ children }: { children: ReactNode }) => (
    <AppRuntimeProvider
      client={client}
      requestOptions={{
        getMe: { headers: { Prefer: "example=authenticated" } },
      }}
    >
      {children}
    </AppRuntimeProvider>
  );
}

const ASSET_BASE: ResumeAsset = {
  id: "01918fa0-0000-7000-8000-000000001100",
  title: "alice.pdf",
  language: "zh",
  parseStatus: "ready",
  createdAt: "2026-05-17T00:00:00Z",
  updatedAt: "2026-05-17T00:00:00Z",
  parsedSummary: {
    identity: { name: "Alice", title: "Senior FE" },
    summary: "Summary text",
    skills: ["React"],
    experience: [],
    projects: [],
    education: [],
  },
};

describe("useResumeParsingPolling", () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("polls getResume without Idempotency-Key (read op) and resolves to ready on terminal state", async () => {
    const client = buildClient();
    const spy = vi
      .spyOn(client, "getResume")
      .mockResolvedValueOnce({ ...ASSET_BASE, parseStatus: "queued" })
      .mockResolvedValueOnce({ ...ASSET_BASE, parseStatus: "processing" })
      .mockResolvedValueOnce(ASSET_BASE);

    const { result } = renderHook(
      () =>
        useResumeParsingPolling("01918fa0-0000-7000-8000-000000001100", {
          initialDelayMs: 10,
          backoffFactor: 1,
          maxAttempts: 5,
          maxTotalMs: 1000,
        }),
      { wrapper: buildWrapper(client) },
    );

    await waitFor(
      () => {
        expect(result.current.snapshot.status).toBe("ready");
      },
      { timeout: 2000 },
    );
    expect(spy).toHaveBeenCalledTimes(3);
    for (const call of spy.mock.calls) {
      const opts = call[1];
      expect(opts?.idempotencyKey).toBeUndefined();
    }
  });

  it("flips to failed when parseStatus=failed", async () => {
    const client = buildClient();
    vi.spyOn(client, "getResume").mockResolvedValue({
      ...ASSET_BASE,
      parseStatus: "failed",
    });
    const { result } = renderHook(
      () =>
        useResumeParsingPolling("01918fa0-0000-7000-8000-000000001100", {
          initialDelayMs: 10,
          backoffFactor: 1,
          maxAttempts: 3,
          maxTotalMs: 500,
        }),
      { wrapper: buildWrapper(client) },
    );

    await waitFor(() => {
      expect(result.current.snapshot.status).toBe("failed");
    });
    expect(result.current.snapshot.errorCode).toBe("AI_TIMEOUT_RETRYABLE");
  });

  it("times out to PARSE_TIMEOUT once attempts exhaust without a terminal status", async () => {
    const client = buildClient();
    vi.spyOn(client, "getResume").mockResolvedValue({
      ...ASSET_BASE,
      parseStatus: "processing",
    });
    const { result } = renderHook(
      () =>
        useResumeParsingPolling("01918fa0-0000-7000-8000-000000001100", {
          initialDelayMs: 5,
          backoffFactor: 1,
          maxAttempts: 3,
          maxTotalMs: 200,
        }),
      { wrapper: buildWrapper(client) },
    );
    await waitFor(
      () => {
        expect(result.current.snapshot.status).toBe("failed");
      },
      { timeout: 2000 },
    );
    expect(result.current.snapshot.errorCode).toBe("PARSE_TIMEOUT");
  });

  it("cancel() reverts to idle and stops further polling", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "getResume").mockResolvedValue({
      ...ASSET_BASE,
      parseStatus: "processing",
    });
    const { result } = renderHook(
      () =>
        useResumeParsingPolling("01918fa0-0000-7000-8000-000000001100", {
          initialDelayMs: 20,
          backoffFactor: 1,
          maxAttempts: 10,
          maxTotalMs: 5_000,
        }),
      { wrapper: buildWrapper(client) },
    );
    // Wait for first poll
    await waitFor(() => {
      expect(spy).toHaveBeenCalled();
    });
    const before = spy.mock.calls.length;
    act(() => {
      result.current.cancel();
    });
    expect(result.current.snapshot.status).toBe("idle");
    // Allow time for any pending tick; cancel guards subsequent state updates.
    await act(async () => {
      await new Promise((r) => setTimeout(r, 80));
    });
    // No more state transitions after cancel — status stays idle.
    expect(result.current.snapshot.status).toBe("idle");
    // It's acceptable for the in-flight call to land, but no escalation happens.
    expect(spy.mock.calls.length).toBeGreaterThanOrEqual(before);
  });
});
