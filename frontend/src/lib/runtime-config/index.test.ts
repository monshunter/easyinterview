import { describe, it, expect, beforeEach } from 'vitest';
import { fetchRuntimeConfig, _resetRuntimeConfigCache } from './index';
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
  beforeEach(() => {
    _resetRuntimeConfigCache();
  });

  it('parses the server response into the typed shape', async () => {
    const stub = (async () =>
      ({
        ok: true,
        status: 200,
        json: async () => fixture,
      }) as unknown as Response) as typeof fetch;
    const got = await fetchRuntimeConfig({ fetchImpl: stub, endpoint: '/api/v1/runtime-config' });
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
    const first = await fetchRuntimeConfig({ fetchImpl: stub });
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
    await expect(fetchRuntimeConfig({ fetchImpl: stub })).rejects.toThrow(/HTTP 500/);
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
    await fetchRuntimeConfig({ fetchImpl: stub });
    await fetchRuntimeConfig({ fetchImpl: stub, forceRefresh: true });
    expect(calls).toBe(2);
  });
});
