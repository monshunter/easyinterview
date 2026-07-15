import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

import {
  COLOR_TOKENS,
  RADIUS_TOKENS,
  SHADOW_TOKENS,
  SPACING_TOKENS,
  TYPOGRAPHY_TOKENS,
  THEME_KEYS,
  MODE_KEYS,
  cssVar,
} from "./tokens";
import { THEME_PALETTE } from "./themes.data";
import { computeCustomAccentOverrides } from "./customAccent";

const HERE = resolve(__dirname);
const TOKENS_TS_PATH = resolve(HERE, "tokens.ts");
const THEMES_CSS_PATH = resolve(HERE, "themes.css");

const HEX_LITERAL = /#[0-9a-fA-F]{3,8}\b/g;

describe("design token module (Phase 1.1)", () => {
  it("exports semantic key constants without hex literals in tokens.ts source", () => {
    const source = readFileSync(TOKENS_TS_PATH, "utf8");
    const matches = source.match(HEX_LITERAL) ?? [];
    expect(
      matches,
      `tokens.ts must export semantic keys only; found hex literals: ${matches.join(", ")}`,
    ).toEqual([]);
  });

  it("exposes the canonical semantic key namespaces", () => {
    expect(COLOR_TOKENS.bg).toMatchObject({
      canvas: "--ei-color-bg-canvas",
      soft: "--ei-color-bg-soft",
      card: "--ei-color-bg-card",
    });
    expect(COLOR_TOKENS.fg).toMatchObject({
      primary: "--ei-color-fg-primary",
      secondary: "--ei-color-fg-secondary",
      tertiary: "--ei-color-fg-tertiary",
      muted: "--ei-color-fg-muted",
    });
    expect(COLOR_TOKENS.rule).toMatchObject({
      strong: "--ei-color-rule-strong",
      soft: "--ei-color-rule-soft",
    });
    expect(COLOR_TOKENS.accent).toMatchObject({
      base: "--ei-color-accent",
      soft: "--ei-color-accent-soft",
    });
    for (const tone of ["amber", "ok", "warn", "danger", "cool"] as const) {
      expect(COLOR_TOKENS[tone]).toMatchObject({
        base: `--ei-color-${tone}`,
        soft: `--ei-color-${tone}-soft`,
      });
    }

    expect(RADIUS_TOKENS).toMatchObject({
      sm: "--ei-radius-sm",
      md: "--ei-radius-md",
      pill: "--ei-radius-pill",
    });

    expect(SHADOW_TOKENS).toMatchObject({
      elev1: "--ei-shadow-elev1",
      elev2: "--ei-shadow-elev2",
    });

    expect(SPACING_TOKENS).toMatchObject({
      "1": "--ei-space-1",
      "2": "--ei-space-2",
      "3": "--ei-space-3",
      "4": "--ei-space-4",
    });

    expect(TYPOGRAPHY_TOKENS.family).toMatchObject({
      serif: "--ei-font-serif",
      sans: "--ei-font-sans",
      mono: "--ei-font-mono",
    });

    expect(THEME_KEYS).toEqual(["ocean", "plum"]);
    expect(MODE_KEYS).toEqual(["light", "dark"]);
  });

  it("provides cssVar helper that returns var(--token) syntax", () => {
    expect(cssVar(COLOR_TOKENS.bg.canvas)).toBe("var(--ei-color-bg-canvas)");
    expect(cssVar(COLOR_TOKENS.accent.soft)).toBe(
      "var(--ei-color-accent-soft)",
    );
  });
});

describe("theme palette data (Phase 1.1)", () => {
  it("defines complete ocean and plum palettes for both display modes", () => {
    expect(Object.keys(THEME_PALETTE)).toEqual([
      "ocean",
      "plum",
    ]);
    for (const theme of THEME_KEYS) {
      for (const mode of MODE_KEYS) {
        const palette = THEME_PALETTE[theme][mode];
        expect(palette.bg, `${theme}/${mode} missing bg`).toBeTruthy();
        expect(palette.ink, `${theme}/${mode} missing ink`).toBeTruthy();
        expect(palette.accent, `${theme}/${mode} missing accent`).toBeTruthy();
        for (const [key, value] of Object.entries(palette)) {
          expect(value, `${theme}/${mode}.${key} must be a hex color`).toMatch(
            /^#[0-9a-f]{6}$/i,
          );
        }
      }
    }
  });

});

describe("themes.css CSS variable wiring (Phase 1.1)", () => {
  const css = readFileSync(THEMES_CSS_PATH, "utf8");
  const colorKeys = [
    "bg-canvas",
    "bg-soft",
    "bg-card",
    "fg-primary",
    "fg-secondary",
    "fg-tertiary",
    "fg-muted",
    "rule-strong",
    "rule-soft",
    "accent",
    "accent-soft",
    "amber",
    "amber-soft",
    "ok",
    "ok-soft",
    "warn",
    "warn-soft",
    "danger",
    "danger-soft",
    "cool",
    "cool-soft",
  ] as const;

  it("defines every base color variable on all 4 theme-mode selectors", () => {
    for (const theme of THEME_KEYS) {
      for (const mode of MODE_KEYS) {
        const selector = `:root[data-theme="${theme}"][data-mode="${mode}"]`;
        const blockMatch = css.match(
          new RegExp(
            `:root\\[data-theme="${theme}"\\]\\[data-mode="${mode}"\\]\\s*\\{([^}]*)\\}`,
            "m",
          ),
        );
        expect(blockMatch, `missing CSS block for ${selector}`).toBeTruthy();
        const block = blockMatch![1] ?? "";
        for (const key of colorKeys) {
          const decl = new RegExp(`--ei-color-${key}\\s*:\\s*[^;]+;`);
          expect(
            decl.test(block),
            `${selector} is missing --ei-color-${key}`,
          ).toBe(true);
        }
      }
    }
  });

  it("defines radius / shadow / spacing / typography tokens at :root", () => {
    expect(/:root\s*\{[^}]*--ei-radius-sm:\s*2px/.test(css)).toBe(true);
    expect(/:root\s*\{[^}]*--ei-radius-md:\s*3px/.test(css)).toBe(true);
    expect(/:root\s*\{[^}]*--ei-radius-pill:\s*999px/.test(css)).toBe(true);
    expect(/:root\s*\{[^}]*--ei-shadow-elev1:[^;]+/.test(css)).toBe(true);
    expect(/:root\s*\{[^}]*--ei-shadow-elev2:[^;]+/.test(css)).toBe(true);
    expect(/:root\s*\{[^}]*--ei-space-1:\s*4px/.test(css)).toBe(true);
    expect(/:root\s*\{[^}]*--ei-space-2:\s*8px/.test(css)).toBe(true);
    expect(/:root\s*\{[^}]*--ei-font-serif:/.test(css)).toBe(true);
    expect(/:root\s*\{[^}]*--ei-font-sans:/.test(css)).toBe(true);
    expect(/:root\s*\{[^}]*--ei-font-mono:/.test(css)).toBe(true);
  });

  it("matches palette data for every theme/mode/key", () => {
    for (const theme of THEME_KEYS) {
      for (const mode of MODE_KEYS) {
        const match = css.match(
          new RegExp(
            `:root\\[data-theme="${theme}"\\]\\[data-mode="${mode}"\\]\\s*\\{([^}]*)\\}`,
            "m",
          ),
        );
        expect(match?.[1], `missing CSS block for ${theme}/${mode}`).toBeTruthy();
        const block = match![1] ?? "";
        const palette = THEME_PALETTE[theme][mode];
        const aliasToCss: Record<string, string> = {
          bg: "bg-canvas",
          bgSoft: "bg-soft",
          bgCard: "bg-card",
          ink: "fg-primary",
          ink2: "fg-secondary",
          ink3: "fg-tertiary",
          ink4: "fg-muted",
          rule: "rule-strong",
          ruleSoft: "rule-soft",
          accent: "accent",
          accentSoft: "accent-soft",
          amber: "amber",
          amberSoft: "amber-soft",
          ok: "ok",
          okSoft: "ok-soft",
          warn: "warn",
          warnSoft: "warn-soft",
          danger: "danger",
          dangerSoft: "danger-soft",
          cool: "cool",
          coolSoft: "cool-soft",
        };
        for (const [aliasKey, value] of Object.entries(palette)) {
          const cssKey = aliasToCss[aliasKey];
          expect(cssKey, `unknown palette alias ${aliasKey}`).toBeTruthy();
          const decl = new RegExp(
            `--ei-color-${cssKey}\\s*:\\s*${escape(value)}\\s*;`,
          );
          expect(
            decl.test(block),
            `${theme}/${mode} CSS missing --ei-color-${cssKey}: ${value}`,
          ).toBe(true);
        }
      }
    }
  });
});

describe("customAccent helper (Phase 1.1)", () => {
  it("only overrides accent / accent-soft variables, not the rest of the palette", () => {
    const overrides = computeCustomAccentOverrides({
      h: 30,
      c: 0.16,
      dark: false,
    });
    expect(overrides).not.toBeNull();
    const keys = Object.keys(overrides!);
    expect(keys.sort()).toEqual([
      "--ei-color-accent",
      "--ei-color-accent-soft",
    ]);
    expect(overrides!["--ei-color-accent"]).toMatch(/^oklch\(/);
    expect(overrides!["--ei-color-accent-soft"]).toMatch(/^oklch\(/);
  });

  it("returns null overrides when accent input is null", () => {
    expect(computeCustomAccentOverrides(null)).toBeNull();
  });

  it("uses the current oklch formula (light=58 / dark=68 lightness)", () => {
    const light = computeCustomAccentOverrides({ h: 30, c: 0.16, dark: false })!;
    const dark = computeCustomAccentOverrides({ h: 30, c: 0.16, dark: true })!;
    expect(light["--ei-color-accent"]).toContain("58%");
    expect(dark["--ei-color-accent"]).toContain("68%");
    // Soft lightness: light=92, dark=28.
    expect(light["--ei-color-accent-soft"]).toContain("92%");
    expect(dark["--ei-color-accent-soft"]).toContain("28%");
  });

  it("clamps chroma to the supported [0, 0.28] range", () => {
    const big = computeCustomAccentOverrides({ h: 30, c: 5, dark: false })!;
    expect(big["--ei-color-accent"]).toContain("0.280");
    const tiny = computeCustomAccentOverrides({ h: 30, c: -1, dark: false })!;
    expect(tiny["--ei-color-accent"]).toContain("0.000");
  });

  it("normalizes hue into [0, 360)", () => {
    const wrap = computeCustomAccentOverrides({ h: 720, c: 0.1, dark: false })!;
    expect(wrap["--ei-color-accent"]).toMatch(/0(\.0)?\)/);
    const negative = computeCustomAccentOverrides({
      h: -45,
      c: 0.1,
      dark: false,
    })!;
    expect(negative["--ei-color-accent"]).toMatch(/315(\.0)?\)/);
  });
});

function escape(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
