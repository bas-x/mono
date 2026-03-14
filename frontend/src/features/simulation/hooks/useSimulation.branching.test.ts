import { describe, expect, it } from 'vitest';

import {
  applyBranchCreatedSummary,
  appendSimulationEvent,
  buildBranchSourceEvent,
  getSimulationCloseFallback,
  mergeSimulationInfos,
  pruneSimulationEventCache,
  rehydrateRunningSimulationState,
} from '@/features/simulation/hooks/useSimulation';

describe('useSimulation branching helpers', () => {
  it('merges simulation summaries with branch metadata', () => {
    const merged = mergeSimulationInfos(
      [
        {
          id: 'base',
          running: true,
          paused: false,
          tick: 42,
          timestamp: '2026-03-12T03:15:05Z',
          untilTick: 96,
        },
      ],
      [
        {
          id: '7f3c2d1a9b8e6f10',
          running: false,
          paused: true,
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
        },
      ],
    );

    expect(merged).toHaveLength(2);
    expect(merged[1]).toMatchObject({
      id: '7f3c2d1a9b8e6f10',
      parentId: 'base',
      splitTick: 42,
      splitTimestamp: '2026-03-12T03:15:05Z',
      sourceEvent: {
        id: 'timeline-evt-17',
        type: 'landing_assignment',
        tick: 41,
      },
    });
  });

  it('dedupes branch summaries when branch_created arrives after local insert', () => {
    const merged = applyBranchCreatedSummary(
      [
        {
          id: 'base',
          running: true,
          paused: false,
          tick: 42,
          timestamp: '2026-03-12T03:15:05Z',
        },
        {
          id: '7f3c2d1a9b8e6f10',
          running: false,
          paused: true,
          tick: 42,
          timestamp: '2026-03-12T03:15:05Z',
          parentId: 'base',
        },
      ],
      {
        type: 'branch_created',
        simulationId: 'base',
        timestamp: '2026-03-12T03:15:05Z',
        branchId: '7f3c2d1a9b8e6f10',
        parentId: 'base',
        splitTick: 42,
        splitTimestamp: '2026-03-12T03:15:05Z',
      },
    );

    expect(merged.filter((simulation) => simulation.id === '7f3c2d1a9b8e6f10')).toHaveLength(1);
    expect(merged[1]).toMatchObject({
      splitTick: 42,
      splitTimestamp: '2026-03-12T03:15:05Z',
    });
  });

  it('builds stable source event metadata from timeline events', () => {
    const sourceEvent = buildBranchSourceEvent({
      type: 'landing_assignment',
      simulationId: 'base',
      tick: 41,
      timestamp: '2026-03-12T03:14:55Z',
    });

    expect(sourceEvent).toEqual({
      id: 'base:landing_assignment:41:2026-03-12T03:14:55Z',
      type: 'landing_assignment',
      tick: 41,
    });
  });

  it('falls back to base when the selected branch closes', () => {
    const fallback = getSimulationCloseFallback('branch-00000001', [
      {
        id: 'base',
        running: true,
        paused: false,
        tick: 42,
        timestamp: '2026-03-12T03:15:05Z',
      },
      {
        id: 'branch-00000001',
        running: false,
        paused: true,
        tick: 42,
        timestamp: '2026-03-12T03:15:05Z',
        parentId: 'base',
      },
    ]);

    expect(fallback).toBe('base');
  });

  it('preserves branch replay history when switching back to a branch', () => {
    const restored = rehydrateRunningSimulationState(
      {
        id: 'branch-00000001',
        running: false,
        paused: true,
        tick: 52,
        timestamp: '2026-03-12T03:20:05Z',
        untilTick: 96,
        parentId: 'base',
        splitTick: 42,
        splitTimestamp: '2026-03-12T03:15:05Z',
      },
      [],
      [],
      {
        status: 'running',
        simulationId: 'branch-00000001',
        isRunnerActive: false,
        isRunnerPaused: true,
        airbases: [],
        aircrafts: [],
        tick: 48,
        time: '2026-03-12T03:18:05Z',
        aircraftPositions: [{ tailNumber: 'BX-101', position: { x: 1, y: 2 }, state: 'Ready', needs: [] }],
        history: {
          41: { aircrafts: [], aircraftPositions: [] },
          48: { aircrafts: [], aircraftPositions: [{ tailNumber: 'BX-101', position: { x: 1, y: 2 }, state: 'Ready', needs: [] }] },
        },
        playbackTick: 41,
        maxTick: 80,
        untilTick: 96,
        parentId: 'base',
        splitTick: 42,
        splitTimestamp: '2026-03-12T03:15:05Z',
      },
    );

    expect(restored.history[41]).toBeDefined();
    expect(restored.playbackTick).toBe(41);
    expect(restored.maxTick).toBe(96);
    expect(restored.aircraftPositions?.[0]?.tailNumber).toBe('BX-101');
  });

  it('retains events per branch while pruning removed branches', () => {
    const cache = appendSimulationEvent(
      appendSimulationEvent(new Map(), {
        type: 'simulation_step',
        simulationId: 'base',
        tick: 42,
        timestamp: '2026-03-12T03:15:05Z',
      }),
      {
        type: 'landing_assignment',
        simulationId: 'branch-00000001',
        tick: 45,
        timestamp: '2026-03-12T03:18:05Z',
        tailNumber: 'BX-101',
      },
    );

    const pruned = pruneSimulationEventCache(cache, ['base', 'branch-00000001']);

    expect(pruned.get('base')).toHaveLength(1);
    expect(pruned.get('branch-00000001')).toHaveLength(1);
    expect(pruneSimulationEventCache(pruned, ['base']).has('branch-00000001')).toBe(false);
  });
});
