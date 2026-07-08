// @vitest-environment jsdom
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const pdfMocks = vi.hoisted(() => {
  const destroyDocument = vi.fn();
  const destroyTask = vi.fn();
  const getDocument = vi.fn();
  const getPage = vi.fn();
  const renderPage = vi.fn();
  return {
    destroyDocument,
    destroyTask,
    getDocument,
    getPage,
    renderPage,
  };
});

vi.mock("pdfjs-dist/legacy/build/pdf.worker.mjs?url", () => ({
  default: "/mock-pdf.worker.mjs",
}));

vi.mock("pdfjs-dist/legacy/build/pdf.mjs", () => ({
  GlobalWorkerOptions: {},
  getDocument: pdfMocks.getDocument,
}));

import { PdfPageStackPreview } from "./PdfPageStackPreview";

const SOURCE_URL = "/api/v1/resumes/resume-1/source";

function mockPdfDocument(pageCount = 2) {
  pdfMocks.renderPage.mockReturnValue({
    cancel: vi.fn(),
    promise: Promise.resolve(),
  });
  pdfMocks.getPage.mockImplementation(async () => ({
    getViewport: ({ scale }: { scale: number }) => ({
      height: 792 * scale,
      width: 612 * scale,
    }),
    render: pdfMocks.renderPage,
  }));
  pdfMocks.getDocument.mockReturnValue({
    destroy: pdfMocks.destroyTask,
    promise: Promise.resolve({
      destroy: pdfMocks.destroyDocument,
      getPage: pdfMocks.getPage,
      numPages: pageCount,
    }),
  });
}

describe("PdfPageStackPreview", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(HTMLCanvasElement.prototype, "getContext", {
      configurable: true,
      value: vi.fn(() => ({
        setTransform: vi.fn(),
      })),
    });
  });

  it("loads the PDF source and renders every page in order without native viewer elements", async () => {
    mockPdfDocument(2);

    render(<PdfPageStackPreview sourceUrl={SOURCE_URL} label="Alice PDF" />);

    const stack = screen.getByTestId("resume-detail-pdf-preview-stack");
    expect(stack).toHaveAttribute("data-source-url", SOURCE_URL);
    expect(stack).toHaveAttribute("aria-label", "Alice PDF");
    expect(pdfMocks.getDocument).toHaveBeenCalledWith({
      url: SOURCE_URL,
      withCredentials: true,
    });

    await screen.findByTestId("resume-detail-pdf-page-1");
    await screen.findByTestId("resume-detail-pdf-page-2");
    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-pdf-page-1")).toHaveAttribute(
        "data-render-state",
        "ready",
      );
      expect(screen.getByTestId("resume-detail-pdf-page-2")).toHaveAttribute(
        "data-render-state",
        "ready",
      );
    });

    expect(pdfMocks.getPage).toHaveBeenNthCalledWith(1, 1);
    expect(pdfMocks.getPage).toHaveBeenNthCalledWith(2, 2);
    expect(pdfMocks.renderPage).toHaveBeenCalledTimes(2);
    expect(document.querySelector("object, iframe, embed")).toBeNull();
  });

  it("shows a simple inline failure state when the PDF cannot be loaded", async () => {
    pdfMocks.getDocument.mockReturnValue({
      destroy: pdfMocks.destroyTask,
      promise: Promise.reject(new Error("missing pdf")),
    });

    render(<PdfPageStackPreview sourceUrl={SOURCE_URL} label="Alice PDF" />);

    expect(await screen.findByText("PDF preview unavailable.")).toBeVisible();
    expect(screen.queryByTestId("resume-detail-pdf-page-1")).toBeNull();
    expect(document.querySelector("object, iframe, embed")).toBeNull();
  });
});
