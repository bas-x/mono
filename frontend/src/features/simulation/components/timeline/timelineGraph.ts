import type { SimulationEvent, SimulationInfo, SourceEvent } from '@/lib/api/types';

import type { SimulationState, TerminalSimulationRecord } from '../../hooks/useSimulation';

export type TimelineLane = {
  id: string;
  label: string;
  shortLabel: string;
  isBase: boolean;
  isActive: boolean;
  startTick: number;
  endTick: number;
  splitTick?: number;
  splitTimestamp?: string;
  sourceEvent?: SourceEvent;
  events: SimulationEvent[];
  terminalRecord?: TerminalSimulationRecord | null;
};

export type TimelineGraph = {
  lanes: TimelineLane[];
  globalMaxTick: number;
};

type BuildTimelineGraphInput = {
  simulations: SimulationInfo[];
  activeSimulationId?: string;
  activeState?: Extract<SimulationState, { status: 'running' }>;
  eventsBySimulation: Map<string, SimulationEvent[]>;
  terminalRecordsBySimulation: Map<string, TerminalSimulationRecord>;
};

export function getTimelineEventKey(event: SimulationEvent): string {
  if (typeof event.id === 'string' && event.id.length > 0) {
    return event.id;
  }

  const tick = typeof event.tick === 'number' ? event.tick : 'na';
  const ordinal = event.sequence ?? event.timestamp ?? 'event';
  return `${event.simulationId}:${event.type}:${tick}:${ordinal}`;
}

function getHighestEventTick(events: SimulationEvent[]): number {
  return events.reduce((highest, event) => {
    const tick = typeof event.tick === 'number' ? event.tick : 0;
    return Math.max(highest, tick);
  }, 0);
}

function getLaneEndTick(
  simulation: SimulationInfo,
  events: SimulationEvent[],
  terminalRecord?: TerminalSimulationRecord | null,
  activeState?: Extract<SimulationState, { status: 'running' }>,
): number {
  return Math.max(
    simulation.tick,
    simulation.untilTick ?? 0,
    getHighestEventTick(events),
    terminalRecord?.tick ?? 0,
    activeState?.tick ?? 0,
    activeState?.maxTick ?? 0,
    activeState?.untilTick ?? 0,
  );
}

function getLaneLabel(simulationId: string): { label: string; shortLabel: string } {
  if (simulationId === 'base') {
    return {
      label: 'Base timeline',
      shortLabel: 'Base',
    };
  }

  const shortId = simulationId.slice(0, 8);
  return {
    label: `Branch ${shortId}`,
    shortLabel: shortId,
  };
}

export function clampTimelineTick(tick: number | null, minTick: number, maxTick: number): number | null {
  if (tick == null) {
    return null;
  }

  const boundedTick = Math.max(minTick, Math.min(maxTick, tick));
  return boundedTick >= maxTick ? null : boundedTick;
}

export function buildTimelineGraph({
  simulations,
  activeSimulationId,
  activeState,
  eventsBySimulation,
  terminalRecordsBySimulation,
}: BuildTimelineGraphInput): TimelineGraph {
  const lanes = simulations
    .map<TimelineLane>((simulation) => {
      const events = (eventsBySimulation.get(simulation.id) ?? [])
        .filter((event) => typeof event.tick !== 'number' || event.tick >= (simulation.splitTick ?? 0));
      const terminalRecord = terminalRecordsBySimulation.get(simulation.id) ?? null;
      const laneActiveState = activeState?.simulationId === simulation.id ? activeState : undefined;
      const labels = getLaneLabel(simulation.id);

      return {
        id: simulation.id,
        label: labels.label,
        shortLabel: labels.shortLabel,
        isBase: simulation.id === 'base',
        isActive: simulation.id === activeSimulationId,
        startTick: simulation.id === 'base' ? 0 : Math.max(0, simulation.splitTick ?? 0),
        endTick: getLaneEndTick(simulation, events, terminalRecord, laneActiveState),
        splitTick: simulation.splitTick,
        splitTimestamp: simulation.splitTimestamp,
        sourceEvent: simulation.sourceEvent,
        events,
        terminalRecord,
      };
    })
    .sort((left, right) => {
      if (left.isBase) {
        return 1;
      }

      if (right.isBase) {
        return -1;
      }

      if ((left.splitTick ?? 0) !== (right.splitTick ?? 0)) {
        return (left.splitTick ?? 0) - (right.splitTick ?? 0);
      }

      return left.id.localeCompare(right.id);
    });

  const globalMaxTick = Math.max(0, ...lanes.map((lane) => lane.endTick));

  return {
    lanes,
    globalMaxTick,
  };
}
