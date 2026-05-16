import type { LooseRoute } from "./normalizeRoute";

/**
 * Parse hash routes used by static preview and pixel-parity gates.
 *
 * Supported shape: `#route=report&reportId=...&sessionId=...`.
 * Route normalization remains owned by App; this helper only extracts the
 * loose route name and string params from the browser hash.
 */
export function parseInitialRouteHash(hash: string): LooseRoute | undefined {
  const source = hash.startsWith("#") ? hash.slice(1) : hash;
  if (!source) return undefined;

  const params = new URLSearchParams(source);
  const name = params.get("route");
  if (!name) return undefined;

  params.delete("route");
  return {
    name,
    params: Object.fromEntries(params.entries()),
  };
}
