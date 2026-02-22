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
  };
}
