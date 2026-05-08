import { useEffect, useState, type FC } from "react";

import { useI18n } from "../../i18n/messages";

export interface JDAssistModalUploadSource {
  source: "upload";
}

export interface JDAssistModalUrlSource {
  source: "url";
  url: string;
}

export type JDAssistModalSource = JDAssistModalUploadSource | JDAssistModalUrlSource;

export interface JDAssistModalProps {
  type: "upload" | "url";
  onClose: () => void;
  onConfirm: (source: JDAssistModalSource) => void;
}

export const JDAssistModal: FC<JDAssistModalProps> = ({
  type,
  onClose,
  onConfirm,
}) => {
  const { t } = useI18n();
  const isUpload = type === "upload";
  const [url, setUrl] = useState("");

  const handleConfirm = () => {
    if (isUpload) {
      onConfirm({ source: "upload" });
    } else {
      onConfirm({ source: "url", url });
    }
  };

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [onClose]);

  return (
    <div
      data-testid={`home-modal-${type}-backdrop`}
      onClick={onClose}
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(24, 20, 16, 0.24)",
        zIndex: 80,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: 24,
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          width: "min(520px, 100%)",
          background: "var(--ei-color-bg-card)",
          border: "1px solid var(--ei-color-rule-strong)",
          borderRadius: 4,
          boxShadow: "0 24px 70px rgba(30, 22, 15, 0.24)",
          padding: 24,
        }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "flex-start",
            gap: 18,
            marginBottom: 18,
          }}
        >
          <div>
            <div
              style={{
                color: "var(--ei-color-fg-tertiary)",
                marginBottom: 6,
                fontSize: 11,
                fontWeight: 500,
                letterSpacing: "0.08em",
                textTransform: "uppercase",
                fontFamily: "var(--ei-font-mono)",
              }}
            >
              {t("home.modalLabel")}
            </div>
            <div
              style={{
                fontSize: 23,
                color: "var(--ei-color-fg-primary)",
                fontFamily: "var(--ei-font-serif)",
              }}
            >
              {isUpload ? t("home.modalUploadTitle") : t("home.modalUrlTitle")}
            </div>
          </div>
          <button
            data-testid={`home-modal-${type}-close`}
            type="button"
            onClick={onClose}
            aria-label="Close"
            style={{
              background: "transparent",
              border: "none",
              color: "var(--ei-color-fg-tertiary)",
              cursor: "pointer",
              padding: 4,
              display: "flex",
              alignItems: "center",
              lineHeight: 1,
              fontSize: 16,
            }}
          >
            ✕
          </button>
        </div>

        {isUpload ? (
          <div
            data-testid="home-modal-upload-dropzone"
            style={{
              border: "1px dashed var(--ei-color-rule-strong)",
              background: "var(--ei-color-bg-soft)",
              borderRadius: 3,
              padding: "30px 22px",
              textAlign: "center",
            }}
          >
            <span
              style={{ fontSize: 24, color: "var(--ei-color-accent)" }}
              aria-hidden
            >
              ↑
            </span>
            <div
              style={{
                fontSize: 15,
                color: "var(--ei-color-fg-primary)",
                marginTop: 12,
                fontWeight: 500,
              }}
            >
              {t("home.modalUploadDropzone")}
            </div>
            <div
              style={{
                fontSize: 12.5,
                color: "var(--ei-color-fg-tertiary)",
                marginTop: 6,
              }}
            >
              {t("home.modalUploadHint")}
            </div>
          </div>
        ) : (
          <div>
            <label
              style={{
                display: "block",
                color: "var(--ei-color-fg-tertiary)",
                marginBottom: 8,
                fontSize: 13,
              }}
            >
              {t("home.modalUrlLabel")}
            </label>
            <input
              data-testid="home-modal-url-input"
              placeholder={t("home.modalUrlPlaceholder")}
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              style={{
                width: "100%",
                boxSizing: "border-box",
                border: "1px solid var(--ei-color-rule-strong)",
                background: "var(--ei-color-bg-soft)",
                color: "var(--ei-color-fg-primary)",
                borderRadius: 3,
                padding: "12px 14px",
                fontSize: 14,
                outline: "none",
                fontFamily: "var(--ei-font-sans)",
              }}
            />
            <div
              style={{
                fontSize: 12.5,
                color: "var(--ei-color-fg-tertiary)",
                marginTop: 8,
              }}
            >
              {t("home.modalUrlHint")}
            </div>
          </div>
        )}

        <div
          style={{
            display: "flex",
            justifyContent: "flex-end",
            gap: 10,
            marginTop: 22,
          }}
        >
          <button
            data-testid={`home-modal-${type}-cancel`}
            type="button"
            onClick={onClose}
            style={{
              background: "transparent",
              border: "1px solid var(--ei-color-rule-strong)",
              color: "var(--ei-color-fg-primary)",
              borderRadius: "var(--ei-radius-sm)",
              padding: "0 16px",
              height: 38,
              fontSize: 14,
              fontWeight: 500,
              cursor: "pointer",
            }}
          >
            {t("home.modalCancel")}
          </button>
          <button
            data-testid={`home-modal-${type}-continue`}
            type="button"
            onClick={handleConfirm}
            style={{
              background: "var(--ei-color-accent)",
              color: "#fff",
              border: "none",
              borderRadius: "var(--ei-radius-sm)",
              padding: "0 16px",
              height: 38,
              fontSize: 14,
              fontWeight: 500,
              cursor: "pointer",
              display: "flex",
              alignItems: "center",
              gap: 8,
            }}
          >
            {t("home.modalContinue")}
          </button>
        </div>
      </div>
    </div>
  );
};
