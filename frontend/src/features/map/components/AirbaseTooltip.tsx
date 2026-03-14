import type { AirbaseDetailsState } from '@/features/map/types';

type AirbaseTooltipProps = {
  airbaseId: string;
  airbaseName: string;
  leftPercent: number;
  topPercent: number;
  region?: string;
  detailsState: AirbaseDetailsState;
};

export function AirbaseTooltip({
  airbaseId,
  airbaseName,
  leftPercent,
  topPercent,
  region,
}: AirbaseTooltipProps) {
  return (
    <aside
      className="pointer-events-none absolute z-20 w-56 -translate-x-1/2 -translate-y-[calc(100%+10px)] rounded-md border border-border bg-surface p-2 shadow-lg"
      style={{ left: `${leftPercent}%`, top: `${topPercent}%` }}
      aria-live="polite"
    >
      <p className="m-0 mb-1 text-xs font-semibold uppercase tracking-[0.08em] text-primary">
        {airbaseName}
      </p>
      <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-2 gap-y-1 text-[11px]">
        <dt className="font-semibold text-text">Airbase ID</dt>
        <dd className="m-0 truncate text-text-muted">{airbaseId}</dd>
        {region ? (
          <>
            <dt className="font-semibold text-text">Region</dt>
            <dd className="m-0 truncate text-text-muted">{region}</dd>
          </>
        ) : null}
      </dl>
    </aside>
  );
}
