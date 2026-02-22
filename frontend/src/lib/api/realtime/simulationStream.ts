import { parseApiConfigFromEnv, SIMULATION_WS_PATH } from '@/lib/api/config';
import { createWebSocketClient } from '@/lib/api/realtime/socket';
import type { ApiConfig, SimulationEventEnvelope, SimulationStreamClient } from '@/lib/api/types';

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function parseSimulationEvent(rawData: string): SimulationEventEnvelope | null {
  const parsed = JSON.parse(rawData) as unknown;

  if (!isRecord(parsed)) {
    return null;
  }

  if (
    typeof parsed.type !== 'string' ||
    typeof parsed.runId !== 'string' ||
    typeof parsed.sequence !== 'number' ||
    typeof parsed.timestamp !== 'string'
  ) {
    return null;
  }

  return {
    type: parsed.type,
    runId: parsed.runId,
    sequence: parsed.sequence,
    timestamp: parsed.timestamp,
    payload: parsed.payload,
  };
}

export function createSimulationStreamClient(overrides?: Partial<ApiConfig>): SimulationStreamClient {
  const config: ApiConfig = {
    ...parseApiConfigFromEnv(),
    ...overrides,
  };

  return createWebSocketClient(config, {
    path: SIMULATION_WS_PATH,
    parseEvent: parseSimulationEvent,
  });
}
