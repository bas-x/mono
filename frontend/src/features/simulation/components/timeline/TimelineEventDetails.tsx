import type { SimulationEvent } from '@/lib/api/types';

type TimelineEventDetailsProps = {
  event: SimulationEvent;
  onClose: () => void;
};

export function TimelineEventDetails({ event, onClose }: TimelineEventDetailsProps) {
  // Omit large or redundant fields for quick display
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { type, simulationId, timestamp, ...rest } = event;

  return (
    <div className="absolute bottom-full left-1/2 z-20 mb-4 w-80 -translate-x-1/2 overflow-hidden rounded-xl border border-[color:var(--color-shell-panel-border)] bg-black/90 shadow-2xl backdrop-blur-xl transition-all">
      <div className="flex items-center justify-between border-b border-white/5 bg-[#111] px-4 py-3">
        <div className="flex flex-col">
          <span className="text-xs font-bold text-white/90 tracking-wide">{type}</span>
          {timestamp && (
            <span className="mt-0.5 font-mono text-[10px] text-white/50">
              {new Date(timestamp).toLocaleTimeString()}
            </span>
          )}
        </div>
        <button 
          type="button"
          onClick={onClose}
          className="rounded-full bg-white/5 p-1.5 text-white/50 transition-colors hover:bg-white/10 hover:text-white"
        >
          <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div className="max-h-48 overflow-y-auto p-4">
        <pre className="whitespace-pre-wrap font-mono text-[10px] text-white/70">
          {JSON.stringify(rest, null, 2)}
        </pre>
      </div>
    </div>
  );
}
