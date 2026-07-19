ALTER TABLE user_settings
  ADD COLUMN theme text NOT NULL DEFAULT 'ocean'
    CHECK (theme IN ('ocean', 'plum')),
  ADD COLUMN custom_accent_hue double precision,
  ADD COLUMN custom_accent_chroma double precision,
  ADD CONSTRAINT user_settings_custom_accent_pair_check CHECK (
    (custom_accent_hue IS NULL AND custom_accent_chroma IS NULL)
    OR
    (custom_accent_hue IS NOT NULL AND custom_accent_chroma IS NOT NULL)
  ),
  ADD CONSTRAINT user_settings_custom_accent_hue_check CHECK (
    custom_accent_hue IS NULL OR (custom_accent_hue >= 0 AND custom_accent_hue < 360)
  ),
  ADD CONSTRAINT user_settings_custom_accent_chroma_check CHECK (
    custom_accent_chroma IS NULL OR (custom_accent_chroma >= 0 AND custom_accent_chroma <= 0.28)
  );
