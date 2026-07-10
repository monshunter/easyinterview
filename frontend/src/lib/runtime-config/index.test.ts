import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { describe, it, expect } from 'vitest';
import { fetchRuntimeConfig } from './index';
import type { RuntimeConfig } from './types';

const fixture: RuntimeConfig = {
  appVersion: '1.2.3',
  defaultUiLanguage: 'zh-CN',
  analyticsEnabled: false,
  featureFlags: {
    practice_hint_enabled: { enabled: true },
  },
};

describe('fetchRuntimeConfig', () => {
  it('does not expose a test-only cache reset API', () => {
    const resetApi = ['_resetRuntimeConfig', 'Cache'].join('');
    expect(readFileSync(resolve(__dirname, 'index.ts'), 'utf8')).not.toContain(resetApi);
  });

  it('parses the server response into the typed shape', async () => {
    const stub = (async () =>
      ({
        ok: true,
        status: 200,
        json: async () => fixture,
      }) as unknown as Response) as typeof fetch;
    const got = await fetchRuntimeConfig({
      fetchImpl: stub,
      endpoint: '/api/v1/runtime-config',
      forceRefresh: true,
    });
    expect(got).toEqual(fixture);
  });

  it('caches the promise so the second call does not re-fetch', async () => {
    let calls = 0;
    const stub = (async () => {
      calls += 1;
      return {
        ok: true,
        status: 200,
        json: async () => fixture,
      } as unknown as Response;
    }) as typeof fetch;
    const first = await fetchRuntimeConfig({ fetchImpl: stub, forceRefresh: true });
    const second = await fetchRuntimeConfig({ fetchImpl: stub });
    expect(first).toEqual(fixture);
    expect(second).toEqual(fixture);
    expect(calls).toBe(1);
  });

  it('rejects on HTTP errors and clears the cache so retries can succeed', async () => {
    let attempt = 0;
    const stub = (async () => {
      attempt += 1;
      if (attempt === 1) {
        return {
          ok: false,
          status: 500,
          json: async () => ({}),
        } as unknown as Response;
      }
      return {
        ok: true,
        status: 200,
        json: async () => fixture,
      } as unknown as Response;
    }) as typeof fetch;
    await expect(fetchRuntimeConfig({ fetchImpl: stub, forceRefresh: true })).rejects.toThrow(/HTTP 500/);
    const recovered = await fetchRuntimeConfig({ fetchImpl: stub });
    expect(recovered).toEqual(fixture);
  });

  it('honors forceRefresh by re-issuing the request', async () => {
    let calls = 0;
    const stub = (async () => {
      calls += 1;
      return {
        ok: true,
        status: 200,
        json: async () => fixture,
      } as unknown as Response;
    }) as typeof fetch;
    await fetchRuntimeConfig({ fetchImpl: stub, forceRefresh: true });
    await fetchRuntimeConfig({ fetchImpl: stub, forceRefresh: true });
    expect(calls).toBe(2);
  });
});
