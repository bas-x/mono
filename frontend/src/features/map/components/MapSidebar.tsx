import { useState, useRef, useEffect } from 'react';
import type { ReactNode } from 'react';

import { AirbaseList } from '@/features/map/components/AirbaseList';
import {
  SelectionDrawer,
  type SelectedAirbaseDetailsState,
} from '@/features/map/components/SelectionDrawer';
import type { Airbase } from '@/features/map/types';
import type { SimulationState } from '@/features/simulation/hooks/useSimulation';

export type ViewMode = 'live' | 'simulate';

type MapSidebarProps = {
  airbases: Airbase[];
  airbaseStatus: 'loading' | 'success' | 'error';
  airbaseMessage?: string;
  viewMode: ViewMode;
  isAirbaseListOpen: boolean;
  selectedAirbaseId: string | null;
  selectedAirbaseDetailsState: SelectedAirbaseDetailsState;
  onModeChange: (nextMode: ViewMode) => void;
  onClearSelection: () => void;
  onResetView: () => void;
  onToggleAirbaseList: () => void;
  onSelectAirbaseFromList: (airbaseId: string) => void;
  onOpenSimulationSheet: () => void;
  onResetSimulation?: () => void;
  isSimulationRunning?: boolean;
  simulations?: Array<{ id: string }>;
  onLoadSimulation?: (id: string) => void;
  simulationState?: SimulationState;
};

type SectionProps = {
  children: ReactNode;
  className?: string;
};

type ModeSectionProps = Pick<MapSidebarProps, 'viewMode' | 'onModeChange'>;

type LiveActionsSectionProps = Pick<
  MapSidebarProps,
  | 'airbases'
  | 'airbaseStatus'
  | 'airbaseMessage'
  | 'isAirbaseListOpen'
  | 'selectedAirbaseId'
  | 'onClearSelection'
  | 'onResetView'
  | 'onToggleAirbaseList'
  | 'onSelectAirbaseFromList'
>;

function mergeClassNames(...parts: Array<string | undefined>) {
  return parts.filter(Boolean).join(' ');
}

function buttonClassName(isActive: boolean) {
  return isActive ? 'shell-button-active' : 'shell-button';
}

function SidebarInsetSection({ children, className }: SectionProps) {
  return <section className={mergeClassNames('space-y-3 px-2', className)}>{children}</section>;
}

function SidebarFlushSection({ children, className }: SectionProps) {
  return <section className={mergeClassNames('space-y-3', className)}>{children}</section>;
}

function ModeSection({ viewMode, onModeChange }: ModeSectionProps) {
  return (
    <SidebarInsetSection>
      <div
        className="shell-panel-soft shell-divider grid grid-cols-1 gap-1 rounded-sm border p-1"
        aria-label="Mode selection"
        role="group"
      >
        <button
          type="button"
          aria-pressed={viewMode === 'live'}
          onClick={() => onModeChange('live')}
          className={`${buttonClassName(viewMode === 'live')} cursor-pointer rounded-[calc(var(--radius-sm)-2px)] px-3 py-2 text-sm font-medium transition-colors`}
        >
          Live
        </button>
        <button
          type="button"
          aria-pressed={viewMode === 'simulate'}
          onClick={() => onModeChange('simulate')}
          className={`${buttonClassName(viewMode === 'simulate')} cursor-pointer rounded-[calc(var(--radius-sm)-2px)] px-3 py-2 text-sm font-medium transition-colors`}
        >
          Simulate
        </button>
      </div>
    </SidebarInsetSection>
  );
}

function LiveActionsSection({
  airbases,
  airbaseStatus,
  airbaseMessage,
  isAirbaseListOpen,
  selectedAirbaseId,
  onClearSelection,
  onResetView,
  onToggleAirbaseList,
  onSelectAirbaseFromList,
}: LiveActionsSectionProps) {
  return (
    <SidebarFlushSection className="shell-divider border-t pt-4">
      <div className="grid gap-2 px-2">
        <button
          type="button"
          onClick={onResetView}
          className="shell-button cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors"
        >
          Full map
        </button>
        <button
          type="button"
          aria-pressed={isAirbaseListOpen}
          onClick={onToggleAirbaseList}
          className={`cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors ${
            isAirbaseListOpen ? 'shell-button-active' : 'shell-button'
          }`}
        >
          By base
        </button>
      </div>
      {isAirbaseListOpen ? (
        airbaseStatus === 'loading' ? (
          <div className="shell-panel-soft shell-divider shell-text-muted mx-0 rounded-2xl border px-3 py-3 text-xs">
            Loading bases...
          </div>
        ) : airbaseStatus === 'error' ? (
          <div className="shell-error-surface mx-0 rounded-2xl border px-3 py-3 text-xs">
            {airbaseMessage ?? 'Unable to load airbases.'}
          </div>
        ) : (
          <AirbaseList
            airbases={airbases}
            selectedAirbaseId={selectedAirbaseId}
            onClearSelection={onClearSelection}
            onSelectAirbase={onSelectAirbaseFromList}
          />
        )
      ) : null}
    </SidebarFlushSection>
  );
}

type SimulateActionsSectionProps = Pick<
  MapSidebarProps,
  | 'onOpenSimulationSheet'
  | 'onResetSimulation'
  | 'isSimulationRunning'
  | 'simulations'
  | 'onLoadSimulation'
  | 'simulationState'
  | 'onSelectAirbaseFromList'
  | 'selectedAirbaseId'
>;

function SimulateActionsSection({
  onOpenSimulationSheet,
  onResetSimulation,
  isSimulationRunning,
  simulations = [],
  onLoadSimulation,
  simulationState,
  onSelectAirbaseFromList,
  selectedAirbaseId,
}: SimulateActionsSectionProps) {
  const [isConfirmingReset, setIsConfirmingReset] = useState(false);
  const [selectedAircraftId, setSelectedAircraftId] = useState<string | null>(null);
  const resetTimeoutRef = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (resetTimeoutRef.current !== null) {
        window.clearTimeout(resetTimeoutRef.current);
      }
    };
  }, []);

  const handleResetClick = () => {
    if (!isConfirmingReset) {
      setIsConfirmingReset(true);
      resetTimeoutRef.current = window.setTimeout(() => {
        setIsConfirmingReset(false);
      }, 3000);
    } else {
      if (resetTimeoutRef.current !== null) {
        window.clearTimeout(resetTimeoutRef.current);
      }
      setIsConfirmingReset(false);
      if (onResetSimulation) {
        onResetSimulation();
      }
    }
  };

  return (
    <SidebarInsetSection className="shell-divider border-t pt-4">
      <button
        type="button"
        onClick={onOpenSimulationSheet}
        className="shell-button cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors w-full"
      >
        Create
      </button>

      <div className="mt-4 flex flex-col gap-2">
        <label htmlFor="simulation-select" className="text-xs font-medium shell-text-muted">
          Current Simulations
        </label>
        {simulations.length === 0 ? (
          <div className="text-xs shell-text-muted italic px-1">No simulations found</div>
        ) : (
          <select
            id="simulation-select"
            className="shell-input w-full rounded-sm border px-2 py-1.5 text-sm"
            defaultValue=""
            onChange={(e) => {
              if (e.target.value && onLoadSimulation) {
                onLoadSimulation(e.target.value);
              }
            }}
          >
            <option value="" disabled>
              Select
            </option>
            {simulations.map((sim) => (
              <option key={sim.id} value={sim.id}>
                {sim.id}
              </option>
            ))}
          </select>
        )}
      </div>

      <button
        type="button"
        onClick={handleResetClick}
        disabled={simulations.length === 0 || !isSimulationRunning}
        className={
          isConfirmingReset
            ? 'cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors w-full mt-4 bg-red-600 text-white border-red-700 hover:bg-red-700'
            : 'shell-button cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors w-full mt-4 disabled:opacity-50 disabled:cursor-not-allowed'
        }
      >
        {isConfirmingReset ? 'Confirm reset' : 'Reset'}
      </button>

      {simulationState?.status === 'running' && (
        <div className="mt-6 flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <h3 className="text-sm font-semibold shell-text-primary">Simulation</h3>
            <div className="text-xs shell-text-muted">
              Status: Running | Tick: {simulationState.tick ?? 0}
            </div>
            <div className="text-xs shell-text-muted">
              Time: {simulationState.time ? new Date(simulationState.time).toLocaleString() : 'n/a'}
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <h3 className="text-sm font-semibold shell-text-primary">
              Airbases ({simulationState.airbases.length})
            </h3>
            <div className="flex flex-col gap-1 max-h-32 overflow-y-auto pr-1">
              {simulationState.airbases.map((base) => (
                <button
                  key={base.id}
                  type="button"
                  onClick={() => onSelectAirbaseFromList && onSelectAirbaseFromList(base.id)}
                  className={`cursor-pointer rounded-sm border px-2 py-1 text-left text-xs font-medium transition-colors w-full truncate ${
                    selectedAirbaseId === base.id ? 'shell-button-active' : 'shell-button'
                  }`}
                >
                  {base.region}
                </button>
              ))}
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <h3 className="text-sm font-semibold shell-text-primary">
              Aircraft ({simulationState.aircrafts.length})
            </h3>
            <div className="flex flex-col gap-1 max-h-48 overflow-y-auto pr-1">
              {simulationState.aircrafts.map((ac) => (
                <button
                  key={ac.tailNumber}
                  type="button"
                  onClick={() => setSelectedAircraftId(ac.tailNumber === selectedAircraftId ? null : ac.tailNumber)}
                  className={`cursor-pointer rounded-sm border px-2 py-1 text-left text-xs font-medium transition-colors w-full truncate flex justify-between ${
                    selectedAircraftId === ac.tailNumber ? 'shell-button-active' : 'shell-button'
                  }`}
                >
                  <span>{ac.tailNumber.substring(0, 8)}</span>
                  <span className="opacity-70">{ac.state}</span>
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </SidebarInsetSection>
  );
}

export function MapSidebar(props: MapSidebarProps) {
  return (
    <aside
      className="shell-panel relative h-full min-h-0 w-full max-w-40"
      aria-label="Map controls"
    >
      <div className="flex h-full min-h-0 flex-col gap-5 overflow-y-auto py-4">
        <ModeSection viewMode={props.viewMode} onModeChange={props.onModeChange} />

        {props.viewMode === 'live' ? (
          <LiveActionsSection
            airbases={props.airbases}
            airbaseStatus={props.airbaseStatus}
            airbaseMessage={props.airbaseMessage}
            isAirbaseListOpen={props.isAirbaseListOpen}
            selectedAirbaseId={props.selectedAirbaseId}
            onClearSelection={props.onClearSelection}
            onResetView={props.onResetView}
            onToggleAirbaseList={props.onToggleAirbaseList}
            onSelectAirbaseFromList={props.onSelectAirbaseFromList}
          />
        ) : (
          <SimulateActionsSection
            onOpenSimulationSheet={props.onOpenSimulationSheet}
            onResetSimulation={props.onResetSimulation}
            isSimulationRunning={props.isSimulationRunning}
            simulations={props.simulations}
            onLoadSimulation={props.onLoadSimulation}
            simulationState={props.simulationState}
            onSelectAirbaseFromList={props.onSelectAirbaseFromList}
            selectedAirbaseId={props.selectedAirbaseId}
          />
        )}
      </div>

      <SelectionDrawer
        viewMode={props.viewMode}
        selectedAirbaseId={props.selectedAirbaseId}
        selectedAirbaseDetailsState={props.selectedAirbaseDetailsState}
        onClearSelection={props.onClearSelection}
      />
    </aside>
  );
}
