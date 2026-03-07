import { useCallback, useEffect, useRef, useState } from 'react';

import type { AirbaseDetails } from '@/features/map/types';
import { Card } from '@/features/ui';
import { ConstellationMap } from '@/features/map/components/ConstellationMap';
import { useApi } from '@/lib/api';

type SelectedAirbaseDetailsState =
  | { status: 'idle' }
  | { status: 'loading'; airbaseId: string }
  | { status: 'success'; details: AirbaseDetails }
  | { status: 'error'; airbaseId: string; message: string };

function toErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return 'Unable to load selected airbase details.';
}

function renderDetailsContent(selectedAirbaseId: string | null, state: SelectedAirbaseDetailsState) {
  if (!selectedAirbaseId || state.status === 'idle') {
    return <p className="m-0 text-xs text-text-muted">Select an airbase on the map to inspect details.</p>;
  }

  if (state.status === 'loading') {
    return <p className="m-0 text-xs text-text-muted">Loading details for {state.airbaseId}…</p>;
  }

  if (state.status === 'error') {
    return (
      <p className="m-0 text-xs text-red-700 dark:text-red-400">
        {state.airbaseId}: {state.message}
      </p>
    );
  }

  const entries = Object.entries(state.details).filter(([key]) => key !== 'id');

  if (entries.length === 0) {
    return <p className="m-0 text-xs text-text-muted">No additional details available.</p>;
  }

  return (
    <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-2 gap-y-1 text-xs">
      {entries.map(([key, value]) => (
        <div key={key} className="contents">
          <dt className="font-semibold text-text">{key}</dt>
          <dd className="m-0 truncate text-text-muted">{String(value)}</dd>
        </div>
      ))}
    </dl>
  );
}

export function MapPanel() {
  const { clients } = useApi();
  const [selectedAirbaseId, setSelectedAirbaseId] = useState<string | null>(null);
  const [selectedAirbaseDetailsState, setSelectedAirbaseDetailsState] =
    useState<SelectedAirbaseDetailsState>({ status: 'idle' });
  const detailsCacheRef = useRef(new Map<string, AirbaseDetails>());
  const requestSequenceRef = useRef(0);
  const activeAbortControllerRef = useRef<AbortController | null>(null);

  const cancelActiveRequest = useCallback(() => {
    if (activeAbortControllerRef.current) {
      activeAbortControllerRef.current.abort();
      activeAbortControllerRef.current = null;
    }
  }, []);

  useEffect(() => {
    return () => {
      cancelActiveRequest();
    };
  }, [cancelActiveRequest]);

  const handleSelectAirbase = useCallback(
    (airbaseId: string | null) => {
      setSelectedAirbaseId(airbaseId);
      cancelActiveRequest();

      if (!airbaseId) {
        setSelectedAirbaseDetailsState({ status: 'idle' });
        return;
      }

      const cachedDetails = detailsCacheRef.current.get(airbaseId);
      if (cachedDetails) {
        setSelectedAirbaseDetailsState({ status: 'success', details: cachedDetails });
        return;
      }

      setSelectedAirbaseDetailsState({ status: 'loading', airbaseId });

      const abortController = new AbortController();
      activeAbortControllerRef.current = abortController;
      requestSequenceRef.current += 1;
      const requestSequence = requestSequenceRef.current;

      clients.map
        .getAirbaseDetails(airbaseId, abortController.signal)
        .then((details) => {
          if (abortController.signal.aborted || requestSequence !== requestSequenceRef.current) {
            return;
          }

          detailsCacheRef.current.set(airbaseId, details);
          setSelectedAirbaseDetailsState({ status: 'success', details });
        })
        .catch((error: unknown) => {
          if (abortController.signal.aborted || requestSequence !== requestSequenceRef.current) {
            return;
          }

          setSelectedAirbaseDetailsState({
            status: 'error',
            airbaseId,
            message: toErrorMessage(error),
          });
        });
    },
    [cancelActiveRequest, clients.map],
  );

  return (
    <Card className="col-span-2" ariaLabel="Map section" title="Map">
      <div className="mt-3 grid items-start gap-3 min-[1100px]:grid-cols-[minmax(0,1fr)_15rem]">
        <ConstellationMap
          className="h-[58vh] min-h-[18rem] min-[900px]:h-[calc(100vh-10rem)]"
          selectedAirbaseId={selectedAirbaseId}
          onSelectAirbase={handleSelectAirbase}
        />

        <aside
          className="h-fit max-h-[calc(100vh-10rem)] overflow-auto rounded-lg border border-border bg-bg p-3"
          aria-label="Selected airbase details"
          aria-live="polite"
        >
          <div className="mb-2 flex items-center justify-between gap-2">
            <h3 className="m-0 text-sm font-semibold text-text">Airbase Details</h3>
            <button
              type="button"
              onClick={() => handleSelectAirbase(null)}
              disabled={!selectedAirbaseId}
              className="cursor-pointer rounded border border-border bg-surface px-2 py-1 text-[11px] text-text-muted transition-colors hover:text-text focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-bg disabled:cursor-not-allowed disabled:opacity-50"
            >
              Clear
            </button>
          </div>
          <p className="m-0 mb-2 text-[11px] text-text-muted">
            Selected: <strong>{selectedAirbaseId ?? 'none'}</strong>
          </p>
          {renderDetailsContent(selectedAirbaseId, selectedAirbaseDetailsState)}
        </aside>
      </div>
    </Card>
  );
}
