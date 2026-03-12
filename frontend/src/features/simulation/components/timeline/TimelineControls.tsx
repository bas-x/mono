import type { SimulationStatus } from '../../hooks/useSimulationControls';
import { FaPlayCircle } from 'react-icons/fa';
import { FaCirclePause } from 'react-icons/fa6';
import { RxResume } from 'react-icons/rx';
import { useEffect } from 'react';

type FilterToggleProps = {
  label: string;
  isActive: boolean;
  onClick: () => void;
  colorClass: string;
};

function FilterToggle({ label, isActive, onClick, colorClass }: FilterToggleProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex items-center gap-1.5 rounded-full px-2 py-1 text-[10px] font-medium uppercase tracking-wider transition-all border ${
        isActive 
          ? 'border-white/20 bg-white/5 text-white/90 shadow-sm' 
          : 'border-transparent text-white/40 hover:bg-white/5 hover:text-white/60'
      }`}
    >
      <div className={`h-1.5 w-1.5 rounded-full transition-colors ${isActive ? colorClass : 'bg-white/20'}`} />
      {label}
    </button>
  );
}

type TimelineControlsProps = {
  status: SimulationStatus;
  isLoading: boolean;
  onStart: () => void;
  onPause: () => void;
  onResume: () => void;
  filters: Record<string, boolean>;
  onToggleFilter: (type: string) => void;
};

export function TimelineControls({ status, isLoading, onStart, onPause, onResume, filters, onToggleFilter }: TimelineControlsProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      if (e.key.toLowerCase() === 's' && status === 'idle' && !isLoading) {
        onStart();
      } else if (e.key.toLowerCase() === 'p' && status === 'running' && !isLoading) {
        onPause();
      } else if (e.key.toLowerCase() === 'r' && status === 'paused' && !isLoading) {
        onResume();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [status, isLoading, onStart, onPause, onResume]);

  return (
    <div className="flex items-center justify-between border-b border-[color:var(--color-shell-panel-border)] pb-3 relative">
      <div className="flex items-center gap-2 flex-1">
        <FilterToggle 
          label="Ticks" 
          isActive={filters.simulation_step !== false} 
          onClick={() => onToggleFilter('simulation_step')} 
          colorClass="bg-white/60"
        />
        <FilterToggle 
          label="Assignments" 
          isActive={filters.landing_assignment !== false} 
          onClick={() => onToggleFilter('landing_assignment')} 
          colorClass="bg-purple-500"
        />
        <FilterToggle 
          label="State" 
          isActive={filters.aircraft_state_change !== false} 
          onClick={() => onToggleFilter('aircraft_state_change')} 
          colorClass="bg-green-500"
        />
        <FilterToggle 
          label="Threats" 
          isActive={filters.ThreatSpawnedEvent !== false} 
          onClick={() => onToggleFilter('ThreatSpawnedEvent')} 
          colorClass="bg-red-500"
        />
      </div>
      <div className="absolute left-1/2 -translate-x-1/2 flex items-center gap-2">
        <button
          type="button"
          onClick={onStart}
          disabled={status !== 'idle' || isLoading}
          className="relative flex items-center justify-center rounded-lg bg-white/10 px-4 py-2 text-white transition-all hover:bg-white/20 active:scale-95 disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-white/10 min-w-[4rem]"
          title="Start Simulation"
        >
          <FaPlayCircle size={18} />
          <span className="absolute right-1 top-0.5 text-[8px] font-bold text-white/40">
            S
          </span>
        </button>

        <button
          type="button"
          onClick={onPause}
          disabled={status !== 'running' || isLoading}
          className="relative flex items-center justify-center rounded-lg bg-amber-500/20 px-4 py-2 text-amber-500 transition-all hover:bg-amber-500/30 active:scale-95 disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-amber-500/20 min-w-[4rem]"
          title="Pause Simulation"
        >
          <FaCirclePause size={18} />
          <span className="absolute right-1 top-0.5 text-[8px] font-bold text-amber-500/40">
            P
          </span>
        </button>

        <button
          type="button"
          onClick={onResume}
          disabled={status !== 'paused' || isLoading}
          className="relative flex items-center justify-center rounded-lg bg-green-500/20 px-4 py-2 text-green-500 transition-all hover:bg-green-500/30 active:scale-95 disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-green-500/20 min-w-[4rem]"
          title="Resume Simulation"
        >
          <RxResume size={18} />
          <span className="absolute right-1 top-0.5 text-[8px] font-bold text-green-500/40">
            R
          </span>
        </button>
      </div>
      <div className="flex items-center justify-end gap-2 w-[120px]">
        <span className={`flex h-2 w-2 rounded-full ${status === 'running' ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)] animate-pulse' : 'bg-white/30'}`} />
        <span className="text-xs font-medium text-white/50 uppercase tracking-widest">
          {status === 'running' ? 'Live' : status}
        </span>
      </div>
    </div>
  );
}
