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
import { translate } from "../../../i18n/messages";
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

function dropFiles(dropzone: HTMLElement, files: File[]): void {
  fireEvent.drop(dropzone, {
    dataTransfer: { files, types: ["Files"] },
  });
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

  it("shows truthful drop copy without the retired central icon and resets drag-active state", async () => {
    renderUploadTab(buildClient());
    const dropzone = await screen.findByTestId("resume-create-upload-dropzone");

    expect(dropzone).toHaveAttribute("data-drag-active", "false");
    expect(dropzone).toHaveTextContent("Drop a PDF / Markdown / TXT resume here");
    expect(translate("zh", "resumeWorkshop.create.upload.dropzoneTitle"))
      .toBe("拖放 PDF / Markdown / TXT 简历到此处");
    expect(translate("zh", "resumeWorkshop.create.upload.dropzoneActiveTitle"))
      .toBe("松开以上传");
    expect(dropzone.querySelector(".ei-resume-create-upload-icon")).toBeNull();

    const dataTransfer = { files: [], types: ["Files"], dropEffect: "none" };
    fireEvent.dragOver(dropzone, { dataTransfer });
    expect(dropzone).toHaveAttribute("data-drag-active", "true");
    expect(dropzone).toHaveTextContent("Release to upload");
    expect(dataTransfer.dropEffect).toBe("copy");

    fireEvent.dragLeave(dropzone, {
      dataTransfer: { files: [], types: ["Files"] },
      relatedTarget: document.body,
    });
    expect(dropzone).toHaveAttribute("data-drag-active", "false");
    expect(dropzone).toHaveTextContent("Drop a PDF / Markdown / TXT resume here");

    fireEvent.dragEnter(dropzone, {
      dataTransfer: { files: [], types: ["Files"] },
    });
    fireEvent.drop(dropzone, {
      dataTransfer: { files: [], types: ["Files"] },
    });
    expect(dropzone).toHaveAttribute("data-drag-active", "false");
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
    ).toHaveTextContent(/不支持|isn't supported/i);
    expect(presignSpy).not.toHaveBeenCalled();
    expect(registerSpy).not.toHaveBeenCalled();
  });

  it("rejects multiple, unsupported, and oversized dropped files before any request", async () => {
    const client = buildClient(1536);
    const presignSpy = vi.spyOn(client, "createUploadPresign");
    const registerSpy = vi.spyOn(client, "registerResume");
    renderUploadTab(client);
    const dropzone = await screen.findByTestId("resume-create-upload-dropzone");

    dropFiles(dropzone, [
      makeFile("first.pdf", 512, "application/pdf"),
      makeFile("second.txt", 512, "text/plain"),
    ]);
    expect(screen.getByTestId("resume-create-upload-error"))
      .toHaveTextContent(/one file|一个文件/i);

    dropFiles(dropzone, [makeFile("photo.jpeg", 512, "image/jpeg")]);
    expect(screen.getByTestId("resume-create-upload-error"))
      .toHaveTextContent(/isn't supported|不支持/i);

    dropFiles(dropzone, [makeFile("large.pdf", 1537, "application/pdf")]);
    expect(screen.getByTestId("resume-create-upload-error"))
      .toHaveTextContent(/1\.5 ?KiB/);
    expect(presignSpy).not.toHaveBeenCalled();
    expect(registerSpy).not.toHaveBeenCalled();
    expect(putSpy).not.toHaveBeenCalled();
  });

  it("ignores a file drop until runtime limits are ready", async () => {
    const client = buildClient();
    const presignSpy = vi.spyOn(client, "createUploadPresign");
    const registerSpy = vi.spyOn(client, "registerResume");
    renderUploadTab(client);
    const dropzone = screen.getByTestId("resume-create-upload-dropzone");

    dropFiles(dropzone, [makeFile("early.pdf", 512, "application/pdf")]);
    await screen.findByText(/Up to 10 MiB|文件最大 10 MiB/);
    expect(presignSpy).not.toHaveBeenCalled();
    expect(registerSpy).not.toHaveBeenCalled();
    expect(putSpy).not.toHaveBeenCalled();
  });

  it("keeps the choose-file button as the explicit keyboard and touch fallback", async () => {
    renderUploadTab(buildClient());
    const input = await screen.findByTestId("resume-create-upload-input");
    const clickSpy = vi.spyOn(input as HTMLInputElement, "click");

    fireEvent.click(screen.getByTestId("resume-create-upload-choose"));
    expect(clickSpy).toHaveBeenCalledTimes(1);
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
    ).toHaveTextContent(/不支持|isn't supported/i);
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

  it("uses the same upload path for one dropped file and ignores another drop while submitting", async () => {
    const client = buildClient();
    const presignSpy = vi.spyOn(client, "createUploadPresign");
    const registerSpy = vi.spyOn(client, "registerResume");
    let finishPut: ((response: Response) => void) | undefined;
    putSpy.mockImplementationOnce(() => new Promise<Response>((resolve) => {
      finishPut = resolve;
    }));

    const { navigate } = renderUploadTab(client);
    const dropzone = await screen.findByTestId("resume-create-upload-dropzone");
    const first = makeFile("dropped.md", 1024, "text/markdown");
    dropFiles(dropzone, [first]);
    await waitFor(() => expect(putSpy).toHaveBeenCalledTimes(1));

    dropFiles(dropzone, [makeFile("duplicate.pdf", 1024, "application/pdf")]);
    expect(presignSpy).toHaveBeenCalledTimes(1);
    expect(registerSpy).not.toHaveBeenCalled();

    finishPut?.(new Response(null, { status: 200 }));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith({
      name: "resume_versions",
      params: { resumeId: registerResumeFixture.scenarios.default.response.body.resumeId },
    }));
    expect(registerSpy).toHaveBeenCalledTimes(1);
    expect(registerSpy.mock.calls[0]?.[0]).toMatchObject({
      sourceType: "upload",
      title: "dropped.md",
    });
    expect(screen.queryByText("duplicate.pdf")).not.toBeInTheDocument();
    expect(document.body).not.toHaveTextContent("boundary-metadata-only");
  });
});
