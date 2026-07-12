import type { FC, MouseEvent, ReactNode } from "react";

import type { TargetJob, TargetJobStatus } from "../../../api/generated/types";
import { useI18n } from "../../i18n/messages";
import {
  buildTargetJobRoundAssumptions,
  resolveTargetJobPracticeProgress,
} from "../../interview-context/roundAssumptions";
import { ResumeWorkshopIcon } from "../resume-workshop/components/ResumeWorkshopIcon";

function statusTone(
  status: TargetJobStatus,
): "muted" | "amber" | "neutral" {
  switch (status) {
    case "draft":
    case "preparing":
      return "muted";
    case "applied":
    case "interviewing":
      return "amber";
    case "offer":
    case "rejected":
    case "archived":
      return "neutral";
  }
}

function statusLabel(status: TargetJobStatus): string {
  switch (status) {
    case "draft":
      return "Draft";
    case "preparing":
      return "Preparing";
    case "applied":
      return "Applied";
    case "interviewing":
      return "Interviewing";
    case "offer":
      return "Offer";
    case "rejected":
      return "Rejected";
    case "archived":
      return "Archived";
  }
}

const statusColorMap: Record<
  string,
  { bg: string; fg: string }
> = {
  amber: {
    bg: "var(--ei-color-amber-soft)",
    fg: "var(--ei-color-amber)",
  },
  neutral: {
    bg: "var(--ei-color-bg-soft)",
    fg: "var(--ei-color-fg-secondary)",
  },
  muted: {
    bg: "transparent",
    fg: "var(--ei-color-fg-tertiary)",
  },
};

// ─── MiniRoundRail ────────────────────────────────────────────────

interface MiniRoundRailProps {
  rounds: Array<{ id: string; name: string }>;
  currentIndex: number | null;
}

const MiniRoundRail: FC<MiniRoundRailProps> = ({ rounds, currentIndex }) => (
  <div style={{ marginTop: 18 }}>
    <div style={{ position: "relative", height: 34 }}>
      <div
        style={{
          position: "absolute",
          top: 9,
          left: 8,
          right: 8,
          height: 1,
          background: "var(--ei-color-rule-strong)",
        }}
      />
      <div
        style={{
          display: "grid",
          gridTemplateColumns: `repeat(${rounds.length}, 1fr)`,
        }}
      >
        {rounds.map((round, i) => {
          const done = currentIndex !== null && i < currentIndex;
          const current = i === currentIndex;
          const align =
            i === 0
              ? "flex-start"
              : i === rounds.length - 1
                ? "flex-end"
                : "center";
          const textAlign =
            i === 0
              ? "left"
              : i === rounds.length - 1
                ? "right"
                : "center";
          return (
            <div
              key={round.id}
              data-round-state={done ? "done" : current ? "current" : "pending"}
              style={{
                position: "relative",
                display: "flex",
                flexDirection: "column",
                alignItems: align,
              }}
            >
              <div
                style={{
                  width: 18,
                  height: 18,
                  borderRadius: 9,
                  border: `1px solid ${
                    done
                      ? "var(--ei-color-ok)"
                      : current
                        ? "var(--ei-color-accent)"
                        : "var(--ei-color-rule-strong)"
                  }`,
                  background: done
                    ? "var(--ei-color-ok)"
                    : current
                      ? "var(--ei-color-accent)"
                      : "var(--ei-color-bg-card)",
                  color:
                    done || current
                      ? "#fff"
                      : "var(--ei-color-fg-tertiary)",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  zIndex: 1,
                }}
              >
                {done ? (
                  <svg
                    width={10}
                    height={10}
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="#fff"
                    strokeWidth={2.2}
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <path d="M5 12l5 5L20 7" />
                  </svg>
                ) : (
                  <span
                    style={{
                      fontSize: 9,
                      fontFamily: "var(--ei-font-mono)",
                    }}
                  >
                    {i + 1}
                  </span>
                )}
              </div>
              <div
                style={{
                  marginTop: 6,
                  fontSize: 10.5,
                  color: current
                    ? "var(--ei-color-fg-primary)"
                    : "var(--ei-color-fg-tertiary)",
                  maxWidth: 68,
                  textAlign,
                  whiteSpace: "nowrap",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                }}
              >
                {round.name}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  </div>
);

// ─── MockInterviewCard ────────────────────────────────────────────

interface MockInterviewCardAction {
  label: string;
  testId: string;
  onClick: () => void | Promise<void>;
  disabled?: boolean;
}

interface MockInterviewCardProps {
  job: TargetJob;
  onClick?: () => void;
  cardTestId?: string;
  bodyTestId?: string;
  railTestId?: string;
  footerTestId?: string;
  footer?: ReactNode;
  primaryAction?: MockInterviewCardAction;
  deleteAction?: MockInterviewCardAction;
}

export const MockInterviewCard: FC<MockInterviewCardProps> = ({
  job,
  onClick,
  cardTestId,
  bodyTestId,
  railTestId,
  footerTestId,
  footer,
  primaryAction,
  deleteAction,
}) => {
  const { t } = useI18n();
  const tone = statusTone(job.status);
  const colors = statusColorMap[tone] ?? { bg: "transparent", fg: "var(--ei-color-fg-tertiary)" };
  const rounds = buildTargetJobRoundAssumptions(job, t);
  const ci = resolveTargetJobPracticeProgress(job).currentIndex;
  const hasFooter = Boolean(footer || primaryAction);

  const runAction = (
    event: MouseEvent<HTMLButtonElement>,
    action: MockInterviewCardAction,
  ) => {
    event.stopPropagation();
    if (!action.disabled) {
      void action.onClick();
    }
  };

  return (
    <div
      data-testid={cardTestId ?? `home-recent-mock-card-${job.id}`}
      onClick={onClick}
      style={{
        background: "var(--ei-color-bg-card)",
        border: "1px solid var(--ei-color-rule-strong)",
        borderRadius: 3,
        padding: 20,
        cursor: onClick ? "pointer" : "default",
        display: "flex",
        flexDirection: "column",
        gap: 14,
        position: "relative",
      }}
    >
      {deleteAction ? (
        <button
          data-testid={deleteAction.testId}
          type="button"
          aria-label={deleteAction.label}
          title={deleteAction.label}
          className="ei-resume-workshop-table-delete"
          disabled={deleteAction.disabled}
          onClick={(event) => runAction(event, deleteAction)}
          style={{
            position: "absolute",
            top: 14,
            right: 14,
            zIndex: 1,
          }}
        >
          <ResumeWorkshopIcon name="trash" size={13} />
        </button>
      ) : null}
      <div
        data-testid={bodyTestId}
        style={{
          display: "flex",
          justifyContent: "space-between",
          gap: 12,
          background: "var(--ei-color-bg-card)",
          paddingRight: deleteAction ? 44 : undefined,
        }}
      >
        <div>
          <div
            style={{
              fontSize: 11,
              fontFamily: "var(--ei-font-mono)",
              color: "var(--ei-color-fg-tertiary)",
              marginBottom: 4,
            }}
          >
            {job.companyName.toUpperCase()} ·{" "}
            {statusLabel(job.status)}
          </div>
          <div
            style={{
              fontSize: 19,
              color: "var(--ei-color-fg-primary)",
              fontFamily: "var(--ei-font-serif)",
              letterSpacing: "-0.01em",
            }}
          >
            {job.title}
          </div>
          <div
            style={{
              fontSize: 12.5,
              color: "var(--ei-color-fg-tertiary)",
              marginTop: 4,
            }}
          >
            {job.locationText || "Location not set"}
          </div>
        </div>
        <div
          style={{
            padding: "3px 8px",
            height: "fit-content",
            background: colors.bg,
            color: colors.fg,
            fontSize: 11,
            fontFamily: "var(--ei-font-mono)",
            borderRadius: 2,
            whiteSpace: "nowrap",
          }}
        >
          {statusLabel(job.status)}
        </div>
      </div>
      <div data-testid={railTestId ?? `home-recent-mock-rail-${job.id}`}>
        <MiniRoundRail
          rounds={rounds}
          currentIndex={ci}
        />
      </div>
      {hasFooter ? (
        <div
          data-testid={footerTestId}
          style={{
            borderTop: "1px solid var(--ei-color-rule-strong)",
            paddingTop: 14,
            background: "var(--ei-color-bg-card)",
            display: "flex",
            justifyContent: "flex-end",
            alignItems: "center",
            gap: 12,
          }}
        >
          {footer}
          {primaryAction ? (
            <button
              data-testid={primaryAction.testId}
              type="button"
              disabled={primaryAction.disabled}
              onClick={(event) => runAction(event, primaryAction)}
              style={{
                flex: "0 0 auto",
                height: 32,
                padding: "0 12px",
                fontSize: 13,
                fontWeight: 500,
                background: "var(--ei-color-accent)",
                color: "#fff",
                border: "1px solid var(--ei-color-accent)",
                borderRadius: 2,
                cursor: primaryAction.disabled ? "not-allowed" : "pointer",
                fontFamily: "var(--ei-font-sans)",
                opacity: primaryAction.disabled ? 0.58 : 1,
              }}
            >
              {primaryAction.label}
            </button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
};
