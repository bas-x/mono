import { describe, expect, it } from 'vitest';

import type { HttpClient } from '@/lib/api/http/client';
import { createHealthServiceClient } from '@/lib/api/services/health';

function createHttpClientStub(status: string, ping: string): HttpClient {
  return {
    requestJson: async <TResponse>() => ({ status } as TResponse),
    requestText: async () => ping,
  };
}

describe('createHealthServiceClient', () => {
  it('returns ok=true when health and ping endpoints are healthy', async () => {
    const healthClient = createHealthServiceClient(
      { apiBaseUrl: 'https://api.example.com' },
      createHttpClientStub('ok', 'pong'),
    );

    const result = await healthClient.ping();

    expect(result.ok).toBe(true);
    expect(result.message).toContain('health=ok');
    expect(result.message).toContain('ping=pong');
  });

  it('returns ok=false when one endpoint response is not expected', async () => {
    const healthClient = createHealthServiceClient(
      { apiBaseUrl: 'https://api.example.com' },
      createHttpClientStub('ok', 'not-pong'),
    );

    const result = await healthClient.ping();

    expect(result.ok).toBe(false);
  });
});
