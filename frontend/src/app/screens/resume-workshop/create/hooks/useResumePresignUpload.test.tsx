// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../../../api/generated/client";
import { AppRuntimeProvider } from "../../../../runtime/AppRuntimeProvider";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../../api/mockTransport";
import {
  useResumePresignUpload,
  type PresignedUploadResult,
} from "./useResumePresignUpload";

import getRuntimeConfigFixture from "../../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../../openapi/fixtures/Auth/getMe.json";
import createUploadPresignFixture from "../../../../../../../openapi/fixtures/Uploads/createUploadPresign.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  createUploadPresignFixture,
];

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

function buildFile(name: string, size: number, type: string): File {
  const blob = new Blob([new Uint8Array(size)], { type });
  return new File([blob], name, { type, lastModified: 1700000000000 });
}

describe("useResumePresignUpload", () => {
  let fetchSpy: ReturnType<typeof vi.fn>;
  let originalFetch: typeof fetch | undefined;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
    fetchSpy = vi.fn(async (input: RequestInfo | URL) => {
      const url = input instanceof URL ? input.href : input instanceof Request ? input.url : String(input);
      if (url.includes("uploads.acme.example")) {
        return new Response(null, { status: 200 });
      }
      throw new Error(`unexpected url: ${url}`);
    });
    // The Phase 2 hook uses the global `fetch` for the signed-URL PUT.
    globalThis.fetch = fetchSpy as unknown as typeof fetch;
  });

  afterEach(() => {
    if (originalFetch) globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("calls createUploadPresign with the file metadata and an Idempotency-Key, then PUTs to the signed URL", async () => {
    const client = buildClient();
    const presignSpy = vi.spyOn(client, "createUploadPresign");
    const { result } = renderHook(() => useResumePresignUpload(), {
      wrapper: buildWrapper(client),
    });

    const file = buildFile("resume.pdf", 1024, "application/pdf");
    let returned: PresignedUploadResult | null = null;
    await act(async () => {
      returned = await result.current.uploadFile(file, {
        contentType: "application/pdf",
      });
    });

    expect(presignSpy).toHaveBeenCalledTimes(1);
    const callArgs = presignSpy.mock.calls[0]!;
    expect(callArgs[0]).toEqual({
      purpose: "resume",
      fileName: "resume.pdf",
      contentType: "application/pdf",
      byteSize: 1024,
    });
    expect(callArgs[1]?.idempotencyKey).toMatch(/^v1\.\d+\.[0-9a-f-]{36}$/);

    const returnedResult = returned as PresignedUploadResult | null;
    expect(returnedResult?.fileObjectId).toBe(
      "01918fa0-0000-7000-8000-000000001100",
    );
    expect(fetchSpy).toHaveBeenCalledTimes(1);
    const putCall = fetchSpy.mock.calls[0]!;
    expect(putCall[1]).toMatchObject({ method: "PUT" });
    expect(putCall[1]?.body).toBe(file);
  });

  it("re-uses the same idempotency key when retrying the same file before TTL expires", async () => {
    const futureFixture = {
      ...createUploadPresignFixture,
      scenarios: {
        default: {
          ...createUploadPresignFixture.scenarios.default,
          response: {
            ...createUploadPresignFixture.scenarios.default.response,
            body: {
              ...createUploadPresignFixture.scenarios.default.response.body,
              expiresAt: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
            },
          },
        },
      },
    };
    const client = new EasyInterviewClient({
      fetch: createFixtureBackedFetch(
        createFixtureRegistry([
          getRuntimeConfigFixture,
          getMeFixture,
          futureFixture,
        ]),
        { scenario: "default" },
      ),
    });
    const presignSpy = vi.spyOn(client, "createUploadPresign");
    const { result } = renderHook(() => useResumePresignUpload(), {
      wrapper: buildWrapper(client),
    });

    const file = buildFile("resume.pdf", 1024, "application/pdf");
    await act(async () => {
      await result.current.uploadFile(file, { contentType: "application/pdf" });
    });
    // Second attempt: TTL not expired → should not re-presign.
    await act(async () => {
      await result.current.uploadFile(file, { contentType: "application/pdf" });
    });
    expect(presignSpy).toHaveBeenCalledTimes(1);
  });

  it("re-presigns with a fresh idempotency key when the signed URL TTL has expired", async () => {
    const expiredFixture = {
      ...createUploadPresignFixture,
      scenarios: {
        default: {
          ...createUploadPresignFixture.scenarios.default,
          response: {
            ...createUploadPresignFixture.scenarios.default.response,
            body: {
              ...createUploadPresignFixture.scenarios.default.response.body,
              expiresAt: new Date(Date.now() - 5_000).toISOString(),
            },
          },
        },
      },
    };
    const client = new EasyInterviewClient({
      fetch: createFixtureBackedFetch(
        createFixtureRegistry([
          getRuntimeConfigFixture,
          getMeFixture,
          expiredFixture,
        ]),
        { scenario: "default" },
      ),
    });
    const presignSpy = vi.spyOn(client, "createUploadPresign");

    // Make the upload PUT fail on first attempt only when expired, succeed on retry.
    let putAttempt = 0;
    fetchSpy.mockImplementation(async (input: RequestInfo | URL) => {
      const url =
        input instanceof URL
          ? input.href
          : input instanceof Request
            ? input.url
            : String(input);
      if (!url.includes("uploads.acme.example")) {
        throw new Error(`unexpected url: ${url}`);
      }
      putAttempt += 1;
      if (putAttempt === 1) throw new Error("UPLOAD_PUT_FAILED:expired");
      return new Response(null, { status: 200 });
    });

    const { result } = renderHook(() => useResumePresignUpload(), {
      wrapper: buildWrapper(client),
    });

    const file = buildFile("resume.pdf", 1024, "application/pdf");
    await act(async () => {
      await result.current.uploadFile(file, { contentType: "application/pdf" });
    });
    expect(presignSpy).toHaveBeenCalledTimes(2);
    const firstKey = presignSpy.mock.calls[0]![1]?.idempotencyKey;
    const secondKey = presignSpy.mock.calls[1]![1]?.idempotencyKey;
    expect(firstKey).toBeTruthy();
    expect(secondKey).toBeTruthy();
    expect(secondKey).not.toBe(firstKey);
  });

  it("does not leak the file binary or the raw URL into localStorage / sessionStorage / console", async () => {
    const setLocal = vi.spyOn(window.localStorage.__proto__, "setItem");
    const setSession = vi.spyOn(window.sessionStorage.__proto__, "setItem");
    const consoleLog = vi.spyOn(console, "log").mockImplementation(() => {});
    const consoleInfo = vi.spyOn(console, "info").mockImplementation(() => {});

    const client = buildClient();
    const { result } = renderHook(() => useResumePresignUpload(), {
      wrapper: buildWrapper(client),
    });

    const file = buildFile("resume.pdf", 1024, "application/pdf");
    await act(async () => {
      await result.current.uploadFile(file, { contentType: "application/pdf" });
    });

    expect(setLocal).not.toHaveBeenCalled();
    expect(setSession).not.toHaveBeenCalled();
    expect(consoleLog).not.toHaveBeenCalled();
    expect(consoleInfo).not.toHaveBeenCalled();
  });
});
