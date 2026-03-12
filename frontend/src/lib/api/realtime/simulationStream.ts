import { parseApiConfigFromEnv, SIMULATION_WS_PATH } from '@/lib/api/config';
import { createWebSocketClient } from '@/lib/api/realtime/socket';
import type { ApiConfig, SimulationEvent, SimulationStreamClient } from '@/lib/api/types';

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function parseSimulationEvent(rawData: string): SimulationEvent | null {
  const parsed = JSON.parse(rawData) as unknown;

  if (!isRecord(parsed)) {
    return null;
  }

  if (
    typeof parsed.type !== 'string' ||
    typeof parsed.simulationId !== 'string' ||
    typeof parsed.timestamp !== 'string'
  ) {
    return null;
  }

  return parsed as SimulationEvent;
}

export function createSimulationStreamClient(overrides?: Partial<ApiConfig>): SimulationStreamClient {
  const config: ApiConfig = {
    ...parseApiConfigFromEnv(),
    ...overrides,
  };

  const client = createWebSocketClient(config, {
    path: SIMULATION_WS_PATH,
    parseEvent: parseSimulationEvent,
  });

  return {
    ...client,
    connect(simulationId: string) {
      const path = SIMULATION_WS_PATH.replace(':simulationId', encodeURIComponent(simulationId));
      return client.connect(path);
    },
  };
}
