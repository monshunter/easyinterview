import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

/**
 * Source-level mirror of ui-design/src/screens-p0-complete.jsx
 * lines 323-331 (eyebrow + serif headline + sub line).
 */
export const HeaderHero: FC = () => {
  const { t } = useI18n();
  return (
    <div>
      <div
        className="ei-label"
        data-testid="generating-header-eyebrow"
        style={{
          color: "var(--ei-ink3)",
          marginBottom: 12,
          letterSpacing: "0.1em",
        }}
      >
        {t("generating.header.eyebrow")}
      </div>
      <h1
        className="ei-serif"
        data-testid="generating-header-title"
        style={{
          fontSize: 34,
          margin: 0,
          color: "var(--ei-ink)",
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
          color: "var(--ei-ink3)",
          marginBottom: 32,
          lineHeight: 1.5,
          maxWidth: 540,
        }}
      >
        {t("generating.header.subtitle")}
      </div>
    </div>
  );
};
