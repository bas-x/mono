import { parseApiConfigFromEnv, SIMULATION_WS_PATH } from '@/lib/api/config';
import { createWebSocketClient } from '@/lib/api/realtime/socket';
import type {
  ApiConfig,
  ServicingSummary,
  SimulationClosedReason,
  SimulationEvent,
  SimulationStreamClient,
} from '@/lib/api/types';

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function parseServicingSummary(value: unknown): ServicingSummary | null {
  if (!isRecord(value)) {
    return null;
  }

  if (
    typeof value.completedVisitCount !== 'number' ||
    typeof value.totalDurationMs !== 'number' ||
    !(typeof value.averageDurationMs === 'number' || value.averageDurationMs === null)
  ) {
    return null;
  }

  return {
    completedVisitCount: value.completedVisitCount,
    totalDurationMs: value.totalDurationMs,
    averageDurationMs: value.averageDurationMs,
  };
}

function parseSimulationClosedReason(value: unknown): SimulationClosedReason | null {
  return value === 'reset' || value === 'cancel' ? value : null;
}

export function parseSimulationEvent(rawData: string): SimulationEvent | null {
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

  if (parsed.type === 'simulation_ended') {
    if (typeof parsed.tick !== 'number') {
      return null;
    }

    const summary = parseServicingSummary(parsed.summary);
    if (!summary) {
      return null;
    }

    return {
      ...(parsed as Record<string, unknown>),
      type: parsed.type,
      simulationId: parsed.simulationId,
      timestamp: parsed.timestamp,
      tick: parsed.tick,
      summary,
    } as SimulationEvent;
  }

  if (parsed.type === 'simulation_closed') {
    if (typeof parsed.tick !== 'number') {
      return null;
    }

    const reason = parseSimulationClosedReason(parsed.reason);
    const summary = parseServicingSummary(parsed.summary);
    if (!reason || !summary) {
      return null;
    }

    return {
      ...(parsed as Record<string, unknown>),
      type: parsed.type,
      simulationId: parsed.simulationId,
      timestamp: parsed.timestamp,
      tick: parsed.tick,
      reason,
      summary,
    } as SimulationEvent;
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
