import { Drawer } from '@/features/ui/components/Drawer';
import type { AirbaseDetails } from '@/features/map/types';

export type SelectedAirbaseDetailsState =
  | { status: 'idle' }
  | { status: 'loading'; airbaseId: string }
  | { status: 'success'; details: AirbaseDetails }
  | { status: 'error'; airbaseId: string; message: string };

type SelectionDrawerProps = {
  viewMode: 'live' | 'simulate';
  selectedAirbaseId: string | null;
  selectedAirbaseDetailsState: SelectedAirbaseDetailsState;
  onClearSelection: () => void;
};

function renderDetailsContent(
  selectedAirbaseId: string | null,
  state: SelectedAirbaseDetailsState,
) {
  if (!selectedAirbaseId) {
    return (
      <p className="shell-text-muted m-0 text-xs">
        Select an airbase on the map to inspect details.
      </p>
    );
  }

  if (state.status === 'idle') {
    return (
      <p className="shell-text-muted m-0 text-xs">Loading details for {selectedAirbaseId}...</p>
    );
  }

  if (state.status === 'loading') {
    return (
      <p className="shell-text-muted m-0 text-xs">Loading details for {state.airbaseId}...</p>
    );
  }

  if (state.status === 'error') {
    return (
      <p className="m-0 text-xs text-primary">
        {state.airbaseId}: {state.message}
      </p>
    );
  }

  const name = typeof state.details.name === 'string' ? state.details.name : null;
  const region = typeof state.details.region === 'string' ? state.details.region : null;

  if (!name && !region) {
    return <p className="shell-text-muted m-0 text-xs">No additional details available.</p>;
  }

  return (
    <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-2 gap-y-1 text-xs">
      {name ? (
        <div className="contents">
          <dt className="font-semibold text-[color:var(--color-shell-text)]">Name</dt>
          <dd className="shell-text-muted m-0 truncate">{name}</dd>
        </div>
      ) : null}
      {region ? (
        <div className="contents">
          <dt className="font-semibold text-[color:var(--color-shell-text)]">Region</dt>
          <dd className="shell-text-muted m-0 truncate">{region}</dd>
        </div>
      ) : null}
    </dl>
  );
}

function SelectionDrawerContent({
  selectedAirbaseId,
  selectedAirbaseDetailsState,
  onClearSelection,
}: Omit<SelectionDrawerProps, 'viewMode'>) {
  const displayName = selectedAirbaseDetailsState.status === 'success'
    && typeof selectedAirbaseDetailsState.details.name === 'string'
    ? selectedAirbaseDetailsState.details.name
    : selectedAirbaseId;

  return (
    <div className="flex h-full min-h-0 flex-col gap-4">
      <div className="shell-divider space-y-2 border-b pb-4">
        <p className="shell-text-muted m-0 text-[0.65rem] font-semibold uppercase tracking-[0.22em]">
          Live Selection
        </p>
        <div className="space-y-1">
          <p className="m-0 text-lg font-semibold text-[color:var(--color-shell-text)]">
            {displayName ?? 'No base selected'}
          </p>
          <p className="shell-text-muted m-0 text-xs">
            Inspect current base details while keeping map controls in the sidebar.
          </p>
        </div>
      </div>

      <div className="min-h-0 flex-1">
        {renderDetailsContent(selectedAirbaseId, selectedAirbaseDetailsState)}
      </div>

      <div className="shell-divider border-t pt-4">
        <button
          type="button"
          onClick={onClearSelection}
          disabled={!selectedAirbaseId}
          className="shell-button cursor-pointer rounded-sm border px-3 py-2 text-sm font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-45"
        >
          Clear selection
        </button>
      </div>
    </div>
  );
}

export function SelectionDrawer({
  viewMode,
  selectedAirbaseId,
  selectedAirbaseDetailsState,
  onClearSelection,
}: SelectionDrawerProps) {
  const isOpen = viewMode === 'live' && selectedAirbaseId !== null;

  return (
    <Drawer
      isOpen={isOpen}
      onClose={onClearSelection}
      positionClassName="fixed inset-y-4 left-20 z-20 flex pl-4"
    >
      <SelectionDrawerContent
        selectedAirbaseId={selectedAirbaseId}
        selectedAirbaseDetailsState={selectedAirbaseDetailsState}
        onClearSelection={onClearSelection}
      />
    </Drawer>
  );
}
