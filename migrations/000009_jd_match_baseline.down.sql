ALTER TABLE async_jobs DROP CONSTRAINT IF EXISTS async_jobs_job_type_check;
ALTER TABLE async_jobs ADD CONSTRAINT async_jobs_job_type_check CHECK (job_type IN ('target_import', 'resume_parse', 'report_generate', 'resume_tailor', 'debrief_generate', 'source_refresh', 'privacy_export', 'privacy_delete', 'email_dispatch'));

DROP INDEX IF EXISTS idx_jd_match_search_runs_user_created_at;
DROP TABLE IF EXISTS jd_match_search_runs;

DROP INDEX IF EXISTS idx_agent_scans_user_created_at;
DROP TABLE IF EXISTS agent_scans;

DROP INDEX IF EXISTS idx_saved_searches_user_created_at;
DROP TABLE IF EXISTS saved_searches;

DROP INDEX IF EXISTS idx_watchlist_items_user_added_at;
DROP TABLE IF EXISTS watchlist_items;

DROP INDEX IF EXISTS idx_jd_match_recommendations_user_recommended_at;
DROP INDEX IF EXISTS idx_jd_match_recommendations_user_active;
DROP TABLE IF EXISTS jd_match_recommendations;
