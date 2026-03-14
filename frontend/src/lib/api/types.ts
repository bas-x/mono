export type ApiMode = 'mock' | 'remote' | 'localhost';

export type ApiConfig = {
  apiBaseUrl: string;
  wsBaseUrl: string;
  mode: ApiMode;
  useMock: boolean;
};

export type HealthPingResult = {
  ok: boolean;
  message: string;
  time: string;
};

export interface HealthServiceClient {
  ping(signal?: AbortSignal): Promise<HealthPingResult>;
}

export type ApiAirbasePoint = {
  x: number;
  y: number;
};

export type ApiAirbase = {
  id: string;
  name: string;
  area: ApiAirbasePoint[];
  infoUrl?: string;
};

export type ApiAirbaseDetails = {
  id: string;
  [key: string]: unknown;
};

export interface MapServiceClient {
  getAirbases(signal?: AbortSignal): Promise<ApiAirbase[]>;
  getAirbaseDetails(idOrUrl: string, signal?: AbortSignal): Promise<ApiAirbaseDetails>;
}

export type SimulationAirbase = {
  id: string;
  name: string;
  location: { x: number; y: number };
  regionId: string;
  region: string;
  metadata?: Record<string, unknown>;
};

export type SimulationAircraftNeed = {
  type: string;
  severity: number;
  requiredCapability: string;
  blocking: boolean;
};

export type AirbaseCapability = {
  recoveryMultiplierPermille: number;
};

export type AirbaseCapabilityMap = Record<string, AirbaseCapability>;

export type SimulationAircraft = {
  tailNumber: string;
  model: string;
  needs: SimulationAircraftNeed[];
  state: string;
  assignedTo?: string;
  assignmentSource?: AssignmentSource;
};

export type SimulationThreat = {
  id: string;
  position: { x: number; y: number };
  createdAt: string;
  createdTick: number;
};

export type AssignmentSource = 'algorithm' | 'human';

export type Assignment = {
  base: string;
  source: AssignmentSource;
};

export type ServicingSummary = {
  completedVisitCount: number;
  totalDurationMs: number;
  averageDurationMs: number | null;
};

export type SimulationClosedReason = 'reset' | 'cancel';

export const SIMULATION_TICKS_PER_SECOND = 64;

export type ApiRatio = {
  numerator: number;
  denominator: number;
};

export type CreateSimulationConstellationOptions = {
  includeRegions?: string[];
  excludeRegions?: string[];
  minPerRegion?: number;
  maxPerRegion?: number;
  maxTotal?: number;
  regionProbability?: ApiRatio;
  maxAttemptsPerAirbase?: number;
};

export type CreateSimulationFleetOptions = {
  aircraftMin?: number;
  aircraftMax?: number;
  needsMin?: number;
  needsMax?: number;
  needsPool?: string[];
  severityMin?: number;
  severityMax?: number;
  blockingChance?: ApiRatio;
};

export type CreateSimulationThreatOptions = {
  spawnChance?: ApiRatio;
  maxActive?: number;
  maxActiveTicks?: number;
};

export type CreateSimulationPhaseDurations = {
  outbound?: number;
  engaged?: number;
  inboundDecision?: number;
  commitApproach?: number;
  servicing?: number;
  ready?: number;
};

export type CreateSimulationNeedRateModel = {
  outboundMilliPerHour?: number;
  engagedMilliPerHour?: number;
  servicingMilliPerHour?: number;
  variancePermille?: number;
};

export type CreateSimulationLifecycleOptions = {
  durations?: CreateSimulationPhaseDurations;
  returnThreshold?: number;
  needRates?: Record<string, CreateSimulationNeedRateModel>;
};

export type CreateSimulationOptions = {
  constellationOpts?: CreateSimulationConstellationOptions;
  fleetOpts?: CreateSimulationFleetOptions;
  threatOpts?: CreateSimulationThreatOptions;
  lifecycleOpts?: CreateSimulationLifecycleOptions;
};

export type CreateBaseSimulationRequest = {
  seed?: string;
  untilTick?: number;
  simulationOptions?: CreateSimulationOptions;
};

export type SourceEvent = {
  id: string;
  type: string;
  tick: number;
};

export type CreateBranchSimulationRequest = {
  sourceEvent?: SourceEvent;
};

export type OverrideAssignmentRequest = {
  baseId: string;
};

export type OverrideAssignmentResponse = {
  aircraft: SimulationAircraft;
  assignment: Assignment;
};

export type SimulationInfo = {
  id: string;
  running: boolean;
  paused: boolean;
  tick: number;
  timestamp: string;
  untilTick?: number;
  parentId?: string;
  splitTick?: number;
  splitTimestamp?: string;
  sourceEvent?: SourceEvent;
};

export interface SimulationServiceClient {
  getSimulations(signal?: AbortSignal): Promise<SimulationInfo[]>;
  getSimulation(simulationId: string, signal?: AbortSignal): Promise<SimulationInfo>;
  createBaseSimulation(
    request: CreateBaseSimulationRequest,
    signal?: AbortSignal,
  ): Promise<{ id: string }>;
  createBranchSimulation(
    simulationId: string,
    request?: CreateBranchSimulationRequest,
    signal?: AbortSignal,
  ): Promise<SimulationInfo>;
  overrideAssignment(
    simulationId: string,
    tailNumber: string,
    request: OverrideAssignmentRequest,
    signal?: AbortSignal,
  ): Promise<OverrideAssignmentResponse>;
  startSimulation(simulationId: string, signal?: AbortSignal): Promise<void>;
  pauseSimulation(simulationId: string, signal?: AbortSignal): Promise<void>;
  resumeSimulation(simulationId: string, signal?: AbortSignal): Promise<void>;
  resetSimulation(simulationId: string, signal?: AbortSignal): Promise<void>;
  getAirbases(simulationId: string, signal?: AbortSignal): Promise<SimulationAirbase[]>;
  getAircrafts(simulationId: string, signal?: AbortSignal): Promise<SimulationAircraft[]>;
}

export type ApiClients = {
  health: HealthServiceClient;
  map: MapServiceClient;
  simulation: SimulationServiceClient;
};

export type ConnectionState =
  | 'idle'
  | 'connecting'
  | 'open'
  | 'reconnecting'
  | 'closed'
  | 'error';

export type SimulationEventType =
  | 'simulation.started'
  | 'simulation.progress'
  | 'simulation.completed'
  | 'simulation.error'
  | (string & {});

export type SimulationEvent = {
  type: string;
  simulationId: string;
  timestamp: string;
  sequence?: number;
  [key: string]: any;
};

export type BranchCreatedEvent = SimulationEvent & {
  type: 'branch_created';
  branchId: string;
  parentId: string;
  splitTick: number;
  splitTimestamp: string;
  sourceEvent?: SourceEvent;
};

export type LandingAssignmentEvent = SimulationEvent & {
  type: 'landing_assignment';
  tailNumber: string;
  baseId: string;
  source: AssignmentSource;
  tick: number;
  needs: SimulationAircraftNeed[];
  capabilities: AirbaseCapabilityMap;
};

export type SimulationEndedEvent = SimulationEvent & {
  type: 'simulation_ended';
  tick: number;
  summary: ServicingSummary;
};

export type SimulationClosedEvent = SimulationEvent & {
  type: 'simulation_closed';
  tick: number;
  reason: SimulationClosedReason;
  summary: ServicingSummary;
};

export type ThreatSpawnedEvent = SimulationEvent & {
  type: 'threat_spawned';
  tick: number;
  threat: SimulationThreat;
};

export type ThreatTargetedEvent = SimulationEvent & {
  type: 'threat_targeted';
  tick: number;
  threat: SimulationThreat;
  tailNumber: string;
};

export type ThreatDespawnedEvent = SimulationEvent & {
  type: 'threat_despawned';
  tick: number;
  threat: SimulationThreat;
};

export type SimulationEventEnvelope = SimulationEvent;

export type Unsubscribe = () => void;

export interface SimulationStreamClient {
  connect(simulationId: string): void;
  disconnect(code?: number, reason?: string): void;
  subscribe(handler: (event: SimulationEvent) => void): Unsubscribe;
  onConnectionStateChange(handler: (state: ConnectionState) => void): Unsubscribe;
}
