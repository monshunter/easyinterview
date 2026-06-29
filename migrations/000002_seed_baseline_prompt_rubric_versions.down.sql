-- F3 prompt-rubric-registry/001-baseline phase 4.5 seed rollback.
-- Reverses the seed inserts for the 9 baseline feature_keys.
BEGIN;

DELETE FROM prompt_versions
 WHERE version = 'v0.1.0' AND feature_key IN ('practice.session.first_question', 'practice.session.follow_up', 'practice.turn.lightweight_observe', 'report.generate', 'report.question_assessment', 'resume.parse', 'resume.tailor.bullet_suggestions', 'resume.tailor.gap_review', 'target.import.parse');

DELETE FROM rubric_versions
 WHERE version = 'v0.1.0' AND feature_key IN ('practice.session.first_question', 'practice.session.follow_up', 'practice.turn.lightweight_observe', 'report.generate', 'report.question_assessment', 'resume.parse', 'resume.tailor.bullet_suggestions', 'resume.tailor.gap_review', 'target.import.parse');

COMMIT;
