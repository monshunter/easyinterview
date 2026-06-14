import { type FC } from "react";

import { useI18n } from "../../i18n/messages";
import { useNavigation } from "../../navigation/NavigationProvider";

interface CompanyIntelEmbedProps {
  companyName?: string;
  locationText?: string | null;
  sourceType?: string;
  summary?: string;
  targetJobId: string;
  jdId?: string;
}

/**
 * Phase 5: CompanyIntelEmbed card — renders getTargetJob summary fields
 * only. Does NOT call getCompanyIntel (missing contract). Company intel is an
 * embedded-only surface, so the action stays on workspace with safe params.
 */
export const CompanyIntelEmbed: FC<CompanyIntelEmbedProps> = ({
  companyName,
  locationText,
  sourceType,
  summary,
  targetJobId,
  jdId,
}) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();

  return (
    <div
      data-testid="workspace-companyintel-summary"
      style={{
        background: "var(--ei-color-bgCard)",
        border: "1px solid var(--ei-color-rule)",
        borderRadius: 3,
        padding: 16,
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
      }}
    >
      <div>
        <div
          className="ei-label"
          style={{
            color: "var(--ei-color-ink3)",
            marginBottom: 6,
          }}
        >
          {t("workspace.intelLabel")}
        </div>
        <div
          className="ei-serif"
          style={{
            fontSize: 15,
            color: "var(--ei-color-ink)",
          }}
        >
          {[companyName, locationText].filter(Boolean).join(" · ") || t("workspace.intelTitle")}
        </div>
        <div
          style={{
            fontSize: 12,
            color: "var(--ei-color-ink3)",
            marginTop: 4,
            lineHeight: 1.5,
          }}
        >
          {summary
            ? summary.slice(0, 120) + (summary.length > 120 ? "…" : "")
            : t("workspace.intelSub")}
        </div>
      </div>
      <button
        data-testid="workspace-companyintel-open"
        onClick={() =>
          navigate({
            name: "workspace",
            params: { targetJobId, jdId: jdId || "" },
          })
        }
        style={{
          background: "transparent",
          border: "1px solid var(--ei-color-rule)",
          borderRadius: 2,
          color: "var(--ei-color-ink2)",
          padding: "5px 10px",
          fontSize: 12,
          cursor: "pointer",
          whiteSpace: "nowrap",
        }}
      >
        {t("workspace.intelOpen")}
      </button>
    </div>
  );
};
