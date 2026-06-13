#!/usr/bin/env node
/**
 * Static fixture server for the pixel-parity Playwright suite.
 *
 * Truth source: docs/spec/frontend-shell/plans/003-ui-design-pixel-parity-
 * gate/plan.md §4 Phase 1.3 + docs/spec/frontend-shell/plans/004-url-
 * addressable-routing/plan.md §6 Phase 4.1. Mounts:
 *   - `frontend/dist/`        at `/`             (production frontend bundle)
 *   - `ui-design/`            at `/ui-design/`   (golden preview prototype)
 *   - `/health`               -> 200 "ok"        (Playwright webServer probe)
 *   - SPA fallback            -> serves `frontend/dist/index.html` for any
 *                                canonical frontend path (see {@link
 *                                FRONTEND_FALLBACK_PATH_PREFIXES}) so reload
 *                                / direct-open of `/workspace`,
 *                                `/auth/login`, etc. lands on the App shell
 *                                instead of 404.
 *   - `/api/*`, `/openapi/*`, `/ui-design/*`, `/health` and file requests
 *     with an extension are NEVER swallowed by the SPA fallback.
 *
 * Offline golden preview: `ui-design/index.html` pins React / ReactDOM /
 * Babel standalone from unpkg, which is unreachable in offline or
 * CDN-blocked environments (2026-06-12 retrospective blocker). The server
 * therefore self-hosts the golden preview without mutating the `ui-design/`
 * truth source:
 *   - `/vendor/react(.dom).development.js` is served from the locally
 *     installed `react` / `react-dom` packages (same pinned 18.3.1).
 *   - `/ui-design/index.html` is rewritten on the fly: unpkg script tags
 *     point at `/vendor/...`, the Babel standalone tag is dropped, and
 *     `type="text/babel"` scripts become plain scripts.
 *   - `/ui-design/**.jsx` is transformed to plain JS at serve time with
 *     esbuild (classic JSX runtime against the global React UMD).
 *
 * Hard requirements:
 *   - Repo-tracked deps only (esbuild + react/react-dom devDependencies);
 *     no network access at runtime.
 *   - Fail loudly with `process.exit(1)` if either source directory is
 *     missing so flaky parity runs surface a clear actionable error.
 */

import { createServer } from "node:http";
import {
  createReadStream,
  statSync,
  existsSync,
  readFileSync,
} from "node:fs";
import { createRequire } from "node:module";
import { dirname, extname, join, normalize, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

import { resolveSpaFallback } from "./spaFallback.mjs";

const require = createRequire(import.meta.url);
const esbuild = require("esbuild");

const VENDOR_FILES = {
  "/vendor/react.development.js": join(
    dirname(require.resolve("react/package.json")),
    "umd",
    "react.development.js",
  ),
  "/vendor/react-dom.development.js": join(
    dirname(require.resolve("react-dom/package.json")),
    "umd",
    "react-dom.development.js",
  ),
};

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
  if (VENDOR_FILES[cleaned]) {
    return { kind: "vendor", absolute: VENDOR_FILES[cleaned] };
  }
  if (cleaned === "/ui-design" || cleaned === "/ui-design/") {
    return { kind: "golden-html", absolute: join(UI_DESIGN_ROOT, "index.html") };
  }
  if (cleaned.startsWith("/ui-design/")) {
    const rel = normalize(cleaned.slice("/ui-design/".length));
    if (rel.startsWith("..") || rel.startsWith(sep + "..")) {
      return { kind: "forbidden" };
    }
    const absolute = join(UI_DESIGN_ROOT, rel);
    if (rel === "index.html") return { kind: "golden-html", absolute };
    if (extname(rel).toLowerCase() === ".jsx") {
      return { kind: "golden-jsx", absolute };
    }
    return { kind: "file", absolute };
  }
  const rel = normalize(cleaned.replace(/^\/+/, ""));
  if (rel.startsWith("..") || rel.startsWith(sep + "..")) {
    return { kind: "forbidden" };
  }
  return { kind: "file", absolute: join(FRONTEND_DIST, rel) };
}

/**
 * Rewrites the golden preview HTML for offline serving: unpkg React /
 * ReactDOM tags point at /vendor/, the Babel standalone tag is removed,
 * external `text/babel` scripts become plain scripts (their .jsx URLs are
 * esbuild-transformed by `golden-jsx` handling), and inline `text/babel`
 * blocks are transformed to plain JS at rewrite time.
 */
function transformGoldenHtml(absolute) {
  let html = readFileSync(absolute, "utf8");
  html = html.replace(
    /<script src="https:\/\/unpkg\.com\/react@[^"]+"[^>]*><\/script>/,
    '<script src="/vendor/react.development.js"></script>',
  );
  html = html.replace(
    /<script src="https:\/\/unpkg\.com\/react-dom@[^"]+"[^>]*><\/script>/,
    '<script src="/vendor/react-dom.development.js"></script>',
  );
  html = html.replace(
    /\s*<script src="https:\/\/unpkg\.com\/@babel\/standalone[^"]+"[^>]*><\/script>/,
    "",
  );
  html = html.replace(/<script type="text\/babel" src=/g, "<script src=");
  html = html.replace(
    /<script type="text\/babel">([\s\S]*?)<\/script>/g,
    (_match, code) => {
      const { code: js } = esbuild.transformSync(code, {
        loader: "jsx",
        jsx: "transform",
      });
      return `<script>${js}</script>`;
    },
  );
  return html;
}

const jsxCache = new Map();

function transformGoldenJsx(absolute) {
  const mtime = statSync(absolute).mtimeMs;
  const cached = jsxCache.get(absolute);
  if (cached && cached.mtime === mtime) return cached.code;
  const source = readFileSync(absolute, "utf8");
  const { code } = esbuild.transformSync(source, {
    loader: "jsx",
    jsx: "transform",
  });
  jsxCache.set(absolute, { mtime, code });
  return code;
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
  if (target.kind === "vendor") {
    res.writeHead(200, {
      "content-type": "application/javascript; charset=utf-8",
      "cache-control": "no-store",
    });
    createReadStream(target.absolute)
      .on("error", () => {
        res.writeHead(500);
        res.end("read error");
      })
      .pipe(res);
    return;
  }
  if (target.kind === "golden-html" || target.kind === "golden-jsx") {
    try {
      const body =
        target.kind === "golden-html"
          ? transformGoldenHtml(target.absolute)
          : transformGoldenJsx(target.absolute);
      res.writeHead(200, {
        "content-type":
          target.kind === "golden-html"
            ? "text/html; charset=utf-8"
            : "application/javascript; charset=utf-8",
        "cache-control": "no-store",
      });
      res.end(body);
    } catch (error) {
      res.writeHead(500, { "content-type": "text/plain; charset=utf-8" });
      res.end(`golden transform error: ${error?.message ?? error}`);
    }
    return;
  }
  let absolute = target.absolute;
  try {
    const stats = statSync(absolute);
    if (stats.isDirectory()) {
      absolute = join(absolute, "index.html");
    }
  } catch {
    // SPA fallback (plan 004 Phase 4.1) — serve `index.html` when the path
    // is a canonical frontend route so reload / direct-open of
    // `/workspace?targetJobId=...` lands on the App shell.
    const fallback = resolveSpaFallback(req.url ?? "/", FRONTEND_DIST);
    if (fallback) {
      const html = fallback.absolute;
      res.writeHead(200, {
        "content-type": "text/html; charset=utf-8",
        "cache-control": "no-store",
      });
      createReadStream(html)
        .on("error", () => {
          res.writeHead(500);
          res.end("read error");
        })
        .pipe(res);
      return;
    }
    // not found — never swallow `/api/*`, `/openapi/*`, scenario script
    // paths or asset 404s with the SPA fallback.
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
