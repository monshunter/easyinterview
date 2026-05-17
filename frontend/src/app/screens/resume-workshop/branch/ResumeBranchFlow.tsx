import {
  useCallback,
  useMemo,
  useState,
  type CSSProperties,
  type FC,
} from "react";

import type {
  ResumeAsset,
  ResumeSeedStrategy,
  ResumeVersion,
} from "../../../../api/generated/types";
import { useI18n, type MessageKey } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { fireResumeWorkshopToast } from "../components/toast";
import { ResumeWorkshopIcon } from "../components/ResumeWorkshopIcon";
import {
  useResumeBranchSource,
  type ResumeBranchSourceStatus,
} from "./useResumeBranchSource";
import {
  useResumeBranchSubmit,
  BranchSubmitError,
} from "./hooks/useResumeBranchSubmit";
import type { BranchSubmitOutcome } from "./hooks/useResumeBranchSubmit";

const FOCUS_OPTIONS: ReadonlyArray<{
  k: BranchFocus;
  label: MessageKey;
}> = [
  { k: "platform", label: "resumeWorkshop.branch.focus.platform" },
  { k: "collaboration", label: "resumeWorkshop.branch.focus.collaboration" },
  { k: "fullstack", label: "resumeWorkshop.branch.focus.fullstack" },
  { k: "leadership", label: "resumeWorkshop.branch.focus.leadership" },
  { k: "custom", label: "resumeWorkshop.branch.focus.custom" },
];

const SEED_OPTIONS: ReadonlyArray<{
  k: ResumeSeedStrategy;
  icon: "layers" | "file" | "sparkle";
  label: MessageKey;
  desc: MessageKey;
}> = [
  {
    k: "copy_master",
    icon: "layers",
    label: "resumeWorkshop.branch.seed.copy_master.label",
    desc: "resumeWorkshop.branch.seed.copy_master.desc",
  },
  {
    k: "blank",
    icon: "file",
    label: "resumeWorkshop.branch.seed.blank.label",
    desc: "resumeWorkshop.branch.seed.blank.desc",
  },
  {
    k: "ai_select",
    icon: "sparkle",
    label: "resumeWorkshop.branch.seed.ai_select.label",
    desc: "resumeWorkshop.branch.seed.ai_select.desc",
  },
];

export type BranchFocus =
  | "platform"
  | "collaboration"
  | "fullstack"
  | "leadership"
  | "custom";

export interface ResumeBranchFormDraft {
  name: string;
  target: string;
  focus: BranchFocus;
  seed: ResumeSeedStrategy;
}

export interface ResumeBranchFlowProps {
  branchOriginalId: string | null;
}

export interface ResumeBranchFlowFormProps {
  branchOriginalId: string;
  original: ResumeAsset;
  master: ResumeVersion;
  onCancel: () => void;
  onSubmit: (draft: ResumeBranchFormDraft) => void | Promise<void>;
  submitting?: boolean;
  errorMessage?: string | null;
}

const SCREEN_WRAPPER: CSSProperties = {
  maxWidth: 980,
  margin: "0 auto",
  padding: "40px 48px 96px",
};

export const ResumeBranchFlow: FC<ResumeBranchFlowProps> = ({
  branchOriginalId,
}) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const source = useResumeBranchSource(branchOriginalId);
  const submitHook = useResumeBranchSubmit();

  const onBack = useCallback(() => {
    navigate({ name: "resume_versions", params: {} });
  }, [navigate]);

  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const handleSubmit = useCallback(
    async (draft: ResumeBranchFormDraft) => {
      if (source.status !== "ready" || !source.master || !branchOriginalId) {
        setErrorMessage(t("resumeWorkshop.branch.error.parentMissing"));
        return;
      }
      setErrorMessage(null);
      try {
        const outcome = await submitHook.submit(draft, {
          parentVersionId: source.master.id,
          targetJobId: draft.target.trim(),
        });
        dispatchSuccess(outcome, draft.seed, navigate, t);
      } catch (rawErr) {
        const friendly = mapBranchErrorToMessage(rawErr, t);
        setErrorMessage(friendly);
      }
    },
    [
      branchOriginalId,
      navigate,
      source.master,
      source.status,
      submitHook,
      t,
    ],
  );

  return (
    <div
      data-testid="resume-branch-flow"
      data-branch-original-id={branchOriginalId ?? ""}
      data-source-status={source.status}
      data-submit-state={submitHook.submitting ? "submitting" : "idle"}
      className="ei-fadein"
      style={SCREEN_WRAPPER}
    >
      {renderBody({
        status: source.status,
        original: source.original,
        master: source.master,
        branchOriginalId,
        onBack,
        onSubmit: handleSubmit,
        submitting: submitHook.submitting,
        errorMessage,
        t,
      })}
    </div>
  );
};

type Navigate = ReturnType<typeof useNavigation>["navigate"];

const dispatchSuccess = (
  outcome: BranchSubmitOutcome,
  seed: ResumeSeedStrategy,
  navigate: Navigate,
  t: (key: MessageKey) => string,
) => {
  if (outcome.kind === "version") {
    const tab = seed === "blank" ? "edit" : "rewrites";
    const toastKey: MessageKey =
      seed === "blank"
        ? "resumeWorkshop.branch.toast.blank"
        : "resumeWorkshop.branch.toast.copyMaster";
    fireResumeWorkshopToast(t(toastKey), "ok");
    navigate({
      name: "resume_versions",
      params: { versionId: outcome.version.id, tab },
    });
    return;
  }
  // ai_select 202 → versionId from accepted envelope; tailorRunId is consumed
  // by Phase 5 polling hook via getResumeVersion / route-restored detail tab.
  fireResumeWorkshopToast(t("resumeWorkshop.branch.toast.aiSelect"), "ok");
  navigate({
    name: "resume_versions",
    params: {
      versionId: outcome.accepted.resumeVersionId,
      tab: "rewrites",
    },
  });
};

const mapBranchErrorToMessage = (
  err: unknown,
  t: (key: MessageKey) => string,
): string => {
  if (err instanceof BranchSubmitError) {
    switch (err.kind) {
      case "validation":
        return t("resumeWorkshop.branch.error.validation");
      case "parent_missing":
      case "target_missing":
      case "cross_user":
        return t("resumeWorkshop.branch.error.parentMissing");
      case "idempotency_conflict":
        return t("resumeWorkshop.branch.error.idempotencyConflict");
      default:
        return t("resumeWorkshop.branch.error.generic");
    }
  }
  return t("resumeWorkshop.branch.error.generic");
};

interface BodyArgs {
  status: ResumeBranchSourceStatus;
  original: ResumeAsset | null;
  master: ResumeVersion | null;
  branchOriginalId: string | null;
  onBack: () => void;
  onSubmit: (draft: ResumeBranchFormDraft) => void | Promise<void>;
  submitting: boolean;
  errorMessage: string | null;
  t: (key: MessageKey) => string;
}

const renderBody = (args: BodyArgs) => {
  const { status, original, master, branchOriginalId } = args;
  const { onBack, onSubmit, submitting, errorMessage, t } = args;

  if (status === "missing-id") {
    return (
      <FallbackPanel
        testId="resume-branch-missing-id"
        eyebrow={t("resumeWorkshop.branch.eyebrow")}
        title={t("resumeWorkshop.branch.missingId.title")}
        body={t("resumeWorkshop.branch.missingId.body")}
        ctaLabel={t("resumeWorkshop.branch.notFound.cta")}
        onCta={onBack}
      />
    );
  }
  if (status === "loading") {
    return (
      <div data-testid="resume-branch-loading" style={LOADING_STYLE}>
        {t("resumeWorkshop.branch.loading")}
      </div>
    );
  }
  if (status === "not-found" || status === "error") {
    return (
      <FallbackPanel
        testId="resume-branch-not-found"
        eyebrow={t("resumeWorkshop.branch.eyebrow")}
        title={t("resumeWorkshop.branch.notFound.title")}
        body={t("resumeWorkshop.branch.notFound.body")}
        ctaLabel={t("resumeWorkshop.branch.notFound.cta")}
        onCta={onBack}
      />
    );
  }
  // status === 'ready'
  if (!original || !master || !branchOriginalId) return null;
  return (
    <ResumeBranchFlowForm
      branchOriginalId={branchOriginalId}
      original={original}
      master={master}
      onCancel={onBack}
      onSubmit={onSubmit}
      submitting={submitting}
      errorMessage={errorMessage}
    />
  );
};

const LOADING_STYLE: CSSProperties = {
  fontFamily: "var(--ei-mono)",
  fontSize: 13,
  color: "var(--ei-color-ink3)",
  padding: "32px 0",
};

interface FallbackPanelProps {
  testId: string;
  eyebrow: string;
  title: string;
  body: string;
  ctaLabel: string;
  onCta: () => void;
}

const FallbackPanel: FC<FallbackPanelProps> = ({
  testId,
  eyebrow,
  title,
  body,
  ctaLabel,
  onCta,
}) => (
  <div data-testid={testId} className="ei-screen-card">
    <span className="ei-text-label">{eyebrow}</span>
    <h2 className="ei-text-title">{title}</h2>
    <p className="ei-text-body">{body}</p>
    <button
      type="button"
      data-testid={`${testId}-cta`}
      className="ei-auth-cta"
      onClick={onCta}
    >
      {ctaLabel}
    </button>
  </div>
);

const formatSourceMeta = (asset: ResumeAsset): string => {
  const parts: string[] = [];
  if (asset.sourceType) parts.push(asset.sourceType);
  if (asset.createdAt) parts.push(asset.createdAt);
  return parts.join(" · ");
};

export const ResumeBranchFlowForm: FC<ResumeBranchFlowFormProps> = ({
  branchOriginalId,
  original,
  master,
  onCancel,
  onSubmit,
  submitting,
  errorMessage,
}) => {
  const { t } = useI18n();
  const [name, setName] = useState("");
  const [target, setTarget] = useState("");
  const [focus, setFocus] = useState<BranchFocus>("platform");
  const [seed, setSeed] = useState<ResumeSeedStrategy>("copy_master");

  const canSubmit = useMemo(
    () => name.trim().length > 0 && target.trim().length > 0,
    [name, target],
  );

  const onSubmitClick = useCallback(() => {
    if (!canSubmit || submitting) return;
    void onSubmit({
      name: name.trim(),
      target: target.trim(),
      focus,
      seed,
    });
  }, [canSubmit, focus, name, onSubmit, seed, submitting, target]);

  return (
    <div
      data-testid="resume-branch-flow-form"
      data-branch-original-id={branchOriginalId}
      data-branch-seed={seed}
      data-branch-focus={focus}
      data-branch-can-submit={canSubmit ? "true" : "false"}
    >
      <button
        type="button"
        data-testid="resume-branch-back"
        onClick={onCancel}
        style={{
          background: "transparent",
          border: "none",
          color: "var(--ei-color-ink3)",
          fontSize: 13,
          marginBottom: 16,
          display: "flex",
          alignItems: "center",
          gap: 6,
          cursor: "pointer",
        }}
      >
        <ResumeWorkshopIcon name="arrowLeft" size={14} />{" "}
        {t("resumeWorkshop.branch.back")}
      </button>
      <div
        className="ei-text-label"
        style={{ color: "var(--ei-color-accent)", marginBottom: 8 }}
      >
        {t("resumeWorkshop.branch.eyebrow")}
      </div>
      <h1
        className="ei-text-display"
        style={{
          fontSize: 32,
          margin: 0,
          color: "var(--ei-color-ink)",
          letterSpacing: "-0.022em",
          lineHeight: 1.2,
        }}
      >
        {t("resumeWorkshop.branch.title")}
      </h1>
      <p
        style={{
          fontSize: 14,
          color: "var(--ei-color-ink3)",
          marginTop: 10,
          maxWidth: 720,
          lineHeight: 1.55,
        }}
      >
        {t("resumeWorkshop.branch.subtitle")}
      </p>

      {/* BRANCHING FROM */}
      <div style={{ marginTop: 22 }}>
        <div
          className="ei-text-label"
          style={{ color: "var(--ei-color-ink3)", marginBottom: 8 }}
        >
          {t("resumeWorkshop.branch.fromLabel")}
        </div>
        <div
          data-testid="resume-branch-from-card"
          className="ei-screen-card"
          style={{
            padding: 0,
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
          }}
        >
          <div
            data-testid="resume-branch-from-original"
            style={{
              padding: "16px 20px",
              borderRight: "1px dotted var(--ei-color-rule)",
            }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: 8,
                marginBottom: 6,
                color: "var(--ei-color-ink3)",
              }}
            >
              <ResumeWorkshopIcon name="file" size={13} />
              <div className="ei-text-label">
                {t("resumeWorkshop.branch.fromOriginalLabel")}
              </div>
            </div>
            <div
              data-testid="resume-branch-from-original-name"
              style={{
                fontSize: 13.5,
                color: "var(--ei-color-ink)",
                fontWeight: 500,
                whiteSpace: "nowrap",
                overflow: "hidden",
                textOverflow: "ellipsis",
              }}
            >
              {original.title || t("resumeWorkshop.branch.fromEmpty")}
            </div>
            <div
              style={{
                fontSize: 12,
                color: "var(--ei-color-ink3)",
                marginTop: 3,
              }}
            >
              {formatSourceMeta(original)}
            </div>
          </div>
          <div
            data-testid="resume-branch-from-master"
            style={{
              padding: "16px 20px",
              background: "var(--ei-color-bg-soft)",
            }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: 8,
                marginBottom: 6,
                color: "var(--ei-color-ink3)",
              }}
            >
              <ResumeWorkshopIcon name="resume" size={13} />
              <div className="ei-text-label">
                {t("resumeWorkshop.branch.fromMasterLabel")}
              </div>
            </div>
            <div
              data-testid="resume-branch-from-master-name"
              style={{
                fontSize: 13.5,
                color: "var(--ei-color-ink)",
                fontWeight: 500,
                whiteSpace: "nowrap",
                overflow: "hidden",
                textOverflow: "ellipsis",
              }}
            >
              {master.displayName || t("resumeWorkshop.branch.fromEmpty")}
            </div>
            <div
              style={{
                fontSize: 12,
                color: "var(--ei-color-ink3)",
                marginTop: 3,
              }}
            >
              {t("resumeWorkshop.branch.fromMasterHint")}
            </div>
          </div>
        </div>
      </div>

      {/* Form: name + target */}
      <div
        style={{
          marginTop: 22,
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: 14,
        }}
      >
        <div className="ei-screen-card" style={{ padding: 20 }}>
          <label
            htmlFor="resume-branch-field-name"
            className="ei-text-label"
            style={{ color: "var(--ei-color-ink3)", marginBottom: 8 }}
          >
            {t("resumeWorkshop.branch.nameLabel")}
          </label>
          <input
            id="resume-branch-field-name"
            data-testid="resume-branch-field-name"
            value={name}
            aria-required="true"
            aria-invalid={!name.trim() ? "true" : "false"}
            onChange={(e) => setName(e.target.value)}
            placeholder={t("resumeWorkshop.branch.namePlaceholder")}
            style={INPUT_STYLE}
          />
          <div style={HINT_STYLE}>{t("resumeWorkshop.branch.nameHint")}</div>
        </div>
        <div className="ei-screen-card" style={{ padding: 20 }}>
          <label
            htmlFor="resume-branch-field-target"
            className="ei-text-label"
            style={{ color: "var(--ei-color-ink3)", marginBottom: 8 }}
          >
            {t("resumeWorkshop.branch.targetLabel")}
          </label>
          <input
            id="resume-branch-field-target"
            data-testid="resume-branch-field-target"
            value={target}
            aria-required="true"
            aria-invalid={!target.trim() ? "true" : "false"}
            onChange={(e) => setTarget(e.target.value)}
            placeholder={t("resumeWorkshop.branch.targetPlaceholder")}
            style={INPUT_STYLE}
          />
          <div style={HINT_STYLE}>{t("resumeWorkshop.branch.targetHint")}</div>
        </div>
      </div>

      {/* Focus chips */}
      <div style={{ marginTop: 14 }}>
        <div
          className="ei-screen-card"
          style={{ padding: 20 }}
          data-testid="resume-branch-focus-card"
        >
          <div
            className="ei-text-label"
            style={{ color: "var(--ei-color-ink3)", marginBottom: 10 }}
          >
            {t("resumeWorkshop.branch.focusLabel")}
          </div>
          <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
            {FOCUS_OPTIONS.map((option) => {
              const on = focus === option.k;
              return (
                <button
                  key={option.k}
                  type="button"
                  data-testid={`resume-branch-focus-chip-${option.k}`}
                  data-selected={on ? "true" : "false"}
                  aria-pressed={on}
                  onClick={() => setFocus(option.k)}
                  style={{
                    padding: "8px 14px",
                    background: on
                      ? "var(--ei-color-accent-soft)"
                      : "transparent",
                    border: `1px solid ${
                      on
                        ? "var(--ei-color-accent)"
                        : "var(--ei-color-rule)"
                    }`,
                    borderRadius: 2,
                    color: on
                      ? "var(--ei-color-accent)"
                      : "var(--ei-color-ink2)",
                    fontFamily: "var(--ei-sans)",
                    fontSize: 13,
                    cursor: "pointer",
                  }}
                >
                  {t(option.label)}
                </button>
              );
            })}
          </div>
          <div style={{ ...HINT_STYLE, marginTop: 10 }}>
            {t("resumeWorkshop.branch.focusHint")}
          </div>
        </div>
      </div>

      {/* Seed cards */}
      <div style={{ marginTop: 14 }}>
        <div
          className="ei-screen-card"
          style={{ padding: 20 }}
          data-testid="resume-branch-seed-card"
        >
          <div
            className="ei-text-label"
            style={{ color: "var(--ei-color-ink3)", marginBottom: 10 }}
          >
            {t("resumeWorkshop.branch.seedLabel")}
          </div>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(3, 1fr)",
              gap: 10,
            }}
          >
            {SEED_OPTIONS.map((option) => {
              const on = seed === option.k;
              return (
                <button
                  key={option.k}
                  type="button"
                  data-testid={`resume-branch-seed-card-${option.k}`}
                  data-selected={on ? "true" : "false"}
                  aria-pressed={on}
                  onClick={() => setSeed(option.k)}
                  style={{
                    textAlign: "left",
                    padding: "14px 14px",
                    background: on
                      ? "var(--ei-color-accent-soft)"
                      : "var(--ei-color-bg)",
                    border: `1px solid ${
                      on
                        ? "var(--ei-color-accent)"
                        : "var(--ei-color-rule)"
                    }`,
                    borderRadius: 2,
                    cursor: "pointer",
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 8,
                      marginBottom: 6,
                      color: on
                        ? "var(--ei-color-accent)"
                        : "var(--ei-color-ink3)",
                    }}
                  >
                    <ResumeWorkshopIcon name={option.icon} size={13} />
                    <div className="ei-text-label">{t(option.label)}</div>
                  </div>
                  <div
                    style={{
                      fontSize: 12,
                      color: "var(--ei-color-ink2)",
                      lineHeight: 1.5,
                    }}
                  >
                    {t(option.desc)}
                  </div>
                </button>
              );
            })}
          </div>
        </div>
      </div>

      {/* Actions */}
      <div
        style={{
          marginTop: 22,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          gap: 12,
        }}
      >
        <div
          data-testid="resume-branch-submit-hint"
          style={{
            fontSize: 12,
            color: "var(--ei-color-ink3)",
            fontFamily: "var(--ei-mono)",
          }}
        >
          {canSubmit
            ? t("resumeWorkshop.branch.readyHint")
            : t("resumeWorkshop.branch.canSubmitHint")}
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          <button
            type="button"
            data-testid="resume-branch-cancel"
            onClick={onCancel}
            style={BTN_GHOST}
          >
            {t("resumeWorkshop.branch.cancel")}
          </button>
          <button
            type="button"
            data-testid="resume-branch-submit"
            disabled={!canSubmit || submitting}
            aria-disabled={!canSubmit || submitting}
            onClick={onSubmitClick}
            style={{
              ...BTN_ACCENT,
              opacity: !canSubmit || submitting ? 0.55 : 1,
              cursor: !canSubmit || submitting ? "not-allowed" : "pointer",
            }}
          >
            {submitting
              ? t("resumeWorkshop.branch.submitting")
              : t("resumeWorkshop.branch.submit")}
          </button>
        </div>
      </div>

      {errorMessage ? (
        <div
          data-testid="resume-branch-error"
          role="alert"
          style={{
            marginTop: 12,
            padding: "10px 14px",
            background: "var(--ei-color-danger-soft)",
            color: "var(--ei-color-danger)",
            border: "1px solid var(--ei-color-danger)",
            borderRadius: 2,
            fontSize: 13,
            fontFamily: "var(--ei-sans)",
          }}
        >
          {errorMessage}
        </div>
      ) : null}
    </div>
  );
};

const INPUT_STYLE: CSSProperties = {
  width: "100%",
  boxSizing: "border-box",
  padding: "10px 12px",
  border: "1px solid var(--ei-color-rule)",
  borderRadius: 2,
  background: "var(--ei-color-bg)",
  fontFamily: "var(--ei-sans)",
  fontSize: 14,
  color: "var(--ei-color-ink)",
};

const HINT_STYLE: CSSProperties = {
  fontSize: 11.5,
  color: "var(--ei-color-ink3)",
  marginTop: 6,
};

const BTN_GHOST: CSSProperties = {
  padding: "8px 14px",
  background: "transparent",
  border: "1px solid var(--ei-color-rule)",
  borderRadius: 2,
  color: "var(--ei-color-ink2)",
  fontSize: 13,
  fontFamily: "var(--ei-sans)",
  cursor: "pointer",
};

const BTN_ACCENT: CSSProperties = {
  padding: "8px 14px",
  background: "var(--ei-color-accent)",
  border: "1px solid var(--ei-color-accent)",
  borderRadius: 2,
  color: "#fff",
  fontSize: 13,
  fontFamily: "var(--ei-sans)",
};
