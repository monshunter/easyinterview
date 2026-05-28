ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_completed_at timestamptz;
ALTER TABLE users ADD COLUMN IF NOT EXISTS terms_accepted_at timestamptz;

UPDATE users
SET
  profile_completed_at = COALESCE(profile_completed_at, created_at),
  terms_accepted_at = COALESCE(terms_accepted_at, created_at),
  updated_at = now()
WHERE deleted_at IS NULL
  AND btrim(COALESCE(display_name, '')) <> ''
  AND (profile_completed_at IS NULL OR terms_accepted_at IS NULL);
