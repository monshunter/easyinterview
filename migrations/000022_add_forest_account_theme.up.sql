ALTER TABLE user_settings
  DROP CONSTRAINT user_settings_theme_check,
  ADD CONSTRAINT user_settings_theme_check
    CHECK (theme IN ('ocean', 'plum', 'forest'));
