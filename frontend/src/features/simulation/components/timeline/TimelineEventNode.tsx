import type { SimulationEvent } from '@/lib/api/types';

type TimelineEventNodeProps = {
  event: SimulationEvent;
  isSelected: boolean;
  onClick: (e: React.MouseEvent<HTMLButtonElement>) => void;
};

export function TimelineEventNode({ event, isSelected, onClick }: TimelineEventNodeProps) {
  const isTick = event.type === 'simulation_step';
  const isAssignment = event.type === 'landing_assignment';
  
  let colorClass = 'bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.6)]';
  let sizeClass = 'h-3 w-3';
  
  if (isTick) {
    colorClass = 'bg-white/20';
    sizeClass = 'h-1.5 w-1.5';
  } else if (isAssignment) {
    colorClass = 'bg-purple-500 shadow-[0_0_8px_rgba(168,85,247,0.6)]';
    sizeClass = 'h-3 w-3';
  } else if (event.type === 'aircraft_state_change') {
    colorClass = 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]';
    sizeClass = 'h-3 w-3';
  } else if (event.type === 'ThreatSpawnedEvent') {
    colorClass = 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.6)]';
    sizeClass = 'h-4 w-4';
  }

  return (
    <button
      type="button"
      onClick={onClick}
      className="group relative flex flex-col items-center justify-center p-2 focus:outline-none"
    >
      <div
        className={`rounded-full transition-all duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] ${sizeClass} ${colorClass} ${
          isSelected 
            ? 'scale-[2] ring-2 ring-white/30 ring-offset-2 ring-offset-black' 
            : 'group-hover:scale-150'
        }`}
      />
      <div className="pointer-events-none absolute -top-8 hidden whitespace-nowrap rounded border border-white/5 bg-black/80 px-2 py-1 text-[10px] text-white/80 opacity-0 shadow-lg backdrop-blur transition-opacity group-hover:block group-hover:opacity-100 z-10">
        {event.type}
      </div>
    </button>
  );
}
