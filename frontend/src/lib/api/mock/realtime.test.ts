import { describe, expect, it, vi } from 'vitest';

import { createMockSimulationStreamClient } from '@/lib/api/mock/realtime';
import { createMockSimulationServiceClient } from '@/lib/api/mock/simulation';

describe('createMockSimulationStreamClient', () => {
  it('emits deterministic sequence values', () => {
    vi.useFakeTimers();

    const client = createMockSimulationStreamClient();
    const sequences: number[] = [];

    client.subscribe((event) => {
      if (typeof event.sequence === 'number') {
        sequences.push(event.sequence);
      }
    });

    client.connect('base');
    vi.advanceTimersByTime(250);
    vi.advanceTimersByTime(1_500);
    vi.advanceTimersByTime(1_500);
    vi.advanceTimersByTime(1_500);

    expect(sequences.length).toBeGreaterThanOrEqual(3);
    expect(sequences.slice(0, 3)).toEqual([1, 2, 3]);

    client.disconnect();
    vi.useRealTimers();
  });

  it('emits a richer mock event stream for the full scenario', () => {
    vi.useFakeTimers();

    const client = createMockSimulationStreamClient();
    const eventTypes: string[] = [];

    client.subscribe((event) => {
      eventTypes.push(event.type);
    });

    client.connect('mock-full-sortie');
    vi.advanceTimersByTime(250 + 1_500 * 6);

    expect(eventTypes).toContain('all_aircraft_positions');
    expect(eventTypes).toContain('landing_assignment');
    expect(eventTypes).toContain('aircraft_state_change');

    client.disconnect();
    vi.useRealTimers();
  });

  it('emits branch_created on the base stream after branch creation', async () => {
    vi.useFakeTimers();

    const service = createMockSimulationServiceClient();
    const client = createMockSimulationStreamClient();
    const branchEvents: string[] = [];

    client.subscribe((event) => {
      if (event.type === 'branch_created') {
        branchEvents.push(String(event.branchId));
      }
    });

    client.connect('base');
    vi.advanceTimersByTime(250);
    await service.createBranchSimulation('base', {
      sourceEvent: {
        id: 'timeline-evt-17',
        type: 'landing_assignment',
        tick: 41,
      },
    });

    expect(branchEvents).toHaveLength(1);
    expect(branchEvents[0]).toMatch(/^branch-/);

    client.disconnect();
    vi.useRealTimers();
  });

  it('emits human landing_assignment on the targeted simulation stream after override', async () => {
    vi.useFakeTimers();

    const service = createMockSimulationServiceClient();
    const client = createMockSimulationStreamClient();
    const seenEvents: Array<{ baseId: string; source: string }> = [];

    client.subscribe((event) => {
      if (event.type === 'landing_assignment' && event.source === 'human') {
        seenEvents.push({ baseId: String(event.baseId), source: String(event.source) });
      }
    });

    client.connect('base');
    vi.advanceTimersByTime(250);

    const aircrafts = await service.getAircrafts('base');
    const airbases = await service.getAirbases('base');
    await service.overrideAssignment('base', aircrafts[0]!.tailNumber, { baseId: airbases[1]!.id });

    expect(seenEvents.at(-1)).toEqual({ baseId: airbases[1]!.id, source: 'human' });

    client.disconnect();
    vi.useRealTimers();
  });
});
