export type ApiConfig = {
  apiBaseUrl: string;
  wsBaseUrl: string;
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

export type ApiClients = {
  health: HealthServiceClient;
  map: MapServiceClient;
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

export type SimulationEventEnvelope<TPayload = unknown> = {
  type: SimulationEventType;
  runId: string;
  sequence: number;
  timestamp: string;
  payload: TPayload;
};

export type Unsubscribe = () => void;

export interface SimulationStreamClient {
  connect(): void;
  disconnect(code?: number, reason?: string): void;
  subscribe(handler: (event: SimulationEventEnvelope) => void): Unsubscribe;
  onConnectionStateChange(handler: (state: ConnectionState) => void): Unsubscribe;
}
