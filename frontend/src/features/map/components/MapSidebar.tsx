import type { ReactNode } from 'react';

import { AirbaseList } from '@/features/map/components/AirbaseList';
import {
  SelectionDrawer,
  type SelectedAirbaseDetailsState,
} from '@/features/map/components/SelectionDrawer';
import type { Airbase } from '@/features/map/types';

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

type SimulateActionsSectionProps = Pick<MapSidebarProps, 'onOpenSimulationSheet'>;

function SimulateActionsSection({ onOpenSimulationSheet }: SimulateActionsSectionProps) {
  return (
    <SidebarInsetSection className="shell-divider border-t pt-4">
      <button
        type="button"
        onClick={onOpenSimulationSheet}
        className="shell-button cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors"
      >
        Create
      </button>
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
          <SimulateActionsSection onOpenSimulationSheet={props.onOpenSimulationSheet} />
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
