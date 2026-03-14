import { describe, expect, it } from 'vitest';

import type { SimulationEvent, SimulationInfo } from '@/lib/api/types';
import type { TerminalSimulationRecord } from '@/features/simulation/hooks/useSimulation';

import {
  buildTimelineGraph,
  clampTimelineTick,
  getTimelineEventKey,
} from '@/features/simulation/components/timeline/timelineGraph';

function createSimulation(partial: Partial<SimulationInfo> & Pick<SimulationInfo, 'id' | 'running' | 'paused' | 'tick' | 'timestamp'>): SimulationInfo {
  return partial;
}

function createEvent(partial: SimulationEvent): SimulationEvent {
  return partial;
}

function createTerminalRecord(partial: TerminalSimulationRecord): TerminalSimulationRecord {
  return partial;
}

describe('timelineGraph', () => {
  it('orders branch lanes above the base lane by split tick', () => {
    const graph = buildTimelineGraph({
      simulations: [
        createSimulation({ id: 'branch-b', running: false, paused: true, tick: 18, timestamp: '2026-03-12T03:16:05Z', parentId: 'base', splitTick: 14, splitTimestamp: '2026-03-12T03:16:05Z' }),
        createSimulation({ id: 'base', running: true, paused: false, tick: 32, timestamp: '2026-03-12T03:15:05Z', untilTick: 64 }),
        createSimulation({ id: 'branch-a', running: false, paused: true, tick: 12, timestamp: '2026-03-12T03:15:35Z', parentId: 'base', splitTick: 8, splitTimestamp: '2026-03-12T03:15:35Z' }),
      ],
      activeSimulationId: 'branch-a',
      activeState: {
        status: 'running',
        simulationId: 'branch-a',
        isRunnerActive: false,
        isRunnerPaused: true,
        airbases: [],
        aircrafts: [],
        activeThreats: [],
        tick: 12,
        time: '2026-03-12T03:15:35Z',
        history: {},
        playbackTick: null,
        maxTick: 36,
        untilTick: 40,
        parentId: 'base',
        splitTick: 8,
        splitTimestamp: '2026-03-12T03:15:35Z',
      },
      eventsBySimulation: new Map(),
      terminalRecordsBySimulation: new Map(),
    });

    expect(graph.lanes.map((lane) => lane.id)).toEqual(['branch-a', 'branch-b', 'base']);
    expect(graph.lanes[0]).toMatchObject({ startTick: 8, isActive: true });
    expect(graph.lanes[2]).toMatchObject({ startTick: 0, isBase: true });
  });

  it('filters branch events before the split and keeps terminal ticks in lane bounds', () => {
    const graph = buildTimelineGraph({
      simulations: [
        createSimulation({
          id: 'base',
          running: true,
          paused: false,
          tick: 30,
          timestamp: '2026-03-12T03:15:05Z',
          untilTick: 64,
        }),
        createSimulation({
          id: 'branch-a',
          running: false,
          paused: true,
          tick: 20,
          timestamp: '2026-03-12T03:16:05Z',
          parentId: 'base',
          splitTick: 12,
          splitTimestamp: '2026-03-12T03:16:05Z',
        }),
      ],
      activeSimulationId: 'base',
      activeState: {
        status: 'running',
        simulationId: 'base',
        isRunnerActive: true,
        isRunnerPaused: false,
        airbases: [],
        aircrafts: [],
        activeThreats: [],
        tick: 30,
        time: '2026-03-12T03:15:05Z',
        history: {},
        playbackTick: null,
        maxTick: 64,
        untilTick: 64,
      },
      eventsBySimulation: new Map<string, SimulationEvent[]>([
        ['branch-a', [
          createEvent({ type: 'landing_assignment', simulationId: 'branch-a', tick: 10, timestamp: '2026-03-12T03:16:00Z' }),
          createEvent({ type: 'landing_assignment', simulationId: 'branch-a', tick: 16, timestamp: '2026-03-12T03:16:20Z' }),
        ]],
      ]),
      terminalRecordsBySimulation: new Map<string, TerminalSimulationRecord>([
        ['branch-a', createTerminalRecord({
          simulationId: 'branch-a',
          tick: 24,
          timestamp: '2026-03-12T03:16:40Z',
          kind: 'closed',
          reason: 'cancel',
          summary: {
            completedVisitCount: 0,
            totalDurationMs: 0,
            averageDurationMs: null,
          },
        })],
      ]),
    });

    const branchLane = graph.lanes.find((lane) => lane.id === 'branch-a');
    expect(branchLane?.events).toHaveLength(1);
    expect(branchLane?.events[0]?.tick).toBe(16);
    expect(branchLane?.endTick).toBe(24);
  });

  it('builds stable timeline event keys from event metadata', () => {
    expect(getTimelineEventKey({
      type: 'landing_assignment',
      simulationId: 'base',
      tick: 41,
      timestamp: '2026-03-12T03:14:55Z',
    })).toBe('base:landing_assignment:41:2026-03-12T03:14:55Z');
  });

  it('clamps branch playback ticks to the split and clears when it reaches lane end', () => {
    expect(clampTimelineTick(4, 12, 64)).toBe(12);
    expect(clampTimelineTick(32, 12, 64)).toBe(32);
    expect(clampTimelineTick(64, 12, 64)).toBeNull();
  });
});
