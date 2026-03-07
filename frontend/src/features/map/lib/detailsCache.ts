import type { AirbaseDetails } from '@/features/map/types';

type CacheEntry = {
  value: AirbaseDetails;
  createdAtMs: number;
};

const detailsCache = new Map<string, CacheEntry>();
const inFlightRequests = new Map<string, Promise<AirbaseDetails>>();

export function getCachedAirbaseDetails(key: string, ttlMs: number): AirbaseDetails | null {
  const entry = detailsCache.get(key);
  if (!entry) {
    return null;
  }

  if (Date.now() - entry.createdAtMs > ttlMs) {
    detailsCache.delete(key);
    return null;
  }

  return entry.value;
}

export function setCachedAirbaseDetails(key: string, details: AirbaseDetails): void {
  detailsCache.set(key, {
    value: details,
    createdAtMs: Date.now(),
  });
}

export function getInFlightAirbaseDetails(key: string): Promise<AirbaseDetails> | null {
  return inFlightRequests.get(key) ?? null;
}

export function setInFlightAirbaseDetails(key: string, request: Promise<AirbaseDetails>): void {
  inFlightRequests.set(key, request);
}

export function clearInFlightAirbaseDetails(key: string): void {
  inFlightRequests.delete(key);
}

export function clearAirbaseDetailsCache(): void {
  detailsCache.clear();
  inFlightRequests.clear();
}
