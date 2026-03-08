import { useMemo } from 'react';

import type { Airbase } from '@/features/map/types';

type AirbaseListProps = {
  airbases: Airbase[];
  selectedAirbaseId: string | null;
  onClearSelection: () => void;
  onSelectAirbase: (airbaseId: string) => void;
};

type ClearSelectionOptionProps = {
  disabled: boolean;
  onClearSelection: () => void;
};

type AirbaseOptionProps = {
  airbase: Airbase;
  isSelected: boolean;
  onClearSelection: () => void;
  onSelectAirbase: (airbaseId: string) => void;
};

function toDisplayName(airbaseId: string): string {
  return airbaseId
    .split(/[-_]/)
    .filter(Boolean)
    .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
    .join(' ');
}

function ClearSelectionOption({ disabled, onClearSelection }: ClearSelectionOptionProps) {
  return (
    <li>
      <button
        type="button"
        onClick={onClearSelection}
        disabled={disabled}
        className="group flex w-full cursor-pointer items-center text-left transition-colors disabled:cursor-not-allowed disabled:opacity-45"
      >
        <span className="shell-clear-pill inline-flex w-full items-center justify-between gap-x-0.5 rounded-md px-2 py-1 text-xs font-medium transition-colors">
          <span className="truncate">Clear selection</span>
          <span className="group relative -mr-1 size-3.5 rounded-xs shell-list-hover">
            <span className="sr-only">Remove</span>
            <svg
              viewBox="0 0 14 14"
              aria-hidden="true"
              className="shell-clear-icon size-3.5 transition-colors"
              fill="none"
            >
              <path d="M4 4l6 6m0-6l-6 6" />
            </svg>
            <span className="absolute -inset-1" aria-hidden="true"></span>
          </span>
        </span>
      </button>
    </li>
  );
}

function AirbaseOption({
  airbase,
  isSelected,
  onClearSelection,
  onSelectAirbase,
}: AirbaseOptionProps) {
  const label = toDisplayName(airbase.id);

  return (
    <li>
      <button
        type="button"
        onClick={() => {
          if (isSelected) {
            onClearSelection();
            return;
          }

          onSelectAirbase(airbase.id);
        }}
        className="flex w-full cursor-pointer items-center text-left transition-colors"
      >
        {isSelected ? (
          <span className="shell-list-selected inline-flex w-full items-center gap-x-1.5 rounded-md px-2 py-1 text-xs font-medium">
            <svg
              viewBox="0 0 6 6"
              aria-hidden="true"
              className="shell-list-selected-dot size-1.5"
            >
              <circle r="3" cx="3" cy="3" />
            </svg>
            {label}
          </span>
        ) : (
          <span className="shell-text-muted shell-list-hover inline-flex w-full truncate rounded-md px-2 py-1 text-xs font-medium transition-colors">
            {label}
          </span>
        )}
      </button>
    </li>
  );
}

export function AirbaseList({
  airbases,
  selectedAirbaseId,
  onClearSelection,
  onSelectAirbase,
}: AirbaseListProps) {
  const sortedAirbases = useMemo(() => {
    return [...airbases].sort((left, right) => left.id.localeCompare(right.id));
  }, [airbases]);

  return (
    <div className="shell-panel-soft max-h-60 overflow-auto p-1 shadow-[inset_0_1px_0_rgba(255,255,255,0.08)]">
      <ul className="m-0 list-none space-y-1 p-0" aria-label="Airbases">
        <ClearSelectionOption
          disabled={!selectedAirbaseId}
          onClearSelection={onClearSelection}
        />
        {sortedAirbases.map((airbase) => (
          <AirbaseOption
            key={airbase.id}
            airbase={airbase}
            isSelected={airbase.id === selectedAirbaseId}
            onClearSelection={onClearSelection}
            onSelectAirbase={onSelectAirbase}
          />
        ))}
      </ul>
    </div>
  );
}
