import { useState, type FC } from "react";

import { useI18n } from "../../i18n/messages";
import type { Route } from "../../routes";

export const JDMatchScreen: FC<{ route: Route }> = ({ route }) => {
  const { t } = useI18n();
  const [tab] = useState<string>("recommended");

  const tabs = [
    { k: "recommended", t: t("jdMatch.tabRecommended") },
    { k: "search", t: t("jdMatch.tabSearch") },
    { k: "watchlist", t: t("jdMatch.tabWatchlist") },
  ];

  return (
    <section
      data-testid={`route-${route.name}`}
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
      className="ei-screen-shell"
      style={{ maxWidth: 1320, padding: "40px 48px 96px" }}
    >
      {/* Hero */}
      <div style={{ marginBottom: 28 }}>
        <div
          data-testid="jdmatch-hero-label"
          className="ei-label"
          style={{
            color: "var(--ei-color-fg-tertiary)",
            marginBottom: 8,
          }}
        >
          {t("jdMatch.heroLabel")}
        </div>
        <h1
          data-testid="jdmatch-hero-title"
          className="ei-serif"
          style={{
            fontSize: 38,
            margin: 0,
            color: "var(--ei-color-fg-primary)",
            letterSpacing: "-0.022em",
            lineHeight: 1.15,
            maxWidth: 820,
          }}
        >
          {t("jdMatch.heroTitle")}
        </h1>
        <div
          data-testid="jdmatch-hero-sub"
          style={{
            fontSize: 14,
            color: "var(--ei-color-fg-tertiary)",
            marginTop: 10,
            maxWidth: 720,
            lineHeight: 1.5,
          }}
        >
          {t("jdMatch.heroSub")}
        </div>
      </div>

      {/* Profile snapshot chip */}
      <div
        data-testid="jdmatch-profile-chip"
        style={{
          padding: "14px 20px",
          background: "var(--ei-color-bg-card)",
          border: "1px solid var(--ei-color-rule-strong)",
          borderLeft: "3px solid var(--ei-color-accent)",
          borderRadius: "var(--ei-radius-sm)",
          marginBottom: 24,
          display: "flex",
          alignItems: "center",
          gap: 20,
          flexWrap: "wrap",
        }}
      >
        <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
          <div
            style={{
              width: 32,
              height: 32,
              borderRadius: 16,
              background: "var(--ei-color-accent-soft)",
              color: "var(--ei-color-accent)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              fontFamily: "var(--ei-font-serif)",
              fontWeight: 600,
              fontSize: 14,
            }}
          >
            {t("jdMatch.profileInitials")}
          </div>
          <div>
            <div
              className="ei-label"
              style={{
                color: "var(--ei-color-fg-tertiary)",
                marginBottom: 2,
              }}
            >
              {t("jdMatch.searchingAs")}
            </div>
            <div
              data-testid="jdmatch-profile-chip-title"
              style={{
                fontSize: 13.5,
                color: "var(--ei-color-fg-primary)",
                fontWeight: 500,
              }}
            >
              {t("jdMatch.profileSummary")}
            </div>
          </div>
        </div>
        <div
          style={{
            height: 32,
            width: 1,
            background: "var(--ei-color-rule-strong)",
          }}
        />
        <div style={{ display: "flex", gap: 5, flexWrap: "wrap" }}>
          {["React", "TypeScript", "Node.js"].map((s) => (
            <span
              key={s}
              style={{
                padding: "2px 8px",
                fontSize: 11,
                fontFamily: "var(--ei-font-mono)",
                background: "var(--ei-color-bg-soft)",
                border: "1px solid var(--ei-color-rule-strong)",
                borderRadius: "var(--ei-radius-pill)",
                color: "var(--ei-color-fg-secondary)",
              }}
            >
              {s}
            </span>
          ))}
        </div>
        <div style={{ flex: 1, minWidth: 40 }} />
        <div
          style={{
            textAlign: "right",
            color: "var(--ei-color-fg-tertiary)",
            fontSize: 11.5,
            lineHeight: 1.45,
          }}
        >
          <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)" }}>
            {t("jdMatch.profileSources")}
          </div>
          <div>{t("jdMatch.profileSourcesStats")}</div>
        </div>
      </div>

      {/* Tabs */}
      <div
        style={{
          display: "flex",
          gap: 0,
          marginBottom: 20,
          borderBottom: "1px solid var(--ei-color-rule-strong)",
        }}
      >
        {tabs.map((tabItem) => (
          <button
            key={tabItem.k}
            data-testid={`jdmatch-tab-${tabItem.k}`}
            style={{
              padding: "12px 22px",
              background: "transparent",
              border: "none",
              borderBottom: `2px solid ${
                tab === tabItem.k
                  ? "var(--ei-color-accent)"
                  : "transparent"
              }`,
              color:
                tab === tabItem.k
                  ? "var(--ei-color-fg-primary)"
                  : "var(--ei-color-fg-tertiary)",
              cursor: "pointer",
              fontFamily: "var(--ei-font-sans)",
              fontSize: 13.5,
              fontWeight: tab === tabItem.k ? 500 : 400,
              marginBottom: -1,
            }}
          >
            {tabItem.t}
          </button>
        ))}
        <div style={{ flex: 1 }} />
        <div
          style={{
            alignSelf: "center",
            fontSize: 11.5,
            color: "var(--ei-color-fg-tertiary)",
            fontFamily: "var(--ei-font-mono)",
            letterSpacing: "0.04em",
            paddingRight: 6,
          }}
        >
          <span
            style={{
              display: "inline-block",
              width: 6,
              height: 6,
              borderRadius: 3,
              background: "var(--ei-color-ok)",
              marginRight: 6,
              verticalAlign: "middle",
            }}
          />
          {t("jdMatch.agentStatus")}
        </div>
      </div>

      {/* Placeholder content */}
      <div
        data-testid="jdmatch-placeholder"
        className="ei-screen-card"
        style={{
          padding: "48px 40px",
          textAlign: "center",
        }}
      >
        <div
          className="ei-serif"
          style={{
            fontSize: 24,
            color: "var(--ei-color-fg-primary)",
            marginBottom: 12,
          }}
        >
          {t("jdMatch.placeholderTitle")}
        </div>
        <div
          style={{
            fontSize: 14,
            color: "var(--ei-color-fg-tertiary)",
            maxWidth: 520,
            margin: "0 auto 24px",
            lineHeight: 1.6,
          }}
        >
          {t("jdMatch.placeholderCopy")}
        </div>
        <div
          data-testid="jdmatch-placeholder-cta"
          style={{
            fontSize: 12,
            color: "var(--ei-color-fg-muted)",
            fontFamily: "var(--ei-font-mono)",
          }}
        >
          {t("jdMatch.placeholderCta")}
        </div>
      </div>
    </section>
  );
};
