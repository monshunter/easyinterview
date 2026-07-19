import { type FC, type ReactNode } from "react";

import { PhoneIcon } from "./PhoneIcon";

export interface TopBarProps {
  company: string;
  title: string;
  elapsed: string;
  budget: string;
  paused: boolean;
  pauseLabel: string;
  resumeLabel: string;
  onTogglePause: () => void;
  interviewerLabel: string;
  phoneDisabledLabel: string;
  finishCta: ReactNode;
}

export const TopBar: FC<TopBarProps> = ({
  company,
  title,
  elapsed,
  budget,
  paused,
  pauseLabel,
  resumeLabel,
  onTogglePause,
  interviewerLabel,
  phoneDisabledLabel,
  finishCta,
}) => (
  <div data-testid="practice-topbar" className="ei-practice-session-header">
    <div className="ei-practice-session-identity">
      <div className="ei-practice-session-status"><span aria-hidden="true" />{company}</div>
      <div data-testid="practice-topbar-title" className="ei-practice-session-title">{title}</div>
    </div>
    <div className="ei-practice-session-actions">
      <span data-testid="practice-topbar-interviewer" className="ei-practice-session-chip"><RoleIcon />{interviewerLabel}</span>
      <span data-testid="practice-topbar-timer" className="ei-practice-session-clock-group">
        <span className="ei-practice-session-timer">{elapsed}</span>
        <span className="ei-practice-session-clock-separator"> / </span>
        <span className="ei-practice-session-chip"><ClockIcon />{budget}</span>
      </span>
      <button data-testid="practice-topbar-pause" type="button" onClick={onTogglePause} aria-pressed={paused} className="ei-practice-session-button"><span aria-hidden="true">{paused ? "▶" : "❚❚"}</span>{paused ? resumeLabel : pauseLabel}</button>
      <span className="ei-practice-session-divider" aria-hidden="true" />
      <button data-testid="practice-topbar-phone-toggle" type="button" disabled aria-disabled="true" aria-label={phoneDisabledLabel} title={phoneDisabledLabel} className="ei-practice-session-phone"><PhoneIcon size={18} /></button>
      <span className="ei-practice-session-divider" aria-hidden="true" />
      {finishCta}
    </div>
  </div>
);

const RoleIcon: FC = () => (
  <svg aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round"><path d="M8 4h8v4H8z" /><path d="M6 8h12v11H6z" /><path d="M9 12h6" /></svg>
);

const ClockIcon: FC = () => (
  <svg aria-hidden="true" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="8" /><path d="M12 8v4l3 2" /></svg>
);
