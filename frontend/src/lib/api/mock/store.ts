import type {
  AssignmentSource,
  CreateBaseSimulationRequest,
  CreateBranchSimulationRequest,
  OverrideAssignmentResponse,
  SimulationEvent,
  SimulationAircraft,
  SimulationInfo,
} from '@/lib/api/types';
import {
  cloneMockSimulationScenario,
  createMockEventTimestamp,
  createMockSimulationFromRequest,
  getMockSimulationScenario,
  listMockSimulationScenarios,
  type MockSimulationScenario,
} from '@/lib/api/mock/scenarios';

const simulationStore = new Map<string, MockSimulationScenario>();
const simulationEventSubscribers = new Map<string, Set<(event: SimulationEvent) => void>>();
let branchCounter = 0;

function initializeStore() {
  simulationStore.clear();
  for (const scenario of listMockSimulationScenarios()) {
    simulationStore.set(scenario.info.id, scenario);
  }
}

initializeStore();

function cloneInfo(info: SimulationInfo): SimulationInfo {
  return {
    ...info,
    sourceEvent: info.sourceEvent ? { ...info.sourceEvent } : undefined,
  };
}

function cloneStoredScenario(scenario: MockSimulationScenario): MockSimulationScenario {
  return cloneMockSimulationScenario(scenario, scenario.info.id);
}

export function resetMockScenarioStore() {
  branchCounter = 0;
  initializeStore();
  simulationEventSubscribers.clear();
}

function createMockHttpError(status: number, message: string) {
  return {
    status,
    body: JSON.stringify({ message }),
  };
}

function cloneAircraft(aircraft: SimulationAircraft): SimulationAircraft {
  return {
    ...aircraft,
    needs: aircraft.needs.map((need) => ({ ...need })),
  };
}

function emitMockSimulationEvent(simulationId: string, event: SimulationEvent) {
  const subscribers = simulationEventSubscribers.get(simulationId);
  if (!subscribers) {
    return;
  }

  subscribers.forEach((handler) => {
    handler(event);
  });
}

export function listStoredMockSimulationScenarios(): MockSimulationScenario[] {
  return Array.from(simulationStore.values()).map((scenario) => cloneStoredScenario(scenario));
}

export function getStoredMockSimulationScenario(simulationId: string): MockSimulationScenario {
  const existingScenario = simulationStore.get(simulationId);
  if (existingScenario) {
    return cloneStoredScenario(existingScenario);
  }

  const fallbackScenario = getMockSimulationScenario(simulationId);
  simulationStore.set(simulationId, fallbackScenario);
  return fallbackScenario;
}

export function updateStoredMockSimulationInfo(simulationId: string, nextInfo: SimulationInfo) {
  const scenario = simulationStore.get(simulationId) ?? getMockSimulationScenario(simulationId);
  simulationStore.set(simulationId, {
    ...cloneStoredScenario(scenario),
    info: cloneInfo(nextInfo),
  });
}

export function createStoredMockBaseSimulation(request: CreateBaseSimulationRequest) {
  const scenario = createMockSimulationFromRequest(request);
  simulationStore.set(scenario.info.id, scenario);
  return scenario;
}

export function resetStoredMockSimulation(simulationId: string) {
  if (simulationId === 'base') {
    simulationStore.delete(simulationId);
    return;
  }

  const scenario = getMockSimulationScenario(simulationId);
  simulationStore.set(simulationId, scenario);
}

export function createStoredMockBranchSimulation(
  simulationId: string,
  request?: CreateBranchSimulationRequest,
): MockSimulationScenario {
  const parentScenario = getStoredMockSimulationScenario(simulationId);
  branchCounter += 1;
  const branchId = `branch-${branchCounter.toString(16).padStart(8, '0')}`;
  const splitTick = parentScenario.info.tick;
  const splitTimestamp = parentScenario.info.timestamp;
  const branchScenario = getMockSimulationScenario(branchId);
  const nextInfo: SimulationInfo = {
    ...branchScenario.info,
    id: branchId,
    running: parentScenario.info.running,
    paused: parentScenario.info.running,
    tick: splitTick,
    timestamp: splitTimestamp,
    untilTick: parentScenario.info.untilTick,
    parentId: simulationId,
    splitTick,
    splitTimestamp,
    sourceEvent: request?.sourceEvent ? { ...request.sourceEvent } : undefined,
  };

  const storedScenario: MockSimulationScenario = {
    ...cloneStoredScenario(branchScenario),
    info: nextInfo,
  };
  simulationStore.set(branchId, storedScenario);

  const branchCreatedEvent: SimulationEvent = {
    type: 'branch_created',
    simulationId: 'base',
    timestamp: splitTimestamp,
    tick: splitTick,
    branchId,
    parentId: simulationId,
    splitTick,
    splitTimestamp,
    sourceEvent: request?.sourceEvent ? { ...request.sourceEvent } : undefined,
  };
  emitMockSimulationEvent('base', branchCreatedEvent);

  return storedScenario;
}

export function overrideStoredMockAssignment(
  simulationId: string,
  tailNumber: string,
  baseId: string,
): OverrideAssignmentResponse {
  const scenario = simulationStore.get(simulationId);
  if (!scenario) {
    throw createMockHttpError(404, 'simulation or aircraft not found');
  }

  const baseExists = scenario.airbases.some((airbase) => airbase.id === baseId);
  if (!baseExists) {
    throw createMockHttpError(400, 'airbase not found');
  }

  const nextAircrafts = scenario.aircrafts.map((aircraft) => {
    if (aircraft.tailNumber !== tailNumber) {
      return cloneAircraft(aircraft);
    }

    if (aircraft.state === 'Committed') {
      throw createMockHttpError(409, 'assignment override too late');
    }

    return {
      ...cloneAircraft(aircraft),
      assignedTo: baseId,
      assignmentSource: 'human' as AssignmentSource,
    };
  });

  const updatedAircraft = nextAircrafts.find((aircraft) => aircraft.tailNumber === tailNumber);
  if (!updatedAircraft) {
    throw createMockHttpError(404, 'simulation or aircraft not found');
  }

  simulationStore.set(simulationId, {
    ...cloneStoredScenario(scenario),
    aircrafts: nextAircrafts,
  });

  emitMockSimulationEvent(simulationId, {
    type: 'landing_assignment',
    simulationId,
    tailNumber,
    baseId,
    source: 'human',
    tick: scenario.info.tick,
    timestamp: createMockEventTimestamp(scenario.info.tick, scenario.info.tick),
  });

  return {
    aircraft: cloneAircraft(updatedAircraft),
    assignment: {
      base: baseId,
      source: 'human',
    },
  };
}

export function createStoredMockSimulationInfoUpdate(
  simulationId: string,
  nextState: Partial<Pick<SimulationInfo, 'running' | 'paused' | 'tick' | 'timestamp' | 'untilTick'>>,
): SimulationInfo {
  const scenario = getStoredMockSimulationScenario(simulationId);
  return {
    ...scenario.info,
    ...nextState,
  };
}

export function subscribeToMockSimulationEvents(
  simulationId: string,
  handler: (event: SimulationEvent) => void,
): () => void {
  const currentSubscribers = simulationEventSubscribers.get(simulationId) ?? new Set();
  currentSubscribers.add(handler);
  simulationEventSubscribers.set(simulationId, currentSubscribers);

  return () => {
    const subscribers = simulationEventSubscribers.get(simulationId);
    if (!subscribers) {
      return;
    }

    subscribers.delete(handler);
    if (subscribers.size === 0) {
      simulationEventSubscribers.delete(simulationId);
    }
  };
}

export function createMockSequenceEvent(event: Omit<SimulationEvent, 'timestamp' | 'sequence'>, index: number): SimulationEvent {
  return {
    ...event,
    timestamp: createMockEventTimestamp(typeof event.tick === 'number' ? event.tick : index, index),
    sequence: index + 1,
  } as SimulationEvent;
}
