import type { FC, KeyboardEvent, MouseEvent, ReactNode } from "react";

import type { TargetJob } from "../../../api/generated/types";
import { useI18n } from "../../i18n/messages";
import {
  buildTargetJobRoundAssumptions,
  resolveTargetJobPracticeProgress,
} from "../../interview-context/roundAssumptions";
import { ResumeWorkshopIcon } from "../resume-workshop/components/ResumeWorkshopIcon";

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
  presentation?: "card" | "home-record";
  recentMeta?: ReactNode;
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
  presentation = "card",
  recentMeta,
}) => {
  const { t } = useI18n();
  const rounds = buildTargetJobRoundAssumptions(job, t);
  const ci = resolveTargetJobPracticeProgress(job).currentIndex;
  const hasFooter = Boolean(footer || primaryAction);
  const location = job.locationText?.trim();

  const runAction = (
    event: MouseEvent<HTMLButtonElement>,
    action: MockInterviewCardAction,
  ) => {
    event.stopPropagation();
    if (!action.disabled) {
      void action.onClick();
    }
  };

  const openFromKeyboard = (event: KeyboardEvent<HTMLElement>) => {
    if (
      event.target !== event.currentTarget ||
      !onClick ||
      (event.key !== "Enter" && event.key !== " ")
    ) return;
    event.preventDefault();
    onClick();
  };

  if (presentation === "home-record") {
    return (
      <article
        data-testid={cardTestId ?? `home-recent-mock-card-${job.id}`}
        data-presentation="home-record"
        className="ei-home-recent-record"
        role={onClick ? "button" : undefined}
        tabIndex={onClick ? 0 : undefined}
        onClick={onClick}
        onKeyDown={openFromKeyboard}
      >
        <span className="ei-home-recent-building" aria-hidden="true">
          <svg width="25" height="25" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
            <path d="M4 21h16M6 21V5h8v16M14 9h4v12M9 8h2M9 12h2M9 16h2M16 12h1M16 16h1" />
          </svg>
        </span>
        <div className="ei-home-recent-main">
          <div className="ei-home-recent-company">{job.companyName}</div>
          <h3 className="ei-home-recent-title">{job.title}</h3>
          <div
            data-testid={railTestId ?? `home-recent-mock-rail-${job.id}`}
            className="ei-home-recent-rail"
          >
            {rounds.map((round, index) => {
              const done = ci !== null && index < ci;
              const current = index === ci;
              return (
                <span
                  key={round.id}
                  className="ei-home-recent-step"
                  data-round-state={done ? "done" : current ? "current" : "pending"}
                >
                  <span className="ei-home-recent-step-mark" aria-hidden="true">
                    {done ? "✓" : index + 1}
                  </span>
                  <span className="ei-home-recent-step-name">{round.name}</span>
                </span>
              );
            })}
          </div>
        </div>
        <div className="ei-home-recent-end" data-testid={footerTestId}>
          {recentMeta}
          {primaryAction ? (
            <button
              data-testid={primaryAction.testId}
              type="button"
              className="ei-home-recent-action"
              disabled={primaryAction.disabled}
              onClick={(event) => runAction(event, primaryAction)}
            >
              {primaryAction.label}
              <svg aria-hidden="true" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                <path d="M5 12h14M13 6l6 6-6 6" />
              </svg>
            </button>
          ) : null}
        </div>
      </article>
    );
  }

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
          className="ei-mock-interview-card-delete"
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
            {job.companyName.toUpperCase()}
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
          {location ? (
            <div
              style={{
                fontSize: 12.5,
                color: "var(--ei-color-fg-tertiary)",
                marginTop: 4,
              }}
            >
              {location}
            </div>
          ) : null}
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
