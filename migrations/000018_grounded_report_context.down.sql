ALTER TABLE practice_plans
  RENAME COLUMN focus_dimension_codes TO focus_competency_codes;

ALTER TABLE feedback_reports
  RENAME COLUMN retry_focus_dimension_codes TO retry_focus_competency_codes;

ALTER TABLE feedback_reports
  DROP COLUMN IF EXISTS generation_context,
  DROP COLUMN IF EXISTS summary;
