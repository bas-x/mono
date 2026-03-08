import type { ReactNode } from 'react';

import { AirbaseList } from '@/features/map/components/AirbaseList';
import type { Airbase, AirbaseDetails } from '@/features/map/types';

export type ViewMode = 'live' | 'simulate';

export type SelectedAirbaseDetailsState =
  | { status: 'idle' }
  | { status: 'loading'; airbaseId: string }
  | { status: 'success'; details: AirbaseDetails }
  | { status: 'error'; airbaseId: string; message: string };

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

type SelectionSectionProps = Pick<
  MapSidebarProps,
  'selectedAirbaseId' | 'selectedAirbaseDetailsState' | 'onClearSelection'
>;

function noop() {}

function mergeClassNames(...parts: Array<string | undefined>) {
  return parts.filter(Boolean).join(' ');
}

function renderDetailsContent(
  selectedAirbaseId: string | null,
  state: SelectedAirbaseDetailsState,
) {
  if (!selectedAirbaseId || state.status === 'idle') {
    return (
      <p className="m-0 text-xs text-zinc-300/80">
        Select an airbase on the map to inspect details.
      </p>
    );
  }

  if (state.status === 'loading') {
    return <p className="m-0 text-xs text-zinc-300/80">Loading details for {state.airbaseId}...</p>;
  }

  if (state.status === 'error') {
    return (
      <p className="m-0 text-xs text-rose-300">
        {state.airbaseId}: {state.message}
      </p>
    );
  }

  const entries = Object.entries(state.details).filter(([key]) => key !== 'id');

  if (entries.length === 0) {
    return <p className="m-0 text-xs text-zinc-300/80">No additional details available.</p>;
  }

  return (
    <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-2 gap-y-1 text-xs">
      {entries.map(([key, value]) => (
        <div key={key} className="contents">
          <dt className="font-semibold text-zinc-50">{key}</dt>
          <dd className="m-0 truncate text-zinc-300/85">{String(value)}</dd>
        </div>
      ))}
    </dl>
  );
}

function buttonClassName(isActive: boolean) {
  return isActive
    ? 'bg-white text-zinc-950 shadow-[0_1px_0_rgba(255,255,255,0.4)]'
    : 'bg-transparent text-zinc-100 hover:bg-white/10';
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
        className="grid grid-cols-1 gap-1 rounded-sm border border-white/18 bg-white/6 p-1"
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
    <SidebarFlushSection className="border-t border-white/10 pt-4">
      <div className="grid gap-2 px-2">
        <button
          type="button"
          onClick={onResetView}
          className="cursor-pointer rounded-sm border border-white/18 bg-white/6 px-3 py-2 text-sm font-medium text-zinc-50 transition-colors hover:bg-white/12"
        >
          Full map
        </button>
        <button
          type="button"
          aria-pressed={isAirbaseListOpen}
          onClick={onToggleAirbaseList}
          className={`cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors ${
            isAirbaseListOpen
              ? 'border-white/40 bg-white/14 text-zinc-50'
              : 'border-white/18 bg-white/6 text-zinc-50 hover:bg-white/12'
          }`}
        >
          By base
        </button>
      </div>
      {isAirbaseListOpen ? (
        airbaseStatus === 'loading' ? (
          <div className="mx-0 rounded-2xl border border-white/12 bg-white/6 px-3 py-3 text-xs text-zinc-300/80">
            Loading bases...
          </div>
        ) : airbaseStatus === 'error' ? (
          <div className="mx-0 rounded-2xl border border-rose-300/30 bg-rose-500/10 px-3 py-3 text-xs text-rose-200">
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

function SimulateActionsSection() {
  return (
    <SidebarInsetSection className="border-t border-white/10 pt-4">
      <button
        type="button"
        onClick={noop}
        className="cursor-pointer rounded-sm border border-white/18 bg-white/6 px-3 py-2 text-sm font-medium text-zinc-50 transition-colors hover:bg-white/12"
      >
        Create
      </button>
    </SidebarInsetSection>
  );
}

function SelectionSection({
  selectedAirbaseId,
  selectedAirbaseDetailsState,
  onClearSelection,
}: SelectionSectionProps) {
  return (
    <SidebarInsetSection className="border-t border-white/10 pt-4">
      <div className="flex items-center justify-between gap-2">
        <div>
          <p className="m-0 text-[0.65rem] font-semibold uppercase tracking-[0.22em] text-zinc-300/70">
            Selection
          </p>
          <p className="m-0 mt-1 text-sm text-zinc-300/80">
            Selected: <strong className="text-zinc-50">{selectedAirbaseId ?? 'none'}</strong>
          </p>
        </div>
        <button
          type="button"
          onClick={onClearSelection}
          disabled={!selectedAirbaseId}
          className="cursor-pointer rounded-sm border border-white/18 bg-white/6 px-2 py-1 text-[11px] font-medium text-zinc-100 transition-colors hover:bg-white/12 disabled:cursor-not-allowed disabled:opacity-45"
        >
          Clear
        </button>
      </div>
      {renderDetailsContent(selectedAirbaseId, selectedAirbaseDetailsState)}
    </SidebarInsetSection>
  );
}

export function MapSidebar(props: MapSidebarProps) {
  return (
    <aside
      className="flex h-full min-h-0 w-full max-w-40 flex-col gap-5 overflow-auto border-l border-white/10 bg-mauve-900 py-4 text-zinc-50"
      aria-label="Map controls"
    >
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
        <SimulateActionsSection />
      )}

      <SelectionSection
        selectedAirbaseId={props.selectedAirbaseId}
        selectedAirbaseDetailsState={props.selectedAirbaseDetailsState}
        onClearSelection={props.onClearSelection}
      />
    </aside>
  );
}
