// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../../api/mockTransport";
import { AppRuntimeProvider } from "../../../../runtime/AppRuntimeProvider";
import {
  buildRegisterPayload,
  useResumeRegistration,
} from "./useResumeRegistration";

import getRuntimeConfigFixture from "../../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../../openapi/fixtures/Auth/getMe.json";
import registerResumeFixture from "../../../../../../../openapi/fixtures/Resumes/registerResume.json";

const FIXTURES = [getRuntimeConfigFixture, getMeFixture, registerResumeFixture];

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

describe("buildRegisterPayload — three sourceType mapper", () => {
  it("maps upload input to { sourceType, fileObjectId, title, language }", () => {
    expect(
      buildRegisterPayload({
        sourceType: "upload",
        fileObjectId: "fobj-1",
        title: "alice.pdf",
        language: "zh",
      }),
    ).toEqual({
      sourceType: "upload",
      fileObjectId: "fobj-1",
      title: "alice.pdf",
      language: "zh",
    });
  });

  it("maps paste input to { sourceType, rawText, title, language }", () => {
    expect(
      buildRegisterPayload({
        sourceType: "paste",
        rawText: "I am Alice…",
        title: "粘贴的简历",
        language: "zh",
      }),
    ).toEqual({
      sourceType: "paste",
      rawText: "I am Alice…",
      title: "粘贴的简历",
      language: "zh",
    });
  });

  it("maps guided input to { sourceType, guidedAnswers, title, language } with all 5 keys preserved", () => {
    const guided = {
      recentRole: "Senior FE @ Foo",
      direction: "frontend platform",
      proofProject: "RSC migration",
      metrics: "LCP 3.2s → 1.4s",
      target: "Staff frontend",
    };
    expect(
      buildRegisterPayload({
        sourceType: "guided",
        guidedAnswers: guided,
        title: "问答简历",
        language: "zh",
      }),
    ).toEqual({
      sourceType: "guided",
      guidedAnswers: guided,
      title: "问答简历",
      language: "zh",
    });
  });

  it("preserves empty-string guided answers (does not trim user input)", () => {
    const guided = {
      recentRole: "",
      direction: "",
      proofProject: "Project A",
      metrics: "",
      target: "Senior FE",
    };
    const payload = buildRegisterPayload({
      sourceType: "guided",
      guidedAnswers: guided,
      title: "问答简历",
      language: "zh",
    });
    expect(payload.guidedAnswers).toEqual(guided);
  });
});

describe("useResumeRegistration", () => {
  it("calls registerResume with an Idempotency-Key + Accept-Language for paste", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "registerResume");
    const { result } = renderHook(() => useResumeRegistration(), {
      wrapper: buildWrapper(client),
    });

    await act(async () => {
      await result.current.register({
        sourceType: "paste",
        rawText: "alpha beta",
        title: "粘贴的简历",
        language: "zh",
      });
    });

    expect(spy).toHaveBeenCalledTimes(1);
    const call = spy.mock.calls[0]!;
    expect(call[0]).toMatchObject({
      sourceType: "paste",
      rawText: "alpha beta",
      title: "粘贴的简历",
      language: "zh",
    });
    expect(call[1]?.idempotencyKey).toMatch(/^v1\.\d+\.[0-9a-f-]{36}$/);
    expect(call[1]?.headers).toMatchObject({ "Accept-Language": "zh" });
  });

  it("calls registerResume with an Idempotency-Key for upload", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "registerResume");
    const { result } = renderHook(() => useResumeRegistration(), {
      wrapper: buildWrapper(client),
    });
    await act(async () => {
      await result.current.register({
        sourceType: "upload",
        fileObjectId: "fobj-1",
        title: "alice.pdf",
        language: "en",
      });
    });
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy.mock.calls[0]![1]?.idempotencyKey).toMatch(
      /^v1\.\d+\.[0-9a-f-]{36}$/,
    );
    expect(spy.mock.calls[0]![1]?.headers).toMatchObject({
      "Accept-Language": "en",
    });
  });

  it("generates a fresh idempotency key per invocation", async () => {
    const client = buildClient();
    const spy = vi.spyOn(client, "registerResume");
    const { result } = renderHook(() => useResumeRegistration(), {
      wrapper: buildWrapper(client),
    });
    await act(async () => {
      await result.current.register({
        sourceType: "paste",
        rawText: "alpha",
        title: "粘贴的简历",
        language: "zh",
      });
    });
    await act(async () => {
      await result.current.register({
        sourceType: "paste",
        rawText: "alpha",
        title: "粘贴的简历",
        language: "zh",
      });
    });
    expect(spy).toHaveBeenCalledTimes(2);
    const keyA = spy.mock.calls[0]![1]?.idempotencyKey;
    const keyB = spy.mock.calls[1]![1]?.idempotencyKey;
    expect(keyA).toBeTruthy();
    expect(keyB).toBeTruthy();
    expect(keyA).not.toBe(keyB);
  });

  it("propagates HTTP errors from registerResume (422 surface, no IK reuse)", async () => {
    const failingClient = new EasyInterviewClient({
      fetch: async () => new Response(
        JSON.stringify({ code: "VALIDATION_FAILED", details: {} }),
        { status: 422, headers: { "Content-Type": "application/json" } },
      ),
    });
    const { result } = renderHook(() => useResumeRegistration(), {
      wrapper: buildWrapper(failingClient),
    });
    await expect(
      act(async () => {
        await result.current.register({
          sourceType: "paste",
          rawText: "alpha",
          title: "粘贴的简历",
          language: "zh",
        });
      }),
    ).rejects.toThrow(/HTTP 422|VALIDATION_FAILED/);
  });
});
