CREATE UNIQUE INDEX IF NOT EXISTS idx_practice_sessions_one_active_per_plan
  ON practice_sessions (user_id, plan_id)
  WHERE status IN ('queued', 'running');
