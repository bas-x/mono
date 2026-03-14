import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type PointerEvent as ReactPointerEvent,
} from 'react';

import type { SimulationEvent, SimulationInfo } from '@/lib/api/types';

import type { SimulationState, TerminalSimulationRecord } from '../../hooks/useSimulation';
import { TimelineEventDetails } from './TimelineEventDetails';
import { TimelineEventNode } from './TimelineEventNode';
import { buildTimelineGraph, clampTimelineTick, getTimelineEventKey } from './timelineGraph';

type TimelineTrackProps = {
  simulations: SimulationInfo[];
  activeSimulationId?: string;
  activeSimulationState?: Extract<SimulationState, { status: 'running' }>;
  timelineEventsBySimulation: Map<string, SimulationEvent[]>;
  timelineTerminalRecordsBySimulation: Map<string, TerminalSimulationRecord>;
  currentTick: number;
  playbackTick: number | null;
  onScrub: (tick: number | null) => void;
  zoom: number;
  onBeforeScrub?: () => void;
  isLive?: boolean;
  onSelectSimulation: (simulationId: string) => unknown;
  onBranchFromEvent?: (event: SimulationEvent) => unknown;
};

function formatSplitLabel(timestamp?: string): string | null {
  if (!timestamp) {
    return null;
  }

  return new Date(timestamp).toLocaleTimeString();
}

export function TimelineTrack({
  simulations,
  activeSimulationId,
  activeSimulationState,
  timelineEventsBySimulation,
  timelineTerminalRecordsBySimulation,
  currentTick,
  playbackTick,
  onScrub,
  zoom,
  onBeforeScrub,
  isLive = false,
  onSelectSimulation,
  onBranchFromEvent,
}: TimelineTrackProps) {
  const graph = useMemo(
    () =>
      buildTimelineGraph({
        simulations,
        activeSimulationId,
        activeState: activeSimulationState,
        eventsBySimulation: timelineEventsBySimulation,
        terminalRecordsBySimulation: timelineTerminalRecordsBySimulation,
      }),
    [
      activeSimulationId,
      activeSimulationState,
      simulations,
      timelineEventsBySimulation,
      timelineTerminalRecordsBySimulation,
    ],
  );

  const scrollRef = useRef<HTMLDivElement>(null);
  const trackRef = useRef<HTMLDivElement>(null);
  const [selectedEvent, setSelectedEvent] = useState<SimulationEvent | null>(null);
  const [selectedLaneId, setSelectedLaneId] = useState<string | null>(activeSimulationId ?? null);
  const [isDraggingUI, setIsDraggingUI] = useState(false);
  const [localTick, setLocalTick] = useState<number | null>(null);
  const scrubRafRef = useRef<number | null>(null);
  const pendingScrubTickRef = useRef<number | null>(null);
  const dragRef = useRef({ isDown: false, isDragging: false, startX: 0, prepared: false });

  useEffect(() => {
    setSelectedLaneId(activeSimulationId ?? null);
  }, [activeSimulationId]);

  useEffect(() => {
    return () => {
      if (scrubRafRef.current != null) {
        window.cancelAnimationFrame(scrubRafRef.current);
      }
    };
  }, []);

  const activeLane = useMemo(() => {
    return graph.lanes.find((lane) => lane.id === activeSimulationId) ?? graph.lanes.at(-1) ?? null;
  }, [activeSimulationId, graph.lanes]);

  const activeTick =
    localTick !== null ? localTick : playbackTick !== null ? playbackTick : currentTick;
  const maxTick = graph.globalMaxTick;
  const minScrubTick = activeLane?.startTick ?? 0;
  const boundedActiveTick = Math.max(minScrubTick, activeTick);
  const progressPercent =
    maxTick === 0 ? 0 : Math.min(100, Math.max(0, (boundedActiveTick / maxTick) * 100));
  const shouldAnimatePlayhead = !isDraggingUI && !isLive;

  useEffect(() => {
    if (!selectedEvent || !scrollRef.current || !trackRef.current || isDraggingUI || maxTick <= 0) {
      return;
    }

    const selectedTick = typeof selectedEvent.tick === 'number' ? selectedEvent.tick : undefined;
    if (selectedTick == null) {
      return;
    }

    const percent = Math.min(100, Math.max(0, (selectedTick / maxTick) * 100));
    const trackWidth = trackRef.current.offsetWidth;
    const viewportWidth = scrollRef.current.clientWidth;
    const desiredScrollLeft = (percent / 100) * trackWidth - viewportWidth / 2;

    scrollRef.current.scrollLeft = Math.max(0, desiredScrollLeft);
  }, [isDraggingUI, maxTick, selectedEvent, zoom]);

  useEffect(() => {
    if (!activeLane) {
      return;
    }

    const clampedTick = clampTimelineTick(playbackTick, activeLane.startTick, maxTick);
    if (clampedTick !== playbackTick) {
      onScrub(clampedTick);
    }
  }, [activeLane, maxTick, onScrub, playbackTick]);

  const getPositionPercent = (tick: number) => {
    if (maxTick === 0) {
      return 0;
    }

    return Math.min(100, Math.max(0, (tick / maxTick) * 100));
  };

  const updateScrubberPosition = (clientX: number, isFinal = false) => {
    if (!trackRef.current || maxTick === 0) {
      return;
    }

    const rect = trackRef.current.getBoundingClientRect();
    const x = Math.max(0, Math.min(clientX - rect.left, rect.width));
    const percentage = x / rect.width;
    const targetTick = Math.max(minScrubTick, Math.round(percentage * maxTick));
    const finalTick = clampTimelineTick(targetTick, minScrubTick, maxTick);

    if (!dragRef.current.prepared) {
      onBeforeScrub?.();
      dragRef.current.prepared = true;
    }

    pendingScrubTickRef.current = finalTick;

    if (dragRef.current.isDragging && !isFinal) {
      setLocalTick(finalTick);

      if (scrubRafRef.current == null) {
        scrubRafRef.current = window.requestAnimationFrame(() => {
          scrubRafRef.current = null;
          onScrub(pendingScrubTickRef.current);
        });
      }
      return;
    }

    if (scrubRafRef.current != null) {
      window.cancelAnimationFrame(scrubRafRef.current);
      scrubRafRef.current = null;
    }

    setLocalTick(null);
    onScrub(finalTick);
  };

  const handlePointerDown = (e: ReactPointerEvent<HTMLDivElement>) => {
    if ((e.target as HTMLElement).closest('button')) {
      return;
    }

    dragRef.current = {
      isDown: true,
      isDragging: false,
      startX: e.clientX,
      prepared: false,
    };
  };

  const handlePointerMove = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (!dragRef.current.isDown) {
      return;
    }

    if (!dragRef.current.isDragging && Math.abs(e.clientX - dragRef.current.startX) > 5) {
      dragRef.current.isDragging = true;
      setIsDraggingUI(true);
      e.currentTarget.setPointerCapture(e.pointerId);
    }

    if (dragRef.current.isDragging) {
      updateScrubberPosition(e.clientX);
    }
  };

  const handlePointerUp = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (!dragRef.current.isDown) {
      return;
    }

    updateScrubberPosition(e.clientX, true);

    if (dragRef.current.isDragging) {
      e.currentTarget.releasePointerCapture(e.pointerId);
    }

    dragRef.current = {
      isDown: false,
      isDragging: false,
      startX: 0,
      prepared: false,
    };
    pendingScrubTickRef.current = null;
    setIsDraggingUI(false);
  };

  const selectedLane = graph.lanes.find(
    (lane) => lane.id === (selectedEvent?.simulationId ?? selectedLaneId ?? activeSimulationId),
  );

  return (
    <>
      <div
        ref={scrollRef}
        className="relative max-h-[150px] w-full overflow-x-auto overflow-y-auto [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden"
      >
        <div className="min-w-full px-4" style={{ width: `${zoom * 100}%` }}>
          <div
            ref={trackRef}
            className="relative min-h-full w-full touch-none"
            onPointerDown={handlePointerDown}
            onPointerMove={handlePointerMove}
            onPointerUp={handlePointerUp}
            onPointerCancel={handlePointerUp}
          >
            <div className="relative flex flex-col gap-4 py-1">
              {graph.lanes.map((lane) => {
                const laneFocused = lane.id === selectedLaneId || lane.isActive;
                const laneSelected = selectedEvent?.simulationId === lane.id;
                const laneProgressPercent = lane.id === activeLane?.id ? progressPercent : null;
                const splitPercent =
                  typeof lane.splitTick === 'number' ? getPositionPercent(lane.splitTick) : null;

                return (
                  <div key={lane.id} className="relative h-16">
                    <button
                      type="button"
                      onClick={() => {
                        setSelectedLaneId(lane.id);
                        void onSelectSimulation(lane.id);
                      }}
                      className={`absolute left-0 top-1/2 z-20 flex w-28 -translate-y-1/2 flex-col items-start rounded-xl border px-3 py-2 text-left transition-all duration-300 ${
                        laneFocused
                          ? 'border-cyan-300/45 bg-cyan-400/12 shadow-[0_0_28px_rgba(34,211,238,0.14)]'
                          : 'border-white/8 bg-white/[0.035] hover:border-white/20 hover:bg-white/[0.06]'
                      }`}
                    >
                      <span className="text-[9px] font-semibold uppercase tracking-[0.34em] text-white/45">
                        {lane.isBase ? 'Primary' : 'Branch'}
                      </span>
                      <span
                        className={`mt-1 text-sm font-semibold ${laneFocused ? 'text-white' : 'text-white/78'}`}
                      >
                        {lane.shortLabel}
                      </span>
                    </button>

                    <div className="ml-32 h-full">
                      <div className="relative h-full">
                        <div
                          className={`pointer-events-none absolute left-0 top-1/2 h-[2px] -translate-y-1/2 rounded-full ${
                            laneFocused
                              ? 'bg-white/14 shadow-[0_0_20px_rgba(125,211,252,0.12)]'
                              : 'bg-white/8'
                          }`}
                          style={{
                            left: `${getPositionPercent(lane.startTick)}%`,
                            width: `${Math.max(0, getPositionPercent(lane.endTick) - getPositionPercent(lane.startTick))}%`,
                          }}
                        />

                        {laneProgressPercent != null ? (
                          <div
                            className={`pointer-events-none absolute top-1/2 h-[4px] -translate-y-1/2 rounded-full bg-amber-400/85 shadow-[0_0_18px_rgba(251,191,36,0.42)] ${shouldAnimatePlayhead ? 'transition-all duration-100 ease-linear' : 'transition-none'}`}
                            style={{
                              left: `${getPositionPercent(lane.startTick)}%`,
                              width: `${Math.max(0, laneProgressPercent - getPositionPercent(lane.startTick))}%`,
                            }}
                          />
                        ) : null}

                        {splitPercent != null ? (
                          <div
                            className="pointer-events-none absolute top-1/2 z-10 -translate-x-1/2 -translate-y-1/2"
                            style={{ left: `${splitPercent}%` }}
                          >
                            <div className="absolute -top-8 left-1/2 -translate-x-1/2 whitespace-nowrap rounded border border-cyan-400/20 bg-cyan-400/10 px-2 py-1 text-[9px] font-semibold uppercase tracking-widest text-cyan-200">
                              Split {lane.splitTick}
                              {formatSplitLabel(lane.splitTimestamp)
                                ? ` · ${formatSplitLabel(lane.splitTimestamp)}`
                                : ''}
                            </div>
                            <div className="h-5 w-px bg-cyan-300/70 shadow-[0_0_8px_rgba(103,232,249,0.45)]" />
                          </div>
                        ) : null}

                        {/* {lane.events.length === 0 ? (
                          <div
                            className="absolute top-1/2 -translate-y-1/2 text-[10px] uppercase tracking-[0.28em] text-white/24"
                            style={{ left: `${getPositionPercent(lane.startTick)}%` }}
                          >
                            Awaiting events
                          </div>
                        ) : null} */}

                        {lane.events.map((event) => {
                          const eventTick = typeof event.tick === 'number' ? event.tick : undefined;
                          if (eventTick == null) {
                            return null;
                          }

                          return (
                            <div
                              key={getTimelineEventKey(event)}
                              className="absolute top-1/2 z-10 -translate-x-1/2 -translate-y-1/2"
                              style={{ left: `${getPositionPercent(eventTick)}%` }}
                            >
                              <TimelineEventNode
                                event={event}
                                isSelected={selectedEvent === event}
                                isDimmed={!laneFocused && !laneSelected}
                                onClick={(clickEvent) => {
                                  clickEvent.stopPropagation();
                                  if (dragRef.current.isDragging) {
                                    return;
                                  }

                                  setSelectedLaneId(lane.id);
                                  setSelectedEvent(event === selectedEvent ? null : event);

                                  if (lane.id === activeLane?.id) {
                                    onBeforeScrub?.();
                                    onScrub(clampTimelineTick(eventTick, lane.startTick, maxTick));
                                  }
                                }}
                              />
                            </div>
                          );
                        })}

                        {lane.terminalRecord ? (
                          <div
                            className="pointer-events-none absolute top-1/2 z-10 -translate-x-1/2 -translate-y-1/2"
                            style={{ left: `${getPositionPercent(lane.terminalRecord.tick)}%` }}
                          >
                            <div
                              className={`h-4 w-4 rounded-full border ${
                                lane.terminalRecord.kind === 'ended'
                                  ? 'border-cyan-300/60 bg-cyan-300/18 shadow-[0_0_14px_rgba(103,232,249,0.38)]'
                                  : 'border-amber-300/60 bg-amber-300/18 shadow-[0_0_14px_rgba(251,191,36,0.34)]'
                              }`}
                            />
                          </div>
                        ) : null}

                        {lane.id === activeLane?.id ? (
                          <div
                            className={`pointer-events-none absolute top-1/2 z-20 h-6 w-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full bg-white shadow-[0_0_10px_rgba(255,255,255,0.8)] ${shouldAnimatePlayhead ? 'transition-[left,transform] duration-100 ease-linear' : 'transition-none'} ${isDraggingUI ? 'scale-125 cursor-grabbing' : 'hover:scale-110 cursor-grab'}`}
                            style={{ left: `${progressPercent}%` }}
                          />
                        ) : null}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>

      {selectedEvent && selectedLane ? (
        <TimelineEventDetails
          event={selectedEvent}
          laneLabel={selectedLane.label}
          canBranchFromEvent={selectedEvent.simulationId === 'base'}
          onBranchFromEvent={onBranchFromEvent}
          onActivateLane={
            selectedLane.id === activeLane?.id
              ? undefined
              : () => void onSelectSimulation(selectedLane.id)
          }
          onClose={() => setSelectedEvent(null)}
        />
      ) : null}
    </>
  );
}
