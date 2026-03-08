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
        className="group flex w-full cursor-pointer items-center text-left text-zinc-200 transition-colors disabled:cursor-not-allowed disabled:opacity-45"
      >
        <span className="inline-flex w-full items-center justify-between gap-x-0.5 rounded-md bg-red-50 px-2 py-1 text-xs font-medium text-red-700 inset-ring inset-ring-red-600/10 transition-colors dark:bg-red-400/10 dark:text-red-400 dark:inset-ring-red-400/20">
          <span className="truncate">Clear selection</span>
          <span className="group relative -mr-1 size-3.5 rounded-xs hover:bg-red-600/20 dark:hover:bg-red-500/30">
            <span className="sr-only">Remove</span>
            <svg
              viewBox="0 0 14 14"
              aria-hidden="true"
              className="size-3.5 stroke-red-600/50 transition-colors group-hover:stroke-red-600/75 dark:stroke-red-400 dark:group-hover:stroke-red-300"
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
        className="flex w-full cursor-pointer items-center text-left text-zinc-200 transition-colors"
      >
        {isSelected ? (
          <span className="inline-flex w-full items-center gap-x-1.5 rounded-md bg-green-100 px-2 py-1 text-xs font-medium text-green-700 dark:bg-green-400/10 dark:text-green-400">
            <svg
              viewBox="0 0 6 6"
              aria-hidden="true"
              className="size-1.5 fill-green-500 dark:fill-green-400"
            >
              <circle r="3" cx="3" cy="3" />
            </svg>
            {label}
          </span>
        ) : (
          <span className="inline-flex w-full truncate rounded-md px-2 py-1 text-xs font-medium transition-colors hover:bg-white/8">
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
    <div className="max-h-60 overflow-auto bg-white/7 p-1 shadow-[inset_0_1px_0_rgba(255,255,255,0.08)] backdrop-blur-sm">
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
