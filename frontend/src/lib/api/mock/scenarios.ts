import type {
  CreateBaseSimulationRequest,
  SimulationAirbase,
  SimulationAircraft,
  SimulationEvent,
  SimulationInfo,
} from '@/lib/api/types';

type MockSimulationScenario = {
  info: SimulationInfo;
  airbases: SimulationAirbase[];
  aircrafts: SimulationAircraft[];
  events: Array<Omit<SimulationEvent, 'simulationId' | 'timestamp' | 'sequence'>>;
};

const MOCK_SIMULATION_START = '2026-01-01T08:00:00.000Z';
const LIGHT_SCENARIO_ID = 'mock-light-sortie';
const FULL_SCENARIO_ID = 'mock-full-sortie';
const LEGACY_SCENARIO_ID = 'base';

function createIsoAtTick(tick: number): string {
  return new Date(Date.parse(MOCK_SIMULATION_START) + tick * 60_000).toISOString();
}

function createNeed(type: string, severity: number, blocking = false) {
  return {
    type,
    severity,
    requiredCapability: type,
    blocking,
  };
}

const LIGHT_AIRBASES: SimulationAirbase[] = [
  {
    id: 'd397eeeddbfae33e',
    location: { x: 109.44765839799018, y: 753.1689567645848 },
    regionId: 'SE-K',
    region: 'Blekinge',
    metadata: { capacity: 2, posture: 'alert' },
  },
  {
    id: 'e32c7ce8a4b77c11',
    location: { x: 84.1112039123401, y: 646.5238701142055 },
    regionId: 'SE-O',
    region: 'Vastra Gotaland',
    metadata: { capacity: 3, posture: 'ready' },
  },
  {
    id: '4b8b4f6e0f91b7bd',
    location: { x: 202.4459181132085, y: 683.2145039186111 },
    regionId: 'SE-I',
    region: 'Gotland',
    metadata: { capacity: 2, posture: 'hold' },
  },
  {
    id: '1ae4f4da1946d172',
    location: { x: 175.9321742199307, y: 592.2041175321494 },
    regionId: 'SE-AB',
    region: 'Stockholm',
    metadata: { capacity: 4, posture: 'surge' },
  },
];

const FULL_AIRBASES: SimulationAirbase[] = [
  ...LIGHT_AIRBASES,
  {
    id: '5c55ed247f82afee',
    location: { x: 269.4434002801482, y: 211.0381271336186 },
    regionId: 'SE-BD',
    region: 'Norrbotten',
    metadata: { capacity: 5, posture: 'dispersed' },
  },
  {
    id: '77d0b19d42cb68a1',
    location: { x: 236.2844107311172, y: 317.4833391172533 },
    regionId: 'SE-Z',
    region: 'Jamtland',
    metadata: { capacity: 3, posture: 'forward' },
  },
  {
    id: 'c4d6652975d0f4c8',
    location: { x: 151.9286048183827, y: 535.8519046032481 },
    regionId: 'SE-U',
    region: 'Vastmanland',
    metadata: { capacity: 4, posture: 'reserve' },
  },
];

const LIGHT_AIRCRAFTS: SimulationAircraft[] = [
  {
    tailNumber: 'BX-101',
    needs: [createNeed('fuel', 28, true)],
    state: 'Inbound',
  },
  {
    tailNumber: 'BX-214',
    needs: [createNeed('crew_support', 18)],
    state: 'Turnaround',
    assignedTo: '1ae4f4da1946d172',
  },
  {
    tailNumber: 'BX-330',
    needs: [],
    state: 'Ready',
    assignedTo: 'e32c7ce8a4b77c11',
  },
];

const FULL_AIRCRAFTS: SimulationAircraft[] = [
  {
    tailNumber: 'BX-101',
    needs: [createNeed('fuel', 32, true), createNeed('munitions', 54)],
    state: 'Inbound',
  },
  {
    tailNumber: 'BX-214',
    needs: [createNeed('crew_support', 22), createNeed('ground_support', 14)],
    state: 'Landing',
  },
  {
    tailNumber: 'BX-330',
    needs: [createNeed('maintenance', 61, true)],
    state: 'Assessment',
  },
  {
    tailNumber: 'BX-441',
    needs: [createNeed('fuel', 26), createNeed('mission_configuration', 45)],
    state: 'Turnaround',
    assignedTo: '77d0b19d42cb68a1',
  },
  {
    tailNumber: 'BX-578',
    needs: [createNeed('protection', 19)],
    state: 'Ready',
    assignedTo: '5c55ed247f82afee',
  },
  {
    tailNumber: 'BX-662',
    needs: [createNeed('repairs', 73, true)],
    state: 'Holding',
  },
];

const LIGHT_EVENTS: MockSimulationScenario['events'] = [
  { type: 'simulation_step', tick: 0 },
  {
    type: 'all_aircraft_positions',
    tick: 0,
    positions: [
      { tailNumber: 'BX-101', position: { x: 112.448, y: 739.112 }, state: 'Inbound', needs: LIGHT_AIRCRAFTS[0]?.needs ?? [] },
      { tailNumber: 'BX-214', position: { x: 173.882, y: 598.744 }, state: 'Turnaround', needs: LIGHT_AIRCRAFTS[1]?.needs ?? [] },
      { tailNumber: 'BX-330', position: { x: 84.111, y: 646.524 }, state: 'Ready', needs: [] },
    ],
  },
  { type: 'landing_assignment', tick: 1, tailNumber: 'BX-101', baseId: 'd397eeeddbfae33e' },
  { type: 'simulation_step', tick: 1 },
  {
    type: 'aircraft_state_change',
    tick: 2,
    tailNumber: 'BX-101',
    aircraft: {
      state: 'Servicing',
      assignedTo: 'd397eeeddbfae33e',
      needs: [createNeed('fuel', 12)],
    },
  },
  { type: 'simulation_step', tick: 2 },
  {
    type: 'all_aircraft_positions',
    tick: 2,
    positions: [
      { tailNumber: 'BX-101', position: { x: 109.448, y: 753.169 }, state: 'Servicing', needs: [createNeed('fuel', 12)] },
      { tailNumber: 'BX-214', position: { x: 175.932, y: 592.204 }, state: 'Turnaround', needs: LIGHT_AIRCRAFTS[1]?.needs ?? [] },
      { tailNumber: 'BX-330', position: { x: 84.111, y: 646.524 }, state: 'Ready', needs: [] },
    ],
  },
  { type: 'simulation_step', tick: 4 },
  {
    type: 'aircraft_state_change',
    tick: 5,
    tailNumber: 'BX-101',
    aircraft: {
      state: 'Ready',
      assignedTo: 'd397eeeddbfae33e',
      needs: [],
    },
  },
  { type: 'simulation_step', tick: 6 },
  { type: 'simulation_ended', tick: 8 },
];

const FULL_EVENTS: MockSimulationScenario['events'] = [
  { type: 'simulation_step', tick: 0 },
  {
    type: 'all_aircraft_positions',
    tick: 0,
    positions: [
      { tailNumber: 'BX-101', position: { x: 117.221, y: 732.402 }, state: 'Inbound', needs: FULL_AIRCRAFTS[0]?.needs ?? [] },
      { tailNumber: 'BX-214', position: { x: 147.228, y: 541.118 }, state: 'Landing', needs: FULL_AIRCRAFTS[1]?.needs ?? [] },
      { tailNumber: 'BX-330', position: { x: 240.772, y: 318.661 }, state: 'Assessment', needs: FULL_AIRCRAFTS[2]?.needs ?? [] },
      { tailNumber: 'BX-441', position: { x: 236.284, y: 317.483 }, state: 'Turnaround', needs: FULL_AIRCRAFTS[3]?.needs ?? [] },
      { tailNumber: 'BX-578', position: { x: 269.443, y: 211.038 }, state: 'Ready', needs: FULL_AIRCRAFTS[4]?.needs ?? [] },
      { tailNumber: 'BX-662', position: { x: 162.004, y: 524.113 }, state: 'Holding', needs: FULL_AIRCRAFTS[5]?.needs ?? [] },
    ],
  },
  { type: 'landing_assignment', tick: 1, tailNumber: 'BX-101', baseId: 'd397eeeddbfae33e' },
  { type: 'landing_assignment', tick: 1, tailNumber: 'BX-214', baseId: 'c4d6652975d0f4c8' },
  { type: 'simulation_step', tick: 1 },
  {
    type: 'aircraft_state_change',
    tick: 2,
    tailNumber: 'BX-214',
    aircraft: {
      state: 'Servicing',
      assignedTo: 'c4d6652975d0f4c8',
      needs: [createNeed('crew_support', 10), createNeed('ground_support', 8)],
    },
  },
  {
    type: 'aircraft_state_change',
    tick: 2,
    tailNumber: 'BX-330',
    aircraft: {
      state: 'Repair',
      assignedTo: '77d0b19d42cb68a1',
      needs: [createNeed('repairs', 51, true)],
    },
  },
  { type: 'simulation_step', tick: 2 },
  {
    type: 'all_aircraft_positions',
    tick: 3,
    positions: [
      { tailNumber: 'BX-101', position: { x: 109.448, y: 753.169 }, state: 'Servicing', needs: [createNeed('fuel', 20), createNeed('munitions', 40)] },
      { tailNumber: 'BX-214', position: { x: 151.929, y: 535.852 }, state: 'Servicing', needs: [createNeed('crew_support', 10), createNeed('ground_support', 8)] },
      { tailNumber: 'BX-330', position: { x: 236.284, y: 317.483 }, state: 'Repair', needs: [createNeed('repairs', 51, true)] },
      { tailNumber: 'BX-441', position: { x: 236.284, y: 317.483 }, state: 'Turnaround', needs: FULL_AIRCRAFTS[3]?.needs ?? [] },
      { tailNumber: 'BX-578', position: { x: 269.443, y: 211.038 }, state: 'Ready', needs: FULL_AIRCRAFTS[4]?.needs ?? [] },
      { tailNumber: 'BX-662', position: { x: 162.004, y: 524.113 }, state: 'Holding', needs: FULL_AIRCRAFTS[5]?.needs ?? [] },
    ],
  },
  { type: 'simulation_step', tick: 4 },
  {
    type: 'aircraft_state_change',
    tick: 5,
    tailNumber: 'BX-101',
    aircraft: {
      state: 'Ready',
      assignedTo: 'd397eeeddbfae33e',
      needs: [],
    },
  },
  {
    type: 'aircraft_state_change',
    tick: 5,
    tailNumber: 'BX-330',
    aircraft: {
      state: 'Taxi',
      assignedTo: '77d0b19d42cb68a1',
      needs: [createNeed('repairs', 18)],
    },
  },
  { type: 'simulation_step', tick: 6 },
  {
    type: 'all_aircraft_positions',
    tick: 7,
    positions: [
      { tailNumber: 'BX-101', position: { x: 118.004, y: 742.101 }, state: 'Ready', needs: [] },
      { tailNumber: 'BX-214', position: { x: 154.11, y: 531.004 }, state: 'Ready', needs: [] },
      { tailNumber: 'BX-330', position: { x: 226.441, y: 329.008 }, state: 'Taxi', needs: [createNeed('repairs', 18)] },
      { tailNumber: 'BX-441', position: { x: 236.284, y: 317.483 }, state: 'Turnaround', needs: FULL_AIRCRAFTS[3]?.needs ?? [] },
      { tailNumber: 'BX-578', position: { x: 269.443, y: 211.038 }, state: 'Ready', needs: FULL_AIRCRAFTS[4]?.needs ?? [] },
      { tailNumber: 'BX-662', position: { x: 171.337, y: 516.772 }, state: 'Holding', needs: FULL_AIRCRAFTS[5]?.needs ?? [] },
    ],
  },
  { type: 'simulation_step', tick: 9 },
  { type: 'simulation_step', tick: 12 },
  { type: 'simulation_ended', tick: 16 },
];

const BASE_SCENARIOS: Record<string, MockSimulationScenario> = {
  [LIGHT_SCENARIO_ID]: {
    info: {
      id: LIGHT_SCENARIO_ID,
      running: false,
      paused: false,
      tick: 0,
      timestamp: createIsoAtTick(0),
      untilTick: 8,
    },
    airbases: LIGHT_AIRBASES,
    aircrafts: LIGHT_AIRCRAFTS,
    events: LIGHT_EVENTS,
  },
  [FULL_SCENARIO_ID]: {
    info: {
      id: FULL_SCENARIO_ID,
      running: false,
      paused: false,
      tick: 0,
      timestamp: createIsoAtTick(0),
      untilTick: 16,
    },
    airbases: FULL_AIRBASES,
    aircrafts: FULL_AIRCRAFTS,
    events: FULL_EVENTS,
  },
};

function cloneScenario(scenario: MockSimulationScenario, simulationId = scenario.info.id): MockSimulationScenario {
  return {
    info: { ...scenario.info, id: simulationId },
    airbases: scenario.airbases.map((airbase) => ({
      ...airbase,
      location: { ...airbase.location },
      metadata: airbase.metadata ? { ...airbase.metadata } : undefined,
    })),
    aircrafts: scenario.aircrafts.map((aircraft) => ({
      ...aircraft,
      needs: aircraft.needs.map((need) => ({ ...need })),
    })),
    events: scenario.events.map((event) => ({
      ...event,
      aircraft: event.aircraft
        ? {
            ...event.aircraft,
            needs: Array.isArray(event.aircraft.needs)
              ? event.aircraft.needs.map((need: Record<string, unknown>) => ({ ...need }))
              : event.aircraft.needs,
          }
        : undefined,
      positions: Array.isArray(event.positions)
        ? event.positions.map((position: Record<string, unknown>) => ({
            ...position,
            position:
              typeof position.position === 'object' && position.position !== null
                ? { ...(position.position as Record<string, unknown>) }
                : position.position,
            needs: Array.isArray(position.needs)
              ? position.needs.map((need: Record<string, unknown>) => ({ ...need }))
              : position.needs,
          }))
        : event.positions,
    })),
  };
}

function normalizeScenarioId(simulationId: string): string {
  return simulationId === LEGACY_SCENARIO_ID ? FULL_SCENARIO_ID : simulationId;
}

function shouldUseLightScenario(request: CreateBaseSimulationRequest): boolean {
  const constellationMax = request.simulationOptions?.constellationOpts?.maxTotal ?? 0;
  const aircraftMax = request.simulationOptions?.fleetOpts?.aircraftMax ?? 0;
  const untilTick = request.untilTick ?? 0;

  return untilTick > 0 && untilTick <= 10 && constellationMax <= 4 && aircraftMax <= 4;
}

export function resolveMockScenarioTemplateId(request?: CreateBaseSimulationRequest): string {
  if (!request) {
    return FULL_SCENARIO_ID;
  }

  return shouldUseLightScenario(request) ? LIGHT_SCENARIO_ID : FULL_SCENARIO_ID;
}

export function listMockSimulationScenarios(): MockSimulationScenario[] {
  return [cloneScenario(BASE_SCENARIOS[LIGHT_SCENARIO_ID]!), cloneScenario(BASE_SCENARIOS[FULL_SCENARIO_ID]!)];
}

export function getMockSimulationScenario(simulationId: string): MockSimulationScenario {
  const normalizedId = normalizeScenarioId(simulationId);
  const baseScenario = BASE_SCENARIOS[normalizedId] ?? BASE_SCENARIOS[FULL_SCENARIO_ID]!;
  return cloneScenario(baseScenario, simulationId);
}

export function createMockSimulationInfoUpdate(
  simulationId: string,
  nextState: Partial<Pick<SimulationInfo, 'running' | 'paused' | 'tick' | 'timestamp' | 'untilTick'>>,
): SimulationInfo {
  const scenario = getMockSimulationScenario(simulationId);
  return {
    ...scenario.info,
    ...nextState,
  };
}

export function createMockSimulationFromRequest(request: CreateBaseSimulationRequest): MockSimulationScenario {
  const templateId = resolveMockScenarioTemplateId(request);
  const scenario = getMockSimulationScenario(templateId);

  return {
    ...scenario,
    info: {
      ...scenario.info,
      untilTick: request.untilTick ?? scenario.info.untilTick,
      timestamp: createIsoAtTick(0),
      tick: 0,
    },
  };
}

export function createMockEventTimestamp(tick: number, fallbackIndex: number): string {
  return createIsoAtTick(Number.isFinite(tick) ? tick : fallbackIndex);
}
