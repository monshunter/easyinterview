-- prompt-rubric-registry/002-output-schema-contract L2 remediation seed rollback.
-- Removes the jd_match prompt/rubric baselines added by migration 000010.

BEGIN;

DELETE FROM prompt_versions
 WHERE version = 'v0.1.0'
   AND feature_key IN ('jd_match.recommendation', 'jd_match.search');

DELETE FROM rubric_versions
 WHERE version = 'v0.1.0'
   AND feature_key IN ('jd_match.recommendation', 'jd_match.search');

COMMIT;
