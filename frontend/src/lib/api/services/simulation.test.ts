import { describe, expect, it, vi } from 'vitest';

import { createSimulationServiceClient } from '@/lib/api/services/simulation';

describe('createSimulationServiceClient', () => {
  it('uses global simulation control endpoints for start pause and resume', async () => {
    const requestJson = vi.fn().mockResolvedValue({});
    const client = createSimulationServiceClient({
      requestJson,
      requestText: async () => {
        throw new Error('simulation client should not call requestText');
      },
    });

    await client.startSimulation('branch-123');
    await client.pauseSimulation('branch-123');
    await client.resumeSimulation('branch-123');

    expect(requestJson).toHaveBeenNthCalledWith(1, '/simulations/start', {
      method: 'POST',
      signal: undefined,
    });
    expect(requestJson).toHaveBeenNthCalledWith(2, '/simulations/pause', {
      method: 'POST',
      signal: undefined,
    });
    expect(requestJson).toHaveBeenNthCalledWith(3, '/simulations/resume', {
      method: 'POST',
      signal: undefined,
    });
  });
});
