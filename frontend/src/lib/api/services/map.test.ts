import { afterEach, describe, expect, it, vi } from 'vitest';

import type { HttpClient } from '@/lib/api/http/client';
import { createMapServiceClient } from '@/lib/api/services/map';

const originalFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = originalFetch;
  vi.restoreAllMocks();
});

function createHttpClientStub<TResponse>(response: TResponse): HttpClient {
  return {
    requestJson: async <T>() => response as unknown as T,
    requestText: async () => '',
  };
}

describe('createMapServiceClient', () => {
  it('returns normalized airbases from /map', async () => {
    const mapClient = createMapServiceClient(
      createHttpClientStub({
        airbases: [
          {
            id: 'alpha',
            area: [
              { x: 1, y: 2 },
              { x: 3, y: 4 },
              { x: 5, y: 6 },
            ],
          },
        ],
      }),
    );

    const result = await mapClient.getAirbases();

    expect(result).toHaveLength(1);
    expect(result[0]?.id).toBe('alpha');
    expect(result[0]?.area).toHaveLength(3);
  });

  it('requests airbase details by id endpoint', async () => {
    const requestJson = vi.fn(async () => ({ id: 'bravo', name: 'Bravo Airbase' })) as unknown as HttpClient['requestJson'];
    const mapClient = createMapServiceClient(
      {
        requestJson,
        requestText: async () => '',
      },
    );

    const result = await mapClient.getAirbaseDetails('bravo');

    expect(requestJson).toHaveBeenCalledWith('/map/airbase/bravo', { signal: undefined });
    expect(result.id).toBe('bravo');
  });

  it('uses absolute infoUrl directly when provided', async () => {
    globalThis.fetch = vi.fn(async () => {
      return new Response(JSON.stringify({ id: 'charlie', status: 'Operational' }), { status: 200 });
    }) as typeof fetch;

    const mapClient = createMapServiceClient(
      createHttpClientStub({ airbases: [] }),
    );

    const result = await mapClient.getAirbaseDetails('https://demo.example.com/map/airbase/charlie');

    expect(globalThis.fetch).toHaveBeenCalledWith(
      'https://demo.example.com/map/airbase/charlie',
      expect.objectContaining({ method: 'GET' }),
    );
    expect(result.id).toBe('charlie');
  });
});
