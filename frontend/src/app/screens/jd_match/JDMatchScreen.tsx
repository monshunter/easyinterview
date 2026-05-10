import { useCallback, useEffect, useMemo, useRef, useState, type FC } from "react";

import { useRequestAuth } from "../../auth/useRequestAuth";
import type { Lang } from "../../i18n/messages";
import { useI18n } from "../../i18n/messages";
import { useNavigation } from "../../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import type { JobMatchRecommendation } from "../../../api/generated/types";

import { RecommendedTab } from "./RecommendedTab";
import { useAgentScanStatus } from "./useAgentScanStatus";
import { useDismissRecommendation } from "./useDismissRecommendation";
import { useJobMatchProfile } from "./useJobMatchProfile";
import { useJobMatchRecommendations } from "./useJobMatchRecommendations";
import { useToggleWatchlist } from "./useToggleWatchlist";

type JdMatchAction =
  | "save"
  | "unsave"
  | "dismiss"
  | "confirm_interview"
  | "run_search"
  | "create_saved_search";

const PROFILE_INITIALS_FALLBACK = "—";

function computeInitials(name?: string | null): string {
  if (!name) return PROFILE_INITIALS_FALLBACK;
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return PROFILE_INITIALS_FALLBACK;
  return parts
    .slice(0, 2)
    .map((p) => (p[0] ?? "").toUpperCase())
    .join("");
}

function formatRelativeMinutes(
  iso: string,
  lang: Lang,
  direction: "past" | "future",
): string {
  const target = new Date(iso).getTime();
  if (Number.isNaN(target)) return "—";
  const now = Date.now();
  const diffMs = direction === "past" ? now - target : target - now;
  const minutes = Math.max(0, Math.round(diffMs / 60000));
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  if (days > 0) return lang === "en" ? `${days}d` : `${days} 天`;
  if (hours > 0) return lang === "en" ? `${hours}h` : `${hours} 小时`;
  return lang === "en" ? `${minutes}m` : `${minutes} 分钟`;
}

type AgentStatusEnum = "idle" | "scanning" | "error" | "loading";

function resolveAgentTone(status: string | undefined): AgentStatusEnum {
  if (status === "idle" || status === "scanning" || status === "error") {
    return status;
  }
  return "loading";
}

function agentToneColor(tone: AgentStatusEnum): string {
  switch (tone) {
    case "scanning":
      return "var(--ei-color-accent)";
    case "error":
      return "var(--ei-color-warn)";
    case "idle":
      return "var(--ei-color-ok)";
    case "loading":
    default:
      return "var(--ei-color-fg-tertiary)";
  }
}

export const JDMatchScreen: FC<{ route: Route }> = ({ route }) => {
  const { t, lang } = useI18n();
  const { navigate } = useNavigation();
  const requestAuth = useRequestAuth();
  const runtime = useAppRuntimeOptional();
  const isUnauthenticated =
    runtime?.auth.status === "unauthenticated" ||
    runtime?.auth.status === "loading";
  const [tab, setTab] = useState<string>("recommended");
  const profileQuery = useJobMatchProfile();
  const agentQuery = useAgentScanStatus(tab);
  const recsQuery = useJobMatchRecommendations();
  const profile = profileQuery.data;
  const agent = agentQuery.data;
  const tone = resolveAgentTone(agent?.status);
  const dotColor = agentToneColor(tone);

  // Local view-model overlays: dismissed ids hide cards; saved overrides toggle
  // the saved flag without mutating the upstream hook's items array. This keeps
  // optimistic updates revertible without touching cached server data.
  const [hiddenIds, setHiddenIds] = useState<Set<string>>(new Set());
  const [savedOverrides, setSavedOverrides] = useState<
    ReadonlyMap<string, boolean>
  >(new Map());
  const [selectedId, setSelectedId] = useState<string | null>(null);

  // First mount or list refresh: pick the first visible item by default.
  const visibleItems = useMemo<JobMatchRecommendation[]>(
    () =>
      recsQuery.items
        .filter((r) => !hiddenIds.has(r.id))
        .map((r) =>
          savedOverrides.has(r.id)
            ? { ...r, saved: savedOverrides.get(r.id) ?? r.saved }
            : r,
        ),
    [recsQuery.items, hiddenIds, savedOverrides],
  );

  useEffect(() => {
    if (visibleItems.length === 0) {
      if (selectedId !== null) setSelectedId(null);
      return;
    }
    if (
      selectedId == null ||
      !visibleItems.some((r) => r.id === selectedId)
    ) {
      setSelectedId(visibleItems[0]!.id);
    }
  }, [visibleItems, selectedId]);

  const applySaved = useCallback((id: string, savedNext: boolean) => {
    setSavedOverrides((prev) => {
      const next = new Map(prev);
      next.set(id, savedNext);
      return next;
    });
  }, []);

  const lastSelectedRef = useRef<string | null>(null);
  const applyHide = useCallback(
    (rec: JobMatchRecommendation) => {
      const wasSelectedAtCall = lastSelectedRef.current;
      setHiddenIds((prev) => {
        const next = new Set(prev);
        next.add(rec.id);
        return next;
      });
      return () => {
        setHiddenIds((prev) => {
          const next = new Set(prev);
          next.delete(rec.id);
          return next;
        });
        if (wasSelectedAtCall === rec.id) setSelectedId(rec.id);
      };
    },
    [],
  );

  // Track selected id snapshot for revert callbacks
  useEffect(() => {
    lastSelectedRef.current = selectedId;
  }, [selectedId]);

  const toggle = useToggleWatchlist({ applyOptimistic: applySaved });
  const dismissCtl = useDismissRecommendation({
    applyOptimisticHide: applyHide,
  });

  const requestAuthForJdMatch = useCallback(
    (rec: JobMatchRecommendation, action: JdMatchAction, label: string) => {
      requestAuth({
        type: "jd_match_action",
        label,
        route: "jd_match",
        params: {
          tab: "recommended",
          selectedJobMatchId: rec.id,
          action,
        },
      });
    },
    [requestAuth],
  );

  const handleConfirmInterview = useCallback(
    (rec: JobMatchRecommendation) => {
      if (isUnauthenticated) {
        requestAuthForJdMatch(
          rec,
          "confirm_interview",
          t("jdMatch.recommended.actionConfirm"),
        );
        return;
      }
      navigate({
        name: "parse",
        params: { source: "jd_match", sourceJobMatchId: rec.id },
      });
    },
    [navigate, isUnauthenticated, requestAuthForJdMatch, t],
  );

  const handleOpenSource = useCallback((rec: JobMatchRecommendation) => {
    if (!rec.sourceUrl) return;
    if (typeof window !== "undefined") {
      window.open(rec.sourceUrl, "_blank", "noopener,noreferrer");
    }
  }, []);

  const handleToggleSave = useCallback(
    (rec: JobMatchRecommendation) => {
      if (isUnauthenticated) {
        const action: JdMatchAction = rec.saved ? "unsave" : "save";
        requestAuthForJdMatch(
          rec,
          action,
          rec.saved
            ? t("jdMatch.recommended.actionUnsave")
            : t("jdMatch.recommended.actionSave"),
        );
        return;
      }
      void toggle.toggleSave(rec);
    },
    [toggle, isUnauthenticated, requestAuthForJdMatch, t],
  );

  const handleDismiss = useCallback(
    (rec: JobMatchRecommendation) => {
      if (isUnauthenticated) {
        requestAuthForJdMatch(
          rec,
          "dismiss",
          t("jdMatch.recommended.actionDismiss"),
        );
        return;
      }
      void dismissCtl.dismiss(rec);
    },
    [dismissCtl, isUnauthenticated, requestAuthForJdMatch, t],
  );

  const initials = computeInitials(profile?.displayName);
  const summaryParts: string[] = [];
  if (profile) {
    if (profile.displayName) summaryParts.push(profile.displayName);
    if (profile.yearsOfExperience != null) {
      summaryParts.push(
        `${profile.yearsOfExperience} ${t("jdMatch.profile.summaryYearsUnit")}`,
      );
    }
    if (profile.locationText) summaryParts.push(profile.locationText);
  }
  const summaryText = summaryParts.length
    ? summaryParts.join(" · ")
    : t("jdMatch.profile.searchingAsLoading");

  const recommendedCount = visibleItems.length;
  const tabs: Array<{ k: string; label: string; count: number | null }> = [
    {
      k: "recommended",
      label: t("jdMatch.tabRecommended"),
      count: recsQuery.loading ? null : recommendedCount,
    },
    { k: "search", label: t("jdMatch.tabSearch"), count: null },
    { k: "watchlist", label: t("jdMatch.tabWatchlist"), count: null },
  ];

  const agentStatusLabel = (() => {
    switch (tone) {
      case "scanning":
        return t("jdMatch.agent.statusScanning");
      case "error":
        return t("jdMatch.agent.statusError");
      case "idle":
        return t("jdMatch.agent.statusIdle");
      case "loading":
      default:
        return t("jdMatch.agent.statusLoading");
    }
  })();

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
            data-testid="jdmatch-profile-chip-avatar"
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
              overflow: "hidden",
            }}
          >
            {profile?.avatarUrl ? (
              <img
                src={profile.avatarUrl}
                alt=""
                style={{ width: "100%", height: "100%", objectFit: "cover" }}
              />
            ) : (
              initials
            )}
          </div>
          <div>
            <div
              className="ei-label"
              style={{
                color: "var(--ei-color-fg-tertiary)",
                marginBottom: 2,
              }}
            >
              {t("jdMatch.profile.searchingAs")}
            </div>
            <div
              data-testid="jdmatch-profile-chip-searching-as"
              style={{
                fontSize: 13.5,
                color: "var(--ei-color-fg-primary)",
                fontWeight: 500,
              }}
            >
              {summaryText}
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
        <div
          data-testid="jdmatch-profile-chip-skills"
          style={{ display: "flex", gap: 5, flexWrap: "wrap" }}
        >
          {(profile?.skills ?? []).map((skill, i) => (
            <span
              key={`${skill}-${i}`}
              data-testid={`jdmatch-profile-chip-skill-${i}`}
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
              {skill}
            </span>
          ))}
        </div>
        <div style={{ flex: 1, minWidth: 40 }} />
        <div
          data-testid="jdmatch-profile-chip-sources"
          style={{
            textAlign: "right",
            color: "var(--ei-color-fg-tertiary)",
            fontSize: 11.5,
            lineHeight: 1.45,
          }}
        >
          <div className="ei-label" style={{ color: "var(--ei-color-fg-tertiary)" }}>
            {t("jdMatch.profile.sourcesHeading")}
          </div>
          {profile?.sources ? (
            <div>
              {`${profile.sources.resumes} ${t("jdMatch.profile.sourcesUnitResumes")} · `}
              {`${profile.sources.jds} ${t("jdMatch.profile.sourcesUnitJds")} · `}
              {`${profile.sources.mocks} ${t("jdMatch.profile.sourcesUnitMocks")} · `}
              {`${profile.sources.debriefs} ${t("jdMatch.profile.sourcesUnitDebriefs")}`}
            </div>
          ) : (
            <div>{t("jdMatch.profile.sourcesEmpty")}</div>
          )}
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
            onClick={() => setTab(tabItem.k)}
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
              display: "flex",
              alignItems: "center",
              gap: 8,
            }}
          >
            <span>{tabItem.label}</span>
            {tabItem.count != null && (
              <span
                data-testid={`jdmatch-tab-${tabItem.k}-count`}
                style={{
                  fontFamily: "var(--ei-font-mono)",
                  fontSize: 10.5,
                  padding: "1px 6px",
                  borderRadius: 10,
                  background:
                    tab === tabItem.k
                      ? "var(--ei-color-accent-soft)"
                      : "var(--ei-color-bg-soft)",
                  color:
                    tab === tabItem.k
                      ? "var(--ei-color-accent)"
                      : "var(--ei-color-fg-tertiary)",
                }}
              >
                {tabItem.count}
              </span>
            )}
          </button>
        ))}
        <div style={{ flex: 1 }} />
        <div
          data-testid="jdmatch-agent-status-badge"
          data-tone={tone}
          style={{
            alignSelf: "center",
            fontSize: 11.5,
            color: "var(--ei-color-fg-tertiary)",
            fontFamily: "var(--ei-font-mono)",
            letterSpacing: "0.04em",
            paddingRight: 6,
            display: "flex",
            alignItems: "center",
            gap: 6,
          }}
        >
          <span
            style={{
              display: "inline-block",
              width: 6,
              height: 6,
              borderRadius: 3,
              background: dotColor,
              verticalAlign: "middle",
            }}
          />
          <span>{agentStatusLabel}</span>
          {tone === "error" && agent?.message ? (
            <span style={{ color: "var(--ei-color-warn)" }}>
              · {agent.message}
            </span>
          ) : null}
          {agent?.lastScanAt ? (
            <span data-testid="jdmatch-agent-status-last-scan">
              {" · "}
              {t("jdMatch.agent.lastScanLabel")}{" "}
              {formatRelativeMinutes(agent.lastScanAt, lang, "past")}
            </span>
          ) : null}
          {agent?.nextScanAt && tone !== "scanning" ? (
            <span data-testid="jdmatch-agent-status-next-scan">
              {" · "}
              {t("jdMatch.agent.nextScanLabel")}{" "}
              {formatRelativeMinutes(agent.nextScanAt, lang, "future")}
            </span>
          ) : null}
        </div>
      </div>

      {/* Tab body */}
      {tab === "recommended" ? (
        <RecommendedTab
          recommendations={visibleItems}
          loading={recsQuery.loading}
          error={recsQuery.error}
          selectedId={selectedId}
          onSelect={setSelectedId}
          onConfirmInterview={handleConfirmInterview}
          onToggleSave={handleToggleSave}
          onOpenSource={handleOpenSource}
          onMarkNotRelevant={handleDismiss}
          onRetry={recsQuery.retry}
        />
      ) : null}
    </section>
  );
};
