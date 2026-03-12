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
};

export function TimelineTrack({ events, currentTick, maxTick, playbackTick, onScrub, zoom }: TimelineTrackProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const trackRef = useRef<HTMLDivElement>(null);
  const [selectedEvent, setSelectedEvent] = useState<SimulationEvent | null>(null);
  const [isDragging, setIsDragging] = useState(false);

  useEffect(() => {
    if (scrollRef.current && !selectedEvent && playbackTick === null && !isDragging) {
      scrollRef.current.scrollLeft = scrollRef.current.scrollWidth;
    }
  }, [events.length, selectedEvent, playbackTick, isDragging, zoom]);

  const updateScrubberPosition = useCallback((clientX: number) => {
    if (!trackRef.current || maxTick === 0) return;
    
    const rect = trackRef.current.getBoundingClientRect();
    const x = Math.max(0, Math.min(clientX - rect.left, rect.width));
    const percentage = x / rect.width;
    
    const clickedTick = Math.max(0, Math.round(percentage * maxTick));
    
    if (clickedTick >= maxTick) {
      onScrub(null);
    } else {
      onScrub(clickedTick);
    }
  }, [maxTick, onScrub]);

  const handlePointerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    setIsDragging(true);
    e.currentTarget.setPointerCapture(e.pointerId);
    updateScrubberPosition(e.clientX);
  };

  const handlePointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!isDragging) return;
    updateScrubberPosition(e.clientX);
  };

  const handlePointerUp = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!isDragging) return;
    setIsDragging(false);
    e.currentTarget.releasePointerCapture(e.pointerId);
  };

  const getPositionPercent = (tick: number) => {
    if (maxTick === 0) return 0;
    return (tick / maxTick) * 100;
  };

  const activeTick = playbackTick !== null ? playbackTick : currentTick;
  const progressPercent = getPositionPercent(activeTick);

  return (
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
            className="absolute left-0 top-0 h-full rounded-full bg-amber-500/80 transition-all duration-200"
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
                className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2"
                style={{ left: `${percent}%` }}
              >
                <TimelineEventNode 
                  event={evt} 
                  isSelected={selectedEvent === evt}
                  onClick={(e) => {
                    e.stopPropagation();
                    setSelectedEvent(evt === selectedEvent ? null : evt);
                    onScrub(evtTick);
                  }}
                />
              </div>
            );
          })}

          <div 
            className={`absolute top-1/2 h-6 w-2.5 -translate-y-1/2 -translate-x-1/2 rounded-full bg-white shadow-[0_0_8px_rgba(255,255,255,0.8)] transition-all duration-200 ${isDragging ? 'scale-125 cursor-grabbing' : 'hover:scale-110 cursor-grab'}`}
            style={{ left: `${progressPercent}%` }}
          />
        </div>
      </div>

      {selectedEvent && (
        <TimelineEventDetails 
          event={selectedEvent} 
          onClose={() => setSelectedEvent(null)} 
        />
      )}
    </div>
  );
}
