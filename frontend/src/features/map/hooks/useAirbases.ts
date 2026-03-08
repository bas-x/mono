import { useEffect, useState } from 'react';

import type { MapServiceClient } from '@/lib/api';

import { MOCK_AIRBASES } from '@/features/map/fixtures/airbases.mock';
import type { Airbase, MapDataSource } from '@/features/map/types';

type AirbaseLoadState =
  | { status: 'loading'; airbases: Airbase[] }
  | { status: 'success'; airbases: Airbase[] }
  | { status: 'error'; airbases: Airbase[]; message: string };

type UseAirbasesOptions = {
  mapClient: MapServiceClient;
  dataSource: MapDataSource;
};

async function loadAirbases(
  mapClient: MapServiceClient,
  dataSource: MapDataSource,
  signal: AbortSignal,
): Promise<Airbase[]> {
  if (dataSource === 'mock') {
    return Promise.resolve(MOCK_AIRBASES);
  }

  if (dataSource === 'api') {
    return mapClient.getAirbases(signal);
  }

  try {
    return await mapClient.getAirbases(signal);
  } catch {
    return MOCK_AIRBASES;
  }
}

function toErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return 'Unable to load airbase map data.';
}

export function useAirbases({ mapClient, dataSource }: UseAirbasesOptions): AirbaseLoadState {
  const [state, setState] = useState<AirbaseLoadState>({
    status: 'loading',
    airbases: [],
  });

  useEffect(() => {
    const abortController = new AbortController();
    Promise.resolve().then(() => {
      if (abortController.signal.aborted) {
        return;
      }
      setState({ status: 'loading', airbases: [] });
    });

    loadAirbases(mapClient, dataSource, abortController.signal)
      .then((airbases) => {
        if (abortController.signal.aborted) {
          return;
        }

        setState({ status: 'success', airbases });
      })
      .catch((error: unknown) => {
        if (abortController.signal.aborted) {
          return;
        }

        setState({
          status: 'error',
          airbases: [],
          message: toErrorMessage(error),
        });
      });

    return () => {
      abortController.abort();
    };
  }, [dataSource, mapClient]);

  return state;
}
