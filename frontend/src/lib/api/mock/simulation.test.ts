import { describe, expect, it } from 'vitest';

import { createMockSimulationServiceClient } from '@/lib/api/mock/simulation';

describe('createMockSimulationServiceClient', () => {
  it('exposes both light and full mock simulations', async () => {
    const client = createMockSimulationServiceClient();

    const simulations = await client.getSimulations();

    expect(simulations.map((simulation) => simulation.id)).toEqual([
      'mock-light-sortie',
      'mock-full-sortie',
    ]);
    expect(simulations[0]?.untilTick).toBe(8);
    expect(simulations[1]?.untilTick).toBe(16);
  });

  it('returns richer mock airbase coverage for the full scenario', async () => {
    const client = createMockSimulationServiceClient();

    const airbases = await client.getAirbases('mock-full-sortie');

    expect(airbases).toHaveLength(7);
    expect(airbases[0]).toMatchObject({
      regionId: 'SE-K',
      region: 'Blekinge',
    });
  });

  it('selects the light scenario when the requested setup is intentionally small', async () => {
    const client = createMockSimulationServiceClient();

    const created = await client.createBaseSimulation({
      untilTick: 8,
      simulationOptions: {
        constellationOpts: { maxTotal: 4 },
        fleetOpts: { aircraftMax: 4 },
      },
    });

    expect(created.id).toBe('mock-light-sortie');
  });

  it('creates branch simulations with lineage metadata', async () => {
    const client = createMockSimulationServiceClient();

    const branch = await client.createBranchSimulation('base', {
      sourceEvent: {
        id: 'timeline-evt-17',
        type: 'landing_assignment',
        tick: 41,
      },
    });

    expect(branch.id).toMatch(/^branch-/);
    expect(branch.parentId).toBe('base');
    expect(branch.splitTick).toBeTypeOf('number');
    expect(branch.splitTimestamp).toBeTypeOf('string');
    expect(branch.sourceEvent).toEqual({
      id: 'timeline-evt-17',
      type: 'landing_assignment',
      tick: 41,
    });
  });
});
