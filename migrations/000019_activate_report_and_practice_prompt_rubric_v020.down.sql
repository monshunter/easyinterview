-- Atomically restore both v0.1 pairs and remove the v0.2 release rows.
BEGIN;

DO $rollback$
BEGIN
  IF (SELECT count(*) FROM prompt_versions WHERE feature_key IN ('report.generate', 'practice.session.chat') AND language = 'multi' AND version = 'v0.1.0') <> 2
    OR (SELECT count(*) FROM rubric_versions WHERE feature_key IN ('report.generate', 'practice.session.chat') AND language = 'multi' AND version = 'v0.1.0') <> 2 THEN
    RAISE EXCEPTION 'rollback invariant: report/practice v0.1 rollback pairs are missing';
  END IF;
END
$rollback$;

UPDATE prompt_versions
SET is_active = (version = 'v0.1.0')
WHERE feature_key IN ('report.generate', 'practice.session.chat')
  AND language = 'multi';

UPDATE rubric_versions
SET is_active = (version = 'v0.1.0')
WHERE feature_key IN ('report.generate', 'practice.session.chat')
  AND language = 'multi';

DELETE FROM prompt_versions
WHERE feature_key IN ('report.generate', 'practice.session.chat')
  AND version = 'v0.2.0'
  AND language = 'multi';

DELETE FROM rubric_versions
WHERE feature_key IN ('report.generate', 'practice.session.chat')
  AND version = 'v0.2.0'
  AND language = 'multi';

DO $rollback$
BEGIN
  IF (SELECT count(*) FROM prompt_versions WHERE feature_key IN ('report.generate', 'practice.session.chat') AND language = 'multi' AND is_active AND version = 'v0.1.0') <> 2
    OR (SELECT count(*) FROM rubric_versions WHERE feature_key IN ('report.generate', 'practice.session.chat') AND language = 'multi' AND is_active AND version = 'v0.1.0') <> 2
    OR EXISTS (SELECT 1 FROM prompt_versions WHERE feature_key IN ('report.generate', 'practice.session.chat') AND language = 'multi' AND version = 'v0.2.0')
    OR EXISTS (SELECT 1 FROM rubric_versions WHERE feature_key IN ('report.generate', 'practice.session.chat') AND language = 'multi' AND version = 'v0.2.0') THEN
    RAISE EXCEPTION 'rollback invariant: report/practice v0.1 pairs were not restored';
  END IF;
END
$rollback$;

COMMIT;
