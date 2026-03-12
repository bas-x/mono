import { useRef, useState, useEffect } from 'react';
import type { SimulationEvent } from '@/lib/api/types';
import { TimelineEventNode } from './TimelineEventNode';
import { TimelineEventDetails } from './TimelineEventDetails';

type TimelineTrackProps = {
  events: SimulationEvent[];
};

export function TimelineTrack({ events }: TimelineTrackProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [selectedEvent, setSelectedEvent] = useState<SimulationEvent | null>(null);

  // Auto-scroll to right when new events arrive
  useEffect(() => {
    if (scrollRef.current && !selectedEvent) {
      scrollRef.current.scrollLeft = scrollRef.current.scrollWidth;
    }
  }, [events.length, selectedEvent]);

  return (
    <div className="relative flex h-20 w-full flex-col justify-center">
      {/* Central line */}
      <div className="absolute left-0 right-0 top-1/2 h-px -translate-y-1/2 bg-white/10" />

      {/* Track container */}
      <div 
        ref={scrollRef}
        className="relative z-10 flex h-full items-center gap-6 overflow-x-auto overflow-y-visible px-4 py-4 [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden"
        style={{ scrollBehavior: 'smooth' }}
      >
        <div className="relative flex flex-col items-center justify-center min-w-[2rem]">
          <div className="h-4 w-1 rounded-full bg-white/40 shadow-[0_0_8px_rgba(255,255,255,0.2)]" />
          <span className="absolute -top-6 text-[8px] font-bold uppercase tracking-widest text-white/50">
            Start
          </span>
        </div>

        {events.length === 0 && (
          <div className="flex-1 text-center text-[10px] uppercase tracking-widest text-white/30">
            Waiting for events...
          </div>
        )}
        {events.map((evt, idx) => (
          <TimelineEventNode 
            key={`${evt.type}-${idx}-${evt.timestamp || idx}`} 
            event={evt} 
            isSelected={selectedEvent === evt}
            onClick={() => setSelectedEvent(evt === selectedEvent ? null : evt)}
          />
        ))}

        {events.length > 0 && (
          <div className="relative flex flex-col items-center justify-center min-w-[2rem] pl-2">
            <div className="h-4 w-1 rounded-full bg-white/40 shadow-[0_0_8px_rgba(255,255,255,0.2)]" />
            <span className="absolute -top-6 text-[8px] font-bold uppercase tracking-widest text-white/50">
              Now
            </span>
          </div>
        )}
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
