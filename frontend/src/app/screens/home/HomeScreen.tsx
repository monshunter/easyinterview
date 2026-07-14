import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type FC,
} from "react";

import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import { useRequestAuth } from "../../auth/useRequestAuth";
import { useI18n, type MessageKey } from "../../i18n/messages";
import { isSelectableInterviewResume } from "../../interview-context/selectableResume";
import { startPracticeFromParams } from "../../interview-context/startPractice";
import { useNavigation } from "../../navigation/NavigationProvider";
import {
  isTargetJobPracticeStartable,
  targetJobDetailRouteParams,
  targetJobPracticeRouteParams,
} from "../../navigation/interviewContext";
import type { Route } from "../../routes";
import { MockInterviewCard } from "./MockInterviewCard";
import {
  consumePendingImportIntent,
  storePendingImportIntent,
  type PendingImportIntent,
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
  const [importing, setImporting] = useState(false);
  const [importError, setImportError] = useState<MessageKey | null>(null);
  const [readyResumes, setReadyResumes] = useState<Resume[]>([]);
  const [selectedResumeId, setSelectedResumeId] = useState("");
  const [resumesLoading, setResumesLoading] = useState(false);
  const [resumeError, setResumeError] = useState<MessageKey | null>(null);
  const [startingRecentJobId, setStartingRecentJobId] = useState<string | null>(null);
  const [recentStartError, setRecentStartError] = useState<MessageKey | null>(null);
  const handledPendingImportId = useRef<string | null>(null);
  const { jobs: rawJobs, loading, error } = useRecentTargetJobs();
  const targetLanguage = lang === "zh" ? "zh-CN" : "en";
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
          return "";
        });
      })
      .catch(() => {
        if (!active) return;
        setReadyResumes([]);
        setSelectedResumeId("");
        setResumeError("home.errors.resumeLoad");
      })
      .finally(() => {
        if (active) setResumesLoading(false);
      });

    return () => {
      active = false;
    };
  }, [lang, runtime]);

  const submitImport = useCallback(async (intent: PendingImportIntent) => {
    if (!runtime) return;
    setImportError(null);
    setImporting(true);
    try {
      const result = await runtime.client.importTargetJob(
        {
          rawText: intent.rawText,
          targetLanguage: intent.targetLanguage,
          resumeId: intent.resumeId,
        },
        { idempotencyKey: intent.idempotencyKey },
      );
      navigate({
        name: "parse",
        params: {
          targetJobId: result.targetJobId,
          resumeId: intent.resumeId,
        },
      });
    } catch {
      setImportError("home.errors.import");
    } finally {
      setImporting(false);
    }
  }, [navigate, runtime]);

  useEffect(() => {
    if (!runtime || runtime.auth.status !== "authenticated") return;
    const opaquePendingImportId = route.params.opaquePendingImportId;
    if (
      !opaquePendingImportId ||
      handledPendingImportId.current === opaquePendingImportId
    ) {
      return;
    }
    handledPendingImportId.current = opaquePendingImportId;

    const intent = consumePendingImportIntent(opaquePendingImportId);
    if (!intent) {
      setImportError("home.pendingImportInvalid");
      navigate({ name: "home", params: {} });
      return;
    }
    void submitImport(intent);
  }, [navigate, route.params.opaquePendingImportId, runtime, submitImport]);

  const handlePasteImport = async () => {
    const rawText = input.trim();
    if (!runtime || !rawText || importing || !selectedResume) return;
    const intent: PendingImportIntent = {
      rawText,
      targetLanguage,
      resumeId: selectedResume.id,
      idempotencyKey: idempotencyKey(),
    };

    if (runtime.auth.status !== "authenticated") {
      const opaquePendingImportId = storePendingImportIntent(intent);
      requestAuth({
        type: "import_jd",
        label: t("home.importBtn"),
        route: "home",
        params: {
          opaquePendingImportId,
        },
      });
      return;
    }

    await submitImport(intent);
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
      } catch {
        setRecentStartError("home.errors.start");
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
        {/* JD paste intake */}
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
              {resumeError ? t(resumeError) : t("home.resumeEmpty")}
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
                border: "1px solid var(--ei-color-accent)",
                borderRadius: 2,
                padding: "0 16px",
                height: 38,
                fontSize: 14,
                fontWeight: 500,
                cursor: canSubmit ? "pointer" : "not-allowed",
                display: "inline-flex",
                alignItems: "center",
                justifyContent: "center",
                gap: 8,
                opacity: canSubmit ? 1 : 0.5,
                fontFamily: "var(--ei-sans)",
                letterSpacing: "-0.005em",
                transition: "transform .08s ease, opacity .15s",
              }}
            >
              {t("home.importBtn")}
              <svg
                aria-hidden="true"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                style={{ flexShrink: 0, display: "inline-block", verticalAlign: "middle" }}
              >
                <path d="M5 12h14M13 6l6 6-6 6" />
              </svg>
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
              {t(importError)}
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
              {t("home.errors.recentLoad")}
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
              {t(recentStartError)}
            </div>
          ) : null}
        </div>
      )}
    </section>
  );
};
