import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { toast } from 'sonner';

import { useApi } from '@/lib/api';
import { extractErrorMessage, getErrorStatus } from '@/lib/api/errors';
import { createMockSimulationStreamClient } from '@/lib/api/mock/realtime';
import { createSimulationStreamClient } from '@/lib/api/realtime/simulationStream';
import { useSimulationStream } from '@/lib/api/useSimulationStream';
import { SIMULATION_TICKS_PER_SECOND } from '@/lib/api/types';
import type {
  Assignment,
  ApiConfig,
  BranchCreatedEvent,
  CreateBaseSimulationRequest,
  ServicingSummary,
  LandingAssignmentEvent,
  OverrideAssignmentResponse,
  SimulationAirbase,
  SimulationAircraft,
  SimulationClosedEvent,
  SimulationClosedReason,
  SimulationEndedEvent,
  SimulationAircraftNeed,
  SimulationEvent,
  SimulationInfo,
  SimulationThreat,
  SimulationStreamClient,
  SourceEvent,
  ThreatDespawnedEvent,
  ThreatSpawnedEvent,
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

export function applyAircraftAssignment(
  aircrafts: SimulationAircraft[],
  tailNumber: string,
  assignment: Assignment,
  needs?: SimulationAircraftNeed[],
): SimulationAircraft[] {
  return aircrafts.map((aircraft) => (
    aircraft.tailNumber === tailNumber
      ? {
          ...aircraft,
          assignedTo: assignment.base,
          assignmentSource: assignment.source,
          needs: needs ?? aircraft.needs,
        }
      : aircraft
  ));
}

export function applyOverrideResponse(
  aircrafts: SimulationAircraft[],
  response: OverrideAssignmentResponse,
): SimulationAircraft[] {
  return aircrafts.map((aircraft) => (
    aircraft.tailNumber === response.aircraft.tailNumber
      ? {
          ...aircraft,
          ...response.aircraft,
          assignmentSource: response.assignment.source,
          assignedTo: response.assignment.base,
        }
      : aircraft
  ));
}

function isLandingAssignmentEvent(event: SimulationEvent): event is LandingAssignmentEvent {
  return event.type === 'landing_assignment'
    && typeof event.tailNumber === 'string'
    && typeof event.baseId === 'string'
    && (event.source === 'algorithm' || event.source === 'human')
    && Array.isArray(event.needs)
    && typeof event.capabilities === 'object'
    && event.capabilities !== null;
}

export function mapOverrideErrorMessage(error: unknown): string {
  const status = getErrorStatus(error);

  switch (status) {
    case 409:
      return 'Override too late';
    case 404:
      return 'Simulation or aircraft no longer exists';
    case 400:
      return 'Invalid assignment target';
    default:
      return extractErrorMessage(error);
  }
}

export type AircraftPosition = {
  tailNumber: string;
  position: { x: number; y: number };
  state: string;
  needs: SimulationAircraftNeed[];
};

export type TerminalSimulationRecord = {
  simulationId: string;
  tick: number;
  timestamp: string;
  kind: 'ended' | 'closed';
  reason?: SimulationClosedReason;
  summary: ServicingSummary;
};

function upsertThreat(threats: SimulationThreat[], threat: SimulationThreat): SimulationThreat[] {
  const existingIndex = threats.findIndex((candidate) => candidate.id === threat.id);
  if (existingIndex === -1) {
    return [...threats, threat];
  }

  return threats.map((candidate) => (candidate.id === threat.id ? threat : candidate));
}

function removeThreat(threats: SimulationThreat[], threatId: string): SimulationThreat[] {
  return threats.filter((candidate) => candidate.id !== threatId);
}

function isSimulationThreat(value: unknown): value is SimulationThreat {
  return typeof value === 'object'
    && value !== null
    && typeof (value as SimulationThreat).id === 'string'
    && typeof (value as SimulationThreat).position?.x === 'number'
    && typeof (value as SimulationThreat).position?.y === 'number'
    && typeof (value as SimulationThreat).createdAt === 'string'
    && typeof (value as SimulationThreat).createdTick === 'number';
}

function isThreatSpawnedEvent(event: SimulationEvent): event is ThreatSpawnedEvent {
  return event.type === 'threat_spawned'
    && typeof event.tick === 'number'
    && isSimulationThreat(event.threat);
}

function isThreatDespawnedEvent(event: SimulationEvent): event is ThreatDespawnedEvent {
  return event.type === 'threat_despawned'
    && typeof event.tick === 'number'
    && isSimulationThreat(event.threat);
}

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
      activeThreats: SimulationThreat[];
      tick?: number;
      time?: string;
      aircraftPositions?: AircraftPosition[];
      history: Record<number, {
        aircraftPositions?: AircraftPosition[];
        aircrafts: SimulationAircraft[];
        activeThreats: SimulationThreat[];
      }>;
      playbackTick?: number | null;
      maxTick?: number;
      untilTick?: number;
      parentId?: string;
      splitTick?: number;
      splitTimestamp?: string;
      sourceEvent?: SourceEvent;
      terminalRecord?: TerminalSimulationRecord;
    }
  | { status: 'error'; message: string };

type RunningSimulationState = Extract<SimulationState, { status: 'running' }>;

export function buildHistorySnapshot(
  current: RunningSimulationState,
  tick: number,
  overrides: Partial<RunningSimulationState['history'][number]> = {},
): RunningSimulationState['history'][number] {
  const existingSnapshot = current.history[tick];

  return {
    aircrafts: overrides.aircrafts ?? existingSnapshot?.aircrafts ?? current.aircrafts,
    aircraftPositions:
      overrides.aircraftPositions ?? existingSnapshot?.aircraftPositions ?? current.aircraftPositions,
    activeThreats: overrides.activeThreats ?? existingSnapshot?.activeThreats ?? current.activeThreats,
  };
}

export function getVisibleAircraftsForPlayback(state: RunningSimulationState): SimulationAircraft[] {
  if (state.playbackTick == null) {
    return state.aircrafts;
  }

  return state.history[state.playbackTick]?.aircrafts ?? state.aircrafts;
}

export function getVisibleThreatsForPlayback(state: RunningSimulationState): SimulationThreat[] {
  if (state.playbackTick == null) {
    return state.activeThreats;
  }

  return state.history[state.playbackTick]?.activeThreats ?? state.activeThreats;
}

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
    activeThreats: cachedState?.activeThreats ?? [],
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
    terminalRecord: cachedState?.terminalRecord,
  };
}

function isBranchCreatedEvent(event: SimulationEvent): event is BranchCreatedEvent {
  return event.type === 'branch_created'
    && typeof event.branchId === 'string'
    && typeof event.parentId === 'string'
    && typeof event.splitTick === 'number'
    && typeof event.splitTimestamp === 'string';
}

function isSimulationEndedEvent(event: SimulationEvent): event is SimulationEndedEvent {
  return event.type === 'simulation_ended'
    && typeof event.tick === 'number'
    && typeof event.summary?.completedVisitCount === 'number'
    && typeof event.summary?.totalDurationMs === 'number'
    && (typeof event.summary?.averageDurationMs === 'number' || event.summary?.averageDurationMs === null);
}

function isSimulationClosedEvent(event: SimulationEvent): event is SimulationClosedEvent {
  return event.type === 'simulation_closed'
    && typeof event.tick === 'number'
    && (event.reason === 'reset' || event.reason === 'cancel')
    && typeof event.summary?.completedVisitCount === 'number'
    && typeof event.summary?.totalDurationMs === 'number'
    && (typeof event.summary?.averageDurationMs === 'number' || event.summary?.averageDurationMs === null);
}

export function createTerminalSimulationRecord(
  event: SimulationEndedEvent | SimulationClosedEvent,
): TerminalSimulationRecord {
  return {
    simulationId: event.simulationId,
    tick: event.tick,
    timestamp: event.timestamp,
    kind: event.type === 'simulation_ended' ? 'ended' : 'closed',
    reason: event.type === 'simulation_closed' ? event.reason : undefined,
    summary: event.summary,
  };
}

export function formatTerminalSummaryDuration(durationMs: number | null): string {
  if (durationMs == null) {
    return 'No completed services yet';
  }

  if (durationMs < 1000) {
    return `${durationMs} ms`;
  }

  const seconds = durationMs / 1000;
  return `${seconds.toFixed(seconds >= 10 ? 0 : 1)} s`;
}

export function formatTerminalSummaryHeadline(record: TerminalSimulationRecord): string {
  if (record.kind === 'ended') {
    return 'Run completed successfully';
  }

  return record.reason === 'reset' ? 'Run stopped and reset' : 'Branch stopped';
}

export function sortTerminalSimulationRecords(
  records: TerminalSimulationRecord[],
): TerminalSimulationRecord[] {
  return [...records].sort((left, right) => {
    if (left.simulationId === 'base') {
      return -1;
    }

    if (right.simulationId === 'base') {
      return 1;
    }

    if (left.timestamp !== right.timestamp) {
      return right.timestamp.localeCompare(left.timestamp);
    }

    if (left.tick !== right.tick) {
      return right.tick - left.tick;
    }

    return left.simulationId.localeCompare(right.simulationId);
  });
}

function sameTerminalRecord(
  left: TerminalSimulationRecord | null,
  right: TerminalSimulationRecord,
): boolean {
  return left?.simulationId === right.simulationId
    && left?.tick === right.tick
    && left?.timestamp === right.timestamp
    && left?.kind === right.kind
    && left?.reason === right.reason;
}

export function useSimulation() {
  const { clients, config } = useApi();
  const [state, setState] = useState<SimulationState>({ status: 'idle' });
  const [simulations, setSimulations] = useState<SimulationInfo[]>([]);
  const [isLoadingSimulations, setIsLoadingSimulations] = useState(false);
  const [latestTerminalRecord, setLatestTerminalRecord] = useState<TerminalSimulationRecord | null>(null);
  const [activeTerminalModalRecord, setActiveTerminalModalRecord] = useState<TerminalSimulationRecord | null>(null);
  const simulationStateCacheRef = useRef<Map<string, RunningSimulationState>>(new Map());
  const eventsBySimulationRef = useRef<Map<string, SimulationEvent[]>>(new Map());
  const terminalRecordsRef = useRef<Map<string, TerminalSimulationRecord>>(new Map());
  const streamClientsRef = useRef<
    Map<string, { client: SimulationStreamClient; unsubscribe: () => void; unsubscribeState: () => void }>
  >(new Map());
  const [eventVersion, setEventVersion] = useState(0);

  const activeSimulationId = state.status === 'running' ? state.simulationId : undefined;
  const stream = useSimulationStream(activeSimulationId);
  const shouldTrackBaseStream = useMemo(() => {
    return simulations.some((simulation) => simulation.id === BASE_SIMULATION_ID) || state.status === 'running';
  }, [simulations, state.status]);
  const baseStream = useSimulationStream(shouldTrackBaseStream ? BASE_SIMULATION_ID : undefined);
  const currentEvents = state.status === 'running'
    ? (eventsBySimulationRef.current.get(state.simulationId) ?? [])
    : [];
  const currentTerminalRecord = state.status === 'running'
    ? (terminalRecordsRef.current.get(state.simulationId) ?? state.terminalRecord ?? null)
    : latestTerminalRecord;
  const terminalRecords = sortTerminalSimulationRecords(Array.from(terminalRecordsRef.current.values()));
  const timelineEventsBySimulation = useMemo(
    () => new Map(eventsBySimulationRef.current),
    [eventVersion],
  );
  const timelineTerminalRecordsBySimulation = useMemo(
    () => new Map(terminalRecordsRef.current),
    [terminalRecords, latestTerminalRecord],
  );

  const registerTerminalRecord = useCallback((record: TerminalSimulationRecord) => {
    terminalRecordsRef.current.set(record.simulationId, record);
    setLatestTerminalRecord(record);
    setActiveTerminalModalRecord((current) => {
      if (sameTerminalRecord(current, record)) {
        return current;
      }

      return record;
    });
  }, []);

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
        if (isSimulationEndedEvent(event) || isSimulationClosedEvent(event)) {
          registerTerminalRecord(createTerminalSimulationRecord(event));
        }
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
  }, [config, recordSimulationEvent, registerTerminalRecord, simulations]);

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

      if (isSimulationClosedEvent(event)) {
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
              [currentTick]: buildHistorySnapshot(current, currentTick),
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
              [currentTick]: buildHistorySnapshot(current, currentTick, { aircrafts: updatedAircrafts }),
            },
          };
        });
      } else if (isLandingAssignmentEvent(event)) {
        setState((current) => {
          if (current.status !== 'running') return current;
          const updatedAircrafts = applyAircraftAssignment(
            current.aircrafts,
            event.tailNumber,
            {
              base: event.baseId,
              source: event.source,
            },
            event.needs,
          );
          const currentTick = current.tick ?? 0;
          return {
            ...current,
            aircrafts: updatedAircrafts,
            history: {
              ...current.history,
              [currentTick]: buildHistorySnapshot(current, currentTick, { aircrafts: updatedAircrafts }),
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
              [currentTick]: buildHistorySnapshot(current, currentTick, {
                aircraftPositions: event.positions,
              }),
            },
          };
        });
      } else if (isThreatSpawnedEvent(event)) {
        setState((current) => {
          if (current.status !== 'running') return current;
          const updatedThreats = upsertThreat(current.activeThreats, event.threat);
          const currentTick = event.tick ?? current.tick ?? 0;
          return {
            ...current,
            activeThreats: updatedThreats,
            history: {
              ...current.history,
              [currentTick]: buildHistorySnapshot(current, currentTick, { activeThreats: updatedThreats }),
            },
          };
        });
      } else if (isThreatDespawnedEvent(event)) {
        setState((current) => {
          if (current.status !== 'running') return current;
          const updatedThreats = removeThreat(current.activeThreats, event.threat.id);
          const currentTick = event.tick ?? current.tick ?? 0;
          return {
            ...current,
            activeThreats: updatedThreats,
            history: {
              ...current.history,
              [currentTick]: buildHistorySnapshot(current, currentTick, { activeThreats: updatedThreats }),
            },
          };
        });
      } else if (isSimulationEndedEvent(event)) {
        const terminalRecord = terminalRecordsRef.current.get(event.simulationId)
          ?? createTerminalSimulationRecord(event);
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
            terminalRecord,
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
    if (event.simulationId !== BASE_SIMULATION_ID) {
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

  const overrideAssignment = useCallback(async (tailNumber: string, baseId: string): Promise<boolean> => {
    if (state.status !== 'running') {
      return false;
    }

    try {
      const response = await clients.simulation.overrideAssignment(state.simulationId, tailNumber, { baseId });
      setState((current) => {
        if (current.status !== 'running' || current.simulationId !== state.simulationId) {
          return current;
        }

        const updatedAircrafts = applyOverrideResponse(current.aircrafts, response);
        const currentTick = current.tick ?? 0;
        return {
          ...current,
          aircrafts: updatedAircrafts,
          history: {
            ...current.history,
            [currentTick]: buildHistorySnapshot(current, currentTick, { aircrafts: updatedAircrafts }),
          },
        };
      });
      toast.success('Assignment override applied');
      return true;
    } catch (error) {
      toast.error(mapOverrideErrorMessage(error));
      if (getErrorStatus(error) === 404) {
        const list = await fetchSimulations();
        if (!list.some((simulation) => simulation.id === state.simulationId)) {
          const fallbackSimulationId = getSimulationCloseFallback(state.simulationId, list);
          if (fallbackSimulationId) {
            await loadSimulation(fallbackSimulationId);
          }
        }
      }
      return false;
    }
  }, [clients.simulation, fetchSimulations, loadSimulation, state]);

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
          terminalRecord: terminalRecordsRef.current.get(simulationInfo.id) ?? current.terminalRecord,
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

  const dismissTerminalModal = useCallback(() => {
    setActiveTerminalModalRecord(null);
  }, []);

  const visibleState = state.status === 'running'
    ? {
        ...state,
        aircrafts: getVisibleAircraftsForPlayback(state),
        activeThreats: getVisibleThreatsForPlayback(state),
        aircraftPositions: state.playbackTick != null && state.history[state.playbackTick]
          ? state.history[state.playbackTick].aircraftPositions ?? state.aircraftPositions
          : state.aircraftPositions,
      }
    : state;

  return {
    state: visibleState,
    activeSimulationId,
    events: currentEvents,
    terminalRecord: currentTerminalRecord,
    latestTerminalRecord,
    terminalRecords,
    timelineEventsBySimulation,
    timelineTerminalRecordsBySimulation,
    activeTerminalModalRecord,
    dismissTerminalModal,
    setPlaybackTick,
    simulations,
    isLoadingSimulations,
    loadSimulation,
    createSimulation,
    createBranchFromEvent,
    overrideAssignment,
    refreshData,
    triggerReset,
    reset,
  };
}
