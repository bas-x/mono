import type { AirbaseDetails } from '@/features/map/types';

export type ViewMode = 'live' | 'simulate';

export type SelectedAirbaseDetailsState =
  | { status: 'idle' }
  | { status: 'loading'; airbaseId: string }
  | { status: 'success'; details: AirbaseDetails }
  | { status: 'error'; airbaseId: string; message: string };

type MapSidebarProps = {
  viewMode: ViewMode;
  selectedAirbaseId: string | null;
  selectedAirbaseDetailsState: SelectedAirbaseDetailsState;
  onModeChange: (nextMode: ViewMode) => void;
  onClearSelection: () => void;
};

const MODE_ACTIONS: Record<ViewMode, string[]> = {
  live: ['Full map', 'By base'],
  simulate: ['Create'],
};

function noop() {}

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

export function MapSidebar({
  viewMode,
  selectedAirbaseId,
  selectedAirbaseDetailsState,
  onModeChange,
  onClearSelection,
}: MapSidebarProps) {
  return (
    <aside
      className="flex h-full min-h-0 w-full max-w-40 flex-col gap-5 overflow-auto border-l border-white/10 bg-mauve-900 px-4 py-4 text-zinc-50"
      aria-label="Map controls"
    >
      <section className="space-y-3">
        <div
          className="grid grid-cols-1 gap-1 rounded-sm border border-white/18 bg-white/6 p-1"
          aria-label="Mode selection"
          role="group"
        >
          <button
            type="button"
            aria-pressed={viewMode === 'live'}
            onClick={() => onModeChange('live')}
            className={`${buttonClassName(viewMode === 'live')} rounded-[calc(var(--radius-sm)-2px)] px-3 py-2 text-sm font-medium transition-colors`}
          >
            Live
          </button>
          <button
            type="button"
            aria-pressed={viewMode === 'simulate'}
            onClick={() => onModeChange('simulate')}
            className={`${buttonClassName(viewMode === 'simulate')} rounded-[calc(var(--radius-sm)-2px)] px-3 py-2 text-sm font-medium transition-colors`}
          >
            Simulate
          </button>
        </div>
      </section>

      <section className="space-y-3 border-t border-white/10 pt-4">
        <div className="grid gap-2">
          {MODE_ACTIONS[viewMode].map((label) => (
            <button
              key={label}
              type="button"
              onClick={noop}
              className="rounded-sm border border-white/18 bg-white/6 px-3 py-2 text-sm font-medium text-zinc-50 transition-colors hover:bg-white/12"
            >
              {label}
            </button>
          ))}
        </div>
      </section>

      <section className="space-y-3 border-t border-white/10 pt-4">
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
            className="rounded-sm border border-white/18 bg-white/6 px-2 py-1 text-[11px] font-medium text-zinc-100 transition-colors hover:bg-white/12 disabled:cursor-not-allowed disabled:opacity-45"
          >
            Clear
          </button>
        </div>
        {renderDetailsContent(selectedAirbaseId, selectedAirbaseDetailsState)}
      </section>
    </aside>
  );
}
