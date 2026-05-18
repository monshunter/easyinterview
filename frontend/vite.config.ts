import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

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
    port: 5173,
  },
  preview: {
    port: 4173,
  },
});
