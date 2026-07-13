import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

/**
 * Source-level mirror of ui-design/src/screens-p0-complete.jsx
 * lines 323-331 (eyebrow + serif headline + sub line).
 */
export const HeaderHero: FC<{ status: "queued" | "generating" }> = ({ status }) => {
  const { t } = useI18n();
  return (
    <div>
      <div
        className="ei-label"
        data-testid="generating-header-eyebrow"
        style={{
          color: "var(--ei-color-fg-tertiary)",
          marginBottom: 12,
          letterSpacing: "0.1em",
        }}
      >
        {t(status === "queued" ? "generating.status.queued" : "generating.status.generating")}
      </div>
      <h1
        className="ei-serif"
        data-testid="generating-header-title"
        style={{
          fontSize: 34,
          margin: 0,
          color: "var(--ei-color-fg-primary)",
          letterSpacing: "-0.02em",
          lineHeight: 1.2,
          marginBottom: 10,
        }}
      >
        {t("generating.header.title")}
      </h1>
      <div
        data-testid="generating-header-subtitle"
        style={{
          fontSize: 14,
          color: "var(--ei-color-fg-tertiary)",
          lineHeight: 1.65,
          maxWidth: 600,
        }}
      >
        {t("generating.header.subtitle")}
      </div>
    </div>
  );
};
