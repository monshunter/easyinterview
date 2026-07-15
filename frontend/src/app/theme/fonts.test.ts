import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const HERE = resolve(__dirname);
const FRONTEND_ROOT = resolve(HERE, "..", "..", "..");
const PACKAGE_JSON = resolve(FRONTEND_ROOT, "package.json");
const FONTS_CSS = resolve(HERE, "fonts.css");
const GLOBAL_CSS = resolve(HERE, "global.css");
const THEMES_CSS = resolve(HERE, "themes.css");
const NOTO_SERIF_SC_ROOT = resolve(
  FRONTEND_ROOT,
  "node_modules",
  "@fontsource",
  "noto-serif-sc",
);

const REQUIRED_FONTSOURCE_PACKAGES = [
  "@fontsource/noto-serif-sc",
  "@fontsource/inter",
  "@fontsource/jetbrains-mono",
];

const REMOVED_FONTSOURCE_PACKAGES = [
  "@fontsource/source-serif-pro",
  "@fontsource/cormorant-garamond",
  "@fontsource/ibm-plex-sans",
  "@fontsource/geist-sans",
];

const LATIN_ONLY_IMPORTS = {
  "@fontsource/inter": [400, 500, 600],
  "@fontsource/jetbrains-mono": [400, 500],
} as const;

describe("open-source font sourcing (Phase 2.1)", () => {
  const pkg = JSON.parse(readFileSync(PACKAGE_JSON, "utf8"));
  const merged = {
    ...(pkg.dependencies ?? {}),
    ...(pkg.devDependencies ?? {}),
  };

  it("declares the single application typography package set", () => {
    for (const name of REQUIRED_FONTSOURCE_PACKAGES) {
      expect(
        merged[name],
        `missing required fontsource dep ${name}`,
      ).toBeTruthy();
    }
  });

  it("does not ship removed font-preset packages or imports", () => {
    const fonts = readFileSync(FONTS_CSS, "utf8");
    for (const name of REMOVED_FONTSOURCE_PACKAGES) {
      expect(merged[name]).toBeUndefined();
      expect(fonts).not.toContain(name);
    }
  });

  it("does not depend on the out-of-scope private brand font names", () => {
    const raw = readFileSync(PACKAGE_JSON, "utf8").toLowerCase();
    expect(raw).not.toContain("copernicus");
    expect(raw).not.toContain("styreneb");
    expect(raw).not.toContain("@fontsource/copernicus");
  });

  it("ships a fonts.css module that imports every required fontsource bundle", () => {
    const fonts = readFileSync(FONTS_CSS, "utf8");
    for (const name of REQUIRED_FONTSOURCE_PACKAGES) {
      const subpath = name.replace("@fontsource/", "");
      const re = new RegExp(`@import\\s+["']@fontsource/${subpath}(/.+)?["'];`);
      expect(
        re.test(fonts),
        `fonts.css must @import ${name}`,
      ).toBe(true);
    }
  });

  it("uses Noto Serif SC unicode-range bundles without duplicate full-font imports", () => {
    const fonts = readFileSync(FONTS_CSS, "utf8");

    for (const weight of ["400", "500"]) {
      expect(fonts).toContain(`@fontsource/noto-serif-sc/${weight}.css`);
      expect(
        readFileSync(resolve(NOTO_SERIF_SC_ROOT, `${weight}.css`), "utf8"),
      ).toMatch(/unicode-range:[^;]*U\+4e00/i);
      expect(fonts).not.toContain(
        `@fontsource/noto-serif-sc/chinese-simplified-${weight}.css`,
      );
    }
  });

  it("limits Western preset fonts to the product's Latin locale subset", () => {
    const fonts = readFileSync(FONTS_CSS, "utf8");

    for (const [packageName, weights] of Object.entries(LATIN_ONLY_IMPORTS)) {
      for (const weight of weights) {
        const latinImport = `${packageName}/latin-${weight}.css`;
        expect(fonts).toContain(latinImport);
        expect(fonts).not.toContain(`${packageName}/${weight}.css`);

        const subsetCss = readFileSync(
          resolve(FRONTEND_ROOT, "node_modules", packageName, `latin-${weight}.css`),
          "utf8",
        );
        expect(subsetCss).toContain(`font-weight: ${weight}`);
      }
    }
  });

  it("global.css imports fonts.css so the production entry pulls fontsource", () => {
    const css = readFileSync(GLOBAL_CSS, "utf8");
    expect(css).toMatch(/@import\s+["']\.\/fonts\.css["'];/);
  });

  it("themes.css font-family chains start with open-source fonts and include system fallback", () => {
    const css = readFileSync(THEMES_CSS, "utf8");
    expect(css).toMatch(
      /--ei-font-serif:\s*"Noto Serif SC"[^;]*Georgia[^;]*serif;/,
    );
    expect(css).toMatch(
      /--ei-font-sans:\s*"Inter"[^;]*sans-serif;/,
    );
    expect(css).toMatch(
      /--ei-font-mono:\s*"JetBrains Mono"[^;]*monospace;/,
    );
  });
});
