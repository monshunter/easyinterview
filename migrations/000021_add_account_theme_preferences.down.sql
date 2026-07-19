ALTER TABLE user_settings
  DROP CONSTRAINT user_settings_custom_accent_chroma_check,
  DROP CONSTRAINT user_settings_custom_accent_hue_check,
  DROP CONSTRAINT user_settings_custom_accent_pair_check,
  DROP COLUMN custom_accent_chroma,
  DROP COLUMN custom_accent_hue,
  DROP COLUMN theme;
