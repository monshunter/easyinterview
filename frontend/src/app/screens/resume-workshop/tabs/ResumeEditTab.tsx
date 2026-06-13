import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type CSSProperties,
  type FC,
} from "react";

import type { Resume } from "../../../../api/generated/types";
import { useI18n, type MessageKey } from "../../../i18n/messages";
import { ResumeWorkshopIcon } from "../components/ResumeWorkshopIcon";

export interface ResumeEditTabProps {
  resume: Resume;
  /**
   * Async save handler. The payload carries the P0 editable fields
   * (displayName + headline + summary). Overwrites this resume via updateResume.
   */
  onSave?: (payload: {
    displayName: string;
    headline: string;
    summary: string;
  }) => Promise<void>;
  /** Optional in-form alert for validation / idempotency errors. */
  errorMessage?: string | null;
  /** True while the host hook is awaiting the updateResume call. */
  saving?: boolean;
}

const safeStringField = (value: unknown): string =>
  typeof value === "string" ? value : "";

export const ResumeEditTab: FC<ResumeEditTabProps> = ({
  resume,
  onSave,
  errorMessage = null,
  saving = false,
}) => {
  const { t } = useI18n();

  const profile = useMemo(
    () => (resume.structuredProfile ?? {}) as Record<string, unknown>,
    [resume.structuredProfile],
  );
  const initialDisplayName = resume.displayName;
  const initialHeadline = useMemo(
    () => safeStringField(profile.headline),
    [profile.headline],
  );
  const initialSummary = useMemo(
    () => safeStringField(profile.summary),
    [profile.summary],
  );

  const [displayName, setDisplayName] = useState(initialDisplayName);
  const [headline, setHeadline] = useState(initialHeadline);
  const [summary, setSummary] = useState(initialSummary);

  useEffect(() => {
    setDisplayName(initialDisplayName);
    setHeadline(initialHeadline);
    setSummary(initialSummary);
  }, [resume.id, initialDisplayName, initialHeadline, initialSummary]);

  const isDirty =
    displayName !== initialDisplayName ||
    headline !== initialHeadline ||
    summary !== initialSummary;

  const handleSave = useCallback(() => {
    if (!onSave || saving || !isDirty) return;
    void onSave({ displayName, headline, summary });
  }, [onSave, saving, isDirty, displayName, headline, summary]);

  return (
    <div
      data-testid="resume-edit-tab"
      data-resume-id={resume.id}
      data-edit-dirty={isDirty ? "true" : "false"}
      data-edit-saving={saving ? "true" : "false"}
    >
      <ScopeBanner
        resumeName={resume.displayName}
        onSave={handleSave}
        disabled={saving || !isDirty}
        saving={saving}
        t={t}
      />

      <div style={CARD_STYLE}>
        <div style={CARD_SECTION_STYLE}>
          <label
            htmlFor="resume-edit-display-name-input"
            className="ei-text-label"
            style={{
              color: "var(--ei-color-ink3)",
              marginBottom: 8,
              display: "block",
            }}
          >
            {t("resumeWorkshop.edit.displayNameLabel")}
          </label>
          <input
            id="resume-edit-display-name-input"
            data-testid="resume-edit-display-name-input"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            style={INPUT_STYLE}
            disabled={saving}
          />
          <label
            htmlFor="resume-edit-headline-input"
            className="ei-text-label"
            style={{
              color: "var(--ei-color-ink3)",
              margin: "16px 0 8px",
              display: "block",
            }}
          >
            {t("resumeWorkshop.edit.headlineLabel")}
          </label>
          <input
            id="resume-edit-headline-input"
            data-testid="resume-edit-headline-input"
            value={headline}
            onChange={(e) => setHeadline(e.target.value)}
            style={INPUT_STYLE}
            disabled={saving}
          />
          <label
            htmlFor="resume-edit-summary-textarea"
            className="ei-text-label"
            style={{
              color: "var(--ei-color-ink3)",
              margin: "16px 0 8px",
              display: "block",
            }}
          >
            {t("resumeWorkshop.edit.summaryLabel")}
          </label>
          <textarea
            id="resume-edit-summary-textarea"
            data-testid="resume-edit-summary-textarea"
            value={summary}
            onChange={(e) => setSummary(e.target.value)}
            style={TEXTAREA_STYLE}
            disabled={saving}
          />
        </div>

        <SectionPlaceholder
          testId="resume-edit-section-experience"
          title={t("resumeWorkshop.edit.section.experience")}
          addLabel={t("resumeWorkshop.edit.section.add")}
          comingSoonLabel={t("resumeWorkshop.edit.section.comingSoon")}
        />
        <SectionPlaceholder
          testId="resume-edit-section-skills"
          title={t("resumeWorkshop.edit.section.skills")}
          addLabel={t("resumeWorkshop.edit.section.add")}
          comingSoonLabel={t("resumeWorkshop.edit.section.comingSoon")}
          isLast
        />
      </div>

      {errorMessage ? (
        <div data-testid="resume-edit-error" role="alert" style={ERROR_STYLE}>
          {errorMessage}
        </div>
      ) : null}
    </div>
  );
};

interface ScopeBannerProps {
  resumeName: string;
  onSave: () => void;
  disabled: boolean;
  saving: boolean;
  t: (key: MessageKey) => string;
}

const ScopeBanner: FC<ScopeBannerProps> = ({
  resumeName,
  onSave,
  disabled,
  saving,
  t,
}) => {
  const message = t("resumeWorkshop.edit.scope.body").replace(
    "{resumeName}",
    resumeName,
  );
  return (
    <div
      data-testid="resume-edit-scope-banner"
      role="status"
      style={SCOPE_BANNER_STYLE}
    >
      <div
        style={{ fontSize: 13, color: "var(--ei-color-ink3)" }}
        data-testid="resume-edit-scope-banner-message"
      >
        <ResumeWorkshopIcon name="check" size={12} /> {message}
      </div>
      <button
        type="button"
        data-testid="resume-edit-save-button"
        onClick={onSave}
        disabled={disabled}
        aria-disabled={disabled}
        style={{
          ...BTN_ACCENT,
          opacity: disabled ? 0.55 : 1,
          cursor: disabled ? "not-allowed" : "pointer",
        }}
      >
        {saving ? t("resumeWorkshop.edit.saving") : t("resumeWorkshop.edit.save")}
      </button>
    </div>
  );
};

interface SectionPlaceholderProps {
  testId: string;
  title: string;
  addLabel: string;
  comingSoonLabel: string;
  isLast?: boolean;
}

const SectionPlaceholder: FC<SectionPlaceholderProps> = ({
  testId,
  title,
  addLabel,
  comingSoonLabel,
  isLast = false,
}) => (
  <div
    data-testid={testId}
    style={{
      padding: "20px 24px",
      borderBottom: isLast ? "none" : "1px solid var(--ei-color-rule)",
    }}
  >
    <div
      style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        marginBottom: 12,
      }}
    >
      <div className="ei-text-label" style={{ color: "var(--ei-color-ink3)" }}>
        {title}
      </div>
      <button
        type="button"
        data-testid={`${testId}-add`}
        onClick={() => {
          if (typeof window === "undefined") return;
          const fn = (
            window as unknown as {
              eiToast?: (msg: string, opts?: { tone?: string }) => void;
            }
          ).eiToast;
          fn?.(comingSoonLabel, { tone: "neutral" });
        }}
        style={BTN_OUTLINE}
      >
        <ResumeWorkshopIcon name="plus" size={11} /> {addLabel}
      </button>
    </div>
    <div
      data-testid={`${testId}-placeholder`}
      style={{
        fontSize: 13,
        color: "var(--ei-color-ink3)",
        fontFamily: "var(--ei-mono)",
      }}
    >
      {comingSoonLabel}
    </div>
  </div>
);

const SCOPE_BANNER_STYLE: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  padding: "10px 14px",
  marginBottom: 16,
  background: "var(--ei-color-bg-soft)",
  border: "1px dotted var(--ei-color-rule)",
  borderRadius: 2,
  gap: 14,
};

const CARD_STYLE: CSSProperties = {
  background: "var(--ei-color-bg-card)",
  border: "1px solid var(--ei-color-rule)",
  borderRadius: 2,
};

const CARD_SECTION_STYLE: CSSProperties = {
  padding: "20px 24px",
  borderBottom: "1px solid var(--ei-color-rule)",
};

const INPUT_STYLE: CSSProperties = {
  width: "100%",
  boxSizing: "border-box",
  padding: "10px 12px",
  border: "1px solid var(--ei-color-rule)",
  borderRadius: 2,
  background: "var(--ei-color-bg)",
  color: "var(--ei-color-ink)",
  fontSize: 16,
  fontFamily: "var(--ei-serif)",
  outline: "none",
};

const TEXTAREA_STYLE: CSSProperties = {
  width: "100%",
  boxSizing: "border-box",
  minHeight: 80,
  padding: "10px 12px",
  border: "1px solid var(--ei-color-rule)",
  borderRadius: 2,
  background: "var(--ei-color-bg)",
  color: "var(--ei-color-ink)",
  fontSize: 13.5,
  lineHeight: 1.6,
  resize: "vertical",
  outline: "none",
};

const ERROR_STYLE: CSSProperties = {
  marginTop: 12,
  padding: "10px 14px",
  background: "var(--ei-color-danger-soft)",
  color: "var(--ei-color-danger)",
  border: "1px solid var(--ei-color-danger)",
  borderRadius: 2,
  fontSize: 13,
};

const BTN_ACCENT: CSSProperties = {
  padding: "6px 14px",
  background: "var(--ei-color-accent)",
  color: "#fff",
  border: "none",
  borderRadius: 2,
  fontSize: 13,
  fontFamily: "var(--ei-sans)",
};

const BTN_OUTLINE: CSSProperties = {
  background: "transparent",
  border: "1px solid var(--ei-color-rule)",
  borderRadius: 2,
  padding: "4px 10px",
  fontSize: 12,
  color: "var(--ei-color-ink3)",
  cursor: "pointer",
  fontFamily: "var(--ei-sans)",
};
