import { useCallback, useRef, useState, type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import { ResumeWorkshopIcon } from "../components/ResumeWorkshopIcon";
import { useResumePresignUpload } from "./hooks/useResumePresignUpload";
import { useResumeRegistration } from "./hooks/useResumeRegistration";
import { deriveDefaultTitle } from "./util/title";

const MAX_RESUME_UPLOAD_BYTES = 10 * 1024 * 1024;
const ALLOWED_EXTENSIONS = [".pdf", ".docx", ".md", ".txt"];

const hasAllowedExtension = (name: string): boolean => {
  const lower = name.toLowerCase();
  return ALLOWED_EXTENSIONS.some((ext) => lower.endsWith(ext));
};

const getMimeFor = (fileName: string): string => {
  const lower = fileName.toLowerCase();
  if (lower.endsWith(".pdf")) return "application/pdf";
  if (lower.endsWith(".docx"))
    return "application/vnd.openxmlformats-officedocument.wordprocessingml.document";
  if (lower.endsWith(".md")) return "text/markdown";
  if (lower.endsWith(".txt")) return "text/plain";
  return "application/octet-stream";
};

export interface UploadTabProps {
  pickedFile: File | null;
  submitting: boolean;
  inlineError: string | null;
  onPickFile: (file: File | null) => void;
  onValidationError: (message: string | null) => void;
  onRegistered: (resumeId: string, sourceLabel: string) => void;
  setSubmitting: (value: boolean) => void;
  setInlineError: (message: string | null) => void;
}

export const UploadTab: FC<UploadTabProps> = ({
  pickedFile,
  submitting,
  inlineError,
  onPickFile,
  onValidationError,
  onRegistered,
  setSubmitting,
  setInlineError,
}) => {
  const { t, lang } = useI18n();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [uploadingMessage, setUploadingMessage] = useState<string | null>(null);
  const upload = useResumePresignUpload();
  const register = useResumeRegistration();

  const performUpload = useCallback(
    async (file: File) => {
      setSubmitting(true);
      setInlineError(null);
      setUploadingMessage(t("resumeWorkshop.create.upload.uploading"));
      try {
        const presigned = await upload.uploadFile(file, {
          contentType: file.type || getMimeFor(file.name),
        });
        const title = deriveDefaultTitle("upload", lang, file.name);
        const registered = await register.register({
          sourceType: "upload",
          fileObjectId: presigned.fileObjectId,
          title,
          language: lang,
        });
        onRegistered(registered.resumeId, file.name);
      } catch (error) {
        const message =
          error instanceof Error
            ? mapUploadError(error.message, t)
            : t("resumeWorkshop.create.errors.uploadFailed");
        setInlineError(message);
      } finally {
        setSubmitting(false);
        setUploadingMessage(null);
      }
    },
    [
      lang,
      onRegistered,
      register,
      setInlineError,
      setSubmitting,
      t,
      upload,
    ],
  );

  const onFileChange = (file: File | null) => {
    if (!file) {
      onPickFile(null);
      return;
    }
    if (!hasAllowedExtension(file.name)) {
      onPickFile(null);
      onValidationError(t("resumeWorkshop.create.errors.extensionInvalid"));
      return;
    }
    if (file.size > MAX_RESUME_UPLOAD_BYTES) {
      onPickFile(null);
      onValidationError(
        t("resumeWorkshop.create.errors.sizeExceeded").replace(
          "{maxMb}",
          `${Math.round(MAX_RESUME_UPLOAD_BYTES / (1024 * 1024))}`,
        ),
      );
      return;
    }
    onPickFile(file);
    void performUpload(file);
  };

  return (
    <div
      className="ei-resume-create-upload"
      data-testid="resume-create-upload-panel"
    >
      <div
        className="ei-resume-create-upload-dropzone"
        data-testid="resume-create-upload-dropzone"
      >
        <div className="ei-resume-create-upload-icon" aria-hidden="true">
          <ResumeWorkshopIcon name="upload" size={24} />
        </div>
        <div className="ei-text-title ei-resume-create-upload-title">
          {t("resumeWorkshop.create.upload.dropzoneTitle")}
        </div>
        <p className="ei-resume-create-upload-body">
          {t("resumeWorkshop.create.upload.dropzoneBody")}
        </p>
        <input
          ref={inputRef}
          type="file"
          accept=".pdf,.docx,.md,.txt"
          data-testid="resume-create-upload-input"
          className="ei-resume-create-upload-input"
          onChange={(event) => {
            const f = event.target.files?.[0] ?? null;
            // Reset value so selecting the same file twice still fires.
            event.target.value = "";
            onFileChange(f);
          }}
        />
        <button
          type="button"
          className="ei-resume-create-cta-accent"
          data-testid="resume-create-upload-choose"
          disabled={submitting}
          onClick={() => inputRef.current?.click()}
        >
          <ResumeWorkshopIcon name="upload" size={14} />
          {t("resumeWorkshop.create.upload.choose")}
        </button>
        {pickedFile ? (
          <div
            className="ei-resume-create-upload-selected"
            data-testid="resume-create-upload-selected"
          >
            {t("resumeWorkshop.create.upload.selectedPrefix")}
            <span className="ei-resume-create-upload-selected-name">
              {pickedFile.name}
            </span>
          </div>
        ) : null}
        {uploadingMessage ? (
          <div
            className="ei-resume-create-upload-progress"
            data-testid="resume-create-upload-progress"
            role="status"
            aria-live="polite"
          >
            {uploadingMessage}
          </div>
        ) : null}
      </div>
      {inlineError ? (
        <div
          className="ei-resume-create-error"
          role="alert"
          data-testid="resume-create-upload-error"
        >
          {inlineError}
        </div>
      ) : null}
    </div>
  );
};

function mapUploadError(
  message: string,
  t: (key: Parameters<ReturnType<typeof useI18n>["t"]>[0]) => string,
): string {
  if (/VALIDATION_FAILED/i.test(message)) {
    return t("resumeWorkshop.create.errors.validation");
  }
  if (/REGISTER_FAILED/i.test(message)) {
    return t("resumeWorkshop.create.errors.registerFailed");
  }
  return t("resumeWorkshop.create.errors.uploadFailed");
}
