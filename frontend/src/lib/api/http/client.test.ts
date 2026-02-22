import { afterEach, describe, expect, it, vi } from 'vitest';

import { buildApiUrl, createHttpClient, HttpClientError } from '@/lib/api/http/client';

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
  vi.restoreAllMocks();
});

describe('buildApiUrl', () => {
  it('normalizes base and path slashes', () => {
    expect(buildApiUrl('https://api.example.com/', '/health')).toBe('https://api.example.com/health');
    expect(buildApiUrl('https://api.example.com', 'health')).toBe('https://api.example.com/health');
  });
});

describe('createHttpClient', () => {
  it('requests JSON with standard headers', async () => {
    globalThis.fetch = vi.fn(async () => {
      return new Response(JSON.stringify({ status: 'ok' }), { status: 200 });
    }) as typeof fetch;

    const client = createHttpClient({ apiBaseUrl: 'https://api.example.com' });
    const result = await client.requestJson<{ status: string }>('/health');

    expect(result.status).toBe('ok');
    expect(globalThis.fetch).toHaveBeenCalledWith(
      'https://api.example.com/health',
      expect.objectContaining({ method: 'GET' }),
    );
  });

  it('throws HttpClientError on non-2xx', async () => {
    globalThis.fetch = vi.fn(async () => {
      return new Response('denied', { status: 403 });
    }) as typeof fetch;

    const client = createHttpClient({ apiBaseUrl: 'https://api.example.com' });

    await expect(client.requestText('/health')).rejects.toBeInstanceOf(HttpClientError);
    await expect(client.requestText('/health')).rejects.toMatchObject({
      status: 403,
      path: '/health',
    });
  });
});
