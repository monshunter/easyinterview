ALTER TABLE user_settings
  ADD COLUMN ui_language text NOT NULL DEFAULT 'zh-CN',
  ADD COLUMN preferred_practice_language text NOT NULL DEFAULT 'en',
  ADD COLUMN region text,
  ADD COLUMN timezone text NOT NULL DEFAULT 'UTC';
