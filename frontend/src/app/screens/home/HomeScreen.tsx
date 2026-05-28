import { useCallback, useEffect, useMemo, useState, type FC } from "react";

import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import { useRequestAuth } from "../../auth/useRequestAuth";
import { useI18n } from "../../i18n/messages";
import { interviewContextFromTargetJob } from "../../navigation/interviewContext";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { JDAssistModal, type JDAssistModalSource } from "./JDAssistModal";
import { MockInterviewCard } from "./MockInterviewCard";
import {
  consumePendingImportSource,
  storePendingImportSource,
  type PendingImportSource,
} from "./pendingImportState";
import { useRecentTargetJobs } from "./useRecentTargetJobs";

function idempotencyKey(): string {
  return `ik-${crypto.randomUUID()}`;
}

export const HomeScreen: FC<{ route: Route }> = ({ route }) => {
  const { lang, t } = useI18n();
  const runtime = useAppRuntimeOptional();
  const requestAuth = useRequestAuth();
  const { navigate } = useNavigation();
  const [input, setInput] = useState("");
  const [assistOpen, setAssistOpen] = useState<"upload" | "url" | null>(null);
  const [importing, setImporting] = useState(false);
  const [importError, setImportError] = useState<string | null>(null);
  const { jobs: rawJobs, loading, error } = useRecentTargetJobs();
  const targetLanguage = lang === "zh" ? "zh-CN" : "en";
  const routeResumeVersionId =
    typeof route.params.resumeVersionId === "string"
      ? route.params.resumeVersionId
      : undefined;
  const showRecentMocks = runtime?.auth.status === "authenticated";

  const jobs = useMemo(() => {
    const sorted = [...rawJobs].sort(
      (a, b) =>
        new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime(),
    );
    return sorted.slice(0, 12);
  }, [rawJobs]);

  const openProtectedRoute = useCallback(
    (next: Route, label: string) => {
      if (!runtime || runtime.auth.status === "authenticated") {
        navigate(next);
        return;
      }
      requestAuth({
        type: "open_protected_route",
        label,
        route: next.name,
        params: next.params,
      });
    },
    [navigate, requestAuth, runtime],
  );

  const submitImportSource = useCallback(async (source: PendingImportSource) => {
    if (!runtime) return;
    setImportError(null);
    setImporting(true);
    try {
      const ik = idempotencyKey();
      let targetJobId: string;

      if (source.source === "upload") {
        const presign = await runtime.client.createUploadPresign(
          {
            purpose: "target_job_attachment",
            fileName: "placeholder.pdf",
            contentType: "application/pdf",
            byteSize: 0,
          },
          { idempotencyKey: ik },
        );
        const result = await runtime.client.importTargetJob(
          {
            source: { type: "file", fileObjectId: presign.fileObjectId },
            targetLanguage,
          },
          { idempotencyKey: ik },
        );
        targetJobId = result.targetJobId;
        setAssistOpen(null);
      } else if (source.source === "url") {
        const result = await runtime.client.importTargetJob(
          {
            source: { type: "url", url: source.url },
            targetLanguage,
          },
          { idempotencyKey: ik },
        );
        targetJobId = result.targetJobId;
        setAssistOpen(null);
      } else {
        const result = await runtime.client.importTargetJob(
          {
            source: { type: "manual_text", rawText: source.rawText },
            targetLanguage,
          },
          { idempotencyKey: ik },
        );
        targetJobId = result.targetJobId;
      }
      navigate({
        name: "parse",
        params: {
          targetJobId,
          source: source.source,
          ...(routeResumeVersionId
            ? { resumeVersionId: routeResumeVersionId }
            : {}),
        },
      });
    } catch (err: unknown) {
      setImportError(
        err instanceof Error ? err.message : String(err),
      );
    } finally {
      setImporting(false);
    }
  }, [navigate, routeResumeVersionId, runtime, targetLanguage]);

  useEffect(() => {
    if (!runtime || runtime.auth.status !== "authenticated") return;
    const pendingImportId = route.params.pendingImportId;
    if (!pendingImportId) return;
    const pendingSource = consumePendingImportSource(pendingImportId);
    if (!pendingSource) {
      setImportError("Pending JD import expired. Please submit the JD again.");
      return;
    }
    void submitImportSource(pendingSource);
  }, [route.params.pendingImportId, runtime, submitImportSource]);

  const handlePasteImport = async () => {
    if (!runtime || !input.trim() || importing) return;
    const source: PendingImportSource = { source: "paste", rawText: input };

    if (runtime.auth.status !== "authenticated") {
      const pendingImportId = storePendingImportSource(source);
      requestAuth({
        type: "import_jd",
        label: t("home.importBtn"),
        route: "home",
        params: {
          source: "paste",
          pendingImportId,
          ...(routeResumeVersionId
            ? { resumeVersionId: routeResumeVersionId }
            : {}),
        },
      });
      return;
    }

    await submitImportSource(source);
  };

  const handleModalConfirm = async (source: JDAssistModalSource) => {
    if (!runtime || importing) return;
    const pendingSource: PendingImportSource = source;

    if (runtime.auth.status !== "authenticated") {
      const pendingImportId = storePendingImportSource(pendingSource);
      requestAuth({
        type: "import_jd",
        label: t("home.importBtn"),
        route: "home",
        params: {
          source: source.source,
          pendingImportId,
          ...(routeResumeVersionId
            ? { resumeVersionId: routeResumeVersionId }
            : {}),
        },
      });
      return;
    }

    await submitImportSource(pendingSource);
  };

  return (
    <section
      data-testid={`route-${route.name}`}
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
      className="ei-screen-shell"
      style={{ padding: "48px 56px 96px" }}
    >
      {/* Hero / import */}
      <div style={{ marginBottom: 56 }}>
        <div
          data-testid="home-hero-label"
          style={{
            color: "var(--ei-color-fg-tertiary)",
            marginBottom: 14,
            fontSize: 11,
            fontWeight: 500,
            letterSpacing: "0.08em",
            textTransform: "uppercase",
            fontFamily: "var(--ei-font-mono)",
          }}
        >
          {t("home.heroLabel")}
        </div>
        <h1
          data-testid="home-hero-title"
          className="ei-text-display"
          style={{ margin: 0, maxWidth: 820, textWrap: "balance" }}
        >
          {t("home.heroTitle")}
        </h1>
        <p
          data-testid="home-hero-sub"
          style={{
            fontSize: 15.5,
            color: "var(--ei-color-fg-secondary)",
            maxWidth: 620,
            marginTop: 16,
            lineHeight: 1.55,
          }}
        >
          {t("home.heroSub")}
        </p>

        {/* JD textarea card */}
        <div
          style={{
            marginTop: 32,
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 3,
            padding: 20,
          }}
        >
          <textarea
            data-testid="home-jd-textarea"
            aria-label={t("home.jdPlaceholder")}
            placeholder={t("home.jdPlaceholder")}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            style={{
              width: "100%",
              minHeight: 120,
              border: "none",
              outline: "none",
              resize: "vertical",
              fontSize: 14.5,
              lineHeight: 1.6,
              color: "var(--ei-color-fg-primary)",
              background: "transparent",
              fontFamily: "var(--ei-font-sans)",
            }}
          />
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginTop: 10,
              paddingTop: 14,
              borderTop: "1px dotted var(--ei-color-rule-strong)",
            }}
          >
            <div
              style={{ display: "flex", gap: 12, alignItems: "center" }}
            >
              <button
                data-testid="home-upload-trigger"
                type="button"
                onClick={() => setAssistOpen("upload")}
                style={{
                  background: "transparent",
                  border: "none",
                  color: "var(--ei-color-fg-tertiary)",
                  fontSize: 13,
                  display: "flex",
                  alignItems: "center",
                  gap: 6,
                  padding: 0,
                  cursor: "pointer",
                }}
              >
                {t("home.orUpload")}
              </button>
              <span style={{ color: "var(--ei-color-rule-strong)" }}>·</span>
              <button
                type="button"
                onClick={() => setAssistOpen("url")}
                style={{
                  background: "transparent",
                  border: "none",
                  color: "var(--ei-color-fg-tertiary)",
                  fontSize: 13,
                  display: "flex",
                  alignItems: "center",
                  gap: 6,
                  padding: 0,
                  cursor: "pointer",
                }}
              >
                URL
              </button>
            </div>
            <button
              data-testid="home-jd-submit"
              type="button"
              disabled={!input.trim() || importing}
              onClick={handlePasteImport}
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
              {t("home.importBtn")}
            </button>
          </div>
        </div>

        {importError && (
          <div
            data-testid="home-import-error"
            style={{
              color: "var(--ei-color-danger)",
              fontSize: 13,
              marginTop: 10,
            }}
          >
            {importError}
          </div>
        )}

        {/* Resume create CTA */}
        <div
          style={{
            marginTop: 16,
            display: "flex",
            gap: 16,
            flexWrap: "wrap",
          }}
        >
          <button
            data-testid="home-resume-create"
            type="button"
            onClick={() =>
              openProtectedRoute(
                { name: "resume_versions", params: { flow: "create" } },
                t("home.resumeCreateLink"),
              )
            }
            style={{
              background: "transparent",
              border: "none",
              color: "var(--ei-color-accent)",
              fontSize: 13,
              padding: 0,
              cursor: "pointer",
              fontWeight: 500,
            }}
          >
            {t("home.resumeCreateLink")}
          </button>
        </div>
      </div>

      {/* Recent mock interviews */}
      {showRecentMocks && (
        <div data-testid="home-recent-mocks" style={{ marginBottom: 48 }}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "flex-end",
              marginBottom: 16,
              gap: 20,
            }}
          >
            <div>
              <div
                style={{
                  color: "var(--ei-color-fg-tertiary)",
                  marginBottom: 8,
                  fontSize: 11,
                  fontWeight: 500,
                  letterSpacing: "0.08em",
                  textTransform: "uppercase",
                  fontFamily: "var(--ei-font-mono)",
                }}
              >
                RECENT
              </div>
              <div
                style={{
                  fontSize: 22,
                  color: "var(--ei-color-fg-primary)",
                  fontFamily: "var(--ei-font-serif)",
                  letterSpacing: "-0.02em",
                }}
              >
                {t("home.recentSection")}
              </div>
              <div
                style={{
                  fontSize: 13,
                  color: "var(--ei-color-fg-tertiary)",
                  marginTop: 4,
                }}
              >
                {t("home.recentSectionSub")}
              </div>
            </div>
          </div>
          {loading ? (
            <div className="ei-skeleton-stripe">
              {t("home.recentSection")}...
            </div>
          ) : error ? (
            <div
              style={{
                color: "var(--ei-color-danger)",
                fontSize: 13,
              }}
            >
              {error.message}
            </div>
          ) : jobs.length === 0 ? (
            <div
              style={{
                background: "var(--ei-color-bg-soft)",
                border: "1px solid var(--ei-color-rule-strong)",
                borderRadius: 3,
                padding: 32,
                textAlign: "center",
              }}
            >
              <p
                style={{
                  color: "var(--ei-color-fg-secondary)",
                  fontSize: 14,
                  margin: 0,
                }}
              >
                {t("home.recentSection")}
              </p>
            </div>
          ) : (
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fill, minmax(320px, 1fr))",
                gap: 16,
              }}
            >
              {jobs.map((j) => (
                <MockInterviewCard
                  key={j.id}
                  job={j}
                  onClick={() =>
                    openProtectedRoute(
                      {
                        name: "workspace",
                        params: interviewContextFromTargetJob(
                          j,
                        ) as unknown as Record<string, string>,
                      },
                      j.title,
                    )
                  }
                />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Auxiliary cards */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: 16,
        }}
      >
        {/* JOB PICKS card */}
        <div
          data-testid="home-aux-jobpicks"
          style={{
            background: "var(--ei-color-bg-soft)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 3,
            padding: 24,
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            gap: 20,
            flexWrap: "wrap",
          }}
        >
          <div style={{ flex: 1, minWidth: 260 }}>
            <div
              style={{
                color: "var(--ei-color-accent)",
                marginBottom: 6,
                fontSize: 11,
                fontWeight: 500,
                letterSpacing: "0.08em",
                textTransform: "uppercase",
                fontFamily: "var(--ei-font-mono)",
              }}
            >
              JOB PICKS
            </div>
            <div
              style={{
                fontSize: 20,
                color: "var(--ei-color-fg-primary)",
                fontFamily: "var(--ei-font-serif)",
                letterSpacing: "-0.02em",
              }}
            >
              {t("home.jobPicksTitle")}
            </div>
            <div
              style={{
                fontSize: 13.5,
                color: "var(--ei-color-fg-secondary)",
                marginTop: 4,
                lineHeight: 1.55,
              }}
            >
              {t("home.jobPicksSub")}
            </div>
          </div>
          <button
            type="button"
            onClick={() =>
              openProtectedRoute(
                { name: "jd_match", params: {} },
                t("home.jobPicksTitle"),
              )
            }
            style={{
              background: "var(--ei-color-bg-canvas)",
              color: "var(--ei-color-fg-primary)",
              border: "1px solid var(--ei-color-rule-strong)",
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
            {t("home.jobPicksBtn")}
          </button>
        </div>

        {/* POST-INTERVIEW card */}
        <div
          data-testid="home-aux-debrief"
          style={{
            background: "var(--ei-color-bg-soft)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 3,
            padding: 24,
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            gap: 20,
            flexWrap: "wrap",
          }}
        >
          <div style={{ flex: 1, minWidth: 260 }}>
            <div
              style={{
                color: "var(--ei-color-accent)",
                marginBottom: 6,
                fontSize: 11,
                fontWeight: 500,
                letterSpacing: "0.08em",
                textTransform: "uppercase",
                fontFamily: "var(--ei-font-mono)",
              }}
            >
              POST-INTERVIEW
            </div>
            <div
              style={{
                fontSize: 20,
                color: "var(--ei-color-fg-primary)",
                fontFamily: "var(--ei-font-serif)",
                letterSpacing: "-0.02em",
              }}
            >
              {t("home.debriefTitle")}
            </div>
            <div
              style={{
                fontSize: 13.5,
                color: "var(--ei-color-fg-secondary)",
                marginTop: 4,
              }}
            >
              {t("home.debriefSub")}
            </div>
          </div>
          <button
            type="button"
            onClick={() =>
              openProtectedRoute(
                { name: "debrief", params: {} },
                t("home.debriefTitle"),
              )
            }
            style={{
              background: "var(--ei-color-bg-canvas)",
              color: "var(--ei-color-fg-primary)",
              border: "1px solid var(--ei-color-rule-strong)",
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
            {t("home.debriefBtn")}
          </button>
        </div>
      </div>

      {assistOpen && (
        <JDAssistModal
          type={assistOpen}
          onClose={() => setAssistOpen(null)}
          onConfirm={handleModalConfirm}
        />
      )}
    </section>
  );
};
