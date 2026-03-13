import { useRef, useState, useEffect, useCallback } from 'react';
import type { SimulationEvent } from '@/lib/api/types';
import { TimelineEventNode } from './TimelineEventNode';
import { TimelineEventDetails } from './TimelineEventDetails';

type TimelineTrackProps = {
  events: SimulationEvent[];
  currentTick: number;
  maxTick: number;
  playbackTick: number | null;
  onScrub: (tick: number | null) => void;
  zoom: number;
  onBeforeScrub?: () => void;
  isLive?: boolean;
};

export function TimelineTrack({ events, currentTick, maxTick, playbackTick, onScrub, zoom, onBeforeScrub, isLive = false }: TimelineTrackProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const trackRef = useRef<HTMLDivElement>(null);
  const [selectedEvent, setSelectedEvent] = useState<SimulationEvent | null>(null);
  const dragRef = useRef({ isDown: false, isDragging: false, startX: 0, prepared: false });
  const [isDraggingUI, setIsDraggingUI] = useState(false);
  const [localTick, setLocalTick] = useState<number | null>(null);
  const scrubRafRef = useRef<number | null>(null);
  const pendingScrubTickRef = useRef<number | null>(null);
  const activeTick = localTick !== null ? localTick : (playbackTick !== null ? playbackTick : currentTick);
  const shouldAnimatePlayhead = !isDraggingUI && !isLive;

  useEffect(() => {
    return () => {
      if (scrubRafRef.current != null) {
        window.cancelAnimationFrame(scrubRafRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (!selectedEvent || !scrollRef.current || !trackRef.current || isDraggingUI) {
      return;
    }

    const selectedTick = selectedEvent.tick as number | undefined;
    if (selectedTick == null || maxTick <= 0) {
      return;
    }

    const percent = Math.min(100, Math.max(0, (selectedTick / maxTick) * 100));
    const trackWidth = trackRef.current.offsetWidth;
    const viewportWidth = scrollRef.current.clientWidth;
    const desiredScrollLeft = (percent / 100) * trackWidth - viewportWidth / 2;

    scrollRef.current.scrollLeft = Math.max(0, desiredScrollLeft);
  }, [isDraggingUI, maxTick, selectedEvent, zoom]);

  const updateScrubberPosition = useCallback((clientX: number, isFinal = false) => {
    if (!trackRef.current || maxTick === 0) return;
    
    const rect = trackRef.current.getBoundingClientRect();
    const x = Math.max(0, Math.min(clientX - rect.left, rect.width));
    const percentage = x / rect.width;
    
    const targetTick = Math.max(0, Math.round(percentage * maxTick));
    const finalTick = targetTick >= maxTick ? null : targetTick;

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
    } else {
      if (scrubRafRef.current != null) {
        window.cancelAnimationFrame(scrubRafRef.current);
        scrubRafRef.current = null;
      }
      setLocalTick(null);
      onScrub(finalTick);
    }
  }, [maxTick, onBeforeScrub, onScrub]);

  const handlePointerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    if ((e.target as HTMLElement).closest('button')) return;

    dragRef.current = {
      isDown: true,
      isDragging: false,
      startX: e.clientX,
      prepared: false,
    };
  };

  const handlePointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!dragRef.current.isDown) return;
    
    if (!dragRef.current.isDragging) {
      if (Math.abs(e.clientX - dragRef.current.startX) > 5) {
        dragRef.current.isDragging = true;
        setIsDraggingUI(true);
        e.currentTarget.setPointerCapture(e.pointerId);
      }
    }

    if (dragRef.current.isDragging) {
      updateScrubberPosition(e.clientX);
    }
  };

  const handlePointerUp = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!dragRef.current.isDown) return;
    
    if (!dragRef.current.isDragging) {
      updateScrubberPosition(e.clientX, true);
    } else {
      updateScrubberPosition(e.clientX, true);
      e.currentTarget.releasePointerCapture(e.pointerId);
    }

    dragRef.current.isDown = false;
    dragRef.current.isDragging = false;
    dragRef.current.prepared = false;
    pendingScrubTickRef.current = null;
    setIsDraggingUI(false);
  };

  const getPositionPercent = (tick: number) => {
    if (maxTick === 0) return 0;
    return Math.min(100, Math.max(0, (tick / maxTick) * 100));
  };

  const progressPercent = getPositionPercent(activeTick);

  return (
    <>
      <div 
        ref={scrollRef}
        className="relative flex h-24 w-full flex-col justify-center overflow-x-auto overflow-y-visible [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden"
      >
        <div className="min-w-full px-4" style={{ width: `${zoom * 100}%` }}>
          <div 
            ref={trackRef}
            className="relative h-2 w-full cursor-pointer rounded-full bg-white/10 touch-none"
            onPointerDown={handlePointerDown}
            onPointerMove={handlePointerMove}
            onPointerUp={handlePointerUp}
            onPointerCancel={handlePointerUp}
          >
            <div 
              className={`absolute left-0 top-0 h-full rounded-full bg-amber-500/80 ${shouldAnimatePlayhead ? 'transition-all duration-100 ease-linear' : 'transition-none'}`}
              style={{ width: `${progressPercent}%` }}
            />

            <div className="absolute left-0 -top-6 text-[8px] font-bold uppercase tracking-widest text-white/50">
              Start
            </div>
            
            <div className="absolute right-0 -top-6 text-[8px] font-bold uppercase tracking-widest text-white/50">
              {activeTick} / {maxTick}
            </div>

            {events.length === 0 && (
              <div className="absolute left-1/2 top-4 -translate-x-1/2 text-[10px] uppercase tracking-widest text-white/30">
                Waiting for events...
              </div>
            )}

            {events.map((evt, idx) => {
              const evtTick = evt.tick as number;
              if (evtTick === undefined) return null;
              
              const percent = getPositionPercent(evtTick);
              
              return (
                <div 
                  key={`${evt.type}-${idx}-${evt.timestamp || idx}`}
                  className="absolute top-1/2 z-10 -translate-y-1/2 -translate-x-1/2"
                  style={{ left: `${percent}%` }}
                >
                  <TimelineEventNode 
                    event={evt} 
                    isSelected={selectedEvent === evt}
                    onClick={(e) => {
                      e.stopPropagation();
                        if (!dragRef.current.isDragging) {
                          setSelectedEvent(evt === selectedEvent ? null : evt);
                          onBeforeScrub?.();
                          onScrub(evtTick);
                        }
                      }}
                  />
                </div>
              );
            })}

            <div 
              className={`absolute top-1/2 z-20 h-6 w-2.5 -translate-y-1/2 -translate-x-1/2 rounded-full bg-white shadow-[0_0_8px_rgba(255,255,255,0.8)] ${shouldAnimatePlayhead ? 'transition-[left,transform] duration-100 ease-linear' : 'transition-none'} ${isDraggingUI ? 'scale-125 cursor-grabbing' : 'hover:scale-110 cursor-grab'}`}
              style={{ left: `${progressPercent}%` }}
            />
          </div>
        </div>
      </div>

      {selectedEvent && (
        <TimelineEventDetails 
          event={selectedEvent} 
          onClose={() => setSelectedEvent(null)} 
        />
      )}
    </>
  );
}
