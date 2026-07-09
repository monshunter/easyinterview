/**
 * D2 design token registry. The TS surface only exposes semantic CSS variable
 * names so callers cannot import raw color literals; all base palette values
 * live in `themes.data.ts` (private to this directory) and are projected onto
 * `:root[data-theme][data-mode]` selectors in `themes.css`.
 *
 * Sources of truth:
 *   - ui-design/src/primitives.jsx (EI_THEMES, EI_FONT_PRESETS, ei-global vars)
 *   - ui-design/src/app.jsx (customAccent oklch formula)
 *
 * Adding a new token requires updating: this registry, themes.data.ts (when
 * it is a per-theme color), themes.css (variable declaration on every base
 * combination), and the matching focused tests in `tokens.test.ts`.
 */

export type Theme = "ocean" | "plum";
export type Mode = "light" | "dark";

export const THEME_KEYS: readonly Theme[] = ["ocean", "plum"];
export const MODE_KEYS: readonly Mode[] = ["light", "dark"];

export const COLOR_TOKENS = {
  bg: {
    canvas: "--ei-color-bg-canvas",
    soft: "--ei-color-bg-soft",
    card: "--ei-color-bg-card",
  },
  fg: {
    primary: "--ei-color-fg-primary",
    secondary: "--ei-color-fg-secondary",
    tertiary: "--ei-color-fg-tertiary",
    muted: "--ei-color-fg-muted",
  },
  rule: {
    strong: "--ei-color-rule-strong",
    soft: "--ei-color-rule-soft",
  },
  accent: {
    base: "--ei-color-accent",
    soft: "--ei-color-accent-soft",
  },
  amber: { base: "--ei-color-amber", soft: "--ei-color-amber-soft" },
  ok: { base: "--ei-color-ok", soft: "--ei-color-ok-soft" },
  warn: { base: "--ei-color-warn", soft: "--ei-color-warn-soft" },
  danger: { base: "--ei-color-danger", soft: "--ei-color-danger-soft" },
  cool: { base: "--ei-color-cool", soft: "--ei-color-cool-soft" },
} as const;

export const RADIUS_TOKENS = {
  sm: "--ei-radius-sm",
  md: "--ei-radius-md",
  pill: "--ei-radius-pill",
} as const;

export const SHADOW_TOKENS = {
  elev1: "--ei-shadow-elev1",
  elev2: "--ei-shadow-elev2",
} as const;

export const SPACING_TOKENS = {
  "1": "--ei-space-1",
  "2": "--ei-space-2",
  "3": "--ei-space-3",
  "4": "--ei-space-4",
  "5": "--ei-space-5",
  "6": "--ei-space-6",
  "7": "--ei-space-7",
  "8": "--ei-space-8",
} as const;

export const TYPOGRAPHY_TOKENS = {
  family: {
    serif: "--ei-font-serif",
    sans: "--ei-font-sans",
    mono: "--ei-font-mono",
  },
  display: {
    size: "--ei-text-display-size",
    line: "--ei-text-display-line",
    weight: "--ei-text-display-weight",
    track: "--ei-text-display-track",
  },
  title: {
    size: "--ei-text-title-size",
    line: "--ei-text-title-line",
    weight: "--ei-text-title-weight",
    track: "--ei-text-title-track",
  },
  body: {
    size: "--ei-text-body-size",
    line: "--ei-text-body-line",
    weight: "--ei-text-body-weight",
    track: "--ei-text-body-track",
  },
  caption: {
    size: "--ei-text-caption-size",
    line: "--ei-text-caption-line",
    weight: "--ei-text-caption-weight",
    track: "--ei-text-caption-track",
  },
  label: {
    size: "--ei-text-label-size",
    line: "--ei-text-label-line",
    weight: "--ei-text-label-weight",
    track: "--ei-text-label-track",
  },
} as const;

export const ROOT_THEME_ATTR = "data-theme";
export const ROOT_MODE_ATTR = "data-mode";
export const ROOT_CUSTOM_ACCENT_ATTR = "data-custom-accent";

export function cssVar(token: string): string {
  return `var(${token})`;
}
