import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type CSSProperties,
  type FC,
} from "react";

import type { ResumeVersion } from "../../../../api/generated/types";
import { useI18n, type MessageKey } from "../../../i18n/messages";
import {
  mapBulletSuggestionToUi,
  type ResumeSuggestionInput,
  type UiBullet,
  type UiBulletStatus,
} from "../adapters/resume";

export interface ResumeRewritesTabActions {
  /** Accept the suggestion (writes status=accepted, no structured_profile mutation). */
  onAccept?: (bulletId: string) => Promise<void> | void;
  /** Reject the suggestion (writes status=rejected). */
  onReject?: (bulletId: string) => Promise<void> | void;
  /**
   * Save a manual edit:
   *  1) update resume version structuredProfile.manualEdits[]
   *  2) call bodyless accept to mark terminal accepted (Phase 4).
   */
  onSaveManualEdit?: (
    bulletId: string,
    editedText: string,
  ) => Promise<void> | void;
  /** Trigger a new tailor run for this version (Phase 5). */
  onRequestRerun?: (mode: "bullet_suggestions" | "gap_review") => void;
}

export interface ResumeRewritesTabProps extends ResumeRewritesTabActions {
  version: ResumeVersion;
  /**
   * Inline status banner shown above the rewrites list (Phase 5 polling). The
   * tab component renders the banner content but does not own polling state.
   */
  pollingBanner?: ReactPollingBanner | null;
  /** Phase 4 wires this to the per-row "saved manual edit pending accept" state. */
  manualEditPendingFor?: string | null;
}

export type ReactPollingBanner =
  | { kind: "info"; message: string }
  | { kind: "danger"; message: string; onRetry: () => void };

const toSuggestionInput = (raw: Record<string, unknown>): ResumeSuggestionInput => {
  const safe = (key: string): string =>
    typeof raw[key] === "string" ? (raw[key] as string) : "";
  const optString = (key: string): string | undefined =>
    typeof raw[key] === "string" ? (raw[key] as string) : undefined;
  return {
    id: safe("id"),
    originalBullet: safe("originalBullet"),
    suggestedBullet: safe("suggestedBullet"),
    reason: safe("reason"),
    status: optString("status"),
    section: optString("section") ?? optString("sectionLabel"),
    decidedAt: typeof raw.decidedAt === "string" ? (raw.decidedAt as string) : null,
    source: optString("source"),
    tailorRunId:
      typeof raw.tailorRunId === "string"
        ? (raw.tailorRunId as string)
        : null,
  };
};

export const ResumeRewritesTab: FC<ResumeRewritesTabProps> = ({
  version,
  onAccept,
  onReject,
  onSaveManualEdit,
  onRequestRerun,
  pollingBanner = null,
  manualEditPendingFor = null,
}) => {
  const { t } = useI18n();
  const bullets = useMemo<UiBullet[]>(
    () =>
      (version.suggestions ?? [])
        .map((entry) => toSuggestionInput(entry))
        .filter((entry) => entry.id !== "")
        .map(mapBulletSuggestionToUi),
    [version.suggestions],
  );

  const [selectedBulletId, setSelectedBulletId] = useState<string | null>(
    bullets[0]?.id ?? null,
  );
  const [editing, setEditing] = useState(false);
  const [editText, setEditText] = useState("");
  const [pendingActionId, setPendingActionId] = useState<string | null>(null);

  useEffect(() => {
    if (bullets.length === 0) {
      setSelectedBulletId(null);
      return;
    }
    if (!bullets.some((b) => b.id === selectedBulletId)) {
      setSelectedBulletId(bullets[0]!.id);
      setEditing(false);
      setEditText("");
    }
  }, [bullets, selectedBulletId]);

  const selected = useMemo(
    () => bullets.find((b) => b.id === selectedBulletId) ?? null,
    [bullets, selectedBulletId],
  );

  const counts = useMemo(() => {
    const result: Record<UiBulletStatus, number> = {
      accepted: 0,
      pending: 0,
      rejected: 0,
    };
    for (const b of bullets) {
      result[b.status] += 1;
    }
    return result;
  }, [bullets]);

  const versionName = version.displayName;

  const handleSelect = useCallback((id: string) => {
    setSelectedBulletId(id);
    setEditing(false);
    setEditText("");
  }, []);

  const wrapAction = useCallback(
    (id: string, fn?: (bulletId: string) => Promise<void> | void) => {
      if (!fn) return undefined;
      return async () => {
        setPendingActionId(id);
        try {
          await fn(id);
        } finally {
          setPendingActionId((current) => (current === id ? null : current));
        }
      };
    },
    [],
  );

  return (
    <div
      data-testid="resume-rewrites-tab"
      data-version-id={version.id}
      data-bullet-count={bullets.length}
      data-pending-count={counts.pending}
      data-accepted-count={counts.accepted}
      data-rejected-count={counts.rejected}
      data-selected-bullet-id={selectedBulletId ?? ""}
      data-editing={editing ? "true" : "false"}
    >
      <ScopeBanner
        versionName={versionName}
        counts={counts}
        translateKey={(key) => t(key)}
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
            bullets={bullets}
            selectedBulletId={selectedBulletId}
            onSelect={handleSelect}
            t={t}
            pendingActionId={pendingActionId}
          />
          <div>
            {selected ? (
              <DiffDetailCard
                bullet={selected}
                editing={editing}
                editText={editText}
                versionName={versionName}
                manualEditPendingFor={manualEditPendingFor}
                pendingActionId={pendingActionId}
                onEditStart={() => {
                  setEditing(true);
                  setEditText(selected.rewritten);
                }}
                onEditChange={(next) => setEditText(next)}
                onEditCancel={() => setEditing(false)}
                onEditSave={async () => {
                  if (!onSaveManualEdit) return;
                  setPendingActionId(selected.id);
                  try {
                    await onSaveManualEdit(selected.id, editText);
                    setEditing(false);
                  } finally {
                    setPendingActionId((current) =>
                      current === selected.id ? null : current,
                    );
                  }
                }}
                onReject={wrapAction(selected.id, onReject)}
                onAccept={wrapAction(selected.id, onAccept)}
                onRerunTailor={
                  onRequestRerun
                    ? () => onRequestRerun("bullet_suggestions")
                    : undefined
                }
                t={t}
              />
            ) : null}
          </div>
        </div>
      )}
    </div>
  );
};

interface ScopeBannerProps {
  versionName: string;
  counts: Record<UiBulletStatus, number>;
  translateKey: (key: MessageKey) => string;
}

const ScopeBanner: FC<ScopeBannerProps> = ({
  versionName,
  counts,
  translateKey,
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
      {translateKey("resumeWorkshop.rewrites.scopeBanner.body").replace(
        "{versionName}",
        versionName,
      )}
    </div>
    <div
      data-testid="resume-rewrites-counts"
      style={{
        fontSize: 11,
        color: "var(--ei-color-ink3)",
        fontFamily: "var(--ei-mono)",
      }}
    >
      {translateKey("resumeWorkshop.rewrites.scopeBanner.counts")
        .replace("{accepted}", String(counts.accepted))
        .replace("{pending}", String(counts.pending))
        .replace("{rejected}", String(counts.rejected))}
    </div>
  </div>
);

interface BulletListProps {
  bullets: UiBullet[];
  selectedBulletId: string | null;
  onSelect: (id: string) => void;
  t: (key: MessageKey) => string;
  pendingActionId: string | null;
}

const STATUS_TO_CHIP_KEY: Record<UiBulletStatus, MessageKey> = {
  accepted: "resumeWorkshop.rewrites.status.accepted",
  pending: "resumeWorkshop.rewrites.status.pending",
  rejected: "resumeWorkshop.rewrites.status.rejected",
};

const STATUS_TO_TONE: Record<UiBulletStatus, string> = {
  accepted: "var(--ei-color-ok)",
  pending: "var(--ei-color-warn)",
  rejected: "var(--ei-color-ink4)",
};

const truncate = (text: string, max: number): string =>
  text.length > max ? `${text.slice(0, max)}…` : text;

const BulletList: FC<BulletListProps> = ({
  bullets,
  selectedBulletId,
  onSelect,
  t,
  pendingActionId,
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
        const tone = STATUS_TO_TONE[b.status];
        const pending = pendingActionId === b.id;
        return (
          <button
            key={b.id}
            type="button"
            role="option"
            aria-selected={active}
            data-testid={`resume-rewrites-bullet-row-${b.id}`}
            data-status={b.status}
            data-source={b.source}
            data-pending={pending ? "true" : "false"}
            onClick={() => onSelect(b.id)}
            style={{
              padding: "14px 16px",
              textAlign: "left",
              cursor: "pointer",
              background: active
                ? "var(--ei-color-bg-soft)"
                : "var(--ei-color-bg-card)",
              border: `1px solid ${
                active
                  ? "var(--ei-color-accent)"
                  : "var(--ei-color-rule)"
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
                {t(STATUS_TO_CHIP_KEY[b.status])}
              </div>
            </div>
            <div
              style={{
                fontSize: 13,
                color: "var(--ei-color-ink2)",
                lineHeight: 1.5,
                textDecoration: b.status === "rejected" ? "line-through" : "none",
                opacity: b.status === "rejected" ? 0.6 : 1,
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
  editing: boolean;
  editText: string;
  versionName: string;
  manualEditPendingFor: string | null;
  pendingActionId: string | null;
  onEditStart: () => void;
  onEditChange: (next: string) => void;
  onEditCancel: () => void;
  onEditSave: () => Promise<void> | void;
  onReject?: () => Promise<void>;
  onAccept?: () => Promise<void>;
  onRerunTailor?: () => void;
  t: (key: MessageKey) => string;
}

const DiffDetailCard: FC<DiffDetailCardProps> = ({
  bullet,
  editing,
  editText,
  versionName,
  manualEditPendingFor,
  pendingActionId,
  onEditStart,
  onEditChange,
  onEditCancel,
  onEditSave,
  onReject,
  onAccept,
  onRerunTailor,
  t,
}) => {
  const isAccepted = bullet.status === "accepted";
  const isRejected = bullet.status === "rejected";
  const manualPending = manualEditPendingFor === bullet.id;
  const acting = pendingActionId === bullet.id;

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
        <div
          className="ei-text-label"
          style={{ color: "var(--ei-color-ink3)" }}
        >
          {bullet.section || t("resumeWorkshop.rewrites.diff.sectionFallback")}
        </div>
        <div style={{ display: "flex", gap: 6 }}>
          <button
            type="button"
            data-testid="resume-rewrites-action-reject"
            onClick={onReject}
            disabled={!onReject || acting}
            aria-disabled={!onReject || acting}
            aria-label={t("resumeWorkshop.rewrites.action.reject")}
            style={{
              ...BTN_OUTLINE,
              background: isRejected
                ? "var(--ei-color-ink2)"
                : "transparent",
              color: isRejected ? "var(--ei-color-bg)" : "var(--ei-color-ink3)",
              opacity: !onReject ? 0.6 : 1,
            }}
          >
            {t("resumeWorkshop.rewrites.action.reject")}
          </button>
          <button
            type="button"
            data-testid="resume-rewrites-action-edit"
            onClick={() => onEditStart()}
            disabled={acting}
            aria-label={t("resumeWorkshop.rewrites.action.edit")}
            aria-pressed={editing}
            style={{
              ...BTN_OUTLINE,
              background: editing
                ? "var(--ei-color-bg-soft)"
                : "transparent",
              borderColor: editing
                ? "var(--ei-color-accent)"
                : "var(--ei-color-rule)",
            }}
          >
            {t("resumeWorkshop.rewrites.action.edit")}
          </button>
          <button
            type="button"
            data-testid="resume-rewrites-action-accept"
            onClick={onAccept}
            disabled={!onAccept || acting}
            aria-disabled={!onAccept || acting}
            aria-label={t("resumeWorkshop.rewrites.action.accept")}
            style={{
              ...BTN_FILLED,
              background: isAccepted
                ? "var(--ei-color-ok)"
                : "var(--ei-color-accent)",
            }}
          >
            {isAccepted
              ? t("resumeWorkshop.rewrites.action.accepted")
              : t("resumeWorkshop.rewrites.action.accept")}
          </button>
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
              background: editing
                ? "var(--ei-color-accent-soft)"
                : "var(--ei-color-ok-soft)",
              color: editing
                ? "var(--ei-color-accent)"
                : "var(--ei-color-ok)",
              fontSize: 10.5,
              fontFamily: "var(--ei-mono)",
              letterSpacing: "0.08em",
              borderRadius: 2,
            }}
          >
            +{" "}
            {editing
              ? t("resumeWorkshop.rewrites.diff.manualEdit")
              : t("resumeWorkshop.rewrites.diff.rewritten")}
          </span>
          <span
            style={{
              fontSize: 11,
              color: "var(--ei-color-ink3)",
              fontFamily: "var(--ei-mono)",
            }}
          >
            {editing
              ? t("resumeWorkshop.rewrites.diff.manualHint").replace(
                  "{versionName}",
                  versionName,
                )
              : t("resumeWorkshop.rewrites.diff.confidence")}
          </span>
        </div>
        {editing ? (
          <div>
            <textarea
              data-testid="resume-rewrites-edit-textarea"
              value={editText}
              onChange={(e) => onEditChange(e.target.value)}
              style={{
                width: "100%",
                minHeight: 110,
                padding: "12px 14px",
                border: "1px solid var(--ei-color-accent)",
                background: "var(--ei-color-accent-soft)",
                color: "var(--ei-color-ink)",
                borderRadius: 2,
                fontFamily: "var(--ei-serif)",
                fontSize: 14.5,
                lineHeight: 1.65,
                resize: "vertical",
                outline: "none",
              }}
            />
            <div
              style={{
                display: "flex",
                justifyContent: "flex-end",
                gap: 8,
                marginTop: 10,
              }}
            >
              <button
                type="button"
                data-testid="resume-rewrites-edit-cancel"
                onClick={onEditCancel}
                style={BTN_GHOST}
              >
                {t("resumeWorkshop.rewrites.action.cancelEdit")}
              </button>
              <button
                type="button"
                data-testid="resume-rewrites-edit-save"
                onClick={() => void onEditSave()}
                style={BTN_FILLED}
              >
                {t("resumeWorkshop.rewrites.action.saveManual")}
              </button>
            </div>
            {manualPending ? (
              <div
                data-testid="resume-rewrites-manual-pending"
                role="alert"
                style={MANUAL_PENDING_STYLE}
              >
                {t("resumeWorkshop.rewrites.error.manualPendingRetry")}
              </div>
            ) : null}
          </div>
        ) : (
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
        )}
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
        {onRerunTailor ? (
          <div style={{ marginTop: 16 }}>
            <button
              type="button"
              data-testid="resume-rewrites-rerun-tailor"
              onClick={onRerunTailor}
              style={BTN_GHOST}
            >
              {t("resumeWorkshop.rewrites.rerun.cta")}
            </button>
          </div>
        ) : null}
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

const MANUAL_PENDING_STYLE: CSSProperties = {
  marginTop: 10,
  padding: "8px 12px",
  background: "var(--ei-color-warn-soft)",
  color: "var(--ei-color-warn)",
  border: "1px solid var(--ei-color-warn)",
  borderRadius: 2,
  fontSize: 12,
};

const BTN_OUTLINE: CSSProperties = {
  padding: "5px 12px",
  fontSize: 12,
  cursor: "pointer",
  background: "transparent",
  color: "var(--ei-color-ink3)",
  border: "1px solid var(--ei-color-rule)",
  borderRadius: 2,
  fontFamily: "var(--ei-sans)",
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
