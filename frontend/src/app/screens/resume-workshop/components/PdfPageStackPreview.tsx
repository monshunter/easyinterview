import type { FC } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import type { PDFDocumentProxy } from "pdfjs-dist/legacy/build/pdf.mjs";

type PDFLoadingTask = ReturnType<
  (typeof import("pdfjs-dist/legacy/build/pdf.mjs"))["getDocument"]
>;

interface PdfPageStackPreviewProps {
  sourceUrl: string;
  label: string;
}

interface PdfPageCanvasProps {
  document: PDFDocumentProxy;
  pageNumber: number;
}

const PAGE_RENDER_SCALE = 1.25;

const PdfPageCanvas: FC<PdfPageCanvasProps> = ({ document, pageNumber }) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const [renderState, setRenderState] = useState<"loading" | "ready" | "error">(
    "loading",
  );

  useEffect(() => {
    let cancelled = false;
    let renderTask: { cancel?: () => void; promise: Promise<unknown> } | null =
      null;

    setRenderState("loading");

    const renderPage = async () => {
      try {
        const page = await document.getPage(pageNumber);
        if (cancelled) return;

        const viewport = page.getViewport({ scale: PAGE_RENDER_SCALE });
        const canvas = canvasRef.current;
        const context = canvas?.getContext("2d");
        if (!canvas || !context) {
          throw new Error("PDF canvas context unavailable");
        }

        const outputScale = window.devicePixelRatio || 1;
        canvas.width = Math.floor(viewport.width * outputScale);
        canvas.height = Math.floor(viewport.height * outputScale);
        canvas.style.width = `${viewport.width}px`;
        canvas.style.height = `${viewport.height}px`;

        if (typeof context.setTransform === "function") {
          context.setTransform(outputScale, 0, 0, outputScale, 0, 0);
        }

        renderTask = page.render({ canvas, canvasContext: context, viewport });
        await renderTask.promise;

        if (!cancelled) setRenderState("ready");
      } catch {
        if (!cancelled) setRenderState("error");
      }
    };

    void renderPage();

    return () => {
      cancelled = true;
      renderTask?.cancel?.();
    };
  }, [document, pageNumber]);

  return (
    <div
      className="ei-resume-detail-pdf-page"
      data-render-state={renderState}
      data-testid={`resume-detail-pdf-page-${pageNumber}`}
    >
      {renderState === "error" ? (
        <p className="ei-text-body ei-resume-detail-pdf-error">
          PDF page unavailable.
        </p>
      ) : null}
      <canvas
        ref={canvasRef}
        className="ei-resume-detail-pdf-canvas"
        aria-label={`PDF page ${pageNumber}`}
      />
    </div>
  );
};

export const PdfPageStackPreview: FC<PdfPageStackPreviewProps> = ({
  sourceUrl,
  label,
}) => {
  const [document, setDocument] = useState<PDFDocumentProxy | null>(null);
  const [pageCount, setPageCount] = useState(0);
  const [loadState, setLoadState] = useState<"loading" | "ready" | "error">(
    "loading",
  );

  useEffect(() => {
    let cancelled = false;
    let loadedDocument: PDFDocumentProxy | null = null;
    let loadingTask: PDFLoadingTask | null = null;

    setDocument(null);
    setPageCount(0);
    setLoadState("loading");

    const loadDocument = async () => {
      try {
        const [pdfjs, worker] = await Promise.all([
          import("pdfjs-dist/legacy/build/pdf.mjs"),
          import("pdfjs-dist/legacy/build/pdf.worker.mjs?url"),
        ]);
        if (cancelled) return;

        pdfjs.GlobalWorkerOptions.workerSrc = worker.default;
        loadingTask = pdfjs.getDocument({
          url: sourceUrl,
          withCredentials: true,
        });
        const pdfDocument = await loadingTask.promise;
        if (cancelled) {
          void pdfDocument.destroy();
          return;
        }
        loadedDocument = pdfDocument;
        setDocument(pdfDocument);
        setPageCount(pdfDocument.numPages);
        setLoadState("ready");
      } catch {
        if (!cancelled) setLoadState("error");
      }
    };

    void loadDocument();

    return () => {
      cancelled = true;
      if (loadedDocument) {
        void loadedDocument.destroy();
        return;
      }
      void loadingTask?.destroy();
    };
  }, [sourceUrl]);

  const pageNumbers = useMemo(
    () => Array.from({ length: pageCount }, (_, index) => index + 1),
    [pageCount],
  );

  return (
    <div
      className="ei-resume-detail-pdf-stack"
      data-source-url={sourceUrl}
      data-testid="resume-detail-pdf-preview-stack"
      aria-label={label}
    >
      {loadState === "loading" ? (
        <div
          className="ei-text-body ei-resume-detail-pdf-loading"
          data-testid="resume-detail-pdf-loading"
        >
          Loading PDF preview...
        </div>
      ) : null}
      {loadState === "error" ? (
        <p className="ei-text-body ei-resume-detail-pdf-error">
          PDF preview unavailable.
        </p>
      ) : null}
      {document
        ? pageNumbers.map((pageNumber) => (
            <PdfPageCanvas
              key={pageNumber}
              document={document}
              pageNumber={pageNumber}
            />
          ))
        : null}
    </div>
  );
};
