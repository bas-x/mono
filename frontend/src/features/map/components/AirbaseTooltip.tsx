import type { AirbaseDetailsState } from '@/features/map/types';

type AirbaseTooltipProps = {
  airbaseId: string;
  leftPercent: number;
  topPercent: number;
  detailsState: AirbaseDetailsState;
};

function renderDetails(detailsState: AirbaseDetailsState) {
  if (detailsState.status === 'idle') {
    return <p className="m-0 text-[11px] text-text-muted">Hover to inspect airbase details.</p>;
  }

  if (detailsState.status === 'loading') {
    return <p className="m-0 text-[11px] text-text-muted">Loading details…</p>;
  }

  if (detailsState.status === 'error') {
    return <p className="m-0 text-[11px] text-red-700 dark:text-red-400">{detailsState.message}</p>;
  }

  const detailEntries = Object.entries(detailsState.details).filter(([key]) => key !== 'id').slice(0, 4);

  if (detailEntries.length === 0) {
    return <p className="m-0 text-[11px] text-text-muted">No additional details available.</p>;
  }

  return (
    <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-2 gap-y-1 text-[11px]">
      {detailEntries.map(([key, value]) => (
        <div key={key} className="contents">
          <dt className="font-semibold text-text">{key}</dt>
          <dd className="m-0 truncate text-text-muted">{String(value)}</dd>
        </div>
      ))}
    </dl>
  );
}

export function AirbaseTooltip({ airbaseId, leftPercent, topPercent, detailsState }: AirbaseTooltipProps) {
  return (
    <aside
      className="pointer-events-none absolute z-20 w-56 -translate-x-1/2 -translate-y-[calc(100%+10px)] rounded-md border border-border bg-surface p-2 shadow-lg"
      style={{ left: `${leftPercent}%`, top: `${topPercent}%` }}
      aria-live="polite"
    >
      <p className="m-0 mb-1 text-xs font-semibold uppercase tracking-[0.08em] text-primary">{airbaseId}</p>
      {renderDetails(detailsState)}
    </aside>
  );
}
