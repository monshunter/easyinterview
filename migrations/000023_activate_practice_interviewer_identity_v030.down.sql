-- Atomically restore practice v0.2 and remove the v0.3 release rows.
BEGIN;

DO $rollback$
BEGIN
  IF (SELECT count(*) FROM prompt_versions WHERE feature_key = 'practice.session.chat' AND language = 'multi' AND version = 'v0.2.0') <> 1
    OR (SELECT count(*) FROM rubric_versions WHERE feature_key = 'practice.session.chat' AND language = 'multi' AND version = 'v0.2.0') <> 1 THEN
    RAISE EXCEPTION 'rollback invariant: practice v0.2 rollback pair is missing';
  END IF;
END
$rollback$;

UPDATE prompt_versions
SET is_active = (version = 'v0.2.0')
WHERE feature_key IN ('practice.session.chat')
  AND language = 'multi';

UPDATE rubric_versions
SET is_active = (version = 'v0.2.0')
WHERE feature_key IN ('practice.session.chat')
  AND language = 'multi';

DELETE FROM prompt_versions
WHERE feature_key = 'practice.session.chat'
  AND version = 'v0.3.0'
  AND language = 'multi';

DELETE FROM rubric_versions
WHERE feature_key = 'practice.session.chat'
  AND version = 'v0.3.0'
  AND language = 'multi';

DO $rollback$
BEGIN
  IF (SELECT count(*) FROM prompt_versions WHERE feature_key = 'practice.session.chat' AND language = 'multi' AND is_active AND version = 'v0.2.0') <> 1
    OR (SELECT count(*) FROM rubric_versions WHERE feature_key = 'practice.session.chat' AND language = 'multi' AND is_active AND version = 'v0.2.0') <> 1
    OR EXISTS (SELECT 1 FROM prompt_versions WHERE feature_key = 'practice.session.chat' AND language = 'multi' AND version = 'v0.3.0')
    OR EXISTS (SELECT 1 FROM rubric_versions WHERE feature_key = 'practice.session.chat' AND language = 'multi' AND version = 'v0.3.0') THEN
    RAISE EXCEPTION 'rollback invariant: practice v0.2 pair was not restored';
  END IF;
END
$rollback$;

COMMIT;
