import type { ApiConfig, ApiMode } from '@/lib/api/types';

export const SIMULATION_WS_PATH = '/ws/simulations/:simulationId/events';

export const REMOTE_API_BASE_URL = 'https://basex.shigure.joshuadematas.me';
export const REMOTE_WS_BASE_URL = 'wss://basex.shigure.joshuadematas.me';

export const LOCALHOST_API_BASE_URL = 'http://localhost:8080';
export const LOCALHOST_WS_BASE_URL = 'ws://localhost:8080';

const DEFAULT_USE_MOCK_API = true;

function normalizeEnvString(value: string | undefined): string | undefined {
  if (!value) {
    return undefined;
  }

  const normalizedValue = value.trim();
  if (!normalizedValue || normalizedValue === 'undefined' || normalizedValue === 'null') {
    return undefined;
  }

  return normalizedValue;
}

function parseUseMock(value: string | undefined): boolean {
  const normalizedValue = normalizeEnvString(value)?.toLowerCase();

  if (normalizedValue === undefined) {
    return DEFAULT_USE_MOCK_API;
  }

  if (normalizedValue === 'true') {
    return true;
  }

  if (normalizedValue === 'false') {
    return false;
  }

  return DEFAULT_USE_MOCK_API;
}

export function resolveBaseUrls(mode: ApiMode, envConfig: { apiBaseUrl: string; wsBaseUrl: string }) {
  if (mode === 'localhost') {
    return {
      apiBaseUrl: LOCALHOST_API_BASE_URL,
      wsBaseUrl: LOCALHOST_WS_BASE_URL,
    };
  }

  return {
    apiBaseUrl: envConfig.apiBaseUrl,
    wsBaseUrl: envConfig.wsBaseUrl,
  };
}

export function parseApiConfigFromEnv(): ApiConfig {
  const apiBaseUrl = normalizeEnvString(import.meta.env.VITE_API_BASE_URL) || REMOTE_API_BASE_URL;
  const wsBaseUrl = normalizeEnvString(import.meta.env.VITE_WS_BASE_URL) || REMOTE_WS_BASE_URL;
  const useMock = parseUseMock(import.meta.env.VITE_USE_MOCK_API);

  return {
    apiBaseUrl,
    wsBaseUrl,
    mode: useMock ? 'mock' : 'remote',
    useMock,
  };
}
