import type { CSSProperties } from 'react';
import { useCallback, useEffect, useRef, useState } from 'react';

import { ConstellationMap } from '@/features/map/components/ConstellationMap';
import {
  MapSidebar,
  type SelectedAirbaseDetailsState,
  type ViewMode,
} from '@/features/map/components/MapSidebar';
import type { AirbaseDetails } from '@/features/map/types';
import { useApi } from '@/lib/api';

type ThemeStyle = CSSProperties & {
  '--color-map-surface': string;
  '--color-map-boundary': string;
  '--color-airbase-default-fill': string;
  '--color-airbase-default-stroke': string;
  '--color-airbase-hover': string;
  '--color-airbase-selected-border': string;
  '--color-airbase-selected-fill': string;
};

const MODE_THEME_STYLES: Record<ViewMode, ThemeStyle> = {
  live: {
    '--color-map-surface': 'oklch(77% 0.112 186)',
    '--color-map-boundary': 'oklch(36% 0.064 221)',
    '--color-airbase-default-fill': 'oklch(31% 0.118 256)',
    '--color-airbase-default-stroke': 'oklch(16% 0.031 258)',
    '--color-airbase-hover': 'oklch(89% 0.11 181)',
    '--color-airbase-selected-border': 'oklch(49% 0.157 34)',
    '--color-airbase-selected-fill': 'oklch(71% 0.173 63)',
  },
  simulate: {
    '--color-map-surface': 'oklch(73% 0.157 42)',
    '--color-map-boundary': 'oklch(35% 0.082 22)',
    '--color-airbase-default-fill': 'oklch(39% 0.138 336)',
    '--color-airbase-default-stroke': 'oklch(21% 0.045 334)',
    '--color-airbase-hover': 'oklch(84% 0.136 32)',
    '--color-airbase-selected-border': 'oklch(37% 0.143 257)',
    '--color-airbase-selected-fill': 'oklch(66% 0.169 275)',
  },
};

function toErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return 'Unable to load selected airbase details.';
}

export function MapPanel() {
  const { clients } = useApi();
  const [viewMode, setViewMode] = useState<ViewMode>('live');
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

  const handleModeChange = useCallback((nextMode: ViewMode) => {
    setViewMode((currentMode) => (currentMode === nextMode ? currentMode : nextMode));
  }, []);

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
    <section
      className="grid h-full min-h-0 min-w-0 overflow-hidden bg-zinc-950 min-[1040px]:grid-cols-[minmax(0,1fr)_10rem]"
      aria-label="Constellation map workspace"
      style={MODE_THEME_STYLES[viewMode]}
    >
      <div className="relative min-h-[55vh] min-w-0 bg-zinc-950 min-[1040px]:min-h-0">
        <ConstellationMap
          className="h-full min-h-full rounded-none border-0"
          mode={viewMode === 'live' ? 'live' : 'static'}
          selectedAirbaseId={selectedAirbaseId}
          onSelectAirbase={handleSelectAirbase}
        />
      </div>

      <MapSidebar
        viewMode={viewMode}
        selectedAirbaseId={selectedAirbaseId}
        selectedAirbaseDetailsState={selectedAirbaseDetailsState}
        onModeChange={handleModeChange}
        onClearSelection={() => handleSelectAirbase(null)}
      />
    </section>
  );
}
