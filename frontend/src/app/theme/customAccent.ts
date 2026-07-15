/**
 * Custom accent override helper. Mirrors the `customAccent` formula in
 * `formal frontend implementation`:
 *
 *   accentL = isDark ? 68 : 58
 *   softL   = isDark ? 28 : 92
 *   softC   = isDark ? min(c * 0.55, 0.10) : min(c * 0.22, 0.05)
 *   accent      = oklch(accentL% c h)
 *   accentSoft  = oklch(softL% softC h)
 *
 * Only `--ei-color-accent` and `--ei-color-accent-soft` are overridden. The
 * rest of the EI_THEMES base palette is left untouched, which matches the
 * static prototype's intent that custom accent is an overlay, not a full
 * theme.
 */

import { COLOR_TOKENS } from "./tokens";

export interface CustomAccentInput {
  /** Hue in degrees; values outside [0, 360) are normalized. */
  h: number;
  /** Chroma in [0, 0.28]; values outside the range are clamped. */
  c: number;
  /** Whether the active mode is dark. Drives the lightness pair. */
  dark: boolean;
}

export type CustomAccentOverrides = {
  [COLOR_TOKENS.accent.base]: string;
  [COLOR_TOKENS.accent.soft]: string;
};

export function computeCustomAccentOverrides(
  input: CustomAccentInput | null | undefined,
): CustomAccentOverrides | null {
  if (!input) return null;
  const { h, c, dark } = input;
  if (
    !Number.isFinite(h) ||
    !Number.isFinite(c) ||
    typeof dark !== "boolean"
  ) {
    return null;
  }
  const hue = ((h % 360) + 360) % 360;
  const chroma = Math.max(0, Math.min(0.28, c));
  const accentL = dark ? 68 : 58;
  const softL = dark ? 28 : 92;
  const softC = dark
    ? Math.min(chroma * 0.55, 0.1)
    : Math.min(chroma * 0.22, 0.05);
  return {
    [COLOR_TOKENS.accent.base]: `oklch(${accentL}% ${chroma.toFixed(3)} ${hue.toFixed(1)})`,
    [COLOR_TOKENS.accent.soft]: `oklch(${softL}% ${softC.toFixed(3)} ${hue.toFixed(1)})`,
  };
}
