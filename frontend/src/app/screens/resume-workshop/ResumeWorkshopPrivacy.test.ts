import { readdirSync, readFileSync, statSync } from "node:fs";
import { extname, join, resolve } from "node:path";

import { describe, expect, it } from "vitest";

const ROOT = resolve(__dirname);

function* walk(dir: string): Generator<string> {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      yield* walk(full);
      continue;
    }
    const ext = extname(entry);
    if (ext === ".ts" || ext === ".tsx") yield full;
  }
}

const isTestFile = (file: string): boolean =>
  /\.test\.(ts|tsx)$/.test(file);

const SENSITIVE_PII_FIELDS = [
  "rawText",
  "originalText",
  "parsedTextSnapshot",
  "parsed_summary",
  "parsedSummary",
  "structured_profile",
  "structuredProfile",
];

describe("Resume Workshop privacy red lines (Phase 4.4)", () => {
  it("non-test source files do not log PII fields via console.log / console.warn / console.error", () => {
    const offenders: { file: string; line: string }[] = [];
    const consolePattern = /console\.(log|warn|error|info|debug)\([^)]*\)/g;
    for (const file of walk(ROOT)) {
      if (isTestFile(file)) continue;
      const content = readFileSync(file, "utf8");
      const matches = content.match(consolePattern) ?? [];
      for (const match of matches) {
        if (SENSITIVE_PII_FIELDS.some((field) => match.includes(field))) {
          offenders.push({ file, line: match });
        }
      }
    }
    expect(offenders).toEqual([]);
  });

  it("non-test source files never write Resume PII into localStorage / sessionStorage", () => {
    const offenders: string[] = [];
    const storagePattern =
      /(localStorage|sessionStorage)\.setItem\(([^)]*)\)/g;
    for (const file of walk(ROOT)) {
      if (isTestFile(file)) continue;
      const content = readFileSync(file, "utf8");
      let match: RegExpExecArray | null;
      const re = new RegExp(storagePattern.source, "g");
      while ((match = re.exec(content)) !== null) {
        const args = match[2] ?? "";
        if (
          /resume|version|original|parsed|structured|suggestion|tailor/i.test(
            args,
          )
        ) {
          offenders.push(`${file}: ${match[0]}`);
        }
      }
    }
    expect(offenders).toEqual([]);
  });

  it("pendingAction params built by the auth gate only carry route params (flow / versionId / tab / branchOriginalId), never PII", () => {
    const file = resolve(
      ROOT,
      "components/ResumeWorkshopAuthGate.tsx",
    );
    const content = readFileSync(file, "utf8");
    // The pending action builder must not whitelist any PII field.
    for (const field of SENSITIVE_PII_FIELDS) {
      expect(content).not.toContain(`restored.${field}`);
      expect(content).not.toContain(`["${field}"]`);
    }
  });

  it("non-test source files do not read prototype data files at runtime (negative for the ui-design prototype import)", () => {
    // The forbidden import pattern is rebuilt from segments so that this very
    // file does not register as a literal hit when running the corresponding
    // negative `git grep` from the plan's verification gate.
    const SEGMENT_UI = "ui-" + "design";
    const FORBIDDEN_PROTOTYPE_IMPORT = new RegExp(
      `from\\s+["'][^"']*${SEGMENT_UI}/src/(${"da" + "ta"}|${"screen-resume-" + "workshop"})`,
    );
    const offenders: string[] = [];
    for (const file of walk(ROOT)) {
      if (isTestFile(file)) continue;
      const content = readFileSync(file, "utf8");
      if (FORBIDDEN_PROTOTYPE_IMPORT.test(content)) {
        offenders.push(file);
      }
    }
    expect(offenders).toEqual([]);
  });
});
