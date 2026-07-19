import type { CSSProperties, FC, KeyboardEvent, MouseEvent, ReactNode } from "react";

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
  <div
    className="ei-workspace-card-rail-track"
    style={{ "--ei-workspace-round-count": rounds.length } as CSSProperties}
  >
    {rounds.map((round, i) => {
      const done = currentIndex !== null && i < currentIndex;
      const current = i === currentIndex;
      return (
        <div
          key={round.id}
          className="ei-workspace-card-round"
          data-round-state={done ? "done" : current ? "current" : "pending"}
        >
          <div className="ei-workspace-card-round-node" aria-hidden="true">
                {done ? (
                  <svg
                    width={16}
                    height={16}
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
                  <span>{i + 1}</span>
                )}
          </div>
          <div className="ei-workspace-card-round-name">{round.name}</div>
        </div>
      );
    })}
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
  presentation?: "workspace-card" | "home-record";
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
  presentation = "workspace-card",
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
    <article
      data-testid={cardTestId ?? `home-recent-mock-card-${job.id}`}
      data-presentation="workspace-card"
      className="ei-workspace-card"
      role={onClick ? "button" : undefined}
      tabIndex={onClick ? 0 : undefined}
      onClick={onClick}
      onKeyDown={openFromKeyboard}
    >
      {deleteAction ? (
        <button
          data-testid={deleteAction.testId}
          type="button"
          aria-label={deleteAction.label}
          title={deleteAction.label}
          className="ei-workspace-card-delete"
          disabled={deleteAction.disabled}
          onClick={(event) => runAction(event, deleteAction)}
        >
          <ResumeWorkshopIcon name="trash" size={20} />
        </button>
      ) : null}
      <div
        data-testid={bodyTestId}
        className="ei-workspace-card-heading"
      >
        <span className="ei-workspace-card-company-icon" aria-hidden="true">
          <svg width="29" height="29" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
            <path d="M4 21h16M6 21V5h8v16M14 9h4v12M9 8h2M9 12h2M9 16h2M16 12h1M16 16h1" />
          </svg>
        </span>
        <div className="ei-workspace-card-heading-copy">
          <div className="ei-workspace-card-company">
            {job.companyName.toUpperCase()}
          </div>
          <h2 className="ei-workspace-card-title">
            {job.title}
          </h2>
          {location ? (
            <div className="ei-workspace-card-location">
              {location}
            </div>
          ) : null}
        </div>
      </div>
      <div
        data-testid={railTestId ?? `home-recent-mock-rail-${job.id}`}
        className="ei-workspace-card-rail"
      >
        <MiniRoundRail
          rounds={rounds}
          currentIndex={ci}
        />
      </div>
      {hasFooter ? (
        <div
          data-testid={footerTestId}
          className="ei-workspace-card-footer"
        >
          {footer}
          {primaryAction ? (
            <button
              data-testid={primaryAction.testId}
              type="button"
              className="ei-workspace-card-primary"
              disabled={primaryAction.disabled}
              onClick={(event) => runAction(event, primaryAction)}
            >
              <svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="9" />
                <path d="m10 8 6 4-6 4z" />
              </svg>
              {primaryAction.label}
            </button>
          ) : null}
        </div>
      ) : null}
    </article>
  );
};
