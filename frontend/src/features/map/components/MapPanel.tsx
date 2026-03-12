import type { CSSProperties } from 'react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { ConstellationMap } from '@/features/map/components/ConstellationMap';
import { MapSidebar, type ViewMode } from '@/features/map/components/MapSidebar';
import type { SelectedAirbaseDetailsState } from '@/features/map/components/SelectionDrawer';
import { useAirbases } from '@/features/map/hooks/useAirbases';
import { createFocusedViewBox } from '@/features/map/lib/geometry';
import { getPlacementBounds, resolveActivePlacementSources } from '@/features/map/lib/placement';
import { SimulationSetupSheet } from '@/features/simulation/components/SimulationSetupSheet';
import { useSimulation } from '@/features/simulation/hooks/useSimulation';
import {
  DEFAULT_SIMULATION_SETUP_FORM_VALUES,
  type SimulationSetupFormValues,
} from '@/features/simulation/types';
import {
  DEFAULT_MAP_VIEW_BOX,
  type Airbase,
  type AirbasePlacementSource,
  type AirbaseDetails,
  type MapDataSource,
  type MapViewBox,
} from '@/features/map/types';
import { useApi } from '@/lib/api';

import { SimulationInfoCard } from '@/features/simulation/components/SimulationInfoCard';
import { SimulationTimeline } from '@/features/simulation/components/timeline/SimulationTimeline';

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
  const dataSource: MapDataSource = 'mock';
  const [viewMode, setViewMode] = useState<ViewMode>('live');
  const [isAirbaseListOpen, setIsAirbaseListOpen] = useState(false);
  const [mapViewBox, setMapViewBox] = useState<MapViewBox>({ ...DEFAULT_MAP_VIEW_BOX });
  const [selectedAirbaseId, setSelectedAirbaseId] = useState<string | null>(null);
  const [isSimulationSheetOpen, setIsSimulationSheetOpen] = useState(false);
  const [simulationSetupValues, setSimulationSetupValues] = useState<SimulationSetupFormValues>(
    DEFAULT_SIMULATION_SETUP_FORM_VALUES,
  );
  const [selectedAirbaseDetailsState, setSelectedAirbaseDetailsState] =
    useState<SelectedAirbaseDetailsState>({ status: 'idle' });
  const {
    state: simulationState,
    setPlaybackTick,
    simulations,
    loadSimulation,
    createSimulation,
    reset: resetSimulation,
    triggerReset,
  } = useSimulation();
  const detailsCacheRef = useRef(new Map<string, AirbaseDetails>());
  const requestSequenceRef = useRef(0);
  const activeAbortControllerRef = useRef<AbortController | null>(null);
  const airbaseState = useAirbases({ mapClient: clients.map, dataSource });
  const activePlacementSources = useMemo<AirbasePlacementSource[]>(() => {
    if (viewMode === 'simulate' && simulationState.status !== 'running') {
      return [];
    }

    return resolveActivePlacementSources({
      viewMode,
      liveAirbases: airbaseState.airbases,
      simulationAirbases: simulationState.status === 'running' ? simulationState.airbases : [],
      hasRunningSimulation: simulationState.status === 'running',
    });
  }, [airbaseState.airbases, simulationState, viewMode]);

  const activePlacementSourceById = useMemo(() => {
    const byId = new Map<string, AirbasePlacementSource>();

    for (const source of activePlacementSources) {
      byId.set(source.id, source);
    }

    return byId;
  }, [activePlacementSources]);

  const activeRenderableAirbases = useMemo<Airbase[]>(() => {
    return activePlacementSources.map((source) => {
      if ('area' in source) {
        return source;
      }

      const bounds = getPlacementBounds(source);
      return {
        id: source.id,
        area: [
          { x: bounds.minX, y: bounds.minY },
          { x: bounds.maxX, y: bounds.minY },
          { x: bounds.maxX, y: bounds.maxY },
          { x: bounds.minX, y: bounds.maxY },
        ],
      };
    });
  }, [activePlacementSources]);

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

  const focusAirbase = useCallback(
    (airbaseId: string) => {
      const source = activePlacementSourceById.get(airbaseId);
      if (!source) {
        return;
      }

      const bounds = getPlacementBounds(source);
      setMapViewBox(createFocusedViewBox(bounds, DEFAULT_MAP_VIEW_BOX));
    },
    [activePlacementSourceById],
  );

  const handleResetView = useCallback(() => {
    setMapViewBox({ ...DEFAULT_MAP_VIEW_BOX });
  }, []);

  const resetWorkspaceState = useCallback(() => {
    cancelActiveRequest();
    detailsCacheRef.current.clear();
    setIsAirbaseListOpen(false);
    setSelectedAirbaseId(null);
    setSelectedAirbaseDetailsState({ status: 'idle' });
    setMapViewBox({ ...DEFAULT_MAP_VIEW_BOX });
  }, [cancelActiveRequest]);

  const handleModeChange = useCallback(
    (nextMode: ViewMode) => {
      if (viewMode === nextMode) {
        return;
      }

      setIsSimulationSheetOpen(false);

      if (nextMode === 'simulate') {
        resetWorkspaceState();
      } else {
        detailsCacheRef.current.clear();
        setSelectedAirbaseDetailsState({ status: 'idle' });
        resetSimulation();
      }

      setViewMode(nextMode);
    },
    [resetWorkspaceState, resetSimulation, viewMode],
  );

  const handleToggleAirbaseList = useCallback(() => {
    setIsAirbaseListOpen((current) => {
      const next = !current;

      if (next && selectedAirbaseId) {
        focusAirbase(selectedAirbaseId);
      }

      return next;
    });
  }, [focusAirbase, selectedAirbaseId]);

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

  const handleClearSelection = useCallback(() => {
    handleSelectAirbase(null);
    setMapViewBox({ ...DEFAULT_MAP_VIEW_BOX });
  }, [handleSelectAirbase]);

  const handleSelectAirbaseFromList = useCallback(
    (airbaseId: string) => {
      if (selectedAirbaseId === airbaseId) {
        handleClearSelection();
        return;
      }

      handleSelectAirbase(airbaseId);
      focusAirbase(airbaseId);
    },
    [focusAirbase, handleClearSelection, handleSelectAirbase, selectedAirbaseId],
  );

  const handleOpenSimulationSheet = useCallback(() => {
    setIsSimulationSheetOpen(true);
  }, []);

  const handleCloseSimulationSheet = useCallback(() => {
    setIsSimulationSheetOpen(false);
  }, []);

  const handleSubmitSimulationSetup = useCallback(
    async (values: SimulationSetupFormValues) => {
      setSimulationSetupValues(values);
      const success = await createSimulation(values.seedHex);
      if (success) {
        setIsSimulationSheetOpen(false);
      }
    },
    [createSimulation],
  );

  return (
    <>
      <section
        className="grid h-full min-h-0 min-w-0 overflow-hidden bg-bg min-[1040px]:grid-cols-[minmax(0,1fr)_10rem]"
        aria-label="Constellation map workspace"
        style={MODE_THEME_STYLES[viewMode]}
      >
        <div className="relative min-h-[55vh] min-w-0 bg-bg min-[1040px]:min-h-0">
          <ConstellationMap
            className="h-full min-h-full rounded-none border-0"
            dataSource={viewMode === 'simulate' ? 'api' : dataSource}
            mode={viewMode === 'live' ? 'live' : 'static'}
            placementSources={activePlacementSources}
            selectedAirbaseId={selectedAirbaseId}
            viewBox={mapViewBox}
            onSelectAirbase={handleSelectAirbase}
            aircraftPositions={simulationState.status === 'running' ? simulationState.aircraftPositions : undefined}
          />
        </div>

        <MapSidebar
          airbases={
            viewMode === 'simulate' && simulationState.status === 'running'
              ? activeRenderableAirbases
              : airbaseState.airbases
          }
          airbaseStatus={
            viewMode === 'simulate'
              ? simulationState.status === 'running'
                ? 'success'
                : 'loading'
              : airbaseState.status
          }
          airbaseMessage={
            viewMode === 'simulate'
              ? simulationState.status === 'error'
                ? simulationState.message
                : undefined
              : airbaseState.status === 'error'
                ? airbaseState.message
                : undefined
          }
          viewMode={viewMode}
          isAirbaseListOpen={isAirbaseListOpen}
          selectedAirbaseId={selectedAirbaseId}
          selectedAirbaseDetailsState={selectedAirbaseDetailsState}
          onModeChange={handleModeChange}
          onClearSelection={handleClearSelection}
          onResetView={handleResetView}
          onToggleAirbaseList={handleToggleAirbaseList}
          onSelectAirbaseFromList={handleSelectAirbaseFromList}
          onOpenSimulationSheet={handleOpenSimulationSheet}
          onResetSimulation={triggerReset}
          isSimulationRunning={simulationState.status === 'running'}
          simulations={simulations}
          onLoadSimulation={loadSimulation}
        />
      </section>

      {viewMode === 'simulate' && simulationState.status === 'running' && (
        <>
          <SimulationInfoCard simulationState={simulationState} simulations={simulations} />
          <SimulationTimeline 
            simulationId={simulationState.simulationId} 
            simulationState={simulationState}
            setPlaybackTick={setPlaybackTick}
          />
        </>
      )}

      <SimulationSetupSheet
        isOpen={viewMode === 'simulate' && isSimulationSheetOpen}
        onClose={handleCloseSimulationSheet}
        defaultValues={simulationSetupValues}
        onSubmit={handleSubmitSimulationSetup}
      />
    </>
  );
}
