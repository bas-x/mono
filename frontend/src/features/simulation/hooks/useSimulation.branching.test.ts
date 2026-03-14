import { describe, expect, it } from 'vitest';

import {
  applyAircraftAssignment,
  applyBranchCreatedSummary,
  applyOverrideResponse,
  appendSimulationEvent,
  buildHistorySnapshot,
  buildBranchSourceEvent,
  getSimulationCloseFallback,
  mapOverrideErrorMessage,
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

  it('keeps the latest landing assignment source for repeated events', () => {
    const afterAlgorithm = applyAircraftAssignment(
      [{ tailNumber: 'BX-101', needs: [], state: 'Inbound' }],
      'BX-101',
      { base: 'base-a', source: 'algorithm' },
    );
    const afterHuman = applyAircraftAssignment(afterAlgorithm, 'BX-101', {
      base: 'base-b',
      source: 'human',
    });

    expect(afterHuman[0]).toMatchObject({
      assignedTo: 'base-b',
      assignmentSource: 'human',
    });
  });

  it('applies override responses synchronously before websocket reconciliation', () => {
    const updated = applyOverrideResponse(
      [{ tailNumber: 'BX-101', needs: [], state: 'Inbound', assignedTo: 'base-a' }],
      {
        aircraft: {
          tailNumber: 'BX-101',
          needs: [],
          state: 'Inbound',
          assignedTo: 'base-b',
        },
        assignment: {
          base: 'base-b',
          source: 'human',
        },
      },
    );

    expect(updated[0]).toMatchObject({
      assignedTo: 'base-b',
      assignmentSource: 'human',
    });
  });

  it('maps override HTTP errors to specific operator messages', () => {
    expect(mapOverrideErrorMessage({ status: 409, body: '{"message":"assignment override too late"}' })).toBe('Override too late');
    expect(mapOverrideErrorMessage({ status: 404, body: '{"message":"simulation or aircraft not found"}' })).toBe('Simulation or aircraft no longer exists');
    expect(mapOverrideErrorMessage({ status: 400, body: '{"message":"invalid base"}' })).toBe('Invalid assignment target');
  });

  it('fills aircrafts when building a history snapshot from positions-only updates', () => {
    const snapshot = buildHistorySnapshot({
      status: 'running',
      simulationId: 'base',
      isRunnerActive: false,
      isRunnerPaused: true,
      airbases: [],
      aircrafts: [{ tailNumber: 'BX-101', needs: [], state: 'Inbound' }],
      tick: 12,
      time: '2026-03-12T03:15:05Z',
      aircraftPositions: [],
      history: {},
      playbackTick: 12,
      maxTick: 12,
      untilTick: 24,
    }, 12, { aircraftPositions: [] });

    expect(snapshot.aircrafts).toHaveLength(1);
    expect(snapshot.aircrafts[0]?.tailNumber).toBe('BX-101');
  });
});
