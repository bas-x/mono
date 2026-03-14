import { describe, expect, it } from 'vitest';

import { parseSimulationEvent } from '@/lib/api/realtime/simulationStream';

describe('parseSimulationEvent', () => {
  it('parses simulation_ended with a direct summary object', () => {
    const event = parseSimulationEvent(JSON.stringify({
      type: 'simulation_ended',
      simulationId: 'base',
      tick: 12,
      timestamp: '2026-03-12T03:15:05Z',
      summary: {
        completedVisitCount: 0,
        totalDurationMs: 0,
        averageDurationMs: null,
      },
    }));

    expect(event).toMatchObject({
      type: 'simulation_ended',
      summary: {
        completedVisitCount: 0,
        totalDurationMs: 0,
        averageDurationMs: null,
      },
    });
  });

  it('parses simulation_closed with reason and summary', () => {
    const event = parseSimulationEvent(JSON.stringify({
      type: 'simulation_closed',
      simulationId: 'branch-1',
      tick: 9,
      timestamp: '2026-03-12T03:16:05Z',
      reason: 'cancel',
      summary: {
        completedVisitCount: 2,
        totalDurationMs: 10000,
        averageDurationMs: 5000,
      },
    }));

    expect(event).toMatchObject({
      type: 'simulation_closed',
      reason: 'cancel',
      summary: {
        completedVisitCount: 2,
        totalDurationMs: 10000,
        averageDurationMs: 5000,
      },
    });
  });

  it('rejects nested terminal summary payloads', () => {
    const event = parseSimulationEvent(JSON.stringify({
      type: 'simulation_closed',
      simulationId: 'base',
      tick: 9,
      timestamp: '2026-03-12T03:16:05Z',
      reason: 'reset',
      summary: {
        servicing: {
          completedVisitCount: 0,
          totalDurationMs: 0,
          averageDurationMs: null,
        },
      },
    }));

    expect(event).toBeNull();
  });

  it('preserves landing_assignment needs and capabilities payloads', () => {
    const event = parseSimulationEvent(JSON.stringify({
      type: 'landing_assignment',
      simulationId: 'base',
      tailNumber: 'BX-101',
      baseId: 'airbase-1',
      source: 'algorithm',
      needs: [
        {
          type: 'fuel',
          severity: 42,
          requiredCapability: 'fuel',
          blocking: false,
        },
      ],
      capabilities: {
        fuel: { recoveryMultiplierPermille: 1300 },
      },
      timestamp: '2026-03-12T03:15:05Z',
    }));

    expect(event).toMatchObject({
      type: 'landing_assignment',
      needs: [
        {
          type: 'fuel',
          severity: 42,
          requiredCapability: 'fuel',
          blocking: false,
        },
      ],
      capabilities: {
        fuel: { recoveryMultiplierPermille: 1300 },
      },
    });
  });
});
