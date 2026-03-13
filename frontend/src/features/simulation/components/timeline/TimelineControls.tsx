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

type ZoomToggleProps = {
  label: string;
  isActive: boolean;
  onClick: () => void;
};

function ZoomToggle({ label, isActive, onClick }: ZoomToggleProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded px-1.5 py-0.5 text-[10px] font-bold transition-all ${
        isActive 
          ? 'bg-amber-500/20 text-amber-500' 
          : 'text-white/40 hover:bg-white/10 hover:text-white/80'
      }`}
    >
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
  zoom: number;
  onZoomChange: (z: number) => void;
  durationLabel?: string | null;
};

export function TimelineControls({ status, isLoading, onStart, onPause, onResume, filters, onToggleFilter, zoom, onZoomChange, durationLabel }: TimelineControlsProps) {
  const isFinished = status === 'finished';
  const isResumeLike = status === 'paused' || status === 'resumable';

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      if (e.key.toLowerCase() === 's' && status === 'idle' && !isLoading) {
        onStart();
      } else if (e.key.toLowerCase() === 'p' && status === 'running' && !isLoading) {
        onPause();
      } else if (e.key.toLowerCase() === 'r' && isResumeLike && !isLoading) {
        onResume();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isLoading, isResumeLike, onPause, onResume, onStart, status]);

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
          label="Positions" 
          isActive={filters.all_aircraft_positions !== false} 
          onClick={() => onToggleFilter('all_aircraft_positions')} 
          colorClass="bg-blue-500"
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
          isActive={filters.threat_spawned !== false} 
          onClick={() => {
            onToggleFilter('threat_spawned');
            onToggleFilter('threat_targeted');
            onToggleFilter('threat_despawned');
          }} 
          colorClass="bg-red-500"
        />
      </div>
      <div className="absolute left-1/2 -translate-x-1/2 flex items-center gap-2">
        <button
          type="button"
          onClick={onStart}
          disabled={status !== 'idle' || isLoading || isFinished}
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
          disabled={status !== 'running' || isLoading || isFinished}
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
          disabled={!isResumeLike || isLoading || isFinished}
          className={`relative flex items-center justify-center rounded-lg px-4 py-2 transition-all active:scale-95 min-w-[4rem] disabled:opacity-30 disabled:cursor-not-allowed ${
            isResumeLike
              ? 'bg-[color:var(--color-primary)]/20 text-[color:var(--color-primary)] hover:bg-[color:var(--color-primary)]/30'
              : 'bg-green-500/20 text-green-500 hover:bg-green-500/30 disabled:hover:bg-green-500/20'
          }`}
          title={status === 'resumable' ? 'Continue Simulation' : 'Resume Simulation'}
        >
          <RxResume size={18} />
          <span className="absolute right-1 top-0.5 text-[8px] font-bold text-current/40">
            R
          </span>
        </button>
      </div>
      <div className="flex items-center justify-end flex-1 gap-4">
        {durationLabel ? (
          <div className="hidden xl:flex items-center rounded-lg border border-[color:var(--color-shell-button-border)] bg-[color:var(--color-shell-button-bg)] px-2.5 py-1 text-[10px] uppercase tracking-wider text-[color:var(--color-shell-text-muted)]">
            End {durationLabel}
          </div>
        ) : null}
        <div className="hidden lg:flex items-center gap-1.5 text-[10px] text-white/40">
          <div className="flex gap-0.5 items-center">
            <span className="flex h-4 min-w-[16px] items-center justify-center rounded border border-white/10 bg-white/5 px-1 font-sans text-[10px]">←</span>
            <span className="flex h-4 min-w-[16px] items-center justify-center rounded border border-white/10 bg-white/5 px-1 font-sans text-[10px]">→</span>
          </div>
          <span>Scrub</span>
          <span className="opacity-50 mx-0.5">|</span>
          <div className="flex gap-0.5 items-center">
            <span className="flex h-4 items-center justify-center rounded border border-white/10 bg-white/5 px-1 font-sans text-[9px] tracking-wide">SHIFT</span>
          </div>
          <span>10x</span>
        </div>
        <div className="flex items-center gap-1 rounded-lg bg-white/5 p-1">
          <ZoomToggle label="1x" isActive={zoom === 1} onClick={() => onZoomChange(1)} />
          <ZoomToggle label="2x" isActive={zoom === 2} onClick={() => onZoomChange(2)} />
          <ZoomToggle label="3x" isActive={zoom === 3} onClick={() => onZoomChange(3)} />
          <ZoomToggle label="5x" isActive={zoom === 5} onClick={() => onZoomChange(5)} />
          <ZoomToggle label="10x" isActive={zoom === 10} onClick={() => onZoomChange(10)} />
        </div>
      </div>
    </div>
  );
}
