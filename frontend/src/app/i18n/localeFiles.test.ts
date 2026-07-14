import { existsSync, readFileSync, readdirSync } from "node:fs";
import { relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import ts from "typescript";
import { describe, expect, it } from "vitest";

const SOURCE_ROOT = fileURLToPath(new URL("../../", import.meta.url));

function sourceFilesUnder(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = resolve(directory, entry.name);
    return entry.isDirectory() ? sourceFilesUnder(path) : [path];
  });
}

function isProductionTypeScript(path: string): boolean {
  const rel = relative(SOURCE_ROOT, path).replaceAll("\\", "/");
  return (
    /\.(ts|tsx)$/.test(rel) &&
    !/\.test\.(ts|tsx)$/.test(rel) &&
    !rel.includes("/__tests__/") &&
    !rel.startsWith("app/__tests__/") &&
    !rel.startsWith("api/generated/") &&
    !rel.startsWith("app/i18n/locales/") &&
    !rel.startsWith("test/")
  );
}

function unwrapExpression(
  expression: ts.Expression | undefined,
): ts.Expression | undefined {
  while (
    expression &&
    (ts.isAsExpression(expression) ||
      ts.isSatisfiesExpression(expression) ||
      ts.isParenthesizedExpression(expression))
  ) {
    expression = expression.expression;
  }
  return expression;
}

function localeKeys(path: string, exportName: string): string[] {
  const source = readFileSync(path, "utf8");
  const sourceFile = ts.createSourceFile(
    path,
    source,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TS,
  );
  const keys: string[] = [];
  const visit = (node: ts.Node): void => {
    if (
      ts.isVariableDeclaration(node) &&
      ts.isIdentifier(node.name) &&
      node.name.text === exportName
    ) {
      const initializer = unwrapExpression(node.initializer);
      if (initializer && ts.isObjectLiteralExpression(initializer)) {
        for (const property of initializer.properties) {
          if (
            ts.isPropertyAssignment(property) &&
            (ts.isStringLiteral(property.name) ||
              ts.isNoSubstitutionTemplateLiteral(property.name))
          ) {
            keys.push(property.name.text);
          }
        }
      }
    }
    ts.forEachChild(node, visit);
  };
  visit(sourceFile);
  return keys;
}

function productionStringLiterals(): Set<string> {
  const literals = new Set<string>();
  for (const path of sourceFilesUnder(SOURCE_ROOT).filter(
    isProductionTypeScript,
  )) {
    const source = readFileSync(path, "utf8");
    const sourceFile = ts.createSourceFile(
      path,
      source,
      ts.ScriptTarget.Latest,
      true,
      path.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
    );
    const visit = (node: ts.Node): void => {
      if (
        ts.isStringLiteral(node) ||
        ts.isNoSubstitutionTemplateLiteral(node)
      ) {
        literals.add(node.text);
      }
      ts.forEachChild(node, visit);
    };
    visit(sourceFile);
  }
  return literals;
}

describe("D1 shell i18n locale file structure", () => {
  it("keeps each supported locale in an independent locale file", () => {
    const zhFile = new URL("./locales/zh.ts", import.meta.url);
    const enFile = new URL("./locales/en.ts", import.meta.url);
    const messagesFile = new URL("./messages.ts", import.meta.url);
    const catalogFile = new URL("./localeCatalog.ts", import.meta.url);

    expect(existsSync(zhFile)).toBe(true);
    expect(existsSync(enFile)).toBe(true);
    expect(existsSync(catalogFile)).toBe(true);

    const zhSource = readFileSync(zhFile, "utf8");
    const enSource = readFileSync(enFile, "utf8");
    const messagesSource = readFileSync(messagesFile, "utf8");
    const catalogSource = readFileSync(catalogFile, "utf8");

    expect(zhSource).toContain('export const zh');
    expect(enSource).toContain('export const en');
    expect(messagesSource).not.toMatch(/\bzh:\s*\{/);
    expect(messagesSource).not.toMatch(/\ben:\s*\{/);
    expect(catalogSource).toContain("SUPPORTED_LOCALES");
    expect(catalogSource).toContain("aliases");
    expect(catalogSource).toContain("shortLabel");
  });

  it("keeps every locale key reachable from production source", () => {
    const zhFile = fileURLToPath(new URL("./locales/zh.ts", import.meta.url));
    const keys = localeKeys(zhFile, "zh");
    const literals = productionStringLiterals();

    expect(keys.length).toBeGreaterThan(0);
    expect(keys.filter((key) => !literals.has(key))).toEqual([]);
  });

  it("contains home.* namespace keys in both zh and en (≥14 keys)", () => {
    const zhSource = readFileSync(
      new URL("./locales/zh.ts", import.meta.url),
      "utf8",
    );
    const enSource = readFileSync(
      new URL("./locales/en.ts", import.meta.url),
      "utf8",
    );

    const requiredKeys = [
      "home.heroLabel",
      "home.heroTitle",
      "home.jdPlaceholder",
      "home.pasteSource",
      "home.importBtn",
      "home.recentSection",
      "home.recentSectionSub",
      "home.recentMore",
      "home.resumeSelect",
      "home.resumeSelectPlaceholder",
      "home.resumeSelectHint",
      "home.resumeSelected",
      "home.resumeLoading",
      "home.resumeEmpty",
      "home.resumeCreateLink",
      "home.pendingImportInvalid",
    ];

    // product-scope D-22 keeps the home debrief CTA outside current scope.
    expect(requiredKeys.length).toBeGreaterThanOrEqual(14);

    for (const key of requiredKeys) {
      expect(zhSource).toContain(`"${key}"`);
      expect(enSource).toContain(`"${key}"`);
    }

    const retiredJdIntakeKeys = [
      "home.uploadSource",
      "home.orUpload",
      "home.modalLabel",
      "home.modalUploadTitle",
      "home.modalUrlTitle",
      "home.modalUploadDropzone",
      "home.modalUploadHint",
      "home.modalUrlLabel",
      "home.modalUrlPlaceholder",
      "home.modalUrlHint",
      "home.modalCancel",
      "home.modalContinue",
    ];
    for (const key of retiredJdIntakeKeys) {
      expect(zhSource).not.toContain(`"${key}"`);
      expect(enSource).not.toContain(`"${key}"`);
    }
  });

  it("does not keep out-of-scope Resume Workshop per-row tree toast keys", () => {
    const zhSource = readFileSync(
      new URL("./locales/zh.ts", import.meta.url),
      "utf8",
    );
    const enSource = readFileSync(
      new URL("./locales/en.ts", import.meta.url),
      "utf8",
    );

    for (const source of [zhSource, enSource]) {
      expect(source).not.toContain("resumeWorkshop.tree.toastSelect");
      expect(source).not.toContain("resumeWorkshop.tree.toastBranch");
    }
  });

  it("does not keep out-of-scope Resume Workshop tree, branch, or unavailable-surface namespaces", () => {
    const zhSource = readFileSync(
      new URL("./locales/zh.ts", import.meta.url),
      "utf8",
    );
    const enSource = readFileSync(
      new URL("./locales/en.ts", import.meta.url),
      "utf8",
    );

    for (const source of [zhSource, enSource]) {
      expect(source).not.toContain("resumeWorkshop.notImplemented.");
      expect(source).not.toContain("resumeWorkshop.branch.");
      expect(source).not.toContain("resumeWorkshop.tree.");
      expect(source).not.toContain("resumeWorkshop.flat.");
      expect(source).not.toContain("resumeWorkshop.stats.");
      expect(source).not.toContain("resumeWorkshop.viewSwitcher.");
      expect(source).not.toContain("resumeWorkshop.openVersion");
      expect(source).not.toContain("resumeWorkshop.list.versionsError");
      expect(source).not.toContain("resumeWorkshop.preview.sidebar.whatSaved.master");
      expect(source).not.toContain("resumeWorkshop.detail.crumbVersions");
      expect(source).not.toContain("resumeWorkshop.detail.comingSoon");
      expect(source).not.toContain("resumeWorkshop.edit.scope.master");
      expect(source).not.toContain("resumeWorkshop.edit.scope.targeted");
      expect(source).not.toMatch(
        /from master|master version|targeted version|create a new suggestion or branch/i,
      );
      expect(source).not.toMatch(/主版本|定制版本|分叉|版本内容/);
    }
  });

  it("does not keep out-of-scope Resume Workshop guided create keys", () => {
    const zhSource = readFileSync(
      new URL("./locales/zh.ts", import.meta.url),
      "utf8",
    );
    const enSource = readFileSync(
      new URL("./locales/en.ts", import.meta.url),
      "utf8",
    );

    for (const source of [zhSource, enSource]) {
      expect(source).not.toContain("resumeWorkshop.create.tabs.guided");
      expect(source).not.toContain("resumeWorkshop.create.guided.");
      expect(source).not.toContain("resumeWorkshop.sourceType.guided");
    }
  });
});
