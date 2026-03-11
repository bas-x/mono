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
      expect.objectContaining({
        method: 'GET',
        headers: expect.any(Headers),
      }),
    );
    const request = vi.mocked(globalThis.fetch).mock.calls[0]?.[1];
    expect(new Headers(request?.headers).get('Accept')).toBe('application/json');
  });

  it('sets application/json content type for JSON request bodies', async () => {
    globalThis.fetch = vi.fn(async () => {
      return new Response(JSON.stringify({ id: 'base' }), { status: 200 });
    }) as typeof fetch;

    const client = createHttpClient({ apiBaseUrl: 'https://api.example.com' });
    await client.requestJson('/simulations/base', {
      method: 'POST',
      body: JSON.stringify({ seed: 'abc123' }),
    });

    const request = vi.mocked(globalThis.fetch).mock.calls[0]?.[1];
    expect(new Headers(request?.headers).get('Content-Type')).toBe('application/json');
    expect(new Headers(request?.headers).get('Accept')).toBe('application/json');
  });

  it('preserves explicit content type overrides', async () => {
    globalThis.fetch = vi.fn(async () => {
      return new Response(JSON.stringify({ ok: true }), { status: 200 });
    }) as typeof fetch;

    const client = createHttpClient({ apiBaseUrl: 'https://api.example.com' });
    await client.requestJson('/custom', {
      method: 'POST',
      body: JSON.stringify({ seed: 'abc123' }),
      headers: { 'Content-Type': 'application/merge-patch+json' },
    });

    const request = vi.mocked(globalThis.fetch).mock.calls[0]?.[1];
    expect(new Headers(request?.headers).get('Content-Type')).toBe('application/merge-patch+json');
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
