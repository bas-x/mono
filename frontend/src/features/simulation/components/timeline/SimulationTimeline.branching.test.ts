import { describe, expect, it } from 'vitest';

import { getTimelineSplitMarker } from '@/features/simulation/components/timeline/SimulationTimeline';

describe('getTimelineSplitMarker', () => {
  it('uses canonical splitTick and splitTimestamp for branch markers', () => {
    const marker = getTimelineSplitMarker({
      status: 'running',
      simulationId: 'branch-00000001',
      isRunnerActive: false,
      isRunnerPaused: true,
      airbases: [],
      aircrafts: [],
      tick: 52,
      time: '2026-03-12T03:20:05Z',
      history: {},
      playbackTick: null,
      maxTick: 96,
      untilTick: 96,
      parentId: 'base',
      splitTick: 42,
      splitTimestamp: '2026-03-12T03:15:05Z',
      sourceEvent: {
        id: 'timeline-evt-17',
        type: 'landing_assignment',
        tick: 41,
      },
    });

    expect(marker).toEqual({
      tick: 42,
      timestamp: '2026-03-12T03:15:05Z',
    });
  });

  it('returns no marker for base simulations', () => {
    const marker = getTimelineSplitMarker({
      status: 'running',
      simulationId: 'base',
      isRunnerActive: true,
      isRunnerPaused: false,
      airbases: [],
      aircrafts: [],
      tick: 42,
      time: '2026-03-12T03:15:05Z',
      history: {},
      playbackTick: null,
      maxTick: 96,
      untilTick: 96,
    });

    expect(marker).toBeNull();
  });
});
