-- product-scope v2.1 D-17 (2026-06-12): the JD-Match / job recommendation
-- module is removed from the product. Drop the five module tables created by
-- 000009 and the jd_match.* prompt / rubric registry rows seeded by 000010.
-- 000009 / 000010 stay in history; this migration is the deletion record.

DROP TABLE IF EXISTS jd_match_search_runs;
DROP TABLE IF EXISTS agent_scans;
DROP TABLE IF EXISTS saved_searches;
DROP TABLE IF EXISTS watchlist_items;
DROP TABLE IF EXISTS jd_match_recommendations;

DELETE FROM rubric_versions WHERE feature_key IN ('jd_match.recommendation', 'jd_match.search');
DELETE FROM prompt_versions WHERE feature_key IN ('jd_match.recommendation', 'jd_match.search');

DELETE FROM async_jobs WHERE job_type IN ('jd_match_agent_scan', 'jd_match_search');

ALTER TABLE async_jobs DROP CONSTRAINT IF EXISTS async_jobs_job_type_check;
ALTER TABLE async_jobs ADD CONSTRAINT async_jobs_job_type_check CHECK (job_type IN ('target_import', 'resume_parse', 'report_generate', 'resume_tailor', 'debrief_generate', 'source_refresh', 'privacy_export', 'privacy_delete', 'email_dispatch'));
