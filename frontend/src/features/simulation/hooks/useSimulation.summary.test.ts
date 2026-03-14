import { describe, expect, it } from 'vitest';

import {
  createTerminalSimulationRecord,
  formatTerminalSummaryDuration,
  formatTerminalSummaryHeadline,
  sortTerminalSimulationRecords,
} from '@/features/simulation/hooks/useSimulation';

describe('terminal summary helpers', () => {
  it('creates natural terminal records from simulation_ended events', () => {
    const record = createTerminalSimulationRecord({
      type: 'simulation_ended',
      simulationId: 'base',
      tick: 16,
      timestamp: '2026-03-12T03:15:05Z',
      summary: {
        completedVisitCount: 0,
        totalDurationMs: 0,
        averageDurationMs: null,
      },
    });

    expect(record).toEqual({
      simulationId: 'base',
      tick: 16,
      timestamp: '2026-03-12T03:15:05Z',
      kind: 'ended',
      reason: undefined,
      summary: {
        completedVisitCount: 0,
        totalDurationMs: 0,
        averageDurationMs: null,
      },
    });
  });

  it('creates closed terminal records with reset and cancel reasons', () => {
    const resetRecord = createTerminalSimulationRecord({
      type: 'simulation_closed',
      simulationId: 'base',
      tick: 9,
      timestamp: '2026-03-12T03:16:05Z',
      reason: 'reset',
      summary: {
        completedVisitCount: 1,
        totalDurationMs: 5000,
        averageDurationMs: 5000,
      },
    });
    const cancelRecord = createTerminalSimulationRecord({
      type: 'simulation_closed',
      simulationId: 'branch-1',
      tick: 9,
      timestamp: '2026-03-12T03:16:05Z',
      reason: 'cancel',
      summary: {
        completedVisitCount: 0,
        totalDurationMs: 0,
        averageDurationMs: null,
      },
    });

    expect(resetRecord.reason).toBe('reset');
    expect(cancelRecord.reason).toBe('cancel');
    expect(formatTerminalSummaryHeadline(resetRecord)).toBe('Run stopped and reset');
    expect(formatTerminalSummaryHeadline(cancelRecord)).toBe('Branch stopped');
  });

  it('formats nullable average durations distinctly from zero durations', () => {
    expect(formatTerminalSummaryDuration(null)).toBe('No completed services yet');
    expect(formatTerminalSummaryDuration(0)).toBe('0 ms');
    expect(formatTerminalSummaryDuration(5000)).toBe('5.0 s');
  });

  it('keeps the terminal record bound to the simulation that emitted it', () => {
    const record = createTerminalSimulationRecord({
      type: 'simulation_closed',
      simulationId: 'branch-42',
      tick: 22,
      timestamp: '2026-03-12T03:16:05Z',
      reason: 'cancel',
      summary: {
        completedVisitCount: 3,
        totalDurationMs: 9000,
        averageDurationMs: 3000,
      },
    });

    expect(record.simulationId).toBe('branch-42');
    expect(record.kind).toBe('closed');
    expect(record.reason).toBe('cancel');
  });

  it('sorts terminal records with base first and newer branch summaries next', () => {
    const sorted = sortTerminalSimulationRecords([
      createTerminalSimulationRecord({
        type: 'simulation_closed',
        simulationId: 'branch-b',
        tick: 10,
        timestamp: '2026-03-12T03:14:05Z',
        reason: 'cancel',
        summary: {
          completedVisitCount: 0,
          totalDurationMs: 0,
          averageDurationMs: null,
        },
      }),
      createTerminalSimulationRecord({
        type: 'simulation_ended',
        simulationId: 'base',
        tick: 16,
        timestamp: '2026-03-12T03:15:05Z',
        summary: {
          completedVisitCount: 1,
          totalDurationMs: 5000,
          averageDurationMs: 5000,
        },
      }),
      createTerminalSimulationRecord({
        type: 'simulation_closed',
        simulationId: 'branch-a',
        tick: 11,
        timestamp: '2026-03-12T03:16:05Z',
        reason: 'cancel',
        summary: {
          completedVisitCount: 2,
          totalDurationMs: 8000,
          averageDurationMs: 4000,
        },
      }),
    ]);

    expect(sorted.map((record) => record.simulationId)).toEqual(['base', 'branch-a', 'branch-b']);
  });
});
