/**
 * Theme palette data. Every value here is transcribed verbatim from
 * formal frontend implementation (EI_THEMES). The matching focused tests in
 * `tokens.test.ts` assert that every
 * hex appears in primitives.jsx so AI cannot invent values.
 *
 * Do NOT export hex literals from `tokens.ts`; palette data is
 * checked by the focused traceability test, while THEME_METADATA drives the
 * TopBar swatches. `themes.css` is checked-in runtime source, not generated.
 */

import type { Mode, Theme } from "./tokens";

export interface PaletteEntry {
  bg: string;
  bgSoft: string;
  bgCard: string;
  ink: string;
  ink2: string;
  ink3: string;
  ink4: string;
  rule: string;
  ruleSoft: string;
  accent: string;
  accentSoft: string;
  amber: string;
  amberSoft: string;
  ok: string;
  okSoft: string;
  warn: string;
  warnSoft: string;
  danger: string;
  dangerSoft: string;
  cool: string;
  coolSoft: string;
}

export const THEME_PALETTE: Record<Theme, Record<Mode, PaletteEntry>> = {
  ocean: {
    light: {
      bg: "#f8fafd",
      bgSoft: "#eef2f7",
      bgCard: "#ffffff",
      ink: "#141821",
      ink2: "#363c4a",
      ink3: "#6b7280",
      ink4: "#a0a8b3",
      rule: "#dde2ec",
      ruleSoft: "#e7ebf2",
      accent: "#3a5fc4",
      accentSoft: "#dde6f7",
      amber: "#c98730",
      amberSoft: "#f3e1c0",
      ok: "#3f8367",
      okSoft: "#d4ebde",
      warn: "#a87832",
      warnSoft: "#f0e0bf",
      danger: "#b3402b",
      dangerSoft: "#f4d6cc",
      cool: "#4a6670",
      coolSoft: "#d6e1e5",
    },
    dark: {
      bg: "#0c0f17",
      bgSoft: "#13182a",
      bgCard: "#0f1320",
      ink: "#e8edf6",
      ink2: "#c4cad8",
      ink3: "#8389a0",
      ink4: "#5d627a",
      rule: "#212740",
      ruleSoft: "#181d2e",
      accent: "#7493d4",
      accentSoft: "#1c2540",
      amber: "#e6a25a",
      amberSoft: "#322411",
      ok: "#74b08c",
      okSoft: "#1a2c20",
      warn: "#d9a868",
      warnSoft: "#332512",
      danger: "#d4694a",
      dangerSoft: "#361d12",
      cool: "#89a4ae",
      coolSoft: "#1a242a",
    },
  },
  plum: {
    light: {
      bg: "#fcf8fa",
      bgSoft: "#f4ebef",
      bgCard: "#ffffff",
      ink: "#1f161b",
      ink2: "#4a3a43",
      ink3: "#7c6c75",
      ink4: "#a8a0a4",
      rule: "#e9dde2",
      ruleSoft: "#f0e6ea",
      accent: "#9c3a5c",
      accentSoft: "#f4dde6",
      amber: "#c98730",
      amberSoft: "#f3e1c0",
      ok: "#5a7a4a",
      okSoft: "#dde7cd",
      warn: "#a87832",
      warnSoft: "#f0e0bf",
      danger: "#a8452a",
      dangerSoft: "#f3d9cd",
      cool: "#5e6480",
      coolSoft: "#dde0eb",
    },
    dark: {
      bg: "#15101a",
      bgSoft: "#1d1620",
      bgCard: "#171120",
      ink: "#f0e6ed",
      ink2: "#d2c5cd",
      ink3: "#988b94",
      ink4: "#6a5e66",
      rule: "#2c2230",
      ruleSoft: "#211826",
      accent: "#c4709a",
      accentSoft: "#3a1f30",
      amber: "#e6a25a",
      amberSoft: "#3b2a16",
      ok: "#8fae7c",
      okSoft: "#24301a",
      warn: "#d9a868",
      warnSoft: "#362812",
      danger: "#d4694a",
      dangerSoft: "#3a1e14",
      cool: "#9da4c0",
      coolSoft: "#1f2240",
    },
  },
};

export interface ThemeMetadataEntry {
  key: Theme;
  swatch: string;
}

/**
 * Topbar swatches transcribed from EI_THEME_LIST in primitives.jsx. Used by
 * the topbar control to render a per-theme dot indicator without importing
 * the per-mode palette.
 */
export const THEME_METADATA: readonly ThemeMetadataEntry[] = [
  { key: "ocean", swatch: "#3a5fc4" },
  { key: "plum", swatch: "#9c3a5c" },
];
