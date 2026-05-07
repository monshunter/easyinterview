#!/usr/bin/env node
/**
 * Static fixture server for the pixel-parity Playwright suite.
 *
 * Truth source: docs/spec/frontend-shell/plans/003-ui-design-pixel-parity-
 * gate/plan.md §4 Phase 1.3. Mounts:
 *   - `frontend/dist/`        at `/`             (production frontend bundle)
 *   - `ui-design/`            at `/ui-design/`   (golden preview prototype)
 *   - `/health`               -> 200 "ok"        (Playwright webServer probe)
 *
 * Hard requirements:
 *   - Node built-ins only (no third-party deps).
 *   - Fail loudly with `process.exit(1)` if either source directory is
 *     missing so flaky parity runs surface a clear actionable error.
 */

import { createServer } from "node:http";
import { createReadStream, statSync, existsSync } from "node:fs";
import { extname, join, normalize, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

const HOST = "127.0.0.1";
const PORT = Number(process.env.PIXEL_PARITY_PORT ?? 4173);

const HERE = resolve(fileURLToPath(import.meta.url), "..");
const FRONTEND_ROOT = resolve(HERE, "..");
const REPO_ROOT = resolve(FRONTEND_ROOT, "..");
const FRONTEND_DIST = resolve(FRONTEND_ROOT, "dist");
const UI_DESIGN_ROOT = resolve(REPO_ROOT, "ui-design");

function fail(message) {
  process.stderr.write(`[serve-pixel-parity] ${message}\n`);
  process.exit(1);
}

if (!existsSync(FRONTEND_DIST)) {
  fail(
    `frontend dist not found at ${FRONTEND_DIST}; run \`pnpm --filter @easyinterview/frontend build\` first`,
  );
}
if (!existsSync(UI_DESIGN_ROOT)) {
  fail(`ui-design directory not found at ${UI_DESIGN_ROOT}`);
}

const MIME = {
  ".html": "text/html; charset=utf-8",
  ".js": "application/javascript; charset=utf-8",
  ".mjs": "application/javascript; charset=utf-8",
  ".jsx": "application/javascript; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".svg": "image/svg+xml",
  ".png": "image/png",
  ".jpg": "image/jpeg",
  ".jpeg": "image/jpeg",
  ".gif": "image/gif",
  ".woff": "font/woff",
  ".woff2": "font/woff2",
  ".ttf": "font/ttf",
  ".otf": "font/otf",
  ".ico": "image/x-icon",
  ".map": "application/json; charset=utf-8",
};

function resolveStaticPath(urlPath) {
  // Strip query string + decode.
  const cleaned = decodeURIComponent(urlPath.split("?")[0] ?? "/");
  if (cleaned === "/health") return { kind: "health" };
  if (cleaned === "/" || cleaned === "") {
    return { kind: "file", absolute: join(FRONTEND_DIST, "index.html") };
  }
  if (cleaned === "/ui-design" || cleaned === "/ui-design/") {
    return { kind: "file", absolute: join(UI_DESIGN_ROOT, "index.html") };
  }
  if (cleaned.startsWith("/ui-design/")) {
    const rel = normalize(cleaned.slice("/ui-design/".length));
    if (rel.startsWith("..") || rel.startsWith(sep + "..")) {
      return { kind: "forbidden" };
    }
    return { kind: "file", absolute: join(UI_DESIGN_ROOT, rel) };
  }
  const rel = normalize(cleaned.replace(/^\/+/, ""));
  if (rel.startsWith("..") || rel.startsWith(sep + "..")) {
    return { kind: "forbidden" };
  }
  return { kind: "file", absolute: join(FRONTEND_DIST, rel) };
}

const server = createServer((req, res) => {
  const target = resolveStaticPath(req.url ?? "/");
  if (target.kind === "health") {
    res.writeHead(200, { "content-type": "text/plain; charset=utf-8" });
    res.end("ok");
    return;
  }
  if (target.kind === "forbidden") {
    res.writeHead(403);
    res.end("forbidden");
    return;
  }
  let absolute = target.absolute;
  try {
    const stats = statSync(absolute);
    if (stats.isDirectory()) {
      absolute = join(absolute, "index.html");
    }
  } catch {
    // not found
    res.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
    res.end(`not found: ${req.url}`);
    return;
  }
  const ext = extname(absolute).toLowerCase();
  const contentType = MIME[ext] ?? "application/octet-stream";
  res.writeHead(200, {
    "content-type": contentType,
    "cache-control": "no-store",
  });
  createReadStream(absolute)
    .on("error", () => {
      res.writeHead(500);
      res.end("read error");
    })
    .pipe(res);
});

server.listen(PORT, HOST, () => {
  process.stdout.write(
    `[serve-pixel-parity] listening on http://${HOST}:${PORT} (frontend dist + /ui-design/)\n`,
  );
});

const shutdown = () => {
  server.close(() => process.exit(0));
};
process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);
