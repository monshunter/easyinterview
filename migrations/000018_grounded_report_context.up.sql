ALTER TABLE feedback_reports
  ADD COLUMN summary text,
  ADD COLUMN generation_context jsonb NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE feedback_reports
  RENAME COLUMN retry_focus_competency_codes TO retry_focus_dimension_codes;

ALTER TABLE practice_plans
  RENAME COLUMN focus_competency_codes TO focus_dimension_codes;
