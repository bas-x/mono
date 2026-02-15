import type { RpcClients } from '@/lib/rpc/types';

const MOCK_PING_TIME = '2026-01-01T00:00:00.000Z';

export function createMockClients(): RpcClients {
  return {
    health: {
      async ping() {
        return {
          ok: true,
          message: 'Mock RPC health check OK',
          time: MOCK_PING_TIME,
        };
      },
    },
  };
}
