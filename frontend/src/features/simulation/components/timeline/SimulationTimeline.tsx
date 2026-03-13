import { useState, useEffect } from 'react';
import { TimelineControls } from './TimelineControls';
import { TimelineTrack } from './TimelineTrack';
import { useSimulationControls } from '../../hooks/useSimulationControls';
import { useSimulationEvents } from '../../hooks/useSimulationEvents';
import type { SimulationState } from '../../hooks/useSimulation';

type SimulationTimelineProps = {
  simulationId: string;
  simulationState: SimulationState;
  setPlaybackTick: (tick: number | null) => void;
};

export function SimulationTimeline({
  simulationId,
  simulationState,
  setPlaybackTick,
}: SimulationTimelineProps) {
  const { status, isLoading, start, pause, resume } = useSimulationControls(simulationId);
  const { events } = useSimulationEvents(simulationId, status === 'paused', status === 'idle');

  const [zoom, setZoom] = useState(1);

  const isRunning = simulationState.status === 'running';
  const currentTick = isRunning ? (simulationState.tick ?? 0) : 0;
  const maxTick = isRunning
    ? Math.max(simulationState.untilTick ?? 0, simulationState.maxTick ?? currentTick)
    : 0;
  const playbackTick = isRunning ? (simulationState.playbackTick ?? null) : null;

  useEffect(() => {
    if (!isRunning) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;

      if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
        e.preventDefault();
        
        const step = e.shiftKey ? 50 : 5;
        const startTick = playbackTick !== null ? playbackTick : currentTick;
        let newTick;

        if (e.key === 'ArrowLeft') {
          newTick = Math.max(0, startTick - step);
        } else {
          newTick = Math.min(maxTick, startTick + step);
        }

        if (newTick >= maxTick) {
          setPlaybackTick(null);
        } else {
          setPlaybackTick(newTick);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isRunning, currentTick, maxTick, playbackTick, setPlaybackTick]);

  const [filters, setFilters] = useState<Record<string, boolean>>({
    simulation_step: false,
    landing_assignment: true,
    aircraft_state_change: true,
    threat_spawned: true,
    threat_targeted: true,
    threat_despawned: true,
    all_aircraft_positions: false,
  });

  const handleToggleFilter = (type: string) => {
    setFilters((prev) => ({ ...prev, [type]: !prev[type] }));
  };

  const filteredEvents = events.filter((e) => {
    if (filters[e.type] === false) return false;

    if (e.type === 'all_aircraft_positions' || e.type === 'simulation_step') {
      const t = e.tick as number;
      if (t !== undefined && t % 5 !== 0) {
        return false;
      }
    }

    return true;
  });

  const getStatusColors = () => {
    switch (status) {
      case 'running':
        return {
          container: 'bg-green-100 text-green-700 dark:bg-green-400/10 dark:text-green-400',
          svg: 'fill-green-500 dark:fill-green-400 animate-pulse',
          label: 'Live',
        };
      case 'paused':
        return {
          container: 'bg-amber-100 text-amber-700 dark:bg-amber-400/10 dark:text-amber-400',
          svg: 'fill-amber-500 dark:fill-amber-400',
          label: 'Paused',
        };
      default:
        return {
          container: 'bg-gray-100 text-gray-700 dark:bg-gray-400/10 dark:text-gray-400',
          svg: 'fill-gray-500 dark:fill-gray-400',
          label: 'Idle',
        };
    }
  };

  const statusStyle = getStatusColors();

  return (
    <div className="pointer-events-none relative z-30 flex w-full justify-center px-4 pb-4 pt-10">
      <div className="pointer-events-auto relative flex w-full max-w-7xl flex-col gap-2 rounded-2xl border border-[color:var(--color-shell-panel-border)] bg-[#0a0a0a]/80 p-4 shadow-[0_8px_32px_rgba(0,0,0,0.8)] backdrop-blur-xl">
        <div className="absolute -top-8 right-8 flex h-8 items-center justify-center rounded-t-lg border-x border-t border-[color:var(--color-shell-panel-border)] bg-[#0a0a0a]/90 px-3 backdrop-blur-xl">
          <span
            className={`inline-flex items-center gap-x-1.5 rounded-md px-1.5 py-0.5 text-xs font-medium uppercase tracking-widest ${statusStyle.container}`}
          >
            <svg viewBox="0 0 6 6" aria-hidden="true" className={`size-1.5 ${statusStyle.svg}`}>
              <circle r="3" cx="3" cy="3" />
            </svg>
            {statusStyle.label}
          </span>
        </div>

        <TimelineControls
          status={status}
          isLoading={isLoading}
          onStart={start}
          onPause={pause}
          onResume={resume}
          filters={filters}
          onToggleFilter={handleToggleFilter}
          zoom={zoom}
          onZoomChange={setZoom}
        />
        <TimelineTrack
          events={filteredEvents}
          currentTick={currentTick}
          maxTick={maxTick}
          playbackTick={playbackTick}
          onScrub={setPlaybackTick}
          zoom={zoom}
        />
      </div>
    </div>
  );
}
