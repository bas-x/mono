import { useEffect, useMemo, useRef, useState } from 'react';

import type { MapServiceClient } from '@/lib/api';

import { resolveMockAirbaseDetails } from '@/features/map/fixtures/airbases.mock';
import {
  clearInFlightAirbaseDetails,
  getCachedAirbaseDetails,
  getInFlightAirbaseDetails,
  setCachedAirbaseDetails,
  setInFlightAirbaseDetails,
} from '@/features/map/lib/detailsCache';
import type { Airbase, AirbaseDetailsState, MapDataSource } from '@/features/map/types';

type UseAirbaseDetailsOptions = {
  mapClient: MapServiceClient;
  hoveredAirbase: Pick<Airbase, 'id' | 'infoUrl'> | null;
  dataSource: MapDataSource;
  enabled: boolean;
  debounceMs: number;
  cacheTtlMs: number;
};

function toLookupKey(airbase: Pick<Airbase, 'id' | 'infoUrl'>): string {
  return airbase.infoUrl ?? airbase.id;
}

function toErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return 'Unable to load airbase details.';
}

async function loadFromSource(
  mapClient: MapServiceClient,
  dataSource: MapDataSource,
  airbase: Pick<Airbase, 'id' | 'infoUrl'>,
  signal: AbortSignal,
) {
  const lookupKey = toLookupKey(airbase);

  if (dataSource === 'mock') {
    return Promise.resolve(resolveMockAirbaseDetails(lookupKey));
  }

  if (dataSource === 'api') {
    return mapClient.getAirbaseDetails(lookupKey, signal);
  }

  try {
    return await mapClient.getAirbaseDetails(lookupKey, signal);
  } catch {
    return resolveMockAirbaseDetails(lookupKey);
  }
}

export function useAirbaseDetails({
  mapClient,
  hoveredAirbase,
  dataSource,
  enabled,
  debounceMs,
  cacheTtlMs,
}: UseAirbaseDetailsOptions): AirbaseDetailsState {
  const [state, setState] = useState<AirbaseDetailsState>({ status: 'idle' });
  const sequenceRef = useRef(0);

  const lookupKey = useMemo(() => {
    if (!hoveredAirbase) {
      return null;
    }

    return toLookupKey(hoveredAirbase);
  }, [hoveredAirbase]);

  useEffect(() => {
    if (!enabled || !hoveredAirbase || !lookupKey) {
      return;
    }

    const cachedDetails = getCachedAirbaseDetails(lookupKey, cacheTtlMs);
    if (cachedDetails) {
      Promise.resolve().then(() => {
        setState({ status: 'success', details: cachedDetails });
      });
      return;
    }

    const sequence = sequenceRef.current + 1;
    sequenceRef.current = sequence;

    const abortController = new AbortController();
    const timer = setTimeout(() => {
      if (abortController.signal.aborted) {
        return;
      }

      setState({ status: 'loading' });

      const inFlightRequest = getInFlightAirbaseDetails(lookupKey);
      if (inFlightRequest) {
        inFlightRequest
          .then((details) => {
            if (sequenceRef.current !== sequence || abortController.signal.aborted) {
              return;
            }
            setState({ status: 'success', details });
          })
          .catch((error: unknown) => {
            if (sequenceRef.current !== sequence || abortController.signal.aborted) {
              return;
            }
            setState({ status: 'error', message: toErrorMessage(error) });
          });
        return;
      }

      const request = loadFromSource(
        mapClient,
        dataSource,
        hoveredAirbase,
        abortController.signal,
      )
        .then((details) => {
          setCachedAirbaseDetails(lookupKey, details);
          return details;
        })
        .finally(() => {
          clearInFlightAirbaseDetails(lookupKey);
        });

      setInFlightAirbaseDetails(lookupKey, request);

      request
        .then((details) => {
          if (sequenceRef.current !== sequence || abortController.signal.aborted) {
            return;
          }
          setState({ status: 'success', details });
        })
        .catch((error: unknown) => {
          if (sequenceRef.current !== sequence || abortController.signal.aborted) {
            return;
          }
          setState({ status: 'error', message: toErrorMessage(error) });
        });
    }, debounceMs);

    return () => {
      abortController.abort();
      clearTimeout(timer);
    };
  }, [cacheTtlMs, dataSource, debounceMs, enabled, hoveredAirbase, lookupKey, mapClient]);

  if (!enabled || !hoveredAirbase || !lookupKey) {
    return { status: 'idle' };
  }

  return state;
}
