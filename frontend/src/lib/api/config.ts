import type { ApiConfig } from '@/lib/api/types';

export const SIMULATION_WS_PATH = '/ws/simulation';

const DEFAULT_API_BASE_URL = 'https://basex.shigure.joshuadematas.me';
const DEFAULT_WS_BASE_URL = 'wss://basex.shigure.joshuadematas.me';
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

export function parseApiConfigFromEnv(): ApiConfig {
  return {
    apiBaseUrl: normalizeEnvString(import.meta.env.VITE_API_BASE_URL) || DEFAULT_API_BASE_URL,
    wsBaseUrl: normalizeEnvString(import.meta.env.VITE_WS_BASE_URL) || DEFAULT_WS_BASE_URL,
    useMock: parseUseMock(import.meta.env.VITE_USE_MOCK_API),
  };
}
