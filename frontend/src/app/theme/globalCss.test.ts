import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const HERE = resolve(__dirname);
const FRONTEND_ROOT = resolve(HERE, "..", "..", "..");
const REPO_ROOT = resolve(HERE, "..", "..", "..", "..");
const MAIN_TSX = resolve(FRONTEND_ROOT, "src", "main.tsx");
const GLOBAL_CSS = resolve(HERE, "global.css");
const THEMES_CSS = resolve(HERE, "themes.css");
const PACKAGE_JSON = resolve(FRONTEND_ROOT, "package.json");
const PRIMITIVES_PATH = resolve(REPO_ROOT, "ui-design/src/primitives.jsx");

describe("global.css entry (Phase 1.3)", () => {
  it("main.tsx imports the colocated global.css entry once", () => {
    const main = readFileSync(MAIN_TSX, "utf8");
    const matches = main.match(/import\s+["']\.\/app\/theme\/global\.css["'];/g);
    expect(matches?.length, "main.tsx should import ./app/theme/global.css exactly once").toBe(1);
  });

  it("global.css imports themes.css so the palette ships in the same entry", () => {
    const css = readFileSync(GLOBAL_CSS, "utf8");
    expect(css).toMatch(/@import\s+["']\.\/themes\.css["'];/);
  });

  it("global.css transcribes ei-global reset / scrollbar / fadein from primitives.jsx", () => {
    const css = readFileSync(GLOBAL_CSS, "utf8");
    const primitives = readFileSync(PRIMITIVES_PATH, "utf8");
    expect(primitives).toContain("box-sizing: border-box");
    expect(css).toContain("box-sizing: border-box");
    expect(css).toMatch(/body\s*\{[^}]*font-family:\s*var\(--ei-font-sans\)/);
    expect(css).toMatch(/\.ei-serif\s*\{[^}]*font-family:\s*var\(--ei-font-serif\)/);
    expect(css).toMatch(/\.ei-mono\s*\{[^}]*font-family:\s*var\(--ei-font-mono\)/);
    expect(css).toMatch(/@keyframes\s+ei-fadein/);
    expect(css).toMatch(/\.ei-scroll/);
  });

  it("themes.css is also still readable as a standalone module", () => {
    const css = readFileSync(THEMES_CSS, "utf8");
    expect(css.length).toBeGreaterThan(0);
  });
});

describe("frontend dependency boundary (Phase 1.3)", () => {
  const pkg = JSON.parse(readFileSync(PACKAGE_JSON, "utf8"));

  it("does not depend on Tailwind / styled-components / Emotion", () => {
    const merged = {
      ...(pkg.dependencies ?? {}),
      ...(pkg.devDependencies ?? {}),
      ...(pkg.peerDependencies ?? {}),
      ...(pkg.optionalDependencies ?? {}),
    };
    const banned = Object.keys(merged).filter((name) =>
      [
        "tailwindcss",
        "postcss-tailwind",
        "styled-components",
      ].includes(name) || name.startsWith("@emotion/"),
    );
    expect(banned, `forbidden CSS framework deps: ${banned.join(", ")}`).toEqual([]);
  });
});
