import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, resolve } from "node:path";

import { describe, expect, it } from "vitest";

const FRONTEND_SRC = resolve(__dirname, "..");

function* walk(dir: string): Generator<string> {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      yield* walk(full);
    } else if (/\.(ts|tsx)$/.test(entry)) {
      yield full;
    }
  }
}

describe("frontend D1 scope guards", () => {
  it("never imports ui-design/src/data.jsx into the active frontend bundle", () => {
    const importPattern = /from\s+["'][^"']*ui-design\/src\/data/;
    const offenders: string[] = [];
    for (const file of walk(FRONTEND_SRC)) {
      const content = readFileSync(file, "utf8");
      if (importPattern.test(content)) {
        offenders.push(file);
      }
    }
    expect(offenders).toEqual([]);
  });
});
