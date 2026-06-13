import type {
  Debrief,
  DebriefQuestion,
  DebriefRiskItem,
  PracticeSession,
  Resume,
  SuggestedDebriefQuestion,
  TargetJob,
} from "../../../api/generated/types";

export type DebriefStep = 0 | 1 | 2;
export type DebriefInputMode = "text" | "voice";

export type DebriefEntrySource =
  | "ai_confirmed"
  | "ai_edited"
  | "manual"
  | "voice_extracted";

export interface DebriefEntry {
  id: string;
  stage?: string;
  questionText: string;
  myAnswerSummary?: string;
  interviewerReaction?: string;
  reflection?: string;
  reaction?: "positive" | "neutral" | "probed" | "skeptical";
  source: DebriefEntrySource;
  tag?: string;
}

export type DebriefPickerKind = "targetJob" | "mockSession" | "resume";

export interface DebriefSelectedContext {
  targetJob: TargetJob | null;
  mockSession: PracticeSession | null;
  resume: Resume | null;
}

export const EMPTY_SELECTED_CONTEXT: DebriefSelectedContext = {
  targetJob: null,
  mockSession: null,
  resume: null,
};

export type DebriefPollingState =
  | "idle"
  | "running"
  | "succeeded"
  | "failed"
  | "timeout";

export type {
  Debrief,
  DebriefQuestion,
  DebriefRiskItem,
  PracticeSession,
  Resume,
  SuggestedDebriefQuestion,
  TargetJob,
};
