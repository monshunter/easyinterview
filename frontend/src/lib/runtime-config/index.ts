/**
 * Runtime-config fetcher (secrets-and-config spec D-2 / §5).
 *
 * `fetchRuntimeConfig()` calls `GET /api/v1/runtime-config` once per page
 * load and caches the result in module-scoped state. D1 frontend-shell
 * wraps this with a React provider; this module stays framework-free so
 * non-React surfaces (ssr smoke tests, vitest, future workers) can reuse
 * it.
 *
 * OpenAPI schema is owned by B2 openapi-v1-contract. The field allowlist
 * is owned by A4 — never extend it without revising the spec.
 */

import type { RuntimeConfig } from './types';

let cached: Promise<RuntimeConfig> | undefined;

export interface FetchOptions {
  /** Override the endpoint (used by tests). Defaults to '/api/v1/runtime-config'. */
  endpoint?: string;
  /** Override the underlying fetch (for tests / SSR). Defaults to the global fetch. */
  fetchImpl?: typeof fetch;
  /** Bypass the module cache (forces a fresh request). */
  forceRefresh?: boolean;
}

/**
 * Returns the runtime configuration document. The first call performs the
 * HTTP request; subsequent calls within the same page load return the
 * cached promise, even if the previous fetch is still in flight (so we
 * never issue duplicate requests on hot startup).
 */
export function fetchRuntimeConfig(options: FetchOptions = {}): Promise<RuntimeConfig> {
  if (!options.forceRefresh && cached) {
    return cached;
  }
  const fetcher = options.fetchImpl ?? globalThis.fetch;
  if (!fetcher) {
    return Promise.reject(new Error('runtime-config: global fetch is not available'));
  }
  const endpoint = options.endpoint ?? '/api/v1/runtime-config';
  const promise = fetcher(endpoint, {
    method: 'GET',
    credentials: 'include',
    headers: { Accept: 'application/json' },
  })
    .then(async (resp) => {
      if (!resp.ok) {
        throw new Error(`runtime-config: HTTP ${resp.status}`);
      }
      return (await resp.json()) as RuntimeConfig;
    })
    .catch((err) => {
      // Reset cache so a transient failure does not pin the module to a
      // rejected promise forever.
      if (cached === promise) {
        cached = undefined;
      }
      throw err;
    });
  cached = promise;
  return promise;
}

/** Clears the module-scoped cache. Tests use this between scenarios. */
export function _resetRuntimeConfigCache(): void {
  cached = undefined;
}

export type { RuntimeConfig, RuntimeFlag } from './types';
