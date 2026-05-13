ALTER TABLE resume_assets
  ADD COLUMN IF NOT EXISTS error_code text;
