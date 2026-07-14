// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { EasyInterviewClient } from "../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { ResumeCreateFlow } from "./ResumeCreateFlow";

import getRuntimeConfigFixture from "../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../openapi/fixtures/Auth/getMe.json";
import createUploadPresignFixture from "../../../../../../openapi/fixtures/Uploads/createUploadPresign.json";
import registerResumeFixture from "../../../../../../openapi/fixtures/Resumes/registerResume.json";
import getResumeFixture from "../../../../../../openapi/fixtures/Resumes/getResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  // Use a future-dated presign so TTL guard doesn't re-fire in tests.
  {
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
  },
  registerResumeFixture,
  getResumeFixture,
];

function buildClient(resumeUploadBytes?: number): EasyInterviewClient {
  const runtimeConfigFixture = {
    ...getRuntimeConfigFixture,
    scenarios: {
      ...getRuntimeConfigFixture.scenarios,
      default: {
        ...getRuntimeConfigFixture.scenarios.default,
        response: {
          ...getRuntimeConfigFixture.scenarios.default.response,
          body: {
            ...getRuntimeConfigFixture.scenarios.default.response.body,
            contentLimits: resumeUploadBytes === undefined
              ? getRuntimeConfigFixture.scenarios.default.response.body.contentLimits
              : {
                  ...getRuntimeConfigFixture.scenarios.default.response.body.contentLimits,
                  resumeUploadBytes,
                },
          },
        },
      },
    },
  };
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry([
      runtimeConfigFixture,
      ...FIXTURES.filter((fixture) => fixture.operationId !== "getRuntimeConfig"),
    ]), {
      scenario: "default",
    }),
  });
}

function renderUploadTab(client: EasyInterviewClient) {
  const navigate = vi.fn();
  const result = render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      >
        <NavigationProvider value={{ navigate }}>
          <ResumeCreateFlow />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
  return { ...result, navigate };
}

function makeFile(name: string, size: number, type: string): File {
  const file = new File(["boundary-metadata-only"], name, {
    type,
    lastModified: 1700000000000,
  });
  Object.defineProperty(file, "size", { configurable: true, value: size });
  return file;
}

describe("UploadTab pre-check + presign + register", () => {
  let originalFetch: typeof fetch | undefined;
  let putSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    originalFetch = globalThis.fetch;
    putSpy = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url =
        input instanceof URL
          ? input.href
          : input instanceof Request
            ? input.url
            : String(input);
      if (url.includes("uploads.acme.example")) {
        // Echo the body length to assert binary handoff.
        const body = init?.body as File;
        return new Response(JSON.stringify({ byteSize: body.size }), {
          status: 200,
        });
      }
      throw new Error(`unexpected fetch url: ${url}`);
    });
    globalThis.fetch = putSpy as unknown as typeof fetch;
  });

  afterEach(() => {
    if (originalFetch) globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  it("rejects an unsupported file extension inline (no presign / register fired)", async () => {
    const client = buildClient();
    const presignSpy = vi.spyOn(client, "createUploadPresign");
    const registerSpy = vi.spyOn(client, "registerResume");

    renderUploadTab(client);
    await waitFor(() =>
      expect(screen.getByTestId("resume-create-upload-input")).toBeInTheDocument(),
    );
    const input = screen.getByTestId(
      "resume-create-upload-input",
    ) as HTMLInputElement;
    fireEvent.change(input, {
      target: { files: [makeFile("photo.jpeg", 1024, "image/jpeg")] },
    });
    expect(
      screen.getByTestId("resume-create-upload-error"),
    ).toHaveTextContent(/不支持|Unsupported/);
    expect(presignSpy).not.toHaveBeenCalled();
    expect(registerSpy).not.toHaveBeenCalled();
  });

  it("rejects DOCX files because the resume module only supports PDF, Markdown, and text upload", async () => {
    const client = buildClient();
    const presignSpy = vi.spyOn(client, "createUploadPresign");
    const registerSpy = vi.spyOn(client, "registerResume");

    renderUploadTab(client);
    await waitFor(() =>
      expect(screen.getByTestId("resume-create-upload-input")).toBeInTheDocument(),
    );
    const input = screen.getByTestId(
      "resume-create-upload-input",
    ) as HTMLInputElement;
    expect(input.accept).toBe(".pdf,.md,.markdown,.txt");
    fireEvent.change(input, {
      target: {
        files: [
          makeFile(
            "resume.docx",
            1024,
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
          ),
        ],
      },
    });
    expect(
      screen.getByTestId("resume-create-upload-error"),
    ).toHaveTextContent(/不支持|Unsupported/);
    expect(presignSpy).not.toHaveBeenCalled();
    expect(registerSpy).not.toHaveBeenCalled();
  });

  it("accepts .markdown resume files using Markdown content type", async () => {
    const client = buildClient();
    const presignSpy = vi.spyOn(client, "createUploadPresign");

    renderUploadTab(client);
    await waitFor(() =>
      expect(screen.getByTestId("resume-create-upload-input")).toBeInTheDocument(),
    );
    const input = screen.getByTestId(
      "resume-create-upload-input",
    ) as HTMLInputElement;
    fireEvent.change(input, {
      target: {
        files: [makeFile("resume.markdown", 1024, "")],
      },
    });
    await waitFor(() => expect(presignSpy).toHaveBeenCalled());
    expect(presignSpy.mock.calls[0]?.[0]?.contentType).toBe(
      "text/markdown",
    );
  });

  it("formats and enforces a non-MiB runtime override without rounding it", async () => {
    const client = buildClient(1536);
    const presignSpy = vi.spyOn(client, "createUploadPresign");
    renderUploadTab(client);
    await waitFor(() =>
      expect(screen.getByTestId("resume-create-upload-dropzone")).toHaveTextContent(/1\.5 ?KiB/),
    );

    const input = screen.getByTestId("resume-create-upload-input");
    fireEvent.change(input, {
      target: { files: [makeFile("exact.txt", 1536, "text/plain")] },
    });
    await waitFor(() => expect(presignSpy).toHaveBeenCalledTimes(1));

    fireEvent.change(input, {
      target: { files: [makeFile("over.txt", 1537, "text/plain")] },
    });
    expect(screen.getByTestId("resume-create-upload-error")).toHaveTextContent(/1\.5 ?KiB/);
    expect(presignSpy).toHaveBeenCalledTimes(1);
  });

  it("completes presign + browser PUT + registerResume + opens the detail directly", async () => {
    const client = buildClient();
    const presignSpy = vi.spyOn(client, "createUploadPresign");
    const registerSpy = vi.spyOn(client, "registerResume");

    const { navigate } = renderUploadTab(client);
    await waitFor(() =>
      expect(screen.getByTestId("resume-create-upload-input")).toBeInTheDocument(),
    );
    const input = screen.getByTestId(
      "resume-create-upload-input",
    ) as HTMLInputElement;
    const file = makeFile("alice.pdf", 2048, "application/pdf");
    fireEvent.change(input, { target: { files: [file] } });

    await waitFor(() =>
      expect(navigate).toHaveBeenCalledWith({
        name: "resume_versions",
        params: {
          resumeId: registerResumeFixture.scenarios.default.response.body.resumeId,
        },
      }),
    );
    expect(presignSpy).toHaveBeenCalledTimes(1);
    expect(presignSpy.mock.calls[0]![1]?.idempotencyKey).toMatch(
      /^v1\.\d+\.[0-9a-f-]{36}$/,
    );
    expect(registerSpy).toHaveBeenCalledTimes(1);
    const registerCall = registerSpy.mock.calls[0]!;
    expect(registerCall[0]).toMatchObject({
      sourceType: "upload",
      fileObjectId: "01918fa0-0000-7000-8000-000000001100",
      title: "alice.pdf",
    });
    expect(registerCall[1]?.idempotencyKey).toMatch(/^v1\.\d+\.[0-9a-f-]{36}$/);
    expect(putSpy).toHaveBeenCalledTimes(1);
    expect(screen.queryByTestId("resume-parse-flow")).not.toBeInTheDocument();
    expect(screen.queryByTestId("resume-preview-confirm")).not.toBeInTheDocument();
    // The create DOM never renders the original file blob.
    const bodyText = document.body.textContent ?? "";
    expect(bodyText).not.toContain("application/pdf");
    expect(bodyText).not.toContain(file.name + ".binary");
  });
});
