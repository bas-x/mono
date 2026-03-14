import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { toast } from 'sonner';

import { useApi } from '@/lib/api';
import { extractErrorMessage, getErrorStatus } from '@/lib/api/errors';
import { createMockSimulationStreamClient } from '@/lib/api/mock/realtime';
import { createSimulationStreamClient } from '@/lib/api/realtime/simulationStream';
import { useSimulationStream } from '@/lib/api/useSimulationStream';
import { SIMULATION_TICKS_PER_SECOND } from '@/lib/api/types';
import type {
  ApiConfig,
  BranchCreatedEvent,
  CreateBaseSimulationRequest,
  SimulationAirbase,
  SimulationAircraft,
  SimulationAircraftNeed,
  SimulationEvent,
  SimulationInfo,
  SimulationStreamClient,
  SourceEvent,
} from '@/lib/api/types';

import {
  durationSecondsToTicks,
  ticksToDurationSeconds,
  type SimulationSetupFormValues,
} from '@/features/simulation/types';

const BASE_SIMULATION_ID = 'base';

function createStandaloneSimulationStreamClient(config: ApiConfig): SimulationStreamClient {
  if (config.mode === 'mock') {
    return createMockSimulationStreamClient();
  }

  return createSimulationStreamClient(config);
}

function parseCsvList(value: string): string[] {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
}

function toPercentRatio(percent: number) {
  return {
    numerator: percent,
    denominator: 100,
  };
}

function buildCreateBaseSimulationRequest(
  values: SimulationSetupFormValues,
): CreateBaseSimulationRequest {
  return {
    seed: values.seedHex || undefined,
    untilTick: durationSecondsToTicks(values.durationSeconds),
    simulationOptions: {
      constellationOpts: {
        includeRegions: parseCsvList(values.includeRegions),
        excludeRegions: parseCsvList(values.excludeRegions),
        minPerRegion: values.minPerRegion,
        maxPerRegion: values.maxPerRegion,
        maxTotal: values.maxTotal,
        regionProbability: toPercentRatio(values.regionProbabilityPercent),
      },
      fleetOpts: {
        aircraftMin: values.aircraftMin,
        aircraftMax: values.aircraftMax,
        needsMin: values.needsMin,
        needsMax: values.needsMax,
        needsPool: values.needsPool,
        severityMin: values.severityMin,
        severityMax: values.severityMax,
        blockingChance: toPercentRatio(values.blockingChancePercent),
      },
    },
  };
}

export function formatSimulationDurationFromTicks(ticks?: number): string | null {
  if (ticks == null || ticks <= 0) {
    return null;
  }

  const seconds = ticksToDurationSeconds(ticks);
  const wholeSeconds = Math.round(seconds);
  const secondsLabel = Math.abs(seconds - wholeSeconds) < 0.05 ? `${wholeSeconds}s` : `${seconds.toFixed(1)}s`;
  return `${secondsLabel} (${ticks} ticks @ ${SIMULATION_TICKS_PER_SECOND}/s)`;
}

function getTimelineEndTick(simulationInfo: Pick<SimulationInfo, 'tick' | 'untilTick'>): number {
  return Math.max(simulationInfo.tick, simulationInfo.untilTick ?? 0);
}

function getLiveTimelineMaxTick(currentTick: number, currentMaxTick?: number, untilTick?: number): number {
  if (untilTick != null) {
    return Math.max(untilTick, currentMaxTick ?? 0);
  }

  return Math.max(currentTick, currentMaxTick ?? 0);
}

function compareSimulationInfo(left: SimulationInfo, right: SimulationInfo): number {
  if (left.id === BASE_SIMULATION_ID) {
    return -1;
  }

  if (right.id === BASE_SIMULATION_ID) {
    return 1;
  }

  const leftTick = left.splitTick ?? Number.MAX_SAFE_INTEGER;
  const rightTick = right.splitTick ?? Number.MAX_SAFE_INTEGER;
  if (leftTick !== rightTick) {
    return leftTick - rightTick;
  }

  const leftTime = left.splitTimestamp ?? left.timestamp;
  const rightTime = right.splitTimestamp ?? right.timestamp;
  if (leftTime !== rightTime) {
    return leftTime.localeCompare(rightTime);
  }

  return left.id.localeCompare(right.id);
}

export function sortSimulationInfos(simulations: SimulationInfo[]): SimulationInfo[] {
  return [...simulations].sort(compareSimulationInfo);
}

export function mergeSimulationInfos(
  current: SimulationInfo[],
  incoming: SimulationInfo[],
): SimulationInfo[] {
  const byId = new Map(current.map((simulation) => [simulation.id, simulation]));

  for (const simulation of incoming) {
    const previous = byId.get(simulation.id);
    byId.set(simulation.id, previous ? { ...previous, ...simulation } : simulation);
  }

  return sortSimulationInfos(Array.from(byId.values()));
}

export function applyBranchCreatedSummary(
  current: SimulationInfo[],
  event: BranchCreatedEvent,
): SimulationInfo[] {
  return mergeSimulationInfos(current, [
    {
      id: event.branchId,
      running: false,
      paused: false,
      tick: event.splitTick,
      timestamp: event.splitTimestamp,
      parentId: event.parentId,
      splitTick: event.splitTick,
      splitTimestamp: event.splitTimestamp,
      sourceEvent: event.sourceEvent,
    },
  ]);
}

export function appendSimulationEvent(
  cache: Map<string, SimulationEvent[]>,
  event: SimulationEvent,
): Map<string, SimulationEvent[]> {
  const next = new Map(cache);
  const currentEvents = next.get(event.simulationId) ?? [];
  next.set(event.simulationId, [...currentEvents, event]);
  return next;
}

export function pruneSimulationEventCache(
  cache: Map<string, SimulationEvent[]>,
  activeSimulationIds: string[],
): Map<string, SimulationEvent[]> {
  const activeIds = new Set(activeSimulationIds);
  return new Map(Array.from(cache.entries()).filter(([simulationId]) => activeIds.has(simulationId)));
}

function getTimelineEventAnchorId(event: SimulationEvent): string {
  if (typeof event.id === 'string' && event.id.length > 0) {
    return event.id;
  }

  const tick = typeof event.tick === 'number' ? event.tick : 'na';
  const ordinal = event.sequence ?? event.timestamp ?? 'event';
  return `${event.simulationId}:${event.type}:${tick}:${ordinal}`;
}

export function buildBranchSourceEvent(event: SimulationEvent): SourceEvent | undefined {
  if (typeof event.tick !== 'number') {
    return undefined;
  }

  return {
    id: getTimelineEventAnchorId(event),
    type: event.type,
    tick: event.tick,
  };
}

export function getSimulationCloseFallback(
  closedSimulationId: string,
  simulations: SimulationInfo[],
): string | null {
  const remaining = simulations.filter((simulation) => simulation.id !== closedSimulationId);
  if (remaining.length === 0) {
    return null;
  }

  const base = remaining.find((simulation) => simulation.id === BASE_SIMULATION_ID);
  return base?.id ?? remaining[0]?.id ?? null;
}

export type AircraftPosition = {
  tailNumber: string;
  position: { x: number; y: number };
  state: string;
  needs: SimulationAircraftNeed[];
};

export type SimulationState =
  | { status: 'idle' }
  | { status: 'creating' }
  | {
      status: 'running';
      simulationId: string;
      isRunnerActive: boolean;
      isRunnerPaused: boolean;
      airbases: SimulationAirbase[];
      aircrafts: SimulationAircraft[];
      tick?: number;
      time?: string;
      aircraftPositions?: AircraftPosition[];
      history: Record<number, { aircraftPositions?: AircraftPosition[]; aircrafts: SimulationAircraft[] }>;
      playbackTick?: number | null;
      maxTick?: number;
      untilTick?: number;
      parentId?: string;
      splitTick?: number;
      splitTimestamp?: string;
      sourceEvent?: SourceEvent;
    }
  | { status: 'error'; message: string };

type RunningSimulationState = Extract<SimulationState, { status: 'running' }>;

export function rehydrateRunningSimulationState(
  simulationInfo: SimulationInfo,
  airbases: SimulationAirbase[],
  aircrafts: SimulationAircraft[],
  cachedState?: RunningSimulationState,
): RunningSimulationState {
  return {
    status: 'running',
    simulationId: simulationInfo.id,
    isRunnerActive: simulationInfo.running,
    isRunnerPaused: simulationInfo.paused,
    airbases,
    aircrafts,
    tick: simulationInfo.tick,
    time: simulationInfo.timestamp,
    aircraftPositions: cachedState?.aircraftPositions,
    history: cachedState?.history ?? {},
    playbackTick: cachedState?.playbackTick ?? null,
    maxTick: Math.max(cachedState?.maxTick ?? 0, getTimelineEndTick(simulationInfo)),
    untilTick: simulationInfo.untilTick,
    parentId: simulationInfo.parentId,
    splitTick: simulationInfo.splitTick,
    splitTimestamp: simulationInfo.splitTimestamp,
    sourceEvent: simulationInfo.sourceEvent,
  };
}

function isBranchCreatedEvent(event: SimulationEvent): event is BranchCreatedEvent {
  return event.type === 'branch_created'
    && typeof event.branchId === 'string'
    && typeof event.parentId === 'string'
    && typeof event.splitTick === 'number'
    && typeof event.splitTimestamp === 'string';
}

export function useSimulation() {
  const { clients, config } = useApi();
  const [state, setState] = useState<SimulationState>({ status: 'idle' });
  const [simulations, setSimulations] = useState<SimulationInfo[]>([]);
  const [isLoadingSimulations, setIsLoadingSimulations] = useState(false);
  const simulationStateCacheRef = useRef<Map<string, RunningSimulationState>>(new Map());
  const eventsBySimulationRef = useRef<Map<string, SimulationEvent[]>>(new Map());
  const streamClientsRef = useRef<
    Map<string, { client: SimulationStreamClient; unsubscribe: () => void; unsubscribeState: () => void }>
  >(new Map());
  const [, setEventVersion] = useState(0);

  const activeSimulationId = state.status === 'running' ? state.simulationId : undefined;
  const stream = useSimulationStream(activeSimulationId);
  const shouldTrackBaseStream = useMemo(() => {
    return simulations.some((simulation) => simulation.id === BASE_SIMULATION_ID) || state.status === 'running';
  }, [simulations, state.status]);
  const baseStream = useSimulationStream(shouldTrackBaseStream ? BASE_SIMULATION_ID : undefined);
  const currentEvents = state.status === 'running'
    ? (eventsBySimulationRef.current.get(state.simulationId) ?? [])
    : [];

  const recordSimulationEvent = useCallback((event: SimulationEvent) => {
    eventsBySimulationRef.current = appendSimulationEvent(eventsBySimulationRef.current, event);
    setEventVersion((version) => version + 1);
  }, []);

  const fetchSimulations = useCallback(async () => {
    setIsLoadingSimulations(true);
    try {
      const list = sortSimulationInfos(await clients.simulation.getSimulations());
      setSimulations(list);
      return list;
    } catch (error) {
      console.error('Failed to fetch simulations', error);
      return [];
    } finally {
      setIsLoadingSimulations(false);
    }
  }, [clients.simulation]);

  const loadSimulation = useCallback(async (id: string) => {
    setState({ status: 'creating' });
    try {
      const [simulationInfo, airbases, aircrafts] = await Promise.all([
        clients.simulation.getSimulation(id),
        clients.simulation.getAirbases(id),
        clients.simulation.getAircrafts(id),
      ]);

      const cachedState = simulationStateCacheRef.current.get(id);
      setSimulations((current) => mergeSimulationInfos(current, [simulationInfo]));
      setState(rehydrateRunningSimulationState(simulationInfo, airbases, aircrafts, cachedState));
      toast.success('Simulation loaded successfully');
    } catch (error) {
      const errorMessage = extractErrorMessage(error);
      toast.error(errorMessage);
      setState({
        status: 'error',
        message: errorMessage,
      });
    }
  }, [clients.simulation]);

  const handleSimulationClosed = useCallback(async (closedSimulationId: string) => {
    simulationStateCacheRef.current.delete(closedSimulationId);
    eventsBySimulationRef.current.delete(closedSimulationId);
    setSimulations((current) => sortSimulationInfos(current.filter((simulation) => simulation.id !== closedSimulationId)));
    const list = await fetchSimulations();

    if (state.status === 'running' && state.simulationId === closedSimulationId) {
      const fallbackSimulationId = getSimulationCloseFallback(closedSimulationId, list);
      if (fallbackSimulationId) {
        await loadSimulation(fallbackSimulationId);
      } else {
        setState({ status: 'idle' });
      }
    }
  }, [fetchSimulations, loadSimulation, state]);

  useEffect(() => {
    if (state.status === 'running') {
      simulationStateCacheRef.current.set(state.simulationId, state);
    }
  }, [state]);

  useEffect(() => {
    const activeSimulationIds = simulations.map((simulation) => simulation.id);
    const activeIdSet = new Set(activeSimulationIds);

    for (const simulationId of activeSimulationIds) {
      if (streamClientsRef.current.has(simulationId)) {
        continue;
      }

      const client = createStandaloneSimulationStreamClient(config);
      const unsubscribe = client.subscribe((event) => {
        recordSimulationEvent(event);
      });
      const unsubscribeState = client.onConnectionStateChange(() => {});
      client.connect(simulationId);
      streamClientsRef.current.set(simulationId, { client, unsubscribe, unsubscribeState });
    }

    for (const [simulationId, streamRecord] of Array.from(streamClientsRef.current.entries())) {
      if (activeIdSet.has(simulationId)) {
        continue;
      }

      streamRecord.unsubscribe();
      streamRecord.unsubscribeState();
      streamRecord.client.disconnect(1000, 'simulation removed');
      streamClientsRef.current.delete(simulationId);
    }

    eventsBySimulationRef.current = pruneSimulationEventCache(
      eventsBySimulationRef.current,
      activeSimulationIds,
    );
  }, [config, recordSimulationEvent, simulations]);

  useEffect(() => {
    return () => {
      for (const streamRecord of streamClientsRef.current.values()) {
        streamRecord.unsubscribe();
        streamRecord.unsubscribeState();
        streamRecord.client.disconnect(1000, 'simulation hook unmounted');
      }
      streamClientsRef.current.clear();
    };
  }, []);

  useEffect(() => {
    fetchSimulations().catch(() => {});
  }, [fetchSimulations]);

  useEffect(() => {
    if (baseStream.state === 'open') {
      fetchSimulations().catch(() => {});
    }
  }, [baseStream.state, fetchSimulations]);

  useEffect(() => {
    if (!shouldTrackBaseStream) {
      return;
    }

    return baseStream.subscribe((event: SimulationEvent) => {
      if (!isBranchCreatedEvent(event)) {
        return;
      }

      setSimulations((current) => applyBranchCreatedSummary(current, event));
    });
  }, [baseStream, shouldTrackBaseStream]);

  useEffect(() => {
    if (state.status !== 'running') {
      return;
    }

    return stream.subscribe((event: SimulationEvent) => {
      if (isBranchCreatedEvent(event)) {
        setSimulations((current) => applyBranchCreatedSummary(current, event));
        return;
      }

      if (event.type === 'simulation_closed') {
        void handleSimulationClosed(event.simulationId);
        return;
      }

      if (event.type === 'simulation_step') {
        setState((current) => {
          if (current.status !== 'running') return current;
          const currentTick = event.tick as number;
          return {
            ...current,
            tick: currentTick,
            maxTick: getLiveTimelineMaxTick(currentTick, current.maxTick, current.untilTick),
            time: event.timestamp,
            history: {
              ...current.history,
              [currentTick]: current.history[currentTick] || {
                aircrafts: current.aircrafts,
                aircraftPositions: current.aircraftPositions,
              },
            },
          };
        });
      } else if (event.type === 'aircraft_state_change') {
        setState((current) => {
          if (current.status !== 'running') return current;
          const updatedAircrafts = current.aircrafts.map((aircraft) =>
            aircraft.tailNumber === event.tailNumber ? { ...aircraft, ...event.aircraft } : aircraft,
          );
          const currentTick = current.tick ?? 0;
          return {
            ...current,
            aircrafts: updatedAircrafts,
            history: {
              ...current.history,
              [currentTick]: { ...current.history[currentTick], aircrafts: updatedAircrafts },
            },
          };
        });
      } else if (event.type === 'landing_assignment') {
        setState((current) => {
          if (current.status !== 'running') return current;
          const updatedAircrafts = current.aircrafts.map((aircraft) =>
            aircraft.tailNumber === event.tailNumber ? { ...aircraft, assignedTo: event.baseId } : aircraft,
          );
          const currentTick = current.tick ?? 0;
          return {
            ...current,
            aircrafts: updatedAircrafts,
            history: {
              ...current.history,
              [currentTick]: { ...current.history[currentTick], aircrafts: updatedAircrafts },
            },
          };
        });
      } else if (event.type === 'all_aircraft_positions') {
        setState((current) => {
          if (current.status !== 'running') return current;
          const currentTick = event.tick ?? current.tick ?? 0;
          return {
            ...current,
            aircraftPositions: event.positions,
            history: {
              ...current.history,
              [currentTick]: { ...current.history[currentTick], aircraftPositions: event.positions },
            },
          };
        });
      } else if (event.type === 'simulation_ended') {
        setState((current) => {
          if (current.status !== 'running') return current;
          const endedTick = event.tick as number;
          const boundedMaxTick = Math.max(endedTick, current.untilTick ?? 0, current.maxTick ?? 0);
          return {
            ...current,
            isRunnerActive: false,
            isRunnerPaused: false,
            tick: endedTick,
            maxTick: boundedMaxTick,
            time: event.timestamp,
            playbackTick:
              current.playbackTick == null ? null : Math.min(current.playbackTick, boundedMaxTick),
          };
        });
      }
    });
  }, [handleSimulationClosed, state.status, stream]);

  const createSimulation = useCallback(async (values: SimulationSetupFormValues): Promise<boolean> => {
    setState({ status: 'creating' });
    try {
      const { id } = await clients.simulation.createBaseSimulation(
        buildCreateBaseSimulationRequest(values),
      );
      toast.success('Simulation created successfully');
      await fetchSimulations();
      await loadSimulation(id);
      return true;
    } catch (error: unknown) {
      const errorMessage = extractErrorMessage(error);
      const statusCode = getErrorStatus(error);

      toast.error(errorMessage);
      setState({
        status: 'error',
        message: errorMessage,
      });

      if (statusCode === 409) {
        await fetchSimulations();
      }
      return false;
    }
  }, [clients.simulation, fetchSimulations, loadSimulation]);

  const createBranchFromEvent = useCallback(async (event: SimulationEvent): Promise<boolean> => {
    if (state.status !== 'running' || state.simulationId !== BASE_SIMULATION_ID) {
      return false;
    }

    try {
      const sourceEvent = buildBranchSourceEvent(event);
      const branchInfo = await clients.simulation.createBranchSimulation(BASE_SIMULATION_ID, sourceEvent ? {
        sourceEvent,
      } : undefined);

      setSimulations((current) => mergeSimulationInfos(current, [branchInfo]));
      toast.success(`Created branch ${branchInfo.id.slice(0, 8)}`);
      await loadSimulation(branchInfo.id);
      return true;
    } catch (error) {
      toast.error(extractErrorMessage(error));
      return false;
    }
  }, [clients.simulation, loadSimulation, state]);

  const refreshData = useCallback(async () => {
    if (state.status !== 'running') return;

    try {
      const [simulationInfo, airbases, aircrafts] = await Promise.all([
        clients.simulation.getSimulation(state.simulationId),
        clients.simulation.getAirbases(state.simulationId),
        clients.simulation.getAircrafts(state.simulationId),
      ]);

      setSimulations((current) => mergeSimulationInfos(current, [simulationInfo]));
      setState((current) => {
        if (current.status !== 'running') return current;
        return {
          ...current,
          isRunnerActive: simulationInfo.running,
          isRunnerPaused: simulationInfo.paused,
          airbases,
          aircrafts,
          tick: simulationInfo.tick,
          time: simulationInfo.timestamp,
          maxTick: Math.max(current.maxTick ?? 0, getTimelineEndTick(simulationInfo)),
          untilTick: simulationInfo.untilTick ?? current.untilTick,
          parentId: simulationInfo.parentId,
          splitTick: simulationInfo.splitTick,
          splitTimestamp: simulationInfo.splitTimestamp,
          sourceEvent: simulationInfo.sourceEvent,
        };
      });
    } catch (error) {
      console.error('Failed to refresh simulation data', error);
    }
  }, [clients.simulation, state]);

  const reset = useCallback(() => {
    setState({ status: 'idle' });
  }, []);

  const triggerReset = useCallback(async () => {
    if (state.status !== 'running') return;

      const resetSimulationId = state.simulationId;

    try {
      await clients.simulation.resetSimulation(resetSimulationId);
      simulationStateCacheRef.current.delete(resetSimulationId);
      eventsBySimulationRef.current.delete(resetSimulationId);
      const list = await fetchSimulations();
      const fallbackSimulationId = getSimulationCloseFallback(resetSimulationId, list);

      if (fallbackSimulationId) {
        await loadSimulation(fallbackSimulationId);
      } else {
        setState({ status: 'idle' });
      }

      toast.success('Simulation reset successfully', {
        action: {
          label: 'Undo',
          onClick: () => console.log('Undo reset not yet implemented on backend'),
        },
      });
    } catch (error) {
      const errorMessage = extractErrorMessage(error);
      toast.error(errorMessage);
    }
  }, [clients.simulation, fetchSimulations, loadSimulation, state]);

  const setPlaybackTick = useCallback((tick: number | null) => {
    setState((current) => {
      if (current.status !== 'running') return current;
      return { ...current, playbackTick: tick };
    });
  }, []);

  const visibleState = state.status === 'running'
    ? {
        ...state,
        aircrafts: state.playbackTick != null && state.history[state.playbackTick]
          ? state.history[state.playbackTick].aircrafts
          : state.aircrafts,
        aircraftPositions: state.playbackTick != null && state.history[state.playbackTick]
          ? state.history[state.playbackTick].aircraftPositions
          : state.aircraftPositions,
      }
    : state;

  return {
    state: visibleState,
    events: currentEvents,
    setPlaybackTick,
    simulations,
    isLoadingSimulations,
    loadSimulation,
    createSimulation,
    createBranchFromEvent,
    refreshData,
    triggerReset,
    reset,
  };
}
