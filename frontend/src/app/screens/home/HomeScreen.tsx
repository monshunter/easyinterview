import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type FC,
} from "react";

import type { ResumeSummary, TargetJob } from "../../../api/generated/types";
import { resolveContentLimits, utf8ByteLength } from "../../../lib/contentLimits";
import { useRequestAuth } from "../../auth/useRequestAuth";
import { useI18n, type MessageKey } from "../../i18n/messages";
import { PracticeLaunchTransition } from "../../interview-context/PracticeLaunchTransition";
import { isSelectableInterviewResume } from "../../interview-context/selectableResume";
import { startPracticeFromParams } from "../../interview-context/startPractice";
import {
  isTargetJobPracticeStartable,
  targetJobDetailRouteParams,
  targetJobPracticeRouteParams,
} from "../../navigation/interviewContext";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import { MockInterviewCard } from "./MockInterviewCard";
import {
  consumePendingImportIntent,
  storePendingImportIntent,
  type PendingImportIntent,
} from "./pendingImportState";
import { useRecentTargetJobs } from "./useRecentTargetJobs";

function idempotencyKey(): string {
  return `ik-${crypto.randomUUID()}`;
}

function sortByMostRecentResume(a: ResumeSummary, b: ResumeSummary): number {
  return Date.parse(b.updatedAt) - Date.parse(a.updatedAt);
}

function resumeMeta(resume: ResumeSummary): string {
  return [resume.language, resume.sourceType, resume.updatedAt.slice(0, 10)]
    .filter(Boolean)
    .join(" · ");
}

function formatRecentDate(value: string, lang: "zh" | "en"): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value.slice(0, 10);
  return new Intl.DateTimeFormat(lang === "zh" ? "zh-CN" : "en-US", {
    month: "numeric",
    day: "numeric",
  }).format(date);
}

const HomeHeroIllustration: FC = () => (
  <svg
    data-testid="home-hero-illustration"
    className="ei-home-hero-illustration"
    viewBox="0 0 280 150"
    fill="none"
    aria-hidden="true"
  >
    <defs>
      <linearGradient id="home-ill-card" x1="0" y1="0" x2="1" y2="1">
        <stop stopColor="currentColor" stopOpacity="0.28" />
        <stop offset="1" stopColor="currentColor" stopOpacity="0.04" />
      </linearGradient>
    </defs>
    <path d="M34 68l4 12 12 4-12 4-4 12-4-12-12-4 12-4 4-12z" fill="currentColor" opacity=".25" />
    <rect x="92" y="22" width="166" height="116" rx="15" fill="url(#home-ill-card)" />
    <rect x="92" y="22" width="166" height="23" rx="15" fill="currentColor" opacity=".14" />
    <circle cx="108" cy="34" r="3" fill="currentColor" opacity=".42" />
    <circle cx="119" cy="34" r="3" fill="currentColor" opacity=".32" />
    <circle cx="130" cy="34" r="3" fill="currentColor" opacity=".24" />
    <rect x="108" y="50" width="112" height="41" rx="10" fill="var(--ei-color-bg-card)" opacity=".78" />
    <circle cx="124" cy="69" r="8" fill="currentColor" opacity=".22" />
    <rect x="140" y="63" width="40" height="9" rx="4.5" fill="currentColor" opacity=".3" />
    <rect x="108" y="100" width="74" height="7" rx="3.5" fill="currentColor" opacity=".13" />
    <rect x="108" y="114" width="56" height="7" rx="3.5" fill="currentColor" opacity=".09" />
    <rect x="218" y="105" width="12" height="33" rx="5" fill="currentColor" opacity=".25" />
    <rect x="235" y="92" width="12" height="46" rx="5" fill="currentColor" opacity=".38" />
    <rect x="65" y="92" width="58" height="35" rx="10" fill="currentColor" opacity=".34" />
    <circle cx="81" cy="109" r="3" fill="white" opacity=".75" />
    <circle cx="94" cy="109" r="3" fill="white" opacity=".75" />
    <circle cx="107" cy="109" r="3" fill="white" opacity=".75" />
  </svg>
);

export const HomeScreen: FC<{ route: Route }> = ({ route }) => {
  const { lang, t } = useI18n();
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const isAuthenticated = runtime?.auth.status === "authenticated";
  const requestAuth = useRequestAuth();
  const { navigate } = useNavigation();
  const [input, setInput] = useState("");
  const [importing, setImporting] = useState(false);
  const [importError, setImportError] = useState<MessageKey | null>(null);
  const [readyResumes, setReadyResumes] = useState<ResumeSummary[]>([]);
  const [selectedResumeId, setSelectedResumeId] = useState("");
  const [resumesLoading, setResumesLoading] = useState(false);
  const [resumeError, setResumeError] = useState<MessageKey | null>(null);
  const [startingRecentJobId, setStartingRecentJobId] = useState<string | null>(null);
  const [recentStartError, setRecentStartError] = useState<MessageKey | null>(null);
  const jdTextareaRef = useRef<HTMLTextAreaElement | null>(null);
  const handledPendingImportId = useRef<string | null>(null);
  const { jobs: rawJobs, loading, error } = useRecentTargetJobs();
  const targetLanguage = lang === "zh" ? "zh-CN" : "en";
  const contentLimits = resolveContentLimits(
    runtime?.runtime.status === "ready" ? runtime.runtime.config : undefined,
  );
  const selectedResume = useMemo(
    () => readyResumes.find((resume) => resume.id === selectedResumeId) ?? null,
    [readyResumes, selectedResumeId],
  );
  const canSubmit = Boolean(contentLimits) && Boolean(input.trim()) && Boolean(selectedResume) && !importing;
  const jobs = useMemo(
    () => [...rawJobs]
      .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
      .slice(0, 3),
    [rawJobs],
  );

  useLayoutEffect(() => {
    const textarea = jdTextareaRef.current;
    if (!textarea) return;

    textarea.style.height = "auto";
    textarea.style.height = `${textarea.scrollHeight}px`;
  }, [input]);

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
    if (!client || !isAuthenticated) {
      setReadyResumes([]);
      setSelectedResumeId("");
      setResumesLoading(false);
      setResumeError(null);
      return;
    }

    let active = true;
    setResumesLoading(true);
    setResumeError(null);
    client
      .listResumes({ headers: { "Accept-Language": lang } })
      .then((page) => {
        if (!active) return;
        const ready = page.items
          .filter(isSelectableInterviewResume)
          .sort(sortByMostRecentResume);
        setReadyResumes(ready);
        setSelectedResumeId((current) =>
          current && ready.some((resume) => resume.id === current) ? current : "",
        );
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
  }, [client, isAuthenticated, lang]);

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
      navigate({ name: "parse", params: { targetJobId: result.targetJobId } });
    } catch {
      setImportError("home.errors.import");
    } finally {
      setImporting(false);
    }
  }, [navigate, runtime]);

  useEffect(() => {
    if (!runtime || runtime.auth.status !== "authenticated") return;
    const opaquePendingImportId = route.params.opaquePendingImportId;
    if (!opaquePendingImportId || handledPendingImportId.current === opaquePendingImportId) return;
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
    if (!runtime || !contentLimits || !rawText || importing || !selectedResume) return;
    if (utf8ByteLength(rawText) > contentLimits.targetJobRawTextBytes) {
      setImportError("home.errors.rawTextTooLarge");
      return;
    }
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
        params: { opaquePendingImportId },
      });
      return;
    }
    await submitImport(intent);
  };

  const openRecentPlan = useCallback(
    (job: TargetJob) => {
      openProtectedRoute(
        { name: "workspace", params: targetJobDetailRouteParams(job) },
        job.title,
      );
    },
    [openProtectedRoute],
  );

  const startRecentInterview = useCallback(
    async (job: TargetJob) => {
      const params = targetJobPracticeRouteParams(job);
      if (!runtime || runtime.auth.status !== "authenticated" || !params.resumeId || !params.roundId) {
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
    <>
      {startingRecentJobId ? <PracticeLaunchTransition /> : null}
      <section
        data-testid={`route-${route.name}`}
        data-route-name={route.name}
        data-route-params={JSON.stringify(route.params)}
        className="ei-screen-shell ei-home-screen"
      >
        <div className="ei-home-content">
          <header className="ei-home-hero">
            <div className="ei-home-hero-copy">
              <h1 data-testid="home-hero-title" className="ei-home-hero-title">
                <span>{t("home.heroTitleLead")}</span>{" "}
                <span data-testid="home-hero-title-accent" className="ei-home-hero-title-accent">
                  {t("home.heroTitleAccent")}
                </span>
              </h1>
              <p data-testid="home-hero-sub" className="ei-home-hero-sub">
                {t("home.heroSubtitle")}
              </p>
            </div>
            <HomeHeroIllustration />
          </header>

          <div data-testid="home-jd-input-card" className="ei-home-intake-card">
            <label className="ei-home-field-label" htmlFor="home-jd-textarea">
              {t("home.pasteSource")}
            </label>
            <div className="ei-home-jd-frame">
              <textarea
                ref={jdTextareaRef}
                id="home-jd-textarea"
                data-testid="home-jd-textarea"
                className="ei-home-jd-textarea"
                aria-label={t("home.jdPlaceholder")}
                placeholder={t("home.jdPlaceholder")}
                value={input}
                onChange={(event) => setInput(event.target.value)}
              />
              <span data-testid="home-jd-counter" className="ei-home-jd-counter">
                {utf8ByteLength(input)} / {contentLimits?.targetJobRawTextBytes ?? "—"}
              </span>
            </div>

            <div className="ei-home-resume-action-row">
              <div className="ei-home-resume-block">
                <label className="ei-home-field-label" htmlFor="home-resume-select">
                  {t("home.resumeSelect")}
                </label>
                <div data-testid="home-resume-row" className="ei-home-resume-row">
                  <span className="ei-home-resume-select-wrap">
                    <svg aria-hidden="true" className="ei-home-resume-icon" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M6 3h8l4 4v14H6zM14 3v5h4M9 12h6M9 16h5" />
                    </svg>
                    <select
                      id="home-resume-select"
                      data-testid="home-resume-select"
                      className="ei-home-resume-select"
                      aria-label={t("home.resumeSelect")}
                      value={selectedResumeId}
                      disabled={resumesLoading || readyResumes.length === 0}
                      onChange={(event) => setSelectedResumeId(event.target.value)}
                    >
                      <option value="">
                        {resumesLoading ? t("home.resumeLoading") : t("home.resumeSelectPlaceholder")}
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
                  </span>
                  <button
                    data-testid="home-resume-create"
                    type="button"
                    className="ei-home-resume-create"
                    onClick={() =>
                      openProtectedRoute(
                        { name: "resume_versions", params: { flow: "create" } },
                        t("home.resumeCreateLink"),
                      )
                    }
                  >
                    {t("home.resumeCreateLink")}
                  </button>
                </div>
                {!resumesLoading && readyResumes.length === 0 ? (
                  <div data-testid="home-resume-empty" className="ei-home-resume-empty">
                    {resumeError ? t(resumeError) : t("home.resumeEmpty")}
                  </div>
                ) : null}
                <span data-testid="home-resume-selection-status" className="ei-home-visually-hidden">
                  {selectedResume
                    ? `${t("home.resumeSelected")} · ${selectedResume.displayName || selectedResume.title}`
                    : t("home.resumeSelectHint")}
                </span>
              </div>

              <div data-testid="home-submit-row" className="ei-home-submit-row">
                <button
                  data-testid="home-jd-submit"
                  type="button"
                  className="ei-home-submit"
                  disabled={!canSubmit}
                  onClick={handlePasteImport}
                >
                  <svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M12 2l1.3 4.2L17.5 7.5l-4.2 1.3L12 13l-1.3-4.2-4.2-1.3 4.2-1.3L12 2zM18.5 13l.8 2.7 2.7.8-2.7.8-.8 2.7-.8-2.7-2.7-.8 2.7-.8.8-2.7z" />
                  </svg>
                  {t("home.importBtn")}
                  <svg aria-hidden="true" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M5 12h14M13 6l6 6-6 6" />
                  </svg>
                </button>
              </div>
            </div>

            <div data-testid="home-privacy-notice" className="ei-home-privacy-notice">
              <svg aria-hidden="true" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 3l7 3v5c0 4.5-2.7 8-7 10-4.3-2-7-5.5-7-10V6l7-3zM9 12l2 2 4-4" />
              </svg>
              {t("home.privacyNotice")}
            </div>
            {importError ? (
              <div data-testid="home-import-error" className="ei-home-error">
                {t(importError)}
              </div>
            ) : null}
          </div>

          {isAuthenticated ? (
            <section data-testid="home-recent-mocks" className="ei-home-recent">
              <header className="ei-home-recent-header">
                <div>
                  <h2>{t("home.recentSection")}</h2>
                  <p>{t("home.recentSectionSub")}</p>
                </div>
                {jobs.length > 0 ? (
                  <button
                    data-testid="home-recent-more"
                    type="button"
                    className="ei-home-recent-more"
                    onClick={() =>
                      openProtectedRoute(
                        { name: "workspace", params: {} },
                        t("home.recentMore"),
                      )
                    }
                  >
                    {t("home.recentMore")}
                    <svg aria-hidden="true" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M5 12h14M13 6l6 6-6 6" />
                    </svg>
                  </button>
                ) : null}
              </header>
              {loading ? (
                <div className="ei-skeleton-stripe">{t("home.recentSection")}...</div>
              ) : error ? (
                <div className="ei-home-error">{t("home.errors.recentLoad")}</div>
              ) : jobs.length === 0 ? (
                <div className="ei-home-recent-empty">{t("home.recentSection")}</div>
              ) : (
                <div data-testid="home-recent-mock-grid" className="ei-home-recent-grid">
                  {jobs.map((job) => (
                    <MockInterviewCard
                      key={job.id}
                      job={job}
                      presentation="home-record"
                      onClick={() => openRecentPlan(job)}
                      recentMeta={
                        <span
                          data-testid={`home-recent-mock-date-${job.id}`}
                          className="ei-home-recent-date"
                        >
                          {t("home.recentLastUsed")} · {formatRecentDate(job.updatedAt, lang)}
                        </span>
                      }
                      primaryAction={{
                        label: t("home.recentContinue"),
                        testId: `home-recent-mock-start-${job.id}`,
                        onClick: () => startRecentInterview(job),
                        disabled: startingRecentJobId === job.id || !isTargetJobPracticeStartable(job),
                      }}
                    />
                  ))}
                </div>
              )}
              {recentStartError ? (
                <div data-testid="home-recent-start-error" className="ei-home-error">
                  {t(recentStartError)}
                </div>
              ) : null}
            </section>
          ) : null}
        </div>
      </section>
    </>
  );
};
