ALTER TABLE privacy_requests DROP CONSTRAINT IF EXISTS privacy_requests_user_id_fkey;
ALTER TABLE privacy_requests ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE privacy_requests
  ADD CONSTRAINT privacy_requests_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
