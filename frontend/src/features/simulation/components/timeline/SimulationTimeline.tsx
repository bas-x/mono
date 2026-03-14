import { useState, useEffect } from 'react';
import { TimelineControls } from './TimelineControls';
import { TimelineTrack } from './TimelineTrack';
import { useSimulationControls } from '../../hooks/useSimulationControls';
import { type SimulationState, type TerminalSimulationRecord } from '../../hooks/useSimulation';
import type { SimulationEvent, SimulationInfo } from '@/lib/api/types';
import { buildTimelineGraph } from './timelineGraph';

type SimulationTimelineProps = {
  simulationId: string;
  activeSimulationId?: string;
  simulationState: SimulationState;
  events: SimulationEvent[];
  simulations: SimulationInfo[];
  timelineEventsBySimulation: Map<string, SimulationEvent[]>;
  timelineTerminalRecordsBySimulation: Map<string, TerminalSimulationRecord>;
  setPlaybackTick: (tick: number | null) => void;
  onRefresh: () => Promise<void>;
  onSelectSimulation: (simulationId: string) => Promise<void>;
  onBranchFromEvent?: (event: SimulationEvent) => unknown;
};

export function getTimelineSplitMarker(simulationState: SimulationState): {
  tick: number;
  timestamp?: string;
} | null {
  if (simulationState.status !== 'running' || typeof simulationState.splitTick !== 'number') {
    return null;
  }

  return {
    tick: simulationState.splitTick,
    timestamp: simulationState.splitTimestamp,
  };
}

export function SimulationTimeline({
  simulationId,
  activeSimulationId,
  simulationState,
  events,
  simulations,
  timelineEventsBySimulation,
  timelineTerminalRecordsBySimulation,
  setPlaybackTick,
  onRefresh,
  onSelectSimulation,
  onBranchFromEvent,
}: SimulationTimelineProps) {
  const { status, isLoading, start, pause, resume } = useSimulationControls(
    simulationId,
    simulationState.status === 'running'
      ? {
          isRunnerActive: simulationState.isRunnerActive,
          isRunnerPaused: simulationState.isRunnerPaused,
          tick: simulationState.tick,
          untilTick: simulationState.untilTick,
        }
      : undefined,
    onRefresh,
  );
  const [zoom, setZoom] = useState(1);

  const isRunning = simulationState.status === 'running';
  const currentTick = isRunning ? (simulationState.tick ?? 0) : 0;
  const highestEventTick = events.reduce((highest, event) => {
    const tick = typeof event.tick === 'number' ? event.tick : 0;
    return Math.max(highest, tick);
  }, 0);
  const timelineGraph = buildTimelineGraph({
    simulations,
    activeSimulationId,
    activeState: isRunning ? simulationState : undefined,
    eventsBySimulation: timelineEventsBySimulation,
    terminalRecordsBySimulation: timelineTerminalRecordsBySimulation,
  });
  const timelineEndTick = isRunning
    ? Math.max(
        simulationState.untilTick ?? 0,
        highestEventTick,
        simulationState.maxTick ?? currentTick,
        currentTick,
        timelineGraph.globalMaxTick,
      )
    : timelineGraph.globalMaxTick;
  const playbackTick = isRunning ? (simulationState.playbackTick ?? null) : null;
  const visibleTick = playbackTick ?? currentTick;
  const isLivePlayback = status === 'running' && playbackTick == null;
  const [filters, setFilters] = useState<Record<string, boolean>>({
    simulation_step: false,
    landing_assignment: true,
    aircraft_state_change: true,
    threat_spawned: true,
    threat_targeted: true,
    threat_despawned: true,
    all_aircraft_positions: false,
  });
  const filteredTimelineEventsBySimulation = new Map(
    Array.from(timelineEventsBySimulation.entries()).map(([id, laneEvents]) => [
      id,
      laneEvents.filter((event) => {
        if (filters[event.type] === false) return false;

        if (event.type === 'all_aircraft_positions' || event.type === 'simulation_step') {
          const tick = event.tick as number;
          if (tick !== undefined && tick % 5 !== 0) {
            return false;
          }
        }

        return true;
      }),
    ]),
  );

  const handlePause = async () => {
    if (status !== 'running') {
      return;
    }

    const paused = await pause();
    if (paused) {
      setPlaybackTick(visibleTick);
    }
  };

  const handleResume = async () => {
    setPlaybackTick(null);
    await resume();
  };

  const handleBeforeScrub = () => {
    if (status === 'running' && !isLoading) {
      void pause();
    }
  };

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
          newTick = Math.min(timelineEndTick, startTick + step);
        }

        if (newTick >= timelineEndTick) {
          setPlaybackTick(null);
        } else {
          setPlaybackTick(newTick);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isRunning, currentTick, playbackTick, setPlaybackTick, timelineEndTick]);

  const handleToggleFilter = (type: string) => {
    setFilters((prev) => ({ ...prev, [type]: !prev[type] }));
  };

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
      <div className="pointer-events-auto relative flex max-h-[250px] w-full max-w-7xl flex-col gap-2 overflow-visible rounded-2xl border border-[color:var(--color-shell-panel-border)] bg-[#0a0a0a]/80 p-4 shadow-[0_8px_32px_rgba(0,0,0,0.8)] backdrop-blur-xl">
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
          onPause={handlePause}
          onResume={handleResume}
          filters={filters}
          onToggleFilter={handleToggleFilter}
          zoom={zoom}
          onZoomChange={setZoom}
        />
        {/* {displayedTerminalRecord ? (
          <div className="rounded-xl border border-white/10 bg-white/[0.04] px-4 py-3 text-xs text-white/80">
            <div className="flex items-center justify-between gap-3">
              <div className="font-semibold uppercase tracking-widest text-white/90">
                {displayedTerminalRecord.kind === 'ended'
                  ? 'Run completed'
                  : displayedTerminalRecord.reason === 'reset'
                    ? 'Run reset'
                    : 'Branch stopped'}
              </div>
              <div className="font-mono text-[10px] text-white/50">
                {displayedTerminalRecord.summary.completedVisitCount} completed services
              </div>
            </div>
            <div className="mt-2 flex flex-wrap gap-4 font-mono text-[11px] text-white/70">
              <span>Total service time {formatTerminalSummaryDuration(displayedTerminalRecord.summary.totalDurationMs)}</span>
              <span>Average service time {formatTerminalSummaryDuration(displayedTerminalRecord.summary.averageDurationMs)}</span>
            </div>
          </div>
        ) : null} */}
        <TimelineTrack
          simulations={simulations}
          activeSimulationId={activeSimulationId}
          activeSimulationState={isRunning ? simulationState : undefined}
          timelineEventsBySimulation={filteredTimelineEventsBySimulation}
          timelineTerminalRecordsBySimulation={timelineTerminalRecordsBySimulation}
          currentTick={currentTick}
          playbackTick={playbackTick}
          onScrub={setPlaybackTick}
          zoom={zoom}
          onBeforeScrub={handleBeforeScrub}
          isLive={isLivePlayback}
          onSelectSimulation={onSelectSimulation}
          onBranchFromEvent={onBranchFromEvent}
        />
      </div>
    </div>
  );
}
