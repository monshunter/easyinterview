BEGIN;
DELETE FROM rubric_versions WHERE version = 'v0.1.0' AND feature_key IN ('practice.session.chat', 'report.generate', 'resume.parse', 'resume.tailor.bullet_suggestions', 'resume.tailor.gap_review', 'target.import.parse');
DELETE FROM prompt_versions WHERE version = 'v0.1.0' AND feature_key IN ('practice.session.chat', 'report.generate', 'resume.parse', 'resume.tailor.bullet_suggestions', 'resume.tailor.gap_review', 'target.import.parse');
COMMIT;
