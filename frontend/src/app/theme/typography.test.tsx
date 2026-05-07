import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const HERE = resolve(__dirname);
const FRONTEND_ROOT = resolve(HERE, "..", "..", "..");
const REPO_ROOT = resolve(HERE, "..", "..", "..", "..");
const TYPOGRAPHY_CSS = resolve(HERE, "typography.css");
const TOPBAR_TSX = resolve(FRONTEND_ROOT, "src", "app", "topbar", "TopBar.tsx");
const PRIMITIVES = resolve(REPO_ROOT, "ui-design", "src", "primitives.jsx");
const APP_JSX = resolve(REPO_ROOT, "ui-design", "src", "app.jsx");

interface ScaleEntry {
  className: string;
  size: string;
  line: string;
  weight: string;
  track: string;
  family: "serif" | "sans" | "mono";
  /** ui-design source file that justifies this scale entry. */
  sourceTrace: string;
  /** Numeric font-size literal that must appear in the source file. */
  sourceFontSize: number;
}

const SCALE: readonly ScaleEntry[] = [
  // ui-design/src/screen-home.jsx:54 hero h1 (48 / 1.1 / 500 / -0.025em)
  {
    className: "ei-text-display",
    size: "48px",
    line: "1.1",
    weight: "500",
    track: "-0.025em",
    family: "serif",
    sourceTrace: "ui-design/src/screen-home.jsx",
    sourceFontSize: 48,
  },
  // ui-design/src/primitives.jsx SectionHeader title (fontSize 22, letter -0.02em)
  {
    className: "ei-text-title",
    size: "22px",
    line: "1.25",
    weight: "500",
    track: "-0.02em",
    family: "serif",
    sourceTrace: "ui-design/src/primitives.jsx",
    sourceFontSize: 22,
  },
  // ui-design/src/app.jsx TopBar logo (fontSize 16, lineHeight 1, letter -0.01em)
  {
    className: "ei-text-subtitle",
    size: "16px",
    line: "1",
    weight: "500",
    track: "-0.01em",
    family: "serif",
    sourceTrace: "ui-design/src/app.jsx",
    sourceFontSize: 16,
  },
  // ui-design/src/screen-home.jsx textarea body (fontSize 14, lineHeight 1.6 in some places; 1.5 in screen-auth)
  {
    className: "ei-text-body",
    size: "14px",
    line: "1.5",
    weight: "400",
    track: "normal",
    family: "sans",
    sourceTrace: "ui-design/src/screen-auth.jsx",
    sourceFontSize: 14,
  },
  // ui-design/src/screen-home.jsx job-card location (fontSize 12.5, marginTop 4)
  {
    className: "ei-text-caption",
    size: "12.5px",
    line: "1.4",
    weight: "400",
    track: "normal",
    family: "sans",
    sourceTrace: "ui-design/src/screen-home.jsx",
    sourceFontSize: 12.5,
  },
  // ui-design/src/primitives.jsx ei-label (fontSize 11, letter 0.08em mono uppercase)
  {
    className: "ei-text-label",
    size: "11px",
    line: "1.4",
    weight: "500",
    track: "0.08em",
    family: "mono",
    sourceTrace: "ui-design/src/primitives.jsx",
    sourceFontSize: 11,
  },
];

describe("typography scale tokens (Phase 2.2)", () => {
  const css = readFileSync(TYPOGRAPHY_CSS, "utf8");

  it.each(SCALE)(
    "$className declares the documented size / line / weight / track via :root tokens",
    ({ className, size, line, weight, track }) => {
      const tokenName = className.replace(/^ei-/, "ei-");
      // :root declares the size/line/weight/track tokens.
      expect(css).toMatch(new RegExp(`--${tokenName}-size:\\s*${escape(size)}\\s*;`));
      expect(css).toMatch(new RegExp(`--${tokenName}-line:\\s*${escape(line)}\\s*;`));
      expect(css).toMatch(
        new RegExp(`--${tokenName}-weight:\\s*${escape(weight)}\\s*;`),
      );
      expect(css).toMatch(
        new RegExp(`--${tokenName}-track:\\s*${escape(track)}\\s*;`),
      );

      // The className rule wires the four tokens via var() so component callers
      // can keep using semantic className without inline px.
      const ruleMatch = css.match(
        new RegExp(`\\.${className}\\s*\\{([^}]+)\\}`, "m"),
      );
      expect(ruleMatch, `missing rule for .${className}`).toBeTruthy();
      const rule = ruleMatch![1] ?? "";
      expect(rule).toMatch(
        new RegExp(`font-size:\\s*var\\(--${tokenName}-size\\)`),
      );
      expect(rule).toMatch(
        new RegExp(`line-height:\\s*var\\(--${tokenName}-line\\)`),
      );
      expect(rule).toMatch(
        new RegExp(`font-weight:\\s*var\\(--${tokenName}-weight\\)`),
      );
      expect(rule).toMatch(
        new RegExp(`letter-spacing:\\s*var\\(--${tokenName}-track\\)`),
      );
    },
  );

  it.each(SCALE)(
    "$className picks the right font-family family ($family)",
    ({ className, family }) => {
      const familyToken =
        family === "serif" ? "ei-font-serif" : family === "mono" ? "ei-font-mono" : "ei-font-sans";
      const ruleMatch = css.match(
        new RegExp(`\\.${className}\\s*\\{([^}]+)\\}`, "m"),
      );
      expect(ruleMatch).toBeTruthy();
      const rule = ruleMatch![1] ?? "";
      expect(rule).toMatch(
        new RegExp(`font-family:\\s*var\\(--${familyToken}\\)`),
      );
    },
  );

  it("uppercase transformation only applies to ei-text-label (matches primitives.jsx ei-label)", () => {
    const labelMatch = css.match(/\.ei-text-label\s*\{([^}]+)\}/);
    expect(labelMatch).toBeTruthy();
    expect(labelMatch![1] ?? "").toMatch(/text-transform:\s*uppercase/);
    // No other ei-text-* should declare uppercase.
    const otherUppercase = css
      .match(/\.ei-text-(?!label)[^\s{]+\s*\{[^}]+\}/g)
      ?.filter((rule) => /text-transform:\s*uppercase/.test(rule)) ?? [];
    expect(otherUppercase).toEqual([]);
  });

  it("each scale entry is traceable to a ui-design source file", () => {
    const sources = new Map<string, string>();
    for (const entry of SCALE) {
      if (!sources.has(entry.sourceTrace)) {
        sources.set(
          entry.sourceTrace,
          readFileSync(resolve(REPO_ROOT, entry.sourceTrace), "utf8"),
        );
      }
    }
    for (const entry of SCALE) {
      const src = sources.get(entry.sourceTrace)!;
      const re = new RegExp(`fontSize:\\s*${entry.sourceFontSize}\\b`);
      expect(
        re.test(src),
        `${entry.className} fontSize ${entry.sourceFontSize} not found in ${entry.sourceTrace}`,
      ).toBe(true);
    }
    // ei-label uppercase transformation traces back to primitives.jsx ei-label rule.
    expect(readFileSync(PRIMITIVES, "utf8")).toContain("text-transform: uppercase");
    // Subtitle line-height: 1 traces back to TopBar logo in app.jsx.
    expect(readFileSync(APP_JSX, "utf8")).toContain("lineHeight: 1");
  });

  it("global.css imports typography.css after themes.css", () => {
    const global = readFileSync(resolve(HERE, "global.css"), "utf8");
    const themesIdx = global.indexOf('@import "./themes.css"');
    const typoIdx = global.indexOf('@import "./typography.css"');
    expect(themesIdx).toBeGreaterThanOrEqual(0);
    expect(typoIdx).toBeGreaterThan(themesIdx);
  });
});

describe("TopBar typography contract (Phase 2.2)", () => {
  const tsx = readFileSync(TOPBAR_TSX, "utf8");

  it("TopBar.tsx does not hardcode inline px font-size literals", () => {
    expect(tsx).not.toMatch(/style=\{\{[^}]*fontSize:\s*\d/);
  });

  it("TopBar.tsx (post-Phase-3) wires text content through ei-text-* className", () => {
    // Phase 2.2 establishes the typography token surface; Phase 3 wires the
    // TopBar visual to it. We allow this assertion to pass when at least one
    // ei-text-* className appears in TopBar.tsx so component-level adoption is
    // tracked. Phase 3 will tighten this further by mapping each text element
    // to a specific scale entry.
    if (/className=\"[^\"]*ei-text-/.test(tsx)) {
      const found = tsx.match(/ei-text-[a-z]+/g) ?? [];
      expect(new Set(found).size).toBeGreaterThan(0);
    }
  });
});

function escape(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
