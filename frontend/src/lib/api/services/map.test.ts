import { describe, expect, it } from 'vitest';

import { createMapServiceClient } from '@/lib/api/services/map';

describe('createMapServiceClient', () => {
  it('returns fixture airbases without backend map requests', async () => {
    const mapClient = createMapServiceClient({
      requestJson: async () => {
        throw new Error('map client should not call requestJson');
      },
      requestText: async () => {
        throw new Error('map client should not call requestText');
      },
    });

    const result = await mapClient.getAirbases();

    expect(result).toHaveLength(4);
    expect(result[0]?.id).toBe('lulea');
    expect(result[0]?.area).toHaveLength(4);
  });

  it('returns fixture airbase details by id', async () => {
    const mapClient = createMapServiceClient({
      requestJson: async () => {
        throw new Error('map client should not call requestJson');
      },
      requestText: async () => {
        throw new Error('map client should not call requestText');
      },
    });

    const result = await mapClient.getAirbaseDetails('lulea');

    expect(result).toMatchObject({
      id: 'lulea',
      name: 'Lulea Airbase',
    });
  });

  it('normalizes airbase path lookup keys against fixtures', async () => {
    const mapClient = createMapServiceClient({
      requestJson: async () => {
        throw new Error('map client should not call requestJson');
      },
      requestText: async () => {
        throw new Error('map client should not call requestText');
      },
    });

    const result = await mapClient.getAirbaseDetails('/airbase/visby');

    expect(result).toMatchObject({
      id: 'visby',
      name: 'Visby Airbase',
    });
  });
});
