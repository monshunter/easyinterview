/**
 * Theme palette data. Every value here is transcribed verbatim from
 * ui-design/src/primitives.jsx (EI_THEMES) and ui-design/src/app.jsx (font
 * presets). The matching focused tests in `tokens.test.ts` assert that every
 * hex appears in primitives.jsx so AI cannot invent values.
 *
 * Do NOT export hex literals from `tokens.ts`; this module is consumed only
 * by the colocated `themes.css` generator and the focused traceability test.
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
  warm: {
    light: {
      bg: "#fdfcf8",
      bgSoft: "#f7f3ea",
      bgCard: "#ffffff",
      ink: "#1c1917",
      ink2: "#44403c",
      ink3: "#78716c",
      ink4: "#a8a29e",
      rule: "#e7e2d6",
      ruleSoft: "#efeadc",
      accent: "#c96442",
      accentSoft: "#fbe8dc",
      amber: "#d9893a",
      amberSoft: "#fbe9ce",
      ok: "#5a7a4a",
      okSoft: "#e7efd9",
      warn: "#b8813a",
      warnSoft: "#f6ead0",
      danger: "#a8452a",
      dangerSoft: "#f6dcd0",
      cool: "#4a6670",
      coolSoft: "#dce6ea",
    },
    dark: {
      bg: "#16130e",
      bgSoft: "#1f1b15",
      bgCard: "#1a1611",
      ink: "#f5f0e4",
      ink2: "#d6cdb8",
      ink3: "#968d7a",
      ink4: "#6b6455",
      rule: "#2d2820",
      ruleSoft: "#24201a",
      accent: "#e08061",
      accentSoft: "#3a2318",
      amber: "#e6a25a",
      amberSoft: "#3b2a16",
      ok: "#8fae7c",
      okSoft: "#24301a",
      warn: "#d9a868",
      warnSoft: "#362812",
      danger: "#d4694a",
      dangerSoft: "#3a1e14",
      cool: "#89a4ae",
      coolSoft: "#1c2830",
    },
  },
  forest: {
    light: {
      bg: "#f9faf3",
      bgSoft: "#eef2e3",
      bgCard: "#ffffff",
      ink: "#181d14",
      ink2: "#3c4434",
      ink3: "#6f7565",
      ink4: "#a3a895",
      rule: "#dde3ce",
      ruleSoft: "#e8ecdc",
      accent: "#5a7d3a",
      accentSoft: "#dfe9c9",
      amber: "#b8813a",
      amberSoft: "#f3e4c4",
      ok: "#5a7a4a",
      okSoft: "#dde7cd",
      warn: "#a87832",
      warnSoft: "#f0e0bf",
      danger: "#a8452a",
      dangerSoft: "#f3d9cd",
      cool: "#4a6670",
      coolSoft: "#d6e1e5",
    },
    dark: {
      bg: "#0e120a",
      bgSoft: "#161b10",
      bgCard: "#11160c",
      ink: "#eaeed8",
      ink2: "#cbcfb0",
      ink3: "#888c70",
      ink4: "#5e6250",
      rule: "#252a1c",
      ruleSoft: "#1c2014",
      accent: "#8fae60",
      accentSoft: "#1f2a14",
      amber: "#d9a868",
      amberSoft: "#352712",
      ok: "#9ab78a",
      okSoft: "#212c16",
      warn: "#cda06a",
      warnSoft: "#332512",
      danger: "#d4694a",
      dangerSoft: "#361d12",
      cool: "#89a4ae",
      coolSoft: "#1a242a",
    },
  },
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

export interface FontPresetEntry {
  key: "editorial" | "modern" | "magazine";
  serif: string;
  sans: string;
}

export const FONT_PRESETS: readonly FontPresetEntry[] = [
  { key: "editorial", serif: "Noto Serif SC", sans: "Inter" },
  { key: "modern", serif: "Source Serif Pro", sans: "Geist" },
  { key: "magazine", serif: "Cormorant Garamond", sans: "IBM Plex Sans" },
];

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
  { key: "warm", swatch: "#c96442" },
  { key: "forest", swatch: "#5a7d3a" },
  { key: "ocean", swatch: "#3a5fc4" },
  { key: "plum", swatch: "#9c3a5c" },
];
