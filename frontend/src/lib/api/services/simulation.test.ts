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

  it('posts branch creation requests with optional source event metadata', async () => {
    const branchInfo = {
      id: '7f3c2d1a9b8e6f10',
      running: false,
      paused: false,
      tick: 42,
      timestamp: '2026-03-12T03:15:05Z',
      untilTick: 96,
      parentId: 'base',
      splitTick: 42,
      splitTimestamp: '2026-03-12T03:15:05Z',
      sourceEvent: {
        id: 'timeline-evt-17',
        type: 'landing_assignment',
        tick: 41,
      },
    };
    const requestJson = vi.fn().mockResolvedValue(branchInfo);
    const client = createSimulationServiceClient({
      requestJson,
      requestText: async () => {
        throw new Error('simulation client should not call requestText');
      },
    });

    const result = await client.createBranchSimulation('base', {
      sourceEvent: {
        id: 'timeline-evt-17',
        type: 'landing_assignment',
        tick: 41,
      },
    });

    expect(result).toEqual(branchInfo);
    expect(requestJson).toHaveBeenCalledWith('/simulations/base/branch', {
      method: 'POST',
      body: JSON.stringify({
        sourceEvent: {
          id: 'timeline-evt-17',
          type: 'landing_assignment',
          tick: 41,
        },
      }),
      signal: undefined,
    });
  });

  it('supports branch creation without source event metadata', async () => {
    const branchInfo = {
      id: '7f3c2d1a9b8e6f10',
      running: false,
      paused: false,
      tick: 42,
      timestamp: '2026-03-12T03:15:05Z',
      parentId: 'base',
      splitTick: 42,
      splitTimestamp: '2026-03-12T03:15:05Z',
    };
    const requestJson = vi.fn().mockResolvedValue(branchInfo);
    const client = createSimulationServiceClient({
      requestJson,
      requestText: async () => {
        throw new Error('simulation client should not call requestText');
      },
    });

    await client.createBranchSimulation('base');

    expect(requestJson).toHaveBeenCalledWith('/simulations/base/branch', {
      method: 'POST',
      body: undefined,
      signal: undefined,
    });
  });

  it('posts assignment overrides and returns aircraft plus assignment metadata', async () => {
    const payload = {
      aircraft: {
        tailNumber: 'BX-101',
        needs: [],
        state: 'Inbound',
        assignedTo: '3a5f',
      },
      assignment: {
        base: '3a5f',
        source: 'human',
      },
    };
    const requestJson = vi.fn().mockResolvedValue(payload);
    const client = createSimulationServiceClient({
      requestJson,
      requestText: async () => {
        throw new Error('simulation client should not call requestText');
      },
    });

    const result = await client.overrideAssignment('branch-123', 'BX-101', { baseId: '3a5f' });

    expect(result).toEqual(payload);
    expect(requestJson).toHaveBeenCalledWith(
      '/simulations/branch-123/aircraft/BX-101/assignment-override',
      {
        method: 'POST',
        body: JSON.stringify({ baseId: '3a5f' }),
        signal: undefined,
      },
    );
  });
});
