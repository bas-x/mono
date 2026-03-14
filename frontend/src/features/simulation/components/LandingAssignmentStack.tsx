import { createPortal } from 'react-dom';

import type { LandingAssignmentEvent, SimulationEvent } from '@/lib/api/types';

type LandingAssignmentStackProps = {
  events: SimulationEvent[];
  portalRoot: Element | null;
};

type CapabilityTone = {
  cardClass: string;
  badgeClass: string;
};

const CAPABILITY_TONES: Record<string, CapabilityTone> = {
  fuel: {
    cardClass: 'border-cyan-300/20 bg-cyan-400/10',
    badgeClass: 'border-cyan-300/25 bg-cyan-300/20 text-cyan-100',
  },
  munitions: {
    cardClass: 'border-rose-300/20 bg-rose-400/10',
    badgeClass: 'border-rose-300/25 bg-rose-300/20 text-rose-100',
  },
  crew_support: {
    cardClass: 'border-emerald-300/20 bg-emerald-400/10',
    badgeClass: 'border-emerald-300/25 bg-emerald-300/20 text-emerald-100',
  },
  ground_support: {
    cardClass: 'border-amber-300/20 bg-amber-400/10',
    badgeClass: 'border-amber-300/25 bg-amber-300/20 text-amber-100',
  },
  repairs: {
    cardClass: 'border-orange-300/20 bg-orange-400/10',
    badgeClass: 'border-orange-300/25 bg-orange-300/20 text-orange-100',
  },
  maintenance: {
    cardClass: 'border-violet-300/20 bg-violet-400/10',
    badgeClass: 'border-violet-300/25 bg-violet-300/20 text-violet-100',
  },
};

function isLandingAssignmentEvent(event: SimulationEvent): event is LandingAssignmentEvent {
  return (
    event.type === 'landing_assignment' &&
    typeof event.tailNumber === 'string' &&
    typeof event.baseId === 'string' &&
    Array.isArray(event.needs) &&
    typeof event.capabilities === 'object' &&
    event.capabilities !== null
  );
}

function toTitleCase(value: string) {
  return value
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}

export function getTopNeedCapabilities(event: LandingAssignmentEvent): Array<{
  capability: string;
  severity: number;
}> {
  const ranked = event.needs
    .map((need) => ({
      capability: need.requiredCapability,
      severity: need.severity,
    }))
    .sort((left, right) => right.severity - left.severity);

  const seen = new Set<string>();
  const topCapabilities: Array<{ capability: string; severity: number }> = [];

  for (const entry of ranked) {
    if (seen.has(entry.capability)) {
      continue;
    }

    seen.add(entry.capability);
    topCapabilities.push(entry);

    if (topCapabilities.length === 3) {
      break;
    }
  }

  if (topCapabilities.length > 0) {
    return topCapabilities;
  }

  return Object.keys(event.capabilities)
    .slice(0, 3)
    .map((capability) => ({ capability, severity: 0 }));
}

export function getLandingAssignmentPrimaryCapability(event: LandingAssignmentEvent): string {
  if (event.needs[0]?.requiredCapability) {
    return event.needs[0].requiredCapability;
  }

  const [firstCapability] = Object.keys(event.capabilities);
  return firstCapability ?? 'assignment';
}

export function getLandingAssignmentTone(event: LandingAssignmentEvent): CapabilityTone {
  return (
    CAPABILITY_TONES[getLandingAssignmentPrimaryCapability(event)] ?? {
      cardClass: 'border-white/10 bg-white/10',
      badgeClass: 'border-white/10 bg-white/10 text-white/90',
    }
  );
}

function getAssignmentSourceBadgeClass(source: LandingAssignmentEvent['source']) {
  return source === 'human'
    ? 'bg-emerald-950 text-emerald-50 ring-1 ring-emerald-50/10'
    : 'bg-slate-950 text-slate-50 ring-1 ring-slate-50/10';
}

function getAssignmentSourceLabel(source: LandingAssignmentEvent['source']) {
  return source === 'human' ? 'Manual' : 'Auto';
}

function formatTailNumber(tailNumber: string) {
  return tailNumber.slice(0, 8);
}

function formatBaseId(baseId: string) {
  return baseId.slice(0, 6);
}

function formatNeedCount(count: number) {
  if (count === 0) {
    return 'No active service needs recorded';
  }

  if (count === 1) {
    return '1 active service need';
  }

  return `${count} active service needs`;
}

export function LandingAssignmentStack({ events, portalRoot }: LandingAssignmentStackProps) {
  if (typeof document === 'undefined' || !portalRoot) {
    return null;
  }

  const assignmentEvents = events.filter(isLandingAssignmentEvent).slice(-5).reverse();

  if (assignmentEvents.length === 0) {
    return null;
  }

  return createPortal(
    <div className="pointer-events-none absolute inset-4 z-20 flex items-start justify-end">
      <div className="pointer-events-auto flex max-h-[min(70vh,36rem)] w-full max-w-[350px] flex-col gap-3 overflow-y-auto pr-1">
        {assignmentEvents.map((event) => {
          const tone = getLandingAssignmentTone(event);
          const topCapabilities = getTopNeedCapabilities(event);

          return (
            <article
              key={`${event.sequence ?? event.timestamp}:${event.tailNumber}:${event.baseId}`}
              className={`rounded-lg border px-4 py-3 shadow-[0_8px_30px_rgba(15,23,42,0.28)] backdrop-blur-xl ${tone.cardClass}`}
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="text-[10px] font-semibold uppercase tracking-[0.22em] text-white/60">
                    Landing assignment
                  </div>
                  <div className="mt-1 text-sm font-semibold text-white/95">
                    {formatTailNumber(event.tailNumber)} {'->'} Base {formatBaseId(event.baseId)}
                  </div>
                  <div className="mt-1 text-xs text-white/70">
                    {event.source === 'human'
                      ? `Operator redirected this aircraft to a new base. ${formatNeedCount(event.needs.length)}.`
                      : `Dispatch assigned this aircraft to the best available base. ${formatNeedCount(event.needs.length)}.`}
                  </div>
                </div>

                <div
                  className={`inline-flex items-center rounded-full px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.18em] shadow-sm ${getAssignmentSourceBadgeClass(event.source)}`}
                >
                  {getAssignmentSourceLabel(event.source)}
                </div>
              </div>

              <div className="mt-3 flex flex-wrap gap-2">
                {topCapabilities.map(({ capability, severity }) => (
                  <span
                    key={capability}
                    className={`rounded-full border px-2 py-0.5 text-[10px] font-medium ${tone.badgeClass}`}
                  >
                    {toTitleCase(capability)}
                    {severity > 0 ? ` ${severity}` : ''}
                  </span>
                ))}
              </div>
            </article>
          );
        })}
      </div>
    </div>,
    portalRoot,
  );
}
