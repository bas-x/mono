import { getMockAirbaseDetails, getMockAirbases } from '@/lib/api/mock/map';
import { createMockSimulationServiceClient } from '@/lib/api/mock/simulation';
import type { ApiClients } from '@/lib/api/types';

const MOCK_PING_TIME = '2026-01-01T00:00:00.000Z';

export function createMockApiClients(): ApiClients {
  return {
    health: {
      async ping() {
        return {
          ok: true,
          message: 'Mock API health check OK',
          time: MOCK_PING_TIME,
        };
      },
    },
    map: {
      async getAirbases() {
        return getMockAirbases();
      },
      async getAirbaseDetails(idOrUrl: string) {
        return getMockAirbaseDetails(idOrUrl);
      },
    },
    simulation: createMockSimulationServiceClient(),
  };
}
