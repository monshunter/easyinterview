import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const devServerPort = envPort("FRONTEND_HOST_PORT", 5173);
const previewServerPort = envPort("FRONTEND_PREVIEW_PORT", 4173);

function envPort(name: string, fallback: number): number {
  const raw = process.env[name]?.trim();
  if (!raw) {
    return fallback;
  }
  const parsed = Number(raw);
  if (!Number.isInteger(parsed) || parsed <= 0 || parsed > 65535) {
    throw new Error(`${name} must be a valid TCP port, got ${raw}`);
  }
  return parsed;
}

/**
 * SPA fallback (plan 004 Phase 4.1) — `appType: "spa"` is Vite's default
 * which already serves `index.html` for any client-side route. We pin the
 * field explicitly so future changes to vite defaults cannot silently
 * break the URL-addressable routing contract (dev / preview both must
 * return `index.html` for `/workspace`, `/auth/login`, etc. while leaving
 * `/api/*` paths alone — the App does not proxy API in dev today).
 */
export default defineConfig({
  plugins: [react()],
  appType: "spa",
  server: {
    port: devServerPort,
  },
  preview: {
    port: previewServerPort,
  },
});
