import { buildApiUrl, type HttpClient } from '@/lib/api/http/client';
import type {
  ApiAirbase,
  ApiAirbaseDetails,
  ApiAirbasePoint,
  MapServiceClient,
} from '@/lib/api/types';

type AirbasesResponse = {
  airbases?: ApiAirbase[];
};

function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value);
}

function normalizePoint(value: unknown): ApiAirbasePoint | null {
  if (!value || typeof value !== 'object') {
    return null;
  }

  const point = value as Record<string, unknown>;
  if (!isFiniteNumber(point.x) || !isFiniteNumber(point.y)) {
    return null;
  }

  return {
    x: point.x,
    y: point.y,
  };
}

function normalizeAirbase(value: unknown): ApiAirbase | null {
  if (!value || typeof value !== 'object') {
    return null;
  }

  const airbase = value as Record<string, unknown>;
  if (typeof airbase.id !== 'string' || !Array.isArray(airbase.area)) {
    return null;
  }

  const points = airbase.area.map(normalizePoint).filter((point): point is ApiAirbasePoint => point !== null);
  if (points.length < 3) {
    return null;
  }

  return {
    id: airbase.id,
    area: points,
    infoUrl: typeof airbase.infoUrl === 'string' ? airbase.infoUrl : undefined,
  };
}

function fallbackId(idOrUrl: string): string {
  const normalized = idOrUrl.replace(/^\/?map\/airbase\//, '').trim();
  if (normalized) {
    return normalized;
  }
  return 'unknown';
}

function normalizeDetails(idOrUrl: string, value: unknown): ApiAirbaseDetails {
  if (!value || typeof value !== 'object') {
    return { id: fallbackId(idOrUrl) };
  }

  const details = value as Record<string, unknown>;
  const id = typeof details.id === 'string' ? details.id : fallbackId(idOrUrl);

  return {
    ...details,
    id,
  };
}

function isAbsoluteUrl(value: string): boolean {
  return value.startsWith('http://') || value.startsWith('https://');
}

async function requestAbsoluteJson(url: string, signal?: AbortSignal): Promise<unknown> {
  const response = await fetch(url, {
    method: 'GET',
    headers: { Accept: 'application/json' },
    signal,
  });

  if (!response.ok) {
    const errorBody = await response.text();
    throw new Error(`Request failed for ${url}: ${response.status} ${errorBody}`);
  }

  return response.json();
}

export function createMapServiceClient(
  httpClient: HttpClient,
): MapServiceClient {
  return {
    async getAirbases(signal?: AbortSignal) {
      const response = await httpClient.requestJson<AirbasesResponse>('/map', { signal });
      if (!Array.isArray(response.airbases)) {
        return [];
      }

      return response.airbases
        .map(normalizeAirbase)
        .filter((airbase): airbase is ApiAirbase => airbase !== null);
    },

    async getAirbaseDetails(idOrUrl: string, signal?: AbortSignal) {
      if (isAbsoluteUrl(idOrUrl)) {
        const response = await requestAbsoluteJson(idOrUrl, signal);
        return normalizeDetails(idOrUrl, response);
      }

      const endpoint = idOrUrl.startsWith('/') ? idOrUrl : `/map/airbase/${encodeURIComponent(idOrUrl)}`;
      const response = await httpClient.requestJson<unknown>(endpoint, { signal });
      return normalizeDetails(idOrUrl, response);
    },
  };
}

export function resolveMapRequestUrl(apiBaseUrl: string, idOrUrl: string): string {
  if (isAbsoluteUrl(idOrUrl)) {
    return idOrUrl;
  }

  const endpoint = idOrUrl.startsWith('/') ? idOrUrl : `/map/airbase/${encodeURIComponent(idOrUrl)}`;
  return buildApiUrl(apiBaseUrl, endpoint);
}
