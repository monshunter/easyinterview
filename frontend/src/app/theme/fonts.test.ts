import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const HERE = resolve(__dirname);
const FRONTEND_ROOT = resolve(HERE, "..", "..", "..");
const PACKAGE_JSON = resolve(FRONTEND_ROOT, "package.json");
const FONTS_CSS = resolve(HERE, "fonts.css");
const GLOBAL_CSS = resolve(HERE, "global.css");
const THEMES_CSS = resolve(HERE, "themes.css");

const REQUIRED_FONTSOURCE_PACKAGES = [
  "@fontsource/noto-serif-sc",
  "@fontsource/inter",
  "@fontsource/source-serif-pro",
  "@fontsource/cormorant-garamond",
  "@fontsource/ibm-plex-sans",
  "@fontsource/jetbrains-mono",
  "@fontsource/geist-sans",
];

describe("open-source font sourcing (Phase 2.1)", () => {
  const pkg = JSON.parse(readFileSync(PACKAGE_JSON, "utf8"));
  const merged = {
    ...(pkg.dependencies ?? {}),
    ...(pkg.devDependencies ?? {}),
  };

  it("declares fontsource packages for every EI_FONT_PRESETS entry", () => {
    for (const name of REQUIRED_FONTSOURCE_PACKAGES) {
      expect(
        merged[name],
        `missing required fontsource dep ${name}`,
      ).toBeTruthy();
    }
  });

  it("does not depend on the non-current private brand font names", () => {
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
