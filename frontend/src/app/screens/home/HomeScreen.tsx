import { useCallback, useEffect, useMemo, useState, type FC } from "react";

import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import { useRequestAuth } from "../../auth/useRequestAuth";
import { useI18n } from "../../i18n/messages";
import { isSelectableInterviewResume } from "../../interview-context/selectableResume";
import { startPracticeFromParams } from "../../interview-context/startPractice";
import { useNavigation } from "../../navigation/NavigationProvider";
import {
  isTargetJobPracticeStartable,
  targetJobDetailRouteParams,
  targetJobPracticeRouteParams,
} from "../../navigation/interviewContext";
import type { Route } from "../../routes";
import { JDAssistModal, type JDAssistModalSource } from "./JDAssistModal";
import { MockInterviewCard } from "./MockInterviewCard";
import {
  consumePendingImportSource,
  storePendingImportSource,
  type PendingImportSource,
} from "./pendingImportState";
import { useRecentTargetJobs } from "./useRecentTargetJobs";
import type { Resume, TargetJob } from "../../../api/generated/types";

function idempotencyKey(): string {
  return `ik-${crypto.randomUUID()}`;
}

function sortByMostRecentResume(a: Resume, b: Resume): number {
  return Date.parse(b.updatedAt) - Date.parse(a.updatedAt);
}

function resumeMeta(resume: Resume): string {
  return [resume.language, resume.sourceType, resume.updatedAt.slice(0, 10)]
    .filter(Boolean)
    .join(" · ");
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
  const [readyResumes, setReadyResumes] = useState<Resume[]>([]);
  const [selectedResumeId, setSelectedResumeId] = useState("");
  const [resumesLoading, setResumesLoading] = useState(false);
  const [resumeError, setResumeError] = useState<string | null>(null);
  const [startingRecentJobId, setStartingRecentJobId] = useState<string | null>(null);
  const [recentStartError, setRecentStartError] = useState<string | null>(null);
  const { jobs: rawJobs, loading, error } = useRecentTargetJobs();
  const targetLanguage = lang === "zh" ? "zh-CN" : "en";
  const routeResumeId =
    typeof route.params.resumeId === "string"
      ? route.params.resumeId
      : undefined;
  const showRecentMocks = runtime?.auth.status === "authenticated";
  const selectedResume = useMemo(
    () => readyResumes.find((resume) => resume.id === selectedResumeId) ?? null,
    [readyResumes, selectedResumeId],
  );
  const canSubmit = Boolean(input.trim()) && Boolean(selectedResume) && !importing;

  const sortedJobs = useMemo(() => {
    return [...rawJobs].sort(
      (a, b) =>
        new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime(),
    );
  }, [rawJobs]);
  const jobs = useMemo(() => sortedJobs.slice(0, 3), [sortedJobs]);
  const hasMoreRecentMocks = sortedJobs.length > jobs.length;

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

  useEffect(() => {
    if (!runtime || runtime.auth.status !== "authenticated") {
      setReadyResumes([]);
      setSelectedResumeId("");
      setResumesLoading(false);
      setResumeError(null);
      return;
    }

    let active = true;
    setResumesLoading(true);
    setResumeError(null);

    runtime.client
      .listResumes({ headers: { "Accept-Language": lang } })
      .then((page) => {
        if (!active) return;
        const ready = page.items
          .filter(isSelectableInterviewResume)
          .sort(sortByMostRecentResume);
        setReadyResumes(ready);
        setSelectedResumeId((current) => {
          if (current && ready.some((resume) => resume.id === current)) {
            return current;
          }
          if (
            routeResumeId &&
            ready.some((resume) => resume.id === routeResumeId)
          ) {
            return routeResumeId;
          }
          return "";
        });
      })
      .catch((err: unknown) => {
        if (!active) return;
        setReadyResumes([]);
        setSelectedResumeId("");
        setResumeError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => {
        if (active) setResumesLoading(false);
      });

    return () => {
      active = false;
    };
  }, [lang, routeResumeId, runtime]);

  const submitImportSource = useCallback(async (source: PendingImportSource, resumeId: string) => {
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
            fileName: "target-job-attachment.pdf",
            contentType: "application/pdf",
            byteSize: 0,
          },
          { idempotencyKey: ik },
        );
        const result = await runtime.client.importTargetJob(
          {
            source: { type: "file", fileObjectId: presign.fileObjectId },
            targetLanguage,
            resumeId,
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
            resumeId,
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
            resumeId,
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
          resumeId,
        },
      });
    } catch (err: unknown) {
      setImportError(
        err instanceof Error ? err.message : String(err),
      );
    } finally {
      setImporting(false);
    }
  }, [navigate, runtime, targetLanguage]);

  useEffect(() => {
    if (!runtime || runtime.auth.status !== "authenticated") return;
    const pendingImportId = route.params.pendingImportId;
    if (!pendingImportId) return;
    const resumeId =
      typeof route.params.resumeId === "string" ? route.params.resumeId : "";
    if (!resumeId) {
      setImportError(t("home.resumeRequired"));
      return;
    }
    const pendingSource = consumePendingImportSource(pendingImportId);
    if (!pendingSource) {
      setImportError("Pending JD import expired. Please submit the JD again.");
      return;
    }
    void submitImportSource(pendingSource, resumeId);
  }, [route.params.pendingImportId, route.params.resumeId, runtime, submitImportSource, t]);

  const handlePasteImport = async () => {
    if (!runtime || !input.trim() || importing || !selectedResume) return;
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
          resumeId: selectedResume.id,
        },
      });
      return;
    }

    await submitImportSource(source, selectedResume.id);
  };

  const handleModalConfirm = async (source: JDAssistModalSource) => {
    if (!runtime || importing || !selectedResume) return;
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
          resumeId: selectedResume.id,
        },
      });
      return;
    }

    await submitImportSource(pendingSource, selectedResume.id);
  };

  const openRecentPlan = useCallback(
    (job: TargetJob) => {
      openProtectedRoute(
        {
          name: "parse",
          params: targetJobDetailRouteParams(job),
        },
        job.title,
      );
    },
    [openProtectedRoute],
  );

  const startRecentInterview = useCallback(
    async (job: TargetJob) => {
      const params = targetJobPracticeRouteParams(job);
      if (
        !runtime ||
        runtime.auth.status !== "authenticated" ||
        !params.resumeId ||
        !params.roundId
      ) {
        openRecentPlan(job);
        return;
      }

      setRecentStartError(null);
      setStartingRecentJobId(job.id);
      try {
        const started = await startPracticeFromParams(runtime.client, params, lang);
        navigate({ name: "practice", params: started.params });
      } catch (err: unknown) {
        setRecentStartError(err instanceof Error ? err.message : String(err));
      } finally {
        setStartingRecentJobId(null);
      }
    },
    [lang, navigate, openRecentPlan, runtime],
  );

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
        {/* JD source choice */}
        <div
          style={{
            marginTop: 32,
          }}
        >
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
            {t("home.pasteSource")}
          </div>
          <div
            data-testid="home-jd-input-card"
            style={{
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
              data-testid="home-jd-source-controls"
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                gap: 12,
                flexWrap: "wrap",
                marginTop: 14,
                paddingTop: 14,
                borderTop: "1px solid var(--ei-color-rule-strong)",
              }}
            >
              <div
                style={{
                  color: "var(--ei-color-fg-tertiary)",
                  fontSize: 12.5,
                  lineHeight: 1.5,
                }}
              >
                {t("home.orUpload")}
              </div>
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 10,
                  flexWrap: "wrap",
                }}
              >
                <button
                  data-testid="home-upload-trigger"
                  type="button"
                  onClick={() => setAssistOpen("upload")}
                  style={{
                    background: "var(--ei-color-bg-soft)",
                    border: "1px solid var(--ei-color-rule-strong)",
                    borderRadius: 3,
                    color: "var(--ei-color-fg-primary)",
                    fontSize: 13,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    minHeight: 34,
                    padding: "0 12px",
                    cursor: "pointer",
                    fontWeight: 500,
                  }}
                >
                  {t("home.uploadSource")}
                </button>
                <button
                  data-testid="home-url-trigger"
                  type="button"
                  onClick={() => setAssistOpen("url")}
                  style={{
                    background: "transparent",
                    border: "1px solid transparent",
                    color: "var(--ei-color-accent)",
                    fontSize: 13,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    minHeight: 34,
                    padding: "0 12px",
                    cursor: "pointer",
                    fontWeight: 500,
                  }}
                >
                  URL
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Resume selection and submit */}
        <div
          style={{
            marginTop: 16,
          }}
        >
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
            {t("home.resumeSelect")}
          </div>
          <div
            data-testid="home-resume-row"
            style={{
              display: "flex",
              gap: 14,
              flexWrap: "wrap",
              alignItems: "center",
            }}
          >
            <select
              data-testid="home-resume-select"
              aria-label={t("home.resumeSelect")}
              value={selectedResumeId}
              disabled={resumesLoading || readyResumes.length === 0}
              onChange={(event) => setSelectedResumeId(event.target.value)}
              style={{
                width: 360,
                maxWidth: "100%",
                flex: "0 1 360px",
                boxSizing: "border-box",
                minHeight: 42,
                border: "1px solid var(--ei-color-rule-strong)",
                borderRadius: 3,
                background: "var(--ei-color-bg-card)",
                color: "var(--ei-color-fg-primary)",
                fontSize: 13.5,
                fontFamily: "var(--ei-font-sans)",
                padding: "0 12px",
                outline: "none",
                cursor:
                  resumesLoading || readyResumes.length === 0
                    ? "not-allowed"
                    : "pointer",
              }}
            >
              <option value="">
                {resumesLoading
                  ? t("home.resumeLoading")
                  : t("home.resumeSelectPlaceholder")}
              </option>
              {readyResumes.map((resume) => (
                <option
                  key={resume.id}
                  data-testid={`home-resume-option-${resume.id}`}
                  value={resume.id}
                >
                  {`${resume.displayName || resume.title} · ${resumeMeta(resume)}`}
                </option>
              ))}
            </select>
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
                minHeight: 42,
                display: "flex",
                alignItems: "center",
              }}
            >
              {t("home.resumeCreateLink")}
            </button>
          </div>
          {!resumesLoading && readyResumes.length === 0 && (
            <div
              data-testid="home-resume-empty"
              style={{
                marginTop: 8,
                border: "1px dashed var(--ei-color-rule-strong)",
                borderRadius: 3,
                padding: "10px 12px",
                color: "var(--ei-color-fg-tertiary)",
                fontSize: 13,
                maxWidth: 360,
              }}
            >
              {resumeError ?? t("home.resumeEmpty")}
            </div>
          )}
          <div
            data-testid="home-resume-selection-status"
            style={{
              marginTop: 8,
              fontSize: 12.5,
              color: selectedResume
                ? "var(--ei-color-fg-secondary)"
                : "var(--ei-color-fg-tertiary)",
            }}
          >
            {selectedResume
              ? `${t("home.resumeSelected")} · ${selectedResume.displayName || selectedResume.title}`
              : t("home.resumeSelectHint")}
          </div>
          <div
            data-testid="home-submit-row"
            style={{
              marginTop: 14,
              display: "flex",
            }}
          >
            <button
              data-testid="home-jd-submit"
              type="button"
              disabled={!canSubmit}
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
            {hasMoreRecentMocks && (
              <button
                data-testid="home-recent-more"
                type="button"
                onClick={() =>
                  openProtectedRoute(
                    { name: "workspace", params: {} },
                    t("home.recentMore"),
                  )
                }
                style={{
                  background: "transparent",
                  border: "none",
                  color: "var(--ei-color-accent)",
                  fontSize: 13,
                  fontWeight: 500,
                  padding: 0,
                  cursor: "pointer",
                  whiteSpace: "nowrap",
                }}
              >
                {t("home.recentMore")}
              </button>
            )}
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
              data-testid="home-recent-mock-grid"
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fill, minmax(300px, 360px))",
                justifyContent: "start",
                gap: 16,
              }}
            >
              {jobs.map((j) => (
                <MockInterviewCard
                  key={j.id}
                  job={j}
                  onClick={() => openRecentPlan(j)}
                  primaryAction={{
                    label: t("home.importBtn"),
                    testId: `home-recent-mock-start-${j.id}`,
                    onClick: () => startRecentInterview(j),
                    disabled:
                      startingRecentJobId === j.id ||
                      !isTargetJobPracticeStartable(j),
                  }}
                />
              ))}
            </div>
          )}
          {recentStartError ? (
            <div
              data-testid="home-recent-start-error"
              style={{
                color: "var(--ei-color-danger)",
                fontSize: 13,
                marginTop: 10,
              }}
            >
              {recentStartError}
            </div>
          ) : null}
        </div>
      )}

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
