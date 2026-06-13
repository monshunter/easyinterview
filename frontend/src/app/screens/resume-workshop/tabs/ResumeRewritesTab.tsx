import {
  useEffect,
  useMemo,
  useState,
  type CSSProperties,
  type FC,
} from "react";

import type {
  Resume,
  ResumeTailorBulletSuggestion,
} from "../../../../api/generated/types";
import { useI18n, type MessageKey } from "../../../i18n/messages";
import {
  mapBulletSuggestionToUi,
  type UiBullet,
} from "../adapters/resume";

export interface AcceptedRewrite {
  original: string;
  rewritten: string;
}

export interface ResumeRewritesTabProps {
  resume: Resume;
  /** Ephemeral tailor-run suggestions (D-20: not persisted until saved). */
  suggestions: ResumeTailorBulletSuggestion[];
  /** Trigger a new tailor run for this resume (Phase 5). */
  onRequestRerun?: (mode: "bullet_suggestions" | "gap_review") => void;
  /** Save accepted rewrites by overwriting this resume (updateResume). */
  onOverwrite?: (accepted: AcceptedRewrite[]) => Promise<void> | void;
  /** Save accepted rewrites as a new resume (duplicateResume). */
  onSaveAsNew?: (accepted: AcceptedRewrite[]) => Promise<void> | void;
  /** True while the host hook is awaiting a save call. */
  saving?: boolean;
  /**
   * Inline status banner shown above the rewrites list (Phase 5 polling). The
   * tab component renders the banner content but does not own polling state.
   */
  pollingBanner?: ReactPollingBanner | null;
}

export type ReactPollingBanner =
  | { kind: "info"; message: string }
  | { kind: "danger"; message: string; onRetry: () => void };

const toUiBullets = (
  suggestions: ResumeTailorBulletSuggestion[],
): UiBullet[] =>
  suggestions.map((suggestion, index) =>
    mapBulletSuggestionToUi({
      id: `bullet-${index}`,
      originalBullet: suggestion.originalBullet,
      suggestedBullet: suggestion.suggestedBullet,
      reason: suggestion.reason,
    }),
  );

export const ResumeRewritesTab: FC<ResumeRewritesTabProps> = ({
  resume,
  suggestions,
  onRequestRerun,
  onOverwrite,
  onSaveAsNew,
  saving = false,
  pollingBanner = null,
}) => {
  const { t } = useI18n();
  const bullets = useMemo<UiBullet[]>(
    () => toUiBullets(suggestions),
    [suggestions],
  );

  const [acceptedIds, setAcceptedIds] = useState<Record<string, boolean>>({});
  const [selectedBulletId, setSelectedBulletId] = useState<string | null>(
    bullets[0]?.id ?? null,
  );
  const [confirmOpen, setConfirmOpen] = useState(false);

  useEffect(() => {
    setAcceptedIds({});
    setConfirmOpen(false);
    setSelectedBulletId(bullets[0]?.id ?? null);
  }, [resume.id, bullets]);

  const decorated = bullets.map((b) => ({
    ...b,
    status: acceptedIds[b.id] ? ("accepted" as const) : ("pending" as const),
  }));
  const selected =
    decorated.find((b) => b.id === selectedBulletId) ?? decorated[0] ?? null;
  const acceptedBullets = decorated.filter((b) => b.status === "accepted");
  const acceptedCount = acceptedBullets.length;

  const acceptBullet = (id: string) => {
    if (acceptedIds[id]) return;
    setAcceptedIds((prev) => ({ ...prev, [id]: true }));
  };

  const handleConfirm = async (mode: "overwrite" | "new") => {
    setConfirmOpen(false);
    const accepted: AcceptedRewrite[] = acceptedBullets.map((b) => ({
      original: b.original,
      rewritten: b.rewritten,
    }));
    if (mode === "new") {
      await onSaveAsNew?.(accepted);
    } else {
      await onOverwrite?.(accepted);
    }
  };

  return (
    <div
      data-testid="resume-rewrites-tab"
      data-resume-id={resume.id}
      data-bullet-count={bullets.length}
      data-accepted-count={acceptedCount}
      data-selected-bullet-id={selectedBulletId ?? ""}
    >
      <ScopeBanner
        acceptedCount={acceptedCount}
        untouchedCount={bullets.length - acceptedCount}
        onPreviewSave={() => setConfirmOpen(true)}
        previewDisabled={acceptedCount === 0 || saving}
        t={t}
      />
      {pollingBanner ? <PollingBanner banner={pollingBanner} t={t} /> : null}
      {bullets.length === 0 ? (
        <EmptyState
          t={t}
          onRerun={
            onRequestRerun
              ? () => onRequestRerun("bullet_suggestions")
              : undefined
          }
        />
      ) : (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1.3fr",
            gap: 20,
          }}
        >
          <BulletList
            bullets={decorated}
            selectedBulletId={selectedBulletId}
            onSelect={setSelectedBulletId}
            t={t}
          />
          <div>
            {selected ? (
              <DiffDetailCard
                bullet={selected}
                onAccept={() => acceptBullet(selected.id)}
                t={t}
              />
            ) : null}
          </div>
        </div>
      )}

      {confirmOpen ? (
        <RewriteSaveConfirmModal
          resumeName={resume.displayName}
          bullets={acceptedBullets}
          saving={saving}
          onClose={() => setConfirmOpen(false)}
          onConfirm={handleConfirm}
          t={t}
        />
      ) : null}
    </div>
  );
};

interface ScopeBannerProps {
  acceptedCount: number;
  untouchedCount: number;
  onPreviewSave: () => void;
  previewDisabled: boolean;
  t: (key: MessageKey) => string;
}

const ScopeBanner: FC<ScopeBannerProps> = ({
  acceptedCount,
  untouchedCount,
  onPreviewSave,
  previewDisabled,
  t,
}) => (
  <div
    data-testid="resume-rewrites-scope-banner"
    role="status"
    aria-live="polite"
    style={{
      display: "flex",
      justifyContent: "space-between",
      alignItems: "center",
      gap: 14,
      padding: "10px 14px",
      marginBottom: 16,
      background: "var(--ei-color-accent-soft)",
      border: "1px solid var(--ei-color-accent)",
      borderRadius: 2,
      flexWrap: "wrap",
    }}
  >
    <div style={{ fontSize: 13, color: "var(--ei-color-ink2)" }}>
      {t("resumeWorkshop.rewrites.scopeBanner.body")}
    </div>
    <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
      <div
        data-testid="resume-rewrites-counts"
        style={{
          fontSize: 11,
          color: "var(--ei-color-ink3)",
          fontFamily: "var(--ei-mono)",
        }}
      >
        {t("resumeWorkshop.rewrites.scopeBanner.counts")
          .replace("{accepted}", String(acceptedCount))
          .replace("{untouched}", String(untouchedCount))}
      </div>
      <button
        type="button"
        data-testid="resume-rewrites-preview-save"
        onClick={onPreviewSave}
        disabled={previewDisabled}
        aria-disabled={previewDisabled}
        style={{
          ...BTN_FILLED,
          opacity: previewDisabled ? 0.55 : 1,
          cursor: previewDisabled ? "not-allowed" : "pointer",
        }}
      >
        {t("resumeWorkshop.rewrites.previewAndSave")}
      </button>
    </div>
  </div>
);

interface BulletListProps {
  bullets: UiBullet[];
  selectedBulletId: string | null;
  onSelect: (id: string) => void;
  t: (key: MessageKey) => string;
}

const truncate = (text: string, max: number): string =>
  text.length > max ? `${text.slice(0, max)}…` : text;

const BulletList: FC<BulletListProps> = ({
  bullets,
  selectedBulletId,
  onSelect,
  t,
}) => (
  <div>
    <div
      className="ei-text-label"
      style={{ color: "var(--ei-color-ink3)", marginBottom: 12 }}
    >
      {t("resumeWorkshop.rewrites.listEyebrow")}
    </div>
    <div
      role="listbox"
      aria-label={t("resumeWorkshop.rewrites.listEyebrow")}
      data-testid="resume-rewrites-bullet-list"
      style={{ display: "flex", flexDirection: "column", gap: 10 }}
    >
      {bullets.map((b) => {
        const active = b.id === selectedBulletId;
        const accepted = b.status === "accepted";
        const tone = accepted
          ? "var(--ei-color-ok)"
          : "var(--ei-color-ink3)";
        return (
          <button
            key={b.id}
            type="button"
            role="option"
            aria-selected={active}
            data-testid={`resume-rewrites-bullet-row-${b.id}`}
            data-status={b.status}
            onClick={() => onSelect(b.id)}
            style={{
              padding: "14px 16px",
              textAlign: "left",
              cursor: "pointer",
              background: active
                ? "var(--ei-color-bg-soft)"
                : "var(--ei-color-bg-card)",
              border: `1px solid ${
                active ? "var(--ei-color-accent)" : "var(--ei-color-rule)"
              }`,
              borderRadius: 2,
              fontFamily: "var(--ei-sans)",
            }}
          >
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "flex-start",
                gap: 10,
                marginBottom: 6,
              }}
            >
              <div
                style={{
                  fontSize: 11,
                  color: "var(--ei-color-ink3)",
                  fontFamily: "var(--ei-mono)",
                  letterSpacing: "0.04em",
                }}
              >
                {b.section || t("resumeWorkshop.rewrites.diff.sectionFallback")}
              </div>
              <div
                data-testid={`resume-rewrites-status-chip-${b.status}-${b.id}`}
                style={{
                  display: "flex",
                  gap: 4,
                  alignItems: "center",
                  fontSize: 10.5,
                  color: tone,
                  fontFamily: "var(--ei-mono)",
                  letterSpacing: "0.04em",
                }}
              >
                <div
                  style={{
                    width: 5,
                    height: 5,
                    borderRadius: 3,
                    background: tone,
                  }}
                />
                {accepted
                  ? t("resumeWorkshop.rewrites.status.accepted")
                  : t("resumeWorkshop.rewrites.status.suggested")}
              </div>
            </div>
            <div
              style={{
                fontSize: 13,
                color: "var(--ei-color-ink2)",
                lineHeight: 1.5,
              }}
            >
              {truncate(b.rewritten, 90)}
            </div>
          </button>
        );
      })}
    </div>
  </div>
);

interface DiffDetailCardProps {
  bullet: UiBullet;
  onAccept: () => void;
  t: (key: MessageKey) => string;
}

const DiffDetailCard: FC<DiffDetailCardProps> = ({ bullet, onAccept, t }) => {
  const isAccepted = bullet.status === "accepted";
  return (
    <div
      data-testid="resume-rewrites-diff-card"
      data-bullet-id={bullet.id}
      data-bullet-status={bullet.status}
      style={DIFF_CARD_STYLE}
    >
      <div
        style={{
          padding: "14px 22px",
          borderBottom: "1px solid var(--ei-color-rule)",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <div className="ei-text-label" style={{ color: "var(--ei-color-ink3)" }}>
          {bullet.section || t("resumeWorkshop.rewrites.diff.sectionFallback")}
        </div>
        <button
          type="button"
          data-testid="resume-rewrites-action-accept"
          onClick={onAccept}
          disabled={isAccepted}
          aria-disabled={isAccepted}
          aria-label={t("resumeWorkshop.rewrites.action.accept")}
          style={{
            ...BTN_FILLED,
            background: isAccepted
              ? "var(--ei-color-ok)"
              : "var(--ei-color-accent)",
            cursor: isAccepted ? "default" : "pointer",
          }}
        >
          {isAccepted
            ? t("resumeWorkshop.rewrites.action.accepted")
            : t("resumeWorkshop.rewrites.action.accept")}
        </button>
      </div>

      <div
        style={{
          padding: "16px 22px",
          borderBottom: "1px dotted var(--ei-color-rule)",
        }}
      >
        <div
          style={{
            display: "flex",
            gap: 10,
            alignItems: "center",
            marginBottom: 8,
          }}
        >
          <span
            style={{
              padding: "2px 8px",
              background: "var(--ei-color-danger-soft)",
              color: "var(--ei-color-danger)",
              fontSize: 10.5,
              fontFamily: "var(--ei-mono)",
              letterSpacing: "0.08em",
              borderRadius: 2,
            }}
          >
            - {t("resumeWorkshop.rewrites.diff.original")}
          </span>
          <span
            style={{
              fontSize: 11,
              color: "var(--ei-color-ink3)",
              fontFamily: "var(--ei-mono)",
            }}
          >
            {t("resumeWorkshop.rewrites.diff.originalFrom")}
          </span>
        </div>
        <div
          data-testid="resume-rewrites-original-text"
          style={{
            fontSize: 14.5,
            color: "var(--ei-color-ink2)",
            lineHeight: 1.65,
            fontFamily: "var(--ei-serif)",
            background: "var(--ei-color-danger-soft)",
            padding: "12px 14px",
            borderRadius: 2,
            borderLeft: "2px solid var(--ei-color-danger)",
          }}
        >
          {bullet.original}
        </div>
      </div>

      <div
        style={{
          padding: "16px 22px",
          borderBottom: "1px dotted var(--ei-color-rule)",
        }}
      >
        <div
          style={{
            display: "flex",
            gap: 10,
            alignItems: "center",
            marginBottom: 8,
          }}
        >
          <span
            style={{
              padding: "2px 8px",
              background: "var(--ei-color-ok-soft)",
              color: "var(--ei-color-ok)",
              fontSize: 10.5,
              fontFamily: "var(--ei-mono)",
              letterSpacing: "0.08em",
              borderRadius: 2,
            }}
          >
            + {t("resumeWorkshop.rewrites.diff.rewritten")}
          </span>
          <span
            style={{
              fontSize: 11,
              color: "var(--ei-color-ink3)",
              fontFamily: "var(--ei-mono)",
            }}
          >
            {t("resumeWorkshop.rewrites.diff.confidence")}
          </span>
        </div>
        <div
          data-testid="resume-rewrites-rewritten-text"
          style={{
            fontSize: 14.5,
            color: "var(--ei-color-ink)",
            lineHeight: 1.65,
            fontFamily: "var(--ei-serif)",
            background: "var(--ei-color-ok-soft)",
            padding: "12px 14px",
            borderRadius: 2,
            borderLeft: "2px solid var(--ei-color-ok)",
          }}
        >
          {bullet.rewritten}
        </div>
      </div>

      <div style={{ padding: "16px 22px" }}>
        <div
          className="ei-text-label"
          style={{ color: "var(--ei-color-ink3)", marginBottom: 8 }}
        >
          {t("resumeWorkshop.rewrites.diff.whyEyebrow")}
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {bullet.why.map((w, i) => (
            <div
              key={`${bullet.id}-why-${i}`}
              data-testid={`resume-rewrites-why-${i}`}
              style={{
                display: "flex",
                gap: 10,
                fontSize: 13,
                color: "var(--ei-color-ink2)",
                lineHeight: 1.5,
              }}
            >
              <span>{w}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

interface RewriteSaveConfirmModalProps {
  resumeName: string;
  bullets: UiBullet[];
  saving: boolean;
  onClose: () => void;
  onConfirm: (mode: "overwrite" | "new") => void;
  t: (key: MessageKey) => string;
}

const RewriteSaveConfirmModal: FC<RewriteSaveConfirmModalProps> = ({
  resumeName,
  bullets,
  saving,
  onClose,
  onConfirm,
  t,
}) => {
  const [mode, setMode] = useState<"overwrite" | "new">("overwrite");
  return (
    <div
      data-testid="resume-rewrites-save-modal-overlay"
      role="presentation"
      style={MODAL_OVERLAY_STYLE}
      onClick={onClose}
    >
      <div
        data-testid="resume-rewrites-save-modal"
        role="dialog"
        aria-modal="true"
        aria-label={t("resumeWorkshop.rewrites.save.title")}
        style={MODAL_CARD_STYLE}
        onClick={(e) => e.stopPropagation()}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "flex-start",
            gap: 18,
            marginBottom: 18,
          }}
        >
          <div>
            <div
              className="ei-text-label"
              style={{ color: "var(--ei-color-accent)", marginBottom: 6 }}
            >
              {t("resumeWorkshop.rewrites.save.eyebrow")}
            </div>
            <div
              className="ei-text-title"
              data-testid="resume-rewrites-save-modal-title"
            >
              {t("resumeWorkshop.rewrites.save.title")}
            </div>
            <div
              style={{
                fontSize: 13,
                color: "var(--ei-color-ink3)",
                marginTop: 6,
                lineHeight: 1.6,
              }}
            >
              {t("resumeWorkshop.rewrites.save.sub")}
            </div>
          </div>
          <button
            type="button"
            data-testid="resume-rewrites-save-modal-close"
            onClick={onClose}
            aria-label={t("resumeWorkshop.rewrites.save.close")}
            style={{
              background: "transparent",
              border: "none",
              color: "var(--ei-color-ink3)",
              cursor: "pointer",
              padding: 4,
            }}
          >
            ×
          </button>
        </div>

        <div
          data-testid="resume-rewrites-save-modal-list"
          style={{
            border: "1px solid var(--ei-color-rule)",
            background: "var(--ei-color-bg-soft)",
            borderRadius: 3,
            padding: 16,
            marginBottom: 16,
          }}
        >
          <div
            className="ei-text-label"
            style={{ color: "var(--ei-color-ink3)", marginBottom: 10 }}
          >
            {t("resumeWorkshop.rewrites.save.acceptedCount").replace(
              "{count}",
              String(bullets.length),
            )}
          </div>
          {bullets.map((b, i) => (
            <div
              key={b.id}
              data-testid={`resume-rewrites-save-modal-item-${b.id}`}
              style={{
                padding: "10px 0",
                borderBottom:
                  i < bullets.length - 1
                    ? "1px dotted var(--ei-color-rule)"
                    : "none",
              }}
            >
              <div
                style={{
                  fontSize: 11,
                  color: "var(--ei-color-ink3)",
                  fontFamily: "var(--ei-mono)",
                  marginBottom: 4,
                }}
              >
                {b.section || t("resumeWorkshop.rewrites.diff.sectionFallback")}
              </div>
              <div
                style={{
                  fontSize: 13.5,
                  color: "var(--ei-color-ink)",
                  lineHeight: 1.6,
                }}
              >
                {b.rewritten}
              </div>
            </div>
          ))}
        </div>

        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            gap: 10,
            marginBottom: 18,
          }}
        >
          {(
            [
              {
                k: "overwrite" as const,
                label: t("resumeWorkshop.rewrites.save.overwriteLabel"),
                desc: t("resumeWorkshop.rewrites.save.overwriteDesc").replace(
                  "{resumeName}",
                  resumeName,
                ),
              },
              {
                k: "new" as const,
                label: t("resumeWorkshop.rewrites.save.newLabel"),
                desc: t("resumeWorkshop.rewrites.save.newDesc"),
              },
            ]
          ).map((m) => {
            const on = mode === m.k;
            return (
              <button
                key={m.k}
                type="button"
                data-testid={`resume-rewrites-save-mode-${m.k}`}
                data-active={on ? "true" : "false"}
                aria-pressed={on}
                onClick={() => setMode(m.k)}
                style={{
                  textAlign: "left",
                  padding: "14px 14px",
                  background: on
                    ? "var(--ei-color-accent-soft)"
                    : "var(--ei-color-bg)",
                  border: `1px solid ${
                    on ? "var(--ei-color-accent)" : "var(--ei-color-rule)"
                  }`,
                  borderRadius: 2,
                  cursor: "pointer",
                }}
              >
                <div
                  className="ei-text-label"
                  style={{
                    color: on ? "var(--ei-color-accent)" : "var(--ei-color-ink3)",
                    marginBottom: 6,
                  }}
                >
                  {m.label}
                </div>
                <div
                  style={{
                    fontSize: 12,
                    color: "var(--ei-color-ink2)",
                    lineHeight: 1.5,
                  }}
                >
                  {m.desc}
                </div>
              </button>
            );
          })}
        </div>

        <div
          style={{ display: "flex", justifyContent: "flex-end", gap: 10 }}
        >
          <button
            type="button"
            data-testid="resume-rewrites-save-cancel"
            onClick={onClose}
            style={BTN_GHOST}
          >
            {t("resumeWorkshop.rewrites.save.cancel")}
          </button>
          <button
            type="button"
            data-testid="resume-rewrites-save-confirm"
            onClick={() => onConfirm(mode)}
            disabled={saving}
            aria-disabled={saving}
            style={{ ...BTN_FILLED, opacity: saving ? 0.55 : 1 }}
          >
            {mode === "new"
              ? t("resumeWorkshop.rewrites.save.confirmNew")
              : t("resumeWorkshop.rewrites.save.confirmOverwrite")}
          </button>
        </div>
      </div>
    </div>
  );
};

interface PollingBannerProps {
  banner: ReactPollingBanner;
  t: (key: MessageKey) => string;
}

const PollingBanner: FC<PollingBannerProps> = ({ banner, t }) => {
  if (banner.kind === "info") {
    return (
      <div
        data-testid="resume-rewrites-polling-banner"
        role="status"
        aria-live="polite"
        style={POLLING_BANNER_INFO}
      >
        {banner.message}
      </div>
    );
  }
  return (
    <div
      data-testid="resume-rewrites-failed-banner"
      role="alert"
      style={POLLING_BANNER_DANGER}
    >
      <span>{banner.message}</span>
      <button
        type="button"
        data-testid="resume-rewrites-polling-retry"
        onClick={banner.onRetry}
        style={BTN_GHOST}
      >
        {t("resumeWorkshop.rewrites.polling.retry")}
      </button>
    </div>
  );
};

interface EmptyStateProps {
  t: (key: MessageKey) => string;
  onRerun?: () => void;
}

const EmptyState: FC<EmptyStateProps> = ({ t, onRerun }) => (
  <div data-testid="resume-rewrites-empty" className="ei-screen-card">
    <span className="ei-text-label">
      {t("resumeWorkshop.rewrites.empty.title")}
    </span>
    <p className="ei-text-body">{t("resumeWorkshop.rewrites.empty.body")}</p>
    {onRerun ? (
      <button
        type="button"
        data-testid="resume-rewrites-rerun-tailor"
        onClick={onRerun}
        style={BTN_GHOST}
      >
        {t("resumeWorkshop.rewrites.rerun.cta")}
      </button>
    ) : null}
  </div>
);

const DIFF_CARD_STYLE: CSSProperties = {
  background: "var(--ei-color-bg-card)",
  border: "1px solid var(--ei-color-rule)",
  borderRadius: 2,
};

const MODAL_OVERLAY_STYLE: CSSProperties = {
  position: "fixed",
  inset: 0,
  background: "rgba(24, 20, 16, 0.24)",
  zIndex: 80,
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  padding: 24,
};

const MODAL_CARD_STYLE: CSSProperties = {
  width: "min(760px, 100%)",
  maxHeight: "88vh",
  overflow: "auto",
  background: "var(--ei-color-bg-card)",
  border: "1px solid var(--ei-color-rule)",
  borderRadius: 4,
  boxShadow: "0 24px 70px rgba(30, 22, 15, 0.24)",
  padding: 24,
};

const POLLING_BANNER_INFO: CSSProperties = {
  marginBottom: 16,
  padding: "10px 14px",
  background: "var(--ei-color-accent-soft)",
  color: "var(--ei-color-accent)",
  border: "1px solid var(--ei-color-accent)",
  borderRadius: 2,
  fontSize: 13,
};

const POLLING_BANNER_DANGER: CSSProperties = {
  marginBottom: 16,
  padding: "10px 14px",
  background: "var(--ei-color-danger-soft)",
  color: "var(--ei-color-danger)",
  border: "1px solid var(--ei-color-danger)",
  borderRadius: 2,
  fontSize: 13,
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: 8,
};

const BTN_FILLED: CSSProperties = {
  padding: "5px 12px",
  fontSize: 12,
  cursor: "pointer",
  background: "var(--ei-color-accent)",
  color: "#fff",
  border: "none",
  borderRadius: 2,
  fontFamily: "var(--ei-sans)",
};

const BTN_GHOST: CSSProperties = {
  padding: "5px 12px",
  fontSize: 12,
  cursor: "pointer",
  background: "transparent",
  color: "var(--ei-color-ink2)",
  border: "1px solid var(--ei-color-rule)",
  borderRadius: 2,
  fontFamily: "var(--ei-sans)",
};
