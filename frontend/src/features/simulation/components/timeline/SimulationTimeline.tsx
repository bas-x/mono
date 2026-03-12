import { useState } from 'react';
import { TimelineControls } from './TimelineControls';
import { TimelineTrack } from './TimelineTrack';
import { useSimulationControls } from '../../hooks/useSimulationControls';
import { useSimulationEvents } from '../../hooks/useSimulationEvents';

type SimulationTimelineProps = {
  simulationId: string;
};

export function SimulationTimeline({ simulationId }: SimulationTimelineProps) {
  const { status, isLoading, start, pause, resume } = useSimulationControls(simulationId);
  const { events } = useSimulationEvents(simulationId, status === 'paused', status === 'idle');

  const [filters, setFilters] = useState<Record<string, boolean>>({
    simulation_step: false,
    landing_assignment: true,
    aircraft_state_change: true,
    ThreatSpawnedEvent: true,
  });

  const handleToggleFilter = (type: string) => {
    setFilters((prev) => ({ ...prev, [type]: !prev[type] }));
  };

  const filteredEvents = events.filter((e) => filters[e.type] !== false);

  return (
    <div className="pointer-events-none fixed bottom-6 left-0 right-0 z-30 flex justify-center px-4">
      <div className="pointer-events-auto flex w-full max-w-7xl flex-col gap-2 rounded-2xl border border-[color:var(--color-shell-panel-border)] bg-[#0a0a0a]/80 p-4 shadow-[0_8px_32px_rgba(0,0,0,0.8)] backdrop-blur-xl">
        <TimelineControls 
          status={status} 
          isLoading={isLoading} 
          onStart={start} 
          onPause={pause} 
          onResume={resume} 
          filters={filters}
          onToggleFilter={handleToggleFilter}
        />
        <TimelineTrack events={filteredEvents} />
      </div>
    </div>
  );
}
