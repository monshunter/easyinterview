ALTER TABLE privacy_requests DROP CONSTRAINT IF EXISTS privacy_requests_user_id_fkey;
DELETE FROM privacy_requests WHERE user_id IS NULL;
ALTER TABLE privacy_requests ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE privacy_requests
  ADD CONSTRAINT privacy_requests_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
